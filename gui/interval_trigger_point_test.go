package gui

import (
	"fynescope/control"
	"fynescope/genericps"
	"fynescope/settings"
	"image"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func setupIntervalTriggerScp() *ScpDesc {
	genericps.Range_5v = 8
	genericps.Range_1v = 6
	genericps.TriggerRising = genericps.ThresholdDirection(2)
	genericps.TriggerFalling = genericps.ThresholdDirection(3)
	genericps.InputRanges = make([]int32, 40)
	genericps.InputRanges[genericps.Range_5v] = 5000
	genericps.InputRanges[genericps.Range_1v] = 1000
	genericps.RangeValuesMv = map[genericps.RangeEnum]float64{
		genericps.Range_5v: 5000.0,
		genericps.Range_1v: 1000.0,
	}

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

func TestIntervalTriggerPointViewer_RiseFallDraw(t *testing.T) {
	scp := setupIntervalTriggerScp()
	img := image.NewRGBA64(image.Rect(0, 0, 100, 100))
	scp.ftScopeSignalScreen = img
	scp.maxScreenTime = 100.0
	scp.triggerSource = 0
	scp.channelViewers = make([]channelViewerDesc, 2)
	scp.Settings.Channels[0].TriggerSource = true
	scp.Settings.Channels[0].Trigger.Mv = 1000
	scp.Settings.Channels[0].Trigger.LowerMv = -1000
	scp.Settings.Channels[0].Trigger.IntervalType = genericps.PwTypeLessThan
	scp.Settings.Channels[0].Trigger.IntervalTimeUpper = 20.0

	viewer := newIntervalTriggerPointViewer(img, scp, false)

	// 1. RiseFall with Rising direction -> Should have 1 arrow at lower threshold
	scp.triggerSettingMsg.Type = control.RiseFall
	scp.Settings.Channels[0].Trigger.Type = settings.TriggerTypeRiseFall
	scp.Settings.Channels[0].Trigger.TriggerDirection = genericps.TriggerRising
	viewer.draw()
	assert.Equal(t, 1, len(viewer.lowerHImgRects), "Should draw 1 handle for RiseFall Rising")
	_, expectedYLower := viewer.timeMv2xy(-1000)
	assert.Equal(t, int(math.Round(float64(expectedYLower))), (viewer.lowerHImgRects[0].Min.Y+viewer.lowerHImgRects[0].Max.Y)/2)

	// 2. RiseFall with Falling direction -> Should have 1 arrow at upper threshold
	scp.Settings.Channels[0].Trigger.TriggerDirection = genericps.TriggerFalling
	viewer.draw()
	assert.Equal(t, 1, len(viewer.lowerHImgRects), "Should draw 1 handle for RiseFall Falling")
	_, expectedYUpper := viewer.timeMv2xy(1000)
	assert.Equal(t, int(math.Round(float64(expectedYUpper))), (viewer.lowerHImgRects[0].Min.Y+viewer.lowerHImgRects[0].Max.Y)/2)

	// 3. WindowPulseWidth -> Should have 2 arrows (both thresholds)
	scp.triggerSettingMsg.Type = control.WindowPulseWidth
	scp.Settings.Channels[0].Trigger.Type = settings.TriggerTypeWindowPulseWidth
	viewer.draw()
	assert.Equal(t, 2, len(viewer.lowerHImgRects), "Should draw 2 handles for WindowPulseWidth")
}
