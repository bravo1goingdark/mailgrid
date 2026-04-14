package logger

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"sync"
)

// csvLogger holds a persistent, buffered file handle for a CSV log file.
// Opening a new file handle on every log call is slow under high concurrency
// and risks interleaved writes. One handle per file, flushed periodically and at shutdown.
type csvLogger struct {
	mu     sync.Mutex
	file   *os.File
	writer *bufio.Writer
}

func newCSVLogger(filename string) (*csvLogger, error) {
	f, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return nil, err
	}
	return &csvLogger{
		file:   f,
		writer: bufio.NewWriterSize(f, 64*1024), // 64 KB write buffer
	}, nil
}

func (l *csvLogger) write(email, subject, status string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if _, err := fmt.Fprintf(l.writer, "%s,%s,%s\n", email, subject, status); err != nil {
		log.Printf("Error writing CSV log: %v", err)
	}
}

func (l *csvLogger) flush() {
	l.mu.Lock()
	defer l.mu.Unlock()
	if err := l.writer.Flush(); err != nil {
		log.Printf("Error flushing CSV log: %v", err)
	}
}

func (l *csvLogger) close() {
	l.flush()
	l.mu.Lock()
	defer l.mu.Unlock()
	if err := l.file.Close(); err != nil {
		log.Printf("Error closing CSV log: %v", err)
	}
}

// Package-level loggers, initialized lazily.
var (
	loggerMu      sync.Mutex
	successLogger *csvLogger
	failedLogger  *csvLogger
)

func getSuccessLogger() *csvLogger {
	loggerMu.Lock()
	defer loggerMu.Unlock()
	if successLogger == nil {
		var err error
		successLogger, err = newCSVLogger("success.csv")
		if err != nil {
			log.Printf("Could not open success.csv: %v", err)
		}
	}
	return successLogger
}

func getFailedLogger() *csvLogger {
	loggerMu.Lock()
	defer loggerMu.Unlock()
	if failedLogger == nil {
		var err error
		failedLogger, err = newCSVLogger("failed.csv")
		if err != nil {
			log.Printf("Could not open failed.csv: %v", err)
		}
	}
	return failedLogger
}

// LogSuccess logs a successful send to stdout and appends to success.csv.
func LogSuccess(email string, subject string) {
	log.Printf("Sent to %s", email)
	if l := getSuccessLogger(); l != nil {
		l.write(email, subject, "OK")
	}
}

// LogFailure logs a permanent failure to stdout and appends to failed.csv.
func LogFailure(email string, subject string) {
	log.Printf("Failed permanently: %s", email)
	if l := getFailedLogger(); l != nil {
		l.write(email, subject, "Failed")
	}
}

// FlushAndClose flushes write buffers and closes all open log file handles.
// Call this once at program exit to ensure all data is written to disk.
func FlushAndClose() {
	loggerMu.Lock()
	s, f := successLogger, failedLogger
	loggerMu.Unlock()

	if s != nil {
		s.close()
	}
	if f != nil {
		f.close()
	}
}

// Errorf logs an error message with formatting.
func Errorf(format string, v ...any) {
	log.Printf("ERROR: "+format, v...)
}

// Warnf logs a warning message with formatting.
func Warnf(format string, v ...any) {
	log.Printf("WARNING: "+format, v...)
}
