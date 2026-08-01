package gui

import (
	"fynescope/genericps"
	"fynescope/settings"
	"image"
	"image/draw"
	"math"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
)

type (
	dftVChannelLabelViewer struct {
		rasterPartition
		chLabelRect           image.Rectangle
		vChannelIndex         int
		leftLabel             bool
		selected              bool
		scp                   *ScpDesc
		displayOffsetFraction float64
	}
)

var (
	_ mouser     = (*dftVChannelLabelViewer)(nil)
	_ dragger    = (*dftVChannelLabelViewer)(nil)
	_ scroller   = (*dftVChannelLabelViewer)(nil)
	_ keyer      = (*dftVChannelLabelViewer)(nil)
	_ drawer     = (*dftVChannelLabelViewer)(nil)
	_ cursorable = (*dftVChannelLabelViewer)(nil)
)

func (cl *dftVChannelLabelViewer) typedKey(x, y float32, keyName fyne.KeyName) {
	switch keyName {
	case fyne.KeyDown:
		cl.scrolled(-scrollDelta, x, y)
	case fyne.KeyUp:
		cl.scrolled(scrollDelta, x, y)
	}
}

func newDftVChannelLabelViewer(img rasterImage, imgRect image.Rectangle, vChannelIndex int,
	scopeSignalScreen image.Rectangle, scp *ScpDesc) dftVChannelLabelViewer {
	cl := dftVChannelLabelViewer{rasterPartition: rasterPartition{img: img,
		imgRect: imgRect, refreshFlag: true},
		chLabelRect: scopeSignalScreen, vChannelIndex: vChannelIndex, scp: scp}
	return cl
}

func (cl *dftVChannelLabelViewer) cursor(x, y float32) (desktop.Cursor, bool) {
	if cl.mouseIn(x, y) {
		return desktop.PointerCursor, true
	}
	return desktop.DefaultCursor, false
}

func (cl *dftVChannelLabelViewer) mouseMoved(x, y float32) {
}

func (cl *dftVChannelLabelViewer) mouseIn(x, y float32) bool {
	p := image.Point{X: int(math.Round(float64(x))), Y: int(math.Round(float64(y)))}
	if p.In(cl.rect()) {
		return true
	}
	return false
}

func (cl *dftVChannelLabelViewer) mouseDown(button desktop.MouseButton, modifier fyne.KeyModifier, x, y float32) {
	if button == desktop.MouseButtonSecondary && cl.mouseIn(x, y) {
		channel := &cl.scp.Settings.VirtualChannels[cl.vChannelIndex]
		if channel.Enabled {
			cl.displayOffsetFraction = 0
			cl.scp.Settings.VirtualChannels[cl.vChannelIndex].DftDisplayVOffset = 0
			
			cl.enableRefresh()
			
			cl.scp.clearAllDftPersistentLayers()
			cl.scp.refreshRasters()
		}
	} else {
		cl.selected = cl.mouseIn(x, y)
	}
}

func (cl *dftVChannelLabelViewer) mouseUp(button desktop.MouseButton, modifier fyne.KeyModifier, x, y float32) {
	cl.selected = false
}

func (cl *dftVChannelLabelViewer) setChDispYOffset(dy, x, y float64, scroll bool) {
	p := image.Point{X: int(x), Y: int(y)}
	h := float64(cl.img.Bounds().Dy())
	channel := &cl.scp.Settings.VirtualChannels[cl.vChannelIndex]
	if channel.Enabled {
		bounds := cl.rect()
		if p.In(bounds) {
			if scroll {
				cl.displayOffsetFraction = dy + cl.scp.offsetNToDftY(channel.DftDisplayVOffset)
			} else {
				cl.displayOffsetFraction += dy
			}
			if cl.displayOffsetFraction < -h {
				cl.displayOffsetFraction = -h
			}
			if cl.displayOffsetFraction > h {
				cl.displayOffsetFraction = h
			}
			cl.scp.Settings.VirtualChannels[cl.vChannelIndex].DftDisplayVOffset = cl.scp.snapYToDftN(cl.displayOffsetFraction)
			cl.enableRefresh()
			cl.scp.clearAllDftPersistentLayers()
			cl.scp.refreshRasters()
		}
	}
}

func (cl *dftVChannelLabelViewer) dragged(dx, dy, x, y float32) {
	if cl.selected {
		cl.setChDispYOffset(float64(dy), float64(x), float64(y), false)
	}
}

