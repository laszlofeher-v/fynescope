package sim

import (
	"math"
)

type SimDigitalFilter struct {
	lpEnabled bool
	lpAlpha   float64
	lpY       float64

	hpEnabled bool
	hpAlpha   float64
	hpXPrev   float64
	hpY       float64

	bpEnabled bool
	bpOmega0  float64
	bpB0, bpB1, bpB2, bpA1, bpA2 float64
	bpX1, bpX2, bpY1, bpY2 float64

	bsEnabled bool
	bsOmega0  float64
	bsB0, bsB1, bsB2, bsA1, bsA2 float64
	bsX1, bsX2, bsY1, bsY2 float64
}

func NewSimDigitalFilter(ch int, dt float64) *SimDigitalFilter {
	desc := &channels[ch]
	f := &SimDigitalFilter{
		lpEnabled: desc.dfLpEnabled,
		hpEnabled: desc.dfHpEnabled,
		bpEnabled: desc.dfBpEnabled,
		bsEnabled: desc.dfBsEnabled,
	}

	if dt <= 0 {
		return f
	}

	if f.lpEnabled {
		fc := desc.dfLpFc
		f.lpAlpha = 1.0 - math.Exp(-2.0*math.Pi*fc*dt)
	}

	if f.hpEnabled {
		fc := desc.dfHpFc
		f.hpAlpha = math.Exp(-2.0*math.Pi*fc*dt)
	}

	if f.bpEnabled {
		fc1 := desc.dfBpFc1
		fc2 := desc.dfBpFc2
		f0 := 0.5 * (fc1 + fc2)
		bw := fc2 - fc1
		if bw <= 0 {
			bw = 1.0
		}
		omega0 := 2.0 * math.Pi * f0 * dt
		f.bpOmega0 = omega0
		if omega0 < math.Pi {
			q := f0 / bw
			if q < 0.1 {
				q = 0.1
			}
			alpha := math.Sin(omega0) / (2.0 * q)
			b0 := math.Sin(omega0) / 2.0
			a0 := 1.0 + alpha
			a1 := -2.0 * math.Cos(omega0)
			a2 := 1.0 - alpha

			f.bpB0 = b0 / a0
			f.bpB1 = 0.0
			f.bpB2 = -b0 / a0
			f.bpA1 = a1 / a0
			f.bpA2 = a2 / a0
		}
	}

	if f.bsEnabled {
		fc1 := desc.dfBsFc1
		fc2 := desc.dfBsFc2
		f0 := 0.5 * (fc1 + fc2)
		bw := fc2 - fc1
		if bw <= 0 {
			bw = 1.0
		}
		omega0 := 2.0 * math.Pi * f0 * dt
		f.bsOmega0 = omega0
		if omega0 < math.Pi {
			q := f0 / bw
			if q < 0.1 {
				q = 0.1
			}
			alpha := math.Sin(omega0) / (2.0 * q)
			b0 := 1.0
			b1 := -2.0 * math.Cos(omega0)
			b2 := 1.0
			a0 := 1.0 + alpha
			a1 := -2.0 * math.Cos(omega0)
			a2 := 1.0 - alpha

			f.bsB0 = b0 / a0
			f.bsB1 = b1 / a0
			f.bsB2 = b2 / a0
			f.bsA1 = a1 / a0
			f.bsA2 = a2 / a0
		}
	}

	return f
}

func (f *SimDigitalFilter) Init(x float64) {
	f.lpY = x

	f.hpXPrev = x
	f.hpY = 0.0

	f.bpX1 = x
	f.bpX2 = x
	f.bpY1 = 0.0
	f.bpY2 = 0.0

	f.bsX1 = x
	f.bsX2 = x
	f.bsY1 = x
	f.bsY2 = x
}

func (f *SimDigitalFilter) Step(x float64) float64 {
	val := x

	// 1. Lowpass
	if f.lpEnabled {
		f.lpY = f.lpY + f.lpAlpha*(val-f.lpY)
		val = f.lpY
	}

	// 2. Highpass
	if f.hpEnabled {
		if f.hpAlpha < 1.0 {
			y := f.hpAlpha*f.hpY + f.hpAlpha*(val-f.hpXPrev)
			f.hpXPrev = val
			f.hpY = y
			val = y
		} else {
			val = 0.0
		}
	}

	// 3. Bandpass
	if f.bpEnabled && f.bpOmega0 < math.Pi {
		out := f.bpB0*val + f.bpB1*f.bpX1 + f.bpB2*f.bpX2 - f.bpA1*f.bpY1 - f.bpA2*f.bpY2
		f.bpX2 = f.bpX1
		f.bpX1 = val
		f.bpY2 = f.bpY1
		f.bpY1 = out
		val = out
	}

	// 4. Bandstop
	if f.bsEnabled && f.bsOmega0 < math.Pi {
		out := f.bsB0*val + f.bsB1*f.bsX1 + f.bsB2*f.bsX2 - f.bsA1*f.bsY1 - f.bsA2*f.bsY2
		f.bsX2 = f.bsX1
		f.bsX1 = val
		f.bsY2 = f.bsY1
		f.bsY1 = out
		val = out
	}

	return val
}
