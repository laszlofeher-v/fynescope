package control

// No fyne dependency allowed in this package
import (
	"errors"
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

// goid returns the numeric id of the calling goroutine (parsed from the
// runtime stack header). Intended for debugging/logging only.
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

// ScopeError classifies the severity of a message reported through
// PscDesc.DisplayStatus.
type ScopeError int

const (
	Fatal   ScopeError = iota // Fallback to idle
	Warning                   // continue
	Info                      // continue
)

const (
	// SincWMultiplier * sampleCount -> sampleCount
	// It must be odd.
	SincWMultiplier = 5
	// initialBufferSize is the number of samples allocated per channel buffer
	// before the first acquisition resizes them.
	initialBufferSize = 2048
	// startTimeout is how long a mode switch waits for the state machine.
	startTimeout = 1000 * time.Millisecond
	// etsCallbackTimeout is the maximum wait for an ETS (equivalent time
	// sampling) callback.
	etsCallbackTimeout = 100000 * time.Millisecond
	// StreamThreshold is the total screen time above which streaming is used.
	StreamThreshold = 2.1 // seconds total screen time
)

type (
	// TriggerDirections holds the threshold direction for each trigger source,
	// used by complex (multi-condition) triggers.
	TriggerDirections struct {
		ChannelA, ChannelB, ChannelC, ChannelD, Ext, Aux genericps.ThresholdDirection
	}
	// TriggerDesc fully describes the trigger configuration sent to the scope.
	TriggerDesc struct {
		Enabled                bool
		TriggerADC             int16
		LowerTriggerADC        int16
		HysteresisADC          uint16
		LowerHysteresisADC     uint16
		UpperHysteresis        int32
		LowerHysteresis        int32
		Source                 genericps.ChannelId
		ThresholdDirection     genericps.ThresholdDirection
		ThresholdMode          genericps.ThresholdModeId
		Mode                   TriggerModes
		Type                   TriggerTypes
		Mv                     int32
		LowerMv                int32
		ComplexProperties      []genericps.TriggerChannelProperties
		ComplexConditions      []genericps.TriggerConditions
		ComplexDirections      []TriggerDirections
		IntervalType           genericps.PulseWidthType
		IntervalTimeLower      float64
		IntervalTimeUpper      float64
		XOffset                float64
		AutoTriggerMs          int16
		DigitalTriggerEnabled  bool
		DigitalDirections      []genericps.DigitalChannelDirections
		DigitalAnalogOperand   genericps.TriggerOperand
		DigitalChannelsOperand genericps.TriggerOperand
		EtsInterleave          int16
		EtsCycles              int16
	}
	// TriggerDescMsg carries a new trigger setting; Done is signalled once applied.
	TriggerDescMsg struct {
		TriggerDesc
		Done chan struct{}
	}
	// getTriggerMsg is the request/reply handshake used by the acquisition
	// goroutine to fetch the latest trigger settings from triggerMonitor.
	getTriggerMsg struct {
		triggerSettings *TriggerDesc
		newSettings     chan bool // true if settings changed since the last request
	}

	// getNumOfEnabledChMsg is used to query the number of enabled analog channels.
	getNumOfEnabledChMsg struct {
		n chan int
	}

	// DigitalPortMsg carries the settings of one digital port (MSO).
	DigitalPortMsg struct {
		Port     genericps.DigitalPort
		Settings settings.DigitalPortSettings
	}

	// getDigitalPortMsg is the request/reply handshake for digital port settings.
	getDigitalPortMsg struct {
		portSettings *DigitalPortMsg
		newSettings  chan bool
	}
	// getChannelMsg is the request/reply handshake for analog channel settings.
	getChannelMsg struct {
		channelSettings *genericps.SetChannelMsg
		newSettings     chan bool
	}

	// getInterpolationModeMsg is the request/reply handshake for the
	// interpolation mode (linear, sinc, ...).
	getInterpolationModeMsg struct {
		ipMode     settings.InterpolationType
		newSetting chan bool
	}

	// getScopeScreenWidthMsg is the request/reply handshake for the display width.
	getScopeScreenWidthMsg struct {
		width      int32
		newSetting chan bool
	}

	// GeneratorDesc describes the signal/arbitrary waveform generator setup.
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
		ArbitraryWaveform                                   []int16
		StartDeltaPhase                                     uint32
		StopDeltaPhase                                      uint32
		DeltaPhaseIncrement                                 uint32
		IndexMode                                           genericps.IndexMode
		SpiDataValue                                        uint32
		I2cAddressValue                                     uint32
	}

	// GeneratorDescMsg carries new generator settings; Done is signalled once applied.
	GeneratorDescMsg struct {
		GeneratorDesc
		Done chan struct{}
	}

	// getGeneratorMsg is the request/reply handshake for generator settings.
	getGeneratorMsg struct {
		generatorSettings *GeneratorDesc
		newSetting        chan bool
	}

	// PscDesc is the central controller of a connected scope (real or demo).
	// It owns the acquisition state machine, the settings channels and the
	// sample buffers. Settings are changed through the Set*Ch channels and
	// picked up by the acquisition goroutine through the get*Ch handshakes.
	PscDesc struct {
		Con *genericps.Connection

		hardwareMu   sync.Mutex
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

		SetDigitalPortCh chan *DigitalPortMsg
		getDigitalPortCh chan *getDigitalPortMsg
		getDigitalPort   getDigitalPortMsg

		SetInterpolationModeCh chan settings.InterpolationType
		getInterpolationModeCh chan *getInterpolationModeMsg
		getInterpolationMode   getInterpolationModeMsg

		ResolutionMode atomic.Int32 // genericps.RatioMode

		SetGeneratorCh chan *GeneratorDescMsg
		SetDemoGenCh   chan *GeneratorDescMsg
		getGeneratorCh chan *getGeneratorMsg
		getGenerator   getGeneratorMsg

		SetScopeScreenWidthCh chan int32
		getScopeScreenWidthCh chan *getScopeScreenWidthMsg
		getScopeScreenWidth   getScopeScreenWidthMsg

		triggerSetting              TriggerDesc
		chEnabled                   []atomic.Bool
		digitalPortsEnabled         [2]atomic.Bool
		triggerTimeOffset           int64     // sub-sample trigger time offset in femtoseconds
		receiveBuffer               [][]int16 // raw data buffer, only for real channel
		receiveBufferMin            [][]int16 // raw data buffer for min values when in ED mode
		digitalReceiveBuffer        [][]int16
		digitalReceiveBufferMin     [][]int16
		displayBuffer               [][]float32 // signal stored in mv
		EtsInBuffer                 []int64
		overSample                  int16
		SamplingTimeInterval        float64 // seconds between samples
		lastTriggerSamplingInterval float64 // sampling interval when the trigger was last sent
		initialTriggerSet           bool    // false until the trigger was sent at least once
		SampleCountRequired         uint64  // total samples to acquire (incl. sinc padding)
		NPre, NPro                  uint64  // samples before / after the trigger point
		XRoundError                 float64 // time rounding error of the trigger position
		timeBase                    uint64
		ipmode                      settings.InterpolationType
		numOfSamplesAcquired        uint64
		downSampleRatioMode         genericps.RatioMode
		downSampleRatio             uint64
		maxValue                    int32
		maxScreenTime               float64
		scopeScreenWidth            float64
		timeBaseDec                 uint64
		minValue                    int32
		RefreshCallback             func(buffers [][]int16, buffersMin [][]int16, digitalBuffers [][]int16, startTimeOffset int64, xRoundError float64, samplingTimeInterval float64)
		RefreshEtsCallback          func(buffers [][]int16, etsOutBuffer []int64, xRoundError float64, samplingTimeInterval float64)
		BufferCallback              func(size int)
		EtsBufferCallback           func(size int)
		DisplayStatus               func(s string, errorType ScopeError)
		refreshTime                 time.Time
		Info                        string
		ScopeModel                  ScopeType
		MaxSamplingRate             uint32
		StreamEnabled               atomic.Bool
	}
)

