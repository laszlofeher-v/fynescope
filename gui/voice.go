package gui

import (
	"fynescope/genericps"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2"
	"gopkg.in/yaml.v3"
)

type VoiceCommands struct {
	RunCommands     []string `yaml:"run_commands"`
	StopCommands    []string `yaml:"stop_commands"`
	EnableCommands  []string `yaml:"enable_commands"`
	DisableCommands []string `yaml:"disable_commands"`
	TriggerCommands []string `yaml:"trigger_commands"`
	X10Commands     []string `yaml:"x10_commands"`
	InvCommands     []string `yaml:"inv_commands"`
	AcCommands      []string `yaml:"ac_commands"`
	DcCommands      []string `yaml:"dc_commands"`
	RisingCommands  []string `yaml:"rising_commands"`
	FallingCommands []string `yaml:"falling_commands"`
	ChannelA        []string `yaml:"channel_a"`
	ChannelB        []string `yaml:"channel_b"`
	ChannelC        []string `yaml:"channel_c"`
	ChannelD        []string `yaml:"channel_d"`
}

var ActiveVoiceCommands VoiceCommands

func init() {
	InitVoiceCommands()
}

func InitVoiceCommands() {
	dir := "voice_commands"
	if err := os.MkdirAll(dir, 0755); err != nil {
		slog.Error("failed to create voice_commands directory", "err", err)
		return
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		slog.Error("failed to read voice_commands directory", "err", err)
		return
	}

	var hasYaml bool
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".yaml" {
			hasYaml = true
			break
		}
	}

	if !hasYaml {
		writeDefaultVoiceCommands(dir)
		entries, _ = os.ReadDir(dir)
	}

	// load and merge all
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".yaml" {
			path := filepath.Join(dir, entry.Name())
			data, err := os.ReadFile(path)
			if err != nil {
				slog.Error("failed to read voice command file", "file", path, "err", err)
				continue
			}

			var cmds VoiceCommands
			if err := yaml.Unmarshal(data, &cmds); err != nil {
				slog.Error("failed to parse voice command file", "file", path, "err", err)
				continue
			}

			mergeVoiceCommands(&ActiveVoiceCommands, &cmds)
		}
	}
}

func mergeVoiceCommands(dest, src *VoiceCommands) {
	dest.RunCommands = append(dest.RunCommands, src.RunCommands...)
	dest.StopCommands = append(dest.StopCommands, src.StopCommands...)
	dest.EnableCommands = append(dest.EnableCommands, src.EnableCommands...)
	dest.DisableCommands = append(dest.DisableCommands, src.DisableCommands...)
	dest.TriggerCommands = append(dest.TriggerCommands, src.TriggerCommands...)
	dest.X10Commands = append(dest.X10Commands, src.X10Commands...)
	dest.InvCommands = append(dest.InvCommands, src.InvCommands...)
	dest.AcCommands = append(dest.AcCommands, src.AcCommands...)
	dest.DcCommands = append(dest.DcCommands, src.DcCommands...)
	dest.RisingCommands = append(dest.RisingCommands, src.RisingCommands...)
	dest.FallingCommands = append(dest.FallingCommands, src.FallingCommands...)
	dest.ChannelA = append(dest.ChannelA, src.ChannelA...)
	dest.ChannelB = append(dest.ChannelB, src.ChannelB...)
	dest.ChannelC = append(dest.ChannelC, src.ChannelC...)
	dest.ChannelD = append(dest.ChannelD, src.ChannelD...)
}

