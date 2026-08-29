package gui

import (
	"fmt"
	"image"
	"image/color"
	"math"

	"fynescope/control"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"golang.org/x/image/draw"
	// "fynescope/checkcolorpick"
)

type digitalRaster struct {
	scp      *ScpDesc
	raster   *canvas.Raster
	tappable *tappableDigitalRaster
	window   fyne.Window

	showInspector bool
	mouseX        float32
	mouseY        float32
}

func (scp *ScpDesc) newDigitalRaster(window fyne.Window) (*fyne.Container, *digitalRaster) {
	dr := &digitalRaster{
		scp:    scp,
		window: window,
	}

	dr.raster = canvas.NewRaster(dr.generate)
	tappable := newTappableDigitalRaster(dr)
	dr.tappable = tappable

	return container.NewMax(container.NewClip(tappable)), dr
}

func (dr *digitalRaster) updateColorPickers() {
}

func (dr *digitalRaster) refresh() {
	if dr.raster != nil {
		canvas.Refresh(dr.raster)
	}
}

func (dr *digitalRaster) generate(w, h int) image.Image {
	if w <= 0 || h <= 0 {
		return image.NewRGBA(image.Rect(0, 0, 1, 1))
	}

	dr.scp.screenLocker.Lock()
	defer dr.scp.screenLocker.Unlock()

	img := image.NewRGBA(image.Rect(0, 0, w, h))
	bgCol := dr.scp.theme.Color(ColorNameSignalBackground, 0)
	draw.Draw(img, img.Bounds(), &image.Uniform{bgCol}, image.ZP, draw.Src)

	if !dr.scp.Settings.Digital.Ports[0].Enabled && !dr.scp.Settings.Digital.Ports[1].Enabled {
		return img
	}

	// Match the horizontal margins and signal area with the analog screen
	var minX, maxX int
	var signalW float64

	if dr.scp.ftScopeSignalScreen != nil && dr.scp.ftScopeSignalScreen.Bounds().Dx() > 0 {
		bounds := dr.scp.ftScopeSignalScreen.Bounds()
		minX = bounds.Min.X
		maxX = bounds.Max.X
		signalW = float64(bounds.Dx()) - 1
	} else {
		leftMargin, rightMargin := dr.scp.clipFtChRangeScrs(float32(w), float32(h))
		minX = int(math.Round(float64(leftMargin)))
		maxX = int(math.Round(float64(float32(w) - rightMargin)))
		signalW = float64(maxX-minX) - 1
	}

	if signalW <= 0 {
		signalW = 1
	}

	// Use the same horizontal division positions as the analog raster (ftDivsX).
	// Because analog and digital rasters have the same total width and same horizontal layout,
	// ftDivsX contains the exact pixel positions for vertical time lines.
	dr.scp.setFtHDivsX()
	gridCol := dr.scp.theme.Color(theme.ColorNameDisabled, 0)
	for _, divX := range dr.scp.ftDivsX {
		x := int(math.Round(float64(divX)))
		if x >= minX && x <= maxX && x >= 0 && x < w {
			drawVerticalDashedLine(img, float32(x), 0, float32(h), gridCol, 4, 6)
		}
	}

	// We have 16 digital channels maximum
	activeChannels := 0
	if dr.scp.Settings.Digital.Ports[0].Enabled {
		activeChannels += 8
	}
	if dr.scp.Settings.Digital.Ports[1].Enabled {
		activeChannels += 8
	}

	if activeChannels == 0 {
		return img
	}

	channelHeight := float64(h) / float64(activeChannels)

	maxScreenTime := dr.scp.maxScreenTime
	if maxScreenTime <= 0 {
		maxScreenTime = 1
	}

	unit := signalW / maxScreenTime
	deltaT := unit * dr.scp.controlSamplingTimeInterval

	leftPadding := float64(control.LeftOut)
	extra := 1.0

	triggerOffsetSeconds := float64(dr.scp.controlTriggerTimeOffset) / 1e15
	expectedOffset := leftPadding * dr.scp.controlSamplingTimeInterval
	if dr.scp.controlSamplingTimeInterval > 0 && math.Abs(triggerOffsetSeconds-expectedOffset) > 10*dr.scp.controlSamplingTimeInterval {
		triggerOffsetSeconds = expectedOffset
	}

	t0 := (-leftPadding*dr.scp.controlSamplingTimeInterval +
		float64(dr.scp.controlXRoundError) +
		triggerOffsetSeconds) * unit

	if dr.scp.timeZoomWindow != nil {
		t0 -= dr.scp.timeZoomBoxOffset * unit
	}
	t0 -= extra * deltaT

	chIdx := 0
	for port := 0; port < 2; port++ {
		if !dr.scp.Settings.Digital.Ports[port].Enabled {
			continue
		}
		var buf []uint8
		if len(dr.scp.digitalDisplayBuffer) > port {
			buf = dr.scp.digitalDisplayBuffer[port]
		}
		samples := len(buf)

		for c := 0; c < 8; c++ {
			if !dr.scp.Settings.Digital.ChannelsEnabled[port*8+c] {
				chIdx++
				continue
			}

			yBase := float64(chIdx) * channelHeight
			yHigh := int(math.Round(yBase + channelHeight*0.2))
			yLow := int(math.Round(yBase + channelHeight*0.8))

			chColor := dr.scp.Settings.Digital.ChannelColors[port*8+c]
			lineCol := color.RGBA{R: chColor.R, G: chColor.G, B: chColor.B, A: 255}
			hatchCol := color.RGBA{R: uint8(float64(chColor.R) * 0.6), G: uint8(float64(chColor.G) * 0.6), B: uint8(float64(chColor.B) * 0.6), A: 8}

			// Draw channel label on the left margin if space is available
			if minX >= 25 {
				label := fmt.Sprintf("D%d", port*8+c)
				if l := dr.scp.Settings.Digital.ChannelLabels[port*8+c]; l != "" {
					runes := []rune(l)
					if len(runes) > 6 {
						l = string(runes[:6])
					}
					label += " " + l
				}
				lblLeft, lblTop, lblRight, lblBottom := dr.scp.boundString(label)
				_ = lblLeft
				_ = lblRight
				dr.scp.addLabel(img, 6, int(math.Round(yBase+channelHeight*0.5-(float64(lblTop+lblBottom)/2))), label, lineCol, 12)
			}

			for x := minX; x <= maxX && x < w; x++ {
				yPos := yLow
				isHigh := false
				sampleIdx := -1
				if samples > 0 && deltaT > 0 {
					sampleIdx = int(math.Round((float64(x) - float64(minX) - t0) / deltaT))
					if sampleIdx >= 0 && sampleIdx < samples {
						val := buf[sampleIdx]
						if (val & (1 << c)) != 0 {
							yPos = yHigh
							isHigh = true
						}
					}
				}
				img.Set(x, yPos, lineCol)

				// Draw diagonal hatches under high level signal
				if isHigh {
					for y := yHigh + 1; y <= yLow; y++ {
						if (x+y)%8 == 0 {
							img.Set(x, y, hatchCol)
						}
					}
				}

				if x > minX && samples > 0 && deltaT > 0 {
					prevSampleIdx := int(math.Round((float64(x-1) - float64(minX) - t0) / deltaT))
					if sampleIdx >= 0 && sampleIdx < samples && prevSampleIdx >= 0 && prevSampleIdx < samples {
						prevVal := buf[prevSampleIdx]
						val := buf[sampleIdx]
						prevBit := (prevVal & (1 << c)) != 0
						currBit := (val & (1 << c)) != 0
						if prevBit != currBit {
							for y := yHigh; y <= yLow; y++ {
								img.Set(x, y, lineCol)
							}
						}
					}
				}
			}
			chIdx++
		}
	}

	if dr.scp.Settings.Digital.Trigger.Enabled {
		triggerTimeOffset := dr.scp.Settings.Time.TriggerTimeOffset
		if dr.scp.timeZoomWindow != nil {
			triggerTimeOffset -= dr.scp.timeZoomBoxOffset
		}
		triggerX := float64(minX) + float64(triggerTimeOffset)*signalW/maxScreenTime

		if triggerX >= float64(minX) && triggerX <= float64(maxX) {
			xInt := int(math.Round(triggerX))
			lineCol := dr.scp.theme.Color(theme.ColorNameForeground, 0)
			for y := 0; y < h; y++ {
				img.Set(xInt, y, lineCol)
			}
		}
	}

	if dr.showInspector {
		crosscol := color.RGBA{180, 180, 180, 180}
		mx := int(dr.mouseX)

		for i := 0; i < h; i++ {
			img.Set(mx, i, crosscol)
		}

		idxF := (float64(mx) - float64(minX) - t0) / deltaT
		idx := int(math.Round(idxF))

		timeAtCursor := (float64(idx)-leftPadding)*dr.scp.controlSamplingTimeInterval - dr.scp.Settings.Time.TriggerTimeOffset
		timeStr := formatTime(timeAtCursor)

		var binStrLSB, binStrMSB string
		var hexValLSBD0, hexValMSBD0 uint16
		var bits int

		for port := 0; port < 2; port++ {
			if !dr.scp.Settings.Digital.Ports[port].Enabled {
				continue
			}
			var buf []uint8
			if len(dr.scp.digitalDisplayBuffer) > port {
				buf = dr.scp.digitalDisplayBuffer[port]
			}
			if idx >= 0 && idx < len(buf) {
				portVal := buf[idx]
				for c := 0; c < 8; c++ {
					chIdx := port*8 + c
					if dr.scp.Settings.Digital.ChannelsEnabled[chIdx] {
						bitValLSB := (portVal >> c) & 1
						bitValMSB := (portVal >> (7 - c)) & 1
						hexValLSBD0 |= (uint16(bitValLSB) << bits)
						hexValMSBD0 |= (uint16(bitValMSB) << bits)
						bits++
					}
				}
			}
		}

		if bits > 0 {
			for i := 0; i < bits; i++ {
				b := (hexValMSBD0 >> i) & 1
				binStrLSB += fmt.Sprintf("%d", b)
				binStrMSB = fmt.Sprintf("%d", b) + binStrMSB
			}
		} else {
			binStrLSB = "0"
			binStrMSB = "0"
		}

		var info []struct {
			text string
			col  color.Color
		}
		info = append(info, struct {
			text string
			col  color.Color
		}{"T: " + timeStr, color.White})

		info = append(info, struct {
			text string
			col  color.Color
		}{fmt.Sprintf("Hexa LSBD0: 0x%X MSBD0: 0x%X", hexValLSBD0, hexValMSBD0), color.White})

		info = append(info, struct {
			text string
			col  color.Color
		}{"LSB: " + binStrLSB, color.White})

		info = append(info, struct {
			text string
			col  color.Color
		}{"MSB: " + binStrMSB, color.White})

		// Draw the box
		lineHeight := 20
		maxW := float32(0)
		for _, item := range info {
			left, _, right, _ := dr.scp.boundString(item.text)
			if right-left > maxW {
				maxW = right - left
			}
		}
		boxWidth := int(maxW) + 15
		boxHeight := len(info)*lineHeight + 10

		x := int(dr.mouseX) + 20
		y := int(dr.mouseY) + 20

		if x+boxWidth > w-2 {
			x = int(dr.mouseX) - boxWidth - 20
		}
		if x < 2 {
			x = 2
		}
		if y+boxHeight > h-2 {
			y = int(dr.mouseY) - boxHeight - 20
		}
		if y < 2 {
			y = 2
		}

		rect := image.Rect(x, y, x+boxWidth, y+boxHeight)
		draw.Draw(img, rect, &image.Uniform{color.RGBA{20, 20, 20, 220}}, image.ZP, draw.Over)

		for i := 0; i < boxWidth; i++ {
			img.Set(x+i, y, color.White)
			img.Set(x+i, y+boxHeight-1, color.White)
		}
		for i := 0; i < boxHeight; i++ {
			img.Set(x, y+i, color.White)
			img.Set(x+boxWidth-1, y+i, color.White)
		}

		for i, item := range info {
			dr.scp.addLabel(img, x+5, y+5+(i+1)*lineHeight-5, item.text, item.col)
		}
	}

	return img
}

