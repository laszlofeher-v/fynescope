package genericps

import (
	"fmt"
	"log/slog"
	"time"
)

type (
	NumOfChannelEnum int
)

const (
	cmdSendTimeout         = 5000 * time.Millisecond
	responseReceiveTimeout = 5000 * time.Millisecond
	MaxDwellTime           = 100000000
)

const (
	DualScope  NumOfChannelEnum = 2
	QuadScope  NumOfChannelEnum = 4
	MaxChannel NumOfChannelEnum = 4
)
const (
	NoSweep = SweepTypeEnum(-1)
)

type (
	RatioMode          int
	ChannelId          int
	Coupling           int
	RangeEnum          int
	PicoInfo           int
	ThresholdDirection int
	EtsMode            int
	TimeUnits          int
	TriggerRespBase    int
	ThresholdModeId    int
	PulseWidthType     int
	WaveTypeEnum       int
	DigitalChannel     int
	DigitalDirection   int
	SweepTypeEnum      int
	ExtraOperations    int
	IndexMode          int
	SigGenTrigType     int
	SigGenTrigSource   int
	HoldOffType        int
	TriggerOperand     int
	DigitalPort        int

	BlockReady     func(handle int16, status int, param any)
	DataReady      func(handle int16, status int, numOfSamples uint32, overflow int16, param any)
	StreamingReady func(handle int16, numOfSamples int32, startIndex uint32, overflow int16,
		triggeredAt uint32, triggered, autoStop int16, param any) (err error)

	TriggerChannelProperties struct {
		ThresholdUpper           int16
		ThresholdUpperHysteresis uint16
		ThresholdLower           int16
		ThresholdLowerHysteresis uint16
		Channel                  ChannelId
		ThresholdMode            ThresholdModeId
	}
	PwqConditions struct {
		ChannelA TriggerRespBase
		ChannelB TriggerRespBase
		ChannelC TriggerRespBase
		ChannelD TriggerRespBase
		External TriggerRespBase
		Aux      TriggerRespBase
		Digital  TriggerRespBase
	}
	TriggerConditions struct {
		ChannelA            TriggerRespBase
		ChannelB            TriggerRespBase
		ChannelC            TriggerRespBase
		ChannelD            TriggerRespBase
		External            TriggerRespBase
		Aux                 TriggerRespBase
		PulseWidthQualifier TriggerRespBase
		Digital             TriggerRespBase
	}
	DigitalChannelDirections struct {
		Channel   DigitalChannel
		Direction DigitalDirection
	}

	RespBase struct {
		status error // error from the scope
	}

	Response interface {
		Status() error
		SetStatus(err error)
	}

	Message interface {
		SetHandle(handle int16)
		SetRspCh(rspCh chan struct{})
		// SetRsp(rsp Response)
		Handle() (handle int16)
		SetStatus(err error)
		// RspCh() (rspCh chan Response)
		RspCh() (rspCh chan struct{})
		Rsp() (rsp Response)
	}

	MsgBase struct {
		handle int16
		rspCh  chan struct{} // only a flag
		rsp    Response      // every message has its own response type
	}

	Connection struct {
		Handle int16
		ID     string
		MsgCh  chan Message
		RspCh  chan struct{}
	}

	NullMsg struct {
		MsgBase
	}
	NullRsp struct {
		RespBase
	}

	GetValuesMsg struct {
		MsgBase
		StartIndex          uint32
		ReqNumOfSamples     uint32
		DownSampleRatio     uint32
		DownSampleRatioMode RatioMode
		SegmentIndex        uint32
	}

	GetValuesRsp struct {
		RespBase
		NumOfSamples uint32
		Overflow     int16
	}

	CloseUnitMsg struct {
		MsgBase
	}
	CloseUnitRsp struct {
		RespBase
	}

	SetChannelMsg struct {
		MsgBase
		Channel      ChannelId
		Enabled      bool
		X10          bool
		CouplingType Coupling
		VoltageRange RangeEnum
		AnalogOffset float32
	}
	SetChannelRsp struct {
		RespBase
	}

	FlashLedMsg struct {
		MsgBase
		Start int16
	}
	FlashLedRsp struct {
		RespBase
	}
	// EnumerateUnitMsg struct {
	// 	MsgBase
	// }
	// EnumerateUnitsRsp struct {
	// 	RespBase
	// }
	GetUnitInfoMsg struct {
		MsgBase
		Info PicoInfo
	}
	GetUnitInfoRsp struct {
		RespBase
		InfoString string
	}
	GetValuesAsyncMsg struct {
		MsgBase
		StartIndex          uint32
		NumOfSamples        uint32
		DownSampleRatio     uint32
		DownSampleRatioMode RatioMode
		LpDataReady         DataReady
		SegmentIndex        uint32
		Param               any
	}
	GetValuesAsyncRsp struct {
		RespBase
	}
	GetValuesBulkMsg struct {
		MsgBase
		ReqNumOfSamples     uint32
		FromSegmentIndex    uint32
		ToSegmentIndex      uint32
		DownSampleRatio     uint32
		DownSampleRatioMode RatioMode
		Overflow            []int16
	}
	GetValuesBulkRsp struct {
		RespBase
		NumOfSamples uint32
	}
	GetValuesOverlappedMsg struct {
		MsgBase
		StartIndex          uint32
		ReqNumOfSamples     uint32
		DownSampleRatio     uint32
		DownSampleRatioMode RatioMode
		SegmentIndex        uint32
		Overflow            []int16
	}
	GetValuesOverlappedRsp struct {
		RespBase
		NumOfSamples uint32
	}
	GetValuesOverlappedBulkMsg struct {
		MsgBase
		StartIndex          uint32
		ReqNumOfSamples     uint32
		DownSampleRatio     uint32
		DownSampleRatioMode RatioMode
		FromSegment         uint32
		ToSegment           uint32
		Overflow            []int16
	}
	GetValuesOverlappedBulkRsp struct {
		RespBase
		NumOfSamples uint32
	}
	GetAnalogueOffsetMsg struct {
		MsgBase
		VoltageRange int
		Coupling     Coupling
	}
	GetAnalogueOffsetRsp struct {
		RespBase
		MaximumVoltage float32
		MinimumVoltage float32
	}
	GetChannelInformationMsg struct {
		MsgBase
		Info    int16
		Probe   int32
		Ranges  []int32
		Channel ChannelId
	}
	GetChannelInformationRsp struct {
		RespBase
		Ranges         []int32
		LengthOfRanges int32
	}
	GetMaxDownSampleRatioMsg struct {
		MsgBase
		NumOfUnaggregatedSamples uint32
		DownSampleRatioMode      RatioMode
		SegmentIndex             int32
	}
	GetMaxDownSampleRatioRsp struct {
		RespBase
		MaxDownSampleRatio uint32
	}
	GetMaxSegmentsMsg struct {
		MsgBase
	}
	GetMaxSegmentsRsp struct {
		RespBase
		MaxSegments uint32
	}

	GetNumOfCapturesMsg struct {
		MsgBase
	}
	GetNumOfCapturesRsp struct {
		RespBase
		NCaptures uint32
	}

	GetNumOfProcessedCapturesMsg struct {
		MsgBase
	}
	GetNumOfProcessedCapturesRsp struct {
		RespBase
		NCaptures uint32
	}

	GetStreamingLatestValuesMsg struct {
		MsgBase
		LpStreamingReadyGoPar StreamingReady
		Param                 any
	}
	GetStreamingLatestValuesRsp struct {
		RespBase
	}

	GetTimebaseMsg struct {
		MsgBase
		TimeBase     uint32
		NumOfSamples int32
		OverSample   int16
		SegmentIndex uint32
	}
	GetTimebaseRsp struct {
		RespBase
		TimeIntervalNanoseconds int32
		MaxSamples              int32
	}

	GetTimebase2Msg struct {
		MsgBase
		TimeBase     uint32
		NumOfSamples int32
		OverSample   int16
		SegmentIndex uint32
	}
	GetTimebase2Rsp struct {
		RespBase
		TimeIntervalNanoseconds float32
		MaxSamples              int32
	}

	MaximumValueMsg struct {
		MsgBase
	}
	MaximumValueRsp struct {
		RespBase
		Value int16
	}

	MinimumValueMsg struct {
		MsgBase
	}
	MinimumValueResp struct {
		RespBase
		Value int16
	}

	SetSimpleTriggerMsg struct {
		MsgBase
		Enable        bool
		Source        ChannelId
		Threshold     int16
		Direction     ThresholdDirection
		Delay         uint32
		AutoTriggerMs int16
	}
	SetSimpleTriggerRsp struct {
		RespBase
	}

	SetDataBufferMsg struct {
		MsgBase
		Ch           ChannelId
		BufferIn     []int16
		SegmentIndex uint32
		Mode         RatioMode
	}
	SetDataBufferRsp struct {
		RespBase
	}

	SetDataBuffersMsg struct {
		MsgBase
		Ch           ChannelId
		BufferMax    []int16
		BufferMin    []int16
		SegmentIndex uint32
		Mode         RatioMode
	}
	SetDataBuffersRsp struct {
		RespBase
	}

	SetUnscaledDataBuffersMsg struct {
		MsgBase
		Ch           ChannelId
		BufferMax    []int16
		BufferMin    []int16
		SegmentIndex uint32
		Mode         RatioMode
	}
	SetUnscaledataBuffersRsp struct {
		RespBase
	}

	SetEtsTimeBufferMsg struct {
		MsgBase
		Buffer []int64
	}
	SetEtsTimeBufferRsp struct {
		RespBase
	}

	SetEtsTimeBuffersMsg struct {
		MsgBase
		TimeUpper, TimeLower []uint32
	}
	SetEtsTimeBuffersRsp struct {
		RespBase
	}

	SetEtsMsg struct {
		MsgBase
		Mode          EtsMode
		EtsCycles     int16
		EtsInterleave int16
	}
	SetEtsRsp struct {
		RespBase
		SampleTimePicoseconds int32
	}

	RunStreamingMsg struct {
		MsgBase
		ReqSampleInterval                           uint32
		SampleIntervalTimeUnits                     TimeUnits
		MaxPreTriggerSamples, MaxPostTriggerSamples uint32
		AutoStop                                    bool
		DownSampleRatio                             uint32
		DownSampleRatioMode                         RatioMode
		OverviewBufferSize                          uint32
	}
	RunStreamingRsp struct {
		RespBase
		SampleInterval uint32
	}

	RunBlockMsg struct {
		MsgBase
		NumOfPreTriggerSamples  int32
		NumOfPostTriggerSamples int32
		TimeBase                uint32
		OverSample              int16
		SegmentIndex            uint32
		LpBlockReadyGoPar       BlockReady
		Param                   any
	}
	RunBlockRsp struct {
		RespBase
		TimeIndisposedMs int32
	}

	SetTriggerChannelPropertiesMsg struct {
		MsgBase
		ChannelProperties []TriggerChannelProperties
		AuxOutputEnable   bool
		AutoTriggerMs     int32
	}
	SetTriggerChannelPropertiesRsp struct {
		RespBase
	}

	SetTriggerChannelConditionsMsg struct {
		MsgBase
		TriggerConditions []TriggerConditions
	}
	SetTriggerChannelConditionsRsp struct {
		RespBase
	}

	SetTriggerChannelDirectionsMsg struct {
		MsgBase
		ChannelA, ChannelB, ChannelC, ChannelD, Ext, Aux ThresholdDirection
	}
	SetTriggerChannelDirectionsRsp struct {
		RespBase
	}

	SetTriggerDelayMsg struct {
		MsgBase
		Delay uint32
	}
	SetTriggerDelayRsp struct {
		RespBase
	}

	SetPulseWidthQualifierMsg struct {
		MsgBase
		Conditions   []PwqConditions
		Direction    ThresholdDirection
		Lower, Upper uint32
		PwType       PulseWidthType
	}
	SetPulseWidthQualifierRsp struct {
		RespBase
	}

	SetTriggerDigitalPortPropertiesMsg struct {
		MsgBase
		DigitalDirections []DigitalChannelDirections
	}
	SetTriggerDigitalPortPropertiesRsp struct {
		RespBase
	}

	StopMsg struct {
		MsgBase
	}
	StopRsp struct {
		RespBase
	}

	SetSigGenBuiltInMsg struct {
		MsgBase
		OffsetVoltage                                       int32
		PkToPK                                              uint32
		WaveType                                            WaveTypeEnum
		StartFrequency, StopFrequency, Increment, DwellTime float32
		SweepType                                           SweepTypeEnum
		Operation                                           ExtraOperations
		Shots, Sweeps                                       uint32
		TriggerType                                         SigGenTrigType
		TriggerSource                                       SigGenTrigSource
		ExtInThreshold                                      int16
	}
	SetSigGenBuiltInRsp struct {
		RespBase
	}

	SetSigGenBuiltInV2Msg struct {
		MsgBase
		OffsetVoltage                                       int32
		PkToPK                                              uint32
		WaveType                                            WaveTypeEnum
		StartFrequency, StopFrequency, Increment, DwellTime float64
		SweepType                                           SweepTypeEnum
		Operation                                           ExtraOperations
		Shots, Sweeps                                       uint32
		TriggerType                                         SigGenTrigType
		TriggerSource                                       SigGenTrigSource
		ExtInThreshold                                      int16
	}
	SetSigGenBuiltInV2Rsp struct {
		RespBase
	}

	SetSimGenMsg struct {
		MsgBase
		Channel                                             ChannelId
		On                                                  bool
		OffsetVoltage                                       int32
		PkToPK                                              uint32
		WaveType                                            WaveTypeEnum
		StartFrequency, StopFrequency, Increment, DwellTime float64
		SweepType                                           SweepTypeEnum
		Operation                                           ExtraOperations
		Shots, Sweeps                                       uint32
		TriggerType                                         SigGenTrigType
		TriggerSource                                       SigGenTrigSource
		ExtInThreshold                                      int16
		Phase                                               float64
	}
	SetSimGenRsp struct {
		RespBase
	}

	SetSimRlcFilterMsg struct {
		MsgBase
		Channel    ChannelId
		GenSource  ChannelId
		Enabled    bool
		FilterType string
		R          float64
		RUnit      string
		L          float64
		LUnit      string
		C          float64
		CUnit      string
	}
	SetSimRlcFilterRsp struct {
		RespBase
	}

	SetSimDigitalFilterMsg struct {
		MsgBase
		Channel         ChannelId
		LowpassEnabled  bool
		LowpassFc       float64
		HighpassEnabled bool
		HighpassFc      float64
		BandpassEnabled bool
		BandpassFc1     float64
		BandpassFc2     float64
		BandstopEnabled bool
		BandstopFc1     float64
		BandstopFc2     float64
	}
	SetSimDigitalFilterRsp struct {
		RespBase
	}

	SigGenFrequencyToPhasenMsg struct {
		Frequency float64
		MsgBase
		IndexMode    IndexMode
		BufferLength uint32
	}
	SigGenFrequencyToPhaseRsp struct {
		RespBase
		Phase uint32
	}

	SetNumOfCapturesMsg struct {
		MsgBase
		NCaptures uint32
	}
	SetNumOfCapturesRsp struct {
		RespBase
	}

	GetTriggerTimeOffsetMsg struct {
		MsgBase
		SegmentIndex uint32
	}
	GetTriggerTimeOffsetRsp struct {
		RespBase
		TimeUpper, TimeLower uint32
		TimeUnits            TimeUnits
	}

	GetTriggerTimeOffset64Msg struct {
		MsgBase
		SegmentIndex uint32
	}
	GetTriggerTimeOffset64Rsp struct {
		RespBase
		Time      int64
		TimeUnits TimeUnits
	}

	GetValuesTriggerTimeOffsetBulkMsg struct {
		MsgBase
		TimesUpper, TimesLower           []uint32
		TimeUnits                        []TimeUnits
		FromSegmentIndex, ToSegmentIndex uint32
	}
	GetValuesTriggerTimeOffsetBulkRsp struct {
		RespBase
	}

	GetValuesTriggerTimeOffsetBulk64Msg struct {
		MsgBase
		Times                            []int64
		TimeUnits                        []TimeUnits
		FromSegmentIndex, ToSegmentIndex uint32
	}
	GetValuesTriggerTimeOffsetBulk64Rsp struct {
		RespBase
	}

	HoldOffMsg struct {
		MsgBase
		HoldOff     uint64
		HoldOffType HoldOffType
	}
	HoldOffRsp struct {
		RespBase
	}

	LsReadyMsg struct {
		MsgBase
	}
	LsReadyRsp struct {
		RespBase
		Ready int16
	}

	TriggerOrPulseWidthQualifierEnabledMsg struct {
		MsgBase
	}
	TriggerOrPulseWidthQualifierEnabledRsp struct {
		RespBase
		TriggerEnabled, PulseWidthQualifierEnabledint16 int16
	}

	MemorySegmentsMsg struct {
		MsgBase
		NSegments uint32
	}
	MemorySegmentsRsp struct {
		RespBase
		NMaxSamples int32
	}

	NumOfStreamingValuesMsg struct {
		MsgBase
	}
	NumOfStreamingValuesRsp struct {
		RespBase
		NumOfValues uint32
	}

	OpenUnitProgressMsg struct {
		MsgBase
	}
	OpenUnitProgressRsp struct {
		RespBase
		RetHandle                 int16
		ProgressPercent, Complete int16
	}

	PingUnitMsg struct {
		MsgBase
	}
	PingUnitRsp struct {
		RespBase
	}

	QueryOutputEdgeDetectMsg struct {
		MsgBase
	}
	QueryOutputEdgeDetectRsp struct {
		RespBase
		State int16
	}

	SetDigitalAnalogTriggerOperandMsg struct {
		MsgBase
		Operand TriggerOperand
	}
	SetDigitalAnalogTriggerOperandRsp struct {
		RespBase
	}

	SetDigitalPortMsg struct {
		MsgBase
		Port       DigitalPort
		Enabled    bool
		Logiclevel int16
	}
	SetDigitalPortRsp struct {
		RespBase
	}

	SetOutputEdgeDetectMsg struct {
		MsgBase
		State int16
	}
	SetOutputEdgeDetectRsp struct {
		RespBase
	}

	SetPulseWidthDigitalPortPropertiesMsg struct {
		MsgBase
		DigitalDirections []DigitalChannelDirections
	}
	SetPulseWidthDigitalPortPropertiesRsp struct {
		RespBase
	}

	SetSigGenArbitraryMsg struct {
		MsgBase
		OffsetVoltage                                                    int32
		PkToPK                                                           uint32
		StartDeltaPhase, StopDeltaPhase, DeltaPhaseIncrement, DwellCount uint32
		ArbitraryWaveform                                                []int16
		SweepType                                                        SweepTypeEnum
		Operation                                                        ExtraOperations
		IndexMode                                                        IndexMode
		Shots, Sweeps                                                    uint32
		TtriggerType                                                     SigGenTrigType
		TriggerSource                                                    SigGenTrigSource
		ExtInThreshold                                                   int16
	}
	SetSigGenArbitraryRsp struct {
		RespBase
	}

	SetSigGenPropertiesArbitraryMsg struct {
		MsgBase
		OffsetVoltage                                                    int32
		StartDeltaPhase, StopDeltaPhase, DeltaPhaseIncrement, DwellCount uint32
		SweepType                                                        SweepTypeEnum
		Operation                                                        ExtraOperations
		IndexMode                                                        IndexMode
		Shots, Sweeps                                                    uint32
		TriggerType                                                      SigGenTrigType
		TriggerSource                                                    SigGenTrigSource
		ExtInThreshold                                                   int16
	}
	SetSigGenPropertiesArbitraryRsp struct {
		RespBase
	}

	SetSigGenPropertiesBuiltInMsg struct {
		MsgBase
		OffsetVoltage                                       int32
		StartFrequency, StopFrequency, Increment, DwellTime float64
		SweepType                                           SweepTypeEnum
		Shots, Sweeps                                       uint32
		TriggerType                                         SigGenTrigType
		TriggerSource                                       SigGenTrigSource
		ExtInThreshold                                      int16
	}
	SetSigGenPropertiesBuiltInRsp struct {
		RespBase
	}

	SigGenArbitraryMinMaxValuesMsg struct {
		MsgBase
	}
	SigGenArbitraryMinMaxValuesRsp struct {
		RespBase
		MinArbitraryWaveformValue, MaxArbitraryWaveformValue int16
		MinArbitraryWaveformSize, MaxArbitraryWaveformSize   uint32
	}

	SigGenSoftwareControlMsg struct {
		MsgBase
		State int16
	}
	SigGenSoftwareControlRsp struct {
		RespBase
	}
)
type (
	Constants interface {
		ChA() ChannelId
		ChB() ChannelId
		ChC() ChannelId
		ChD() ChannelId
		RatioModeNone() RatioMode
		RatioModeAggregate() RatioMode
		RatioModeDecimate() RatioMode
		RatioModeAverage() RatioMode
		Ac() Coupling
		Dc() Coupling
		Range_10mv() RangeEnum
		Range_20mv() RangeEnum
		Range_50mv() RangeEnum
		Range_100mv() RangeEnum
		Range_200mv() RangeEnum
		Range_500mv() RangeEnum
		Range_1v() RangeEnum
		Range_2v() RangeEnum
		Range_5v() RangeEnum
		Range_10v() RangeEnum
		Range_20v() RangeEnum
		Range_50v() RangeEnum
		// Level, Window                              ThresholdModeId
		// CondDontCare, CondTrue, CondFalse, CondMax TriggerRespBase
		// TriggerAbove, TriggerBelow, TriggerRaising, TriggerFalling, TriggerRisingOrFalling,
		// TriggerAboveLower, TriggerBelowLower, TriggerRisingLower, TriggerFallingLower,
		// TriggerOutside, TriggerInside, TriggerEnter, TriggerExit, TriggerEnterOrExit,
		// TriggerPositiveRunt, TriggerNegativeRunt, TriggerNone ThresholdDirection
		// Dch0, Dch1, Dch2, Dch3, Dch4, Dch5, Dch6, Dch7,
		// Dch8, Dch9, Dch10, Dch11, Dch12, Dch13, Dch14, Dch15,
		// Dch16, Dch17, Dch18, Dch19, Dch20, Dch21, Dch22, Dch23,
		// Dch24, Dch25, Dch26, Dch27, Dch28, Dch29, Dch30, Dch31, DchMax DigitalChannel
		// DigitalDontCare, DigitalDirectionLow, DigitalDirectionHigh, DigitalDirectionRising,
		// DigitalDirectionFalling, DigitalDirectionRisingOrFalling, DigitalMaxDirection DigitalDirection
		// SweepUp, SweepDown, SweepUpDown, SweepDownUp, SweepMaxTypes           SweepTypeEnum
		// EsOff, WhiteNoise, Prbs                                               ExtraOperations
		// Single, Dual, Quad, MaxIndexModes                                     IndexMode
		// SigGenRising, SigGenFalling, SigGenGateHigh, SigGenGateLow            SigGenTrigType
		// SigGenNone, SigGenScopeTrig, SigGenAuxIn, SigGenExtIn, SigGenSoftTrig SigGenTrigSource
		// TuFs, TuPs, TuNs, TuUs, TuMs, TuS                                     TimeUnits
		// HofTime, MaxHoldOffTime                                               HoldOffType
		// OperandNone, OperandOr, OperandAnd, OperandThen                       TriggerOperand
		// Port0, Port1, Port2, Port3, MaxDigitalPorts                           DigitalPort
		// PicoDriverVersion, PicoUsbVersion, PicoHardwareVersion, PicoVariantInfo,
		// PicoBatchAndSerial, PicoCalDate, PicoKernelVersion, PicoDigitalHardwareVersion,
		// PicoAnalogueHardwareVersion, PicoFirmwareVersion1, PicoFirmwareVersion2,
		// PicoMacAddress, PicoShadowCall, PicoIppVersion, PicoDriverPath,
		// PicoFirmwareVersion3, PicoFrontPanelFirmwareVersion3, PicoBootloaderVersion PicoInfo
		// EtsOff, EtsFast, EtsSlow, EtsMax EtsMode
		// PwTypeNone, PwTypeLessThan, PwTypeGreaterThan,
		// PwTypeInRange, PwTypeOutOfRange PulseWidthType
		// Sine, Square, Triangle, RampUp, RampDown,
		// SinC, Gaussian, HalfSine, DcVoltage WaveTypeEnum
		// inputRanges       []int32
		// ChannelInfoRanges int16
		// RangeValuesMv     map[RangeEnum]float32
		// SineMaxFrequency, SquareMaxFrequency, TriangleMaxFrequency, SinCMaxFrequency,
		// RampMaxFrequency, HalfSineMaxFrequency, GaussianMaxFrequency,
		// PrbsMaxFrequency, PrbsMinFrequency, MinFrequency float64
		// AwgMinSigGenBufferSize, AwgMaxSigGenBufferSize int16
	}
)

