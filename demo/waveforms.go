package demo

import (
	"log/slog"
	"math"
	"math/rand"
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

// NewArbitraryWaveformGenerator creates a waveform generator that plays back a custom []int16 buffer.
// The data values are expected to be signed 16-bit integers and are scaled to the range [-1.0, 1.0].
func NewArbitraryWaveformGenerator(data []int16) WaveformGenerator {
	if len(data) == 0 {
		return dcVoltageWave
	}
	// To avoid capturing the slice and allocating, we make a copy of the slice header or data.
	// We'll just capture the slice since it is passed from the control message.
	return func(t float64, freq float64) float64 {
		// Wrap phase t (in radians) into [0, 2π)
		phase := math.Mod(t, 2*math.Pi)
		if phase < 0 {
			phase += 2 * math.Pi
		}
		// Map the phase to an array index
		idx := int((phase / (2 * math.Pi)) * float64(len(data)))
		if idx >= len(data) {
			idx = len(data) - 1
		}
		return float64(data[idx]) / 32767.0
	}
}

var prbsBlockOffset uint64

// AdvancePRBS randomizes the PRBS sequence offset for the next simulation block.
// Calling this at the start of each block simulates a continuously running PRBS generator.
func AdvancePRBS() {
	prbsBlockOffset = rand.Uint64()
}

// NewPrbsGenerator returns a WaveformGenerator that produces a PRBS signal
// (Pseudo-Random Binary Sequence) using a 15-bit maximal-length LFSR.
// The bit-clock period is 1/freq, so the configured frequency controls the
// NewPrbsGenerator returns a WaveformGenerator that produces a PRBS signal.
// To ensure it is entirely stateless and can evaluate time non-monotonically
// without corrupting the sequence, it uses a SplitMix64 hash of the bit period index.
// This provides an infinitely non-repeating pseudo-random sequence.
func NewPrbsGenerator() WaveformGenerator {
	return func(t float64, freq float64) float64 {
		if freq <= 0 {
			return 0
		}
		// Which bit period are we in?
		// `t` here is phase in radians: t = t_seconds * freq * 2π
		// So t / (2π) = t_seconds * freq = number of elapsed bit periods.
		// Use math.Floor to ensure correct uniform bit periods across t=0.
		bitIndex := int64(math.Floor(t / (2 * math.Pi)))

		// Use a fast, high-quality 64-bit integer hash (SplitMix64)
		// to deterministically generate a bit for this specific bit index.
		x := uint64(bitIndex) + prbsBlockOffset
		x += 0x9e3779b97f4a7c15 // Weyl constant
		x = (x ^ (x >> 30)) * 0xbf58476d1ce4e5b9
		x = (x ^ (x >> 27)) * 0x94d049bb133111eb
		x = x ^ (x >> 31)

		if x&1 == 1 {
			return 1.0
		}
		return -1.0
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

// NewSpiClockGenerator creates a waveform generator for SPI clock (CPOL=0, CPHA=0).
// Clock pulses high in the middle of each bit period.
func NewSpiClockGenerator() WaveformGenerator {
	return func(t float64, freq float64) float64 {
		if freq <= 0 {
			return -1.0
		}
		// Calculate the fractional part of the bit period.
		frac := math.Mod(t/(2*math.Pi), 1.0)
		if frac < 0 {
			frac += 1.0
		}
		if frac >= 0.25 && frac < 0.75 {
			return 1.0
		}
		return -1.0
	}
}

// NewSpiDataGenerator creates a waveform generator for SPI data (MOSI).
// Outputs a 20-bit pattern (5 hex digits) MSB first continuously.
func NewSpiDataGenerator(data uint32) WaveformGenerator {
	slog.Debug("NewSpiDataGenerator", "data", data)
	return func(t float64, freq float64) float64 {
		if freq <= 0 {
			return -1.0
		}
		bitIndex := int64(math.Floor(t / (2 * math.Pi)))
		// We have 16 bits. MSB first.
		idx := int(bitIndex % 16)
		if idx < 0 {
			idx += 16
		}
		shift := 15 - idx
		bit := (data >> shift) & 1
		if bit == 1 {
			return 1.0
		}
		return -1.0
	}
}
