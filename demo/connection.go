// Package demo implements a software-simulated oscilloscope backend for fynescope,
// modeling acquisition hardware, waveform synthesis, signal generators, and trigger systems.
package demo

import (
	"fmt"
	"fynescope/genericps"
	"log/slog"
	"time"
)

const (
	cmdSendTimeout         = 5000 * time.Millisecond
	responseReceiveTimeout = 5000 * time.Millisecond
)

// checkHandle validates that the simulated device handle is positive.
func checkHandle(handle int16) error {
	if handle <= 0 {
		return fmt.Errorf("invalid handle: %d", handle)
	}
	return nil
}

// Stop does nothing. Probably it should terminate any ongoing background
// activities for the demo connection.
// TODO terminate background activities
func Stop() (err error) {
	return
}

// getValues handles GetValuesMsg by calling simGetValues and returning captured
// waveform data.
func getValues(m *genericps.GetValuesMsg) {
	var (
		numOfSamples uint32
		overflow     int16
		err          error
	)

	numOfSamples, overflow, err = simGetValues(m.Handle(), uint32(m.StartIndex),
		uint32(m.ReqNumOfSamples),
		uint32(m.DownSampleRatio),
		RatioMode(m.DownSampleRatioMode),
		uint32(m.SegmentIndex))
	response := m.Rsp().(*genericps.GetValuesRsp)
	response.SetStatus(err)
	response.NumOfSamples = uint64(numOfSamples)
	response.Overflow = overflow
	m.RspCh() <- struct{}{}
}

