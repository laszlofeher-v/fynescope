package gui

import (
	"fynescope/genericps"
	"image"
	"image/color"
	"image/draw"
	"log/slog"
	"math"

	"fynescope/settings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
)

type complexHandleType int

const (
	complexHandleNone complexHandleType = iota
	complexHandleMain
	complexHandleUpperHyst
	complexHandleLower
	complexHandleLowerHyst
	complexHandleIntervalLower
	complexHandleIntervalUpper
)

func isWindowType(triggerType string) bool {
	return triggerType == settings.TriggerTypeWindow ||
		triggerType == settings.TriggerTypeWindowPulseWidth ||
		triggerType == settings.TriggerTypeWindowDropout ||
		triggerType == settings.TriggerTypeRunt ||
		triggerType == settings.TriggerTypeRiseFall
}

func isAdvancedType(triggerType string) bool {
	return triggerType == settings.TriggerTypeAdvanced ||
		triggerType == settings.TriggerTypeInterval ||
		triggerType == settings.TriggerTypePulseWidth ||
		triggerType == settings.TriggerTypeDropout
}

func isTimeTriggerType(triggerType string) bool {
	return triggerType == settings.TriggerTypeInterval ||
		triggerType == settings.TriggerTypePulseWidth ||
		triggerType == settings.TriggerTypeDropout ||
		triggerType == settings.TriggerTypeWindowPulseWidth ||
		triggerType == settings.TriggerTypeWindowDropout ||
		triggerType == settings.TriggerTypeRiseFall
}

type complexHit struct {
	channelIndex int
	handle       complexHandleType
}

type complexTriggerPointViewer struct {
	rasterPartition
	scp         *ScpDesc
	hoveredHit  complexHit
	selectedHit complexHit

	mainRects     map[int]image.Rectangle
	uhRects       map[int]image.Rectangle
	lRects        map[int]image.Rectangle
	lhRects       map[int]image.Rectangle
	intLowerRects map[int][]image.Rectangle
	intUpperRects map[int][]image.Rectangle
	isTimeZoom    bool
}

var (
	_ mouser     = (*complexTriggerPointViewer)(nil)
	_ dragger    = (*complexTriggerPointViewer)(nil)
	_ scroller   = (*complexTriggerPointViewer)(nil)
	_ drawer     = (*complexTriggerPointViewer)(nil)
	_ cursorable = (*complexTriggerPointViewer)(nil)
)

func newComplexTriggerPointViewer(img rasterImage, scp *ScpDesc, isTimeZoom bool) *complexTriggerPointViewer {
	return &complexTriggerPointViewer{
		rasterPartition: rasterPartition{img: img},
		scp:             scp,
		hoveredHit:      complexHit{-1, complexHandleNone},
		selectedHit:     complexHit{-1, complexHandleNone},
		mainRects:       make(map[int]image.Rectangle),
		uhRects:         make(map[int]image.Rectangle),
		lRects:          make(map[int]image.Rectangle),
		lhRects:         make(map[int]image.Rectangle),
		intLowerRects:   make(map[int][]image.Rectangle),
		intUpperRects:   make(map[int][]image.Rectangle),
		isTimeZoom:      isTimeZoom,
	}
}

func (tp *complexTriggerPointViewer) signalScreen() draw.RGBA64Image {
	if tp.isTimeZoom {
		return tp.scp.timeZoomScopeSignalScreen
	}
	return tp.scp.ftScopeSignalScreen
}

func (tp *complexTriggerPointViewer) maxScreenTime() float64 {
	if tp.isTimeZoom {
		return tp.scp.timeZoomMaxScreenTime
	}
	return tp.scp.maxScreenTime
}

func (tp *complexTriggerPointViewer) raster() *screenRaster {
	if tp.isTimeZoom {
		return tp.scp.timeZoomRaster
	}
	return tp.scp.ftRaster
}

