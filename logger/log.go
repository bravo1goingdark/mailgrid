package logger

import (
	"fmt"
	"log"
	"os"
)

// LogSuccess logs and appends a successful email to CSV.
func LogSuccess(email string, subject string) {
	log.Printf("Sent to %s", email)
	appendToCSV("success.csv", email, subject, "OK")
}

// LogFailure logs and appends a failed email to CSV.
func LogFailure(email string, subject string) {
	log.Printf("Failed permanently: %s", email)
	appendToCSV("failed.csv", email, subject, "Failed")
}

// appendToCSV writes a log entry to the specified CSV file.
func appendToCSV(filename, email, subject, status string) {
	f, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("Could not write to log file %s: %v", filename, err)
		return
	}
	defer func() {
		if err := f.Close(); err != nil {
			log.Printf("Could not close log file %s: %v", filename, err)
		}
	}()

	if _, err := fmt.Fprintf(f, "%s,%s,%s\n", email, subject, status); err != nil {
		log.Printf("Error writing to CSV %s: %v", filename, err)
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
