package main

import (
	"fmt"
	"math"
)

var prbsBlockOffset uint64 = 12345

func NewPrbsGenerator() func(t float64, freq float64) float64 {
	return func(t float64, freq float64) float64 {
		if freq <= 0 {
			return 0
		}
		bitIndex := int64(t / (2 * math.Pi))
		
		x := uint64(bitIndex) + prbsBlockOffset
		x += 0x9e3779b97f4a7c15 // Weyl constant
		x = (x ^ (x >> 30)) * 0xbf58476d1ce4e5b9
		x = (x ^ (x >> 27)) * 0x94d049bb133111eb
		x = x ^ (x >> 31)

		if x&1 == 1 {
			return 1.0
		}
		return -1.0
	}
}

func main() {
	gen := NewPrbsGenerator()
	freq := 4000.0 // 4 kHz
	
	dt := 2.0e-6 // 2 us

	lastVal := -2.0
	pulseStart := -1.0

	for i := 0; i < 500000; i++ {
		t := float64(i) * dt
		phase := (t * freq) * math.Pi * 2
		val := gen(phase, freq)
		
		if lastVal != -2.0 && val != lastVal {
			if lastVal == -1.0 && val == 1.0 {
				pulseStart = t
			} else if lastVal == 1.0 && val == -1.0 && pulseStart != -1.0 {
				duration := t - pulseStart
				if duration < 240e-6 {
					fmt.Printf("SHORT PULSE DETECTED! t=%v, duration=%v\n", t, duration)
				}
			}
		}
		lastVal = val
	}
	fmt.Println("Test finished")
}
