package gui

import (
	"fynescope/genericps"
	"image"
	"log/slog"
	"math"

	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
)

type (
	advTriggerPointViewer struct {
		triggerPointViewer
		uhImgRect  image.Rectangle
		uhSelected bool
	}
)

var (
	_ mouser     = (*advTriggerPointViewer)(nil)
	_ dragger    = (*advTriggerPointViewer)(nil)
	_ scroller   = (*advTriggerPointViewer)(nil)
	_ drawer     = (*advTriggerPointViewer)(nil)
	_ cursorable = (*advTriggerPointViewer)(nil)
)

func (tp *advTriggerPointViewer) cursor(x, y float32) (desktop.Cursor, bool) {
	if tp.scp.inStreamMode() {
		return desktop.DefaultCursor, false
	}
	cp, ok := tp.triggerPointViewer.cursor(x, y)
	if ok {
		return cp, ok
	}
	if tp.mouseAtHysteresisPoint(x, y) {
		return desktop.PointerCursor, true
	}
	return desktop.DefaultCursor, false
}

func (tp *advTriggerPointViewer) mouseAtHysteresisPoint(x, y float32) bool {
	p := image.Point{X: int(math.Round(float64(x))), Y: int(math.Round(float64(y)))}
	if p.In(tp.uhImgRect) {
		return true
	}
	return false
}

func (tp *advTriggerPointViewer) mouseMoved(x, y float32) {
	if tp.scp.inStreamMode() {
		return
	}
	tp.triggerPointViewer.mouseMoved(x, y)
	if tp.triggerPointViewer.mouseAt {
		tp.mouseAt = false
		return
	}
	prev := tp.mouseAt
	if tp.mouseAtHysteresisPoint(x, y) {
		tp.mouseAt = true
	} else {
		tp.mouseAt = false
	}
	if prev != tp.mouseAt {
		tp.enableRefresh()
		tp.scp.ftRaster.Refresh()
	}
}

func (tp *advTriggerPointViewer) mouseDown(button desktop.MouseButton, x, y float32) {
	if tp.scp.inStreamMode() {
		return
	}
	tp.triggerPointViewer.mouseDown(button, x, y)
	if !tp.selected {
		tp.uhSelected = tp.mouseAtHysteresisPoint(x, y)
	}
}

func (tp *advTriggerPointViewer) mouseUp(button desktop.MouseButton, x, y float32) {
	if tp.scp.inStreamMode() {
		return
	}
	tp.triggerPointViewer.mouseUp(button, x, y)
	prev := tp.uhSelected
	tp.uhSelected = false
	if prev {
		if !tp.mouseAtHysteresisPoint(x, y) {
			tp.mouseAt = false
		}
		tp.enableRefresh()
		tp.scp.ftRaster.Refresh()
	}
}
func (scp *ScpDesc) SetTriggerUpperHysteresis(mv int32) {
	if scp.triggerSettingMsg.UpperHysteresis != mv {
		scp.triggerSettingMsg.UpperHysteresis = mv
		scp.triggerSettingMsg.HysteresisADC = uint16(scp.mvToUAdc(mv, scp.Settings.Channels[scp.triggerSource].VRange))
		scp.psControl.SetTriggerCh <- &scp.triggerSettingMsg
		<-scp.triggerSettingMsg.Done
	}
}

func (tp *advTriggerPointViewer) setHysteresisDispOffset(dyh float32) {
	bounds := tp.scp.ftScopeSignalScreen.Bounds()
	h := float64(bounds.Dy())
	if tp.scp.triggerSource < 0 || int(tp.scp.triggerSource) >= len(tp.scp.Settings.Channels) {
		slog.Error("setHysteresisDispOffset index error", "tp.scp.triggerSource", tp.scp.triggerSource)
		return
	}
	channel := &tp.scp.Settings.Channels[tp.scp.triggerSource]
	yScale := 2 * genericps.RangeValuesMv[channel.VRange] / h
	if yScale < 1 {
		yScale = 1
	}
	d := int32(math.Round(yScale * float64(dyh)))
	if d > 0 || channel.Trigger.Hysteresis > 0 {
		channel.Trigger.Hysteresis += d
	}
	tp.scp.SetTriggerUpperHysteresis(channel.Trigger.Hysteresis)
	tp.enableRefresh()
	tp.scp.ftRaster.Refresh()
}

