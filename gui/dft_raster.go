package gui

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"log/slog"
	"time"

	"fynescope/genericps"
	"fynescope/settings"
	"math"
	"math/cmplx"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"gonum.org/v1/gonum/dsp/fourier"
	"gonum.org/v1/gonum/dsp/window"
)

type dftViewer struct {
	rasterPartition
	scp                 *ScpDesc
	selected            bool
	showInspector       bool
	mouseX, mouseY      float32
	inspectorLastX      float32
	inspectorLastY      float32
	magnitudesCache     [][]float64
	mCache              int
	fsCache             float64
	inspectorSumV       []float64
	inspectorSumVCur    []float64
	inspectorDispV      []float64
	inspectorDispVCur   []float64
	inspectorSamples    int
	inspectorLastUpdate time.Time

	// Reference point state for interval measurement
	refActive bool
	refX      float32
	refY      float32
}

var (
	fft       *fourier.FFT
	samples   []float64
	fftResult []complex128
	m         int
)

type (
	frqLabelViewer struct {
		rasterPartition
		scp      *ScpDesc
		selected bool
	}
)

func niceStep(span float64) float64 {
	if span <= 0 || math.IsNaN(span) || math.IsInf(span, 0) {
		return 1
	}
	exp := math.Floor(math.Log10(span))
	frac := span / math.Pow(10, exp)
	var niceFrac float64
	switch {
	case frac < 1.5:
		niceFrac = 1
	case frac < 3.5:
		niceFrac = 2
	case frac < 7.5:
		niceFrac = 5
	default:
		niceFrac = 10
	}
	return niceFrac * math.Pow(10, exp)
}

var (
	_ mouser     = (*frqLabelViewer)(nil)
	_ dragger    = (*frqLabelViewer)(nil)
	_ scroller   = (*frqLabelViewer)(nil)
	_ keyer      = (*frqLabelViewer)(nil)
	_ cursorable = (*frqLabelViewer)(nil)
	_ drawer     = (*frqLabelViewer)(nil)
)

func (frql *frqLabelViewer) cursor(x, y float32) (desktop.Cursor, bool) {
	if frql.mousIn(x, y) {
		return desktop.PointerCursor, true
	}
	return desktop.DefaultCursor, false
}

func (frql *frqLabelViewer) mouseMoved(x, y float32) {
}
func (frql *frqLabelViewer) mousIn(x, y float32) bool {
	p := image.Point{X: int(math.Round(float64(x))), Y: int(math.Round(float64(y)))}
	if p.In(frql.rect()) {
		return true
	}
	return false
}
func (frql *frqLabelViewer) mouseDown(button desktop.MouseButton, modifier fyne.KeyModifier, x, y float32) {
	frql.selected = frql.mousIn(x, y)
}
func (tl *frqLabelViewer) mouseUp(button desktop.MouseButton, modifier fyne.KeyModifier, x, y float32) {
	tl.selected = false
}

func (frql *frqLabelViewer) setDispFreqOffset(dx float32) {
	span := frql.scp.Settings.Dft.MaxFreq - frql.scp.Settings.Dft.MinFreq
	if span < 0 {
		span = 0
	}
	w := float32(frql.scp.dftScopeSignalScreen.Bounds().Dx()) - 1
	if w > 0 {
		freqDelta := (float64(-dx) / float64(w)) * span
		newMin := frql.scp.Settings.Dft.MinFreq + freqDelta
		newMax := frql.scp.Settings.Dft.MaxFreq + freqDelta
		fs := 1.0 / float64(frql.scp.psControl.SamplingTimeInterval)
		maxFreqAvailable := fs / 2

		if newMin < 0 {
			newMin = 0
			newMax = span
		}
		if newMax > maxFreqAvailable {
			newMax = maxFreqAvailable
			newMin = newMax - span
		}
		if newMin < 0 {
			newMin = 0
		}

		frql.scp.Settings.Dft.MinFreq = newMin
		frql.scp.Settings.Dft.MaxFreq = newMax

		if frql.scp.dftMinFreqDisp != nil {
			frql.scp.dftMinFreqDisp.SilentSetFloatValue(newMin, 2)
		}
		if frql.scp.dftMaxFreqDisp != nil {
			frql.scp.dftMaxFreqDisp.SilentSetFloatValue(newMax, 2)
		}

		frql.scp.setDftHDivsX()
		frql.scp.clearAllDftPersistentLayers()
		frql.enableRefresh()
		frql.scp.refreshRasters()
		frql.scp.SaveSettings()
	}
}

func (frql *frqLabelViewer) dragged(dx, dy, x, y float32) {
	if frql.selected {
		frql.setDispFreqOffset(dx)
	}
}

func (frql *frqLabelViewer) scrolled(delta, x, y float32) {
	if !frql.mousIn(x, y) {
		return
	}
	nX := (float32(frql.scp.dftScopeSignalScreen.Bounds().Dx()) / float32(numberOfDivs)) / 10
	frql.setDispFreqOffset(delta * nX)
}

func (frql *frqLabelViewer) typedKey(x, y float32, keyName fyne.KeyName) {
	switch keyName {
	case fyne.KeyLeft:
		frql.scrolled(-scrollDelta, x, y)
	case fyne.KeyRight:
		frql.scrolled(scrollDelta, x, y)
	}
}

func (frql *frqLabelViewer) clear() {
	draw.Draw(frql.img, frql.rect(), &image.Uniform{theme.BackgroundColor()}, image.ZP, draw.Src)
}

