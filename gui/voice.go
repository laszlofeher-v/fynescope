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

	// Helper to check for multiple keywords
	containsAny := func(s string, subs ...string) bool {
		for _, sub := range subs {
			if strings.Contains(s, sub) {
				return true
			}
		}
		return false
	}

	// Ensure execution happens on the main UI thread, as many
	// of these commands will update Fyne widgets (checkboxes, buttons).
	fyne.Do(func() {
		// Run / Stop commands
		if containsAny(cmd, "start", "run", "iniciar", "arrancar", "démarrer", "starten", "indulás", "indítás", "futás") {
			// Try to start if stopped
			if scp.streamEnableButton != nil && !scp.streamEnableButton.Disabled() {
				scp.runblockButton.Tapped(&fyne.PointEvent{})
			}
		} else if containsAny(cmd, "stop", "halt", "detener", "parar", "arrêter", "stoppen", "állj", "leállítás", "megállítás") {
			// Trigger the stop action if running
			if scp.runblockButton != nil && !scp.runblockButton.Disabled() {
				scp.runblockButton.Tapped(&fyne.PointEvent{})
			}
		}

		// Channel Enable/Disable commands
		enable := containsAny(cmd, "enable", "turn on", "show", "habilitar", "encender", "activer", "allumer", "afficher", "aktivieren", "einschalten", "zeigen", "engedélyezés", "bekapcsolás", "mutat")
		disable := containsAny(cmd, "disable", "turn off", "hide", "deshabilitar", "apagar", "désactiver", "éteindre", "masquer", "deaktivieren", "ausschalten", "verstecken", "tiltás", "kikapcsolás", "elrejt")

		if enable || disable {
			// Find which channel is mentioned
			var ch genericps.ChannelId = -1
			if containsAny(cmd, "channel a", "ch a", "canal a", "voie a", "kanal a", "a csatorna", "csatorna a") {
				ch = genericps.ChA
			} else if containsAny(cmd, "channel b", "ch b", "canal b", "voie b", "kanal b", "b csatorna", "csatorna b") {
				ch = genericps.ChB
			} else if containsAny(cmd, "channel c", "ch c", "canal c", "voie c", "kanal c", "c csatorna", "csatorna c") {
				ch = genericps.ChC
			} else if containsAny(cmd, "channel d", "ch d", "canal d", "voie d", "kanal d", "d csatorna", "csatorna d") {
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
