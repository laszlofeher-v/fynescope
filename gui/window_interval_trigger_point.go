package gui

import (
	"fynescope/genericps"
	"image"
	"math"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
)

type windowIntervalTriggerPointViewer struct {
	windowTriggerPointViewer
	lowerHImgRect image.Rectangle
	upperHImgRect image.Rectangle
	lowerSelected bool
	upperSelected bool
}

var (
	_ mouser     = (*windowIntervalTriggerPointViewer)(nil)
	_ dragger    = (*windowIntervalTriggerPointViewer)(nil)
	_ scroller   = (*windowIntervalTriggerPointViewer)(nil)
	_ drawer     = (*windowIntervalTriggerPointViewer)(nil)
	_ cursorable = (*windowIntervalTriggerPointViewer)(nil)
)

func newWindowIntervalTriggerPointViewer(img rasterImage, scp *ScpDesc, isTimeZoom bool) *windowIntervalTriggerPointViewer {
	tp := &windowIntervalTriggerPointViewer{
		windowTriggerPointViewer: *newWindowTriggerPointViewer(img, scp, isTimeZoom),
	}
	return tp
}

func (tp *windowIntervalTriggerPointViewer) cursor(x, y float32) (desktop.Cursor, bool) {
	if tp.scp.inStreamMode() {
		return desktop.DefaultCursor, false
	}
	cp, ok := tp.windowTriggerPointViewer.cursor(x, y)
	if ok {
		return cp, ok
	}
	if tp.mouseAtIntervalPoint(x, y) {
		return desktop.PointerCursor, true
	}
	return desktop.DefaultCursor, false
}

func (tp *windowIntervalTriggerPointViewer) mouseAtIntervalPoint(x, y float32) bool {
	p := image.Point{X: int(math.Round(float64(x))), Y: int(math.Round(float64(y)))}
	if p.In(tp.lowerHImgRect) || p.In(tp.upperHImgRect) {
		return true
	}
	return false
}

func (tp *windowIntervalTriggerPointViewer) mouseMoved(x, y float32) {
	if tp.scp.inStreamMode() {
		return
	}
	tp.windowTriggerPointViewer.mouseMoved(x, y)
	if tp.windowTriggerPointViewer.mouseAt || tp.windowTriggerPointViewer.uhMouseAt || tp.windowTriggerPointViewer.lMouseAt || tp.windowTriggerPointViewer.lhMouseAt {
		return
	}
	prev := tp.mouseAtIntervalPoint(float32(tp.lowerHImgRect.Min.X+tp.lowerHImgRect.Dx()/2), y) || tp.mouseAtIntervalPoint(float32(tp.upperHImgRect.Min.X+tp.upperHImgRect.Dx()/2), y)
	
	if tp.mouseAtIntervalPoint(x, y) {
		if !prev {
			tp.enableRefresh()
			if tp.raster() != nil {
				tp.raster().Refresh()
			}
		}
	} else {
		if prev {
			tp.enableRefresh()
			if tp.raster() != nil {
				tp.raster().Refresh()
			}
		}
	}
}

func (tp *windowIntervalTriggerPointViewer) mouseDown(button desktop.MouseButton, modifier fyne.KeyModifier, x, y float32) {
	if tp.scp.inStreamMode() {
		return
	}
	tp.windowTriggerPointViewer.mouseDown(button, modifier, x, y)
	if !tp.selected && !tp.uhSelected && !tp.lSelected && !tp.lhSelected {
		p := image.Point{X: int(math.Round(float64(x))), Y: int(math.Round(float64(y)))}
		tp.lowerSelected = p.In(tp.lowerHImgRect)
		tp.upperSelected = p.In(tp.upperHImgRect)
	}
}

func (tp *windowIntervalTriggerPointViewer) mouseUp(button desktop.MouseButton, modifier fyne.KeyModifier, x, y float32) {
	if tp.scp.inStreamMode() {
		return
	}
	tp.windowTriggerPointViewer.mouseUp(button, modifier, x, y)
	prev := tp.lowerSelected || tp.upperSelected
	tp.lowerSelected = false
	tp.upperSelected = false
	if prev {
		tp.enableRefresh()
		if tp.raster() != nil {
			tp.raster().Refresh()
		}
	}
}

