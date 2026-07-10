package sim

import (
	"math"
	"testing"
)

func TestNewWaveformGenerator(t *testing.T) {
	tests := []struct {
		waveType WaveTypeEnum
		name     string
	}{
		{Sine, "Sine"},
		{HalfSine, "HalfSine"},
		{Gaussian, "Gaussian"},
		{SinC, "SinC"},
		{Square, "Square"},
		{Triangle, "Triangle"},
		{RampUp, "RampUp"},
		{RampDown, "RampDown"},
		{DcVoltage, "DcVoltage"},
		{WaveTypeEnum(999), "UnknownDefaultToSine"}, // Unknown type defaults to Sine
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := NewWaveformGenerator(tt.waveType)
			if gen == nil {
				t.Errorf("NewWaveformGenerator(%v) returned nil", tt.waveType)
			}
		})
	}
}

func TestWaveforms_Sine(t *testing.T) {
	gen := NewWaveformGenerator(Sine)
	
	valZero := gen(0, 1000)
	if math.Abs(valZero) > 1e-9 {
		t.Errorf("Sine(0) = %v, expected 0", valZero)
	}

	valPi2 := gen(math.Pi/2, 1000)
	if math.Abs(valPi2-1.0) > 1e-9 {
		t.Errorf("Sine(pi/2) = %v, expected 1.0", valPi2)
	}
	
	val3Pi2 := gen(3*math.Pi/2, 1000)
	if math.Abs(val3Pi2-(-1.0)) > 1e-9 {
		t.Errorf("Sine(3pi/2) = %v, expected -1.0", val3Pi2)
	}
}

func TestWaveforms_HalfSine(t *testing.T) {
	gen := NewWaveformGenerator(HalfSine)
	
	// HalfSine uses math.Abs(math.Sin(t/2))
	valPi := gen(math.Pi, 1000)
	if math.Abs(valPi-1.0) > 1e-9 {
		t.Errorf("HalfSine(pi) = %v, expected 1.0", valPi)
	}
	
	val3Pi := gen(3*math.Pi, 1000) // math.Sin(3pi/2) is -1, Abs is 1
	if math.Abs(val3Pi-1.0) > 1e-9 {
		t.Errorf("HalfSine(3pi) = %v, expected 1.0", val3Pi)
	}
}

func TestWaveforms_DcVoltage(t *testing.T) {
	gen := NewWaveformGenerator(DcVoltage)
	for i := 0.0; i < 10.0; i += 1.5 {
		if val := gen(i, 1000); val != 0 {
			t.Errorf("DcVoltage(%v) = %v, expected 0", i, val)
		}
	}
}

func TestWaveforms_PrbsGenerator(t *testing.T) {
	gen := NewPrbsGenerator()
	if gen == nil {
		t.Fatal("NewPrbsGenerator returned nil")
	}

	freq := 1000.0
	// time t where bitIndex will change
	t0 := 0.0
	val0 := gen(t0, freq)
	if val0 != 1.0 && val0 != -1.0 {
		t.Errorf("PRBS generator output must be 1.0 or -1.0, got %v", val0)
	}
	
	// Ensure that for the same bit index period, the value is constant
	val1 := gen(0.5*(2*math.Pi/freq), freq)
	if val1 != val0 {
		t.Errorf("PRBS changed within the same bit period")
	}

	// Move to the next bit period and ensure it is still valid
	t2 := 1.5 * (2 * math.Pi / freq)
	val2 := gen(t2, freq)
	if val2 != 1.0 && val2 != -1.0 {
		t.Errorf("PRBS generator output must be 1.0 or -1.0, got %v", val2)
	}
}

func TestWaveforms_Square(t *testing.T) {
	SetRaiseFallTimePercent(0.01) // Ensure default state
	gen := NewWaveformGenerator(Square)
	
	valHigh := gen(math.Pi/2, 1000) // In the middle of the high phase
	if math.Abs(valHigh-(-1.0)) > 1e-9 {
		t.Errorf("Square(pi/2) = %v, expected -1.0", valHigh)
	}

	valLow := gen(3*math.Pi/2, 1000) // In the middle of the low phase
	if math.Abs(valLow-1.0) > 1e-9 {
		t.Errorf("Square(3pi/2) = %v, expected 1.0", valLow)
	}
}
