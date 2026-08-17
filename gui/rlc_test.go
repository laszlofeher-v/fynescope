package gui

import (
	"fynescope/settings"
	"testing"

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
}
