package gui

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNumOfMeasurements(t *testing.T) {
	scp := &ScpDesc{}

	// Test maxScreenTime > 0.1
	scp.maxScreenTime = 0.5
	assert.Equal(t, 1, scp.numOfMeasurements(), "Should return 1 when maxScreenTime > 0.1")

	// Test maxScreenTime <= 0.1
	scp.maxScreenTime = 0.1
	assert.Equal(t, 32, scp.numOfMeasurements(), "Should return 32 when maxScreenTime <= 0.1")
}

func TestMeasureFrq(t *testing.T) {
	// Generate a simulated square/sine wave
	buffer := make([]float32, 1000)
	
	// Create a 10Hz signal assuming a 1ms timeInterval (total time 1 second)
	// 10 cycles per 1000 samples -> period = 100 samples
	for i := 0; i < 1000; i++ {
		// Use a sine wave: sin(2 * pi * f * t)
		// f = 10, t = i * 0.001
		buffer[i] = float32(math.Sin(2 * math.Pi * 10 * (float64(i) * 0.001)))
	}

	// mean = 0, top = 0.5, bottom = -0.5
	// timeInterval = 0.001 (1ms between samples)
	frq, period := measureFrq(buffer, 0, 0.5, -0.5, 0.001)

	// Since we generated exactly 10 cycles in 1 second, frequency should be approx 10Hz
	assert.InDelta(t, 10.0, frq, 0.5, "Frequency should be around 10Hz")
	assert.InDelta(t, 0.1, period, 0.01, "Period should be around 0.1s")
}

func TestMeasureFrq_InsufficientData(t *testing.T) {
	buffer := []float32{1.0} // Not enough samples
	frq, period := measureFrq(buffer, 0, 0.5, -0.5, 0.001)
	assert.Equal(t, 0.0, frq)
	assert.Equal(t, 0.0, period)
}
