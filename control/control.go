package control

// No fyne dependency allowed in this package
import (
	"fmt"
	"fynescope/genericps"
	"fynescope/settings"
	"log/slog"
	"sync"
	"sync/atomic"

	"runtime"
	"strconv"
	"strings"
	"time"
)

func goid() int {
	var buf [64]byte
	n := runtime.Stack(buf[:], false)
	idField := strings.Fields(strings.TrimPrefix(string(buf[:n]), "goroutine "))[0]
	id, err := strconv.Atoi(idField)
	if err != nil {
		panic(fmt.Sprintf("cannot get goroutine id: %v", err))
	}
	return id
}

type ScopeError int

const (
	Fatal   ScopeError = iota // Fallback to idle
	Warning                   // continue
	Info                      // continue
)

const (
	// SincWMultiplier * sampleCount -> sampleCount
	// It must be odd.
	SincWMultiplier    = 5
	initialBufferSize  = 2048
	startTimeout       = 1000 * time.Millisecond
	haltTimeout        = 1000 * time.Millisecond
	etsCallbackTimeout = 100000 * time.Millisecond
	StreamThreshold    = 2.1 // seconds total screen time
)

type (
	TriggerDirections struct {
		ChannelA, ChannelB, ChannelC, ChannelD, Ext, Aux genericps.ThresholdDirection
	}
	TriggerDesc struct {
		Enabled            bool
		TriggerADC         int16
		LowerTriggerADC    int16
		HysteresisADC      uint16
		LowerHysteresisADC uint16
		UpperHysteresis    int32
		LowerHysteresis    int32
		Source             genericps.ChannelId
		ThresholdDirection genericps.ThresholdDirection
		ThresholdMode      genericps.ThresholdModeId
		Mode               TriggerModes
		Type               TriggerTypes
		Mv                 int32
		LowerMv            int32
		ComplexProperties  []genericps.TriggerChannelProperties
		ComplexConditions  []genericps.TriggerConditions
		ComplexDirections  []TriggerDirections
		IntervalType       genericps.PulseWidthType
		IntervalTimeLower  float64
		IntervalTimeUpper  float64
		// TimeOffset         int32
		// TriggerPreRatio float32
		XOffset       float64
		AutoTriggerMs int16
	}
	TriggerDescMsg struct {
		TriggerDesc
		Done chan struct{}
	}
	getTriggerMsg struct {
		triggerSettings *TriggerDesc
		newSettings     chan bool
	}

	getNumOfEnabledChMsg struct {
		n chan int
	}
	getChannelMsg struct {
		channelSettings *genericps.SetChannelMsg
		newSettings     chan bool
	}

	getInterpolationModeMsg struct {
		ipMode     settings.InterpolationType
		newSetting chan bool
	}

	getScopeScreenWidthMsg struct {
		width      int32
		newSetting chan bool
	}

	getMaxScreenTimeMsg struct {
		maxScreenTime float64
		newSetting    chan bool
	}

	GeneratorDesc struct {
		OffsetVoltage                                       int32
		PkToPK                                              uint32
		WaveType                                            genericps.WaveTypeEnum
		StartFrequency, StopFrequency, Increment, DwellTime float64
		SweepType                                           genericps.SweepTypeEnum
		Operation                                           genericps.ExtraOperations
		Shots, Sweeps                                       uint32
		TriggerType                                         genericps.SigGenTrigType
		TriggerSource                                       genericps.SigGenTrigSource
		ExtInThreshold                                      int16
		Phase                                               float64
		Channel                                             genericps.ChannelId
		On                                                  bool
	}
	GeneratorDescMsg struct {
		GeneratorDesc
		Done chan struct{}
	}

	getGeneratorMsg struct {
		generatorSettings *GeneratorDesc
		newSetting        chan bool
	}

	PscDesc struct {
		Con *genericps.Connection

		shutdownCh   chan struct{} // closed by Shutdown() to stop all monitor goroutines
		shutdownOnce sync.Once

		stateChannel   chan state
		stopChannel    chan struct{}
		restartChannel chan struct{}

		SetTriggerCh chan *TriggerDescMsg
		getTriggerCh chan *getTriggerMsg
		getTrigger   getTriggerMsg

		SetChannelCh      chan *settings.ChSettings
		getChannelCh      chan *getChannelMsg
		getNumOfEnabledCh chan *getNumOfEnabledChMsg
		getChannel        getChannelMsg
		getNumOfEnabled   getNumOfEnabledChMsg

		SetInterpolationModeCh chan settings.InterpolationType
		getInterpolationModeCh chan *getInterpolationModeMsg
		getInterpolationMode   getInterpolationModeMsg

		SetGeneratorCh chan *GeneratorDescMsg
		SetSimGenCh    chan *GeneratorDescMsg
		getGeneratorCh chan *getGeneratorMsg
		getGenerator   getGeneratorMsg

		SetScopeScreenWidthCh chan int32
		getScopeScreenWidthCh chan *getScopeScreenWidthMsg
		getScopeScreenWidth   getScopeScreenWidthMsg

		SetMaxScreenTimeCh chan float64
		getMaxScreenTimeCh chan *getMaxScreenTimeMsg
		getMaxScreenTime   getMaxScreenTimeMsg

		triggerSetting       TriggerDesc
		chEnabled            []atomic.Bool
		triggerTimeOffset    int64
		receiveBuffer        [][]int16   // raw data buffer, only for real channel
		displayBuffer        [][]float32 // signal stored in mv
		EtsInBuffer          []int64
		overSample                  int16
		SamplingTimeInterval        float64
		lastTriggerSamplingInterval float64
		SampleCountRequired         int32
		NPre, NPro           int32
		XRoundError          float64
		timeBase             uint32
		ipmode               settings.InterpolationType
		numOfSamplesAcquired uint32
		downSampleRatioMode  genericps.RatioMode
		downSampleRatio      uint32
		maxValue             int16
		maxScreenTime        float64
		scopeScreenWidth     float64
		timeBaseDec          uint32
		minValue             int16
		RefreshCallback      func(buffers [][]int16, startTimeOffset int64,
			xRoundError, samplingTimeInterval float64)
		RefreshEtsCallback func(buffers [][]int16, etsOutBuffer []int64, xRoundError float64)
		BufferCallback     func(size int)
		EtsBufferCallback  func(size int)
		DisplayStatus      func(s string, errorType ScopeError)
		refreshTime        time.Time
		Info               string
		MaxSamplingRate    uint32
		StreamEnabled      atomic.Bool
	}
)

