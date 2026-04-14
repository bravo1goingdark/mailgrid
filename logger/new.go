package logger

import (
	"log"
	"os"

	"github.com/sirupsen/logrus"
)

// Init configures the global logrus instance and redirects the stdlib log
// package through it. Call once at program startup before any logging occurs.
//
//   - level:  "debug", "info", "warn", "error" (default "info" for unknown values)
//   - format: "json" for structured JSON output; anything else uses text format
func Init(level, format string) {
	// Set log level
	lvl, err := logrus.ParseLevel(level)
	if err != nil {
		lvl = logrus.InfoLevel
	}
	logrus.SetLevel(lvl)

	// Set formatter
	if format == "json" {
		logrus.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: "2006-01-02T15:04:05Z07:00",
		})
	} else {
		logrus.SetFormatter(&logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: "2006-01-02T15:04:05",
		})
	}

	logrus.SetOutput(os.Stderr)

	// Redirect the stdlib log package through logrus so that existing
	// log.Printf calls are formatted consistently.
	log.SetOutput(logrus.StandardLogger().Writer())
	log.SetFlags(0) // logrus supplies its own timestamp
}

// Logger returns a minimal logger compatible with the scheduler's Logger interface.
// The name is used as a structured "component" field on every log entry.
func New(name string) interface {
	Infof(string, ...any)
	Warnf(string, ...any)
	Errorf(string, ...any)
} {
	return &structLogger{
		entry: logrus.WithField("component", name),
	}
}

type structLogger struct {
	entry *logrus.Entry
}

func (l *structLogger) Infof(format string, v ...any)  { l.entry.Infof(format, v...) }
func (l *structLogger) Warnf(format string, v ...any)  { l.entry.Warnf(format, v...) }
func (l *structLogger) Errorf(format string, v ...any) { l.entry.Errorf(format, v...) }
