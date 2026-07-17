package gui

import (
	"fynescope/genericps"
	"log/slog"
	"strings"

	"fyne.io/fyne/v2"
)

// ExecuteVoiceCommand processes natural language text commands
// and executes the corresponding UI/backend logic.
func (scp *ScpDesc) ExecuteVoiceCommand(cmd string) {
	cmd = strings.ToLower(strings.TrimSpace(cmd))
	slog.Debug("ExecuteVoiceCommand parsing", "cmd", cmd)

	// Ensure execution happens on the main UI thread, as many
	// of these commands will update Fyne widgets (checkboxes, buttons).
	fyne.Do(func() {
		// Run / Stop commands
		if strings.Contains(cmd, "start") || strings.Contains(cmd, "run") {
			// Try to start if stopped
			if scp.streamEnableButton != nil && !scp.streamEnableButton.Disabled() {
				scp.runblockButton.Tapped(&fyne.PointEvent{})
			}
		} else if strings.Contains(cmd, "stop") || strings.Contains(cmd, "halt") {
			// Trigger the stop action if running
			if scp.runblockButton != nil && !scp.runblockButton.Disabled() {
				scp.runblockButton.Tapped(&fyne.PointEvent{})
			}
		}

		// Channel Enable/Disable commands
		enable := false
		disable := false
		if strings.Contains(cmd, "enable") || strings.Contains(cmd, "turn on") || strings.Contains(cmd, "show") {
			enable = true
		} else if strings.Contains(cmd, "disable") || strings.Contains(cmd, "turn off") || strings.Contains(cmd, "hide") {
			disable = true
		}

		if enable || disable {
			// Find which channel is mentioned
			var ch genericps.ChannelId = -1
			if strings.Contains(cmd, "channel a") || strings.Contains(cmd, "ch a") {
				ch = genericps.ChA
			} else if strings.Contains(cmd, "channel b") || strings.Contains(cmd, "ch b") {
				ch = genericps.ChB
			} else if strings.Contains(cmd, "channel c") || strings.Contains(cmd, "ch c") {
				ch = genericps.ChC
			} else if strings.Contains(cmd, "channel d") || strings.Contains(cmd, "ch d") {
				ch = genericps.ChD
			}

			if ch != -1 && int(ch) < int(scp.channelCount) {
				// We update the UI checkbox directly, which triggers the logic
				if scp.channelViewers[ch].enableCheckbox != nil {
					if scp.channelViewers[ch].enableCheckbox.Val != enable {
						scp.channelViewers[ch].enableCheckbox.Tapped(&fyne.PointEvent{})
					}
				} else {
					scp.EnableChannel(ch, enable)
				}
			}
		}
	})
}
