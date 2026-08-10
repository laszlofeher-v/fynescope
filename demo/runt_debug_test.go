package demo

import (
	"fmt"
	"math"
	"testing"
)

func TestNewRuntStateLogic(t *testing.T) {
	td := NewTriggerDetector(true, 50, 0, TriggerPositiveRunt, ChA)
	td.SetChannelProperties([]TriggerChannelProperties{
		{ThresholdUpper: 50, ThresholdUpperHysteresis: 0, ThresholdLower: 10, ThresholdLowerHysteresis: 0, Channel: ChA},
	})
	
	cfg := td.channels[ChA]
	cfg.ThresholdMode = Window
	cfg.ThresholdLower = 10
	cfg.ThresholdLowerHysteresis = 0

	amplitude := 80.0
	freq := 0.01
	signalFunc := func(time float64, ch ChannelId) float64 {
		if ch != ChA {
			return 0
		}
		return amplitude * math.Sin(time * freq * 2 * math.Pi)
	}

	state := ChannelTriggerState{}
	dt := 1.0

	for tVal := 0.0; tVal < 1000.0; tVal += dt {
		level := signalFunc(tVal, ChA)
		
		_, fired, offset := td.evaluateWindowTrigger(cfg, &state, level, signalFunc, tVal, dt, ChA)
		
		if fired {
			fmt.Printf("FIRED at tVal=%v, level=%v, offset=%v\n", tVal, level, offset)
			fmt.Printf("  Actual trigger point was %v\n", tVal+offset)
			
			// If it fired, let's see what the signal was at the actual trigger point
			trigTime := tVal + offset
			fmt.Printf("  Signal at trigger point = %v\n", signalFunc(trigTime, ChA))
			t.Errorf("Should not fire on full swing")
			return
		}
	}
	fmt.Println("No trigger fired! Logic is correct.")
}
