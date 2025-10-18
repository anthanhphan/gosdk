// Copyright (c) 2025 anthanhphan <an.thanhphan.work@gmail.com>

package logger

// Level represents the log level for filtering log messages.
type Level string

func (l Level) isValid() bool {
	return l == LevelDebug || l == LevelInfo || l == LevelWarn || l == LevelError
}

func levelValues() []string {
	return []string{string(LevelDebug), string(LevelInfo), string(LevelWarn), string(LevelError)}
}

// Encoding represents the output format for log messages.
type Encoding string

func (e Encoding) isValid() bool {
	return e == EncodingJSON || e == EncodingConsole
}

func encodingValues() []string {
	return []string{string(EncodingJSON), string(EncodingConsole)}
}
