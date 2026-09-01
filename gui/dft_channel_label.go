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

func (cl *dftChannelLabelViewer) mouseDown(button desktop.MouseButton, modifier fyne.KeyModifier, x, y float32) {
	if button == desktop.MouseButtonSecondary && cl.mouseIn(x, y) {
		channel := &cl.scp.Settings.Channels[cl.channelIndex]
		if channel.Enabled {
			channelViewer := &cl.scp.channelViewers[cl.channelIndex]

			channelViewer.dftDisplayOffsetFraction = 0
			channelViewer.dftDisplayOffsetInt = 0
			cl.scp.Settings.Channels[cl.channelIndex].DftDisplayVOffset = 0

			channelViewer.label.enableRefresh()
			channelViewer.dftLabel.enableRefresh()

			cl.scp.clearAllDftPersistentLayers()
			cl.scp.refreshRasters()
		}
	} else {
		cl.selected = cl.mouseIn(x, y)
	}
}
func (cl *dftChannelLabelViewer) mouseUp(button desktop.MouseButton, modifier fyne.KeyModifier, x, y float32) {
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

			cl.scp.clearAllDftPersistentLayers()
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
	nY := float64(cl.img.Bounds().Dy()) / float64(numberOfDivs)
	if delta < 0 {
		cl.setChDispYOffset(nY, float64(x), float64(y), true)
	} else if delta > 0 {
		cl.setChDispYOffset(-nY, float64(x), float64(y), true)
	}
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

		if cl.scp.Settings.Dft.DisplayMode != settings.ModeVoltage {
			unitName := cl.scp.Settings.Dft.DisplayMode
			if unitName == settings.ModeArbitraryDB {
				unitName = "dB"
			}
			left, _, right, _ := cl.scp.boundString(unitName)
			dy := float32(yBounds.Dy()-1.0) / 10.0
			xoffset := left - right
			if !channelViewer.leftLabel {
				xoffset = -float32(xBounds.Dx())
			}

			stride := 1
			if dy < float32(fontSize)+6.0 {
				stride = 2
			}
			lastDrawnY := -100000.0
			unitDrawn := false

			unitY := float64(cl.scp.dftDivsY[0]) + yOffset - float64(fontSize)/2.0 - 2.0
			if unitY >= minY && unitY <= maxY {
				cl.scp.addLabel(cl.rasterPartition.img, int(math.Round(x+float64(xoffset))),
					int(math.Round(unitY)),
					unitName, channel.Col[cl.scp.Settings.ChannelColorIndex])
				lastDrawnY = unitY
				unitDrawn = true
			}

			for i, y := range cl.scp.dftDivsY {
				yo := float64(y) + yOffset
				if yo > maxY || yo < minY {
					continue
				}
				if i%stride != 0 && i != 0 && i != len(cl.scp.dftDivsY)-1 {
					continue
				}
				v := float64(i) * -10.0
				vstr := strconv.FormatFloat(v, 'f', 0, 64)
				if !unitDrawn {
					vstr += " " + unitName
					unitDrawn = true
				}
				left, top, right, bottom := cl.scp.boundString(vstr)
				drawY := float64(y) + yOffset - float64(top-bottom)/2.0 - 1.0
				if math.Abs(drawY-lastDrawnY) < float64(fontSize)-1.0 {
					continue
				}
				xoffset := left - right - 1
				if !channelViewer.leftLabel {
					xoffset = -float32(xBounds.Dx())
				}
				cl.scp.addLabel(cl.rasterPartition.img, int(math.Round(x+float64(xoffset))),
					int(math.Round(drawY)), vstr,
					channel.Col[cl.scp.Settings.ChannelColorIndex])
				lastDrawnY = drawY
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

			stride := 1
			if dy < float32(fontSize)+6.0 {
				stride = 2
			}
			lastDrawnY := -100000.0
			unitDrawn := false

			unitY := float64(cl.scp.dftDivsY[0]) + yOffset - float64(fontSize)/2.0 - 2.0
			if unitY >= minY && unitY <= maxY {
				cl.scp.addLabel(cl.rasterPartition.img, int(math.Round(x+float64(xoffset))),
					int(math.Round(unitY)),
					unitName, channel.Col[cl.scp.Settings.ChannelColorIndex])
				lastDrawnY = unitY
				unitDrawn = true
			}

			dv := maxV / 10.0
			for i, y := range cl.scp.dftDivsY {
				yo := float64(y) + yOffset
				if yo > maxY || yo < minY {
					continue
				}
				if i%stride != 0 && i != 0 && i != len(cl.scp.dftDivsY)-1 {
					continue
				}
				v := maxV - float64(i)*dv
				vstr := strconv.FormatFloat(v, 'f', 1, 64)
				if !unitDrawn {
					vstr += " " + unitName
					unitDrawn = true
				}
				left, top, right, bottom := cl.scp.boundString(vstr)
				drawY := float64(y) + yOffset - float64(top-bottom)/2.0 - 1.0
				if math.Abs(drawY-lastDrawnY) < float64(fontSize)-1.0 {
					continue
				}
				xoffset := left - right - 1
				if !channelViewer.leftLabel {
					xoffset = -float32(xBounds.Dx())
				}
				cl.scp.addLabel(cl.rasterPartition.img, int(math.Round(x+float64(xoffset))),
					int(math.Round(drawY)), vstr,
					channel.Col[cl.scp.Settings.ChannelColorIndex])
				lastDrawnY = drawY
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
	effectiveH := float32(yBounds.Dy())
	dy := (effectiveH - 1.0) / 10.0
	xoffset := left - right
	if !channelViewer.leftLabel {
		xoffset = -float32(xBounds.Dx())
	}
	yOffset := cl.scp.offsetNToFtY(channelViewer.displayOffsetInt)

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
			unitName, channel.Col[cl.scp.Settings.ChannelColorIndex])
		lastDrawnY = unitY
		unitDrawn = true
	}

	v := startValue
	for i, y := range cl.scp.ftDivsY {
		drawY := float64(y) + yOffset
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
		if !channelViewer.leftLabel {
			xoffset = -float32(xBounds.Dx())
		}
		cl.scp.addLabel(cl.rasterPartition.img, int(math.Round(x+float64(xoffset))),
			int(math.Round(lblDrawY)), vstr,
			channel.Col[cl.scp.Settings.ChannelColorIndex])
		lastDrawnY = lblDrawY
		v = v - dv
	}
	cl.disableRefresh()
}

func (cl *dftChannelLabelViewer) clear() {
	draw.Draw(cl.img, cl.rect(), &image.Uniform{theme.BackgroundColor()},
		image.ZP, draw.Src)
}