var (
	ChA, ChB, ChC, ChD                                                     ChannelId
	RatioModeNone, RatioModeAggregate, RatioModeDecimate, RatioModeAverage RatioMode
	Ac, Dc                                                                 Coupling
	Range_10mv, Range_20mv, Range_50mv, Range_100mv, Range_200mv, Range_500mv,
	Range_1v, Range_2v, Range_5v, Range_10v, Range_20v, Range_50v RangeEnum
	Level, Window                              ThresholdModeId
	CondDontCare, CondTrue, CondFalse, CondMax TriggerRespBase
	TriggerAbove, TriggerBelow, TriggerRaising, TriggerFalling, TriggerRisingOrFalling,
	TriggerAboveLower, TriggerBelowLower, TriggerRisingLower, TriggerFallingLower,
	TriggerOutside, TriggerInside, TriggerEnter, TriggerExit, TriggerEnterOrExit,
	TriggerPositiveRunt, TriggerNegativeRunt, TriggerNone ThresholdDirection
	Dch0, Dch1, Dch2, Dch3, Dch4, Dch5, Dch6, Dch7,
	Dch8, Dch9, Dch10, Dch11, Dch12, Dch13, Dch14, Dch15,
	Dch16, Dch17, Dch18, Dch19, Dch20, Dch21, Dch22, Dch23,
	Dch24, Dch25, Dch26, Dch27, Dch28, Dch29, Dch30, Dch31, DchMax DigitalChannel
	DigitalDontCare, DigitalDirectionLow, DigitalDirectionHigh, DigitalDirectionRising,
	DigitalDirectionFalling, DigitalDirectionRisingOrFalling, DigitalMaxDirection DigitalDirection
	SweepUp, SweepDown, SweepUpDown, SweepDownUp, SweepMaxTypes           SweepTypeEnum
	EsOff, WhiteNoise, Prbs                                               ExtraOperations
	Single, Dual, Quad, MaxIndexModes                                     IndexMode
	SigGenRising, SigGenFalling, SigGenGateHigh, SigGenGateLow            SigGenTrigType
	SigGenNone, SigGenScopeTrig, SigGenAuxIn, SigGenExtIn, SigGenSoftTrig SigGenTrigSource
	TuFs, TuPs, TuNs, TuUs, TuMs, TuS                                     TimeUnits
	HofTime, MaxHoldOffTime                                               HoldOffType
	OperandNone, OperandOr, OperandAnd, OperandThen                       TriggerOperand
	Port0, Port1, Port2, Port3, MaxDigitalPorts                           DigitalPort
	PicoDriverVersion, PicoUsbVersion, PicoHardwareVersion, PicoVariantInfo,
	PicoBatchAndSerial, PicoCalDate, PicoKernelVersion, PicoDigitalHardwareVersion,
	PicoAnalogueHardwareVersion, PicoFirmwareVersion1, PicoFirmwareVersion2,
	PicoMacAddress, PicoShadowCall, PicoIppVersion, PicoDriverPath,
	PicoFirmwareVersion3, PicoFrontPanelFirmwareVersion3, PicoBootloaderVersion PicoInfo
	EtsOff, EtsFast, EtsSlow, EtsMax EtsMode
	PwTypeNone, PwTypeLessThan, PwTypeGreaterThan,
	PwTypeInRange, PwTypeOutOfRange PulseWidthType
	Sine, Square, Triangle, RampUp, RampDown,
	SinC, Gaussian, HalfSine, DcVoltage WaveTypeEnum
	InputRanges       []int32
	ChannelInfoRanges int16
	RangeValuesMv     map[RangeEnum]float64
	SineMaxFrequency, SquareMaxFrequency, TriangleMaxFrequency, SinCMaxFrequency,
	RampMaxFrequency, HalfSineMaxFrequency, GaussianMaxFrequency,
	PrbsMaxFrequency, PrbsMinFrequency, MinFrequency float64
	AwgMinSigGenBufferSize, AwgMaxSigGenBufferSize int16
)

