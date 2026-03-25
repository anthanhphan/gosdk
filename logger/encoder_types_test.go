package logger

import (
	"bytes"
	"io"
	"testing"
)

func TestLogger_Encoder_NumbersAndBools(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(&Config{
		LogLevel:    LevelDebug,
		LogEncoding: EncodingJSON,
	}, []io.Writer{AddSync(&buf)})

	// This should cover appendJSONFloat, appendJSONSignedInt, appendJSONUnsignedInt, and Any field resolution
	logger.Debugw("test types",
		"float64", 3.14159,
		"float32", float32(2.718),
		"int", -42,
		"int8", int8(-8),
		"int16", int16(-16),
		"int32", int32(-32),
		"int64", int64(-64),
		"uint", uint(42),
		"uint8", uint8(8),
		"uint16", uint16(16),
		"uint32", uint32(32),
		"uint64", uint64(64),
		"bool", true,
		"false_bool", false,
		"byte", byte(255),
		"rune", rune(100),
		"nil_val", nil,
	)

	logger.Sync()

	output := buf.String()
	if !bytes.Contains(buf.Bytes(), []byte("3.14159")) {
		t.Errorf("expected float64 in output, got: %s", output)
	}
	if !bytes.Contains(buf.Bytes(), []byte("-42")) {
		t.Errorf("expected int in output, got: %s", output)
	}
}

func TestLogger_ConsoleEncoder_Types(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(&Config{
		LogLevel:    LevelDebug,
		LogEncoding: EncodingConsole,
	}, []io.Writer{AddSync(&buf)})

	logger.Debugw("console types",
		"float", 3.14,
		"bool", true,
		"int", -10,
	)

	logger.Sync()
}
