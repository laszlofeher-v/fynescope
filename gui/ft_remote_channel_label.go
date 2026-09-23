package gui

import (
	"fynescope/genericps"
	"image"
	"image/draw"
	"math"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
)

type (
	ftRemoteChannelLabelViewer struct {
		rasterPartition
		chLabelRect   image.Rectangle
		remoteChanIdx int
		leftLabel     bool
		selected      bool
		scp           *ScpDesc
		isTimeZoom    bool

		displayOffsetFraction float64
	}
)

func (cl *ftRemoteChannelLabelViewer) raster() *screenRaster {
	if cl.isTimeZoom {
		return cl.scp.timeZoomRaster
	}
	return cl.scp.ftRaster
}

var (
	_ mouser     = (*ftRemoteChannelLabelViewer)(nil)
	_ dragger    = (*ftRemoteChannelLabelViewer)(nil)
	_ scroller   = (*ftRemoteChannelLabelViewer)(nil)
	_ keyer      = (*ftRemoteChannelLabelViewer)(nil)
	_ drawer     = (*ftRemoteChannelLabelViewer)(nil)
	_ cursorable = (*ftRemoteChannelLabelViewer)(nil)
)

func (cl *ftRemoteChannelLabelViewer) typedKey(x, y float32, keyName fyne.KeyName) {
	switch keyName {
	case fyne.KeyDown:
		cl.scrolled(-scrollDelta, x, y)
	case fyne.KeyUp:
		cl.scrolled(scrollDelta, x, y)
	}
}

func newFtRemoteChannelLabelViewer(img rasterImage, imgRect image.Rectangle, remoteChanIdx int,
	scopeSignalScreen image.Rectangle, leftLabel bool, scp *ScpDesc, isTimeZoom bool) ftRemoteChannelLabelViewer {
	cl := ftRemoteChannelLabelViewer{
		rasterPartition: rasterPartition{
			img:         img,
			imgRect:     imgRect,
			refreshFlag: true,
		},
		chLabelRect:   scopeSignalScreen,
		remoteChanIdx: remoteChanIdx,
		leftLabel:     leftLabel,
		scp:           scp,
		isTimeZoom:    isTimeZoom,
	}
	return cl
}

func (cl *ftRemoteChannelLabelViewer) cursor(x, y float32) (desktop.Cursor, bool) {
	if cl.mouseIn(x, y) {
		return desktop.PointerCursor, true
	}
	return desktop.DefaultCursor, false
}

func (cl *ftRemoteChannelLabelViewer) mouseMoved(x, y float32) {
}

func (cl *ftRemoteChannelLabelViewer) mouseIn(x, y float32) bool {
	p := image.Point{X: int(math.Round(float64(x))), Y: int(math.Round(float64(y)))}
	return p.In(cl.rect())
}

func (cl *ftRemoteChannelLabelViewer) mouseDown(button desktop.MouseButton, modifier fyne.KeyModifier, x, y float32) {
	if button == desktop.MouseButtonSecondary && cl.mouseIn(x, y) {
		cl.scp.remoteChannelsMu.Lock()
		if cl.remoteChanIdx < len(cl.scp.remoteChannels) {
			rch := &cl.scp.remoteChannels[cl.remoteChanIdx]
			if rch.Enabled {
				rch.DisplayVOffset = 0
				cl.displayOffsetFraction = 0
				cl.enableRefresh()
				cl.scp.clearAllFtPersistentLayers()
				cl.scp.refreshRasters()
			}
		}
		cl.scp.remoteChannelsMu.Unlock()
	} else {
		cl.selected = cl.mouseIn(x, y)
	}
}

func (cl *ftRemoteChannelLabelViewer) mouseUp(button desktop.MouseButton, modifier fyne.KeyModifier, x, y float32) {
	cl.selected = false
}

func (cl *ftRemoteChannelLabelViewer) setChDispYOffset(dy, x, y float64, scroll bool) {
	p := image.Point{X: int(x), Y: int(y)}
	cl.scp.remoteChannelsMu.Lock()
	defer cl.scp.remoteChannelsMu.Unlock()

	if cl.remoteChanIdx >= len(cl.scp.remoteChannels) {
		return
	}
	rch := &cl.scp.remoteChannels[cl.remoteChanIdx]
	h := float64(cl.img.Bounds().Dy())
	if rch.Enabled {
		bounds := cl.rect()
		if p.In(bounds) {
			if scroll {
				cl.displayOffsetFraction = dy + cl.scp.offsetNToFtY(rch.DisplayVOffset)
			} else {
				cl.displayOffsetFraction += dy
			}
			if cl.displayOffsetFraction < -h {
				cl.displayOffsetFraction = -h
			}
			if cl.displayOffsetFraction > h {
				cl.displayOffsetFraction = h
			}
			rch.DisplayVOffset = cl.scp.snapYToFtN(cl.displayOffsetFraction)

			cl.enableRefresh()
			cl.scp.clearAllFtPersistentLayers()
			cl.scp.refreshRasters()
		}
	}
}

