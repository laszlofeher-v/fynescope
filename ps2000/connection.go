//go:build !demo && ps2000

package ps2000

import (
	"fmt"
	"fynescope/genericps"
	"log/slog"
	"sync"
	"time"
)

const (
	cmdSendTimeout         = 5000 * time.Millisecond
	responseReceiveTimeout = 5000 * time.Millisecond
)

var (
	buffersMutex sync.Mutex
	scopeBuffers = make(map[int16]map[genericps.ChannelId][]int16)
)

func Stop() (err error) {
	return
}

func getValues(m *genericps.GetValuesMsg) {
	buffersMutex.Lock()
	chMap := scopeBuffers[m.Handle()]
	buffersMutex.Unlock()
	
	bufA := chMap[genericps.ChA]
	bufB := chMap[genericps.ChB]
	bufC := chMap[genericps.ChC]
	bufD := chMap[genericps.ChD]

	numOfSamples, overflow, err := ps2000GetValues(m.Handle(),
		m.ReqNumOfSamples, bufA, bufB, bufC, bufD)
		
	response := m.Rsp().(*genericps.GetValuesRsp)
	response.SetStatus(err)
	response.NumOfSamples = numOfSamples
	response.Overflow = overflow
	m.RspCh() <- struct{}{}
}

func closeUnit(m *genericps.CloseUnitMsg) {
	slog.Info("Close unit")
	err := ps2000CloseUnit(m.Handle())
	response := m.Rsp().(*genericps.CloseUnitRsp)
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

func setChannel(m *genericps.SetChannelMsg) {
	err := ps2000SetChannel(m.Handle(), ChannelId(m.Channel),
		m.Enabled, Coupling(m.CouplingType),
		RangeEnum(m.VoltageRange), m.AnalogOffset)
	response := m.Rsp().(*genericps.SetChannelRsp)
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

func flashLed(m *genericps.FlashLedMsg) {
	err := ps2000FlashLed(m.Handle(), m.Start)
	response := m.Rsp().(*genericps.FlashLedRsp)
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

func getUnitInfo(m *genericps.GetUnitInfoMsg) {
	s, err := ps2000GetUnitInfo(m.Handle(), PicoInfo(m.Info))
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
	err := fmt.Errorf("Not Supported on ps2000")
	response := m.Rsp().(*genericps.GetValuesAsyncRsp)
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

func getValuesBulk(m *genericps.GetValuesBulkMsg) {
	err := fmt.Errorf("Not Supported on ps2000")
	response := m.Rsp().(*genericps.GetValuesBulkRsp)
	response.NumOfSamples = 0
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

func getValuesOverlapped(m *genericps.GetValuesOverlappedMsg) {
	err := fmt.Errorf("Not Supported on ps2000")
	response := m.Rsp().(*genericps.GetValuesOverlappedRsp)
	response.NumOfSamples = 0
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

func getValuesOverlappedBulk(m *genericps.GetValuesOverlappedBulkMsg) {
	err := fmt.Errorf("Not Supported on ps2000")
	response := m.Rsp().(*genericps.GetValuesOverlappedBulkRsp)
	response.NumOfSamples = 0
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

func getAnalogueOffsetMsg(m *genericps.GetAnalogueOffsetMsg) {
	err := fmt.Errorf("Not Supported on ps2000")
	response := m.Rsp().(*genericps.GetAnalogueOffsetRsp)
	response.MinimumVoltage = 0
	response.MaximumVoltage = 0
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

func getChannelInformation(m *genericps.GetChannelInformationMsg) {
	err := fmt.Errorf("Not Supported on ps2000")
	response := m.Rsp().(*genericps.GetChannelInformationRsp)
	response.Ranges = m.Ranges
	response.LengthOfRanges = 0
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

func getMaxDownSampleRatio(m *genericps.GetMaxDownSampleRatioMsg) {
	err := fmt.Errorf("Not Supported on ps2000")
	response := m.Rsp().(*genericps.GetMaxDownSampleRatioRsp)
	response.MaxDownSampleRatio = 1
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

func getMaxSegments(m *genericps.GetMaxSegmentsMsg) {
	err := fmt.Errorf("Not Supported on ps2000")
	response := m.Rsp().(*genericps.GetMaxSegmentsRsp)
	response.MaxSegments = 1
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

func getNumberOfCaptures(m *genericps.GetNumOfCapturesMsg) {
	err := fmt.Errorf("Not Supported on ps2000")
	response := m.Rsp().(*genericps.GetNumOfCapturesRsp)
	response.NCaptures = 1
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

func getNumberOfProcessedCaptures(m *genericps.GetNumOfProcessedCapturesMsg) {
	err := fmt.Errorf("Not Supported on ps2000")
	response := m.Rsp().(*genericps.GetNumOfProcessedCapturesRsp)
	response.NCaptures = 1
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

func getStreamingLatestValues(m *genericps.GetStreamingLatestValuesMsg) {
	err := ps2000GetStreamingLatestValues(m.Handle(), genericps.StreamingReady(m.LpStreamingReadyGoPar), m.Param)
	response := m.Rsp().(*genericps.GetStreamingLatestValuesRsp)
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

func getTimebase(m *genericps.GetTimebaseMsg) {
	timeIntervalNanoseconds, maxSamples, err := ps2000GetTimebase(m.Handle(), m.TimeBase, m.NumOfSamples, m.OverSample, m.SegmentIndex)
	response := m.Rsp().(*genericps.GetTimebaseRsp)
	response.TimeIntervalNanoseconds = timeIntervalNanoseconds
	response.MaxSamples = maxSamples
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

func getTimebase2(m *genericps.GetTimebase2Msg) {
	err := fmt.Errorf("Not Supported on ps2000")
	response := m.Rsp().(*genericps.GetTimebase2Rsp)
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

func maximumValue(m *genericps.MaximumValueMsg) {
	value, err := ps2000MaximumValue(m.Handle())
	response := m.Rsp().(*genericps.MaximumValueRsp)
	response.Value = value
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

func minimumValue(m *genericps.MinimumValueMsg) {
	value, err := ps2000MinimumValue(m.Handle())
	response := m.Rsp().(*genericps.MinimumValueResp)
	response.Value = value
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

func setSimpleTrigger(m *genericps.SetSimpleTriggerMsg) {
	err := ps2000SetSimpleTrigger(m.Handle(), m.Enable, ChannelId(m.Source),
		m.Threshold, ThresholdDirection(m.Direction), m.Delay, m.AutoTriggerMs)
	response := m.Rsp().(*genericps.SetSimpleTriggerRsp)
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

func setDataBuffer(m *genericps.SetDataBufferMsg) {
	buffersMutex.Lock()
	if scopeBuffers[m.Handle()] == nil {
		scopeBuffers[m.Handle()] = make(map[genericps.ChannelId][]int16)
	}
	scopeBuffers[m.Handle()][m.Ch] = m.BufferIn
	buffersMutex.Unlock()
	
	response := m.Rsp().(*genericps.SetDataBufferRsp)
	response.SetStatus(nil)
	m.RspCh() <- struct{}{}
}

func setDataBuffers(m *genericps.SetDataBuffersMsg) {
	err := fmt.Errorf("Not Supported on ps2000")
	response := m.Rsp().(*genericps.SetDataBuffersRsp)
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

func setUnscaledDataBuffers(m *genericps.SetUnscaledDataBuffersMsg) {
	err := fmt.Errorf("Not Supported on ps2000")
	response := m.Rsp().(*genericps.SetUnscaledataBuffersRsp)
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

func setEtsTimeBuffer(m *genericps.SetEtsTimeBufferMsg) {
	err := fmt.Errorf("Not Supported on ps2000")
	response := m.Rsp().(*genericps.SetEtsTimeBufferRsp)
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

func setEtsTimeBuffers(m *genericps.SetEtsTimeBuffersMsg) {
	err := fmt.Errorf("Not Supported on ps2000")
	response := m.Rsp().(*genericps.SetEtsTimeBufferRsp)
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}
func setEts(m *genericps.SetEtsMsg) {
	err := fmt.Errorf("Not Supported on ps2000")
	response := m.Rsp().(*genericps.SetEtsRsp)
	response.SampleTimePicoseconds = 0
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

func runStreaming(m *genericps.RunStreamingMsg) {
	sampleInterval, err := ps2000RunStreaming(m.Handle(), m.ReqSampleInterval, TimeUnits(m.SampleIntervalTimeUnits), m.MaxPreTriggerSamples,
		m.MaxPostTriggerSamples, m.AutoStop, m.DownSampleRatio, RatioMode(m.DownSampleRatioMode), m.OverviewBufferSize)
	response := m.Rsp().(*genericps.RunStreamingRsp)
	response.SampleInterval = sampleInterval
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

func runBlock(m *genericps.RunBlockMsg) {
	timeIndisposedMs, err := ps2000RunBlock(m.Handle(), m.NumOfPreTriggerSamples, m.NumOfPostTriggerSamples,
		m.TimeBase, m.OverSample, m.SegmentIndex, genericps.BlockReady(m.LpBlockReadyGoPar), m.Param)
	response := m.Rsp().(*genericps.RunBlockRsp)
	response.TimeIndisposedMs = timeIndisposedMs
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

func setTriggerChannelProperties(m *genericps.SetTriggerChannelPropertiesMsg) {
	err := fmt.Errorf("Not Supported on ps2000")
	response := m.Rsp().(*genericps.SetTriggerChannelPropertiesRsp)
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

func setTriggerChannelConditions(m *genericps.SetTriggerChannelConditionsMsg) {
	err := fmt.Errorf("Not Supported on ps2000")
	response := m.Rsp().(*genericps.SetTriggerChannelConditionsRsp)
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

func setTriggerChannelDirections(m *genericps.SetTriggerChannelDirectionsMsg) {
	err := fmt.Errorf("Not Supported on ps2000")
	response := m.Rsp().(*genericps.SetTriggerChannelDirectionsRsp)
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

func setTriggerDelay(m *genericps.SetTriggerDelayMsg) {
	err := fmt.Errorf("Not Supported on ps2000")
	response := m.Rsp().(*genericps.SetTriggerDelayRsp)
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

func setPulseWidthQualifier(m *genericps.SetPulseWidthQualifierMsg) {
	err := fmt.Errorf("Not Supported on ps2000")
	response := m.Rsp().(*genericps.SetPulseWidthQualifierRsp)
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

func setTriggerDigitalPortProperties(m *genericps.SetTriggerDigitalPortPropertiesMsg) {
	err := fmt.Errorf("Not Supported on ps2000")
	response := m.Rsp().(*genericps.SetTriggerDigitalPortPropertiesRsp)
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

func stopReq(m *genericps.StopMsg) {
	err := ps2000Stop(m.Handle())
	response := m.Rsp().(*genericps.StopRsp)
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

func setSigGenBuiltIn(m *genericps.SetSigGenBuiltInMsg) {
	err := ps2000SetSigGenBuiltIn(m.Handle(), m.OffsetVoltage, m.PkToPK, WaveTypeEnum(m.WaveType), m.StartFrequency,
		m.StopFrequency, m.Increment, m.DwellTime, SweepTypeEnum(m.SweepType),
		ExtraOperations(m.Operation), m.Shots, m.Sweeps, SigGenTrigType(m.TriggerType),
		SigGenTrigSource(m.TriggerSource), m.ExtInThreshold)
	response := m.Rsp().(*genericps.SetSigGenBuiltInRsp)
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

func setSigGenBuiltInV2(m *genericps.SetSigGenBuiltInV2Msg) {
	err := fmt.Errorf("Not Supported on ps2000")
	response := m.Rsp().(*genericps.SetSigGenBuiltInV2Rsp)
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

func sigGenFrequencyToPhase(m *genericps.SigGenFrequencyToPhasenMsg) {
	err := fmt.Errorf("Not Supported on ps2000")
	response := m.Rsp().(*genericps.SigGenFrequencyToPhaseRsp)
	response.SetStatus(err)
	response.Phase = 0
	m.RspCh() <- struct{}{}
}

func setNumOfCaptures(m *genericps.SetNumOfCapturesMsg) {
	err := fmt.Errorf("Not Supported on ps2000")
	response := m.Rsp().(*genericps.SetNumOfCapturesRsp)
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

func getTriggerTimeOffset(m *genericps.GetTriggerTimeOffsetMsg) {
	err := fmt.Errorf("Not Supported on ps2000")
	response := m.Rsp().(*genericps.GetTriggerTimeOffsetRsp)
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

func getTriggerTimeOffset64(m *genericps.GetTriggerTimeOffset64Msg) {
	err := fmt.Errorf("Not Supported on ps2000")
	response := m.Rsp().(*genericps.GetTriggerTimeOffset64Rsp)
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

func getValuesTriggerTimeOffsetBulk(m *genericps.GetValuesTriggerTimeOffsetBulkMsg) {
	err := fmt.Errorf("Not Supported on ps2000")
	response := m.Rsp().(*genericps.GetValuesTriggerTimeOffsetBulkRsp)
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

func getValuesTriggerTimeOffsetBulk64(m *genericps.GetValuesTriggerTimeOffsetBulk64Msg) {
	err := fmt.Errorf("Not Supported on ps2000")
	response := m.Rsp().(*genericps.GetValuesTriggerTimeOffsetBulk64Rsp)
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

func holdOff(m *genericps.HoldOffMsg) {
	err := fmt.Errorf("Not Supported on ps2000")
	response := m.Rsp().(*genericps.HoldOffRsp)
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

func lsReady(m *genericps.LsReadyMsg) {
	ready, err := ps2000IsReady(m.Handle())
	response := m.Rsp().(*genericps.LsReadyRsp)
	response.Ready = ready
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

func triggerOrPulseWidthQualifierEnabled(m *genericps.TriggerOrPulseWidthQualifierEnabledMsg) {
	err := fmt.Errorf("Not Supported on ps2000")
	response := m.Rsp().(*genericps.TriggerOrPulseWidthQualifierEnabledRsp)
	response.TriggerEnabled = 0
	response.PulseWidthQualifierEnabledint16 = 0
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

func memorySegments(m *genericps.MemorySegmentsMsg) {
	err := fmt.Errorf("Not Supported on ps2000")
	response := m.Rsp().(*genericps.MemorySegmentsRsp)
	response.SetStatus(err)
	response.NMaxSamples = 0
	m.RspCh() <- struct{}{}
}

func numOfStreamingValues(m *genericps.NumOfStreamingValuesMsg) {
	err := fmt.Errorf("Not Supported on ps2000")
	response := m.Rsp().(*genericps.NumOfStreamingValuesRsp)
	response.NumOfValues = 0
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

func pingUnit(m *genericps.PingUnitMsg) {
	err := fmt.Errorf("Not Supported on ps2000")
	response := m.Rsp().(*genericps.PingUnitRsp)
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

func queryOutputEdgeDetect(m *genericps.QueryOutputEdgeDetectMsg) {
	err := fmt.Errorf("Not Supported on ps2000")
	response := m.Rsp().(*genericps.QueryOutputEdgeDetectRsp)
	response.State = 0
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

func setDigitalAnalogTriggerOperand(m *genericps.SetDigitalAnalogTriggerOperandMsg) {
	err := fmt.Errorf("Not Supported on ps2000")
	response := m.Rsp().(*genericps.SetDigitalAnalogTriggerOperandRsp)
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

func setDigitalPort(m *genericps.SetDigitalPortMsg) {
	err := fmt.Errorf("Not Supported on ps2000")
	response := m.Rsp().(*genericps.SetDigitalAnalogTriggerOperandRsp)
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

func setOutputEdgeDetect(m *genericps.SetOutputEdgeDetectMsg) {
	err := fmt.Errorf("Not Supported on ps2000")
	response := m.Rsp().(*genericps.SetOutputEdgeDetectRsp)
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

func setPulseWidthDigitalPortProperties(m *genericps.SetPulseWidthDigitalPortPropertiesMsg) {
	err := fmt.Errorf("Not Supported on ps2000")
	response := m.Rsp().(*genericps.SetPulseWidthDigitalPortPropertiesRsp)
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

func setSigGenArbitrary(m *genericps.SetSigGenArbitraryMsg) {
	err := fmt.Errorf("Not Supported on ps2000")
	response := m.Rsp().(*genericps.SetSigGenArbitraryRsp)
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

func setSigGenPropertiesArbitrary(m *genericps.SetSigGenPropertiesArbitraryMsg) {
	err := fmt.Errorf("Not Supported on ps2000")
	response := m.Rsp().(*genericps.SetSigGenPropertiesArbitraryRsp)
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

func setSigGenPropertiesBuiltIn(m *genericps.SetSigGenPropertiesBuiltInMsg) {
	err := fmt.Errorf("Not Supported on ps2000")
	response := m.Rsp().(*genericps.SetSigGenPropertiesBuiltInRsp)
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

func sigGenArbitraryMinMaxValues(m *genericps.SigGenArbitraryMinMaxValuesMsg) {
	err := fmt.Errorf("Not Supported on ps2000")
	response := m.Rsp().(*genericps.SigGenArbitraryMinMaxValuesRsp)
	response.MinArbitraryWaveformValue = 0
	response.MaxArbitraryWaveformValue = 0
	response.MinArbitraryWaveformSize = 0
	response.MaxArbitraryWaveformSize = 0
	response.SetStatus(err)
	m.RspCh() <- struct{}{}
}

func sigGenSoftwareControl(m *genericps.SigGenSoftwareControlMsg) {
	err := fmt.Errorf("Not Supported on ps2000")
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
		stopReq(m)
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
	}
}
