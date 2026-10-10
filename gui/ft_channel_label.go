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
	// ftChannelLabelViewer renders voltage scale division labels and handles interactive
	// vertical offset adjustments (dragging, scrolling, and right-click reset) for a physical
	// hardware channel on time-domain (FT, TZ) or frequency-domain (DFT) screen rasters.
	ftChannelLabelViewer struct {
		rasterPartition
		chLabelRect  image.Rectangle // Bounding rectangle of the channel label display column.
		channelIndex int             // Zero-based index of the physical channel (0=A, 1=B, etc.).
		selected     bool            // True while this channel label is being actively dragged.
		scp          *ScpDesc        // Pointer to the parent oscilloscope descriptor.
		isTimeZoom   bool            // True if this viewer is attached to the time-zoom viewport raster.
	}
)

// raster returns the target raster screen for this viewer, selecting either the
// time-zoom raster or the main time-domain (FT) raster.
func (cl *ftChannelLabelViewer) raster() *screenRaster {
	if cl.isTimeZoom {
		return cl.scp.timeZoomRaster
	}
	return cl.scp.ftRaster
}

// Compile-time assertions verifying that ftChannelLabelViewer implements the necessary
// interactive partition interfaces.
var (
	_ mouser     = (*ftChannelLabelViewer)(nil)
	_ dragger    = (*ftChannelLabelViewer)(nil)
	_ scroller   = (*ftChannelLabelViewer)(nil)
	_ keyer      = (*ftChannelLabelViewer)(nil)
	_ drawer     = (*ftChannelLabelViewer)(nil)
	_ cursorable = (*ftChannelLabelViewer)(nil)
)

// typedKey handles keyboard navigation (up and down arrow keys) to adjust the channel's vertical position.
func (cl *ftChannelLabelViewer) typedKey(x, y float32, keyName fyne.KeyName) {
	switch keyName {
	case fyne.KeyDown:
		cl.scrolled(-scrollDelta, x, y)
	case fyne.KeyUp:
		cl.scrolled(scrollDelta, x, y)
	}
}

// newFtChannelLabelViewer constructs an ftChannelLabelViewer partition for the given channel index.
func newFtChannelLabelViewer(img rasterImage, imgRect image.Rectangle, channelIndex int,
	scopeSignalScreen image.Rectangle, scp *ScpDesc, isTimeZoom bool) ftChannelLabelViewer {
	cl := ftChannelLabelViewer{
		rasterPartition: rasterPartition{
			img:         img,
			imgRect:     imgRect,
			refreshFlag: true,
		},
		chLabelRect:  scopeSignalScreen,
		channelIndex: channelIndex,
		scp:          scp,
		isTimeZoom:   isTimeZoom,
	}
	return cl
}

// cursor returns a pointer cursor when the mouse hovers over the interactive label region.
func (cl *ftChannelLabelViewer) cursor(x, y float32) (desktop.Cursor, bool) {
	if cl.mouseIn(x, y) {
		return desktop.PointerCursor, true
	}
	return desktop.DefaultCursor, false
}

// mouseMoved handles mouse hover motion events.
func (cl *ftChannelLabelViewer) mouseMoved(x, y float32) {
}

// mouseIn reports whether the coordinate (x, y) lies within the label viewer partition bounds.
func (cl *ftChannelLabelViewer) mouseIn(x, y float32) bool {
	p := image.Point{X: int(math.Round(float64(x))), Y: int(math.Round(float64(y)))}
	if p.In(cl.rect()) {
		return true
	}
	return false
}

