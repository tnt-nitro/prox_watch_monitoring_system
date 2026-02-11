package watcher

import "testing"

func TestCounter_Initial(t *testing.T) {
	c := NewCounter()
	if c.GetCount() != 0 {
		t.Errorf("Expected initial count to be 0, got %d", c.GetCount())
	}
}

func TestCounter_Increment(t *testing.T) {
	c := NewCounter()
	
	// Increment 1×
	if err := c.Increment(); err != nil {
		t.Fatalf("Increment() failed: %v", err)
	}
	if c.GetCount() != 1 {
		t.Errorf("Expected count to be 1 after increment, got %d", c.GetCount())
	}
}

func TestCounter_IncrementMultiple(t *testing.T) {
	c := NewCounter()
	
	// Increment mehrfach
	for i := 1; i <= 5; i++ {
		if err := c.Increment(); err != nil {
			t.Fatalf("Increment() failed: %v", err)
		}
		if c.GetCount() != i {
			t.Errorf("Expected count to be %d after %d increments, got %d", i, i, c.GetCount())
		}
	}
	
	if c.GetCount() != 5 {
		t.Errorf("Expected final count to be 5, got %d", c.GetCount())
	}
}

func TestCounter_Reset(t *testing.T) {
	c := NewCounter()
	
	// Increment mehrfach
	for i := 0; i < 3; i++ {
		if err := c.Increment(); err != nil {
			t.Fatalf("Increment() failed: %v", err)
		}
	}
	
	if c.GetCount() != 3 {
		t.Errorf("Expected count to be 3 before reset, got %d", c.GetCount())
	}
	
	// Reset
	if err := c.Reset(); err != nil {
		t.Fatalf("Reset() failed: %v", err)
	}
	
	if c.GetCount() != 0 {
		t.Errorf("Expected count to be 0 after reset, got %d", c.GetCount())
	}
}

func TestCounter_NoNegativeState(t *testing.T) {
	c := NewCounter()
	
	// Reset ohne vorheriges Increment
	if err := c.Reset(); err != nil {
		t.Fatalf("Reset() failed: %v", err)
	}
	
	if c.GetCount() < 0 {
		t.Errorf("Counter should never be negative, got %d", c.GetCount())
	}
	
	// Increment nach Reset
	if err := c.Increment(); err != nil {
		t.Fatalf("Increment() failed: %v", err)
	}
	
	if c.GetCount() != 1 {
		t.Errorf("Expected count to be 1 after increment, got %d", c.GetCount())
	}
}

func TestCounter_ResetAfterIncrement(t *testing.T) {
	c := NewCounter()
	
	// Increment → Reset → Increment
	if err := c.Increment(); err != nil {
		t.Fatalf("Increment() failed: %v", err)
	}
	
	if err := c.Reset(); err != nil {
		t.Fatalf("Reset() failed: %v", err)
	}
	
	if err := c.Increment(); err != nil {
		t.Fatalf("Increment() failed: %v", err)
	}
	
	if c.GetCount() != 1 {
		t.Errorf("Expected count to be 1 after reset and increment, got %d", c.GetCount())
	}
}
