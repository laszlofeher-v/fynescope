package gui

import (
	"image/color"
	"testing"

	"fynescope/control"
	"fynescope/selectscroll"
	"fynescope/settings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/stretchr/testify/assert"
)

func TestTriggerStatusLabel_InitialAndStoppedState(t *testing.T) {
	test.NewApp()
	scp := &ScpDesc{
		running: false,
		Settings: &settings.PsSettings{
			Trigger: settings.TriggerSettings{
				Mode: settings.TriggerModeRepeat,
			},
		},
	}
	scp.triggerStatusLabel = widget.NewLabel(triggerStopStat)
	scp.triggerStatusLabel.TextStyle = fyne.TextStyle{Bold: true, Monospace: true}
	scp.triggerStatusLabel.Importance = widget.DangerImportance

	// Verify initial state
	assert.Equal(t, triggerStopStat, scp.triggerStatusLabel.Text)
	assert.Equal(t, widget.DangerImportance, scp.triggerStatusLabel.Importance)

	// When stopped, trigger state notifications should remain "Stop"
	scp.updateTriggerState(true)
	assert.Equal(t, triggerStopStat, scp.triggerStatusLabel.Text)
	assert.Equal(t, widget.DangerImportance, scp.triggerStatusLabel.Importance)

	scp.updateTriggerState(false)
	assert.Equal(t, triggerStopStat, scp.triggerStatusLabel.Text)
	assert.Equal(t, widget.DangerImportance, scp.triggerStatusLabel.Importance)

	// Calling updateTriggerStateRunning while stopped should not change label
	scp.updateTriggerStateRunning()
	assert.Equal(t, triggerStopStat, scp.triggerStatusLabel.Text)
	assert.Equal(t, widget.DangerImportance, scp.triggerStatusLabel.Importance)
}

func TestTriggerStatusLabel_RunningRepeatMode(t *testing.T) {
	test.NewApp()
	scp := &ScpDesc{
		running: true,
		Settings: &settings.PsSettings{
			Trigger: settings.TriggerSettings{
				Mode: settings.TriggerModeRepeat,
			},
		},
	}
	scp.triggerStatusLabel = widget.NewLabel(triggerStopStat)
	scp.triggerStatusLabel.TextStyle = fyne.TextStyle{Bold: true, Monospace: true}
	scp.triggerStatusLabel.Importance = widget.DangerImportance

	// Started running: waiting for trigger
	scp.updateTriggerStateRunning()
	assert.Equal(t, triggerWaitStat, scp.triggerStatusLabel.Text)
	assert.Equal(t, widget.WarningImportance, scp.triggerStatusLabel.Importance)

	// Fired trigger: Trig'd
	scp.updateTriggerState(true)
	assert.Equal(t, triggeredStat, scp.triggerStatusLabel.Text)
	assert.Equal(t, widget.SuccessImportance, scp.triggerStatusLabel.Importance)

	// Trigger released / waiting again: Wait
	scp.updateTriggerState(false)
	assert.Equal(t, triggerWaitStat, scp.triggerStatusLabel.Text)
	assert.Equal(t, widget.WarningImportance, scp.triggerStatusLabel.Importance)

	// Stopped: Stop
	scp.running = false
	scp.updateTriggerStateStopped()
	assert.Equal(t, triggerStopStat, scp.triggerStatusLabel.Text)
	assert.Equal(t, widget.DangerImportance, scp.triggerStatusLabel.Importance)
}

func TestTriggerStatusLabel_RunningAutoMode(t *testing.T) {
	test.NewApp()
	scp := &ScpDesc{
		running: true,
		Settings: &settings.PsSettings{
			Trigger: settings.TriggerSettings{
				Mode: settings.TriggerModeAuto,
			},
		},
	}
	scp.triggerStatusLabel = widget.NewLabel(triggerStopStat)
	scp.triggerStatusLabel.TextStyle = fyne.TextStyle{Bold: true, Monospace: true}
	scp.triggerStatusLabel.Importance = widget.DangerImportance

	// Started running: waiting in Auto mode displays "Auto"
	scp.updateTriggerStateRunning()
	assert.Equal(t, triggerAutoStat, scp.triggerStatusLabel.Text)
	assert.Equal(t, widget.WarningImportance, scp.triggerStatusLabel.Importance)

	// Fired trigger: Trig'd
	scp.updateTriggerState(true)
	assert.Equal(t, triggeredStat, scp.triggerStatusLabel.Text)
	assert.Equal(t, widget.SuccessImportance, scp.triggerStatusLabel.Importance)

	// Trigger released / auto timeout: Auto
	scp.updateTriggerState(false)
	assert.Equal(t, triggerAutoStat, scp.triggerStatusLabel.Text)
	assert.Equal(t, widget.WarningImportance, scp.triggerStatusLabel.Importance)

	// Stopped: Stop
	scp.running = false
	scp.updateTriggerStateStopped()
	assert.Equal(t, triggerStopStat, scp.triggerStatusLabel.Text)
	assert.Equal(t, widget.DangerImportance, scp.triggerStatusLabel.Importance)
}

