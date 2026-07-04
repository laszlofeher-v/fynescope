package gui

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"log/slog"
	"math"
	"time"

	"fynescope/genericps"
	"fynescope/settings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/driver/desktop"
)

type (
	fvViewer struct {
		rasterPartition
		scp                 *ScpDesc
		labelBounds         map[int]image.Rectangle
		selectedChannel     int
		showInspector       bool
		mouseX, mouseY      float32
		inspectorLastX      float32
		inspectorLastY      float32
		inspectorSumV       []float32
		inspectorSumVCur    []float32
		inspectorDispV      []float32
		inspectorDispVCur   []float32
		inspectorSamples    int
		inspectorLastUpdate time.Time
	}
)

func newFvViewer(img rasterImage, imgRect image.Rectangle, scp *ScpDesc) *fvViewer {
	fv := &fvViewer{rasterPartition: rasterPartition{img: img, imgRect: imgRect, refreshFlag: true},
		scp: scp, labelBounds: make(map[int]image.Rectangle), selectedChannel: -1}
	return fv
}

func (fv *fvViewer) draw() {
	if !fv.scp.shouldDrawRaster(fvTabIndex) {
		return
	}
	bounds := fv.scp.fvScopeSignalScreen.Bounds()
	draw.Draw(fv.scp.fvScopeFullScreen, fv.imgRect, &image.Uniform{fv.scp.theme.Color(ColorNameSignalBackground, 0)}, image.ZP, draw.Src)

	// Draw grid
	gridCol := fv.scp.theme.Color(ColorNameDivision, 0)
	w := float64(bounds.Dx())
	h := float64(bounds.Dy())

	// Find X channel (argument)
	var xCh int = -1
	for i := 0; i < int(fv.scp.channelCount); i++ {
		ch := fv.scp.Settings.Channels[i]
		if ch.FvMode == settings.FvArgument && ch.Enabled {
			xCh = i
			break
		}
	}

	drawDivs := func(xf, yf float64, col color.Color) {
		// Vertical dotted line
		ixf := int(math.Round(xf))
		if ixf >= bounds.Min.X && ixf <= bounds.Max.X {
			counter := 0
			for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
				if counter%10 < 4 {
					fv.scp.fvScopeFullScreen.Set(ixf, y, col)
				}
				counter++
			}
		}
		// Horizontal dotted line
		iyf := int(math.Round(yf))
		if iyf >= bounds.Min.Y && iyf <= bounds.Max.Y {
			counter := 0
			for x := bounds.Min.X; x < bounds.Max.X; x++ {
				if counter%10 < 4 {
					fv.scp.fvScopeFullScreen.Set(x, iyf, col)
				}
				counter++
			}
		}
	}

	fv.labelBounds = make(map[int]image.Rectangle)

	if xCh == -1 {
		// No X channel, draw default grid
		for i := 0; i <= numberOfDivs; i++ {
			xf := float64(bounds.Min.X) + float64(i)*w/float64(numberOfDivs)
			yf := float64(bounds.Min.Y) + float64(i)*h/float64(numberOfDivs)
			drawDivs(xf, yf, gridCol)
		}
		return
	}

	xRange := genericps.RangeValuesMv[fv.scp.Settings.Channels[xCh].VRange]
	xScale := w / (2.0 * float64(xRange))
	xOffset := fv.offsetNToFv(fv.scp.Settings.Channels[xCh].DisplayVOffset)
	xZero := float64(bounds.Min.X) + w/2.0 + xOffset
	xCol := fv.scp.Settings.Channels[xCh].Col[fv.scp.Settings.ChannelColorIndex]

	fv.labelBounds[xCh] = image.Rect(bounds.Min.X, bounds.Max.Y, bounds.Max.X, bounds.Max.Y+40)

	// Draw grid and X labels
	for i := 0; i <= numberOfDivs; i++ {
		xf := float64(bounds.Min.X) + float64(i)*w/float64(numberOfDivs) + xOffset
		yf := float64(bounds.Min.Y) + float64(i)*h/float64(numberOfDivs)
		drawDivs(xf, yf, gridCol)

		// X labels
		vx := -float64(xRange) + float64(i)*2.0*float64(xRange)/float64(numberOfDivs)
		vstr := fv.formatVoltage(float32(vx), fv.scp.Settings.Channels[xCh].VRange)
		left, _, right, _ := fv.scp.boundString(vstr)
		if xf >= float64(bounds.Min.X) && xf <= float64(bounds.Max.X) {
			fv.scp.addLabel(fv.scp.fvScopeFullScreen, int(math.Round(xf-float64(right-left)/2)), bounds.Max.Y+20, vstr, xCol)
		}
	}

	yLabelOffsetLeft := 0
	yLabelOffsetRight := 0
	yCount := 0
	for i := 0; i < int(fv.scp.channelCount); i++ {
		ch := fv.scp.Settings.Channels[i]
		if ch.FvMode == settings.FvValue && ch.Enabled {
			yRange := genericps.RangeValuesMv[ch.VRange]
			col := ch.Col[fv.scp.Settings.ChannelColorIndex]
			yOffset := fv.offsetNToFv(ch.DisplayVOffset)

			var minX, maxX int
			if yCount%2 == 0 {
				minX = bounds.Min.X - 60 - yLabelOffsetLeft
				maxX = bounds.Min.X - yLabelOffsetLeft
			} else {
				minX = bounds.Max.X + yLabelOffsetRight
				maxX = bounds.Max.X + 60 + yLabelOffsetRight
			}
			fv.labelBounds[i] = image.Rect(minX, bounds.Min.Y, maxX, bounds.Max.Y)

			// Draw Y labels
			for j := 0; j <= numberOfDivs; j++ {
				yf := float64(bounds.Min.Y) + float64(j)*h/float64(numberOfDivs) + yOffset
				vy := float64(yRange) - float64(j)*2.0*float64(yRange)/float64(numberOfDivs)
				vstr := fv.formatVoltage(float32(vy), ch.VRange)
				left, top, right, bottom := fv.scp.boundString(vstr)

				var lx int
				if yCount%2 == 0 {
					// Left side
					lx = bounds.Min.X - int(right-left) - 5 - yLabelOffsetLeft
				} else {
					// Right side
					lx = bounds.Max.X + 5 + yLabelOffsetRight
				}
				if yf >= float64(bounds.Min.Y) && yf <= float64(bounds.Max.Y) {
					fv.scp.addLabel(fv.scp.fvScopeFullScreen, lx, int(math.Round(yf-float64(top-bottom)/2)), vstr, col)
				}
			}

			if yCount%2 == 0 {
				yLabelOffsetLeft += 60
			} else {
				yLabelOffsetRight += 60
			}
			yCount++
		}
	}

	// Plot all Y channels (values)
	xBuffer := fv.scp.displayBuffers[xCh]
	if len(xBuffer) == 0 {
		return
	}

	for i := 0; i < int(fv.scp.channelCount); i++ {
		ch := fv.scp.Settings.Channels[i]
		if ch.FvMode == settings.FvValue && ch.Enabled {
			yBuffer := fv.scp.displayBuffers[i]
			if len(yBuffer) == 0 {
				continue
			}

			yRange := genericps.RangeValuesMv[ch.VRange]
			yScale := h / (2.0 * float64(yRange))
			yOffset := fv.offsetNToFv(ch.DisplayVOffset)
			yZero := float64(bounds.Min.Y) + h/2.0 + yOffset
			col := ch.Col[fv.scp.Settings.ChannelColorIndex]

			samples := len(xBuffer)
			if len(yBuffer) < samples {
				samples = len(yBuffer)
			}

			if samples < 2 {
				continue
			}

			prevX := xZero + float64(xBuffer[0])*xScale
			prevY := yZero - float64(yBuffer[0])*yScale
			for s := 1; s < samples; s++ {
				currX := xZero + float64(xBuffer[s])*xScale
				currY := yZero - float64(yBuffer[s])*yScale
				drawLine(fv.scp.fvScopeSignalScreen, float32(prevX), float32(prevY), float32(currX), float32(currY), col)
				prevX, prevY = currX, currY
			}
		}
	}

	if fv.showInspector {
		fv.drawInspector(w, h, bounds)
	}
}

