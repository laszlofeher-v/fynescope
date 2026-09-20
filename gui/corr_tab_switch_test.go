package gui

import (
	"fynescope/control"
	"fynescope/settings"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/theme"
	"github.com/stretchr/testify/assert"
)

func TestAllTabsTap(t *testing.T) {
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

	ftContent := container.NewVBox()
	scp.ftTab = container.NewTabItem(tabNames[ftTabIndex], ftContent)
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

	scp.corrLayout = container.NewMax()
	scp.corrTab = container.NewTabItem(tabNames[corrTabIndex], scp.corrLayout)
	scp.newCorrPanel(scp.corrLayout, true)

	scp.controlTab = container.NewAppTabs(
		scp.ftTab, scp.fvTab, scp.dftTab, scp.ffTab, scp.rlcTab, scp.digPortTab, scp.genTab, scp.extgenTab, scp.digGenTab, scp.vchTab, scp.decodeTab, scp.filterTab, scp.corrTab)

	scp.Settings.Window.Function = ftTabIndex
	scp.Settings.Window.LastDispFunction = ftTabIndex

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
			tab == scp.corrTab {
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

	w.SetContent(scp.controlTab)
	w.Resize(fyne.NewSize(300, 600)) // Narrow width like controlTab on side of oscilloscope!

	r := test.WidgetRenderer(scp.controlTab)
	assert.NotNil(t, r)
	buttons := r.Objects()[0].(*fyne.Container).Objects[0].(*fyne.Container).Objects
	t.Logf("Number of tab buttons rendered in bar: %d (out of %d total tabs)", len(buttons), len(scp.controlTab.Items))

	for i, b := range buttons {
		t.Logf("Button %d: %+v", i, b)
	}

	// Select corr tab
	scp.controlTab.Select(scp.corrTab)
	t.Logf("Selected after select corr: %v", scp.controlTab.Selected().Text)
	assert.Equal(t, scp.corrTab, scp.controlTab.Selected())
	assert.True(t, scp.corrLayout.Visible())
	assert.False(t, ftContent.Visible())

	// Select corr tab again
	scp.controlTab.Select(scp.corrTab)
	assert.Equal(t, scp.corrTab, scp.controlTab.Selected())

	t.Logf("Tapping via TapCanvas at (10, 10)...")
	test.TapCanvas(w.Canvas(), fyne.NewPos(10, 10))
	t.Logf("Selected after TapCanvas at (10, 10): %v", scp.controlTab.Selected().Text)
	assert.Equal(t, scp.ftTab, scp.controlTab.Selected())
}