func (tp *complexTriggerPointViewer) timeMv2xy(mv int32, channelIndex int) (x, y float32) {
	bounds := tp.signalScreen().Bounds()
	zeroOffset := float64(bounds.Min.Y + bounds.Dy()/2)
	h := float64(bounds.Dy())
	channel := &tp.scp.Settings.Channels[channelIndex]
	channelViewer := &tp.scp.channelViewers[channelIndex]
	yScale := h / float64(2.0*genericps.RangeValuesMv[channel.VRange])
	yOffset := float64(0)
	if channelViewer.displayOffsetInt != 0 {
		yOffset = tp.scp.offsetNToFtY(channelViewer.displayOffsetInt)
	}
	if channel.Inverted {
		mv = -mv
	}
	y = float32(-yScale*float64(mv) + yOffset + zeroOffset)
	triggerTimeOffset := tp.scp.Settings.Time.TriggerTimeOffset
	if !tp.isTimeZoom {
		triggerTimeOffset -= tp.scp.timeZoomBoxOffset
	}
	x = float32(bounds.Min.X) + float32(triggerTimeOffset)*
		float32(tp.signalScreen().Bounds().Dx()-1)/float32(tp.maxScreenTime())
	return
}

func (tp *complexTriggerPointViewer) y2mv(y float64, channelIndex int) (mv float64) {
	bounds := tp.signalScreen().Bounds()
	zeroOffset := float64(bounds.Min.Y + bounds.Dy()/2)
	h := float64(bounds.Dy())
	channel := &tp.scp.Settings.Channels[channelIndex]
	channelViewer := &tp.scp.channelViewers[channelIndex]
	yScale := h / float64(2.0*genericps.RangeValuesMv[channel.VRange])
	yOffset := float64(0)
	if channelViewer.displayOffsetInt != 0 {
		yOffset = tp.scp.offsetNToFtY(channelViewer.displayOffsetInt)
	}
	mv = (y - yOffset - zeroOffset) / (-yScale)
	if channel.Inverted {
		mv = -mv
	}
	return
}

func (tp *complexTriggerPointViewer) getHit(x, y float32) complexHit {
	p := image.Point{X: int(math.Round(float64(x))), Y: int(math.Round(float64(y)))}

	// Check main rects
	for chIdx, rect := range tp.mainRects {
		if p.In(rect) {
			return complexHit{chIdx, complexHandleMain}
		}
	}
	// Check upper hysteresis rects
	for chIdx, rect := range tp.uhRects {
		if p.In(rect) {
			return complexHit{chIdx, complexHandleUpperHyst}
		}
	}
	// Check lower rects
	for chIdx, rect := range tp.lRects {
		if p.In(rect) {
			return complexHit{chIdx, complexHandleLower}
		}
	}
	// Check lower hysteresis rects
	for chIdx, rect := range tp.lhRects {
		if p.In(rect) {
			return complexHit{chIdx, complexHandleLowerHyst}
		}
	}
	// Check interval lower rects
	for chIdx, rects := range tp.intLowerRects {
		for _, rect := range rects {
			if p.In(rect) {
				return complexHit{chIdx, complexHandleIntervalLower}
			}
		}
	}
	// Check interval upper rects
	for chIdx, rects := range tp.intUpperRects {
		for _, rect := range rects {
			if p.In(rect) {
				return complexHit{chIdx, complexHandleIntervalUpper}
			}
		}
	}

	return complexHit{-1, complexHandleNone}
}

func (tp *complexTriggerPointViewer) cursor(x, y float32) (desktop.Cursor, bool) {
	if tp.scp.inStreamMode() {
		return desktop.DefaultCursor, false
	}
	hit := tp.getHit(x, y)
	if hit.handle != complexHandleNone || tp.selectedHit.handle != complexHandleNone {
		return desktop.PointerCursor, true
	}
	return desktop.DefaultCursor, false
}

func (tp *complexTriggerPointViewer) mouseMoved(x, y float32) {
	if tp.scp.inStreamMode() {
		return
	}
	prev := tp.hoveredHit
	tp.hoveredHit = tp.getHit(x, y)

	if prev != tp.hoveredHit {
		tp.enableRefresh()
		if tp.raster() != nil {
			tp.raster().Refresh()
		}
	}
}

func (tp *complexTriggerPointViewer) mouseDown(button desktop.MouseButton, modifier fyne.KeyModifier, x, y float32) {
	if tp.scp.inStreamMode() {
		return
	}
	tp.selectedHit = tp.getHit(x, y)
}

