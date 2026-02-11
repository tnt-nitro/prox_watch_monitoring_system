package journal

import (
	"context"
	"testing"
	"time"
)

func TestFakeReader_Read(t *testing.T) {
	now := time.Now()
	entries := []Entry{
		NewFakeEntry(3, "systemd", now),
		NewFakeEntry(4, "kernel", now.Add(1*time.Minute)),
		NewFakeEntry(6, "systemd", now.Add(2*time.Minute)),
	}

	reader := NewFakeReader(entries)
	ctx := context.Background()

	ch, err := reader.Read(ctx)
	if err != nil {
		t.Fatalf("Read() error = %v, want nil", err)
	}

	var received []Entry
	for entry := range ch {
		received = append(received, entry)
	}

	if len(received) != len(entries) {
		t.Errorf("Read() received %d entries, want %d", len(received), len(entries))
	}

	for i, entry := range received {
		if entry.Priority != entries[i].Priority {
			t.Errorf("Read() entry[%d].Priority = %d, want %d", i, entry.Priority, entries[i].Priority)
		}
		if entry.Source != entries[i].Source {
			t.Errorf("Read() entry[%d].Source = %s, want %s", i, entry.Source, entries[i].Source)
		}
	}
}

func TestFakeReader_Close(t *testing.T) {
	reader := NewFakeReader([]Entry{})
	ctx := context.Background()

	ch, err := reader.Read(ctx)
	if err != nil {
		t.Fatalf("Read() error = %v, want nil", err)
	}

	if err := reader.Close(); err != nil {
		t.Errorf("Close() error = %v, want nil", err)
	}

	// Channel should still be open, but reader is closed
	// Try to read from closed reader should fail
	_, err = reader.Read(ctx)
	if err != ErrReaderClosed {
		t.Errorf("Read() after Close() error = %v, want ErrReaderClosed", err)
	}

	// Close channel
	<-ch
}

func TestFakeReader_ContextCancellation(t *testing.T) {
	now := time.Now()
	entries := []Entry{
		NewFakeEntry(3, "systemd", now),
		NewFakeEntry(4, "kernel", now.Add(1*time.Minute)),
	}

	reader := NewFakeReader(entries)
	ctx, cancel := context.WithCancel(context.Background())

	ch, err := reader.Read(ctx)
	if err != nil {
		t.Fatalf("Read() error = %v, want nil", err)
	}

	// Cancel context after first entry
	<-ch
	cancel()

	// Channel should close
	_, ok := <-ch
	if ok {
		t.Error("Read() channel should be closed after context cancellation")
	}
}

func TestNewFakeEntry(t *testing.T) {
	now := time.Now()
	entry := NewFakeEntry(3, "systemd", now)

	if entry.Priority != 3 {
		t.Errorf("NewFakeEntry() Priority = %d, want 3", entry.Priority)
	}
	if entry.Source != "systemd" {
		t.Errorf("NewFakeEntry() Source = %s, want systemd", entry.Source)
	}
	if !entry.Timestamp.Equal(now) {
		t.Errorf("NewFakeEntry() Timestamp = %v, want %v", entry.Timestamp, now)
	}
}

func TestSystemdReader_Close(t *testing.T) {
	reader := NewSystemdReader()

	if err := reader.Close(); err != nil {
		t.Errorf("Close() error = %v, want nil", err)
	}

	// Try to read from closed reader should fail
	ctx := context.Background()
	_, err := reader.Read(ctx)
	if err != ErrReaderClosed {
		t.Errorf("Read() after Close() error = %v, want ErrReaderClosed", err)
	}
}
