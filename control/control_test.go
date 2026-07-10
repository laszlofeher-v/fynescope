package control

import (
	"testing"
)

func TestGoid(t *testing.T) {
	// The exact ID isn't predictable, but we can verify it returns a positive integer
	// and that calling it twice in the same goroutine returns the same ID.
	id1 := goid()
	id2 := goid()

	if id1 <= 0 {
		t.Errorf("Expected positive goroutine ID, got %d", id1)
	}

	if id1 != id2 {
		t.Errorf("Expected identical goroutine IDs for the same goroutine, got %d and %d", id1, id2)
	}

	// Verify a new goroutine gets a different ID
	ch := make(chan int)
	go func() {
		ch <- goid()
	}()

	id3 := <-ch
	if id3 == id1 {
		t.Errorf("Expected different goroutine ID for a new goroutine, got %d for both", id1)
	}
}
