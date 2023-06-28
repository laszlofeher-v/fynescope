package psi

// Receives messages from the GUI
// Calls go ps2000a API wrapper

import (
	"fmt"
	"psgo/ps2000a"
	// "psgo/ps2000asim"
	// "time"
)

type (
	Stat struct {
		err error
	}
)

type (
	GetValuesMsg struct {
		Unit
		StartIndex          uint32
		ReqNumOfSamples     uint32
		DownSampleRatio     uint32
		DownSampleRatioMode ps2000a.RatioMode
		SegmentIndex        uint32
	}
	GetValuesRsp struct {
		Stat
		NumOfSamples uint32
		Overflow     int16
	}
)

func getValues(m *GetValuesMsg) {
	numOfSamples, overflow, err := m.ps.GetValues(m.StartIndex, m.ReqNumOfSamples, m.DownSampleRatio, m.DownSampleRatioMode,
		m.SegmentIndex)
	response := &GetValuesRsp{Stat: Stat{err: err}, NumOfSamples: numOfSamples, Overflow: overflow}
	m.respCh <- response
}

type (
	CloseUnitMsg struct {
		Unit
	}
	CloseUnitRsp struct {
		Stat
	}
)

func closeUnit(m *CloseUnitMsg) {
	fmt.Println("Close unit")
	err := m.ps.CloseUnit()
	response := &CloseUnitRsp{Stat: Stat{err: err}}
	m.respCh <- response
	close(m.cmdCh)
}

type (
	SetChannelMsg struct {
		Unit
		Channel      ps2000a.ChannelId
		Enabled      bool
		CouplingType ps2000a.Coupling
		VoltageRange ps2000a.RangeEnum
		AnalogOffset float32
	}
	SetChannelRsp struct {
		Stat
	}
)

func setChannel(m *SetChannelMsg) {
	err := m.ps.SetChannel(m.Channel, m.Enabled, m.CouplingType, m.VoltageRange, m.AnalogOffset)
	response := &SetChannelRsp{Stat: Stat{err: err}}
	m.respCh <- response
}

type (
	FlashLedMsg struct {
		Unit
		Start int16
	}
	FlashLedRsp struct {
		Stat
	}
)

func flashLed(m *FlashLedMsg) {
	err := m.ps.FlashLed(m.Start)
	response := &FlashLedRsp{Stat: Stat{err: err}}
	m.respCh <- response
}

type (
	GetUnitInfoMsg struct {
		Unit
		Info ps2000a.PicoInfo
	}
	GetUnitInfoRsp struct {
		Stat
		InfoString string
	}
)

func getUnitInfo(m *GetUnitInfoMsg) {
	s, err := m.ps.GetUnitInfo(m.Info)
	response := &GetUnitInfoRsp{Stat: Stat{err: err}, InfoString: s}
	m.respCh <- response
}

type (
	GetValuesAsyncMsg struct {
		Unit
		StartIndex          uint32
		NumOfSamples        uint32
		DownSampleRatio     uint32
		DownSampleRatioMode ps2000a.RatioMode
		LpDataReady         ps2000a.DataReady
		SegmentIndex        uint32
		Param               interface{}
	}
	GetValuesAsyncRsp struct {
		Stat
	}
)

func getValuesAsync(m *GetValuesAsyncMsg) {
	err := m.ps.GetValuesAsync(m.StartIndex, m.NumOfSamples, m.DownSampleRatio, m.DownSampleRatioMode,
		m.LpDataReady, m.SegmentIndex, m.Param)
	response := &GetValuesAsyncRsp{Stat: Stat{err: err}}
	m.respCh <- response
}

type (
	GetValuesBulkMsg struct {
		Unit
		ReqNumOfSamples     uint32
		FromSegmentIndex    uint32
		ToSegmentIndex      uint32
		DownSampleRatio     uint32
		DownSampleRatioMode ps2000a.RatioMode
		Overflow            []int16
	}
	GetValuesBulkRsp struct {
		Stat
		NumOfSamples uint32
	}
)

func getValuesBulk(m *GetValuesBulkMsg) {
	numOfSamples, err := m.ps.GetValuesBulk(m.ReqNumOfSamples, m.FromSegmentIndex, m.ToSegmentIndex,
		m.DownSampleRatio, m.DownSampleRatioMode, m.Overflow)
	response := &GetValuesBulkRsp{Stat: Stat{err: err}, NumOfSamples: numOfSamples}
	m.respCh <- response
}

type (
	GetValuesOverlappedMsg struct {
		Unit
		StartIndex          uint32
		ReqNumOfSamples     uint32
		DownSampleRatio     uint32
		DownSampleRatioMode ps2000a.RatioMode
		SegmentIndex        uint32
		Overflow            []int16
	}
	GetValuesOverlappedRsp struct {
		Stat
		NumOfSamples uint32
	}
)

