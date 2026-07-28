//go:build ps5000a

package ps5000a

// #cgo CFLAGS: -g -Wall -I/opt/picoscope/include/libps6000a
// #cgo LDFLAGS: -L/opt/picoscope/lib/ -lps5000a
// #include <stdlib.h>
// #include "/opt/picoscope/include/libps5000a/PicoStatus.h"
// #include "/opt/picoscope/include/libps5000a/ps5000aApi.h"
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
	// MaxChannelBuffers = C.PS5000A_MAX_CHANNEL_BUFFERS
	//	MaxDigitalPorts   = C.PS5000A_MAX_DIGITAL_PORTS
	MaxChannels       = C.PS5000A_MAX_CHANNELS
	ChannelInfoRanges = C.PS5000A_CI_RANGES
)

type NumOfChannelEnum int

const (
	DualScope NumOfChannelEnum = 2
	QuadScope NumOfChannelEnum = 4
)

type RangeEnum int

const (
	Range_10mv  RangeEnum = C.PS5000A_10MV
	Range_20mv  RangeEnum = C.PS5000A_20MV
	Range_50mv  RangeEnum = C.PS5000A_50MV
	Range_100mv RangeEnum = C.PS5000A_100MV
	Range_200mv RangeEnum = C.PS5000A_200MV
	Range_500mv RangeEnum = C.PS5000A_500MV
	Range_1v    RangeEnum = C.PS5000A_1V
	Range_2v    RangeEnum = C.PS5000A_2V
	Range_5v    RangeEnum = C.PS5000A_5V
	Range_10v   RangeEnum = C.PS5000A_10V
	Range_20v   RangeEnum = C.PS5000A_20V
	Range_50v   RangeEnum = C.PS5000A_50V
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
	ChA ChannelId = C.PS5000A_CHANNEL_A
	ChB ChannelId = C.PS5000A_CHANNEL_B
	ChC ChannelId = C.PS5000A_CHANNEL_C
	ChD ChannelId = C.PS5000A_CHANNEL_D
)

type ThresholdModeId int

const (
	Level  ThresholdModeId = C.PS5000A_LEVEL
	Window ThresholdModeId = C.PS5000A_WINDOW
)

type Coupling int

const (
	Ac Coupling = C.PS5000A_AC
	Dc Coupling = C.PS5000A_DC
)

type ThresholdDirection int

