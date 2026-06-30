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


