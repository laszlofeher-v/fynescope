package gui

import (
	"fynescope/control"
	"fynescope/genericps"
	"fynescope/settings"
	"image/color"
	"testing"

	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"
	"fynescope/checkcolorpick"
	"fynescope/disp7"
	"github.com/stretchr/testify/assert"
)

func TestNewDftPanel(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()

	reqCh := make(chan genericps.Message)
	scp := &ScpDesc{
		App:   app,
		theme: Theme(settings.DarkTheme),
		Settings: &settings.PsSettings{
			Theme: settings.DarkTheme,
			Dft: settings.DftSettings{
				Window:          settings.WindowRectangular,
				DisplayMode:     settings.ModeVoltage,
				Bins:            1024,
				ArbitraryDbRefV: 1.0,
				MinFreq:         0,
				MaxFreq:         10000,
			},
			Channels: []settings.ChSettings{
				{Enabled: true, Col: [2]color.NRGBA{{R: 255, G: 255, B: 255, A: 255}, {R: 255, G: 255, B: 255, A: 255}}},
				{Enabled: false, Col: [2]color.NRGBA{{R: 255, G: 255, B: 255, A: 255}, {R: 255, G: 255, B: 255, A: 255}}},
			},
		},
		channelCount:   2,
		channelViewers: []channelViewerDesc{
			{
				enableCheckbox:  &checkcolorpick.CheckColorPick{},
				triggerCheckbox: &widget.Check{},
				minV:            &disp7.DigitArray{},
				maxV:            &disp7.DigitArray{},
				offset:          &disp7.DigitArray{},
				frq:             &disp7.DigitArray{},
				period:          &disp7.DigitArray{},
			},
			{
				enableCheckbox:  &checkcolorpick.CheckColorPick{},
				triggerCheckbox: &widget.Check{},
				minV:            &disp7.DigitArray{},
				maxV:            &disp7.DigitArray{},
				offset:          &disp7.DigitArray{},
				frq:             &disp7.DigitArray{},
				period:          &disp7.DigitArray{},
			},
		},
		psControl: &control.PscDesc{
			Con: &genericps.Connection{
				MsgCh: reqCh,
			},
		},
	}

	// Mock the scope control so ChannelRanges doesn't block forever
	go func() {
		for {
			msg, ok := <-reqCh
			if !ok {
				break
			}
			if getInfoMsg, ok := msg.(*genericps.GetChannelInformationMsg); ok {
				rsp := getInfoMsg.Rsp().(*genericps.GetChannelInformationRsp)
				rsp.LengthOfRanges = 0
			}

			ch := msg.RspCh()
			if ch != nil {
				close(ch)
			}
		}
	}()
	defer close(reqCh)

	layout := container.NewVBox()

	// Call the function under test
	scp.newDftPanel(layout)

	// Verify that key UI components have been initialized
	assert.NotNil(t, scp.dftMinFreqDisp, "dftMinFreqDisp should be initialized")
	assert.NotNil(t, scp.dftMaxFreqDisp, "dftMaxFreqDisp should be initialized")
	assert.NotNil(t, scp.binWidthLabel, "binWidthLabel should be initialized")
	assert.NotNil(t, scp.dftDataCollectionTimeLabel, "dftDataCollectionTimeLabel should be initialized")

	for i := 0; i < int(scp.channelCount); i++ {
		assert.NotNil(t, scp.channelViewers[i].dftCheckbox, "dftCheckbox should be initialized for channel %d", i)
		assert.NotNil(t, scp.channelViewers[i].dftPersistenceCheckbox, "dftPersistenceCheckbox should be initialized for channel %d", i)
		assert.NotNil(t, scp.channelViewers[i].dftNameLabel, "dftNameLabel should be initialized for channel %d", i)
		// Because there can be exactly one or more selects created
		assert.Greater(t, len(scp.channelViewers[i].vRangeSelects), 0, "vRangeSelects should be populated")
		assert.Greater(t, len(scp.channelViewers[i].x10Checkboxes), 0, "x10Checkboxes should be populated")
	}
}
