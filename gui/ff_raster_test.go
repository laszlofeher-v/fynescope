package gui

import (
	"math"
	"testing"
)

func TestSmoothPhase(t *testing.T) {
	// 1. Empty slice
	t.Run("Empty slice", func(t *testing.T) {
		res := smoothPhase([]bodePoint{}, 5)
		if len(res) != 0 {
			t.Errorf("expected empty result, got len %d", len(res))
		}
	})

	// 2. Constant phase
	t.Run("Constant phase", func(t *testing.T) {
		pts := []bodePoint{
			{phase: 45.0},
			{phase: 45.0},
			{phase: 45.0},
			{phase: 45.0},
			{phase: 45.0},
		}
		res := smoothPhase(pts, 3)
		if len(res) != len(pts) {
			t.Fatalf("expected len %d, got %d", len(pts), len(res))
		}
		for i, val := range res {
			if math.Abs(val-45.0) > 1e-5 {
				t.Errorf("at index %d: expected 45.0, got %f", i, val)
			}
		}
	})

	// 3. Phase wrap around (e.g. crossing 180 degrees)
	t.Run("Phase wrap around", func(t *testing.T) {
		// A simple arithmetic average of 178 and -178 is 0,
		// but vector averaging should yield 180 (or -180).
		pts := []bodePoint{
			{phase: 176.0},
			{phase: 178.0},
			{phase: -178.0},
			{phase: -176.0},
		}
		res := smoothPhase(pts, 3)
		if len(res) != len(pts) {
			t.Fatalf("expected len %d, got %d", len(pts), len(res))
		}
		// The smoothed phase at index 1 (averaged with 176 and -178)
		// should be close to 178.7 degrees.
		// The smoothed phase at index 2 (averaged with 178, -178, -176)
		// should be close to -178.7 degrees (or 180.0 / -180.0 range).
		// We assert that none of them is close to 0 (which would happen with a naive average).
		for i, val := range res {
			if math.Abs(val) < 90.0 {
				t.Errorf("at index %d: expected phase near wrap boundary (abs > 90), got %f", i, val)
			}
		}
	})
}
