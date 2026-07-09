package gui

import (
	"image"
	"math"

	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
)

type intervalTriggerPointViewer struct {
	advTriggerPointViewer
	lowerHImgRect image.Rectangle
	upperHImgRect image.Rectangle
	lowerSelected bool
	upperSelected bool
}

var (
	_ mouser     = (*intervalTriggerPointViewer)(nil)
	_ dragger    = (*intervalTriggerPointViewer)(nil)
	_ scroller   = (*intervalTriggerPointViewer)(nil)
	_ drawer     = (*intervalTriggerPointViewer)(nil)
	_ cursorable = (*intervalTriggerPointViewer)(nil)
)

func newIntervalTriggerPointViewer(img rasterImage, scp *ScpDesc) *intervalTriggerPointViewer {
	tp := &intervalTriggerPointViewer{
		advTriggerPointViewer: *newAdvTriggerPointViewer(img, scp),
	}
	return tp
}

func (tp *intervalTriggerPointViewer) cursor(x, y float32) (desktop.Cursor, bool) {
	if tp.scp.inStreamMode() {
		return desktop.DefaultCursor, false
	}
	cp, ok := tp.advTriggerPointViewer.cursor(x, y)
	if ok {
		return cp, ok
	}
	if tp.mouseAtIntervalPoint(x, y) {
		return desktop.PointerCursor, true
	}
	return desktop.DefaultCursor, false
}

func (tp *intervalTriggerPointViewer) mouseAtIntervalPoint(x, y float32) bool {
	p := image.Point{X: int(math.Round(float64(x))), Y: int(math.Round(float64(y)))}
	if p.In(tp.lowerHImgRect) || p.In(tp.upperHImgRect) {
		return true
	}
	return false
}

func (tp *intervalTriggerPointViewer) mouseMoved(x, y float32) {
	if tp.scp.inStreamMode() {
		return
	}
	tp.advTriggerPointViewer.mouseMoved(x, y)
	if tp.advTriggerPointViewer.mouseAt {
		tp.mouseAt = false
		return
	}
	prev := tp.mouseAt
	if tp.mouseAtIntervalPoint(x, y) {
		tp.mouseAt = true
	} else {
		tp.mouseAt = false
	}
	if prev != tp.mouseAt {
		tp.enableRefresh()
		if tp.scp.ftRaster != nil {
			tp.scp.ftRaster.Refresh()
		}
	}
}

func (tp *intervalTriggerPointViewer) mouseDown(button desktop.MouseButton, x, y float32) {
	if tp.scp.inStreamMode() {
		return
	}
	tp.advTriggerPointViewer.mouseDown(button, x, y)
	if !tp.selected && !tp.uhSelected {
		p := image.Point{X: int(math.Round(float64(x))), Y: int(math.Round(float64(y)))}
		tp.lowerSelected = p.In(tp.lowerHImgRect)
		tp.upperSelected = p.In(tp.upperHImgRect)
	}
}

func (tp *intervalTriggerPointViewer) mouseUp(button desktop.MouseButton, x, y float32) {
	if tp.scp.inStreamMode() {
		return
	}
	tp.advTriggerPointViewer.mouseUp(button, x, y)
	prev := tp.lowerSelected || tp.upperSelected
	tp.lowerSelected = false
	tp.upperSelected = false
	if prev {
		if !tp.mouseAtIntervalPoint(x, y) {
			tp.mouseAt = false
		}
		tp.enableRefresh()
		if tp.scp.ftRaster != nil {
			tp.scp.ftRaster.Refresh()
		}
	}
}

func (tp *intervalTriggerPointViewer) dragged(dx, dy, x, y float32) {
	if tp.scp.inStreamMode() {
		return
	}
	if tp.scp.triggerSource < 0 || int(tp.scp.triggerSource) >= len(tp.scp.Settings.Channels) {
		return
	}
	tp.advTriggerPointViewer.dragged(dx, dy, x, y)

	if !tp.lowerSelected && !tp.upperSelected {
		return
	}

	channel := &tp.scp.Settings.Channels[tp.scp.triggerSource]
	bounds := tp.scp.ftScopeSignalScreen.Bounds()
	w := float64(bounds.Dx())

	if w <= 0 || tp.scp.maxScreenTime <= 0 {
		return
	}

	// Convert x coordinate to time offset from trigger point
	triggerX, _ := tp.timeMv2xy(channel.Trigger.Mv)
	timeOffset := (float64(x-triggerX) / w) * tp.scp.maxScreenTime
	if timeOffset < 0 {
		timeOffset = -timeOffset
	}

	if tp.lowerSelected {
		if channel.Trigger.IntervalTimeUpper > 0 && timeOffset > channel.Trigger.IntervalTimeUpper {
			timeOffset = channel.Trigger.IntervalTimeUpper
		}
		channel.Trigger.IntervalTimeLower = timeOffset
	} else if tp.upperSelected {
		if timeOffset < channel.Trigger.IntervalTimeLower {
			timeOffset = channel.Trigger.IntervalTimeLower
		}
		channel.Trigger.IntervalTimeUpper = timeOffset
	}

	tp.scp.triggerSettingMsg.IntervalTimeLower = channel.Trigger.IntervalTimeLower
	tp.scp.triggerSettingMsg.IntervalTimeUpper = channel.Trigger.IntervalTimeUpper

	tp.scp.psControl.SetTriggerCh <- &tp.scp.triggerSettingMsg
	<-tp.scp.triggerSettingMsg.Done

	tp.enableRefresh()
	if tp.scp.ftRaster != nil {
		tp.scp.ftRaster.Refresh()
	}
}