func getValuesOverlapped(m *GetValuesOverlappedMsg) {
	numOfSamples, err := m.ps.GetValuesOverlapped(m.StartIndex, m.ReqNumOfSamples,
		m.DownSampleRatio, m.DownSampleRatioMode, m.SegmentIndex, m.Overflow)
	response := &GetValuesOverlappedRsp{Stat: Stat{err: err}, NumOfSamples: numOfSamples}
	m.respCh <- response
}

type (
	GetValuesOverlappedBulkMsg struct {
		Unit
		StartIndex          uint32
		ReqNumOfSamples     uint32
		DownSampleRatio     uint32
		DownSampleRatioMode ps2000a.RatioMode
		FromSegment         uint32
		ToSegment           uint32
		Overflow            []int16
	}
	GetValuesOverlappedBulkRsp struct {
		Stat
		NumOfSamples uint32
	}
)

func getValuesOverlappedBulk(m *GetValuesOverlappedBulkMsg) {
	numOfSamples, err := m.ps.GetValuesOverlappedBulk(m.StartIndex, m.ReqNumOfSamples,
		m.DownSampleRatio, m.DownSampleRatioMode, m.FromSegment, m.ToSegment, m.Overflow)
	response := &GetValuesOverlappedBulkRsp{Stat: Stat{err: err}, NumOfSamples: numOfSamples}
	m.respCh <- response
}

type (
	GetAnalogueOffsetMsg struct {
		Unit
		VoltageRange int
		Coupling     ps2000a.Coupling
	}
	GetAnalogueOffsetRsp struct {
		Stat
		MaximumVoltage float32
		MinimumVoltage float32
	}
)

func getAnalogueOffsetMsg(m *GetAnalogueOffsetMsg) {
	maximumVoltage, minimumVoltage, err := m.ps.GetAnalogueOffset(m.VoltageRange, m.Coupling)
	response := &GetAnalogueOffsetRsp{Stat: Stat{err: err}, MinimumVoltage: minimumVoltage, MaximumVoltage: maximumVoltage}
	m.respCh <- response
}

type (
	GetChannelInformationMsg struct {
		Unit
		Info    int16
		Probe   int32
		Ranges  []int32
		Channel ps2000a.ChannelId
	}
	GetChannelInformationRsp struct {
		Stat
		LengthOfRanges int32
	}
)

func getChannelInformation(m *GetChannelInformationMsg) {
	lengthOfRanges, err := m.ps.GetChannelInformation(m.Info, m.Probe, m.Ranges, m.Channel)
	response := &GetChannelInformationRsp{Stat: Stat{err: err}, LengthOfRanges: lengthOfRanges}
	m.respCh <- response
}

type (
	GetMaxDownSampleRatioMsg struct {
		Unit
		NumOfUnaggregatedSamples uint32
		DownSampleRatioMode      ps2000a.RatioMode
		SegmentIndex             int32
	}
	GetMaxDownSampleRatioRsp struct {
		Stat
		MaxDownSampleRatio uint32
	}
)

func getMaxDownSampleRatio(m *GetMaxDownSampleRatioMsg) {
	maxDownSampleRatio, err := m.ps.GetMaxDownSampleRatio(m.NumOfUnaggregatedSamples, m.DownSampleRatioMode, m.SegmentIndex)
	response := &GetMaxDownSampleRatioRsp{Stat: Stat{err: err}, MaxDownSampleRatio: maxDownSampleRatio}
	m.respCh <- response
}

type (
	GetMaxSegmentsMsg struct {
		Unit
	}
	GetMaxSegmentsRsp struct {
		Stat
		MaxSegments uint32
	}
)

func getMaxSegments(m *GetMaxSegmentsMsg) {
	maxSegments, err := m.ps.GetMaxSegments()
	response := &GetMaxSegmentsRsp{Stat: Stat{err: err}, MaxSegments: maxSegments}
	m.respCh <- response
}

type (
	GetNumOfCapturesMsg struct {
		Unit
	}
	GetNumOfCapturesRsp struct {
		Stat
		NCaptures uint32
	}
)

func getNumberOfCaptures(m *GetNumOfCapturesMsg) {
	numOfCaptures, err := m.ps.GetNoOfCaptures()
	response := &GetNumOfCapturesRsp{Stat: Stat{err: err}, NCaptures: numOfCaptures}
	m.respCh <- response
}

type (
	GetNumOfProcessedCapturesMsg struct {
		Unit
	}
	GetNumOfProcessedCapturesRsp struct {
		Stat
		NCaptures uint32
	}
)

func getNumberOfProcessedCaptures(m *GetNumOfProcessedCapturesMsg) {
	numOfCaptures, err := m.ps.GetNoOfProcessedCaptures()
	response := &GetNumOfProcessedCapturesRsp{Stat: Stat{err: err}, NCaptures: numOfCaptures}
	m.respCh <- response
}

type (
	GetStreamingLatestValuesMsg struct {
		Unit
		LpStreamingReadyGoPar ps2000a.StreamingReady
		Param                 interface{}
	}
	GetStreamingLatestValuesRsp struct {
		Stat
	}
)

