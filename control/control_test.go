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

func TestGeneratorDesc_Equals(t *testing.T) {
	g1 := &GeneratorDesc{
		OffsetVoltage:     100,
		PkToPK:            200,
		ArbitraryWaveform: []int16{1, 2, 3},
	}
	g2 := &GeneratorDesc{
		OffsetVoltage:     100,
		PkToPK:            200,
		ArbitraryWaveform: []int16{1, 2, 3},
	}
	if !g1.Equals(g2) {
		t.Errorf("expected g1 equals g2")
	}

	g2.PkToPK = 300
	if g1.Equals(g2) {
		t.Errorf("expected g1 not equals g2 after PkToPK change")
	}

	g2.PkToPK = 200
	g2.ArbitraryWaveform = []int16{1, 2}
	if g1.Equals(g2) {
		t.Errorf("expected g1 not equals g2 with different waveform length")
	}

	g2.ArbitraryWaveform = []int16{1, 2, 4}
	if g1.Equals(g2) {
		t.Errorf("expected g1 not equals g2 with different waveform data")
	}
}
