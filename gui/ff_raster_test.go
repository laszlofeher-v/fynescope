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

func TestFormatFreqHighRes(t *testing.T) {
	tests := []struct {
		input    float64
		expected string
	}{
		{0, "0 Hz"},
		{100, "100.0 Hz"},
		{123.45, "123.45 Hz"},
		{-123.45, "-123.45 Hz"},
		{1000, "1.00 kHz"},
		{1500, "1.50 kHz"},
		{1525.5, "1.5255 kHz"},
		{-15400, "-15.40 kHz"},
		{13000000, "13.000 MHz"},
		{13456780, "13.45678 MHz"},
		{13456789, "13.45679 MHz"},
		{-13500000, "-13.500 MHz"},
	}

	for _, tc := range tests {
		actual := formatFreqHighRes(tc.input)
		if actual != tc.expected {
			t.Errorf("formatFreqHighRes(%g): expected %q, got %q", tc.input, tc.expected, actual)
		}
	}
}

func TestDecimalsForStep(t *testing.T) {
	tests := []struct {
		step     float64
		expected int
	}{
		{10.0, 0},
		{5.0, 0},
		{1.0, 0},
		{0.5, 1},
		{0.2, 1},
		{0.1, 1},
		{0.05, 2},
		{0.02, 2},
		{0.01, 2},
		{0.005, 3},
		{0.001, 3},
		{0.0001, 4},
	}

	for _, tc := range tests {
		actual := decimalsForStep(tc.step)
		if actual != tc.expected {
			t.Errorf("decimalsForStep(%g): expected %d, got %d", tc.step, tc.expected, actual)
		}
	}
}

func TestFormatDecadeFreq(t *testing.T) {
	tests := []struct {
		input    float64
		expected string
	}{
		{100, "100Hz"},
		{1000, "1kHz"},
		{2000, "2kHz"},
		{1500, "1.5kHz"},
		{1000000, "1MHz"},
		{2000000, "2MHz"},
		{1500000, "1.5MHz"},
		{10000000, "10MHz"},
	}

	for _, tc := range tests {
		actual := formatDecadeFreq(tc.input)
		if actual != tc.expected {
			t.Errorf("formatDecadeFreq(%g): expected %q, got %q", tc.input, tc.expected, actual)
		}
	}
}

