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
	dftChannelLabelViewer struct {
		rasterPartition
		chLabelRect  image.Rectangle
		channelIndex int
		selected     bool
		scp          *ScpDesc
	}
)

var (
	_ mouser     = (*dftChannelLabelViewer)(nil)
	_ dragger    = (*dftChannelLabelViewer)(nil)
	_ scroller   = (*dftChannelLabelViewer)(nil)
	_ keyer      = (*dftChannelLabelViewer)(nil)
	_ drawer     = (*dftChannelLabelViewer)(nil)
	_ cursorable = (*dftChannelLabelViewer)(nil)
)

func (cl *dftChannelLabelViewer) typedKey(x, y float32, keyName fyne.KeyName) {
	switch keyName {
	case fyne.KeyDown:
		cl.scrolled(-scrollDelta, x, y)
	case fyne.KeyUp:
		cl.scrolled(scrollDelta, x, y)
	}
}

func newDftChannelLabelViewer(img rasterImage, imgRect image.Rectangle, channelIndex int,
	scopeSignalScreen image.Rectangle, scp *ScpDesc) dftChannelLabelViewer {
	cl := dftChannelLabelViewer{rasterPartition: rasterPartition{img: img,
		imgRect: imgRect, refreshFlag: true},
		chLabelRect: scopeSignalScreen, channelIndex: channelIndex, scp: scp}
	return cl
}

func (cl *dftChannelLabelViewer) cursor(x, y float32) (desktop.Cursor, bool) {
	if cl.mouseIn(x, y) {
		return desktop.PointerCursor, true
	}
	return desktop.DefaultCursor, false
}

func (cl *dftChannelLabelViewer) mouseMoved(x, y float32) {
}

func (cl *dftChannelLabelViewer) mouseIn(x, y float32) bool {
	p := image.Point{X: int(math.Round(float64(x))), Y: int(math.Round(float64(y)))}
	if p.In(cl.rect()) {
		return true
	}
	return false
}

func (cl *dftChannelLabelViewer) mouseDown(button desktop.MouseButton, x, y float32) {
	cl.selected = cl.mouseIn(x, y)
}
func (cl *dftChannelLabelViewer) mouseUp(button desktop.MouseButton, x, y float32) {
	cl.selected = false
	// tl.scp.dispTimeOffsetRatio = tl.scp.xOffsetRatio(tl.scp.dispTimeOffsetAbsolute)
}

func (cl *dftChannelLabelViewer) setChDispYOffset(dy, x, y float64, scroll bool) {
	p := image.Point{X: int(x), Y: int(y)}
	h := float64(cl.img.Bounds().Dy())
	channel := &cl.scp.Settings.Channels[cl.channelIndex]
	channelViewer := &cl.scp.channelViewers[cl.channelIndex]
	if channel.Enabled {
		// Use cl.rect() instead of hardcoded channelViewer.label.imgRect
		bounds := cl.rect()
		if p.In(bounds) {
			if scroll {
				channelViewer.dftDisplayOffsetFraction = dy +
					cl.scp.offsetNToDftY(channelViewer.dftDisplayOffsetInt)
			} else {
				channelViewer.dftDisplayOffsetFraction += dy
			}
			if channelViewer.dftDisplayOffsetFraction < -h {
				channelViewer.dftDisplayOffsetFraction = -h
			}
			if channelViewer.dftDisplayOffsetFraction > h {
				channelViewer.dftDisplayOffsetFraction = h
			}
			channelViewer.dftDisplayOffsetInt =
				cl.scp.snapYToDftN(channelViewer.dftDisplayOffsetFraction)
			cl.scp.Settings.Channels[cl.channelIndex].DftDisplayVOffset =
				channelViewer.dftDisplayOffsetInt

			channelViewer.label.enableRefresh()
			channelViewer.dftLabel.enableRefresh()

			cl.scp.clearDftPersistentLayer(genericps.ChannelId(cl.channelIndex))
			cl.scp.refreshRasters()
		}
	}
}
func (cl *dftChannelLabelViewer) dragged(dx, dy, x, y float32) {
	if cl.selected {
		cl.setChDispYOffset(float64(dy), float64(x), float64(y), false)
	}
}

