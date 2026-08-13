package gui

import (
	"encoding/binary"
	"encoding/csv"
	"fmt"
	"fynescope/control"
	"fynescope/disp7"
	"image"
	"image/color"
	"math"
	"os"
	"strconv"
	"strings"

	"fynescope/selectscroll"
	"fynescope/settings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
)

const (
	awgModeFree = "Free"
	awgModeLine = "Line"
)

type awgEditorWidget struct {
	widget.BaseWidget
	mode               string
	values             []float64 // normalized -1.0 to 1.0
	lastX, lastY       float32
	currentX, currentY float32
	drawing            bool
	raster             *canvas.Raster
	uartMode           bool
	samplesPerBit      float64
	pointColors        []color.NRGBA
}

func newAwgEditorWidget() *awgEditorWidget {
	w := &awgEditorWidget{
		mode:   awgModeFree,
		values: make([]float64, 32768),
	}
	w.raster = canvas.NewRaster(w.drawImage)
	w.ExtendBaseWidget(w)
	return w
}

func (w *awgEditorWidget) CreateRenderer() fyne.WidgetRenderer {
	return &awgEditorRenderer{w: w, raster: w.raster}
}

type awgEditorRenderer struct {
	w      *awgEditorWidget
	raster *canvas.Raster
}

func (r *awgEditorRenderer) Destroy() {}

func (r *awgEditorRenderer) Layout(size fyne.Size) {
	r.raster.Resize(size)
}

func (r *awgEditorRenderer) MinSize() fyne.Size {
	return fyne.NewSize(400, 300)
}

func (r *awgEditorRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{r.raster}
}

func (r *awgEditorRenderer) Refresh() {
	canvas.Refresh(r.raster)
}

func (w *awgEditorWidget) Tapped(e *fyne.PointEvent) {
	if w.mode == awgModeLine {
		if !w.drawing {
			w.drawing = true
			w.lastX = e.Position.X
			w.lastY = e.Position.Y
			w.currentX = e.Position.X
			w.currentY = e.Position.Y
		} else {
			w.drawLineToPoints(w.lastX, w.lastY, e.Position.X, e.Position.Y)
			w.lastX = e.Position.X
			w.lastY = e.Position.Y
			w.drawing = false
		}
		w.Refresh()
	} else if w.mode == awgModeFree {
		w.setValueAtCoord(e.Position.X, e.Position.Y)
		w.Refresh()
	}
}

func (w *awgEditorWidget) Dragged(e *fyne.DragEvent) {
	if w.mode == awgModeFree {
		if !w.drawing {
			w.drawing = true
			w.lastX = e.Position.X
			w.lastY = e.Position.Y
		}
		w.drawLineToPoints(w.lastX, w.lastY, e.Position.X, e.Position.Y)
		w.lastX = e.Position.X
		w.lastY = e.Position.Y
		w.Refresh()
	} else if w.mode == awgModeLine {
		if !w.drawing {
			w.drawing = true
			w.lastX = e.Position.X
			w.lastY = e.Position.Y
		}
		w.currentX = e.Position.X
		w.currentY = e.Position.Y
		w.Refresh()
	}
}

func (w *awgEditorWidget) DragEnd() {
	if w.mode == awgModeLine && w.drawing {
		w.drawLineToPoints(w.lastX, w.lastY, w.currentX, w.currentY)
		w.lastX = w.currentX
		w.lastY = w.currentY
	}
	w.drawing = false
	w.Refresh()
}

func (w *awgEditorWidget) setValueAtCoord(x, y float32) {
	if w.Size().Width <= 0 || w.Size().Height <= 0 {
		return
	}
	if x < 0 || x >= w.Size().Width || y < 0 || y >= w.Size().Height {
		return
	}

	val := 1.0 - (float64(y) / float64(w.Size().Height) * 2.0)

	startX := float64(x) - 0.5
	endX := float64(x) + 0.5

	startIdx := int(math.Round(startX / float64(w.Size().Width) * float64(len(w.values))))
	endIdx := int(math.Round(endX / float64(w.Size().Width) * float64(len(w.values))))

	if startIdx < 0 {
		startIdx = 0
	}
	if endIdx >= len(w.values) {
		endIdx = len(w.values) - 1
	}

	for i := startIdx; i <= endIdx; i++ {
		w.values[i] = val
	}
}