func (tp *complexTriggerPointViewer) mouseUp(button desktop.MouseButton, modifier fyne.KeyModifier, x, y float32) {
	if tp.scp.inStreamMode() {
		return
	}
	refresh := false
	if tp.selectedHit.handle != complexHandleNone {
		tp.selectedHit = complexHit{-1, complexHandleNone}
		tp.hoveredHit = tp.getHit(x, y)
		refresh = true

		// If lower bounds were changed, save settings
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

func (tp *complexTriggerPointViewer) setDispOffset(dx, x, y float32, chIdx int) {
	bounds := tp.signalScreen().Bounds()
	if int(x) < bounds.Min.X || int(x) > bounds.Max.X ||
		int(y) < bounds.Min.Y || int(y) > bounds.Max.Y {
		return
	}
	mv := tp.y2mv(float64(y), chIdx)
	channel := &tp.scp.Settings.Channels[chIdx]
	bound := float64(genericps.InputRanges[channel.VRange])
	if mv < -bound || mv > bound {
		return
	}

	tp.scp.addFtXOffset(float64(dx))
	tp.scp.setTriggerTime(tp.scp.Settings.Time.TriggerTimeOffset)

	newMv := int32(math.Round(float64(mv)))
	if isWindowType(channel.Trigger.Type) || channel.Trigger.ThresholdMode == genericps.Window {
		minThresholdDiff := genericps.GetMinThresholdDiff(channel.VRange)
		if newMv < channel.Trigger.LowerMv+minThresholdDiff {
			newMv = channel.Trigger.LowerMv + minThresholdDiff
		}
	}
	channel.Trigger.Mv = newMv

	tp.scp.buildComplexTriggerMessage()
	t := tp.scp.triggerSettingMsg
	t.Done = make(chan struct{}, 1)
	go func() {
		tp.scp.psControl.SetTriggerCh <- &t
		<-t.Done
	}()

	lw := tp.scp.ftBottomLabelViewer.(*timeLabelViewer)
	tp.scp.clearAllFtPersistentLayers()
	tp.scp.clearAllDftPersistentLayers()
	lw.enableRefresh()
	tp.enableRefresh()

	if tp.scp.ftRaster != nil {
		tp.scp.ftRaster.Refresh()
	}
	if tp.scp.digitalRaster != nil {
		tp.scp.digitalRaster.refresh()
	}
}

func (tp *complexTriggerPointViewer) setLowerDispOffset(dx, x, y float32, chIdx int) {
	bounds := tp.signalScreen().Bounds()
	if int(x) < bounds.Min.X || int(x) > bounds.Max.X ||
		int(y) < bounds.Min.Y || int(y) > bounds.Max.Y {
		return
	}
	mv := tp.y2mv(float64(y), chIdx)
	channel := &tp.scp.Settings.Channels[chIdx]
	bound := float64(genericps.InputRanges[channel.VRange])
	if mv < -bound || mv > bound {
		return
	}
	tp.scp.addFtXOffset(float64(dx))
	tp.scp.setTriggerTime(tp.scp.Settings.Time.TriggerTimeOffset)

	newMv := int32(math.Round(float64(mv)))
	if isWindowType(channel.Trigger.Type) || channel.Trigger.ThresholdMode == genericps.Window {
		minThresholdDiff := genericps.GetMinThresholdDiff(channel.VRange)
		if newMv > channel.Trigger.Mv-minThresholdDiff {
			newMv = channel.Trigger.Mv - minThresholdDiff
		}
	}
	channel.Trigger.LowerMv = newMv

	tp.scp.buildComplexTriggerMessage()
	t := tp.scp.triggerSettingMsg
	t.Done = make(chan struct{}, 1)
	go func() {
		tp.scp.psControl.SetTriggerCh <- &t
		<-t.Done
	}()

	lw := tp.scp.ftBottomLabelViewer.(*timeLabelViewer)
	tp.scp.clearAllFtPersistentLayers()
	tp.scp.clearAllDftPersistentLayers()
	lw.enableRefresh()
	tp.enableRefresh()

	if tp.raster() != nil {
		tp.raster().Refresh()
	}
	if tp.scp.digitalRaster != nil {
		tp.scp.digitalRaster.refresh()
	}
}

func (tp *complexTriggerPointViewer) dragged(dx, dy, x, y float32) {
	if tp.scp.inStreamMode() {
		return
	}
	if tp.selectedHit.handle == complexHandleNone {
		return
	}

	chIdx := tp.selectedHit.channelIndex
	channel := &tp.scp.Settings.Channels[chIdx]

	if tp.selectedHit.handle == complexHandleMain {
		tp.setDispOffset(dx, x, y, chIdx)
		return
	}

	if tp.selectedHit.handle == complexHandleLower {
		tp.setLowerDispOffset(dx, x, y, chIdx)
		return
	}

	newH := int32(math.Round(tp.y2mv(float64(y), chIdx)))

	if tp.selectedHit.handle == complexHandleUpperHyst {
		if isWindowType(channel.Trigger.Type) || channel.Trigger.ThresholdMode == genericps.Window {
			switch channel.Trigger.TriggerDirection {
			case genericps.TriggerRising, genericps.TriggerInside, genericps.TriggerOutside, genericps.TriggerEnter, genericps.TriggerEnterOrExit:
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
		} else {
			switch channel.Trigger.TriggerDirection {
			case genericps.TriggerRising:
				if newH <= channel.Trigger.Mv {
					channel.Trigger.Hysteresis = channel.Trigger.Mv - newH
				}
			case genericps.TriggerFalling:
				if newH >= channel.Trigger.Mv {
					channel.Trigger.Hysteresis = newH - channel.Trigger.Mv
				}
			default:
				slog.Error("advTrigger", "TriggerDirection", channel.Trigger.TriggerDirection)
			}
		}

		tp.scp.buildComplexTriggerMessage()
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
		return
	}

	if tp.selectedHit.handle == complexHandleLowerHyst {
		switch channel.Trigger.TriggerDirection {
		case genericps.TriggerRising, genericps.TriggerInside, genericps.TriggerOutside, genericps.TriggerEnter, genericps.TriggerEnterOrExit:
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

		tp.scp.buildComplexTriggerMessage()
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
		return
	}

	if tp.selectedHit.handle == complexHandleIntervalLower || tp.selectedHit.handle == complexHandleIntervalUpper {
		bounds := tp.signalScreen().Bounds()
		w := float64(bounds.Dx() - 1)
		if w <= 0 || tp.maxScreenTime() <= 0 {
			return
		}
		triggerX, _ := tp.timeMv2xy(channel.Trigger.Mv, chIdx)
		timeOffset := (float64(triggerX-x) / w) * tp.maxScreenTime()
		if timeOffset < 0 {
			timeOffset = 0
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
			channel.Trigger.IntervalTimeUpper = timeOffset
			channel.Trigger.IntervalTimeLower = timeOffset
		} else {
			if tp.selectedHit.handle == complexHandleIntervalLower {
				if channel.Trigger.IntervalTimeUpper > 0 && timeOffset > channel.Trigger.IntervalTimeUpper {
					timeOffset = channel.Trigger.IntervalTimeUpper
				}
				channel.Trigger.IntervalTimeLower = timeOffset
			} else {
				if timeOffset < channel.Trigger.IntervalTimeLower {
					timeOffset = channel.Trigger.IntervalTimeLower
				}
				channel.Trigger.IntervalTimeUpper = timeOffset
			}
		}
		if genericps.ChannelId(chIdx) == tp.scp.triggerSource {
			tp.scp.triggerSettingMsg.IntervalTimeLower = channel.Trigger.IntervalTimeLower
			tp.scp.triggerSettingMsg.IntervalTimeUpper = channel.Trigger.IntervalTimeUpper
		}
		tp.scp.buildComplexTriggerMessage()
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
		return
	}
}

func (tp *complexTriggerPointViewer) scrolled(delta, x, y float32) {
}

func (tp *complexTriggerPointViewer) draw() {
	if tp.scp.getActiveFunctionIndex() == dftTabIndex || tp.scp.inStreamMode() {
		return
	}

	tp.mainRects = make(map[int]image.Rectangle)
	tp.uhRects = make(map[int]image.Rectangle)
	tp.lRects = make(map[int]image.Rectangle)
	tp.lhRects = make(map[int]image.Rectangle)
	tp.intLowerRects = make(map[int][]image.Rectangle)
	tp.intUpperRects = make(map[int][]image.Rectangle)

	bounds := tp.signalScreen().Bounds()
	w := float64(bounds.Dx() - 1)

	for i, ch := range tp.scp.Settings.Channels {
		chCfg := ch.Trigger
		if chCfg.Condition != genericps.CondDontCare && ch.Enabled {
			x, y := tp.timeMv2xy(chCfg.Mv, i)
			bound := tp.signalScreen().Bounds()
			maxY := float32(bound.Max.Y)
			minY := float32(bound.Min.Y)
			if y > maxY {
				y = maxY
			}
			if y < minY {
				y = minY
			}

			halfRectSize := float32(triggerPointR * 2)
			rectSize2 := 2 * halfRectSize

			tp.mainRects[i] = image.Rect(
				int(math.Round(float64(x-halfRectSize))),
				int(math.Round(float64(y-halfRectSize))),
				int(math.Round(float64(x+halfRectSize))),
				int(math.Round(float64(y+halfRectSize))),
			)

			col := ch.Col[tp.scp.Settings.ChannelColorIndex]

			// Main Point Color
			var mainCol color.Color = col
			if tp.selectedHit.channelIndex == i && tp.selectedHit.handle == complexHandleMain ||
				tp.hoveredHit.channelIndex == i && tp.hoveredHit.handle == complexHandleMain {
				mainCol = theme.SelectionColor()
				drawCircle(tp.signalScreen(), x, y, triggerPointR, mainCol)
				drawCircle(tp.signalScreen(), x, y, triggerPointR-1, mainCol)
				drawCircle(tp.signalScreen(), x, y, triggerPointR-2, col)
				drawCircle(tp.signalScreen(), x, y, triggerPointR-3, col)
				drawCircle(tp.signalScreen(), x, y, triggerPointR-4, col)
				drawCircle(tp.signalScreen(), x, y, triggerPointR-5, col)
			} else {
				drawCircle(tp.signalScreen(), x, y, triggerPointR, mainCol)
				drawCircle(tp.signalScreen(), x, y, triggerPointR-1, mainCol)
				drawCircle(tp.signalScreen(), x, y, triggerPointR-2, mainCol)
				drawCircle(tp.signalScreen(), x, y, triggerPointR-3, mainCol)
				drawCircle(tp.signalScreen(), x, y, triggerPointR-4, mainCol)
			}

			isWin := isWindowType(chCfg.Type) || chCfg.ThresholdMode == genericps.Window
			isAdv := isAdvancedType(chCfg.Type) || (!isWin && chCfg.Type != settings.TriggerTypeSimple && chCfg.Type != "")

			if isWin {
				var yh float32
				_, yh = tp.timeMv2xy(chCfg.Mv+chCfg.Hysteresis, i)
				if chCfg.TriggerDirection == genericps.TriggerFalling || chCfg.TriggerDirection == genericps.TriggerExit {
					_, yh = tp.timeMv2xy(chCfg.Mv-chCfg.Hysteresis, i)
				}
				if yh > maxY {
					yh = maxY
				}
				if yh < minY {
					yh = minY
				}

				tp.uhRects[i] = image.Rect(int(math.Round(float64(x-rectSize2))),
					int(math.Round(float64(yh-rectSize2))),
					int(math.Round(float64(x+rectSize2))),
					int(math.Round(float64(rectSize2+yh))))

				lx, ly := tp.timeMv2xy(chCfg.LowerMv, i)
				if ly > maxY {
					ly = maxY
				}
				if ly < minY {
					ly = minY
				}

				tp.lRects[i] = image.Rect(int(math.Round(float64(lx-halfRectSize))),
					int(math.Round(float64(ly-halfRectSize))),
					int(math.Round(float64(lx+halfRectSize))),
					int(math.Round(float64(ly+halfRectSize))))

				var lyh float32
				_, lyh = tp.timeMv2xy(chCfg.LowerMv-chCfg.LowerHysteresis, i)
				if chCfg.TriggerDirection == genericps.TriggerFalling || chCfg.TriggerDirection == genericps.TriggerExit {
					_, lyh = tp.timeMv2xy(chCfg.LowerMv+chCfg.LowerHysteresis, i)
				}
				if lyh > maxY {
					lyh = maxY
				}
				if lyh < minY {
					lyh = minY
				}

				tp.lhRects[i] = image.Rect(int(math.Round(float64(lx-rectSize2))),
					int(math.Round(float64(lyh-rectSize2))),
					int(math.Round(float64(lx+rectSize2))),
					int(math.Round(float64(rectSize2+lyh))))

				// Draw Upper Hysteresis
				var uhCol color.Color = col
				if tp.selectedHit.channelIndex == i && tp.selectedHit.handle == complexHandleUpperHyst ||
					tp.hoveredHit.channelIndex == i && tp.hoveredHit.handle == complexHandleUpperHyst {
					uhCol = theme.SelectionColor()
				}
				drawLine(tp.signalScreen(), x, y, x, yh, uhCol)
				drawLine(tp.signalScreen(), x-halfRectSize, yh, x+halfRectSize, yh, uhCol)

				// Draw Lower Point
				var lCol color.Color = col
				if tp.selectedHit.channelIndex == i && tp.selectedHit.handle == complexHandleLower ||
					tp.hoveredHit.channelIndex == i && tp.hoveredHit.handle == complexHandleLower {
					lCol = theme.SelectionColor()
					drawCircle(tp.signalScreen(), lx, ly, triggerPointR, lCol)
					drawCircle(tp.signalScreen(), lx, ly, triggerPointR-1, lCol)
					drawCircle(tp.signalScreen(), lx, ly, triggerPointR-2, col)
					drawCircle(tp.signalScreen(), lx, ly, triggerPointR-3, col)
					drawCircle(tp.signalScreen(), lx, ly, triggerPointR-4, col)
					drawCircle(tp.signalScreen(), lx, ly, triggerPointR-5, col)
				} else {
					drawCircle(tp.signalScreen(), lx, ly, triggerPointR, lCol)
					drawCircle(tp.signalScreen(), lx, ly, triggerPointR-1, lCol)
					drawCircle(tp.signalScreen(), lx, ly, triggerPointR-2, lCol)
					drawCircle(tp.signalScreen(), lx, ly, triggerPointR-3, lCol)
					drawCircle(tp.signalScreen(), lx, ly, triggerPointR-4, lCol)
				}

				// Draw Lower Hysteresis
				var lhCol color.Color = col
				if tp.selectedHit.channelIndex == i && tp.selectedHit.handle == complexHandleLowerHyst ||
					tp.hoveredHit.channelIndex == i && tp.hoveredHit.handle == complexHandleLowerHyst {
					lhCol = theme.SelectionColor()
				}
				drawLine(tp.signalScreen(), lx, ly, lx, lyh, lhCol)
				drawLine(tp.signalScreen(), lx-halfRectSize, lyh, lx+halfRectSize, lyh, lhCol)

			} else if isAdv {
				// Draw Advanced (Level) Trigger Hysteresis
				var yh float32
				_, yh = tp.timeMv2xy(chCfg.Mv-chCfg.Hysteresis, i)
				if chCfg.TriggerDirection == genericps.TriggerFalling {
					_, yh = tp.timeMv2xy(chCfg.Mv+chCfg.Hysteresis, i)
				}
				if yh > maxY {
					yh = maxY
				}
				if yh < minY {
					yh = minY
				}

				tp.uhRects[i] = image.Rect(int(math.Round(float64(x-rectSize2))),
					int(math.Round(float64(yh-rectSize2))),
					int(math.Round(float64(x+rectSize2))),
					int(math.Round(float64(rectSize2+yh))))

				var uhCol color.Color = col
				if tp.selectedHit.channelIndex == i && tp.selectedHit.handle == complexHandleUpperHyst ||
					tp.hoveredHit.channelIndex == i && tp.hoveredHit.handle == complexHandleUpperHyst {
					uhCol = theme.SelectionColor()
				}
				drawLine(tp.signalScreen(), x, y, x, yh, uhCol)
				drawLine(tp.signalScreen(), x-halfRectSize, yh, x+halfRectSize, yh, uhCol)
			}

			// Draw Time Handles if applicable
			if isTimeTriggerType(chCfg.Type) && w > 0 && tp.maxScreenTime() > 0 {
				pwType := chCfg.IntervalType
				isSingle := intervalSingleModeTypes[pwType]

				yList := []float32{y}
				if isWin {
					_, yLower := tp.timeMv2xy(chCfg.LowerMv, i)
					yList = append(yList, yLower)
				}

				for _, currY := range yList {
					if isSingle {
						var singleTime float64
						if pwType == genericps.PwTypeLessThan {
							singleTime = chCfg.IntervalTimeUpper
						} else {
							singleTime = chCfg.IntervalTimeLower
						}
						singleDx := float32((singleTime / tp.maxScreenTime()) * w)
						xSingle := x - singleDx

						rect := image.Rect(
							int(math.Round(float64(xSingle-rectSize2))),
							int(math.Round(float64(currY-rectSize2))),
							int(math.Round(float64(xSingle+rectSize2))),
							int(math.Round(float64(currY+rectSize2))))
						tp.intLowerRects[i] = append(tp.intLowerRects[i], rect)

						var colSingle color.Color = col
						if (tp.selectedHit.channelIndex == i && tp.selectedHit.handle == complexHandleIntervalLower) ||
							(tp.hoveredHit.channelIndex == i && tp.hoveredHit.handle == complexHandleIntervalLower) {
							colSingle = theme.SelectionColor()
						}

						drawLine(tp.signalScreen(), x, currY, xSingle, currY, colSingle)
						if pwType == genericps.PwTypeLessThan {
							drawLine(tp.signalScreen(), xSingle-halfRectSize, currY-halfRectSize, xSingle-halfRectSize, currY+halfRectSize, colSingle)
							drawLine(tp.signalScreen(), xSingle-halfRectSize, currY-halfRectSize, xSingle, currY, colSingle)
							drawLine(tp.signalScreen(), xSingle-halfRectSize, currY+halfRectSize, xSingle, currY, colSingle)
						} else {
							drawLine(tp.signalScreen(), xSingle+halfRectSize, currY-halfRectSize, xSingle+halfRectSize, currY+halfRectSize, colSingle)
							drawLine(tp.signalScreen(), xSingle+halfRectSize, currY-halfRectSize, xSingle, currY, colSingle)
							drawLine(tp.signalScreen(), xSingle+halfRectSize, currY+halfRectSize, xSingle, currY, colSingle)
						}
					} else {
						lowerDx := float32((chCfg.IntervalTimeLower / tp.maxScreenTime()) * w)
						upperDx := float32((chCfg.IntervalTimeUpper / tp.maxScreenTime()) * w)

						xLower := x - lowerDx
						xUpper := x - upperDx

						lRect := image.Rect(
							int(math.Round(float64(xLower-rectSize2))),
							int(math.Round(float64(currY-rectSize2))),
							int(math.Round(float64(xLower+rectSize2))),
							int(math.Round(float64(currY+rectSize2))))
						tp.intLowerRects[i] = append(tp.intLowerRects[i], lRect)

						uRect := image.Rect(
							int(math.Round(float64(xUpper-rectSize2))),
							int(math.Round(float64(currY-rectSize2))),
							int(math.Round(float64(xUpper+rectSize2))),
							int(math.Round(float64(currY+rectSize2))))
						tp.intUpperRects[i] = append(tp.intUpperRects[i], uRect)

						var colLower color.Color = col
						if (tp.selectedHit.channelIndex == i && tp.selectedHit.handle == complexHandleIntervalLower) ||
							(tp.hoveredHit.channelIndex == i && tp.hoveredHit.handle == complexHandleIntervalLower) {
							colLower = theme.SelectionColor()
						}

						var colUpper color.Color = col
						if (tp.selectedHit.channelIndex == i && tp.selectedHit.handle == complexHandleIntervalUpper) ||
							(tp.hoveredHit.channelIndex == i && tp.hoveredHit.handle == complexHandleIntervalUpper) {
							colUpper = theme.SelectionColor()
						}

						drawLine(tp.signalScreen(), x, currY, xLower, currY, colLower)
						drawLine(tp.signalScreen(), x, currY, xUpper, currY, colUpper)

						switch pwType {
						case genericps.PwTypeInRange:
							drawLine(tp.signalScreen(), xLower+halfRectSize, currY-halfRectSize, xLower+halfRectSize, currY+halfRectSize, colLower)
							drawLine(tp.signalScreen(), xLower+halfRectSize, currY-halfRectSize, xLower, currY, colLower)
							drawLine(tp.signalScreen(), xLower+halfRectSize, currY+halfRectSize, xLower, currY, colLower)

							drawLine(tp.signalScreen(), xUpper-halfRectSize, currY-halfRectSize, xUpper-halfRectSize, currY+halfRectSize, colUpper)
							drawLine(tp.signalScreen(), xUpper-halfRectSize, currY-halfRectSize, xUpper, currY, colUpper)
							drawLine(tp.signalScreen(), xUpper-halfRectSize, currY+halfRectSize, xUpper, currY, colUpper)
						case genericps.PwTypeOutOfRange:
							drawLine(tp.signalScreen(), xLower-halfRectSize, currY-halfRectSize, xLower-halfRectSize, currY+halfRectSize, colLower)
							drawLine(tp.signalScreen(), xLower-halfRectSize, currY-halfRectSize, xLower, currY, colLower)
							drawLine(tp.signalScreen(), xLower-halfRectSize, currY+halfRectSize, xLower, currY, colLower)

							drawLine(tp.signalScreen(), xUpper+halfRectSize, currY-halfRectSize, xUpper+halfRectSize, currY+halfRectSize, colUpper)
							drawLine(tp.signalScreen(), xUpper+halfRectSize, currY-halfRectSize, xUpper, currY, colUpper)
							drawLine(tp.signalScreen(), xUpper+halfRectSize, currY+halfRectSize, xUpper, currY, colUpper)
						}
					}
				}
			}

			if genericps.ChannelId(i) == tp.scp.triggerSource {
				if tp.scp.triggerThresholdDisp.Value != int(chCfg.Mv) {
					tp.scp.triggerThresholdDisp.SilentSetValue(int(chCfg.Mv))
					tp.scp.triggerThresholdDisp.Refresh()
				}
				if tp.scp.triggerLowerThresholdDisp != nil {
					if tp.scp.triggerLowerThresholdDisp.Value != int(chCfg.LowerMv) {
						tp.scp.triggerLowerThresholdDisp.SilentSetValue(int(chCfg.LowerMv))
						tp.scp.triggerLowerThresholdDisp.Refresh()
					}
				}
				currentHysteresis := int(chCfg.Hysteresis)
				currentLowerHysteresis := int(chCfg.LowerHysteresis)
				if chCfg.Type == settings.TriggerTypeDropout || chCfg.Type == settings.TriggerTypeWindowDropout {
					currentHysteresis = int(chCfg.DropoutHysteresis)
					currentLowerHysteresis = int(chCfg.DropoutHysteresis)
				}

				if tp.scp.triggerHysteresisDisp.Value != currentHysteresis {
					tp.scp.triggerHysteresisDisp.SilentSetValue(currentHysteresis)
					tp.scp.triggerHysteresisDisp.Refresh()
				}
				if tp.scp.triggerLowerHysteresisDisp != nil {
					if tp.scp.triggerLowerHysteresisDisp.Value != currentLowerHysteresis {
						tp.scp.triggerLowerHysteresisDisp.SilentSetValue(currentLowerHysteresis)
						tp.scp.triggerLowerHysteresisDisp.Refresh()
					}
				}

				// Update time displays for active source if time trigger
				if isTimeTriggerType(chCfg.Type) {
					pwType := chCfg.IntervalType
					isSingle := intervalSingleModeTypes[pwType]
					if isSingle {
						if tp.scp.intervalTimeSingleDisp != nil {
							var singleTime float64
							if pwType == genericps.PwTypeLessThan {
								singleTime = chCfg.IntervalTimeUpper
							} else {
								singleTime = chCfg.IntervalTimeLower
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
							val := int(math.Round(chCfg.IntervalTimeLower / multiplier))
							if tp.scp.intervalTimeLowerDisp.Value != val {
								tp.scp.intervalTimeLowerDisp.SilentSetValue(val)
								tp.scp.intervalTimeLowerDisp.Refresh()
							}
						}
						if tp.scp.intervalTimeUpperDisp != nil {
							unit := getBaseTimeUnit(tp.scp.Settings.Time.Unit)
							multiplier := getIntervalUnitMultiplier(unit)
							val := int(math.Round(chCfg.IntervalTimeUpper / multiplier))
							if tp.scp.intervalTimeUpperDisp.Value != val {
								tp.scp.intervalTimeUpperDisp.SilentSetValue(val)
								tp.scp.intervalTimeUpperDisp.Refresh()
							}
						}
					}
				}
			}
		}
	}
}

func (tp *complexTriggerPointViewer) clear() {
}