func (cl *dftChannelLabelViewer) scrolled(delta, x, y float32) {
	nY := (float64(cl.img.Bounds().Dy()) / float64(numberOfDivs)) / 10
	cl.setChDispYOffset(float64(-delta)*nY, float64(x), float64(y), true)
}

func (cl *dftChannelLabelViewer) draw() {
	if !cl.refreshFlag {
		return
	}
	cl.clear()
	channel := &cl.scp.Settings.Channels[cl.channelIndex]
	channelViewer := &cl.scp.channelViewers[cl.channelIndex]

	xBounds := cl.rect()
	yBounds := cl.chLabelRect.Bounds()
	x := float64(xBounds.Max.X)

	// DFT labels (dB or Voltage)
	if cl.scp.shouldDrawRaster(dftTabIndex) {
		if !channel.Enabled {
			cl.disableRefresh()
			return
		}

		yOffset := cl.scp.offsetNToDftY(channelViewer.dftDisplayOffsetInt)
		maxY := float64(yBounds.Max.Y)
		minY := float64(yBounds.Min.Y)

		if cl.scp.Settings.Dft.DisplayMode == "dB" {
			unitName := "dB"
			left, _, right, _ := cl.scp.boundString(unitName)
			dy := float32(yBounds.Dy()-1.0) / 10.0
			xoffset := left - right
			if !channelViewer.leftLabel {
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
				if !channelViewer.leftLabel {
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
			if !channelViewer.leftLabel {
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
				if !channelViewer.leftLabel {
					xoffset = -float32(xBounds.Dx())
				}
				cl.scp.addLabel(cl.rasterPartition.img, int(math.Round(x+float64(xoffset))),
					int(math.Round(float64(y)+yOffset-float64(top-bottom)/2)-1), vstr,
					channel.Col[cl.scp.Settings.ChannelColorIndex])
			}
		}
		cl.disableRefresh()
		return
	}
	startValue := genericps.RangeValuesMv[channel.VRange]
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
	if !channelViewer.leftLabel {
		xoffset = -float32(xBounds.Dx())
	}
	yOffset := cl.scp.offsetNToFtY(channelViewer.displayOffsetInt)
	if yOffset >= 0 {
		cl.scp.addLabel(cl.rasterPartition.img, int(math.Round(x+float64(xoffset))),
			int(math.Round(float64(cl.scp.ftDivsY[0])+yOffset+float64(dy+fontSize)/2)),
			unitName, channel.Col[cl.scp.Settings.ChannelColorIndex])
	} else {
		cl.scp.addLabel(cl.rasterPartition.img, int(math.Round(x+float64(xoffset))),
			int(math.Round(float64(cl.scp.ftDivsY[len(cl.scp.ftDivsY)-1])+yOffset-float64(dy-fontSize)/2)),
			unitName, channel.Col[cl.scp.Settings.ChannelColorIndex])
	}
	v := startValue
	for _, y := range cl.scp.ftDivsY {
		if float64(y)+yOffset > float64(yBounds.Max.Y) {
			break
		}
		vstr := strconv.FormatFloat(float64(v), 'f', 1, 64)
		left, top, right, bottom := cl.scp.boundString(vstr)
		xoffset := left - right - 1
		if !channelViewer.leftLabel {
			xoffset = -float32(xBounds.Dx()) //+ 2
		}
		cl.scp.addLabel(cl.rasterPartition.img, int(math.Round(x+float64(xoffset))),
			int(math.Round(float64(y)+yOffset-float64(top-bottom)/2)-1), vstr,
			channel.Col[cl.scp.Settings.ChannelColorIndex])
		v = v - dv
	}
	cl.disableRefresh()
}

func (cl *dftChannelLabelViewer) clear() {
	draw.Draw(cl.img, cl.rect(), &image.Uniform{theme.BackgroundColor()},
		image.ZP, draw.Src)
}
