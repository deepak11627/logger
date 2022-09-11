package log

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

type Log struct {
	Level      string `json:"level"`
	Msg        string `json:"msg"`
	Service    string `json:"service"`
	Version    string `json:"version"`
	InstanceID string `json:"instance_id"`
}

func TestThatNewLoggerDefaultValues(t *testing.T) {
	buf := new(bytes.Buffer)
	l, err := NewLogger("myservice", "1.0.0", WithOutput(buf))
	if err != nil {
		t.Errorf("failed to create logger instance, error: %s", err)
	}
	assert.Equal(t, defaultLevel, l.config.logLevel)
	assert.Equal(t, "/logs/myservice.log", l.config.logPath)
	assert.Equal(t, maxcount, l.config.rotationCount)
	assert.Equal(t, maxsize, l.config.rotationSize)
	assert.Equal(t, defaultEncoding, l.config.encoding)
}

func TestThatNewLoggerWritesExpectedJSONLog(t *testing.T) {
	buf := new(bytes.Buffer)
	l, err := NewLogger("myservice", "1.0.0", WithOutput(buf))
	if err != nil {
		t.Errorf("failed to create logger instance, error: %s", err)
	}
	l.Info("this is a test message")
	var tmpLog Log
	err = json.Unmarshal(buf.Bytes(), &tmpLog)
	if err != nil {
		t.Errorf("failed to unmarshal the generated log, error: %s", err)
	}
	assert.Equal(t, "INFO", tmpLog.Level)
	assert.Equal(t, "this is a test message", tmpLog.Msg)
	assert.Equal(t, "myservice", tmpLog.Service)
	assert.Equal(t, "1.0.0", tmpLog.Version)
}

func TestThatNewLoggerCreateLoggerWithInstance(t *testing.T) {
	buf := new(bytes.Buffer)
	l, err := NewLogger("myservice", "1.0.0", WithOutput(buf), WithInstanceID("my_instance"))
	if err != nil {
		t.Errorf("failed to create logger instance, error: %s", err)
	}
	l.Info("this is a test message")
	var tmpLog Log
	err = json.Unmarshal(buf.Bytes(), &tmpLog)
	if err != nil {
		t.Errorf("failed to unmarshal the generated log, error: %s", err)
	}
	assert.Equal(t, "INFO", tmpLog.Level)
	assert.Equal(t, "my_instance", tmpLog.InstanceID)
}
