package journal

import (
	"context"
	"time"
)

// FakeReader is a fake reader for testing.
// It generates synthetic entries without accessing systemd journal.
type FakeReader struct {
	entries []Entry
	closed  bool
}

// NewFakeReader creates a new fake reader with predefined entries.
func NewFakeReader(entries []Entry) *FakeReader {
	return &FakeReader{
		entries: entries,
		closed:  false,
	}
}

// Read starts streaming fake journal entries.
func (f *FakeReader) Read(ctx context.Context) (<-chan Entry, error) {
	if f.closed {
		return nil, ErrReaderClosed
	}

	entries := make(chan Entry, len(f.entries))

	go func() {
		defer close(entries)

		for _, entry := range f.entries {
			select {
			case entries <- entry:
			case <-ctx.Done():
				return
			}
		}
	}()

	return entries, nil
}

// Close closes the fake reader.
func (f *FakeReader) Close() error {
	f.closed = true
	return nil
}

// NewFakeEntry creates a fake journal entry for testing.
func NewFakeEntry(priority int, source string, timestamp time.Time) Entry {
	return Entry{
		Priority:  priority,
		Source:    source,
		Timestamp: timestamp,
	}
}
