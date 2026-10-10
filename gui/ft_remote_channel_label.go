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
	// ftRemoteChannelLabelViewer renders voltage scale division labels and handles interactive
	// vertical offset adjustments (dragging, scrolling, and right-click reset) for a remote
	// (network-streamed) oscilloscope channel on the time-domain raster screen.
	ftRemoteChannelLabelViewer struct {
		rasterPartition
		chLabelRect   image.Rectangle // Bounding rectangle of the channel label display column.
		remoteChanIdx int             // Index within ScpDesc.remoteChannels slice.
		leftLabel     bool            // True if labels are aligned on the left edge of the signal screen.
		selected      bool            // True while the label is being actively dragged by the user.
		scp           *ScpDesc        // Pointer to the parent oscilloscope descriptor.
		isTimeZoom    bool            // True if this viewer is associated with the time-zoom viewport raster.

		displayOffsetFraction float64 // Accumulated fractional pixel offset for smooth vertical dragging.
	}
)

// raster returns the target raster screen for this viewer, selecting either the
// time-zoom raster or the main time-domain (FT) raster.
func (cl *ftRemoteChannelLabelViewer) raster() *screenRaster {
	if cl.isTimeZoom {
		return cl.scp.timeZoomRaster
	}
	return cl.scp.ftRaster
}

// Compile-time assertions verifying that ftRemoteChannelLabelViewer implements the necessary
// interactive partition interfaces.
var (
	_ mouser     = (*ftRemoteChannelLabelViewer)(nil)
	_ dragger    = (*ftRemoteChannelLabelViewer)(nil)
	_ scroller   = (*ftRemoteChannelLabelViewer)(nil)
	_ keyer      = (*ftRemoteChannelLabelViewer)(nil)
	_ drawer     = (*ftRemoteChannelLabelViewer)(nil)
	_ cursorable = (*ftRemoteChannelLabelViewer)(nil)
)

// typedKey handles keyboard navigation (up and down arrow keys) to adjust the channel's vertical position.
func (cl *ftRemoteChannelLabelViewer) typedKey(x, y float32, keyName fyne.KeyName) {
	switch keyName {
	case fyne.KeyDown:
		cl.scrolled(-scrollDelta, x, y)
	case fyne.KeyUp:
		cl.scrolled(scrollDelta, x, y)
	}
}

// newFtRemoteChannelLabelViewer constructs an ftRemoteChannelLabelViewer partition for the given remote channel index.
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

// cursor returns a pointer cursor when the mouse hovers over the interactive label region.
func (cl *ftRemoteChannelLabelViewer) cursor(x, y float32) (desktop.Cursor, bool) {
	if cl.mouseIn(x, y) {
		return desktop.PointerCursor, true
	}
	return desktop.DefaultCursor, false
}

// mouseMoved handles mouse hover motion events.
func (cl *ftRemoteChannelLabelViewer) mouseMoved(x, y float32) {
}

// mouseIn reports whether the coordinate (x, y) lies within the label viewer partition bounds.
func (cl *ftRemoteChannelLabelViewer) mouseIn(x, y float32) bool {
	p := image.Point{X: int(math.Round(float64(x))), Y: int(math.Round(float64(y)))}
	return p.In(cl.rect())
}

// mouseDown handles mouse button clicks. Secondary (right) click resets the channel's vertical
// display offset back to zero. Primary (left) click begins dragging.
func (cl *ftRemoteChannelLabelViewer) mouseDown(button desktop.MouseButton,
	modifier fyne.KeyModifier, x, y float32) {
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

// mouseUp releases the active drag selection.
func (cl *ftRemoteChannelLabelViewer) mouseUp(button desktop.MouseButton,
	modifier fyne.KeyModifier, x, y float32) {
	cl.selected = false
}

// setChDispYOffset adjusts the remote channel's vertical display offset by dy pixels.
// When scroll is true, dy is treated as an incremental jump added to the snapped division offset.
// Display offset is clamped to the partition height and snapped to vertical divisions.
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
			// Clamp offset to screen bounds
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

// dragged processes mouse drag events to dynamically update the vertical display offset.
func (cl *ftRemoteChannelLabelViewer) dragged(dx, dy, x, y float32) {
	if cl.selected {
		cl.setChDispYOffset(float64(dy), float64(x), float64(y), false)
	}
}

// scrolled processes mouse wheel scrolling by shifting by one division (nY) per step.
func (cl *ftRemoteChannelLabelViewer) scrolled(delta, x, y float32) {
	nY := float64(cl.img.Bounds().Dy()) / float64(numberOfDivs)
	if delta < 0 {
		cl.setChDispYOffset(nY, float64(x), float64(y), true)
	} else if delta > 0 {
		cl.setChDispYOffset(-nY, float64(x), float64(y), true)
	}
}

// draw renders the voltage scale divisions, unit label ("V" or "mV"), and tick values
// along the vertical grid corresponding to the remote channel's voltage range and vertical offset.
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

	// Format voltage unit as V or mV depending on range magnitude
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

	// Determine vertical position for unit header label
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
			unitName, rch.Color)
		unitDrawn = true
	}

	// Render voltage value labels for each division line
	v := startValue
	for i, y_px := range cl.scp.ftDivsY {
		if y_px < 0 {
			v = v - dv
			continue
		}
		drawY := float64(y_px) + yOffset
		if drawY > float64(yBounds.Max.Y)+float64(fontSize)/2.0 ||
			drawY < float64(yBounds.Min.Y)-float64(fontSize)/2.0 {
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

// clear erases the partition bounding box with the current theme background color.
func (cl *ftRemoteChannelLabelViewer) clear() {
	draw.Draw(cl.img, cl.rect(), &image.Uniform{theme.BackgroundColor()},
		image.ZP, draw.Src)
}