type triggerTestScopeController struct {
	mockRemoteScopeController
	triggerStateCb  func(triggered bool)
	displayStatusCb func(s string, errorType control.ScopeError)
}

func (m *triggerTestScopeController) SetTriggerStateCallback(cb func(triggered bool)) {
	m.triggerStateCb = cb
}

func (m *triggerTestScopeController) SetDisplayStatus(cb func(s string, errorType control.ScopeError)) {
	m.displayStatusCb = cb
}

func (m *triggerTestScopeController) ShowDisplayStatus(s string, errorType control.ScopeError) {
	if m.displayStatusCb != nil {
		m.displayStatusCb(s, errorType)
	}
}

func TestTriggerStatusLabel_CallbackIntegration(t *testing.T) {
	test.NewApp()
	mockCtrl := &triggerTestScopeController{}
	scp := &ScpDesc{
		running:   true,
		psControl: mockCtrl,
		Settings: &settings.PsSettings{
			Trigger: settings.TriggerSettings{
				Mode: settings.TriggerModeRepeat,
			},
		},
	}
	scp.triggerStatusLabel = widget.NewLabel(triggerStopStat)
	scp.triggerStatusLabel.TextStyle = fyne.TextStyle{Bold: true, Monospace: true}
	scp.triggerStatusLabel.Importance = widget.DangerImportance

	// Wire up callback as done in gui.go
	mockCtrl.SetTriggerStateCallback(func(triggered bool) {
		scp.updateTriggerState(triggered)
	})

	assert.NotNil(t, mockCtrl.triggerStateCb)

	// Trigger fired
	mockCtrl.triggerStateCb(true)
	assert.Equal(t, triggeredStat, scp.triggerStatusLabel.Text)
	assert.Equal(t, widget.SuccessImportance, scp.triggerStatusLabel.Importance)

	// Trigger waiting
	mockCtrl.triggerStateCb(false)
	assert.Equal(t, triggerWaitStat, scp.triggerStatusLabel.Text)
	assert.Equal(t, widget.WarningImportance, scp.triggerStatusLabel.Importance)
}

func TestTriggerStatusLabel_FatalError(t *testing.T) {
	test.NewApp()
	mockCtrl := &triggerTestScopeController{}
	scp := &ScpDesc{
		running:        true,
		psControl:      mockCtrl,
		runblockButton: widget.NewButtonWithIcon("", theme.MediaPauseIcon(), nil),
		Settings: &settings.PsSettings{
			Trigger: settings.TriggerSettings{
				Mode: settings.TriggerModeRepeat,
			},
		},
	}
	scp.triggerStatusLabel = widget.NewLabel(triggerWaitStat)
	scp.triggerStatusLabel.TextStyle = fyne.TextStyle{Bold: true, Monospace: true}
	scp.triggerStatusLabel.Importance = widget.WarningImportance

	scp.initStatus()

	// Simulate fatal error
	mockCtrl.ShowDisplayStatus("Fatal error occurred", control.Fatal)

	assert.False(t, scp.running)
	assert.Equal(t, triggerStopStat, scp.triggerStatusLabel.Text)
	assert.Equal(t, widget.DangerImportance, scp.triggerStatusLabel.Importance)
}

