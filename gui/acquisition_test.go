package gui

import (
	"fynescope/control"
	"testing"

	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"
	"github.com/stretchr/testify/assert"
)

func TestInStreamMode(t *testing.T) {
	scp := &ScpDesc{}

	// When psControl is nil
	assert.False(t, scp.inStreamMode())

	scp.psControl = &control.PscDesc{}

	// When below threshold
	scp.maxScreenTime = control.StreamThreshold - 1
	scp.psControl.StreamEnabled.Store(true)
	assert.False(t, scp.inStreamMode())

	// When above threshold but stream disabled
	scp.maxScreenTime = control.StreamThreshold
	scp.psControl.StreamEnabled.Store(false)
	assert.False(t, scp.inStreamMode())

	// When above threshold and stream enabled
	scp.psControl.StreamEnabled.Store(true)
	assert.True(t, scp.inStreamMode())
}

func TestUpdateStreamButtonState(t *testing.T) {
	test.NewApp()
	scp := &ScpDesc{}

	// When psControl or button is nil, it should not panic
	scp.updateStreamButtonState()

	scp.psControl = &control.PscDesc{}
	scp.streamEnableButton = widget.NewButton("", nil)

	scp.psControl.StreamEnabled.Store(true)
	scp.updateStreamButtonState()
	assert.Equal(t, streamEnabledLabel, scp.streamEnableButton.Text)

	scp.psControl.StreamEnabled.Store(false)
	scp.updateStreamButtonState()
	assert.Equal(t, streamDisabledLabel, scp.streamEnableButton.Text)
}
