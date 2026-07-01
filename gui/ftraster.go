package gui

import (
	"fmt"
	"fynescope/control"
	"fynescope/genericps"
	"fynescope/settings"
	"image"
	"image/color"
	"image/draw"
	"log/slog"
	"math"
	"time"

	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/driver/desktop"
)

const (
	e15 = 1e15
)

type (
	signalViewer struct {
		rasterPartition
		scp            *ScpDesc
		showInspector  bool
		mouseX, mouseY float32

		inspectorLastX, inspectorLastY float32
		inspectorSumV                  []float32
		inspectorSumVCur               []float32
		inspectorSamples               int
		inspectorLastUpdate            time.Time
		inspectorDispV                 []float32
		inspectorDispVCur              []float32
	}
)

var (
	savedW float64
)

func newSignalViewer(img rasterImage, imgRect image.Rectangle,
	scp *ScpDesc) *signalViewer {
	sv := &signalViewer{rasterPartition: rasterPartition{img: img, imgRect: imgRect, refreshFlag: true},
		scp: scp}
	return sv
}

func (sv *signalViewer) mouseIn(x, y float32) bool {
	p := image.Point{X: int(math.Round(float64(x))), Y: int(math.Round(float64(y)))}
	if p.In(sv.imgRect) {
		return true
	}
	return false
}

func (sv *signalViewer) mouseDown(button desktop.MouseButton, x, y float32) {
	if button == desktop.RightMouseButton && sv.mouseIn(x, y) {
		sv.showInspector = true
		sv.mouseX = x
		sv.mouseY = y
		sv.enableRefresh()
		canvas.Refresh(sv.scp.ftRaster)
	}
}

func (sv *signalViewer) mouseUp(button desktop.MouseButton, x, y float32) {
	if button == desktop.RightMouseButton {
		sv.showInspector = false
		sv.enableRefresh()
		canvas.Refresh(sv.scp.ftRaster)
	}
}

func (sv *signalViewer) mouseMoved(x, y float32) {
	if sv.showInspector {
		sv.mouseX = x
		sv.mouseY = y
		// Clamp to signal viewer bounds
		if sv.mouseX < float32(sv.imgRect.Min.X) {
			sv.mouseX = float32(sv.imgRect.Min.X)
		}
		if sv.mouseX > float32(sv.imgRect.Max.X-1) { // -1 because it's the last pixel
			sv.mouseX = float32(sv.imgRect.Max.X - 1)
		}
		if sv.mouseY < float32(sv.imgRect.Min.Y) {
			sv.mouseY = float32(sv.imgRect.Min.Y)
		}
		if sv.mouseY > float32(sv.imgRect.Max.Y-1) {
			sv.mouseY = float32(sv.imgRect.Max.Y - 1)
		}
		sv.enableRefresh()
		canvas.Refresh(sv.scp.ftRaster)
	}
}

func (sv *signalViewer) dragged(dx, dy, x, y float32) {
	if sv.showInspector {
		sv.mouseX = x
		sv.mouseY = y
		if sv.mouseX < float32(sv.imgRect.Min.X) {
			sv.mouseX = float32(sv.imgRect.Min.X)
		}
		if sv.mouseX > float32(sv.imgRect.Max.X-1) {
			sv.mouseX = float32(sv.imgRect.Max.X - 1)
		}
		if sv.mouseY < float32(sv.imgRect.Min.Y) {
			sv.mouseY = float32(sv.imgRect.Min.Y)
		}
		if sv.mouseY > float32(sv.imgRect.Max.Y-1) {
			sv.mouseY = float32(sv.imgRect.Max.Y - 1)
		}
		sv.enableRefresh()
		canvas.Refresh(sv.scp.ftRaster)
	}
}

func (sv *signalViewer) draw() {
	if !sv.scp.shouldDrawRaster(ftTabIndex) {
		return
	}
	bounds := sv.scp.ftScopeSignalScreen.Bounds()
	h := float64(bounds.Dy())
	w := float64(bounds.Dx())
	sv.scp.drawFtDivisions()
	zeroOffset := bounds.Min.Y + bounds.Dy()/2

	if sv.scp.triggerSettingMsg.Mode == control.ETS {
		etsDx := float64(w) / (sv.scp.maxScreenTime * 1e15)
		sv.drawETS(w, h, bounds, zeroOffset, etsDx)
	} else {
		deltaT := (w / sv.scp.maxScreenTime) * sv.scp.controlSamplingTimeInterval
		sv.drawNormal(w, h, bounds, zeroOffset, deltaT)
	}

	if sv.showInspector {
		sv.drawInspector(w, h, bounds)
	}
}