func TestTriggerStatusLabel_ModeChangeTransitions(t *testing.T) {
	test.NewApp()
	mockCtrl := &triggerTestScopeController{}
	scp := &ScpDesc{
		running:        true,
		psControl:      mockCtrl,
		runblockButton: widget.NewButtonWithIcon("", theme.MediaPauseIcon(), nil),
		Settings: &settings.PsSettings{
			Trigger: settings.TriggerSettings{
				Mode: settings.TriggerModeAuto,
			},
		},
		repartition: make(chan struct{}, 1),
	}
	scp.triggerStatusLabel = widget.NewLabel(triggerAutoStat)
	scp.triggerStatusLabel.TextStyle = fyne.TextStyle{Bold: true, Monospace: true}
	scp.triggerStatusLabel.Importance = widget.WarningImportance

	// Switch from Auto to Repeat while running and not triggered
	scp.onTriggerModeChange(settings.TriggerModeRepeat, selectscroll.None)
	assert.Equal(t, triggerWaitStat, scp.triggerStatusLabel.Text)
	assert.Equal(t, widget.WarningImportance, scp.triggerStatusLabel.Importance)

	// Switch from Repeat to Auto while running and not triggered
	scp.onTriggerModeChange(settings.TriggerModeAuto, selectscroll.None)
	assert.Equal(t, triggerAutoStat, scp.triggerStatusLabel.Text)
	assert.Equal(t, widget.WarningImportance, scp.triggerStatusLabel.Importance)

	// If currently Trig'd, mode change should preserve Trig'd
	scp.triggerStatusLabel.Text = triggeredStat
	scp.triggerStatusLabel.Importance = widget.SuccessImportance
	scp.onTriggerModeChange(settings.TriggerModeRepeat, selectscroll.None)
	assert.Equal(t, triggeredStat, scp.triggerStatusLabel.Text)
	assert.Equal(t, widget.SuccessImportance, scp.triggerStatusLabel.Importance)

	// Switch to Single while running: should stop acquisition and show Stop
	scp.onTriggerModeChange(settings.TriggerModeSingle, selectscroll.None)
	assert.False(t, scp.running)
	assert.Equal(t, triggerStopStat, scp.triggerStatusLabel.Text)
	assert.Equal(t, widget.DangerImportance, scp.triggerStatusLabel.Importance)
}

func TestTriggerStatusLabel_RunningETSMode(t *testing.T) {
	test.NewApp()
	scp := &ScpDesc{
		running: true,
		Settings: &settings.PsSettings{
			Trigger: settings.TriggerSettings{
				Mode: settings.TriggerModeETS,
			},
		},
	}
	scp.triggerStatusLabel = widget.NewLabel(triggerStopStat)
	scp.triggerStatusLabel.TextStyle = fyne.TextStyle{Bold: true, Monospace: true}
	scp.triggerStatusLabel.Importance = widget.DangerImportance

	// Started running in ETS mode: waiting for trigger
	scp.updateTriggerStateRunning()
	assert.Equal(t, triggerWaitStat, scp.triggerStatusLabel.Text)
	assert.Equal(t, widget.WarningImportance, scp.triggerStatusLabel.Importance)

	// Fired trigger: Trig'd
	scp.updateTriggerState(true)
	assert.Equal(t, triggeredStat, scp.triggerStatusLabel.Text)
	assert.Equal(t, widget.SuccessImportance, scp.triggerStatusLabel.Importance)

	// Trigger released / waiting again (e.g. trigger point out of signal range): Wait
	scp.updateTriggerState(false)
	assert.Equal(t, triggerWaitStat, scp.triggerStatusLabel.Text)
	assert.Equal(t, widget.WarningImportance, scp.triggerStatusLabel.Importance)

	// Stopped: Stop
	scp.running = false
	scp.updateTriggerStateStopped()
	assert.Equal(t, triggerStopStat, scp.triggerStatusLabel.Text)
	assert.Equal(t, widget.DangerImportance, scp.triggerStatusLabel.Importance)
}