func (cl *dftVChannelLabelViewer) scrolled(delta, x, y float32) {
	nY := float64(cl.img.Bounds().Dy()) / float64(numberOfDivs)
	if delta < 0 {
		cl.setChDispYOffset(nY, float64(x), float64(y), true)
	} else if delta > 0 {
		cl.setChDispYOffset(-nY, float64(x), float64(y), true)
	}
}

func (cl *dftVChannelLabelViewer) draw() {
	if !cl.refreshFlag {
		return
	}
	cl.clear()
	channel := &cl.scp.Settings.VirtualChannels[cl.vChannelIndex]

	xBounds := cl.rect()
	yBounds := cl.chLabelRect.Bounds()
	x := float64(xBounds.Max.X)

	// DFT labels (dB or Voltage)
	if cl.scp.shouldDrawRaster(dftTabIndex) {
		if !channel.Enabled {
			cl.disableRefresh()
			return
		}

		yOffset := cl.scp.offsetNToDftY(channel.DftDisplayVOffset)
		maxY := float64(yBounds.Max.Y)
		minY := float64(yBounds.Min.Y)

		if cl.scp.Settings.Dft.DisplayMode != settings.ModeVoltage {
			unitName := cl.scp.Settings.Dft.DisplayMode
			if unitName == settings.ModeArbitraryDB {
				unitName = "dB"
			}
			left, _, right, _ := cl.scp.boundString(unitName)
			dy := float32(yBounds.Dy()-1.0) / 10.0
			xoffset := left - right
			if !cl.leftLabel {
				xoffset = -float32(xBounds.Dx())
			}

			// Draw unit name "dB"
			cl.scp.addLabel(cl.rasterPartition.img, int(math.Round(x+float64(xoffset))),
				int(math.Round(float64(cl.scp.dftDivsY[0])+yOffset+float64(dy+fontSize)/2)),
				unitName, channel.Col[cl.scp.Settings.ChannelColorIndex])

			for i, y := range cl.scp.dftDivsY {
				yo := float64(y) + yOffset
				if yo > maxY || yo < minY {
					continue
				}
				v := float64(i) * -10.0
				vstr := strconv.FormatFloat(v, 'f', 0, 64)
				left, top, right, bottom := cl.scp.boundString(vstr)
				xoffset := left - right - 1
				if !cl.leftLabel {
					xoffset = -float32(xBounds.Dx())
				}
				cl.scp.addLabel(cl.rasterPartition.img, int(math.Round(x+float64(xoffset))),
					int(math.Round(float64(y)+yOffset-float64(top-bottom)/2)-1), vstr,
					channel.Col[cl.scp.Settings.ChannelColorIndex])
			}
		} else {
			// Voltage mode for DFT
			unitName := "V"
			maxV := genericps.RangeValuesMv[channel.VRange]
			if maxV < 1000.0 {
				unitName = "mV"
			} else {
				maxV /= 1000.0
			}

			left, _, right, _ := cl.scp.boundString(unitName)
			dy := float32(yBounds.Dy()-1.0) / 10.0
			xoffset := left - right
			if !cl.leftLabel {
				xoffset = -float32(xBounds.Dx())
			}

			// Draw unit name
			cl.scp.addLabel(cl.rasterPartition.img, int(math.Round(x+float64(xoffset))),
				int(math.Round(float64(cl.scp.dftDivsY[0])+yOffset+float64(dy+fontSize)/2)),
				unitName, channel.Col[cl.scp.Settings.ChannelColorIndex])

			dv := maxV / 10.0
			for i, y := range cl.scp.dftDivsY {
				yo := float64(y) + yOffset
				if yo > maxY || yo < minY {
					continue
				}
				v := maxV - float64(i)*dv
				vstr := strconv.FormatFloat(v, 'f', 1, 64)
				left, top, right, bottom := cl.scp.boundString(vstr)
				xoffset := left - right - 1
				if !cl.leftLabel {
					xoffset = -float32(xBounds.Dx())
				}
				cl.scp.addLabel(cl.rasterPartition.img, int(math.Round(x+float64(xoffset))),
					int(math.Round(float64(y)+yOffset-float64(top-bottom)/2)-1), vstr,
					channel.Col[cl.scp.Settings.ChannelColorIndex])
			}
		}
	}
	cl.disableRefresh()
}

func (cl *dftVChannelLabelViewer) clear() {
	draw.Draw(cl.img, cl.rect(), &image.Uniform{theme.BackgroundColor()},
		image.ZP, draw.Src)
}
