package gui

import (
	"fynescope/control"
	"image"
	"image/draw"

	"fyne.io/fyne/v2"
)

func (scp *ScpDesc) timeZoomGenerator(wInt int, hInt int) image.Image {
	defer scp.screenLocker.Unlock()
	scp.screenLocker.Lock()

	w := float32(wInt)
	h := float32(hInt)

	if scp.timeZoomScopeFullScreen == nil || scp.timeZoomScopeFullScreen.Bounds().Dx() != wInt || scp.timeZoomScopeFullScreen.Bounds().Dy() != hInt {
		scp.timeZoomScopeFullScreen = scp.newScopeScreen(image.Point{wInt, hInt})
		switch scp.triggerSettingMsg.Type {
		case control.Simple:
			scp.timeZoomTriggerPoint = newTriggerPointViewer(scp.timeZoomScopeFullScreen, scp, true)
		case control.Advanced:
			scp.timeZoomTriggerPoint = newAdvTriggerPointViewer(scp.timeZoomScopeFullScreen, scp, true)
		case control.Window:
			scp.timeZoomTriggerPoint = newWindowTriggerPointViewer(scp.timeZoomScopeFullScreen, scp, true)
		case control.Complex:
			scp.timeZoomTriggerPoint = newComplexTriggerPointViewer(scp.timeZoomScopeFullScreen, scp, true)
		case control.Interval, control.PulseWidth:
			scp.timeZoomTriggerPoint = newIntervalTriggerPointViewer(scp.timeZoomScopeFullScreen, scp, true)
		default:
			scp.timeZoomTriggerPoint = newTriggerPointViewer(scp.timeZoomScopeFullScreen, scp, true)
		}
		scp.partitionTzScreen(w, h)
		draw.Draw(scp.timeZoomScopeFullScreen, scp.timeZoomScopeFullScreen.Bounds(), &image.Uniform{scp.theme.Color(ColorNameSignalBackground, 0)}, image.ZP, draw.Src)
		scp.setTzVDivsY()
		scp.setTzHDivsX()
	} else if getFlag(scp.tzRepartition) {
		switch scp.triggerSettingMsg.Type {
		case control.Simple:
			scp.timeZoomTriggerPoint = newTriggerPointViewer(scp.timeZoomScopeFullScreen, scp, true)
		case control.Advanced:
			scp.timeZoomTriggerPoint = newAdvTriggerPointViewer(scp.timeZoomScopeFullScreen, scp, true)
		case control.Window:
			scp.timeZoomTriggerPoint = newWindowTriggerPointViewer(scp.timeZoomScopeFullScreen, scp, true)
		case control.Complex:
			scp.timeZoomTriggerPoint = newComplexTriggerPointViewer(scp.timeZoomScopeFullScreen, scp, true)
		case control.Interval, control.PulseWidth:
			scp.timeZoomTriggerPoint = newIntervalTriggerPointViewer(scp.timeZoomScopeFullScreen, scp, true)
		default:
			scp.timeZoomTriggerPoint = newTriggerPointViewer(scp.timeZoomScopeFullScreen, scp, true)
		}
		scp.partitionTzScreen(w, h)
		draw.Draw(scp.timeZoomScopeFullScreen, scp.timeZoomScopeFullScreen.Bounds(), &image.Uniform{scp.theme.Color(ColorNameSignalBackground, 0)}, image.ZP, draw.Src)
		scp.setTzVDivsY()
		scp.setTzHDivsX()
	} else {
		draw.Draw(scp.timeZoomScopeFullScreen, scp.timeZoomScopeSignalScreen.Bounds(), &image.Uniform{scp.theme.Color(ColorNameSignalBackground, 0)}, image.ZP, draw.Src)
		scp.setTzVDivsY()
		scp.setTzHDivsX()
	}

	for i := range scp.timeZoomDrawers {
		scp.timeZoomDrawers[i].draw()
	}
	return scp.timeZoomScopeFullScreen
}

func (scp *ScpDesc) openTimeZoomWindow() {
	if scp.timeZoomWindow != nil {
		scp.timeZoomWindow.RequestFocus()
		return
	}
	scp.timeZoomWindow = scp.App.NewWindow("Time Zoom")
	scp.timeZoomWindow.SetOnClosed(func() {
		scp.timeZoomWindow = nil
		scp.timeZoomRaster = nil
		scp.timeZoomScopeFullScreen = nil
		scp.timeZoomScopeSignalScreen = nil
		scp.timeZoomDrawers = nil
		scp.timeZoomBoxOffset = 0
		scp.setMaxScreenTime()
		scp.clearAllFtPersistentLayers()
		scp.clearAllDftPersistentLayers()
		scp.refreshRasters()
		if scp.ftBottomLabelViewer != nil {
			scp.ftBottomLabelViewer.(*timeLabelViewer).enableRefresh()
		}
	})

	scp.timeZoomMaxScreenTime = scp.maxScreenTime
	scp.timeZoomTimeDiv = scp.timeDiv
	scp.timeZoomTimeUnit = scp.timeUnit
	scp.timeZoomBoxOffset = 0

	// Trigger repartition for Time Zoom
	setFlag(scp.tzRepartition)

	scp.timeZoomRaster = scp.newScreenRaster(scp.timeZoomGenerator, scp.timeZoomWindow, false, false, false)
	scp.timeZoomRaster.disableInput = false

	scp.timeZoomWindow.SetContent(scp.timeZoomRaster)
	scp.timeZoomWindow.Resize(fyne.NewSize(800, 600))
	scp.timeZoomWindow.Show()
}
