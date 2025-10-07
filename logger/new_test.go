package logger

import (
	"bytes"
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	logger := New("test-logger")
	assert.NotNil(t, logger)

	// Verify it implements the expected interface
	assert.Implements(t, (*interface {
		Infof(string, ...any)
		Warnf(string, ...any)
		Errorf(string, ...any)
	})(nil), logger)
}

func TestStdLogger_Infof(t *testing.T) {
	logger := New("test")

	var buf bytes.Buffer
	oldOutput := log.Writer()
	log.SetOutput(&buf)
	defer log.SetOutput(oldOutput)

	logger.Infof("This is info: %s", "test")

	output := buf.String()
	assert.Contains(t, output, "INFO: This is info: test")
}

func TestStdLogger_Warnf(t *testing.T) {
	logger := New("test")

	var buf bytes.Buffer
	oldOutput := log.Writer()
	log.SetOutput(&buf)
	defer log.SetOutput(oldOutput)

	logger.Warnf("This is warning: %d", 123)

	output := buf.String()
	assert.Contains(t, output, "WARN: This is warning: 123")
}

func TestStdLogger_Errorf(t *testing.T) {
	logger := New("test")

	var buf bytes.Buffer
	oldOutput := log.Writer()
	log.SetOutput(&buf)
	defer log.SetOutput(oldOutput)

	logger.Errorf("This is error: %v", "test error")

	output := buf.String()
	assert.Contains(t, output, "ERROR: This is error: test error")
}

func TestStdLogger_NoFormatting(t *testing.T) {
	logger := New("test")

	var buf bytes.Buffer
	oldOutput := log.Writer()
	log.SetOutput(&buf)
	defer log.SetOutput(oldOutput)

	logger.Infof("Simple info")
	logger.Warnf("Simple warning")
	logger.Errorf("Simple error")

	output := buf.String()
	assert.Contains(t, output, "INFO: Simple info")
	assert.Contains(t, output, "WARN: Simple warning")
	assert.Contains(t, output, "ERROR: Simple error")
}

func TestNew_IgnoresName(t *testing.T) {
	// The name parameter should be ignored but function should work the same
	logger1 := New("name1")
	logger2 := New("name2")
	logger3 := New("")

	// All should work identically
	var buf1, buf2, buf3 bytes.Buffer
	oldOutput := log.Writer()
	defer log.SetOutput(oldOutput)

	log.SetOutput(&buf1)
	logger1.Infof("test")

	log.SetOutput(&buf2)
	logger2.Infof("test")

	log.SetOutput(&buf3)
	logger3.Infof("test")

	// All should produce the same output pattern
	assert.Contains(t, buf1.String(), "INFO: test")
	assert.Contains(t, buf2.String(), "INFO: test")
	assert.Contains(t, buf3.String(), "INFO: test")
}