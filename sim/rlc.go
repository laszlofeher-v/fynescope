package sim

import (
	"math"
	"strings"
)

type RlcFilter struct {
	a1, a2     float64
	b0, b1, b2 float64
	x1, x2     float64
	y1, y2     float64
}

func unitMultiplier(unit string) float64 {
	switch unit {
	case "mΩ", "mH", "mF":
		return 1e-3
	case "Ω", "H", "F":
		return 1.0
	case "kΩ":
		return 1e3
	case "MΩ":
		return 1e6
	case "µH", "µF":
		return 1e-6
	case "nF":
		return 1e-9
	case "pF":
		return 1e-12
	default:
		return 1.0
	}
}

func NewRlcFilter(filterType string, r float64, runit string, l float64, lunit string, c float64, cunit string, dt float64) *RlcFilter {
	rVal := r * unitMultiplier(runit)
	lVal := l * unitMultiplier(lunit)
	cVal := c * unitMultiplier(cunit)

	f := &RlcFilter{}
	if dt <= 0 {
		return f
	}

	switch {
	case strings.Contains(filterType, "Lowpass RC") || strings.Contains(filterType, "Lowpass RL"):
		tau := rVal * cVal
		if strings.Contains(filterType, "RL") {
			if rVal == 0 {
				rVal = 1e-6
			}
			tau = lVal / rVal
		}
		a0 := 1.0 + 2.0*tau/dt
		f.b0 = 1.0 / a0
		f.b1 = 1.0 / a0
		f.a1 = (1.0 - 2.0*tau/dt) / a0

	case strings.Contains(filterType, "Highpass RC") || strings.Contains(filterType, "Highpass RL"):
		tau := rVal * cVal
		if strings.Contains(filterType, "RL") {
			if rVal == 0 {
				rVal = 1e-6
			}
			tau = lVal / rVal
		}
		a0 := 1.0 + 2.0*tau/dt
		f.b0 = (2.0 * tau / dt) / a0
		f.b1 = -(2.0 * tau / dt) / a0
		f.a1 = (1.0 - 2.0*tau/dt) / a0

	case strings.Contains(filterType, "Lowpass LC"):
		zeta := 0.2 // slight damping to prevent infinite resonance
		w0 := 1.0 / math.Sqrt(lVal*cVal)
		if math.IsNaN(w0) || math.IsInf(w0, 0) {
			f.b0 = 1.0 // pass through
			break
		}
		K := w0 * dt / 2.0
		norm := 1.0 + 2.0*zeta*K + K*K
		f.b0 = (K * K) / norm
		f.b1 = 2.0 * f.b0
		f.b2 = f.b0
		f.a1 = 2.0 * (K*K - 1.0) / norm
		f.a2 = (1.0 - 2.0*zeta*K + K*K) / norm

	case strings.Contains(filterType, "Highpass LC"):
		zeta := 0.2
		w0 := 1.0 / math.Sqrt(lVal*cVal)
		if math.IsNaN(w0) || math.IsInf(w0, 0) {
			f.b0 = 1.0 // pass through
			break
		}
		K := w0 * dt / 2.0
		norm := 1.0 + 2.0*zeta*K + K*K
		f.b0 = 1.0 / norm
		f.b1 = -2.0 / norm
		f.b2 = 1.0 / norm
		f.a1 = 2.0 * (K*K - 1.0) / norm
		f.a2 = (1.0 - 2.0*zeta*K + K*K) / norm
	}

	return f
}

func (f *RlcFilter) Step(x0 float64) float64 {
	y0 := f.b0*x0 + f.b1*f.x1 + f.b2*f.x2 - f.a1*f.y1 - f.a2*f.y2
	f.x2 = f.x1
	f.x1 = x0
	f.y2 = f.y1
	f.y1 = y0
	return y0
}