func (w *awgEditorWidget) drawLineToPoints(x0, y0, x1, y1 float32) {
	if w.Size().Width <= 0 {
		return
	}
	dist := math.Sqrt(float64((x1-x0)*(x1-x0) + (y1-y0)*(y1-y0)))
	steps := int(dist) + 1
	for i := 0; i <= steps; i++ {
		t := float32(i) / float32(steps)
		cx := x0 + (x1-x0)*t
		cy := y0 + (y1-y0)*t
		w.setValueAtCoord(cx, cy)
	}
}

func (w *awgEditorWidget) clear() {
	for i := range w.values {
		w.values[i] = 0.0
	}
	w.lastX = 0
	w.lastY = w.Size().Height / 2.0
	w.drawing = false
	w.Refresh()
}

func (w *awgEditorWidget) generateWaveform(waveType string) {
	w.uartMode = false
	w.pointColors = nil
	n := len(w.values)
	if n == 0 {
		return
	}
	for i := 0; i < n; i++ {
		t := float64(i) / float64(n)
		var val float64
		switch waveType {
		case "Sine":
			val = math.Sin(2.0 * math.Pi * t)
		case "Square":
			if t < 0.5 {
				val = 1.0
			} else {
				val = -1.0
			}
		case "Triangle":
			if t < 0.25 {
				val = t * 4.0
			} else if t < 0.75 {
				val = 1.0 - (t-0.25)*4.0
			} else {
				val = -1.0 + (t-0.75)*4.0
			}
		case "RampUp":
			val = -1.0 + t*2.0
		case "RampDown":
			val = 1.0 - t*2.0
		case "SinC":
			x := (t - 0.5) * 4.0 * math.Pi
			if x == 0 {
				val = 1.0
			} else {
				val = math.Sin(x) / x
			}
		case "Gaussian":
			x := (t - 0.5) * 4.0
			val = math.Exp(-x * x)
		case "HalfSine":
			val = math.Sin(math.Pi * t)
		case "DC":
			val = 0.0
		default:
			val = 0.0
		}
		w.values[i] = val
	}
	w.Refresh()
}

func (w *awgEditorWidget) resizeValues(newSize int) {
	if newSize == len(w.values) {
		return
	}
	newValues := make([]float64, newSize)
	for i := range newValues {
		origIdx := int(float64(i) / float64(newSize) * float64(len(w.values)))
		if origIdx >= len(w.values) {
			origIdx = len(w.values) - 1
		}
		newValues[i] = w.values[origIdx]
	}
	w.values = newValues
	w.Refresh()
}

