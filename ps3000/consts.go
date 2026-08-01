//go:build !demo && ps3000

package ps3000

// #cgo CFLAGS: -g -Wall
// #cgo LDFLAGS: -L/opt/picoscope/lib/ -lps3000
// #include <stdlib.h>
// #include "/opt/picoscope/include/libps3000/ps3000.h"
import "C"
import "fynescope/genericps"

const (
	MaxChannels       = C.PS3000_MAX_CHANNELS
	ChannelInfoRanges = 12 // manually set, since ps3000 doesn't have PS3000_CI_RANGES
)

type NumOfChannelEnum int

const (
	DualScope NumOfChannelEnum = 2
	QuadScope NumOfChannelEnum = 4
)

type RangeEnum int

const (
	Range_10mv  RangeEnum = C.PS3000_10MV
	Range_20mv  RangeEnum = C.PS3000_20MV
	Range_50mv  RangeEnum = C.PS3000_50MV
	Range_100mv RangeEnum = C.PS3000_100MV
	Range_200mv RangeEnum = C.PS3000_200MV
	Range_500mv RangeEnum = C.PS3000_500MV
	Range_1v    RangeEnum = C.PS3000_1V
	Range_2v    RangeEnum = C.PS3000_2V
	Range_5v    RangeEnum = C.PS3000_5V
	Range_10v   RangeEnum = C.PS3000_10V
	Range_20v   RangeEnum = C.PS3000_20V
	Range_50v   RangeEnum = C.PS3000_50V
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
	ChA ChannelId = C.PS3000_CHANNEL_A
	ChB ChannelId = C.PS3000_CHANNEL_B
	ChC ChannelId = C.PS3000_CHANNEL_C
	ChD ChannelId = C.PS3000_CHANNEL_D
)

// ps3000 doesn't have ThresholdModeId, so use dummy values
type ThresholdModeId int
const (
	Level  ThresholdModeId = 0
	Window ThresholdModeId = 1
)

type Coupling int
const (
	Ac Coupling = 0 // C.PS3000_AC // ps3000 actually uses int 0/1 for dc coupling (0 for AC, 1 for DC)
	Dc Coupling = 1 // C.PS3000_DC
)

type ThresholdDirection int

const (
	TriggerAbove           ThresholdDirection = C.ABOVE
	TriggerBelow           ThresholdDirection = C.BELOW
	TriggerRising          ThresholdDirection = C.RISING
	TriggerFalling         ThresholdDirection = C.FALLING
	TriggerRisingOrFalling ThresholdDirection = C.RISING_OR_FALLING
	TriggerOutside         ThresholdDirection = C.OUTSIDE
	TriggerInside          ThresholdDirection = C.INSIDE
	TriggerEnter           ThresholdDirection = C.ENTER
	TriggerExit            ThresholdDirection = C.EXIT
	TriggerEnterOrExit     ThresholdDirection = C.ENTER_OR_EXIT
	TriggerNone            ThresholdDirection = C.NONE
)

type PicoInfo int

const (
	PicoDriverVersion              PicoInfo = C.PS3000_DRIVER_VERSION
	PicoUsbVersion                 PicoInfo = C.PS3000_USB_VERSION
	PicoHardwareVersion            PicoInfo = C.PS3000_HARDWARE_VERSION
	PicoVariantInfo                PicoInfo = C.PS3000_VARIANT_INFO
	PicoBatchAndSerial             PicoInfo = C.PS3000_BATCH_AND_SERIAL
	PicoCalDate                    PicoInfo = C.PS3000_CAL_DATE
	PicoKernelVersion              PicoInfo = C.PS3000_KERNEL_DRIVER_VERSION
	PicoDriverPath                 PicoInfo = C.PS3000_DRIVER_PATH
)

type TimeUnits C.PS3000_TIME_UNITS

const (
	TuFs TimeUnits = C.PS3000_FS
	TuPs TimeUnits = C.PS3000_PS
	TuNs TimeUnits = C.PS3000_NS
	TuUs TimeUnits = C.PS3000_US
	TuMs TimeUnits = C.PS3000_MS
	TuS  TimeUnits = C.PS3000_S
)

// ps3000 doesn't have ratio modes
type RatioMode int
const (
	RatioModeNone      RatioMode = 0
	RatioModeAggregate RatioMode = 1
	RatioModeDecimate  RatioMode = 2
	RatioModeAverage   RatioMode = 3
)

const (
	MinThresholdDiff       = 100
)

