package control

import (
	_ "fynescope/demo"
	"fynescope/genericps"
	"fynescope/settings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestEtsTimes_2407B(t *testing.T) {
	ps := &PscDesc{
		Info:            "2407B",
		ScopeModel:      StringToScopeType("2407B"),
		MaxSamplingRate: MaxSampling500M,
	}

	tests := []struct {
		sampleTime int32
		wantInter  int16
		wantCycles int16
		wantErr    bool
	}{
		{1000, 2, 4, false},
		{500, 4, 8, false},
		{400, 5, 10, false},
		{200, 10, 20, false},
		{100, 20, 40, false},
		{50, 40, 80, false},
		{40, 0, 0, true},   // Below 50
		{1001, 0, 0, true}, // Above 1000
	}

	for _, tt := range tests {
		cycles, inter, err := ps.EtsTimes(tt.sampleTime)
		if tt.wantErr {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
			assert.Equal(t, tt.wantInter, inter)
			assert.Equal(t, tt.wantCycles, cycles)
		}
	}
}

func TestEtsTimes_2207SIM(t *testing.T) {
	ps := &PscDesc{
		Info:            "2207SIM",
		ScopeModel:      StringToScopeType("2207SIM"),
		MaxSamplingRate: MaxSampling500M,
	}
	cycles, inter, err := ps.EtsTimes(200)
	assert.NoError(t, err)
	assert.Equal(t, int16(10), inter)
	assert.Equal(t, int16(20), cycles)
}

func TestEtsBlockModeBasic(t *testing.T) {
	con := genericps.NewConnection()
	handle, err := genericps.OpenDemo(con, genericps.DemoId)
	con.Handle = handle
	if err != nil {
		t.Fatalf("Failed to open simulator: %v", err)
	}
	defer con.CloseUnit()

	chSettings := make([]settings.ChSettings, 4)
	for i := range chSettings {
		chSettings[i] = settings.ChSettings{
			ID:      genericps.ChannelId(i),
			Enabled: i == 0,
		}
	}

	psControl := &PscDesc{
		Con:                    con,
		stateChannel:           make(chan state, 1),
		restartChannel:         make(chan struct{}, 1),
		stopChannel:            make(chan struct{}, 1),
		SetChannelCh:           make(chan *settings.ChSettings, 1),
		getChannelCh:           make(chan *getChannelMsg, 1),
		getNumOfEnabledCh:      make(chan *getNumOfEnabledChMsg, 1),
		SetScopeScreenWidthCh:  make(chan int32, 1),
		getScopeScreenWidthCh:  make(chan *getScopeScreenWidthMsg, 1),
		SetInterpolationModeCh: make(chan settings.InterpolationType, 1),
		getInterpolationModeCh: make(chan *getInterpolationModeMsg, 1),
		SetGeneratorCh:         make(chan *GeneratorDescMsg, 1),
		getGeneratorCh:         make(chan *getGeneratorMsg, 1),
		SetTriggerCh:           make(chan *TriggerDescMsg, 1),
		getTriggerCh:           make(chan *getTriggerMsg, 1),
		RefreshCallback: func(buffers [][]int16, buffersMin [][]int16, digitalBuffers [][]int16, triggerTimeOffset int64,
			xRoundError, samplingTimeInterval float64) {
		},
		RefreshEtsCallback: func(buffers [][]int16, etsOutBuffer []int64, xRoundError float64, samplingTimeInterval float64) {
		},
		BufferCallback:    func(size int) {},
		EtsBufferCallback: func(size int) {},
		DisplayStatus:     func(status string, level ScopeError) {},
	}
	psControl.scopeScreenWidth = 1000
	psControl.getChannel.newSettings = make(chan bool, 1)
	psControl.getNumOfEnabled.n = make(chan int, 1)
	psControl.getScopeScreenWidth.newSetting = make(chan bool, 1)
	psControl.getInterpolationMode.newSetting = make(chan bool, 1)
	psControl.getGenerator.newSetting = make(chan bool, 1)
	psControl.getTrigger.newSettings = make(chan bool, 1)

	go psControl.channelStateMachine(4)
	go psControl.generatorMonitor()
	go psControl.demoGeneratorMonitor()
	go psControl.interpolationMonitor()
	go psControl.triggerMonitor()

	psControl.SetChannelCh <- &chSettings[0]
	time.Sleep(10 * time.Millisecond)

	psControl.SetMaxScreenTime(1e-7)
	time.Sleep(10 * time.Millisecond)

	done := make(chan struct{})
	go func() {
		etsBlockMode(psControl)
		close(done)
	}()

	time.Sleep(100 * time.Millisecond)

	// Stop it
	psControl.stopChannel <- struct{}{}

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("etsBlockMode did not stop in time")
	}

	assert.NotNil(t, handle)
}