// mouseDown handles mouse button clicks. Secondary (right) click resets the channel's vertical
// display offset back to zero across all associated views (main, time-zoom, and DFT).
// Primary (left) click begins dragging.
func (cl *ftChannelLabelViewer) mouseDown(button desktop.MouseButton, modifier fyne.KeyModifier, x, y float32) {
	if button == desktop.MouseButtonSecondary && cl.mouseIn(x, y) {
		channel := &cl.scp.Settings.Channels[cl.channelIndex]
		if channel.Enabled {
			channelViewer := &cl.scp.channelViewers[cl.channelIndex]

			channelViewer.displayOffsetFraction = 0
			channelViewer.displayOffsetInt = 0
			cl.scp.Settings.Channels[cl.channelIndex].DisplayVOffset = 0

			channelViewer.label.enableRefresh()
			if channelViewer.tzLabel != nil {
				channelViewer.tzLabel.enableRefresh()
			}
			channelViewer.dftLabel.enableRefresh()

			cl.scp.clearAllFtPersistentLayers()
			cl.scp.refreshRasters()
		}
	} else {
		cl.selected = cl.mouseIn(x, y)
	}
}

// mouseUp releases the active drag selection.
func (cl *ftChannelLabelViewer) mouseUp(button desktop.MouseButton, modifier fyne.KeyModifier, x, y float32) {
	cl.selected = false
}

// setChDispYOffset adjusts the physical channel's vertical display offset by dy pixels.
// When scroll is true, dy is treated as an incremental jump added to the snapped division offset.
// Display offset is clamped to the partition height and snapped to vertical divisions.
func (cl *ftChannelLabelViewer) setChDispYOffset(dy, x, y float64, scroll bool) {
	p := image.Point{X: int(x), Y: int(y)}
	h := float64(cl.img.Bounds().Dy())
	channel := &cl.scp.Settings.Channels[cl.channelIndex]
	channelViewer := &cl.scp.channelViewers[cl.channelIndex]
	if channel.Enabled {
		bounds := cl.rect()
		if p.In(bounds) {
			if scroll {
				channelViewer.displayOffsetFraction = dy +
					cl.scp.offsetNToFtY(channelViewer.displayOffsetInt)
			} else {
				channelViewer.displayOffsetFraction += dy
			}
			// Clamp vertical offset to screen bounds
			if channelViewer.displayOffsetFraction < -h {
				channelViewer.displayOffsetFraction = -h
			}
			if channelViewer.displayOffsetFraction > h {
				channelViewer.displayOffsetFraction = h
			}
			channelViewer.displayOffsetInt =
				cl.scp.snapYToFtN(channelViewer.displayOffsetFraction)
			cl.scp.Settings.Channels[cl.channelIndex].DisplayVOffset =
				channelViewer.displayOffsetInt

			channelViewer.label.enableRefresh()
			if channelViewer.tzLabel != nil {
				channelViewer.tzLabel.enableRefresh()
			}
			channelViewer.dftLabel.enableRefresh()

			cl.scp.clearAllFtPersistentLayers()
			cl.scp.refreshRasters()
		}
	}
}

// dragged processes mouse drag events to dynamically update the vertical display offset.
func (cl *ftChannelLabelViewer) dragged(dx, dy, x, y float32) {
	if cl.selected {
		cl.setChDispYOffset(float64(dy), float64(x), float64(y), false)
	}
}

// scrolled processes mouse wheel scrolling by shifting by one division (nY) per step.
func (cl *ftChannelLabelViewer) scrolled(delta, x, y float32) {
	nY := float64(cl.img.Bounds().Dy()) / float64(numberOfDivs)
	if delta < 0 {
		cl.setChDispYOffset(nY, float64(x), float64(y), true)
	} else if delta > 0 {
		cl.setChDispYOffset(-nY, float64(x), float64(y), true)
	}
}

