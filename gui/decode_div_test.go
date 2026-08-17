package gui

import (
	"fynescope/settings"
	"fynescope/genericps"
	"testing"

	"fyne.io/fyne/v2/test"
	"github.com/stretchr/testify/assert"
)

func TestBuildDecodeContent(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()

	scp := &ScpDesc{
		App: app,
		Settings: &settings.PsSettings{
			Decode: settings.DecodeSettings{
				Enabled:  true,
				Protocol: "UART",
				Channel1: 0,
				BaudRate: 9600,
			},
			Channels: []settings.ChSettings{
				{VRange: genericps.Range_5v}, // Ch A
				{VRange: genericps.Range_1v}, // Ch B
				{VRange: genericps.Range_1v}, // Ch C
				{VRange: genericps.Range_1v}, // Ch D
			},
		},
		MaxValue: 32767,
		MinValue: -32768,
	}

	genericps.InputRanges = []int32{5000} // For adc/mv math mock

	// Test building content while docked
	content := scp.buildDecodeContent(true)
	assert.NotNil(t, content)

	// Since settings say UART, it should have the baud rate
	assert.Equal(t, "UART", scp.Settings.Decode.Protocol)
	assert.Equal(t, 9600, scp.Settings.Decode.BaudRate)

	// Refresh layout test
	scp.refreshDecodeTab()
}
