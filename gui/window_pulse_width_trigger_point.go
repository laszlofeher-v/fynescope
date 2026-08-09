package gui

import (
	"fynescope/genericps"
	"image"
	"math"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
)

// windowPulseWidthTriggerPointViewer combines:
//   - intervalTriggerPointViewer: the upper trigger point (circle + hysteresis handle)
//     plus the horizontal pulse-width time handles (lowerHImgRect / upperHImgRect).
//   - A dedicated lower voltage-threshold trigger point (circle + hysteresis handle)
//     that defines the bottom boundary of the window, exactly as in windowTriggerPointViewer.
type windowPulseWidthTriggerPointViewer struct {
	intervalTriggerPointViewer        // upper circle + hysteresis + PW time handles
	lImgRect   image.Rectangle       // lower threshold circle hit-test rect
	lSelected  bool
	lMouseAt   bool
	lhImgRect  image.Rectangle       // lower hysteresis handle hit-test rect
	lhSelected bool
	lhMouseAt  bool
}

var (
	_ mouser     = (*windowPulseWidthTriggerPointViewer)(nil)
	_ dragger    = (*windowPulseWidthTriggerPointViewer)(nil)
	_ scroller   = (*windowPulseWidthTriggerPointViewer)(nil)
	_ drawer     = (*windowPulseWidthTriggerPointViewer)(nil)
	_ cursorable = (*windowPulseWidthTriggerPointViewer)(nil)
)

func newWindowPulseWidthTriggerPointViewer(img rasterImage, scp *ScpDesc, isTimeZoom bool) *windowPulseWidthTriggerPointViewer {
	tp := &windowPulseWidthTriggerPointViewer{
		intervalTriggerPointViewer: *newIntervalTriggerPointViewer(img, scp, isTimeZoom),
	}
	return tp
}

// mouseAtLowerPoint checks if the provided (x, y) coordinates fall within the
// hit-test rectangle of the lower voltage threshold circle.
func (tp *windowPulseWidthTriggerPointViewer) mouseAtLowerPoint(x, y float32) bool {
	p := image.Point{X: int(math.Round(float64(x))), Y: int(math.Round(float64(y)))}
	return p.In(tp.lImgRect)
}

// mouseAtLowerHysteresisPoint checks if the provided (x, y) coordinates fall within
// the hit-test rectangle of the lower hysteresis handle.
func (tp *windowPulseWidthTriggerPointViewer) mouseAtLowerHysteresisPoint(x, y float32) bool {
	p := image.Point{X: int(math.Round(float64(x))), Y: int(math.Round(float64(y)))}
	return p.In(tp.lhImgRect)
}

// cursor determines the appropriate mouse cursor to display based on the current
// pointer position. It returns a pointer cursor if hovering over an interactive element.
func (tp *windowPulseWidthTriggerPointViewer) cursor(x, y float32) (desktop.Cursor, bool) {
	if tp.scp.inStreamMode() {
		return desktop.DefaultCursor, false
	}
	cp, ok := tp.intervalTriggerPointViewer.cursor(x, y)
	if ok {
		return cp, ok
	}
	if tp.mouseAtLowerPoint(x, y) || tp.mouseAtLowerHysteresisPoint(x, y) {
		return desktop.PointerCursor, true
	}
	return desktop.DefaultCursor, false
}

// mouseMoved tracks the pointer's movement to update hover states for the lower
// threshold and hysteresis handles, triggering a UI refresh when the state changes.
func (tp *windowPulseWidthTriggerPointViewer) mouseMoved(x, y float32) {
	if tp.scp.inStreamMode() {
		return
	}
	tp.intervalTriggerPointViewer.mouseMoved(x, y)

	prevL := tp.lMouseAt
	prevLh := tp.lhMouseAt
	tp.lMouseAt = tp.mouseAtLowerPoint(x, y)
	tp.lhMouseAt = tp.mouseAtLowerHysteresisPoint(x, y)

	if prevL != tp.lMouseAt || prevLh != tp.lhMouseAt {
		tp.enableRefresh()
		if tp.raster() != nil {
			tp.raster().Refresh()
		}
	}
}

// mouseDown handles click events. It registers selection on the lower threshold or
// lower hysteresis handle if the click is within their bounds and no upper handles are claimed.
func (tp *windowPulseWidthTriggerPointViewer) mouseDown(button desktop.MouseButton, modifier fyne.KeyModifier, x, y float32) {
	if tp.scp.inStreamMode() {
		return
	}
	tp.intervalTriggerPointViewer.mouseDown(button, modifier, x, y)
	// Only pick up lower handles if no upper/PW handle was claimed
	if !tp.selected && !tp.uhSelected && !tp.lowerSelected && !tp.upperSelected {
		tp.lSelected = tp.mouseAtLowerPoint(x, y)
		tp.lhSelected = tp.mouseAtLowerHysteresisPoint(x, y)
	}
}

