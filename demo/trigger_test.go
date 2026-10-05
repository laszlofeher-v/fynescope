package demo

import (
	"math"
	"testing"
)

func TestTriggerDetector_FindTriggerPoint_IntervalLessThan(t *testing.T) {
	// Setup simple trigger detector
	td := NewTriggerDetector(true, 50, 10, TriggerRising, ChA)

	// Configure PWQ for "Less Than" 100 units
	conds := []PwqConditions{{ChannelA: CondTrue, ChannelB: CondDontCare, ChannelC: CondDontCare, ChannelD: CondDontCare}}
	td.SetPulseWidthQualifier(conds, TriggerRisingLower, 100, 0, PwTypeLessThan)

	// Create a mock signal function that creates a rising edge at t=100 and another rising edge at t=125.
	// Interval is 25, which is < 100, so it should trigger exactly at t=125.
	signalFunc := func(time float64, ch ChannelId) float64 {
		if ch != ChA {
			return 0
		}
		// Starts low (0), rises at t=100, falls at t=110, rises at t=125
		if (time >= 100 && time < 110) || (time >= 125 && time < 135) {
			return 100 // High
		}
		return 0 // Low
	}

	// Run FindTriggerPoint
	dt := 1.0
	maxTime := 300.0
	found, triggerTime := td.FindTriggerPoint(signalFunc, 1000, maxTime, dt)
	if !found {
		t.Errorf("Expected trigger to be found")
	}

	// We expect the trigger to happen precisely at t=125
	// Due to dt being 1.0 and checking logic edgeTriggerTime = t - dt, it will be 124
	if math.Abs(triggerTime-124.0) > 1.0 {
		t.Errorf("Expected trigger near t=124, got %v", triggerTime)
	}
}

func TestTriggerDetector_FindTriggerPoint_IntervalGreaterThan(t *testing.T) {
	// Setup simple trigger detector
	td := NewTriggerDetector(true, 50, 10, TriggerFalling, ChA)

	// Configure PWQ for "Greater Than" 100 units
	conds := []PwqConditions{{ChannelA: CondTrue, ChannelB: CondDontCare, ChannelC: CondDontCare, ChannelD: CondDontCare}}
	td.SetPulseWidthQualifier(conds, TriggerFallingLower, 100, 0, PwTypeGreaterThan)

	// Create a mock signal function that creates two falling edges.
	// Falling edge at t=100
	// Falling edge at t=250
	// Interval is 150, which is > 100, so it should trigger at t=250.
	signalFunc := func(time float64, ch ChannelId) float64 {
		if ch != ChA {
			return 0
		}
		// Starts high (100)
		// Falls to low (0) at t=100, rises at 110
		// Falls to low (0) at t=250, rises at 260
		if time < 100 {
			return 100
		}
		if time >= 100 && time < 110 {
			return 0
		}
		if time >= 110 && time < 250 {
			return 100
		}
		if time >= 250 && time < 260 {
			return 0
		}
		return 100
	}

	dt := 1.0
	maxTime := 400.0
	found, triggerTime := td.FindTriggerPoint(signalFunc, 1000, maxTime, dt)
	if !found {
		t.Errorf("Expected trigger to be found")
	}

	// Due to dt being 1.0 and checking logic edgeTriggerTime = t - dt, it will be 249
	if math.Abs(triggerTime-249.0) > 1.0 {
		t.Errorf("Expected trigger near t=249, got %v", triggerTime)
	}
}

