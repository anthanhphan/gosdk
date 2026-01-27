// Copyright (c) 2025 anthanhphan <an.thanhphan.work@gmail.com>

package logger

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

func TestLogger_DirectMethods(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(&Config{
		LogLevel:    LevelDebug,
		LogEncoding: EncodingJSON,
	}, []io.Writer{&buf})

	t.Run("Debug direct call", func(t *testing.T) {
		buf.Reset()
		logger.Debug("test debug")
		if !strings.Contains(buf.String(), "test debug") {
			t.Errorf("Expected debug log, got: %s", buf.String())
		}
	})

	t.Run("Debugf direct call", func(t *testing.T) {
		buf.Reset()
		logger.Debugf("test %s", "debugf")
		if !strings.Contains(buf.String(), "test debugf") {
			t.Errorf("Expected debugf log, got: %s", buf.String())
		}
	})

	t.Run("Debugw direct call", func(t *testing.T) {
		buf.Reset()
		logger.Debugw("test", "key", "debugw")
		if !strings.Contains(buf.String(), "test") || !strings.Contains(buf.String(), "debugw") {
			t.Errorf("Expected debugw log, got: %s", buf.String())
		}
	})

	t.Run("Info direct call", func(t *testing.T) {
		buf.Reset()
		logger.Info("test info")
		if !strings.Contains(buf.String(), "test info") {
			t.Errorf("Expected info log, got: %s", buf.String())
		}
	})

	t.Run("Infof direct call", func(t *testing.T) {
		buf.Reset()
		logger.Infof("test %s", "infof")
		if !strings.Contains(buf.String(), "test infof") {
			t.Errorf("Expected infof log, got: %s", buf.String())
		}
	})

	t.Run("Infow direct call", func(t *testing.T) {
		buf.Reset()
		logger.Infow("test", "key", "infow")
		if !strings.Contains(buf.String(), "test") || !strings.Contains(buf.String(), "infow") {
			t.Errorf("Expected infow log, got: %s", buf.String())
		}
	})

	t.Run("Warn direct call", func(t *testing.T) {
		buf.Reset()
		logger.Warn("test warn")
		if !strings.Contains(buf.String(), "test warn") {
			t.Errorf("Expected warn log, got: %s", buf.String())
		}
	})

	t.Run("Warnf direct call", func(t *testing.T) {
		buf.Reset()
		logger.Warnf("test %s", "warnf")
		if !strings.Contains(buf.String(), "test warnf") {
			t.Errorf("Expected warnf log, got: %s", buf.String())
		}
	})

	t.Run("Warnw direct call", func(t *testing.T) {
		buf.Reset()
		logger.Warnw("test", "key", "warnw")
		if !strings.Contains(buf.String(), "test") || !strings.Contains(buf.String(), "warnw") {
			t.Errorf("Expected warnw log, got: %s", buf.String())
		}
	})

	t.Run("Error direct call", func(t *testing.T) {
		buf.Reset()
		logger.Error("test error")
		if !strings.Contains(buf.String(), "test error") {
			t.Errorf("Expected error log, got: %s", buf.String())
		}
	})

	t.Run("Errorf direct call", func(t *testing.T) {
		buf.Reset()
		logger.Errorf("test %s", "errorf")
		if !strings.Contains(buf.String(), "test errorf") {
			t.Errorf("Expected errorf log, got: %s", buf.String())
		}
	})

	t.Run("Errorw direct call", func(t *testing.T) {
		buf.Reset()
		logger.Errorw("test", "key", "errorw")
		if !strings.Contains(buf.String(), "test") || !strings.Contains(buf.String(), "errorw") {
			t.Errorf("Expected errorw log, got: %s", buf.String())
		}
	})
}

func TestLogger_ShouldLog_EdgeCases(t *testing.T) {
	logger := NewLogger(&Config{
		LogLevel:    LevelInfo,
		LogEncoding: EncodingJSON,
	}, []io.Writer{os.Stdout})

	t.Run("invalid level returns false", func(t *testing.T) {
		if logger.shouldLog(Level("invalid")) {
			t.Error("shouldLog with invalid level should return false")
		}
	})

	t.Run("debug below info returns false", func(t *testing.T) {
		if logger.shouldLog(LevelDebug) {
			t.Error("shouldLog debug when config is info should return false")
		}
	})

	t.Run("error above info returns true", func(t *testing.T) {
		if !logger.shouldLog(LevelError) {
			t.Error("shouldLog error when config is info should return true")
		}
	})
}

func TestLogger_SetClosers(t *testing.T) {
	logger := NewLogger(&Config{
		LogLevel:    LevelInfo,
		LogEncoding: EncodingJSON,
	}, nil)

	t.Run("empty closers does nothing", func(t *testing.T) {
		logger.setClosers(nil)
		if len(logger.closers) != 0 {
			t.Error("setClosers with nil should not add closers")
		}
	})

	t.Run("add closers", func(t *testing.T) {
		var buf bytes.Buffer
		logger.setClosers([]io.Closer{&mockCloser{&buf}})
		if len(logger.closers) != 1 {
			t.Error("setClosers should add closer")
		}
	})
}

type mockCloser struct {
	*bytes.Buffer
}

func (*mockCloser) Close() error {
	return nil
}

func TestLogger_FlushOutputs(t *testing.T) {
	t.Run("flush with file outputs", func(t *testing.T) {
		tmpFile, err := os.CreateTemp("", "test_flush_*.log")
		if err != nil {
			t.Fatal(err)
		}
		defer func() { _ = os.Remove(tmpFile.Name()) }()
		defer func() { _ = tmpFile.Close() }()

		logger := NewLogger(&Config{
			LogLevel:    LevelInfo,
			LogEncoding: EncodingJSON,
		}, []io.Writer{tmpFile})

		logger.Info("test message")
		logger.flushOutputs()
	})
}

func TestLogger_CloseOutputs(t *testing.T) {
	t.Run("close with file outputs", func(t *testing.T) {
		tmpFile, err := os.CreateTemp("", "test_close_*.log")
		if err != nil {
			t.Fatal(err)
		}
		defer func() { _ = os.Remove(tmpFile.Name()) }()

		logger := NewLogger(&Config{
			LogLevel:    LevelInfo,
			LogEncoding: EncodingJSON,
		}, []io.Writer{tmpFile})
		logger.setClosers([]io.Closer{tmpFile})

		logger.closeOutputs()
	})

	t.Run("close with nil closers", func(_ *testing.T) {
		logger := NewLogger(&Config{
			LogLevel:    LevelInfo,
			LogEncoding: EncodingJSON,
		}, nil)
		logger.closeOutputs()
	})
}

func TestAsyncLogger_LogEdgeCases(t *testing.T) {
	var buf bytes.Buffer
	baseLogger := NewLogger(&Config{
		LogLevel:    LevelWarn,
		LogEncoding: EncodingJSON,
	}, []io.Writer{&buf})

	al := NewAsyncLogger(baseLogger, 1)
	defer func() { _ = al.Close() }()

	t.Run("async log with filtered level", func(t *testing.T) {
		al.log(LevelDebug, 0, "should not log")
		al.Flush()
		if strings.Contains(buf.String(), "should not log") {
			t.Error("should filter debug when level is warn")
		}
	})
}