func (w *awgEditorWidget) drawImage(width, height int) image.Image {
	if width <= 0 || height <= 0 {
		return image.NewNRGBA(image.Rect(0, 0, 1, 1))
	}
	img := image.NewNRGBA(image.Rect(0, 0, width, height))

	bgColor := color.NRGBA{R: 20, G: 20, B: 20, A: 255}
	gridColor := color.NRGBA{R: 45, G: 45, B: 45, A: 255}
	zeroLineColor := color.NRGBA{R: 80, G: 80, B: 80, A: 255}
	waveColor := color.NRGBA{R: 0, G: 220, B: 255, A: 255}
	previewColor := color.NRGBA{R: 255, G: 200, B: 0, A: 255}

	// Fast fill background
	for i := 0; i < len(img.Pix); i += 4 {
		img.Pix[i] = bgColor.R
		img.Pix[i+1] = bgColor.G
		img.Pix[i+2] = bgColor.B
		img.Pix[i+3] = bgColor.A
	}

	// Draw horizontal grid lines (8 divisions)
	for i := 1; i < 8; i++ {
		y := (height * i) / 8
		c := gridColor
		if i == 4 {
			c = zeroLineColor
		}
		for x := 0; x < width; x++ {
			if i == 4 || x%4 < 2 {
				img.SetNRGBA(x, y, c)
			}
		}
	}

	// Draw vertical grid lines or bit divisions
	if w.uartMode && w.samplesPerBit > 0 {
		bitDivColor := color.NRGBA{R: 80, G: 80, B: 150, A: 255}
		numBits := int(float64(len(w.values)) / w.samplesPerBit)
		pixelPerBit := (w.samplesPerBit / float64(len(w.values))) * float64(width)

		if pixelPerBit >= 4.0 {
			for b := 0; b <= numBits; b++ {
				sampleIdx := float64(b) * w.samplesPerBit
				x := int(sampleIdx / float64(len(w.values)) * float64(width))
				if x >= 0 && x < width {
					for y := 0; y < height; y++ {
						if y%4 < 2 {
							img.SetNRGBA(x, y, bitDivColor)
						}
					}
				}
			}
		}
	} else {
		for i := 1; i < 10; i++ {
			x := (width * i) / 10
			for y := 0; y < height; y++ {
				if y%4 < 2 {
					img.SetNRGBA(x, y, gridColor)
				}
			}
		}
	}

	// Draw waveform curve
	n := len(w.values)
	if n > 0 {
		var prevX, prevY int
		for x := 0; x < width; x++ {
			idx := int(float64(x) / float64(width) * float64(n))
			if idx >= n {
				idx = n - 1
			}
			val := w.values[idx]
			segmentColor := waveColor

			if len(w.pointColors) == n {
				segmentColor = w.pointColors[idx]
			}

			if w.mode == awgModeLine && w.drawing {
				minX, maxX := w.lastX, w.currentX
				if minX > maxX {
					minX, maxX = maxX, minX
				}
				startPixel := int(math.Round(float64(minX)))
				endPixel := int(math.Round(float64(maxX)))

				if x >= startPixel && x <= endPixel {
					if w.currentX != w.lastX {
						t := (float32(x) - w.lastX) / (w.currentX - w.lastX)
						if t < 0 {
							t = 0
						} else if t > 1 {
							t = 1
						}
						interpY := w.lastY + (w.currentY-w.lastY)*t
						val = 1.0 - (float64(interpY) / float64(height) * 2.0)
					} else {
						val = 1.0 - (float64(w.currentY) / float64(height) * 2.0)
					}
					segmentColor = previewColor
				}
			}

			y := int((1.0 - val) / 2.0 * float64(height))
			if y < 0 {
				y = 0
			} else if y >= height {
				y = height - 1
			}

			if x > 0 {
				drawLineBresenham(img, prevX, prevY, x, y, segmentColor)
			} else {
				img.SetNRGBA(x, y, segmentColor)
			}
			prevX = x
			prevY = y
		}
	}

	return img
}

func drawLineBresenham(img *image.NRGBA, x0, y0, x1, y1 int, col color.NRGBA) {
	dx := x1 - x0
	if dx < 0 {
		dx = -dx
	}
	dy := y1 - y0
	if dy < 0 {
		dy = -dy
	}
	sx := -1
	if x0 < x1 {
		sx = 1
	}
	sy := -1
	if y0 < y1 {
		sy = 1
	}
	err := dx - dy

	w := img.Bounds().Max.X
	h := img.Bounds().Max.Y

	for {
		if x0 >= 0 && x0 < w && y0 >= 0 && y0 < h {
			img.SetNRGBA(x0, y0, col)
			if y0+1 < h {
				img.SetNRGBA(x0, y0+1, col)
			}
		}

		if x0 == x1 && y0 == y1 {
			break
		}
		e2 := 2 * err
		if e2 > -dy {
			err -= dy
			x0 += sx
		}
		if e2 < dx {
			err += dx
			y0 += sy
		}
	}
}

