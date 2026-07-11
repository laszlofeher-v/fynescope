package sim

import (
	"fynescope/genericps"
	"log/slog"
	"time"
)

const (
	notImplemented                     = "Not implemented"
	maxValue                           = 32515
	mega                               = 1024 * 1024
	maxTriggerTest                     = 10000000
	callDelayMs                        = 100
	MinChannels                        = 1
	MaxChannels                        = 4
	DefaultChannels                    = 4
	ChannelInfoRanges                  = 0
	DualScope         NumOfChannelEnum = 2
	QuadScope         NumOfChannelEnum = 4
)

var (
	numberOfChannels       = 4
	scopeVariantInfo       = "2407SIM"
	scopeBathAndSerialInfo = "1/1"
	MaxChannelBuffers      = 8
)

var (
	timeout = 8 * time.Second
)

const (
	normal returnStatus = iota
	faulty
	timeoutNormal
	timeoutfaulty
)

type NumOfChannelEnum int

type RangeEnum int

const (
	Range_10mv RangeEnum = iota
	Range_20mv
	Range_50mv
	Range_100mv
	Range_200mv
	Range_500mv
	Range_1v
	Range_2v
	Range_5v
	Range_10v
	Range_20v
	Range_50v
)

var (
	RangeValuesMv = map[RangeEnum]float64{
		Range_10mv:  10.0,
		Range_20mv:  20.0,
		Range_50mv:  50.0,
		Range_100mv: 100.0,
		Range_200mv: 200.0,
		Range_500mv: 500.0,
		Range_1v:    1000.0,
		Range_2v:    2000.0,
		Range_5v:    5000.0,
		Range_10v:   10000.0,
		Range_20v:   20000.0,
		Range_50v:   50000.0,
	}
)

type ChannelId int

const (
	ChA ChannelId = iota
	ChB
	ChC
	ChD
)

type ThresholdModeId int

const (
	Level ThresholdModeId = iota
	Window
)

type HoldOffType int

const (
	HofTime HoldOffType = iota
	MaxHoldOffTime
)

type Coupling int

const (
	Ac Coupling = iota
	Dc
)

type ThresholdDirection int

const (
	TriggerAbove ThresholdDirection = iota
	TriggerBelow
	TriggerRising
	TriggerFalling
	TriggerRisingOrFalling
	TriggerAboveLower
	TriggerBelowLower
	TriggerRisingLower
	TriggerFallingLower
	TriggerOutside
	TriggerInside
	TriggerEnter
	TriggerExit
	TriggerEnterOrExit
	TriggerPositiveRunt
	TriggerNegativeRunt
	TriggerNone
)

type PicoInfo int

const (
	PicoDriverVersion PicoInfo = iota
	PicoUsbVersion
	PicoHardwareVersion
	PicoVariantInfo
	PicoBatchAndSerial
	PicoCalDate
	PicoKernelVersion
	PicoDigitalHardwareVersion
	PicoAnalogueHardwareVersion
	PicoFirmwareVersion1
	PicoFirmwareVersion2
	PicoMacAddress
	PicoShadowCall
	PicoIppVersion
	PicoDriverPath
	PicoFirmwareVersion3
	PicoFrontPanelFirmwareVersion3
	PicoBootloaderVersion
)

type TimeUnits int

const (
	TuFs TimeUnits = iota
	TuPs
	TuNs
	TuUs
	TuMs
	TuS
)

type RatioMode int

const (
	RatioModeNone RatioMode = iota
	RatioModeAggregate
	RatioModeDecimate
	RatioModeAverage
)

const (
	AwgMinSigGenBufferSize = 1
	AwgMaxSigGenBufferSize = 8192
	MinThresholdDiff       = 100
)

type TriggerState int

const (
	CondDontCare TriggerState = iota
	CondTrue
	CondFalse
	CondMax
)

type PulseWidthType int

const (
	PwTypeNone PulseWidthType = iota
	PwTypeLessThan
	PwTypeGreaterThan
	PwTypeInRange
	PwTypeOutOfRange
)

type EtsMode int

const (
	EtsOff EtsMode = iota
	EtsFast
	EtsSlow
	EtsMax
)
const (
	Ps2207MaxEtsCyscles       = 500
	Ps2207MaxEtsMaxInterleave = 40
)

type DigitalChannel int