func NewConnection() (con *Connection) {
	con = &Connection{}
	con.ID = ""
	con.MsgCh = make(chan Message)
	con.RspCh = make(chan struct{})
	return
}

// func (c Connection) EnumerateUnits() (count int16, serials string, serialLth int16, err error) {
// 	msg := &EnumerateUnitMsg{}
// 	msg.rsp = &EnumerateUnitsRsp{}
// 	c.Send(msg)
// 	rsp := msg.Rsp().(*EnumerateUnitsRsp)
// 	err = rsp.Status()
// 	return
// }

func OpenUnitAsync(serial string) (err error) {
	return
}

func OpenUnitProgress() (retHandle int16, progressPercent, complete int16, err error) {
	return
}

func (c Connection) Send(msg Message) {
	// sends msg to the scope and receives the response
	// handles timeout
	msg.SetHandle(c.Handle)
	msg.SetRspCh(c.RspCh)
	select {
	case c.MsgCh <- msg:
	case <-time.After(cmdSendTimeout):
		msg.SetStatus(fmt.Errorf("Timeout. Could not send %t\n", msg))
		return
	}
	select {
	case <-c.RspCh:
		// slog.Debug("response received", "c", c)
	case <-time.After(responseReceiveTimeout):
		msg.SetStatus(fmt.Errorf("Timeout. Could not receive response of %t\n", msg))
	}
}