func getStreamingLatestValues(m *GetStreamingLatestValuesMsg) {
	err := m.ps.GetStreamingLatestValues(m.LpStreamingReadyGoPar, m.Param)
	response := &GetStreamingLatestValuesRsp{Stat: Stat{err: err}}
	m.respCh <- response
}

type (
	GetTimebaseMsg struct {
		Unit
		TimeBase     uint32
		NumOfSamples int32
		OverSample   int16
		SegmentIndex uint32
	}
	GetTimebaseRsp struct {
		Stat
		TimeIntervalNanoseconds int32
		MaxSamples              int32
	}
)

func getTimebase(m *GetTimebaseMsg) {
	timeIntervalNanoseconds, maxSamples, err := m.ps.GetTimebase(m.TimeBase, m.NumOfSamples, m.OverSample, m.SegmentIndex)
	response := &GetTimebaseRsp{Stat: Stat{err: err}, TimeIntervalNanoseconds: timeIntervalNanoseconds, MaxSamples: maxSamples}
	m.respCh <- response
}

type (
	GetTimebase2Msg struct {
		Unit
		TimeBase     uint32
		NumOfSamples int32
		OverSample   int16
		SegmentIndex uint32
	}
	GetTimebase2Rsp struct {
		Stat
		TimeIntervalNanoseconds float32
		MaxSamples              int32
	}
)

func getTimebase2(m *GetTimebase2Msg) {
	timeIntervalNanoseconds, maxSamples, err := m.ps.GetTimebase2(m.TimeBase, m.NumOfSamples, m.OverSample, m.SegmentIndex)
	response := &GetTimebase2Rsp{Stat: Stat{err: err}, TimeIntervalNanoseconds: timeIntervalNanoseconds, MaxSamples: maxSamples}
	m.respCh <- response
}

type (
	MaximumValueMsg struct {
		Unit
	}
	MaximumValueRsp struct {
		Stat
		Value int16
	}
)

func maximumValue(m *MaximumValueMsg) {
	value, err := m.ps.MaximumValue()
	response := &MaximumValueRsp{Stat: Stat{err: err}, Value: value}
	m.respCh <- response
}

type (
	MinimumValueMsg struct {
		Unit
	}
	MinimumValueResp struct {
		Stat
		Value int16
	}
)

func minimumValue(m *MinimumValueMsg) {
	value, err := m.ps.MinimumValue()
	response := &MinimumValueResp{Stat: Stat{err: err}, Value: value}
	m.respCh <- response
}

type (
	SetSimpleTriggerMsg struct {
		Unit
		Eable         bool
		Source        ps2000a.ChannelId
		Threshold     int16
		Direction     ps2000a.ThresholdDirection
		Delay         uint32
		AutoTriggerMs int16
	}
	SetSimpleTriggerRsp struct {
		Stat
	}
)

func setSimpleTrigger(m *SetSimpleTriggerMsg) {
	err := m.ps.SetSimpleTrigger(m.Eable, m.Source, m.Threshold, m.Direction, m.Delay, m.AutoTriggerMs)
	response := &SetSimpleTriggerRsp{Stat: Stat{err: err}}
	m.respCh <- response
}

type (
	SetDataBufferMsg struct {
		Unit
		Ch           ps2000a.ChannelId
		BufferIn     []int16
		SegmentIndex uint32
		Mode         ps2000a.RatioMode
	}
	SetDataBufferRsp struct {
		Stat
	}
)

func setDataBuffer(m *SetDataBufferMsg) {
	err := m.ps.SetDataBuffer(m.Ch, m.BufferIn, m.SegmentIndex, m.Mode)
	response := &SetDataBufferRsp{Stat: Stat{err: err}}
	m.respCh <- response
}

type (
	SetDataBuffersMsg struct {
		Unit
		Ch           ps2000a.ChannelId
		BufferMax    []int16
		BufferMin    []int16
		SegmentIndex uint32
		Mode         ps2000a.RatioMode
	}
	SetDataBuffersRsp struct {
		Stat
	}
)

func setDataBuffers(m *SetDataBuffersMsg) {
	err := m.ps.SetDataBuffers(m.Ch, m.BufferMax, m.BufferMin, m.SegmentIndex, m.Mode)
	response := &SetDataBuffersRsp{Stat: Stat{err: err}}
	m.respCh <- response
}

type (
	SetUnscaledDataBuffersMsg struct {
		Unit
		Ch           ps2000a.ChannelId
		BufferMax    []int16
		BufferMin    []int16
		SegmentIndex uint32
		Mode         ps2000a.RatioMode
	}
	SetUnscaledataBuffersRsp struct {
		Stat
	}
)

func setUnscaledDataBuffers(m *SetUnscaledDataBuffersMsg) {
	err := m.ps.SetUnscaledDataBuffers(m.Ch, m.BufferMax, m.BufferMin, m.SegmentIndex, m.Mode)
	response := &SetUnscaledataBuffersRsp{Stat: Stat{err: err}}
	m.respCh <- response
}

