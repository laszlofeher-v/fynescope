package demo

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
	filter := NewRlcFilter("Lowpass RC", 1, "kΩ", 0, "H", 1, "µF", 0, dt)

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
	filter := NewRlcFilter("Highpass RC", 1, "kΩ", 0, "H", 1, "µF", 0, dt)

	// Feed DC (0 Hz) -> should block completely
	output := 1.0
	for i := 0; i < 10000; i++ {
		output = filter.Step(1.0)
	}
	if math.Abs(output) > 1e-2 {
		t.Errorf("Highpass RC failed to block DC; output is %v", output)
	}
}

func TestNewRlcFilter_LowpassLC(t *testing.T) {
	dt := 1e-6 // 1 MHz sample rate
	// L = 1 mH, C = 1 µF -> f0 ~ 5 kHz

	// Test 1: Low Q (Q=0.5, critically damped) -> should pass DC with no significant overshoot
	filterCrit := NewRlcFilter("Lowpass LC", 0, "Ω", 1, "mH", 1, "µF", 0.5, dt)
	maxCrit := 0.0
	outputCrit := 0.0
	for i := 0; i < 10000; i++ {
		outputCrit = filterCrit.Step(1.0)
		if outputCrit > maxCrit {
			maxCrit = outputCrit
		}
	}
	if math.Abs(outputCrit-1.0) > 1e-2 {
		t.Errorf("Lowpass LC critically damped attenuated DC signal; output is %v", outputCrit)
	}
	if maxCrit > 1.05 {
		t.Errorf("Lowpass LC critically damped overshot significantly: max = %v", maxCrit)
	}

	// Test 2: Resonant Q (Q=5.0) -> step response has significant resonance/overshoot
	filterRes := NewRlcFilter("Lowpass LC", 0, "Ω", 1, "mH", 1, "µF", 5.0, dt)
	maxRes := 0.0
	for i := 0; i < 10000; i++ {
		out := filterRes.Step(1.0)
		if out > maxRes {
			maxRes = out
		}
	}
	if maxRes < 1.5 {
		t.Errorf("Lowpass LC with Q=5.0 expected overshoot > 1.5, got %v", maxRes)
	}

	// Test 3: Default Q when Q <= 0 (should default to 2.5)
	filterDef := NewRlcFilter("Lowpass LC", 0, "Ω", 1, "mH", 1, "µF", 0, dt)
	filter25 := NewRlcFilter("Lowpass LC", 0, "Ω", 1, "mH", 1, "µF", 2.5, dt)
	for i := 0; i < 50; i++ {
		outDef := filterDef.Step(1.0)
		out25 := filter25.Step(1.0)
		if math.Abs(outDef-out25) > 1e-9 {
			t.Errorf("Step %d: Q=0 (%v) did not match Q=2.5 (%v)", i, outDef, out25)
		}
	}
}

func TestNewRlcFilter_HighpassLC(t *testing.T) {
	dt := 1e-6
	filter := NewRlcFilter("Highpass LC", 0, "Ω", 1, "mH", 1, "µF", 2.5, dt)

	// Feed DC (1.0) -> Highpass should block DC completely
	output := 1.0
	for i := 0; i < 10000; i++ {
		output = filter.Step(1.0)
	}
	if math.Abs(output) > 1e-2 {
		t.Errorf("Highpass LC failed to block DC; output is %v", output)
	}
}

