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

func TestWindowPulseWidthTriggerPointViewer_MouseEvents(t *testing.T) {
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
	tp := newWindowPulseWidthTriggerPointViewer(img, scp, false)

	// Inject rects to test mouse hover bounds since draw() hasn't populated them
	tp.lImgRect = image.Rect(10, 10, 20, 20)
	tp.lhImgRect = image.Rect(30, 30, 40, 40)

	// Test bounds for lImgRect
	assert.True(t, tp.mouseAtLowerPoint(15, 15))
	assert.False(t, tp.mouseAtLowerPoint(25, 25))

	// Test bounds for lhImgRect
	assert.True(t, tp.mouseAtLowerHysteresisPoint(35, 35))
	assert.False(t, tp.mouseAtLowerHysteresisPoint(45, 45))

	// Test mouseMoved triggers hover states
	tp.mouseMoved(15, 15)
	assert.True(t, tp.lMouseAt)
	assert.False(t, tp.lhMouseAt)

	// Test cursor for lower threshold
	_, ok := tp.cursor(15, 15)
	assert.True(t, ok)

	// Test cursor for lower hysteresis
	_, ok = tp.cursor(35, 35)
	assert.True(t, ok)

	// Test cursor outside
	_, ok = tp.cursor(100, 100)
	assert.False(t, ok)

	// Test mouseDown sets selection
	tp.mouseMoved(15, 15)
	tp.mouseDown(0, 0, 15, 15) // inside lImgRect
	assert.True(t, tp.lSelected)
	assert.False(t, tp.lhSelected)

	// Test mouseUp clears selection
	tp.mouseUp(0, 0, 15, 15)
	assert.False(t, tp.lSelected)

	// Test mouseDown for lhImgRect
	tp.mouseMoved(35, 35)
	tp.mouseDown(0, 0, 35, 35)
	assert.True(t, tp.lhSelected)
	assert.False(t, tp.lSelected)
	tp.mouseUp(0, 0, 35, 35)
	assert.False(t, tp.lhSelected)
}
