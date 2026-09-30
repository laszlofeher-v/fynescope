package gui

import (
	"testing"

	"fyne.io/fyne/v2/test"
	"fynescope/control"
	"fynescope/genericps"
	"fynescope/settings"
	"github.com/stretchr/testify/assert"
)

type mockRemoteScopeController struct {
	info string
	min  int32
	max  int32
}

func (m *mockRemoteScopeController) SetTrigger(msg *control.TriggerDescMsg)                       {}
func (m *mockRemoteScopeController) SetChannel(msg *settings.ChSettings)                         {}
func (m *mockRemoteScopeController) SetDigitalPort(msg *control.DigitalPortMsg)                  {}
func (m *mockRemoteScopeController) SetGenerator(msg *control.GeneratorDescMsg)                 {}
func (m *mockRemoteScopeController) SetDemoGen(msg *control.GeneratorDescMsg)                   {}
func (m *mockRemoteScopeController) SetInterpolationMode(mode settings.InterpolationType)       {}
func (m *mockRemoteScopeController) SetScopeScreenWidth(width float64)                           {}
func (m *mockRemoteScopeController) SetMaxScreenTime(t float64)                                  {}
func (m *mockRemoteScopeController) SuggestSampleCount(sc uint64)                                {}
func (m *mockRemoteScopeController) SetETSMode() error                                           { return nil }
func (m *mockRemoteScopeController) SetBlockMode() error                                         { return nil }
func (m *mockRemoteScopeController) Shutdown()                                                   {}
func (m *mockRemoteScopeController) Stop() error                                                 { return nil }
func (m *mockRemoteScopeController) GetCon() *genericps.Connection                               { return nil }
func (m *mockRemoteScopeController) GetInfo() string                                             { return m.info }
func (m *mockRemoteScopeController) GetScopeModel() control.ScopeType                            { return control.Scope2206B_MSO }
func (m *mockRemoteScopeController) GetMaxSamplingRate() uint32                                  { return 1000000000 }
func (m *mockRemoteScopeController) GetStreamEnabled() bool                                      { return false }
func (m *mockRemoteScopeController) SetStreamEnabled(b bool)                                     {}
func (m *mockRemoteScopeController) GetSamplingTimeInterval() float64                            { return 1e-6 }
func (m *mockRemoteScopeController) GetTimeBase() uint64                                         { return 1 }
func (m *mockRemoteScopeController) ChannelRanges(chIndex genericps.ChannelId) ([]int32, error) {
	return []int32{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11}, nil
}
func (m *mockRemoteScopeController) NumberOfEnabledAnalogChannels() int { return 2 }
func (m *mockRemoteScopeController) MinMaxValues() (min, max int32, err error) {
	return m.min, m.max, nil
}
func (m *mockRemoteScopeController) UnitVariantInfo() (info string, err error) {
	return m.info, nil
}
func (m *mockRemoteScopeController) UnitBatchAndSerialInfo() (info string, err error) {
	return "TEST/1", nil
}
func (m *mockRemoteScopeController) SetDigitalPortEnabled(port int, enabled bool) {}
func (m *mockRemoteScopeController) SetRefreshCallback(cb func(buffers [][]int16, buffersMin [][]int16, digitalBuffers [][]int16, startTimeOffset int64, xRoundError float64, samplingTimeInterval float64)) {}
func (m *mockRemoteScopeController) SetRefreshEtsCallback(cb func(buffers [][]int16, etsOutBuffer []int64, xRoundError float64, samplingTimeInterval float64)) {}
func (m *mockRemoteScopeController) SetBufferCallback(cb func(size int)) {}
func (m *mockRemoteScopeController) SetEtsBufferCallback(cb func(size int)) {}
func (m *mockRemoteScopeController) SetDisplayStatus(cb func(s string, errorType control.ScopeError)) {}
func (m *mockRemoteScopeController) ShowDisplayStatus(s string, errorType control.ScopeError) {}
func (m *mockRemoteScopeController) NewChannels(numberOfChannels int) {}
func (m *mockRemoteScopeController) SetMaxSamplingRate(rate uint32) {}
func (m *mockRemoteScopeController) RequestRestart() {}
func (m *mockRemoteScopeController) SetResolutionMode(mode genericps.RatioMode) {}
func (m *mockRemoteScopeController) CallBufferCallback(size int) {}
func (m *mockRemoteScopeController) CallRefreshCallback(buffers [][]int16, buffersMin [][]int16, digitalBuffers [][]int16, startTimeOffset int64, xRoundError float64, samplingTimeInterval float64) {}
func (m *mockRemoteScopeController) SetInfo(info string) { m.info = info }
func (m *mockRemoteScopeController) SetScopeModel(model control.ScopeType) {}
func (m *mockRemoteScopeController) GetEtsLimits() (maxInterleave, maxCycles int16) { return 0, 0 }
func (m *mockRemoteScopeController) GetAnalogueOffset(voltageRange int, coupling genericps.Coupling) (maximumVoltage, minimumVoltage float32, err error) {
	return 0, 0, nil
}
func (m *mockRemoteScopeController) IsDemo() bool { return false }
func (m *mockRemoteScopeController) SetDemoDigitalGen(port0Enabled, port1Enabled bool, freq float64, dir genericps.DigitalDemoGenDirection, enc genericps.DigitalDemoGenEncoding, mode genericps.DigitalDemoGenMode, bitDelay float64) error {
	return nil
}
func (m *mockRemoteScopeController) SetDemoRlcFilter(channel genericps.ChannelId, genSource genericps.ChannelId, enabled bool, filterType string, r float64, runit string, l float64, lunit string, cval float64, cunit string) error {
	return nil
}

func TestMenuWithController_2206BMSO(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()

	scp := &ScpDesc{App: app}
	ctrl := &mockRemoteScopeController{
		info: "2206BMSO",
		min:  -32768,
		max:  32767,
	}

	cfg := settings.NewDefaultSettings()
	err := scp.MenuWithController(ctrl, cfg, "")
	assert.NoError(t, err)
	assert.Equal(t, int32(32767), scp.MaxValue)
	assert.Equal(t, int32(-32768), scp.MinValue)
	assert.True(t, scp.IsMSO)
	assert.NotNil(t, scp.Window)
}
