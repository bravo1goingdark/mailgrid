package logger

import "log"

// New returns a minimal logger compatible with the scheduler's Logger interface.
// Name is currently ignored but kept for API compatibility.
func New(_ string) interface{ Infof(string, ...any); Warnf(string, ...any); Errorf(string, ...any) } {
	return &stdLogger{}
}

type stdLogger struct{}

func (*stdLogger) Infof(format string, v ...any)  { log.Printf("INFO: "+format, v...) }
func (*stdLogger) Warnf(format string, v ...any)  { log.Printf("WARN: "+format, v...) }
func (*stdLogger) Errorf(format string, v ...any) { log.Printf("ERROR: "+format, v...) }

