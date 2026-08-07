package web

import "fynescope/settings"

type ScopeController interface {
	GetStatus() string
	GetSettings() *settings.PsSettings
	ApplySettingsJSON(data []byte) error
	RunScope() error
	StopScope() error
	AutoRange() error
	ExecuteVoiceCommand(string)
}
