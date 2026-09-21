package gui

import (
	"fmt"
	"fynescope/selectscroll"
	"image"
	"image/color"
	"image/draw"
	"math"
	"math/cmplx"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"gonum.org/v1/gonum/dsp/fourier"
)

// ─────────────────────────────────────────────────────────────────────────────
// Scalar correlation helpers
// ─────────────────────────────────────────────────────────────────────────────

// computeCorrelation returns the Pearson correlation coefficient between two
// equal-length sample slices. It returns NaN if either slice has zero variance
// or if the slices are empty / mismatched in length.
func computeCorrelation(a, b []float32) float64 {
	n := len(a)
	if n == 0 || n != len(b) {
		return math.NaN()
	}

	var sumA, sumB float64
	for i := 0; i < n; i++ {
		sumA += float64(a[i])
		sumB += float64(b[i])
	}
	meanA := sumA / float64(n)
	meanB := sumB / float64(n)

	var cov, varA, varB float64
	for i := 0; i < n; i++ {
		da := float64(a[i]) - meanA
		db := float64(b[i]) - meanB
		cov += da * db
		varA += da * da
		varB += db * db
	}
	if varA == 0 || varB == 0 {
		return math.NaN()
	}
	return cov / math.Sqrt(varA*varB)
}

// computeCrossCorrelation returns the normalized cross-correlation between two
// equal-length sample slices at lag 0 (the dot product normalised by energies).
func computeCrossCorrelation(a, b []float32) float64 {
	n := len(a)
	if n == 0 || n != len(b) {
		return math.NaN()
	}

	var cov, energyA, energyB float64
	for i := 0; i < n; i++ {
		fa := float64(a[i])
		fb := float64(b[i])
		cov += fa * fb
		energyA += fa * fa
		energyB += fb * fb
	}
	if energyA == 0 || energyB == 0 {
		return math.NaN()
	}
	return cov / math.Sqrt(energyA*energyB)
}

// ─────────────────────────────────────────────────────────────────────────────
// Full XCF sequence (all lags)
// ─────────────────────────────────────────────────────────────────────────────

// xcfMaxSamples caps how many samples are used so the FFT stays fast.
const xcfMaxSamples = 4096

// computeXCFSequence computes the normalised cross-correlation function for all
// lags from -(n-1) to +(n-1) using an FFT approach.  The returned slice has
// length 2n-1; index k corresponds to lag k-(n-1).  All values lie in [-1, 1].
// Returns nil if the inputs are invalid or have zero energy.
func computeXCFSequence(a, b []float32) []float64 {
	n := len(a)
	if n == 0 || n != len(b) {
		return nil
	}

	// Limit sample count for performance.
	if n > xcfMaxSamples {
		start := (n - xcfMaxSamples) / 2
		a = a[start : start+xcfMaxSamples]
		b = b[start : start+xcfMaxSamples]
		n = xcfMaxSamples
	}

	// Zero-pad to next power-of-two >= 2n to make the convolution linear.
	size := 1
	for size < 2*n {
		size <<= 1
	}

	fa := make([]float64, size)
	fb := make([]float64, size)
	for i := 0; i < n; i++ {
		fa[i] = float64(a[i])
		fb[i] = float64(b[i])
	}

	// Auto-correlation energies for normalisation.
	var raa0, rbb0 float64
	for i := 0; i < n; i++ {
		raa0 += fa[i] * fa[i]
		rbb0 += fb[i] * fb[i]
	}
	norm := math.Sqrt(raa0 * rbb0)
	if norm == 0 {
		return nil
	}

	// FFT-based cross-correlation: IFFT( FFT(a) · conj(FFT(b)) ).
	fftObj := fourier.NewFFT(size)
	cA := fftObj.Coefficients(nil, fa)
	cB := fftObj.Coefficients(nil, fb)
	for i := range cA {
		cA[i] = cA[i] * cmplx.Conj(cB[i])
	}
	// gonum Sequence returns size * IDFT (unnormalised).
	xcfFull := fftObj.Sequence(nil, cA)

	// scale = size (IDFT correction) × norm (energy normalisation).
	scale := float64(size) * norm

	// Extract lags -(n-1) … +(n-1); output length 2n-1.
	// lag >= 0 → xcfFull[lag]
	// lag <  0 → xcfFull[size+lag]
	out := make([]float64, 2*n-1)
	for lag := -(n - 1); lag <= n-1; lag++ {
		idx := lag
		if idx < 0 {
			idx += size
		}
		out[lag+(n-1)] = xcfFull[idx] / scale
	}
	return out
}

