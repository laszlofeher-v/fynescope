package control

import (
	_ "fynescope/demo"
	"fynescope/genericps"
	"fynescope/settings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestBlockModeBasic(t *testing.T) {
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
		BufferCallback: func(size int) {},
		DisplayStatus:  func(status string, level ScopeError) {},
	}
	psControl.StreamEnabled.Store(true)
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

	psControl.SetMaxScreenTime(1.0)
	time.Sleep(10 * time.Millisecond)

	done := make(chan struct{})
	go func() {
		// Run blockMode
		blockMode(psControl)
		close(done)
	}()

	time.Sleep(50 * time.Millisecond)

	// Stop it
	psControl.stopChannel <- struct{}{}

	select {
	case <-done:
	case <-time.After(1 * time.Second):
		t.Fatal("blockMode did not stop in time")
	}

	assert.NotNil(t, handle)
}

func TestBlockModeHugeScopeScreenWidth(t *testing.T) {
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
		BufferCallback: func(size int) {},
		DisplayStatus:  func(status string, level ScopeError) {},
	}
	psControl.StreamEnabled.Store(true)
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

	// Set astronomical screen width directly as reported in fuzzer crash
	psControl.scopeScreenWidth = 524800000000
	psControl.SetMaxScreenTime(0.001)
	time.Sleep(10 * time.Millisecond)

	done := make(chan struct{})
	go func() {
		blockMode(psControl)
		close(done)
	}()

	time.Sleep(50 * time.Millisecond)
	psControl.stopChannel <- struct{}{}

	select {
	case <-done:
	case <-time.After(1 * time.Second):
		t.Fatal("blockMode did not stop in time")
	}

	// SampleCountRequired must have been clamped to maxSampleCount (64MiB = 67108864) and NOT 524800000000
	assert.LessOrEqual(t, psControl.SampleCountRequired, uint64(67108864))
	assert.Greater(t, psControl.SampleCountRequired, uint64(0))
}

func TestBlockModeMaxScreenTimeSampleLimit(t *testing.T) {
	con := genericps.NewConnection()
	handle, err := genericps.OpenDemo(con, genericps.DemoId)
	con.Handle = handle
	if err != nil {
		t.Fatalf("Failed to open demo: %v", err)
	}
	defer func() {
		for i := 1; i < 4; i++ {
			_ = con.SetChannel(genericps.ChannelId(i), false, genericps.Dc, 7, 0)
		}
		con.CloseUnit()
	}()

	chSettings := make([]settings.ChSettings, 4)
	for i := range chSettings {
		chSettings[i] = settings.ChSettings{
			ID:      genericps.ChannelId(i),
			Enabled: true,
		}
	}

	var fatalLogged string
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
		BufferCallback: func(size int) {},
		DisplayStatus: func(status string, level ScopeError) {
			if level == Fatal {
				fatalLogged = status
			}
		},
	}
	psControl.StreamEnabled.Store(true)
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

	for i := range chSettings {
		psControl.SetChannelCh <- &chSettings[i]
	}
	time.Sleep(10 * time.Millisecond)

	// Simulate large maxScreenTime which previously could produce >1.38 billion samples
	psControl.scopeScreenWidth = 1382
	psControl.SetMaxScreenTime(2.764)
	time.Sleep(10 * time.Millisecond)

	done := make(chan struct{})
	go func() {
		blockMode(psControl)
		close(done)
	}()

	time.Sleep(50 * time.Millisecond)
	psControl.stopChannel <- struct{}{}

	select {
	case <-done:
	case <-time.After(1 * time.Second):
		t.Fatal("blockMode did not stop in time")
	}

	assert.Empty(t, fatalLogged, "RunBlock should not fail with fatal error (e.g. too many samples)")
	assert.LessOrEqual(t, psControl.SampleCountRequired, uint64(67108864))
	assert.Greater(t, psControl.SampleCountRequired, uint64(0))
}
