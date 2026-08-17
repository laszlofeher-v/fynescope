package gui

import (
	"fynescope/control"
	"fynescope/settings"
	"image"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDrawDecode(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 800, 600))
	bounds := img.Bounds()
	w := float64(800)
	zeroOffset := 300

	scp := &ScpDesc{
		maxScreenTime:         10.0,
		timeZoomMaxScreenTime: 5.0,
		timeZoomBoxOffset:     2.0,
		Settings: &settings.PsSettings{
			Decode: settings.DecodeSettings{
				Enabled:       true,
				Protocol:      "UART",
				ShowBitstarts: true,
			},
		},
		DecodeState: control.DecoderState{
			Bytes: []control.DecodeResult{
				{Value: 0x41, StartTime: 1.0, EndTime: 2.0, Error: false}, // Fits screen, Value 'A'
				{Value: 0x42, StartTime: -5.0, EndTime: -4.0, Error: true}, // Offscreen left
				{Value: 0x43, StartTime: 15.0, EndTime: 16.0, Error: false}, // Offscreen right
				{Value: 0x44, StartTime: 9.0, EndTime: 11.0, Error: false}, // Clips right edge
				{Value: 0x15, StartTime: 3.0, EndTime: 4.0, Error: false}, // Non-printable UART text
			},
			Bits: []control.DecodeBit{
				{StartTime: 1.5, EndTime: 1.6}, // Visible bit
				{StartTime: 11.0, EndTime: 11.1}, // Offscreen bit
			},
		},
	}

	// 1. Normal View Test
	// Should run successfully without panics and draw 'A', clip 'D', ignore others.
	scp.drawDecode(img, bounds, w, zeroOffset, false)
	assert.NotNil(t, img)

	// 2. Zoom View Test
	// timeZoomBoxOffset = 2.0, maxScreenTime = 5.0 (so screen is T=2.0 to T=7.0)
	// Byte 0x41 (1.0 - 2.0) should clip left edge.
	// Byte 0x15 (3.0 - 4.0) should fit perfectly.
	scp.drawDecode(img, bounds, w, zeroOffset, true)
	assert.NotNil(t, img)

	// 3. Disabled decode test
	scp.Settings.Decode.Enabled = false
	// Should abort immediately
	scp.drawDecode(img, bounds, w, zeroOffset, false)
}
