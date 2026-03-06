// Copyright (c) 2025 anthanhphan <an.thanhphan.work@gmail.com>

package logger

import (
	"sync"
	"time"
)

// Entry represents a log entry. CallerInfo is inlined to avoid pointer allocation.
type Entry struct {
	Time          time.Time
	Level         Level
	Message       string
	CallerFile    string // empty if caller is disabled
	CallerLine    int
	CallerDefined bool
	Stacktrace    string
	Fields        []Field
}

// maxPooledFields is the maximum field slice capacity retained in the pool.
const maxPooledFields = 32

var entryPool = sync.Pool{
	New: func() any {
		return &Entry{
			Fields: make([]Field, 0, 16),
		}
	},
}

// getEntry retrieves an Entry from the pool and resets it.
func getEntry() *Entry {
	e := entryPool.Get().(*Entry)
	e.Fields = e.Fields[:0]
	e.CallerFile = ""
	e.CallerLine = 0
	e.CallerDefined = false
	e.Stacktrace = ""
	e.Message = ""
	return e
}

// putEntry returns an Entry to the pool for reuse.
func putEntry(e *Entry) {
	if e == nil {
		return
	}
	if cap(e.Fields) > maxPooledFields {
		return
	}
	// Clear field references for GC
	fields := e.Fields[:cap(e.Fields)]
	for i := range fields {
		fields[i] = Field{}
	}
	e.Fields = e.Fields[:0]
	e.CallerFile = ""
	e.CallerDefined = false
	entryPool.Put(e)
}
