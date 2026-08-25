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

	if scp.virtualChWindow != nil {
		scp.virtualChWindow.Close()
	}
}

func TestVirtualChannelDialog_UndockDedup(t *testing.T) {
	scp := &ScpDesc{
		App:           test.NewApp(),
		psControl:     &control.PscDesc{},
		theme:         theme.DefaultTheme(),
		tzRepartition: createFlag(),
		repartition:   createFlag(),
		Window:        test.NewWindow(container.NewVBox()),
		controlTab:    container.NewAppTabs(),
		Settings:      settings.NewDefaultSettings(),
		channelCount:  genericps.QuadScope,
	}

	content := scp.buildVirtualChannelContent(true)
	scp.vchTab = container.NewTabItem("VCh", content)
	scp.controlTab.Append(scp.vchTab)

	// Simulate first undock
	w1 := test.NewWindow(container.NewVBox())
	scp.virtualChWindow = w1

	// Calling openVirtualChannelDialog or undock should not create a new window
	scp.openVirtualChannelDialog()
	assert.Equal(t, w1, scp.virtualChWindow)

	w1.Close()
}
