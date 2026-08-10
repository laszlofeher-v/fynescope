package demo

import (
	"math"
	"testing"
)

func TestTriggerDetector_PositiveRunt(t *testing.T) {
	// Setup trigger detector for Positive Runt
	// Upper: 50, Lower: 10
	td := NewTriggerDetector(true, 50, 0, TriggerPositiveRunt, ChA)
	td.SetChannelProperties([]TriggerChannelProperties{
		{ThresholdUpper: 50, ThresholdUpperHysteresis: 0, ThresholdLower: 10, ThresholdLowerHysteresis: 0, Channel: ChA},
	})
	
	// Set threshold mode to Window so evaluateWindowTrigger is used
	td.channels[ChA].ThresholdMode = Window
	td.channels[ChA].ThresholdLower = 10
	td.channels[ChA].ThresholdLowerHysteresis = 0

	signalFunc := func(time float64, ch ChannelId) float64 {
		if ch != ChA {
			return 0
		}
		// Full swing (pulse to 60): Rises > 10, then Rises > 50, then Falls < 50, then Falls < 10
		// Should NOT trigger.
		if time >= 100 && time < 150 {
			if time < 110 {
				return (time - 100) * 6 // 0 to 60
			} else if time >= 140 {
				return (150 - time) * 6 // 60 to 0
			}
			return 60
		}

		// Positive Runt pulse: Rises > 10, peaks at 30, then falls < 10.
		// SHOULD trigger at the moment it falls back < 10 (approx t=250)
		if time >= 200 && time < 250 {
			if time < 210 {
				return (time - 200) * 3 // 0 to 30
			} else if time >= 240 {
				return (250 - time) * 3 // 30 to 0
			}
			return 30
		}

		return 0
	}

	dt := 1.0
	maxTime := 400.0
	found, triggerTime := td.FindTriggerPoint(signalFunc, 1000, maxTime, dt)
	if !found {
		t.Errorf("Expected positive runt trigger to be found")
	}

	// Trigger should fire at the INITIAL rising edge of the runt pulse
	// (when signal first crosses 10 upwards, at approx t=203.3).
	// The real scope also fires at the initial edge and just confirms runt via lookahead.
	if math.Abs(triggerTime-203.0) > 5.0 {
		t.Errorf("Expected positive runt trigger near t=203 (initial rising edge), got %v", triggerTime)
	}
}

func TestTriggerDetector_NegativeRunt(t *testing.T) {
	// Setup trigger detector for Negative Runt
	// Upper: -10, Lower: -50
	td := NewTriggerDetector(true, -10, 0, TriggerNegativeRunt, ChA)
	td.SetChannelProperties([]TriggerChannelProperties{
		{ThresholdUpper: -10, ThresholdUpperHysteresis: 0, ThresholdLower: -50, ThresholdLowerHysteresis: 0, Channel: ChA},
	})
	
	// Set threshold mode to Window so evaluateWindowTrigger is used
	td.channels[ChA].ThresholdMode = Window
	td.channels[ChA].ThresholdLower = -50
	td.channels[ChA].ThresholdLowerHysteresis = 0

	signalFunc := func(time float64, ch ChannelId) float64 {
		if ch != ChA {
			return 0
		}
		// Full swing (pulse to -60): Falls < -10, then Falls < -50, then Rises > -50, then Rises > -10
		// Should NOT trigger.
		if time >= 100 && time < 150 {
			if time < 110 {
				return -(time - 100) * 6 // 0 to -60
			} else if time >= 140 {
				return -(150 - time) * 6 // -60 to 0
			}
			return -60
		}

		// Negative Runt pulse: Falls < -10, dips to -30, then rises > -10.
		// SHOULD trigger at the moment it rises back > -10 (approx t=246)
		if time >= 200 && time < 250 {
			if time < 210 {
				return -(time - 200) * 3 // 0 to -30
			} else if time >= 240 {
				return -(250 - time) * 3 // -30 to 0
			}
			return -30
		}

		return 0
	}

	dt := 1.0
	maxTime := 400.0
	found, triggerTime := td.FindTriggerPoint(signalFunc, 1000, maxTime, dt)
	if !found {
		t.Errorf("Expected negative runt trigger to be found")
	}

	// Trigger should fire at the INITIAL falling edge of the runt pulse
	// (when signal first crosses -10 downwards, at approx t=203.3).
	// The real scope also fires at the initial edge and just confirms runt via lookahead.
	if math.Abs(triggerTime-203.0) > 5.0 {
		t.Errorf("Expected negative runt trigger near t=203 (initial falling edge), got %v", triggerTime)
	}
}
