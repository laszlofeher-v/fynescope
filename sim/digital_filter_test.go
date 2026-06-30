package sim

import (
	"math"
	"testing"
)

func TestSimDigitalFilter(t *testing.T) {
	// Initialize a channelDesc for channel 0
	ch := 0
	dt := 1.0 / 1000.0 // 1 kHz sampling rate

	t.Run("Lowpass Filter", func(t *testing.T) {
		channels[ch] = channelDesc{
			dfLpEnabled: true,
			dfLpFc:      10.0, // 10 Hz cutoff
		}

		filter := NewSimDigitalFilter(ch, dt)
		if !filter.lpEnabled {
			t.Fatal("Expected lowpass to be enabled")
		}

		// Input a step signal from 0 to 1
		filter.Init(0.0)
		var val float64
		for i := 0; i < 500; i++ {
			val = filter.Step(1.0)
		}

		// Lowpass filter should eventually reach 1.0
		if math.Abs(val-1.0) > 0.1 {
			t.Errorf("Expected lowpass output to reach close to 1.0, got %f", val)
		}
	})

	t.Run("Highpass Filter", func(t *testing.T) {
		channels[ch] = channelDesc{
			dfHpEnabled: true,
			dfHpFc:      10.0, // 10 Hz cutoff
		}

		filter := NewSimDigitalFilter(ch, dt)
		if !filter.hpEnabled {
			t.Fatal("Expected highpass to be enabled")
		}

		// Input a step signal from 0 to 1
		filter.Init(1.0)
		var val float64
		for i := 0; i < 500; i++ {
			val = filter.Step(1.0)
		}

		// Highpass filter should block the DC step and return to ~0
		if math.Abs(val) > 0.1 {
			t.Errorf("Expected highpass output to block DC and reach close to 0, got %f", val)
		}
	})

	t.Run("Bandpass Filter", func(t *testing.T) {
		channels[ch] = channelDesc{
			dfBpEnabled: true,
			dfBpFc1:     40.0,
			dfBpFc2:     60.0,
		}

		filter := NewSimDigitalFilter(ch, dt)
		if !filter.bpEnabled {
			t.Fatal("Expected bandpass to be enabled")
		}

		// Test a 50 Hz sine wave (center frequency)
		filter.Init(0.0)
		var maxAmp float64
		for i := 0; i < 500; i++ {
			time := float64(i) * dt
			val := filter.Step(math.Sin(2 * math.Pi * 50.0 * time))
			if math.Abs(val) > maxAmp {
				maxAmp = math.Abs(val)
			}
		}

		// Center frequency should pass through with high amplitude
		if maxAmp < 0.2 {
			t.Errorf("Expected bandpass center frequency to pass through, got max amplitude %f", maxAmp)
		}
	})

	t.Run("Bandstop Filter", func(t *testing.T) {
		channels[ch] = channelDesc{
			dfBsEnabled: true,
			dfBsFc1:     40.0,
			dfBsFc2:     60.0,
		}

		filter := NewSimDigitalFilter(ch, dt)
		if !filter.bsEnabled {
			t.Fatal("Expected bandstop to be enabled")
		}

		// Test a 50 Hz sine wave (rejected frequency)
		filter.Init(0.0)
		var maxAmp float64
		// Wait for initial transient to settle, then measure amplitude
		for i := 0; i < 200; i++ {
			time := float64(i) * dt
			filter.Step(math.Sin(2 * math.Pi * 50.0 * time))
		}
		for i := 200; i < 500; i++ {
			time := float64(i) * dt
			val := filter.Step(math.Sin(2 * math.Pi * 50.0 * time))
			if math.Abs(val) > maxAmp {
				maxAmp = math.Abs(val)
			}
		}

		// Rejected frequency should be strongly attenuated
		if maxAmp > 0.5 {
			t.Errorf("Expected bandstop center frequency to be attenuated, got max amplitude %f", maxAmp)
		}
	})
}