func (scp *ScpDesc) showAwgEditor(applyCb func([]int16)) {
	if scp.awgWindow != nil {
		scp.awgWindow.Show()
		return
	}

	scp.awgWindow = scp.App.NewWindow("AWG Waveform Editor")
	scp.awgWindow.Resize(fyne.NewSize(800, 600))

	editor := newAwgEditorWidget()

	modeSelect := widget.NewRadioGroup([]string{awgModeFree, awgModeLine}, func(mode string) {
		editor.mode = mode
		editor.drawing = false
		editor.Refresh()
	})
	modeSelect.SetSelected(awgModeFree)
	modeSelect.Horizontal = true

	var pointsDisp *disp7.DigitArray

	options := []string{"Sine", "Square", "Triangle", "RampUp", "RampDown", "SinC", "Gaussian", "HalfSine", "DC", "UART"}

	waveformFile := settings.WaveformFileName(scp.SettingFileName)
	if _, err := os.Stat(waveformFile); err == nil {
		options = append(options, "Last Waveform")
	}

	for i := range scp.Settings.DemoGenPanel {
		demoWfFile := settings.WaveformDemoFileName(scp.SettingFileName, i)
		if _, err := os.Stat(demoWfFile); err == nil {
			options = append(options, fmt.Sprintf("Last Demo Waveform %d", i+1))
		}
	}

	chNames := []string{"Ch A", "Ch B", "Ch C", "Ch D"}
	for i, ch := range scp.Settings.Channels {
		if ch.Enabled {
			options = append(options, chNames[i])
		}
	}

	var selectedWaveformType string

	waveformSelect := selectscroll.NewSelectScroll(options, func(selected string, ex selectscroll.Exception) {
		selectedWaveformType = selected
		if selected == "Last Waveform" {
			if wfData, err := os.ReadFile(waveformFile); err == nil {
				wfLen := len(wfData) / 2
				if pointsDisp != nil {
					pointsDisp.SetValue(wfLen)
				}
				editor.resizeValues(wfLen)
				for i := 0; i < wfLen; i++ {
					v := int16(binary.LittleEndian.Uint16(wfData[i*2 : i*2+2]))
					if i < len(editor.values) {
						editor.values[i] = float64(v) / 32767.0
					}
				}
				editor.Refresh()
			}
		} else if strings.HasPrefix(selected, "Last Demo Waveform") {
			var ch int
			if _, err := fmt.Sscanf(selected, "Last Demo Waveform %d", &ch); err == nil {
				demoWfFile := settings.WaveformDemoFileName(scp.SettingFileName, ch-1)
				if wfData, err := os.ReadFile(demoWfFile); err == nil {
					wfLen := len(wfData) / 2
					if pointsDisp != nil {
						pointsDisp.SetValue(wfLen)
					}
					editor.resizeValues(wfLen)
					for i := 0; i < wfLen; i++ {
						v := int16(binary.LittleEndian.Uint16(wfData[i*2 : i*2+2]))
						if i < len(editor.values) {
							editor.values[i] = float64(v) / 32767.0
						}
					}
					editor.Refresh()
				}
			}
		} else if strings.HasPrefix(selected, "Ch ") {
			chIdx := -1
			switch selected {
			case "Ch A":
				chIdx = 0
			case "Ch B":
				chIdx = 1
			case "Ch C":
				chIdx = 2
			case "Ch D":
				chIdx = 3
			}
			if chIdx >= 0 && chIdx < len(scp.displayBuffers) {
				buf := scp.displayBuffers[chIdx]
				n := len(buf)
				if n > 0 {
					var leftPadding float64
					var extra float64
					if scp.Settings.Time.Interpolation == settings.Sinc {
						displaySamples := n / control.SincWMultiplier
						leftPadding = float64(n-displaySamples) / 2.0
						extra = 0
					} else {
						leftPadding = float64(control.LeftOut)
						extra = 1
					}

					triggerOffsetSeconds := float64(scp.controlTriggerTimeOffset) / 1e15
					expectedOffset := leftPadding * scp.controlSamplingTimeInterval
					if math.Abs(triggerOffsetSeconds-expectedOffset) > 10*scp.controlSamplingTimeInterval {
						triggerOffsetSeconds = expectedOffset
					}

					offsetTime := -leftPadding*scp.controlSamplingTimeInterval + scp.controlXRoundError + triggerOffsetSeconds - extra*scp.controlSamplingTimeInterval

					startIndex := int(math.Ceil(-offsetTime / scp.controlSamplingTimeInterval))
					endIndex := int(math.Floor((scp.maxScreenTime - offsetTime) / scp.controlSamplingTimeInterval))

					if startIndex < 0 {
						startIndex = 0
					}
					if endIndex >= n {
						endIndex = n - 1
					}

					if startIndex <= endIndex {
						visBuf := buf[startIndex : endIndex+1]
						visN := len(visBuf)

						minVal := float32(math.MaxFloat32)
						maxVal := float32(-math.MaxFloat32)
						for _, v := range visBuf {
							if v < minVal {
								minVal = v
							}
							if v > maxVal {
								maxVal = v
							}
						}

						if pointsDisp != nil {
							pointsDisp.SetValue(visN)
						}
						editor.resizeValues(visN)

						if maxVal > minVal {
							diff := float64(maxVal - minVal)
							for i, v := range visBuf {
								if i < len(editor.values) {
									editor.values[i] = float64(v-minVal)/diff*2.0 - 1.0
								}
							}
						} else {
							for i := range visBuf {
								if i < len(editor.values) {
									editor.values[i] = 0
								}
							}
						}
						editor.Refresh()
					}
				}
			}
		} else if selected == "UART" {
			editor.clear()
		} else {
			editor.generateWaveform(selected)
		}
	}, "Select Waveform")

	clearBtn := widget.NewButton("Clear", func() {
		editor.clear()
		waveformSelect.ClearSelected()
	})

	importCsvBtn := widget.NewButton("Import CSV", func() {
		fd := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
			if err != nil || reader == nil {
				return
			}
			defer reader.Close()

			csvReader := csv.NewReader(reader)
			// Handle potentially varying number of fields per record
			csvReader.FieldsPerRecord = -1
			records, err := csvReader.ReadAll()
			if err != nil {
				dialog.ShowError(err, scp.awgWindow)
				return
			}

			editor.values = nil
			for _, row := range records {
				for _, col := range row {
					val, err := strconv.ParseFloat(strings.TrimSpace(col), 64)
					if err == nil {
						editor.values = append(editor.values, val)
						break
					}
				}
			}
			if len(editor.values) > 0 {
				if pointsDisp != nil {
					pointsDisp.SetValue(len(editor.values))
				}
				editor.Refresh()
			}
		}, scp.awgWindow)
		fd.SetFilter(storage.NewExtensionFileFilter([]string{".csv"}))
		if cwd, err := os.Getwd(); err == nil {
			if lister, err := storage.ListerForURI(storage.NewFileURI(cwd)); err == nil {
				fd.SetLocation(lister)
			}
		}
		fd.Show()
	})

	exportCsvBtn := widget.NewButton("Export CSV", func() {
		fd := dialog.NewFileSave(func(writer fyne.URIWriteCloser, err error) {
			if err != nil || writer == nil {
				return
			}
			defer writer.Close()

			csvWriter := csv.NewWriter(writer)
			for _, val := range editor.values {
				strVal := strconv.FormatFloat(val, 'f', -1, 64)
				err := csvWriter.Write([]string{strVal})
				if err != nil {
					dialog.ShowError(err, scp.awgWindow)
					return
				}
			}
			csvWriter.Flush()
			if err := csvWriter.Error(); err != nil {
				dialog.ShowError(err, scp.awgWindow)
			}
		}, scp.awgWindow)
		fd.SetFilter(storage.NewExtensionFileFilter([]string{".csv"}))
		fd.SetFileName("waveform.csv")
		if cwd, err := os.Getwd(); err == nil {
			if lister, err := storage.ListerForURI(storage.NewFileURI(cwd)); err == nil {
				fd.SetLocation(lister)
			}
		}
		fd.Show()
	})

	patternGenBtn := widget.NewButton("Pattern Gen", func() {
		if scp.patternWindow != nil {
			scp.patternWindow.Show()
			return
		}

		if selectedWaveformType == "UART" {
			scp.showUartPatternGenerator(editor)
			return
		}

		scp.patternWindow = scp.App.NewWindow("Pattern Generator")
		scp.patternWindow.Resize(fyne.NewSize(500, 250))

		binEntry := widget.NewEntry()
		hexEntry := widget.NewEntry()
		highEntry := widget.NewEntry()
		highEntry.SetText("100")
		lowEntry := widget.NewEntry()
		lowEntry.SetText("0")

		updatingBin := false
		updatingHex := false

		updateWaveform := func() {
			binStr := ""
			for _, c := range binEntry.Text {
				if c == '0' || c == '1' {
					binStr += string(c)
				}
			}
			if len(binStr) == 0 {
				return
			}

			hlPct, errH := strconv.ParseFloat(highEntry.Text, 64)
			llPct, errL := strconv.ParseFloat(lowEntry.Text, 64)
			if errH != nil || errL != nil {
				return
			}
			hl := hlPct / 100.0
			ll := llPct / 100.0

			N := len(editor.values)
			L := len(binStr)
			for i := 0; i < N; i++ {
				bitIdx := (i * L) / N
				if binStr[bitIdx] == '1' {
					editor.values[i] = hl
				} else {
					editor.values[i] = ll
				}
			}
			editor.Refresh()
		}

		binToHex := func(bin string) string {
			cleanBin := ""
			for _, c := range bin {
				if c == '0' || c == '1' {
					cleanBin += string(c)
				}
			}
			if len(cleanBin) == 0 {
				return ""
			}
			pad := len(cleanBin) % 4
			if pad != 0 {
				cleanBin = strings.Repeat("0", 4-pad) + cleanBin
			}
			var hexBuilder strings.Builder
			for i := 0; i < len(cleanBin); i += 4 {
				chunk := cleanBin[i : i+4]
				val, _ := strconv.ParseUint(chunk, 2, 8)
				hexBuilder.WriteString(fmt.Sprintf("%X", val))
			}
			return hexBuilder.String()
		}

		hexToBin := func(hexStr string) string {
			cleanHex := ""
			for _, c := range hexStr {
				if (c >= '0' && c <= '9') || (c >= 'A' && c <= 'F') || (c >= 'a' && c <= 'f') {
					cleanHex += string(c)
				}
			}
			var binBuilder strings.Builder
			for _, c := range cleanHex {
				val, _ := strconv.ParseUint(string(c), 16, 8)
				binBuilder.WriteString(fmt.Sprintf("%04b", val))
			}
			return binBuilder.String()
		}

		binEntry.OnChanged = func(s string) {
			if updatingHex || updatingBin {
				return
			}
			updatingBin = true
			cleanBin := ""
			for _, c := range s {
				if c == '0' || c == '1' {
					cleanBin += string(c)
				}
			}
			if s != cleanBin {
				binEntry.SetText(cleanBin)
			}
			hexEntry.SetText(binToHex(cleanBin))
			updatingBin = false
			updateWaveform()
		}

		hexEntry.OnChanged = func(s string) {
			if updatingBin || updatingHex {
				return
			}
			updatingHex = true
			cleanHex := ""
			for _, c := range s {
				if (c >= '0' && c <= '9') || (c >= 'A' && c <= 'F') || (c >= 'a' && c <= 'f') {
					cleanHex += string(c)
				}
			}
			if s != cleanHex {
				hexEntry.SetText(cleanHex)
			}
			binEntry.SetText(hexToBin(cleanHex))
			updatingHex = false
			updateWaveform()
		}

		highEntry.OnChanged = func(s string) {
			updateWaveform()
		}

		lowEntry.OnChanged = func(s string) {
			updateWaveform()
		}

		form := widget.NewForm(
			widget.NewFormItem("Binary Input", binEntry),
			widget.NewFormItem("Hex Input", hexEntry),
			widget.NewFormItem("High Level (%)", highEntry),
			widget.NewFormItem("Low Level (%)", lowEntry),
		)

		closeBtn := widget.NewButton("Close", func() {
			scp.patternWindow.Close()
		})

		content := container.NewBorder(nil, container.NewHBox(layout.NewSpacer(), closeBtn), nil, nil, form)
		scp.patternWindow.SetContent(content)
		scp.patternWindow.SetOnClosed(func() {
			scp.patternWindow = nil
		})
		scp.patternWindow.Show()
	})

	var err error
	pointsDisp, err = disp7.NewCustomDisp7Array(5, 0, 32768, 10,
		disp7.UnSigned, disp7.NoTrailingZeroes, scp.awgWindow,
		scp.theme.Color(ColorNameGeneratorDisp, 0),
		disp7.ReadWrite, 1.0*disp7.DefaultDigitWidth,
		disp7.DeafultDigitHeight, 1,
		disp7.DefaultVCursorSpace, "Points: ", "")
	_ = err
	pointsDisp.SilentSetValue(32768)
	pointsDisp.OnChanged = func(val float64) {
		editor.resizeValues(int(val))
	}

	applyBtn := widget.NewButton("Apply to Generator", func() {
		waveform := make([]int16, len(editor.values))
		for i, v := range editor.values {
			waveform[i] = int16(v * 32767.0)
		}
		if applyCb != nil {
			applyCb(waveform)
		}
	})
	applyBtn.Importance = widget.HighImportance

	toolbar := container.NewHBox(
		widget.NewLabel("Draw Mode:"),
		modeSelect,
		widget.NewLabel("Waveform:"),
		waveformSelect,
		clearBtn,
		importCsvBtn,
		exportCsvBtn,
		patternGenBtn,
		pointsDisp,
	)

	bottomBar := container.NewHBox(
		applyBtn,
	)

	content := container.NewBorder(toolbar, bottomBar, nil, nil, editor)
	scp.awgWindow.SetContent(content)

	scp.awgWindow.SetOnClosed(func() {
		scp.awgWindow = nil
		if scp.patternWindow != nil {
			scp.patternWindow.Close()
			scp.patternWindow = nil
		}
	})

	scp.awgWindow.Show()
}

