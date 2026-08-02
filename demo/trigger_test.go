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