func (frql *frqLabelViewer) draw() {
	if !frql.refreshFlag {
		return
	}
	if !frql.scp.shouldDrawRaster(dftTabIndex) {
		return
	}
	bounds := frql.scp.dftScopeSignalScreen.Bounds()
	w := float32(bounds.Dx()) - 1
	if w < 1 {
		return
	}
	l, t, r, b := frql.scp.boundString("100M")
	maxLblWidth := r - l
	lblHeight := b - t

	// Frequency steps: avoid overlapping by calculating based on width
	labelSpacing := maxLblWidth + 10 // add some padding between labels
	if labelSpacing < 10 {
		labelSpacing = 50 // fallback
	}

	numDivs := int(w / labelSpacing)
	if numDivs > 10 {
		numDivs = 10
	}
	if numDivs < 2 {
		numDivs = 2
	}
	labelBounds := bounds
	labelBounds.Min.Y = bounds.Max.Y
	labelBounds.Max.Y += int(math.Ceil(float64(lblHeight))) + 8
	labelBounds.Min.X -= int(math.Ceil(float64(maxLblWidth)))
	labelBounds.Max.X += int(math.Ceil(float64(maxLblWidth)))
	draw.Draw(frql.scp.dftScopeFullScreen, labelBounds,
		&image.Uniform{frql.scp.theme.Color(ColorNameSignalBackground, 0)},
		image.ZP, draw.Src)

	minFreq, maxFreqPlot := frql.scp.getDftFreqRange()

	if !frql.scp.Settings.Dft.XAxisLog {
		if numDivs <= 0 {
			numDivs = 1
		}
		step := niceStep(maxFreqPlot / float64(numDivs))
		firstFreq := math.Floor(minFreq/step) * step

		for i := 0; i < 20; i++ { // Draw up to 20 potential labels
			freq := firstFreq + float64(i)*step
			if freq < 0 {
				continue
			}

			fraction := (freq - minFreq) / maxFreqPlot
			x := float32(bounds.Min.X) + float32(fraction)*w

			if x < float32(bounds.Min.X)-maxLblWidth/2 {
				continue
			}
			if x > float32(bounds.Max.X)+maxLblWidth/2 {
				break
			}

			label := formatFreq(freq)
			lblL, _, lblR, _ := frql.scp.boundString(label)
			lblW := lblR - lblL

			frql.scp.addLabel(frql.scp.dftScopeFullScreen, int(x-lblW/2),
				bounds.Max.Y+int(math.Ceil(float64(-t)))+4, label, theme.ForegroundColor())
		}
	} else {
		logMin := math.Log10(math.Max(minFreq, 1e-6))
		maxFreq := minFreq + maxFreqPlot
		logMax := math.Log10(math.Max(maxFreq, math.Max(minFreq, 1e-6)*1.001))
		logRange := logMax - logMin

		getX := func(f float64) float32 {
			if f <= 0 {
				f = 1e-6
			}
			return float32(float64(bounds.Min.X) + ((math.Log10(f)-logMin)/logRange)*float64(w))
		}

		startDecade := int(math.Floor(logMin))
		endDecade := int(math.Ceil(logMax))

		for dec := startDecade; dec <= endDecade; dec++ {
			base := math.Pow(10, float64(dec))
			for j := 1; j < 10; j++ {
				freq := base * float64(j)
				if freq < minFreq-1e-6 || freq > minFreq+maxFreqPlot+1e-6 {
					continue
				}

				x := getX(freq)

				if x < float32(bounds.Min.X)-maxLblWidth/2 {
					continue
				}
				if x > float32(bounds.Max.X)+maxLblWidth/2 {
					break
				}

				if j == 1 {
					label := formatFreq(freq)
					lblL, _, lblR, _ := frql.scp.boundString(label)
					lblW := lblR - lblL

					frql.scp.addLabel(frql.scp.dftScopeFullScreen, int(x-lblW/2),
						bounds.Max.Y+int(math.Ceil(float64(-t)))+4, label, theme.ForegroundColor())
				}
			}
		}
	}
	frql.disableRefresh()
}

func newFrqLabelViewer(img rasterImage, imgRect image.Rectangle, scp *ScpDesc) *frqLabelViewer {
	frql := &frqLabelViewer{rasterPartition: rasterPartition{img: img, imgRect: imgRect, refreshFlag: true},
		scp: scp}
	return frql
}
func newDftViewer(img rasterImage, imgRect image.Rectangle, scp *ScpDesc) *dftViewer {
	return &dftViewer{
		rasterPartition: rasterPartition{img: img, imgRect: imgRect, refreshFlag: true},
		scp:             scp,
		magnitudesCache: make([][]float64, 4),
	}
}

var (
	_ mouser     = (*dftViewer)(nil)
	_ dragger    = (*dftViewer)(nil)
	_ scroller   = (*dftViewer)(nil)
	_ keyer      = (*dftViewer)(nil)
	_ cursorable = (*dftViewer)(nil)
	_ drawer     = (*dftViewer)(nil)
)

func (dv *dftViewer) cursor(x, y float32) (desktop.Cursor, bool) {
	if dv.mousIn(x, y) {
		return desktop.CrosshairCursor, true
	}
	return desktop.DefaultCursor, false
}

func (dv *dftViewer) mouseMoved(x, y float32) {
	if dv.showInspector {
		dv.mouseX = x
		dv.mouseY = y
		if dv.mouseX < float32(dv.imgRect.Min.X) {
			dv.mouseX = float32(dv.imgRect.Min.X)
		}
		if dv.mouseX > float32(dv.imgRect.Max.X-1) {
			dv.mouseX = float32(dv.imgRect.Max.X - 1)
		}
		if dv.mouseY < float32(dv.imgRect.Min.Y) {
			dv.mouseY = float32(dv.imgRect.Min.Y)
		}
		if dv.mouseY > float32(dv.imgRect.Max.Y-1) {
			dv.mouseY = float32(dv.imgRect.Max.Y - 1)
		}
		dv.enableRefresh()
		canvas.Refresh(dv.scp.dftRaster)
	}
}

func (dv *dftViewer) mousIn(x, y float32) bool {
	p := image.Point{X: int(math.Round(float64(x))), Y: int(math.Round(float64(y)))}
	return p.In(dv.rect())
}

func (dv *dftViewer) mouseDown(button desktop.MouseButton, modifier fyne.KeyModifier, x, y float32) {
	if button == desktop.RightMouseButton && dv.mousIn(x, y) {
		if modifier&fyne.KeyModifierShift != 0 {
			dv.refActive = true
			dv.refX = x
			dv.refY = y
		} else {
			dv.showInspector = true
		}
		dv.mouseX = x
		dv.mouseY = y
		dv.enableRefresh()
		canvas.Refresh(dv.scp.dftRaster)
		return
	}
	dv.selected = dv.mousIn(x, y)
}

func (dv *dftViewer) mouseUp(button desktop.MouseButton, modifier fyne.KeyModifier, x, y float32) {
	if button == desktop.RightMouseButton {
		dv.showInspector = false
		dv.enableRefresh()
		canvas.Refresh(dv.scp.dftRaster)
		return
	}
	dv.selected = false
}

