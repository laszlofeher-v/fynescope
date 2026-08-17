package gui

import (
	"fynescope/genericps"
	"fynescope/settings"
	"image"
	"testing"

	"github.com/stretchr/testify/assert"
)

func setupComplexTriggerScp() *ScpDesc {
	genericps.InputRanges = []int32{5000, 1000}
	genericps.RangeValuesMv = map[genericps.RangeEnum]float64{
		genericps.RangeEnum(30): 5000.0,
		genericps.RangeEnum(20): 1000.0,
	}
	return &ScpDesc{
		channelCount: 2,
		Settings: &settings.PsSettings{
			Channels: []settings.ChSettings{
				{VRange: genericps.RangeEnum(30), Inverted: false},
				{VRange: genericps.RangeEnum(20), Inverted: true},
			},
			Time: settings.TimeSettings{
				TriggerTimeOffset: 0,
			},
		},
		channelViewers: []channelViewerDesc{
			{displayOffsetInt: 0},
			{displayOffsetInt: 0},
		},
		dftDivsX: []float32{10}, // Just dummy data to prevent nil maps or divisions
		ftScopeSignalScreen: image.NewRGBA64(image.Rect(0, 0, 800, 600)),
		timeZoomScopeSignalScreen: image.NewRGBA64(image.Rect(0, 0, 800, 600)),
	}
}

func TestComplexTrigger_timeMv2xy(t *testing.T) {
	scp := setupComplexTriggerScp()
	img := scp.ftScopeSignalScreen

	viewer := newComplexTriggerPointViewer(img, scp, false)

	// In newComplexTriggerPointViewer, it might use methods that expect other maps.
	// We mainly want to test timeMv2xy
	// h = 600
	// zeroOffset = 300
	// yScale = 600 / (2 * 5000) = 0.06

	// mv = 1000
	// y = - (0.06 * 1000) + 300 = 240
	x, y := viewer.timeMv2xy(1000, 0)
	assert.Equal(t, float32(240), y)

	// Test Inverted channel (Channel 1 is 1V range, inverted)
	// h = 600
	// yScale = 600 / (2 * 1000) = 0.3
	// mv = 500 -> inverted to -500
	// y = - (0.3 * -500) + 300 = 300 + 150 = 450
	x2, y2 := viewer.timeMv2xy(500, 1)
	assert.Equal(t, float32(450), y2)

	_ = x
	_ = x2
}

func TestComplexTrigger_y2mv(t *testing.T) {
	scp := setupComplexTriggerScp()
	img := scp.ftScopeSignalScreen

	viewer := newComplexTriggerPointViewer(img, scp, false)

	// Reverse of the above
	mv := viewer.y2mv(240, 0)
	assert.Equal(t, 1000.0, mv)

	mv2 := viewer.y2mv(450, 1)
	assert.Equal(t, 500.0, mv2)
}
