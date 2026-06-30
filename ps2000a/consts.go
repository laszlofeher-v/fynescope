//go:build !noscope

package ps2000a

// #cgo CFLAGS: -g -Wall
// #cgo LDFLAGS: -L/opt/picoscope/lib/ -lps2000a
// #include <stdlib.h>
// #include "/opt/picoscope/include/libps2000/ps2000.h"
// #include "/opt/picoscope/include/libps2000a/PicoStatus.h"
// #include "/opt/picoscope/include/libps2000a/ps2000aApi.h"
/*
// Forward declarations
int lpDataReady(int16_t handle, PICO_STATUS status, uint32_t noOfSamples,
				int16_t overflow, void * pParameter);
int lpStreamingReady(int16_t handle, int32_t noOfSamples, uint32_t startIndex,
                int16_t overflow, uint32_t triggerAt, int16_t triggered,
                int16_t autoStop, void * pParameter);
*/
import "C"
import "fynescope/genericps"

const (
	MaxChannelBuffers = C.PS2000A_MAX_CHANNEL_BUFFERS
	//	MaxDigitalPorts   = C.PS2000A_MAX_DIGITAL_PORTS
	MaxChannels       = C.PS2000A_MAX_CHANNELS
	ChannelInfoRanges = C.PS2000A_CI_RANGES
)

type NumOfChannelEnum int

const (
	DualScope NumOfChannelEnum = 2
	QuadScope NumOfChannelEnum = 4
)

type RangeEnum int

const (
	Range_10mv  RangeEnum = C.PS2000A_10MV
	Range_20mv  RangeEnum = C.PS2000A_20MV
	Range_50mv  RangeEnum = C.PS2000A_50MV
	Range_100mv RangeEnum = C.PS2000A_100MV
	Range_200mv RangeEnum = C.PS2000A_200MV
	Range_500mv RangeEnum = C.PS2000A_500MV
	Range_1v    RangeEnum = C.PS2000A_1V
	Range_2v    RangeEnum = C.PS2000A_2V
	Range_5v    RangeEnum = C.PS2000A_5V
	Range_10v   RangeEnum = C.PS2000A_10V
	Range_20v   RangeEnum = C.PS2000A_20V
	Range_50v   RangeEnum = C.PS2000A_50V
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
	ChA ChannelId = C.PS2000A_CHANNEL_A
	ChB ChannelId = C.PS2000A_CHANNEL_B
	ChC ChannelId = C.PS2000A_CHANNEL_C
	ChD ChannelId = C.PS2000A_CHANNEL_D
)

type ThresholdModeId int

const (
	Level  ThresholdModeId = C.PS2000A_LEVEL
	Window ThresholdModeId = C.PS2000A_WINDOW
)

type Coupling int

const (
	Ac Coupling = C.PS2000A_AC
	Dc Coupling = C.PS2000A_DC
)

type ThresholdDirection int

const (
	TriggerAbove           ThresholdDirection = C.PS2000A_ABOVE
	TriggerBelow           ThresholdDirection = C.PS2000A_BELOW
	TriggerRaising         ThresholdDirection = C.PS2000A_RISING
	TriggerFalling         ThresholdDirection = C.PS2000A_FALLING
	TriggerRisingOrFalling ThresholdDirection = C.PS2000A_RISING_OR_FALLING
	TriggerAboveLower      ThresholdDirection = C.PS2000A_ABOVE_LOWER
	TriggerBelowLower      ThresholdDirection = C.PS2000A_BELOW_LOWER
	TriggerRisingLower     ThresholdDirection = C.PS2000A_RISING_LOWER
	TriggerFallingLower    ThresholdDirection = C.PS2000A_FALLING_LOWER
	TriggerOutside         ThresholdDirection = C.PS2000A_OUTSIDE
	TriggerInside          ThresholdDirection = C.PS2000A_INSIDE
	TriggerEnter           ThresholdDirection = C.PS2000A_ENTER
	TriggerExit            ThresholdDirection = C.PS2000A_EXIT
	TriggerEnterOrExit     ThresholdDirection = C.PS2000A_ENTER_OR_EXIT
	TriggerPositiveRunt    ThresholdDirection = C.PS2000A_POSITIVE_RUNT
	TriggerNegativeRunt    ThresholdDirection = C.PS2000A_NEGATIVE_RUNT
	TriggerNone            ThresholdDirection = C.PS2000A_NONE
)

type PicoInfo int