func (dv *dftViewer) dragged(dx, dy, x, y float32) {
	if dv.showInspector {
		dv.mouseX = x
		dv.mouseY = y
		if dv.mouseX < float32(dv.imgRect.Min.X) {
			dv.mouseX = float32(dv.imgRect.Min.X)
		}
		if dv.mouseX > float32(dv.imgRect.Max.X-1) {
			dv.mouseX = float32(dv.imgRect.Max.X - 1)
		}
		if dv.mouseY < float32(dv.imgRect.Min.Y) {
			dv.mouseY = float32(dv.imgRect.Min.Y)
		}
		if dv.mouseY > float32(dv.imgRect.Max.Y-1) {
			dv.mouseY = float32(dv.imgRect.Max.Y - 1)
		}
		dv.enableRefresh()
		canvas.Refresh(dv.scp.dftRaster)
	}
	if dv.selected {
		dv.scp.dftBottomLabelViewer.(*frqLabelViewer).setDispFreqOffset(dx)
	}
}

func (dv *dftViewer) scrolled(delta, x, y float32) {
}

func (dv *dftViewer) typedKey(x, y float32, keyName fyne.KeyName) {
	if keyName == fyne.KeyDelete && dv.mousIn(x, y) {
		dv.refActive = false
		dv.enableRefresh()
		canvas.Refresh(dv.scp.dftRaster)
	}
	dv.scp.dftBottomLabelViewer.(*frqLabelViewer).typedKey(x, y, keyName)
}
func (scp *ScpDesc) snapYToDftN(y float64) int {
	h := scp.dftScopeSignalScreen.Bounds().Dy()
	yRasterDiv := (float64(h) / float64(numberOfDivs)) / 5
	n := int(math.Round((y / yRasterDiv)))
	return n
}

