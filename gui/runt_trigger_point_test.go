package gui

import (
	"fynescope/genericps"
	"fynescope/settings"
	"image"
	"testing"

	"github.com/stretchr/testify/assert"
)

func setupRuntTriggerScp() *ScpDesc {
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

func TestRuntTriggerPointViewer_HitTesting(t *testing.T) {
	scp := setupRuntTriggerScp()

	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	viewer := newRuntTriggerPointViewer(img, scp, false)

	// Set up hitboxes
	viewer.imgRect = image.Rect(-10, -10, -5, -5) // Out of the way to avoid advTriggerPointViewer
	viewer.uhImgRect = image.Rect(10, 10, 20, 20)
	viewer.lImgRect = image.Rect(30, 30, 40, 40)
	viewer.lhImgRect = image.Rect(50, 50, 60, 60)

	// Test outside
	assert.False(t, viewer.mouseAtUpperHysteresisPoint(0, 0))
	assert.False(t, viewer.mouseAtLowerPoint(0, 0))
	assert.False(t, viewer.mouseAtLowerHysteresisPoint(0, 0))

	cursor, ok := viewer.cursor(-100, -100)
	assert.False(t, ok)
	assert.NotNil(t, cursor)

	// Test upper hysteresis hit
	assert.True(t, viewer.mouseAtUpperHysteresisPoint(15, 15))
	cursor, ok = viewer.cursor(15, 15)
	assert.True(t, ok)
	assert.NotNil(t, cursor)

	// Test lower hit
	assert.True(t, viewer.mouseAtLowerPoint(35, 35))
	cursor, ok = viewer.cursor(35, 35)
	assert.True(t, ok)
	assert.NotNil(t, cursor)

	// Test lower hysteresis hit
	assert.True(t, viewer.mouseAtLowerHysteresisPoint(55, 55))
	cursor, ok = viewer.cursor(55, 55)
	assert.True(t, ok)
	assert.NotNil(t, cursor)
}