type (
	SetEtsTimeBufferMsg struct {
		Unit
		Buffer []int64
	}
	SetEtsTimeBufferRsp struct {
		Stat
	}
)

func setEtsTimeBuffer(m *SetEtsTimeBufferMsg) {
	err := m.ps.SetEtsTimeBuffer(m.Buffer)
	response := &SetEtsTimeBufferRsp{Stat: Stat{err: err}}
	m.respCh <- response
}

type (
	SetEtsTimeBuffersMsg struct {
		Unit
		TimeUpper, TimeLower []uint32
	}
	SetEtsTimeBuffersRsp struct {
		Stat
	}
)

func setEtsTimeBuffers(m *SetEtsTimeBuffersMsg) {
	err := m.ps.SetEtsTimeBuffers(m.TimeUpper, m.TimeLower)
	response := &SetEtsTimeBufferRsp{Stat: Stat{err: err}}
	m.respCh <- response
}

type (
	SetEtsMsg struct {
		Unit
		Mode          ps2000a.EtsMode
		EtsCycles     int16
		EtsInterleave int16
	}
	SetEtsRsp struct {
		Stat
		SampleTimePicoseconds int32
	}
)

func setEts(m *SetEtsMsg) {
	sampleTimePicoseconds, err := m.ps.SetEts(m.Mode, m.EtsCycles, m.EtsInterleave)
	response := &SetEtsRsp{Stat: Stat{err: err}, SampleTimePicoseconds: sampleTimePicoseconds}
	m.respCh <- response
}

type (
	RunStreamingMsg struct {
		Unit
		ReqSampleInterval                           uint32
		SampleIntervalTimeUnits                     ps2000a.TimeUnits
		MaxPreTriggerSamples, MaxPostTriggerSamples uint32
		AutoStop                                    bool
		DownSampleRatio                             uint32
		DownSampleRatioMode                         ps2000a.RatioMode
		OverviewBufferSize                          uint32
	}
	RunStreamingRsp struct {
		Stat
		SampleInterval uint32
	}
)

func runStreaming(m *RunStreamingMsg) {
	sampleInterval, err := m.ps.RunStreaming(m.ReqSampleInterval, m.SampleIntervalTimeUnits, m.MaxPostTriggerSamples,
		m.MaxPostTriggerSamples, m.AutoStop, m.DownSampleRatio, m.DownSampleRatioMode, m.OverviewBufferSize)
	response := &RunStreamingRsp{Stat: Stat{err: err}, SampleInterval: sampleInterval}
	m.respCh <- response
}

type (
	RunBlockMsg struct {
		Unit
		NumOfPreTriggerSamples  int32
		NumOfPostTriggerSamples int32
		TimeBase                uint32
		OverSample              int16
		SegmentIndex            uint32
		LpBlockReadyGoPar       ps2000a.BlockReady
		Param                   interface{}
	}
	RunBlockRsp struct {
		Stat
		TimeIndisposedMs int32
	}
)

func runBlock(m *RunBlockMsg) {
	timeIndisposeMs, err := m.ps.RunBlock(m.NumOfPreTriggerSamples, m.NumOfPostTriggerSamples,
		m.TimeBase, m.OverSample, m.SegmentIndex, m.LpBlockReadyGoPar, m.Param)
	response := &RunBlockRsp{Stat: Stat{err: err}, TimeIndisposedMs: timeIndisposeMs}
	m.respCh <- response
}

type (
	SetTriggerChannelPropertiesMsg struct {
		Unit
		ChannelProperties []ps2000a.TriggerChannelProperties
		AuxOutputEnable   bool
		AutoTriggerMs     int32
	}
	SetTriggerChannelPropertiesRsp struct {
		Stat
	}
)

func setTriggerChannelProperties(m *SetTriggerChannelPropertiesMsg) {
	err := m.ps.SetTriggerChannelProperties(m.ChannelProperties, m.AuxOutputEnable, m.AutoTriggerMs)
	response := &SetTriggerChannelPropertiesRsp{Stat: Stat{err: err}}
	m.respCh <- response
}

type (
	SetTriggerChannelConditionsMsg struct {
		Unit
		TriggerConditions []ps2000a.TriggerConditions
	}
	SetTriggerChannelConditionsRsp struct {
		Stat
	}
)

func setTriggerChannelConditions(m *SetTriggerChannelConditionsMsg) {
	err := m.ps.SetTriggerChannelConditions(m.TriggerConditions)
	response := &SetTriggerChannelConditionsRsp{Stat: Stat{err: err}}
	m.respCh <- response
}

type (
	SetTriggerChannelDirectionsMsg struct {
		Unit
		ChannelA, ChannelB, ChannelC, ChannelD, Ext, Aux ps2000a.ThresholdDirection
	}
	SetTriggerChannelDirectionsRsp struct {
		Stat
	}
)