func TestTriggerDetector_FindTriggerPoint_IntervalInRange(t *testing.T) {
	td := NewTriggerDetector(true, 50, 10, TriggerRising, ChA)

	conds := []PwqConditions{{ChannelA: CondTrue, ChannelB: CondDontCare, ChannelC: CondDontCare, ChannelD: CondDontCare}}
	// Interval must be between 80 and 120
	td.SetPulseWidthQualifier(conds, TriggerRisingLower, 80, 120, PwTypeInRange)

	signalFunc := func(time float64, ch ChannelId) float64 {
		if ch != ChA {
			return 0
		}
		// Edges:
		// t=100 (Rise 1)
		// t=110 (Rise 2) - Interval 10, NOT in range
		// t=150 (Rise 3) - Interval 40, NOT in range
		// t=250 (Rise 4) - Interval 100, IN range!
		if (time >= 100 && time < 105) ||
			(time >= 110 && time < 115) ||
			(time >= 150 && time < 155) ||
			(time >= 250 && time < 255) {
			return 100
		}
		return 0
	}

	dt := 1.0
	maxTime := 500.0
	found, triggerTime := td.FindTriggerPoint(signalFunc, 1000, maxTime, dt)
	if !found {
		t.Errorf("Expected trigger to be found")
	}

	// Due to dt being 1.0 and checking logic edgeTriggerTime = t - dt, it will be 249
	if math.Abs(triggerTime-249.0) > 1.0 {
		t.Errorf("Expected trigger near t=249, got %v", triggerTime)
	}
}

func TestTriggerDetector_FindTriggerPoint_TrueInterval(t *testing.T) {
	td := NewTriggerDetector(true, 50, 10, TriggerRising, ChA)

	conds := []PwqConditions{{ChannelA: CondTrue, ChannelB: CondDontCare, ChannelC: CondDontCare, ChannelD: CondDontCare}}
	// Interval must be Greater Than 100
	td.SetPulseWidthQualifier(conds, TriggerRising, 100, 0, PwTypeGreaterThan)

	signalFunc := func(time float64, ch ChannelId) float64 {
		if ch != ChA {
			return 0
		}
		// Intervals:
		// t=100 (Rise 1)
		// t=120 (Rise 2) - Interval 20. Not > 100.
		// t=150 (Rise 3) - Interval 30. Not > 100.
		// t=300 (Rise 4) - Interval 150. IS > 100!
		// It should trigger at t=300.
		if (time >= 100 && time < 110) ||
			(time >= 120 && time < 130) ||
			(time >= 150 && time < 160) ||
			(time >= 300 && time < 310) {
			return 100
		}
		return 0
	}

	dt := 1.0
	maxTime := 500.0
	found, triggerTime := td.FindTriggerPoint(signalFunc, 1000, maxTime, dt)
	if !found {
		t.Errorf("Expected trigger to be found")
	}

	// Due to dt being 1.0 and checking logic edgeTriggerTime = t - dt, it will be 299
	if math.Abs(triggerTime-299.0) > 1.0 {
		t.Errorf("Expected trigger near t=299, got %v", triggerTime)
	}
}

func TestTriggerDetector_FindTriggerPoint_RiseTime(t *testing.T) {
	td := NewTriggerDetector(true, 80, 5, TriggerRising, ChA)
	props := []TriggerChannelProperties{{
		ThresholdUpper:           80,
		ThresholdUpperHysteresis: 5,
		ThresholdLower:           20,
		ThresholdLowerHysteresis: 5,
		Channel:                  ChA,
		ThresholdMode:            Level,
	}}
	td.SetChannelProperties(props)
	td.SetChannelConditions([]TriggerConditions{{ChannelA: CondTrue, PulseWidthQualifier: CondTrue}})
	td.SetChannelDirections(TriggerRising, TriggerNone, TriggerNone, TriggerNone)

	conds := []PwqConditions{{ChannelA: CondTrue, ChannelB: CondDontCare, ChannelC: CondDontCare, ChannelD: CondDontCare}}
	// Rise time must be Less Than 50 units
	td.SetPulseWidthQualifier(conds, TriggerRisingLower, 50, 0, PwTypeLessThan)

	signalFunc := func(time float64, ch ChannelId) float64 {
		if ch != ChA {
			return 0
		}
		// Signal starts at 0.
		// At t=100 starts rising, reaches 20 at t=104, reaches 80 at t=120, tops at 100.
		// Transition time from 20 to 80 is 16 units (< 50).
		if time < 100 {
			return 0
		}
		if time >= 100 && time < 125 {
			return (time - 100) * 4.0 // at t=105: 20, at t=120: 80
		}
		return 100
	}

	dt := 1.0
	maxTime := 300.0
	found, triggerTime := td.FindTriggerPoint(signalFunc, 1000, maxTime, dt)
	if !found {
		t.Errorf("Expected Rise Time trigger to be found")
	}

	// Trigger should occur when crossing upper threshold around t=120
	if math.Abs(triggerTime-119.0) > 2.0 {
		t.Errorf("Expected trigger near t=119, got %v", triggerTime)
	}
}

