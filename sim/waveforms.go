package sim

import (
	"log/slog"
	"math"
)

// RaiseFallTimePercent is kept for backwards compatibility; use SetRaiseFallTimePercent
// from params.go to change it safely across goroutines.
func init() { SetRaiseFallTimePercent(0.01) } // Default to 1% (T/100)

// WaveformGenerator generates waveform values at a given time point.
// The time parameter t is in radians (already multiplied by 2π).
type WaveformGenerator func(t float64, freq float64) float64

// NewWaveformGenerator creates a waveform generator function for the specified wave type.
// The returned function takes a time parameter in radians and returns the waveform value [-1, 1].
func NewWaveformGenerator(waveType WaveTypeEnum) WaveformGenerator {
	switch waveType {
	case Sine:
		return sineWave
	case HalfSine:
		return halfSineWave
	case Gaussian:
		return gaussianWave
	case SinC:
		return sinCWave
	case Square:
		return squareWave
	case Triangle:
		return triangleWave
	case RampUp:
		return rampUpWave
	case RampDown:
		return rampDownWave
	case DcVoltage:
		return dcVoltageWave
	default:
		slog.Error("Unknown waveType type. Default to sine wave.")
		// Default to sine wave for unknown types
		return sineWave
	}
}

// NewPrbsGenerator returns a WaveformGenerator that produces a PRBS signal
// (Pseudo-Random Binary Sequence) using a 15-bit maximal-length LFSR.
// The bit-clock period is 1/freq, so the configured frequency controls the
// bit-rate. At each bit boundary the LFSR advances one step producing a
// deterministic but pseudo-random ±1 sequence.
func NewPrbsGenerator() WaveformGenerator {
	// 15-bit LFSR with taps at positions 15 and 14 (polynomial x^15 + x^14 + 1)
	lfsr := uint32(1)
	prevBit := float64(0) // current output level (+1 or -1)
	prevBitIndex := int64(-1)

	return func(t float64, freq float64) float64 {
		if freq <= 0 {
			return 0
		}
		// Which bit period are we in?
		bitIndex := int64(t * freq / (2 * math.Pi))
		if bitIndex != prevBitIndex {
			// Advance LFSR for each elapsed bit period
			steps := bitIndex - prevBitIndex
			if steps < 0 {
				steps = 1
			}
			if steps > 65535 {
				steps = 65535
			}
			for i := int64(0); i < steps; i++ {
				// Galois LFSR: tap bits 15 and 14 (1-indexed)
				bit := (lfsr ^ (lfsr >> 1)) & 1
				lfsr = (lfsr >> 1) | (bit << 14)
				if lfsr == 0 {
					lfsr = 1 // avoid zero state
				}
			}
			prevBitIndex = bitIndex
			if lfsr&1 == 1 {
				prevBit = 1.0
			} else {
				prevBit = -1.0
			}
		}
		return prevBit
	}
}

// sineWave generates a standard sine wave.
// Returns values in range [-1, 1].
func sineWave(t float64, freq float64) float64 {
	return math.Sin(t)
}

// halfSineWave generates a half-wave rectified sine wave.
// Returns values in range [0, 1].
func halfSineWave(t float64, freq float64) float64 {
	return math.Abs(math.Sin(t / 2))
}

// gaussianWave generates a Gaussian pulse waveform.
// Returns values in range [-1, 1].
func gaussianWave(t float64, freq float64) float64 {
	const l = 1
	x := math.Mod(t, l*2*math.Pi)
	if x < 0 {
		x = -x
	}
	if x >= l*math.Pi {
		x = l*2*math.Pi - x
	}
	return math.Exp2(-x*x)*2 - 1.0
}

// sinCWave generates a sinc (sin(x)/x) waveform.
// Returns values in range [-1, 1].
func sinCWave(t float64, freq float64) float64 {
	const l = 10
	t = 10 * t
	x := math.Mod(t, l*2*math.Pi)
	if x < 0 {
		x = -x
	}
	if x >= l*math.Pi {
		x = l*2*math.Pi - x
	}
	if x != 0 {
		return math.Sin(x) / x
	}
	return 1
}

// squareWave generates a square wave.
// Returns values in range [-1, 1].
func squareWave(t float64, freq float64) float64 {
	x := math.Mod(t, 2*math.Pi)
	if x < 0 {
		x += 2 * math.Pi
	}
	rfp := GetRaiseFallTimePercent()
	dt := rfp * 2 * math.Pi
	switch {
	case x < dt:
		return 1.0 - 2.0*(x/dt)
	case x < math.Pi:
		return -1.0
	case x < math.Pi+dt:
		return -1.0 + 2.0*((x-math.Pi)/dt)
	default:
		return 1.0
	}
}

// triangleWave generates a triangle wave.
// Returns values in range [-1, 1].
func triangleWave(t float64, freq float64) float64 {
	x := math.Mod(t, 2*math.Pi)
	if x > 0 {
		switch {
		case x <= math.Pi/2:
			return x / (math.Pi / 2)
		case x <= math.Pi+math.Pi/2:
			return (math.Pi/2-x)/(math.Pi/2) + 1
		default:
			return (x-math.Pi)/(math.Pi/2) - 2
		}
	} else {
		x = -x
		switch {
		case x <= math.Pi/2:
			return (math.Pi/2-x)/(math.Pi/2) - 1
		case x <= math.Pi+math.Pi/2:
			return x/(math.Pi/2) - 2
		default:
			return (-x+math.Pi)/(math.Pi/2) + 2
		}
	}
}

// rampUpWave generates a rising sawtooth (ramp up) wave.
// Returns values in range [-1, 1].
func rampUpWave(t float64, freq float64) float64 {
	x := math.Mod(t, 2*math.Pi)
	if x < 0 {
		x += 2 * math.Pi
	}
	rfp := GetRaiseFallTimePercent()
	dt := rfp * 2 * math.Pi
	if x < dt {
		return 1.0 - 2.0*(x/dt)
	}
	return -1.0 + 2.0*((x-dt)/(2*math.Pi-dt))
}

// rampDownWave generates a falling sawtooth (ramp down) wave.
// Returns values in range [-1, 1].
func rampDownWave(t float64, freq float64) float64 {
	x := math.Mod(t, 2*math.Pi)
	if x < 0 {
		x += 2 * math.Pi
	}
	rfp := GetRaiseFallTimePercent()
	dt := rfp * 2 * math.Pi
	if x < dt {
		return -1.0 + 2.0*(x/dt)
	}
	return 1.0 - 2.0*((x-dt)/(2*math.Pi-dt))
}

// dcVoltageWave generates a DC (constant) voltage.
// Always returns 0.
func dcVoltageWave(t float64, freq float64) float64 {
	return 0
}


