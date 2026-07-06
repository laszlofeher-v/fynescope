package gui

import (
	"fynescope/genericps"
	"image"
	"math"

	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
)

type complexTriggerPointViewer struct {
	rasterPartition
	scp             *ScpDesc
	selectedChannel int
	hoveredChannel  int
	channelRects    map[int]image.Rectangle
}

var (
	_ mouser     = (*complexTriggerPointViewer)(nil)
	_ dragger    = (*complexTriggerPointViewer)(nil)
	_ scroller   = (*complexTriggerPointViewer)(nil)
	_ drawer     = (*complexTriggerPointViewer)(nil)
	_ cursorable = (*complexTriggerPointViewer)(nil)
)

func newComplexTriggerPointViewer(img rasterImage, scp *ScpDesc) *complexTriggerPointViewer {
	return &complexTriggerPointViewer{
		rasterPartition: rasterPartition{img: img},
		scp:             scp,
		selectedChannel: -1,
		hoveredChannel:  -1,
		channelRects:    make(map[int]image.Rectangle),
	}
}

func (tp *complexTriggerPointViewer) timeMv2xy(mv int32, channelIndex int) (x, y float32) {
	bounds := tp.scp.ftScopeSignalScreen.Bounds()
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
	x = float32(bounds.Min.X) + float32(tp.scp.Settings.Time.TriggerTimeOffset)*
		float32(tp.scp.ftScopeSignalScreen.Bounds().Dx())/float32(tp.scp.maxScreenTime)
	return
}

func (tp *complexTriggerPointViewer) y2mv(y float64, channelIndex int) (mv float64) {
	bounds := tp.scp.ftScopeSignalScreen.Bounds()
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

func (tp *complexTriggerPointViewer) mouseInChannel(x, y float32) int {
	p := image.Point{X: int(math.Round(float64(x))), Y: int(math.Round(float64(y)))}
	for chIdx, rect := range tp.channelRects {
		if p.In(rect) {
			return chIdx
		}
	}
	return -1
}

func (tp *complexTriggerPointViewer) cursor(x, y float32) (desktop.Cursor, bool) {
	if tp.scp.inStreamMode() {
		return desktop.DefaultCursor, false
	}
	chIdx := tp.mouseInChannel(x, y)
	if chIdx != -1 || tp.selectedChannel != -1 {
		return desktop.PointerCursor, true
	}
	return desktop.DefaultCursor, false
}

func (tp *complexTriggerPointViewer) mouseMoved(x, y float32) {
	if tp.scp.inStreamMode() {
		return
	}
	prev := tp.hoveredChannel
	tp.hoveredChannel = tp.mouseInChannel(x, y)
	
	if prev != tp.hoveredChannel {
		tp.enableRefresh()
		if tp.scp.ftRaster != nil {
			tp.scp.ftRaster.Refresh()
		}
	}
}

func (tp *complexTriggerPointViewer) mouseDown(button desktop.MouseButton, x, y float32) {
	if tp.scp.inStreamMode() {
		return
	}
	tp.selectedChannel = tp.mouseInChannel(x, y)
}

func (tp *complexTriggerPointViewer) mouseUp(button desktop.MouseButton, x, y float32) {
	tp.selectedChannel = -1
}

func (tp *complexTriggerPointViewer) setDispOffset(dx, x, y float32, chIdx int) {
	bounds := tp.scp.ftScopeSignalScreen.Bounds()
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
	tp.scp.Settings.Channels[chIdx].Trigger.Mv = newMv
	// For level trigger we keep LowerMv independent as requested by the user.

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

func (tp *complexTriggerPointViewer) dragged(dx, dy, x, y float32) {
	if tp.scp.inStreamMode() {
		return
	}
	if tp.selectedChannel != -1 {
		tp.setDispOffset(dx, x, y, tp.selectedChannel)
	}
}

func (tp *complexTriggerPointViewer) scrolled(delta, x, y float32) {
}

func (tp *complexTriggerPointViewer) draw() {
	if tp.scp.controlTab.SelectedIndex() == dftTabIndex || tp.scp.inStreamMode() {
		return
	}
	
	tp.channelRects = make(map[int]image.Rectangle)
	
	for i, ch := range tp.scp.Settings.Channels {
		chCfg := ch.Trigger
		if chCfg.Condition != genericps.CondDontCare && ch.Enabled {
			x, y := tp.timeMv2xy(chCfg.Mv, i)
			bound := tp.scp.ftScopeSignalScreen.Bounds()
			maxY := float32(bound.Max.Y)
			minY := float32(bound.Min.Y)
			switch {
			case y > maxY:
				y = maxY
			case y < minY:
				y = minY
			}
			
			rectSize := float32(triggerPointR * 2)
			rect := image.Rect(
				int(math.Round(float64(x-rectSize))),
				int(math.Round(float64(y-rectSize))),
				int(math.Round(float64(x+rectSize))),
				int(math.Round(float64(y+rectSize))),
			)
			tp.channelRects[i] = rect
			
			col := tp.scp.Settings.Channels[i].Col[tp.scp.Settings.ChannelColorIndex]
			if tp.selectedChannel == i || tp.hoveredChannel == i {
				// Highlight outline if hovered or selected
				drawCircle(tp.scp.ftScopeSignalScreen, x, y, triggerPointR, theme.SelectionColor())
				drawCircle(tp.scp.ftScopeSignalScreen, x, y, triggerPointR-2, col)
			} else {
				drawCircle(tp.scp.ftScopeSignalScreen, x, y, triggerPointR, col)
			}
		}
	}
}

func (tp *complexTriggerPointViewer) clear() {
}
