package sim

import (
	"math"
	"testing"
)

func TestUnitMultiplier(t *testing.T) {
	tests := []struct {
		unit     string
		expected float64
	}{
		{"mΩ", 1e-3},
		{"Ω", 1.0},
		{"kΩ", 1e3},
		{"MΩ", 1e6},
		{"µH", 1e-6},
		{"nF", 1e-9},
		{"pF", 1e-12},
		{"Unknown", 1.0},
	}

	for _, tt := range tests {
		t.Run(tt.unit, func(t *testing.T) {
			val := unitMultiplier(tt.unit)
			if val != tt.expected {
				t.Errorf("unitMultiplier(%q) = %v, expected %v", tt.unit, val, tt.expected)
			}
		})
	}
}

func TestNewAcCouplingFilter(t *testing.T) {
	// 1 MHz sample rate
	dt := 1e-6
	filter := NewAcCouplingFilter(dt)

	if filter == nil {
		t.Fatal("NewAcCouplingFilter returned nil")
	}

	// Test DC blocking
	// If we feed a constant 1.0 into it, the output should eventually decay to 0.
	output := 1.0
	for i := 0; i < 1000000; i++ { // 1 second of simulation
		output = filter.Step(1.0)
	}

	if math.Abs(output) > 1e-2 {
		t.Errorf("AC coupling filter failed to block DC; output after 1s is %v", output)
	}
}

func TestNewRlcFilter_LowpassRC(t *testing.T) {
	// Lowpass RC filter
	// Cutoff freq = 1 / (2*pi*R*C)
	// R = 1 kOhm, C = 1 uF -> fc ~ 159 Hz
	dt := 1e-5 // 100 kHz sample rate
	filter := NewRlcFilter("Lowpass RC", 1, "kΩ", 0, "H", 1, "µF", dt)

	// Feed DC (0 Hz) -> should pass completely
	output := 0.0
	for i := 0; i < 10000; i++ {
		output = filter.Step(1.0)
	}
	if math.Abs(output-1.0) > 1e-2 {
		t.Errorf("Lowpass RC attenuated DC signal; output is %v", output)
	}
}

func TestNewRlcFilter_HighpassRC(t *testing.T) {
	// Highpass RC filter
	dt := 1e-5
	filter := NewRlcFilter("Highpass RC", 1, "kΩ", 0, "H", 1, "µF", dt)

	// Feed DC (0 Hz) -> should block completely
	output := 1.0
	for i := 0; i < 10000; i++ {
		output = filter.Step(1.0)
	}
	if math.Abs(output) > 1e-2 {
		t.Errorf("Highpass RC failed to block DC; output is %v", output)
	}
}