func (fv *fvViewer) drawInspector(w, h float64, bounds image.Rectangle) {
	// Find X channel (argument)
	var xCh int = -1
	for i := 0; i < int(fv.scp.channelCount); i++ {
		ch := fv.scp.Settings.Channels[i]
		if ch.FvMode == settings.FvArgument && ch.Enabled {
			xCh = i
			break
		}
	}

	crosscol := color.RGBA{180, 180, 180, 180}
	mx := int(fv.mouseX)
	my := int(fv.mouseY)
	for i := bounds.Min.X; i < bounds.Max.X; i++ {
		fv.scp.fvScopeFullScreen.Set(i, my, crosscol)
	}
	for i := bounds.Min.Y; i < bounds.Max.Y; i++ {
		fv.scp.fvScopeFullScreen.Set(mx, i, crosscol)
	}

	if xCh == -1 {
		return
	}

	// Calculate cursor X voltage
	xRange := genericps.RangeValuesMv[fv.scp.Settings.Channels[xCh].VRange]
	xScale := w / (2.0 * float64(xRange))
	xOffset := fv.offsetNToFv(fv.scp.Settings.Channels[xCh].DisplayVOffset)
	xZero := float64(bounds.Min.X) + w/2.0 + xOffset

	v_cursor_x := float32((float64(fv.mouseX) - xZero) / xScale)

	var info []struct {
		text string
		col  color.Color
	}

	moved := false
	if fv.mouseX != fv.inspectorLastX || fv.mouseY != fv.inspectorLastY {
		moved = true
		fv.inspectorLastX = fv.mouseX
		fv.inspectorLastY = fv.mouseY
	}

	if fv.inspectorSumV == nil || len(fv.inspectorSumV) != len(fv.scp.channelViewers) {
		fv.inspectorSumV = make([]float32, len(fv.scp.channelViewers))
		fv.inspectorSumVCur = make([]float32, len(fv.scp.channelViewers))
		fv.inspectorDispV = make([]float32, len(fv.scp.channelViewers))
		fv.inspectorDispVCur = make([]float32, len(fv.scp.channelViewers))
	}

	if moved {
		for i := range fv.inspectorSumV {
			fv.inspectorSumV[i] = 0
			fv.inspectorSumVCur[i] = 0
		}
		fv.inspectorSamples = 0
	}

	xBuffer := fv.scp.displayBuffers[xCh]
	if len(xBuffer) == 0 {
		return
	}

	bestIdx := 0
	minDiffX := math.MaxFloat64
	for s := 0; s < len(xBuffer); s++ {
		cx := xZero + float64(xBuffer[s])*xScale
		diff := math.Abs(cx - float64(fv.mouseX))
		if diff < minDiffX {
			minDiffX = diff
			bestIdx = s
		}
	}

	instV := make([]float32, len(fv.scp.channelViewers))
	instVCur := make([]float32, len(fv.scp.channelViewers))

	instV[xCh] = xBuffer[bestIdx]
	instVCur[xCh] = v_cursor_x

	for channelIndex := 0; channelIndex < int(fv.scp.channelCount); channelIndex++ {
		if channelIndex == xCh {
			continue
		}
		channel := &fv.scp.Settings.Channels[channelIndex]
		if channel.FvMode == settings.FvValue && channel.Enabled && len(fv.scp.displayBuffers) > channelIndex {
			displayBuffer := fv.scp.displayBuffers[channelIndex]
			if len(displayBuffer) == 0 || bestIdx >= len(displayBuffer) {
				continue
			}

			yRange := genericps.RangeValuesMv[channel.VRange]
			yScale := h / (2.0 * float64(yRange))
			yOffset := fv.offsetNToFv(channel.DisplayVOffset)
			yZero := float64(bounds.Min.Y) + h/2.0 + yOffset

			v_cursor := float32((yZero - float64(fv.mouseY)) / yScale)

			instV[channelIndex] = displayBuffer[bestIdx]
			instVCur[channelIndex] = v_cursor
		}
	}

	for i := range fv.scp.channelViewers {
		fv.inspectorSumV[i] += instV[i]
		fv.inspectorSumVCur[i] += instVCur[i]
	}
	fv.inspectorSamples++

	now := time.Now()
	updateDisplay := false
	if moved || now.Sub(fv.inspectorLastUpdate) >= 500*time.Millisecond {
		updateDisplay = true
		fv.inspectorLastUpdate = now
	}

	if updateDisplay {
		for i := range fv.scp.channelViewers {
			if fv.inspectorSamples > 0 {
				fv.inspectorDispV[i] = fv.inspectorSumV[i] / float32(fv.inspectorSamples)
				fv.inspectorDispVCur[i] = fv.inspectorSumVCur[i] / float32(fv.inspectorSamples)
			}
			fv.inspectorSumV[i] = 0
			fv.inspectorSumVCur[i] = 0
		}
		fv.inspectorSamples = 0
	}

	xCol := fv.scp.Settings.Channels[xCh].Col[fv.scp.Settings.ChannelColorIndex]
	info = append(info, struct {
		text string
		col  color.Color
	}{fmt.Sprintf("Ch%c(X): %s (Cur: %s)", 'A'+xCh, fv.formatVoltage(fv.inspectorDispV[xCh], fv.scp.Settings.Channels[xCh].VRange), fv.formatVoltage(fv.inspectorDispVCur[xCh], fv.scp.Settings.Channels[xCh].VRange)), xCol})

	for channelIndex := 0; channelIndex < int(fv.scp.channelCount); channelIndex++ {
		if channelIndex == xCh {
			continue
		}
		channel := &fv.scp.Settings.Channels[channelIndex]
		if channel.FvMode == settings.FvValue && channel.Enabled && len(fv.scp.displayBuffers) > channelIndex && len(fv.scp.displayBuffers[channelIndex]) > 0 {
			v := fv.inspectorDispV[channelIndex]
			v_cursor := fv.inspectorDispVCur[channelIndex]
			col := channel.Col[fv.scp.Settings.ChannelColorIndex]
			info = append(info, struct {
				text string
				col  color.Color
			}{fmt.Sprintf("Ch%c(Y): %s (Cur: %s)", 'A'+channelIndex, fv.formatVoltage(v, channel.VRange), fv.formatVoltage(v_cursor, channel.VRange)), col})
		}
	}

	lineHeight := 20
	maxW := float32(0)
	for _, item := range info {
		left, _, right, _ := fv.scp.boundString(item.text)
		if right-left > maxW {
			maxW = right - left
		}
	}
	boxWidth := int(maxW) + 15
	boxHeight := len(info)*lineHeight + 10

	xBox := int(fv.mouseX) + 10
	yBox := int(fv.mouseY)

	if xBox+boxWidth > bounds.Max.X-2 {
		xBox = int(fv.mouseX) - boxWidth - 10
	}
	if xBox < bounds.Min.X+2 {
		xBox = bounds.Min.X + 2
	}
	if yBox+boxHeight > bounds.Max.Y-2 {
		yBox = bounds.Max.Y - boxHeight - 2
	}
	if yBox < bounds.Min.Y+2 {
		yBox = bounds.Min.Y + 2
	}

	rect := image.Rect(xBox, yBox, xBox+boxWidth, yBox+boxHeight)
	draw.Draw(fv.scp.fvScopeFullScreen, rect, &image.Uniform{color.RGBA{20, 20, 20, 220}}, image.ZP, draw.Over)
	for i := 0; i < boxWidth; i++ {
		fv.scp.fvScopeFullScreen.Set(xBox+i, yBox, color.White)
		fv.scp.fvScopeFullScreen.Set(xBox+i, yBox+boxHeight-1, color.White)
	}
	for i := 0; i < boxHeight; i++ {
		fv.scp.fvScopeFullScreen.Set(xBox, yBox+i, color.White)
		fv.scp.fvScopeFullScreen.Set(xBox+boxWidth-1, yBox+i, color.White)
	}

	for i, item := range info {
		fv.scp.addLabel(fv.scp.fvScopeFullScreen, xBox+8, yBox+10+i*lineHeight+15, item.text, item.col)
	}
}