// Equals reports whether two generator descriptions are identical,
// including the arbitrary waveform contents.
func (g *GeneratorDesc) Equals(other *GeneratorDesc) bool {
	if g.OffsetVoltage != other.OffsetVoltage || g.PkToPK != other.PkToPK ||
		g.WaveType != other.WaveType || g.StartFrequency != other.StartFrequency ||
		g.StopFrequency != other.StopFrequency || g.Increment != other.Increment ||
		g.DwellTime != other.DwellTime || g.SweepType != other.SweepType ||
		g.Operation != other.Operation || g.Shots != other.Shots ||
		g.Sweeps != other.Sweeps || g.TriggerType != other.TriggerType ||
		g.TriggerSource != other.TriggerSource || g.ExtInThreshold != other.ExtInThreshold ||
		g.Phase != other.Phase || g.Channel != other.Channel || g.On != other.On ||
		g.StartDeltaPhase != other.StartDeltaPhase || g.StopDeltaPhase != other.StopDeltaPhase ||
		g.DeltaPhaseIncrement != other.DeltaPhaseIncrement || g.IndexMode != other.IndexMode ||
		g.SpiDataValue != other.SpiDataValue || g.I2cAddressValue != other.I2cAddressValue {
		return false
	}
	if len(g.ArbitraryWaveform) != len(other.ArbitraryWaveform) {
		return false
	}
	for i, v := range g.ArbitraryWaveform {
		if v != other.ArbitraryWaveform[i] {
			return false
		}
	}
	return true
}