func (tp *intervalTriggerPointViewer) draw() {
	if tp.scp.controlTab.SelectedIndex() == dftTabIndex || tp.scp.inStreamMode() {
		return
	}
	tp.advTriggerPointViewer.draw()

	if tp.scp.triggerSource != dontCare {
		channel := &tp.scp.Settings.Channels[tp.scp.triggerSource]
		x, y := tp.timeMv2xy(channel.Trigger.Mv)
		bounds := tp.scp.ftScopeSignalScreen.Bounds()

		w := float64(bounds.Dx())

		if w <= 0 || tp.scp.maxScreenTime <= 0 {
			return
		}

		lowerDx := float32((channel.Trigger.IntervalTimeLower / tp.scp.maxScreenTime) * w)
		upperDx := float32((channel.Trigger.IntervalTimeUpper / tp.scp.maxScreenTime) * w)

		var xLower float32
		var xUpper float32

		// Pulse width is measured after the trigger point
		xLower = x + lowerDx
		xUpper = x + upperDx

		halfRectSize := float32(triggerPointR * 2)
		rectSize2 := 2 * halfRectSize

		tp.lowerHImgRect = image.Rect(
			int(math.Round(float64(xLower-rectSize2))),
			int(math.Round(float64(y-rectSize2))),
			int(math.Round(float64(xLower+rectSize2))),
			int(math.Round(float64(y+rectSize2))))

		tp.upperHImgRect = image.Rect(
			int(math.Round(float64(xUpper-rectSize2))),
			int(math.Round(float64(y-rectSize2))),
			int(math.Round(float64(xUpper+rectSize2))),
			int(math.Round(float64(y+rectSize2))))

		colLower := theme.ForegroundColor()
		if tp.lowerSelected || (tp.mouseAt && tp.mouseAtIntervalPoint(float32(tp.lowerHImgRect.Min.X+tp.lowerHImgRect.Dx()/2), y)) {
			colLower = theme.SelectionColor()
		}

		colUpper := theme.ForegroundColor()
		if tp.upperSelected || (tp.mouseAt && tp.mouseAtIntervalPoint(float32(tp.upperHImgRect.Min.X+tp.upperHImgRect.Dx()/2), y)) {
			colUpper = theme.SelectionColor()
		}

		// Draw horizontal line from trigger point to xLower
		drawLine(tp.scp.ftScopeSignalScreen, x, y, xLower, y, colLower)
		// Draw vertical handle at xLower
		drawLine(tp.scp.ftScopeSignalScreen, xLower, y-halfRectSize, xLower, y+halfRectSize, colLower)

		// Draw horizontal line from trigger point to xUpper
		drawLine(tp.scp.ftScopeSignalScreen, x, y, xUpper, y, colUpper)
		// Draw vertical handle at xUpper
		drawLine(tp.scp.ftScopeSignalScreen, xUpper, y-halfRectSize, xUpper, y+halfRectSize, colUpper)

		// Update UI labels if necessary
		if tp.scp.intervalTimeLowerDisp != nil {
			unit := channel.Trigger.IntervalTimeUnit
			if unit == "" && tp.scp.intervalUnitSelect != nil {
				unit = tp.scp.intervalUnitSelect.Selected
			}
			multiplier := getIntervalUnitMultiplier(unit)
			val := int(math.Round(channel.Trigger.IntervalTimeLower / multiplier))
			if tp.scp.intervalTimeLowerDisp.Value != val {
				tp.scp.intervalTimeLowerDisp.SilentSetValue(val)
				tp.scp.intervalTimeLowerDisp.Refresh()
			}
		}

		if tp.scp.intervalTimeUpperDisp != nil {
			unit := channel.Trigger.IntervalTimeUnit
			if unit == "" && tp.scp.intervalUnitSelect != nil {
				unit = tp.scp.intervalUnitSelect.Selected
			}
			multiplier := getIntervalUnitMultiplier(unit)
			val := int(math.Round(channel.Trigger.IntervalTimeUpper / multiplier))
			if tp.scp.intervalTimeUpperDisp.Value != val {
				tp.scp.intervalTimeUpperDisp.SilentSetValue(val)
				tp.scp.intervalTimeUpperDisp.Refresh()
			}
		}
	}
}

func (tp *intervalTriggerPointViewer) scrolled(delta, x, y float32) {
	if tp.scp.inStreamMode() {
		return
	}
	tp.advTriggerPointViewer.scrolled(delta, x, y)
}
