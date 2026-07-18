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

	// Create a mock signal function that creates a rising edge at t=100 and falling at t=125
	// Pulse Width should be 25, which is < 100, so it should trigger exactly at t=125 (end of pulse).
	signalFunc := func(time float64, ch ChannelId) float64 {
		if ch != ChA {
			return 0
		}
		// Signal logic to simulate edges:
		// Starts low (0), goes high at t=100, goes low at t=125
		if time >= 100 && time < 125 {
			return 100 // High
		}
		return 0 // Low
	}

	// Run FindTriggerPoint
	dt := 1.0
	maxTime := 300.0
	triggerTime := td.FindTriggerPoint(signalFunc, 1000, maxTime, dt)

	// We expect the trigger to happen precisely at t=125 (the falling edge marking the end of the pulse)
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
	td.SetPulseWidthQualifier(conds, TriggerFalling, 100, 0, PwTypeGreaterThan)

	// Create a mock signal function that creates a negative pulse
	// Falling edge at t=100
	// Rising edge at t=250
	// Pulse Width is 150, which is > 100, so it should trigger at t=250.
	signalFunc := func(time float64, ch ChannelId) float64 {
		if ch != ChA {
			return 0
		}
		// Starts high (100)
		// Falls to low (0) at t=100
		// Rises to high (100) at t=250
		if time < 100 {
			return 100
		}
		if time >= 100 && time < 250 {
			return 0
		}
		return 100
	}

	dt := 1.0
	maxTime := 400.0
	triggerTime := td.FindTriggerPoint(signalFunc, 1000, maxTime, dt)

	// Due to dt being 1.0 and checking logic edgeTriggerTime = t - dt, it will be 249
	if math.Abs(triggerTime-249.0) > 1.0 {
		t.Errorf("Expected trigger near t=249, got %v", triggerTime)
	}
}

func TestTriggerDetector_FindTriggerPoint_IntervalInRange(t *testing.T) {
	td := NewTriggerDetector(true, 50, 10, TriggerRising, ChA)

	conds := []PwqConditions{{ChannelA: CondTrue, ChannelB: CondDontCare, ChannelC: CondDontCare, ChannelD: CondDontCare}}
	// Pulse Width must be between 80 and 120
	td.SetPulseWidthQualifier(conds, TriggerRising, 80, 120, PwTypeInRange)

	signalFunc := func(time float64, ch ChannelId) float64 {
		if ch != ChA {
			return 0
		}
		// Edges:
		// t=100 (Rise 1) to t=110 (Fall 1) - Width 10, NOT in range
		// t=150 (Rise 2) to t=160 (Fall 2) - Width 10, NOT in range
		// t=250 (Rise 3) to t=350 (Fall 3) - Width 100, IN range!
		if time >= 100 && time < 110 { return 100 }
		if time >= 150 && time < 160 { return 100 }
		if time >= 250 && time < 350 { return 100 }
		return 0
	}

	dt := 1.0
	maxTime := 500.0
	triggerTime := td.FindTriggerPoint(signalFunc, 1000, maxTime, dt)

	// Due to dt being 1.0 and checking logic edgeTriggerTime = t - dt, it will be 349
	if math.Abs(triggerTime-349.0) > 1.0 {
		t.Errorf("Expected trigger near t=349, got %v", triggerTime)
	}
}
