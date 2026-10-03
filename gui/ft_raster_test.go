package gui

import (
	"fynescope/control"
	"fynescope/settings"
	"image"
	"image/color"
	"math"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
	"github.com/stretchr/testify/assert"
)

func TestSignalViewer_MouseIn(t *testing.T) {
	rect := image.Rect(10, 20, 100, 50)
	img := image.NewRGBA(image.Rect(0, 0, 200, 100))
	scp := &ScpDesc{
		Settings: &settings.PsSettings{},
	}

	sv := newSignalViewer(img, rect, scp, false)

	assert.True(t, sv.mouseIn(15, 25), "Point inside should be true")
	assert.True(t, sv.mouseIn(10, 20), "Point on top-left border should be true")

	// Max bounds are exclusive in image.Rect
	assert.False(t, sv.mouseIn(100, 49), "Point on right border should be false")
	assert.False(t, sv.mouseIn(99, 50), "Point on bottom border should be false")
	assert.False(t, sv.mouseIn(5, 25), "Point outside left should be false")
	assert.False(t, sv.mouseIn(15, 10), "Point outside top should be false")
}

func TestSignalViewer_MouseEvents(t *testing.T) {
	rect := image.Rect(10, 20, 100, 50)
	img := image.NewRGBA(image.Rect(0, 0, 200, 100))
	scp := &ScpDesc{
		Settings: &settings.PsSettings{},
	}

	sv := newSignalViewer(img, rect, scp, false)

	// Initial state
	assert.False(t, sv.showInspector)
	assert.False(t, sv.refActive)

	// Mouse down inside rect with RightMouseButton -> shows inspector
	sv.mouseDown(desktop.MouseButtonSecondary, 0, 15, 25)
	assert.True(t, sv.showInspector)
	assert.False(t, sv.refActive)

	// Mouse up
	sv.mouseUp(desktop.MouseButtonSecondary, 0, 15, 25)
	assert.False(t, sv.showInspector)

	// Mouse down inside rect with RightMouseButton + Shift -> sets reference
	sv.mouseDown(desktop.MouseButtonSecondary, fyne.KeyModifierShift, 15, 25)
	assert.True(t, sv.refActive)
	assert.True(t, sv.refDragging)

	sv.mouseUp(desktop.MouseButtonSecondary, fyne.KeyModifierShift, 15, 25)
	assert.False(t, sv.refDragging)
	// refActive remains true after mouseUp until deleted
}

func TestSignalViewer_TypedKey(t *testing.T) {
	rect := image.Rect(10, 20, 100, 50)
	img := image.NewRGBA(image.Rect(0, 0, 200, 100))
	scp := &ScpDesc{
		Settings: &settings.PsSettings{},
	}

	sv := newSignalViewer(img, rect, scp, false)

	// Make sure typedKey doesn't panic on a dummy scope
	sv.typedKey(20, 30, fyne.KeyDown)
	sv.typedKey(20, 30, fyne.KeyUp)

	assert.True(t, true)
}

func TestSignalViewer_SincTriggerPointAlignment(t *testing.T) {
	bounds := image.Rect(0, 0, 1000, 500)
	targetImg := image.NewRGBA(bounds)
	red := color.NRGBA{R: 255, G: 0, B: 0, A: 255}

	scp := &ScpDesc{
		Settings: settings.NewDefaultSettings(),
		ftScopeSignalScreen: targetImg,
		maxScreenTime:       100e-9, // 100 ns screen (e.g. 10 ns/div)
	}
	scp.Settings.Time.Interpolation = settings.Sinc
	scp.Settings.Time.TriggerTimeOffset = 50e-9 // 50 ns (center of screen, pixel x = 500)
	scp.channelViewers = make([]channelViewerDesc, 1)
	scp.Settings.Channels = make([]settings.ChSettings, 1)
	scp.Settings.Channels[0].Enabled = true
	scp.Settings.Channels[0].VRange = 7 // e.g. 5V (5000 mV)
	scp.Settings.Channels[0].Col[scp.Settings.ChannelColorIndex] = red

	// 20 MHz signal at 1 GS/s (dt = 1 ns). Period = 50 ns (2 cycles across 100 ns screen).
	// Buffer size 1024 samples (clamped by minSampleCount).
	dt := 1e-9
	totalSamples := 1024
	displayRange := float64(totalSamples) / control.SincWMultiplier // 204.8
	leftRightRange := (float64(totalSamples) - displayRange) / 2.0  // 409.6
	xOffset := 50e-9
	nPre := uint64(math.Round(xOffset/dt + leftRightRange)) // 460
	xRoundError := xOffset - dt*(float64(nPre)-leftRightRange)

	scp.controlSamplingTimeInterval = dt
	scp.controlXRoundError = xRoundError
	scp.controlTriggerTimeOffset = 0 // Sub-sample jitter from simulator or hardware

	buf := make([]float32, totalSamples)
	// Sine wave crossing 0 at sample nPre (the trigger point)
	for i := range buf {
		timeFromTrig := float64(int(i)-int(nPre)) * dt
		val := math.Sin(2 * math.Pi * 20e6 * timeFromTrig)
		// Scale to mV
		buf[i] = float32(1000 * val)
	}
	scp.displayBuffers = [][]float32{buf}

	scp.channelCount = 1
	sv := newSignalViewer(targetImg, bounds, scp, false)
	unit := 1000.0 / scp.maxScreenTime
	deltaT := dt * unit
	zeroOffset := 250 // middle of height 500

	sv.drawNormal(1000.0, 500.0, bounds, zeroOffset, deltaT)

	// Inspect all pixels at x = 500 to find where the line was drawn
	var coloredYs []int
	for y := 0; y < 500; y++ {
		p := targetImg.RGBAAt(500, y)
		if p.R == 255 && p.A == 255 {
			coloredYs = append(coloredYs, y)
		}
	}
	tAtCursor, instV, _ := sv.calcValuesAt(500, 250, 1000.0, 500.0, bounds)
	t.Logf("Colored Ys at x=500: %v, controlXRoundError: %v, deltaT: %v, tAtCursor: %v, instV: %v", coloredYs, scp.controlXRoundError, deltaT, tAtCursor, instV)

	hasColor := false
	for _, y := range coloredYs {
		if math.Abs(float64(y-250)) <= 2 {
			hasColor = true
			break
		}
	}
	assert.True(t, hasColor, "Signal with SinC interpolation should pass directly through the trigger point at (500, 250)")

	// Also verify that cursor / value calculation at trigger position evaluates to ~0 mV.
	assert.InDelta(t, float32(0.0), instV[0], 50.0, "Calculated voltage at trigger point should be ~0 mV")
}
