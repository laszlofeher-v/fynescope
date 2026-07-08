package gui

import (
	"fynescope/genericps"
	"fynescope/settings"
	"image"
	"log/slog"
	"math"

	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
)

const triggerPointR = 8

type (
	triggerPointViewer struct {
		rasterPartition
		scp      *ScpDesc
		selected bool
		mouseAt  bool
	}
)

var (
	_ mouser     = (*triggerPointViewer)(nil)
	_ dragger    = (*triggerPointViewer)(nil)
	_ scroller   = (*triggerPointViewer)(nil)
	_ drawer     = (*triggerPointViewer)(nil)
	_ cursorable = (*triggerPointViewer)(nil)
)

func (tp *triggerPointViewer) cursor(x, y float32) (desktop.Cursor, bool) {
	if tp.mouseIn(x, y) || tp.selected {
		return desktop.PointerCursor, true
	}
	return desktop.DefaultCursor, false
}

func (tp *triggerPointViewer) mouseIn(x, y float32) bool {
	if tp.scp.inStreamMode() {
		return false
	}
	p := image.Point{X: int(math.Round(float64(x))), Y: int(math.Round(float64(y)))}
	if p.In(tp.rect()) {
		return true
	}
	return false
}

func (tp *triggerPointViewer) mouseMoved(x, y float32) {
	prev := tp.mouseAt
	if tp.mouseIn(x, y) {
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

func (tp *triggerPointViewer) mouseDown(button desktop.MouseButton, x, y float32) {
	if tp.scp.inStreamMode() {
		return
	}
	tp.selected = tp.mouseIn(x, y)
	slog.Debug("mouseDown", "tp.selected", tp.selected)
}
func (tp *triggerPointViewer) mouseUp(button desktop.MouseButton, x, y float32) {
	tp.selected = false
}

func (scp *ScpDesc) setTriggerTime(xOffset float64) {
	newOffset := float64(xOffset)
	if scp.triggerSettingMsg.XOffset != newOffset {
		scp.triggerSettingMsg.XOffset = newOffset
		scp.psControl.SetTriggerCh <- &scp.triggerSettingMsg
		<-scp.triggerSettingMsg.Done
	}
}

func (tp *triggerPointViewer) y2mv(y float64) (mv float64) {
	bounds := tp.scp.ftScopeSignalScreen.Bounds()
	zeroOffset := float64(bounds.Min.Y + bounds.Dy()/2)
	h := float64(bounds.Dy())
	channel := &tp.scp.Settings.Channels[tp.scp.triggerSource]
	channelViewer := &tp.scp.channelViewers[tp.scp.triggerSource]
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

func (tp *triggerPointViewer) timeMv2xy(mv int32) (x, y float32) {
	bounds := tp.scp.ftScopeSignalScreen.Bounds()
	zeroOffset := float64(bounds.Min.Y + bounds.Dy()/2)
	h := float64(bounds.Dy())
	channel := &tp.scp.Settings.Channels[tp.scp.triggerSource]
	channelViewer := &tp.scp.channelViewers[tp.scp.triggerSource]
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

func (tp *triggerPointViewer) setDispOffset(dx, x, y float32) {
	bounds := tp.scp.ftScopeSignalScreen.Bounds()        // if new position is outside
	if int(x) < bounds.Min.X || int(x) > bounds.Max.X || // then return
		int(y) < bounds.Min.Y || int(y) > bounds.Max.Y {
		return
	}
	mv := tp.y2mv(float64(y)) // new trigger value
	channel := &tp.scp.Settings.Channels[tp.scp.triggerSource]
	bound := float64(genericps.InputRanges[channel.VRange])
	if mv < -bound || mv > bound {
		return
	}
	tp.scp.addFtXOffset(float64(dx))
	tp.scp.setTriggerTime(tp.scp.Settings.Time.TriggerTimeOffset)
	newMv := int32(math.Round(float64(mv)))
	if tp.scp.Settings.Trigger.Type == settings.TriggerTypeWindow {
		if newMv < channel.Trigger.LowerMv {
			newMv = channel.Trigger.LowerMv
		}
	}
	channel.Trigger.Mv = newMv
	tp.scp.triggerSettingMsg.TriggerADC = int16(tp.scp.mvToAdc(channel.Trigger.Mv, channel.VRange))
	tp.scp.triggerSettingMsg.Mv = channel.Trigger.Mv
	tp.scp.psControl.SetTriggerCh <- &tp.scp.triggerSettingMsg
	<-tp.scp.triggerSettingMsg.Done
	lw := tp.scp.ftBottomLabelViewer.(*timeLabelViewer)
	tp.scp.clearAllFtPersistentLayers()
	tp.scp.clearAllDftPersistentLayers()
	lw.enableRefresh()
	tp.enableRefresh()
	slog.Debug("setDispOffset")
	// tp.scp.psControl.Restart()
	if tp.scp.ftRaster != nil {
		tp.scp.ftRaster.Refresh()
	}
}
func (tp *triggerPointViewer) dragged(dx, dy, x, y float32) {
	if tp.scp.triggerSource < 0 || int(tp.scp.triggerSource) >= len(tp.scp.Settings.Channels) {
		return
	}
	if tp.selected {
		tp.setDispOffset(dx, x, y)
	}
}

func (tp *triggerPointViewer) scrolled(delta, x, y float32) {
}

func (tp *triggerPointViewer) draw() {
	if tp.scp.controlTab.SelectedIndex() == dftTabIndex || tp.scp.inStreamMode() {
		return
	}
	// slog.Debug("triggerPointViewer draw 1")
	if tp.scp.triggerSource != dontCare {
		// slog.Debug("triggerPointViewer draw 2")
		channel := &tp.scp.Settings.Channels[tp.scp.triggerSource]
		if channel.TriggerSource {
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
			rectSize := float32(triggerPointR * 2)
			tp.imgRect = image.Rect(int(math.Round(float64(x-rectSize))),
				int(math.Round(float64(y-rectSize))),
				int(math.Round(float64(x+rectSize))),
				int(math.Round(float64(y+rectSize))))
			drawCircle(tp.scp.ftScopeSignalScreen, x, y, triggerPointR, theme.ForegroundColor())
			if tp.scp.triggerThresholdDisp.Value != int(channel.Trigger.Mv) {
				tp.scp.triggerThresholdDisp.SilentSetValue(int(channel.Trigger.Mv))
				tp.scp.triggerThresholdDisp.Refresh()
				// if tp.scp.psControl.TriggerSetting.Mode == control.ETS {
				// tp.scp.psControl.Restart()
				// }
			}
		}
	}
}
func (tp *triggerPointViewer) clear() {

}

func newTriggerPointViewer(img rasterImage, scp *ScpDesc) *triggerPointViewer {
	imgRect := image.Rect(-triggerPointR,
		-triggerPointR,
		triggerPointR,
		triggerPointR)
	tp := &triggerPointViewer{rasterPartition: rasterPartition{img: img, imgRect: imgRect}, scp: scp}
	return tp
}
