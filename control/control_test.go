package control

import (
	"sync/atomic"
	"testing"
	"time"

	_ "fynescope/demo"
	"fynescope/genericps"
	"fynescope/settings"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGoid(t *testing.T) {
	id1 := goid()
	id2 := goid()

	if id1 <= 0 {
		t.Errorf("Expected positive goroutine ID, got %d", id1)
	}

	if id1 != id2 {
		t.Errorf("Expected identical goroutine IDs for the same goroutine, got %d and %d", id1, id2)
	}

	ch := make(chan int)
	go func() {
		ch <- goid()
	}()

	id3 := <-ch
	if id3 == id1 {
		t.Errorf("Expected different goroutine ID for a new goroutine, got %d for both", id1)
	}
}

func TestGeneratorDesc_Equals(t *testing.T) {
	g1 := &GeneratorDesc{
		OffsetVoltage:     100,
		PkToPK:            200,
		ArbitraryWaveform: []int16{1, 2, 3},
	}
	g2 := &GeneratorDesc{
		OffsetVoltage:     100,
		PkToPK:            200,
		ArbitraryWaveform: []int16{1, 2, 3},
	}
	assert.True(t, g1.Equals(g2))

	g2.PkToPK = 300
	assert.False(t, g1.Equals(g2))

	g2.PkToPK = 200
	g2.ArbitraryWaveform = []int16{1, 2}
	assert.False(t, g1.Equals(g2))

	g2.ArbitraryWaveform = []int16{1, 2, 4}
	assert.False(t, g1.Equals(g2))
}

func TestNewControlAndShutdown(t *testing.T) {
	con := genericps.NewConnection()
	handle, err := genericps.OpenDemo(con, genericps.DemoId)
	require.NoError(t, err)
	con.Handle = handle
	defer con.CloseUnit()

	psControl := NewControl(con)
	require.NotNil(t, psControl)

	psControl.Shutdown()
	// Calling shutdown twice should be safe
	psControl.Shutdown()
}

func TestPscDesc_SettersAndRestart(t *testing.T) {
	psControl := &PscDesc{
		restartChannel: make(chan struct{}, 10),
	}

	psControl.SetScopeScreenWidth(100.0)
	assert.Equal(t, 100.0, psControl.scopeScreenWidth)
	assert.Len(t, psControl.restartChannel, 1)

	// Same width should not trigger restart
	psControl.SetScopeScreenWidth(100.0)
	assert.Len(t, psControl.restartChannel, 1)

	psControl.SetMaxScreenTime(0.5)
	assert.Equal(t, 0.5, psControl.maxScreenTime)
	assert.Len(t, psControl.restartChannel, 2)

	psControl.SuggestSampleCount(2000)
	assert.Equal(t, uint64(2000), psControl.SampleCountRequired)
	assert.Len(t, psControl.restartChannel, 3)
}

func TestPscDesc_Channels(t *testing.T) {
	psControl := &PscDesc{
		shutdownCh:        make(chan struct{}),
		SetChannelCh:      make(chan *settings.ChSettings, 10),
		getChannelCh:      make(chan *getChannelMsg, 10),
		getNumOfEnabledCh: make(chan *getNumOfEnabledChMsg, 10),
		getNumOfEnabled: getNumOfEnabledChMsg{
			n: make(chan int, 10),
		},
		digitalPortsEnabled: [2]atomic.Bool{},
	}
	defer close(psControl.shutdownCh)

	psControl.NewChannels(2)
	assert.Len(t, psControl.receiveBuffer, 2)
	assert.Len(t, psControl.displayBuffer, 2)
	psControl.SetChannelCh <- &settings.ChSettings{ID: genericps.ChA, Enabled: true}
	psControl.SetChannelCh <- &settings.ChSettings{ID: genericps.ChB, Enabled: true}
	time.Sleep(10 * time.Millisecond)

	nAnalog := psControl.NumberOfEnabledAnalogChannels()
	assert.Equal(t, 2, nAnalog)

	psControl.digitalPortsEnabled[0].Store(true)
	nTotal := psControl.numberOfEnabledChannels()
	assert.Equal(t, 3, nTotal)
}

func TestPscDesc_InfoAndRanges(t *testing.T) {
	con := genericps.NewConnection()
	handle, err := genericps.OpenDemo(con, genericps.DemoId)
	require.NoError(t, err)
	con.Handle = handle
	defer con.CloseUnit()

	psControl := &PscDesc{Con: con, Info: "ps6000a"}

	variant, err := psControl.UnitVariantInfo()
	assert.NoError(t, err)
	assert.NotEmpty(t, variant)

	serial, err := psControl.UnitBatchAndSerialInfo()
	assert.NoError(t, err)
	assert.NotEmpty(t, serial)

	minVal, maxVal, err := psControl.MinMaxValues()
	assert.NoError(t, err)
	assert.Less(t, minVal, maxVal)

	ranges, err := psControl.ChannelRanges(genericps.ChA)
	assert.NoError(t, err)
	assert.NotEmpty(t, ranges)
}

func TestPscDesc_ModeTransitions(t *testing.T) {
	con := genericps.NewConnection()
	handle, err := genericps.OpenDemo(con, genericps.DemoId)
	require.NoError(t, err)
	con.Handle = handle
	defer con.CloseUnit()

	psControl := &PscDesc{
		Con:            con,
		stateChannel:   make(chan state, 2),
		stopChannel:    make(chan struct{}, 2),
		restartChannel: make(chan struct{}, 2),
	}

	err = psControl.stopHardware()
	assert.NoError(t, err)

	err = psControl.Stop()
	assert.NoError(t, err)
	// Duplicate stop request should be ignored silently
	err = psControl.Stop()
	assert.NoError(t, err)

	err = psControl.SetETSMode()
	assert.NoError(t, err)
	st := <-psControl.stateChannel
	assert.NotNil(t, st)

	err = psControl.SetBlockMode()
	assert.NoError(t, err)
	st = <-psControl.stateChannel
	assert.NotNil(t, st)
}

func TestPscDesc_SendTriggerAllTypes(t *testing.T) {
	con := genericps.NewConnection()
	handle, err := genericps.OpenDemo(con, genericps.DemoId)
	require.NoError(t, err)
	con.Handle = handle
	defer con.CloseUnit()

	psControl := &PscDesc{
		Con: con,
		triggerSetting: TriggerDesc{
			Enabled:       true,
			Source:        genericps.ChA,
			TriggerADC:    1000,
			Mode:          Auto,
			AutoTriggerMs: 1000,
		},
	}

	types := []TriggerTypes{
		Simple, Advanced, Complex, Window, Interval, PulseWidth,
		WindowPulseWidth, Dropout, WindowDropout, Runt, RiseFall,
	}

	for _, tt := range types {
		psControl.triggerSetting.Type = tt
		err := psControl.sendTrigger()
		assert.NoError(t, err, "sendTrigger failed for type %v", tt)
	}

	// Disabled trigger -> simple trigger
	psControl.triggerSetting.Enabled = false
	err = psControl.sendTrigger()
	assert.NoError(t, err)

	// Disabled analog trigger with digital trigger enabled
	psControl.triggerSetting.DigitalTriggerEnabled = true
	err = psControl.sendTrigger()
	assert.NoError(t, err)
}

func TestPscDesc_SetIpMode(t *testing.T) {
	psControl := &PscDesc{
		getInterpolationModeCh: make(chan *getInterpolationModeMsg, 1),
		getInterpolationMode: getInterpolationModeMsg{
			newSetting: make(chan bool, 1),
			ipMode:     settings.Sinc,
		},
	}

	go func() {
		msg := <-psControl.getInterpolationModeCh
		msg.newSetting <- true
	}()

	psControl.setIpMode()
	assert.Equal(t, settings.Sinc, psControl.ipmode)
}