func TestTriggerStatusLabel_ModeChangeTransitionsETS(t *testing.T) {
	test.NewApp()
	mockCtrl := &triggerTestScopeController{}
	scp := &ScpDesc{
		running:        true,
		psControl:      mockCtrl,
		runblockButton: widget.NewButtonWithIcon("", theme.MediaPauseIcon(), nil),
		channelViewers: []channelViewerDesc{
			{
				triggerCheckbox: widget.NewCheck("", nil),
			},
		},
		Settings: &settings.PsSettings{
			Trigger: settings.TriggerSettings{
				Mode: settings.TriggerModeAuto,
			},
			Channels: make([]settings.ChSettings, 4),
		},
		repartition: make(chan struct{}, 1),
	}
	scp.triggerStatusLabel = widget.NewLabel(triggeredStat)
	scp.triggerStatusLabel.TextStyle = fyne.TextStyle{Bold: true, Monospace: true}
	scp.triggerStatusLabel.Importance = widget.SuccessImportance

	// Switch from Auto to ETS while running and currently Trig'd
	// Status must immediately transition to "Wait" (not remaining Trig'd)
	scp.onTriggerModeChange(settings.TriggerModeETS, selectscroll.None)
	assert.Equal(t, triggerWaitStat, scp.triggerStatusLabel.Text)
	assert.Equal(t, widget.WarningImportance, scp.triggerStatusLabel.Importance)

	// Trigger fires in ETS mode -> Trig'd
	scp.updateTriggerState(true)
	assert.Equal(t, triggeredStat, scp.triggerStatusLabel.Text)
	assert.Equal(t, widget.SuccessImportance, scp.triggerStatusLabel.Importance)

	// Trigger point out of signal range in ETS mode -> Wait
	scp.updateTriggerState(false)
	assert.Equal(t, triggerWaitStat, scp.triggerStatusLabel.Text)
	assert.Equal(t, widget.WarningImportance, scp.triggerStatusLabel.Importance)

	// Switch back to Auto while running -> transitions to "Auto"
	scp.onTriggerModeChange(settings.TriggerModeAuto, selectscroll.None)
	assert.Equal(t, triggerAutoStat, scp.triggerStatusLabel.Text)
	assert.Equal(t, widget.WarningImportance, scp.triggerStatusLabel.Importance)
}

func TestTriggerStatusLabel_ThemeColors(t *testing.T) {
	darkWarning := ScpDarkTheme{}.Color(theme.ColorNameWarning, theme.VariantDark)
	assert.Equal(t, color.NRGBA{R: 255, G: 165, B: 0, A: 255}, darkWarning)

	lightWarning := ScpLightTheme{}.Color(theme.ColorNameWarning, theme.VariantLight)
	assert.Equal(t, color.NRGBA{R: 217, G: 119, B: 6, A: 255}, lightWarning)
}

func TestTriggerStatusLabel_AutoModeUnsupportedTriggerTypes(t *testing.T) {
	test.NewApp()
	mockCtrl := &triggerTestScopeController{}

	unsupportedTypes := []string{
		settings.TriggerTypeWindowPulseWidth,
		settings.TriggerTypeWindowDropout,
		settings.TriggerTypeInterval,
		settings.TriggerTypePulseWidth,
		settings.TriggerTypeDropout,
		settings.TriggerTypeRunt,
		settings.TriggerTypeRiseFall,
	}

	for _, tType := range unsupportedTypes {
		t.Run(tType, func(t *testing.T) {
			scp := &ScpDesc{
				running:   true,
				psControl: mockCtrl,
				Settings: &settings.PsSettings{
					Trigger: settings.TriggerSettings{
						Mode: settings.TriggerModeAuto,
						Type: tType,
					},
					Channels: make([]settings.ChSettings, 4),
				},
				channelViewers: make([]channelViewerDesc, 4),
				repartition:    make(chan struct{}, 1),
			}
			scp.triggerStatusLabel = widget.NewLabel(triggerStopStat)
			scp.triggerStatusLabel.TextStyle = fyne.TextStyle{Bold: true, Monospace: true}
			scp.triggerStatusLabel.Importance = widget.DangerImportance

			// When running in Auto mode but with an unsupported auto-trigger type:
			// Waiting should display "Wait" instead of "Auto"
			scp.updateTriggerStateRunning()
			assert.Equal(t, triggerWaitStat, scp.triggerStatusLabel.Text)
			assert.Equal(t, widget.WarningImportance, scp.triggerStatusLabel.Importance)

			// When triggered: should still display "Trig'd"
			scp.updateTriggerState(true)
			assert.Equal(t, triggeredStat, scp.triggerStatusLabel.Text)
			assert.Equal(t, widget.SuccessImportance, scp.triggerStatusLabel.Importance)

			// When untriggered: should display "Wait" instead of "Auto"
			scp.updateTriggerState(false)
			assert.Equal(t, triggerWaitStat, scp.triggerStatusLabel.Text)
			assert.Equal(t, widget.WarningImportance, scp.triggerStatusLabel.Importance)

			// Switching trigger type from Simple to this unsupported type while running
			scp.Settings.Trigger.Type = settings.TriggerTypeSimple
			scp.updateTriggerState(false)
			assert.Equal(t, triggerAutoStat, scp.triggerStatusLabel.Text)

			scp.onTriggerTypeChange(tType, selectscroll.None)
			assert.Equal(t, triggerWaitStat, scp.triggerStatusLabel.Text)
		})
	}
}
