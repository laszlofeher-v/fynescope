package demo

import (
	"fmt"
	"testing"
)

func TestPhaseNoiseAmplitude(t *testing.T) {
	SetPhaseNoiseDegree(int(ChA), 5.0) // 5 degrees
	
	desc := &channels[ChA]
	desc.genSource = ChA
	desc.genWaveFunction = sineWave
	desc.genOn = true
	desc.genPkToPk = 2000000 // 1000mV * 2
	desc.vrange = 8 // 1V/div maybe

	// Get full scale amplitude
	SetPhaseNoiseDegree(int(ChA), 0.0)
	desc.phase = 90
	fullScale := calculateSampleLevelAtTime(0, ChA)
	fmt.Printf("Full scale amplitude: %f\n", fullScale)

	SetPhaseNoiseDegree(int(ChA), 5.0)
	desc.phase = 0
	maxVal := -1000000.0
	for i := 0; i < 10000; i++ {
		val := calculateSampleLevelAtTime(0, ChA)
		if val > maxVal { maxVal = val }
	}
	
	fmt.Printf("Max val at 0 with 5 deg noise: %f\n", maxVal)
	fmt.Printf("Ratio: %f\n", maxVal / fullScale)
}