func (tp *advTriggerPointViewer) dragged(dx, dy, x, y float32) {
	if tp.scp.inStreamMode() {
		return
	}
	tp.triggerPointViewer.dragged(dx, dy, x, y) // call base class method
	if !tp.uhSelected {                         // mouse down/up set it
		return // 								   cursor is somewhere else
	}
	channel := &tp.scp.Settings.Channels[tp.scp.triggerSource]
	newH := int32(math.Round(tp.y2mv(float64(y))))
	switch {
	case channel.Trigger.TriggerDirection == genericps.TriggerRaising:
		if newH <= channel.Trigger.Mv {
			channel.Trigger.Hysteresis = channel.Trigger.Mv - newH
		}
	case channel.Trigger.TriggerDirection == genericps.TriggerFalling:
		if newH >= channel.Trigger.Mv {
			channel.Trigger.Hysteresis = -channel.Trigger.Mv + newH
		}
	default:
		slog.Error("advTrigger", "TriggerDirection", channel.Trigger.TriggerDirection)
	}
	tp.scp.SetTriggerUpperHysteresis(channel.Trigger.Hysteresis)
	tp.enableRefresh()
	tp.scp.ftRaster.Refresh()
}

func (tp *advTriggerPointViewer) scrolled(delta, x, y float32) {
	if tp.scp.inStreamMode() {
		return
	}
	tp.mouseDown(desktop.MouseButtonTertiary, x, y)
	if tp.selected {
		switch {
		case delta > 0:
			tp.setHysteresisDispOffset(1)
		case delta < 0:
			tp.setHysteresisDispOffset(-1)
		default:
			return
		}
		tp.enableRefresh()
		tp.scp.ftRaster.Refresh()
	}
	tp.selected = false
	tp.uhSelected = false
}

func (tp *advTriggerPointViewer) draw() {
	if tp.scp.controlTab.SelectedIndex() == dftTabIndex || tp.scp.inStreamMode() {
		return
	}
	if tp.scp.triggerSource != dontCare {
		channel := &tp.scp.Settings.Channels[tp.scp.triggerSource]
		x, y := tp.timeMv2xy(channel.Trigger.Mv)
		bound := tp.scp.ftScopeSignalScreen.Bounds()
		maxY := float32(bound.Max.Y)
		minY := float32(bound.Min.Y)
		switch {
		case y > maxY:
			y = maxY
		case y < minY:
			y = minY
		}
		halfRectSize := float32(triggerPointR * 2)
		tp.imgRect = image.Rect(int(math.Round(float64(x-halfRectSize))),
			int(math.Round(float64(y-halfRectSize))),
			int(math.Round(float64(x+halfRectSize))),
			int(math.Round(float64(y+halfRectSize))))
		var yh float32
		_, yh = tp.timeMv2xy(channel.Trigger.Mv - channel.Trigger.Hysteresis)
		if channel.Trigger.TriggerDirection == genericps.TriggerFalling {
			_, yh = tp.timeMv2xy(channel.Trigger.Mv + channel.Trigger.Hysteresis)
		}
		switch {
		case yh > maxY:
			yh = maxY
		case yh < minY:
			yh = minY
		}
		rectSize2 := 2 * halfRectSize
		tp.uhImgRect = image.Rect(int(math.Round(float64(x-rectSize2))),
			int(math.Round(float64( /*y-*/ yh-rectSize2))),
			int(math.Round(float64(x+rectSize2))),
			int(math.Round(float64( /*y+*/ rectSize2+yh))))
		col := theme.ForegroundColor()
		if tp.selected || tp.triggerPointViewer.mouseAt {
			col = theme.SelectionColor()
		}
		drawCircle(tp.scp.ftScopeSignalScreen, x, y, triggerPointR, col)
		col = theme.ForegroundColor()
		if tp.uhSelected || tp.mouseAt {
			col = theme.SelectionColor()
		}
		drawLine(tp.scp.ftScopeSignalScreen, x, y, x, yh, col)
		drawLine(tp.scp.ftScopeSignalScreen, x-halfRectSize, yh, x+halfRectSize, yh, col)
		if tp.scp.triggerThresholdDisp.Value != int(channel.Trigger.Mv) {
			tp.scp.triggerThresholdDisp.SilentSetValue(int(channel.Trigger.Mv))
			tp.scp.triggerThresholdDisp.Refresh()
		}
		if tp.scp.triggerHysteresisDisp.Value != int(channel.Trigger.Hysteresis) {
			tp.scp.triggerHysteresisDisp.SilentSetValue(int(channel.Trigger.Hysteresis))
			tp.scp.triggerHysteresisDisp.Refresh()
		}
	}
}

func newAdvTriggerPointViewer(img rasterImage, scp *ScpDesc) *advTriggerPointViewer {
	imgRect := image.Rect(int(math.Round(-triggerPointR)),
		int(math.Round(-triggerPointR)),
		int(math.Round(triggerPointR)),
		int(math.Round(triggerPointR)))
	uhImgRect := image.Rect(int(math.Round(-triggerPointR)),
		int(math.Round(triggerPointR)+100.0),
		int(math.Round(triggerPointR)),
		int(math.Round(triggerPointR+100.0)))
	tp := &advTriggerPointViewer{uhImgRect: uhImgRect, triggerPointViewer: triggerPointViewer{rasterPartition: rasterPartition{img: img, imgRect: imgRect}, scp: scp}}
	return tp
}