func (dv *dftViewer) draw() {
	if !dv.scp.shouldDrawRaster(dftTabIndex) {
		return
	}

	fs := 1.0 / float64(dv.scp.psControl.SamplingTimeInterval) // Sampling frequency in Hz
	maxFreqAvailable := fs / 2
	maxFreqPlot := dv.scp.Settings.Dft.MaxFreq
	if maxFreqPlot > maxFreqAvailable {
		maxFreqPlot = maxFreqAvailable
	}

	bounds := dv.scp.dftScopeSignalScreen.Bounds()
	h := float32(bounds.Dy())
	w := float32(bounds.Dx()) - 1
	if w < 1 {
		return
	}

	// Draw divisions (optional, or simplified)
	dv.scp.drawDftDivisions()
	totalChannels := int(dv.scp.channelCount) + len(dv.scp.Settings.VirtualChannels)
	if len(dv.magnitudesCache) < totalChannels {
		newCache := make([][]float64, totalChannels)
		copy(newCache, dv.magnitudesCache)
		dv.magnitudesCache = newCache
	}
	
	for chIdx := 0; chIdx < totalChannels; chIdx++ {
		var enabled bool
		if chIdx < int(dv.scp.channelCount) {
			enabled = dv.scp.Settings.Channels[chIdx].Enabled
		} else {
			enabled = dv.scp.Settings.VirtualChannels[chIdx-int(dv.scp.channelCount)].Enabled
		}
		
		if !enabled {
			if chIdx < len(dv.magnitudesCache) {
				dv.magnitudesCache[chIdx] = nil
			}
			continue
		}

		displayBuffer := dv.scp.displayBuffers[chIdx]
		if len(displayBuffer) < 2 {
			slog.Debug("dftdraw", "chIdx", chIdx, "len", len(displayBuffer))
			continue
		}
		if m == 0 || fft == nil || len(samples) != m || len(fftResult) != m/2+1 {
			bins := dv.scp.Settings.Dft.Bins
			if bins < 128 {
				bins = 128
				dv.scp.Settings.Dft.Bins = bins
			}
			m = bins * 2
			fft = fourier.NewFFT(m)
			samples = make([]float64, m)
			fftResult = make([]complex128, m/2+1)
		}

		nsig := len(displayBuffer)
		if nsig > m {
			nsig = m
		}

		for i := range samples {
			if i < nsig {
				samples[i] = float64(displayBuffer[i])
			} else {
				// Zero padding
				samples[i] = 0
			}
		}

		applyWindow(samples[:nsig], dv.scp.Settings.Dft.Window)
		fftResult = fft.Coefficients(fftResult, samples)
		magnitudes := make([]float64, m/2)
		const dbFloor = -100.0

		// Determine max voltage range for the channel to normalize visualization
		// genericps.InputRanges[channel.VRange] gives the full range in mV (e.g. 10000 for +/- 5V?)
		// Actually inputRanges string array has values like "±5V", but genericps.InputRanges is an array of ints.
		// Let's assume genericps.InputRanges[channel.VRange] is the max voltage in mV.
		// Wait, looking at gui.go: adcToMv uses genericps.InputRanges.
		// Let's use the channel's set range as the full scale for display.
		// The FFT magnitude is normalized such that a pure sine wave of amplitude A has peak magnitude A.

		normFactor := float64(nsig) / 2.0
		normFactor *= dv.scp.getCoherentGain(dv.scp.Settings.Dft.Window, nsig)

		var vRange genericps.RangeEnum
		var col color.NRGBA
		var dftDisplayOffsetInt int
		var dftPersistence bool
		if chIdx < int(dv.scp.channelCount) {
			ch := &dv.scp.Settings.Channels[chIdx]
			vRange = ch.VRange
			col = ch.Col[dv.scp.Settings.ChannelColorIndex]
			dftDisplayOffsetInt = dv.scp.channelViewers[chIdx].dftDisplayOffsetInt
			dftPersistence = ch.DftPersistence
		} else {
			vchIdx := chIdx - int(dv.scp.channelCount)
			vch := &dv.scp.Settings.VirtualChannels[vchIdx]
			vRange = vch.VRange
			col = vch.Col[dv.scp.Settings.ChannelColorIndex]
			dftPersistence = false
			if vchIdx < len(dv.scp.ftVChannelLabels) {
				dftDisplayOffsetInt = dv.scp.snapYToDftN(dv.scp.ftVChannelLabels[vchIdx].displayOffsetFraction)
			} else {
				dftDisplayOffsetInt = 0
			}
		}

		yScale := 1.0 / float32(genericps.RangeValuesMv[vRange])
		for i := 0; i < m/2; i++ {
			mag := cmplx.Abs(fftResult[i]) / normFactor // Magnitude in mV (since input samples are in mV)
			val := float64(float32(mag) * yScale)

			if dv.scp.Settings.Dft.DisplayMode == settings.ModeVoltage {
				magnitudes[i] = val
			} else {
				// dB plot
				if mag < 1e-10 { // Avoid log(0)
					magnitudes[i] = dbFloor
				} else {
					var db float64
					if dv.scp.Settings.Dft.DisplayMode == settings.ModeDBFS {
						db = 20 * math.Log10(val)
					} else {
						vPeak := val * float64(genericps.RangeValuesMv[vRange]) / 1000.0
						vRms := vPeak / math.Sqrt(2)

						switch dv.scp.Settings.Dft.DisplayMode {
						case settings.ModeDBV:
							db = 20 * math.Log10(vRms/1.0)
						case settings.ModeDBU:
							db = 20 * math.Log10(vRms/0.7746)
						case settings.ModeDBM:
							db = 20 * math.Log10(vRms/0.2236)
						case settings.ModeArbitraryDB:
							ref := dv.scp.Settings.Dft.ArbitraryDbRefV
							if ref <= 0 {
								ref = 1e-6
							}
							db = 20 * math.Log10(vPeak/ref)
						default:
							db = 20 * math.Log10(val)
						}
					}

					if db < dbFloor {
						db = dbFloor
					}
					magnitudes[i] = db
				}
			}
		}
		dv.magnitudesCache[chIdx] = magnitudes

		yOffset := dv.scp.offsetNToDftY(dftDisplayOffsetInt)
		prevX := float32(bounds.Min.X)

		minFreq, maxFreqPlot := dv.scp.getDftFreqRange()
		fs := 1.0 / float64(dv.scp.psControl.SamplingTimeInterval) // Sampling frequency in Hz
		maxFreqAvailable := fs / 2

		minBinIdx := int(math.Round((minFreq / maxFreqAvailable) * float64(m/2)))
		if minBinIdx < 0 {
			minBinIdx = 0
		}

		var targetImg rasterImage = dv.scp.dftScopeSignalScreen
		if dftPersistence {
			if chIdx >= len(dv.scp.dftPersistentLayers) {
				newLayers := make([]*image.RGBA, chIdx+1)
				copy(newLayers, dv.scp.dftPersistentLayers)
				dv.scp.dftPersistentLayers = newLayers
			}
			if dv.scp.dftPersistentLayers[chIdx] == nil || dv.scp.dftPersistentLayers[chIdx].Bounds() != bounds {
				dv.scp.dftPersistentLayers[chIdx] = image.NewRGBA(bounds)
			}
			targetImg = dv.scp.dftPersistentLayers[chIdx]
		}

		var startY float32
		if dv.scp.Settings.Dft.DisplayMode == settings.ModeVoltage {
			startY = float32(float64(bounds.Min.Y) + (1.0-magnitudes[minBinIdx])*float64(h) + yOffset)
		} else {
			startY = float32(float64(bounds.Min.Y) + (magnitudes[minBinIdx]/dbFloor)*float64(h) + yOffset)
		}

		maxBinIdxPlot := int(math.Round(((minFreq + maxFreqPlot) / maxFreqAvailable) * float64(m/2)))
		if maxBinIdxPlot > m/2 {
			maxBinIdxPlot = m / 2
		}
		if maxBinIdxPlot <= minBinIdx {
			maxBinIdxPlot = minBinIdx + 1
		}

		prevY := startY
		logMin := math.Log10(math.Max(minFreq, 1e-6))
		maxFreq := minFreq + maxFreqPlot
		logMax := math.Log10(math.Max(maxFreq, math.Max(minFreq, 1e-6)*1.001))
		logRange := logMax - logMin

		getX := func(f float64) float32 {
			if f <= 0 {
				f = 1e-6
			}
			return float32(float64(bounds.Min.X) + ((math.Log10(f)-logMin)/logRange)*float64(w))
		}

		for i := minBinIdx; i < maxBinIdxPlot; i++ {
			if i >= len(magnitudes) {
				break
			}
			binFreq := float64(i) * (maxFreqAvailable / float64(m/2))
			var x float32
			if dv.scp.Settings.Dft.XAxisLog {
				x = getX(binFreq)
			} else {
				fraction := (binFreq - minFreq) / maxFreqPlot
				x = float32(bounds.Min.X) + float32(fraction)*w
			}

			var y float32

			if dv.scp.Settings.Dft.DisplayMode == settings.ModeVoltage {
				y = float32(float64(bounds.Min.Y) + (1.0-magnitudes[i])*float64(h) + yOffset)
			} else {
				y = float32(float64(bounds.Min.Y) + (magnitudes[i]/dbFloor)*float64(h) + yOffset)
			}
			if i > minBinIdx {
				drawLine(targetImg, prevX, prevY, x, y, col)
			}
			prevX = x
			prevY = y
		}

		if dftPersistence && dv.scp.dftPersistentLayers[chIdx] != nil {
			img, ok := dv.scp.dftScopeSignalScreen.(draw.Image)
			if ok {
				draw.Draw(img, bounds, dv.scp.dftPersistentLayers[chIdx], bounds.Min, draw.Over)
			}
		}
	}
	dv.mCache = m
	dv.fsCache = fs

	if dv.showInspector || dv.refActive {
		dv.drawInspector(float64(w), float64(h), bounds)
	}
}

