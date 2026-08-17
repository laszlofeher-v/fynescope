package gui

import (
	"fynescope/genericps"
	"testing"

	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"
	"github.com/stretchr/testify/assert"
)

func TestMergeVoiceCommands(t *testing.T) {
	dest := &VoiceCommands{
		RunCommands: []string{"run"},
	}
	src := &VoiceCommands{
		RunCommands: []string{"start"},
		ChannelA:    []string{"ch a"},
	}
	mergeVoiceCommands(dest, src)
	assert.Contains(t, dest.RunCommands, "run")
	assert.Contains(t, dest.RunCommands, "start")
	assert.Contains(t, dest.ChannelA, "ch a")
}

func TestExecuteVoiceCommand(t *testing.T) {
	test.NewApp()

	// Backup original and set deterministic commands for testing
	originalCmds := ActiveVoiceCommands
	defer func() {
		ActiveVoiceCommands = originalCmds
	}()

	ActiveVoiceCommands = VoiceCommands{
		RunCommands:     []string{"run"},
		StopCommands:    []string{"stop"},
		EnableCommands:  []string{"enable"},
		DisableCommands: []string{"disable"},
		TriggerCommands: []string{"trigger"},
		InvCommands:     []string{"invert"},
		AcCommands:      []string{"ac"},
		DcCommands:      []string{"dc"},
		RisingCommands:  []string{"rising"},
		FallingCommands: []string{"falling"},
		ChannelA:        []string{"channel a"},
	}

	scp := &ScpDesc{
		channelCount: genericps.QuadScope,
	}

	var runblockClicked bool
	scp.runblockButton = widget.NewButton("Run/Stop", func() {
		runblockClicked = true
	})

	scp.streamEnableButton = widget.NewButton("Stream", func() {})

	scp.channelViewers = make([]channelViewerDesc, scp.channelCount)
	for i := range scp.channelViewers {
		scp.channelViewers[i] = channelViewerDesc{}
	}

	var invertToggled bool
	scp.channelViewers[genericps.ChA].invertCheckbox = widget.NewCheck("Invert", func(b bool) {
		invertToggled = b
	})

	// Test Run
	scp.ExecuteVoiceCommand("run")
	assert.True(t, runblockClicked)

	// Test Stop
	runblockClicked = false
	scp.ExecuteVoiceCommand("stop")
	assert.True(t, runblockClicked)

	// Test Channel A Invert Toggle
	scp.ExecuteVoiceCommand("invert channel a")
	assert.True(t, invertToggled)

	// Test Channel A Invert Enable
	scp.channelViewers[genericps.ChA].invertCheckbox.SetChecked(false) // reset state
	invertToggled = false
	scp.ExecuteVoiceCommand("enable invert channel a")
	assert.True(t, invertToggled)

	// Test Channel A Invert Disable
	scp.ExecuteVoiceCommand("disable invert channel a")
	assert.False(t, invertToggled)
}