const (
	Dch0 DigitalChannel = iota
	Dch1
	Dch2
	Dch3
	Dch4
	Dch5
	Dch6
	Dch7
	Dch8
	Dch9
	Dch10
	Dch11
	Dch12
	Dch13
	Dch14
	Dch15
	Dch16
	Dch17
	Dch18
	Dch19
	Dch20
	Dch21
	Dch22
	Dch23
	Dch24
	Dch25
	Dch26
	Dch27
	Dch28
	Dch29
	Dch30
	Dch31
	DchMax
)

type DigitalDirection int

const (
	DigitalDontCare DigitalDirection = iota
	DigitalDirectionLow
	DigitalDirectionHigh
	DigitalDirectionRising
	DigitalDirectionFalling
	DigitalDirectionRisingOrFalling
	DigitalMaxDirection
)

type WaveTypeEnum int16

const (
	Sine WaveTypeEnum = iota
	Square
	Triangle
	RampUp
	RampDown
	SinC
	Gaussian
	HalfSine
	DcVoltage
)

type SweepTypeEnum int

const (
	SweepUp SweepTypeEnum = iota
	SweepDown
	SweepUpDown
	SweepDownUp
	SweepMaxTypes
)

type ExtraOperations int

const (
	EsOff ExtraOperations = iota
	WhiteNoise
	Prbs
)

const (
	// 1 GHz
	SineMaxFrequency     = 100000000
	SquareMaxFrequency   = 100000000
	TriangleMaxFrequency = 100000000
	SinCMaxFrequency     = 100000000
	RampMaxFrequency     = 100000000
	HalfSineMaxFrequency = 100000000
	GaussianMaxFrequency = 100000000
	PrbsMaxFrequency     = 100000000
	PrbsMinFrequency     = 100000000
	MinFrequency         = 0
)

const (
	MaxSweepShots              = (1 << 30) - 1
	ShotSweepTriggerContinuous = 0xFFFFFFFF
)

type SigGenTrigType int

const (
	SigGenRising SigGenTrigType = iota
	SigGenFalling
	SigGenGateHigh
	SigGenGateLow
)

type SigGenTrigSource int

const (
	SigGenNone SigGenTrigSource = iota
	SigGenScopeTrig
	SigGenAuxIn
	SigGenExtIn
	SigGenSoftTrig
)

type IndexMode int

const (
	Single IndexMode = iota
	Dual
	Quad
	MaxIndexModes
)

type TriggerOperand int

const (
	OperandNone TriggerOperand = iota
	OperandOr
	OperandAnd
	OperandThen
)

type DigitalPort int

const (
	Port0 DigitalPort = iota // digital channel 0 - 7
	Port1                    // digital channel 8 - 15
	Port2                    // digital channel 16 - 23
	Port3                    // digital channel 24 - 31
	MaxDigitalPorts
)

var (
	inputRanges []int32 = []int32{
		10,
		20,
		50,
		100,
		200,
		500,
		1000,
		2000,
		5000,
		10000,
		20000,
		50000,
	}
)

