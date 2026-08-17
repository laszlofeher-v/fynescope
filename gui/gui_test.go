package gui

import (
	"fynescope/genericps"
	"testing"
)

func TestFlags(t *testing.T) {
	// Test nil flag
	var nilFlag chan struct{}
	setFlag(nilFlag) // Should not panic
	if getFlag(nilFlag) {
		t.Error("Expected getFlag(nil) to return false")
	}

	// Test normal flag
	flag := createFlag()
	if getFlag(flag) {
		t.Error("Expected initial flag to be false")
	}

	setFlag(flag)
	if !getFlag(flag) {
		t.Error("Expected flag to be true after setFlag")
	}

	if getFlag(flag) {
		t.Error("Expected flag to be cleared after getFlag")
	}

	// Test setting multiple times (should not block/panic)
	setFlag(flag)
	setFlag(flag)
	if !getFlag(flag) {
		t.Error("Expected flag to be true")
	}
}

func TestAdcConversions(t *testing.T) {
	// Mock InputRanges since it's normally populated by device drivers
	genericps.InputRanges = make([]int32, 20)
	genericps.InputRanges[genericps.Range_2v] = 2000

	scp := &ScpDesc{
		MaxValue: 32767,
		MinValue: -32767,
	}

	// Range 2V means +/- 2000mV
	// max raw 32767 should be 2000mV
	// However, the formula is: float64(raw) * float64(InputRanges[chRange]) / float64(scp.MaxValue)
	// InputRanges[Range_2v] is 2000.
	
	// Test adcToMv
	mv := scp.adcToMv(32767, genericps.Range_2v)
	if mv != 2000 {
		t.Errorf("Expected 2000, got %v", mv)
	}

	mv = scp.adcToMv(-32767, genericps.Range_2v)
	if mv != -2000 {
		t.Errorf("Expected -2000, got %v", mv)
	}

	mv = scp.adcToMv(0, genericps.Range_2v)
	if mv != 0 {
		t.Errorf("Expected 0, got %v", mv)
	}

	// Test mvToAdc
	adc := scp.mvToAdc(2000, genericps.Range_2v)
	if adc != 32767 {
		t.Errorf("Expected 32767, got %v", adc)
	}

	adc = scp.mvToAdc(-2000, genericps.Range_2v)
	if adc != -32767 {
		t.Errorf("Expected -32767, got %v", adc)
	}

	// Test mvToAdc clamping
	adc = scp.mvToAdc(3000, genericps.Range_2v) // exceeding max
	if adc != 32767 {
		t.Errorf("Expected 32767 (clamped), got %v", adc)
	}

	adc = scp.mvToAdc(-3000, genericps.Range_2v) // exceeding min
	if adc != -32767 {
		t.Errorf("Expected -32767 (clamped), got %v", adc)
	}

	// Test mvToUAdc clamping (only positive clamping)
	adc = scp.mvToUAdc(3000, genericps.Range_2v)
	if adc != 32767 {
		t.Errorf("Expected 32767, got %v", adc)
	}
	
	// mvToUAdc doesn't clamp negative, so it will return a highly negative value
	adc = scp.mvToUAdc(-3000, genericps.Range_2v)
	expectedNeg := int32(-3000 * 32767 / 2000)
	if adc != expectedNeg {
		t.Errorf("Expected %v, got %v", expectedNeg, adc)
	}
}