const (
	PicoDriverVersion              PicoInfo = C.PICO_DRIVER_VERSION
	PicoUsbVersion                 PicoInfo = C.PICO_USB_VERSION
	PicoHardwareVersion            PicoInfo = C.PICO_HARDWARE_VERSION
	PicoVariantInfo                PicoInfo = C.PICO_VARIANT_INFO
	PicoBatchAndSerial             PicoInfo = C.PICO_BATCH_AND_SERIAL
	PicoCalDate                    PicoInfo = C.PICO_CAL_DATE
	PicoKernelVersion              PicoInfo = C.PICO_KERNEL_VERSION
	PicoDigitalHardwareVersion     PicoInfo = C.PICO_DIGITAL_HARDWARE_VERSION
	PicoAnalogueHardwareVersion    PicoInfo = C.PICO_ANALOGUE_HARDWARE_VERSION
	PicoFirmwareVersion1           PicoInfo = C.PICO_FIRMWARE_VERSION_1
	PicoFirmwareVersion2           PicoInfo = C.PICO_FIRMWARE_VERSION_2
	PicoMacAddress                 PicoInfo = C.PICO_MAC_ADDRESS
	PicoShadowCall                 PicoInfo = C.PICO_SHADOW_CAL
	PicoIppVersion                 PicoInfo = C.PICO_IPP_VERSION
	PicoDriverPath                 PicoInfo = C.PICO_DRIVER_PATH
	PicoFirmwareVersion3           PicoInfo = C.PICO_FIRMWARE_VERSION_3
	PicoFrontPanelFirmwareVersion3 PicoInfo = C.PICO_FRONT_PANEL_FIRMWARE_VERSION
	PicoBootloaderVersion          PicoInfo = C.PICO_BOOTLOADER_VERSION
)

type TimeUnits C.PS2000A_TIME_UNITS

const (
	TuFs TimeUnits = C.PS2000A_FS
	TuPs TimeUnits = C.PS2000A_PS
	TuNs TimeUnits = C.PS2000A_NS
	TuUs TimeUnits = C.PS2000A_US
	TuMs TimeUnits = C.PS2000A_MS
	TuS  TimeUnits = C.PS2000A_S
)

type RatioMode int

const (
	RatioModeNone      RatioMode = C.PS2000A_RATIO_MODE_NONE
	RatioModeAggregate RatioMode = C.PS2000A_RATIO_MODE_AGGREGATE
	RatioModeDecimate  RatioMode = C.PS2000A_RATIO_MODE_DECIMATE
	RatioModeAverage   RatioMode = C.PS2000A_RATIO_MODE_AVERAGE
)

const (
	AwgMinSigGenBufferSize = 1
	AwgMaxSigGenBufferSize = 8192
)

type TriggerState int

const (
	CondDontCare TriggerState = C.PS2000A_CONDITION_DONT_CARE
	CondTrue     TriggerState = C.PS2000A_CONDITION_TRUE
	CondFalse    TriggerState = C.PS2000A_CONDITION_FALSE
	CondMax      TriggerState = C.PS2000A_CONDITION_MAX
)

type PulseWidthType int

const (
	PwTypeNone        PulseWidthType = C.PS2000A_PW_TYPE_NONE
	PwTypeLessThan    PulseWidthType = C.PS2000A_PW_TYPE_LESS_THAN
	PwTypeGreaterThan PulseWidthType = C.PS2000A_PW_TYPE_GREATER_THAN
	PwTypeInRange     PulseWidthType = C.PS2000A_PW_TYPE_IN_RANGE
	PwTypeOutOfRange  PulseWidthType = C.PS2000A_PW_TYPE_OUT_OF_RANGE
)

type EtsMode int

const (
	EtsOff  EtsMode = C.PS2000A_ETS_OFF
	EtsFast EtsMode = C.PS2000A_ETS_FAST
	EtsSlow EtsMode = C.PS2000A_ETS_SLOW
	EtsMax  EtsMode = C.PS2000A_ETS_MODES_MAX
)
const (
	Ps2207MaxEtsCyscles       = C.PS2207_MAX_ETS_CYCLES
	Ps2207MaxEtsMaxInterleave = C.PS2207_MAX_INTERLEAVE
)

type DigitalChannel int

