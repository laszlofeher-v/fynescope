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
	windowTriggerPointViewer struct {
		triggerPointViewer
		uhImgRect  image.Rectangle
		uhSelected bool
		uhMouseAt  bool
		lImgRect   image.Rectangle
		lSelected  bool
		lMouseAt   bool
		lhImgRect  image.Rectangle
		lhSelected bool
		lhMouseAt  bool
	}
)

var (
	_ mouser     = (*windowTriggerPointViewer)(nil)
	_ dragger    = (*windowTriggerPointViewer)(nil)
	_ scroller   = (*windowTriggerPointViewer)(nil)
	_ drawer     = (*windowTriggerPointViewer)(nil)
	_ cursorable = (*windowTriggerPointViewer)(nil)
)

func (tp *windowTriggerPointViewer) cursor(x, y float32) (desktop.Cursor, bool) {
	if tp.scp.inStreamMode() {
		return desktop.DefaultCursor, false
	}
	cp, ok := tp.triggerPointViewer.cursor(x, y)
	if ok {
		return cp, ok
	}
	if tp.mouseAtUpperHysteresisPoint(x, y) || tp.mouseAtLowerHysteresisPoint(x, y) || tp.mouseAtLowerPoint(x, y) {
		return desktop.PointerCursor, true
	}
	return desktop.DefaultCursor, false
}

func (tp *windowTriggerPointViewer) mouseAtUpperHysteresisPoint(x, y float32) bool {
	p := image.Point{X: int(math.Round(float64(x))), Y: int(math.Round(float64(y)))}
	return p.In(tp.uhImgRect)
}

func (tp *windowTriggerPointViewer) mouseAtLowerPoint(x, y float32) bool {
	p := image.Point{X: int(math.Round(float64(x))), Y: int(math.Round(float64(y)))}
	return p.In(tp.lImgRect)
}

func (tp *windowTriggerPointViewer) mouseAtLowerHysteresisPoint(x, y float32) bool {
	p := image.Point{X: int(math.Round(float64(x))), Y: int(math.Round(float64(y)))}
	return p.In(tp.lhImgRect)
}

func (tp *windowTriggerPointViewer) mouseMoved(x, y float32) {
	if tp.scp.inStreamMode() {
		return
	}
	prevMain := tp.mouseAt
	prevUh := tp.uhMouseAt
	prevL := tp.lMouseAt
	prevLh := tp.lhMouseAt

	tp.mouseAt = tp.mouseIn(x, y)
	tp.uhMouseAt = tp.mouseAtUpperHysteresisPoint(x, y)
	tp.lMouseAt = tp.mouseAtLowerPoint(x, y)
	tp.lhMouseAt = tp.mouseAtLowerHysteresisPoint(x, y)

	if prevUh != tp.uhMouseAt || prevL != tp.lMouseAt || prevLh != tp.lhMouseAt || prevMain != tp.mouseAt {
		tp.enableRefresh()
		tp.scp.ftRaster.Refresh()
	}
}

func (tp *windowTriggerPointViewer) mouseDown(button desktop.MouseButton, x, y float32) {
	if tp.scp.inStreamMode() {
		return
	}
	if tp.mouseAt {
		tp.selected = true
	} else if tp.uhMouseAt {
		tp.uhSelected = true
	} else if tp.lMouseAt {
		tp.lSelected = true
	} else if tp.lhMouseAt {
		tp.lhSelected = true
	}
}

func (tp *windowTriggerPointViewer) mouseUp(button desktop.MouseButton, x, y float32) {
	if tp.scp.inStreamMode() {
		return
	}
	refresh := false
	saveLower := false
	if tp.selected {
		tp.selected = false
		tp.mouseAt = tp.mouseIn(x, y)
		refresh = true
	}
	if tp.uhSelected {
		tp.uhSelected = false
		tp.uhMouseAt = tp.mouseAtUpperHysteresisPoint(x, y)
		refresh = true
	}
	if tp.lSelected {
		tp.lSelected = false
		tp.lMouseAt = tp.mouseAtLowerPoint(x, y)
		refresh = true
		saveLower = true
	}
	if tp.lhSelected {
		tp.lhSelected = false
		tp.lhMouseAt = tp.mouseAtLowerHysteresisPoint(x, y)
		refresh = true
		saveLower = true
	}
	if saveLower {
		setFlag(tp.scp.repartition)
		tp.scp.SaveSettings()
	}
	if refresh {
		tp.enableRefresh()
		tp.scp.ftRaster.Refresh()
	}
}

