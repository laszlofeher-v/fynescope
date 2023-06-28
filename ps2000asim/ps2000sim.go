package ps2000asim

import (
	"fmt"
	"psgo/ps2000a"
	"time"
)

const timeout = 8 * time.Second

type returnStatus int

const (
	normal returnStatus = iota
	faulty
	timeoutNormal
	timeoutfaulty
)

var behaviour returnStatus

type (
	SimDesc struct {
		handle int16
	}
)

func boolToint16(b bool) int16 {
	if b {
		return int16(1)
	}
	return int16(0)
}

func nextInt16() func() int16 {
	i := int16(0)
	return func() int16 {
		i++
		return i
	}
}

var uniqueHandle = nextInt16()

func (ps *SimDesc) Handle() int16 {
	return ps.handle
}

func (ps *SimDesc) EnumerateUnits() (count int16, serials string, serialLth int16, err error) {
	nf := func() {
		count = 1
		serials = "AQ005/139,VDR61/356,ZOR14/107"
		serialLth = int16(len(serials))
		err = nil
	}
	ff := func() {
		err = fmt.Errorf("EnumerateUnits error")
	}
	switch behaviour {
	case normal:
		nf()
	case faulty:
		ff()
	case timeoutNormal:
		time.Sleep(timeout)
		nf()
	case timeoutfaulty:
		time.Sleep(timeout)
		ff()
	}
	return
}

func (ps *SimDesc) OpenUnit(serial string) (err error) {
	fmt.Println("OpenUnit called")
	nf := func() {
		ps.handle = uniqueHandle()
		err = nil
	}
	ff := func() {
		err = fmt.Errorf("OpenUnit error")
	}
	switch behaviour {
	case normal:
		nf()
	case faulty:
		ff()
	case timeoutNormal:
		time.Sleep(timeout)
		nf()
	case timeoutfaulty:
		time.Sleep(timeout)
		ff()
	}
	return
}
func (ps *SimDesc) OpenUnitAsync(serial string) (err error) {
	nf := func() {
		ps.handle = uniqueHandle()
		err = nil
	}
	ff := func() {
		err = fmt.Errorf("OpenUnit error")
	}
	switch behaviour {
	case normal:
		nf()
	case faulty:
		ff()
	case timeoutNormal:
		time.Sleep(timeout)
		nf()
	case timeoutfaulty:
		time.Sleep(timeout)
		ff()
	}
	return
}

func (ps *SimDesc) CloseUnit() (err error) {
	return
}

func (ps *SimDesc) GetUnitInfo(info ps2000a.PicoInfo) (infoString string, err error) {
	return
}

func (ps *SimDesc) FlashLed(start int16) (err error) {
	return
}

func lpDataReadyGo(handle int16, status int, noOfSamples uint32, overflow int16, param interface{}) {
	return
}

func (ps *SimDesc) GetValuesAsync(startIndex, noOfSamples, downSampleRatio uint32,
	downSampleRatioMode ps2000a.RatioMode, lpDataReadyGoPar ps2000a.DataReady, segmentIndex uint32,
	param interface{}) (err error) {
	return
}

func (ps *SimDesc) GetValues(startIndex, reqNoOfSamples, downSampleRatio uint32,
	downSampleRatioMode ps2000a.RatioMode, segmentIndex uint32) (noOfSamples uint32, overflow int16, err error) {
	return
}

func (ps *SimDesc) GetValuesBulk(reqNoOfSamples uint32, fromSegmentIndex, toSegmentIndex, downSampleRatio uint32,
	downSampleRatioMode ps2000a.RatioMode, overflow []int16) (noSamples uint32, err error) {
	return
}

func (ps *SimDesc) GetValuesOverlapped(startIndex, reqNoOfSamples, downSampleRatio uint32,
	downSampleRatioMode ps2000a.RatioMode, segmentIndex uint32, overflow []int16) (noSamples uint32, err error) {
	return
}

func (ps *SimDesc) GetValuesOverlappedBulk(startIndex, reqNoOfSamples, downSampleRatio uint32,
	downSampleRatioMode ps2000a.RatioMode, fromSegmentIndex, toSegmentIndex uint32, overflow []int16) (noSamples uint32, err error) {
	return
}

func (ps *SimDesc) GetAnalogueOffset(voltageRange int, coupling ps2000a.Coupling) (maximumVoltage, minimumVoltage float32, err error) {
	return
}

func (ps *SimDesc) GetChannelInformation(info int16, probe int32, ranges []int32, channels ps2000a.ChannelId) (lengthOfRanges int32, err error) {
	lengthOfRanges = int32(len(ranges))
	return
}

func (ps *SimDesc) GetMaxDownSampleRatio(noOfUnaggregatedSamples uint32, downSampleRatioMode ps2000a.RatioMode, segmentIndex int32) (maxDownSampleRatio uint32, err error) {
	return
}

func (ps *SimDesc) GetMaxSegments() (maxSegments uint32, err error) {
	return
}

func (ps *SimDesc) GetNoOfCaptures() (nCaptures uint32, err error) {
	return
}

