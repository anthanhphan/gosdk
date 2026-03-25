package logger

import (
	"bytes"
	"io"
	"testing"
)

type mockWriteSyncer struct {
	io.Writer
	syncCalled bool
}

func (m *mockWriteSyncer) Sync() error {
	m.syncCalled = true
	return nil
}

func TestLogger_writeEntry_MultipleOutputs(t *testing.T) {
	var buf1, buf2 bytes.Buffer
	logger := NewLogger(&Config{
		LogLevel:    LevelInfo,
		LogEncoding: EncodingJSON,
	}, []io.Writer{&buf1, &buf2})

	logger.Info("multiple outputs test")
	logger.Sync()

	if buf1.Len() == 0 || buf2.Len() == 0 {
		t.Error("writeEntry failed to write to multiple outputs")
	}

	// Test zero outputs
	loggerZero := NewLogger(&Config{
		LogLevel:    LevelInfo,
		LogEncoding: EncodingJSON,
	}, []io.Writer{}) // NewLogger defaults to stdout if empty, let's force empty
	loggerZero.outputs = nil
	loggerZero.Info("zero outputs test") // Should not panic
	loggerZero.Sync()
}

func TestLogger_shouldLog_EdgeCases(t *testing.T) {
	logger := NewLogger(&Config{
		LogLevel: LevelInfo,
	}, []io.Writer{})

	// Test invalid level
	if logger.shouldLog(Level("invalid")) {
		t.Error("shouldLog should return false for invalid log level")
	}

	// Test invalid config level
	logger.config.LogLevel = Level("invalid")
	if !logger.shouldLog(LevelInfo) {
		t.Error("shouldLog should return true if config level is invalid")
	}
}

func TestLogger_stopOutputs_NonBuffered(t *testing.T) {
	var buf bytes.Buffer
	mock := &mockWriteSyncer{Writer: &buf}

	logger := NewLogger(&Config{
		LogLevel: LevelInfo,
	}, []io.Writer{})

	// Inject a non-buffered WriteSyncer
	logger.outputs = []WriteSyncer{mock}

	logger.stopOutputs()

	if !mock.syncCalled {
		t.Error("stopOutputs should call Sync on non-buffered WriteSyncers")
	}
}

func TestLogger_parseKeysAndValues_Odd(t *testing.T) {
	logger := NewLogger(&Config{
		LogLevel: LevelInfo,
	}, []io.Writer{})

	// Test parseKeysAndValues directly via Warnw
	var buf bytes.Buffer
	logger.outputs = []WriteSyncer{AddSync(&buf)}

	// Odd number of keysAndValues
	logger.Warnw("odd args", "key1", "val1", "key2_missing_val")

	if !bytes.Contains(buf.Bytes(), []byte("key2_missing_val")) {
		t.Error("parseKeysAndValues failed to format odd number of arguments")
	}
}
