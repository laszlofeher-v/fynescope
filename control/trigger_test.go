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