func (c Connection) CloseUnit() (err error) {
	msg := &CloseUnitMsg{}
	msg.rsp = &CloseUnitRsp{}
	c.Send(msg)
	rsp := msg.Rsp().(*CloseUnitRsp)
	err = rsp.Status()
	// switch r := resp.(type) {
	// case *CloseUnitRsp:
	// 	err = r.Status()
	// default:
	// }
	return
}

func (c Connection) FlashLed(start int16) (err error) {
	var (
		msg FlashLedMsg
		rsp FlashLedRsp
	)
	c.Send(&msg)
	err = rsp.status
	return
}
func (c Connection) PingUnit() (err error) {
	return
}
func (c Connection) GetAnalogueOffset(voltageRange int, coupling Coupling) (maximumVoltage, minimumVoltage float32, err error) {
	msg := &GetAnalogueOffsetMsg{VoltageRange: voltageRange, Coupling: coupling}
	msg.rsp = &GetAnalogueOffsetRsp{}
	c.Send(msg)
	rsp := msg.Rsp().(*GetAnalogueOffsetRsp)
	err = rsp.Status()
	maximumVoltage = rsp.MaximumVoltage
	minimumVoltage = rsp.MinimumVoltage
	return
}
func (c Connection) GetChannelInformation(info int16, probe int32, ranges []int32, channels ChannelId) (lengthOfRanges int32, err error) {
	msg := &GetChannelInformationMsg{Info: info,
		Probe: probe, Ranges: ranges, Channel: channels}
	msg.rsp = &GetChannelInformationRsp{}
	c.Send(msg)
	rsp := msg.Rsp().(*GetChannelInformationRsp)
	err = rsp.Status()
	lengthOfRanges = rsp.LengthOfRanges
	return
}
func (c Connection) GetMaxDownSampleRatio(numOfUnaggregatedSamples uint32, downSampleRatioMode RatioMode, segmentIndex int32) (maxDownSampleRatio uint32, err error) {
	msg := &GetMaxDownSampleRatioMsg{NumOfUnaggregatedSamples: numOfUnaggregatedSamples, DownSampleRatioMode: downSampleRatioMode, SegmentIndex: segmentIndex}
	msg.rsp = &GetMaxDownSampleRatioRsp{}
	c.Send(msg)
	rsp := msg.Rsp().(*GetMaxDownSampleRatioRsp)
	maxDownSampleRatio = rsp.MaxDownSampleRatio
	err = rsp.Status()
	return
}
func (c Connection) GetMaxSegments() (maxSegments uint32, err error) {
	msg := &GetMaxSegmentsMsg{}
	msg.rsp = &GetMaxSegmentsRsp{}
	c.Send(msg)
	rsp := msg.Rsp().(*GetMaxSegmentsRsp)
	maxSegments = rsp.MaxSegments
	err = rsp.Status()
	return
}
func (c Connection) GetNumOfCaptures() (nCaptures uint32, err error) {
	msg := &GetNumOfCapturesMsg{}
	msg.rsp = &GetNumOfCapturesRsp{}
	c.Send(msg)
	rsp := msg.Rsp().(*GetNumOfCapturesRsp)
	nCaptures = rsp.NCaptures
	err = rsp.Status()
	return
}
func (c Connection) GetNumOfProcessedCaptures() (nCaptures uint32, err error) {
	msg := &GetNumOfProcessedCapturesMsg{}
	msg.rsp = &GetNumOfProcessedCapturesRsp{}
	c.Send(msg)
	rsp := msg.Rsp().(*GetNumOfProcessedCapturesRsp)
	nCaptures = rsp.NCaptures
	err = rsp.Status()
	return
}
func (c Connection) GetStreamingLatestValues(lpStreamingReadyGoPar StreamingReady, param any) (err error) {
	msg := &GetStreamingLatestValuesMsg{LpStreamingReadyGoPar: lpStreamingReadyGoPar, Param: param}
	msg.rsp = &GetStreamingLatestValuesRsp{}
	c.Send(msg)
	rsp := msg.Rsp().(*GetStreamingLatestValuesRsp)
	err = rsp.Status()
	return
}
func (c Connection) GetTimebase(timeBase uint32, numOfSamples int32, overSample int16, segmentIndex uint32) (timeIntervalNanoseconds, maxSamples int32, err error) {
	msg := &GetTimebaseMsg{TimeBase: timeBase, NumOfSamples: numOfSamples, OverSample: overSample, SegmentIndex: segmentIndex}
	msg.rsp = &GetTimebaseRsp{}
	c.Send(msg)
	rsp := msg.Rsp().(*GetTimebaseRsp)
	timeIntervalNanoseconds = rsp.TimeIntervalNanoseconds
	maxSamples = rsp.MaxSamples
	err = rsp.Status()
	return
}
func (c Connection) GetTimebase2(timeBase uint32, numOfSamples int32, overSample int16, segmentIndex uint32) (timeIntervalNanoseconds float32, maxSamples int32, err error) {
	msg := &GetTimebase2Msg{TimeBase: timeBase, NumOfSamples: numOfSamples, OverSample: overSample, SegmentIndex: segmentIndex}
	msg.rsp = &GetTimebase2Rsp{}
	c.Send(msg)
	rsp := msg.Rsp().(*GetTimebase2Rsp)
	timeIntervalNanoseconds = rsp.TimeIntervalNanoseconds
	maxSamples = rsp.MaxSamples
	slog.Debug("GetTimebase2", "rsp", rsp)
	err = rsp.Status()
	return
}
func (c Connection) GetTriggerTimeOffset(segmentIndex uint32) (timeUpper, timeLower uint32, timeUnits TimeUnits, err error) {
	msg := &GetTriggerTimeOffsetMsg{SegmentIndex: segmentIndex}
	msg.rsp = &GetTriggerTimeOffsetRsp{}
	c.Send(msg)
	rsp := msg.Rsp().(*GetTriggerTimeOffsetRsp)
	timeUpper = rsp.TimeUpper
	timeLower = rsp.TimeLower
	timeUnits = rsp.TimeUnits
	err = rsp.Status()
	return
}
func (c Connection) GetTriggerTimeOffset64(segmentIndex uint32) (time int64, timeUnits TimeUnits, err error) {
	msg := &GetTriggerTimeOffset64Msg{SegmentIndex: segmentIndex}
	msg.rsp = &GetTriggerTimeOffset64Rsp{}
	c.Send(msg)
	rsp := msg.Rsp().(*GetTriggerTimeOffset64Rsp)
	time = rsp.Time
	timeUnits = rsp.TimeUnits
	err = rsp.Status()
	return
}
func (c Connection) GetUnitInfo(info PicoInfo) (infoString string, err error) {
	msg := &GetUnitInfoMsg{Info: info}
	msg.rsp = &GetUnitInfoRsp{}
	c.Send(msg)
	rsp := msg.Rsp().(*GetUnitInfoRsp)
	infoString = rsp.InfoString
	err = rsp.Status()
	return
}