func (sv *signalViewer) drawETS(w, h float64, bounds image.Rectangle, zeroOffset int, etsDx float64) {
	fMinY := float64(bounds.Min.Y)
	fMaxY := float64(bounds.Max.Y)
	fMinX := float64(bounds.Min.X)
	fMaxX := float64(bounds.Max.X)
	unit := (w) / float64(sv.scp.maxScreenTime)
	for channelIndex := range sv.scp.channelViewers {
		channelViewer := &sv.scp.channelViewers[channelIndex]
		channel := &sv.scp.Settings.Channels[channelIndex]
		if len(sv.scp.displayBuffers) > 0 {
			displayBuffer := sv.scp.displayBuffers[channelIndex]
			if len(displayBuffer) > 0 && channel.Enabled {
				yScale := h / float64((2.0 * genericps.RangeValuesMv[channel.VRange]))
				col := channel.Col[sv.scp.Settings.ChannelColorIndex]
				yOffset := float64(0)
				if channelViewer.displayOffsetInt != 0 {
					yOffset = sv.scp.offsetNToFtY(channel.DisplayVOffset)
				}
				
				var targetImg draw.Image = sv.scp.ftScopeSignalScreen.(draw.Image)
				if channel.Persistence {
					if sv.scp.ftPersistentLayers[channelIndex] == nil || sv.scp.ftPersistentLayers[channelIndex].Bounds() != bounds {
						sv.scp.ftPersistentLayers[channelIndex] = image.NewRGBA(bounds)
					}
					targetImg = sv.scp.ftPersistentLayers[channelIndex]
				}
				
				etsDrawDot := func() {
					startX := float64(bounds.Min.X) //+ float64(sv.scp.controlXRoundError)*unit
					// slog.Debug("etsDrawRaw", "controlXRoundError", sv.scp.controlXRoundError)
					offsetFloat := float64(zeroOffset) + yOffset
					s := displayBuffer[0]
					if channel.Inverted {
						s = -s
					}
					for i := 1; i < len(sv.scp.etsBuffer) && i < len(displayBuffer); i = i + 1 {
						s := displayBuffer[i]
						if channel.Inverted {
							s = -s
						}
						y := -yScale*float64(s) + offsetFloat
						switch { // TODO not real clip
						case y > fMaxY:
							y = fMaxY
						case y < fMinY:
							y = fMinY
						}
						x := sv.scp.Settings.Time.TriggerTimeOffset*unit + (float64(sv.scp.etsBuffer[i]))*etsDx + startX
						// slog.Debug("draw", "i", i, "etsBuffer", sv.scp.etsBuffer[i], "etsDx", etsDx, "x", x)
						switch { // TODO not real clip
						case x > fMaxX:
							x = fMaxX
						case x < fMinX:
							x = fMinX
						}
						const r = 5.0
						drawCircle(targetImg, float32(x), float32(y), r, col)
						drawCircle(targetImg, float32(x), float32(y), r-1, col)
						drawCircle(targetImg, float32(x), float32(y), r-2, col)
					}
				}
				// Raw drawer
				etsDrawRaw := func() {
					startX := float64(bounds.Min.X) //+ float64(sv.scp.controlXRoundError)*unit
					offsetFloat := float64(zeroOffset) + yOffset
					prevX := startX
					s := displayBuffer[0]
					if channel.Inverted {
						s = -s
					}
					prevY := -yScale*float64(s) + offsetFloat
					switch { // TODO not real clip
					case prevY > fMaxY:
						prevY = fMaxY
					case prevY < fMinY:
						prevY = fMinY
					}
					for i := 1; i < len(sv.scp.etsBuffer) && i < len(displayBuffer); i = i + 1 {
						s := displayBuffer[i]
						if channel.Inverted {
							s = -s
						}
						y := -yScale*float64(s) + offsetFloat
						switch { // TODO not real clip
						case y > fMaxY:
							y = fMaxY
						case y < fMinY:
							y = fMinY
						}
						x := sv.scp.Settings.Time.TriggerTimeOffset*unit + (float64(sv.scp.etsBuffer[i]))*etsDx + startX
						switch {
						case x > fMaxX:
							x = fMaxX
						case x < fMinX:
							x = fMinX
						}
						err := drawLine(targetImg, float32(prevX), float32(prevY), float32(x), float32(prevY), col)
						if err != nil {
							slog.Debug("ets", "x", x, "y", y, "fMinX", fMinX, "fMaxX", fMaxX, "fMinY", fMinY, "fMaxY", fMaxY)
							panic("draw error")
						}
						err = drawLine(targetImg, float32(x), float32(prevY), float32(x), float32(y), col)
						prevX = x
						if err != nil {
							slog.Debug("ets", "x", x, "y", y, "fMinX", fMinX, "fMaxX", fMaxX)
							panic("draw error")
						}
						prevY = y
					}
				}

				// Linear interpolation 1
				etsDrawLinear := func() {
					startX := float64(bounds.Min.X)
					offsetFloat := float64(zeroOffset) + yOffset
					prevX := startX
					s := displayBuffer[0]
					if channel.Inverted {
						s = -s
					}
					prevY := -yScale*float64(s) + offsetFloat
					for i := 1; i < len(sv.scp.etsBuffer) && i < len(displayBuffer); i = i + 1 {
						s := displayBuffer[i]
						if channel.Inverted {
							s = -s
						}
						y := -yScale*float64(s) + offsetFloat
						x := sv.scp.Settings.Time.TriggerTimeOffset*unit + (float64(sv.scp.etsBuffer[i]))*etsDx + startX
						drawLine(targetImg, float32(prevX), float32(prevY), float32(x), float32(y), col)
						prevX = x
						prevY = y
					}
				}

				switch sv.scp.Settings.Time.Interpolation {
				case settings.Dot:
					etsDrawDot()
				case settings.Raw:
					etsDrawRaw()
				case settings.Linear:
					etsDrawLinear()
				case settings.Sinc:
					slog.Error("sinc is not allowed in ETS")
					etsDrawLinear()
				default:
					panic("Undefine interpolation mode")
				}
				
				if channel.Persistence {
					draw.Draw(sv.scp.ftScopeSignalScreen, bounds, sv.scp.ftPersistentLayers[channelIndex], bounds.Min, draw.Over)
				}
			}
		}
	}
}

