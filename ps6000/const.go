//go:build ps6000

package ps6000

// #cgo CFLAGS: -g -Wall -I/opt/picoscope/include/libps6000a
// #cgo LDFLAGS: -L/opt/picoscope/lib/ -lps6000
// #include <stdlib.h>
// #include "/opt/picoscope/include/libps6000/PicoStatus.h"
// #include "/opt/picoscope/include/libps6000/ps6000Api.h"
/*
// Forward declarations
int lpBlockReady(int16_t handle, PICO_STATUS status, void * pParameter);
int lpDataReady(int16_t handle, PICO_STATUS status, uint32_t noOfSamples,
				int16_t overflow, void * pParameter);
int lpStreamingReady(int16_t handle, int32_t noOfSamples, uint32_t startIndex,
                int16_t overflow, uint32_t triggerAt, int16_t triggered,
                int16_t autoStop, void * pParameter);
*/
import "C"

const (
	MaxChannelBuffers = C.PS6000_MAX_CHANNEL_BUFFERS
	//	MaxDigitalPorts   = C.PS6000_MAX_DIGITAL_PORTS
	MaxChannels       = C.PS6000_MAX_CHANNELS
	ChannelInfoRanges = C.PS6000_CI_RANGES
)

type NumOfChannelEnum int

const (
	DualScope NumOfChannelEnum = 2
	QuadScope NumOfChannelEnum = 4
)

type RangeEnum int

const (
	Range_10mv  RangeEnum = C.PS6000_10MV
	Range_20mv  RangeEnum = C.PS6000_20MV
	Range_50mv  RangeEnum = C.PS6000_50MV
	Range_100mv RangeEnum = C.PS6000_100MV
	Range_200mv RangeEnum = C.PS6000_200MV
	Range_500mv RangeEnum = C.PS6000_500MV
	Range_1v    RangeEnum = C.PS6000_1V
	Range_2v    RangeEnum = C.PS6000_2V
	Range_5v    RangeEnum = C.PS6000_5V
	Range_10v   RangeEnum = C.PS6000_10V
	Range_20v   RangeEnum = C.PS6000_20V
	Range_50v   RangeEnum = C.PS6000_50V
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
	ChA ChannelId = C.PS6000_CHANNEL_A
	ChB ChannelId = C.PS6000_CHANNEL_B
	ChC ChannelId = C.PS6000_CHANNEL_C
	ChD ChannelId = C.PS6000_CHANNEL_D
)

type ThresholdModeId int

const (
	Level  ThresholdModeId = C.PS6000_LEVEL
	Window ThresholdModeId = C.PS6000_WINDOW
)

type Coupling int

const (
	Ac Coupling = C.PS6000_AC
	Dc Coupling = C.PS6000_DC
)

type ThresholdDirection int

const (
	TriggerAbove           ThresholdDirection = C.PS6000_ABOVE
	TriggerBelow           ThresholdDirection = C.PS6000_BELOW
	TriggerRaising         ThresholdDirection = C.PS6000_RISING
	TriggerFalling         ThresholdDirection = C.PS6000_FALLING
	TriggerRisingOrFalling ThresholdDirection = C.PS6000_RISING_OR_FALLING
	TriggerAboveLower      ThresholdDirection = C.PS6000_ABOVE_LOWER
	TriggerBelowLower      ThresholdDirection = C.PS6000_BELOW_LOWER
	TriggerRisingLower     ThresholdDirection = C.PS6000_RISING_LOWER
	TriggerFallingLower    ThresholdDirection = C.PS6000_FALLING_LOWER
	TriggerOutside         ThresholdDirection = C.PS6000_OUTSIDE
	TriggerInside          ThresholdDirection = C.PS6000_INSIDE
	TriggerEnter           ThresholdDirection = C.PS6000_ENTER
	TriggerExit            ThresholdDirection = C.PS6000_EXIT
	TriggerEnterOrExit     ThresholdDirection = C.PS6000_ENTER_OR_EXIT
	TriggerPositiveRunt    ThresholdDirection = C.PS6000_POSITIVE_RUNT
	TriggerNegativeRunt    ThresholdDirection = C.PS6000_NEGATIVE_RUNT
	TriggerNone            ThresholdDirection = C.PS6000_NONE
)

