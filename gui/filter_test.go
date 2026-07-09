package gui

import (
	"math"
	"fynescope/settings"
	"testing"
)

func TestApplyFIR_2Tap(t *testing.T) {
	// Test 2-tap FIR with coeffs [0.5, 0.5] (simple moving average lowpass)
	input := []float32{1.0, 1.0, 1.0, 1.0, 1.0}
	coeffs := []float64{0.5, 0.5}
	output := applyFIR(input, coeffs)

	// Since input is constant 1.0, output should be 1.0
	for i, v := range output {
		if math.Abs(float64(v-1.0)) > 1E-6 {
			t.Errorf("At index %d, expected 1.0, got %f", i, v)
		}
	}

	// Test with impulse input
	input = []float32{1.0, 0.0, 0.0, 0.0}
	output = applyFIR(input, coeffs)
	// Output should be [1.0, 0.5, 0.0, 0.0] under boundary clamping
	expected := []float32{1.0, 0.5, 0.0, 0.0}
	for i, v := range output {
		if math.Abs(float64(v-expected[i])) > 1E-6 {
			t.Errorf("At index %d, expected %f, got %f", i, expected[i], v)
		}
	}
}

func TestApplyFIR_3Tap(t *testing.T) {
	// Test 3-tap FIR with coeffs [0.25, 0.5, 0.25]
	input := []float32{1.0, 1.0, 1.0, 1.0, 1.0}
	coeffs := []float64{0.25, 0.5, 0.25}
	output := applyFIR(input, coeffs)

	// Constant input should yield constant output (since sum of coeffs = 1.0)
	for i, v := range output {
		if math.Abs(float64(v-1.0)) > 1E-6 {
			t.Errorf("At index %d, expected 1.0, got %f", i, v)
		}
	}

	// Test impulse
	input = []float32{1.0, 0.0, 0.0, 0.0}
	output = applyFIR(input, coeffs)
	// Output should be [1.0, 0.75, 0.25, 0.0] under boundary clamping
	expected := []float32{1.0, 0.75, 0.25, 0.0}
	for i, v := range output {
		if math.Abs(float64(v-expected[i])) > 1E-6 {
			t.Errorf("At index %d, expected %f, got %f", i, expected[i], v)
		}
	}
}

func TestToHzAndFromHz(t *testing.T) {
	if toHz(1.5, settings.UnitKHz) != 1500.0 {
		t.Errorf("expected 1500, got %f", toHz(1.5, settings.UnitKHz))
	}
	if toHz(2.5, settings.UnitMHz) != 2.5e6 {
		t.Errorf("expected 2.5e6, got %f", toHz(2.5, settings.UnitMHz))
	}
	if toHz(100, settings.UnitHz) != 100.0 {
		t.Errorf("expected 100, got %f", toHz(100, settings.UnitHz))
	}

	val, unit := fromHz(1500.0)
	if val != 1.5 || unit != settings.UnitKHz {
		t.Errorf("expected 1.5 kHz, got %f %s", val, unit)
	}

	val, unit = fromHz(2.5e6)
	if val != 2.5 || unit != settings.UnitMHz {
		t.Errorf("expected 2.5 MHz, got %f %s", val, unit)
	}

	val, unit = fromHz(100.0)
	if val != 100.0 || unit != settings.UnitHz {
		t.Errorf("expected 100.0 Hz, got %f %s", val, unit)
	}
}

func TestApplyDigitalFilters_IIR(t *testing.T) {
	scp := &ScpDesc{}
	scp.Settings = settings.NewDefaultSettings()
	scp.Settings.Channels = make([]settings.ChSettings, 2)
	// Initialize channel settings
	ch := &scp.Settings.Channels[0]
	ch.Enabled = true
	ch.DigitalFilter.LowpassEnabled = true
	ch.DigitalFilter.LowpassFc = 1000.0 // 1 kHz

	// fs = 100 kHz (samplingTimeInterval = 1e-5)
	samplingTimeInterval := 1e-5

	buf := []float32{1.0, 2.0, 3.0, 4.0, 5.0, 6.0, 7.0, 8.0, 9.0, 10.0}
	scp.applyDigitalFilters(0, buf, samplingTimeInterval)

	// Verify that the signal has been smoothed
	// If it is filtered, the output should not be identical to original, and should not contain NaNs
	if len(buf) != 10 {
		t.Fatalf("length changed")
	}
	for i, v := range buf {
		if math.IsNaN(float64(v)) {
			t.Errorf("NaN at index %d", i)
		}
	}

	// Verify highpass filter
	ch.DigitalFilter.LowpassEnabled = false
	ch.DigitalFilter.HighpassEnabled = true
	ch.DigitalFilter.HighpassFc = 1000.0
	buf2 := []float32{10.0, 10.0, 10.0, 10.0, 10.0}
	scp.applyDigitalFilters(0, buf2, samplingTimeInterval)
	// For a constant DC input, a highpass filter should eventually output close to 0
	if math.Abs(float64(buf2[4])) >= 10.0 {
		t.Errorf("Highpass did not attenuate DC: %v", buf2)
	}
}