func (tp *windowIntervalTriggerPointViewer) dragged(dx, dy, x, y float32) {
	if tp.scp.inStreamMode() {
		return
	}
	if tp.scp.triggerSource < 0 || int(tp.scp.triggerSource) >= len(tp.scp.Settings.Channels) {
		return
	}
	tp.windowTriggerPointViewer.dragged(dx, dy, x, y)

	if !tp.lowerSelected && !tp.upperSelected {
		return
	}

	channel := &tp.scp.Settings.Channels[tp.scp.triggerSource]
	bounds := tp.signalScreen().Bounds()
	w := float64(bounds.Dx() - 1)

	if w <= 0 || tp.maxScreenTime() <= 0 {
		return
	}

	triggerX, _ := tp.timeMv2xy(channel.Trigger.Mv)
	timeOffset := (float64(x-triggerX) / w) * tp.maxScreenTime()
	if timeOffset < 0 {
		timeOffset = -timeOffset
	}

	minTime, maxTime := tp.scp.getScreenTimeLimits()
	if timeOffset < minTime {
		timeOffset = minTime
	}
	if timeOffset > maxTime {
		timeOffset = maxTime
	}

	pwType := channel.Trigger.IntervalType
	isSingle := intervalSingleModeTypes[pwType]

	if isSingle {
		if pwType == genericps.PwTypeLessThan {
			channel.Trigger.IntervalTimeUpper = timeOffset
			channel.Trigger.IntervalTimeLower = timeOffset
		} else {
			channel.Trigger.IntervalTimeUpper = timeOffset
			channel.Trigger.IntervalTimeLower = timeOffset
		}
	} else {
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
	}

	tp.scp.triggerSettingMsg.IntervalTimeLower = channel.Trigger.IntervalTimeLower
	tp.scp.triggerSettingMsg.IntervalTimeUpper = channel.Trigger.IntervalTimeUpper

	t := tp.scp.triggerSettingMsg
	t.Done = make(chan struct{}, 1)
	go func() {
		tp.scp.psControl.SetTriggerCh <- &t
		<-t.Done
	}()

	tp.enableRefresh()
	if tp.raster() != nil {
		tp.raster().Refresh()
	}
}

