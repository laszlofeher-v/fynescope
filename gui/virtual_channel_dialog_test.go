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

func TestVirtualChannelDialog_BuildContent(t *testing.T) {
	test.NewApp()
	scp := &ScpDesc{
		App:           test.NewApp(),
		psControl:     &control.PscDesc{},
		theme:         theme.DefaultTheme(),
		tzRepartition: createFlag(),
		repartition:   createFlag(),
		Window:        test.NewWindow(container.NewVBox()), // provide a valid window
		controlTab:    container.NewAppTabs(),
		Settings:      settings.NewDefaultSettings(),
		channelCount:  genericps.QuadScope,
	}

	content := scp.buildVirtualChannelContent(true)
	assert.NotNil(t, content)

	// Verify that disp7 digits were successfully initialized during build
	assert.NotNil(t, scp.vchMinV)
	assert.NotNil(t, scp.vchMaxV)
	assert.NotNil(t, scp.vchFrq)
	assert.NotNil(t, scp.vchPeriod)

	// Open Virtual Channel Dialog
	scp.vchTab = container.NewTabItem("VCh", content)
	scp.controlTab.Append(scp.vchTab)
	
	// Open while docked
	scp.openVirtualChannelDialog()
	assert.Equal(t, scp.vchTab, scp.controlTab.Selected())

	// Test close of Virtual Channel Window by firing the callback if possible, or just skip it since it's an inline function
	if scp.virtualChWindow != nil {
		scp.virtualChWindow.Close()
	}
}