func (sv *signalViewer) drawNormal(w, h float64, bounds image.Rectangle, zeroOffset int, deltaT float64) {
	unit := (w) / float64(sv.scp.maxScreenTime)

	for channelIndex := range sv.scp.channelViewers {
		channelViewer := &sv.scp.channelViewers[channelIndex]
		channel := &sv.scp.Settings.Channels[channelIndex]
		if len(sv.scp.displayBuffers) > 0 {
			displayBuffer := sv.scp.displayBuffers[channelIndex]
			if displayBuffer != nil && channel.Enabled {
				yScale := h / float64(2.0*genericps.RangeValuesMv[channel.VRange])
				col := channel.Col[sv.scp.Settings.ChannelColorIndex]
				yOffset := float64(0)
				if channelViewer.displayOffsetInt != 0 {
					yOffset = sv.scp.offsetNToFtY(channelViewer.displayOffsetInt)
				}
				offsetFloat := float64(zeroOffset) + yOffset

				// Correctly identify padding and compensation based on interpolation mode
				var leftPadding float64
				var extra float64
				if sv.scp.Settings.Time.Interpolation == settings.Sinc {
					totalSamples := len(displayBuffer)
					displaySamples := totalSamples / control.SincWMultiplier
					leftPadding = float64(totalSamples-displaySamples) / 2.0
					extra = 0
				} else {
					leftPadding = float64(control.LeftOut)
					extra = 1 // Compensation for different XRoundError definition in non-Sinc mode
				}

				// t0 is the pixel position of the first sample (index 0) relative to bounds.Min.X
				t0 := (-leftPadding*sv.scp.controlSamplingTimeInterval +
					float64(sv.scp.controlXRoundError) +
					float64(sv.scp.controlTriggerTimeOffset)/1e15) * unit
				t0 -= extra * deltaT
				
				var targetImg draw.Image = sv.scp.ftScopeSignalScreen.(draw.Image)
				if channel.Persistence {
					if sv.scp.ftPersistentLayers[channelIndex] == nil || sv.scp.ftPersistentLayers[channelIndex].Bounds() != bounds {
						sv.scp.ftPersistentLayers[channelIndex] = image.NewRGBA(bounds)
					}
					targetImg = sv.scp.ftPersistentLayers[channelIndex]
				}

				drawDot := func() {
					s0 := displayBuffer[0]
					if channel.Inverted {
						s0 = -s0
					}
					for i := 1; i < len(displayBuffer); i++ {
						x := t0 + float64(i)*deltaT + float64(bounds.Min.X)
						s := displayBuffer[i]
						if channel.Inverted {
							s = -s
						}
						y := -yScale*float64(s) + offsetFloat
						const r = 5.0
						drawCircle(targetImg, float32(x), float32(y), r, col)
						drawCircle(targetImg, float32(x), float32(y), r-1, col)
						drawCircle(targetImg, float32(x), float32(y), r-2, col)
					}
				} //drawPoint

				drawRaw := func() {
					var prevX float64 = t0 + float64(bounds.Min.X)
					s0 := displayBuffer[0]
					if channel.Inverted {
						s0 = -s0
					}
					var prevY float64 = -yScale*float64(s0) + offsetFloat

					for i := 1; i < len(displayBuffer); i++ {
						x := t0 + float64(i)*deltaT + float64(bounds.Min.X)
						s := displayBuffer[i]
						if channel.Inverted {
							s = -s
						}
						y := -yScale*float64(s) + offsetFloat

						// Horizontal segment (from prev sample to current sample x)
						drawLine(targetImg, float32(prevX), float32(prevY), float32(x), float32(prevY), col)
						// Vertical segment
						drawLine(targetImg, float32(x), float32(prevY), float32(x), float32(y), col)

						prevX, prevY = x, y
					}
				} //drawRaw

				// Linear interpolation
				drawLinear := func() {
					var prevX float64 = t0 + float64(bounds.Min.X)
					s0 := displayBuffer[0]
					if channel.Inverted {
						s0 = -s0
					}
					var prevY float64 = -yScale*float64(s0) + offsetFloat

					for i := 1; i < len(displayBuffer); i++ {
						x := t0 + float64(i)*deltaT + float64(bounds.Min.X)
						s := displayBuffer[i]
						if channel.Inverted {
							s = -s
						}
						y := -yScale*float64(s) + offsetFloat

						drawLine(targetImg, float32(prevX), float32(prevY), float32(x), float32(y), col)
						prevX, prevY = x, y
					}
				} //drawLinear

				// sinc interpolation
				drawSinc := func() {
					t := t0 // Exact pixel position of sample 0
					totalSamples := len(displayBuffer)

					sincInterpolation := func(nf float64, n int) float64 {
						a := math.Pi * (nf - float64(n))
						if a == 0 {
							return 1.0
						}
						return math.Sin(a) / a
					}

					var prevX, prevY float32
					first := true
					for x := bounds.Min.X; x < bounds.Max.X; x++ {
						// Align at pixel center (+0.5) for precise sub-pixel matching
						nf := (float64(x-bounds.Min.X) + 0.5 - t) / deltaT
						v := float64(0)
						centerN := int(math.Round(nf))
						window := 1000
						startN := centerN - window
						if startN < 0 {
							startN = 0
						}
						endN := centerN + window
						if endN > totalSamples {
							endN = totalSamples
						}
						for n := startN; n < endN; n++ {
							xn := displayBuffer[n]
							v += sincInterpolation(nf, n) * float64(xn)
						}
						if channel.Inverted {
							v = -v
						}
						ys := -yScale*v + offsetFloat
						currX := float32(x)
						currY := float32(ys)
						if !first {
							drawLine(targetImg, prevX, prevY, currX, currY, col)
						}
						prevX, prevY = currX, currY
						first = false
					}
				} //drawSinc

				switch sv.scp.Settings.Time.Interpolation {
				case settings.Dot:
					drawDot()
				case settings.Raw:
					drawRaw()
				case settings.Linear:
					drawLinear()
				case settings.Sinc:
					drawSinc()
				default:
					panic("Undefined interpolation mode")
				}
				
				if channel.Persistence {
					draw.Draw(sv.scp.ftScopeSignalScreen, bounds, sv.scp.ftPersistentLayers[channelIndex], bounds.Min, draw.Over)
				}
			}
		}
	}
}

