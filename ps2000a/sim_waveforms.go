//go:build demo

package ps2000a

import (
	"fynescope/genericps"
	"log/slog"
	"math"
)

// WaveformGenerator generates waveform values at a given time point.
// The time parameter t is in radians (already multiplied by 2π).
type WaveformGenerator func(t float64, freq float64) float64

// NewWaveformGenerator creates a waveform generator function for the specified wave type.
// The returned function takes a time parameter in radians and returns the waveform value [-1, 1].
func NewWaveformGenerator(waveType genericps.WaveTypeEnum) WaveformGenerator {
	switch waveType {
	case genericps.Sine:
		return sineWave
	case genericps.HalfSine:
		return halfSineWave
	case genericps.Gaussian:
		return gaussianWave
	case genericps.SinC:
		return sinCWave
	case genericps.Square:
		return squareWave
	case genericps.Triangle:
		return triangleWave
	case genericps.RampUp:
		return rampUpWave
	case genericps.RampDown:
		return rampDownWave
	case genericps.DcVoltage:
		return dcVoltageWave
	default:
		slog.Error("Unknown waveType type. Default to sine wave.")
		// Default to sine wave for unknown types
		return sineWave
	}
}

// NewPrbsGenerator returns a WaveformGenerator that produces a PRBS signal.
func NewPrbsGenerator() WaveformGenerator {
	return func(t float64, freq float64) float64 {
		if freq <= 0 {
			return 0
		}
		// Which bit period are we in?
		bitIndex := int64(t / (2 * math.Pi))
		
		// Use a fast, high-quality 64-bit integer hash (SplitMix64) 
		x := uint64(bitIndex)
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
func sineWave(t float64, freq float64) float64 {
	return math.Sin(t)
}

// halfSineWave generates a half-wave rectified sine wave.
func halfSineWave(t float64, freq float64) float64 {
	return math.Abs(math.Sin(t / 2))
}

// gaussianWave generates a Gaussian pulse waveform.
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
func squareWave(t float64, freq float64) float64 {
	x := math.Mod(t, 2*math.Pi)
	if x < 0 {
		x += 2 * math.Pi
	}
	if x < math.Pi {
		return -1.0
	}
	return 1.0
}

// triangleWave generates a triangle wave.
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
func rampUpWave(t float64, freq float64) float64 {
	x := math.Mod(t, 2*math.Pi)
	if x < 0 {
		x += 2 * math.Pi
	}
	return -1.0 + 2.0*(x/(2*math.Pi))
}

// rampDownWave generates a falling sawtooth (ramp down) wave.
func rampDownWave(t float64, freq float64) float64 {
	x := math.Mod(t, 2*math.Pi)
	if x < 0 {
		x += 2 * math.Pi
	}
	return 1.0 - 2.0*(x/(2*math.Pi))
}

// dcVoltageWave generates a DC (constant) voltage.
func dcVoltageWave(t float64, freq float64) float64 {
	return 0
}
