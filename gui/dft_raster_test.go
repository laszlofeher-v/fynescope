package gui

import (
	"fynescope/settings"
	"image"
	"image/draw"
	"math"
	"testing"

	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/theme"
	"github.com/stretchr/testify/assert"
)

func TestNiceStep(t *testing.T) {
	tests := []struct {
		span     float64
		expected float64
	}{
		{0, 1},
		{-10, 1},
		{1, 1},
		{1.2, 1},
		{2.5, 2},
		{5.0, 5},
		{8.0, 10},
		{12.0, 10},
		{25.0, 20},
		{50.0, 50},
		{80.0, 100},
		{100, 100},
		{150, 200},
		{300, 200},
		{600, 500},
		{0.12, 0.1},
		{0.25, 0.2},
		{0.6, 0.5},
	}

	for _, tc := range tests {
		actual := niceStep(tc.span)
		assert.InDelta(t, tc.expected, actual, 0.0001, "niceStep(%f)", tc.span)
	}
}

func TestFrqLabelViewer_MousIn(t *testing.T) {
	rect := image.Rect(10, 20, 100, 50)
	frql := &frqLabelViewer{
		rasterPartition: rasterPartition{
			imgRect: rect,
		},
	}

	assert.True(t, frql.mousIn(15, 25), "Point inside should be true")
	assert.True(t, frql.mousIn(10, 20), "Point on top-left border should be true")

	// Max bounds are exclusive in image.Rect
	assert.False(t, frql.mousIn(100, 49), "Point on right border should be false")
	assert.False(t, frql.mousIn(99, 50), "Point on bottom border should be false")
	assert.False(t, frql.mousIn(5, 25), "Point outside left should be false")
	assert.False(t, frql.mousIn(15, 10), "Point outside top should be false")
}

func TestFrqLabelViewer_MouseEvents(t *testing.T) {
	rect := image.Rect(10, 20, 100, 50)
	frql := &frqLabelViewer{
		rasterPartition: rasterPartition{
			imgRect: rect,
		},
	}

	// Initial state
	assert.False(t, frql.selected)

	// Mouse down inside rect
	frql.mouseDown(0, 0, 15, 25)
	assert.True(t, frql.selected)

	// Mouse up
	frql.mouseUp(0, 0, 15, 25)
	assert.False(t, frql.selected)

	// Mouse down outside rect
	frql.mouseDown(0, 0, 5, 25)
	assert.False(t, frql.selected)
}

func TestFormatAxisFreq(t *testing.T) {
	tests := []struct {
		freq float64
		step float64
		want string
	}{
		{0, 1e5, "0"},
		{100e3, 1e5, "100k"},
		{1e6, 1e5, "1M"},
		{1.2e6, 1e5, "1.2M"},
		{10e6, 20e3, "10M"},
		{10.02e6, 20e3, "10.02M"},
		{10.04e6, 20e3, "10.04M"},
		{1.005e6, 5e3, "1.005M"},
		{1.25e6, 250e3, "1.25M"},
		{999.9, 0.2, "999.9"},
		{999.9, 1.0, "1k"},
	}

	for _, tc := range tests {
		got := formatAxisFreq(tc.freq, tc.step)
		assert.Equal(t, tc.want, got, "formatAxisFreq(%g, %g)", tc.freq, tc.step)
	}
}

func TestLinearAxisStep_AlignsWithGridDivisions(t *testing.T) {
	cases := []struct {
		span    float64
		numDivs int
	}{
		{200e3, 4},
		{200e3, 5},
		{500e3, 3},
		{1e6, 6},
		{1e6, 10},
		{25e3, 7},
	}

	for _, tc := range cases {
		gridStep := niceStep(tc.span / 10.0)
		labelStep := linearAxisStep(tc.span, tc.numDivs)
		assert.LessOrEqual(t, tc.span/labelStep, float64(tc.numDivs)+1e-9, "span=%g numDivs=%d", tc.span, tc.numDivs)
		ratio := labelStep / gridStep
		assert.InDelta(t, math.Round(ratio), ratio, 1e-9, "labelStep=%g must be an integer multiple of gridStep=%g", labelStep, gridStep)
	}
}

func TestFrqLabelViewer_RedrawsWhenFreqRangeOrLogChangesEvenWithoutRefreshFlag(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()

	fullImg := image.NewRGBA(image.Rect(0, 0, 800, 400))
	dftTab := container.NewTabItem("FFT", container.NewVBox())

	scp := &ScpDesc{
		Settings: &settings.PsSettings{
			Dft: settings.DftSettings{
				MinFreq:  0,
				MaxFreq:  500e3,
				XAxisLog: false,
			},
		},
		controlSamplingTimeInterval: 1e-6, // 1 MS/s -> maxFreqAvailable = 500 kHz
		theme:                       theme.DefaultTheme(),
		dftScopeFullScreen:          fullImg,
		dftTab:                      dftTab,
		controlTab:                  container.NewAppTabs(dftTab),
	}
	scp.dftScopeSignalScreen = fullImg.SubImage(image.Rect(50, 10, 750, 350)).(draw.RGBA64Image)
	frql := newFrqLabelViewer(fullImg, image.Rect(0, 358, 800, 400), scp)

	// First draw renders initial 0..500k range and clears refreshFlag.
	frql.draw()
	assert.False(t, frql.refreshFlag)
	assert.True(t, frql.hasLastDraw)
	assert.Equal(t, 0.0, frql.lastMinFreq)
	assert.Equal(t, 500e3, frql.lastMaxFreqPlot)
	assert.False(t, frql.lastXAxisLog)

	// Simulate hardware RefreshCallback changing controlSamplingTimeInterval WITHOUT calling enableRefresh():
	// Nyquist shrinks from 500 kHz to 100 kHz -> frql must detect the change and redraw.
	scp.controlSamplingTimeInterval = 5e-6 // 200 kS/s -> maxFreqAvailable = 100 kHz
	frql.draw()
	assert.InDelta(t, 100e3, frql.lastMaxFreqPlot, 1e-6, "expected labels to update when controlSamplingTimeInterval changed")

	// Simulate Log X toggle WITHOUT calling enableRefresh():
	scp.Settings.Dft.XAxisLog = true
	frql.draw()
	assert.True(t, frql.lastXAxisLog, "expected labels to update when XAxisLog changed")

	// Simulate MinFreq/MaxFreq zoom WITHOUT calling enableRefresh():
	scp.Settings.Dft.XAxisLog = false
	scp.Settings.Dft.MinFreq = 20e3
	scp.Settings.Dft.MaxFreq = 80e3
	frql.draw()
	assert.Equal(t, 20e3, frql.lastMinFreq)
	assert.Equal(t, 60e3, frql.lastMaxFreqPlot)
}