type PicoInfo int

const (
	PicoDriverVersion              PicoInfo = C.PICO_DRIVER_VERSION
	PicoUsbVersion                 PicoInfo = C.PICO_USB_VERSION
	PicoHardwareVersion            PicoInfo = C.PICO_HARDWARE_VERSION
	PicoVariantInfo                PicoInfo = C.PICO_VARIANT_INFO
	PicoBatchAndSerial             PicoInfo = C.PICO_BATCH_AND_SERIAL
	PicoCalDate                    PicoInfo = C.PICO_CAL_DATE
	PicoKernelVarsion              PicoInfo = C.PICO_KERNEL_VERSION
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

type TimeUnits C.PS6000_TIME_UNITS

const (
	TuFs TimeUnits = C.PS6000_FS
	TuPs TimeUnits = C.PS6000_PS
	TuNs TimeUnits = C.PS6000_NS
	TuUs TimeUnits = C.PS6000_US
	TuMs TimeUnits = C.PS6000_MS
	TuS  TimeUnits = C.PS6000_S
)

type RatioMode int

const (
	RatioModeNone      RatioMode = C.RATIO_MODE_NONE
	RatioModeAggregate RatioMode = C.RATIO_MODE_AGGREGATE
	RatioModeDecimate  RatioMode = C.RATIO_MODE_DECIMATE
	RatioModeAverage   RatioMode = C.RATIO_MODE_AVERAGE
)

const (
	AwgMinSigGenBufferSize = 1
	AwgMaxSigGenBufferSize = 8192
	MinThresholdDiff       = 100
)

type TriggerState int

const (
	CondDontCare TriggerState = C.PS6000_CONDITION_DONT_CARE
	CondTrue     TriggerState = C.PS6000_CONDITION_TRUE
	CondFalse    TriggerState = C.PS6000_CONDITION_FALSE
	CondMax      TriggerState = C.PS6000_CONDITION_MAX
)

type PulseWidthType int

const (
	PwTypeNone        PulseWidthType = C.PS6000_PW_TYPE_NONE
	PwTypeLessThan    PulseWidthType = C.PS6000_PW_TYPE_LESS_THAN
	PwTypeGreaterThan PulseWidthType = C.PS6000_PW_TYPE_GREATER_THAN
	PwTypeInRange     PulseWidthType = C.PS6000_PW_TYPE_IN_RANGE
	PwTypeOutOfRange  PulseWidthType = C.PS6000_PW_TYPE_OUT_OF_RANGE
)

type EtsMode int

const (
	EtsOff  EtsMode = C.PS6000_ETS_OFF
	EtsFast EtsMode = C.PS6000_ETS_FAST
	EtsSlow EtsMode = C.PS6000_ETS_SLOW
	EtsMax  EtsMode = C.PS6000_ETS_MODES_MAX
)
const (
	Ps2207MaxEtsCyscles       = C.PS3207A_MAX_ETS_CYCLES
	Ps2207MaxEtsMaxInterleave = C.PS3207A_MAX_INTERLEAVE
)

type DigitalChannel int
const (
	DchMax DigitalChannel = 0
)

type DigitalDirection int
const (
	DigitalMaxDirection DigitalDirection = 0
)

type WaveTypeEnum int16

const (
	Sine      WaveTypeEnum = C.PS6000_SINE
	Square    WaveTypeEnum = C.PS6000_SQUARE
	Triangle  WaveTypeEnum = C.PS6000_TRIANGLE
	RampUp    WaveTypeEnum = C.PS6000_RAMP_UP
	RampDown  WaveTypeEnum = C.PS6000_RAMP_DOWN
	SinC      WaveTypeEnum = C.PS6000_SINC
	Gaussian  WaveTypeEnum = C.PS6000_GAUSSIAN
	HalfSine  WaveTypeEnum = C.PS6000_HALF_SINE
	DcVoltage WaveTypeEnum = C.PS6000_DC_VOLTAGE
)

type SweepTypeEnum C.SWEEP_TYPE

const (
	SweepUp       SweepTypeEnum = C.PS6000_UP
	SweepDown     SweepTypeEnum = C.PS6000_DOWN
	SweepUpDown   SweepTypeEnum = C.PS6000_UPDOWN
	SweepDownUp   SweepTypeEnum = C.PS6000_DOWNUP
	SweepMaxTypes SweepTypeEnum = C.PS6000_MAX_SWEEP_TYPES
)

type ExtraOperations int

const (
	EsOff      ExtraOperations = C.PS6000_ES_OFF
	WhiteNoise ExtraOperations = C.PS6000_WHITENOISE
	Prbs       ExtraOperations = C.PS6000_PRBS
)

const (
	SineMaxFrequency     = C.PS6000_SINE_MAX_FREQUENCY
	SquareMaxFrequency   = C.PS6000_SQUARE_MAX_FREQUENCY
	TriangleMaxFrequency = C.PS6000_TRIANGLE_MAX_FREQUENCY
	SinCMaxFrequency     = C.PS6000_SINC_MAX_FREQUENCY
	RampMaxFrequency     = C.PS6000_RAMP_MAX_FREQUENCY
	HalfSineMaxFrequency = C.PS6000_HALF_SINE_MAX_FREQUENCY
	GaussianMaxFrequency = C.PS6000_GAUSSIAN_MAX_FREQUENCY
	PrbsMaxFrequency     = C.PS6000_PRBS_MAX_FREQUENCY
	PrbsMinFrequency     = C.PS6000_PRBS_MIN_FREQUENCY
	MinFrequency         = C.PS6000_MIN_FREQUENCY
)

const (
	MaxSweepShots              = C.MAX_SWEEPS_SHOTS
	ShotSweepTriggerContinuous = C.PS6000_SHOT_SWEEP_TRIGGER_CONTINUOUS_RUN
)

type SigGenTrigType int

const (
	SigGenRising   SigGenTrigType = C.PS6000_SIGGEN_RISING
	SigGenFalling  SigGenTrigType = C.PS6000_SIGGEN_FALLING
	SigGenGateHigh SigGenTrigType = C.PS6000_SIGGEN_GATE_HIGH
	SigGenGateLow  SigGenTrigType = C.PS6000_SIGGEN_GATE_LOW
)

type SigGenTrigSource int

const (
	SigGenNone      SigGenTrigSource = C.PS6000_SIGGEN_NONE
	SigGenScopeTrig SigGenTrigSource = C.PS6000_SIGGEN_SCOPE_TRIG
	SigGenAuxIn     SigGenTrigSource = C.PS6000_SIGGEN_AUX_IN
	SigGenExtIn     SigGenTrigSource = C.PS6000_SIGGEN_EXT_IN
	SigGenSoftTrig  SigGenTrigSource = C.PS6000_SIGGEN_SOFT_TRIG
)

type IndexMode int

const (
	Single        IndexMode = C.PS6000_SINGLE
	Dual          IndexMode = C.PS6000_DUAL
	Quad          IndexMode = C.PS6000_QUAD
	MaxIndexModes IndexMode = C.PS6000_MAX_INDEX_MODES
)

type HoldOffType int
const (
	HofTime HoldOffType = 0
	MaxHoldOffTime HoldOffType = 1
)

type DigitalPort int
const (
	MaxDigitalPorts DigitalPort = 0
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