// Shutdown signals all monitor goroutines launched by NewControl to exit.
// It is safe to call multiple times.
func (psControl *PscDesc) Shutdown() {
	psControl.shutdownOnce.Do(func() {
		close(psControl.shutdownCh)
	})
}

// NewControl creates a controller for the given connection, allocates all
// communication channels and starts the state machine and monitor goroutines.
func NewControl(con *genericps.Connection) *PscDesc {
	slog.Debug("NewControl")
	psControl := &PscDesc{Con: con}
	psControl.shutdownCh = make(chan struct{})
	psControl.StreamEnabled.Store(true)
	psControl.stateChannel = make(chan state, 1)
	psControl.restartChannel = make(chan struct{}, 1) // non blocking
	psControl.stopChannel = make(chan struct{}, 1)    // non blocking

	psControl.SetGeneratorCh = make(chan *GeneratorDescMsg)
	psControl.SetDemoGenCh = make(chan *GeneratorDescMsg)
	psControl.getGeneratorCh = make(chan *getGeneratorMsg)
	psControl.getGenerator.newSetting = make(chan bool)

	psControl.SetChannelCh = make(chan *settings.ChSettings, 16)
	psControl.getChannelCh = make(chan *getChannelMsg)
	psControl.getChannel.newSettings = make(chan bool)
	psControl.getNumOfEnabledCh = make(chan *getNumOfEnabledChMsg)
	psControl.getNumOfEnabled.n = make(chan int)

	psControl.SetDigitalPortCh = make(chan *DigitalPortMsg, 16)
	psControl.getDigitalPortCh = make(chan *getDigitalPortMsg)
	psControl.getDigitalPort.newSettings = make(chan bool)

	psControl.SetInterpolationModeCh = make(chan settings.InterpolationType)
	psControl.getInterpolationModeCh = make(chan *getInterpolationModeMsg)
	psControl.getInterpolationMode.newSetting = make(chan bool)

	psControl.SetScopeScreenWidthCh = make(chan int32)
	psControl.getScopeScreenWidthCh = make(chan *getScopeScreenWidthMsg)
	psControl.getScopeScreenWidth.newSetting = make(chan bool)

	psControl.SetTriggerCh = make(chan *TriggerDescMsg)
	psControl.getTriggerCh = make(chan *getTriggerMsg)
	psControl.getTrigger.triggerSettings = &psControl.triggerSetting
	psControl.getTrigger.newSettings = make(chan bool)
	go psControl.stateMachine()
	go psControl.triggerMonitor()
	go psControl.generatorMonitor()
	go psControl.demoGeneratorMonitor()
	go psControl.interpolationMonitor()
	go psControl.digitalPortMonitor()
	return psControl
}

// GetAnalogueOffset returns the maximum and minimum analogue offset voltage
// allowed for the given voltage range and coupling.
func (psControl *PscDesc) GetAnalogueOffset(voltageRange int,
	coupling genericps.Coupling) (maximumVoltage, minimumVoltage float32, err error) {
	if psControl.Con == nil {
		return 0, 0, errors.New("no connection")
	}
	maximumVoltage, minimumVoltage, err =
		psControl.Con.GetAnalogueOffset(voltageRange, coupling)
	return
}

// setChannel fetches pending channel settings one by one and applies them to
// the scope until no more new settings are available.
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

