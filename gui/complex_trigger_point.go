package gui

import (
	"fynescope/genericps"
	"image"
	"image/color"
	"image/draw"
	"log/slog"
	"math"

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
)

type complexHit struct {
	channelIndex int
	handle       complexHandleType
}

type complexTriggerPointViewer struct {
	rasterPartition
	scp         *ScpDesc
	hoveredHit  complexHit
	selectedHit complexHit

	mainRects map[int]image.Rectangle
	uhRects   map[int]image.Rectangle
	lRects    map[int]image.Rectangle
	lhRects   map[int]image.Rectangle
	isTimeZoom bool
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
	if channel.Trigger.Type == "Window" || channel.Trigger.ThresholdMode == genericps.Window {
		if newMv < channel.Trigger.LowerMv+genericps.MinThresholdDiff {
			newMv = channel.Trigger.LowerMv + genericps.MinThresholdDiff
		}
	}
	channel.Trigger.Mv = newMv

	tp.scp.buildComplexTriggerMessage()
	tp.scp.psControl.SetTriggerCh <- &tp.scp.triggerSettingMsg
	<-tp.scp.triggerSettingMsg.Done

	lw := tp.scp.ftBottomLabelViewer.(*timeLabelViewer)
	tp.scp.clearAllFtPersistentLayers()
	tp.scp.clearAllDftPersistentLayers()
	lw.enableRefresh()
	tp.enableRefresh()

	if tp.scp.ftRaster != nil {
		tp.scp.ftRaster.Refresh()
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
	if channel.Trigger.Type == "Window" || channel.Trigger.ThresholdMode == genericps.Window {
		if newMv > channel.Trigger.Mv-genericps.MinThresholdDiff {
			newMv = channel.Trigger.Mv - genericps.MinThresholdDiff
		}
	}
	// if newMv > channel.Trigger.Mv {
	// 	newMv = channel.Trigger.Mv
	// }
	channel.Trigger.LowerMv = newMv

	tp.scp.buildComplexTriggerMessage()
	tp.scp.psControl.SetTriggerCh <- &tp.scp.triggerSettingMsg
	<-tp.scp.triggerSettingMsg.Done

	lw := tp.scp.ftBottomLabelViewer.(*timeLabelViewer)
	tp.scp.clearAllFtPersistentLayers()
	tp.scp.clearAllDftPersistentLayers()
	lw.enableRefresh()
	tp.enableRefresh()

	if tp.raster() != nil {
		tp.raster().Refresh()
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
		if channel.Trigger.Type == "Window" || channel.Trigger.ThresholdMode == genericps.Window {
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
		} else if channel.Trigger.Type == "Advanced" {
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
		tp.scp.psControl.SetTriggerCh <- &tp.scp.triggerSettingMsg
		<-tp.scp.triggerSettingMsg.Done
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
		tp.scp.psControl.SetTriggerCh <- &tp.scp.triggerSettingMsg
		<-tp.scp.triggerSettingMsg.Done
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
	if tp.scp.controlTab.SelectedIndex() == dftTabIndex || tp.scp.inStreamMode() {
		return
	}

	tp.mainRects = make(map[int]image.Rectangle)
	tp.uhRects = make(map[int]image.Rectangle)
	tp.lRects = make(map[int]image.Rectangle)
	tp.lhRects = make(map[int]image.Rectangle)

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

			if chCfg.Type == "Window" || chCfg.ThresholdMode == genericps.Window {
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
			} else if chCfg.Type == "Advanced" {
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
				if tp.scp.triggerHysteresisDisp.Value != int(chCfg.Hysteresis) {
					tp.scp.triggerHysteresisDisp.SilentSetValue(int(chCfg.Hysteresis))
					tp.scp.triggerHysteresisDisp.Refresh()
				}
				if tp.scp.triggerLowerHysteresisDisp != nil {
					if tp.scp.triggerLowerHysteresisDisp.Value != int(chCfg.LowerHysteresis) {
						tp.scp.triggerLowerHysteresisDisp.SilentSetValue(int(chCfg.LowerHysteresis))
						tp.scp.triggerLowerHysteresisDisp.Refresh()
					}
				}
			}
		}
	}
}

func (tp *complexTriggerPointViewer) clear() {
}
