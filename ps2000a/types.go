//go:build !noscope

package ps2000a

type (
	PwqConditions struct {
		ChannelA TriggerState
		ChannelB TriggerState
		ChannelC TriggerState
		ChannelD TriggerState
		External TriggerState
		Aux      TriggerState
		Digital  TriggerState
	}
	Pwq struct {
		Conditions  PwqConditions
		NConditions int16
		Direction   ThresholdDirection
		Lower       uint32
		Upper       uint32
		Type        PulseWidthType
	}

	ChannelDesc struct {
		CoupleType Coupling
		Range      RangeEnum
		Enabled    bool
		Inverted   bool
		Offset     float32
	}
	PsDesc struct {
		handle int16
	}
	BlockReady     func(handle int16, status int, param any)
	DataReady      func(handle int16, status int, noOfSamples uint32, overflow int16, param any)
	StreamingReady func(handle int16, noOfSamples int32, startIndex uint32, overflow int16,
		triggeredAt uint32, triggered, autoStop int16, param any) (err error)

	// Picoscope2000Handler interface {
	// 	Handle() int16
	// 	OpenUnit(serial string) (err error)
	// 	OpenUnitAsync(serial string) (err error)
	// 	OpenUnitProgress() (retHandle, progressPercent, complete int16, err error)
	// 	EnumerateUnits(bufferLen int16) (count int16, serials string, serialLth int16, err error)
	// 	CloseUnit() (err error)
	// 	FlashLed(start int16) (err error)
	// 	PingUnit() (err error)
	// 	GetAnalogueOffset(voltageRange int, coupling Coupling) (maximumVoltage, minimumVoltage float32, err error)
	// 	GetChannelInformation(info int16, probe int32, ranges []int32, channels ChannelId) (lengthOfRanges int32, err error)
	// 	GetMaxDownSampleRatio(noOfUnaggregatedSamples uint32, downSampleRatioMode RatioMode, segmentIndex int32) (maxDownSampleRatio uint32, err error)
	// 	GetMaxSegments() (maxSegments uint32, err error)
	// 	GetNoOfCaptures() (nCaptures uint32, err error)
	// 	GetNoOfProcessedCaptures() (nCaptures uint32, err error)
	// 	GetStreamingLatestValues(lpStreamingReadyGoPar StreamingReady, param any) (err error)
	// 	GetTimebase(timeBase uint32, noOfSamples int32, overSample int16, segmentIndex uint32) (timeIntervalNanoseconds, maxSamples int32, err error)
	// 	GetTimebase2(timeBase uint32, noOfSamples int32, overSample int16, segmentIndex uint32) (timeIntervalNanoseconds float32, maxSamples int32, err error)
	// 	GetTriggerTimeOffset(segmentIndex uint32) (timeUpper, timeLower uint32, timeUnits TimeUnits, err error)
	// 	GetTriggerTimeOffset64(segmentIndex uint32) (time int64, timeUnits TimeUnits, err error)
	// 	GetUnitInfo(info PicoInfo) (infoString string, err error)
	// 	GetValues(startIndex, reqNoOfSamples, downSampleRatio uint32,
	// 		downSampleRatioMode RatioMode, segmentIndex uint32) (noOfSamples uint32, overflow int16, err error)
	// 	GetValuesAsync(startIndex, noOfSamples, downSampleRatio uint32,
	// 		downSampleRatioMode RatioMode, lpDataReadyGoPar DataReady, segmentIndex uint32,
	// 		param any) (err error)
	// 	GetValuesBulk(reqNoOfSamples uint32, fromSegmentIndex, toSegmentIndex, downSampleRatio uint32,
	// 		downSampleRatioMode RatioMode, overflow []int16) (noSamples uint32, err error)
	// 	GetValuesOverlapped(startIndex, reqNoOfSamples, downSampleRatio uint32,
	// 		downSampleRatioMode RatioMode, segmentIndex uint32, overflow []int16) (noSamples uint32, err error)
	// 	GetValuesOverlappedBulk(startIndex, reqNoOfSamples, downSampleRatio uint32,
	// 		downSampleRatioMode RatioMode, fromSegmentIndex, toSegmentIndex uint32, overflow []int16) (noSamples uint32, err error)
	// 	GetValuesTriggerTimeOffsetBulk(timesUpper, timesLower []uint32, timeUnits []TimeUnits,
	// 		fromSegmentIndex, toSegmentIndex uint32) (err error)
	// 	GetValuesTriggerTimeOffsetBulk64(times []int64, timeUnits []TimeUnits,
	// 		fromSegmentIndex, toSegmentIndex uint32) (err error)
	// 	HoldOff(holdOff uint64, holdOffType HoldOffType) (err error)
	// 	LsReady() (ready int16, err error)
	// 	MaximumValue() (value int16, err error)
	// 	MemorySegments(nSegments uint32) (nMaxSamples int32, err error)
	// 	MinimumValue() (value int16, err error)
	// 	NoOfStreamingValues() (noOfValues uint32, err error)
	// 	QueryOutputEdgeDetect() (state int16, err error)
	// 	RunBlock(noOfPreTriggerSamples, noOfPostTriggerSamples int32,
	// 		timeBase uint32, overSample int16, segmentIndex uint32, lpBlockReadyGoPar BlockReady,
	// 		param any) (timeIndisposedMs int32, err error)
	// 	RunStreaming(reqSampleInterval uint32, sampleIntervalTimeUnits TimeUnits,
	// 		maxPreTriggerSamples, maxPostTriggerSamples uint32,
	// 		autoStop bool, downSampleRatio uint32, downSampleRatioMode RatioMode,
	// 		overviewBufferSize uint32) (sampleInterval uint32, err error)
	// 	SetChannel(channel ChannelId, enabled bool, couplingType Coupling, voltageRange RangeEnum, analogOffset float32) (err error)
	// 	SetDataBuffer(ch ChannelId, bufferIn []int16, segmentIndex uint32,
	// 		mode RatioMode) (err error)
	// 	SetDataBuffers(ch ChannelId, bufferMax, bufferMin []int16, segmentIndex uint32, mode RatioMode) (err error)
	// 	SetUnscaledDataBuffers(ch ChannelId, bufferMax, bufferMin []int16, segmentIndex uint32, mode RatioMode) (err error)
	// 	SetDigitalAnalogTriggerOperand(operand TriggerOperand) (err error)
	// 	SetDigitalPort(port DigitalPort, enabled bool, logiclevel int16) (err error)
	// 	SetEts(mode EtsMode, etsCycles int16, etsInterLeave int16) (sampleTimePicoseconds int32, err error)
	// 	SetEtsTimeBuffer(buffer []int64) (err error)
	// 	SetEtsTimeBuffers(timeUpper, timeLower []uint32) (err error)
	// 	SetNoCaptures(nCaptures uint32) (err error)
	// 	SetOutputEdgeDetect(state int16) (err error)
	// 	SetPulseWidthDigitalPortProperties(digitalDirections []DigitalChannelDirections) (err error)
	// 	SetPulseWidthQualifier(conditions []PwqConditions, direction ThresholdDirection, lower, upper uint32,
	// 		pwType PulseWidthType) (err error)
	// 	SetSigGenArbitrary(offsetVoltage int32, pkToPK uint32,
	// 		startDeltaPhase, stopDeltaPhase, deltaPhaseIncrement, dwellCount uint32,
	// 		arbitraryWaveform []int16, sweepType SweepTypeEnum, operation ExtraOperations,
	// 		indexMode IndexMode, shots, sweeps uint32, triggerType SigGenTrigType,
	// 		triggerSource SigGenTrigSource, extInThreshold int16) (err error)
	// 	SigGenArbitraryMinMaxValues() (minArbitraryWaveformValue, maxArbitraryWaveformValue int16,
	// 		minArbitraryWaveformSize, maxArbitraryWaveformSize uint32, err error)
	// 	SetSigGenBuiltIn(offsetVoltage int32, pkToPK uint32, waveType WaveTypeEnum,
	// 		startFrequency, stopFrequency, increment, dwellTime float32, sweepType SweepTypeEnum,
	// 		operation ExtraOperations, shots, sweeps uint32, triggerType SigGenTrigType,
	// 		triggerSource SigGenTrigSource, extInThreshold int16) (err error)
	// 	SetSigGenBuiltInV2(offsetVoltage int32, pkToPK uint32, waveType WaveTypeEnum,
	// 		startFrequency, stopFrequency, increment, dwellTime float64, sweepType SweepTypeEnum,
	// 		operation ExtraOperations, shots, sweeps uint32, triggerType SigGenTrigType,
	// 		triggerSource SigGenTrigSource, extInThreshold int16) (err error)
	// 	SetSigGenPropertiesArbitrary(offsetVoltage int32,
	// 		startDeltaPhase, stopDeltaPhase, deltaPhaseIncrement, dwellCount uint32,
	// 		sweepType SweepTypeEnum, operation ExtraOperations,
	// 		indexMode IndexMode, shots, sweeps uint32, triggerType SigGenTrigType,
	// 		triggerSource SigGenTrigSource, extInThreshold int16) (err error)
	// 	SetSigGenPropertiesBuiltIn(offsetVoltage int32,
	// 		startFrequency, stopFrequency, increment, dwellTime float64,
	// 		sweepType SweepTypeEnum,
	// 		shots, sweeps uint32, triggerType SigGenTrigType,
	// 		triggerSource SigGenTrigSource, extInThreshold int16) (err error)
	// 	SetSimpleTrigger(enable bool, source ChannelId, threshold int16,
	// 		direction ThresholdDirection, delay uint32, autoTriggerMs int16) (err error)
	// 	SetTriggerChannelConditions(triggerConditions []TriggerConditions) (err error)
	// 	SetTriggerChannelDirections(channelA, channelB, channelC, channelD, ext, aux ThresholdDirection) (err error)
	// 	SetTriggerChannelProperties(channelProperties []TriggerChannelProperties, auxOutputEnable bool,
	// 		autoTriggerMs int32) (err error)
	// 	SetTriggerDelay(delay uint32) (err error)
	// 	SetTriggerDigitalPortProperties(digitalDirections []DigitalChannelDirections) (err error)
	// 	SigGenFrequencyToPhase(frequency float64, indexMode IndexMode, bufferLength uint32) (phase uint32, err error)
	// 	Stop() (err error)
	// 	TriggerOrPulseWidthQualifierEnabled() (triggerEnabled, pulseWidthQualifierEnabledint16 int16, err error)
	// 	SigGenSoftwareControl(state int16) (err error)
	// }
	TriggerChannelProperties struct {
		ThresholdUpper           int16
		ThresholdUpperHysteresis uint16
		ThresholdLower           int16
		ThresholdLowerHysteresis uint16
		Channel                  ChannelId
		ThresholdMode            ThresholdModeId
	}
	TriggerConditions struct {
		ChannelA            TriggerState
		ChannelB            TriggerState
		ChannelC            TriggerState
		ChannelD            TriggerState
		External            TriggerState
		Aux                 TriggerState
		PulseWidthQualifier TriggerState
		Digital             TriggerState
	}

	TriggerDirections struct {
		ChannelA ThresholdDirection
		ChannelB ThresholdDirection
		ChannelC ThresholdDirection
		ChannelD ThresholdDirection
		Ext      ThresholdDirection
		Aux      ThresholdDirection
	}

	DigitalChannelDirections struct {
		Channel   DigitalChannel
		Direction DigitalDirection
	}
)