// setTrigger sends the trigger to the scope if the settings changed, if a
// time dependent trigger needs rescaling after a sampling interval change, or
// if no trigger has been sent yet.
func (psControl *PscDesc) setTrigger() (err error) {
	psControl.getTriggerCh <- &psControl.getTrigger   // ask for data
	newSettings := <-psControl.getTrigger.newSettings // wait for data

	samplingIntervalChanged := psControl.SamplingTimeInterval != psControl.lastTriggerSamplingInterval
	timeDependentTrigger := psControl.triggerSetting.Type == Interval || psControl.triggerSetting.Type == PulseWidth || psControl.triggerSetting.Type == Dropout || psControl.triggerSetting.Type == WindowDropout || psControl.triggerSetting.Type == WindowPulseWidth || psControl.triggerSetting.Type == RiseFall
	if newSettings || (samplingIntervalChanged && timeDependentTrigger) || !psControl.initialTriggerSet {
		err = psControl.sendTrigger() // 			   send to the scope
		if err != nil {
			slog.Error("setTrigger", "error", err)
			return
		}
		psControl.initialTriggerSet = true
		psControl.lastTriggerSamplingInterval = psControl.SamplingTimeInterval
	}
	return
}

// setIpMode fetches the current interpolation mode.
func (psControl *PscDesc) setIpMode() {
	psControl.getInterpolationModeCh <- &psControl.getInterpolationMode
	if <-psControl.getInterpolationMode.newSetting { // wait for data
		psControl.ipmode = psControl.getInterpolationMode.ipMode
	}
}

// setEverything applies interpolation, generator, channel and digital port
// settings; used when (re)starting an acquisition.
func (psControl *PscDesc) setEverything() (err error) {
	psControl.setIpMode()
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
	err = psControl.setDigitalPort()
	if err != nil {
		slog.Error("setDigitalPort", "error", err)
		return
	}
	return
}

// sendTrigger dispatches to the sender matching the configured trigger type.
// When the trigger is disabled, a digital trigger or a simple (auto) trigger is used.
func (psControl *PscDesc) sendTrigger() (err error) {
	if !psControl.triggerSetting.Enabled {
		if psControl.hasActiveDigitalTrigger() {
			return psControl.sendDigitalTrigger()
		}
		return psControl.sendSimpleTrigger()
	}

	switch psControl.triggerSetting.Type {
	case Simple:
		err = psControl.sendSimpleTrigger()
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
	case WindowPulseWidth:
		err = psControl.sendWindowPulseWidthTrigger()
	case Dropout:
		err = psControl.sendDropOutTrigger()
	case WindowDropout:
		err = psControl.sendWindowDropoutTrigger()
	case Runt:
		err = psControl.sendRuntTrigger()
	case RiseFall:
		err = psControl.sendRiseFallTrigger()
	}

	return
}

// SetScopeScreenWidth sets the display width in pixels (capped at 1e6) and
// restarts the acquisition if it changed.
func (psControl *PscDesc) SetScopeScreenWidth(w float64) {
	if w > 1000000 {
		w = 1000000
	}
	if psControl.scopeScreenWidth != w {
		psControl.scopeScreenWidth = w
		psControl.requestRestart()
	}
}

// SetMaxScreenTime sets the maximum displayed time span in seconds and
// restarts the acquisition if it changed.
func (psControl *PscDesc) SetMaxScreenTime(t float64) {
	if psControl.maxScreenTime != t {
		psControl.maxScreenTime = t
		psControl.requestRestart()
	}
}

// SuggestSampleCount sets the required sample count and restarts the
// acquisition if it changed.
func (psControl *PscDesc) SuggestSampleCount(sc uint64) {
	if psControl.SampleCountRequired != sc {
		psControl.SampleCountRequired = sc
		psControl.requestRestart()
	}
}

// numberOfEnabledChannels counts enabled analog channels plus one if any
// digital port is enabled.
func (psControl *PscDesc) numberOfEnabledChannels() (n int) {
	psControl.getNumOfEnabledCh <- &psControl.getNumOfEnabled
	n = <-psControl.getNumOfEnabled.n
	if psControl.digitalPortsEnabled[0].Load() || psControl.digitalPortsEnabled[1].Load() {
		n++
	}
	return
}

// numberOfEnabledAnalogChannels counts enabled analog channels only.
func (psControl *PscDesc) numberOfEnabledAnalogChannels() (n int) {
	psControl.getNumOfEnabledCh <- &psControl.getNumOfEnabled
	return <-psControl.getNumOfEnabled.n
}

// NumberOfEnabledAnalogChannels is the exported form of numberOfEnabledAnalogChannels.
func (psControl *PscDesc) NumberOfEnabledAnalogChannels() int {
	return psControl.numberOfEnabledAnalogChannels()
}

// SetDigitalPortEnabled marks digital port 0 or 1 as enabled/disabled.
func (psControl *PscDesc) SetDigitalPortEnabled(port int, enabled bool) {
	if port >= 0 && port < 2 {
		psControl.digitalPortsEnabled[port].Store(enabled)
	}
}

