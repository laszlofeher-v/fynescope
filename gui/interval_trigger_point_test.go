package gui

import (
	"fynescope/genericps"
	"fynescope/settings"
	"image"
	"testing"

	"github.com/stretchr/testify/assert"
)

func setupIntervalTriggerScp() *ScpDesc {
	genericps.InputRanges = make([]int32, 40)
	genericps.InputRanges[genericps.Range_5v] = 5000
	genericps.InputRanges[genericps.Range_1v] = 1000

	return &ScpDesc{
		channelCount: 2,
		Settings: &settings.PsSettings{
			Channels: []settings.ChSettings{
				{VRange: genericps.Range_5v, Inverted: false},
				{VRange: genericps.Range_1v, Inverted: true},
			},
			Time: settings.TimeSettings{
				TimeDiv: "10us",
			},
		},
	}
}

func TestIntervalTriggerPointViewer_MouseAtIntervalPoint(t *testing.T) {
	scp := setupIntervalTriggerScp()

	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	viewer := newIntervalTriggerPointViewer(img, scp, false)

	// Inject dummy rectangles
	viewer.lowerHImgRects = []image.Rectangle{
		image.Rect(10, 10, 20, 20),
	}
	viewer.upperHImgRects = []image.Rectangle{
		image.Rect(30, 30, 40, 40),
	}

	assert.True(t, viewer.mouseAtIntervalPoint(15, 15), "Should be true for point inside lower rect")
	assert.True(t, viewer.mouseAtIntervalPoint(35, 35), "Should be true for point inside upper rect")
	assert.False(t, viewer.mouseAtIntervalPoint(25, 25), "Should be false for point outside rects")
	assert.False(t, viewer.mouseAtIntervalPoint(0, 0), "Should be false for point outside rects")
}

func TestIntervalTriggerPointViewer_Cursor(t *testing.T) {
	scp := setupIntervalTriggerScp()

	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	viewer := newIntervalTriggerPointViewer(img, scp, false)

	viewer.lowerHImgRects = []image.Rectangle{
		image.Rect(10, 10, 20, 20),
	}

	// Test outside
	cursor, ok := viewer.cursor(-100, -100)
	assert.False(t, ok)
	assert.NotNil(t, cursor)

	// Test inside lower rect
	cursor, ok = viewer.cursor(15, 15)
	assert.True(t, ok)
	assert.NotNil(t, cursor)
}