func setTriggerChannelDirections(m *SetTriggerChannelDirectionsMsg) {
	err := m.ps.SetTriggerChannelDirections(m.ChannelA, m.ChannelB, m.ChannelC, m.ChannelD,
		m.Ext, m.Aux)
	response := &SetTriggerChannelDirectionsRsp{Stat: Stat{err: err}}
	m.respCh <- response
}

type (
	SetTriggerDelayMsg struct {
		Unit
		Delay uint32
	}
	SetTriggerDelayRsp struct {
		Stat
	}
)

func setTriggerDelay(m *SetTriggerDelayMsg) {
	err := m.ps.SetTriggerDelay(m.Delay)
	response := &SetTriggerDelayRsp{Stat: Stat{err: err}}
	m.respCh <- response
}

type (
	SetPulseWidthQualifierMsg struct {
		Unit
		Conditions   []ps2000a.PwqConditions
		Direction    ps2000a.ThresholdDirection
		Lower, upper uint32
		PwType       ps2000a.PulseWidthType
	}
	SetPulseWidthQualifierRsp struct {
		Stat
	}
)

func setPulseWidthQualifier(m *SetPulseWidthQualifierMsg) {
	err := m.ps.SetPulseWidthQualifier(m.Conditions, m.Direction, m.Lower, m.upper, m.PwType)
	response := &SetPulseWidthQualifierRsp{Stat: Stat{err: err}}
	m.respCh <- response
}

type (
	SetTriggerDigitalPortPropertiesMsg struct {
		Unit
		DigitalDirections []ps2000a.DigitalChannelDirections
	}
	SetTriggerDigitalPortPropertiesRsp struct {
		Stat
	}
)

func setTriggerDigitalPortProperties(m *SetTriggerDigitalPortPropertiesMsg) {
	err := m.ps.SetTriggerDigitalPortProperties(m.DigitalDirections)
	response := &SetTriggerDigitalPortPropertiesRsp{Stat: Stat{err: err}}
	m.respCh <- response
}

type (
	StopMsg struct {
		Unit
	}
	StopRsp struct {
		Stat
	}
)

func stop(m *StopMsg) {
	err := m.ps.Stop()
	response := &StopRsp{Stat: Stat{err: err}}
	m.respCh <- response
}

type (
	SetSigGenBuiltInMsg struct {
		Unit
		OffsetVoltage                                       int32
		PkToPK                                              uint32
		WaveType                                            ps2000a.WaveTypeEnum
		StartFrequency, StopFrequency, Increment, DwellTime float32
		SweepType                                           ps2000a.SweepTypeEnum
		Operation                                           ps2000a.ExtraOperations
		Shots, Sweeps                                       uint32
		TriggerType                                         ps2000a.SigGenTrigType
		TriggerSource                                       ps2000a.SigGenTrigSource
		ExtInThreshold                                      int16
	}
	SetSigGenBuiltInRsp struct {
		Stat
	}
)

func setSigGenBuiltIn(m *SetSigGenBuiltInMsg) {
	err := m.ps.SetSigGenBuiltIn(m.OffsetVoltage, m.PkToPK, m.WaveType, m.StartFrequency,
		m.StopFrequency, m.Increment, m.DwellTime, m.SweepType, m.Operation, m.Shots, m.Sweeps, m.TriggerType,
		m.TriggerSource, m.ExtInThreshold)
	response := &SetSigGenBuiltInRsp{Stat: Stat{err: err}}
	m.respCh <- response
}

type (
	SetSigGenBuiltInV2Msg struct {
		Unit
		OffsetVoltage                                       int32
		PkToPK                                              uint32
		WaveType                                            ps2000a.WaveTypeEnum
		StartFrequency, StopFrequency, Increment, DwellTime float64
		SweepType                                           ps2000a.SweepTypeEnum
		Operation                                           ps2000a.ExtraOperations
		Shots, Sweeps                                       uint32
		TriggerType                                         ps2000a.SigGenTrigType
		TriggerSource                                       ps2000a.SigGenTrigSource
		ExtInThreshold                                      int16
	}
	SetSigGenBuiltInV2Rsp struct {
		Stat
	}
)

func setSigGenBuiltInV2(m *SetSigGenBuiltInV2Msg) {
	err := m.ps.SetSigGenBuiltInV2(m.OffsetVoltage, m.PkToPK, m.WaveType, m.StartFrequency,
		m.StopFrequency, m.Increment, m.DwellTime, m.SweepType, m.Operation, m.Shots, m.Sweeps, m.TriggerType,
		m.TriggerSource, m.ExtInThreshold)
	response := &SetSigGenBuiltInV2Rsp{Stat: Stat{err: err}}
	m.respCh <- response
}

type (
	SigGenFrequencyToPhasenMsg struct {
		Unit
		Frequency    float64
		IndexMode    ps2000a.IndexMode
		BufferLength uint32
	}
	SigGenFrequencyToPhaseRsp struct {
		Stat
		Phase uint32
	}
)