// ─────────────────────────────────────────────────────────────────────────────
// corrXCFRaster — Fyne widget that renders the XCF waveform
// ─────────────────────────────────────────────────────────────────────────────

type corrXCFRaster struct {
	widget.BaseWidget
	mu    sync.Mutex
	xcf   []float64
	col   color.RGBA
	nameA string
	nameB string
	scp   *ScpDesc
}

func newCorrXCFRaster(scp *ScpDesc) *corrXCFRaster {
	cr := &corrXCFRaster{scp: scp}
	cr.ExtendBaseWidget(cr)
	return cr
}

// SetXCF stores new XCF data and triggers a redraw.  Safe to call from any goroutine.
func (cr *corrXCFRaster) SetXCF(xcf []float64, col color.RGBA, nameA, nameB string) {
	cr.mu.Lock()
	cr.xcf = xcf
	cr.col = col
	cr.nameA = nameA
	cr.nameB = nameB
	cr.mu.Unlock()
	fyne.Do(func() {
		canvas.Refresh(cr)
	})
}

// generate is the canvas.Raster generator called on every resize/refresh.
func (cr *corrXCFRaster) generate(w, h int) image.Image {
	cr.mu.Lock()
	xcf := make([]float64, len(cr.xcf))
	copy(xcf, cr.xcf)
	col := cr.col
	nameA := cr.nameA
	nameB := cr.nameB
	cr.mu.Unlock()

	img := image.NewRGBA(image.Rect(0, 0, w, h))

	// Background.
	var bgColor color.Color = color.RGBA{5, 5, 5, 255}
	if cr.scp != nil && cr.scp.theme != nil {
		bgColor = cr.scp.theme.Color(ColorNameSignalBackground, 0)
	}
	draw.Draw(img, img.Bounds(), &image.Uniform{bgColor}, image.Point{}, draw.Src)

	// Margins (pixels).
	const (
		leftM   = 44
		rightM  = 8
		topM    = 18
		bottomM = 34
		lblSize = 11.0
	)
	sigL := leftM
	sigR := w - rightM
	sigT := topM
	sigB := h - bottomM
	sigW := sigR - sigL
	sigH := sigB - sigT
	if sigW <= 2 || sigH <= 2 {
		return img
	}

	var fgColor color.Color = white
	if cr.scp != nil && cr.scp.theme != nil {
		fgColor = cr.scp.theme.Color(theme.ColorNameForeground, 0)
	}

	// ── Grid ──────────────────────────────────────────────────────────────
	gridCol := color.RGBA{50, 50, 50, 255}
	zeroCol := color.RGBA{80, 80, 80, 255}
	for _, v := range []float64{1.0, 0.5, 0.0, -0.5, -1.0} {
		yf := float32(sigT) + float32(sigH)*float32((1.0-v)/2.0)
		c := color.Color(gridCol)
		if v == 0.0 {
			c = zeroCol
		}
		drawHorizontalLine(img, float32(sigL), float32(sigR), yf, c)
	}

	// ── Y-axis labels ─────────────────────────────────────────────────────
	if cr.scp != nil {
		for _, v := range []float64{1.0, 0.5, 0.0, -0.5, -1.0} {
			yPx := int(float64(sigT)+float64(sigH)*(1.0-v)/2.0) + int(lblSize)/2
			cr.scp.addLabel(img, 0, yPx, fmt.Sprintf("%.1f", v), fgColor, lblSize)
		}
	}

	// ── Border ────────────────────────────────────────────────────────────
	borderCol := color.RGBA{80, 80, 80, 255}
	drawHorizontalLine(img, float32(sigL), float32(sigR), float32(sigT), borderCol)
	drawHorizontalLine(img, float32(sigL), float32(sigR), float32(sigB), borderCol)
	drawVerticalLine(img, float32(sigL), float32(sigT), float32(sigB), borderCol)
	drawVerticalLine(img, float32(sigR), float32(sigT), float32(sigB), borderCol)

	if len(xcf) == 0 {
		if cr.scp != nil {
			cr.scp.addLabel(img, sigL+sigW/4, sigT+sigH/2, "No data – enable two channels", fgColor, lblSize)
		}
		return img
	}

	halfLen := (len(xcf) - 1) / 2

	// ── Centre vertical (lag = 0) ─────────────────────────────────────────
	centerX := float32(sigL) + float32(sigW)/2.0
	drawVerticalLine(img, centerX, float32(sigT), float32(sigB), color.RGBA{70, 70, 70, 255})

	// ── XCF waveform ──────────────────────────────────────────────────────
	n := len(xcf)
	clamp := func(v float64) float64 {
		if v > 1 {
			return 1
		}
		if v < -1 {
			return -1
		}
		return v
	}
	toY := func(v float64) float32 {
		return float32(sigT) + float32(sigH)*float32((1.0-clamp(v))/2.0)
	}
	prevX := float32(sigL)
	prevY := toY(xcf[0])
	for i := 1; i < n; i++ {
		x := float32(sigL) + float32(sigW)*float32(i)/float32(n-1)
		y := toY(xcf[i])
		drawLine(img, prevX, prevY, x, y, col)
		prevX, prevY = x, y
	}

	// ── Peak annotation ───────────────────────────────────────────────────
	peakIdx := 0
	peakVal := xcf[0]
	for i, v := range xcf {
		if math.Abs(v) > math.Abs(peakVal) {
			peakVal = v
			peakIdx = i
		}
	}
	peakLag := peakIdx - halfLen
	if cr.scp != nil {
		cr.scp.addLabel(img, sigL+4, sigT+int(lblSize)+4,
			fmt.Sprintf("peak  lag=%+9d  r=%+.4f", peakLag, peakVal), fgColor, lblSize)
	}

	// ── X-axis labels ─────────────────────────────────────────────────────
	if cr.scp != nil {
		cr.scp.addLabel(img, sigL, h-6, fmt.Sprintf("-%d", halfLen), fgColor, lblSize-1)
		cr.scp.addLabel(img, int(centerX)-5, h-6, "0", fgColor, lblSize-1)
		posLbl := fmt.Sprintf("+%d", halfLen)
		cr.scp.addLabel(img, sigR-len(posLbl)*7, h-6, posLbl, fgColor, lblSize-1)

		// Pair title.
		title := nameA + " ↔ " + nameB + "  XCF"
		titleX := sigL + sigW/2 - len(title)*3
		cr.scp.addLabel(img, titleX, h-18, title, col, lblSize)
	}

	return img
}