const (
	Dch0   DigitalChannel = C.PS2000A_DIGITAL_CHANNEL_0
	Dch1   DigitalChannel = C.PS2000A_DIGITAL_CHANNEL_1
	Dch2   DigitalChannel = C.PS2000A_DIGITAL_CHANNEL_2
	Dch3   DigitalChannel = C.PS2000A_DIGITAL_CHANNEL_3
	Dch4   DigitalChannel = C.PS2000A_DIGITAL_CHANNEL_4
	Dch5   DigitalChannel = C.PS2000A_DIGITAL_CHANNEL_5
	Dch6   DigitalChannel = C.PS2000A_DIGITAL_CHANNEL_6
	Dch7   DigitalChannel = C.PS2000A_DIGITAL_CHANNEL_7
	Dch8   DigitalChannel = C.PS2000A_DIGITAL_CHANNEL_8
	Dch9   DigitalChannel = C.PS2000A_DIGITAL_CHANNEL_9
	Dch10  DigitalChannel = C.PS2000A_DIGITAL_CHANNEL_10
	Dch11  DigitalChannel = C.PS2000A_DIGITAL_CHANNEL_11
	Dch12  DigitalChannel = C.PS2000A_DIGITAL_CHANNEL_12
	Dch13  DigitalChannel = C.PS2000A_DIGITAL_CHANNEL_13
	Dch14  DigitalChannel = C.PS2000A_DIGITAL_CHANNEL_14
	Dch15  DigitalChannel = C.PS2000A_DIGITAL_CHANNEL_15
	Dch16  DigitalChannel = C.PS2000A_DIGITAL_CHANNEL_16
	Dch17  DigitalChannel = C.PS2000A_DIGITAL_CHANNEL_17
	Dch18  DigitalChannel = C.PS2000A_DIGITAL_CHANNEL_18
	Dch19  DigitalChannel = C.PS2000A_DIGITAL_CHANNEL_19
	Dch20  DigitalChannel = C.PS2000A_DIGITAL_CHANNEL_20
	Dch21  DigitalChannel = C.PS2000A_DIGITAL_CHANNEL_21
	Dch22  DigitalChannel = C.PS2000A_DIGITAL_CHANNEL_22
	Dch23  DigitalChannel = C.PS2000A_DIGITAL_CHANNEL_23
	Dch24  DigitalChannel = C.PS2000A_DIGITAL_CHANNEL_24
	Dch25  DigitalChannel = C.PS2000A_DIGITAL_CHANNEL_25
	Dch26  DigitalChannel = C.PS2000A_DIGITAL_CHANNEL_26
	Dch27  DigitalChannel = C.PS2000A_DIGITAL_CHANNEL_27
	Dch28  DigitalChannel = C.PS2000A_DIGITAL_CHANNEL_28
	Dch29  DigitalChannel = C.PS2000A_DIGITAL_CHANNEL_29
	Dch30  DigitalChannel = C.PS2000A_DIGITAL_CHANNEL_30
	Dch31  DigitalChannel = C.PS2000A_DIGITAL_CHANNEL_31
	DchMax DigitalChannel = C.PS2000A_MAX_DIGITAL_CHANNELS
)

type DigitalDirection int

const (
	DigitalDontCare                 DigitalDirection = C.PS2000A_DIGITAL_DONT_CARE
	DigitalDirectionLow             DigitalDirection = C.PS2000A_DIGITAL_DIRECTION_LOW
	DigitalDirectionHigh            DigitalDirection = C.PS2000A_DIGITAL_DIRECTION_HIGH
	DigitalDirectionRising          DigitalDirection = C.PS2000A_DIGITAL_DIRECTION_RISING
	DigitalDirectionFalling         DigitalDirection = C.PS2000A_DIGITAL_DIRECTION_FALLING
	DigitalDirectionRisingOrFalling DigitalDirection = C.PS2000A_DIGITAL_DIRECTION_RISING_OR_FALLING
	DigitalMaxDirection             DigitalDirection = C.PS2000A_DIGITAL_MAX_DIRECTION
)

type WaveTypeEnum int16

const (
	Sine      WaveTypeEnum = C.PS2000A_SINE
	Square    WaveTypeEnum = C.PS2000A_SQUARE
	Triangle  WaveTypeEnum = C.PS2000A_TRIANGLE
	RampUp    WaveTypeEnum = C.PS2000A_RAMP_UP
	RampDown  WaveTypeEnum = C.PS2000A_RAMP_DOWN
	SinC      WaveTypeEnum = C.PS2000A_SINC
	Gaussian  WaveTypeEnum = C.PS2000A_GAUSSIAN
	HalfSine  WaveTypeEnum = C.PS2000A_HALF_SINE
	DcVoltage WaveTypeEnum = C.PS2000A_DC_VOLTAGE
)

type SweepTypeEnum C.PS2000A_SWEEP_TYPE

