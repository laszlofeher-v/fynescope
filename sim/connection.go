package sim

import (
	// "reflect"
	// "fmt"
	"log/slog"
	"fynescope/genericps"
	"time"
)

const (
	cmdSendTimeout         = 5000 * time.Millisecond
	responseReceiveTimeout = 5000 * time.Millisecond
)

func Stop() (err error) {
	// err = s.ps.Stop()
	return
}

func getValues(m *genericps.GetValuesMsg) {
	var (
		numOfSamples uint32
		overflow     int16
		err          error
	)
	// numOfSamples, overflow, err = m.Connection().Ps().GetValues(m.StartIndex,
	numOfSamples, overflow, err = simGetValues(m.Handle(), m.StartIndex,
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
	err := simCloseUnit(m.Handle())
	response := m.Rsp().(*genericps.CloseUnitRsp)
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

func setChannel(m *genericps.SetChannelMsg) {
	var err error
	err = simSetChannel(m.Handle(), ChannelId(m.Channel),
		m.Enabled, Coupling(m.CouplingType),
		RangeEnum(m.VoltageRange), m.AnalogOffset)
	response := m.Rsp().(*genericps.SetChannelRsp)
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

func flashLed(m *genericps.FlashLedMsg) {
	err := simFlashLed(m.Handle(), m.Start)
	response := m.Rsp().(*genericps.FlashLedRsp)
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

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

func getValuesAsync(m *genericps.GetValuesAsyncMsg) {
	var err error
	err = simGetValuesAsync(m.Handle(), m.StartIndex,
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
	numOfSamples, err = simGetValuesBulk(m.Handle(), m.ReqNumOfSamples, m.FromSegmentIndex, m.ToSegmentIndex,
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
	numOfSamples, err = simGetValuesOverlapped(m.Handle(), m.StartIndex, m.ReqNumOfSamples,
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
	numOfSamples, err = simGetValuesOverlappedBulk(m.Handle(), m.StartIndex, m.ReqNumOfSamples,
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
	maximumVoltage, minimumVoltage, err = simGetAnalogueOffset(m.Handle(), m.VoltageRange, Coupling(m.Coupling))
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
	lengthOfRanges, err = simGetChannelInformation(m.Handle(), m.Info, m.Probe, m.Ranges, ChannelId(m.Channel))
	m.Ranges = m.Ranges[0:lengthOfRanges]
	// response := &GetChannelInformationRsp{Stat: Stat{err: err}, Ranges: m.Ranges, LengthOfRanges: lengthOfRanges}
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
	maxDownSampleRatio, err = simGetMaxDownSampleRatio(m.Handle(), m.NumOfUnaggregatedSamples, RatioMode(m.DownSampleRatioMode), m.SegmentIndex)
	response := m.Rsp().(*genericps.GetMaxDownSampleRatioRsp)
	response.MaxDownSampleRatio = maxDownSampleRatio
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

func getMaxSegments(m *genericps.GetMaxSegmentsMsg) {
	maxSegments, err := simGetMaxSegments(m.Handle())
	response := m.Rsp().(*genericps.GetMaxSegmentsRsp)
	response.MaxSegments = maxSegments
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

func getNumberOfCaptures(m *genericps.GetNumOfCapturesMsg) {
	numOfCaptures, err := simGetNoOfCaptures(m.Handle())
	response := m.Rsp().(*genericps.GetNumOfCapturesRsp)
	response.NCaptures = numOfCaptures
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

func getNumberOfProcessedCaptures(m *genericps.GetNumOfProcessedCapturesMsg) {
	numOfCaptures, err := simGetNoOfProcessedCaptures(m.Handle())
	// response := &GetNumOfProcessedCapturesRsp{Stat: Stat{err: err}, NCaptures: numOfCaptures}
	response := m.Rsp().(*genericps.GetNumOfProcessedCapturesRsp)
	response.NCaptures = numOfCaptures
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

func getStreamingLatestValues(m *genericps.GetStreamingLatestValuesMsg) {
	var (
		err error
	)
	err = simGetStreamingLatestValues(m.Handle(), StreamingReady(m.LpStreamingReadyGoPar), m.Param)
	// response := &GetStreamingLatestValuesRsp{Stat: Stat{err: err}}
	response := m.Rsp().(*genericps.GetStreamingLatestValuesRsp)
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

func getTimebase(m *genericps.GetTimebaseMsg) {
	timeIntervalNanoseconds, maxSamples, err := simGetTimebase(m.Handle(), m.TimeBase, m.NumOfSamples, m.OverSample, m.SegmentIndex)
	// response := &GetTimebaseRsp{Stat: Stat{err: err}, TimeIntervalNanoseconds: timeIntervalNanoseconds, MaxSamples: maxSamples}
	response := m.Rsp().(*genericps.GetTimebaseRsp)
	response.TimeIntervalNanoseconds = timeIntervalNanoseconds
	response.MaxSamples = maxSamples
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

func getTimebase2(m *genericps.GetTimebase2Msg) {
	timeIntervalNanoseconds, maxSamples, err := simGetTimebase2(m.Handle(), m.TimeBase, m.NumOfSamples, m.OverSample, m.SegmentIndex)
	// response := &GetTimebase2Rsp{Stat: Stat{err: err}, TimeIntervalNanoseconds: timeIntervalNanoseconds, MaxSamples: maxSamples}
	response := m.Rsp().(*genericps.GetTimebase2Rsp)
	response.TimeIntervalNanoseconds = timeIntervalNanoseconds
	response.MaxSamples = maxSamples
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

func maximumValue(m *genericps.MaximumValueMsg) {
	value, err := simMaximumValue(m.Handle())
	// response := &MaximumValueRsp{Stat: Stat{err: err}, Value: value}
	response := m.Rsp().(*genericps.MaximumValueRsp)
	response.Value = value
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

func minimumValue(m *genericps.MinimumValueMsg) {
	value, err := simMinimumValue(m.Handle())
	// response := &MinimumValueResp{Stat: Stat{err: err}, Value: value}
	response := m.Rsp().(*genericps.MinimumValueResp)
	response.Value = value
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

func setSimpleTrigger(m *genericps.SetSimpleTriggerMsg) {
	var (
		err error
	)
	err = simSetSimpleTrigger(m.Handle(), m.Enable, ChannelId(m.Source),
		m.Threshold, ThresholdDirection(m.Direction), m.Delay, m.AutoTriggerMs)
	// response := &SetSimpleTriggerRsp{Stat: Stat{err: err}}
	response := m.Rsp().(*genericps.SetSimpleTriggerRsp)
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

func setDataBuffer(m *genericps.SetDataBufferMsg) {
	var (
		err error
	)
	err = simSetDataBuffer(m.Handle(), ChannelId(m.Ch), m.BufferIn, m.SegmentIndex, RatioMode(m.Mode))
	// response := &SetDataBufferRsp{Stat: Stat{err: err}}
	response := m.Rsp().(*genericps.SetDataBufferRsp)
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

func setDataBuffers(m *genericps.SetDataBuffersMsg) {
	var (
		err error
	)
	err = simSetDataBuffers(m.Handle(), ChannelId(m.Ch), m.BufferMax, m.BufferMin, m.SegmentIndex, RatioMode(m.Mode))
	// response := &SetDataBuffersRsp{Stat: Stat{err: err}}
	response := m.Rsp().(*genericps.SetDataBuffersRsp)
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

func setUnscaledDataBuffers(m *genericps.SetUnscaledDataBuffersMsg) {
	var (
		err error
	)
	err = simSetUnscaledDataBuffers(m.Handle(), ChannelId(m.
		Ch), m.BufferMax, m.BufferMin, m.SegmentIndex, RatioMode(m.Mode))
	// response := &SetUnscaledataBuffersRsp{Stat: Stat{err: err}}
	response := m.Rsp().(*genericps.SetUnscaledataBuffersRsp)
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

func setEtsTimeBuffer(m *genericps.SetEtsTimeBufferMsg) {
	err := simSetEtsTimeBuffer(m.Handle(), m.Buffer)
	// response := &SetEtsTimeBufferRsp{Stat: Stat{err: err}}
	response := m.Rsp().(*genericps.SetEtsTimeBufferRsp)
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

func setEtsTimeBuffers(m *genericps.SetEtsTimeBuffersMsg) {
	err := simSetEtsTimeBuffers(m.Handle(), m.TimeUpper, m.TimeLower)
	// response := &SetEtsTimeBufferRsp{Stat: Stat{err: err}}
	response := m.Rsp().(*genericps.SetEtsTimeBufferRsp)
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}
func setEts(m *genericps.SetEtsMsg) {
	var (
		err                   error
		sampleTimePicoseconds int32
	)
	sampleTimePicoseconds, err = simSetEts(m.Handle(), EtsMode(m.Mode), m.EtsCycles, m.EtsInterleave)
	// response := &SetEtsRsp{Stat: Stat{err: err}, SampleTimePicoseconds: sampleTimePicoseconds}
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
	sampleInterval, err = simRunStreaming(m.Handle(), m.ReqSampleInterval, TimeUnits(m.SampleIntervalTimeUnits), m.MaxPostTriggerSamples,
		m.MaxPostTriggerSamples, m.AutoStop, m.DownSampleRatio, RatioMode(m.DownSampleRatioMode), m.OverviewBufferSize)
	// response := &RunStreamingRsp{Stat: Stat{err: err}, SampleInterval: sampleInterval}
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
	timeIndisposedMs, err = simRunBlock(m.Handle(), m.NumOfPreTriggerSamples, m.NumOfPostTriggerSamples,
		m.TimeBase, m.OverSample, m.SegmentIndex, BlockReady(m.LpBlockReadyGoPar), m.Param)
	// response := &RunBlockRsp{Stat: Stat{err: err}, TimeIndisposedMs: timeIndisposedMs}
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
	err = simSetTriggerChannelProperties(m.Handle(), tcp, m.AuxOutputEnable, m.AutoTriggerMs)
	// response := &SetTriggerChannelPropertiesRsp{Stat: Stat{err: err}}
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
	err = simSetTriggerChannelConditions(m.Handle(), tc)
	// response := &SetTriggerChannelConditionsRsp{Stat: Stat{err: err}}
	response := m.Rsp().(*genericps.SetTriggerChannelConditionsRsp)
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

func setTriggerChannelDirections(m *genericps.SetTriggerChannelDirectionsMsg) {
	var (
		err error
	)
	err = simSetTriggerChannelDirections(m.Handle(), ThresholdDirection(m.ChannelA),
		ThresholdDirection(m.ChannelB), ThresholdDirection(m.ChannelC),
		ThresholdDirection(m.ChannelD), ThresholdDirection(m.Ext), ThresholdDirection(m.Aux))
	// response := &SetTriggerChannelDirectionsRsp{Stat: Stat{err: err}}
	response := m.Rsp().(*genericps.SetTriggerChannelDirectionsRsp)
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

func setTriggerDelay(m *genericps.SetTriggerDelayMsg) {
	err := simSetTriggerDelay(m.Handle(), m.Delay)
	// response := &SetTriggerDelayRsp{Stat: Stat{err: err}}
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
	err = simSetPulseWidthQualifier(m.Handle(), c, ThresholdDirection(m.Direction), m.Lower, m.Upper, PulseWidthType(m.PwType))
	// response := &SetPulseWidthQualifierRsp{Stat: Stat{err: err}}
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
	err = simSetTriggerDigitalPortProperties(m.Handle(), tdp)
	// response := &SetTriggerDigitalPortPropertiesRsp{Stat: Stat{err: err}}
	response := m.Rsp().(*genericps.SetTriggerDigitalPortPropertiesRsp)
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

func stop(m *genericps.StopMsg) {
	err := simStop(m.Handle())
	// response := &StopRsp{Stat: Stat{err: err}}
	response := m.Rsp().(*genericps.StopRsp)
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

func setSigGenBuiltIn(m *genericps.SetSigGenBuiltInMsg) {
	var (
		err error
	)
	err = simSetSigGenBuiltIn(m.Handle(), m.OffsetVoltage, m.PkToPK, WaveTypeEnum(m.WaveType), m.StartFrequency,
		m.StopFrequency, m.Increment, m.DwellTime, SweepTypeEnum(m.SweepType),
		ExtraOperations(m.Operation), m.Shots, m.Sweeps, SigGenTrigType(m.TriggerType),
		SigGenTrigSource(m.TriggerSource), m.ExtInThreshold)
	// response := &SetSigGenBuiltInRsp{Stat: Stat{err: err}}
	response := m.Rsp().(*genericps.SetSigGenBuiltInRsp)
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

func setSigGenBuiltInV2(m *genericps.SetSigGenBuiltInV2Msg) {
	var (
		err error
	)
	err = simSetSigGenBuiltInV2(m.Handle(), m.OffsetVoltage, m.PkToPK, WaveTypeEnum(m.WaveType), m.StartFrequency,
		m.StopFrequency, m.Increment, m.DwellTime, SweepTypeEnum(m.SweepType),
		ExtraOperations(m.Operation), m.Shots, m.Sweeps, SigGenTrigType(m.TriggerType),
		SigGenTrigSource(m.TriggerSource), m.ExtInThreshold)
	// response := &SetSigGenBuiltInV2Rsp{Stat: Stat{err: err}}
	response := m.Rsp().(*genericps.SetSigGenBuiltInV2Rsp)
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

func sigGenFrequencyToPhase(m *genericps.SigGenFrequencyToPhasenMsg) {
	var (
		err   error
		phase uint32
	)
	phase, err = simSigGenFrequencyToPhase(m.Handle(), m.Frequency, IndexMode(m.IndexMode), m.BufferLength)
	// response := &SigGenFrequencyToPhaseRsp{Stat: Stat{err: err}, Phase: phase}
	response := m.Rsp().(*genericps.SigGenFrequencyToPhaseRsp)
	response.SetStatus(err)
	response.Phase = phase
	m.RspCh() <- struct{}{}
}

func setNumOfCaptures(m *genericps.SetNumOfCapturesMsg) {
	err := simSetNoCaptures(m.Handle(), m.NCaptures)
	// response := &SetNumOfCapturesRsp{Stat: Stat{err: err}}
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
	timeUpper, timeLower, timeUnits, err = simGetTriggerTimeOffset(m.Handle(), m.SegmentIndex)
	// response := &GetTriggerTimeOffsetRsp{Stat: Stat{err: err}, TimeUpper: timeUpper,
	// TimeLower: timeLower, TimeUnits: TimeUnits(timeUnits)}
	response := m.Rsp().(*genericps.GetTriggerTimeOffsetRsp)
	response.TimeLower = timeLower
	response.TimeUnits = genericps.TimeUnits(timeUnits)
	response.TimeUpper = timeUpper
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

func getTriggerTimeOffset64(m *genericps.GetTriggerTimeOffset64Msg) {
	time, timeUnits, err := simGetTriggerTimeOffset64(m.Handle(), m.SegmentIndex)
	// response := GetTriggerTimeOffset64RspPool.Get().(*GetTriggerTimeOffset64Rsp)
	// response.Stat.err = err
	// response.Time = time
	// response.TimeUnits = TimeUnits(timeUnits)
	// response := &GetTriggerTimeOffset64Rsp{Stat: Stat{err: err}, Time: time, TimeUnits: timeUnits}
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
	err = simGetValuesTriggerTimeOffsetBulk(m.Handle(), m.TimesUpper, m.TimesLower, tu, m.FromSegmentIndex, m.ToSegmentIndex)
	// response := &GetValuesTriggerTimeOffsetBulkRsp{Stat: Stat{err: err}}
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
	err = simGetValuesTriggerTimeOffsetBulk64(m.Handle(), m.Times, tu, m.FromSegmentIndex, m.ToSegmentIndex)
	// response := &GetValuesTriggerTimeOffsetBulk64Rsp{Stat: Stat{err: err}}
	response := m.Rsp().(*genericps.GetValuesTriggerTimeOffsetBulk64Rsp)
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

func holdOff(m *genericps.HoldOffMsg) {
	err := simHoldOff(m.Handle(), m.HoldOff, HoldOffType(m.HoldOffType))
	// response := &HoldOffRsp{Stat: Stat{err: err}}
	response := m.Rsp().(*genericps.HoldOffRsp)
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}
func simIsReady(handle int16) (ready int16, err error) {
	ready = 1
	return
}

func lsReady(m *genericps.LsReadyMsg) {
	ready, err := simIsReady(m.Handle())
	// response := &LsReadyRsp{Stat: Stat{err: err}, Ready: ready}
	response := m.Rsp().(*genericps.LsReadyRsp)
	response.Ready = ready
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

func triggerOrPulseWidthQualifierEnabled(m *genericps.TriggerOrPulseWidthQualifierEnabledMsg) {
	triggerEnabled, pulseWidthQualifierEnabledint16, err := simTriggerOrPulseWidthQualifierEnabled(m.Handle())
	// response := &TriggerOrPulseWidthQualifierEnabledRsp{Stat: Stat{err: err},
	// 	TriggerEnabled: triggerEnabled, PulseWidthQualifierEnabledint16: pulseWidthQualifierEnabledint16}
	response := m.Rsp().(*genericps.TriggerOrPulseWidthQualifierEnabledRsp)
	response.TriggerEnabled = triggerEnabled
	response.PulseWidthQualifierEnabledint16 = pulseWidthQualifierEnabledint16
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

func memorySegments(m *genericps.MemorySegmentsMsg) {
	nMaxSamples, err := simMemorySegments(m.Handle(), m.NSegments)
	// response := &MemorySegmentsRsp{Stat: Stat{err: err}, NMaxSamples: nMaxSamples}
	response := m.Rsp().(*genericps.MemorySegmentsRsp)
	response.SetStatus(err)
	response.NMaxSamples = nMaxSamples
	m.RspCh() <- struct{}{}
}

func numOfStreamingValues(m *genericps.NumOfStreamingValuesMsg) {
	numOfValues, err := simNoOfStreamingValues(m.Handle())
	// response := &NumOfStreamingValuesRsp{Stat: Stat{err: err}, NumOfValues: numOfValues}
	response := m.Rsp().(*genericps.NumOfStreamingValuesRsp)
	response.NumOfValues = numOfValues
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

// func openUnitProgress(m *genericps.OpenUnitProgressMsg) {
// 	retHandle, progressPercent, complete, err := ps2000aOpenUnitProgress(m.Handle())
// 	// response := &OpenUnitProgressRsp{Stat: Stat{err: err}, RetHandle: retHandle, ProgressPercent: progressPercent, Complete: complete}
// 	response := m.Rsp().(*genericps.OpenUnitProgressRsp)
// 	response.RetHandle = retHandle
// 	response.ProgressPercent = progressPercent
// 	response.Complete = complete
// 	response.SetStatus(err)
// 	m.RspCh() <- struct{}{}
// }

func pingUnit(m *genericps.PingUnitMsg) {
	err := simPingUnit(m.Handle())
	// response := &PingUnitRsp{Stat: Stat{err: err}}
	response := m.Rsp().(*genericps.PingUnitRsp)
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

func queryOutputEdgeDetect(m *genericps.QueryOutputEdgeDetectMsg) {
	state, err := simQueryOutputEdgeDetect(m.Handle())
	// response := &QueryOutputEdgeDetectRsp{Stat: Stat{err: err}, State: state}
	response := m.Rsp().(*genericps.QueryOutputEdgeDetectRsp)
	response.State = state
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

func setDigitalAnalogTriggerOperand(m *genericps.SetDigitalAnalogTriggerOperandMsg) {
	var (
		err error
	)
	err = simSetDigitalAnalogTriggerOperand(m.Handle(), TriggerOperand(m.Operand))
	// response := &SetDigitalAnalogTriggerOperandRsp{Stat: Stat{err: err}}
	response := m.Rsp().(*genericps.SetDigitalAnalogTriggerOperandRsp)
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

func setDigitalPort(m *genericps.SetDigitalPortMsg) {
	var (
		err error
	)
	err = simSetDigitalPort(m.Handle(), DigitalPort(m.Port), m.Enabled, m.Logiclevel)
	// response := &SetDigitalPortRsp{Stat: Stat{err: err}}
	response := m.Rsp().(*genericps.SetDigitalAnalogTriggerOperandRsp)
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

func setOutputEdgeDetect(m *genericps.SetOutputEdgeDetectMsg) {
	err := simSetOutputEdgeDetect(m.Handle(), m.State)
	// response := &SetOutputEdgeDetectRsp{Stat: Stat{err: err}}
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
	err = simSetPulseWidthDigitalPortProperties(m.Handle(), dd)
	// response := &SetPulseWidthDigitalPortPropertiesRsp{Stat: Stat{err: err}}
	response := m.Rsp().(*genericps.SetPulseWidthDigitalPortPropertiesRsp)
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

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
	// response := &SetSigGenArbitraryRsp{Stat: Stat{err: err}}
	response := m.Rsp().(*genericps.SetSigGenArbitraryRsp)
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

func setSigGenPropertiesArbitrary(m *genericps.SetSigGenPropertiesArbitraryMsg) {
	err := simSetSigGenPropertiesArbitrary(m.Handle(), m.OffsetVoltage, m.StartDeltaPhase,
		m.StopDeltaPhase, m.DeltaPhaseIncrement, m.DwellCount, SweepTypeEnum(m.SweepType),
		ExtraOperations(m.Operation), IndexMode(m.IndexMode), m.Shots, m.Sweeps,
		SigGenTrigType(m.TriggerType),
		SigGenTrigSource(m.TriggerSource), m.ExtInThreshold)
	// response := &SetSigGenPropertiesArbitraryRsp{Stat: Stat{err: err}}
	response := m.Rsp().(*genericps.SetSigGenPropertiesArbitraryRsp)
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

func setSigGenPropertiesBuiltIn(m *genericps.SetSigGenPropertiesBuiltInMsg) {
	err := simSetSigGenPropertiesBuiltIn(m.Handle(), m.OffsetVoltage, m.StartFrequency,
		m.StopFrequency, m.Increment, m.DwellTime, SweepTypeEnum(m.SweepType),
		m.Shots, m.Sweeps,
		SigGenTrigType(m.TriggerType), SigGenTrigSource(m.TriggerSource), m.ExtInThreshold)
	// response := &SetSigGenPropertiesBuiltInRsp{Stat: Stat{err: err}}
	response := m.Rsp().(*genericps.SetSigGenPropertiesBuiltInRsp)
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

func sigGenArbitraryMinMaxValues(m *genericps.SigGenArbitraryMinMaxValuesMsg) {
	minArbitraryWaveformValue, maxArbitraryWaveformValue, minArbitraryWaveformSize,
		maxArbitraryWaveformSize, err := simSigGenArbitraryMinMaxValues(m.Handle())
	// response := &SigGenArbitraryMinMaxValuesRsp{Stat: Stat{err: err}, MinArbitraryWaveformValue: minArbitraryWaveformValue,
	// 	MaxArbitraryWaveformValue: maxArbitraryWaveformValue, MinArbitraryWaveformSize: minArbitraryWaveformSize,
	// 	MaxArbitraryWaveformSize: maxArbitraryWaveformSize}
	response := m.Rsp().(*genericps.SigGenArbitraryMinMaxValuesRsp)
	response.MinArbitraryWaveformValue = minArbitraryWaveformValue
	response.MaxArbitraryWaveformValue = maxArbitraryWaveformValue
	response.MinArbitraryWaveformSize = minArbitraryWaveformSize
	response.MaxArbitraryWaveformSize = maxArbitraryWaveformSize
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

func sigGenSoftwareControl(m *genericps.SigGenSoftwareControlMsg) {
	err := simSigGenSoftwareControl(m.Handle(), m.State)
	// response := &SigGenSoftwareControlRsp{Stat: Stat{err: err}}
	response := m.Rsp().(*genericps.SigGenSoftwareControlRsp)
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

func setSimGen(m *genericps.SetSimGenMsg) {
	s := SimDesc{handle: m.Handle()}
	err := s.SetSimGen(m.Channel, m.On, m.OffsetVoltage, m.PkToPK, genericps.WaveTypeEnum(m.WaveType),
		m.StartFrequency, m.StopFrequency, m.Increment, m.DwellTime, genericps.SweepTypeEnum(m.SweepType),
		genericps.ExtraOperations(m.Operation), m.Shots, m.Sweeps, genericps.SigGenTrigType(m.TriggerType),
		genericps.SigGenTrigSource(m.TriggerSource), m.ExtInThreshold, m.Phase)
	response := m.Rsp().(*genericps.SetSimGenRsp)
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

func setSimRlcFilter(m *genericps.SetSimRlcFilterMsg) {
	s := SimDesc{handle: m.Handle()}
	err := s.SetSimRlcFilter(m.Channel, m.GenSource, m.Enabled, m.FilterType, m.R, m.RUnit, m.L, m.LUnit, m.C, m.CUnit)
	response := m.Rsp().(*genericps.SetSimRlcFilterRsp)
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

func setSimDigitalFilter(m *genericps.SetSimDigitalFilterMsg) {
	s := SimDesc{handle: m.Handle()}
	err := s.SetSimDigitalFilter(m.Channel, m.LowpassEnabled, m.LowpassFc, m.HighpassEnabled, m.HighpassFc,
		m.BandpassEnabled, m.BandpassFc1, m.BandpassFc2, m.BandstopEnabled, m.BandstopFc1, m.BandstopFc2)
	response := m.Rsp().(*genericps.SetSimDigitalFilterRsp)
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