// mouseUp handles the release of a click event. It clears selection states for
// the lower handles and triggers a settings save and UI refresh if any were selected.
func (tp *windowPulseWidthTriggerPointViewer) mouseUp(button desktop.MouseButton, modifier fyne.KeyModifier, x, y float32) {
	if tp.scp.inStreamMode() {
		return
	}
	tp.intervalTriggerPointViewer.mouseUp(button, modifier, x, y)
	refresh := false
	saveLower := false
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
		if tp.raster() != nil {
			tp.raster().Refresh()
		}
	}
}

// setLowerDispOffset translates vertical dragging of the lower threshold circle
// into a voltage change (LowerMv), updates the oscilloscope settings, and sends
// the new configuration to the device.
func (tp *windowPulseWidthTriggerPointViewer) setLowerDispOffset(dx, x, y float32) {
	bounds := tp.signalScreen().Bounds()
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
	if newMv > channel.Trigger.Mv-genericps.MinThresholdDiff {
		newMv = channel.Trigger.Mv - genericps.MinThresholdDiff
	}
	channel.Trigger.LowerMv = newMv
	tp.scp.triggerSettingMsg.LowerMv = newMv
	tp.scp.triggerSettingMsg.LowerTriggerADC = int16(tp.scp.mvToAdc(newMv, channel.VRange))
	t := tp.scp.triggerSettingMsg
	t.Done = make(chan struct{}, 1)
	go func() {
		tp.scp.psControl.SetTriggerCh <- &t
		<-t.Done
	}()
	tp.scp.clearAllFtPersistentLayers()
	tp.scp.clearAllDftPersistentLayers()
	tp.enableRefresh()
	if tp.raster() != nil {
		tp.raster().Refresh()
	}
}

// setLowerHysteresisDispOffset adjusts the lower threshold hysteresis in response
// to a scroll event. It calculates a step size based on the current voltage range
// and pushes the updated hysteresis to the device.
func (tp *windowPulseWidthTriggerPointViewer) setLowerHysteresisDispOffset(dyh float32) {
	bounds := tp.signalScreen().Bounds()
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
		t := tp.scp.triggerSettingMsg
		t.Done = make(chan struct{}, 1)
		go func() {
			tp.scp.psControl.SetTriggerCh <- &t
			<-t.Done
		}()
	}
	tp.enableRefresh()
	if tp.raster() != nil {
		tp.raster().Refresh()
	}
}

// dragged handles pointer drag events. It delegates to the appropriate update
// function for the lower threshold/hysteresis if they are selected; otherwise,
// it falls back to the embedded interval viewer for upper/time handle dragging.
func (tp *windowPulseWidthTriggerPointViewer) dragged(dx, dy, x, y float32) {
	if tp.scp.inStreamMode() {
		return
	}
	if tp.scp.triggerSource < 0 || int(tp.scp.triggerSource) >= len(tp.scp.Settings.Channels) {
		return
	}

	// Lower threshold drag
	if tp.lSelected {
		tp.setLowerDispOffset(dx, x, y)
		return
	}

	// Lower hysteresis drag
	if tp.lhSelected {
		channel := &tp.scp.Settings.Channels[tp.scp.triggerSource]
		newH := int32(math.Round(tp.y2mv(float64(y))))
		switch channel.Trigger.TriggerDirection {
		case genericps.TriggerRising, genericps.TriggerInside, genericps.TriggerOutside, genericps.TriggerEnter, genericps.TriggerEnterOrExit:
			if newH <= channel.Trigger.LowerMv {
				channel.Trigger.LowerHysteresis = channel.Trigger.LowerMv - newH
			}
		case genericps.TriggerFalling, genericps.TriggerExit:
			if newH >= channel.Trigger.LowerMv {
				channel.Trigger.LowerHysteresis = -channel.Trigger.LowerMv + newH
			}
		}
		if tp.scp.triggerSettingMsg.LowerHysteresis != channel.Trigger.LowerHysteresis {
			tp.scp.triggerSettingMsg.LowerHysteresis = channel.Trigger.LowerHysteresis
			tp.scp.triggerSettingMsg.LowerHysteresisADC = uint16(tp.scp.mvToUAdc(channel.Trigger.LowerHysteresis, channel.VRange))
			t := tp.scp.triggerSettingMsg
			t.Done = make(chan struct{}, 1)
			go func() {
				tp.scp.psControl.SetTriggerCh <- &t
				<-t.Done
			}()
		}
		tp.enableRefresh()
		if tp.raster() != nil {
			tp.raster().Refresh()
		}
		return
	}

	// Delegate upper trigger point and PW time handle drags to interval viewer
	tp.intervalTriggerPointViewer.dragged(dx, dy, x, y)
}

