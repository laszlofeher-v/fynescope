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
		LowerTriggerADC: 1500,
		ThresholdMode:   genericps.Window,
	}

	props := psControl.getValidTriggerProperties()
	if props[0].ThresholdUpper != 1500 || props[0].ThresholdLower != 500 {
		t.Errorf("Expected bounds to be swapped to 1500 and 500, got upper: %v, lower: %v", props[0].ThresholdUpper, props[0].ThresholdLower)
	}

	// Case 2: lower == upper (should expand window to minimum 512 counts)
	psControl.triggerSetting = TriggerDesc{
		TriggerADC:      500,
		LowerTriggerADC: 500,
		ThresholdMode:   genericps.Window,
	}

	props = psControl.getValidTriggerProperties()
	if props[0].ThresholdUpper != 1012 || props[0].ThresholdLower != 500 {
		t.Errorf("Expected upper bound to be expanded to 1012, got upper: %v, lower: %v", props[0].ThresholdUpper, props[0].ThresholdLower)
	}

	// Case 3: lower == upper near max int16 (should decrement lower by 512)
	psControl.triggerSetting = TriggerDesc{
		TriggerADC:      32700,
		LowerTriggerADC: 32700,
		ThresholdMode:   genericps.Window,
	}

	props = psControl.getValidTriggerProperties()
	if props[0].ThresholdUpper != 32700 || props[0].ThresholdLower != 32188 {
		t.Errorf("Expected lower bound to be decremented to 32188, got upper: %v, lower: %v", props[0].ThresholdUpper, props[0].ThresholdLower)
	}

	// Case 4: window < 512 counts (e.g. upper=256, lower=0) should expand to at least 512 counts
	psControl.triggerSetting = TriggerDesc{
		TriggerADC:      256,
		LowerTriggerADC: 0,
		ThresholdMode:   genericps.Window,
	}
	props = psControl.getValidTriggerProperties()
	if props[0].ThresholdUpper != 512 || props[0].ThresholdLower != 0 {
		t.Errorf("Expected window to expand to upper: 512, lower: 0, got upper: %v, lower: %v", props[0].ThresholdUpper, props[0].ThresholdLower)
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

	// Case 1: Window mode with upper hysteresis = 32767 (excessive hysteresis)
	psControl.triggerSetting = TriggerDesc{
		TriggerADC:         1000,
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
	if p.ThresholdUpper != 1000 || p.ThresholdLower != 0 {
		t.Errorf("Expected upper: 1000, lower: 0, got upper: %v, lower: %v", p.ThresholdUpper, p.ThresholdLower)
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

func TestPscDesc_buildTriggerConditions(t *testing.T) {
	channels := []genericps.ChannelId{genericps.ChA, genericps.ChB, genericps.ChC, genericps.ChD}
	for _, ch := range channels {
		psControl := &PscDesc{}
		psControl.triggerSetting.Source = ch
		psControl.triggerSetting.DigitalTriggerEnabled = false

		conds := psControl.buildTriggerConditions(genericps.CondTrue, genericps.CondFalse)
		if len(conds) != 1 {
			t.Fatalf("expected 1 condition, got %d", len(conds))
		}
		c := conds[0]
		if c.PulseWidthQualifier != genericps.CondFalse {
			t.Errorf("expected pwqCond false, got %v", c.PulseWidthQualifier)
		}
		if c.Digital != genericps.CondDontCare {
			t.Errorf("expected digital CondDontCare, got %v", c.Digital)
		}
		switch ch {
		case genericps.ChA:
			if c.ChannelA != genericps.CondTrue || c.ChannelB != genericps.CondDontCare {
				t.Errorf("ChA condition mismatch: %+v", c)
			}
		case genericps.ChB:
			if c.ChannelB != genericps.CondTrue || c.ChannelA != genericps.CondDontCare {
				t.Errorf("ChB condition mismatch: %+v", c)
			}
		case genericps.ChC:
			if c.ChannelC != genericps.CondTrue || c.ChannelA != genericps.CondDontCare {
				t.Errorf("ChC condition mismatch: %+v", c)
			}
		case genericps.ChD:
			if c.ChannelD != genericps.CondTrue || c.ChannelA != genericps.CondDontCare {
				t.Errorf("ChD condition mismatch: %+v", c)
			}
		}

		// With digital enabled
		psControl.triggerSetting.DigitalTriggerEnabled = true
		condsDig := psControl.buildTriggerConditions(genericps.CondTrue, genericps.CondTrue)
		if condsDig[0].Digital != genericps.CondTrue {
			t.Errorf("expected digital CondTrue when enabled, got %v", condsDig[0].Digital)
		}
	}
}

func TestPscDesc_buildPwqConditions(t *testing.T) {
	channels := []genericps.ChannelId{genericps.ChA, genericps.ChB, genericps.ChC, genericps.ChD}
	for _, ch := range channels {
		psControl := &PscDesc{}
		psControl.triggerSetting.Source = ch

		pwqConds := psControl.buildPwqConditions(genericps.CondTrue)
		if len(pwqConds) != 1 {
			t.Fatalf("expected 1 pwq condition, got %d", len(pwqConds))
		}
		p := pwqConds[0]
		switch ch {
		case genericps.ChA:
			if p.ChannelA != genericps.CondTrue || p.ChannelB != genericps.CondDontCare {
				t.Errorf("ChA pwq mismatch: %+v", p)
			}
		case genericps.ChB:
			if p.ChannelB != genericps.CondTrue || p.ChannelA != genericps.CondDontCare {
				t.Errorf("ChB pwq mismatch: %+v", p)
			}
		case genericps.ChC:
			if p.ChannelC != genericps.CondTrue || p.ChannelA != genericps.CondDontCare {
				t.Errorf("ChC pwq mismatch: %+v", p)
			}
		case genericps.ChD:
			if p.ChannelD != genericps.CondTrue || p.ChannelA != genericps.CondDontCare {
				t.Errorf("ChD pwq mismatch: %+v", p)
			}
		}
	}
}

func TestPscDesc_computePwqSamples(t *testing.T) {
	ps := &PscDesc{}
	ps.scopeScreenWidth = 10000
	ps.SamplingTimeInterval = 10 // maxPwqSamples = 1000

	ps.triggerSetting.IntervalTimeLower = 500  // 50 samples
	ps.triggerSetting.IntervalTimeUpper = 2000 // 200 samples
	ps.triggerSetting.IntervalType = genericps.PwTypeNone

	low, up := ps.computePwqSamples()
	if low != 50 || up != 200 {
		t.Errorf("expected low=50, up=200, got low=%d, up=%d", low, up)
	}

	// Test InRange swap/adjust when lower >= upper
	ps.triggerSetting.IntervalTimeLower = 2000 // 200 samples
	ps.triggerSetting.IntervalTimeUpper = 500  // 50 samples
	ps.triggerSetting.IntervalType = genericps.PwTypeInRange

	low, up = ps.computePwqSamples()
	if low >= up {
		t.Errorf("expected low < up for InRange, got low=%d, up=%d", low, up)
	}

	// Test clamping to maxPwqSamples
	ps.triggerSetting.IntervalType = genericps.PwTypeNone
	ps.triggerSetting.IntervalTimeLower = 500000
	ps.triggerSetting.IntervalTimeUpper = 500000
	low, up = ps.computePwqSamples()
	if low != 1000 || up != 1000 {
		t.Errorf("expected samples clamped to 1000, got low=%d, up=%d", low, up)
	}
}

func TestPscDesc_autoTriggerMilliseconds32(t *testing.T) {
	ps := &PscDesc{}
	ps.triggerSetting.Mode = Auto
	if ps.autoTriggerMilliseconds32() != autoTriggerMs {
		t.Errorf("expected %d, got %d", autoTriggerMs, ps.autoTriggerMilliseconds32())
	}
	ps.triggerSetting.Mode = Repeat
	if ps.autoTriggerMilliseconds32() != 0 {
		t.Errorf("expected 0 for Repeat mode, got %d", ps.autoTriggerMilliseconds32())
	}
}

func TestWindowChannelDir(t *testing.T) {
	tests := []struct {
		in   genericps.ThresholdDirection
		want genericps.ThresholdDirection
	}{
		{genericps.TriggerEnter, genericps.TriggerOutside},
		{genericps.TriggerEnterOrExit, genericps.TriggerOutside},
		{genericps.TriggerExit, genericps.TriggerInside},
		{genericps.TriggerRising, genericps.TriggerRising},
		{genericps.TriggerFalling, genericps.TriggerFalling},
	}
	for _, tt := range tests {
		if got := windowChannelDir(tt.in); got != tt.want {
			t.Errorf("windowChannelDir(%v) = %v, want %v", tt.in, got, tt.want)
		}
	}
}

func TestAdjustPwqSamples(t *testing.T) {
	// LessThan
	it, low, up := adjustPwqSamples(genericps.PwTypeLessThan, 10, 50)
	if it != genericps.PwTypeLessThan || low != 50 || up != 0 {
		t.Errorf("LessThan: got it=%v, low=%d, up=%d", it, low, up)
	}

	// GreaterThan
	it, low, up = adjustPwqSamples(genericps.PwTypeGreaterThan, 10, 50)
	if it != genericps.PwTypeGreaterThan || low != 10 || up != 0 {
		t.Errorf("GreaterThan: got it=%v, low=%d, up=%d", it, low, up)
	}

	// InRange
	it, low, up = adjustPwqSamples(genericps.PwTypeInRange, 10, 50)
	if it != genericps.PwTypeInRange || low != 10 || up != 50 {
		t.Errorf("InRange: got it=%v, low=%d, up=%d", it, low, up)
	}
}

func TestPscDesc_adjustPwqSamplesRiseFall(t *testing.T) {
	ps := &PscDesc{}

	// LessThan
	it, low, up := ps.adjustPwqSamplesRiseFall(genericps.PwTypeLessThan, 10, 50)
	if it != genericps.PwTypeLessThan || low != 50 || up != 0 {
		t.Errorf("LessThan: got it=%v, low=%d, up=%d", it, low, up)
	}

	// GreaterThan converts to InRange with upper = maxPwqSamples - 1 (32766)
	it, low, up = ps.adjustPwqSamplesRiseFall(genericps.PwTypeGreaterThan, 20, 100)
	if it != genericps.PwTypeInRange || low != 20 || up != 32766 {
		t.Errorf("GreaterThan: got it=%v, low=%d, up=%d", it, low, up)
	}
}