func (dv *dftViewer) calcValuesAt(mx, my float32, w, h float64, bounds image.Rectangle) (freqAtCursor float64, instV, instVCur []float64) {
	minFreq, maxFreqPlot := dv.scp.getDftFreqRange()
	maxFreqAvailable := dv.fsCache / 2

	fractionAtCursor := (float64(mx) - float64(bounds.Min.X)) / w
	if dv.scp.Settings.Dft.XAxisLog {
		logMin := math.Log10(math.Max(minFreq, 1e-6))
		maxFreq := minFreq + maxFreqPlot
		logMax := math.Log10(math.Max(maxFreq, math.Max(minFreq, 1e-6)*1.001))
		logRange := logMax - logMin
		
		logF := logMin + fractionAtCursor*logRange
		freqAtCursor = math.Pow(10, logF)
	} else {
		freqAtCursor = minFreq + fractionAtCursor*maxFreqPlot
	}

	n := int(dv.scp.channelCount) + len(dv.scp.Settings.VirtualChannels)
	instV = make([]float64, n)
	instVCur = make([]float64, n)

	binIdx := int(math.Round((freqAtCursor / maxFreqAvailable) * float64(dv.mCache/2)))

	for chIdx := 0; chIdx < n; chIdx++ {
		var enabled bool
		var dftDisplayOffsetInt int
		if chIdx < int(dv.scp.channelCount) {
			enabled = dv.scp.Settings.Channels[chIdx].Enabled
			dftDisplayOffsetInt = dv.scp.channelViewers[chIdx].dftDisplayOffsetInt
		} else {
			vchIdx := chIdx - int(dv.scp.channelCount)
			enabled = dv.scp.Settings.VirtualChannels[vchIdx].Enabled
			if vchIdx < len(dv.scp.ftVChannelLabels) {
				dftDisplayOffsetInt = dv.scp.snapYToDftN(dv.scp.ftVChannelLabels[vchIdx].displayOffsetFraction)
			} else {
				dftDisplayOffsetInt = 0
			}
		}
		if enabled && len(dv.magnitudesCache) > chIdx && len(dv.magnitudesCache[chIdx]) > 0 {
			magnitudes := dv.magnitudesCache[chIdx]
			var val float64
			if binIdx >= 0 && binIdx < len(magnitudes) {
				val = magnitudes[binIdx]
			}

			yOffset := dv.scp.offsetNToDftY(dftDisplayOffsetInt)
			var v_cursor float64
			dbFloor := -100.0
			if dv.scp.Settings.Dft.DisplayMode == settings.ModeVoltage {
				v_cursor = (float64(bounds.Min.Y) + h + yOffset - float64(my)) / h
			} else {
				v_cursor = (float64(my) - float64(bounds.Min.Y) - yOffset) / h * dbFloor
			}

			instV[chIdx] = val
			instVCur[chIdx] = v_cursor
		}
	}
	return freqAtCursor, instV, instVCur
}