func (sv *signalViewer) drawInspector(w, h float64, bounds image.Rectangle) {
	if sv.scp.maxScreenTime == 0 {
		return
	}

	crosscol := color.RGBA{180, 180, 180, 180}
	mx := int(sv.mouseX)
	my := int(sv.mouseY)
	for i := bounds.Min.X; i < bounds.Max.X; i++ {
		sv.scp.ftScopeFullScreen.Set(i, my, crosscol)
	}
	for i := bounds.Min.Y; i < bounds.Max.Y; i++ {
		sv.scp.ftScopeFullScreen.Set(mx, i, crosscol)
	}

	unit := w / sv.scp.maxScreenTime
	tAtCursor := (float64(sv.mouseX)-float64(bounds.Min.X))/unit - sv.scp.Settings.Time.TriggerTimeOffset

	var info []struct {
		text string
		col  color.Color
	}
	info = append(info, struct {
		text string
		col  color.Color
	}{"T: " + sv.formatTime(tAtCursor), color.White})

	moved := false
	if sv.mouseX != sv.inspectorLastX || sv.mouseY != sv.inspectorLastY {
		moved = true
		sv.inspectorLastX = sv.mouseX
		sv.inspectorLastY = sv.mouseY
	}

	if sv.inspectorSumV == nil || len(sv.inspectorSumV) != len(sv.scp.channelViewers) {
		sv.inspectorSumV = make([]float32, len(sv.scp.channelViewers))
		sv.inspectorSumVCur = make([]float32, len(sv.scp.channelViewers))
		sv.inspectorDispV = make([]float32, len(sv.scp.channelViewers))
		sv.inspectorDispVCur = make([]float32, len(sv.scp.channelViewers))
	}

	if moved {
		for i := range sv.inspectorSumV {
			sv.inspectorSumV[i] = 0
			sv.inspectorSumVCur[i] = 0
		}
		sv.inspectorSamples = 0
	}

	instV := make([]float32, len(sv.scp.channelViewers))
	instVCur := make([]float32, len(sv.scp.channelViewers))

	for channelIndex := range sv.scp.channelViewers {
		channel := &sv.scp.Settings.Channels[channelIndex]
		if channel.Enabled && len(sv.scp.displayBuffers) > channelIndex {
			displayBuffer := sv.scp.displayBuffers[channelIndex]
			if len(displayBuffer) == 0 {
				continue
			}

			var v float32
			if sv.scp.triggerSettingMsg.Mode == control.ETS {
				// Find nearest in ETS
				etsDx := w / (sv.scp.maxScreenTime * 1e15)
				startX := float64(bounds.Min.X)
				// targetTime is in femtoseconds relative to trigger point
				targetTime := (float64(sv.mouseX)-startX)/etsDx - sv.scp.Settings.Time.TriggerTimeOffset*1e15

				// Binary search or linear search in etsBuffer
				bestIdx := 0
				minDiff := math.MaxFloat64
				for i, val := range sv.scp.etsBuffer {
					diff := math.Abs(float64(val) - targetTime)
					if diff < minDiff {
						minDiff = diff
						bestIdx = i
					}
				}
				if bestIdx < len(displayBuffer) {
					v = displayBuffer[bestIdx]
				}
			} else {
				// Normal mode
				var leftPadding float64
				if sv.scp.Settings.Time.Interpolation == settings.Sinc {
					totalSamples := len(displayBuffer)
					displaySamples := totalSamples / control.SincWMultiplier
					leftPadding = float64(totalSamples-displaySamples) / 2.0
				} else { //TODO Why added 1.5?
					leftPadding = float64(control.LeftOut) + 1.5
				}

				deltaT_samples := sv.scp.controlSamplingTimeInterval
				// Time of sample 0 relative to trigger point
				t_start_of_buffer_rel_to_trigger := -leftPadding*deltaT_samples +
					float64(sv.scp.controlXRoundError) +
					float64(sv.scp.controlTriggerTimeOffset)/1e15 -
					sv.scp.Settings.Time.TriggerTimeOffset

				// Time at sample i is t_start_of_buffer_rel_to_trigger + i * deltaT_samples
				// So i = (tAtCursor - t_start_of_buffer_rel_to_trigger) / deltaT_samples
				i := (tAtCursor - t_start_of_buffer_rel_to_trigger) / deltaT_samples
				idx := int(math.Round(i))
				if idx >= 0 && idx < len(displayBuffer) {
					v = displayBuffer[idx]
				}
			}
			// Cursor voltage for this channel
			zeroOffset := float64(bounds.Min.Y) + h/2.0
			yScale := h / float64(2.0*genericps.RangeValuesMv[channel.VRange])
			channelViewer := &sv.scp.channelViewers[channelIndex]
			yOffset := float64(0)
			if channelViewer.displayOffsetInt != 0 {
				yOffset = sv.scp.offsetNToFtY(channelViewer.displayOffsetInt)
			}
			offsetFloat := zeroOffset + yOffset
			v_cursor := float32((offsetFloat - float64(sv.mouseY)) / yScale)

			instV[channelIndex] = v
			instVCur[channelIndex] = v_cursor
		}
	}

	for i := range sv.scp.channelViewers {
		sv.inspectorSumV[i] += instV[i]
		sv.inspectorSumVCur[i] += instVCur[i]
	}
	sv.inspectorSamples++

	now := time.Now()
	updateDisplay := false
	if moved || now.Sub(sv.inspectorLastUpdate) >= 500*time.Millisecond {
		updateDisplay = true
		sv.inspectorLastUpdate = now
	}

	if updateDisplay {
		for i := range sv.scp.channelViewers {
			if sv.inspectorSamples > 0 {
				sv.inspectorDispV[i] = sv.inspectorSumV[i] / float32(sv.inspectorSamples)
				sv.inspectorDispVCur[i] = sv.inspectorSumVCur[i] / float32(sv.inspectorSamples)
			}
			sv.inspectorSumV[i] = 0
			sv.inspectorSumVCur[i] = 0
		}
		sv.inspectorSamples = 0
	}

	for channelIndex := range sv.scp.channelViewers {
		channel := &sv.scp.Settings.Channels[channelIndex]
		if channel.Enabled && len(sv.scp.displayBuffers) > channelIndex && len(sv.scp.displayBuffers[channelIndex]) > 0 {
			v := sv.inspectorDispV[channelIndex]
			v_cursor := sv.inspectorDispVCur[channelIndex]
			col := channel.Col[sv.scp.Settings.ChannelColorIndex]
			info = append(info, struct {
				text string
				col  color.Color
			}{fmt.Sprintf("Ch%c: %s (Cur: %s)", 'A'+channelIndex, sv.formatVoltage(v, channel.VRange), sv.formatVoltage(v_cursor, channel.VRange)), col})
		}
	}

	// Draw the box
	lineHeight := 20
	// Calculate box width based on content
	maxW := float32(0)
	for _, item := range info {
		left, _, right, _ := sv.scp.boundString(item.text)
		if right-left > maxW {
			maxW = right - left
		}
	}
	boxWidth := int(maxW) + 15
	boxHeight := len(info)*lineHeight + 10

	x := int(sv.mouseX) + 10
	y := int(sv.mouseY)

	// Flip to left if it would touch or go beyond the right edge (with small margin)
	if x+boxWidth > bounds.Max.X-2 {
		x = int(sv.mouseX) - boxWidth - 10
	}
	// Constrain to left edge
	if x < bounds.Min.X+2 {
		x = bounds.Min.X + 2
	}

	// Ensure it's within Y bounds
	if y+boxHeight > bounds.Max.Y-2 {
		y = bounds.Max.Y - boxHeight - 2
	}
	if y < bounds.Min.Y+2 {
		y = bounds.Min.Y + 2
	}

	rect := image.Rect(x, y, x+boxWidth, y+boxHeight)
	draw.Draw(sv.scp.ftScopeFullScreen, rect, &image.Uniform{color.RGBA{20, 20, 20, 220}}, image.ZP, draw.Over)
	// Draw border (inside the box bounds)
	for i := 0; i < boxWidth; i++ {
		sv.scp.ftScopeFullScreen.Set(x+i, y, color.White)
		sv.scp.ftScopeFullScreen.Set(x+i, y+boxHeight-1, color.White)
	}
	for i := 0; i < boxHeight; i++ {
		sv.scp.ftScopeFullScreen.Set(x, y+i, color.White)
		sv.scp.ftScopeFullScreen.Set(x+boxWidth-1, y+i, color.White)
	}

	for i, item := range info {
		sv.scp.addLabel(sv.scp.ftScopeFullScreen, x+5, y+5+(i+1)*lineHeight-5, item.text, item.col)
	}
}