// ── Fyne widget boilerplate ────────────────────────────────────────────────

type corrXCFRasterRenderer struct {
	bg     *canvas.Raster
	widget *corrXCFRaster
}

func (cr *corrXCFRaster) CreateRenderer() fyne.WidgetRenderer {
	bg := canvas.NewRaster(cr.generate)
	return &corrXCFRasterRenderer{bg: bg, widget: cr}
}

func (r *corrXCFRasterRenderer) Layout(size fyne.Size) {
	r.bg.Resize(size)
	r.bg.Move(fyne.NewPos(0, 0))
}
func (r *corrXCFRasterRenderer) MinSize() fyne.Size           { return r.widget.MinSize() }
func (r *corrXCFRasterRenderer) Refresh()                     { canvas.Refresh(r.bg) }
func (r *corrXCFRasterRenderer) Destroy()                     {}
func (r *corrXCFRasterRenderer) Objects() []fyne.CanvasObject { return []fyne.CanvasObject{r.bg} }

func (cr *corrXCFRaster) MinSize() fyne.Size { return fyne.NewSize(300, 200) }

// ─────────────────────────────────────────────────────────────────────────────
// Panel builder
// ─────────────────────────────────────────────────────────────────────────────

// newCorrPanel builds the correlation UI into the given container.
// panel should be a container.NewMax() so the scroll view fills the tab area.
func (scp *ScpDesc) newCorrPanel(panel *fyne.Container, undockable bool) {
	vbox := container.New(layout.NewVBoxLayout())

	// ── Undock button ──────────────────────────────────────────────────────
	if undockable {
		undockBtn := widget.NewButtonWithIcon("Undock", theme.ViewFullScreenIcon(), func() {
			if scp.corrWindow != nil {
				scp.corrWindow.RequestFocus()
				return
			}
			scp.corrWindow = scp.App.NewWindow("Correlation")
			winContent := container.NewMax()
			scp.newCorrPanel(winContent, false)
			scp.controlTab.Remove(scp.corrTab)
			scp.corrWindow.SetContent(winContent)
			scp.corrWindow.SetOnClosed(func() {
				scp.corrWindow = nil
				scp.dockTab(scp.corrTab)
				scp.controlTab.SelectIndex(ftTabIndex)
				// Rebuild the tab panel so mode-switching closures are fresh.
				fyne.Do(func() {
					scp.corrLayout.RemoveAll()
					scp.newCorrPanel(scp.corrLayout, true)
					scp.corrLayout.Refresh()
				})
			})
			scp.corrWindow.Resize(fyne.NewSize(520, 580))
			scp.controlTab.SelectIndex(ftTabIndex)
			scp.corrWindow.Show()
			fyne.Do(winContent.Refresh)
		})
		addToTest(undockBtn, "corrUndockBtn", corrTabIndex)
		vbox.Add(container.NewHBox(undockBtn))
		vbox.Add(widget.NewSeparator())
	}

	// ── Title ──────────────────────────────────────────────────────────────
	title := widget.NewLabelWithStyle("Correlation", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	vbox.Add(title)

	// ── Type selector (shared via ScpDesc so it survives undock/redock) ────
	pearsonBox := container.New(layout.NewVBoxLayout())
	xcfBox := container.New(layout.NewVBoxLayout())
	applyMode := func(mode string) {
		if mode == "Cross-Correlation" {
			pearsonBox.Hide()
			xcfBox.Show()
		} else {
			xcfBox.Hide()
			pearsonBox.Show()
		}
		scp.UpdateCorrelation()
	}
	// ── Type selector (recreated so its callbacks bind to the correct containers) ────
	if scp.corrSelectedType == "" {
		scp.corrSelectedType = "Pearson r"
	}
	scp.corrTypeSelect = selectscroll.NewSelectScroll([]string{"Pearson r", "Cross-Correlation"}, func(s string, _ selectscroll.Exception) {
		scp.corrSelectedType = s
		applyMode(s)
	}, "Cross-Correlation")
	scp.corrTypeSelect.SetSelected(scp.corrSelectedType)
	
	vbox.Add(container.NewHBox(widget.NewLabel("Type:"), scp.corrTypeSelect))
	vbox.Add(widget.NewSeparator())

	// ─────────────────────────────────────────────────────────────────────
	// Pearson mode content
	// ─────────────────────────────────────────────────────────────────────

	hint := widget.NewLabel("Correlation between enabled channel pairs, updated live.\nRange: −1 (inverse) … 0 (none) … +1 (perfect)")
	hint.Wrapping = fyne.TextWrapWord
	pearsonBox.Add(hint)
	pearsonBox.Add(widget.NewSeparator())

	headerHBox := container.NewHBox(
		corrCell("Pair", 70, true),
		corrCell("Result", 100, true),
		corrCell("Strength", 140, true),
	)
	pearsonBox.Add(headerHBox)
	pearsonBox.Add(widget.NewSeparator())

	nch := int(scp.channelCount)
	for i := 0; i < nch; i++ {
		for j := i + 1; j < nch; j++ {
			ci, cj := i, j
			nameI := string(rune('A' + ci))
			nameJ := string(rune('A' + cj))

			col := scp.Settings.Channels[ci].Col[scp.Settings.ChannelColorIndex]
			pairLabel := canvas.NewText(nameI+"↔"+nameJ, col)
			pairLabel.TextStyle.Bold = true
			pairLabel.TextSize = 13

			rLabel := widget.NewLabel("---")
			rLabel.Alignment = fyne.TextAlignLeading
			strengthLabel := widget.NewLabel("")

			scp.corrLabels[ci][cj] = rLabel
			scp.corrStrengthLabels[ci][cj] = strengthLabel

			row := container.NewHBox(
				container.New(layout.NewGridWrapLayout(fyne.NewSize(70, 24)), pairLabel),
				container.New(layout.NewGridWrapLayout(fyne.NewSize(100, 24)), rLabel),
				container.New(layout.NewGridWrapLayout(fyne.NewSize(140, 24)), strengthLabel),
			)
			pearsonBox.Add(row)
		}
	}

	// ─────────────────────────────────────────────────────────────────────
	// XCF mode content
	// ─────────────────────────────────────────────────────────────────────

	// Build pair options for all channel pairs.
	var pairOptions []string
	for i := 0; i < nch; i++ {
		for j := i + 1; j < nch; j++ {
			pairOptions = append(pairOptions, string(rune('A'+i))+"↔"+string(rune('A'+j)))
		}
	}

	// Pair selector — created fresh each time so closures reference correct containers.
	var xcfPairSelect *selectscroll.SelectScroll
	if len(pairOptions) > 0 {
		xcfPairSelect = selectscroll.NewSelectScroll(pairOptions, func(s string, _ selectscroll.Exception) {
			scp.corrSelectedPair = s
			scp.UpdateCorrelation()
		}, "AAA")
		// Restore previously selected pair if still valid.
		sel := scp.corrSelectedPair
		valid := false
		for _, opt := range pairOptions {
			if opt == sel {
				valid = true
				break
			}
		}
		if valid {
			xcfPairSelect.SetSelected(sel)
		} else {
			xcfPairSelect.SetSelected(pairOptions[0])
			scp.corrSelectedPair = pairOptions[0]
		}
		xcfBox.Add(container.NewHBox(widget.NewLabel("Pair:"), xcfPairSelect))
	}

	// XCF raster — one instance, shared between tab and undocked window
	// (they are never displayed simultaneously).
	if scp.corrXCFRaster == nil {
		scp.corrXCFRaster = newCorrXCFRaster(scp)
	}
	xcfBox.Add(scp.corrXCFRaster)

	// ─────────────────────────────────────────────────────────────────────
	// Add both mode boxes to the vbox; toggle visibility on mode change.
	// In Fyne VBox, hidden objects take no space.
	// ─────────────────────────────────────────────────────────────────────
	vbox.Add(pearsonBox)
	vbox.Add(xcfBox)

	// Apply current mode immediately.
	applyMode(scp.corrSelectedType)

	panel.Add(container.NewVScroll(vbox))
}

// ─────────────────────────────────────────────────────────────────────────────
// Helpers
// ─────────────────────────────────────────────────────────────────────────────

// corrCell is a helper that returns a label in a fixed-width cell for column alignment.
func corrCell(text string, width float32, bold bool) *fyne.Container {
	lbl := widget.NewLabel(text)
	lbl.TextStyle.Bold = bold
	return container.New(layout.NewGridWrapLayout(fyne.NewSize(width, 24)), lbl)
}

// corrStrength returns a human-readable label for a correlation coefficient.
func corrStrength(r float64) string {
	abs := math.Abs(r)
	switch {
	case abs >= 0.9:
		return "Very strong"
	case abs >= 0.7:
		return "Strong"
	case abs >= 0.5:
		return "Moderate"
	case abs >= 0.3:
		return "Weak"
	default:
		return "Negligible"
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// UpdateCorrelation — called every acquisition cycle
// ─────────────────────────────────────────────────────────────────────────────

// UpdateCorrelation computes the selected correlation metric for every enabled
// channel pair and pushes results to the UI.  Safe to call from any goroutine.
func (scp *ScpDesc) UpdateCorrelation() {
	nch := int(scp.channelCount)
	if nch < 2 {
		return
	}

	isTabSelected := scp.controlTab != nil && scp.controlTab.Selected() == scp.corrTab
	isWindowOpen := scp.corrWindow != nil
	if !isTabSelected && !isWindowOpen {
		return
	}

	isCross := scp.corrTypeSelect != nil && scp.corrTypeSelect.Selected == "Cross-Correlation"

	// ── Cross-Correlation: full XCF drawn on the raster ───────────────────
	if isCross {
		if scp.corrXCFRaster == nil {
			return
		}
		// Parse selected pair "X↔Y".
		sel := scp.corrSelectedPair
		chI, chJ := -1, -1
		for i := 0; i < nch; i++ {
			for j := i + 1; j < nch; j++ {
				opt := string(rune('A'+i)) + "↔" + string(rune('A'+j))
				if opt == sel {
					chI, chJ = i, j
				}
			}
		}
		if chI < 0 || chI >= len(scp.displayBuffers) || chJ >= len(scp.displayBuffers) {
			return
		}
		if !scp.Settings.Channels[chI].Enabled || !scp.Settings.Channels[chJ].Enabled {
			scp.corrXCFRaster.SetXCF(nil, color.RGBA{}, "", "")
			return
		}

		xcf := computeXCFSequence(scp.displayBuffers[chI], scp.displayBuffers[chJ])
		col := toRGBA(scp.Settings.Channels[chI].Col[scp.Settings.ChannelColorIndex])
		nameI := string(rune('A' + chI))
		nameJ := string(rune('A' + chJ))
		scp.corrXCFRaster.SetXCF(xcf, col, nameI, nameJ)
		return
	}

	// ── Pearson r: scalar table update ────────────────────────────────────
	if scp.corrLabels[0][1] == nil {
		return
	}

	type result struct {
		i, j int
		r    float64
	}
	results := make([]result, 0, nch*(nch-1)/2)

	for i := 0; i < nch; i++ {
		for j := i + 1; j < nch; j++ {
			if scp.corrLabels[i][j] == nil {
				continue
			}
			chI := &scp.Settings.Channels[i]
			chJ := &scp.Settings.Channels[j]
			if !chI.Enabled || !chJ.Enabled {
				results = append(results, result{i, j, math.NaN()})
				continue
			}
			if i >= len(scp.displayBuffers) || j >= len(scp.displayBuffers) {
				results = append(results, result{i, j, math.NaN()})
				continue
			}
			r := computeCorrelation(scp.displayBuffers[i], scp.displayBuffers[j])
			results = append(results, result{i, j, r})
		}
	}

	fyne.Do(func() {
		for _, res := range results {
			lbl := scp.corrLabels[res.i][res.j]
			str := scp.corrStrengthLabels[res.i][res.j]
			if lbl == nil {
				continue
			}
			if math.IsNaN(res.r) {
				lbl.SetText("N/A")
				if str != nil {
					str.SetText("—")
				}
			} else {
				lbl.SetText(fmt.Sprintf("%.4f", res.r))
				if str != nil {
					str.SetText(corrStrength(res.r))
				}
			}
		}
	})
}

// toRGBA converts any color.Color to color.RGBA, discarding precision.
func toRGBA(c color.Color) color.RGBA {
	r, g, b, a := c.RGBA()
	return color.RGBA{uint8(r >> 8), uint8(g >> 8), uint8(b >> 8), uint8(a >> 8)}
}