func sigGenFrequencyToPhase(m *SigGenFrequencyToPhasenMsg) {
	phase, err := m.ps.SigGenFrequencyToPhase(m.Frequency, m.IndexMode, m.BufferLength)
	response := &SigGenFrequencyToPhaseRsp{Stat: Stat{err: err}, Phase: phase}
	m.respCh <- response
}

type (
	SetNumOfCapturesMsg struct {
		Unit
		NCaptures uint32
	}
	SetNumOfCapturesRsp struct {
		Stat
	}
)

func setNumOfCaptures(m *SetNumOfCapturesMsg) {
	err := m.ps.SetNoCaptures(m.NCaptures)
	response := &SetNumOfCapturesRsp{Stat: Stat{err: err}}
	m.respCh <- response
}

type (
	GetTriggerTimeOffsetsMsg struct {
		Unit
		SegmentIndex uint32
	}
	GetTriggerTimeOffsetRsp struct {
		Stat
		TimeUpper, TimeLower uint32
		TimeUnits            ps2000a.TimeUnits
	}
)

func getTriggerTimeOffset(m *GetTriggerTimeOffsetsMsg) {
	timeUpper, timeLower, timeUnits, err := m.ps.GetTriggerTimeOffset(m.SegmentIndex)
	response := &GetTriggerTimeOffsetRsp{Stat: Stat{err: err}, TimeUpper: timeUpper, TimeLower: timeLower, TimeUnits: timeUnits}
	m.respCh <- response
}

type (
	GetTriggerTimeOffsets64Msg struct {
		Unit
		SegmentIndex uint32
	}
	GetTriggerTimeOffset64Rsp struct {
		Stat
		Time      int64
		TimeUnits ps2000a.TimeUnits
	}
)

func getTriggerTimeOffset64(m *GetTriggerTimeOffsets64Msg) {
	time, timeUnits, err := m.ps.GetTriggerTimeOffset64(m.SegmentIndex)
	response := &GetTriggerTimeOffset64Rsp{Stat: Stat{err: err}, Time: time, TimeUnits: timeUnits}
	m.respCh <- response
}

type (
	GetValuesTriggerTimeOffsetBulkMsg struct {
		Unit
		TimesUpper, TimesLower           []uint32
		TimeUnits                        []ps2000a.TimeUnits
		FromSegmentIndex, ToSegmentIndex uint32
	}
	GetValuesTriggerTimeOffsetBulkRsp struct {
		Stat
	}
)

func getValuesTriggerTimeOffsetBulk(m *GetValuesTriggerTimeOffsetBulkMsg) {
	err := m.ps.GetValuesTriggerTimeOffsetBulk(m.TimesUpper, m.TimesLower, m.TimeUnits, m.FromSegmentIndex, m.ToSegmentIndex)
	response := &GetValuesTriggerTimeOffsetBulkRsp{Stat: Stat{err: err}}
	m.respCh <- response
}

type (
	GetValuesTriggerTimeOffsetBulk64Msg struct {
		Unit
		Times                            []int64
		TimeUnits                        []ps2000a.TimeUnits
		FromSegmentIndex, ToSegmentIndex uint32
	}
	GetValuesTriggerTimeOffsetBulk64Rsp struct {
		Stat
	}
)

func getValuesTriggerTimeOffsetBulk64(m *GetValuesTriggerTimeOffsetBulk64Msg) {
	err := m.ps.GetValuesTriggerTimeOffsetBulk64(m.Times, m.TimeUnits, m.FromSegmentIndex, m.ToSegmentIndex)
	response := &GetValuesTriggerTimeOffsetBulk64Rsp{Stat: Stat{err: err}}
	m.respCh <- response
}

type (
	HoldOffMsg struct {
		Unit
		HoldOff     uint64
		HoldOffType ps2000a.HoldOffType
	}
	HoldOffRsp struct {
		Stat
	}
)

func holdOff(m *HoldOffMsg) {
	err := m.ps.HoldOff(m.HoldOff, m.HoldOffType)
	response := &HoldOffRsp{Stat: Stat{err: err}}
	m.respCh <- response
}

type (
	LsReadyMsg struct {
		Unit
	}
	LsReadyRsp struct {
		Stat
		Ready int16
	}
)

func lsReady(m *LsReadyMsg) {
	ready, err := m.ps.LsReady()
	response := &LsReadyRsp{Stat: Stat{err: err}, Ready: ready}
	m.respCh <- response
}

type (
	TriggerOrPulseWidthQualifierEnabledMsg struct {
		Unit
	}
	TriggerOrPulseWidthQualifierEnabledRsp struct {
		Stat
		TriggerEnabled, PulseWidthQualifierEnabledint16 int16
	}
)

func triggerOrPulseWidthQualifierEnabled(m *TriggerOrPulseWidthQualifierEnabledMsg) {
	triggerEnabled, pulseWidthQualifierEnabledint16, err := m.ps.TriggerOrPulseWidthQualifierEnabled()
	response := &TriggerOrPulseWidthQualifierEnabledRsp{Stat: Stat{err: err},
		TriggerEnabled: triggerEnabled, PulseWidthQualifierEnabledint16: pulseWidthQualifierEnabledint16}
	m.respCh <- response
}

