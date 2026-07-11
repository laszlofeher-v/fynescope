package sim

import (
	"math"
	"testing"
)

func TestTriggerDetector_FindTriggerPoint_IntervalLessThan(t *testing.T) {
	// Setup simple trigger detector
	td := NewTriggerDetector(true, 50, 10, TriggerRising, ChA)

	// Configure PWQ for "Less Than" 100 units
	conds := []PwqConditions{{ChannelA: CondTrue, ChannelB: CondDontCare, ChannelC: CondDontCare, ChannelD: CondDontCare}}
	td.SetPulseWidthQualifier(conds, TriggerRising, 100, 0, PwTypeLessThan)

	// Create a mock signal function that creates a rising edge at t=100 and t=150
	// Interval should be 50, which is < 100, so it should trigger exactly at t=150.
	signalFunc := func(time float64, ch ChannelId) float64 {
		if ch != ChA {
			return 0
		}
		// Signal logic to simulate edges:
		// Starts low (0), goes high at t=100, goes low at t=125, goes high at t=150
		if time >= 100 && time < 125 {
			return 100 // High
		}
		if time >= 150 && time < 175 {
			return 100 // High
		}
		return 0 // Low
	}

	// Run FindTriggerPoint
	dt := 1.0
	maxTime := 300.0
	triggerTime := td.FindTriggerPoint(signalFunc, 1000, maxTime, dt)

	// We expect the trigger to happen precisely at t=150 (the second rising edge)
	// Due to dt being 1.0 and checking logic edgeTriggerTime = t - dt, it will be 149
	if math.Abs(triggerTime-149.0) > 1.0 {
		t.Errorf("Expected trigger near t=149, got %v", triggerTime)
	}
}

func TestTriggerDetector_FindTriggerPoint_IntervalGreaterThan(t *testing.T) {
	// Setup simple trigger detector
	td := NewTriggerDetector(true, 50, 10, TriggerFalling, ChA)

	// Configure PWQ for "Greater Than" 100 units
	conds := []PwqConditions{{ChannelA: CondTrue, ChannelB: CondDontCare, ChannelC: CondDontCare, ChannelD: CondDontCare}}
	td.SetPulseWidthQualifier(conds, TriggerFalling, 100, 0, PwTypeGreaterThan)

	// Create a mock signal function that creates falling edges
	// First falling edge at t=100
	// Second falling edge at t=250
	// Interval is 150, which is > 100, so it should trigger at t=250.
	signalFunc := func(time float64, ch ChannelId) float64 {
		if ch != ChA {
			return 0
		}
		// Starts high (100)
		// Falls to low (0) at t=100
		// Rises to high (100) at t=200
		// Falls to low (0) at t=250
		if time < 100 {
			return 100
		}
		if time >= 100 && time < 200 {
			return 0
		}
		if time >= 200 && time < 250 {
			return 100
		}
		return 0
	}

	dt := 1.0
	maxTime := 400.0
	triggerTime := td.FindTriggerPoint(signalFunc, 1000, maxTime, dt)

	if math.Abs(triggerTime-249.0) > 1.0 {
		t.Errorf("Expected trigger near t=249, got %v", triggerTime)
	}
}

func TestTriggerDetector_FindTriggerPoint_IntervalInRange(t *testing.T) {
	td := NewTriggerDetector(true, 50, 10, TriggerRising, ChA)

	conds := []PwqConditions{{ChannelA: CondTrue, ChannelB: CondDontCare, ChannelC: CondDontCare, ChannelD: CondDontCare}}
	// Interval must be between 80 and 120
	td.SetPulseWidthQualifier(conds, TriggerRising, 80, 120, PwTypeInRange)

	signalFunc := func(time float64, ch ChannelId) float64 {
		if ch != ChA {
			return 0
		}
		// Edges:
		// t=100 (Rise 1)
		// t=150 (Rise 2 - Interval 50, NOT in range)
		// t=250 (Rise 3 - Interval 100, IN range)
		if time >= 100 && time < 110 { return 100 }
		if time >= 150 && time < 160 { return 100 }
		if time >= 250 && time < 260 { return 100 }
		return 0
	}

	dt := 1.0
	maxTime := 400.0
	triggerTime := td.FindTriggerPoint(signalFunc, 1000, maxTime, dt)

	if math.Abs(triggerTime-249.0) > 1.0 {
		t.Errorf("Expected trigger near t=249, got %v", triggerTime)
	}
}