func writeDefaultVoiceCommands(dir string) {
	en := VoiceCommands{
		RunCommands:     []string{"start", "run"},
		StopCommands:    []string{"stop", "halt"},
		EnableCommands:  []string{"enable", "turn on", "show"},
		DisableCommands: []string{"disable", "turn off", "hide"},
		TriggerCommands: []string{"trigger", "set trigger"},
		X10Commands:     []string{"x10", "times 10", "times ten"},
		InvCommands:     []string{"invert", "inv"},
		AcCommands:      []string{"ac"},
		DcCommands:      []string{"dc"},
		RisingCommands:  []string{"rising", "rise"},
		FallingCommands: []string{"falling", "fall"},
		ChannelA:        []string{"channel a", "ch a"},
		ChannelB:        []string{"channel b", "ch b"},
		ChannelC:        []string{"channel c", "ch c"},
		ChannelD:        []string{"channel d", "ch d"},
	}
	writeYaml(filepath.Join(dir, "en.yaml"), en)

	es := VoiceCommands{
		RunCommands:     []string{"iniciar", "arrancar"},
		StopCommands:    []string{"detener", "parar"},
		EnableCommands:  []string{"habilitar", "encender"},
		DisableCommands: []string{"deshabilitar", "apagar"},
		TriggerCommands: []string{"disparador", "gatillo", "trigger"},
		X10Commands:     []string{"x10", "por diez"},
		InvCommands:     []string{"invertir", "inv"},
		AcCommands:      []string{"ac"},
		DcCommands:      []string{"dc"},
		RisingCommands:  []string{"ascendente", "subida"},
		FallingCommands: []string{"descendente", "bajada"},
		ChannelA:        []string{"canal a"},
		ChannelB:        []string{"canal b"},
		ChannelC:        []string{"canal c"},
		ChannelD:        []string{"canal d"},
	}
	writeYaml(filepath.Join(dir, "es.yaml"), es)

	fr := VoiceCommands{
		RunCommands:     []string{"démarrer"},
		StopCommands:    []string{"arrêter"},
		EnableCommands:  []string{"activer", "allumer", "afficher"},
		DisableCommands: []string{"désactiver", "éteindre", "masquer"},
		TriggerCommands: []string{"déclencheur", "trigger"},
		X10Commands:     []string{"x10", "fois dix"},
		InvCommands:     []string{"inverser", "inv"},
		AcCommands:      []string{"ac"},
		DcCommands:      []string{"dc"},
		RisingCommands:  []string{"montant"},
		FallingCommands: []string{"descendant"},
		ChannelA:        []string{"voie a"},
		ChannelB:        []string{"voie b"},
		ChannelC:        []string{"voie c"},
		ChannelD:        []string{"voie d"},
	}
	writeYaml(filepath.Join(dir, "fr.yaml"), fr)

	de := VoiceCommands{
		RunCommands:     []string{"starten"},
		StopCommands:    []string{"stoppen"},
		EnableCommands:  []string{"aktivieren", "einschalten", "zeigen"},
		DisableCommands: []string{"deaktivieren", "ausschalten", "verstecken"},
		TriggerCommands: []string{"trigger", "auslöser"},
		X10Commands:     []string{"x10", "mal zehn"},
		InvCommands:     []string{"invertieren", "inv"},
		AcCommands:      []string{"ac"},
		DcCommands:      []string{"dc"},
		RisingCommands:  []string{"steigend"},
		FallingCommands: []string{"fallend"},
		ChannelA:        []string{"kanal a"},
		ChannelB:        []string{"kanal b"},
		ChannelC:        []string{"kanal c"},
		ChannelD:        []string{"kanal d"},
	}
	writeYaml(filepath.Join(dir, "de.yaml"), de)

	hu := VoiceCommands{
		RunCommands:     []string{"indulás", "indítás", "futás"},
		StopCommands:    []string{"állj", "leállítás", "megállítás"},
		EnableCommands:  []string{"engedélyezés", "bekapcsolás", "mutat"},
		DisableCommands: []string{"tiltás", "kikapcsolás", "elrejt"},
		TriggerCommands: []string{"trigger", "indítási feltétel"},
		X10Commands:     []string{"x10", "szer tíz"},
		InvCommands:     []string{"invertálás", "inv"},
		AcCommands:      []string{"ac", "váltakozó"},
		DcCommands:      []string{"dc", "egyen"},
		RisingCommands:  []string{"felfutó"},
		FallingCommands: []string{"lefutó"},
		ChannelA:        []string{"a csatorna", "csatorna a"},
		ChannelB:        []string{"b csatorna", "csatorna b"},
		ChannelC:        []string{"c csatorna", "csatorna c"},
		ChannelD:        []string{"d csatorna", "csatorna d"},
	}
	writeYaml(filepath.Join(dir, "hu.yaml"), hu)
}

func writeYaml(path string, cmds VoiceCommands) {
	data, err := yaml.Marshal(cmds)
	if err == nil {
		os.WriteFile(path, data, 0644)
	}
}