func (c Connection) GetValues(startIndex, reqNumOfSamples, downSampleRatio uint32,
	downSampleRatioMode RatioMode, segmentIndex uint32) (numOfSamples uint32, overflow int16, err error) {
	msg := &GetValuesMsg{StartIndex: startIndex, ReqNumOfSamples: reqNumOfSamples,
		DownSampleRatio: downSampleRatio, DownSampleRatioMode: downSampleRatioMode,
		SegmentIndex: segmentIndex}
	msg.rsp = &GetValuesRsp{}
	c.Send(msg)
	rsp := msg.Rsp().(*GetValuesRsp)
	err = rsp.Status()
	numOfSamples = rsp.NumOfSamples
	overflow = rsp.Overflow
	return
}
func (c Connection) GetValuesAsync(startIndex, numOfSamples, downSampleRatio uint32,
	downSampleRatioMode RatioMode, lpDataReadyGoPar DataReady, segmentIndex uint32,
	param any) (err error) {
	msg := &GetValuesAsyncMsg{StartIndex: startIndex, NumOfSamples: numOfSamples,
		DownSampleRatio: downSampleRatio, DownSampleRatioMode: downSampleRatioMode,
		LpDataReady: lpDataReadyGoPar, SegmentIndex: segmentIndex}
	msg.rsp = &GetValuesAsyncRsp{}
	c.Send(msg)
	rsp := msg.Rsp().(*GetValuesAsyncRsp)
	err = rsp.Status()
	return
}
func (c Connection) GetValuesBulk(reqNumOfSamples uint32, fromSegmentIndex, toSegmentIndex, downSampleRatio uint32,
	downSampleRatioMode RatioMode, overflow []int16) (numOfSamples uint32, err error) {
	msg := &GetValuesBulkMsg{ReqNumOfSamples: reqNumOfSamples, FromSegmentIndex: fromSegmentIndex,
		ToSegmentIndex: toSegmentIndex, DownSampleRatio: downSampleRatio, DownSampleRatioMode: downSampleRatioMode,
		Overflow: overflow}
	msg.rsp = &GetValuesBulkRsp{}
	c.Send(msg)
	rsp := msg.Rsp().(*GetValuesBulkRsp)
	numOfSamples = rsp.NumOfSamples
	err = rsp.Status()
	return
}
func (c Connection) GetValuesOverlapped(startIndex, reqNumOfSamples, downSampleRatio uint32,
	downSampleRatioMode RatioMode, segmentIndex uint32, overflow []int16) (numOfSamples uint32, err error) {
	msg := &GetValuesOverlappedMsg{StartIndex: startIndex, ReqNumOfSamples: reqNumOfSamples,
		DownSampleRatio: downSampleRatio, DownSampleRatioMode: downSampleRatioMode,
		SegmentIndex: segmentIndex, Overflow: overflow}
	msg.rsp = &GetValuesOverlappedRsp{}
	c.Send(msg)
	rsp := msg.Rsp().(*GetValuesOverlappedRsp)
	numOfSamples = rsp.NumOfSamples
	err = rsp.Status()
	return
}
func (c Connection) GetValuesOverlappedBulk(startIndex, reqNumOfSamples, downSampleRatio uint32,
	downSampleRatioMode RatioMode, fromSegmentIndex, toSegmentIndex uint32, overflow []int16) (numOfSamples uint32, err error) {
	msg := &GetValuesOverlappedBulkMsg{StartIndex: startIndex, ReqNumOfSamples: reqNumOfSamples,
		DownSampleRatio: downSampleRatio, DownSampleRatioMode: downSampleRatioMode,
		FromSegment: fromSegmentIndex, ToSegment: toSegmentIndex, Overflow: overflow}
	msg.rsp = &GetValuesOverlappedBulkRsp{}
	c.Send(msg)
	rsp := msg.Rsp().(*GetValuesOverlappedBulkRsp)
	numOfSamples = rsp.NumOfSamples
	err = rsp.Status()
	return
}
func (c Connection) GetValuesTriggerTimeOffsetBulk(timesUpper, timesLower []uint32, timeUnits []TimeUnits,
	fromSegmentIndex, toSegmentIndex uint32) (err error) {
	msg := &GetValuesTriggerTimeOffsetBulkMsg{TimesUpper: timesUpper, TimesLower: timesLower,
		TimeUnits: timeUnits, FromSegmentIndex: fromSegmentIndex, ToSegmentIndex: toSegmentIndex}
	msg.rsp = &GetValuesTriggerTimeOffsetBulkRsp{}
	c.Send(msg)
	rsp := msg.Rsp().(*GetValuesTriggerTimeOffsetBulkRsp)
	err = rsp.Status()
	return
}
func (c Connection) GetValuesTriggerTimeOffsetBulk64(times []int64, timeUnits []TimeUnits,
	fromSegmentIndex, toSegmentIndex uint32) (err error) {
	msg := &GetValuesTriggerTimeOffsetBulk64Msg{Times: times, TimeUnits: timeUnits,
		FromSegmentIndex: fromSegmentIndex, ToSegmentIndex: toSegmentIndex}
	msg.rsp = &GetValuesTriggerTimeOffsetBulk64Rsp{}
	c.Send(msg)
	rsp := msg.Rsp().(*GetValuesTriggerTimeOffsetBulk64Rsp)
	err = rsp.Status()
	return
}
func (c Connection) HoldOff(holdOff uint64, holdOffType HoldOffType) (err error) {
	msg := &HoldOffMsg{HoldOff: holdOff, HoldOffType: holdOffType}
	msg.rsp = &HoldOffRsp{}
	c.Send(msg)
	rsp := msg.Rsp().(*HoldOffRsp)
	err = rsp.Status()
	return
}
func (c Connection) LsReady() (ready int16, err error) {
	msg := &LsReadyMsg{}
	msg.rsp = &LsReadyRsp{}
	c.Send(msg)
	rsp := msg.Rsp().(*LsReadyRsp)
	ready = rsp.Ready
	err = rsp.Status()
	return
}
func (c Connection) MaximumValue() (value int16, err error) {
	msg := &MaximumValueMsg{}
	msg.rsp = &MaximumValueRsp{}
	c.Send(msg)
	rsp := msg.Rsp().(*MaximumValueRsp)
	value = rsp.Value
	err = rsp.Status()
	return
}
func (c Connection) MemorySegments(nSegments uint32) (nMaxSamples int32, err error) {
	msg := &MemorySegmentsMsg{NSegments: nSegments}
	msg.rsp = &MemorySegmentsRsp{}
	c.Send(msg)
	rsp := msg.Rsp().(*MemorySegmentsRsp)
	nMaxSamples = rsp.NMaxSamples
	err = rsp.Status()
	return
}
func (c Connection) MinimumValue() (value int16, err error) {
	msg := &MinimumValueMsg{}
	msg.rsp = &MinimumValueResp{}
	c.Send(msg)
	rsp := msg.Rsp().(*MinimumValueResp)
	value = rsp.Value
	err = rsp.Status()
	return
}
func (c Connection) NumOfStreamingValues() (numOfValues uint32, err error) {
	msg := &NumOfStreamingValuesMsg{}
	msg.rsp = &NumOfStreamingValuesRsp{}
	c.Send(msg)
	rsp := msg.Rsp().(*NumOfStreamingValuesRsp)
	numOfValues = rsp.NumOfValues
	err = rsp.Status()
	return
}
func (c Connection) QueryOutputEdgeDetect() (state int16, err error) {
	msg := &QueryOutputEdgeDetectMsg{}
	msg.rsp = &QueryOutputEdgeDetectRsp{}
	c.Send(msg)
	rsp := msg.Rsp().(*QueryOutputEdgeDetectRsp)
	state = rsp.State
	err = rsp.Status()
	return
}
func (c Connection) RunBlock(numOfPreTriggerSamples, numOfPostTriggerSamples int32,
	timeBase uint32, overSample int16, segmentIndex uint32, lpBlockReadyGoPar BlockReady,
	param any) (timeIndisposedMs int32, err error) {
	msg := &RunBlockMsg{NumOfPreTriggerSamples: numOfPreTriggerSamples, NumOfPostTriggerSamples: numOfPostTriggerSamples,
		TimeBase: timeBase, OverSample: overSample, SegmentIndex: segmentIndex, LpBlockReadyGoPar: lpBlockReadyGoPar, Param: param}
	msg.rsp = &RunBlockRsp{}
	c.Send(msg)
	rsp := msg.Rsp().(*RunBlockRsp)
	timeIndisposedMs = rsp.TimeIndisposedMs
	err = rsp.Status()
	return
}
func (c Connection) RunStreaming(reqSampleInterval uint32, sampleIntervalTimeUnits TimeUnits,
	maxPreTriggerSamples, maxPostTriggerSamples uint32,
	autoStop bool, downSampleRatio uint32, downSampleRatioMode RatioMode,
	overviewBufferSize uint32) (sampleInterval uint32, err error) {
	msg := &RunStreamingMsg{ReqSampleInterval: reqSampleInterval, SampleIntervalTimeUnits: sampleIntervalTimeUnits,
		MaxPreTriggerSamples: maxPreTriggerSamples, MaxPostTriggerSamples: maxPostTriggerSamples,
		AutoStop: autoStop, DownSampleRatio: downSampleRatio, DownSampleRatioMode: downSampleRatioMode,
		OverviewBufferSize: overviewBufferSize}
	msg.rsp = &RunStreamingRsp{}
	c.Send(msg)
	rsp := msg.Rsp().(*RunStreamingRsp)
	sampleInterval = rsp.SampleInterval
	err = rsp.Status()
	return
}