func TestTriggerDetector_FindTriggerPoint_FallTime(t *testing.T) {
	td := NewTriggerDetector(true, 80, 5, TriggerFallingLower, ChA)
	props := []TriggerChannelProperties{{
		ThresholdUpper:           80,
		ThresholdUpperHysteresis: 5,
		ThresholdLower:           20,
		ThresholdLowerHysteresis: 5,
		Channel:                  ChA,
		ThresholdMode:            Level,
	}}
	td.SetChannelProperties(props)
	td.SetChannelConditions([]TriggerConditions{{ChannelA: CondTrue, PulseWidthQualifier: CondTrue}})
	td.SetChannelDirections(TriggerFallingLower, TriggerNone, TriggerNone, TriggerNone)

	conds := []PwqConditions{{ChannelA: CondTrue, ChannelB: CondDontCare, ChannelC: CondDontCare, ChannelD: CondDontCare}}
	// Fall time must be Less Than 50 units
	td.SetPulseWidthQualifier(conds, TriggerFalling, 50, 0, PwTypeLessThan)

	signalFunc := func(time float64, ch ChannelId) float64 {
		if ch != ChA {
			return 0
		}
		// Signal starts at 100.
		// At t=100 starts falling, reaches 80 at t=105, reaches 20 at t=120, bottoms at 0.
		// Transition time from 80 to 20 is 15 units (< 50).
		if time < 100 {
			return 100
		}
		if time >= 100 && time < 125 {
			return 100.0 - (time-100)*4.0 // at t=105: 80, at t=120: 20
		}
		return 0
	}

	dt := 1.0
	maxTime := 300.0
	found, triggerTime := td.FindTriggerPoint(signalFunc, 1000, maxTime, dt)
	if !found {
		t.Errorf("Expected Fall Time trigger to be found")
	}

	// Trigger should occur when crossing lower threshold around t=120
	if math.Abs(triggerTime-119.0) > 2.0 {
		t.Errorf("Expected trigger near t=119, got %v", triggerTime)
	}
}

func setupWindowPwqDetector(dir ThresholdDirection, lower, upper uint32, pwType PulseWidthType) *TriggerDetector {
	td := NewTriggerDetector(true, 80, 5, dir, ChA)
	props := []TriggerChannelProperties{{
		ThresholdUpper:           80,
		ThresholdUpperHysteresis: 5,
		ThresholdLower:           20,
		ThresholdLowerHysteresis: 5,
		Channel:                  ChA,
		ThresholdMode:            Window,
	}}
	td.SetChannelProperties(props)
	td.SetChannelConditions([]TriggerConditions{{ChannelA: CondTrue, PulseWidthQualifier: CondTrue}})
	td.SetChannelDirections(dir, TriggerNone, TriggerNone, TriggerNone)

	conds := []PwqConditions{{ChannelA: CondTrue, ChannelB: CondDontCare, ChannelC: CondDontCare, ChannelD: CondDontCare}}
	td.SetPulseWidthQualifier(conds, dir, lower, upper, pwType)
	return td
}