func (fv *fvViewer) formatVoltage(mv float32, vRange genericps.RangeEnum) string {
	if genericps.RangeValuesMv[vRange] >= 1000 {
		return fmt.Sprintf("%.1fV", mv/1000.0)
	}
	return fmt.Sprintf("%.0fmV", mv)
}

func (fv *fvViewer) mouseInSignalScreen(x, y float32) bool {
	p := image.Point{X: int(math.Round(float64(x))), Y: int(math.Round(float64(y)))}
	return p.In(fv.scp.fvScopeSignalScreen.Bounds())
}

func (fv *fvViewer) cursor(x, y float32) (desktop.Cursor, bool) {
	if fv.mouseInSignalScreen(x, y) {
		return desktop.CrosshairCursor, true
	}
	return desktop.DefaultCursor, false
}

func (fv *fvViewer) mouseIn(x, y float32) bool {
	p := image.Point{X: int(math.Round(float64(x))), Y: int(math.Round(float64(y)))}
	for _, bounds := range fv.labelBounds {
		if p.In(bounds) {
			return true
		}
	}
	return false
}

func (fv *fvViewer) mouseDown(button desktop.MouseButton, x, y float32) {
	if button == desktop.RightMouseButton && fv.mouseInSignalScreen(x, y) {
		fv.showInspector = true
		fv.mouseX = x
		fv.mouseY = y
		fv.enableRefresh()
		canvas.Refresh(fv.scp.fvRaster)
		return
	}
	p := image.Point{X: int(math.Round(float64(x))), Y: int(math.Round(float64(y)))}
	fv.selectedChannel = -1
	for chIdx, bounds := range fv.labelBounds {
		if p.In(bounds) {
			if button == desktop.MouseButtonSecondary || button == desktop.RightMouseButton {
				channelViewer := &fv.scp.channelViewers[chIdx]
				channelViewer.displayOffsetFraction = 0
				channelViewer.displayOffsetInt = 0
				fv.scp.Settings.Channels[chIdx].DisplayVOffset = 0
				channelViewer.label.enableRefresh()
				channelViewer.dftLabel.enableRefresh()
				fv.scp.refreshRasters()
				return
			}
			fv.selectedChannel = chIdx
			break
		}
	}
}