// NewChannels starts the channel state machine and allocates the raw and
// display buffers for the given number of analog channels and 2 digital ports.
func (psControl *PscDesc) NewChannels(numberOfChannels int) {
	go psControl.channelStateMachine(numberOfChannels)
	psControl.receiveBuffer = make([][]int16, numberOfChannels)
	psControl.receiveBufferMin = make([][]int16, numberOfChannels)
	psControl.digitalReceiveBuffer = make([][]int16, 2)
	psControl.digitalReceiveBufferMin = make([][]int16, 2)
	psControl.displayBuffer = make([][]float32, numberOfChannels)
	for i := 0; i < numberOfChannels; i++ {
		psControl.receiveBuffer[i] = make([]int16, initialBufferSize)
		psControl.receiveBufferMin[i] = make([]int16, initialBufferSize)
		psControl.displayBuffer[i] = make([]float32, initialBufferSize)
	}
	for i := 0; i < 2; i++ {
		psControl.digitalReceiveBuffer[i] = make([]int16, initialBufferSize)
		psControl.digitalReceiveBufferMin[i] = make([]int16, initialBufferSize)
	}
}

// ChannelRanges returns the voltage ranges supported by the given channel.
// Some 6000a variants do not report ranges, so a fixed list is used for them.
func (psControl *PscDesc) ChannelRanges(chIndex genericps.ChannelId) (ranges []int32,
	err error) {
	if psControl == nil || psControl.Con == nil {
		return nil, nil
	}
	allowedRanges := make([]int32, 32)
	length, err := psControl.Con.GetChannelInformation(genericps.ChannelInfoRanges,
		0, allowedRanges, chIndex)
	if err != nil {
		if strings.Contains(err.Error(), "Not supported") {
			if strings.Contains(psControl.Info, "6000a") || strings.Contains(psControl.Info, "64AEM") || strings.Contains(psControl.Info, "ps6000a") {
				return []int32{
					int32(genericps.Range_10mv), int32(genericps.Range_20mv), int32(genericps.Range_50mv),
					int32(genericps.Range_100mv), int32(genericps.Range_200mv), int32(genericps.Range_500mv),
					int32(genericps.Range_1v), int32(genericps.Range_2v), int32(genericps.Range_5v),
					int32(genericps.Range_10v), int32(genericps.Range_20v),
				}, nil
			}
		}
		slog.Error("Get ch info", "err", err)
		return
	}
	return allowedRanges[:length], err
}

// UnitVariantInfo returns the variant (model) string of the connected unit.
func (psControl *PscDesc) UnitVariantInfo() (info string, err error) {
	info, err = psControl.Con.GetUnitInfo(genericps.PicoVariantInfo)
	return
}

// UnitBatchAndSerialInfo returns the batch and serial number of the unit.
func (psControl *PscDesc) UnitBatchAndSerialInfo() (info string, err error) {
	info, err = psControl.Con.GetUnitInfo(genericps.PicoBatchAndSerial)
	return
}

// MinMaxValues queries the ADC min/max values and caches them in the controller.
func (psControl *PscDesc) MinMaxValues() (min, max int32, err error) {
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

// stopHardware stops the running acquisition on the device, serialized by hardwareMu.
func (psControl *PscDesc) stopHardware() (err error) {
	psControl.hardwareMu.Lock()
	defer psControl.hardwareMu.Unlock()
	if psControl.Con == nil {
		return nil
	}
	return psControl.Con.Stop()
}

// Stop requests the acquisition to stop. It never blocks.
func (psControl *PscDesc) Stop() (err error) {
	select {
	case psControl.stopChannel <- struct{}{}:
	default:
		// A stop request is already pending.
		slog.Debug("Stop request already pending")
	}
	return
}

// SetETSMode stops the current acquisition and switches to ETS block mode.
func (psControl *PscDesc) SetETSMode() (err error) {
	_ = psControl.Stop()
	select {
	case <-psControl.stateChannel:
	default:
	}
	select {
	case psControl.stateChannel <- etsBlockMode:
	case <-time.After(startTimeout):
		err = fmt.Errorf("Could not start ETS mode")
	}
	return
}

// SetBlockMode stops the current acquisition and switches to block mode.
func (psControl *PscDesc) SetBlockMode() (err error) {
	_ = psControl.Stop()
	select {
	case <-psControl.stateChannel:
	default:
	}
	select {
	case psControl.stateChannel <- blockMode:
	case <-time.After(startTimeout):
		err = fmt.Errorf("Could not start block mode")
	}
	return
}
