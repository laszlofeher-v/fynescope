package gui

import (
	"fynescope/selectscroll"
	"fynescope/settings"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"github.com/stretchr/testify/assert"
)

func TestNewRlcPanel(t *testing.T) {
	scp := &ScpDesc{
		channelCount: 2,
		Settings: &settings.PsSettings{
			Channels: make([]settings.ChSettings, 2),
		},
	}
	scp.channelViewers = make([]channelViewerDesc, 2)
	scp.channelViewers[0] = channelViewerDesc{}
	scp.channelViewers[1] = channelViewerDesc{}

	// Create a dummy container
	panel := container.NewVBox()

	// Call newRlcPanel
	scp.newRlcPanel(panel)

	// Panel should have 1 child (the VBox created inside newRlcPanel)
	assert.Len(t, panel.Objects, 1)

	// Ensure the rlcNameLabel is set for both channels
	assert.NotNil(t, scp.channelViewers[0].rlcNameLabel)
	assert.NotNil(t, scp.channelViewers[1].rlcNameLabel)

	assert.Equal(t, "Ch A:", scp.channelViewers[0].rlcNameLabel.Text)
	assert.Equal(t, "Ch B:", scp.channelViewers[1].rlcNameLabel.Text)

	// Ensure Q defaults to 2.5
	assert.Equal(t, 2.5, scp.Settings.Channels[0].RlcFilter.Q)
	assert.Equal(t, 2.5, scp.Settings.Channels[1].RlcFilter.Q)

	// Ensure rlcQ controls are registered for test proxy
	controlsMtx.Lock()
	_, hasQA := controls[rlcQId+"A"]
	_, hasQB := controls[rlcQId+"B"]
	controlsMtx.Unlock()
	assert.True(t, hasQA, "expected rlcQA control registered")
	assert.True(t, hasQB, "expected rlcQB control registered")

	// Verify visibility: child 0 is title, child 1 is channel A controls container
	vbox := panel.Objects[0].(*fyne.Container)
	chAControls := vbox.Objects[1].(*fyne.Container)
	// chAControls objects: [0]=row1, [1]=typeSelect, [2]=rBox, [3]=lBox, [4]=cBox, [5]=qBox
	assert.Len(t, chAControls.Objects, 6)
	typeSelect := chAControls.Objects[1].(*selectscroll.SelectScroll)
	rBox := chAControls.Objects[2].(*fyne.Container)
	lBox := chAControls.Objects[3].(*fyne.Container)
	cBox := chAControls.Objects[4].(*fyne.Container)
	qBox := chAControls.Objects[5].(*fyne.Container)

	// In Lowpass RC (default): R and C visible, L and Q hidden
	assert.True(t, rBox.Visible())
	assert.False(t, lBox.Visible())
	assert.True(t, cBox.Visible())
	assert.False(t, qBox.Visible())

	// Switch to Lowpass LC: L, C, and Q visible, R hidden
	typeSelect.SetSelected("Lowpass LC")
	assert.False(t, rBox.Visible())
	assert.True(t, lBox.Visible())
	assert.True(t, cBox.Visible())
	assert.True(t, qBox.Visible())

	// Switch to Highpass LC: L, C, and Q visible, R hidden
	typeSelect.SetSelected("Highpass LC")
	assert.False(t, rBox.Visible())
	assert.True(t, lBox.Visible())
	assert.True(t, cBox.Visible())
	assert.True(t, qBox.Visible())

	// Switch to Disabled: all hidden
	typeSelect.SetSelected(settings.RlcFilterTypeDisabled)
	assert.False(t, rBox.Visible())
	assert.False(t, lBox.Visible())
	assert.False(t, cBox.Visible())
	assert.False(t, qBox.Visible())
}
