package control

import (
	"fynescope/genericps"
	"testing"
)

func TestPscDesc_getValidTriggerProperties(t *testing.T) {
	psControl := &PscDesc{}

	// Set up a basic trigger setting
	psControl.triggerSetting = TriggerDesc{
		TriggerADC:         1000,
		LowerTriggerADC:    -1000,
		HysteresisADC:      100,
		LowerHysteresisADC: 50,
		Source:             genericps.ChA,
		ThresholdMode:      genericps.Level,
	}

	props := psControl.getValidTriggerProperties()

	if len(props) != 1 {
		t.Fatalf("Expected 1 property, got %d", len(props))
	}

	p := props[0]
	if p.ThresholdUpper != 1000 {
		t.Errorf("Expected ThresholdUpper 1000, got %v", p.ThresholdUpper)
	}
	if p.ThresholdUpperHysteresis != 100 {
		t.Errorf("Expected ThresholdUpperHysteresis 100, got %v", p.ThresholdUpperHysteresis)
	}
	if p.ThresholdLower != -1000 {
		t.Errorf("Expected ThresholdLower -1000, got %v", p.ThresholdLower)
	}
	if p.ThresholdLowerHysteresis != 50 {
		t.Errorf("Expected ThresholdLowerHysteresis 50, got %v", p.ThresholdLowerHysteresis)
	}
	if p.Channel != genericps.ChA {
		t.Errorf("Expected Channel ChA, got %v", p.Channel)
	}
	if p.ThresholdMode != genericps.Level {
		t.Errorf("Expected ThresholdMode Level, got %v", p.ThresholdMode)
	}
}

func TestPscDesc_getValidTriggerProperties_WindowCorrection(t *testing.T) {
	psControl := &PscDesc{}

	// Case 1: lower > upper (should swap)
	psControl.triggerSetting = TriggerDesc{
		TriggerADC:      500,
		LowerTriggerADC: 1000,
		ThresholdMode:   genericps.Window,
	}

	props := psControl.getValidTriggerProperties()
	if props[0].ThresholdUpper != 1000 || props[0].ThresholdLower != 500 {
		t.Errorf("Expected bounds to be swapped to 1000 and 500, got upper: %v, lower: %v", props[0].ThresholdUpper, props[0].ThresholdLower)
	}

	// Case 2: lower == upper (should expand window to minimum 256 counts)
	psControl.triggerSetting = TriggerDesc{
		TriggerADC:      500,
		LowerTriggerADC: 500,
		ThresholdMode:   genericps.Window,
	}

	props = psControl.getValidTriggerProperties()
	if props[0].ThresholdUpper != 756 || props[0].ThresholdLower != 500 {
		t.Errorf("Expected upper bound to be expanded to 756, got upper: %v, lower: %v", props[0].ThresholdUpper, props[0].ThresholdLower)
	}

	// Case 3: lower == upper near max int16 (should decrement lower)
	psControl.triggerSetting = TriggerDesc{
		TriggerADC:      32700,
		LowerTriggerADC: 32700,
		ThresholdMode:   genericps.Window,
	}

	props = psControl.getValidTriggerProperties()
	if props[0].ThresholdUpper != 32700 || props[0].ThresholdLower != 32444 {
		t.Errorf("Expected lower bound to be decremented to 32444, got upper: %v, lower: %v", props[0].ThresholdUpper, props[0].ThresholdLower)
	}
}

func TestPscDesc_getValidTriggerProperties_RiseFall(t *testing.T) {
	psControl := &PscDesc{}

	// RiseFall with initial Level mode should automatically become Window mode in properties
	psControl.triggerSetting = TriggerDesc{
		Type:            RiseFall,
		TriggerADC:      2000,
		LowerTriggerADC: 500,
		ThresholdMode:   genericps.Level,
	}

	props := psControl.getValidTriggerProperties()
	if len(props) != 1 {
		t.Fatalf("Expected 1 property, got %d", len(props))
	}
	if props[0].ThresholdMode != genericps.Window {
		t.Errorf("Expected ThresholdMode Window for RiseFall, got %v", props[0].ThresholdMode)
	}
	if props[0].ThresholdUpper != 2000 || props[0].ThresholdLower != 500 {
		t.Errorf("Expected upper: 2000, lower: 500, got upper: %v, lower: %v", props[0].ThresholdUpper, props[0].ThresholdLower)
	}
}

func TestPscDesc_getValidTriggerProperties_HysteresisClamping(t *testing.T) {
	psControl := &PscDesc{}

	// Case 1: Window mode with upper hysteresis = 32767 (the exact error report)
	psControl.triggerSetting = TriggerDesc{
		TriggerADC:         327,
		LowerTriggerADC:    0,
		HysteresisADC:      32767,
		LowerHysteresisADC: 0,
		Source:             genericps.ChA,
		ThresholdMode:      genericps.Window,
	}

	props := psControl.getValidTriggerProperties()
	p := props[0]
	if p.ThresholdMode != genericps.Window {
		t.Errorf("Expected ThresholdMode Window, got %v", p.ThresholdMode)
	}
	if p.ThresholdUpper != 327 || p.ThresholdLower != 0 {
		t.Errorf("Expected upper: 327, lower: 0, got upper: %v, lower: %v", p.ThresholdUpper, p.ThresholdLower)
	}
	// Verify that upper - upperHyst >= lower + lowerHyst (no crossover)
	if int32(p.ThresholdUpper)-int32(p.ThresholdUpperHysteresis) < int32(p.ThresholdLower)+int32(p.ThresholdLowerHysteresis) {
		t.Errorf("Upper hysteresis overlaps lower threshold: upper=%d, uh=%d, lower=%d, lh=%d",
			p.ThresholdUpper, p.ThresholdUpperHysteresis, p.ThresholdLower, p.ThresholdLowerHysteresis)
	}
	if p.ThresholdUpperHysteresis == 32767 {
		t.Errorf("Expected ThresholdUpperHysteresis to be clamped from 32767, got %v", p.ThresholdUpperHysteresis)
	}

	// Case 2: Level mode with excessive hysteresis
	psControl.triggerSetting = TriggerDesc{
		Type:               Advanced,
		TriggerADC:         100,
		LowerTriggerADC:    100,
		HysteresisADC:      50000,
		LowerHysteresisADC: 0,
		Source:             genericps.ChA,
		ThresholdMode:      genericps.Level,
	}
	props = psControl.getValidTriggerProperties()
	p = props[0]
	if p.ThresholdUpperHysteresis > 32767 {
		t.Errorf("Expected upper hysteresis in level mode to be <= 32767, got %v", p.ThresholdUpperHysteresis)
	}
}

