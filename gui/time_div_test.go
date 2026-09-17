package gui

import (
	"fynescope/control"
	"fynescope/selectscroll"
	"fynescope/settings"
	"testing"

	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"
	"github.com/stretchr/testify/assert"
)

func TestTimeDiv_GetTimeUnitName(t *testing.T) {
	initTimeMaps()
	assert.Equal(t, "ps", getTimeUnitName(-12))
	assert.Equal(t, "ns", getTimeUnitName(-9))
	assert.Equal(t, "µs", getTimeUnitName(-6))
	assert.Equal(t, "ms", getTimeUnitName(-3))
}

func TestTimeDiv_SwitchUpTimeUnit(t *testing.T) {
	initTimeMaps()
	dt, unitName := switchUpTimeUnit(2000, -9)
	assert.Equal(t, float32(2), dt)
	assert.Equal(t, "µs", unitName)

	dt, unitName = switchUpTimeUnit(100, -6)
	assert.Equal(t, float32(0.1), dt)
	assert.Equal(t, "ms", unitName)
}

func TestTimeDiv_CheckTimeZoomConstraint(t *testing.T) {
	test.NewApp()
	scp := &ScpDesc{
		psControl: &control.PscDesc{},
	}
	scp.initStatus()

	scp.timeZoomWindow = nil
	assert.True(t, scp.checkTimeZoomConstraint(500, -9))

	scp.timeZoomWindow = test.NewWindow(widget.NewLabel(""))
	scp.timeZoomMaxScreenTime = 1000 * 1e-9 * 10
	// 5000ns < 10000ns original
	assert.True(t, scp.checkTimeZoomConstraint(500, -9))
	assert.False(t, scp.checkTimeZoomConstraint(5000, -9))
}

func TestTimeDiv_SampleUnitUp(t *testing.T) {
	test.NewApp()
	scp := &ScpDesc{
		sampleRateSelect: selectscroll.NewSelectScroll([]string{"1", "2", "3"}, func(s string, e selectscroll.Exception) {}, ""),
		sampleUnitSelect: selectscroll.NewSelectScroll([]string{"Hz", "kHz", "MHz"}, func(s string, e selectscroll.Exception) {}, ""),
	}

	scp.sampleUnitSelect.SetSelectedIndex(1) // kHz
	scp.sampleUnitUp()

	assert.Equal(t, 2, scp.sampleRateSelect.SelectedIndex())
	assert.Equal(t, 0, scp.sampleUnitSelect.SelectedIndex()) // Hz
}

func TestTimeDiv_SampleUnitDown(t *testing.T) {
	test.NewApp()
	scp := &ScpDesc{
		sampleRateSelect: selectscroll.NewSelectScroll([]string{"1", "2", "3"}, func(s string, e selectscroll.Exception) {}, ""),
		sampleUnitSelect: selectscroll.NewSelectScroll([]string{"Hz", "kHz", "MHz"}, func(s string, e selectscroll.Exception) {}, ""),
	}

	scp.sampleUnitSelect.SetSelectedIndex(1) // kHz
	scp.sampleUnitDown()

	assert.Equal(t, 0, scp.sampleRateSelect.SelectedIndex())
	assert.Equal(t, 2, scp.sampleUnitSelect.SelectedIndex()) // MHz
}

func TestTimeDiv_UpdateTriggerModeOptions_2206BMSO(t *testing.T) {
	test.NewApp()
	scopeModel := control.StringToScopeType("2206BMSO")
	assert.Equal(t, control.Scope2206B_MSO, scopeModel)
	assert.True(t, scopeModel.IsETSCapable())
	assert.True(t, scopeModel.IsMSO())

	scp := &ScpDesc{
		psControl: &control.PscDesc{
			ScopeModel: scopeModel,
			Info:       "2206BMSO",
		},
		IsMSO: true,
		Settings: &settings.PsSettings{
			Digital: settings.DigitalSettings{
				Ports: [2]settings.DigitalPortSettings{
					{Enabled: false},
					{Enabled: false},
				},
			},
			Trigger: settings.TriggerSettings{
				Mode: settings.TriggerModeRepeat,
			},
		},
		triggerModeSelect: selectscroll.NewSelectScroll(triggerModeOptions, func(s string, e selectscroll.Exception) {}, settings.TriggerModeRepeat),
	}

	// When digital ports are disabled, ETS mode should be available in trigger mode options
	scp.updateTriggerModeOptions()
	opts := scp.triggerModeSelect.Options
	assert.Contains(t, opts, settings.TriggerModeETS)

	// When digital port 0 is enabled, ETS mode should be removed
	scp.Settings.Digital.Ports[0].Enabled = true
	scp.updateTriggerModeOptions()
	opts = scp.triggerModeSelect.Options
	assert.NotContains(t, opts, settings.TriggerModeETS)

	// When digital port 0 is disabled again, ETS mode should be restored
	scp.Settings.Digital.Ports[0].Enabled = false
	scp.updateTriggerModeOptions()
	opts = scp.triggerModeSelect.Options
	assert.Contains(t, opts, settings.TriggerModeETS)
}