func (c Connection) SetChannel(channel ChannelId, enabled bool, couplingType Coupling,
	voltageRange RangeEnum, analogOffset float32) (err error) {
	msg := &SetChannelMsg{Channel: channel, Enabled: enabled, CouplingType: couplingType,
		VoltageRange: voltageRange, AnalogOffset: analogOffset}
	msg.rsp = &SetChannelRsp{}
	c.Send(msg)
	rsp := msg.Rsp().(*SetChannelRsp)
	err = rsp.Status()
	return
}
func (c Connection) SetDataBuffer(ch ChannelId, bufferIn []int16, segmentIndex uint32,
	mode RatioMode) (err error) {
	msg := &SetDataBufferMsg{Ch: ch, BufferIn: bufferIn, SegmentIndex: segmentIndex, Mode: mode}
	msg.rsp = &SetDataBufferRsp{}
	c.Send(msg)
	rsp := msg.Rsp().(*SetDataBufferRsp)
	err = rsp.Status()
	return
}
func (c Connection) SetDataBuffers(ch ChannelId, bufferMax, bufferMin []int16, segmentIndex uint32, mode RatioMode) (err error) {
	msg := &SetDataBuffersMsg{Ch: ch, BufferMax: bufferMax, BufferMin: bufferMin, SegmentIndex: segmentIndex, Mode: mode}
	msg.rsp = &SetDataBuffersRsp{}
	c.Send(msg)
	rsp := msg.Rsp().(*SetDataBuffersRsp)
	err = rsp.Status()
	return
}
func (c Connection) SetUnscaledDataBuffers(ch ChannelId, bufferMax, bufferMin []int16, segmentIndex uint32, mode RatioMode) (err error) {
	msg := &SetUnscaledDataBuffersMsg{Ch: ch, BufferMax: bufferMax, BufferMin: bufferMin, SegmentIndex: segmentIndex, Mode: mode}
	msg.rsp = &SetUnscaledataBuffersRsp{}
	c.Send(msg)
	rsp := msg.Rsp().(*SetUnscaledataBuffersRsp)
	err = rsp.Status()
	return
}
func (c Connection) SetDigitalAnalogTriggerOperand(operand TriggerOperand) (err error) {
	msg := &SetDigitalAnalogTriggerOperandMsg{Operand: operand}
	msg.rsp = &SetDigitalAnalogTriggerOperandRsp{}
	c.Send(msg)
	rsp := msg.Rsp().(*SetDigitalAnalogTriggerOperandRsp)
	err = rsp.Status()
	return
}
func (c Connection) SetDigitalPort(port DigitalPort, enabled bool, logiclevel int16) (err error) {
	msg := &SetDigitalPortMsg{Port: port, Enabled: enabled, Logiclevel: logiclevel}
	msg.rsp = &SetDigitalPortRsp{}
	c.Send(msg)
	rsp := msg.Rsp().(*SetDigitalPortRsp)
	err = rsp.Status()
	return
}
func (c Connection) SetEts(mode EtsMode, etsCycles int16, etsInterLeave int16) (sampleTimePicoseconds int32, err error) {
	msg := &SetEtsMsg{Mode: mode, EtsCycles: etsCycles, EtsInterleave: etsInterLeave}
	msg.rsp = &SetEtsRsp{}
	c.Send(msg)
	rsp := msg.Rsp().(*SetEtsRsp)
	sampleTimePicoseconds = rsp.SampleTimePicoseconds
	err = rsp.Status()
	return
}
func (c Connection) SetEtsTimeBuffer(buffer []int64) (err error) {
	msg := &SetEtsTimeBufferMsg{Buffer: buffer}
	msg.rsp = &SetEtsTimeBufferRsp{}
	c.Send(msg)
	rsp := msg.Rsp().(*SetEtsTimeBufferRsp)
	err = rsp.Status()
	return
}
func (c Connection) SetEtsTimeBuffers(timeUpper, timeLower []uint32) (err error) {
	msg := &SetEtsTimeBuffersMsg{TimeUpper: timeUpper, TimeLower: timeLower}
	msg.rsp = &SetEtsTimeBuffersRsp{}
	c.Send(msg)
	rsp := msg.Rsp().(*SetEtsTimeBuffersRsp)
	err = rsp.Status()
	return
}
func (c Connection) SetNoCaptures(nCaptures uint32) (err error) {
	msg := &SetNumOfCapturesMsg{NCaptures: nCaptures}
	msg.rsp = &SetNumOfCapturesRsp{}
	c.Send(msg)
	rsp := msg.Rsp().(*SetNumOfCapturesRsp)
	err = rsp.Status()
	return
}
func (c Connection) SetOutputEdgeDetect(state int16) (err error) {
	msg := &SetOutputEdgeDetectMsg{State: state}
	msg.rsp = &SetOutputEdgeDetectRsp{}
	c.Send(msg)
	rsp := msg.Rsp().(*SetOutputEdgeDetectRsp)
	err = rsp.Status()
	return
}
func (c Connection) SetPulseWidthDigitalPortProperties(digitalDirections []DigitalChannelDirections) (err error) {
	msg := &SetPulseWidthDigitalPortPropertiesMsg{DigitalDirections: digitalDirections}
	msg.rsp = &SetPulseWidthDigitalPortPropertiesRsp{}
	c.Send(msg)
	rsp := msg.Rsp().(*SetPulseWidthDigitalPortPropertiesRsp)
	err = rsp.Status()
	return
}
func (c Connection) SetPulseWidthQualifier(conditions []PwqConditions, direction ThresholdDirection, lower, upper uint32,
	pwType PulseWidthType) (err error) {
	msg := &SetPulseWidthQualifierMsg{Conditions: conditions, Direction: direction, Lower: lower, Upper: upper, PwType: pwType}
	msg.rsp = &SetPulseWidthQualifierRsp{}
	c.Send(msg)
	rsp := msg.Rsp().(*SetPulseWidthQualifierRsp)
	err = rsp.Status()
	return
}
func (c Connection) SetSigGenArbitrary(offsetVoltage int32, pkToPK uint32,
	startDeltaPhase, stopDeltaPhase, deltaPhaseIncrement, dwellCount uint32,
	arbitraryWaveform []int16, sweepType SweepTypeEnum, operation ExtraOperations,
	indexMode IndexMode, shots, sweeps uint32, triggerType SigGenTrigType,
	triggerSource SigGenTrigSource, extInThreshold int16) (err error) {
	msg := &SetSigGenArbitraryMsg{OffsetVoltage: offsetVoltage, PkToPK: pkToPK,
		StartDeltaPhase: startDeltaPhase, StopDeltaPhase: stopDeltaPhase, DeltaPhaseIncrement: deltaPhaseIncrement,
		DwellCount: dwellCount, ArbitraryWaveform: arbitraryWaveform, SweepType: sweepType,
		Operation: operation, IndexMode: indexMode, Shots: shots, Sweeps: sweeps,
		TtriggerType: triggerType, TriggerSource: triggerSource, ExtInThreshold: extInThreshold}
	msg.rsp = &SetSigGenArbitraryRsp{}
	c.Send(msg)
	rsp := msg.Rsp().(*SetSigGenArbitraryRsp)
	err = rsp.Status()
	return
}
func (c Connection) SigGenArbitraryMinMaxValues() (minArbitraryWaveformValue, maxArbitraryWaveformValue int16,
	minArbitraryWaveformSize, maxArbitraryWaveformSize uint32, err error) {
	msg := &SigGenArbitraryMinMaxValuesMsg{}
	msg.rsp = &SigGenArbitraryMinMaxValuesRsp{}
	c.Send(msg)
	rsp := msg.Rsp().(*SigGenArbitraryMinMaxValuesRsp)
	minArbitraryWaveformValue = rsp.MinArbitraryWaveformValue
	maxArbitraryWaveformValue = rsp.MaxArbitraryWaveformValue
	minArbitraryWaveformSize = rsp.MinArbitraryWaveformSize
	maxArbitraryWaveformSize = rsp.MaxArbitraryWaveformSize
	err = rsp.Status()
	return
}
func (c Connection) SetSigGenBuiltIn(offsetVoltage int32, pkToPK uint32, waveType WaveTypeEnum,
	startFrequency, stopFrequency, increment, dwellTime float32, sweepType SweepTypeEnum,
	operation ExtraOperations, shots, sweeps uint32, triggerType SigGenTrigType,
	triggerSource SigGenTrigSource, extInThreshold int16) (err error) {
	msg := &SetSigGenBuiltInMsg{OffsetVoltage: offsetVoltage, PkToPK: pkToPK, WaveType: waveType,
		StartFrequency: startFrequency, StopFrequency: stopFrequency, Increment: increment, DwellTime: dwellTime,
		SweepType: sweepType, Operation: operation, Shots: shots, Sweeps: sweeps, TriggerType: triggerType,
		TriggerSource: triggerSource, ExtInThreshold: extInThreshold}
	msg.rsp = &SetSigGenBuiltInRsp{}
	c.Send(msg)
	rsp := msg.Rsp().(*SetSigGenBuiltInRsp)
	err = rsp.Status()
	return
}
func (c Connection) SetSigGenBuiltInV2(offsetVoltage int32, pkToPK uint32, waveType WaveTypeEnum,
	startFrequency, stopFrequency, increment, dwellTime float64, sweepType SweepTypeEnum,
	operation ExtraOperations, shots, sweeps uint32, triggerType SigGenTrigType,
	triggerSource SigGenTrigSource, extInThreshold int16) (err error) {
	msg := &SetSigGenBuiltInV2Msg{OffsetVoltage: offsetVoltage, PkToPK: pkToPK, WaveType: waveType,
		StartFrequency: startFrequency, StopFrequency: stopFrequency, Increment: increment, DwellTime: dwellTime,
		SweepType: sweepType, Operation: operation, Shots: shots, Sweeps: sweeps, TriggerType: triggerType,
		TriggerSource: triggerSource, ExtInThreshold: extInThreshold}
	msg.rsp = &SetSigGenBuiltInV2Rsp{}
	c.Send(msg)
	rsp := msg.Rsp().(*SetSigGenBuiltInV2Rsp)
	err = rsp.Status()
	return
}
func (c Connection) SetSimGen(channel ChannelId, on bool, offsetVoltage int32, pkToPK uint32, waveType WaveTypeEnum,
	startFrequency, stopFrequency, increment, dwellTime float64, sweepType SweepTypeEnum,
	operation ExtraOperations, shots, sweeps uint32, triggerType SigGenTrigType,
	triggerSource SigGenTrigSource, extInThreshold int16, phase float64) (err error) {
	msg := &SetSimGenMsg{Channel: channel, On: on, OffsetVoltage: offsetVoltage, PkToPK: pkToPK, WaveType: waveType,
		StartFrequency: startFrequency, StopFrequency: stopFrequency, Increment: increment, DwellTime: dwellTime,
		SweepType: sweepType, Operation: operation, Shots: shots, Sweeps: sweeps, TriggerType: triggerType,
		TriggerSource: triggerSource, ExtInThreshold: extInThreshold, Phase: phase}
	msg.rsp = &SetSimGenRsp{}
	c.Send(msg)
	rsp := msg.Rsp().(*SetSimGenRsp)
	err = rsp.Status()
	return
}

