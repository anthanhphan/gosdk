// Copyright (c) 2025 anthanhphan <an.thanhphan.work@gmail.com>

package logger

import (
	"bufio"
	"io"
	"sync"
	"time"
)

// WriteSyncer extends io.Writer with a Sync method for flushing buffered data.
type WriteSyncer interface {
	io.Writer
	Sync() error
}

// AddSync converts an io.Writer to a WriteSyncer. If the writer already
// implements Sync(), it is used directly. Otherwise a no-op Sync is added.
func AddSync(w io.Writer) WriteSyncer {
	if ws, ok := w.(WriteSyncer); ok {
		return ws
	}
	return writerWrapper{Writer: w}
}

type writerWrapper struct {
	io.Writer
}

func (writerWrapper) Sync() error { return nil }

// lockedWriteSyncer wraps a WriteSyncer with a mutex for concurrent safety.
type lockedWriteSyncer struct {
	mu sync.Mutex
	ws WriteSyncer
}

// Lock wraps a WriteSyncer with a mutex, making it safe for concurrent use.
func Lock(ws WriteSyncer) WriteSyncer {
	return &lockedWriteSyncer{ws: ws}
}

func (s *lockedWriteSyncer) Write(p []byte) (int, error) {
	s.mu.Lock()
	n, err := s.ws.Write(p)
	s.mu.Unlock()
	return n, err
}

func (s *lockedWriteSyncer) Sync() error {
	s.mu.Lock()
	err := s.ws.Sync()
	s.mu.Unlock()
	return err
}

const (
	// defaultBufSize is the default buffer size for BufferedWriteSyncer (256KB).
	defaultBufSize = 256 * 1024

	// defaultFlushInterval is how often the buffer auto-flushes.
	defaultFlushInterval = 100 * time.Millisecond
)

// BufferedWriteSyncer buffers writes in memory and flushes to the underlying
// WriteSyncer either when the buffer is full or on a timer. This decouples
// I/O from the caller (request path), avoiding syscall blocking per log entry.
type BufferedWriteSyncer struct {
	underlying WriteSyncer
	mu         sync.Mutex
	buf        *bufio.Writer
	ticker     *time.Ticker
	stop       chan struct{} // signals the flush goroutine to stop
	done       chan struct{} // closed when flush goroutine exits
	stopOnce   sync.Once     // ensures Stop() is idempotent
}

// NewBufferedWriteSyncer creates a BufferedWriteSyncer that wraps the given
// WriteSyncer. Writes are buffered up to bufSize bytes and flushed every
// flushInterval. If bufSize <= 0, defaultBufSize (256KB) is used.
// If flushInterval <= 0, defaultFlushInterval (100ms) is used.
//
// IMPORTANT: Call Stop() before program exit to flush remaining data.
func NewBufferedWriteSyncer(ws WriteSyncer, bufSize int, flushInterval time.Duration) *BufferedWriteSyncer {
	if bufSize <= 0 {
		bufSize = defaultBufSize
	}
	if flushInterval <= 0 {
		flushInterval = defaultFlushInterval
	}

	bws := &BufferedWriteSyncer{
		underlying: ws,
		buf:        bufio.NewWriterSize(ws, bufSize),
		ticker:     time.NewTicker(flushInterval),
		stop:       make(chan struct{}),
		done:       make(chan struct{}),
	}

	go bws.flushLoop()
	return bws
}

func (s *BufferedWriteSyncer) flushLoop() {
	defer close(s.done)
	for {
		select {
		case <-s.ticker.C:
			s.mu.Lock()
			_ = s.buf.Flush()
			s.mu.Unlock()
		case <-s.stop:
			return
		}
	}
}

// Write appends data to the in-memory buffer. If the buffer is full,
// bufio.Writer automatically flushes to the underlying writer.
func (s *BufferedWriteSyncer) Write(p []byte) (int, error) {
	s.mu.Lock()
	n, err := s.buf.Write(p)
	s.mu.Unlock()
	return n, err
}

// Sync flushes the buffer and calls Sync on the underlying writer.
func (s *BufferedWriteSyncer) Sync() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.buf.Flush(); err != nil {
		return err
	}
	return s.underlying.Sync()
}

// Stop flushes any buffered data, stops the flush goroutine, and syncs.
// This must be called before program exit. It is safe to call multiple times.
func (s *BufferedWriteSyncer) Stop() error {
	var err error
	s.stopOnce.Do(func() {
		s.ticker.Stop()
		close(s.stop) // signal goroutine
		<-s.done      // wait for goroutine exit

		// Final flush + sync
		s.mu.Lock()
		defer s.mu.Unlock()

		if flushErr := s.buf.Flush(); flushErr != nil {
			err = flushErr
			return
		}
		err = s.underlying.Sync()
	})
	return err
}
