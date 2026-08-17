//go:build demo

package ps2000a

import (
	"fynescope/genericps"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewWaveformGenerator(t *testing.T) {
	tests := []struct {
		name     string
		waveType genericps.WaveTypeEnum
	}{
		{"Sine", genericps.Sine},
		{"HalfSine", genericps.HalfSine},
		{"Gaussian", genericps.Gaussian},
		{"SinC", genericps.SinC},
		{"Square", genericps.Square},
		{"Triangle", genericps.Triangle},
		{"RampUp", genericps.RampUp},
		{"RampDown", genericps.RampDown},
		{"DcVoltage", genericps.DcVoltage},
		{"Unknown (Fallback to Sine)", 999},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := NewWaveformGenerator(tt.waveType)
			assert.NotNil(t, gen)

			// Execute just to ensure no panics
			val := gen(0, 1000)
			// DC voltage returns 0, others generally start at 0, 1, or -1.
			// Here we just ensure we can call it.
			_ = val
		})
	}
}

func TestNewPrbsGenerator(t *testing.T) {
	gen := NewPrbsGenerator()
	assert.NotNil(t, gen)

	// freq <= 0 should return 0
	assert.Equal(t, 0.0, gen(100, 0))

	// Should return either 1.0 or -1.0
	val1 := gen(0, 1000)
	assert.True(t, val1 == 1.0 || val1 == -1.0)

	val2 := gen(1000, 1000)
	assert.True(t, val2 == 1.0 || val2 == -1.0)
}

func TestWaveformsFormulas(t *testing.T) {
	t.Run("sineWave", func(t *testing.T) {
		assert.InDelta(t, 0.0, sineWave(0, 1000), 1e-6)
		assert.InDelta(t, 1.0, sineWave(math.Pi/2, 1000), 1e-6)
		assert.InDelta(t, -1.0, sineWave(3*math.Pi/2, 1000), 1e-6)
	})

	t.Run("halfSineWave", func(t *testing.T) {
		assert.InDelta(t, 0.0, halfSineWave(0, 1000), 1e-6)
		// math.Abs(math.Sin((Pi) / 2)) = 1
		assert.InDelta(t, 1.0, halfSineWave(math.Pi, 1000), 1e-6)
		// negative time
		assert.InDelta(t, 1.0, halfSineWave(-math.Pi, 1000), 1e-6)
	})

	t.Run("squareWave", func(t *testing.T) {
		assert.Equal(t, -1.0, squareWave(0, 1000))
		assert.Equal(t, 1.0, squareWave(math.Pi, 1000))
		assert.Equal(t, -1.0, squareWave(2*math.Pi, 1000))
		assert.Equal(t, 1.0, squareWave(-math.Pi, 1000))
	})

	t.Run("triangleWave", func(t *testing.T) {
		assert.InDelta(t, 0.0, triangleWave(0, 1000), 1e-6)
		assert.InDelta(t, 1.0, triangleWave(math.Pi/2, 1000), 1e-6)
		assert.InDelta(t, 0.0, triangleWave(math.Pi, 1000), 1e-6)
		assert.InDelta(t, -1.0, triangleWave(3*math.Pi/2, 1000), 1e-6)

		// Test negative time
		assert.InDelta(t, 0.0, triangleWave(0, 1000), 1e-6)
		assert.InDelta(t, -1.0, triangleWave(-math.Pi/2, 1000), 1e-6)
	})

	t.Run("rampUpWave", func(t *testing.T) {
		assert.InDelta(t, -1.0, rampUpWave(0, 1000), 1e-6)
		assert.InDelta(t, 0.0, rampUpWave(math.Pi, 1000), 1e-6)
		// Approach 1.0 just before 2*Pi
		assert.InDelta(t, 0.999, rampUpWave(2*math.Pi-0.001*math.Pi, 1000), 1e-2)

		// Negative time
		assert.InDelta(t, 0.0, rampUpWave(-math.Pi, 1000), 1e-6)
	})

	t.Run("rampDownWave", func(t *testing.T) {
		assert.InDelta(t, 1.0, rampDownWave(0, 1000), 1e-6)
		assert.InDelta(t, 0.0, rampDownWave(math.Pi, 1000), 1e-6)
		// Approach -1.0 just before 2*Pi
		assert.InDelta(t, -0.999, rampDownWave(2*math.Pi-0.001*math.Pi, 1000), 1e-2)
	})

	t.Run("dcVoltageWave", func(t *testing.T) {
		assert.Equal(t, 0.0, dcVoltageWave(0, 1000))
		assert.Equal(t, 0.0, dcVoltageWave(123.456, 1000))
	})

	t.Run("sinCWave", func(t *testing.T) {
		assert.InDelta(t, 1.0, sinCWave(0, 1000), 1e-6)
	})

	t.Run("gaussianWave", func(t *testing.T) {
		assert.InDelta(t, 1.0, gaussianWave(0, 1000), 1e-6)
	})
}
