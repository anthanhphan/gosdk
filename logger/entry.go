// Copyright (c) 2025 anthanhphan <an.thanhphan.work@gmail.com>

package logger

import "time"

// Entry represents a log entry.
type Entry struct {
	Time       time.Time
	Level      Level
	Message    string
	Caller     *CallerInfo
	Stacktrace string
	Fields     map[string]interface{}
}

// CallerInfo represents information about the caller.
type CallerInfo struct {
	File string
	Line int
}