func (sv *signalViewer) formatVoltage(mv float32, vRange genericps.RangeEnum) string {
	if genericps.InputRanges[vRange] >= 1000 {
		return fmt.Sprintf("%.3f V", mv/1000.0)
	}
	return fmt.Sprintf("%.1f mV", mv)
}

func (sv *signalViewer) formatTime(seconds float64) string {
	absT := math.Abs(seconds)
	unit := "s"
	val := seconds

	if absT < 1e-9 {
		val *= 1e12
		unit = "ps"
	} else if absT < 1e-6 {
		val *= 1e9
		unit = "ns"
	} else if absT < 1e-3 {
		val *= 1e6
		unit = "µs"
	} else if absT < 1 {
		val *= 1e3
		unit = "ms"
	}

	return fmt.Sprintf("%.2f %s", val, unit)
}

func (scp *ScpDesc) addFtXOffset(dx float64) {
	scp.Settings.Time.TriggerTimeOffset += scp.maxScreenTime * dx / float64(scp.ftScopeSignalScreen.Bounds().Dx())
	switch {
	case scp.Settings.Time.TriggerTimeOffset < 0:
		scp.Settings.Time.TriggerTimeOffset = 0
	case scp.Settings.Time.TriggerTimeOffset > scp.maxScreenTime:
		scp.Settings.Time.TriggerTimeOffset = scp.maxScreenTime
	default:
	}
}

