package gui

import (
	"fynescope/control"
	"fynescope/genericps"
	"fynescope/settings"
	"image"
	"testing"

	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/theme"
	"github.com/stretchr/testify/assert"
)

func newWindowTriggerPointViewerForTest(img rasterImage, scp *ScpDesc, isTimeZoom bool) *windowTriggerPointViewer {
	return &windowTriggerPointViewer{
		triggerPointViewer: *newTriggerPointViewer(img, scp, isTimeZoom),
	}
}

func TestWindowTriggerPointViewer_MouseEvents(t *testing.T) {
	test.NewApp()
	scp := &ScpDesc{
		App:                 test.NewApp(),
		psControl:           &control.PscDesc{},
		theme:               theme.DefaultTheme(),
		tzRepartition:       createFlag(),
		controlTab:          container.NewAppTabs(),
		Settings:            settings.NewDefaultSettings(),
		channelCount:        genericps.QuadScope,
		ftScopeSignalScreen: image.NewRGBA(image.Rect(0, 0, 100, 100)),
		maxScreenTime:       100.0,
		repartition:         createFlag(),
	}
	scp.channelViewers = make([]channelViewerDesc, scp.channelCount)
	for i := range scp.channelViewers {
		scp.channelViewers[i] = channelViewerDesc{}
	}

	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	tp := newWindowTriggerPointViewerForTest(img, scp, false)

	// Inject rects to test mouse hover bounds since draw() hasn't populated them
	tp.uhImgRect = image.Rect(10, 10, 20, 20)
	tp.lImgRect = image.Rect(30, 30, 40, 40)
	tp.lhImgRect = image.Rect(50, 50, 60, 60)
	// Base trigger point rect
	tp.imgRect = image.Rect(-8, -8, 8, 8)

	// Test bounds for upper hysteresis
	assert.True(t, tp.mouseAtUpperHysteresisPoint(15, 15))
	assert.False(t, tp.mouseAtUpperHysteresisPoint(25, 25))

	// Test bounds for lower threshold
	assert.True(t, tp.mouseAtLowerPoint(35, 35))
	assert.False(t, tp.mouseAtLowerPoint(45, 45))

	// Test bounds for lower hysteresis
	assert.True(t, tp.mouseAtLowerHysteresisPoint(55, 55))
	assert.False(t, tp.mouseAtLowerHysteresisPoint(65, 65))

	// Test mouseMoved triggers hover states
	tp.mouseMoved(15, 15)
	assert.True(t, tp.uhMouseAt)
	assert.False(t, tp.lMouseAt)
	assert.False(t, tp.lhMouseAt)

	// Test cursor for upper hysteresis
	_, ok := tp.cursor(15, 15)
	assert.True(t, ok)

	// Test cursor for lower threshold
	_, ok = tp.cursor(35, 35)
	assert.True(t, ok)

	// Test cursor outside
	_, ok = tp.cursor(100, 100)
	assert.False(t, ok)

	// Test mouseDown sets selection
	tp.mouseMoved(35, 35)
	tp.mouseDown(0, 0, 35, 35) // inside lImgRect
	assert.True(t, tp.lSelected)
	assert.False(t, tp.uhSelected)

	// Test mouseUp clears selection
	tp.mouseUp(0, 0, 35, 35)
	assert.False(t, tp.lSelected)

	// Test mouseDown for lhImgRect
	tp.mouseMoved(55, 55)
	tp.mouseDown(0, 0, 55, 55)
	assert.True(t, tp.lhSelected)
	assert.False(t, tp.lSelected)
	tp.mouseUp(0, 0, 55, 55)
	assert.False(t, tp.lhSelected)
}