func (tp *windowTriggerPointViewer) setLowerDispOffset(dx, x, y float32) {
	bounds := tp.scp.ftScopeSignalScreen.Bounds()
	if int(x) < bounds.Min.X || int(x) > bounds.Max.X ||
		int(y) < bounds.Min.Y || int(y) > bounds.Max.Y {
		return
	}
	mv := tp.y2mv(float64(y))
	channel := &tp.scp.Settings.Channels[tp.scp.triggerSource]
	bound := float64(genericps.InputRanges[channel.VRange])
	if mv < -bound || mv > bound {
		return
	}
	tp.scp.addFtXOffset(float64(dx))
	tp.scp.setTriggerTime(tp.scp.Settings.Time.TriggerTimeOffset)
	newMv := int32(math.Round(float64(mv)))
	if newMv > channel.Trigger.Mv {
		newMv = channel.Trigger.Mv
	}
	channel.Trigger.LowerMv = newMv
	tp.scp.triggerSettingMsg.LowerMv = newMv
	tp.scp.triggerSettingMsg.LowerTriggerADC = int16(tp.scp.mvToAdc(newMv, channel.VRange))
	tp.scp.psControl.SetTriggerCh <- &tp.scp.triggerSettingMsg
	<-tp.scp.triggerSettingMsg.Done
	tp.scp.clearAllFtPersistentLayers()
	tp.scp.clearAllDftPersistentLayers()
	tp.enableRefresh()
	if tp.scp.ftRaster != nil {
		tp.scp.ftRaster.Refresh()
	}
}

func (tp *windowTriggerPointViewer) dragged(dx, dy, x, y float32) {
	if tp.scp.inStreamMode() {
		return
	}

	if tp.selected {
		tp.triggerPointViewer.dragged(dx, dy, x, y)
		return
	}

	if tp.lSelected {
		tp.setLowerDispOffset(dx, x, y)
		return
	}

	channel := &tp.scp.Settings.Channels[tp.scp.triggerSource]
	newH := int32(math.Round(tp.y2mv(float64(y))))

	if tp.uhSelected {
		switch channel.Trigger.TriggerDirection {
		case genericps.TriggerRaising, genericps.TriggerInside, genericps.TriggerOutside, genericps.TriggerEnter, genericps.TriggerEnterOrExit:
			if newH >= channel.Trigger.Mv {
				channel.Trigger.Hysteresis = newH - channel.Trigger.Mv
			}
		case genericps.TriggerFalling, genericps.TriggerExit:
			if newH <= channel.Trigger.Mv {
				channel.Trigger.Hysteresis = channel.Trigger.Mv - newH
			}
		default:
			slog.Error("windowTrigger", "TriggerDirection", channel.Trigger.TriggerDirection)
		}
		tp.scp.SetTriggerUpperHysteresis(channel.Trigger.Hysteresis)
		tp.enableRefresh()
		tp.scp.ftRaster.Refresh()
		return
	}

	if tp.lhSelected {
		channel := &tp.scp.Settings.Channels[tp.scp.triggerSource]
		newH := int32(math.Round(tp.y2mv(float64(y))))
		switch channel.Trigger.TriggerDirection {
		case genericps.TriggerRaising, genericps.TriggerInside, genericps.TriggerOutside, genericps.TriggerEnter, genericps.TriggerEnterOrExit:
			if newH <= channel.Trigger.LowerMv {
				channel.Trigger.LowerHysteresis = channel.Trigger.LowerMv - newH
			}
		case genericps.TriggerFalling, genericps.TriggerExit:
			if newH >= channel.Trigger.LowerMv {
				channel.Trigger.LowerHysteresis = -channel.Trigger.LowerMv + newH
			}
		default:
			slog.Error("windowTrigger", "TriggerDirection", channel.Trigger.TriggerDirection)
		}
		// Update without repartition (same as SetTriggerUpperHysteresis)
		if tp.scp.triggerSettingMsg.LowerHysteresis != channel.Trigger.LowerHysteresis {
			tp.scp.triggerSettingMsg.LowerHysteresis = channel.Trigger.LowerHysteresis
			tp.scp.triggerSettingMsg.LowerHysteresisADC = uint16(tp.scp.mvToUAdc(channel.Trigger.LowerHysteresis, channel.VRange))
			tp.scp.psControl.SetTriggerCh <- &tp.scp.triggerSettingMsg
			<-tp.scp.triggerSettingMsg.Done
		}
		tp.enableRefresh()
		tp.scp.ftRaster.Refresh()
		return
	}
}

