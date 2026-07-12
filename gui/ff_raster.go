package gui

import (
	"fmt"
	"fynescope/genericps"
	"image"
	"image/color"
	"image/draw"
	"log/slog"
	"math"
	"math/cmplx"
	"sort"
	"strconv"
	"time"

	"gonum.org/v1/gonum/dsp/fourier"

	"fynescope/control/scpi"
	"fynescope/disp7"
	"fynescope/selectscroll"
	"fynescope/settings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// bodePoint represents a single measured data point in a frequency response (Bode) sweep.
// It stores both the amplitude ratio and the relative phase offset at a specific frequency,
// along with vectors for phase averaging to prevent discontinuities at +/-180 degrees.
type bodePoint struct {
	targetFreq float64 // The commanded/target sweep frequency set on the signal generator.
	freq       float64 // The actual frequency measured from reference channel zero-crossings.
	amp        float64 // The running-average amplitude (linear peak-to-peak voltage or dB ratio).
	ampSum     float64 // Accumulated amplitude sum used for running-average computation.
	phase      float64 // The relative phase in degrees, wrapped within the range (-180, +180].
	phaseCos   float64 // Accumulated cosine components used for vector-based phase averaging.
	phaseSin   float64 // Accumulated sine components used for vector-based phase averaging.
	count      int     // Number of measurements blended into this point.
	needsDraw  bool    // Flag indicating that this point was updated and needs rendering on the next display refresh.
}

type (
	// ffViewer implements the custom scope screen drawer and interaction controller for the F(f) Bode plot panel.
	// It handles drawing grid divisions, frequency response curves, processing user drags for scrolling axes,
	// and displaying cursor measurement overlays.
	ffViewer struct {
		rasterPartition                                 // Embedded base raster properties including refresh flag and image pointers.
		scp                     *ScpDesc                // Reference to the global program state descriptor.
		labelBounds             map[int]image.Rectangle // Bounding boxes of interactive screen labels for dragging/interaction (-1 for X-axis, channel index for Y-axis).
		selectedChannel         int                     // The currently active channel index under interaction (-1 for X-axis/frequency scroll, -2 for none).
		frequencyOffsetFraction float64                 // Accumulates fine mouse dragging increments along the logarithmic frequency axis.
		lastMaxDb               float64                 // Cached maximum decibel value from the last full layout render.
		lastMinDb               float64                 // Cached minimum decibel value from the last full layout render.

		fft       *fourier.FFT // Pre-configured FFT structure for fast signal processing.
		samples   []float64    // Temporary buffer for FFT input samples.
		fftResult []complex128 // Temporary buffer for FFT output coefficients.
		m         int          // Internal variable for gonum DSP operations.

		showInspector         bool      // Toggled on when the user right-clicks and drags inside the signal screen.
		mouseX, mouseY        float32   // Current mouse coordinates on the scope raster image.
		inspectorLastX        float32   // Previous X coordinate to detect mouse movement.
		inspectorLastY        float32   // Previous Y coordinate to detect mouse movement.
		inspectorSumAmp       []float64 // Sum of amplitudes at current cursor frequency across channels.
		inspectorSumPhase     []float64 // Sum of phases at current cursor frequency across channels.
		inspectorSumAmpCur    []float64 // Sum of amplitudes at current cursor Y-position across channels.
		inspectorSumPhaseCur  []float64 // Sum of phases at current cursor Y-position across channels.
		inspectorDispAmp      []float64 // Averaged amplitude displayed in the inspector overlay.
		inspectorDispPhase    []float64 // Averaged phase displayed in the inspector overlay.
		inspectorDispAmpCur   []float64 // Averaged Y-position amplitude displayed in the inspector overlay.
		inspectorDispPhaseCur []float64 // Averaged Y-position phase displayed in the inspector overlay.
		inspectorSamples      int       // Count of data frames collected in the current averaging period.
		inspectorLastUpdate   time.Time // Timestamp of the last visual update of the inspector values.
		// cached scratch slices reused per drawInspector/drawChannels call to avoid per-frame allocs
		instAmp         []float64
		instPhase       []float64
		instAmpCur      []float64
		instPhaseCur    []float64
		smoothedPhaseBuf []float64
	}
)

var (
	_ cursorable = (*ffViewer)(nil)
)

// newFfViewer constructs a new ffViewer instance, initializing default interactive selections.
func newFfViewer(img rasterImage, imgRect image.Rectangle, scp *ScpDesc) *ffViewer {
	return &ffViewer{
		rasterPartition: rasterPartition{img: img, imgRect: imgRect, refreshFlag: true},
		scp:             scp,
		labelBounds:     make(map[int]image.Rectangle),
		selectedChannel: -2, // -2: none, -1: x-axis
	}
}

// draw renders the complete frequency response visual interface on the scope screen.
// It handles drawing the grid system (logarithmic vertical grid lines for frequency,
// horizontal grid lines for division reference), adding axis scale labels (e.g. Hz, kHz, V, dB, degrees),
// and calling the routines to draw individual channel response curves and the inspector popup.
func (ff *ffViewer) draw() {
	if !ff.scp.shouldDrawRaster(ffTabIndex) {
		return
	}
	fullRefresh := ff.refreshFlag
	ff.refreshFlag = false

	if fullRefresh {
		ff.labelBounds = make(map[int]image.Rectangle)
	}
	bounds := ff.scp.ffScopeSignalScreen.Bounds()
	w := float64(bounds.Dx())
	h := float64(bounds.Dy())

	// Background and Grid
	gridCol := ff.scp.Settings.Channels[ff.scp.Settings.ChannelColorIndex].Col[ff.scp.Settings.ChannelColorIndex]
	gridCol = color.NRGBA{50, 50, 50, 255}

	// Frequency Axis (Horizontal) using Ff parameters
	minFreq := ff.scp.Settings.Ff.MinFreq
	maxFreq := ff.scp.Settings.Ff.MaxFreq
	freqRange := maxFreq - minFreq
	if freqRange <= 0 {
		freqRange = 1000 // safe fallback
	}

	// Add horizontal drag/scroll hit box for X axis
	ff.labelBounds[-1] = image.Rect(bounds.Min.X, bounds.Max.Y, bounds.Max.X, bounds.Max.Y+40)

	drawVDiv := func(xf float64, col color.Color) {
		ixf := int(math.Round(xf))
		if ixf >= bounds.Min.X && ixf <= bounds.Max.X {
			counter := 0
			for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
				if counter%10 < 4 {
					ff.scp.ffScopeFullScreen.Set(ixf, y, col)
				}
				counter++
			}
		}
	}

	drawHDiv := func(yf float64, col color.Color) {
		iyf := int(math.Round(yf))
		if iyf >= bounds.Min.Y && iyf <= bounds.Max.Y {
			counter := 0
			for x := bounds.Min.X; x < bounds.Max.X; x++ {
				if counter%10 < 4 {
					ff.scp.ffScopeFullScreen.Set(x, iyf, col)
				}
				counter++
			}
		}
	}

	channelIndex := ff.scp.displayMovedDivs - 1
	var yOffsetShift float64
	if channelIndex >= 0 {
		channelViewer := &ff.scp.channelViewers[channelIndex]
		yOffsetShift = channelViewer.ffDisplayOffsetFraction
	}

	if fullRefresh {
		// Draw horizontal linear grid (amplitude)
		for i := -40; i <= 48; i++ {
			yf := float64(bounds.Min.Y) + float64(i)*h/8.0

			if channelIndex >= 0 {
				drawHDiv(yf, color.NRGBA{50, 50, 50, 255})
				col := ff.scp.Settings.Channels[channelIndex].Col[ff.scp.Settings.ChannelColorIndex]
				counter := 0
				yfShifted := yf + yOffsetShift
				for x := bounds.Min.X; x < bounds.Max.X; x++ {
					if counter%10 < 4 {
						if yfShifted >= float64(bounds.Min.Y) && yfShifted < float64(bounds.Max.Y) {
							ff.scp.ffScopeFullScreen.Set(x, int(math.Round(yfShifted)), col)
						}
					}
					counter++
				}
			} else {
				drawHDiv(yf, gridCol)
			}
		}
	}

	if fullRefresh {
		// Draw vertical logarithmic grid (frequency)
		logMin := math.Log10(math.Max(minFreq, 1E-6))
		logMax := math.Log10(math.Max(maxFreq, math.Max(minFreq, 1E-6)*1.001))
		logRange := logMax - logMin

		getX := func(f float64) float64 {
			if f <= 0 {
				f = 1E-6
			}
			return float64(bounds.Min.X) + ((math.Log10(f)-logMin)/logRange)*w
		}

		startDecade := int(math.Floor(logMin))
		endDecade := int(math.Ceil(logMax))

		for dec := startDecade; dec <= endDecade; dec++ {
			base := math.Pow(10, float64(dec))
			for j := 1; j < 10; j++ {
				f := base * float64(j)
				xf := getX(f)
				if xf >= float64(bounds.Min.X) && xf <= float64(bounds.Max.X) {
					col := gridCol
					if channelIndex >= 0 {
						col = color.NRGBA{50, 50, 50, 255}
					}

					if j == 1 {
						ixf := int(math.Round(xf))
						for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
							ff.scp.ffScopeFullScreen.Set(ixf, y, col)
						}

						vstr := fmt.Sprintf("%gHz", f)
						if f >= 1000000 {
							vstr = fmt.Sprintf("%.1fMHz", f/1000000.0)
						} else if f >= 1000 {
							vstr = fmt.Sprintf("%.1fkHz", f/1000.0)
						}
						left, _, right, _ := ff.scp.boundString(vstr)
						ff.scp.addLabel(ff.scp.ffScopeFullScreen, int(math.Round(xf-float64(right-left)/2)), bounds.Max.Y+20, vstr, color.NRGBA{200, 200, 200, 255})
					} else {
						drawVDiv(xf, col)
					}
				}
			}
		}
	}

	yLabelOffsetLeft := 0
	yLabelOffsetRight := 0
	yCount := 0

	if fullRefresh {
		for i := 0; i < int(ff.scp.channelCount); i++ {
			ch := ff.scp.Settings.Channels[i]
			if !ch.Enabled {
				continue
			}

			col := ch.Col[ff.scp.Settings.ChannelColorIndex]
			yOffset := ff.offsetNToFf(ch.FfDisplayVOffset)

			var minX, maxX int
			if yCount%2 == 0 {
				minX = bounds.Min.X - 60 - yLabelOffsetLeft
				maxX = bounds.Min.X - yLabelOffsetLeft
			} else {
				minX = bounds.Max.X + yLabelOffsetRight
				maxX = bounds.Max.X + 60 + yLabelOffsetRight
			}
			ff.labelBounds[i] = image.Rect(minX, bounds.Min.Y, maxX, bounds.Max.Y)

			var unitName string
			if ff.scp.Settings.Dft.DisplayMode == settings.ModeVoltage {
				maxV := genericps.RangeValuesMv[ch.VRange]
				if maxV < 1000.0 {
					unitName = "mV"
				} else {
					unitName = "V"
					maxV /= 1000.0
				}

				lx_unit := 0
				left_u, _, right_u, _ := ff.scp.boundString(unitName)
				if yCount%2 == 0 {
					lx_unit = bounds.Min.X - int(right_u-left_u) - 5 - yLabelOffsetLeft
				} else {
					lx_unit = bounds.Max.X + 5 + yLabelOffsetRight
				}

				divH := h / 8.0
				j_min := math.Ceil(-yOffset / divH)
				yf_top := float64(bounds.Min.Y) + j_min*divH + yOffset
				ff.scp.addLabel(ff.scp.ffScopeFullScreen, lx_unit, int(math.Round(yf_top+divH/2+10)), unitName, col)

				dv := maxV / 8.0
				for j := -40; j <= 48; j++ {
					yf := float64(bounds.Min.Y) + float64(j)*divH + yOffset
					v := maxV - float64(j)*dv
					vstr := strconv.FormatFloat(v, 'f', 1, 64)

					left, top, right, bottom := ff.scp.boundString(vstr)
					var lx int
					if yCount%2 == 0 {
						lx = bounds.Min.X - int(right-left) - 5 - yLabelOffsetLeft
					} else {
						lx = bounds.Max.X + 5 + yLabelOffsetRight
					}
					if yf >= float64(bounds.Min.Y) && yf <= float64(bounds.Max.Y) {
						ff.scp.addLabel(ff.scp.ffScopeFullScreen, lx, int(math.Round(yf-float64(top-bottom)/2)), vstr, col)
					}
				}
			} else {
				unitName = "dB"
				lx_unit := 0
				left_u, _, right_u, _ := ff.scp.boundString(unitName)
				if yCount%2 == 0 {
					lx_unit = bounds.Min.X - int(right_u-left_u) - 5 - yLabelOffsetLeft
				} else {
					lx_unit = bounds.Max.X + 5 + yLabelOffsetRight
				}

				divH := h / 8.0
				j_min := math.Ceil(-yOffset / divH)
				yf_top := float64(bounds.Min.Y) + j_min*divH + yOffset
				ff.scp.addLabel(ff.scp.ffScopeFullScreen, lx_unit, int(math.Round(yf_top+divH/2+10)), unitName, col)

				for j := -40; j <= 48; j++ {
					yf := float64(bounds.Min.Y) + float64(j)*divH + yOffset
					v := float64(j) * -10.0
					vstr := strconv.FormatFloat(v, 'f', 0, 64)

					left, top, right, bottom := ff.scp.boundString(vstr)
					var lx int
					if yCount%2 == 0 {
						lx = bounds.Min.X - int(right-left) - 5 - yLabelOffsetLeft
					} else {
						lx = bounds.Max.X + 5 + yLabelOffsetRight
					}
					if yf >= float64(bounds.Min.Y) && yf <= float64(bounds.Max.Y) {
						ff.scp.addLabel(ff.scp.ffScopeFullScreen, lx, int(math.Round(yf-float64(top-bottom)/2)), vstr, col)
					}
				}
			}

			if yCount%2 == 0 {
				yLabelOffsetLeft += 60
			} else {
				yLabelOffsetRight += 60
			}
			yCount++
		}

		hasPhase := false
		for i := 0; i < int(ff.scp.channelCount); i++ {
			if ff.scp.Settings.Channels[i].Enabled && ff.scp.Settings.Channels[i].FfPhaseEnabled {
				hasPhase = true
				break
			}
		}

		if hasPhase {
			col := color.NRGBA{160, 160, 160, 255}
			unitName := "°"
			lx_unit := 0
			left_u, _, right_u, _ := ff.scp.boundString(unitName)
			if yCount%2 == 0 {
				lx_unit = bounds.Min.X - int(right_u-left_u) - 5 - yLabelOffsetLeft
			} else {
				lx_unit = bounds.Max.X + 5 + yLabelOffsetRight
			}
			ff.scp.addLabel(ff.scp.ffScopeFullScreen, lx_unit, int(math.Round(float64(bounds.Min.Y)+float64(h)/float64(numberOfDivs)/2+10)), unitName, col)

			divHPhase := h / 8.0
			for j := 0; j <= 8; j++ {
				yf := float64(bounds.Min.Y) + float64(j)*divHPhase
				v := 180.0 - float64(j)*45.0
				vstr := strconv.FormatFloat(v, 'f', 0, 64)

				left, top, right, bottom := ff.scp.boundString(vstr)
				var lx int
				if yCount%2 == 0 {
					lx = bounds.Min.X - int(right-left) - 5 - yLabelOffsetLeft
				} else {
					lx = bounds.Max.X + 5 + yLabelOffsetRight
				}
				if yf >= float64(bounds.Min.Y) && yf <= float64(bounds.Max.Y) {
					ff.scp.addLabel(ff.scp.ffScopeFullScreen, lx, int(math.Round(yf-float64(top-bottom)/2)), vstr, col)
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

	ff.drawChannels(minFreq, freqRange, w, h)

	if ff.showInspector {
		ff.drawInspector(w, h, bounds)
	}
}

// drawDashedLine draws a dashed segment between two coordinates on the target image.
// This is primarily used for phase plots to visually distinguish them from amplitude plots.
func drawDashedLine(img draw.Image, x0, y0, x1, y1 float32, c color.Color) {
	dx := x1 - x0
	dy := y1 - y0
	dist := float32(math.Hypot(float64(dx), float64(dy)))
	if dist == 0 {
		return
	}
	dashLen := float32(4.0)
	gapLen := float32(4.0)
	step := dashLen + gapLen

	for t := float32(0.0); t < dist; t += step {
		tEnd := t + dashLen
		if tEnd > dist {
			tEnd = dist
		}
		xStart := x0 + dx*(t/dist)
		yStart := y0 + dy*(t/dist)
		xEnd := x0 + dx*(tEnd/dist)
		yEnd := y0 + dy*(tEnd/dist)
		_ = drawLine(img, xStart, yStart, xEnd, yEnd, c)
	}
}

// drawChannels plots the amplitude (always) and phase (if enabled) frequency response curves for all enabled scope channels.
// Amplitude curves are drawn as solid lines (scaled as either linear peak voltage or logarithmic decibels).
// Phase curves are drawn as dashed, dimmed lines mapped on a fixed Y-scale from -180 to 180 degrees.
// To avoid messy vertical segments when phase wraps between -180 and +180 degrees, transitions exceeding 270 degrees are not joined.
func (ff *ffViewer) drawChannels(minFreq, freqRange, w, h float64) {
	bounds := ff.scp.ffScopeSignalScreen.Bounds()

	logMin := math.Log10(math.Max(minFreq, 1E-6))
	logMax := math.Log10(math.Max(minFreq+freqRange, math.Max(minFreq, 1E-6)*1.001))
	logRange := logMax - logMin

	getX := func(f float64) float64 {
		if f <= 0 {
			f = 1E-6
		}
		return float64(bounds.Min.X) + ((math.Log10(f)-logMin)/logRange)*w
	}

	for chIdx := 0; chIdx < int(ff.scp.channelCount); chIdx++ {
		ch := ff.scp.Settings.Channels[chIdx]
		if !ch.Enabled {
			continue
		}

		ff.scp.ffLocker.Lock()
		pts := make([]bodePoint, len(ff.scp.bodeBuffers[chIdx]))
		copy(pts, ff.scp.bodeBuffers[chIdx])
		ff.scp.ffLocker.Unlock()

		if len(pts) < 1 {
			continue
		}

		col := ch.Col[ff.scp.Settings.ChannelColorIndex]
		yOffset := ff.offsetNToFf(ch.FfDisplayVOffset)

		if ch.Enabled {
			var prevX, prevY float32
			first := true

			for _, pt := range pts {
				if pt.freq < minFreq {
					continue
				}
				if pt.freq > minFreq+freqRange {
					break
				}

				x := float32(getX(pt.freq))
				var y float32

				if ff.scp.Settings.Dft.DisplayMode == settings.ModeVoltage {
					val := pt.amp / genericps.RangeValuesMv[ch.VRange]
					if val > 1.0 {
						val = 1.0
					}
					y = float32(float64(bounds.Min.Y) + (1.0-val)*h + yOffset)
				} else {
					dbFloor := -80.0
					val := pt.amp / genericps.RangeValuesMv[ch.VRange]
					var db float64
					if val < 1e-10 {
						db = dbFloor
					} else {
						db = 20 * math.Log10(val)
					}
					if db < dbFloor {
						db = dbFloor
					}
					if db > 0 {
						db = 0
					}
					y = float32(float64(bounds.Min.Y) + (db/dbFloor)*h + yOffset)
				}

				if !first {
					_ = drawLine(ff.scp.ffScopeSignalScreen, prevX, prevY, x, y, col)
				} else {
					first = false
				}
				prevX = x
				prevY = y
			}
		}

		if ch.FfPhaseEnabled {
			r, g, b, a := col.RGBA()
			dimmedCol := color.NRGBA{
				R: uint8(r >> 9),
				G: uint8(g >> 9),
				B: uint8(b >> 9),
				A: uint8(a >> 8),
			}

			if len(ff.smoothedPhaseBuf) != len(pts) {
				ff.smoothedPhaseBuf = make([]float64, len(pts))
			}
			smoothedPhases := smoothPhaseInto(pts, 7, ff.smoothedPhaseBuf)

			var prevX, prevY float32
			var prevPhase float64
			first := true

			for idx, pt := range pts {
				if pt.freq < minFreq {
					continue
				}
				if pt.freq > minFreq+freqRange {
					break
				}

				x := float32(getX(pt.freq))
				phase := smoothedPhases[idx]
				// Fixed scale: -180 to 180 degrees mapped to height h
				y := float32(float64(bounds.Min.Y) + (180.0-phase)/360.0*h)

				if !first {
					// Avoid drawing ugly vertical lines when the phase wraps around between -180 and 180 degrees.
					// We only draw the line segment if the difference between successive phases is less than 270 degrees.
					if math.Abs(phase-prevPhase) < 270.0 {
						drawDashedLine(ff.scp.ffScopeSignalScreen, prevX, prevY, x, y, dimmedCol)
					}
				} else {
					first = false
				}
				prevX = x
				prevY = y
				prevPhase = phase
			}
		}
	}
}

// drawInspector draws a measurement overlay (crosshairs + dynamic info panel) on the screen.
// It is active while the user holds the right mouse button. It identifies the closest
// frequency point to the mouse cursor, reads/interpolates both amplitude and phase for each channel,
// and averages these values over short epochs to prevent visual jitter.
func (ff *ffViewer) drawInspector(w, h float64, bounds image.Rectangle) {
	minFreq := ff.scp.Settings.Ff.MinFreq
	maxFreq := ff.scp.Settings.Ff.MaxFreq
	freqRange := maxFreq - minFreq
	if freqRange <= 0 {
		freqRange = 1000
	}

	crosscol := color.RGBA{180, 180, 180, 180}
	mx := int(ff.mouseX)
	my := int(ff.mouseY)
	for i := bounds.Min.X; i < bounds.Max.X; i++ {
		ff.scp.ffScopeFullScreen.Set(i, my, crosscol)
	}
	for i := bounds.Min.Y; i < bounds.Max.Y; i++ {
		ff.scp.ffScopeFullScreen.Set(mx, i, crosscol)
	}

	logMin := math.Log10(math.Max(minFreq, 1E-6))
	logMax := math.Log10(math.Max(minFreq+freqRange, math.Max(minFreq, 1E-6)*1.001))
	logRange := logMax - logMin

	fractionAtCursor := (float64(ff.mouseX) - float64(bounds.Min.X)) / w
	logF := logMin + fractionAtCursor*logRange
	freqAtCursor := math.Pow(10, logF)

	var info []struct {
		text string
		col  color.Color
	}
	info = append(info, struct {
		text string
		col  color.Color
	}{"F: " + formatFreq(freqAtCursor) + "Hz", color.White})

	moved := false
	if ff.mouseX != ff.inspectorLastX || ff.mouseY != ff.inspectorLastY {
		moved = true
		ff.inspectorLastX = ff.mouseX
		ff.inspectorLastY = ff.mouseY
	}

	if ff.inspectorSumAmp == nil || len(ff.inspectorSumAmp) != len(ff.scp.channelViewers) {
		ff.inspectorSumAmp = make([]float64, len(ff.scp.channelViewers))
		ff.inspectorSumPhase = make([]float64, len(ff.scp.channelViewers))
		ff.inspectorSumAmpCur = make([]float64, len(ff.scp.channelViewers))
		ff.inspectorSumPhaseCur = make([]float64, len(ff.scp.channelViewers))

		ff.inspectorDispAmp = make([]float64, len(ff.scp.channelViewers))
		ff.inspectorDispPhase = make([]float64, len(ff.scp.channelViewers))
		ff.inspectorDispAmpCur = make([]float64, len(ff.scp.channelViewers))
		ff.inspectorDispPhaseCur = make([]float64, len(ff.scp.channelViewers))
	}

	if moved {
		for i := range ff.inspectorSumAmp {
			ff.inspectorSumAmp[i] = 0
			ff.inspectorSumPhase[i] = 0
			ff.inspectorSumAmpCur[i] = 0
			ff.inspectorSumPhaseCur[i] = 0
		}
		ff.inspectorSamples = 0
	}

	n := len(ff.scp.channelViewers)
	if len(ff.instAmp) != n {
		ff.instAmp = make([]float64, n)
		ff.instPhase = make([]float64, n)
		ff.instAmpCur = make([]float64, n)
		ff.instPhaseCur = make([]float64, n)
	}
	instAmp := ff.instAmp
	instPhase := ff.instPhase
	instAmpCur := ff.instAmpCur
	instPhaseCur := ff.instPhaseCur
	// zero out reused slices
	for i := range instAmp {
		instAmp[i] = 0
		instPhase[i] = 0
		instAmpCur[i] = 0
		instPhaseCur[i] = 0
	}

	for chIdx := 0; chIdx < int(ff.scp.channelCount); chIdx++ {
		ch := ff.scp.Settings.Channels[chIdx]
		if !ch.Enabled {
			continue
		}

		ff.scp.ffLocker.Lock()
		pts := make([]bodePoint, len(ff.scp.bodeBuffers[chIdx]))
		copy(pts, ff.scp.bodeBuffers[chIdx])
		ff.scp.ffLocker.Unlock()

		if len(pts) == 0 {
			continue
		}

		if len(ff.smoothedPhaseBuf) != len(pts) {
			ff.smoothedPhaseBuf = make([]float64, len(pts))
		}
		smoothedPhases := smoothPhaseInto(pts, 7, ff.smoothedPhaseBuf)

		var bestPt *bodePoint
		bestIdx := -1
		minDiff := math.MaxFloat64
		for i := range pts {
			diff := math.Abs(pts[i].freq - freqAtCursor)
			if diff < minDiff {
				minDiff = diff
				bestPt = &pts[i]
				bestIdx = i
			}
		}

		if bestPt != nil {
			instAmp[chIdx] = bestPt.amp
			instPhase[chIdx] = smoothedPhases[bestIdx]
		}

		yOffset := ff.offsetNToFf(ch.FfDisplayVOffset)
		instPhaseCur[chIdx] = 180.0 - (float64(ff.mouseY)-float64(bounds.Min.Y))/h*360.0

		if ff.scp.Settings.Dft.DisplayMode == settings.ModeVoltage {
			val_cursor := (float64(bounds.Min.Y) + h + yOffset - float64(ff.mouseY)) / h
			instAmpCur[chIdx] = val_cursor * float64(genericps.RangeValuesMv[ch.VRange])
		} else {
			dbFloor := -80.0
			instAmpCur[chIdx] = (float64(ff.mouseY) - float64(bounds.Min.Y) - yOffset) / h * dbFloor
		}
	}

	for i := range ff.scp.channelViewers {
		ff.inspectorSumAmp[i] += instAmp[i]
		ff.inspectorSumPhase[i] += instPhase[i]
		ff.inspectorSumAmpCur[i] += instAmpCur[i]
		ff.inspectorSumPhaseCur[i] += instPhaseCur[i]
	}
	ff.inspectorSamples++

	now := time.Now()
	updateDisplay := false
	if moved || now.Sub(ff.inspectorLastUpdate) >= 500*time.Millisecond {
		updateDisplay = true
		ff.inspectorLastUpdate = now
	}

	if updateDisplay {
		for i := range ff.scp.channelViewers {
			if ff.inspectorSamples > 0 {
				ff.inspectorDispAmp[i] = ff.inspectorSumAmp[i] / float64(ff.inspectorSamples)
				ff.inspectorDispPhase[i] = ff.inspectorSumPhase[i] / float64(ff.inspectorSamples)
				ff.inspectorDispAmpCur[i] = ff.inspectorSumAmpCur[i] / float64(ff.inspectorSamples)
				ff.inspectorDispPhaseCur[i] = ff.inspectorSumPhaseCur[i] / float64(ff.inspectorSamples)
			}
			ff.inspectorSumAmp[i] = 0
			ff.inspectorSumPhase[i] = 0
			ff.inspectorSumAmpCur[i] = 0
			ff.inspectorSumPhaseCur[i] = 0
		}
		ff.inspectorSamples = 0
	}

	for chIdx := 0; chIdx < int(ff.scp.channelCount); chIdx++ {
		ch := ff.scp.Settings.Channels[chIdx]
		if !ch.Enabled {
			continue
		}

		col := ch.Col[ff.scp.Settings.ChannelColorIndex]
		var ampStr, ampCurStr, phaseStr, phaseCurStr string

		if ch.Enabled {
			if ff.scp.Settings.Dft.DisplayMode == settings.ModeVoltage {
				ampStr = formatVoltageFloat64(ff.inspectorDispAmp[chIdx], ch.VRange)
				ampCurStr = formatVoltageFloat64(ff.inspectorDispAmpCur[chIdx], ch.VRange)
			} else {
				val := ff.inspectorDispAmp[chIdx] / float64(genericps.RangeValuesMv[ch.VRange])
				var db float64
				if val < 1e-10 {
					db = -80.0
				} else {
					db = 20 * math.Log10(val)
				}
				if db < -80.0 {
					db = -80.0
				}
				if db > 0 {
					db = 0
				}
				ampStr = fmt.Sprintf("%.1fdB", db)
				ampCurStr = fmt.Sprintf("%.1fdB", ff.inspectorDispAmpCur[chIdx])
			}
		}

		if ch.FfPhaseEnabled {
			phaseStr = fmt.Sprintf("%.1f°", ff.inspectorDispPhase[chIdx])
			phaseCurStr = fmt.Sprintf("%.1f°", ff.inspectorDispPhaseCur[chIdx])
		}

		var line string
		if ch.Enabled && ch.FfPhaseEnabled {
			line = fmt.Sprintf("Ch%c: %s/%s (Cur: %s/%s)", 'A'+chIdx, ampStr, phaseStr, ampCurStr, phaseCurStr)
		} else if ch.Enabled {
			line = fmt.Sprintf("Ch%c: %s (Cur: %s)", 'A'+chIdx, ampStr, ampCurStr)
		} else {
			line = fmt.Sprintf("Ch%c: %s (Cur: %s)", 'A'+chIdx, phaseStr, phaseCurStr)
		}

		info = append(info, struct {
			text string
			col  color.Color
		}{line, col})
	}

	lineHeight := 20
	maxW := float32(0)
	for _, item := range info {
		left, _, right, _ := ff.scp.boundString(item.text)
		if right-left > maxW {
			maxW = right - left
		}
	}
	boxWidth := int(maxW) + 15
	boxHeight := len(info)*lineHeight + 10

	xBox := int(ff.mouseX) + 10
	yBox := int(ff.mouseY)

	if xBox+boxWidth > bounds.Max.X-2 {
		xBox = int(ff.mouseX) - boxWidth - 10
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
	draw.Draw(ff.scp.ffScopeFullScreen, rect, &image.Uniform{color.RGBA{20, 20, 20, 220}}, image.ZP, draw.Over)
	for i := 0; i < boxWidth; i++ {
		ff.scp.ffScopeFullScreen.Set(xBox+i, yBox, color.White)
		ff.scp.ffScopeFullScreen.Set(xBox+i, yBox+boxHeight-1, color.White)
	}
	for i := 0; i < boxHeight; i++ {
		ff.scp.ffScopeFullScreen.Set(xBox, yBox+i, color.White)
		ff.scp.ffScopeFullScreen.Set(xBox+boxWidth-1, yBox+i, color.White)
	}

	for i, item := range info {
		ff.scp.addLabel(ff.scp.ffScopeFullScreen, xBox+8, yBox+10+i*lineHeight+15, item.text, item.col)
	}
}

// fyneToImg converts Fyne logical UI coordinates to actual raster image pixel coordinates.
func (ff *ffViewer) fyneToImg(x, y float32) (int, int) {
	if ff.scp.ffRaster == nil {
		return int(math.Round(float64(x))), int(math.Round(float64(y)))
	}
	size := ff.scp.ffRaster.Size()
	if size.Width == 0 || size.Height == 0 {
		return int(math.Round(float64(x))), int(math.Round(float64(y)))
	}
	scaleX := float32(ff.img.Bounds().Dx()) / size.Width
	scaleY := float32(ff.img.Bounds().Dy()) / size.Height
	return int(math.Round(float64(x * scaleX))), int(math.Round(float64(y * scaleY)))
}

// mouseInSignalScreen checks whether the given Fyne coordinate is located within the active scope screen area (excluding external label gutters).
func (ff *ffViewer) mouseInSignalScreen(x, y float32) bool {
	ix, iy := ff.fyneToImg(x, y)
	p := image.Point{X: ix, Y: iy}
	return p.In(ff.scp.ffScopeSignalScreen.Bounds())
}

// cursor returns the appropriate desktop cursor type based on the mouse coordinate.
// It displays a Crosshair when over the signal screen, a Pointer when hovering interactive labels,
// and falls back to Default otherwise.
func (ff *ffViewer) cursor(x, y float32) (desktop.Cursor, bool) {
	if ff.mouseInSignalScreen(x, y) {
		return desktop.CrosshairCursor, true
	}
	ix, iy := ff.fyneToImg(x, y)
	p := image.Point{X: ix, Y: iy}
	for _, bounds := range ff.labelBounds {
		if p.In(bounds) {
			return desktop.PointerCursor, true
		}
	}
	return desktop.DefaultCursor, false
}

// mouseIn returns true if the coordinate is within any interactive label hit boxes.
func (ff *ffViewer) mouseIn(x, y float32) bool {
	ix, iy := ff.fyneToImg(x, y)
	p := image.Point{X: ix, Y: iy}
	for _, bounds := range ff.labelBounds {
		if p.In(bounds) {
			return true
		}
	}
	return false
}

// mouseDown handles mouse press events. Right-clicking toggles the measurement inspector overlay,
// while left-clicking is checked against label hitboxes to register scroll/drag focus.
func (ff *ffViewer) mouseDown(button desktop.MouseButton, x, y float32) {
	if button == desktop.RightMouseButton && ff.mouseInSignalScreen(x, y) {
		ff.showInspector = true
		ix, iy := ff.fyneToImg(x, y)
		ff.mouseX = float32(ix)
		ff.mouseY = float32(iy)
		ff.scp.ffFullRefresh = true
		ff.enableRefresh()
		canvas.Refresh(ff.scp.ffRaster)
		return
	}
	ix, iy := ff.fyneToImg(x, y)
	p := image.Point{X: ix, Y: iy}
	ff.selectedChannel = -2
	for chIdx, bounds := range ff.labelBounds {
		if p.In(bounds) {
			if button == desktop.MouseButtonSecondary || button == desktop.RightMouseButton {
				if chIdx >= 0 {
					channelViewer := &ff.scp.channelViewers[chIdx]
					channelViewer.ffDisplayOffsetFraction = 0
					ff.scp.Settings.Channels[chIdx].FfDisplayVOffset = 0
					ff.scp.ffFullRefresh = true
					ff.scp.refreshRasters()
				}
				return
			}
			ff.selectedChannel = chIdx
			if chIdx >= 0 {
				ff.scp.displayMovedDivs = chIdx + 1
			}
			break
		}
	}
}

// mouseUp handles mouse release events, hiding the inspector if active, and resetting interactive states.
func (ff *ffViewer) mouseUp(button desktop.MouseButton, x, y float32) {
	if button == desktop.RightMouseButton {
		ff.showInspector = false
		ff.scp.ffFullRefresh = true
		ff.enableRefresh()
		canvas.Refresh(ff.scp.ffRaster)
		return
	}
	ff.selectedChannel = -2
	ff.scp.displayMovedDivs = 0
}

// mouseMoved handles mouse movement events, updating the inspector position and forcing a refresh if it's currently visible.
func (ff *ffViewer) mouseMoved(x, y float32) {
	if ff.showInspector {
		ix, iy := ff.fyneToImg(x, y)
		ff.mouseX = float32(ix)
		ff.mouseY = float32(iy)
		bounds := ff.scp.ffScopeSignalScreen.Bounds()
		if ff.mouseX < float32(bounds.Min.X) {
			ff.mouseX = float32(bounds.Min.X)
		}
		if ff.mouseX > float32(bounds.Max.X-1) {
			ff.mouseX = float32(bounds.Max.X - 1)
		}
		if ff.mouseY < float32(bounds.Min.Y) {
			ff.mouseY = float32(bounds.Min.Y)
		}
		if ff.mouseY > float32(bounds.Max.Y-1) {
			ff.mouseY = float32(bounds.Max.Y - 1)
		}
		ff.scp.ffFullRefresh = true
		ff.enableRefresh()
		canvas.Refresh(ff.scp.ffRaster)
	}
}

// offsetNToFf converts integer division offsets to Ff display vertical pixel offsets.
func (ff *ffViewer) offsetNToFf(n int) float64 {
	h := float64(ff.scp.ffScopeSignalScreen.Bounds().Dy())
	yRasterDiv := (h / float64(numberOfDivs)) / 5.0
	return float64(n) * yRasterDiv
}

// snapYToFfN rounds a visual Y-pixel offset to the nearest division grid step.
func (ff *ffViewer) snapYToFfN(y float64) int {
	h := float64(ff.scp.ffScopeSignalScreen.Bounds().Dy())
	yRasterDiv := (h / float64(numberOfDivs)) / 5.0
	return int(math.Round(y / yRasterDiv))
}

// setChDispOffset implements vertical and horizontal offset changes.
// When chIndex == -1, it modifies the logarithmic frequency sweep range (Start/Stop frequency),
// recalculates step values, updates local generator configuration, and resets the sweep.
// Otherwise, it updates vertical offset positions for the specified scope channel.
func (ff *ffViewer) setChDispOffset(chIndex int, dy float64, scroll bool) {
	if chIndex == -1 {
		logMin := math.Log10(math.Max(ff.scp.Settings.Ff.MinFreq, 1E-6))
		logMax := math.Log10(math.Max(ff.scp.Settings.Ff.MaxFreq, math.Max(ff.scp.Settings.Ff.MinFreq, 1E-6)*1.001))
		diff := logMax - logMin
		shift := -diff * (dy / float64(ff.img.Bounds().Dx()))

		ff.scp.Settings.Ff.MinFreq = math.Pow(10, logMin+shift)
		ff.scp.Settings.Ff.MaxFreq = math.Pow(10, logMax+shift)

		if ff.scp.ffMinFreqDisp != nil {
			ff.scp.ffMinFreqDisp.SetFloatValue(ff.scp.Settings.Ff.MinFreq, 2)
		}
		if ff.scp.ffMaxFreqDisp != nil {
			ff.scp.ffMaxFreqDisp.SetFloatValue(ff.scp.Settings.Ff.MaxFreq, 2)
		}
		ff.scp.SaveSettings()
		ff.scp.applyFfGenSettings(ff.scp.Settings.FfGen.On)
		if scroll {
			ff.frequencyOffsetFraction = 0
		} else {
			ff.frequencyOffsetFraction += dy
		}
		ff.scp.ResetFfSweep()
		return
	}
	// Vertical shift
	h := float64(ff.img.Bounds().Dy())
	ch := &ff.scp.Settings.Channels[chIndex]
	channelViewer := &ff.scp.channelViewers[chIndex]
	if scroll {
		channelViewer.ffDisplayOffsetFraction = dy + ff.offsetNToFf(ch.FfDisplayVOffset)
	} else {
		channelViewer.ffDisplayOffsetFraction += dy
	}
	if channelViewer.ffDisplayOffsetFraction < -h {
		channelViewer.ffDisplayOffsetFraction = -h
	}
	if channelViewer.ffDisplayOffsetFraction > h {
		channelViewer.ffDisplayOffsetFraction = h
	}
	ch.FfDisplayVOffset = ff.snapYToFfN(channelViewer.ffDisplayOffsetFraction)
	ff.scp.ffFullRefresh = true
	ff.scp.refreshRasters()
}

// dragged handles click-and-drag interactions.
// If the inspector is active, it moves the inspector coordinate.
// If an interactive label is focused, it performs vertical offset panning for channels
// or horizontal frequency range scrolling.
func (ff *ffViewer) dragged(dx, dy, x, y float32) {
	if ff.showInspector {
		ix, iy := ff.fyneToImg(x, y)
		ff.mouseX = float32(ix)
		ff.mouseY = float32(iy)
		bounds := ff.scp.ffScopeSignalScreen.Bounds()
		if ff.mouseX < float32(bounds.Min.X) {
			ff.mouseX = float32(bounds.Min.X)
		}
		if ff.mouseX > float32(bounds.Max.X-1) {
			ff.mouseX = float32(bounds.Max.X - 1)
		}
		if ff.mouseY < float32(bounds.Min.Y) {
			ff.mouseY = float32(bounds.Min.Y)
		}
		if ff.mouseY > float32(bounds.Max.Y-1) {
			ff.mouseY = float32(bounds.Max.Y - 1)
		}
		ff.scp.ffFullRefresh = true
		ff.enableRefresh()
		canvas.Refresh(ff.scp.ffRaster)
	}
	if ff.selectedChannel != -2 {
		if ff.selectedChannel == -1 {
			ff.setChDispOffset(-1, float64(dx), false)
		} else {
			ff.setChDispOffset(ff.selectedChannel, float64(dy), false)
		}
	}
}

// scrolled handles mouse wheel scroll events over interactive label coordinates.
func (ff *ffViewer) scrolled(delta, x, y float32) {
	ix, iy := ff.fyneToImg(x, y)
	p := image.Point{X: ix, Y: iy}
	for chIdx, bounds := range ff.labelBounds {
		if p.In(bounds) {
			if chIdx == -1 {
				ff.setChDispOffset(-1, float64(-delta)*100, true) // Arbitrary horizontal scroll scale
			} else {
				nY := (float64(ff.img.Bounds().Dy()) / float64(numberOfDivs)) / 10
				ff.setChDispOffset(chIdx, float64(-delta)*nY, true)
			}
			break
		}
	}
}

// typedKey handles arrow key presses when focusing interactive label coordinates.
func (ff *ffViewer) typedKey(x, y float32, keyName fyne.KeyName) {
	ix, iy := ff.fyneToImg(x, y)
	p := image.Point{X: ix, Y: iy}
	for chIdx, bounds := range ff.labelBounds {
		if p.In(bounds) {
			if chIdx == -1 {
				switch keyName {
				case fyne.KeyLeft:
					ff.scrolled(-scrollDelta, x, y)
				case fyne.KeyRight:
					ff.scrolled(scrollDelta, x, y)
				}
			} else {
				switch keyName {
				case fyne.KeyDown:
					ff.scrolled(-scrollDelta, x, y)
				case fyne.KeyUp:
					ff.scrolled(scrollDelta, x, y)
				}
			}
			break
		}
	}
}

// ffRasterGenerator is the primary generator callback used by Fyne to draw/redraw
// the entire F(f) Bode tab scope image.
// It recalculates display margins based on which channels are enabled (Amplitude) and have Phase axes active,
// updates the sub-image bounding rect for the signal screen, and delegates drawing tasks to registered drawers.
func (scp *ScpDesc) ffRasterGenerator(wInt int, hInt int) image.Image {
	// slog.Debug("ffRasterGenerator called", "w", wInt, "h", hInt)
	defer scp.screenLocker.Unlock()
	scp.screenLocker.Lock()

	// fullRefresh tells us if we must clear the raster and redraw the grid and labels.
	// We check the global ffFullRefresh flag which is set when the Run button is clicked,
	// or on window size/bounds changes.
	fullRefresh := scp.ffFullRefresh
	scp.ffFullRefresh = false

	if scp.ffScopeFullScreen == nil || scp.ffScopeFullScreen.Bounds().Dx() != wInt || scp.ffScopeFullScreen.Bounds().Dy() != hInt {
		scp.ffScopeFullScreen = scp.newScopeScreen(image.Point{wInt, hInt})
		fullRefresh = true
		if scp.ffViewer != nil {
			// Update existing viewer's image and dimensions
			fv, ok := scp.ffViewer.(*ffViewer)
			if ok {
				fv.img = scp.ffScopeFullScreen
				fv.imgRect = image.Rect(0, 0, wInt, hInt)
			}
		}
	}

	ip := scp.ffScopeFullScreen.(*image.RGBA)

	if fullRefresh {
		bgCol := scp.theme.Color(theme.ColorNameBackground, 0)
		draw.Draw(ip, ip.Bounds(), &image.Uniform{bgCol}, image.Point{}, draw.Src)
	}

	w := float64(wInt)
	h := float64(hInt)
	var leftMargin, rightMargin, topMargin, bottomMargin float64

	var ampChannelsCount int
	var hasPhase bool
	for i := 0; i < int(scp.channelCount); i++ {
		ch := scp.Settings.Channels[i]
		if ch.Enabled {
			if ch.Enabled {
				ampChannelsCount++
			}
			if ch.FfPhaseEnabled {
				hasPhase = true
			}
		}
	}
	ffActiveAxesCount := ampChannelsCount
	if hasPhase {
		ffActiveAxesCount++
	}

	leftColumnCount := (ffActiveAxesCount + 1) / 2
	rightColumnCount := ffActiveAxesCount / 2

	leftMargin = 35.0 + float64(leftColumnCount)*60.0
	rightMargin = 35.0 + float64(rightColumnCount)*60.0
	topMargin = float64(scp.Settings.Window.Height) * 0.05
	bottomMargin = float64(scp.Settings.Window.Height) * 0.08

	// F(f) doesn't need to be square like F(v). It can use the whole available space.
	wSigMax := w - leftMargin - rightMargin
	hSigMax := h - topMargin - bottomMargin

	scp.ffScopeSignalScreen = ip.SubImage(image.Rect(
		int(leftMargin),
		int(topMargin),
		int(leftMargin+wSigMax),
		int(topMargin+hSigMax),
	)).(draw.RGBA64Image)

	if scp.ffViewer == nil {
		scp.ffViewer = newFfViewer(scp.ffScopeFullScreen, image.Rect(0, 0, wInt, hInt), scp)
		scp.addFfDrawer(scp.ffViewer)
	}

	for i := range scp.ffDrawers {
		if fullRefresh {
			scp.ffDrawers[i].enableRefresh()
		}
		scp.ffDrawers[i].draw()
	}

	return ip.SubImage(ip.Bounds())
}

// newFfPanel initializes the graphical control panel container for the F(f) tab.
// It constructs channel checkboxes (Enabled, Phase, Reference), voltage range dropdowns,
// digit displays (Start/Stop frequency, Step, Delta T), and binds state updates
// to sweep calculations and generator registers.
func (scp *ScpDesc) newFfPanel(panel *fyne.Container) {
	// Clamp configuration settings to safe valid ranges to prevent out-of-bounds panics or initialization errors
	if scp.Settings.Ff.ReferenceChannel < 0 || scp.Settings.Ff.ReferenceChannel >= int(scp.channelCount) {
		scp.Settings.Ff.ReferenceChannel = 0
	}

	maxPossibleFreq := genericps.SineMaxFrequency
	if scp.Settings.Ff.UseExternalGen {
		maxPossibleFreq = 100000000.0 // 100 MHz
	} else {
		// Ensure enough digits are allocated for external generator even if currently using internal generator
		maxPossibleFreq = 100000000.0 // 100 MHz
	}

	if scp.Settings.Ff.MinFreq < genericps.MinFrequency {
		scp.Settings.Ff.MinFreq = genericps.MinFrequency
	}
	// Initial bounds check
	if scp.Settings.Ff.MaxFreq > maxPossibleFreq {
		scp.Settings.Ff.MaxFreq = maxPossibleFreq
	}
	if scp.Settings.Ff.MinFreq > scp.Settings.Ff.MaxFreq {
		scp.Settings.Ff.MinFreq = scp.Settings.Ff.MaxFreq
	}

	if scp.Settings.Ff.DeltaT < 0.001 {
		scp.Settings.Ff.DeltaT = 1.0
	}

	maxV := 2000000 // 2V peak-to-peak
	if scp.Settings.Ff.Amplitude <= 0 {
		scp.Settings.Ff.Amplitude = 2000000
	} else if scp.Settings.Ff.Amplitude > uint32(maxV) {
		scp.Settings.Ff.Amplitude = uint32(maxV)
	}

	vbox := container.New(layout.NewVBoxLayout())
	var refChecks []*widget.Check

	for i := 0; i < int(scp.channelCount); i++ {
		chIndex := genericps.ChannelId(i)
		chName := channelNames[i]

		// Channel Label
		text := "Ch " + chName + ":"
		if scp.isDigitalFilterEnabled(chIndex) {
			text += " ⚠️"
		}
		label := canvas.NewText(text, scp.Settings.Channels[chIndex].Col[scp.Settings.ChannelColorIndex])
		label.TextStyle = fyne.TextStyle{Bold: true}
		label.TextSize = theme.TextSize()
		scp.channelViewers[chIndex].ffNameLabel = label

		// Enabled Checkbox
		enabledCheck := widget.NewCheck("Enabled", func(b bool) {
			scp.EnableChannel(chIndex, b)
		})
		enabledCheck.SetChecked(scp.Settings.Channels[chIndex].Enabled)
		scp.channelViewers[chIndex].enableChecks = append(scp.channelViewers[chIndex].enableChecks, enabledCheck)

		// Phase Checkbox
		phaseCheck := widget.NewCheck("Phase", func(b bool) {
			scp.Settings.Channels[chIndex].FfPhaseEnabled = b
			scp.ResetFfSweep()
		})
		phaseCheck.SetChecked(scp.Settings.Channels[chIndex].FfPhaseEnabled)

		// Ref Radio (Check)
		refCheck := widget.NewCheck("Ref", nil)
		refChecks = append(refChecks, refCheck)

		// Range Selector
		rangesEnum, _ := scp.psControl.ChannelRanges(chIndex)
		var ranges []string
		for _, r := range rangesEnum {
			ranges = append(ranges, inputRanges[r])
		}
		vRange := selectscroll.NewSelectScroll(ranges, func(opt string, ex selectscroll.Exception) {
			scp.changeChannelRange(chIndex, opt)
		}, "+500m")
		scp.channelViewers[chIndex].vRangeSelects = append(scp.channelViewers[chIndex].vRangeSelects, vRange)
		vr := scp.Settings.Channels[chIndex].VRange
		if s, ok := rangeEnumToString[vr]; ok {
			vRange.SetSelected(s)
		}

		// X10 Checkbox
		x10Check := widget.NewCheck("X10", func(c bool) {
			scp.changeChannelX10(chIndex, c)
		})
		x10Check.SetChecked(scp.Settings.Channels[chIndex].X10)
		scp.channelViewers[chIndex].x10Checkboxes = append(scp.channelViewers[chIndex].x10Checkboxes, x10Check)

		// Arrange settings to minimize width
		row1 := container.New(layout.NewHBoxLayout(), label, enabledCheck, phaseCheck, refCheck)
		row2 := container.New(layout.NewHBoxLayout(), widget.NewLabel("Range:"), vRange, x10Check)

		chBox := container.New(layout.NewVBoxLayout(), row1, row2)
		if i > 0 {
			vbox.Add(layout.NewSpacer())
		}
		vbox.Add(chBox)

		addToTest(enabledCheck, ffEnableId+chName)
		addToTest(phaseCheck, ffPhaseCheckId+chName)
		addToTest(refCheck, ffRefCheckId+chName)
		addToTest(vRange, ffVRangeId+chName)
		addToTest(x10Check, ffX10Id+chName)
	}

	// Declare disp7 widgets first so they can be referenced in the OnChanged closure
	// deltaDisp, deltaTDisp removed

	for i := 0; i < int(scp.channelCount); i++ {
		idx := i
		refChecks[idx].OnChanged = func(b bool) {
			if b {
				scp.Settings.Ff.ReferenceChannel = idx
				for j, rc := range refChecks {
					if j != idx {
						rc.SetChecked(false)
					}
				}

				// Update disp7 widget colors based on reference channel
				refCol := scp.Settings.Channels[idx].Col[scp.Settings.ChannelColorIndex]
				if scp.ffMinFreqDisp != nil {
					scp.ffMinFreqDisp.SetOncolor(refCol)
					scp.ffMaxFreqDisp.SetOncolor(refCol)
					// deltaDisp and deltaTDisp color updates removed
					if scp.ffCurrentFreqDisp != nil {
						scp.ffCurrentFreqDisp.SetOncolor(refCol)
					}
				}

				scp.ResetFfSweep()
			}
		}
		if scp.Settings.Ff.ReferenceChannel == i {
			refChecks[i].SetChecked(true)
		}
	}

	panel.Add(vbox)

	// Horizontal Sweep Disp7 Widgets
	slog.Debug("newFfPanel", "maxPossibleFreq", maxPossibleFreq)
	numOfFractionDigits := 2
	numOfDigits := numOfFractionDigits
	f := int(math.Round(maxPossibleFreq))
	for f > 0 {
		f /= 10
		numOfDigits++
	}
	size := float32(0.8)
	refCol := scp.Settings.Channels[scp.Settings.Ff.ReferenceChannel].Col[scp.Settings.ChannelColorIndex]

	scp.ffMinFreqDisp, _ = disp7.NewCustomDisp7Array(numOfDigits, numOfFractionDigits,
		int(maxPossibleFreq)*pow10tab[numOfFractionDigits],
		int(genericps.MinFrequency)*pow10tab[numOfFractionDigits],
		disp7.UnSigned, disp7.NoTrailingZeroes, scp.Window,
		refCol,
		disp7.ReadWrite, size*disp7.DefaultDigitWidth,
		disp7.DeafultDigitHeight, 1,
		disp7.DefaultVCursorSpace, "Start:", " Hz")

	scp.ffMaxFreqDisp, _ = disp7.NewCustomDisp7Array(numOfDigits, numOfFractionDigits,
		int(maxPossibleFreq)*100, int(genericps.MinFrequency)*pow10tab[numOfFractionDigits],
		disp7.UnSigned, disp7.NoTrailingZeroes, scp.Window,
		refCol,
		disp7.ReadWrite, size*disp7.DefaultDigitWidth,
		disp7.DeafultDigitHeight, 1,
		disp7.DefaultVCursorSpace, "Stop :", " Hz")

	scp.ffMinFreqDisp.SetFloatValue(scp.Settings.Ff.MinFreq, 2)
	scp.ffMaxFreqDisp.SetFloatValue(scp.Settings.Ff.MaxFreq, 2)

	scp.ffCurrentFreqDisp, _ = disp7.NewCustomDisp7Array(numOfDigits, numOfFractionDigits,
		int(maxPossibleFreq)*pow10tab[numOfFractionDigits],
		int(genericps.MinFrequency)*pow10tab[numOfFractionDigits],
		disp7.UnSigned, disp7.NoTrailingZeroes, scp.Window,
		refCol,
		disp7.ReaOnly, size*disp7.DefaultDigitWidth,
		disp7.DeafultDigitHeight, 1,
		disp7.DefaultVCursorSpace, "Freq :", " Hz")
	scp.updateFfCurrentFreq()

	scp.ffStepFreqDisp, _ = disp7.NewCustomDisp7Array(3, 0,
		500,
		5,
		disp7.UnSigned, disp7.NoTrailingZeroes, scp.Window,
		refCol,
		disp7.ReadWrite, size*disp7.DefaultDigitWidth,
		disp7.DeafultDigitHeight, 1,
		disp7.DefaultVCursorSpace, "Pts/dec:", "")

	scp.ffStepFreqDisp.OnChanged = func(v float64) {
		scp.Settings.Ff.PtsDec = v
		scp.SaveSettings()
	}
	if scp.Settings.Ff.PtsDec < 5 {
		scp.Settings.Ff.PtsDec = 5
	}
	scp.ffStepFreqDisp.SetValue(int(scp.Settings.Ff.PtsDec))

	scp.ffDeltaTDisp, _ = disp7.NewCustomDisp7Array(5, 3,
		10000, 0,
		disp7.UnSigned, disp7.NoTrailingZeroes, scp.Window,
		refCol,
		disp7.ReadWrite, size*disp7.DefaultDigitWidth,
		disp7.DeafultDigitHeight, 1,
		disp7.DefaultVCursorSpace, "ΔT:", " s")
	scp.ffDeltaTDisp.SetFloatValue(scp.Settings.Ff.DeltaT, 3)

	syncGenStartStopAndStep := func() {
		scp.Settings.FfGen.StartFrequency = scp.Settings.Ff.MinFreq
		scp.Settings.FfGen.StopFrequency = scp.Settings.Ff.MaxFreq
		scp.Settings.FfGen.Frequency = scp.Settings.Ff.MinFreq
		scp.Settings.FfGen.Dwelltime = scp.Settings.Ff.DeltaT
		scp.SaveSettings()

		if scp.psControl != nil && scp.psControl.Con != nil && scp.psControl.Con.ID == genericps.SimId {
			scp.applyFfSimGenSettings(false)
			scp.applyFfSimGenSettings(scp.Settings.FfGen.On)
		} else {
			scp.applyFfGenSettings(false)
			scp.applyFfGenSettings(scp.Settings.FfGen.On)
		}
	}

	scp.ffDeltaTDisp.OnChanged = func(v float64) {
		scp.Settings.Ff.DeltaT = v / 1000.0
		scp.ResetFfSweep()
		syncGenStartStopAndStep()
	}

	scp.ffMinFreqDisp.OnChanged = func(v float64) {
		scp.Settings.Ff.MinFreq = v / 100.0
		if scp.Settings.Ff.MinFreq > scp.Settings.Ff.MaxFreq {
			scp.Settings.Ff.MaxFreq = scp.Settings.Ff.MinFreq
			if scp.ffMaxFreqDisp != nil {
				scp.ffMaxFreqDisp.SetFloatValue(scp.Settings.Ff.MaxFreq, 2)
			}
		}
		scp.ResetFfSweep()
		syncGenStartStopAndStep()
	}
	scp.ffMaxFreqDisp.OnChanged = func(v float64) {
		scp.Settings.Ff.MaxFreq = v / 100.0
		if scp.Settings.Ff.MaxFreq < scp.Settings.Ff.MinFreq {
			scp.Settings.Ff.MinFreq = scp.Settings.Ff.MaxFreq
			if scp.ffMinFreqDisp != nil {
				scp.ffMinFreqDisp.SetFloatValue(scp.Settings.Ff.MinFreq, 2)
			}
		}
		scp.ResetFfSweep()
		syncGenStartStopAndStep()
	}

	if scp.Settings.Dft.DisplayMode == "" {
		scp.Settings.Dft.DisplayMode = "dB"
	}
	dispModeSelect := selectscroll.NewSelectScroll([]string{settings.ModeVoltage, settings.ModeDB}, func(opt string, ex selectscroll.Exception) {
		scp.Settings.Dft.DisplayMode = opt
		scp.ResetFfSweep()
		scp.SaveSettings()
	}, settings.ModeVoltage)
	dispModeSelect.SilentSetSelected(scp.Settings.Dft.DisplayMode)

	dispModeControls := container.NewHBox(widget.NewLabel(" Mode:"), dispModeSelect)

	// Generator controls container
	genHeader := widget.NewLabelWithStyle("Generator Settings", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})

	isSim := false
	if scp.psControl != nil && scp.psControl.Con != nil && scp.psControl.Con.ID == genericps.SimId {
		isSim = true
	}

	var genPanel *fyne.Container
	var genErr error
	if isSim {
		genPanel, genErr = scp.newFfSimGenPanel()
	} else {
		genPanel, genErr = scp.newFfGenPanel()
	}
	if genErr != nil {
		slog.Error("Failed to create generator panel", "err", genErr)
	}

	scp.useExtGenCheck = widget.NewCheck("Use external generator", func(checked bool) {
		scp.Settings.Ff.UseExternalGen = checked
		scp.updateFfWidgetLimits()
		scp.SaveSettings()
		
		if scp.psControl != nil && scp.psControl.Con != nil && scp.psControl.Con.ID == genericps.SimId {
			scp.applyFfSimGenSettings(scp.Settings.FfGen.On)
		} else {
			scp.applyFfGenSettings(scp.Settings.FfGen.On)
		}
	})
	scp.useExtGenCheck.SetChecked(scp.Settings.Ff.UseExternalGen)
	if !scp.ExtGenEnabled || !scp.extGen.Connected() {
		scp.useExtGenCheck.Hide()
	}

	genVBox := container.NewVBox(
		layout.NewSpacer(),
		genHeader,
		scp.useExtGenCheck,
	)
	if genPanel != nil {
		genVBox.Add(genPanel)
	}

	infoHead := widget.NewLabelWithStyle("Status / Info", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	genVBox.Add(layout.NewSpacer())
	genVBox.Add(infoHead)
	genVBox.Add(scp.ffCurrentFreqDisp)

	genSettings := container.New(layout.NewVBoxLayout(), scp.ffMinFreqDisp,
		scp.ffMaxFreqDisp, scp.ffDeltaTDisp, scp.ffStepFreqDisp,
		dispModeControls, genVBox)
	panel.Add(genSettings)

	addToTest(scp.ffMinFreqDisp, ffMinFreqId)
	addToTest(scp.ffMaxFreqDisp, ffMaxFreqId)
	addToTest(scp.ffCurrentFreqDisp, ffCurrentFreqId)
	addToTest(dispModeSelect, ffDispModeSelectId)
	addToTest(scp.useExtGenCheck, ffExtGenSelectId)

	scp.updateFfWidgetLimits()
}

// updateFfWidgetLimits adjusts the min/max limits of the frequency display widgets
// to match the selected generator (internal scope generator or external USB device).
// It also clamps any stored settings values that fall outside the new limits.
func (scp *ScpDesc) updateFfWidgetLimits() {
	if scp.ffMinFreqDisp == nil || scp.ffMaxFreqDisp == nil || scp.ffCurrentFreqDisp == nil {
		return
	}

	var minF, maxF float64
	if scp.Settings.Ff.UseExternalGen {
		minF = 0.01        // 10 mHz
		maxF = 100000000.0 // 100 MHz
	} else {
		minF = genericps.MinFrequency
		maxF = genericps.SineMaxFrequency
	}

	fractionWidth := 2
	scale := int(math.Pow10(fractionWidth))

	minLimitVal := int(minF * float64(scale))
	maxLimitVal := int(maxF * float64(scale))

	scp.ffMinFreqDisp.Value = int(scp.Settings.Ff.MinFreq * float64(scale))
	scp.ffMaxFreqDisp.Value = int(scp.Settings.Ff.MaxFreq * float64(scale))

	scp.ffMinFreqDisp.SetMinMax(minLimitVal, maxLimitVal)
	scp.ffMaxFreqDisp.SetMinMax(minLimitVal, maxLimitVal)
	scp.ffCurrentFreqDisp.SetMinMax(minLimitVal, maxLimitVal)

	// In case the values were clamped, update settings to match
	scp.Settings.Ff.MinFreq = float64(scp.ffMinFreqDisp.Value) / float64(scale)
	scp.Settings.Ff.MaxFreq = float64(scp.ffMaxFreqDisp.Value) / float64(scale)

}

// ResetFfSweep clears all recorded Bode sweep point data from the channels' buffers
// and schedules a full scope screen refresh to restart the frequency sweep.
func (scp *ScpDesc) ResetFfSweep() {
	scp.ffLocker.Lock()
	for i := range scp.bodeBuffers {
		scp.bodeBuffers[i] = nil
	}
	scp.ffLocker.Unlock()
	scp.ffFullRefresh = true
	scp.refreshRasters()
}

// generateLogFrequencies generates logarithmically spaced frequency points
// between minFreq and maxFreq. The DeltaFreq setting controls the number of
// points per decade (e.g. DeltaFreq=50 means 50 points per decade).
func generateLogFrequencies(minFreq, maxFreq, pointsPerDecade float64) []float64 {
	if minFreq <= 0 {
		minFreq = 1.0
	}
	if maxFreq <= minFreq {
		return []float64{minFreq}
	}
	if pointsPerDecade < 5 {
		pointsPerDecade = 5
	}
	if pointsPerDecade > 500 {
		pointsPerDecade = 500
	}

	logMin := math.Log10(minFreq)
	logMax := math.Log10(maxFreq)
	decades := logMax - logMin
	totalPoints := int(math.Ceil(decades * pointsPerDecade))
	if totalPoints < 2 {
		totalPoints = 2
	}

	freqs := make([]float64, totalPoints+1)
	for i := 0; i <= totalPoints; i++ {
		logF := logMin + (logMax-logMin)*float64(i)/float64(totalPoints)
		freqs[i] = math.Pow(10, logF)
	}
	return freqs
}

// startFfSweep launches the application-controlled Bode frequency sweep.
// It generates logarithmically spaced frequency points and steps through them,
// setting the generator to a fixed frequency at each step and waiting for
// DeltaT seconds of dwell time before advancing. This replaces the old
// approach of relying on the simulator's SweepController.
func (scp *ScpDesc) startFfSweep() {
	scp.stopFfSweep() // stop any existing sweep
	scp.ResetFfSweep() // clear old sweep data

	scp.ffSweepQuit = make(chan struct{})
	scp.ffSweepDataReady = make(chan struct{}, 1)
	quit := scp.ffSweepQuit

	pointsPerDecade := scp.Settings.Ff.PtsDec
	if pointsPerDecade <= 0 {
		pointsPerDecade = 50
	}

	freqs := generateLogFrequencies(
		scp.Settings.Ff.MinFreq,
		scp.Settings.Ff.MaxFreq,
		pointsPerDecade,
	)

	dwellTime := scp.Settings.Ff.DeltaT
	if dwellTime < 0.01 {
		dwellTime = 0.01
	}

	// Pre-allocate bodeBuffers so no append growth occurs during the sweep.
	scp.ffLocker.Lock()
	for i := range scp.bodeBuffers {
		scp.bodeBuffers[i] = make([]bodePoint, 0, len(freqs))
	}
	scp.ffLocker.Unlock()

	go func() {
		slog.Debug("startFfSweep", "points", len(freqs), "ppd", pointsPerDecade,
			"min", scp.Settings.Ff.MinFreq, "max", scp.Settings.Ff.MaxFreq,
			"dwell", dwellTime)

		for _, f := range freqs {
			// Drain any stale data ready signals
			for {
				select {
				case <-scp.ffSweepDataReady:
				default:
					goto Drained
				}
			}
		Drained:
			select {
			case <-quit:
				slog.Debug("startFfSweep: quit signal received")
				return
			default:
			}

			// Set the current target frequency
			scp.currentFfFreq = f

			// Update the frequency display
			scp.measuredFfFreq = f
			fyne.Do(func() {
				scp.updateFfCurrentFreq()
				// Set the generator to this fixed frequency
				scp.setGeneratorFreq(f)
				// Update acquisition parameters for the new target frequency
				scp.updateAcquisitionParameters()
			})

			// Dwell at this frequency — the scope's RefreshCallback will
			// call processFfData which records the Bode points
			// We wait for both dwellTime (for the physical system to settle)
			// and maxScreenTime (to fill the scope buffer with the new signal).
			totalWait := dwellTime
			if scp.maxScreenTime > 0 {
				totalWait += scp.maxScreenTime
			}

			scp.ffLocker.Lock()
			scp.ffSweepAcquireTime = time.Now().Add(time.Duration(totalWait * float64(time.Second)))
			scp.ffLocker.Unlock()

			timer := time.NewTimer(time.Duration(totalWait * float64(time.Second)))
			dwellExpired := false
			dataReady := false

			for !dwellExpired || !dataReady {
				select {
				case <-quit:
					timer.Stop()
					slog.Debug("startFfSweep: quit during dwell/acq wait")
					return
				case <-timer.C:
					dwellExpired = true
				case <-scp.ffSweepDataReady:
					dataReady = true
				}
			}
		}

		slog.Debug("startFfSweep: sweep complete")
		// Automatically stop the run block mode when the sweep completes
		fyne.Do(func() {
			scp.StopRunning()
		})
	}()
}

// stopFfSweep signals the running sweep goroutine to stop.
func (scp *ScpDesc) stopFfSweep() {
	if scp.ffSweepQuit != nil {
		close(scp.ffSweepQuit)
		scp.ffSweepQuit = nil
	}
}

// addFfDrawer registers a custom drawing component (like ffViewer) to be called
// by the raster generator during screen rendering.
func (scp *ScpDesc) addFfDrawer(d drawer) {
	scp.ffDrawers = append(scp.ffDrawers, d)
}

// deleteFfDrawer removes d from the ffDrawers slice (order-preserving).
func (scp *ScpDesc) deleteFfDrawer(d drawer) {
	for i, v := range scp.ffDrawers {
		if v == d {
			scp.ffDrawers = append(scp.ffDrawers[:i], scp.ffDrawers[i+1:]...)
			return
		}
	}
}

// measurePeriod estimates the signal period (in seconds) from a sample buffer
// by detecting rising zero-crossings relative to the signal mean and averaging
// the intervals between them. Returns an error if fewer than 2 crossings are found.
func (scp *ScpDesc) measurePeriod(buf []float32, samplingInterval float64) (float64, error) {
	if len(buf) < 10 {
		return 0, fmt.Errorf("buffer too short")
	}

	// Calculate average to find zero-crossing level
	sum := 0.0
	for _, v := range buf {
		sum += float64(v)
	}
	avg := sum / float64(len(buf))

	// Find min/max to calculate hysteresis
	minVal := buf[0]
	maxVal := buf[0]
	for _, v := range buf {
		if v < minVal {
			minVal = v
		}
		if v > maxVal {
			maxVal = v
		}
	}
	pkToPk := float64(maxVal - minVal)
	hysteresis := 0.1 * pkToPk
	if hysteresis < 10.0 { // fallback minimum threshold
		hysteresis = 10.0
	}

	// Schmitt trigger state machine
	// state: -1 = low, 1 = high, 0 = unknown
	state := 0
	var crossings []float64

	// First determine initial state
	if float64(buf[0])-avg > hysteresis {
		state = 1
	} else if float64(buf[0])-avg < -hysteresis {
		state = -1
	}

	for i := 0; i < len(buf)-1; i++ {
		v1 := float64(buf[i]) - avg
		v2 := float64(buf[i+1]) - avg

		switch state {
		case -1:
			if v2 > hysteresis {
				// Transition to high state (rising edge crossing)
				// Estimate crossing at avg (0) using linear interpolation between v1 and v2
				t := float64(i) + (-v1 / (v2 - v1))
				crossings = append(crossings, t)
				state = 1
			}
		case 1:
			if v2 < -hysteresis {
				// Transition to low state (falling edge)
				state = -1
			}
		default:
			// Unknown state: wait until we cross threshold
			if v2 > hysteresis {
				state = 1
			} else if v2 < -hysteresis {
				state = -1
			}
		}
	}

	// We need at least 2 crossings to measure a period
	if len(crossings) < 2 {
		return 0, fmt.Errorf(ErrFrequencyCannotBeDetected)
	}

	// Calculate average period in samples
	totalT := 0.0
	for i := 0; i < len(crossings)-1; i++ {
		totalT += (crossings[i+1] - crossings[i])
	}
	avgTInSamples := totalT / float64(len(crossings)-1)

	return avgTInSamples * samplingInterval, nil
}

// processFfData is the core signal processing function called when new scope channel buffers arrive.
// It estimates the actual frequency on the reference channel, computes FFTs for all enabled channels,
// determines the signal magnitude and relative phase shift at the measured peak frequency bin,
// performs envelope tracking (peak-hold) to keep the highest values, blends new phase estimates using vector averaging,
// and appends/sorts new sweep points into the channels' Bode buffers.
func (scp *ScpDesc) processFfData() {
	scp.ffLocker.Lock()
	acquireTime := scp.ffSweepAcquireTime
	scp.ffLocker.Unlock()

	// Only process data if the required wait time for the current sweep step has elapsed.
	// This ensures we do not accumulate FFTs containing historical data from the previous frequency.
	if time.Now().Before(acquireTime) {
		slog.Info("processFfData early return: before acquireTime", "now", time.Now(), "acquireTime", acquireTime)
		return
	}

	refCh := scp.Settings.Ff.ReferenceChannel
	if !scp.Settings.Channels[refCh].Enabled {
		slog.Info("processFfData early return: refCh not enabled", "refCh", refCh)
		return
	}

	// 1. Measure actual frequency using measurePeriod on the reference channel
	refBuf := scp.displayBuffers[refCh]
	samplingInterval := float64(scp.psControl.SamplingTimeInterval)
	if samplingInterval <= 0 {
		samplingInterval = 1E-6 // fallback
	}

	period, err := scp.measurePeriod(refBuf, samplingInterval)
	var measuredFreq float64
	if err == nil && period > 0 {
		measuredFreq = 1.0 / period
	} else {
		// If frequency cannot be detected via zero-crossings (e.g., signal attenuated at high freq),
		// fall back to the known sweep target frequency if we are currently sweeping.
		if scp.currentFfFreq > 0 {
			measuredFreq = scp.currentFfFreq
		} else {
			// Not sweeping and no signal, skip measurement
			slog.Info("processFfData early return: no signal and no sweep target", "err", err, "period", period)
			return
		}
	}

	// Use the app-controlled target frequency for Bode point placement.
	// The sweep goroutine manages currentFfFreq and the frequency display.
	// Fall back to measuredFreq if no target is set (legacy/non-sweep mode).
	targetFreq := scp.currentFfFreq
	if targetFreq <= 0 {
		targetFreq = measuredFreq
	}

	// 2. Compute FFT for each channel to get magnitude at measured frequency
	samplesLen := len(refBuf)
	if samplesLen < 4 {
		slog.Info("processFfData early return: samplesLen < 4", "len", samplesLen)
		return
	}

	scp.ensureFfFft(samplesLen)
	fft := scp.ffFftObj
	fftBuf := scp.ffFftBuf
	fftResult := scp.ffFftResult

	// We apply the same window configuration as DFT raster
	window := settings.WindowFlatTop
	normFactor := (float64(samplesLen) / 2.0) * scp.getCoherentGain(window, samplesLen)
	if normFactor <= 0 {
		normFactor = 1.0
	}

	// Find the peak bin in the reference channel around the measured frequency
	var refPeakBin int
	var refPhase float64
	{
		for j, v := range refBuf {
			fftBuf[j] = float64(v)
		}
		applyWindow(fftBuf, window)
		fftResult = fft.Coefficients(fftResult, fftBuf)

		// Expected bin
		expectedBin := int(math.Round(measuredFreq * float64(samplesLen) * samplingInterval))
		if expectedBin >= len(fftResult) {
			expectedBin = len(fftResult) - 1
		}
		if expectedBin < 1 {
			expectedBin = 1
		}

		// Search in a window around expectedBin
		searchStart := expectedBin - 5
		if searchStart < 1 {
			searchStart = 1
		}
		searchEnd := expectedBin + 5
		if searchEnd >= len(fftResult) {
			searchEnd = len(fftResult) - 1
		}

		maxMag := -1.0
		refPeakBin = expectedBin
		for bin := searchStart; bin <= searchEnd; bin++ {
			mag := cmplx.Abs(fftResult[bin])
			if mag > maxMag {
				maxMag = mag
				refPeakBin = bin
			}
		}
		if refPeakBin < 1 {
			refPeakBin = 1
		}
		refPhase = cmplx.Phase(fftResult[refPeakBin])
	}

	// Now compute magnitudes and phases for all enabled channels at refPeakBin
	scp.ffLocker.Lock()
	defer scp.ffLocker.Unlock()

	for i := 0; i < int(scp.channelCount); i++ {
		ch := scp.Settings.Channels[i]
		if !ch.Enabled {
			continue
		}

		chBuf := scp.displayBuffers[i]
		if len(chBuf) != samplesLen {
			continue
		}

		for j, v := range chBuf {
			fftBuf[j] = float64(v)
		}
		applyWindow(fftBuf, window)
		chFftResult := fft.Coefficients(scp.ffFftResult, fftBuf)

		// Get magnitude and phase at the reference peak bin
		cVal := chFftResult[refPeakBin]
		mag := cmplx.Abs(cVal) / normFactor
		chPhaseRad := cmplx.Phase(cVal)

		// Phase relative to the reference channel
		relPhaseRad := chPhaseRad - refPhase

		// Wrap/normalize to (-Pi, Pi]
		for relPhaseRad > math.Pi {
			relPhaseRad -= 2 * math.Pi
		}
		for relPhaseRad <= -math.Pi {
			relPhaseRad += 2 * math.Pi
		}

		// Convert to degrees
		relPhaseDeg := relPhaseRad * 180.0 / math.Pi

		// Find if we already have a bodePoint for this targetFreq (close match).
		// 5% tolerance absorbs the jitter from measurePeriod so repeated measurements
		// at the same generator frequency always land in the same bucket.
		const relThreshold = 0.05 // 5% frequency matching tolerance
		found := false
		for idx, pt := range scp.bodeBuffers[i] {
			relDiff := math.Abs(pt.targetFreq - targetFreq)
			if pt.targetFreq > 0 {
				relDiff /= pt.targetFreq
			}
			if relDiff < relThreshold {
				// Accumulate amplitude as a running average (same as phase) to
				// smooth out measurement noise and eliminate zig-zag spikes.
				scp.bodeBuffers[i][idx].ampSum += mag
				scp.bodeBuffers[i][idx].amp = scp.bodeBuffers[i][idx].ampSum / float64(scp.bodeBuffers[i][idx].count+1)
				// Blend phase using running vector average for smooth results.
				newRad := relPhaseDeg * math.Pi / 180.0
				scp.bodeBuffers[i][idx].phaseCos += math.Cos(newRad)
				scp.bodeBuffers[i][idx].phaseSin += math.Sin(newRad)
				scp.bodeBuffers[i][idx].phase = math.Atan2(
					scp.bodeBuffers[i][idx].phaseSin,
					scp.bodeBuffers[i][idx].phaseCos,
				) * 180.0 / math.Pi
				scp.bodeBuffers[i][idx].count++
				scp.bodeBuffers[i][idx].needsDraw = true
				scp.ffFullRefresh = true
				found = true
				break
			}
		}

		if !found {
			initRad := relPhaseDeg * math.Pi / 180.0
			newPt := bodePoint{
				targetFreq: targetFreq,
				freq:       targetFreq,
				amp:        mag,
				ampSum:     mag,
				phase:      relPhaseDeg,
				phaseCos:   math.Cos(initRad),
				phaseSin:   math.Sin(initRad),
				count:      1,
				needsDraw:  true,
			}
			scp.bodeBuffers[i] = append(scp.bodeBuffers[i], newPt)

			// Sort by frequency
			sort.Slice(scp.bodeBuffers[i], func(a, b int) bool {
				return scp.bodeBuffers[i][a].freq < scp.bodeBuffers[i][b].freq
			})
			scp.ffFullRefresh = true
		}
	}

	if scp.ffSweepDataReady != nil {
		slog.Info("processFfData: sending to ffSweepDataReady", "freq", targetFreq)
		select {
		case scp.ffSweepDataReady <- struct{}{}:
		default:
		}
	}
}

// digitEntry is a customized Fyne entry widget designed to input numerical digits.
// It overrides the MinSize method to enforce a minimum width suitable for multi-digit numbers.
type digitEntry struct {
	widget.Entry
}

// MinSize returns the minimum size required by the digit entry widget.
// Enforces a minimum width of 90 units to ensure numerical values aren't clipped.
func (d *digitEntry) MinSize() fyne.Size {
	m := d.Entry.MinSize()
	// 90 units of width comfortably fits 8-10 digits in standard Fyne layouts
	if m.Width < 90 {
		m.Width = 90
	}
	return m
}

// newDigitEntry creates and initializes a new digit entry widget.
func newDigitEntry() *digitEntry {
	e := &digitEntry{}
	e.ExtendBaseWidget(e)
	return e
}

// updateFfCurrentFreq retrieves the latest measured frequency
// and updates the corresponding seven-segment frequency display widget.
func (scp *ScpDesc) updateFfCurrentFreq() {
	if scp.ffCurrentFreqDisp == nil {
		return
	}
	f := scp.measuredFfFreq
	if f <= 0 {
		f = 0
	}
	scp.ffCurrentFreqDisp.SetFloatValue(f, 2)
	scp.ffCurrentFreqDisp.Refresh()
}

// ensureFfFft lazily initialises (or re-initialises when the sample count changes) the shared
// FFT object and its input/output buffers used by processFfData.  Calling this before every
// use avoids the ~O(n) twiddle-factor allocation that fourier.NewFFT would incur on every
// scope buffer callback.
func (scp *ScpDesc) ensureFfFft(samplesLen int) {
	if scp.ffFftSamples == samplesLen && scp.ffFftObj != nil {
		return
	}
	scp.ffFftObj = fourier.NewFFT(samplesLen)
	scp.ffFftBuf = make([]float64, samplesLen)
	scp.ffFftResult = make([]complex128, samplesLen/2+1)
	scp.ffFftSamples = samplesLen
}

// smoothPhase applies a moving average window over the phase responses.
// To handle the -180 to +180 degrees discontinuity correctly, it converts phases to unit vectors (sine/cosine),
// computes the moving average of these components, and transforms the resulting average vector back to degrees.
func smoothPhase(pts []bodePoint, windowSize int) []float64 {
	return smoothPhaseInto(pts, windowSize, make([]float64, len(pts)))
}

// smoothPhaseInto is like smoothPhase but writes results into the caller-supplied output buffer (len must equal len(pts)).
// Use this when the caller can cache the output buffer to avoid per-call heap allocation.
func smoothPhaseInto(pts []bodePoint, windowSize int, out []float64) []float64 {
	if len(pts) == 0 {
		return out
	}
	if windowSize < 1 {
		windowSize = 1
	}
	halfWin := windowSize / 2

	for i := range pts {
		sumCos := 0.0
		sumSin := 0.0
		count := 0
		for k := -halfWin; k <= halfWin; k++ {
			idx := i + k
			if idx >= 0 && idx < len(pts) {
				rad := pts[idx].phase * math.Pi / 180.0
				sumCos += math.Cos(rad)
				sumSin += math.Sin(rad)
				count++
			}
		}
		if count > 0 {
			avgCos := sumCos / float64(count)
			avgSin := sumSin / float64(count)
			smoothedRad := math.Atan2(avgSin, avgCos)
			out[i] = smoothedRad * 180.0 / math.Pi
		} else {
			out[i] = pts[i].phase
		}
	}
	return out
}

func (scp *ScpDesc) newFfGenPanel() (box *fyne.Container, err error) {
	checked := func(c bool) {
		scp.Settings.FfGen.On = c
		scp.applyFfGenSettings(c)
		scp.SaveSettings()
	}
	check := widget.NewCheck("On", checked)
	check.Checked = scp.Settings.FfGen.On

	size := float32(0.8)
	refCol := scp.Settings.Channels[scp.Settings.Ff.ReferenceChannel].Col[scp.Settings.ChannelColorIndex]
	maxV := 2000000

	scp.ffAmpDisp, err = disp7.NewCustomDisp7Array(7, 6, maxV, 0,
		disp7.SignedHidden, disp7.NoTrailingZeroes, scp.Window,
		refCol,
		disp7.ReadWrite, size*disp7.DefaultDigitWidth,
		disp7.DeafultDigitHeight, 1,
		disp7.DefaultVCursorSpace, "Amp  :", " V")
	if err != nil {
		return nil, err
	}
	scp.ffAmpDisp.SetValue(int(scp.Settings.FfGen.Amplitude))
	scp.ffAmpDisp.OnChanged = func(v float64) {
		scp.Settings.FfGen.Amplitude = uint32(v)
		scp.SaveSettings()
		if scp.ExtGenEnabled && scp.Settings.Ff.UseExternalGen && scp.extGen.Connected() {
			scp.extGen.SetAmplitude(scpi.Ch1, float64(scp.Settings.FfGen.Amplitude)/1000000.0)
		} else if scp.running {
			scp.applyFfGenSettings(check.Checked)
		}
	}

	scp.ffOffsetDisp, err = disp7.NewCustomDisp7Array(7, 6,
		maxV, -maxV,
		disp7.Signed, disp7.NoTrailingZeroes, scp.Window,
		refCol,
		disp7.ReadWrite, size*disp7.DefaultDigitWidth,
		disp7.DeafultDigitHeight, 1,
		disp7.DefaultVCursorSpace, "Offset   :", " V")
	if err != nil {
		return nil, err
	}
	scp.ffOffsetDisp.SetValue(int(scp.Settings.FfGen.OffsetVoltage))
	scp.ffOffsetDisp.OnChanged = func(v float64) {
		scp.Settings.FfGen.OffsetVoltage = int32(v)
		scp.SaveSettings()
		if scp.ExtGenEnabled && scp.Settings.Ff.UseExternalGen && scp.extGen.Connected() {
			scp.extGen.SetOffset(scpi.Ch1, float64(scp.Settings.FfGen.OffsetVoltage)/1000000.0)
		} else if scp.running {
			scp.applyFfGenSettings(check.Checked)
		}
	}

	top := container.NewHBox(check, widget.NewLabel("Wave: Sine"), widget.NewLabel("Sweep: Up"))
	box = container.NewVBox(top, scp.ffAmpDisp, scp.ffOffsetDisp)
	return box, nil
}

func (scp *ScpDesc) newFfSimGenPanel() (box *fyne.Container, err error) {
	checked := func(c bool) {
		scp.Settings.FfGen.On = c
		scp.applyFfSimGenSettings(c)
		scp.SaveSettings()
	}
	check := widget.NewCheck("On", checked)
	check.Checked = scp.Settings.FfGen.On

	size := float32(0.8)
	chCol := scp.Settings.Channels[0].Col[scp.Settings.ChannelColorIndex]
	maxV := 2000000

	scp.ffAmpDisp, err = disp7.NewCustomDisp7Array(7, 6, maxV, 0,
		disp7.SignedHidden, disp7.NoTrailingZeroes, scp.Window,
		chCol,
		disp7.ReadWrite, size*disp7.DefaultDigitWidth,
		disp7.DeafultDigitHeight, 1,
		disp7.DefaultVCursorSpace, "Amplitude:", " V")
	if err != nil {
		return nil, err
	}
	scp.ffAmpDisp.SetValue(int(scp.Settings.FfGen.Amplitude))
	scp.ffAmpDisp.OnChanged = func(v float64) {
		scp.Settings.FfGen.Amplitude = uint32(v)
		scp.SaveSettings()
		if scp.ExtGenEnabled && scp.Settings.Ff.UseExternalGen && scp.extGen.Connected() {
			scp.extGen.SetAmplitude(scpi.Ch1, float64(scp.Settings.FfGen.Amplitude)/1000000.0)
		} else {
			scp.applyFfSimGenSettings(check.Checked)
		}
	}

	scp.ffOffsetDisp, err = disp7.NewCustomDisp7Array(7, 6,
		maxV, -maxV,
		disp7.Signed, disp7.NoTrailingZeroes, scp.Window,
		chCol,
		disp7.ReadWrite, size*disp7.DefaultDigitWidth,
		disp7.DeafultDigitHeight, 1,
		disp7.DefaultVCursorSpace, "Offset   :", " V")
	if err != nil {
		return nil, err
	}
	scp.ffOffsetDisp.SetValue(int(scp.Settings.FfGen.OffsetVoltage))
	scp.ffOffsetDisp.OnChanged = func(v float64) {
		scp.Settings.FfGen.OffsetVoltage = int32(v)
		scp.SaveSettings()
		if scp.ExtGenEnabled && scp.Settings.Ff.UseExternalGen && scp.extGen.Connected() {
			scp.extGen.SetOffset(scpi.Ch1, float64(scp.Settings.FfGen.OffsetVoltage)/1000000.0)
		} else {
			scp.applyFfSimGenSettings(check.Checked)
		}
	}


	top := container.NewHBox(check, widget.NewLabel("Wave: Sine"), widget.NewLabel("Channel: Ch A"))
	box = container.NewVBox(
		top,
		scp.ffAmpDisp,
		scp.ffOffsetDisp,
	)
	return box, nil
}
