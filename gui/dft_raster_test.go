package gui

import (
	"image"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNiceStep(t *testing.T) {
	tests := []struct {
		span     float64
		expected float64
	}{
		{0, 1},
		{-10, 1},
		{1, 1},
		{1.2, 1},
		{2.5, 2},
		{5.0, 5},
		{8.0, 10},
		{12.0, 10},
		{25.0, 20},
		{50.0, 50},
		{80.0, 100},
		{100, 100},
		{150, 200},
		{300, 200},
		{600, 500},
		{0.12, 0.1},
		{0.25, 0.2},
		{0.6, 0.5},
	}

	for _, tc := range tests {
		actual := niceStep(tc.span)
		assert.InDelta(t, tc.expected, actual, 0.0001, "niceStep(%f)", tc.span)
	}
}

func TestFrqLabelViewer_MousIn(t *testing.T) {
	rect := image.Rect(10, 20, 100, 50)
	frql := &frqLabelViewer{
		rasterPartition: rasterPartition{
			imgRect: rect,
		},
	}

	assert.True(t, frql.mousIn(15, 25), "Point inside should be true")
	assert.True(t, frql.mousIn(10, 20), "Point on top-left border should be true")
	
	// Max bounds are exclusive in image.Rect
	assert.False(t, frql.mousIn(100, 49), "Point on right border should be false")
	assert.False(t, frql.mousIn(99, 50), "Point on bottom border should be false")
	assert.False(t, frql.mousIn(5, 25), "Point outside left should be false")
	assert.False(t, frql.mousIn(15, 10), "Point outside top should be false")
}

func TestFrqLabelViewer_MouseEvents(t *testing.T) {
	rect := image.Rect(10, 20, 100, 50)
	frql := &frqLabelViewer{
		rasterPartition: rasterPartition{
			imgRect: rect,
		},
	}

	// Initial state
	assert.False(t, frql.selected)

	// Mouse down inside rect
	frql.mouseDown(0, 0, 15, 25)
	assert.True(t, frql.selected)

	// Mouse up
	frql.mouseUp(0, 0, 15, 25)
	assert.False(t, frql.selected)

	// Mouse down outside rect
	frql.mouseDown(0, 0, 5, 25)
	assert.False(t, frql.selected)
}