// draw renders the scale labels along the vertical axis for the channel.
// When the active tab is DFT, it renders frequency-domain magnitude labels (in dB or Voltage).
// Otherwise, it renders time-domain voltage division labels (in V or mV).
func (cl *ftChannelLabelViewer) draw() {
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

		if cl.scp.Settings.Dft.DisplayUnit == "dB" {
			unitName := "dB"
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

			// Compute position for DFT unit indicator
			var unitY float64
			if yOffset >= 0 {
				unitY = float64(cl.scp.dftDivsY[0]) + yOffset + float64(dy+fontSize)/2.0
			} else {
				lastY := cl.scp.dftDivsY[0]
				for _, yVal := range cl.scp.dftDivsY {
					if yVal >= 0 {
						lastY = yVal
					}
				}
				unitY = float64(lastY) + yOffset - float64(dy-fontSize)/2.0
			}
			if unitY >= minY && unitY <= maxY {
				cl.scp.addLabel(cl.rasterPartition.img, int(math.Round(x+float64(xoffset))),
					int(math.Round(unitY)),
					unitName, channel.Col[cl.scp.Settings.ChannelColorIndex])
				unitDrawn = true
			}

			// Draw decibel values down each division (e.g. 0 dB, -10 dB, -20 dB, ...)
			for i, y := range cl.scp.dftDivsY {
				yo := float64(y) + yOffset
				if yo > maxY+float64(fontSize)/2.0 || yo < minY-float64(fontSize)/2.0 {
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
				if unitDrawn && math.Abs(drawY-unitY) < float64(fontSize)-1.0 {
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
			// Linear Voltage mode for DFT
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

			var unitY float64
			if yOffset >= 0 {
				unitY = float64(cl.scp.dftDivsY[0]) + yOffset + float64(dy+fontSize)/2.0
			} else {
				lastY := cl.scp.dftDivsY[0]
				for _, yVal := range cl.scp.dftDivsY {
					if yVal >= 0 {
						lastY = yVal
					}
				}
				unitY = float64(lastY) + yOffset - float64(dy-fontSize)/2.0
			}
			if unitY >= minY && unitY <= maxY {
				cl.scp.addLabel(cl.rasterPartition.img, int(math.Round(x+float64(xoffset))),
					int(math.Round(unitY)),
					unitName, channel.Col[cl.scp.Settings.ChannelColorIndex])
				unitDrawn = true
			}

			// Render linear voltage increments for DFT
			dv := maxV / 10.0
			for i, y := range cl.scp.dftDivsY {
				yo := float64(y) + yOffset
				if yo > maxY+float64(fontSize)/2.0 || yo < minY-float64(fontSize)/2.0 {
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
				if unitDrawn && math.Abs(drawY-unitY) < float64(fontSize)-1.0 {
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

	// Time-domain (FT / TZ) voltage mode labels
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

	var unitY float64
	if yOffset >= 0 {
		unitY = float64(cl.scp.ftDivsY[0]) + yOffset + float64(dy+fontSize)/2.0
	} else {
		lastY := cl.scp.ftDivsY[0]
		for _, yVal := range cl.scp.ftDivsY {
			if yVal >= 0 {
				lastY = yVal
			}
		}
		unitY = float64(lastY) + yOffset - float64(dy-fontSize)/2.0
	}
	if unitY >= float64(yBounds.Min.Y) && unitY <= float64(yBounds.Max.Y) {
		cl.scp.addLabel(cl.rasterPartition.img, int(math.Round(x+float64(xoffset))),
			int(math.Round(unitY)),
			unitName, channel.Col[cl.scp.Settings.ChannelColorIndex])
		unitDrawn = true
	}

	// Render voltage division values across the time-domain vertical scale
	v := startValue
	for i, y := range cl.scp.ftDivsY {
		if y < 0 {
			v = v - dv
			continue
		}
		drawY := float64(y) + yOffset
		if drawY > float64(yBounds.Max.Y)+float64(fontSize)/2.0 || drawY < float64(yBounds.Min.Y)-float64(fontSize)/2.0 {
			v = v - dv
			continue
		}
		// Skip interleaved ticks when labels would crowd together
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
		// Prevent overlapping adjacent labels
		if math.Abs(lblDrawY-lastDrawnY) < float64(fontSize)-1.0 {
			v = v - dv
			continue
		}
		if unitDrawn && math.Abs(lblDrawY-unitY) < float64(fontSize)-1.0 {
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

// clear erases the partition bounding box with the current theme background color.
func (cl *ftChannelLabelViewer) clear() {
	draw.Draw(cl.img, cl.rect(), &image.Uniform{theme.BackgroundColor()},
		image.ZP, draw.Src)
}