// ExecuteVoiceCommand processes natural language text commands
// and executes the corresponding UI/backend logic.
func (scp *ScpDesc) ExecuteVoiceCommand(cmd string) {
	cmd = strings.ToLower(strings.TrimSpace(cmd))
	slog.Debug("ExecuteVoiceCommand parsing", "cmd", cmd)

	// Helper to check for multiple keywords
	containsAny := func(s string, subs []string) bool {
		for _, sub := range subs {
			if sub != "" && strings.Contains(s, sub) {
				return true
			}
		}
		return false
	}

	// Ensure execution happens on the main UI thread, as many
	// of these commands will update Fyne widgets (checkboxes, buttons).
	fyne.Do(func() {
		// Run / Stop commands
		if containsAny(cmd, ActiveVoiceCommands.RunCommands) {
			// Try to start if stopped
			if scp.streamEnableButton != nil && !scp.streamEnableButton.Disabled() {
				scp.runblockButton.Tapped(&fyne.PointEvent{})
			}
		} else if containsAny(cmd, ActiveVoiceCommands.StopCommands) {
			// Trigger the stop action if running
			if scp.runblockButton != nil && !scp.runblockButton.Disabled() {
				scp.runblockButton.Tapped(&fyne.PointEvent{})
			}
		}

		// Channel Enable/Disable commands
		enable := containsAny(cmd, ActiveVoiceCommands.EnableCommands)
		disable := containsAny(cmd, ActiveVoiceCommands.DisableCommands)

		// property commands
		inv := containsAny(cmd, ActiveVoiceCommands.InvCommands)
		x10 := containsAny(cmd, ActiveVoiceCommands.X10Commands)
		ac := containsAny(cmd, ActiveVoiceCommands.AcCommands)
		dc := containsAny(cmd, ActiveVoiceCommands.DcCommands)
		rising := containsAny(cmd, ActiveVoiceCommands.RisingCommands)
		falling := containsAny(cmd, ActiveVoiceCommands.FallingCommands)

		// Check if a specific channel is mentioned
		var ch genericps.ChannelId = -1
		if containsAny(cmd, ActiveVoiceCommands.ChannelA) {
			ch = genericps.ChA
		} else if containsAny(cmd, ActiveVoiceCommands.ChannelB) {
			ch = genericps.ChB
		} else if containsAny(cmd, ActiveVoiceCommands.ChannelC) {
			ch = genericps.ChC
		} else if containsAny(cmd, ActiveVoiceCommands.ChannelD) {
			ch = genericps.ChD
		}

		if ch != -1 && int(ch) < int(scp.channelCount) {
			cv := scp.channelViewers[ch]
			handled := false

			if inv {
				handled = true
				if cv.invertCheckbox != nil {
					if disable {
						cv.invertCheckbox.SetChecked(false)
					} else if enable {
						cv.invertCheckbox.SetChecked(true)
					} else {
						// Toggle
						cv.invertCheckbox.SetChecked(!cv.invertCheckbox.Checked)
					}
				}
			}

			if x10 {
				handled = true
				if disable {
					scp.changeChannelX10(ch, false)
				} else if enable {
					scp.changeChannelX10(ch, true)
				} else {
					// Toggle
					current := scp.Settings.Channels[ch].X10
					scp.changeChannelX10(ch, !current)
				}
			}

			if ac {
				handled = true
				if cv.acdcSelect != nil {
					cv.acdcSelect.SetSelected("AC")
				}
			} else if dc {
				handled = true
				if cv.acdcSelect != nil {
					cv.acdcSelect.SetSelected("DC")
				}
			}

			if rising {
				handled = true
				if cv.triggerDirectionSelect != nil {
					cv.triggerDirectionSelect.SetSelected("Rising")
				}
			} else if falling {
				handled = true
				if cv.triggerDirectionSelect != nil {
					cv.triggerDirectionSelect.SetSelected("Falling")
				}
			}

			if !handled && (enable || disable) {
				// We update the UI checkbox directly, which triggers the logic
				if cv.enableCheckbox != nil {
					if cv.enableCheckbox.Val != enable {
						cv.enableCheckbox.Tapped(&fyne.PointEvent{})
					}
				} else {
					scp.EnableChannel(ch, enable)
				}
			}

			// Trigger source setting commands
			if containsAny(cmd, ActiveVoiceCommands.TriggerCommands) {
				if cv.triggerCheckbox != nil {
					cv.triggerCheckbox.SetChecked(true)
				}
			}
		} else {
			// If no channel is specified but we have a trigger channel active,
			// or if we have global commands, we can handle them here if needed.
		}
	})
}