// scrolled handles mouse scroll events. Scrolling while hovering over the lower
// threshold or lower hysteresis elements adjusts the hysteresis value; otherwise,
// it delegates the event to the interval viewer.
func (tp *windowPulseWidthTriggerPointViewer) scrolled(delta, x, y float32) {
	if tp.scp.inStreamMode() {
		return
	}
	switch {
	case tp.lMouseAt || tp.lSelected:
		if delta == 0 {
			return
		}
		if delta > 0 {
			tp.setLowerHysteresisDispOffset(1)
		} else {
			tp.setLowerHysteresisDispOffset(-1)
		}
	case tp.lhMouseAt || tp.lhSelected:
		if delta == 0 {
			return
		}
		if delta > 0 {
			tp.setLowerHysteresisDispOffset(1)
		} else {
			tp.setLowerHysteresisDispOffset(-1)
		}
	default:
		tp.intervalTriggerPointViewer.scrolled(delta, x, y)
	}
}

// draw renders the window pulse width trigger UI onto the signal screen.
// It relies on the embedded intervalTriggerPointViewer to draw the upper
// threshold, upper hysteresis, and time handles, then calculates hit-test
// rectangles and draws the lower threshold circle and hysteresis indicator.
func (tp *windowPulseWidthTriggerPointViewer) draw() {
	if tp.scp.controlTab.SelectedIndex() == dftTabIndex || tp.scp.inStreamMode() {
		return
	}

	// Draw upper trigger circle + upper hysteresis handle + PW time handles
	tp.intervalTriggerPointViewer.draw()

	if tp.scp.triggerSource == dontCare {
		return
	}
	channel := &tp.scp.Settings.Channels[tp.scp.triggerSource]
	if !channel.TriggerSource {
		return
	}

	scrnBound := tp.signalScreen().Bounds()
	maxY := float32(scrnBound.Max.Y)
	minY := float32(scrnBound.Min.Y)

	halfRectSize := float32(triggerPointR * 2)
	rectSize2 := 2 * halfRectSize

	// Lower threshold circle
	lx, ly := tp.timeMv2xy(channel.Trigger.LowerMv)
	if ly > maxY {
		ly = maxY
	}
	if ly < minY {
		ly = minY
	}
	tp.lImgRect = image.Rect(
		int(math.Round(float64(lx-halfRectSize))),
		int(math.Round(float64(ly-halfRectSize))),
		int(math.Round(float64(lx+halfRectSize))),
		int(math.Round(float64(ly+halfRectSize))))

	// Lower hysteresis handle position
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
	tp.lhImgRect = image.Rect(
		int(math.Round(float64(lx-rectSize2))),
		int(math.Round(float64(lyh-rectSize2))),
		int(math.Round(float64(lx+rectSize2))),
		int(math.Round(float64(rectSize2+lyh))))

	// Draw lower threshold circle
	col := theme.ForegroundColor()
	if tp.lSelected || tp.lMouseAt {
		col = theme.SelectionColor()
	}
	drawCircle(tp.signalScreen(), lx, ly, triggerPointR, col)

	// Draw lower hysteresis handle
	col = theme.ForegroundColor()
	if tp.lhSelected || tp.lhMouseAt {
		col = theme.SelectionColor()
	}
	drawLine(tp.signalScreen(), lx, ly, lx, lyh, col)
	drawLine(tp.signalScreen(), lx-halfRectSize, lyh, lx+halfRectSize, lyh, col)

	// Update lower threshold and hysteresis display widgets
	if tp.scp.triggerLowerThresholdDisp != nil && tp.scp.triggerLowerThresholdDisp.Value != int(channel.Trigger.LowerMv) {
		tp.scp.triggerLowerThresholdDisp.SilentSetValue(int(channel.Trigger.LowerMv))
		tp.scp.triggerLowerThresholdDisp.Refresh()
	}
	if tp.scp.triggerLowerHysteresisDisp != nil && tp.scp.triggerLowerHysteresisDisp.Value != int(channel.Trigger.LowerHysteresis) {
		tp.scp.triggerLowerHysteresisDisp.SilentSetValue(int(channel.Trigger.LowerHysteresis))
		tp.scp.triggerLowerHysteresisDisp.Refresh()
	}
}
