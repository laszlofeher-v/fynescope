package gui

import (
	"fynescope/control"
	"fynescope/genericps"
	"image"
	"math"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
)

type intervalTriggerPointViewer struct {
	advTriggerPointViewer
	lowerHImgRects []image.Rectangle
	upperHImgRects []image.Rectangle
	lowerSelected  bool
	upperSelected  bool
}

var (
	_ mouser     = (*intervalTriggerPointViewer)(nil)
	_ dragger    = (*intervalTriggerPointViewer)(nil)
	_ scroller   = (*intervalTriggerPointViewer)(nil)
	_ drawer     = (*intervalTriggerPointViewer)(nil)
	_ cursorable = (*intervalTriggerPointViewer)(nil)
)

func newIntervalTriggerPointViewer(img rasterImage, scp *ScpDesc, isTimeZoom bool) *intervalTriggerPointViewer {
	tp := &intervalTriggerPointViewer{
		advTriggerPointViewer: *newAdvTriggerPointViewer(img, scp, isTimeZoom),
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
	for _, rect := range tp.lowerHImgRects {
		if p.In(rect) {
			return true
		}
	}
	for _, rect := range tp.upperHImgRects {
		if p.In(rect) {
			return true
		}
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
		if tp.raster() != nil {
			tp.raster().Refresh()
		}
	}
}

func (tp *intervalTriggerPointViewer) mouseDown(button desktop.MouseButton, modifier fyne.KeyModifier, x, y float32) {
	if tp.scp.inStreamMode() {
		return
	}
	tp.advTriggerPointViewer.mouseDown(button, modifier, x, y)
	if !tp.selected && !tp.uhSelected {
		p := image.Point{X: int(math.Round(float64(x))), Y: int(math.Round(float64(y)))}
		tp.lowerSelected = false
		tp.upperSelected = false
		for _, rect := range tp.lowerHImgRects {
			if p.In(rect) {
				tp.lowerSelected = true
				break
			}
		}
		for _, rect := range tp.upperHImgRects {
			if p.In(rect) {
				tp.upperSelected = true
				break
			}
		}
	}
}

func (tp *intervalTriggerPointViewer) mouseUp(button desktop.MouseButton, modifier fyne.KeyModifier, x, y float32) {
	if tp.scp.inStreamMode() {
		return
	}
	tp.advTriggerPointViewer.mouseUp(button, modifier, x, y)
	prev := tp.lowerSelected || tp.upperSelected
	tp.lowerSelected = false
	tp.upperSelected = false
	if prev {
		if !tp.mouseAtIntervalPoint(x, y) {
			tp.mouseAt = false
		}
		tp.enableRefresh()
		if tp.raster() != nil {
			tp.raster().Refresh()
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
	bounds := tp.signalScreen().Bounds()
	w := float64(bounds.Dx() - 1)

	if w <= 0 || tp.maxScreenTime() <= 0 {
		return
	}

	// Convert x coordinate to time offset from trigger point
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
		// Single mode: either handle controls the single ΔT
		if pwType == genericps.PwTypeLessThan {
			channel.Trigger.IntervalTimeUpper = timeOffset
			channel.Trigger.IntervalTimeLower = timeOffset
		} else { // GreaterThan
			channel.Trigger.IntervalTimeUpper = timeOffset
			channel.Trigger.IntervalTimeLower = timeOffset
		}
	} else {
		// Range mode: separate lower/upper handles
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

func (tp *intervalTriggerPointViewer) draw() {
	if tp.scp.controlTab.SelectedIndex() == dftTabIndex || tp.scp.inStreamMode() {
		return
	}
	tp.advTriggerPointViewer.draw()

	if tp.scp.triggerSource != dontCare {
		channel := &tp.scp.Settings.Channels[tp.scp.triggerSource]
		if !channel.TriggerSource {
			return
		}
		x, y := tp.timeMv2xy(channel.Trigger.Mv)

		yList := []float32{}
		if tp.scp.triggerSettingMsg.Type == control.WindowPulseWidth || tp.scp.triggerSettingMsg.Type == control.WindowDropout {
			// For Window Pulse Width, the user wants arrows at BOTH thresholds
			_, yUpper := tp.timeMv2xy(channel.Trigger.Mv)
			_, yLower := tp.timeMv2xy(channel.Trigger.LowerMv)
			yList = []float32{yUpper, yLower}
		} else {
			yList = []float32{y}
		}

		bounds := tp.signalScreen().Bounds()
		w := float64(bounds.Dx() - 1)

		if w <= 0 || tp.maxScreenTime() <= 0 {
			return
		}

		pwType := channel.Trigger.IntervalType
		isSingle := intervalSingleModeTypes[pwType]

		halfRectSize := float32(triggerPointR * 2)
		rectSize2 := 2 * halfRectSize

		tp.lowerHImgRects = nil
		tp.upperHImgRects = nil

		for _, currY := range yList {
			if isSingle {
				// Single handle mode: show only one horizontal handle for the single ΔT
				var singleTime float64
				if pwType == genericps.PwTypeLessThan {
					singleTime = channel.Trigger.IntervalTimeUpper
				} else { // GreaterThan
					singleTime = channel.Trigger.IntervalTimeLower
				}
				singleDx := float32((singleTime / tp.maxScreenTime()) * w)
				xSingle := x - singleDx

				// Add to lowerHImgRects for hit-testing
				rect := image.Rect(
					int(math.Round(float64(xSingle-rectSize2))),
					int(math.Round(float64(currY-rectSize2))),
					int(math.Round(float64(xSingle+rectSize2))),
					int(math.Round(float64(currY+rectSize2))))
				tp.lowerHImgRects = append(tp.lowerHImgRects, rect)

				colSingle := theme.ForegroundColor()
				if tp.lowerSelected || (tp.mouseAt && tp.mouseAtIntervalPoint(xSingle, currY)) {
					colSingle = theme.SelectionColor()
				}

				// Draw horizontal line from trigger point to the single handle
				drawLine(tp.signalScreen(), x, currY, xSingle, currY, colSingle)
				if pwType == genericps.PwTypeLessThan {
					// Point RIGHT (towards origin, meaning smaller time/less than)
					drawLine(tp.signalScreen(), xSingle-halfRectSize, currY-halfRectSize, xSingle-halfRectSize, currY+halfRectSize, colSingle)
					drawLine(tp.signalScreen(), xSingle-halfRectSize, currY-halfRectSize, xSingle, currY, colSingle)
					drawLine(tp.signalScreen(), xSingle-halfRectSize, currY+halfRectSize, xSingle, currY, colSingle)
				} else {
					// Point LEFT (away from origin, meaning larger time/greater than)
					drawLine(tp.signalScreen(), xSingle+halfRectSize, currY-halfRectSize, xSingle+halfRectSize, currY+halfRectSize, colSingle)
					drawLine(tp.signalScreen(), xSingle+halfRectSize, currY-halfRectSize, xSingle, currY, colSingle)
					drawLine(tp.signalScreen(), xSingle+halfRectSize, currY+halfRectSize, xSingle, currY, colSingle)
				}

			} else {
				// Range mode: two time handles + single trigger point
				lowerDx := float32((channel.Trigger.IntervalTimeLower / tp.maxScreenTime()) * w)
				upperDx := float32((channel.Trigger.IntervalTimeUpper / tp.maxScreenTime()) * w)

				xLower := x - lowerDx
				xUpper := x - upperDx

				lRect := image.Rect(
					int(math.Round(float64(xLower-rectSize2))),
					int(math.Round(float64(currY-rectSize2))),
					int(math.Round(float64(xLower+rectSize2))),
					int(math.Round(float64(currY+rectSize2))))
				tp.lowerHImgRects = append(tp.lowerHImgRects, lRect)

				uRect := image.Rect(
					int(math.Round(float64(xUpper-rectSize2))),
					int(math.Round(float64(currY-rectSize2))),
					int(math.Round(float64(xUpper+rectSize2))),
					int(math.Round(float64(currY+rectSize2))))
				tp.upperHImgRects = append(tp.upperHImgRects, uRect)

				colLower := theme.ForegroundColor()
				if tp.lowerSelected || (tp.mouseAt && tp.mouseAtIntervalPoint(xLower, currY)) {
					colLower = theme.SelectionColor()
				}

				colUpper := theme.ForegroundColor()
				if tp.upperSelected || (tp.mouseAt && tp.mouseAtIntervalPoint(xUpper, currY)) {
					colUpper = theme.SelectionColor()
				}

				// Draw horizontal line from trigger point to xLower
				drawLine(tp.signalScreen(), x, currY, xLower, currY, colLower)
				// Draw horizontal line from trigger point to xUpper
				drawLine(tp.signalScreen(), x, currY, xUpper, currY, colUpper)

				switch pwType {
				case genericps.PwTypeInRange:
					// xLower points LEFT (inside)
					drawLine(tp.signalScreen(), xLower+halfRectSize, currY-halfRectSize, xLower+halfRectSize, currY+halfRectSize, colLower)
					drawLine(tp.signalScreen(), xLower+halfRectSize, currY-halfRectSize, xLower, currY, colLower)
					drawLine(tp.signalScreen(), xLower+halfRectSize, currY+halfRectSize, xLower, currY, colLower)
					
					// xUpper points RIGHT (inside)
					drawLine(tp.signalScreen(), xUpper-halfRectSize, currY-halfRectSize, xUpper-halfRectSize, currY+halfRectSize, colUpper)
					drawLine(tp.signalScreen(), xUpper-halfRectSize, currY-halfRectSize, xUpper, currY, colUpper)
					drawLine(tp.signalScreen(), xUpper-halfRectSize, currY+halfRectSize, xUpper, currY, colUpper)
				case genericps.PwTypeOutOfRange:
					// xLower points RIGHT (outside)
					drawLine(tp.signalScreen(), xLower-halfRectSize, currY-halfRectSize, xLower-halfRectSize, currY+halfRectSize, colLower)
					drawLine(tp.signalScreen(), xLower-halfRectSize, currY-halfRectSize, xLower, currY, colLower)
					drawLine(tp.signalScreen(), xLower-halfRectSize, currY+halfRectSize, xLower, currY, colLower)
					
					// xUpper points LEFT (outside)
					drawLine(tp.signalScreen(), xUpper+halfRectSize, currY-halfRectSize, xUpper+halfRectSize, currY+halfRectSize, colUpper)
					drawLine(tp.signalScreen(), xUpper+halfRectSize, currY-halfRectSize, xUpper, currY, colUpper)
					drawLine(tp.signalScreen(), xUpper+halfRectSize, currY+halfRectSize, xUpper, currY, colUpper)
				}
			}
		}

		// Update time displays once per draw call
		if isSingle {
			if tp.scp.intervalTimeSingleDisp != nil {
				var singleTime float64
				if pwType == genericps.PwTypeLessThan {
					singleTime = channel.Trigger.IntervalTimeUpper
				} else { // GreaterThan
					singleTime = channel.Trigger.IntervalTimeLower
				}
				unit := getBaseTimeUnit(tp.scp.Settings.Time.Unit)
				multiplier := getIntervalUnitMultiplier(unit)
				val := int(math.Round(singleTime / multiplier))
				if tp.scp.intervalTimeSingleDisp.Value != val {
					tp.scp.intervalTimeSingleDisp.SilentSetValue(val)
					tp.scp.intervalTimeSingleDisp.Refresh()
				}
			}
		} else {
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

func (tp *intervalTriggerPointViewer) scrolled(delta, x, y float32) {
	if tp.scp.inStreamMode() {
		return
	}
	tp.advTriggerPointViewer.scrolled(delta, x, y)
}