func (fv *fvViewer) mouseUp(button desktop.MouseButton, x, y float32) {
	if button == desktop.RightMouseButton {
		fv.showInspector = false
		fv.enableRefresh()
		canvas.Refresh(fv.scp.fvRaster)
		return
	}
	fv.selectedChannel = -1
}

func (fv *fvViewer) mouseMoved(x, y float32) {
	if fv.showInspector {
		fv.mouseX = x
		fv.mouseY = y
		bounds := fv.scp.fvScopeSignalScreen.Bounds()
		if fv.mouseX < float32(bounds.Min.X) {
			fv.mouseX = float32(bounds.Min.X)
		}
		if fv.mouseX > float32(bounds.Max.X-1) {
			fv.mouseX = float32(bounds.Max.X - 1)
		}
		if fv.mouseY < float32(bounds.Min.Y) {
			fv.mouseY = float32(bounds.Min.Y)
		}
		if fv.mouseY > float32(bounds.Max.Y-1) {
			fv.mouseY = float32(bounds.Max.Y - 1)
		}
		fv.enableRefresh()
		canvas.Refresh(fv.scp.fvRaster)
	}
}

func (fv *fvViewer) typedKey(x, y float32, keyName fyne.KeyName) {
	p := image.Point{X: int(math.Round(float64(x))), Y: int(math.Round(float64(y)))}
	for chIdx, bounds := range fv.labelBounds {
		if p.In(bounds) {
			if fv.scp.Settings.Channels[chIdx].FvMode == settings.FvArgument {
				switch keyName {
				case fyne.KeyLeft:
					fv.scrolled(-scrollDelta, x, y)
				case fyne.KeyRight:
					fv.scrolled(scrollDelta, x, y)
				}
			} else {
				switch keyName {
				case fyne.KeyDown:
					fv.scrolled(-scrollDelta, x, y)
				case fyne.KeyUp:
					fv.scrolled(scrollDelta, x, y)
				}
			}
			break
		}
	}
}