func (cl *ftRemoteChannelLabelViewer) dragged(dx, dy, x, y float32) {
	if cl.selected {
		cl.setChDispYOffset(float64(dy), float64(x), float64(y), false)
	}
}

func (cl *ftRemoteChannelLabelViewer) scrolled(delta, x, y float32) {
	nY := float64(cl.img.Bounds().Dy()) / float64(numberOfDivs)
	if delta < 0 {
		cl.setChDispYOffset(nY, float64(x), float64(y), true)
	} else if delta > 0 {
		cl.setChDispYOffset(-nY, float64(x), float64(y), true)
	}
}

func (cl *ftRemoteChannelLabelViewer) draw() {
	if !cl.refreshFlag {
		return
	}
	cl.clear()

	cl.scp.remoteChannelsMu.RLock()
	if cl.remoteChanIdx >= len(cl.scp.remoteChannels) {
		cl.scp.remoteChannelsMu.RUnlock()
		cl.disableRefresh()
		return
	}
	rch := cl.scp.remoteChannels[cl.remoteChanIdx]
	cl.scp.remoteChannelsMu.RUnlock()

	if !rch.Enabled {
		cl.disableRefresh()
		return
	}

	xBounds := cl.rect()
	yBounds := cl.chLabelRect.Bounds()
	x := float64(xBounds.Max.X)

	startValue := genericps.RangeValuesMv[rch.VRange]
	var unitName string
	if startValue >= 1000.0 {
		startValue = startValue / 1000.0
		unitName = "V"
	} else {
		unitName = "mV"
	}
	left, _, right, _ := cl.scp.boundString(unitName)
	dv := startValue / 5.0
	effectiveH := float32(yBounds.Dy())
	dy := (effectiveH - 1.0) / 10.0
	xoffset := left - right
	if !cl.leftLabel {
		xoffset = -float32(xBounds.Dx())
	}
	yOffset := cl.scp.offsetNToFtY(rch.DisplayVOffset)
	stride := 1
	if dy < float32(fontSize)+6.0 {
		stride = 2
	}
	lastDrawnY := -100000.0
	unitDrawn := false

	unitY := float64(cl.scp.ftDivsY[0]) + yOffset - float64(fontSize)/2.0 - 2.0
	if yOffset < 0 {
		unitY = float64(cl.scp.ftDivsY[len(cl.scp.ftDivsY)-1]) + yOffset + float64(fontSize)*1.2
	}
	if unitY >= float64(yBounds.Min.Y) && unitY <= float64(yBounds.Max.Y) {
		cl.scp.addLabel(cl.rasterPartition.img, int(math.Round(x+float64(xoffset))),
			int(math.Round(unitY)),
			unitName, rch.Color)
		lastDrawnY = unitY
		unitDrawn = true
	}

	v := startValue
	for i, y_px := range cl.scp.ftDivsY {
		if y_px < 0 {
			v = v - dv
			continue
		}
		drawY := float64(y_px) + yOffset
		if drawY > float64(yBounds.Max.Y) || drawY < float64(yBounds.Min.Y) {
			v = v - dv
			continue
		}
		if i%stride != 0 && i != 0 && i != len(cl.scp.ftDivsY)-1 {
			v = v - dv
			continue
		}
		vstr := strconv.FormatFloat(float64(v), 'f', 1, 64)
		if !unitDrawn {
			vstr += " " + unitName
			unitDrawn = true
		}
		left, top, right, bottom := cl.scp.boundString(vstr)
		lblDrawY := drawY - float64(top-bottom)/2.0 - 1.0
		if math.Abs(lblDrawY-lastDrawnY) < float64(fontSize)-1.0 {
			v = v - dv
			continue
		}
		xoffset := left - right - 1
		if !cl.leftLabel {
			xoffset = -float32(xBounds.Dx())
		}
		cl.scp.addLabel(cl.rasterPartition.img, int(math.Round(x+float64(xoffset))),
			int(math.Round(lblDrawY)), vstr,
			rch.Color)
		lastDrawnY = lblDrawY
		v = v - dv
	}
	cl.disableRefresh()
}

func (cl *ftRemoteChannelLabelViewer) clear() {
	draw.Draw(cl.img, cl.rect(), &image.Uniform{theme.BackgroundColor()},
		image.ZP, draw.Src)
}