type TriggerState int
const (
	CondDontCare TriggerState = C.CONDITION_DONT_CARE
	CondTrue     TriggerState = C.CONDITION_TRUE
	CondFalse    TriggerState = C.CONDITION_FALSE
	CondMax      TriggerState = C.CONDITION_MAX
)

type PulseWidthType int
const (
	PwTypeNone        PulseWidthType = C.PW_TYPE_NONE
	PwTypeLessThan    PulseWidthType = C.PW_TYPE_LESS_THAN
	PwTypeGreaterThan PulseWidthType = C.PW_TYPE_GREATER_THAN
	PwTypeInRange     PulseWidthType = C.PW_TYPE_IN_RANGE
	PwTypeOutOfRange  PulseWidthType = C.PW_TYPE_OUT_OF_RANGE
)

type EtsMode int
const (
	EtsOff  EtsMode = C.PS3000_ETS_OFF
	EtsFast EtsMode = C.PS3000_ETS_FAST
	EtsSlow EtsMode = C.PS3000_ETS_SLOW
	EtsMax  EtsMode = C.PS3000_ETS_MODES_MAX
)

type DigitalChannel int
const (
	DchMax DigitalChannel = 0
)

type DigitalDirection int
const (
	DigitalMaxDirection             DigitalDirection = 0
)

type WaveTypeEnum int32
const (
	Sine      WaveTypeEnum = C.PS3000_SINE
	Square    WaveTypeEnum = C.PS3000_SQUARE
	Triangle  WaveTypeEnum = C.PS3000_TRIANGLE
	RampUp    WaveTypeEnum = 3
	RampDown  WaveTypeEnum = 4
	SinC      WaveTypeEnum = 5
	Gaussian  WaveTypeEnum = 6
	HalfSine  WaveTypeEnum = 7
	DcVoltage WaveTypeEnum = 8
)

type SweepTypeEnum int32
const (
	SweepUp       SweepTypeEnum = 0
	SweepDown     SweepTypeEnum = 1
	SweepUpDown   SweepTypeEnum = 2
	SweepDownUp   SweepTypeEnum = 3
	SweepMaxTypes SweepTypeEnum = 4
)

type ExtraOperations int
const (
	EsOff      ExtraOperations = 0
	WhiteNoise ExtraOperations = 1
	Prbs       ExtraOperations = 2
)

type SigGenTrigType int
const (
	SigGenRising   SigGenTrigType = 0
	SigGenFalling  SigGenTrigType = 1
	SigGenGateHigh SigGenTrigType = 2
	SigGenGateLow  SigGenTrigType = 3
)

type SigGenTrigSource int
const (
	SigGenNone      SigGenTrigSource = 0
	SigGenScopeTrig SigGenTrigSource = 1
	SigGenAuxIn     SigGenTrigSource = 2
	SigGenExtIn     SigGenTrigSource = 3
	SigGenSoftTrig  SigGenTrigSource = 4
)

type IndexMode int
const (
	Single        IndexMode = 0
	Dual          IndexMode = 1
	Quad          IndexMode = 2
	MaxIndexModes IndexMode = 3
)

type HoldOffType int
const (
	HofTime        HoldOffType = 0
	MaxHoldOffTime HoldOffType = 1
)

type TriggerOperand int
const (
	OperandNone TriggerOperand = 0
	OperandOr   TriggerOperand = 1
	OperandAnd  TriggerOperand = 2
	OperandThen TriggerOperand = 3
)

type DigitalPort int
const (
	Port0           DigitalPort = 0
	Port1           DigitalPort = 1
	Port2           DigitalPort = 2
	Port3           DigitalPort = 3
	MaxDigitalPorts DigitalPort = 4
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
	return inputRanges[int(r)]
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
	genericps.TriggerOutside = genericps.ThresholdDirection(TriggerOutside)
	genericps.TriggerInside = genericps.ThresholdDirection(TriggerInside)
	genericps.TriggerEnter = genericps.ThresholdDirection(TriggerEnter)
	genericps.TriggerExit = genericps.ThresholdDirection(TriggerExit)
	genericps.TriggerEnterOrExit = genericps.ThresholdDirection(TriggerEnterOrExit)
	genericps.TriggerNone = genericps.ThresholdDirection(TriggerNone)
	genericps.DchMax = genericps.DigitalChannel(DchMax)
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
	genericps.PicoDriverPath = genericps.PicoInfo(PicoDriverPath)
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
	genericps.MinThresholdDiff = MinThresholdDiff
}
