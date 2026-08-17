package gui

import (
	"fynescope/settings"
	"testing"

	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/test"
	"github.com/stretchr/testify/assert"
)

func TestNewGenPanel(t *testing.T) {
	test.NewApp()

	scp := &ScpDesc{
		Settings: &settings.PsSettings{
			Theme: settings.DarkTheme,
			GenPanel: settings.GeneratorSettings{
				On:        true,
				Frequency: 1000,
				Amplitude: 1000,
			},
		},
		theme: Theme(settings.DarkTheme),
	}

	cont := container.NewVBox()
	err := scp.newGenPanel(cont)
	
	assert.NoError(t, err)
	assert.NotNil(t, cont)
	assert.Greater(t, len(cont.Objects), 0, "Container should have objects added")
}