func (tp *windowTriggerPointViewer) setUpperHysteresisDispOffset(dyh float32) {
	bounds := tp.scp.ftScopeSignalScreen.Bounds()
	h := float64(bounds.Dy())
	if tp.scp.triggerSource < 0 || int(tp.scp.triggerSource) >= len(tp.scp.Settings.Channels) {
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

func (tp *windowTriggerPointViewer) setLowerHysteresisDispOffset(dyh float32) {
	bounds := tp.scp.ftScopeSignalScreen.Bounds()
	h := float64(bounds.Dy())
	if tp.scp.triggerSource < 0 || int(tp.scp.triggerSource) >= len(tp.scp.Settings.Channels) {
		return
	}
	channel := &tp.scp.Settings.Channels[tp.scp.triggerSource]
	yScale := 2 * genericps.RangeValuesMv[channel.VRange] / h
	if yScale < 1 {
		yScale = 1
	}
	d := int32(math.Round(yScale * float64(dyh)))
	if d > 0 || channel.Trigger.LowerHysteresis > 0 {
		channel.Trigger.LowerHysteresis += d
	}
	if tp.scp.triggerSettingMsg.LowerHysteresis != channel.Trigger.LowerHysteresis {
		tp.scp.triggerSettingMsg.LowerHysteresis = channel.Trigger.LowerHysteresis
		tp.scp.triggerSettingMsg.LowerHysteresisADC = uint16(tp.scp.mvToUAdc(channel.Trigger.LowerHysteresis, channel.VRange))
		tp.scp.psControl.SetTriggerCh <- &tp.scp.triggerSettingMsg
		<-tp.scp.triggerSettingMsg.Done
	}
	tp.enableRefresh()
	tp.scp.ftRaster.Refresh()
}

func (tp *windowTriggerPointViewer) scrolled(delta, x, y float32) {
	if tp.scp.inStreamMode() {
		return
	}
	// Determine which handle to scroll based on where the mouse is
	switch {
	case tp.mouseAt || tp.selected:
		// Scroll over the upper trigger point: adjust upper hysteresis
		if delta == 0 {
			return
		}
		if delta > 0 {
			tp.setUpperHysteresisDispOffset(1)
		} else {
			tp.setUpperHysteresisDispOffset(-1)
		}
	case tp.uhMouseAt || tp.uhSelected:
		// Scroll over the upper hysteresis handle: adjust upper hysteresis
		if delta == 0 {
			return
		}
		if delta > 0 {
			tp.setUpperHysteresisDispOffset(1)
		} else {
			tp.setUpperHysteresisDispOffset(-1)
		}
	case tp.lMouseAt || tp.lSelected:
		// Scroll over the lower trigger point: adjust lower hysteresis
		if delta == 0 {
			return
		}
		if delta > 0 {
			tp.setLowerHysteresisDispOffset(1)
		} else {
			tp.setLowerHysteresisDispOffset(-1)
		}
	case tp.lhMouseAt || tp.lhSelected:
		// Scroll over the lower hysteresis handle: adjust lower hysteresis
		if delta == 0 {
			return
		}
		if delta > 0 {
			tp.setLowerHysteresisDispOffset(1)
		} else {
			tp.setLowerHysteresisDispOffset(-1)
		}
	}
}


func (tp *windowTriggerPointViewer) draw() {
	if tp.scp.controlTab.SelectedIndex() == dftTabIndex || tp.scp.inStreamMode() {
		return
	}
	if tp.scp.triggerSource != dontCare {
		channel := &tp.scp.Settings.Channels[tp.scp.triggerSource]
		bound := tp.scp.ftScopeSignalScreen.Bounds()
		maxY := float32(bound.Max.Y)
		minY := float32(bound.Min.Y)

		// Upper Point
		x, y := tp.timeMv2xy(channel.Trigger.Mv)
		if y > maxY {
			y = maxY
		}
		if y < minY {
			y = minY
		}

		halfRectSize := float32(triggerPointR * 2)
		tp.imgRect = image.Rect(int(math.Round(float64(x-halfRectSize))),
			int(math.Round(float64(y-halfRectSize))),
			int(math.Round(float64(x+halfRectSize))),
			int(math.Round(float64(y+halfRectSize))))

		var yh float32
		_, yh = tp.timeMv2xy(channel.Trigger.Mv + channel.Trigger.Hysteresis)
		if channel.Trigger.TriggerDirection == genericps.TriggerFalling || channel.Trigger.TriggerDirection == genericps.TriggerExit {
			_, yh = tp.timeMv2xy(channel.Trigger.Mv - channel.Trigger.Hysteresis)
		}
		if yh > maxY {
			yh = maxY
		}
		if yh < minY {
			yh = minY
		}

		rectSize2 := 2 * halfRectSize
		tp.uhImgRect = image.Rect(int(math.Round(float64(x-rectSize2))),
			int(math.Round(float64(yh-rectSize2))),
			int(math.Round(float64(x+rectSize2))),
			int(math.Round(float64(rectSize2+yh))))

		// Lower Point
		lx, ly := tp.timeMv2xy(channel.Trigger.LowerMv)
		if ly > maxY {
			ly = maxY
		}
		if ly < minY {
			ly = minY
		}

		tp.lImgRect = image.Rect(int(math.Round(float64(lx-halfRectSize))),
			int(math.Round(float64(ly-halfRectSize))),
			int(math.Round(float64(lx+halfRectSize))),
			int(math.Round(float64(ly+halfRectSize))))

		var lyh float32
		_, lyh = tp.timeMv2xy(channel.Trigger.LowerMv - channel.Trigger.LowerHysteresis)
		if channel.Trigger.TriggerDirection == genericps.TriggerFalling || channel.Trigger.TriggerDirection == genericps.TriggerExit {
			_, lyh = tp.timeMv2xy(channel.Trigger.LowerMv + channel.Trigger.LowerHysteresis)
		}
		if lyh > maxY {
			lyh = maxY
		}
		if lyh < minY {
			lyh = minY
		}

		tp.lhImgRect = image.Rect(int(math.Round(float64(lx-rectSize2))),
			int(math.Round(float64(lyh-rectSize2))),
			int(math.Round(float64(lx+rectSize2))),
			int(math.Round(float64(rectSize2+lyh))))

		// Draw Upper
		col := theme.ForegroundColor()
		if tp.selected || tp.mouseAt {
			col = theme.SelectionColor()
		}
		drawCircle(tp.scp.ftScopeSignalScreen, x, y, triggerPointR, col)

		col = theme.ForegroundColor()
		if tp.uhSelected || tp.uhMouseAt {
			col = theme.SelectionColor()
		}
		drawLine(tp.scp.ftScopeSignalScreen, x, y, x, yh, col)
		drawLine(tp.scp.ftScopeSignalScreen, x-halfRectSize, yh, x+halfRectSize, yh, col)

		// Draw Lower
		col = theme.ForegroundColor()
		if tp.lSelected || tp.lMouseAt {
			col = theme.SelectionColor()
		}
		drawCircle(tp.scp.ftScopeSignalScreen, lx, ly, triggerPointR, col)

		col = theme.ForegroundColor()
		if tp.lhSelected || tp.lhMouseAt {
			col = theme.SelectionColor()
		}
		drawLine(tp.scp.ftScopeSignalScreen, lx, ly, lx, lyh, col)
		drawLine(tp.scp.ftScopeSignalScreen, lx-halfRectSize, lyh, lx+halfRectSize, lyh, col)

		// Update Disp7s
		if tp.scp.triggerThresholdDisp.Value != int(channel.Trigger.Mv) {
			tp.scp.triggerThresholdDisp.SilentSetValue(int(channel.Trigger.Mv))
			tp.scp.triggerThresholdDisp.Refresh()
		}
		if tp.scp.triggerHysteresisDisp.Value != int(channel.Trigger.Hysteresis) {
			tp.scp.triggerHysteresisDisp.SilentSetValue(int(channel.Trigger.Hysteresis))
			tp.scp.triggerHysteresisDisp.Refresh()
		}
		if tp.scp.triggerLowerThresholdDisp != nil && tp.scp.triggerLowerThresholdDisp.Value != int(channel.Trigger.LowerMv) {
			tp.scp.triggerLowerThresholdDisp.SilentSetValue(int(channel.Trigger.LowerMv))
			tp.scp.triggerLowerThresholdDisp.Refresh()
		}
		if tp.scp.triggerLowerHysteresisDisp != nil && tp.scp.triggerLowerHysteresisDisp.Value != int(channel.Trigger.LowerHysteresis) {
			tp.scp.triggerLowerHysteresisDisp.SilentSetValue(int(channel.Trigger.LowerHysteresis))
			tp.scp.triggerLowerHysteresisDisp.Refresh()
		}
	}
}

func newWindowTriggerPointViewer(img rasterImage, scp *ScpDesc) *windowTriggerPointViewer {
	imgRect := image.Rect(int(math.Round(-triggerPointR)),
		int(math.Round(-triggerPointR)),
		int(math.Round(triggerPointR)),
		int(math.Round(triggerPointR)))
	// We init with huge rects, they will be refined in draw()
	tp := &windowTriggerPointViewer{triggerPointViewer: triggerPointViewer{rasterPartition: rasterPartition{img: img, imgRect: imgRect}, scp: scp}}
	return tp
}