func (c Connection) SetSimRlcFilter(channel ChannelId, genSource ChannelId, enabled bool, filterType string, r float64, runit string, l float64, lunit string, cval float64, cunit string) (err error) {
	msg := &SetSimRlcFilterMsg{Channel: channel, GenSource: genSource, Enabled: enabled, FilterType: filterType, R: r, RUnit: runit, L: l, LUnit: lunit, C: cval, CUnit: cunit}
	msg.rsp = &SetSimRlcFilterRsp{}
	c.Send(msg)
	rsp := msg.Rsp().(*SetSimRlcFilterRsp)
	err = rsp.Status()
	return
}

func (c Connection) SetSimDigitalFilter(channel ChannelId, lpEnabled bool, lpFc float64, hpEnabled bool, hpFc float64, bpEnabled bool, bpFc1, bpFc2 float64, bsEnabled bool, bsFc1, bsFc2 float64) (err error) {
	msg := &SetSimDigitalFilterMsg{
		Channel:         channel,
		LowpassEnabled:  lpEnabled,
		LowpassFc:       lpFc,
		HighpassEnabled: hpEnabled,
		HighpassFc:      hpFc,
		BandpassEnabled: bpEnabled,
		BandpassFc1:     bpFc1,
		BandpassFc2:     bpFc2,
		BandstopEnabled: bsEnabled,
		BandstopFc1:     bsFc1,
		BandstopFc2:     bsFc2,
	}
	msg.rsp = &SetSimDigitalFilterRsp{}
	c.Send(msg)
	rsp := msg.Rsp().(*SetSimDigitalFilterRsp)
	err = rsp.Status()
	return
}