func (scp *ScpDesc) snapYToFtN(y float64) int {
	h := scp.ftScopeSignalScreen.Bounds().Dy()
	yRasterDiv := (float64(h) / float64(numberOfDivs)) / 5
	n := int(math.Round((y / yRasterDiv)))
	return n
}

func (scp *ScpDesc) offsetNToFtY(n int) float64 {
	h := float64(scp.ftScopeSignalScreen.Bounds().Dy())
	yRasterDiv := (h / float64(numberOfDivs)) / 5.0
	return float64(n) * yRasterDiv
}

func (scp *ScpDesc) setFtVDivsY() {
	if scp.ftScopeSignalScreen == nil {
		return
	}
	bounds := scp.ftScopeSignalScreen.Bounds()
	h := float32(bounds.Dy())
	dh := (h - 1) / numberOfDivs
	for i, y := 0, float32(bounds.Min.Y); y <= float32(bounds.Max.Y); i, y = i+1, y+dh {
		scp.ftDivsY[i] = y
	}
}

func (scp *ScpDesc) setFtHDivsX() {
	if scp.ftScopeSignalScreen == nil {
		return
	}
	bounds := scp.ftScopeSignalScreen.Bounds()
	w := float64(bounds.Dx()) - 1
	dx := (w) / float64(numberOfDivs)
	unit := w / float64(scp.maxScreenTime)
	offset := float64(scp.Settings.Time.TriggerTimeOffset) * unit
	n := float64(math.Round(float64(offset / dx)))
	offset = float64(snapN(float32(offset-n*dx), xSnapValue))
	x := float64(bounds.Min.X)
	for i := range scp.ftDivsX {
		scp.ftDivsX[i] = float32(x + offset)
		x = x + dx
	}
}