func (ps *SimDesc) GetNoOfProcessedCaptures() (nCaptures uint32, err error) {
	return
}

func lpStreamingReadyGo(handle int16, noOfSamples int32, startIndex uint32, overflow int16,
	triggeredAt uint32, triggered, autoStop int16, param interface{}) {
	return
}

func (ps *SimDesc) GetStreamingLatestValues(lpStreamingReadyGoPar ps2000a.StreamingReady, param interface{}) (err error) {
	return
}

func (ps *SimDesc) GetTimebase(timeBase uint32, noOfSamples int32, overSample int16, segmentIndex uint32) (timeIntervalNanoseconds, maxSamples int32, err error) {
	return
}

func (ps *SimDesc) GetTimebase2(timeBase uint32, noOfSamples int32, overSample int16, segmentIndex uint32) (timeIntervalNanoseconds float32, maxSamples int32, err error) {
	return
}

func (ps *SimDesc) SetChannel(channel ps2000a.ChannelId, enabled bool, couplingType ps2000a.Coupling, voltageRange ps2000a.RangeEnum, analogOffset float32) (err error) {
	fmt.Println("SetChannel called")
	return
}

func (ps *SimDesc) MaximumValue() (value int16, err error) {
	return
}

func (ps *SimDesc) MinimumValue() (value int16, err error) {
	return
}

func (ps *SimDesc) SetSimpleTrigger(enable bool, source ps2000a.ChannelId, threshold int16,
	direction ps2000a.ThresholdDirection, delay uint32, autoTriggerMs int16) (err error) {
	return
}

func (ps *SimDesc) SetDataBuffer(ch ps2000a.ChannelId, bufferIn []int16, segmentIndex uint32,
	mode ps2000a.RatioMode) (err error) {
	return
}

func (ps *SimDesc) SetDataBuffers(ch ps2000a.ChannelId, bufferMax, bufferMin []int16, segmentIndex uint32, mode ps2000a.RatioMode) (err error) {
	return
}

func (ps *SimDesc) SetUnscaledDataBuffers(ch ps2000a.ChannelId, bufferMax, bufferMin []int16, segmentIndex uint32, mode ps2000a.RatioMode) (err error) {
	return
}

func (ps *SimDesc) SetEtsTimeBuffer(buffer []int64) (err error) {
	return
}

func (ps *SimDesc) SetEtsTimeBuffers(timeUpper, timeLower []uint32) (err error) {
	return
}

func (ps *SimDesc) SetEts(mode ps2000a.EtsMode, etsCycles int16, etsInterLeave int16) (sampleTimePicoseconds int32, err error) {
	return
}

func (ps *SimDesc) RunStreaming(reqSampleInterval uint32, sampleIntervalTimeUnits ps2000a.TimeUnits,
	maxPreTriggerSamples, maxPostTriggerSamples uint32,
	autoStop bool, downSampleRatio uint32, downSampleRatioMode ps2000a.RatioMode,
	overviewBufferSize uint32) (sampleInterval uint32, err error) {
	return
}

var regLpBlockReadyGo ps2000a.BlockReady // registered go callback function

func lpBlockReadyGo(handle int16, status int, noOfSamples uint32, overflow int16, param interface{}) {
	return
}

func (ps *SimDesc) RunBlock(noOfPreTriggerSamples, noOfPostTriggerSamples int32,
	timeBase uint32, overSample int16, segmentIndex uint32, lpBlockReadyGoPar ps2000a.BlockReady,
	param interface{}) (timeIndisposedMs int32, err error) {
	return
}

func (ps *SimDesc) SetTriggerChannelProperties(channelProperties []ps2000a.TriggerChannelProperties, auxOutputEnable bool,
	autoTriggerMs int32) (err error) {
	return

}

func (ps *SimDesc) SetTriggerChannelConditions(triggerConditions []ps2000a.TriggerConditions) (err error) {
	return
}

func (ps *SimDesc) SetTriggerChannelDirections(channelA, channelB, channelC, channelD, ext, aux ps2000a.ThresholdDirection) (err error) {
	return
}

func (ps *SimDesc) SetTriggerDelay(delay uint32) (err error) {
	return
}

func (ps *SimDesc) SetPulseWidthQualifier(conditions []ps2000a.PwqConditions, direction ps2000a.ThresholdDirection, lower, upper uint32,
	pwType ps2000a.PulseWidthType) (err error) {
	return
}
func (ps *SimDesc) SetTriggerDigitalPortProperties(digitalDirections []ps2000a.DigitalChannelDirections) (err error) {
	return
}

func (ps *SimDesc) Stop() (err error) {
	return
}

func (ps *SimDesc) SetSigGenBuiltIn(offsetVoltage int32, pkToPK uint32, waveType ps2000a.WaveTypeEnum,
	startFrequency, stopFrequency, increment, dwellTime float32, sweepType ps2000a.SweepTypeEnum,
	operation ps2000a.ExtraOperations, shots, sweeps uint32, triggerType ps2000a.SigGenTrigType,
	triggerSource ps2000a.SigGenTrigSource, extInThreshold int16) (err error) {
	return
}

