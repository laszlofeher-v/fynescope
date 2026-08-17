package gui

import (
	"image"
	"image/color"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/image/math/fixed"
)

func TestAbs(t *testing.T) {
	assert.Equal(t, 5, abs(5))
	assert.Equal(t, 5, abs(-5))
	assert.Equal(t, 0, abs(0))
}

func TestI26_6Conversions(t *testing.T) {
	// fixed.Int26_6 represents value * 64
	val := fixed.I(10) // 10 * 64

	assert.Equal(t, 10.0, i26_6ToFloat64(val))
	assert.Equal(t, float32(10.0), i26_6ToFloat32(val))
}

func TestDrawLine(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 10, 10))
	c := color.RGBA{R: 255, A: 255}

	err := drawLine(img, 1, 1, 5, 5, c)
	assert.NoError(t, err)

	// Check if starting point is colored
	pixelColor := img.At(1, 1).(color.RGBA)
	assert.Equal(t, c, pixelColor)
	
	// Check if end point is colored
	pixelColor = img.At(5, 5).(color.RGBA)
	assert.Equal(t, c, pixelColor)
}

func TestDrawCircle(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 20, 20))
	c := color.RGBA{G: 255, A: 255}

	drawCircle(img, 10, 10, 5, c)

	// Center should NOT be colored (just the outline is drawn)
	centerColor := img.At(10, 10).(color.RGBA)
	assert.NotEqual(t, c, centerColor)

	// A point on the radius should be colored (algorithm draws at r-1)
	edgeColor := img.At(14, 10).(color.RGBA)
	assert.Equal(t, c, edgeColor)
}
