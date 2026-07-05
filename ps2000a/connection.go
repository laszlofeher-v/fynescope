//go:build !noscope

package ps2000a

import (
	"fynescope/genericps"
	"log/slog"
	"reflect"
	"time"
)

const (
	cmdSendTimeout         = 5000 * time.Millisecond
	responseReceiveTimeout = 5000 * time.Millisecond
)

func Stop() (err error) {
	return
}

func getValues(m *genericps.GetValuesMsg) {
	var (
		numOfSamples uint32
		overflow     int16
		err          error
	)
	numOfSamples, overflow, err = ps2000aGetValues(m.Handle(), m.StartIndex,
		m.ReqNumOfSamples,
		m.DownSampleRatio,
		RatioMode(m.DownSampleRatioMode),
		m.SegmentIndex)
	response := m.Rsp().(*genericps.GetValuesRsp)
	response.SetStatus(err)
	response.NumOfSamples = numOfSamples
	response.Overflow = overflow
	m.RspCh() <- struct{}{}
}

func closeUnit(m *genericps.CloseUnitMsg) {
	slog.Info("Close unit")
	err := ps2000aCloseUnit(m.Handle())
	response := m.Rsp().(*genericps.CloseUnitRsp)
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

func setChannel(m *genericps.SetChannelMsg) {
	var err error
	err = ps2000aSetChannel(m.Handle(), ChannelId(m.Channel),
		m.Enabled, Coupling(m.CouplingType),
		RangeEnum(m.VoltageRange), m.AnalogOffset)
	response := m.Rsp().(*genericps.SetChannelRsp)
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

func flashLed(m *genericps.FlashLedMsg) {
	err := ps2000aFlashLed(m.Handle(), m.Start)
	response := m.Rsp().(*genericps.FlashLedRsp)
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

func getUnitInfo(m *genericps.GetUnitInfoMsg) {
	var (
		err error
		s   string
	)
	s, err = ps2000aGetUnitInfo(m.Handle(), PicoInfo(m.Info))
	if err != nil {
		slog.Error("PicoInfo ", "error:", err)
	}
	slog.Info("GetUnitInfo ", "info:", s)
	response := m.Rsp().(*genericps.GetUnitInfoRsp)
	response.InfoString = s
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

func getValuesAsync(m *genericps.GetValuesAsyncMsg) {
	var err error
	err = ps2000aGetValuesAsync(m.Handle(), m.StartIndex,
		m.NumOfSamples,
		m.DownSampleRatio,
		RatioMode(m.DownSampleRatioMode),
		DataReady(m.LpDataReady),
		m.SegmentIndex,
		m.Param)
	response := m.Rsp().(*genericps.GetValuesAsyncRsp)
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

func getValuesBulk(m *genericps.GetValuesBulkMsg) {
	var (
		err          error
		numOfSamples uint32
	)
	numOfSamples, err = ps2000aGetValuesBulk(m.Handle(), m.ReqNumOfSamples, m.FromSegmentIndex, m.ToSegmentIndex,
		m.DownSampleRatio, RatioMode(m.DownSampleRatioMode), m.Overflow)
	response := m.Rsp().(*genericps.GetValuesBulkRsp)
	response.NumOfSamples = numOfSamples
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

func getValuesOverlapped(m *genericps.GetValuesOverlappedMsg) {
	var (
		err          error
		numOfSamples uint32
	)
	numOfSamples, err = ps2000aGetValuesOverlapped(m.Handle(), m.StartIndex, m.ReqNumOfSamples,
		m.DownSampleRatio, RatioMode(m.DownSampleRatioMode), m.SegmentIndex, m.Overflow)
	response := m.Rsp().(*genericps.GetValuesOverlappedRsp)
	response.NumOfSamples = numOfSamples
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

func getValuesOverlappedBulk(m *genericps.GetValuesOverlappedBulkMsg) {
	var (
		err          error
		numOfSamples uint32
	)
	numOfSamples, err = ps2000aGetValuesOverlappedBulk(m.Handle(), m.StartIndex, m.ReqNumOfSamples,
		m.DownSampleRatio, RatioMode(m.DownSampleRatioMode), m.FromSegment, m.ToSegment, m.Overflow)
	response := m.Rsp().(*genericps.GetValuesOverlappedBulkRsp)
	response.NumOfSamples = numOfSamples
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

func getAnalogueOffsetMsg(m *genericps.GetAnalogueOffsetMsg) {
	var (
		err                            error
		maximumVoltage, minimumVoltage float32
	)
	maximumVoltage, minimumVoltage, err = ps2000aGetAnalogueOffset(m.Handle(), m.VoltageRange, Coupling(m.Coupling))
	response := m.Rsp().(*genericps.GetAnalogueOffsetRsp)
	response.MinimumVoltage = minimumVoltage
	response.MaximumVoltage = maximumVoltage
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

func getChannelInformation(m *genericps.GetChannelInformationMsg) {
	var (
		err            error
		lengthOfRanges int32
	)
	lengthOfRanges, err = ps2000aGetChannelInformation(m.Handle(), m.Info, m.Probe, m.Ranges, ChannelId(m.Channel))
	m.Ranges = m.Ranges[0:lengthOfRanges]
	response := m.Rsp().(*genericps.GetChannelInformationRsp)
	response.Ranges = m.Ranges
	response.LengthOfRanges = lengthOfRanges
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

func getMaxDownSampleRatio(m *genericps.GetMaxDownSampleRatioMsg) {
	var (
		err                error
		maxDownSampleRatio uint32
	)
	maxDownSampleRatio, err = ps2000aGetMaxDownSampleRatio(m.Handle(), m.NumOfUnaggregatedSamples, RatioMode(m.DownSampleRatioMode), m.SegmentIndex)
	response := m.Rsp().(*genericps.GetMaxDownSampleRatioRsp)
	response.MaxDownSampleRatio = maxDownSampleRatio
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

func getMaxSegments(m *genericps.GetMaxSegmentsMsg) {
	maxSegments, err := ps2000aGetMaxSegments(m.Handle())
	response := m.Rsp().(*genericps.GetMaxSegmentsRsp)
	response.MaxSegments = maxSegments
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

func getNumberOfCaptures(m *genericps.GetNumOfCapturesMsg) {
	numOfCaptures, err := ps2000aGetNumOfCaptures(m.Handle())
	response := m.Rsp().(*genericps.GetNumOfCapturesRsp)
	response.NCaptures = numOfCaptures
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

func getNumberOfProcessedCaptures(m *genericps.GetNumOfProcessedCapturesMsg) {
	numOfCaptures, err := ps2000aGetNumOfProcessedCaptures(m.Handle())
	response := m.Rsp().(*genericps.GetNumOfProcessedCapturesRsp)
	response.NCaptures = numOfCaptures
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

func getStreamingLatestValues(m *genericps.GetStreamingLatestValuesMsg) {
	var (
		err error
	)
	err = ps2000aGetStreamingLatestValues(m.Handle(), StreamingReady(m.LpStreamingReadyGoPar), m.Param)
	response := m.Rsp().(*genericps.GetStreamingLatestValuesRsp)
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

func getTimebase(m *genericps.GetTimebaseMsg) {
	timeIntervalNanoseconds, maxSamples, err := ps2000aGetTimebase(m.Handle(), m.TimeBase, m.NumOfSamples, m.OverSample, m.SegmentIndex)
	response := m.Rsp().(*genericps.GetTimebaseRsp)
	response.TimeIntervalNanoseconds = timeIntervalNanoseconds
	response.MaxSamples = maxSamples
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

func getTimebase2(m *genericps.GetTimebase2Msg) {
	timeIntervalNanoseconds, maxSamples, err := ps2000aGetTimebase2(m.Handle(), m.TimeBase, m.NumOfSamples, m.OverSample, m.SegmentIndex)
	response := m.Rsp().(*genericps.GetTimebase2Rsp)
	response.TimeIntervalNanoseconds = timeIntervalNanoseconds
	response.MaxSamples = maxSamples
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

func maximumValue(m *genericps.MaximumValueMsg) {
	value, err := ps2000aMaximumValue(m.Handle())
	response := m.Rsp().(*genericps.MaximumValueRsp)
	response.Value = value
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

func minimumValue(m *genericps.MinimumValueMsg) {
	value, err := ps2000aMinimumValue(m.Handle())
	response := m.Rsp().(*genericps.MinimumValueResp)
	response.Value = value
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

func setSimpleTrigger(m *genericps.SetSimpleTriggerMsg) {
	var (
		err error
	)
	err = ps2000aSetSimpleTrigger(m.Handle(), m.Enable, ChannelId(m.Source),
		m.Threshold, ThresholdDirection(m.Direction), m.Delay, m.AutoTriggerMs)
	response := m.Rsp().(*genericps.SetSimpleTriggerRsp)
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

func setDataBuffer(m *genericps.SetDataBufferMsg) {
	var (
		err error
	)
	err = ps2000aSetDataBuffer(m.Handle(), ChannelId(m.Ch), m.BufferIn, m.SegmentIndex, RatioMode(m.Mode))
	response := m.Rsp().(*genericps.SetDataBufferRsp)
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

func setDataBuffers(m *genericps.SetDataBuffersMsg) {
	var (
		err error
	)
	err = ps2000aSetDataBuffers(m.Handle(), ChannelId(m.Ch), m.BufferMax, m.BufferMin, m.SegmentIndex, RatioMode(m.Mode))
	response := m.Rsp().(*genericps.SetDataBuffersRsp)
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

func setUnscaledDataBuffers(m *genericps.SetUnscaledDataBuffersMsg) {
	var (
		err error
	)
	err = ps2000aSetUnscaledDataBuffers(m.Handle(), ChannelId(m.
		Ch), m.BufferMax, m.BufferMin, m.SegmentIndex, RatioMode(m.Mode))
	response := m.Rsp().(*genericps.SetUnscaledataBuffersRsp)
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

func setEtsTimeBuffer(m *genericps.SetEtsTimeBufferMsg) {
	err := ps2000aSetEtsTimeBuffer(m.Handle(), m.Buffer)
	response := m.Rsp().(*genericps.SetEtsTimeBufferRsp)
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

func setEtsTimeBuffers(m *genericps.SetEtsTimeBuffersMsg) {
	err := ps2000aSetEtsTimeBuffers(m.Handle(), m.TimeUpper, m.TimeLower)
	response := m.Rsp().(*genericps.SetEtsTimeBufferRsp)
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}
func setEts(m *genericps.SetEtsMsg) {
	var (
		err                   error
		sampleTimePicoseconds int32
	)
	sampleTimePicoseconds, err = ps2000aSetEts(m.Handle(), EtsMode(m.Mode), m.EtsCycles, m.EtsInterleave)
	response := m.Rsp().(*genericps.SetEtsRsp)
	response.SampleTimePicoseconds = sampleTimePicoseconds
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

func runStreaming(m *genericps.RunStreamingMsg) {
	var (
		err            error
		sampleInterval uint32
	)
	sampleInterval, err = ps2000aRunStreaming(m.Handle(), m.ReqSampleInterval, TimeUnits(m.SampleIntervalTimeUnits), m.MaxPostTriggerSamples,
		m.MaxPostTriggerSamples, m.AutoStop, m.DownSampleRatio, RatioMode(m.DownSampleRatioMode), m.OverviewBufferSize)
	response := m.Rsp().(*genericps.RunStreamingRsp)
	response.SampleInterval = sampleInterval
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

func runBlock(m *genericps.RunBlockMsg) {
	var (
		err              error
		timeIndisposedMs int32
	)
	timeIndisposedMs, err = ps2000aRunBlock(m.Handle(), m.NumOfPreTriggerSamples, m.NumOfPostTriggerSamples,
		m.TimeBase, m.OverSample, m.SegmentIndex, BlockReady(m.LpBlockReadyGoPar), m.Param)
	response := m.Rsp().(*genericps.RunBlockRsp)
	response.TimeIndisposedMs = timeIndisposedMs
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

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
	slog.Debug("trigger", "tcp", tcp)
	err = ps2000aSetTriggerChannelProperties(m.Handle(), tcp, m.AuxOutputEnable, m.AutoTriggerMs)
	response := m.Rsp().(*genericps.SetTriggerChannelPropertiesRsp)
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

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
	err = ps2000aSetTriggerChannelConditions(m.Handle(), tc)
	response := m.Rsp().(*genericps.SetTriggerChannelConditionsRsp)
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

func setTriggerChannelDirections(m *genericps.SetTriggerChannelDirectionsMsg) {
	var (
		err error
	)
	err = ps2000aSetTriggerChannelDirections(m.Handle(), ThresholdDirection(m.ChannelA),
		ThresholdDirection(m.ChannelB), ThresholdDirection(m.ChannelC),
		ThresholdDirection(m.ChannelD), ThresholdDirection(m.Ext), ThresholdDirection(m.Aux))
	response := m.Rsp().(*genericps.SetTriggerChannelDirectionsRsp)
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

func setTriggerDelay(m *genericps.SetTriggerDelayMsg) {
	err := ps2000aSetTriggerDelay(m.Handle(), m.Delay)
	response := m.Rsp().(*genericps.SetTriggerDelayRsp)
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

func setPulseWidthQualifier(m *genericps.SetPulseWidthQualifierMsg) {
	var (
		err error
	)
	c := make([]PwqConditions, len(m.Conditions))
	for i := range c {
		c[i].Aux = TriggerState(m.Conditions[i].Aux)
		c[i].ChannelA = TriggerState(m.Conditions[i].ChannelA)
		c[i].ChannelB = TriggerState(m.Conditions[i].ChannelB)
		c[i].ChannelC = TriggerState(m.Conditions[i].ChannelC)
		c[i].ChannelD = TriggerState(m.Conditions[i].ChannelD)
	}
	err = ps2000aSetPulseWidthQualifier(m.Handle(), c, ThresholdDirection(m.Direction), m.Lower, m.Upper, PulseWidthType(m.PwType))
	response := m.Rsp().(*genericps.SetPulseWidthQualifierRsp)
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

func setTriggerDigitalPortProperties(m *genericps.SetTriggerDigitalPortPropertiesMsg) {
	var (
		err error
	)
	tdp := make([]DigitalChannelDirections, len(m.DigitalDirections))
	for i := range tdp {
		tdp[i].Channel = DigitalChannel(m.DigitalDirections[i].Channel)
		tdp[i].Direction = DigitalDirection(m.DigitalDirections[i].Direction)
	}
	err = ps2000aSetTriggerDigitalPortProperties(m.Handle(), tdp)
	response := m.Rsp().(*genericps.SetTriggerDigitalPortPropertiesRsp)
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

func stop(m *genericps.StopMsg) {
	err := ps2000aStop(m.Handle())
	response := m.Rsp().(*genericps.StopRsp)
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

func setSigGenBuiltIn(m *genericps.SetSigGenBuiltInMsg) {
	var (
		err error
	)
	err = ps2000aSetSigGenBuiltIn(m.Handle(), m.OffsetVoltage, m.PkToPK, WaveTypeEnum(m.WaveType), m.StartFrequency,
		m.StopFrequency, m.Increment, m.DwellTime, SweepTypeEnum(m.SweepType),
		ExtraOperations(m.Operation), m.Shots, m.Sweeps, SigGenTrigType(m.TriggerType),
		SigGenTrigSource(m.TriggerSource), m.ExtInThreshold)
	response := m.Rsp().(*genericps.SetSigGenBuiltInRsp)
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

func setSigGenBuiltInV2(m *genericps.SetSigGenBuiltInV2Msg) {
	var (
		err error
	)
	err = ps2000aSetSigGenBuiltInV2(m.Handle(), m.OffsetVoltage, m.PkToPK, WaveTypeEnum(m.WaveType), m.StartFrequency,
		m.StopFrequency, m.Increment, m.DwellTime, SweepTypeEnum(m.SweepType),
		ExtraOperations(m.Operation), m.Shots, m.Sweeps, SigGenTrigType(m.TriggerType),
		SigGenTrigSource(m.TriggerSource), m.ExtInThreshold)
	response := m.Rsp().(*genericps.SetSigGenBuiltInV2Rsp)
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

func sigGenFrequencyToPhase(m *genericps.SigGenFrequencyToPhasenMsg) {
	var (
		err   error
		phase uint32
	)
	phase, err = ps2000aSigGenFrequencyToPhase(m.Handle(), m.Frequency, IndexMode(m.IndexMode), m.BufferLength)
	response := m.Rsp().(*genericps.SigGenFrequencyToPhaseRsp)
	response.SetStatus(err)
	response.Phase = phase
	m.RspCh() <- struct{}{}
}

func setNumOfCaptures(m *genericps.SetNumOfCapturesMsg) {
	err := ps2000aSetNoCaptures(m.Handle(), m.NCaptures)
	response := m.Rsp().(*genericps.SetNumOfCapturesRsp)
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

func getTriggerTimeOffset(m *genericps.GetTriggerTimeOffsetMsg) {
	var (
		err                  error
		timeUpper, timeLower uint32
		timeUnits            TimeUnits
	)
	timeUpper, timeLower, timeUnits, err = ps2000aGetTriggerTimeOffset(m.Handle(), m.SegmentIndex)
	response := m.Rsp().(*genericps.GetTriggerTimeOffsetRsp)
	response.TimeLower = timeLower
	response.TimeUnits = genericps.TimeUnits(timeUnits)
	response.TimeUpper = timeUpper
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

func getTriggerTimeOffset64(m *genericps.GetTriggerTimeOffset64Msg) {
	time, timeUnits, err := ps2000aGetTriggerTimeOffset64(m.Handle(), m.SegmentIndex)
	response := m.Rsp().(*genericps.GetTriggerTimeOffset64Rsp)
	response.Time = time
	response.TimeUnits = genericps.TimeUnits(timeUnits)
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

func getValuesTriggerTimeOffsetBulk(m *genericps.GetValuesTriggerTimeOffsetBulkMsg) {
	var (
		err error
	)
	tu := make([]TimeUnits, len(m.TimeUnits))
	for i := range tu {
		tu[i] = TimeUnits(m.TimeUnits[i])
	}
	err = ps2000aGetValuesTriggerTimeOffsetBulk(m.Handle(), m.TimesUpper, m.TimesLower, tu, m.FromSegmentIndex, m.ToSegmentIndex)
	response := m.Rsp().(*genericps.GetValuesTriggerTimeOffsetBulkRsp)
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

func getValuesTriggerTimeOffsetBulk64(m *genericps.GetValuesTriggerTimeOffsetBulk64Msg) {
	var (
		err error
	)
	tu := make([]TimeUnits, len(m.TimeUnits))
	for i := range tu {
		tu[i] = TimeUnits(m.TimeUnits[i])
	}
	err = ps2000aGetValuesTriggerTimeOffsetBulk64(m.Handle(), m.Times, tu, m.FromSegmentIndex, m.ToSegmentIndex)
	response := m.Rsp().(*genericps.GetValuesTriggerTimeOffsetBulk64Rsp)
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

func holdOff(m *genericps.HoldOffMsg) {
	err := ps2000aHoldOff(m.Handle(), m.HoldOff, HoldOffType(m.HoldOffType))
	response := m.Rsp().(*genericps.HoldOffRsp)
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

func lsReady(m *genericps.LsReadyMsg) {
	ready, err := ps2000aIsReady(m.Handle())
	response := m.Rsp().(*genericps.LsReadyRsp)
	response.Ready = ready
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

func triggerOrPulseWidthQualifierEnabled(m *genericps.TriggerOrPulseWidthQualifierEnabledMsg) {
	triggerEnabled, pulseWidthQualifierEnabledint16, err := ps2000aTriggerOrPulseWidthQualifierEnabled(m.Handle())
	response := m.Rsp().(*genericps.TriggerOrPulseWidthQualifierEnabledRsp)
	response.TriggerEnabled = triggerEnabled
	response.PulseWidthQualifierEnabledint16 = pulseWidthQualifierEnabledint16
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

func memorySegments(m *genericps.MemorySegmentsMsg) {
	nMaxSamples, err := ps2000aMemorySegments(m.Handle(), m.NSegments)
	response := m.Rsp().(*genericps.MemorySegmentsRsp)
	response.SetStatus(err)
	response.NMaxSamples = nMaxSamples
	m.RspCh() <- struct{}{}
}

func numOfStreamingValues(m *genericps.NumOfStreamingValuesMsg) {
	numOfValues, err := ps2000aNumOfStreamingValues(m.Handle())
	response := m.Rsp().(*genericps.NumOfStreamingValuesRsp)
	response.NumOfValues = numOfValues
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

func pingUnit(m *genericps.PingUnitMsg) {
	err := ps2000aPingUnit(m.Handle())
	response := m.Rsp().(*genericps.PingUnitRsp)
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

func queryOutputEdgeDetect(m *genericps.QueryOutputEdgeDetectMsg) {
	state, err := ps2000aQueryOutputEdgeDetect(m.Handle())
	response := m.Rsp().(*genericps.QueryOutputEdgeDetectRsp)
	response.State = state
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

func setDigitalAnalogTriggerOperand(m *genericps.SetDigitalAnalogTriggerOperandMsg) {
	var (
		err error
	)
	err = ps2000aSetDigitalAnalogTriggerOperand(m.Handle(), TriggerOperand(m.Operand))
	response := m.Rsp().(*genericps.SetDigitalAnalogTriggerOperandRsp)
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

func setDigitalPort(m *genericps.SetDigitalPortMsg) {
	var (
		err error
	)
	err = ps2000aSetDigitalPort(m.Handle(), DigitalPort(m.Port), m.Enabled, m.Logiclevel)
	response := m.Rsp().(*genericps.SetDigitalAnalogTriggerOperandRsp)
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

func setOutputEdgeDetect(m *genericps.SetOutputEdgeDetectMsg) {
	err := ps2000aSetOutputEdgeDetect(m.Handle(), m.State)
	response := m.Rsp().(*genericps.SetOutputEdgeDetectRsp)
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

func setPulseWidthDigitalPortProperties(m *genericps.SetPulseWidthDigitalPortPropertiesMsg) {
	var (
		err error
	)
	dd := make([]DigitalChannelDirections, len(m.DigitalDirections))
	for i := range dd {
		dd[i].Channel = DigitalChannel(m.DigitalDirections[i].Channel)
		dd[i].Direction = DigitalDirection(m.DigitalDirections[i].Direction)
	}
	err = ps2000aSetPulseWidthDigitalPortProperties(m.Handle(), dd)
	response := m.Rsp().(*genericps.SetPulseWidthDigitalPortPropertiesRsp)
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

func setSigGenArbitrary(m *genericps.SetSigGenArbitraryMsg) {
	var (
		err error
	)
	err = ps2000aSetSigGenArbitrary(m.Handle(), m.OffsetVoltage, m.PkToPK, m.StartDeltaPhase,
		m.StopDeltaPhase, m.DeltaPhaseIncrement, m.DwellCount, m.ArbitraryWaveform,
		SweepTypeEnum(m.SweepType), ExtraOperations(m.Operation),
		IndexMode(m.IndexMode), m.Shots, m.Sweeps,
		SigGenTrigType(m.TtriggerType),
		SigGenTrigSource(m.TriggerSource), m.ExtInThreshold)
	response := m.Rsp().(*genericps.SetSigGenArbitraryRsp)
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

func setSigGenPropertiesArbitrary(m *genericps.SetSigGenPropertiesArbitraryMsg) {
	err := ps2000aSetSigGenPropertiesArbitrary(m.Handle(), m.OffsetVoltage, m.StartDeltaPhase,
		m.StopDeltaPhase, m.DeltaPhaseIncrement, m.DwellCount, SweepTypeEnum(m.SweepType),
		ExtraOperations(m.Operation), IndexMode(m.IndexMode), m.Shots, m.Sweeps,
		SigGenTrigType(m.TriggerType),
		SigGenTrigSource(m.TriggerSource), m.ExtInThreshold)
	response := m.Rsp().(*genericps.SetSigGenPropertiesArbitraryRsp)
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

func setSigGenPropertiesBuiltIn(m *genericps.SetSigGenPropertiesBuiltInMsg) {
	err := ps2000aSetSigGenPropertiesBuiltIn(m.Handle(), m.OffsetVoltage, m.StartFrequency,
		m.StopFrequency, m.Increment, m.DwellTime, SweepTypeEnum(m.SweepType),
		m.Shots, m.Sweeps,
		SigGenTrigType(m.TriggerType), SigGenTrigSource(m.TriggerSource), m.ExtInThreshold)
	response := m.Rsp().(*genericps.SetSigGenPropertiesBuiltInRsp)
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

func sigGenArbitraryMinMaxValues(m *genericps.SigGenArbitraryMinMaxValuesMsg) {
	minArbitraryWaveformValue, maxArbitraryWaveformValue, minArbitraryWaveformSize,
		maxArbitraryWaveformSize, err := ps2000aSigGenArbitraryMinMaxValues(m.Handle())
	response := m.Rsp().(*genericps.SigGenArbitraryMinMaxValuesRsp)
	response.MinArbitraryWaveformValue = minArbitraryWaveformValue
	response.MaxArbitraryWaveformValue = maxArbitraryWaveformValue
	response.MinArbitraryWaveformSize = minArbitraryWaveformSize
	response.MaxArbitraryWaveformSize = maxArbitraryWaveformSize
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

func sigGenSoftwareControl(m *genericps.SigGenSoftwareControlMsg) {
	err := ps2000aSigGenSoftwareControl(m.Handle(), m.State)
	response := m.Rsp().(*genericps.SigGenSoftwareControlRsp)
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

func dispatch(msg genericps.Message) {
	switch m := msg.(type) {
	case *genericps.NullMsg:
		m.RspCh() <- struct{}{}
	case *genericps.SetChannelMsg:
		setChannel(m)
	case *genericps.GetUnitInfoMsg:
		getUnitInfo(m)
	case *genericps.CloseUnitMsg:
		closeUnit(m)
	case *genericps.FlashLedMsg:
		flashLed(m)
	case *genericps.GetValuesAsyncMsg:
		getValuesAsync(m)
	case *genericps.GetValuesBulkMsg:
		getValuesBulk(m)
	case *genericps.GetValuesOverlappedBulkMsg:
		getValuesOverlappedBulk(m)
	case *genericps.GetValuesOverlappedMsg:
		getValuesOverlapped(m)
	case *genericps.GetValuesMsg:
		getValues(m)
	case *genericps.GetAnalogueOffsetMsg:
		getAnalogueOffsetMsg(m)
	case *genericps.GetChannelInformationMsg:
		getChannelInformation(m)
	case *genericps.GetMaxDownSampleRatioMsg:
		getMaxDownSampleRatio(m)
	case *genericps.GetMaxSegmentsMsg:
		getMaxSegments(m)
	case *genericps.GetNumOfCapturesMsg:
		getNumberOfCaptures(m)
	case *genericps.GetNumOfProcessedCapturesMsg:
		getNumberOfProcessedCaptures(m)
	case *genericps.GetTimebaseMsg:
		getTimebase(m)
	case *genericps.GetTimebase2Msg:
		getTimebase2(m)
	case *genericps.MaximumValueMsg:
		maximumValue(m)
	case *genericps.MinimumValueMsg:
		minimumValue(m)
	case *genericps.SetSimpleTriggerMsg:
		setSimpleTrigger(m)
	case *genericps.SetDataBufferMsg:
		setDataBuffer(m)
	case *genericps.SetDataBuffersMsg:
		setDataBuffers(m)
	case *genericps.SetUnscaledDataBuffersMsg:
		setUnscaledDataBuffers(m)
	case *genericps.SetEtsTimeBufferMsg:
		setEtsTimeBuffer(m)
	case *genericps.SetEtsTimeBuffersMsg:
		setEtsTimeBuffers(m)
	case *genericps.SetEtsMsg:
		setEts(m)
	case *genericps.RunStreamingMsg:
		runStreaming(m)
	case *genericps.GetStreamingLatestValuesMsg:
		getStreamingLatestValues(m)
	case *genericps.RunBlockMsg:
		runBlock(m)
	case *genericps.SetTriggerChannelPropertiesMsg:
		setTriggerChannelProperties(m)
	case *genericps.SetTriggerChannelConditionsMsg:
		setTriggerChannelConditions(m)
	case *genericps.SetTriggerChannelDirectionsMsg:
		setTriggerChannelDirections(m)
	case *genericps.SetTriggerDelayMsg:
		setTriggerDelay(m)
	case *genericps.SetPulseWidthQualifierMsg:
		setPulseWidthQualifier(m)
	case *genericps.SetTriggerDigitalPortPropertiesMsg:
		setTriggerDigitalPortProperties(m)
	case *genericps.StopMsg:
		stop(m)
	case *genericps.SetSigGenBuiltInMsg:
		setSigGenBuiltIn(m)
	case *genericps.SetSigGenBuiltInV2Msg:
		setSigGenBuiltInV2(m)
	case *genericps.SigGenFrequencyToPhasenMsg:
		sigGenFrequencyToPhase(m)
	case *genericps.SetNumOfCapturesMsg:
		setNumOfCaptures(m)
	case *genericps.GetTriggerTimeOffsetMsg:
		getTriggerTimeOffset(m)
	case *genericps.GetTriggerTimeOffset64Msg:
		getTriggerTimeOffset64(m)
	case *genericps.GetValuesTriggerTimeOffsetBulkMsg:
		getValuesTriggerTimeOffsetBulk(m)
	case *genericps.GetValuesTriggerTimeOffsetBulk64Msg:
		getValuesTriggerTimeOffsetBulk64(m)
	case *genericps.HoldOffMsg:
		holdOff(m)
	case *genericps.LsReadyMsg:
		lsReady(m)
	case *genericps.TriggerOrPulseWidthQualifierEnabledMsg:
		triggerOrPulseWidthQualifierEnabled(m)
	case *genericps.MemorySegmentsMsg:
		memorySegments(m)
	case *genericps.NumOfStreamingValuesMsg:
		numOfStreamingValues(m)
	case *genericps.PingUnitMsg:
		pingUnit(m)
	case *genericps.QueryOutputEdgeDetectMsg:
		queryOutputEdgeDetect(m)
	case *genericps.SetDigitalAnalogTriggerOperandMsg:
		setDigitalAnalogTriggerOperand(m)
	case *genericps.SetDigitalPortMsg:
		setDigitalPort(m)
	case *genericps.SetOutputEdgeDetectMsg:
		setOutputEdgeDetect(m)
	case *genericps.SetPulseWidthDigitalPortPropertiesMsg:
		setPulseWidthDigitalPortProperties(m)
	case *genericps.SetSigGenArbitraryMsg:
		setSigGenArbitrary(m)
	case *genericps.SetSigGenPropertiesArbitraryMsg:
		setSigGenPropertiesArbitrary(m)
	case *genericps.SetSigGenPropertiesBuiltInMsg:
		setSigGenPropertiesBuiltIn(m)
	case *genericps.SigGenArbitraryMinMaxValuesMsg:
		sigGenArbitraryMinMaxValues(m)
	case *genericps.SigGenSoftwareControlMsg:
		sigGenSoftwareControl(m)
	default:
		slog.Error("dispatch unhandled", "type", reflect.TypeOf(msg), "msg", msg)
	}
}
