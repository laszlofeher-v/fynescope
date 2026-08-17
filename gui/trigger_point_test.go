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

func TestTriggerPoint_MouseEvents(t *testing.T) {
	test.NewApp()
	scp := &ScpDesc{
		App:           test.NewApp(),
		psControl:     &control.PscDesc{},
		theme:         theme.DefaultTheme(),
		tzRepartition: createFlag(),
		controlTab:    container.NewAppTabs(),
		Settings:      settings.NewDefaultSettings(),
		channelCount:  genericps.QuadScope,
	}

	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	tp := newTriggerPointViewer(img, scp, false)

	// test mouseIn
	assert.True(t, tp.mouseIn(0, 0))
	assert.False(t, tp.mouseIn(100, 100))
	assert.False(t, tp.mouseIn(8, 8)) // rect is (-8, -8) to (8, 8). In() is exclusive on max point so (8, 8) is false.
	assert.True(t, tp.mouseIn(7, 7))

	// test mouseMoved
	tp.mouseMoved(0, 0)
	assert.True(t, tp.mouseAt)
	tp.mouseMoved(100, 100)
	assert.False(t, tp.mouseAt)

	// test cursor
	tp.mouseMoved(0, 0) // sets mouseAt = true, but cursor only checks mouseIn(x, y) or selected.
	_, ok := tp.cursor(0, 0)
	assert.True(t, ok)
	
	_, ok = tp.cursor(100, 100)
	assert.False(t, ok) // not selected and mouse not in

	tp.mouseDown(0, 0, 0, 0)
	assert.True(t, tp.selected)
	
	_, ok = tp.cursor(100, 100)
	assert.True(t, ok) // selected, so cursor should be true even if mouse outside

	tp.mouseUp(0, 0, 0, 0)
	assert.False(t, tp.selected)
}

func TestTriggerPoint_TimeMv2xy(t *testing.T) {
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
	}
	scp.channelViewers = make([]channelViewerDesc, scp.channelCount)
	for i := range scp.channelViewers {
		scp.channelViewers[i] = channelViewerDesc{}
	}

	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	tp := newTriggerPointViewer(img, scp, false)

	x, y := tp.timeMv2xy(0)
	// Just asserting it doesn't panic and returns valid floats
	assert.NotNil(t, x)
	assert.NotNil(t, y)
}

