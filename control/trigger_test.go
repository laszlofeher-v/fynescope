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

func TestPscDesc_getValidTriggerProperties_WindowPanic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic when lower >= upper in Window mode")
		}
	}()

	psControl := &PscDesc{}
	psControl.triggerSetting = TriggerDesc{
		TriggerADC:         500,
		LowerTriggerADC:    1000, // Invalid: lower > upper
		ThresholdMode:      genericps.Window,
	}

	// This should panic
	_ = psControl.getValidTriggerProperties()
}
