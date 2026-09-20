package gui

import (
	"fynescope/control"
	"fynescope/genericps"
	"fynescope/settings"
	"testing"

	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/theme"
	"github.com/stretchr/testify/assert"
)

func TestTimeZoom_OpenClose(t *testing.T) {
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

	scp.openTimeZoomWindow()
	assert.NotNil(t, scp.timeZoomWindow)
	assert.NotNil(t, scp.timeZoomRaster)

	// Calling again should just focus, not create a second window
	scp.openTimeZoomWindow()
	assert.NotNil(t, scp.timeZoomWindow)

	// Trigger the close callback and verify cleanup
	scp.timeZoomWindow.Close()

	assert.Nil(t, scp.timeZoomWindow)
	assert.Nil(t, scp.timeZoomRaster)
	assert.Nil(t, scp.timeZoomScopeFullScreen)
}