func (scp *ScpDesc) showUartPatternGenerator(editor *awgEditorWidget) {
	scp.patternWindow = scp.App.NewWindow("UART Pattern Generator")
	scp.patternWindow.Resize(fyne.NewSize(500, 400))

	textEntry := widget.NewEntry()
	textEntry.MultiLine = true
	highEntry := widget.NewEntry()
	highEntry.SetText("100")
	lowEntry := widget.NewEntry()
	lowEntry.SetText("0")
	suggestedFreqLabel := widget.NewLabel("N/A")
	var updateWaveform func()

	baudrates := []string{"300", "600", "1200", "2400", "4800", "9600", "14400", "19200", "38400", "57600", "115200", "128000", "256000"}
	baudrateSelect := selectscroll.NewSelectScroll(baudrates, func(s string, ex selectscroll.Exception) {
		if updateWaveform != nil {
			updateWaveform()
		}
	}, "Select Baudrate")
	baudrateSelect.SetSelected("115200")

	bitNumsSelect := selectscroll.NewSelectScroll([]string{"5", "6", "7", "8", "9"}, func(s string, ex selectscroll.Exception) {
		if updateWaveform != nil {
			updateWaveform()
		}
	}, "Data Bits")
	bitNumsSelect.SetSelected("8")
	paritySelect := selectscroll.NewSelectScroll([]string{"None", "Even", "Odd", "Mark", "Space"}, func(s string, ex selectscroll.Exception) {
		if updateWaveform != nil {
			updateWaveform()
		}
	}, "Parity")
	paritySelect.SetSelected("None")
	stopBitsSelect := selectscroll.NewSelectScroll([]string{"1", "1.5", "2"}, func(s string, ex selectscroll.Exception) {
		if updateWaveform != nil {
			updateWaveform()
		}
	}, "Stop Bits")
	stopBitsSelect.SetSelected("1")

	updateWaveform = func() {
		hlPct, errH := strconv.ParseFloat(highEntry.Text, 64)
		llPct, errL := strconv.ParseFloat(lowEntry.Text, 64)
		if errH != nil || errL != nil {
			return
		}
		hl := hlPct / 100.0
		ll := llPct / 100.0

		bitrate, errB := strconv.ParseFloat(baudrateSelect.Selected, 64)
		if errB != nil || bitrate <= 0 {
			return
		}

		dataBits, _ := strconv.Atoi(bitNumsSelect.Selected)
		if dataBits == 0 {
			dataBits = 8
		}

		stopBits := 1.0
		switch stopBitsSelect.Selected {
		case "1.5":
			stopBits = 1.5
		case "2":
			stopBits = 2.0
		}

		parity := paritySelect.Selected
		text := textEntry.Text
		N := len(editor.values)

		editor.pointColors = make([]color.NRGBA, N)
		defaultColor := color.NRGBA{R: 0, G: 220, B: 255, A: 255}
		stopBitColor := color.NRGBA{R: 255, G: 100, B: 100, A: 255}

		for i := 0; i < N; i++ {
			editor.values[i] = hl // Idle state is high
			editor.pointColors[i] = defaultColor
		}

		if len(text) == 0 {
			editor.Refresh()
			return
		}

		parityBits := 0.0
		if parity != "None" {
			parityBits = 1.0
		}
		charBits := 1.0 + float64(dataBits) + parityBits + stopBits
		totalBits := float64(len(text)) * charBits

		var samplesPerBit float64
		if totalBits > 0 {
			// Ensure at least 50% idle time or 20 bits of idle time so the oscilloscope can trigger cleanly between sequences
			paddedBits := totalBits * 2.0
			if paddedBits < totalBits+20.0 {
				paddedBits = totalBits + 20.0
			}
			S := int(float64(N) / paddedBits)
			if S >= 1 {
				suggestedFreq := (float64(S) * bitrate) / float64(N)
				suggestedFreqLabel.SetText(fmt.Sprintf("%.4f Hz", suggestedFreq))
				samplesPerBit = float64(S)
			} else {
				suggestedFreqLabel.SetText("Text too long")
				samplesPerBit = 1.0
			}
		} else {
			suggestedFreqLabel.SetText("N/A")
			samplesPerBit = 1.0
		}

		for i := 0; i < N; i++ {
			bitIdxFloat := float64(i) / samplesPerBit

			charIdx := int(bitIdxFloat / charBits)
			if charIdx >= len(text) {
				continue // Idle high
			}

			intraCharBit := bitIdxFloat - float64(charIdx)*charBits
			c := text[charIdx]
			val := hl
			ptColor := defaultColor

			if intraCharBit < 1.0 {
				// Start bit
				val = ll
			} else if intraCharBit < 1.0+float64(dataBits) {
				// Data bit (LSB first)
				bitPos := int(intraCharBit - 1.0)
				bitVal := (c >> bitPos) & 1
				if bitVal == 1 {
					val = hl
				} else {
					val = ll
				}
			} else if intraCharBit < 1.0+float64(dataBits)+parityBits {
				// Parity bit
				onesCount := 0
				for b := 0; b < dataBits; b++ {
					if (c>>b)&1 == 1 {
						onesCount++
					}
				}
				parityBit := false
				switch parity {
				case "Even":
					parityBit = (onesCount%2 != 0)
				case "Odd":
					parityBit = (onesCount%2 == 0)
				case "Mark":
					parityBit = true
				case "Space":
					parityBit = false
				}
				if parityBit {
					val = hl
				} else {
					val = ll
				}
			} else {
				// Stop bits
				val = hl
				ptColor = stopBitColor
			}

			editor.values[i] = val
			editor.pointColors[i] = ptColor
		}
		editor.uartMode = true
		editor.samplesPerBit = samplesPerBit
		editor.Refresh()
	}

	textEntry.OnChanged = func(s string) { updateWaveform() }
	highEntry.OnChanged = func(s string) { updateWaveform() }
	lowEntry.OnChanged = func(s string) { updateWaveform() }

	form := widget.NewForm(
		widget.NewFormItem("Text", textEntry),
		widget.NewFormItem("High Level (%)", highEntry),
		widget.NewFormItem("Low Level (%)", lowEntry),
		widget.NewFormItem("Baudrate", baudrateSelect),
		widget.NewFormItem("Data Bits", bitNumsSelect),
		widget.NewFormItem("Parity", paritySelect),
		widget.NewFormItem("Stop Bits", stopBitsSelect),
		widget.NewFormItem("Suggested AWG Freq", suggestedFreqLabel),
	)

	closeBtn := widget.NewButton("Close", func() {
		scp.patternWindow.Close()
	})

	content := container.NewBorder(nil, container.NewHBox(layout.NewSpacer(), closeBtn), nil, nil, form)
	scp.patternWindow.SetContent(content)
	scp.patternWindow.SetOnClosed(func() {
		scp.patternWindow = nil
	})
	scp.patternWindow.Show()
}