func (dv *dftViewer) drawInspector(w, h float64, bounds image.Rectangle) {
	if dv.fsCache == 0 || dv.mCache == 0 {
		return
	}

	crosscol := color.RGBA{180, 180, 180, 180}
	mx := int(dv.mouseX)
	my := int(dv.mouseY)
	for i := bounds.Min.X; i < bounds.Max.X; i++ {
		dv.scp.dftScopeFullScreen.Set(i, my, crosscol)
	}
	for i := bounds.Min.Y; i < bounds.Max.Y; i++ {
		dv.scp.dftScopeFullScreen.Set(mx, i, crosscol)
	}

	if dv.refActive {
		refcol := color.RGBA{255, 255, 0, 180}
		rx := int(dv.refX)
		ry := int(dv.refY)
		for i := bounds.Min.X; i < bounds.Max.X; i++ {
			dv.scp.dftScopeFullScreen.Set(i, ry, refcol)
		}
		for i := bounds.Min.Y; i < bounds.Max.Y; i++ {
			dv.scp.dftScopeFullScreen.Set(rx, i, refcol)
		}
	}

	if !dv.showInspector {
		return
	}

	freqAtCursor, instVLocal, instVCurLocal := dv.calcValuesAt(dv.mouseX, dv.mouseY, w, h, bounds)

	var refFreq float64
	var refInstV, refInstVCur []float64
	if dv.refActive {
		refFreq, refInstV, refInstVCur = dv.calcValuesAt(dv.refX, dv.refY, w, h, bounds)
	}

	var info []struct {
		text string
		col  color.Color
	}
	info = append(info, struct {
		text string
		col  color.Color
	}{"F: " + formatFreq(freqAtCursor) + "Hz", color.White})

	if dv.refActive {
		df := freqAtCursor - refFreq
		info = append(info, struct {
			text string
			col  color.Color
		}{"ΔF: " + formatFreq(df) + "Hz", color.White})
	}

	moved := false
	if dv.mouseX != dv.inspectorLastX || dv.mouseY != dv.inspectorLastY {
		moved = true
		dv.inspectorLastX = dv.mouseX
		dv.inspectorLastY = dv.mouseY
	}

	totalChannels := int(dv.scp.channelCount) + len(dv.scp.Settings.VirtualChannels)
	if dv.inspectorSumV == nil || len(dv.inspectorSumV) != totalChannels {
		dv.inspectorSumV = make([]float64, totalChannels)
		dv.inspectorSumVCur = make([]float64, totalChannels)
		dv.inspectorDispV = make([]float64, totalChannels)
		dv.inspectorDispVCur = make([]float64, totalChannels)
	}

	if moved {
		for i := range dv.inspectorSumV {
			dv.inspectorSumV[i] = 0
			dv.inspectorSumVCur[i] = 0
		}
		dv.inspectorSamples = 0
	}

	for i := 0; i < totalChannels; i++ {
		dv.inspectorSumV[i] += instVLocal[i]
		dv.inspectorSumVCur[i] += instVCurLocal[i]
	}
	dv.inspectorSamples++

	now := time.Now()
	updateDisplay := false
	if moved || now.Sub(dv.inspectorLastUpdate) >= 500*time.Millisecond {
		updateDisplay = true
		dv.inspectorLastUpdate = now
	}

	if updateDisplay {
		for i := 0; i < totalChannels; i++ {
			if dv.inspectorSamples > 0 {
				dv.inspectorDispV[i] = dv.inspectorSumV[i] / float64(dv.inspectorSamples)
				dv.inspectorDispVCur[i] = dv.inspectorSumVCur[i] / float64(dv.inspectorSamples)
			}
			dv.inspectorSumV[i] = 0
			dv.inspectorSumVCur[i] = 0
		}
		dv.inspectorSamples = 0
	}

	for chIdx := 0; chIdx < totalChannels; chIdx++ {
		var enabled bool
		var col color.NRGBA
		var vRange genericps.RangeEnum
		var chName string
		
		if chIdx < int(dv.scp.channelCount) {
			enabled = dv.scp.Settings.Channels[chIdx].Enabled
			col = dv.scp.Settings.Channels[chIdx].Col[dv.scp.Settings.ChannelColorIndex]
			vRange = dv.scp.Settings.Channels[chIdx].VRange
			chName = fmt.Sprintf("Ch%c", 'A'+chIdx)
		} else {
			vchIdx := chIdx - int(dv.scp.channelCount)
			enabled = dv.scp.Settings.VirtualChannels[vchIdx].Enabled
			col = dv.scp.Settings.VirtualChannels[vchIdx].Col[dv.scp.Settings.ChannelColorIndex]
			vRange = dv.scp.Settings.VirtualChannels[vchIdx].VRange
			chName = dv.scp.Settings.VirtualChannels[vchIdx].Name
		}
		if enabled && len(dv.magnitudesCache) > chIdx && len(dv.magnitudesCache[chIdx]) > 0 {
			v := dv.inspectorDispV[chIdx]
			v_cursor := dv.inspectorDispVCur[chIdx]

			var valStr, curStr string
			if dv.scp.Settings.Dft.DisplayMode == settings.ModeVoltage {
				mv := v * float64(genericps.RangeValuesMv[vRange])
				mvCur := v_cursor * float64(genericps.RangeValuesMv[vRange])
				valStr = formatVoltageFloat64(mv, vRange)
				curStr = formatVoltageFloat64(mvCur, vRange)
			} else {
				unitStr := dv.scp.Settings.Dft.DisplayMode
				if unitStr == settings.ModeArbitraryDB {
					unitStr = "dB"
				}
				valStr = fmt.Sprintf("%.1f%s", v, unitStr)
				curStr = fmt.Sprintf("%.1f%s", v_cursor, unitStr)
			}

			text := fmt.Sprintf("%s: %s (Cur: %s)", chName, valStr, curStr)
			if dv.refActive {
				dvV := v - refInstV[chIdx]
				dvCurV := v_cursor - refInstVCur[chIdx]

				var dvValStr, dvCurStr string
				if dv.scp.Settings.Dft.DisplayMode == settings.ModeVoltage {
					mv := dvV * float64(genericps.RangeValuesMv[vRange])
					mvCur := dvCurV * float64(genericps.RangeValuesMv[vRange])
					dvValStr = formatVoltageFloat64(mv, vRange)
					dvCurStr = formatVoltageFloat64(mvCur, vRange)
				} else {
					unitStr := dv.scp.Settings.Dft.DisplayMode
					if unitStr == settings.ModeArbitraryDB {
						unitStr = "dB"
					}
					dvValStr = fmt.Sprintf("%.1f%s", dvV, unitStr)
					dvCurStr = fmt.Sprintf("%.1f%s", dvCurV, unitStr)
				}
				text += fmt.Sprintf(" ΔV: %s (ΔCur: %s)", dvValStr, dvCurStr)
			}

			info = append(info, struct {
				text string
				col  color.Color
			}{text, col})
		}
	}

	// Draw the box
	lineHeight := 20
	maxW := float32(0)
	for _, item := range info {
		left, _, right, _ := dv.scp.boundString(item.text)
		if right-left > maxW {
			maxW = right - left
		}
	}
	boxWidth := int(maxW) + 15
	boxHeight := len(info)*lineHeight + 10

	xBox := int(dv.mouseX) + 20
	yBox := int(dv.mouseY) + 20

	if xBox+boxWidth > bounds.Max.X-2 {
		xBox = int(dv.mouseX) - boxWidth - 20
	}
	if xBox < bounds.Min.X+2 {
		xBox = bounds.Min.X + 2
	}
	if yBox+boxHeight > bounds.Max.Y-2 {
		yBox = int(dv.mouseY) - boxHeight - 20
	}
	if yBox < bounds.Min.Y+2 {
		yBox = bounds.Min.Y + 2
	}

	rect := image.Rect(xBox, yBox, xBox+boxWidth, yBox+boxHeight)
	draw.Draw(dv.scp.dftScopeFullScreen, rect, &image.Uniform{color.RGBA{20, 20, 20, 220}}, image.ZP, draw.Over)
	for i := 0; i < boxWidth; i++ {
		dv.scp.dftScopeFullScreen.Set(xBox+i, yBox, color.White)
		dv.scp.dftScopeFullScreen.Set(xBox+i, yBox+boxHeight-1, color.White)
	}
	for i := 0; i < boxHeight; i++ {
		dv.scp.dftScopeFullScreen.Set(xBox, yBox+i, color.White)
		dv.scp.dftScopeFullScreen.Set(xBox+boxWidth-1, yBox+i, color.White)
	}

	for i, item := range info {
		dv.scp.addLabel(dv.scp.dftScopeFullScreen, xBox+8, yBox+10+i*lineHeight+15, item.text, item.col)
	}
}

func formatVoltageFloat64(mv float64, vRange genericps.RangeEnum) string {
	if genericps.RangeValuesMv[vRange] >= 1000 {
		return fmt.Sprintf("%.1fV", mv/1000.0)
	}
	return fmt.Sprintf("%.0fmV", mv)
}

func formatFreq(f float64) string {
	if f == 0 {
		return "0"
	}
	if f >= 1e6 {
		return fmt.Sprintf("%.3gM", f/1e6)
	}
	if f >= 1e3 {
		return fmt.Sprintf("%.3gk", f/1e3)
	}
	return fmt.Sprintf("%.3g", f)
}

func formatTime(t float64) string {
	if t >= 1.0 {
		return fmt.Sprintf("%.3gs", t)
	}
	if t >= 1e-3 {
		return fmt.Sprintf("%.3gms", t*1e3)
	}
	if t >= 1e-6 {
		return fmt.Sprintf("%.3gµs", t*1e6)
	}
	if t >= 1e-9 {
		return fmt.Sprintf("%.3gns", t*1e9)
	}
	return fmt.Sprintf("%.3gps", t*1e12)
}

func (scp *ScpDesc) updateBinWidth() {
	if scp.binWidthLabel == nil {
		return
	}
	if scp.psControl.SamplingTimeInterval == 0 {
		fyne.Do(func() { scp.binWidthLabel.SetText("BW: -") })
		return
	}
	fs := 1.0 / float64(scp.psControl.SamplingTimeInterval)
	bw := fs / float64(2*scp.Settings.Dft.Bins)
	text := fmt.Sprintf("BW: %sHz", formatFreq(bw))
	fyne.Do(func() { scp.binWidthLabel.SetText(text) })
}

func (scp *ScpDesc) updateDftDataCollectionTime() {
	if scp.dftDataCollectionTimeLabel == nil {
		return
	}
	// maxScreenTime is N/fs in DFT mode
	text := fmt.Sprintf("Coll: %s", formatTime(scp.maxScreenTime))
	fyne.Do(func() { scp.dftDataCollectionTimeLabel.SetText(text) })
}

