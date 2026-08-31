package gui

import (
	"image"
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