func (tp *windowIntervalTriggerPointViewer) draw() {
	if tp.scp.controlTab.SelectedIndex() == dftTabIndex || tp.scp.inStreamMode() {
		return
	}
	tp.windowTriggerPointViewer.draw()

	if tp.scp.triggerSource != dontCare {
		channel := &tp.scp.Settings.Channels[tp.scp.triggerSource]
		if !channel.TriggerSource {
			return
		}
		x, yUpper := tp.timeMv2xy(channel.Trigger.Mv)
		_, yLower := tp.timeMv2xy(channel.Trigger.LowerMv)
		y := (yUpper + yLower) / 2
		bounds := tp.signalScreen().Bounds()

		w := float64(bounds.Dx() - 1)

		if w <= 0 || tp.maxScreenTime() <= 0 {
			return
		}

		pwType := channel.Trigger.IntervalType
		isSingle := intervalSingleModeTypes[pwType]

		halfRectSize := float32(triggerPointR * 2)
		rectSize2 := 2 * halfRectSize
		
		if isSingle {
			var singleTime float64
			if pwType == genericps.PwTypeLessThan {
				singleTime = channel.Trigger.IntervalTimeUpper
			} else {
				singleTime = channel.Trigger.IntervalTimeLower
			}
			singleDx := float32((singleTime / tp.maxScreenTime()) * w)
			xSingle := x - singleDx

			tp.lowerHImgRect = image.Rect(
				int(math.Round(float64(xSingle-rectSize2))),
				int(math.Round(float64(y-rectSize2))),
				int(math.Round(float64(xSingle+rectSize2))),
				int(math.Round(float64(y+rectSize2))))
			tp.upperHImgRect = image.Rect(0, 0, 0, 0)

			colSingle := theme.ForegroundColor()
			if tp.lowerSelected || tp.mouseAtIntervalPoint(xSingle, y) {
				colSingle = theme.SelectionColor()
			}

			drawLine(tp.signalScreen(), x, y, xSingle, y, colSingle)
			if pwType == genericps.PwTypeLessThan {
				drawLine(tp.signalScreen(), xSingle-halfRectSize, y-halfRectSize, xSingle-halfRectSize, y+halfRectSize, colSingle)
				drawLine(tp.signalScreen(), xSingle-halfRectSize, y-halfRectSize, xSingle, y, colSingle)
				drawLine(tp.signalScreen(), xSingle-halfRectSize, y+halfRectSize, xSingle, y, colSingle)
			} else {
				drawLine(tp.signalScreen(), xSingle+halfRectSize, y-halfRectSize, xSingle+halfRectSize, y+halfRectSize, colSingle)
				drawLine(tp.signalScreen(), xSingle+halfRectSize, y-halfRectSize, xSingle, y, colSingle)
				drawLine(tp.signalScreen(), xSingle+halfRectSize, y+halfRectSize, xSingle, y, colSingle)
			}

			if tp.scp.intervalTimeSingleDisp != nil {
				unit := getBaseTimeUnit(tp.scp.Settings.Time.Unit)
				multiplier := getIntervalUnitMultiplier(unit)
				val := int(math.Round(singleTime / multiplier))
				if tp.scp.intervalTimeSingleDisp.Value != val {
					tp.scp.intervalTimeSingleDisp.SilentSetValue(val)
					tp.scp.intervalTimeSingleDisp.Refresh()
				}
			}
		} else {
			lowerDx := float32((channel.Trigger.IntervalTimeLower / tp.maxScreenTime()) * w)
			upperDx := float32((channel.Trigger.IntervalTimeUpper / tp.maxScreenTime()) * w)

			xLower := x - lowerDx
			xUpper := x - upperDx

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
			if tp.lowerSelected || tp.mouseAtIntervalPoint(xLower, y) {
				colLower = theme.SelectionColor()
			}

			colUpper := theme.ForegroundColor()
			if tp.upperSelected || tp.mouseAtIntervalPoint(xUpper, y) {
				colUpper = theme.SelectionColor()
			}

			drawLine(tp.signalScreen(), x, y, xLower, y, colLower)
			drawLine(tp.signalScreen(), x, y, xUpper, y, colUpper)

			switch pwType {
			case genericps.PwTypeInRange:
				drawLine(tp.signalScreen(), xLower+halfRectSize, y-halfRectSize, xLower+halfRectSize, y+halfRectSize, colLower)
				drawLine(tp.signalScreen(), xLower+halfRectSize, y-halfRectSize, xLower, y, colLower)
				drawLine(tp.signalScreen(), xLower+halfRectSize, y+halfRectSize, xLower, y, colLower)
				
				drawLine(tp.signalScreen(), xUpper-halfRectSize, y-halfRectSize, xUpper-halfRectSize, y+halfRectSize, colUpper)
				drawLine(tp.signalScreen(), xUpper-halfRectSize, y-halfRectSize, xUpper, y, colUpper)
				drawLine(tp.signalScreen(), xUpper-halfRectSize, y+halfRectSize, xUpper, y, colUpper)
			case genericps.PwTypeOutOfRange:
				drawLine(tp.signalScreen(), xLower-halfRectSize, y-halfRectSize, xLower-halfRectSize, y+halfRectSize, colLower)
				drawLine(tp.signalScreen(), xLower-halfRectSize, y-halfRectSize, xLower, y, colLower)
				drawLine(tp.signalScreen(), xLower-halfRectSize, y+halfRectSize, xLower, y, colLower)
				
				drawLine(tp.signalScreen(), xUpper+halfRectSize, y-halfRectSize, xUpper+halfRectSize, y+halfRectSize, colUpper)
				drawLine(tp.signalScreen(), xUpper+halfRectSize, y-halfRectSize, xUpper, y, colUpper)
				drawLine(tp.signalScreen(), xUpper+halfRectSize, y+halfRectSize, xUpper, y, colUpper)
			}

			if tp.scp.intervalTimeLowerDisp != nil {
				unit := getBaseTimeUnit(tp.scp.Settings.Time.Unit)
				multiplier := getIntervalUnitMultiplier(unit)
				val := int(math.Round(channel.Trigger.IntervalTimeLower / multiplier))
				if tp.scp.intervalTimeLowerDisp.Value != val {
					tp.scp.intervalTimeLowerDisp.SilentSetValue(val)
					tp.scp.intervalTimeLowerDisp.Refresh()
				}
			}

			if tp.scp.intervalTimeUpperDisp != nil {
				unit := getBaseTimeUnit(tp.scp.Settings.Time.Unit)
				multiplier := getIntervalUnitMultiplier(unit)
				val := int(math.Round(channel.Trigger.IntervalTimeUpper / multiplier))
				if tp.scp.intervalTimeUpperDisp.Value != val {
					tp.scp.intervalTimeUpperDisp.SilentSetValue(val)
					tp.scp.intervalTimeUpperDisp.Refresh()
				}
			}
		}
	}
}

func (tp *windowIntervalTriggerPointViewer) scrolled(delta, x, y float32) {
	if tp.scp.inStreamMode() {
		return
	}
	tp.windowTriggerPointViewer.scrolled(delta, x, y)
}