func (fv *fvViewer) offsetNToFv(n int) float64 {
	h := float64(fv.scp.fvScopeSignalScreen.Bounds().Dy())
	yRasterDiv := (h / float64(numberOfDivs)) / 5.0
	return float64(n) * yRasterDiv
}

func (fv *fvViewer) snapYToFvN(y float64) int {
	h := float64(fv.scp.fvScopeSignalScreen.Bounds().Dy())
	yRasterDiv := (h / float64(numberOfDivs)) / 5.0
	return int(math.Round(y / yRasterDiv))
}

func (fv *fvViewer) setChDispOffset(chIndex int, dy float64, scroll bool) {
	h := float64(fv.img.Bounds().Dy())
	channelViewer := &fv.scp.channelViewers[chIndex]
	if scroll {
		channelViewer.displayOffsetFraction = dy + fv.offsetNToFv(fv.scp.Settings.Channels[chIndex].DisplayVOffset)
	} else {
		channelViewer.displayOffsetFraction += dy
	}
	if channelViewer.displayOffsetFraction < -h {
		channelViewer.displayOffsetFraction = -h
	}
	if channelViewer.displayOffsetFraction > h {
		channelViewer.displayOffsetFraction = h
	}
	channelViewer.displayOffsetInt = fv.snapYToFvN(channelViewer.displayOffsetFraction)
	fv.scp.Settings.Channels[chIndex].DisplayVOffset = channelViewer.displayOffsetInt

	channelViewer.label.enableRefresh()
	channelViewer.dftLabel.enableRefresh()
	fv.scp.refreshRasters()
}

