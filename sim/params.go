package sim

import (
	"math"
	"sync/atomic"
)

// simParams holds the simulator parameters that are written by the GUI goroutine
// and read concurrently by the sampling goroutine. Each field is stored as a
// uint64 bit-pattern and accessed via sync/atomic so that reads and writes are
// always race-free without needing a mutex.
var simParams struct {
	noiseAmplitude       [MaxChannels]atomic.Uint64 // float64 bits
	phaseNoiseDegree     [MaxChannels]atomic.Uint64 // float64 bits
	raiseFallTimePercent atomic.Uint64              // float64 bits
	triggerTimeOffset    atomic.Uint64              // float64 bits
}

// SetNoiseAmplitude sets the noise amplitude (GUI → sampling goroutine).
func SetNoiseAmplitude(ch int, v float64) {
	if ch >= 0 && ch < MaxChannels {
		simParams.noiseAmplitude[ch].Store(math.Float64bits(v))
	}
}

// GetNoiseAmplitude returns the current noise amplitude.
func GetNoiseAmplitude(ch int) float64 {
	if ch >= 0 && ch < MaxChannels {
		return math.Float64frombits(simParams.noiseAmplitude[ch].Load())
	}
	return 0
}

// SetPhaseNoiseDegree sets the phase noise in degrees (GUI → sampling goroutine).
func SetPhaseNoiseDegree(ch int, v float64) {
	if ch >= 0 && ch < MaxChannels {
		simParams.phaseNoiseDegree[ch].Store(math.Float64bits(v))
	}
}

// GetPhaseNoiseDegree returns the current phase noise in degrees.
func GetPhaseNoiseDegree(ch int) float64 {
	if ch >= 0 && ch < MaxChannels {
		return math.Float64frombits(simParams.phaseNoiseDegree[ch].Load())
	}
	return 0
}

// SetRaiseFallTimePercent sets the rise/fall time percent (GUI → sampling goroutine).
func SetRaiseFallTimePercent(v float64) {
	simParams.raiseFallTimePercent.Store(math.Float64bits(v))
}

// GetRaiseFallTimePercent returns the current rise/fall time percent.
func GetRaiseFallTimePercent() float64 {
	return math.Float64frombits(simParams.raiseFallTimePercent.Load())
}

// SetTriggerTimeOffset sets the trigger time offset (sampling goroutine → read goroutine).
func SetTriggerTimeOffset(v float64) {
	simParams.triggerTimeOffset.Store(math.Float64bits(v))
}

// GetTriggerTimeOffset returns the current trigger time offset.
func GetTriggerTimeOffset() float64 {
	return math.Float64frombits(simParams.triggerTimeOffset.Load())
}