func (c Connection) SetSigGenPropertiesArbitrary(offsetVoltage int32,
	startDeltaPhase, stopDeltaPhase, deltaPhaseIncrement, dwellCount uint32,
	sweepType SweepTypeEnum, operation ExtraOperations,
	indexMode IndexMode, shots, sweeps uint32, triggerType SigGenTrigType,
	triggerSource SigGenTrigSource, extInThreshold int16) (err error) {
	msg := &SetSigGenPropertiesArbitraryMsg{OffsetVoltage: offsetVoltage,
		StartDeltaPhase: startDeltaPhase, StopDeltaPhase: stopDeltaPhase, DeltaPhaseIncrement: deltaPhaseIncrement,
		DwellCount: dwellCount, SweepType: sweepType, Operation: operation, IndexMode: indexMode,
		Shots: shots, Sweeps: sweeps, TriggerType: triggerType, TriggerSource: triggerSource,
		ExtInThreshold: extInThreshold}
	msg.rsp = &SetSigGenPropertiesArbitraryRsp{}
	c.Send(msg)
	rsp := msg.Rsp().(*SetSigGenPropertiesArbitraryRsp)
	err = rsp.Status()
	return
}
func (c Connection) SetSigGenPropertiesBuiltIn(offsetVoltage int32,
	startFrequency, stopFrequency, increment, dwellTime float64,
	sweepType SweepTypeEnum,
	shots, sweeps uint32, triggerType SigGenTrigType,
	triggerSource SigGenTrigSource, extInThreshold int16) (err error) {
	msg := &SetSigGenPropertiesBuiltInMsg{OffsetVoltage: offsetVoltage,
		StartFrequency: startFrequency, StopFrequency: stopFrequency, Increment: increment, DwellTime: dwellTime,
		SweepType: sweepType, Shots: shots, Sweeps: sweeps, TriggerType: triggerType,
		TriggerSource: triggerSource, ExtInThreshold: extInThreshold}
	msg.rsp = &SetSigGenPropertiesBuiltInRsp{}
	c.Send(msg)
	rsp := msg.Rsp().(*SetSigGenPropertiesBuiltInRsp)
	err = rsp.Status()
	return
}
func (c Connection) SetSimpleTrigger(enable bool, source ChannelId, threshold int16,
	direction ThresholdDirection, delay uint32, autoTriggerMs int16) (err error) {
	msg := &SetSimpleTriggerMsg{Enable: enable, Source: source, Threshold: threshold,
		Direction: direction, Delay: delay, AutoTriggerMs: autoTriggerMs}
	msg.rsp = &SetSimpleTriggerRsp{}
	c.Send(msg)
	rsp := msg.Rsp().(*SetSimpleTriggerRsp)
	err = rsp.Status()
	return
}
func (c Connection) SetTriggerChannelConditions(triggerConditions []TriggerConditions) (err error) {
	msg := &SetTriggerChannelConditionsMsg{TriggerConditions: triggerConditions}
	msg.rsp = &SetTriggerChannelConditionsRsp{}
	c.Send(msg)
	rsp := msg.Rsp().(*SetTriggerChannelConditionsRsp)
	err = rsp.Status()
	return
}
func (c Connection) SetTriggerChannelDirections(channelA, channelB, channelC, channelD, ext, aux ThresholdDirection) (err error) {
	msg := &SetTriggerChannelDirectionsMsg{ChannelA: channelA, ChannelB: channelB,
		ChannelC: channelC, ChannelD: channelD, Ext: ext, Aux: aux}
	msg.rsp = &SetTriggerChannelDirectionsRsp{}
	c.Send(msg)
	rsp := msg.Rsp().(*SetTriggerChannelDirectionsRsp)
	err = rsp.Status()
	return
}
func (c Connection) SetTriggerChannelProperties(channelProperties []TriggerChannelProperties, auxOutputEnable bool,
	autoTriggerMs int32) (err error) {
	msg := &SetTriggerChannelPropertiesMsg{ChannelProperties: channelProperties, AuxOutputEnable: auxOutputEnable, AutoTriggerMs: autoTriggerMs}
	msg.rsp = &SetTriggerChannelPropertiesRsp{}
	c.Send(msg)
	rsp := msg.Rsp().(*SetTriggerChannelPropertiesRsp)
	err = rsp.Status()
	return
}
func (c Connection) SetTriggerDelay(delay uint32) (err error) {
	msg := &SetTriggerDelayMsg{Delay: delay}
	msg.rsp = &SetTriggerDelayRsp{}
	c.Send(msg)
	rsp := msg.Rsp().(*SetTriggerDelayRsp)
	err = rsp.Status()
	return
}
func (c Connection) SetTriggerDigitalPortProperties(digitalDirections []DigitalChannelDirections) (err error) {
	msg := &SetTriggerDigitalPortPropertiesMsg{DigitalDirections: digitalDirections}
	msg.rsp = &SetTriggerDigitalPortPropertiesRsp{}
	c.Send(msg)
	rsp := msg.Rsp().(*SetTriggerDigitalPortPropertiesRsp)
	err = rsp.Status()
	return
}
func (c Connection) SigGenFrequencyToPhase(frequency float64, indexMode IndexMode, bufferLength uint32) (phase uint32, err error) {
	msg := &SigGenFrequencyToPhasenMsg{Frequency: frequency, IndexMode: indexMode, BufferLength: bufferLength}
	msg.rsp = &SigGenFrequencyToPhaseRsp{}
	c.Send(msg)
	rsp := msg.Rsp().(*SigGenFrequencyToPhaseRsp)
	phase = rsp.Phase
	err = rsp.Status()
	return
}
func (c Connection) Stop() (err error) {
	msg := &StopMsg{}
	msg.rsp = &StopRsp{}
	c.Send(msg)
	rsp := msg.Rsp().(*StopRsp)
	err = rsp.Status()
	return
}
func (c Connection) TriggerOrPulseWidthQualifierEnabled() (triggerEnabled, pulseWidthQualifierEnabledint16 int16, err error) {
	msg := &TriggerOrPulseWidthQualifierEnabledMsg{}
	msg.rsp = &TriggerOrPulseWidthQualifierEnabledRsp{}
	c.Send(msg)
	rsp := msg.Rsp().(*TriggerOrPulseWidthQualifierEnabledRsp)
	triggerEnabled = rsp.TriggerEnabled
	pulseWidthQualifierEnabledint16 = rsp.PulseWidthQualifierEnabledint16
	err = rsp.Status()
	return
}
func (c Connection) SigGenSoftwareControl(state int16) (err error) {
	msg := &SigGenSoftwareControlMsg{State: state}
	msg.rsp = &SigGenSoftwareControlRsp{}
	c.Send(msg)
	rsp := msg.Rsp().(*SigGenSoftwareControlRsp)
	err = rsp.Status()
	return
}

func (m *MsgBase) Status() (err error) {
	err = m.rsp.Status()
	return
}

func (m *MsgBase) SetStatus(err error) {
	m.rsp.SetStatus(err)
}

func (m *MsgBase) Handle() (handle int16) {
	handle = m.handle
	return
}

func (m *MsgBase) SetHandle(handle int16) {
	m.handle = handle
}

func (m *MsgBase) SetRspCh(rspCh chan struct{}) {
	m.rspCh = rspCh
}

func (m *MsgBase) RspCh() (rspCh chan struct{}) {
	rspCh = m.rspCh
	return
}
func (m *MsgBase) Rsp() (rsp Response) {
	rsp = m.rsp
	return
}
func (r *RespBase) SetStatus(err error) {
	r.status = err
}
func (r *RespBase) Status() (status error) {
	status = r.status
	return
}

func TimeUnitToVal(tu TimeUnits) float64 {
	switch {
	case tu == TuFs:
		return 1e-15
	case tu == TuPs:
		return 1e-12
	case tu == TuNs:
		return 1e-9
	case tu == TuUs:
		return 1e-6
	case tu == TuMs:
		return 1e-3
	case tu == TuS:
		return 1
	}
	return 0
}

// type (
// 	InputRangesFunc = func(r RangeEnum) int32
// )

// var InputRanges InputRangesFunc