func (scp *ScpDesc) drawFtDivisions() {
	if scp.ftScopeSignalScreen == nil {
		return
	}
	bounds := scp.ftScopeSignalScreen.Bounds()
	drawDivs := func(yOffset float32, col color.Color) {
		draw.Draw(scp.ftScopeFullScreen, bounds, &image.Uniform{scp.theme.Color(ColorNameSignalBackground, 0)}, image.ZP, draw.Src)
		for _, v := range scp.ftDivsY {
			counter := 0
			for x := float64(bounds.Min.X); x <= float64(bounds.Max.X); x = x + 1.0 {
				if counter%10 < 4 {
					scp.ftScopeSignalScreen.Set(int(math.Round(x)), int(math.Round(float64(v+yOffset))), col)
				}
				counter++
			}
		}
		for _, v := range scp.ftDivsX {
			counter := 0
			for y := float64(bounds.Min.Y); y <= float64(bounds.Max.Y); y = y + 1.0 {
				if counter%10 < 4 {
					scp.ftScopeSignalScreen.Set(int(math.Round(float64(v))), int(math.Round(float64(y))), col)
				}
				counter++
			}
		}
	}
	channellIndex := scp.displayMovedDivs - 1
	col := scp.theme.Color(ColorNameDivision, 0)
	if channellIndex >= 0 {
		if scp.displayMovedDivs > 0 && scp.channelViewers[channellIndex].displayOffsetInt != 0 {
			drawDivs(0, gray)
			yOffset := scp.offsetNToFtY(scp.channelViewers[channellIndex].displayOffsetInt)
			drawDivs(float32(yOffset), scp.Settings.Channels[channellIndex].Col[scp.Settings.ChannelColorIndex])
		} else {
			drawDivs(0, col)
		}
	} else {
		drawDivs(0, col)
	}
}
func (scp *ScpDesc) clipFtChRangeScrs(w, h float32) (leftMargin, rightMargin float32) {
	numberOfEnabledChannels, _ := scp.numberOfEnabledChannels()
	if numberOfEnabledChannels == 0 {
		leftMargin = defaultLeftMargin
		rightMargin = defaultRightMargin
		return
	}
	leftColumnCount := numberOfEnabledChannels / 2
	rightColumnCount := numberOfEnabledChannels / 2
	if numberOfEnabledChannels%2 != 0 {
		leftColumnCount++
	}
	leftMargin = float32(leftColumnCount) * scp.rangeMargin
	if rightColumnCount == 0 {
		rightMargin = defaultRightMargin
	} else {
		rightMargin = float32(rightColumnCount) * scp.rangeMargin
	}
	start := float32(0)
	end := scp.rangeMargin
	for channelIndex := range scp.channelViewers {
		channel := &scp.Settings.Channels[channelIndex]
		channelViewer := &scp.channelViewers[channelIndex]
		if channel.Enabled {
			channelViewer.label = newFtChannelLabelViewer(scp.ftScopeFullScreen,
				image.Rect(int(math.Round(float64(start))), 0, int(math.Round(float64(end))), int(math.Round(float64(h-defaultTimeMargin)))),
				channelIndex, image.Rect(int(math.Round(float64(leftMargin))), defaultTopMargin,
					int(math.Round(float64(w-rightMargin))), int(math.Round(float64(h-defaultBottomMargin)))), scp)
			scp.addFtDrawer(&channelViewer.label)
			switch {
			case leftColumnCount > 1:
				channelViewer.leftLabel = true
				leftColumnCount--
				start = end
				end += scp.rangeMargin
			case leftColumnCount == 1:
				channelViewer.leftLabel = true
				leftColumnCount--
				start = w - rightMargin
				end = start + scp.rangeMargin
			default:
				channelViewer.leftLabel = false
				start = end
				end += scp.rangeMargin
			}
			channelViewer.hasScreenPartition = true
		} else {
			channelViewer.hasScreenPartition = false
		}
	}
	return
}