const (
	SweepUp       SweepTypeEnum = C.PS2000A_UP
	SweepDown     SweepTypeEnum = C.PS2000A_DOWN
	SweepUpDown   SweepTypeEnum = C.PS2000A_UPDOWN
	SweepDownUp   SweepTypeEnum = C.PS2000A_DOWNUP
	SweepMaxTypes SweepTypeEnum = C.PS2000A_MAX_SWEEP_TYPES
)

type ExtraOperations int

const (
	EsOff      ExtraOperations = C.PS2000A_ES_OFF
	WhiteNoise ExtraOperations = C.PS2000A_WHITENOISE
	Prbs       ExtraOperations = C.PS2000A_PRBS
)

const (
	SineMaxFrequency     = C.PS2000A_SINE_MAX_FREQUENCY
	SquareMaxFrequency   = C.PS2000A_SQUARE_MAX_FREQUENCY
	TriangleMaxFrequency = C.PS2000A_TRIANGLE_MAX_FREQUENCY
	SinCMaxFrequency     = C.PS2000A_SINC_MAX_FREQUENCY
	RampMaxFrequency     = C.PS2000A_RAMP_MAX_FREQUENCY
	HalfSineMaxFrequency = C.PS2000A_HALF_SINE_MAX_FREQUENCY
	GaussianMaxFrequency = C.PS2000A_GAUSSIAN_MAX_FREQUENCY
	PrbsMaxFrequency     = C.PS2000A_PRBS_MAX_FREQUENCY
	PrbsMinFrequency     = C.PS2000A_PRBS_MIN_FREQUENCY
	MinFrequency         = C.PS2000A_MIN_FREQUENCY
)

const (
	MaxSweepShots              = C.PS2000A_MAX_SWEEPS_SHOTS
	ShotSweepTriggerContinuous = C.PS2000A_SHOT_SWEEP_TRIGGER_CONTINUOUS_RUN
)

type SigGenTrigType int

const (
	SigGenRising   SigGenTrigType = C.PS2000A_SIGGEN_RISING
	SigGenFalling  SigGenTrigType = C.PS2000A_SIGGEN_FALLING
	SigGenGateHigh SigGenTrigType = C.PS2000A_SIGGEN_GATE_HIGH
	SigGenGateLow  SigGenTrigType = C.PS2000A_SIGGEN_GATE_LOW
)

type SigGenTrigSource int

const (
	SigGenNone      SigGenTrigSource = C.PS2000A_SIGGEN_NONE
	SigGenScopeTrig SigGenTrigSource = C.PS2000A_SIGGEN_SCOPE_TRIG
	SigGenAuxIn     SigGenTrigSource = C.PS2000A_SIGGEN_AUX_IN
	SigGenExtIn     SigGenTrigSource = C.PS2000A_SIGGEN_EXT_IN
	SigGenSoftTrig  SigGenTrigSource = C.PS2000A_SIGGEN_SOFT_TRIG
)

type IndexMode int

const (
	Single        IndexMode = C.PS2000A_SINGLE
	Dual          IndexMode = C.PS2000A_DUAL
	Quad          IndexMode = C.PS2000A_QUAD
	MaxIndexModes IndexMode = C.PS2000A_MAX_INDEX_MODES
)

type HoldOffType C.PS2000A_HOLDOFF_TYPE

const (
	HofTime        HoldOffType = C.PS2000A_TIME
	MaxHoldOffTime HoldOffType = C.PS2000A_MAX_HOLDOFF_TYPE
)

type TriggerOperand C.PS2000A_TRIGGER_OPERAND

const (
	OperandNone TriggerOperand = C.PS2000A_OPERAND_NONE
	OperandOr   TriggerOperand = C.PS2000A_OPERAND_OR
	OperandAnd  TriggerOperand = C.PS2000A_OPERAND_AND
	OperandThen TriggerOperand = C.PS2000A_OPERAND_THEN
)

type DigitalPort C.PS2000A_DIGITAL_PORT

const (
	Port0           DigitalPort = C.PS2000A_DIGITAL_PORT0 // digital channel 0 - 7
	Port1           DigitalPort = C.PS2000A_DIGITAL_PORT1 // digital channel 8 - 15
	Port2           DigitalPort = C.PS2000A_DIGITAL_PORT2 // digital channel 16 - 23
	Port3           DigitalPort = C.PS2000A_DIGITAL_PORT3 // digital channel 24 - 31
	MaxDigitalPorts DigitalPort = C.PS2000A_MAX_DIGITAL_PORTS
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
	genericps.TriggerRaising = genericps.ThresholdDirection(TriggerRaising)
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
}
