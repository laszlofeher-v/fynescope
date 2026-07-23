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
	ftVChannelLabelViewer struct {
		rasterPartition
		chLabelRect   image.Rectangle
		vChannelIndex int
		leftLabel     bool
		selected      bool
		scp           *ScpDesc
		isTimeZoom    bool
		
		displayOffsetFraction float64
	}
)

func (cl *ftVChannelLabelViewer) raster() *screenRaster {
	if cl.isTimeZoom {
		return cl.scp.timeZoomRaster
	}
	return cl.scp.ftRaster
}

var (
	_ mouser     = (*ftVChannelLabelViewer)(nil)
	_ dragger    = (*ftVChannelLabelViewer)(nil)
	_ scroller   = (*ftVChannelLabelViewer)(nil)
	_ keyer      = (*ftVChannelLabelViewer)(nil)
	_ drawer     = (*ftVChannelLabelViewer)(nil)
	_ cursorable = (*ftVChannelLabelViewer)(nil)
)

func (cl *ftVChannelLabelViewer) typedKey(x, y float32, keyName fyne.KeyName) {
	switch keyName {
	case fyne.KeyDown:
		cl.scrolled(-scrollDelta, x, y)
	case fyne.KeyUp:
		cl.scrolled(scrollDelta, x, y)
	}
}

func newFtVChannelLabelViewer(img rasterImage, imgRect image.Rectangle, vChannelIndex int,
	scopeSignalScreen image.Rectangle, leftLabel bool, scp *ScpDesc, isTimeZoom bool) ftVChannelLabelViewer {
	cl := ftVChannelLabelViewer{rasterPartition: rasterPartition{img: img,
		imgRect: imgRect, refreshFlag: true},
		chLabelRect: scopeSignalScreen, vChannelIndex: vChannelIndex, leftLabel: leftLabel, scp: scp, isTimeZoom: isTimeZoom}
	return cl
}

func (cl *ftVChannelLabelViewer) cursor(x, y float32) (desktop.Cursor, bool) {
	if cl.mouseIn(x, y) {
		return desktop.PointerCursor, true
	}
	return desktop.DefaultCursor, false
}

func (cl *ftVChannelLabelViewer) mouseMoved(x, y float32) {
}

func (cl *ftVChannelLabelViewer) mouseIn(x, y float32) bool {
	p := image.Point{X: int(math.Round(float64(x))), Y: int(math.Round(float64(y)))}
	if p.In(cl.rect()) {
		return true
	}
	return false
}

func (cl *ftVChannelLabelViewer) mouseDown(button desktop.MouseButton, modifier fyne.KeyModifier, x, y float32) {
	if button == desktop.MouseButtonSecondary && cl.mouseIn(x, y) {
		vch := &cl.scp.Settings.VirtualChannels[cl.vChannelIndex]
		if vch.Enabled {
			vch.DisplayVOffset = 0
			cl.displayOffsetFraction = 0
			cl.enableRefresh()
			cl.scp.clearAllFtPersistentLayers()
			cl.scp.refreshRasters()
		}
	} else {
		cl.selected = cl.mouseIn(x, y)
	}
}

func (cl *ftVChannelLabelViewer) mouseUp(button desktop.MouseButton, modifier fyne.KeyModifier, x, y float32) {
	cl.selected = false
}

func (cl *ftVChannelLabelViewer) setChDispYOffset(dy, x, y float64, scroll bool) {
	p := image.Point{X: int(x), Y: int(y)}
	vch := &cl.scp.Settings.VirtualChannels[cl.vChannelIndex]
	h := float64(cl.img.Bounds().Dy())
	if vch.Enabled {
		bounds := cl.rect()
		if p.In(bounds) {
			if scroll {
				cl.displayOffsetFraction = dy + cl.scp.offsetNToFtY(vch.DisplayVOffset)
			} else {
				cl.displayOffsetFraction += dy
			}
			if cl.displayOffsetFraction < -h {
				cl.displayOffsetFraction = -h
			}
			if cl.displayOffsetFraction > h {
				cl.displayOffsetFraction = h
			}
			vch.DisplayVOffset = cl.scp.snapYToFtN(cl.displayOffsetFraction)

			cl.enableRefresh()
			cl.scp.clearAllFtPersistentLayers()
			cl.scp.refreshRasters()
		}
	}
}

func (cl *ftVChannelLabelViewer) dragged(dx, dy, x, y float32) {
	if cl.selected {
		cl.setChDispYOffset(float64(dy), float64(x), float64(y), false)
	}
}

func (cl *ftVChannelLabelViewer) scrolled(delta, x, y float32) {
	nY := (float64(cl.img.Bounds().Dy()) / float64(numberOfDivs)) / 10
	cl.setChDispYOffset(float64(-delta)*nY, float64(x), float64(y), true)
}

func (cl *ftVChannelLabelViewer) draw() {
	if !cl.refreshFlag {
		return
	}
	cl.clear()
	vch := &cl.scp.Settings.VirtualChannels[cl.vChannelIndex]
	if !vch.Enabled {
		cl.disableRefresh()
		return
	}

	xBounds := cl.rect()
	yBounds := cl.chLabelRect.Bounds()
	x := float64(xBounds.Max.X)

	startValue := genericps.RangeValuesMv[vch.VRange]
	var unitName string
	if startValue >= 1000.0 {
		startValue = startValue / 1000.0
		unitName = "V"
	} else {
		unitName = "mV"
	}
	left, _, right, _ := cl.scp.boundString(unitName)
	dv := startValue / 5.0
	dy := float32(yBounds.Dy()-1.0) / 10.0
	xoffset := left - right
	if !cl.leftLabel {
		xoffset = -float32(xBounds.Dx())
	}
	yOffset := cl.scp.offsetNToFtY(vch.DisplayVOffset)
	if yOffset >= 0 {
		cl.scp.addLabel(cl.rasterPartition.img, int(math.Round(x+float64(xoffset))),
			int(math.Round(float64(cl.scp.ftDivsY[0])+yOffset+float64(dy+fontSize)/2)),
			unitName, vch.Col[cl.scp.Settings.ChannelColorIndex])
	} else {
		cl.scp.addLabel(cl.rasterPartition.img, int(math.Round(x+float64(xoffset))),
			int(math.Round(float64(cl.scp.ftDivsY[len(cl.scp.ftDivsY)-1])+yOffset-float64(dy-fontSize)/2)),
			unitName, vch.Col[cl.scp.Settings.ChannelColorIndex])
	}
	v := startValue
	for _, y_px := range cl.scp.ftDivsY {
		if float64(y_px)+yOffset > float64(yBounds.Max.Y) {
			break
		}
		vstr := strconv.FormatFloat(float64(v), 'f', 1, 64)
		left, top, right, bottom := cl.scp.boundString(vstr)
		xoffset := left - right - 1
		if !cl.leftLabel {
			xoffset = -float32(xBounds.Dx())
		}
		cl.scp.addLabel(cl.rasterPartition.img, int(math.Round(x+float64(xoffset))),
			int(math.Round(float64(y_px)+yOffset-float64(top-bottom)/2)-1), vstr,
			vch.Col[cl.scp.Settings.ChannelColorIndex])
		v = v - dv
	}
	cl.disableRefresh()
}

func (cl *ftVChannelLabelViewer) clear() {
	draw.Draw(cl.img, cl.rect(), &image.Uniform{theme.BackgroundColor()},
		image.ZP, draw.Src)
}
