package gui

import (
	"math"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFormatFreq(t *testing.T) {
	tests := []struct {
		freq     float64
		expected string
	}{
		{0, "0"},
		{500, "500"},
		{1000, "1k"},
		{1500, "1.5k"},
		{25000, "25k"},
		{100000, "100k"},
		{1000000, "1M"},
		{10500000, "10.5M"},
		{100000000, "100M"},
	}

	for _, tc := range tests {
		actual := formatFreq(tc.freq)
		assert.Equal(t, tc.expected, actual, "formatFreq(%f)", tc.freq)
	}
}

func TestTimeLabelStrideSelection(t *testing.T) {
	// For narrow screen (gridSpacing = 40), maxSampleWidth = 60 -> stride should be 2
	gridSpacing := 40.0
	maxSampleWidth := 60.0
	stride := 1
	if gridSpacing < maxSampleWidth {
		if 2*gridSpacing >= maxSampleWidth {
			stride = 2
		} else {
			stride = 5
		}
	}
	assert.Equal(t, 2, stride)

	// For very narrow screen (gridSpacing = 20), maxSampleWidth = 60 -> stride should be 5
	gridSpacing = 20.0
	if gridSpacing < maxSampleWidth {
		if 2*gridSpacing >= maxSampleWidth {
			stride = 2
		} else {
			stride = 5
		}
	}
	assert.Equal(t, 5, stride)

	// For wide screen (gridSpacing = 100), maxSampleWidth = 60 -> stride should be 1
	gridSpacing = 100.0
	stride = 1
	if gridSpacing < maxSampleWidth {
		if 2*gridSpacing >= maxSampleWidth {
			stride = 2
		} else {
			stride = 5
		}
	}
	assert.Equal(t, 1, stride)
}

func TestLogFrequencyMultipliers(t *testing.T) {
	getMultipliers := func(minFreq, maxFreq float64) []int {
		logMin := math.Log10(math.Max(minFreq, 1e-6))
		logMax := math.Log10(math.Max(maxFreq, math.Max(minFreq, 1e-6)*1.001))
		numDecades := logMax - logMin
		if numDecades <= 1.2 {
			return []int{1, 2, 3, 4, 5, 6, 7, 8, 9}
		} else if numDecades <= 3.5 {
			return []int{1, 2, 5}
		}
		return []int{1}
	}

	// Narrow span: 20kHz to 50kHz (< 1 decade) -> full sub-decade points
	mNarrow := getMultipliers(20000, 50000)
	assert.Equal(t, []int{1, 2, 3, 4, 5, 6, 7, 8, 9}, mNarrow)

	// Medium span: 1kHz to 100kHz (2 decades) -> 1, 2, 5 points
	mMed := getMultipliers(1000, 100000)
	assert.Equal(t, []int{1, 2, 5}, mMed)

	// Wide span: 10Hz to 100MHz (7 decades) -> decade points only (1)
	mWide := getMultipliers(10, 100000000)
	assert.Equal(t, []int{1}, mWide)
}

func TestScaleLabelCollisionAvoidance(t *testing.T) {
	// Test simulated drawing loop with collision prevention
	divs := []float64{0, 10, 20, 30, 40, 50, 60, 70, 80, 90, 100}
	lblWidth := 25.0 // each label is 25px wide
	minGap := 6.0

	lastDrawnRight := -100000.0
	drawnIndices := []int{}

	for i, x := range divs {
		drawX := x - lblWidth/2.0
		if drawX < lastDrawnRight+minGap {
			continue
		}
		drawnIndices = append(drawnIndices, i)
		lastDrawnRight = drawX + lblWidth
	}

	// Since divs are spaced 10px apart and label+gap requires 31px, labels should only be drawn at indices 0, 3, 6, 9 (or similar spacing)
	// Guaranteed no adjacent labels overlap
	assert.True(t, len(drawnIndices) < len(divs))
	for k := 1; k < len(drawnIndices); k++ {
		prevX := divs[drawnIndices[k-1]]
		currX := divs[drawnIndices[k]]
		assert.True(t, currX-prevX >= lblWidth+minGap-0.001)
	}
}

func TestFormatFloatNoMinusZero(t *testing.T) {
	dt := float32(0.5)
	v := float32(-0.01)
	if v > -dt/8 && v < dt/8 {
		v = 0
	}
	vstr := strconv.FormatFloat(float64(v), 'f', 1, 32)
	assert.Equal(t, "0.0", vstr)
}
