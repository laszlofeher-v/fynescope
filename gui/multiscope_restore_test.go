package gui

import (
	"fynescope/control"
	"fynescope/settings"
	"testing"

	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/theme"
	"github.com/stretchr/testify/assert"
)

func TestMultiscopeTabRestorationAndRasters(t *testing.T) {
	app := test.NewApp()
	w := app.NewWindow("Test")
	scp := &ScpDesc{
		channelCount: 2,
		App:          app,
		Window:       w,
		Settings:     settings.NewDefaultSettings(),
		psControl:    &control.PscDesc{},
		theme:        theme.DefaultTheme(),
	}

	// Set saved settings: multiscope was the last function, with f(v) as last displayed function
	scp.Settings.Window.Function = multiTabIndex
	scp.Settings.Window.LastDispFunction = fvTabIndex

	makeRaster := func() *screenRaster {
		r := &screenRaster{scp: scp}
		r.ExtendBaseWidget(r)
		return r
	}

	scp.ftRaster = makeRaster()
	scp.fvRaster = makeRaster()
	scp.dftRaster = makeRaster()
	scp.ffRaster = makeRaster()
	scp.activeRasterContainer = container.NewMax(scp.ftRaster, scp.fvRaster, scp.dftRaster, scp.ffRaster)

	scp.ftTab = container.NewTabItem(tabNames[ftTabIndex], container.NewVBox())
	scp.fvTab = container.NewTabItem(tabNames[fvTabIndex], container.NewVBox())
	scp.dftTab = container.NewTabItem(tabNames[dftTabIndex], container.NewVBox())
	scp.ffTab = container.NewTabItem(tabNames[ffTabIndex], container.NewVBox())
	scp.rlcTab = container.NewTabItem(tabNames[rlcTabIndex], container.NewVBox())
	scp.digPortTab = container.NewTabItem(tabNames[digPortTabIndex], container.NewVBox())
	scp.genTab = container.NewTabItem(tabNames[genTabIndex], container.NewVBox())
	scp.extgenTab = container.NewTabItem(tabNames[extgenTabIndex], container.NewVBox())
	scp.digGenTab = container.NewTabItem(tabNames[digGenTabIndex], container.NewVBox())
	scp.vchTab = container.NewTabItem(tabNames[vchTabIndex], container.NewVBox())
	scp.decodeTab = container.NewTabItem(tabNames[decodeTabIndex], container.NewVBox())
	scp.filterTab = container.NewTabItem(tabNames[filterTabIndex], container.NewVBox())
	scp.corrTab = container.NewTabItem(tabNames[corrTabIndex], container.NewVBox())
	scp.multiLayout = container.NewVBox()
	scp.multiTab = container.NewTabItem(tabNames[multiTabIndex], scp.multiLayout)

	scp.controlTab = container.NewAppTabs(
		scp.ftTab, scp.fvTab, scp.dftTab, scp.ffTab, scp.rlcTab, scp.digPortTab, scp.genTab, scp.extgenTab, scp.digGenTab, scp.vchTab, scp.decodeTab, scp.filterTab, scp.corrTab, scp.multiTab,
	)

	// Simulate build2000Gui startup logic
	initTabItem := scp.getTabItem(scp.Settings.Window.Function)
	isTabPresent := false
	for _, item := range scp.controlTab.Items {
		if item == initTabItem {
			isTabPresent = true
			break
		}
	}
	if !isTabPresent {
		initTabItem = scp.ftTab
	}

	scp.controlTab.OnSelected = func(tab *container.TabItem) {
		prevTab := scp.Settings.Window.Function
		newTab := scp.getFunctionIndex(tab)
		scp.handleTabTransition(prevTab, newTab)

		scp.Settings.Window.Function = newTab
		switch newTab {
		case ftTabIndex, fvTabIndex, dftTabIndex, ffTabIndex:
			scp.Settings.Window.LastDispFunction = newTab
		}

		targetFunction := scp.Settings.Window.Function
		if tab == scp.genTab ||
			tab == scp.filterTab ||
			tab == scp.extgenTab ||
			tab == scp.vchTab ||
			tab == scp.digPortTab ||
			tab == scp.corrTab ||
			tab == scp.multiTab {
			targetFunction = scp.Settings.Window.LastDispFunction
		}

		switch targetFunction {
		case dftTabIndex:
			scp.ftRaster.Hide()
			scp.fvRaster.Hide()
			scp.ffRaster.Hide()
			scp.dftRaster.Show()
		case fvTabIndex:
			scp.ftRaster.Hide()
			scp.dftRaster.Hide()
			scp.ffRaster.Hide()
			scp.fvRaster.Show()
		case ffTabIndex:
			scp.ftRaster.Hide()
			scp.dftRaster.Hide()
			scp.fvRaster.Hide()
			scp.ffRaster.Show()
		default:
			scp.dftRaster.Hide()
			scp.fvRaster.Hide()
			scp.ffRaster.Hide()
			scp.ftRaster.Show()
		}
	}

	if initTabItem != nil {
		if scp.controlTab.Selected() == initTabItem {
			if scp.controlTab.OnSelected != nil {
				scp.controlTab.OnSelected(initTabItem)
			}
		} else {
			scp.controlTab.Select(initTabItem)
		}
	}

	// 1. Verify multiscope tab is selected on startup (not f(v))
	assert.Equal(t, scp.multiTab, scp.controlTab.Selected(), "multiscope tab should be selected on startup")
	assert.Equal(t, multiTabIndex, scp.Settings.Window.Function, "Function setting should remain multiTabIndex")

	// 2. Verify f(v) raster is visible and active
	assert.True(t, scp.fvRaster.Visible(), "f(v) raster should be visible when multiscope was opened with LastDispFunction == f(v)")
	assert.False(t, scp.ftRaster.Visible(), "f(t) raster should be hidden")
	assert.False(t, scp.dftRaster.Visible(), "dft raster should be hidden")
	assert.False(t, scp.ffRaster.Visible(), "ff raster should be hidden")

	// 3. Verify shouldDrawRaster returns true for fvTabIndex when multiscope tab is selected
	assert.True(t, scp.shouldDrawRaster(fvTabIndex), "shouldDrawRaster should be true for f(v) raster when multiscope tab is selected")
	assert.False(t, scp.shouldDrawRaster(ftTabIndex), "shouldDrawRaster should be false for f(t) raster when LastDispFunction is f(v)")
}