type (
	MemorySegmentsMsg struct {
		Unit
		NSegments uint32
	}
	MemorySegmentsRsp struct {
		Stat
		NMaxSamples int32
	}
)

func memorySegments(m *MemorySegmentsMsg) {
	nMaxSamples, err := m.ps.MemorySegments(m.NSegments)
	response := &MemorySegmentsRsp{Stat: Stat{err: err}, NMaxSamples: nMaxSamples}
	m.respCh <- response
}

type (
	NumOfStreamingValuesMsg struct {
		Unit
	}
	NumOfStreamingValuesRsp struct {
		Stat
		NumOfValues uint32
	}
)

func numOfStreamingValues(m *NumOfStreamingValuesMsg) {
	numOfValues, err := m.ps.NoOfStreamingValues()
	response := &NumOfStreamingValuesRsp{Stat: Stat{err: err}, NumOfValues: numOfValues}
	m.respCh <- response
}

type (
	OpenUnitProgressMsg struct {
		Unit
	}
	OpenUnitProgressRsp struct {
		Stat
		RetHandle, ProgressPercent, Complete int16
	}
)

func openUnitProgress(m *OpenUnitProgressMsg) {
	retHandle, progressPercent, complete, err := m.ps.OpenUnitProgress()
	response := &OpenUnitProgressRsp{Stat: Stat{err: err}, RetHandle: retHandle, ProgressPercent: progressPercent, Complete: complete}
	m.respCh <- response
}

type (
	PingUnitMsg struct {
		Unit
	}
	PingUnitRsp struct {
		Stat
	}
)

func pingUnit(m *PingUnitMsg) {
	err := m.ps.PingUnit()
	response := &PingUnitRsp{Stat: Stat{err: err}}
	m.respCh <- response
}

type (
	QueryOutputEdgeDetectMsg struct {
		Unit
	}
	QueryOutputEdgeDetectRsp struct {
		Stat
		State int16
	}
)

func queryOutputEdgeDetect(m *QueryOutputEdgeDetectMsg) {
	state, err := m.ps.QueryOutputEdgeDetect()
	response := &QueryOutputEdgeDetectRsp{Stat: Stat{err: err}, State: state}
	m.respCh <- response
}

type (
	SetDigitalAnalogTriggerOperandMsg struct {
		Unit
		Operand ps2000a.TriggerOperand
	}
	SetDigitalAnalogTriggerOperandRsp struct {
		Stat
	}
)

func setDigitalAnalogTriggerOperand(m *SetDigitalAnalogTriggerOperandMsg) {
	err := m.ps.SetDigitalAnalogTriggerOperand(m.Operand)
	response := &SetDigitalAnalogTriggerOperandRsp{Stat: Stat{err: err}}
	m.respCh <- response
}

type (
	SetDigitalPortMsg struct {
		Unit
		Port       ps2000a.DigitalPort
		Enabled    bool
		Logiclevel int16
	}
	SetDigitalPortRsp struct {
		Stat
	}
)

func setDigitalPort(m *SetDigitalPortMsg) {
	err := m.ps.SetDigitalPort(m.Port, m.Enabled, m.Logiclevel)
	response := &SetDigitalPortRsp{Stat: Stat{err: err}}
	m.respCh <- response
}

type (
	SetOutputEdgeDetectMsg struct {
		Unit
		State int16
	}
	SetOutputEdgeDetectRsp struct {
		Stat
	}
)

func setOutputEdgeDetect(m *SetOutputEdgeDetectMsg) {
	err := m.ps.SetOutputEdgeDetect(m.State)
	response := &SetOutputEdgeDetectRsp{Stat: Stat{err: err}}
	m.respCh <- response
}

type (
	SetPulseWidthDigitalPortPropertiesMsg struct {
		Unit
		DigitalDirections []ps2000a.DigitalChannelDirections
	}
	SetPulseWidthDigitalPortPropertiesRsp struct {
		Stat
	}
)

func setPulseWidthDigitalPortProperties(m *SetPulseWidthDigitalPortPropertiesMsg) {
	err := m.ps.SetPulseWidthDigitalPortProperties(m.DigitalDirections)
	response := &SetPulseWidthDigitalPortPropertiesRsp{Stat: Stat{err: err}}
	m.respCh <- response
}

type (
	SetSigGenArbitraryMsg struct {
		Unit
		OffsetVoltage                                                    int32
		PkToPK                                                           uint32
		StartDeltaPhase, StopDeltaPhase, DeltaPhaseIncrement, DwellCount uint32
		ArbitraryWaveform                                                []int16
		SweepType                                                        ps2000a.SweepTypeEnum
		Operation                                                        ps2000a.ExtraOperations
		IndexMode                                                        ps2000a.IndexMode
		Shots, Sweeps                                                    uint32
		TtriggerType                                                     ps2000a.SigGenTrigType
		TriggerSource                                                    ps2000a.SigGenTrigSource
		ExtInThreshold                                                   int16
	}
	SetSigGenArbitraryRsp struct {
		Stat
	}
)