func (fv *fvViewer) dragged(dx, dy, x, y float32) {
	if fv.showInspector {
		fv.mouseX = x
		fv.mouseY = y
		bounds := fv.scp.fvScopeSignalScreen.Bounds()
		if fv.mouseX < float32(bounds.Min.X) {
			fv.mouseX = float32(bounds.Min.X)
		}
		if fv.mouseX > float32(bounds.Max.X-1) {
			fv.mouseX = float32(bounds.Max.X - 1)
		}
		if fv.mouseY < float32(bounds.Min.Y) {
			fv.mouseY = float32(bounds.Min.Y)
		}
		if fv.mouseY > float32(bounds.Max.Y-1) {
			fv.mouseY = float32(bounds.Max.Y - 1)
		}
		fv.enableRefresh()
		canvas.Refresh(fv.scp.fvRaster)
	}
	if fv.selectedChannel != -1 {
		// X-axis label drag vs Y-axis label drag
		if fv.scp.Settings.Channels[fv.selectedChannel].FvMode == settings.FvArgument {
			fv.setChDispOffset(fv.selectedChannel, float64(dx), false)
		} else {
			fv.setChDispOffset(fv.selectedChannel, float64(dy), false)
		}
	}
}

func (fv *fvViewer) scrolled(delta, x, y float32) {
	p := image.Point{X: int(math.Round(float64(x))), Y: int(math.Round(float64(y)))}
	for chIdx, bounds := range fv.labelBounds {
		if p.In(bounds) {
			nY := (float64(fv.img.Bounds().Dy()) / float64(numberOfDivs)) / 10
			if fv.scp.Settings.Channels[chIdx].FvMode == settings.FvArgument {
				fv.setChDispOffset(chIdx, float64(delta)*nY, true)
			} else {
				fv.setChDispOffset(chIdx, float64(-delta)*nY, true)
			}
			break
		}
	}
}

func (scp *ScpDesc) fvRasterGenerator(wInt int, hInt int) image.Image {
	ws := scp.Window.Canvas().Size()
	scp.Settings.Window.Height = ws.Height
	scp.Settings.Window.Width = ws.Width
	slog.Debug("fvRasterGenerator called", "w", wInt, "h", hInt)
	defer scp.screenLocker.Unlock()
	scp.screenLocker.Lock()

	if scp.fvScopeFullScreen == nil || scp.fvScopeFullScreen.Bounds().Dx() != wInt || scp.fvScopeFullScreen.Bounds().Dy() != hInt {
		scp.fvScopeFullScreen = scp.newScopeScreen(image.Point{wInt, hInt})
		if scp.ftScopeSignalScreen == nil {
			scp.updateAcquisitionParameters()
		}
	}

	w, h := float64(wInt), float64(hInt)
	yCountTotal := 0
	for i := 0; i < int(scp.channelCount); i++ {
		ch := scp.Settings.Channels[i]
		if ch.FvMode == settings.FvValue && ch.Enabled {
			yCountTotal++
		}
	}
	leftCols := (yCountTotal + 1) / 2
	rightCols := yCountTotal / 2

	// Define margins
	leftMargin := float64(defaultLeftMargin + float64(leftCols)*60.0)
	rightMargin := float64(defaultRightMargin + float64(rightCols)*60.0)
	topMargin := float64(defaultTopMargin)
	bottomMargin := float64(defaultBottomMargin)

	ip := scp.fvScopeFullScreen.(*image.RGBA)

	wSigMax := w - leftMargin - rightMargin
	hSigMax := h - topMargin - bottomMargin

	sigDim := math.Min(wSigMax, hSigMax)
	if sigDim < 0 {
		sigDim = 0
	}

	// Center the square horizontally in the available signal area, but keep it at the top vertically
	xOffset := (wSigMax - sigDim) / 2

	scp.fvScopeSignalScreen = ip.SubImage(image.Rect(
		int(leftMargin+xOffset),
		int(topMargin),
		int(leftMargin+xOffset+sigDim),
		int(topMargin+sigDim),
	)).(draw.RGBA64Image)

	// Set rect for drawers
	for _, d := range scp.fvDrawers {
		d.setRect(image.Rect(0, 0, wInt, hInt))
		d.draw()
	}
	return scp.fvScopeFullScreen
}