func (scp *ScpDesc) dftSampleUnitUp() {
	if scp.dftSampleUnitSelect == nil || scp.dftSampleRateSelect == nil {
		return
	}
	index := scp.dftSampleUnitSelect.SelectedIndex()
	if index < len(scp.dftSampleUnitSelect.Options)-1 {
		scp.dftSampleRateSelect.SilentSetSelectedIndex(0) // Set to "1"
		scp.Settings.Dft.SampleRate = scp.dftSampleRateSelect.Selected
		scp.dftSampleUnitSelect.SetSelectedIndex(index + 1)
		scp.Settings.Dft.SampleRateUnit = scp.dftSampleUnitSelect.Selected
	}
}

func (scp *ScpDesc) dftSampleUnitDown() {
	if scp.dftSampleUnitSelect == nil || scp.dftSampleRateSelect == nil {
		return
	}
	index := scp.dftSampleUnitSelect.SelectedIndex()
	if index > 0 {
		scp.dftSampleRateSelect.SilentSetSelectedIndex(len(scp.dftSampleRateSelect.Options) - 1) // Set to "500"
		scp.Settings.Dft.SampleRate = scp.dftSampleRateSelect.Selected
		scp.dftSampleUnitSelect.SetSelectedIndex(index - 1)
		scp.Settings.Dft.SampleRateUnit = scp.dftSampleUnitSelect.Selected
	}
}


func applyWindow(samples []float64, windowType string) {
	n := len(samples)
	if n <= 1 {
		return
	}
	switch windowType {
	case settings.WindowBartlettHann:
		window.BartlettHann(samples)
	case settings.WindowBlackman:
		window.Blackman(samples)
	case settings.WindowBlackmanHarris:
		window.BlackmanHarris(samples)
	case settings.WindowBlackmanNuttall:
		window.BlackmanNuttall(samples)
	case settings.WindowFlatTop:
		window.FlatTop(samples)
	case settings.WindowHamming:
		window.Hamming(samples)
	case settings.WindowHann:
		window.Hann(samples)
	case settings.WindowNuttall:
		window.Nuttall(samples)
	case settings.WindowLanczos:
		window.Lanczos(samples)
	case settings.WindowTriangular:
		window.Triangular(samples)
	case settings.WindowRectangular:
		// Do nothing
	}
}

func (scp *ScpDesc) getCoherentGain(windowType string, n int) float64 {
	if windowType == settings.WindowRectangular || n <= 0 {
		return 1.0
	}
	temp := make([]float64, n)
	for i := range temp {
		temp[i] = 1.0
	}
	applyWindow(temp, windowType)
	sum := 0.0
	for _, v := range temp {
		sum += v
	}
	return sum / float64(n)
}

func (scp *ScpDesc) offsetNToDftY(n int) float64 {
	if scp.dftScopeSignalScreen == nil {
		return 0
	}
	h := float64(scp.dftScopeSignalScreen.Bounds().Dy())
	yRasterDiv := (h / float64(numberOfDivs)) / 5.0
	return float64(n) * yRasterDiv
}

func (scp *ScpDesc) setDftVDivsY() {
	if scp.dftScopeSignalScreen == nil {
		return
	}
	bounds := scp.dftScopeSignalScreen.Bounds()
	h := float32(bounds.Dy())
	dh := (h - 1) / numberOfDivs
	for i, y := 0, float32(bounds.Min.Y); y <= float32(bounds.Max.Y); i, y = i+1, y+dh {
		scp.dftDivsY[i] = y
	}
}

func (scp *ScpDesc) setDftHDivsX() {
	if scp.dftScopeSignalScreen == nil {
		return
	}
	bounds := scp.dftScopeSignalScreen.Bounds()
	w := float64(bounds.Dx()) - 1
	if w < 1 {
		return
	}
	minFreq, span := scp.getDftFreqRange()

	if !scp.Settings.Dft.XAxisLog {
		// Calculate nice step for approximately 10 divisions
		step := niceStep(span / 10.0)
		firstFreq := math.Floor(minFreq/step) * step

		if len(scp.dftDivsX) < 11 {
			scp.dftDivsX = make([]float32, 11)
		}
		scp.dftDivsX = scp.dftDivsX[:11]
		for i := range scp.dftDivsX {
			freq := firstFreq + float64(i)*step
			x := float32(bounds.Min.X) + float32((freq-minFreq)/span*w)
			scp.dftDivsX[i] = x
		}
	} else {
		logMin := math.Log10(math.Max(minFreq, 1e-6))
		maxFreq := minFreq + span
		logMax := math.Log10(math.Max(maxFreq, math.Max(minFreq, 1e-6)*1.001))
		logRange := logMax - logMin

		getX := func(f float64) float32 {
			if f <= 0 {
				f = 1e-6
			}
			return float32(float64(bounds.Min.X) + ((math.Log10(f)-logMin)/logRange)*float64(w))
		}

		startDecade := int(math.Floor(logMin))
		endDecade := int(math.Ceil(logMax))

		idx := 0
		for dec := startDecade; dec <= endDecade; dec++ {
			base := math.Pow(10, float64(dec))
			for j := 1; j < 10; j++ {
				freq := base * float64(j)
				if freq < minFreq-1e-6 || freq > minFreq+span+1e-6 {
					continue
				}

				x := getX(freq)
				if idx >= cap(scp.dftDivsX) {
					newSlice := make([]float32, cap(scp.dftDivsX)*2+10)
					copy(newSlice, scp.dftDivsX)
					scp.dftDivsX = newSlice
				}
				if idx >= len(scp.dftDivsX) {
					scp.dftDivsX = scp.dftDivsX[:idx+1]
				}
				scp.dftDivsX[idx] = x
				idx++
			}
		}
		scp.dftDivsX = scp.dftDivsX[:idx]
	}
}