func makeWindowPulseSignal(pulseWidth float64) func(float64, ChannelId) float64 {
	return func(time float64, ch ChannelId) float64 {
		if ch != ChA {
			return 0
		}
		// Baseline 0. Pulse at 50 (inside window [20, 80]) from t=100 to t=100+pulseWidth.
		if time >= 100 && time <= 100+pulseWidth {
			return 50.0
		}
		return 0.0
	}
}

func TestTriggerDetector_WindowPulseWidth_LessThan(t *testing.T) {
	// Threshold = 50 units.
	td := setupWindowPwqDetector(TriggerOutside, 50, 0, PwTypeLessThan)

	// Pulse width 200 > 50: must NOT fire Less Than.
	found, _ := td.FindTriggerPoint(makeWindowPulseSignal(200), 1000, 1000, 1.0)
	if found {
		t.Errorf("Window PW Less Than mistakenly fired for pulse width 200 > 50")
	}

	// Pulse width 30 < 50: MUST fire Less Than.
	found, triggerTime := td.FindTriggerPoint(makeWindowPulseSignal(30), 1000, 1000, 1.0)
	if !found {
		t.Errorf("Window PW Less Than failed to fire for pulse width 30 < 50")
	}
	if math.Abs(triggerTime-130.0) > 2.0 {
		t.Errorf("Expected trigger near t=130, got %v", triggerTime)
	}
}

func TestTriggerDetector_WindowPulseWidth_GreaterThan(t *testing.T) {
	// Threshold = 50 units.
	td := setupWindowPwqDetector(TriggerOutside, 50, 0, PwTypeGreaterThan)

	// Pulse width 200 > 50: MUST fire Greater Than.
	found, triggerTime := td.FindTriggerPoint(makeWindowPulseSignal(200), 1000, 1000, 1.0)
	if !found {
		t.Errorf("Window PW Greater Than failed to fire for pulse width 200 > 50")
	}
	if triggerTime < 145 || triggerTime > 305 {
		t.Errorf("Expected trigger between t=150 and t=300, got %v", triggerTime)
	}

	// Pulse width 30 < 50: must NOT fire Greater Than.
	found, _ = td.FindTriggerPoint(makeWindowPulseSignal(30), 1000, 1000, 1.0)
	if found {
		t.Errorf("Window PW Greater Than mistakenly fired for pulse width 30 < 50")
	}
}

func TestTriggerDetector_WindowPulseWidth_InRange(t *testing.T) {
	// Range = [80, 120] units.
	td := setupWindowPwqDetector(TriggerOutside, 80, 120, PwTypeInRange)

	// Pulse width 100 in [80, 120]: MUST fire In Range.
	found, triggerTime := td.FindTriggerPoint(makeWindowPulseSignal(100), 1000, 1000, 1.0)
	if !found {
		t.Errorf("Window PW In Range failed to fire for pulse width 100 in [80, 120]")
	}
	if math.Abs(triggerTime-200.0) > 2.0 {
		t.Errorf("Expected trigger near t=200, got %v", triggerTime)
	}

	// Pulse width 200 outside [80, 120]: must NOT fire In Range.
	found, _ = td.FindTriggerPoint(makeWindowPulseSignal(200), 1000, 1000, 1.0)
	if found {
		t.Errorf("Window PW In Range mistakenly fired for pulse width 200 outside [80, 120]")
	}

	// Pulse width 30 outside [80, 120]: must NOT fire In Range.
	found, _ = td.FindTriggerPoint(makeWindowPulseSignal(30), 1000, 1000, 1.0)
	if found {
		t.Errorf("Window PW In Range mistakenly fired for pulse width 30 outside [80, 120]")
	}
}