func (scp *ScpDesc) partitionFtScreen(w, h float32) {
	ip := scp.ftScopeFullScreen.(*image.RGBA)
	scp.ftDrawers = nil
	leftMargin, rightMargin := scp.clipFtChRangeScrs(w, h)
	scp.ftScopeSignalScreen = ip.SubImage(image.Rect(int(math.Round(float64(leftMargin))),
		defaultTopMargin, int(math.Round(float64(w-rightMargin))),
		int(math.Round(float64(h-defaultBottomMargin))))).(draw.RGBA64Image)
	scp.psControl.SetScopeScreenWidth(float64(scp.ftScopeSignalScreen.Bounds().Dx()))
	scp.ftBottomLabelViewer = newTimelLabelViewer(scp.ftScopeFullScreen,
		image.Rect(int(math.Round(0)), int(math.Round(float64(h-defaultTimeMargin))),
			int(math.Round(float64(w))), int(math.Round(float64(h)))), scp)
	scp.addFtDrawer(scp.ftBottomLabelViewer)
	scp.addFtDrawer(newSignalViewer(scp.ftScopeFullScreen, scp.ftScopeSignalScreen.Bounds(), scp))
	scp.addFtDrawer(scp.triggerPoint)
}

func (scp *ScpDesc) ftRasterGenerator(wInt int, hInt int) image.Image {
	ws := scp.Window.Canvas().Size()
	scp.Settings.Window.Height = ws.Height
	scp.Settings.Window.Width = ws.Width
	defer scp.screenLocker.Unlock()
	scp.screenLocker.Lock()
	w := float32(wInt)
	h := float32(hInt)
	rect := scp.ftScopeFullScreen.Bounds()
	if wInt != rect.Max.X-rect.Min.X || hInt != rect.Max.Y-rect.Min.Y { // window resized
		slog.Debug("RESIZED")
		scp.ftScopeFullScreen = scp.newScopeScreen(image.Point{wInt, hInt})
		if scp.triggerSettingMsg.Type == control.Simple {
			scp.triggerPoint = newTriggerPointViewer(scp.ftScopeFullScreen, scp)
		} else {
			scp.triggerPoint = newAdvTriggerPointViewer(scp.ftScopeFullScreen, scp)
		}
		rect = scp.ftScopeFullScreen.Bounds()
		w = float32(rect.Dx())
		h = float32(rect.Dy())
		draw.Draw(scp.ftScopeFullScreen, scp.ftScopeFullScreen.Bounds(), &image.Uniform{scp.theme.Color(ColorNameSignalBackground, 0)}, image.ZP, draw.Src)
		getFlag(scp.repartition) //clear flag
		scp.partitionFtScreen(w, h)
		scp.setFtVDivsY()
		scp.setFtHDivsX()
	} else if getFlag(scp.repartition) {
		slog.Debug("REPARTITION")
		if scp.triggerSettingMsg.Type == control.Simple {
			scp.triggerPoint = newTriggerPointViewer(scp.ftScopeFullScreen, scp)
		} else {
			scp.triggerPoint = newAdvTriggerPointViewer(scp.ftScopeFullScreen, scp)
		}
		scp.partitionFtScreen(w, h)
		draw.Draw(scp.ftScopeFullScreen, scp.ftScopeFullScreen.Bounds(), &image.Uniform{scp.theme.Color(ColorNameSignalBackground, 0)}, image.ZP, draw.Src)
		scp.setFtVDivsY()
		scp.setFtHDivsX()
	} else if scp.ftScopeSignalScreen == nil {
		slog.Debug("ftScopeSignalScreen == nil")
		scp.setFtVDivsY()
		scp.setFtHDivsX()
	} else if getFlag(scp.themeChanged) {
		slog.Debug("themeChanged")
		draw.Draw(scp.ftScopeFullScreen, scp.ftScopeFullScreen.Bounds(), &image.Uniform{scp.theme.Color(ColorNameSignalBackground, 0)}, image.ZP, draw.Src)
		for channelIndex := range scp.channelViewers {
			channelViewer := &scp.channelViewers[channelIndex]
			channelViewer.label.enableRefresh()
		}
		if scp.ftBottomLabelViewer != nil {
			tl := scp.ftBottomLabelViewer.(*timeLabelViewer)
			tl.enableRefresh()
		}
	} else { // display signal only
		draw.Draw(scp.ftScopeFullScreen, scp.ftScopeSignalScreen.Bounds(), &image.Uniform{scp.theme.Color(ColorNameSignalBackground, 0)}, image.ZP, draw.Src)
	}
	scp.setFtVDivsY()
	scp.setFtHDivsX()
	for i := range scp.ftDrawers {
		scp.ftDrawers[i].draw()
	}
	return scp.ftScopeFullScreen
}