func (scp *ScpDesc) drawDftDivisions() {
	if scp.dftScopeSignalScreen == nil {
		return
	}
	bounds := scp.dftScopeSignalScreen.Bounds()
	drawDivs := func(yOffset float32, col color.Color) {
		for _, v := range scp.dftDivsY {
			counter := 0
			for x := float64(bounds.Min.X); x < float64(bounds.Max.X); x = x + 1.0 {
				if counter%10 < 4 {
					scp.dftScopeSignalScreen.Set(int(math.Round(x)), int(math.Round(float64(v+yOffset))), col)
				}
				counter++
			}
		}
		for _, v := range scp.dftDivsX {
			if v < 0 {
				continue
			}
			counter := 0
			for y := float64(bounds.Min.Y); y < float64(bounds.Max.Y); y = y + 1.0 {
				if counter%10 < 4 {
					scp.dftScopeSignalScreen.Set(int(math.Round(float64(v))), int(math.Round(float64(y))), col)
				}
				counter++
			}
		}
	}

	channelIndex := scp.displayMovedDivs - 1
	col := scp.theme.Color(ColorNameDivision, 0)
	if channelIndex >= 0 {
		if scp.displayMovedDivs > 0 && scp.channelViewers[channelIndex].dftDisplayOffsetInt != 0 {
			drawDivs(0, gray)
			yOffset := scp.offsetNToDftY(scp.channelViewers[channelIndex].dftDisplayOffsetInt)
			drawDivs(float32(yOffset), scp.Settings.Channels[channelIndex].Col[scp.Settings.ChannelColorIndex])
		} else {
			drawDivs(0, col)
		}
	} else {
		drawDivs(0, col)
	}
}

func (scp *ScpDesc) clipDftChRangeScrs(w, h float32) (leftMargin, rightMargin float32) {
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
			channelViewer.dftLabel = newDftChannelLabelViewer(scp.dftScopeFullScreen,
				image.Rect(int(math.Round(float64(start))), 0, int(math.Round(float64(end))), int(math.Round(float64(h-defaultTimeMargin)))),
				channelIndex, image.Rect(int(math.Round(float64(leftMargin))), defaultTopMargin,
					int(math.Round(float64(w-rightMargin))), int(math.Round(float64(h-defaultBottomMargin)))), scp)
			scp.addDftDrawer(&channelViewer.dftLabel)
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

func (scp *ScpDesc) partitionDftScreen(w, h float32) {
	ip := scp.dftScopeFullScreen.(*image.RGBA)
	scp.dftDrawers = nil
	leftMargin, rightMargin := scp.clipDftChRangeScrs(w, h)
	scp.dftScopeSignalScreen = ip.SubImage(image.Rect(int(math.Round(float64(leftMargin))),
		defaultTopMargin, int(math.Round(float64(w-rightMargin))),
		int(math.Round(float64(h-defaultBottomMargin))))).(draw.RGBA64Image)
	scp.dftBottomLabelViewer = newFrqLabelViewer(scp.dftScopeFullScreen,
		image.Rect(int(math.Round(0)), int(math.Round(float64(h-defaultTimeMargin))),
			int(math.Round(float64(w))), int(math.Round(float64(h)))), scp)
	scp.addDftDrawer(scp.dftBottomLabelViewer)
	scp.addDftDrawer(newDftViewer(scp.dftScopeFullScreen, scp.dftScopeSignalScreen.Bounds(), scp))
}

func (scp *ScpDesc) dftRasterGenerator(wInt int, hInt int) image.Image {
	ws := scp.Window.Canvas().Size()
	scp.Settings.Window.Height = ws.Height
	scp.Settings.Window.Width = ws.Width
	defer scp.screenLocker.Unlock()
	scp.screenLocker.Lock()
	w := float32(wInt)
	h := float32(hInt)
	rect := scp.dftScopeFullScreen.Bounds()
	if wInt != rect.Max.X-rect.Min.X || hInt != rect.Max.Y-rect.Min.Y { // window resized
		scp.dftScopeFullScreen = scp.newScopeScreen(image.Point{wInt, hInt})
		rect = scp.dftScopeFullScreen.Bounds()
		w = float32(rect.Dx())
		h = float32(rect.Dy())
		draw.Draw(scp.dftScopeFullScreen, scp.dftScopeFullScreen.Bounds(), &image.Uniform{scp.theme.Color(ColorNameSignalBackground, 0)}, image.ZP, draw.Src)
		scp.partitionDftScreen(w, h)
		scp.setDftVDivsY()
		scp.setDftHDivsX()
	} else if getFlag(scp.repartition) {
		scp.partitionDftScreen(w, h)
		draw.Draw(scp.dftScopeFullScreen, scp.dftScopeFullScreen.Bounds(), &image.Uniform{scp.theme.Color(ColorNameSignalBackground, 0)}, image.ZP, draw.Src)
		scp.setDftVDivsY()
		scp.setDftHDivsX()
	} else if scp.dftScopeSignalScreen == nil {
		scp.setDftVDivsY()
		scp.setDftHDivsX()
	} else if getFlag(scp.themeChanged) {
		draw.Draw(scp.dftScopeFullScreen, scp.dftScopeFullScreen.Bounds(), &image.Uniform{scp.theme.Color(ColorNameSignalBackground, 0)}, image.ZP, draw.Src)
	} else {
		draw.Draw(scp.dftScopeFullScreen, scp.dftScopeSignalScreen.Bounds(), &image.Uniform{scp.theme.Color(ColorNameSignalBackground, 0)}, image.ZP, draw.Src)
	}
	scp.setDftVDivsY()
	scp.setDftHDivsX()
	for i := range scp.dftDrawers {
		// slog.Debug("dftDrawers", "i", i)
		scp.dftDrawers[i].draw()
	}
	return scp.dftScopeFullScreen
}

func (scp *ScpDesc) getDftFreqRange() (minFreq, maxFreqPlot float64) {
	minFreq = scp.Settings.Dft.MinFreq
	maxFreqPlot = scp.Settings.Dft.MaxFreq - minFreq
	if maxFreqPlot <= 0 {
		maxFreqPlot = 1e6 // Default to 1MHz if 0
	}
	if scp.psControl != nil && scp.psControl.SamplingTimeInterval > 0 {
		fs := 1.0 / float64(scp.psControl.SamplingTimeInterval)
		maxFreqAvailable := fs / 2
		if maxFreqPlot > maxFreqAvailable {
			maxFreqPlot = maxFreqAvailable
		}
		if minFreq > maxFreqAvailable-maxFreqPlot {
			minFreq = maxFreqAvailable - maxFreqPlot
		}
	}
	if minFreq < 0 {
		minFreq = 0
	}
	return minFreq, maxFreqPlot
}