type tappableDigitalRaster struct {
	widget.BaseWidget
	dr *digitalRaster
}

func newTappableDigitalRaster(dr *digitalRaster) *tappableDigitalRaster {
	t := &tappableDigitalRaster{dr: dr}
	t.ExtendBaseWidget(t)
	return t
}

func (t *tappableDigitalRaster) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(t.dr.raster)
}

func (t *tappableDigitalRaster) Tapped(event *fyne.PointEvent) {}

func (t *tappableDigitalRaster) TappedSecondary(event *fyne.PointEvent) {
	t.dr.showInspector = !t.dr.showInspector
	if t.dr.showInspector {
		t.dr.mouseX = event.Position.X
		t.dr.mouseY = event.Position.Y
	}
	t.dr.raster.Refresh()
}

func (t *tappableDigitalRaster) MouseMoved(event *desktop.MouseEvent) {
	if t.dr.showInspector {
		t.dr.mouseX = event.Position.X
		t.dr.mouseY = event.Position.Y
		t.dr.raster.Refresh()
	}
}

func (t *tappableDigitalRaster) MouseIn(event *desktop.MouseEvent) {
	if t.dr.showInspector {
		t.dr.mouseX = event.Position.X
		t.dr.mouseY = event.Position.Y
		t.dr.raster.Refresh()
	}
}

func (t *tappableDigitalRaster) MouseOut() {}

func (t *tappableDigitalRaster) Dragged(event *fyne.DragEvent) {
	if !t.dr.scp.Settings.Digital.Trigger.Enabled {
		return
	}
	dx := event.Dragged.DX
	t.dr.scp.addFtXOffset(float64(dx))
	t.dr.scp.setTriggerTime(t.dr.scp.Settings.Time.TriggerTimeOffset)
	t.dr.scp.clearAllFtPersistentLayers()
	t.dr.scp.clearAllDftPersistentLayers()
	if t.dr.scp.ftBottomLabelViewer != nil {
		t.dr.scp.ftBottomLabelViewer.(*timeLabelViewer).enableRefresh()
	}
	if t.dr.scp.timeZoomBottomLabelViewer != nil {
		t.dr.scp.timeZoomBottomLabelViewer.(*timeLabelViewer).enableRefresh()
	}
	t.dr.scp.refreshRasters()
}

func (t *tappableDigitalRaster) DragEnd() {}