const (
	TriggerAbove           ThresholdDirection = C.PS5000A_ABOVE
	TriggerBelow           ThresholdDirection = C.PS5000A_BELOW
	TriggerRaising         ThresholdDirection = C.PS5000A_RISING
	TriggerFalling         ThresholdDirection = C.PS5000A_FALLING
	TriggerRisingOrFalling ThresholdDirection = C.PS5000A_RISING_OR_FALLING
	TriggerAboveLower      ThresholdDirection = C.PS5000A_ABOVE_LOWER
	TriggerBelowLower      ThresholdDirection = C.PS5000A_BELOW_LOWER
	TriggerRisingLower     ThresholdDirection = C.PS5000A_RISING_LOWER
	TriggerFallingLower    ThresholdDirection = C.PS5000A_FALLING_LOWER
	TriggerOutside         ThresholdDirection = C.PS5000A_OUTSIDE
	TriggerInside          ThresholdDirection = C.PS5000A_INSIDE
	TriggerEnter           ThresholdDirection = C.PS5000A_ENTER
	TriggerExit            ThresholdDirection = C.PS5000A_EXIT
	TriggerEnterOrExit     ThresholdDirection = C.PS5000A_ENTER_OR_EXIT
	TriggerPositiveRunt    ThresholdDirection = C.PS5000A_POSITIVE_RUNT
	TriggerNegativeRunt    ThresholdDirection = C.PS5000A_NEGATIVE_RUNT
	TriggerNone            ThresholdDirection = C.PS5000A_NONE
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

type TimeUnits C.PS5000A_TIME_UNITS

const (
	TuFs TimeUnits = C.PS5000A_FS
	TuPs TimeUnits = C.PS5000A_PS
	TuNs TimeUnits = C.PS5000A_NS
	TuUs TimeUnits = C.PS5000A_US
	TuMs TimeUnits = C.PS5000A_MS
	TuS  TimeUnits = C.PS5000A_S
)

type RatioMode int

const (
	RatioModeNone      RatioMode = C.PS5000A_RATIO_MODE_NONE
	RatioModeAggregate RatioMode = C.PS5000A_RATIO_MODE_AGGREGATE
	RatioModeDecimate  RatioMode = C.PS5000A_RATIO_MODE_DECIMATE
	RatioModeAverage   RatioMode = C.PS5000A_RATIO_MODE_AVERAGE
)

const (
	AwgMinSigGenBufferSize = 1
	AwgMaxSigGenBufferSize = 8192
	MinThresholdDiff       = 100
)

type TriggerState int

const (
	CondDontCare TriggerState = C.PS5000A_CONDITION_DONT_CARE
	CondTrue     TriggerState = C.PS5000A_CONDITION_TRUE
	CondFalse    TriggerState = C.PS5000A_CONDITION_FALSE
	CondMax      TriggerState = C.PS5000A_CONDITION_MAX
)

type PulseWidthType int

const (
	PwTypeNone        PulseWidthType = C.PS5000A_PW_TYPE_NONE
	PwTypeLessThan    PulseWidthType = C.PS5000A_PW_TYPE_LESS_THAN
	PwTypeGreaterThan PulseWidthType = C.PS5000A_PW_TYPE_GREATER_THAN
	PwTypeInRange     PulseWidthType = C.PS5000A_PW_TYPE_IN_RANGE
	PwTypeOutOfRange  PulseWidthType = C.PS5000A_PW_TYPE_OUT_OF_RANGE
)

type EtsMode int

const (
	EtsOff  EtsMode = C.PS5000A_ETS_OFF
	EtsFast EtsMode = C.PS5000A_ETS_FAST
	EtsSlow EtsMode = C.PS5000A_ETS_SLOW
	EtsMax  EtsMode = C.PS5000A_ETS_MODES_MAX
)
const (
	Ps5242aMaxEtsCyscles       = C.PS5242A_MAX_ETS_CYCLES
	Ps5242aMaxEtsMaxInterleave = C.PS5242A_MAX_ETS_INTERLEAVE
	Ps5243aMaxEtsCyscles       = C.PS5243A_MAX_ETS_CYCLES
	Ps5243aMaxEtsMaxInterleave = C.PS5243A_MAX_ETS_INTERLEAVE
	Ps5244aMaxEtsCyscles       = C.PS5244A_MAX_ETS_CYCLES
	Ps5244aMaxEtsMaxInterleave = C.PS5244A_MAX_ETS_INTERLEAVE
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
	Sine      WaveTypeEnum = C.PS5000A_SINE
	Square    WaveTypeEnum = C.PS5000A_SQUARE
	Triangle  WaveTypeEnum = C.PS5000A_TRIANGLE
	RampUp    WaveTypeEnum = C.PS5000A_RAMP_UP
	RampDown  WaveTypeEnum = C.PS5000A_RAMP_DOWN
	SinC      WaveTypeEnum = C.PS5000A_SINC
	Gaussian  WaveTypeEnum = C.PS5000A_GAUSSIAN
	HalfSine  WaveTypeEnum = C.PS5000A_HALF_SINE
	DcVoltage WaveTypeEnum = C.PS5000A_DC_VOLTAGE
)

type SweepTypeEnum C.PS5000A_SWEEP_TYPE

const (
	SweepUp       SweepTypeEnum = C.PS5000A_UP
	SweepDown     SweepTypeEnum = C.PS5000A_DOWN
	SweepUpDown   SweepTypeEnum = C.PS5000A_UPDOWN
	SweepDownUp   SweepTypeEnum = C.PS5000A_DOWNUP
	SweepMaxTypes SweepTypeEnum = C.PS5000A_MAX_SWEEP_TYPES
)

type ExtraOperations int

const (
	EsOff      ExtraOperations = C.PS5000A_ES_OFF
	WhiteNoise ExtraOperations = C.PS5000A_WHITENOISE
	Prbs       ExtraOperations = C.PS5000A_PRBS
)

const (
	SineMaxFrequency     = C.PS5000A_SINE_MAX_FREQUENCY
	SquareMaxFrequency   = C.PS5000A_SQUARE_MAX_FREQUENCY
	TriangleMaxFrequency = C.PS5000A_TRIANGLE_MAX_FREQUENCY
	SinCMaxFrequency     = C.PS5000A_SINC_MAX_FREQUENCY
	RampMaxFrequency     = C.PS5000A_RAMP_MAX_FREQUENCY
	HalfSineMaxFrequency = C.PS5000A_HALF_SINE_MAX_FREQUENCY
	GaussianMaxFrequency = C.PS5000A_GAUSSIAN_MAX_FREQUENCY
	// PrbsMaxFrequency     = C.PS5000A_PRBS_MAX_FREQUENCY
	// PrbsMinFrequency     = C.PS5000A_PRBS_MIN_FREQUENCY
	MinFrequency = C.PS5000A_MIN_FREQUENCY
)

const (
	MaxSweepShots              = C.MAX_SWEEPS_SHOTS
	ShotSweepTriggerContinuous = C.PS5000A_SHOT_SWEEP_TRIGGER_CONTINUOUS_RUN
)

type SigGenTrigType int

const (
	SigGenRising   SigGenTrigType = C.PS5000A_SIGGEN_RISING
	SigGenFalling  SigGenTrigType = C.PS5000A_SIGGEN_FALLING
	SigGenGateHigh SigGenTrigType = C.PS5000A_SIGGEN_GATE_HIGH
	SigGenGateLow  SigGenTrigType = C.PS5000A_SIGGEN_GATE_LOW
)

type SigGenTrigSource int

const (
	SigGenNone      SigGenTrigSource = C.PS5000A_SIGGEN_NONE
	SigGenScopeTrig SigGenTrigSource = C.PS5000A_SIGGEN_SCOPE_TRIG
	SigGenAuxIn     SigGenTrigSource = C.PS5000A_SIGGEN_AUX_IN
	SigGenExtIn     SigGenTrigSource = C.PS5000A_SIGGEN_EXT_IN
	SigGenSoftTrig  SigGenTrigSource = C.PS5000A_SIGGEN_SOFT_TRIG
)

type IndexMode int

const (
	Single        IndexMode = C.PS5000A_SINGLE
	Dual          IndexMode = C.PS5000A_DUAL
	Quad          IndexMode = C.PS5000A_QUAD
	MaxIndexModes IndexMode = C.PS5000A_MAX_INDEX_MODES
)

type HoldOffType int

const (
	HofTime        HoldOffType = 0
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
