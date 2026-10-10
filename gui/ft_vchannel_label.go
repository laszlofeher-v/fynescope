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
	// ftVChannelLabelViewer renders voltage scale division labels and handles interactive
	// vertical offset adjustments (dragging, scrolling, and right-click reset) for a math/virtual
	// channel on the time-domain raster screen.
	ftVChannelLabelViewer struct {
		rasterPartition
		chLabelRect   image.Rectangle // Bounding rectangle of the virtual channel label display column.
		vChannelIndex int             // Zero-based index within ScpDesc.Settings.VirtualChannels slice.
		leftLabel     bool            // True if labels are aligned on the left edge of the signal area.
		selected      bool            // True while this virtual channel label is being actively dragged.
		scp           *ScpDesc        // Pointer to the parent oscilloscope descriptor.
		isTimeZoom    bool            // True if attached to the time-zoom viewport raster.

		displayOffsetFraction float64 // Accumulated fractional pixel offset for smooth vertical dragging.
	}
)

// raster returns the target raster screen for this viewer, selecting either the
// time-zoom raster or the main time-domain (FT) raster.
func (cl *ftVChannelLabelViewer) raster() *screenRaster {
	if cl.isTimeZoom {
		return cl.scp.timeZoomRaster
	}
	return cl.scp.ftRaster
}

// Compile-time assertions verifying that ftVChannelLabelViewer implements the necessary
// interactive partition interfaces.
var (
	_ mouser     = (*ftVChannelLabelViewer)(nil)
	_ dragger    = (*ftVChannelLabelViewer)(nil)
	_ scroller   = (*ftVChannelLabelViewer)(nil)
	_ keyer      = (*ftVChannelLabelViewer)(nil)
	_ drawer     = (*ftVChannelLabelViewer)(nil)
	_ cursorable = (*ftVChannelLabelViewer)(nil)
)

// typedKey handles keyboard navigation (up and down arrow keys) to adjust the virtual channel's vertical position.
func (cl *ftVChannelLabelViewer) typedKey(x, y float32, keyName fyne.KeyName) {
	switch keyName {
	case fyne.KeyDown:
		cl.scrolled(-scrollDelta, x, y)
	case fyne.KeyUp:
		cl.scrolled(scrollDelta, x, y)
	}
}

// newFtVChannelLabelViewer constructs an ftVChannelLabelViewer partition for the given virtual channel index.
func newFtVChannelLabelViewer(img rasterImage, imgRect image.Rectangle, vChannelIndex int,
	scopeSignalScreen image.Rectangle, leftLabel bool, scp *ScpDesc,
	isTimeZoom bool) ftVChannelLabelViewer {
	cl := ftVChannelLabelViewer{
		rasterPartition: rasterPartition{
			img:         img,
			imgRect:     imgRect,
			refreshFlag: true,
		},
		chLabelRect:   scopeSignalScreen,
		vChannelIndex: vChannelIndex,
		leftLabel:     leftLabel,
		scp:           scp,
		isTimeZoom:    isTimeZoom,
	}
	return cl
}

// cursor returns a pointer cursor when the mouse hovers over the interactive label region.
func (cl *ftVChannelLabelViewer) cursor(x, y float32) (desktop.Cursor, bool) {
	if cl.mouseIn(x, y) {
		return desktop.PointerCursor, true
	}
	return desktop.DefaultCursor, false
}

// mouseMoved handles mouse hover motion events.
func (cl *ftVChannelLabelViewer) mouseMoved(x, y float32) {
}

// mouseIn reports whether the coordinate (x, y) lies within the label viewer partition bounds.
func (cl *ftVChannelLabelViewer) mouseIn(x, y float32) bool {
	p := image.Point{X: int(math.Round(float64(x))), Y: int(math.Round(float64(y)))}
	if p.In(cl.rect()) {
		return true
	}
	return false
}

// mouseDown handles mouse button clicks. Secondary (right) click resets the virtual channel's vertical
// display offset back to zero. Primary (left) click begins dragging.
func (cl *ftVChannelLabelViewer) mouseDown(button desktop.MouseButton,
	modifier fyne.KeyModifier, x, y float32) {
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

// mouseUp releases the active drag selection.
func (cl *ftVChannelLabelViewer) mouseUp(button desktop.MouseButton,
	modifier fyne.KeyModifier, x, y float32) {
	cl.selected = false
}

// setChDispYOffset adjusts the virtual channel's vertical display offset by dy pixels.
// When scroll is true, dy is treated as an incremental jump added to the snapped division offset.
// Display offset is clamped to the partition height and snapped to vertical divisions.
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
			// Clamp offset to screen bounds
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

// dragged processes mouse drag events to dynamically update the vertical display offset.
func (cl *ftVChannelLabelViewer) dragged(dx, dy, x, y float32) {
	if cl.selected {
		cl.setChDispYOffset(float64(dy), float64(x), float64(y), false)
	}
}

// scrolled processes mouse wheel scrolling by shifting by one division (nY) per step.
func (cl *ftVChannelLabelViewer) scrolled(delta, x, y float32) {
	nY := float64(cl.img.Bounds().Dy()) / float64(numberOfDivs)
	if delta < 0 {
		cl.setChDispYOffset(nY, float64(x), float64(y), true)
	} else if delta > 0 {
		cl.setChDispYOffset(-nY, float64(x), float64(y), true)
	}
}

// draw renders the voltage scale divisions, unit label ("V" or "mV"), and tick values
// along the vertical grid corresponding to the virtual channel's voltage range and vertical offset.
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

	// Format voltage unit as V or mV depending on range magnitude
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
	effectiveH := float32(yBounds.Dy())
	dy := (effectiveH - 1.0) / 10.0
	xoffset := left - right
	if !cl.leftLabel {
		xoffset = -float32(xBounds.Dx())
	}
	yOffset := cl.scp.offsetNToFtY(vch.DisplayVOffset)
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
			unitName, vch.Col[cl.scp.Settings.ChannelColorIndex])
		unitDrawn = true
	}

	// Render voltage division values across the time-domain vertical scale
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
			vch.Col[cl.scp.Settings.ChannelColorIndex])
		lastDrawnY = lblDrawY
		v = v - dv
	}
	cl.disableRefresh()
}

// clear erases the partition bounding box with the current theme background color.
func (cl *ftVChannelLabelViewer) clear() {
	draw.Draw(cl.img, cl.rect(), &image.Uniform{theme.BackgroundColor()},
		image.ZP, draw.Src)
}