func (ps *SimDesc) SetSigGenBuiltInV2(offsetVoltage int32, pkToPK uint32, waveType ps2000a.WaveTypeEnum,
	startFrequency, stopFrequency, increment, dwellTime float64, sweepType ps2000a.SweepTypeEnum,
	operation ps2000a.ExtraOperations, shots, sweeps uint32, triggerType ps2000a.SigGenTrigType,
	triggerSource ps2000a.SigGenTrigSource, extInThreshold int16) (err error) {
	return
}

func (ps *SimDesc) SigGenFrequencyToPhase(frequency float64, indexMode ps2000a.IndexMode, bufferLength uint32) (phase uint32, err error) {
	return
}

func (ps *SimDesc) SetNoCaptures(nCaptures uint32) (err error) {
	return
}

func (ps *SimDesc) GetTriggerTimeOffset(segmentIndex uint32) (timeUpper, timeLower uint32, timeUnits ps2000a.TimeUnits, err error) {
	return
}

func (ps *SimDesc) GetTriggerTimeOffset64(segmentIndex uint32) (time int64, timeUnits ps2000a.TimeUnits, err error) {
	return

}

func (ps *SimDesc) GetValuesTriggerTimeOffsetBulk(timesUpper, timesLower []uint32, timeUnits []ps2000a.TimeUnits,
	fromSegmentIndex, toSegmentIndex uint32) (err error) {
	return
}

func (ps *SimDesc) GetValuesTriggerTimeOffsetBulk64(times []int64, timeUnits []ps2000a.TimeUnits,
	fromSegmentIndex, toSegmentIndex uint32) (err error) {
	return
}

func (ps *SimDesc) HoldOff(holdOff uint64, holdOffType ps2000a.HoldOffType) (err error) {
	return
}

func (ps *SimDesc) LsReady() (ready int16, err error) {
	return
}

func (ps *SimDesc) TriggerOrPulseWidthQualifierEnabled() (triggerEnabled, pulseWidthQualifierEnabledint16 int16, err error) {
	return
}

func (ps *SimDesc) MemorySegments(nSegments uint32) (nMaxSamples int32, err error) {
	return
}

func (ps *SimDesc) NoOfStreamingValues() (noOfValues uint32, err error) {
	return
}

func (ps *SimDesc) OpenUnitProgress() (retHandle, progressPercent, complete int16, err error) {
	return
}

func (ps *SimDesc) PingUnit() (err error) {
	return
}

func (ps *SimDesc) QueryOutputEdgeDetect() (state int16, err error) {
	return
}

func (ps *SimDesc) SetDigitalAnalogTriggerOperand(operand ps2000a.TriggerOperand) (err error) {
	return
}

func (ps *SimDesc) SetDigitalPort(port ps2000a.DigitalPort, enabled bool, logiclevel int16) (err error) {
	return
}

func (ps *SimDesc) SetOutputEdgeDetect(state int16) (err error) {
	return
}

func (ps *SimDesc) SetPulseWidthDigitalPortProperties(digitalDirections []ps2000a.DigitalChannelDirections) (err error) {
	return
}

func (ps *SimDesc) SetSigGenArbitrary(offsetVoltage int32, pkToPK uint32,
	startDeltaPhase, stopDeltaPhase, deltaPhaseIncrement, dwellCount uint32,
	arbitraryWaveform []int16, sweepType ps2000a.SweepTypeEnum, operation ps2000a.ExtraOperations,
	indexMode ps2000a.IndexMode, shots, sweeps uint32, triggerType ps2000a.SigGenTrigType,
	triggerSource ps2000a.SigGenTrigSource, extInThreshold int16) (err error) {
	return
}

func (ps *SimDesc) SetSigGenPropertiesArbitrary(offsetVoltage int32,
	startDeltaPhase, stopDeltaPhase, deltaPhaseIncrement, dwellCount uint32,
	sweepType ps2000a.SweepTypeEnum, operation ps2000a.ExtraOperations,
	indexMode ps2000a.IndexMode, shots, sweeps uint32, triggerType ps2000a.SigGenTrigType,
	triggerSource ps2000a.SigGenTrigSource, extInThreshold int16) (err error) {
	return
}

func (ps *SimDesc) SetSigGenPropertiesBuiltIn(offsetVoltage int32,
	startFrequency, stopFrequency, increment, dwellTime float64,
	sweepType ps2000a.SweepTypeEnum,
	shots, sweeps uint32, triggerType ps2000a.SigGenTrigType,
	triggerSource ps2000a.SigGenTrigSource, extInThreshold int16) (err error) {
	return
}

func (ps *SimDesc) SigGenArbitraryMinMaxValues() (minArbitraryWaveformValue, maxArbitraryWaveformValue int16,
	minArbitraryWaveformSize, maxArbitraryWaveformSize uint32, err error) {
	return
}

func (ps *SimDesc) SigGenSoftwareControl(state int16) (err error) {
	return
}