// Shutdown signals all monitor goroutines launched by NewControl to exit.
// It is safe to call multiple times.
func (psControl *PscDesc) Shutdown() {
	psControl.shutdownOnce.Do(func() {
		close(psControl.shutdownCh)
	})
}

func NewControl(con *genericps.Connection) *PscDesc {
	slog.Debug("NewControl")
	psControl := &PscDesc{Con: con}
	psControl.shutdownCh = make(chan struct{})
	psControl.StreamEnabled.Store(true)
	psControl.stateChannel = make(chan state)
	psControl.restartChannel = make(chan struct{}, 1) // non blocking
	psControl.stopChannel = make(chan struct{}, 1)    // non blocking

	psControl.SetGeneratorCh = make(chan *GeneratorDescMsg)
	psControl.SetSimGenCh = make(chan *GeneratorDescMsg)
	psControl.getGeneratorCh = make(chan *getGeneratorMsg)
	psControl.getGenerator.newSetting = make(chan bool)

	psControl.SetChannelCh = make(chan *settings.ChSettings)
	psControl.getChannelCh = make(chan *getChannelMsg)
	psControl.getChannel.newSettings = make(chan bool)
	psControl.getNumOfEnabledCh = make(chan *getNumOfEnabledChMsg)
	psControl.getNumOfEnabled.n = make(chan int)

	psControl.SetInterpolationModeCh = make(chan settings.InterpolationType)
	psControl.getInterpolationModeCh = make(chan *getInterpolationModeMsg)
	psControl.getInterpolationMode.newSetting = make(chan bool)

	psControl.SetScopeScreenWidthCh = make(chan int32)
	psControl.getScopeScreenWidthCh = make(chan *getScopeScreenWidthMsg)
	psControl.getScopeScreenWidth.newSetting = make(chan bool)

	psControl.SetMaxScreenTimeCh = make(chan float64)
	psControl.getMaxScreenTimeCh = make(chan *getMaxScreenTimeMsg)
	psControl.getMaxScreenTime.newSetting = make(chan bool)

	psControl.SetTriggerCh = make(chan *TriggerDescMsg)
	psControl.getTriggerCh = make(chan *getTriggerMsg)
	psControl.getTrigger.triggerSettings = &psControl.triggerSetting
	psControl.getTrigger.newSettings = make(chan bool)
	go psControl.stateMachine()
	go psControl.triggerMonitor()
	go psControl.generatorMonitor()
	go psControl.simGeneratorMonitor()
	go psControl.interpolationMonitor()
	go psControl.screenTimeMonitor()
	return psControl
}