func InputRanges(r RangeEnum) int32 {
	idx := int(r)
	if idx < 0 {
		slog.Error("InputRanges: negative range enum", "r", r)
		idx = 0
	} else if idx >= len(inputRanges) {
		slog.Error("InputRanges: range enum out of bounds", "r", r, "max", len(inputRanges)-1)
		idx = len(inputRanges) - 1
	}
	return inputRanges[idx]
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
func loadConstants() {
	genericps.ChA = genericps.ChannelId(ChA)
	genericps.ChB = genericps.ChannelId(ChB)
	genericps.ChC = genericps.ChannelId(ChC)
	genericps.ChD = genericps.ChannelId(ChD)
	genericps.RatioModeNone = genericps.RatioMode(RatioModeNone)
	genericps.RatioModeAggregate = genericps.RatioMode(RatioModeAggregate)
	genericps.RatioModeDecimate = genericps.RatioMode(RatioModeDecimate)
	genericps.RatioModeAverage = genericps.RatioMode(RatioModeAverage)
	genericps.Ac = genericps.Coupling(Ac)
	genericps.Dc = genericps.Coupling(Dc)
	genericps.Range_10mv = genericps.RangeEnum(Range_10mv)
	genericps.Range_20mv = genericps.RangeEnum(Range_20mv)
	genericps.Range_50mv = genericps.RangeEnum(Range_50mv)
	genericps.Range_100mv = genericps.RangeEnum(Range_100mv)
	genericps.Range_200mv = genericps.RangeEnum(Range_200mv)
	genericps.Range_500mv = genericps.RangeEnum(Range_500mv)
	genericps.Range_1v = genericps.RangeEnum(Range_1v)
	genericps.Range_2v = genericps.RangeEnum(Range_2v)
	genericps.Range_5v = genericps.RangeEnum(Range_5v)
	genericps.Range_10v = genericps.RangeEnum(Range_10v)
	genericps.Range_20v = genericps.RangeEnum(Range_20v)
	genericps.Range_50v = genericps.RangeEnum(Range_50v)
	genericps.Level = genericps.ThresholdModeId(Level)
	genericps.Window = genericps.ThresholdModeId(Window)
	genericps.CondDontCare = genericps.TriggerRespBase(CondDontCare)
	genericps.CondTrue = genericps.TriggerRespBase(CondTrue)
	genericps.CondFalse = genericps.TriggerRespBase(CondFalse)
	genericps.CondMax = genericps.TriggerRespBase(CondMax)
	genericps.TriggerAbove = genericps.ThresholdDirection(TriggerAbove)
	genericps.TriggerBelow = genericps.ThresholdDirection(TriggerBelow)
	genericps.TriggerRising = genericps.ThresholdDirection(TriggerRising)
	genericps.TriggerFalling = genericps.ThresholdDirection(TriggerFalling)
	genericps.TriggerRisingOrFalling = genericps.ThresholdDirection(TriggerRisingOrFalling)
	genericps.TriggerAboveLower = genericps.ThresholdDirection(TriggerAboveLower)
	genericps.TriggerBelowLower = genericps.ThresholdDirection(TriggerBelowLower)
	genericps.TriggerRisingLower = genericps.ThresholdDirection(TriggerRisingLower)
	genericps.TriggerFallingLower = genericps.ThresholdDirection(TriggerFallingLower)
	genericps.TriggerOutside = genericps.ThresholdDirection(TriggerOutside)
	genericps.TriggerInside = genericps.ThresholdDirection(TriggerInside)
	genericps.TriggerEnter = genericps.ThresholdDirection(TriggerEnter)
	genericps.TriggerExit = genericps.ThresholdDirection(TriggerExit)
	genericps.TriggerEnterOrExit = genericps.ThresholdDirection(TriggerEnterOrExit)
	genericps.TriggerPositiveRunt = genericps.ThresholdDirection(TriggerPositiveRunt)
	genericps.TriggerNegativeRunt = genericps.ThresholdDirection(TriggerNegativeRunt)
	genericps.TriggerNone = genericps.ThresholdDirection(TriggerNone)
	genericps.Dch0 = genericps.DigitalChannel(Dch0)
	genericps.Dch1 = genericps.DigitalChannel(Dch1)
	genericps.Dch2 = genericps.DigitalChannel(Dch2)
	genericps.Dch3 = genericps.DigitalChannel(Dch3)
	genericps.Dch4 = genericps.DigitalChannel(Dch4)
	genericps.Dch5 = genericps.DigitalChannel(Dch5)
	genericps.Dch6 = genericps.DigitalChannel(Dch6)
	genericps.Dch7 = genericps.DigitalChannel(Dch7)
	genericps.Dch8 = genericps.DigitalChannel(Dch8)
	genericps.Dch9 = genericps.DigitalChannel(Dch9)
	genericps.Dch10 = genericps.DigitalChannel(Dch10)
	genericps.Dch11 = genericps.DigitalChannel(Dch11)
	genericps.Dch12 = genericps.DigitalChannel(Dch12)
	genericps.Dch13 = genericps.DigitalChannel(Dch13)
	genericps.Dch14 = genericps.DigitalChannel(Dch14)
	genericps.Dch15 = genericps.DigitalChannel(Dch15)
	genericps.Dch16 = genericps.DigitalChannel(Dch16)
	genericps.Dch17 = genericps.DigitalChannel(Dch17)
	genericps.Dch18 = genericps.DigitalChannel(Dch18)
	genericps.Dch19 = genericps.DigitalChannel(Dch19)
	genericps.Dch20 = genericps.DigitalChannel(Dch20)
	genericps.Dch21 = genericps.DigitalChannel(Dch21)
	genericps.Dch22 = genericps.DigitalChannel(Dch22)
	genericps.Dch23 = genericps.DigitalChannel(Dch23)
	genericps.Dch24 = genericps.DigitalChannel(Dch24)
	genericps.Dch25 = genericps.DigitalChannel(Dch25)
	genericps.Dch26 = genericps.DigitalChannel(Dch26)
	genericps.Dch27 = genericps.DigitalChannel(Dch27)
	genericps.Dch28 = genericps.DigitalChannel(Dch28)
	genericps.Dch29 = genericps.DigitalChannel(Dch29)
	genericps.Dch30 = genericps.DigitalChannel(Dch30)
	genericps.Dch31 = genericps.DigitalChannel(Dch31)
	genericps.DchMax = genericps.DigitalChannel(DchMax)
	genericps.DigitalDontCare = genericps.DigitalDirection(DigitalDontCare)
	genericps.DigitalDirectionLow = genericps.DigitalDirection(DigitalDirectionLow)
	genericps.DigitalDirectionHigh = genericps.DigitalDirection(DigitalDirectionHigh)
	genericps.DigitalDirectionRising = genericps.DigitalDirection(DigitalDirectionRising)
	genericps.DigitalDirectionFalling = genericps.DigitalDirection(DigitalDirectionFalling)
	genericps.DigitalDirectionRisingOrFalling = genericps.DigitalDirection(DigitalDirectionRisingOrFalling)
	genericps.DigitalMaxDirection = genericps.DigitalDirection(DigitalMaxDirection)
	genericps.SweepUp = genericps.SweepTypeEnum(SweepUp)
	genericps.SweepDown = genericps.SweepTypeEnum(SweepDown)
	genericps.SweepUpDown = genericps.SweepTypeEnum(SweepUpDown)
	genericps.SweepDownUp = genericps.SweepTypeEnum(SweepDownUp)
	genericps.SweepMaxTypes = genericps.SweepTypeEnum(SweepMaxTypes)
	genericps.EsOff = genericps.ExtraOperations(EsOff)
	genericps.WhiteNoise = genericps.ExtraOperations(WhiteNoise)
	genericps.Prbs = genericps.ExtraOperations(Prbs)
	genericps.Single = genericps.IndexMode(Single)
	genericps.Dual = genericps.IndexMode(Dual)
	genericps.Quad = genericps.IndexMode(Quad)
	genericps.MaxIndexModes = genericps.IndexMode(MaxIndexModes)
	genericps.SigGenRising = genericps.SigGenTrigType(SigGenRising)
	genericps.SigGenFalling = genericps.SigGenTrigType(SigGenFalling)
	genericps.SigGenGateHigh = genericps.SigGenTrigType(SigGenGateHigh)
	genericps.SigGenGateLow = genericps.SigGenTrigType(SigGenGateLow)
	genericps.SigGenNone = genericps.SigGenTrigSource(SigGenNone)
	genericps.SigGenScopeTrig = genericps.SigGenTrigSource(SigGenScopeTrig)
	genericps.SigGenAuxIn = genericps.SigGenTrigSource(SigGenAuxIn)
	genericps.SigGenExtIn = genericps.SigGenTrigSource(SigGenExtIn)
	genericps.SigGenSoftTrig = genericps.SigGenTrigSource(SigGenSoftTrig)
	genericps.TuFs = genericps.TimeUnits(TuFs)
	genericps.TuPs = genericps.TimeUnits(TuPs)
	genericps.TuNs = genericps.TimeUnits(TuNs)
	genericps.TuUs = genericps.TimeUnits(TuUs)
	genericps.TuMs = genericps.TimeUnits(TuMs)
	genericps.TuS = genericps.TimeUnits(TuS)
	genericps.HofTime = genericps.HoldOffType(HofTime)
	genericps.MaxHoldOffTime = genericps.HoldOffType(MaxHoldOffTime)
	genericps.OperandNone = genericps.TriggerOperand(OperandNone)
	genericps.OperandOr = genericps.TriggerOperand(OperandOr)
	genericps.OperandAnd = genericps.TriggerOperand(OperandAnd)
	genericps.OperandThen = genericps.TriggerOperand(OperandThen)
	genericps.Port0 = genericps.DigitalPort(Port0)
	genericps.Port1 = genericps.DigitalPort(Port1)
	genericps.Port2 = genericps.DigitalPort(Port2)
	genericps.Port3 = genericps.DigitalPort(Port3)
	genericps.MaxDigitalPorts = genericps.DigitalPort(MaxDigitalPorts)
	genericps.PicoDriverVersion = genericps.PicoInfo(PicoDriverVersion)
	genericps.PicoUsbVersion = genericps.PicoInfo(PicoUsbVersion)
	genericps.PicoHardwareVersion = genericps.PicoInfo(PicoHardwareVersion)
	genericps.PicoVariantInfo = genericps.PicoInfo(PicoVariantInfo)
	genericps.PicoBatchAndSerial = genericps.PicoInfo(PicoBatchAndSerial)
	genericps.PicoCalDate = genericps.PicoInfo(PicoCalDate)
	genericps.PicoKernelVersion = genericps.PicoInfo(PicoKernelVersion)
	genericps.PicoDigitalHardwareVersion = genericps.PicoInfo(PicoDigitalHardwareVersion)
	genericps.PicoAnalogueHardwareVersion = genericps.PicoInfo(PicoAnalogueHardwareVersion)
	genericps.PicoFirmwareVersion1 = genericps.PicoInfo(PicoFirmwareVersion1)
	genericps.PicoFirmwareVersion2 = genericps.PicoInfo(PicoFirmwareVersion2)
	genericps.PicoMacAddress = genericps.PicoInfo(PicoMacAddress)
	genericps.PicoShadowCall = genericps.PicoInfo(PicoShadowCall)
	genericps.PicoIppVersion = genericps.PicoInfo(PicoIppVersion)
	genericps.PicoDriverPath = genericps.PicoInfo(PicoDriverPath)
	genericps.PicoFirmwareVersion3 = genericps.PicoInfo(PicoFirmwareVersion3)
	genericps.PicoFrontPanelFirmwareVersion3 = genericps.PicoInfo(PicoFrontPanelFirmwareVersion3)
	genericps.PicoBootloaderVersion = genericps.PicoInfo(PicoBootloaderVersion)
	genericps.EtsOff = genericps.EtsMode(EtsOff)
	genericps.EtsFast = genericps.EtsMode(EtsFast)
	genericps.EtsSlow = genericps.EtsMode(EtsSlow)
	genericps.EtsMax = genericps.EtsMode(EtsMax)
	genericps.PwTypeNone = genericps.PulseWidthType(PwTypeNone)
	genericps.PwTypeLessThan = genericps.PulseWidthType(PwTypeLessThan)
	genericps.PwTypeGreaterThan = genericps.PulseWidthType(PwTypeGreaterThan)
	genericps.PwTypeInRange = genericps.PulseWidthType(PwTypeInRange)
	genericps.PwTypeOutOfRange = genericps.PulseWidthType(PwTypeOutOfRange)
	genericps.Sine = genericps.WaveTypeEnum(Sine)
	genericps.Square = genericps.WaveTypeEnum(Square)
	genericps.Triangle = genericps.WaveTypeEnum(Triangle)
	genericps.RampUp = genericps.WaveTypeEnum(RampUp)
	genericps.RampDown = genericps.WaveTypeEnum(RampDown)
	genericps.SinC = genericps.WaveTypeEnum(SinC)
	genericps.Gaussian = genericps.WaveTypeEnum(Gaussian)
	genericps.HalfSine = genericps.WaveTypeEnum(HalfSine)
	genericps.DcVoltage = genericps.WaveTypeEnum(DcVoltage)
	genericps.InputRanges = inputRanges
	genericps.ChannelInfoRanges = int16(ChannelInfoRanges)
	genericps.RangeValuesMv = make(map[genericps.RangeEnum]float64)
	for k, v := range RangeValuesMv {
		kg := genericps.RangeEnum(k)
		genericps.RangeValuesMv[kg] = v
	}
	genericps.SineMaxFrequency = (SineMaxFrequency)
	genericps.SquareMaxFrequency = (SquareMaxFrequency)
	genericps.TriangleMaxFrequency = (TriangleMaxFrequency)
	genericps.SinCMaxFrequency = (SinCMaxFrequency)
	genericps.RampMaxFrequency = (RampMaxFrequency)
	genericps.HalfSineMaxFrequency = (HalfSineMaxFrequency)
	genericps.GaussianMaxFrequency = (GaussianMaxFrequency)
	genericps.PrbsMaxFrequency = (PrbsMaxFrequency)
	genericps.PrbsMinFrequency = (PrbsMinFrequency)
	genericps.MinFrequency = (MinFrequency)
	genericps.AwgMinSigGenBufferSize = (AwgMinSigGenBufferSize)
	genericps.AwgMaxSigGenBufferSize = (AwgMaxSigGenBufferSize)
	genericps.MinThresholdDiff = MinThresholdDiff
}
