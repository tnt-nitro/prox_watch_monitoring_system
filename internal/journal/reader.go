package journal

import (
	"context"
	"time"
)

// Reader is the interface for reading journal entries.
type Reader interface {
	// Read starts streaming journal entries.
	// Returns a channel of entries and an error if initialization fails.
	Read(ctx context.Context) (<-chan Entry, error)

	// Close closes the reader and stops streaming.
	Close() error
}

// Entry represents a journal entry (metadata only, no log text).
type Entry struct {
	Priority  int       // syslog priority
	Source    string    // journal source (systemd, kernel, etc.)
	Timestamp time.Time // entry timestamp
	// No log text stored
	// No IPs
	// No hostnames
}

// SystemdReader implements Reader using systemd journal.
type SystemdReader struct {
	closed bool
}

// NewSystemdReader creates a new systemd journal reader.
func NewSystemdReader() *SystemdReader {
	return &SystemdReader{
		closed: false,
	}
}

// Read starts streaming journal entries.
func (r *SystemdReader) Read(ctx context.Context) (<-chan Entry, error) {
	if r.closed {
		return nil, ErrReaderClosed
	}

	entries := make(chan Entry, 100) // Buffered channel

	go func() {
		defer close(entries)

		// Open journal for reading
		// Note: This requires github.com/coreos/go-systemd/v22/journal
		// For MVP, we'll use a simplified approach
		// Full implementation would use journal.Reader

		// Placeholder: In real implementation, this would:
		// 1. Open journal connection
		// 2. Seek to end (follow mode)
		// 3. Read entries in loop
		// 4. Extract only metadata (Priority, Source, Timestamp)
		// 5. Send to channel
		// 6. Handle context cancellation

		// For now, return empty channel
		// This will be implemented when systemd dependency is added
		<-ctx.Done()
	}()

	return entries, nil
}

// Close closes the reader.
func (r *SystemdReader) Close() error {
	r.closed = true
	return nil
}

// Errors
var (
	ErrReaderClosed = &ReaderError{Message: "reader is closed"}
)

// ReaderError represents a reader error.
type ReaderError struct {
	Message string
}

func (e *ReaderError) Error() string {
	return e.Message
}
