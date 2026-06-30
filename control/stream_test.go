package control

import (
	"fynescope/genericps"
	"fynescope/settings"
	_ "fynescope/sim"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestStreamModeTransition(t *testing.T) {
	con := genericps.NewConnection()
	handle, err := genericps.OpenSimulator(con, "sim")
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
		SetMaxScreenTimeCh:     make(chan float64, 1),
		getMaxScreenTimeCh:     make(chan *getMaxScreenTimeMsg, 1),
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
		RefreshCallback:        func(buffers [][]int16, triggerTimeOffset int64, xRoundError, samplingTimeInterval float64) {},
		BufferCallback:         func(size int) {},
		DisplayStatus:          func(status string, level ScopeError) {},
	}
	psControl.StreamEnabled.Store(true)
	psControl.getChannel.newSettings = make(chan bool, 1)
	psControl.getNumOfEnabled.n = make(chan int, 1)
	psControl.getMaxScreenTime.newSetting = make(chan bool, 1)
	psControl.getScopeScreenWidth.newSetting = make(chan bool, 1)
	psControl.getInterpolationMode.newSetting = make(chan bool, 1)
	psControl.getGenerator.newSetting = make(chan bool, 1)
	psControl.getTrigger.newSettings = make(chan bool, 1)

	go psControl.screenTimeMonitor()
	go psControl.channelStateMachine(4)
	go psControl.generatorMonitor()
	go psControl.simGeneratorMonitor()
	go psControl.interpolationMonitor()
	go psControl.triggerMonitor()

	psControl.SetChannelCh <- &chSettings[0]
	time.Sleep(10 * time.Millisecond)

	// Start with maxScreenTime < StreamThreshold
	psControl.SetMaxScreenTimeCh <- 1.0 // 1.0 < 2.1
	time.Sleep(10 * time.Millisecond)

	// Run stateMachine in a background goroutine
	go psControl.stateMachine()

	// Switch to blockMode
	psControl.stateChannel <- blockMode
	time.Sleep(50 * time.Millisecond)

	// Update maxScreenTime >= StreamThreshold
	psControl.SetMaxScreenTimeCh <- 3.0 // 3.0 >= 2.1
	time.Sleep(10 * time.Millisecond)

	// Trigger restart to make blockMode re-evaluate and transition to streamMode
	psControl.restartChannel <- struct{}{}
	time.Sleep(100 * time.Millisecond)

	// Clean up
	_ = psControl.Stop()
	psControl.stateChannel <- nil
	time.Sleep(10 * time.Millisecond)

	assert.NotNil(t, handle)
}
