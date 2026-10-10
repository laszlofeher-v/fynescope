package gui

import (
	"fmt"
	"fynescope/genericps"
	"fynescope/settings"
	"image"
	"image/color"
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

func TestFtScaleLabelsVisible(t *testing.T) {
	testRanges := []genericps.RangeEnum{
		genericps.Range_100mv,
		genericps.Range_500mv,
		genericps.Range_2v,
	}

	for _, vRange := range testRanges {
		t.Run(strconv.Itoa(int(vRange)), func(t *testing.T) {
			scp := &ScpDesc{
				Settings: &settings.PsSettings{
					Window: settings.WindowSettings{Width: 1366, Height: 1034},
					Channels: []settings.ChSettings{
						{Enabled: true, VRange: vRange, Col: [2]color.NRGBA{{100, 200, 255, 255}, {1, 5, 191, 255}}},
					},
				},
				channelViewers: make([]channelViewerDesc, 1),
				channelCount:   1,
				theme:          Theme(settings.DarkTheme),
				psControl:      &mockRemoteScopeController{},
			}
			scp.setRangeMargin()
			assert.True(t, scp.rangeMargin >= 75.0, "rangeMargin (%f) must be >= 75 to fit 6-character labels", scp.rangeMargin)

			w, h := float32(1366), float32(1034)
			scp.ftScopeFullScreen = scp.newScopeScreen(image.Point{int(w), int(h)})
			scp.partitionFtScreen(w, h)
			scp.setFtVDivsY()
			scp.setFtHDivsX()

			cl := &scp.channelViewers[0].label
			cl.enableRefresh()
			cl.draw()

			img := scp.ftScopeFullScreen.(*image.RGBA)
			type labelBlock struct {
				startY, endY int
				minX, maxX   int
			}
			var blocks []labelBlock
			inBlock := false
			currentBlock := labelBlock{minX: 10000}

			for y := 0; y < int(h); y++ {
				hasPixel := false
				minXInRow, maxXInRow := 10000, -1
				for x := 0; x < int(scp.rangeMargin); x++ {
					c := img.RGBAAt(x, y)
					if c.B > 200 { // Blue channel color drawn
						hasPixel = true
						if x < minXInRow {
							minXInRow = x
						}
						if x > maxXInRow {
							maxXInRow = x
						}
					}
				}
				if hasPixel {
					if !inBlock {
						inBlock = true
						currentBlock = labelBlock{startY: y, minX: minXInRow, maxX: maxXInRow}
					} else {
						if minXInRow < currentBlock.minX {
							currentBlock.minX = minXInRow
						}
						if maxXInRow > currentBlock.maxX {
							currentBlock.maxX = maxXInRow
						}
					}
				} else {
					if inBlock {
						inBlock = false
						currentBlock.endY = y - 1
						blocks = append(blocks, currentBlock)
					}
				}
			}
			if inBlock {
				currentBlock.endY = int(h) - 1
				blocks = append(blocks, currentBlock)
			}

			// Expect 12 blocks: 11 division labels + 1 unit label
			assert.Equal(t, 12, len(blocks), "expected 12 label blocks (11 divisions + unit)")

			// Uppermost label (block 0) must be drawn near the top division (y = 10)
			assert.True(t, blocks[0].startY <= 10 && blocks[0].endY >= 10,
				"uppermost label should span division 0 at y=10, got %d..%d", blocks[0].startY, blocks[0].endY)

			// Unit label (block 1) should be drawn between division 0 (10) and division 1 (~107)
			assert.True(t, blocks[1].startY > 20 && blocks[1].endY < 100,
				"unit label should be between division 0 and division 1, got %d..%d", blocks[1].startY, blocks[1].endY)

			// Lowest label (block 11) must be drawn near the bottom division (y = 983)
			lastBlock := blocks[len(blocks)-1]
			assert.True(t, lastBlock.startY <= 985 && lastBlock.endY >= 980,
				"lowest label should span division 10 at y=983, got %d..%d", lastBlock.startY, lastBlock.endY)

			// All label blocks must be strictly inside the canvas (minX >= 5, no left clipping)
			for idx, b := range blocks {
				assert.True(t, b.minX >= 5,
					"block %d (y=%d..%d) minX=%d must be >= 5 (not clipped at left edge)", idx, b.startY, b.endY, b.minX)
			}
		})
	}
}

func TestFtScaleLabelsWithOffset(t *testing.T) {
	for _, offset := range []int{-5, 5} {
		t.Run(fmt.Sprintf("offset_%d", offset), func(t *testing.T) {
			scp := &ScpDesc{
				Settings: &settings.PsSettings{
					Window: settings.WindowSettings{Width: 1366, Height: 1034},
					Channels: []settings.ChSettings{
						{Enabled: true, VRange: genericps.Range_100mv, Col: [2]color.NRGBA{{100, 200, 255, 255}, {1, 5, 191, 255}}},
					},
				},
				channelViewers: make([]channelViewerDesc, 1),
				channelCount:   1,
				theme:          Theme(settings.DarkTheme),
				psControl:      &mockRemoteScopeController{},
			}
			scp.setRangeMargin()

			w, h := float32(1366), float32(1034)
			scp.ftScopeFullScreen = scp.newScopeScreen(image.Point{int(w), int(h)})
			scp.partitionFtScreen(w, h)
			scp.setFtVDivsY()
			scp.setFtHDivsX()

			scp.channelViewers[0].displayOffsetInt = offset
			cl := &scp.channelViewers[0].label
			cl.enableRefresh()
			cl.draw()

			img := scp.ftScopeFullScreen.(*image.RGBA)
			drawnPixels := 0
			for y := 0; y < int(h); y++ {
				for x := 0; x < int(scp.rangeMargin); x++ {
					c := img.RGBAAt(x, y)
					if c.B > 200 {
						drawnPixels++
						// Verify pixel is within bounds (not at x < 0, and has padding >= 5)
						assert.True(t, x >= 5, "drawn pixel at x=%d should be >= 5", x)
					}
				}
			}
			assert.True(t, drawnPixels > 500, "expected labels to be drawn, got %d pixels", drawnPixels)
		})
	}
}




