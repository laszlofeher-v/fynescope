package gui

import (
	"encoding/json"
	"fynescope/settings"
)

// GetStatus returns the current run status of the oscilloscope.
func (scp *ScpDesc) GetStatus() string {
	if scp.running {
		return "Running"
	}
	return "Stopped"
}

// GetSettings returns a pointer to the current application settings.
func (scp *ScpDesc) GetSettings() *settings.PsSettings {
	return scp.Settings
}

// ApplySettingsJSON unmarshals a JSON payload over the existing settings struct (deep merge)
// and handles applying those settings to the hardware/UI.
func (scp *ScpDesc) ApplySettingsJSON(data []byte) error {
	if err := json.Unmarshal(data, scp.Settings); err != nil {
		return err
	}
	
	// Save the new settings to disk.
	scp.SaveSettings()
	
	return nil
}

// RunScope starts the oscilloscope capture.
func (scp *ScpDesc) RunScope() error {
	scp.ExecuteVoiceCommand("run")
	return nil
}

// StopScope stops the oscilloscope capture.
func (scp *ScpDesc) StopScope() error {
	scp.ExecuteVoiceCommand("stop")
	return nil
}

// AutoRange triggers the auto range feature.
func (scp *ScpDesc) AutoRange() error {
	scp.ExecuteVoiceCommand("auto range")
	return nil
}