func (psControl *PscDesc) setMaxScreenTime() {
	psControl.getMaxScreenTimeCh <- &psControl.getMaxScreenTime
	if <-psControl.getMaxScreenTime.newSetting { // wait for data
		psControl.maxScreenTime = psControl.getMaxScreenTime.maxScreenTime
	}
}
func (psControl *PscDesc) getAnalogueOffset(voltageRange int,
	coupling genericps.Coupling) (maximumVoltage, minimumVoltage float32, err error) {
	maximumVoltage, minimumVoltage, err =
		psControl.Con.GetAnalogueOffset(voltageRange, coupling)
	return
}
func (psControl *PscDesc) setChannel() (err error) {
	psControl.getChannelCh <- &psControl.getChannel
	for <-psControl.getChannel.newSettings { // wait for data
		chset := psControl.getChannel.channelSettings
		err = psControl.Con.SetChannel(chset.Channel, chset.Enabled,
			chset.CouplingType, chset.VoltageRange, chset.AnalogOffset)
		if err != nil {
			slog.Error("SetChannel", "channels SetChannel:", err)
			return
		}
		psControl.getChannelCh <- &psControl.getChannel
	}
	return
}

func (psControl *PscDesc) setTrigger() (err error) {
	psControl.getTriggerCh <- &psControl.getTrigger   // ask for data
	newSettings := <-psControl.getTrigger.newSettings // wait for data

	samplingIntervalChanged := psControl.SamplingTimeInterval != psControl.lastTriggerSamplingInterval
	timeDependentTrigger := psControl.triggerSetting.Type == Interval || psControl.triggerSetting.Type == PulseWidth

	if newSettings || (samplingIntervalChanged && timeDependentTrigger) {
		err = psControl.sendTrigger() // 			   send to the scope
		if err != nil {
			slog.Error("setTrigger", "error", err)
			return
		}
		psControl.lastTriggerSamplingInterval = psControl.SamplingTimeInterval
	}
	return
}

func (psControl *PscDesc) setIpMode() {
	psControl.getInterpolationModeCh <- &psControl.getInterpolationMode
	if <-psControl.getInterpolationMode.newSetting { // wait for data
		psControl.ipmode = psControl.getInterpolationMode.ipMode
	}
}

func (psControl *PscDesc) setEverything() (err error) {
	psControl.setIpMode()
	psControl.setMaxScreenTime()
	err = psControl.setGenerator()
	if err != nil {
		slog.Error("setGenerator", "error", err)
		return
	}
	err = psControl.setChannel()
	if err != nil {
		slog.Error("setChannel", "error", err)
		return
	}
	return
}

