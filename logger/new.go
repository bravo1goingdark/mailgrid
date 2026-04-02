package logger

import "log"

// Logger returns a minimal logger compatible with the scheduler's Logger interface.
// The name is used as a prefix for all log messages.
func New(name string) interface {
	Infof(string, ...any)
	Warnf(string, ...any)
	Errorf(string, ...any)
} {
	return &stdLogger{
		name: name,
	}
}

type stdLogger struct {
	name string
}

func (l *stdLogger) prefix() string {
	if l.name == "" {
		return ""
	}
	return "[" + l.name + "] "
}

func (l *stdLogger) Infof(format string, v ...any)  { log.Printf("INFO: "+l.prefix()+format, v...) }
func (l *stdLogger) Warnf(format string, v ...any)  { log.Printf("WARN: "+l.prefix()+format, v...) }
func (l *stdLogger) Errorf(format string, v ...any) { log.Printf("ERROR: "+l.prefix()+format, v...) }
