package gui

import (
	"image"
	"image/color"
	"testing"

	"fynescope/settings"

	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/theme"
	"github.com/stretchr/testify/assert"
)

func TestDigitalRaster_Generation(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()
	win := app.NewWindow("Test")

	scp := &ScpDesc{
		Settings: &settings.PsSettings{
			Digital: settings.DigitalSettings{
				Ports: [2]settings.DigitalPortSettings{
					{Enabled: true, Threshold: 0},
					{Enabled: false, Threshold: 0},
				},
			},
			Time: settings.TimeSettings{
				TriggerTimeOffset: 0.0,
			},
		},
		maxScreenTime:               1.0,
		controlSamplingTimeInterval: 0.001,
		theme:                       theme.DefaultTheme(),
	}
	scp.ftScopeFullScreen = image.NewRGBA(image.Rect(0, 0, 800, 400))
	scp.ftScopeSignalScreen = scp.ftScopeFullScreen.(*image.RGBA).SubImage(image.Rect(50, 10, 750, 350)).(rasterImage)
	scp.setFtHDivsX()

	c, dr := scp.newDigitalRaster(win)
	assert.NotNil(t, c)
	assert.NotNil(t, dr)
	assert.NotNil(t, dr.raster)

	// Test generation with w=800, h=200 with high bit data
	scp.digitalDisplayBuffer = [][]uint8{
		make([]uint8, 1000),
		make([]uint8, 0),
	}
	for i := range scp.digitalDisplayBuffer[0] {
		scp.digitalDisplayBuffer[0][i] = 0xFF // all bits high
	}

	for i := 0; i < 8; i++ {
		scp.Settings.Digital.ChannelsEnabled[i] = true
	}

	img := dr.generate(800, 200)
	assert.NotNil(t, img)
	assert.Equal(t, 800, img.Bounds().Dx())
	assert.Equal(t, 200, img.Bounds().Dy())

	// Test trigger offset shift
	scp.Settings.Time.TriggerTimeOffset = 0.2
	scp.setFtHDivsX()
	imgShifted := dr.generate(800, 200)
	assert.NotNil(t, imgShifted)
	assert.Equal(t, 800, imgShifted.Bounds().Dx())
}

func TestDigitalRaster_NoEmptyRowsForDisabledChannels(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()
	win := app.NewWindow("Test")

	scp := &ScpDesc{
		Settings: &settings.PsSettings{
			Digital: settings.DigitalSettings{
				Ports: [2]settings.DigitalPortSettings{
					{Enabled: true, Threshold: 0},
					{Enabled: false, Threshold: 0},
				},
				ChannelColors: [16]color.NRGBA{
					{R: 255, G: 0, B: 0, A: 255}, // D0 - Red
					{R: 0, G: 255, B: 0, A: 255}, // D1 - Green
					{R: 0, G: 0, B: 255, A: 255}, // D2 - Blue
				},
			},
			Time: settings.TimeSettings{
				TriggerTimeOffset: 0.0,
			},
		},
		maxScreenTime:               1.0,
		controlSamplingTimeInterval: 0.001,
		theme:                       theme.DefaultTheme(),
	}
	// Enable only D0 and D2. D1 and D3-D7 are disabled.
	scp.Settings.Digital.ChannelsEnabled[0] = true
	scp.Settings.Digital.ChannelsEnabled[1] = false
	scp.Settings.Digital.ChannelsEnabled[2] = true

	scp.ftScopeFullScreen = image.NewRGBA(image.Rect(0, 0, 800, 200))
	scp.ftScopeSignalScreen = scp.ftScopeFullScreen.(*image.RGBA).SubImage(image.Rect(50, 0, 750, 200)).(rasterImage)
	scp.setFtHDivsX()

	_, dr := scp.newDigitalRaster(win)

	scp.digitalDisplayBuffer = [][]uint8{
		make([]uint8, 1000),
		make([]uint8, 0),
	}
	// D0=0 (low), D2=1 (high)
	for i := range scp.digitalDisplayBuffer[0] {
		scp.digitalDisplayBuffer[0][i] = 0x04 // bit 2 is high, bit 0 is low
	}

	h := 200
	w := 800
	img := dr.generate(w, h).(*image.RGBA)

	// Since only 2 channels are enabled, activeChannels = 2, channelHeight = 100.
	// Row 0 (D0): yBase=0, yHigh=20, yLow=80. D0 is low, so signal line at y=80 with Red.
	// Row 1 (D2): yBase=100, yHigh=120, yLow=180. D2 is high, so signal line at y=120 with Blue.
	// If empty rows were present (e.g. h/8 = 25), D2 would be at yBase = 2 * 25 = 50.
	// Check that at x=400, y=80 is Red (D0 low line)
	redPixel := img.RGBAAt(400, 80)
	assert.Equal(t, uint8(255), redPixel.R, "D0 low signal should be at y=80 (row height 100)")

	// Check that at x=400, y=120 is Blue (D2 high line)
	bluePixel := img.RGBAAt(400, 120)
	assert.Equal(t, uint8(255), bluePixel.B, "D2 high signal should be at y=120 (row height 100)")
}