func (psControl *PscDesc) sendTrigger() (err error) {
	switch psControl.triggerSetting.Type {
	case Simple:
		// if !psControl.triggerSetting.Enabled {
		err = psControl.sendSimpleTrigger()
		// }
	case Advanced:
		err = psControl.sendAdvancedTrigger()
	case Complex:
		err = psControl.sendComplexTrigger()
	case Window:
		err = psControl.sendWindowTrigger()
	case Interval:
		err = psControl.sendIntervalTrigger()
	case PulseWidth:
		err = psControl.sendPulseWidthTrigger()
	}

	return
}

func (psControl *PscDesc) SetScopeScreenWidth(w float64) {
	if psControl.scopeScreenWidth != w {
		psControl.scopeScreenWidth = w
		psControl.requestRestart()
	}
}

func (psControl *PscDesc) SuggestSampleCount(sc int32) {
	if psControl.SampleCountRequired != sc {
		psControl.SampleCountRequired = sc
		psControl.requestRestart()
	}
}

func (psControl *PscDesc) numberOfEnabledChannels() (n int) {
	psControl.getNumOfEnabledCh <- &psControl.getNumOfEnabled
	return <-psControl.getNumOfEnabled.n
}

func (psControl *PscDesc) NewChannels(numberOfChannels int) {
	go psControl.channelStateMachine(numberOfChannels)
	psControl.receiveBuffer = make([][]int16, numberOfChannels)
	psControl.displayBuffer = make([][]float32, numberOfChannels)
	for i := 0; i < numberOfChannels; i++ {
		psControl.receiveBuffer[i] = make([]int16, initialBufferSize)
		psControl.displayBuffer[i] = make([]float32, initialBufferSize)
	}
}

func (psControl *PscDesc) ChannelRanges(chIndex genericps.ChannelId) (ranges []int32,
	err error) {
	allowedRanges := make([]int32, 32)
	length, err := psControl.Con.GetChannelInformation(genericps.ChannelInfoRanges,
		0, allowedRanges, chIndex)
	if err != nil {
		slog.Error("Get ch info", "err", err)
		return
	}
	return allowedRanges[:length], err
}

func (psControl *PscDesc) UnitVariantInfo() (info string, err error) {
	info, err = psControl.Con.GetUnitInfo(genericps.PicoVariantInfo)
	return
}
func (psControl *PscDesc) UnitBatchAndSerialInfo() (info string, err error) {
	info, err = psControl.Con.GetUnitInfo(genericps.PicoBatchAndSerial)
	return
}

func (psControl *PscDesc) MinMaxValues() (min, max int16, err error) {
	max, err = psControl.Con.MaximumValue()
	if err != nil {
		slog.Error("MaximumValue", "error", err)
	}
	if err != nil {
		slog.Error("MinimumValue", "error", err)
		return
	}
	min, err = psControl.Con.MinimumValue()
	psControl.maxValue = max
	psControl.minValue = min
	return
}

func (psControl *PscDesc) Stop() (err error) {
	select {
	case psControl.stopChannel <- struct{}{}:
	case <-time.After(haltTimeout):
		err = fmt.Errorf("Halt send timeout")
		slog.Error("Halt send timeout", "error", err)
		_ = psControl.Con.Stop() // Try to stop anyway
		return
	}
	return
}

func (psControl *PscDesc) SetETSMode() (err error) {
	select {
	case psControl.stateChannel <- etsBlockMode:
	case <-time.After(startTimeout):
		err = fmt.Errorf("Could not start ETS mode")
	}
	return
}

func (psControl *PscDesc) SetBlockMode() (err error) {
	select {
	case psControl.stateChannel <- blockMode:
	case <-time.After(startTimeout):
		err = fmt.Errorf("Could not start block mode")
	}
	return
}