func setSigGenArbitrary(m *SetSigGenArbitraryMsg) {
	err := m.ps.SetSigGenArbitrary(m.OffsetVoltage, m.PkToPK, m.StartDeltaPhase,
		m.StopDeltaPhase, m.DeltaPhaseIncrement, m.DwellCount, m.ArbitraryWaveform,
		m.SweepType, m.Operation, m.IndexMode, m.Shots, m.Sweeps, m.TtriggerType,
		m.TriggerSource, m.ExtInThreshold)
	response := &SetSigGenArbitraryRsp{Stat: Stat{err: err}}
	m.respCh <- response
}

type (
	SetSigGenPropertiesArbitraryMsg struct {
		Unit
		OffsetVoltage                                                    int32
		StartDeltaPhase, StopDeltaPhase, DeltaPhaseIncrement, DwellCount uint32
		SweepType                                                        ps2000a.SweepTypeEnum
		Operation                                                        ps2000a.ExtraOperations
		IndexMode                                                        ps2000a.IndexMode
		Shots, Sweeps                                                    uint32
		TriggerType                                                      ps2000a.SigGenTrigType
		TriggerSource                                                    ps2000a.SigGenTrigSource
		ExtInThreshold                                                   int16
	}
	SetSigGenPropertiesArbitraryRsp struct {
		Stat
	}
)

func setSigGenPropertiesArbitrary(m *SetSigGenPropertiesArbitraryMsg) {
	err := m.ps.SetSigGenPropertiesArbitrary(m.OffsetVoltage, m.StartDeltaPhase,
		m.StopDeltaPhase, m.DeltaPhaseIncrement, m.DwellCount, m.SweepType,
		m.Operation, m.IndexMode, m.Shots, m.Sweeps, m.TriggerType, m.TriggerSource, m.ExtInThreshold)
	response := &SetSigGenPropertiesArbitraryRsp{Stat: Stat{err: err}}
	m.respCh <- response
}

type (
	SetSigGenPropertiesBuiltInMsg struct {
		Unit
		OffsetVoltage                                       int32
		StartFrequency, StopFrequency, Increment, DwellTime float64
		SweepType                                           ps2000a.SweepTypeEnum
		Shots, Sweeps                                       uint32
		TriggerType                                         ps2000a.SigGenTrigType
		TriggerSource                                       ps2000a.SigGenTrigSource
		ExtInThreshold                                      int16
	}
	SetSigGenPropertiesBuiltInRsp struct {
		Stat
	}
)

func setSigGenPropertiesBuiltIn(m *SetSigGenPropertiesBuiltInMsg) {
	err := m.ps.SetSigGenPropertiesBuiltIn(m.OffsetVoltage, m.StartFrequency,
		m.StopFrequency, m.Increment, m.DwellTime, m.SweepType, m.Shots, m.Sweeps,
		m.TriggerType, m.TriggerSource, m.ExtInThreshold)
	response := &SetSigGenPropertiesBuiltInRsp{Stat: Stat{err: err}}
	m.respCh <- response
}

type (
	SigGenArbitraryMinMaxValuesMsg struct {
		Unit
	}
	SigGenArbitraryMinMaxValuesRsp struct {
		Stat
		MinArbitraryWaveformValue, MaxArbitraryWaveformValue int16
		MinArbitraryWaveformSize, MaxArbitraryWaveformSize   uint32
	}
)

func sigGenArbitraryMinMaxValues(m *SigGenArbitraryMinMaxValuesMsg) {
	minArbitraryWaveformValue, maxArbitraryWaveformValue, minArbitraryWaveformSize,
		maxArbitraryWaveformSize, err := m.ps.SigGenArbitraryMinMaxValues()
	response := &SigGenArbitraryMinMaxValuesRsp{Stat: Stat{err: err}, MinArbitraryWaveformValue: minArbitraryWaveformValue,
		MaxArbitraryWaveformValue: maxArbitraryWaveformValue, MinArbitraryWaveformSize: minArbitraryWaveformSize,
		MaxArbitraryWaveformSize: maxArbitraryWaveformSize}
	m.respCh <- response
}

type (
	SigGenSoftwareControlMsg struct {
		Unit
		State int16
	}
	SigGenSoftwareControlRsp struct {
		Stat
	}
)

func sigGenSoftwareControl(m *SigGenSoftwareControlMsg) {
	err := m.ps.SigGenSoftwareControl(m.State)
	response := &SigGenSoftwareControlRsp{Stat: Stat{err: err}}
	m.respCh <- response
}

// type (
// 	Msg struct {
// 		Unit
// 	}
// 	Rsp struct {
// 		Stat
// 	}
// )

// func (m *) {
// 	err := m.ps.()
// 	response := &{Stat: Stat{err: err},}
// 	m.respCh <- response
// }