// closeUnit handles CloseUnitMsg by closing the simulated oscilloscope unit.
func closeUnit(m *genericps.CloseUnitMsg) {
	slog.Info("Close unit")
	err := simCloseUnit(m.Handle())
	response := m.Rsp().(*genericps.CloseUnitRsp)
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

// setChannel handles SetChannelMsg by configuring channel enable, coupling,
// range, and analog offset.
func setChannel(m *genericps.SetChannelMsg) {
	var err error
	err = simSetChannel(m.Handle(), ChannelId(m.Channel),
		m.Enabled, Coupling(m.CouplingType),
		RangeEnum(m.VoltageRange), m.AnalogOffset)
	response := m.Rsp().(*genericps.SetChannelRsp)
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

// flashLed handles FlashLedMsg by simulating the front-panel LED flash sequence.
func flashLed(m *genericps.FlashLedMsg) {
	err := simFlashLed(m.Handle(), m.Start)
	response := m.Rsp().(*genericps.FlashLedRsp)
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

// getUnitInfo handles GetUnitInfoMsg by retrieving hardware metadata strings
// for the simulated scope.
func getUnitInfo(m *genericps.GetUnitInfoMsg) {
	var (
		err error
		s   string
	)
	s, err = simGetUnitInfo(m.Handle(), PicoInfo(m.Info))
	if err != nil {
		slog.Error("PicoInfo ", "error:", err)
	}
	response := m.Rsp().(*genericps.GetUnitInfoRsp)
	response.InfoString = s
	response.SetStatus(err)
	slog.Info("GetUnitInfo ", "info:", s)
	slog.Debug("GetUnitInfo", "rsp", m.Rsp().(*genericps.GetUnitInfoRsp))
	m.RspCh() <- struct{}{}
}

// getValuesAsync handles GetValuesAsyncMsg by initiating asynchronous sample
// retrieval with a callback.
func getValuesAsync(m *genericps.GetValuesAsyncMsg) {
	var err error
	err = simGetValuesAsync(m.Handle(), uint32(m.StartIndex),
		uint32(m.NumOfSamples),
		uint32(m.DownSampleRatio),
		RatioMode(m.DownSampleRatioMode),
		DataReady(m.LpDataReady),
		uint32(m.SegmentIndex),
		m.Param)
	response := m.Rsp().(*genericps.GetValuesAsyncRsp)
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

// getValuesBulk handles GetValuesBulkMsg by retrieving samples across multiple
// memory segments.
func getValuesBulk(m *genericps.GetValuesBulkMsg) {
	var (
		err          error
		numOfSamples uint32
	)
	numOfSamples, err = simGetValuesBulk(m.Handle(), uint32(m.ReqNumOfSamples),
		uint32(m.FromSegmentIndex), uint32(m.ToSegmentIndex),
		uint32(m.DownSampleRatio), RatioMode(m.DownSampleRatioMode), m.Overflow)
	response := m.Rsp().(*genericps.GetValuesBulkRsp)
	response.NumOfSamples = uint64(numOfSamples)
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

// getValuesOverlapped handles GetValuesOverlappedMsg by reading samples with
// non-blocking overlapped I/O.
func getValuesOverlapped(m *genericps.GetValuesOverlappedMsg) {
	var (
		err          error
		numOfSamples uint32
	)
	numOfSamples, err = simGetValuesOverlapped(m.Handle(), m.StartIndex, m.ReqNumOfSamples,
		m.DownSampleRatio, RatioMode(m.DownSampleRatioMode), m.SegmentIndex, m.Overflow)
	response := m.Rsp().(*genericps.GetValuesOverlappedRsp)
	response.NumOfSamples = numOfSamples
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

// getValuesOverlappedBulk handles GetValuesOverlappedBulkMsg by reading overlapped
// samples across segments.
func getValuesOverlappedBulk(m *genericps.GetValuesOverlappedBulkMsg) {
	var (
		err          error
		numOfSamples uint32
	)
	numOfSamples, err = simGetValuesOverlappedBulk(m.Handle(), m.StartIndex, m.ReqNumOfSamples,
		m.DownSampleRatio, RatioMode(m.DownSampleRatioMode), m.FromSegment, m.ToSegment, m.Overflow)
	response := m.Rsp().(*genericps.GetValuesOverlappedBulkRsp)
	response.NumOfSamples = numOfSamples
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

// getAnalogueOffsetMsg handles GetAnalogueOffsetMsg by returning minimum and
// maximum allowable offset voltages.
func getAnalogueOffsetMsg(m *genericps.GetAnalogueOffsetMsg) {
	var (
		err                            error
		maximumVoltage, minimumVoltage float32
	)
	maximumVoltage, minimumVoltage, err = simGetAnalogueOffset(m.Handle(), m.VoltageRange, Coupling(m.Coupling))
	response := m.Rsp().(*genericps.GetAnalogueOffsetRsp)
	response.MinimumVoltage = minimumVoltage
	response.MaximumVoltage = maximumVoltage
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

// getChannelInformation handles GetChannelInformationMsg by returning supported
// voltage ranges for a channel.
func getChannelInformation(m *genericps.GetChannelInformationMsg) {
	var (
		err            error
		lengthOfRanges int32
	)
	lengthOfRanges, err = simGetChannelInformation(m.Handle(), m.Info, m.Probe,
		m.Ranges, ChannelId(m.Channel))
	m.Ranges = m.Ranges[0:lengthOfRanges]

	response := m.Rsp().(*genericps.GetChannelInformationRsp)
	response.Ranges = m.Ranges
	response.LengthOfRanges = lengthOfRanges
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

// getMaxDownSampleRatio handles GetMaxDownSampleRatioMsg by querying max
// downsampling capability.
func getMaxDownSampleRatio(m *genericps.GetMaxDownSampleRatioMsg) {
	var (
		err                error
		maxDownSampleRatio uint32
	)
	maxDownSampleRatio, err = simGetMaxDownSampleRatio(m.Handle(),
		m.NumOfUnaggregatedSamples, RatioMode(m.DownSampleRatioMode), m.SegmentIndex)
	response := m.Rsp().(*genericps.GetMaxDownSampleRatioRsp)
	response.MaxDownSampleRatio = maxDownSampleRatio
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

// getMaxSegments handles GetMaxSegmentsMsg by returning the maximum number of
// memory segments supported.
func getMaxSegments(m *genericps.GetMaxSegmentsMsg) {
	maxSegments, err := simGetMaxSegments(m.Handle())
	response := m.Rsp().(*genericps.GetMaxSegmentsRsp)
	response.MaxSegments = maxSegments
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

// getNumberOfCaptures handles GetNumOfCapturesMsg by returning total completed
// waveform captures.
func getNumberOfCaptures(m *genericps.GetNumOfCapturesMsg) {
	numOfCaptures, err := simGetNoOfCaptures(m.Handle())
	response := m.Rsp().(*genericps.GetNumOfCapturesRsp)
	response.NCaptures = uint64(numOfCaptures)
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

// getNumberOfProcessedCaptures handles GetNumOfProcessedCapturesMsg by returning
// processed capture counts.
func getNumberOfProcessedCaptures(m *genericps.GetNumOfProcessedCapturesMsg) {
	numOfCaptures, err := simGetNoOfProcessedCaptures(m.Handle())

	response := m.Rsp().(*genericps.GetNumOfProcessedCapturesRsp)
	response.NCaptures = uint64(numOfCaptures)
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

// getStreamingLatestValues handles GetStreamingLatestValuesMsg by streaming
// real-time values to the callback.
func getStreamingLatestValues(m *genericps.GetStreamingLatestValuesMsg) {
	var (
		err error
	)
	err = simGetStreamingLatestValues(m.Handle(),
		StreamingReady(m.LpStreamingReadyGoPar), m.Param)
	response := m.Rsp().(*genericps.GetStreamingLatestValuesRsp)
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

// getTimebase handles GetTimebaseMsg by computing time intervals and max
// samples for integer timebases.
func getTimebase(m *genericps.GetTimebaseMsg) {
	timeIntervalNanoseconds, maxSamples, err := simGetTimebase(m.Handle(),
		uint32(m.TimeBase), int32(m.NumOfSamples), m.OverSample, m.SegmentIndex)

	response := m.Rsp().(*genericps.GetTimebaseRsp)
	response.TimeIntervalNanoseconds = float64(timeIntervalNanoseconds)
	response.MaxSamples = uint64(maxSamples)
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

// getTimebase2 handles GetTimebase2Msg by computing time intervals with
// sub-nanosecond precision.
func getTimebase2(m *genericps.GetTimebase2Msg) {
	timeIntervalNanoseconds, maxSamples, err := simGetTimebase2(m.Handle(),
		uint32(m.TimeBase), uint32(m.NumOfSamples), m.OverSample, uint32(m.SegmentIndex))

	response := m.Rsp().(*genericps.GetTimebase2Rsp)
	response.TimeIntervalNanoseconds = float64(timeIntervalNanoseconds)
	response.MaxSamples = uint64(maxSamples)
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

// maximumValue handles MaximumValueMsg by returning the maximum raw ADC sample
// value (e.g. 32767).
func maximumValue(m *genericps.MaximumValueMsg) {
	value, err := simMaximumValue(m.Handle())

	response := m.Rsp().(*genericps.MaximumValueRsp)
	response.Value = value
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

// minimumValue handles MinimumValueMsg by returning the minimum raw ADC sample
// value (e.g. -32767).
func minimumValue(m *genericps.MinimumValueMsg) {
	value, err := simMinimumValue(m.Handle())

	response := m.Rsp().(*genericps.MinimumValueResp)
	response.Value = value
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

// setSimpleTrigger handles SetSimpleTriggerMsg by configuring a basic edge
// trigger condition.
func setSimpleTrigger(m *genericps.SetSimpleTriggerMsg) {
	var (
		err error
	)
	err = simSetSimpleTrigger(m.Handle(), m.Enable, ChannelId(m.Source),
		m.Threshold, ThresholdDirection(m.Direction), m.Delay, m.AutoTriggerMs)

	response := m.Rsp().(*genericps.SetSimpleTriggerRsp)
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

// setDataBuffer handles SetDataBufferMsg by binding a single destination buffer
// for channel samples.
func setDataBuffer(m *genericps.SetDataBufferMsg) {
	var (
		err error
	)
	err = simSetDataBuffer(m.Handle(), ChannelId(m.Ch), m.BufferIn,
		uint32(m.SegmentIndex), RatioMode(m.Mode))
	response := m.Rsp().(*genericps.SetDataBufferRsp)
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

// setDataBuffers handles SetDataBuffersMsg by binding min/max buffers for
// downsampled/aggregated samples.
func setDataBuffers(m *genericps.SetDataBuffersMsg) {
	var (
		err error
	)
	err = simSetDataBuffers(m.Handle(), ChannelId(m.Ch), m.BufferMax, m.BufferMin,
		m.SegmentIndex, RatioMode(m.Mode))
	response := m.Rsp().(*genericps.SetDataBuffersRsp)
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

// setUnscaledDataBuffers handles SetUnscaledDataBuffersMsg by binding raw
// unscaled ADC data buffers.
func setUnscaledDataBuffers(m *genericps.SetUnscaledDataBuffersMsg) {
	var (
		err error
	)
	err = simSetUnscaledDataBuffers(m.Handle(), ChannelId(m.
		Ch), m.BufferMax, m.BufferMin, m.SegmentIndex, RatioMode(m.Mode))
	response := m.Rsp().(*genericps.SetUnscaledataBuffersRsp)
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

// setEtsTimeBuffer handles SetEtsTimeBufferMsg by registering an ETS (Equivalent
// Time Sampling) timing buffer.
func setEtsTimeBuffer(m *genericps.SetEtsTimeBufferMsg) {
	err := simSetEtsTimeBuffer(m.Handle(), m.Buffer)
	response := m.Rsp().(*genericps.SetEtsTimeBufferRsp)
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

// setEtsTimeBuffers handles SetEtsTimeBuffersMsg by registering upper and lower
// 32-bit ETS time buffers.
func setEtsTimeBuffers(m *genericps.SetEtsTimeBuffersMsg) {
	err := simSetEtsTimeBuffers(m.Handle(), m.TimeUpper, m.TimeLower)
	response := m.Rsp().(*genericps.SetEtsTimeBufferRsp)
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

// setEts handles SetEtsMsg by configuring Equivalent Time Sampling mode parameters.
func setEts(m *genericps.SetEtsMsg) {
	var (
		err                   error
		sampleTimePicoseconds int32
	)
	sampleTimePicoseconds, err = simSetEts(m.Handle(), EtsMode(m.Mode),
		m.EtsCycles, m.EtsInterleave)
	response := m.Rsp().(*genericps.SetEtsRsp)
	response.SampleTimePicoseconds = sampleTimePicoseconds
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

// runStreaming handles RunStreamingMsg by starting simulated continuous
// streaming acquisition.
func runStreaming(m *genericps.RunStreamingMsg) {
	var (
		err            error
		sampleInterval uint32
	)
	sampleInterval, err = simRunStreaming(m.Handle(), m.ReqSampleInterval,
		TimeUnits(m.SampleIntervalTimeUnits), m.MaxPostTriggerSamples,
		m.MaxPostTriggerSamples, m.AutoStop, m.DownSampleRatio,
		RatioMode(m.DownSampleRatioMode), m.OverviewBufferSize)
	response := m.Rsp().(*genericps.RunStreamingRsp)
	response.SampleInterval = sampleInterval
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

// runBlock handles RunBlockMsg by starting simulated block-mode acquisition.
func runBlock(m *genericps.RunBlockMsg) {
	var (
		err              error
		timeIndisposedMs int32
	)
	timeIndisposedMs, err = simRunBlock(m.Handle(), int32(m.NumOfPreTriggerSamples),
		int32(m.NumOfPostTriggerSamples),
		uint32(m.TimeBase), m.OverSample, uint32(m.SegmentIndex),
		BlockReady(m.LpBlockReadyGoPar), m.Param)
	response := m.Rsp().(*genericps.RunBlockRsp)
	response.TimeIndisposedMs = timeIndisposedMs
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

// setTriggerChannelProperties handles SetTriggerChannelPropertiesMsg by setting
// channel trigger thresholds and hysteresis.
func setTriggerChannelProperties(m *genericps.SetTriggerChannelPropertiesMsg) {
	var (
		err error
	)
	tcp := make([]TriggerChannelProperties, len(m.ChannelProperties))
	for i := range tcp {
		tcp[i].Channel = ChannelId(m.ChannelProperties[i].Channel)
		tcp[i].ThresholdLower = m.ChannelProperties[i].ThresholdLower
		tcp[i].ThresholdLowerHysteresis = m.ChannelProperties[i].ThresholdLowerHysteresis
		tcp[i].ThresholdMode = ThresholdModeId(m.ChannelProperties[i].ThresholdMode)
		tcp[i].ThresholdUpper = m.ChannelProperties[i].ThresholdUpper
		tcp[i].ThresholdUpperHysteresis = m.ChannelProperties[i].ThresholdUpperHysteresis
	}
	err = simSetTriggerChannelProperties(m.Handle(), tcp, m.AuxOutputEnable, m.AutoTriggerMs)

	response := m.Rsp().(*genericps.SetTriggerChannelPropertiesRsp)
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

// setTriggerChannelConditions handles SetTriggerChannelConditionsMsg by
// configuring complex multi-channel trigger conditions.
func setTriggerChannelConditions(m *genericps.SetTriggerChannelConditionsMsg) {
	var (
		err error
	)
	tc := make([]TriggerConditions, len(m.TriggerConditions))
	for i := range tc {
		tc[i].Aux = TriggerState(m.TriggerConditions[i].Aux)
		tc[i].ChannelA = TriggerState(m.TriggerConditions[i].ChannelA)
		tc[i].ChannelB = TriggerState(m.TriggerConditions[i].ChannelB)
		tc[i].ChannelC = TriggerState(m.TriggerConditions[i].ChannelC)
		tc[i].ChannelD = TriggerState(m.TriggerConditions[i].ChannelD)
		tc[i].Digital = TriggerState(m.TriggerConditions[i].Digital)
		tc[i].External = TriggerState(m.TriggerConditions[i].External)
		tc[i].PulseWidthQualifier = TriggerState(m.TriggerConditions[i].PulseWidthQualifier)
	}
	err = simSetTriggerChannelConditions(m.Handle(), tc)
	response := m.Rsp().(*genericps.SetTriggerChannelConditionsRsp)
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

// setTriggerChannelDirections handles SetTriggerChannelDirectionsMsg
// by configuring edge directions (rising/falling) per channel.
func setTriggerChannelDirections(m *genericps.SetTriggerChannelDirectionsMsg) {
	var (
		err error
	)
	err = simSetTriggerChannelDirections(m.Handle(), ThresholdDirection(m.ChannelA),
		ThresholdDirection(m.ChannelB), ThresholdDirection(m.ChannelC),
		ThresholdDirection(m.ChannelD), ThresholdDirection(m.Ext), ThresholdDirection(m.Aux))
	response := m.Rsp().(*genericps.SetTriggerChannelDirectionsRsp)
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

// setTriggerDelay handles SetTriggerDelayMsg by setting sample delay count after
// trigger event before capture.
func setTriggerDelay(m *genericps.SetTriggerDelayMsg) {
	err := simSetTriggerDelay(m.Handle(), m.Delay)
	response := m.Rsp().(*genericps.SetTriggerDelayRsp)
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

// setPulseWidthQualifier handles SetPulseWidthQualifierMsg by setting
// pulse-width qualifier parameters.
func setPulseWidthQualifier(m *genericps.SetPulseWidthQualifierMsg) {
	var err error
	c := make([]PwqConditions, len(m.Conditions))
	for i := range c {
		c[i].Aux = TriggerState(m.Conditions[i].Aux)
		c[i].ChannelA = TriggerState(m.Conditions[i].ChannelA)
		c[i].ChannelB = TriggerState(m.Conditions[i].ChannelB)
		c[i].ChannelC = TriggerState(m.Conditions[i].ChannelC)
		c[i].ChannelD = TriggerState(m.Conditions[i].ChannelD)
		c[i].External = TriggerState(m.Conditions[i].External)
		c[i].Digital = TriggerState(m.Conditions[i].Digital)
	}
	err = simSetPulseWidthQualifier(m.Handle(), c, ThresholdDirection(m.Direction),
		m.Lower, m.Upper, PulseWidthType(m.PwType))
	response := m.Rsp().(*genericps.SetPulseWidthQualifierRsp)
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

// setTriggerDigitalPortProperties handles SetTriggerDigitalPortPropertiesMsg
// by configuring digital channel trigger directions.
func setTriggerDigitalPortProperties(m *genericps.SetTriggerDigitalPortPropertiesMsg) {
	var (
		err error
	)
	tdp := make([]DigitalChannelDirections, len(m.DigitalDirections))
	for i := range tdp {
		tdp[i].Channel = DigitalChannel(m.DigitalDirections[i].Channel)
		tdp[i].Direction = DigitalDirection(m.DigitalDirections[i].Direction)
	}
	err = simSetTriggerDigitalPortProperties(m.Handle(), tdp)

	response := m.Rsp().(*genericps.SetTriggerDigitalPortPropertiesRsp)
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

// stop handles StopMsg by halting data collection on the simulated oscilloscope.
func stop(m *genericps.StopMsg) {
	err := simStop(m.Handle())
	response := m.Rsp().(*genericps.StopRsp)
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

// setSigGenBuiltIn handles SetSigGenBuiltInMsg by configuring standard
// built-in signal generator waveforms and sweeps.
func setSigGenBuiltIn(m *genericps.SetSigGenBuiltInMsg) {
	var (
		err error
	)
	err = simSetSigGenBuiltIn(m.Handle(), m.OffsetVoltage, m.PkToPK,
		WaveTypeEnum(m.WaveType), m.StartFrequency,
		m.StopFrequency, m.Increment, m.DwellTime, SweepTypeEnum(m.SweepType),
		ExtraOperations(m.Operation), m.Shots, m.Sweeps, SigGenTrigType(m.TriggerType),
		SigGenTrigSource(m.TriggerSource), m.ExtInThreshold)
	response := m.Rsp().(*genericps.SetSigGenBuiltInRsp)
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

// setSigGenBuiltInV2 handles SetSigGenBuiltInV2Msg with extended options.
func setSigGenBuiltInV2(m *genericps.SetSigGenBuiltInV2Msg) {
	var (
		err error
	)
	err = simSetSigGenBuiltInV2(m.Handle(), m.OffsetVoltage, m.PkToPK,
		WaveTypeEnum(m.WaveType), m.StartFrequency,
		m.StopFrequency, m.Increment, m.DwellTime, SweepTypeEnum(m.SweepType),
		ExtraOperations(m.Operation), m.Shots, m.Sweeps, SigGenTrigType(m.TriggerType),
		SigGenTrigSource(m.TriggerSource), m.ExtInThreshold)
	response := m.Rsp().(*genericps.SetSigGenBuiltInV2Rsp)
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

// sigGenFrequencyToPhase handles SigGenFrequencyToPhasenMsg by calculating
// delta phase increment from frequency.
func sigGenFrequencyToPhase(m *genericps.SigGenFrequencyToPhasenMsg) {
	var (
		err   error
		phase uint32
	)
	phase, err = simSigGenFrequencyToPhase(m.Handle(), m.Frequency,
		IndexMode(m.IndexMode), m.BufferLength)
	response := m.Rsp().(*genericps.SigGenFrequencyToPhaseRsp)
	response.SetStatus(err)
	response.Phase = phase
	m.RspCh() <- struct{}{}
}

// setNumOfCaptures handles SetNumOfCapturesMsg by configuring the number of
// waveforms to capture in rapid block mode.
func setNumOfCaptures(m *genericps.SetNumOfCapturesMsg) {
	err := simSetNoCaptures(m.Handle(), m.NCaptures)
	response := m.Rsp().(*genericps.SetNumOfCapturesRsp)
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

// getTriggerTimeOffset handles GetTriggerTimeOffsetMsg by returning 32-bit
// trigger time offsets.
func getTriggerTimeOffset(m *genericps.GetTriggerTimeOffsetMsg) {
	var (
		err                  error
		timeUpper, timeLower uint32
		timeUnits            TimeUnits
	)
	timeUpper, timeLower, timeUnits, err = simGetTriggerTimeOffset(m.Handle(),
		uint32(m.SegmentIndex))
	response := m.Rsp().(*genericps.GetTriggerTimeOffsetRsp)
	response.TimeLower = timeLower
	response.TimeUnits = genericps.TimeUnits(timeUnits)
	response.TimeUpper = timeUpper
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

// getTriggerTimeOffset64 handles GetTriggerTimeOffset64Msg by returning 64-bit
// trigger time offsets.
func getTriggerTimeOffset64(m *genericps.GetTriggerTimeOffset64Msg) {
	time, timeUnits, err := simGetTriggerTimeOffset64(m.Handle(), uint32(m.SegmentIndex))
	response := m.Rsp().(*genericps.GetTriggerTimeOffset64Rsp)
	response.Time = time
	response.TimeUnits = genericps.TimeUnits(timeUnits)
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

// getValuesTriggerTimeOffsetBulk handles GetValuesTriggerTimeOffsetBulkMsg by
// retrieving 32-bit time offset.
func getValuesTriggerTimeOffsetBulk(m *genericps.GetValuesTriggerTimeOffsetBulkMsg) {
	var (
		err error
	)
	tu := make([]TimeUnits, len(m.TimeUnits))
	for i := range tu {
		tu[i] = TimeUnits(m.TimeUnits[i])
	}
	err = simGetValuesTriggerTimeOffsetBulk(m.Handle(), m.TimesUpper,
		m.TimesLower, tu, m.FromSegmentIndex, m.ToSegmentIndex)
	response := m.Rsp().(*genericps.GetValuesTriggerTimeOffsetBulkRsp)
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

// getValuesTriggerTimeOffsetBulk64 handles GetValuesTriggerTimeOffsetBulk64Msg
// by retrieving 64-bit time offset.
func getValuesTriggerTimeOffsetBulk64(m *genericps.GetValuesTriggerTimeOffsetBulk64Msg) {
	var (
		err error
	)
	tu := make([]TimeUnits, len(m.TimeUnits))
	for i := range tu {
		tu[i] = TimeUnits(m.TimeUnits[i])
	}
	err = simGetValuesTriggerTimeOffsetBulk64(m.Handle(), m.Times, tu, m.FromSegmentIndex, m.ToSegmentIndex)
	response := m.Rsp().(*genericps.GetValuesTriggerTimeOffsetBulk64Rsp)
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

// holdOff handles HoldOffMsg by configuring hold-off time duration between trigger events.
func holdOff(m *genericps.HoldOffMsg) {
	err := simHoldOff(m.Handle(), m.HoldOff, HoldOffType(m.HoldOffType))
	response := m.Rsp().(*genericps.HoldOffRsp)
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

// simIsReady checks whether the simulated oscilloscope handle is ready for data transfer.
func simIsReady(handle int16) (ready int16, err error) {
	if handle > 0 {
		ready = 1
	}
	return
}

// lsReady handles IsReadyMsg by reporting whether capture has finished and data is available.
func lsReady(m *genericps.IsReadyMsg) {
	ready, err := simIsReady(m.Handle())
	response := m.Rsp().(*genericps.IsReadyRsp)
	response.Ready = ready
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

// triggerOrPulseWidthQualifierEnabled handles TriggerOrPulseWidthQualifierEnabledMsg
// by reporting trigger/PWQ state.
func triggerOrPulseWidthQualifierEnabled(m *genericps.TriggerOrPulseWidthQualifierEnabledMsg) {
	triggerEnabled, pulseWidthQualifierEnabledint16, err := simTriggerOrPulseWidthQualifierEnabled(m.Handle())
	response := m.Rsp().(*genericps.TriggerOrPulseWidthQualifierEnabledRsp)
	response.TriggerEnabled = triggerEnabled
	response.PulseWidthQualifierEnabledint16 = pulseWidthQualifierEnabledint16
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

// memorySegments handles MemorySegmentsMsg by partitioning sample memory into segments.
func memorySegments(m *genericps.MemorySegmentsMsg) {
	nMaxSamples, err := simMemorySegments(m.Handle(), m.NSegments)
	response := m.Rsp().(*genericps.MemorySegmentsRsp)
	response.SetStatus(err)
	response.NMaxSamples = nMaxSamples
	m.RspCh() <- struct{}{}
}

// numOfStreamingValues handles NumOfStreamingValuesMsg by returning count of
// captured streaming samples.
func numOfStreamingValues(m *genericps.NumOfStreamingValuesMsg) {
	numOfValues, err := simNoOfStreamingValues(m.Handle())
	response := m.Rsp().(*genericps.NumOfStreamingValuesRsp)
	response.NumOfValues = numOfValues
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

// pingUnit handles PingUnitMsg by verifying that the simulated scope responds
// to heartbeat checks.
func pingUnit(m *genericps.PingUnitMsg) {
	err := simPingUnit(m.Handle())
	response := m.Rsp().(*genericps.PingUnitRsp)
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

// queryOutputEdgeDetect handles QueryOutputEdgeDetectMsg by reporting whether
// edge detection is active.
func queryOutputEdgeDetect(m *genericps.QueryOutputEdgeDetectMsg) {
	state, err := simQueryOutputEdgeDetect(m.Handle())
	response := m.Rsp().(*genericps.QueryOutputEdgeDetectRsp)
	response.State = state
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

// setDigitalAnalogTriggerOperand handles SetDigitalAnalogTriggerOperandMsg by
// setting boolean logic between analog and digital triggers.
func setDigitalAnalogTriggerOperand(m *genericps.SetDigitalAnalogTriggerOperandMsg) {
	var (
		err error
	)
	err = simSetDigitalAnalogTriggerOperand(m.Handle(), TriggerOperand(m.Operand))
	response := m.Rsp().(*genericps.SetDigitalAnalogTriggerOperandRsp)
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

// setDigitalChannelsTriggerOperand handles SetDigitalChannelsTriggerOperandMsg
// by setting boolean logic between digital channels.
func setDigitalChannelsTriggerOperand(m *genericps.SetDigitalChannelsTriggerOperandMsg) {
	var (
		err error
	)
	err = simSetDigitalChannelsTriggerOperand(m.Handle(), TriggerOperand(m.Operand))
	response := m.Rsp().(*genericps.SetDigitalChannelsTriggerOperandRsp)
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

// setDigitalPort handles SetDigitalPortMsg by enabling/disabling a digital port
// and setting logic threshold.
func setDigitalPort(m *genericps.SetDigitalPortMsg) {
	var (
		err error
	)
	err = simSetDigitalPort(m.Handle(), DigitalPort(m.Port), m.Enabled, m.Logiclevel)
	response := m.Rsp().(*genericps.SetDigitalPortRsp)
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

// setOutputEdgeDetect handles SetOutputEdgeDetectMsg by enabling/disabling
// output edge detection.
func setOutputEdgeDetect(m *genericps.SetOutputEdgeDetectMsg) {
	err := simSetOutputEdgeDetect(m.Handle(), m.State)
	response := m.Rsp().(*genericps.SetOutputEdgeDetectRsp)
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

// setPulseWidthDigitalPortProperties handles SetPulseWidthDigitalPortPropertiesMsg
// by configuring digital PWQ properties.
func setPulseWidthDigitalPortProperties(m *genericps.SetPulseWidthDigitalPortPropertiesMsg) {
	var (
		err error
	)
	dd := make([]DigitalChannelDirections, len(m.DigitalDirections))
	for i := range dd {
		dd[i].Channel = DigitalChannel(m.DigitalDirections[i].Channel)
		dd[i].Direction = DigitalDirection(m.DigitalDirections[i].Direction)
	}
	err = simSetPulseWidthDigitalPortProperties(m.Handle(), dd)
	response := m.Rsp().(*genericps.SetPulseWidthDigitalPortPropertiesRsp)
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

// setSigGenArbitrary handles SetSigGenArbitraryMsg by programming arbitrary
// waveform buffer, frequency, and sweep options.
func setSigGenArbitrary(m *genericps.SetSigGenArbitraryMsg) {
	var (
		err error
	)
	err = simSetSigGenArbitrary(m.Handle(), m.OffsetVoltage, m.PkToPK, m.StartDeltaPhase,
		m.StopDeltaPhase, m.DeltaPhaseIncrement, m.DwellCount, m.ArbitraryWaveform,
		SweepTypeEnum(m.SweepType), ExtraOperations(m.Operation),
		IndexMode(m.IndexMode), m.Shots, m.Sweeps,
		SigGenTrigType(m.TtriggerType),
		SigGenTrigSource(m.TriggerSource), m.ExtInThreshold)
	response := m.Rsp().(*genericps.SetSigGenArbitraryRsp)
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

// setSigGenPropertiesArbitrary handles SetSigGenPropertiesArbitraryMsg
// by updating AWG run parameters.
func setSigGenPropertiesArbitrary(m *genericps.SetSigGenPropertiesArbitraryMsg) {
	err := simSetSigGenPropertiesArbitrary(m.Handle(), m.OffsetVoltage, m.StartDeltaPhase,
		m.StopDeltaPhase, m.DeltaPhaseIncrement, m.DwellCount, SweepTypeEnum(m.SweepType),
		ExtraOperations(m.Operation), IndexMode(m.IndexMode), m.Shots, m.Sweeps,
		SigGenTrigType(m.TriggerType),
		SigGenTrigSource(m.TriggerSource), m.ExtInThreshold)
	response := m.Rsp().(*genericps.SetSigGenPropertiesArbitraryRsp)
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

// setSigGenPropertiesBuiltIn handles SetSigGenPropertiesBuiltInMsg
// by updating standard function generator parameters.
func setSigGenPropertiesBuiltIn(m *genericps.SetSigGenPropertiesBuiltInMsg) {
	err := simSetSigGenPropertiesBuiltIn(m.Handle(), m.OffsetVoltage, m.StartFrequency,
		m.StopFrequency, m.Increment, m.DwellTime, SweepTypeEnum(m.SweepType),
		m.Shots, m.Sweeps,
		SigGenTrigType(m.TriggerType), SigGenTrigSource(m.TriggerSource), m.ExtInThreshold)
	response := m.Rsp().(*genericps.SetSigGenPropertiesBuiltInRsp)
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

// sigGenArbitraryMinMaxValues handles SigGenArbitraryMinMaxValuesMsg
// by returning buffer limits and sample range for AWG.
func sigGenArbitraryMinMaxValues(m *genericps.SigGenArbitraryMinMaxValuesMsg) {
	minArbitraryWaveformValue, maxArbitraryWaveformValue, minArbitraryWaveformSize,
		maxArbitraryWaveformSize, err := simSigGenArbitraryMinMaxValues(m.Handle())
	response := m.Rsp().(*genericps.SigGenArbitraryMinMaxValuesRsp)
	response.MinArbitraryWaveformValue = minArbitraryWaveformValue
	response.MaxArbitraryWaveformValue = maxArbitraryWaveformValue
	response.MinArbitraryWaveformSize = minArbitraryWaveformSize
	response.MaxArbitraryWaveformSize = maxArbitraryWaveformSize
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

// sigGenSoftwareControl handles SigGenSoftwareControlMsg by triggering
// or gating the signal generator via software.
func sigGenSoftwareControl(m *genericps.SigGenSoftwareControlMsg) {
	err := simSigGenSoftwareControl(m.Handle(), m.State)
	response := m.Rsp().(*genericps.SigGenSoftwareControlRsp)
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

// setDemoGen handles SetDemoGenMsg by configuring demo signal generator
// properties for simulated channel synthesis.
func setDemoGen(m *genericps.SetDemoGenMsg) {
	response := m.Rsp().(*genericps.SetDemoGenRsp)
	if err := checkHandle(m.Handle()); err != nil {
		response.SetStatus(err)
		m.RspCh() <- struct{}{}
		return
	}
	s := SimDesc{handle: m.Handle()}
	err := s.SetDemoGen(m.Channel, m.On, m.OffsetVoltage, m.PkToPK, genericps.WaveTypeEnum(m.WaveType),
		m.StartFrequency, m.StopFrequency, m.Increment, m.DwellTime, genericps.SweepTypeEnum(m.SweepType),
		genericps.ExtraOperations(m.Operation), m.Shots, m.Sweeps, genericps.SigGenTrigType(m.TriggerType),
		genericps.SigGenTrigSource(m.TriggerSource), m.ExtInThreshold, m.Phase, m.ArbitraryWaveform, m.SpiDataValue, m.I2cAddressValue)
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

// setDemoRlcFilter handles SetDemoRlcFilterMsg by configuring an interactive
// virtual RLC filter stage for a channel.
func setDemoRlcFilter(m *genericps.SetDemoRlcFilterMsg) {
	response := m.Rsp().(*genericps.SetDemoRlcFilterRsp)
	if err := checkHandle(m.Handle()); err != nil {
		response.SetStatus(err)
		m.RspCh() <- struct{}{}
		return
	}
	s := SimDesc{handle: m.Handle()}
	err := s.SetDemoRlcFilter(m.Channel, m.GenSource, m.Enabled, m.FilterType, m.R, m.RUnit, m.L, m.LUnit, m.C, m.CUnit, m.Q)
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

// setDemoDigitalGen handles SetDemoDigitalGenMsg by configuring simulated
// digital protocol pattern generation.
func setDemoDigitalGen(m *genericps.SetDemoDigitalGenMsg) {
	response := m.Rsp().(*genericps.SetDemoDigitalGenRsp)
	if err := checkHandle(m.Handle()); err != nil {
		response.SetStatus(err)
		m.RspCh() <- struct{}{}
		return
	}
	s := SimDesc{handle: m.Handle()}
	err := s.SetDemoDigitalGen(m.Port0Enabled, m.Port1Enabled, m.Frequency, m.Direction, m.Encoding, m.Mode, m.BitDelay)
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}