func TestTriggerDetector_WindowPulseWidth_OutOfRange(t *testing.T) {
	// Range = [80, 120] units.
	td := setupWindowPwqDetector(TriggerOutside, 80, 120, PwTypeOutOfRange)

	// Pulse width 100 in [80, 120]: must NOT fire Out Of Range.
	found, _ := td.FindTriggerPoint(makeWindowPulseSignal(100), 1000, 1000, 1.0)
	if found {
		t.Errorf("Window PW Out Of Range mistakenly fired for pulse width 100 in range [80, 120]")
	}

	// Pulse width 200 outside [80, 120]: MUST fire Out Of Range.
	found, triggerTime := td.FindTriggerPoint(makeWindowPulseSignal(200), 1000, 1000, 1.0)
	if !found {
		t.Errorf("Window PW Out Of Range failed to fire for pulse width 200 outside [80, 120]")
	}
	if math.Abs(triggerTime-300.0) > 2.0 {
		t.Errorf("Expected trigger near t=300, got %v", triggerTime)
	}

	// Pulse width 30 outside [80, 120]: MUST fire Out Of Range.
	found, triggerTime = td.FindTriggerPoint(makeWindowPulseSignal(30), 1000, 1000, 1.0)
	if !found {
		t.Errorf("Window PW Out Of Range failed to fire for pulse width 30 outside [80, 120]")
	}
	if math.Abs(triggerTime-130.0) > 2.0 {
		t.Errorf("Expected trigger near t=130, got %v", triggerTime)
	}
}

func TestTriggerDetector_WindowPulseWidth_FullSwingIgnored(t *testing.T) {
	// Full-swing signal: 0 to 100 with finite slew rate (rise takes 10 units, 6 inside window [20, 80]).
	fullSwingSignal := func(time float64, ch ChannelId) float64 {
		if ch != ChA {
			return 0
		}
		cycle := math.Mod(time, 1000)
		if cycle < 10 {
			return cycle * 10.0
		} else if cycle < 500 {
			return 100.0
		} else if cycle < 510 {
			return 100.0 - (cycle-500.0)*10.0
		}
		return 0.0
	}

	// Less Than 50: must not fire on the 6-unit full-swing transition.
	tdLess := setupWindowPwqDetector(TriggerOutside, 50, 0, PwTypeLessThan)
	found, _ := tdLess.FindTriggerPoint(fullSwingSignal, 1000, 2000, 1.0)
	if found {
		t.Errorf("Window PW Less Than mistakenly fired on full-swing transition")
	}

	// Out Of Range [80, 120]: must not fire on the 6-unit full-swing transition.
	tdOut := setupWindowPwqDetector(TriggerOutside, 80, 120, PwTypeOutOfRange)
	found, _ = tdOut.FindTriggerPoint(fullSwingSignal, 1000, 2000, 1.0)
	if found {
		t.Errorf("Window PW Out Of Range mistakenly fired on full-swing transition")
	}
}

func TestTriggerDetector_SimpleTrigger_ThresholdNotMet(t *testing.T) {
	td := NewTriggerDetector(true, 500, 10, TriggerRising, ChA)
	// Flat 0 signal, never crosses 500
	signalFunc := func(time float64, ch ChannelId) float64 {
		return 0
	}
	found, _ := td.FindTriggerPoint(signalFunc, 1000, 1000, 1.0)
	if found {
		t.Errorf("Trigger should NOT be found when signal never crosses threshold")
	}
}

func TestTriggerDetector_SimpleTrigger_ThresholdMet(t *testing.T) {
	td := NewTriggerDetector(true, 500, 10, TriggerRising, ChA)
	// Signal rises from 0 to 1000 at t=100
	signalFunc := func(time float64, ch ChannelId) float64 {
		if time >= 100 {
			return 1000
		}
		return 0
	}
	found, triggerTime := td.FindTriggerPoint(signalFunc, 1000, 500, 1.0)
	if !found {
		t.Errorf("Trigger SHOULD be found when signal rises above threshold")
	}
	if math.Abs(triggerTime-100.0) > 2.0 {
		t.Errorf("Expected trigger near t=100, got %v", triggerTime)
	}
}

