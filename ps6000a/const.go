//go:build !demo && ps6000a

package ps6000a

// #cgo CFLAGS: -g -Wall -I/opt/picoscope/include/libps6000a
// #cgo LDFLAGS: -L/opt/picoscope/lib/ -lps6000a
// #include <stdlib.h>
// #include "/opt/picoscope/include/libps3000/ps3000.h"
// #include "/opt/picoscope/include/libps6000a/PicoStatus.h"
// #include "/opt/picoscope/include/libps6000a/ps6000aApi.h"
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
	MaxChannelBuffers = 2
	//	MaxDigitalPorts   = 4
	MaxChannels       = 8
	ChannelInfoRanges = 1
)

type NumOfChannelEnum int

const (
	DualScope NumOfChannelEnum = 2
	QuadScope NumOfChannelEnum = 4
)

type RangeEnum int

// const (
// 	Range_10mv  RangeEnum = C.PICO_10MV
// 	Range_20mv  RangeEnum = C.PICO_20MV
// 	Range_50mv  RangeEnum = C.PICO_50MV
// 	Range_100mv RangeEnum = C.PICO_100MV
// 	Range_200mv RangeEnum = C.PICO_200MV
// 	Range_500mv RangeEnum = C.PICO_500MV
// 	Range_1v    RangeEnum = C.PICO_1V
// 	Range_2v    RangeEnum = C.PICO_2V
// 	Range_5v    RangeEnum = C.PICO_5V
// 	Range_10v   RangeEnum = C.PICO_10V
// 	Range_20v   RangeEnum = C.PICO_20V
// 	Range_50v   RangeEnum = C.PICO_50V
// )

// var (
// 	RangeValuesMv = map[RangeEnum]float64{
// 		Range_10mv:  10.0,
// 		Range_20mv:  20.0,
// 		Range_50mv:  50.0,
// 		Range_100mv: 100.0,
// 		Range_200mv: 200.0,
// 		Range_500mv: 500.0,
// 		Range_1v:    1000.0,
// 		Range_2v:    2000.0,
// 		Range_5v:    5000.0,
// 		Range_10v:   10000.0,
// 		Range_20v:   20000.0,
// 		Range_50v:   50000.0,
// 	}
// )

type ChannelId int

const (
	ChA ChannelId = C.PICO_CHANNEL_A
	ChB ChannelId = C.PICO_CHANNEL_B
	ChC ChannelId = C.PICO_CHANNEL_C
	ChD ChannelId = C.PICO_CHANNEL_D
)

type ThresholdModeId int

const (
	Level  ThresholdModeId = C.PICO_LEVEL
	Window ThresholdModeId = C.PICO_WINDOW
)

type Coupling int

const (
	Ac Coupling = C.PICO_AC
	Dc Coupling = C.PICO_DC
)

type ThresholdDirection int

const (
	TriggerAbove           ThresholdDirection = C.PICO_ABOVE
	TriggerBelow           ThresholdDirection = C.PICO_BELOW
	TriggerRaising         ThresholdDirection = C.PICO_RISING
	TriggerFalling         ThresholdDirection = C.PICO_FALLING
	TriggerRisingOrFalling ThresholdDirection = C.PICO_RISING_OR_FALLING
	TriggerAboveLower      ThresholdDirection = C.PICO_ABOVE_LOWER
	TriggerBelowLower      ThresholdDirection = C.PICO_BELOW_LOWER
	TriggerRisingLower     ThresholdDirection = C.PICO_RISING_LOWER
	TriggerFallingLower    ThresholdDirection = C.PICO_FALLING_LOWER
	TriggerOutside         ThresholdDirection = C.PICO_OUTSIDE
	TriggerInside          ThresholdDirection = C.PICO_INSIDE
	TriggerEnter           ThresholdDirection = C.PICO_ENTER
	TriggerExit            ThresholdDirection = C.PICO_EXIT
	TriggerEnterOrExit     ThresholdDirection = C.PICO_ENTER_OR_EXIT
	TriggerPositiveRunt    ThresholdDirection = C.PICO_POSITIVE_RUNT
	TriggerNegativeRunt    ThresholdDirection = C.PICO_NEGATIVE_RUNT
	TriggerNone            ThresholdDirection = C.PICO_NONE
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

type TimeUnits C.PICO_TIME_UNITS

const (
	TuFs TimeUnits = C.PICO_FS
	TuPs TimeUnits = C.PICO_PS
	TuNs TimeUnits = C.PICO_NS
	TuUs TimeUnits = C.PICO_US
	TuMs TimeUnits = C.PICO_MS
	TuS  TimeUnits = C.PICO_S
)

type RatioMode int

const (
	RatioModeAggregate    RatioMode = C.PICO_RATIO_MODE_AGGREGATE
	RatioModeDecimate     RatioMode = C.PICO_RATIO_MODE_DECIMATE
	RatioModeAverage      RatioMode = C.PICO_RATIO_MODE_AVERAGE
	RatioModeDistribution RatioMode = C.PICO_RATIO_MODE_DISTRIBUTION
	RatioModeSum          RatioMode = C.PICO_RATIO_MODE_SUM
)

const (
	AwgMinSigGenBufferSize = 1
	AwgMaxSigGenBufferSize = 8192
	MinThresholdDiff       = 100
)

type TriggerState int

const (
	CondDontCare TriggerState = C.PICO_CONDITION_DONT_CARE
	CondTrue     TriggerState = C.PICO_CONDITION_TRUE
	CondFalse    TriggerState = C.PICO_CONDITION_FALSE
	// CondMax      TriggerState = C.PICO_CONDITION_MAX
)

type PulseWidthType int

const (
	PwTypeNone        PulseWidthType = C.PICO_PW_TYPE_NONE
	PwTypeLessThan    PulseWidthType = C.PICO_PW_TYPE_LESS_THAN
	PwTypeGreaterThan PulseWidthType = C.PICO_PW_TYPE_GREATER_THAN
	PwTypeInRange     PulseWidthType = C.PICO_PW_TYPE_IN_RANGE
	PwTypeOutOfRange  PulseWidthType = C.PICO_PW_TYPE_OUT_OF_RANGE
)

type EtsMode int

const (
	EtsOff  EtsMode = C.PICO_ETS_OFF
	EtsFast EtsMode = C.PICO_ETS_FAST
	EtsSlow EtsMode = C.PICO_ETS_SLOW
)
const (
	Ps2207MaxEtsCyscles       = 0
	Ps2207MaxEtsMaxInterleave = 0
)

type DigitalChannel int

const (
	Dch0 DigitalChannel = C.PICO_PORT_DIGITAL_CHANNEL0
	Dch1 DigitalChannel = C.PICO_PORT_DIGITAL_CHANNEL1
	Dch2 DigitalChannel = C.PICO_PORT_DIGITAL_CHANNEL2
	Dch3 DigitalChannel = C.PICO_PORT_DIGITAL_CHANNEL3
	Dch4 DigitalChannel = C.PICO_PORT_DIGITAL_CHANNEL4
	Dch5 DigitalChannel = C.PICO_PORT_DIGITAL_CHANNEL5
	Dch6 DigitalChannel = C.PICO_PORT_DIGITAL_CHANNEL6
	Dch7 DigitalChannel = C.PICO_PORT_DIGITAL_CHANNEL7
)

type DigitalDirection int

const (
	DigitalDontCare                 DigitalDirection = C.PICO_DIGITAL_DONT_CARE
	DigitalDirectionLow             DigitalDirection = C.PICO_DIGITAL_DIRECTION_LOW
	DigitalDirectionHigh            DigitalDirection = C.PICO_DIGITAL_DIRECTION_HIGH
	DigitalDirectionRising          DigitalDirection = C.PICO_DIGITAL_DIRECTION_RISING
	DigitalDirectionFalling         DigitalDirection = C.PICO_DIGITAL_DIRECTION_FALLING
	DigitalDirectionRisingOrFalling DigitalDirection = C.PICO_DIGITAL_DIRECTION_RISING_OR_FALLING
	DigitalMaxDirection             DigitalDirection = C.PICO_DIGITAL_MAX_DIRECTION
)

type WaveTypeEnum int16

const (
	Sine      WaveTypeEnum = C.PICO_SINE
	Square    WaveTypeEnum = C.PICO_SQUARE
	Triangle  WaveTypeEnum = C.PICO_TRIANGLE
	RampUp    WaveTypeEnum = C.PICO_RAMP_UP
	RampDown  WaveTypeEnum = C.PICO_RAMP_DOWN
	SinC      WaveTypeEnum = C.PICO_SINC
	Gaussian  WaveTypeEnum = C.PICO_GAUSSIAN
	HalfSine  WaveTypeEnum = C.PICO_HALF_SINE
	DcVoltage WaveTypeEnum = C.PICO_DC_VOLTAGE
)

type SweepTypeEnum C.PICO_SWEEP_TYPE

const (
	SweepUp     SweepTypeEnum = C.PICO_UP
	SweepDown   SweepTypeEnum = C.PICO_DOWN
	SweepUpDown SweepTypeEnum = C.PICO_UPDOWN
	SweepDownUp SweepTypeEnum = C.PICO_DOWNUP
)

type ExtraOperations int

const (
	EsOff      ExtraOperations = 0
	WhiteNoise ExtraOperations = 1
	Prbs       ExtraOperations = 2
)

const (
	SineMaxFrequency     = 0
	SquareMaxFrequency   = 0
	TriangleMaxFrequency = 0
	SinCMaxFrequency     = 0
	RampMaxFrequency     = 0
	HalfSineMaxFrequency = 0
	GaussianMaxFrequency = 0
	PrbsMaxFrequency     = 0
	PrbsMinFrequency     = 0
	MinFrequency         = 0
)

const (
	MaxSweepShots              = 0
	ShotSweepTriggerContinuous = 0
)

type SigGenTrigType int

const (
	SigGenRising   SigGenTrigType = C.PICO_SIGGEN_RISING
	SigGenFalling  SigGenTrigType = C.PICO_SIGGEN_FALLING
	SigGenGateHigh SigGenTrigType = C.PICO_SIGGEN_GATE_HIGH
	SigGenGateLow  SigGenTrigType = C.PICO_SIGGEN_GATE_LOW
)

type SigGenTrigSource int

const (
	SigGenNone      SigGenTrigSource = C.PICO_SIGGEN_NONE
	SigGenScopeTrig SigGenTrigSource = C.PICO_SIGGEN_SCOPE_TRIG
	SigGenAuxIn     SigGenTrigSource = C.PICO_SIGGEN_AUX_IN
	SigGenExtIn     SigGenTrigSource = C.PICO_SIGGEN_EXT_IN
	SigGenSoftTrig  SigGenTrigSource = C.PICO_SIGGEN_SOFT_TRIG
)

type IndexMode int

const (
	Single IndexMode = C.PICO_SINGLE
	Dual   IndexMode = C.PICO_DUAL
	Quad   IndexMode = C.PICO_QUAD
)

type HoldOffType int

const (
	HofTime        HoldOffType = 0
	MaxHoldOffTime HoldOffType = 1
)

type DigitalPort int

const (
	Port0           DigitalPort = 128 // digital channel 0 - 7
	Port1           DigitalPort = 129 // digital channel 8 - 15
	Port2           DigitalPort = 130 // digital channel 16 - 23
	Port3           DigitalPort = 131 // digital channel 24 - 31
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

type ConnectProbe int32

const (
	ConnectProbeNone                 ConnectProbe = C.PICO_CONNECT_PROBE_NONE
	ConnectProbeD9Bnc                ConnectProbe = C.PICO_CONNECT_PROBE_D9_BNC
	ConnectProbeD92xBnc              ConnectProbe = C.PICO_CONNECT_PROBE_D9_2X_BNC
	ConnectProbeDifferential         ConnectProbe = C.PICO_CONNECT_PROBE_DIFFERENTIAL
	ConnectProbeCurrentClamp200_2ka  ConnectProbe = C.PICO_CONNECT_PROBE_CURRENT_CLAMP_200_2KA
	ConnectProbeCurrentClamp40a      ConnectProbe = C.PICO_CONNECT_PROBE_CURRENT_CLAMP_40A
	ConnectProbeCat3Hv1kv            ConnectProbe = C.PICO_CONNECT_PROBE_CAT3_HV_1KV
	ConnectProbeCurrentClamp2000arms ConnectProbe = C.PICO_CONNECT_PROBE_CURRENT_CLAMP_2000ARMS

	BncPlusPremiumTestLeadBlue   ConnectProbe = C.PICO_BNC_PLUS_PREMIUM_TEST_LEAD_BLUE
	BncPlusPremiumTestLeadRed    ConnectProbe = C.PICO_BNC_PLUS_PREMIUM_TEST_LEAD_RED
	BncPlusPremiumTestLeadGreen  ConnectProbe = C.PICO_BNC_PLUS_PREMIUM_TEST_LEAD_GREEN
	BncPlusPremiumTestLeadYellow ConnectProbe = C.PICO_BNC_PLUS_PREMIUM_TEST_LEAD_YELLOW
	BncPlusCopProbe              ConnectProbe = C.PICO_BNC_PLUS_COP_PROBE

	BncPlusTemperatureProbe        ConnectProbe = C.PICO_BNC_PLUS_TEMPERATURE_PROBE
	BncPlus100aCurrentClamp        ConnectProbe = C.PICO_BNC_PLUS_100A_CURRENT_CLAMP
	BncPlusHtPickup                ConnectProbe = C.PICO_BNC_PLUS_HT_PICKUP
	BncPlusX10ScopeProbe           ConnectProbe = C.PICO_BNC_PLUS_X10_SCOPE_PROBE
	BncPlus2000aCurrentClamp       ConnectProbe = C.PICO_BNC_PLUS_2000A_CURRENT_CLAMP
	BncPlusPressureSensor          ConnectProbe = C.PICO_BNC_PLUS_PRESSURE_SENSOR
	BncPlusResistanceLead          ConnectProbe = C.PICO_BNC_PLUS_RESISTANCE_LEAD
	BncPlus60aCurrentClamp         ConnectProbe = C.PICO_BNC_PLUS_60A_CURRENT_CLAMP
	BncPlusOpticalSensor           ConnectProbe = C.PICO_BNC_PLUS_OPTICAL_SENSOR
	BncPlus60aCurrentClampV2       ConnectProbe = C.PICO_BNC_PLUS_60A_CURRENT_CLAMP_V2
	BncPlusHighVoltageDifferential ConnectProbe = C.PICO_BNC_PLUS_HIGH_VOLTAGE_DIFFERENTIAL
	BncPlusTriAxialAccelerometer   ConnectProbe = C.PICO_BNC_PLUS_TRI_AXIAL_ACCELEROMETER
	BncPlusMicrophone              ConnectProbe = C.PICO_BNC_PLUS_MICROPHONE

	PassiveProbeX10       ConnectProbe = C.PICO_PASSIVE_PROBE_X10
	ActiveX10_750mhz      ConnectProbe = C.PICO_ACTIVE_X10_750MHZ
	ActiveX10_1_3ghz      ConnectProbe = C.PICO_ACTIVE_X10_1_3GHZ
	PassiveProbeX10_50ohm ConnectProbe = C.PICO_PASSIVE_PROBE_X10_50OHM

	ConnectProbeIntelligent ConnectProbe = C.PICO_CONNECT_PROBE_INTELLIGENT
	ConnectProbeUnknown     ConnectProbe = C.PICO_CONNECT_PROBE_UNKNOWN_PROBE
	ConnectProbeFault       ConnectProbe = C.PICO_CONNECT_PROBE_FAULT_PROBE
)

type ConnectProbeRange int32

const (
	X1Probe10mv   ConnectProbeRange = C.PICO_X1_PROBE_10MV
	X1Probe20mv   ConnectProbeRange = C.PICO_X1_PROBE_20MV
	X1Probe50mv   ConnectProbeRange = C.PICO_X1_PROBE_50MV
	X1Probe100mv  ConnectProbeRange = C.PICO_X1_PROBE_100MV
	X1Probe200mv  ConnectProbeRange = C.PICO_X1_PROBE_200MV
	X1Probe500mv  ConnectProbeRange = C.PICO_X1_PROBE_500MV
	X1Probe1v     ConnectProbeRange = C.PICO_X1_PROBE_1V
	X1Probe2v     ConnectProbeRange = C.PICO_X1_PROBE_2V
	X1Probe5v     ConnectProbeRange = C.PICO_X1_PROBE_5V
	X1Probe10v    ConnectProbeRange = C.PICO_X1_PROBE_10V
	X1Probe20v    ConnectProbeRange = C.PICO_X1_PROBE_20V
	X1Probe50v    ConnectProbeRange = C.PICO_X1_PROBE_50V
	X1Probe100v   ConnectProbeRange = C.PICO_X1_PROBE_100V
	X1Probe200v   ConnectProbeRange = C.PICO_X1_PROBE_200V
	X1ProbeRanges ConnectProbeRange = C.PICO_X1_PROBE_RANGES

	X10Probe100mv  ConnectProbeRange = C.PICO_X10_PROBE_100MV
	X10Probe200mv  ConnectProbeRange = C.PICO_X10_PROBE_200MV
	X10Probe500mv  ConnectProbeRange = C.PICO_X10_PROBE_500MV
	X10Probe1v     ConnectProbeRange = C.PICO_X10_PROBE_1V
	X10Probe2v     ConnectProbeRange = C.PICO_X10_PROBE_2V
	X10Probe5v     ConnectProbeRange = C.PICO_X10_PROBE_5V
	X10Probe10v    ConnectProbeRange = C.PICO_X10_PROBE_10V
	X10Probe20v    ConnectProbeRange = C.PICO_X10_PROBE_20V
	X10Probe50v    ConnectProbeRange = C.PICO_X10_PROBE_50V
	X10Probe100v   ConnectProbeRange = C.PICO_X10_PROBE_100MV
	X10Probe200v   ConnectProbeRange = C.PICO_X10_PROBE_200MV
	X10Probe500v   ConnectProbeRange = C.PICO_X10_PROBE_500MV
	X10ProbeRanges ConnectProbeRange = C.PICO_X10_PROBE_RANGES

	Ps4000aResistance315k      ConnectProbeRange = C.PICO_PS4000A_RESISTANCE_315K
	Ps4000aResistance1100k     ConnectProbeRange = C.PICO_PS4000A_RESISTANCE_1100K
	Ps4000aResistance10m       ConnectProbeRange = C.PICO_PS4000A_RESISTANCE_10M
	Ps4000aMaxResistanceRanges ConnectProbeRange = C.PICO_PS4000A_MAX_RESISTANCE_RANGES
	Ps4000aResistanceAdcvFlag  ConnectProbeRange = C.PICO_PS4000A_RESISTANCE_ADCV_FLAG

	ConnectProbeOff ConnectProbeRange = C.PICO_CONNECT_PROBE_OFF

	D9Bnc10mv      ConnectProbeRange = C.PICO_D9_BNC_10MV
	D9Bnc20mv      ConnectProbeRange = C.PICO_D9_BNC_20MV
	D9Bnc50mv      ConnectProbeRange = C.PICO_D9_BNC_50MV
	D9Bnc100mv     ConnectProbeRange = C.PICO_D9_BNC_100MV
	D9Bnc200mv     ConnectProbeRange = C.PICO_D9_BNC_200MV
	D9Bnc500mv     ConnectProbeRange = C.PICO_D9_BNC_500MV
	D9Bnc1v        ConnectProbeRange = C.PICO_D9_BNC_1V
	D9Bnc2v        ConnectProbeRange = C.PICO_D9_BNC_2V
	D9Bnc5v        ConnectProbeRange = C.PICO_D9_BNC_5V
	D9Bnc10v       ConnectProbeRange = C.PICO_D9_BNC_10V
	D9Bnc20v       ConnectProbeRange = C.PICO_D9_BNC_20V
	D9Bnc50v       ConnectProbeRange = C.PICO_D9_BNC_50V
	D9Bnc100v      ConnectProbeRange = C.PICO_D9_BNC_100V
	D9Bnc200v      ConnectProbeRange = C.PICO_D9_BNC_200V
	MaxD9BncRanges ConnectProbeRange = C.PICO_MAX_D9_BNC_RANGES

	D9_2xBnc10mv      ConnectProbeRange = C.PICO_D9_2X_BNC_10MV
	D9_2xBnc20mv      ConnectProbeRange = C.PICO_D9_2X_BNC_20MV
	D9_2xBnc50mv      ConnectProbeRange = C.PICO_D9_2X_BNC_50MV
	D9_2xBnc100mv     ConnectProbeRange = C.PICO_D9_2X_BNC_100MV
	D9_2xBnc200mv     ConnectProbeRange = C.PICO_D9_2X_BNC_200MV
	D9_2xBnc500mv     ConnectProbeRange = C.PICO_D9_2X_BNC_500MV
	D9_2xBnc1v        ConnectProbeRange = C.PICO_D9_2X_BNC_1V
	D9_2xBnc2v        ConnectProbeRange = C.PICO_D9_2X_BNC_2V
	D9_2xBnc5v        ConnectProbeRange = C.PICO_D9_2X_BNC_5V
	D9_2xBnc10v       ConnectProbeRange = C.PICO_D9_2X_BNC_10V
	D9_2xBnc20v       ConnectProbeRange = C.PICO_D9_2X_BNC_20V
	D9_2xBnc50v       ConnectProbeRange = C.PICO_D9_2X_BNC_50V
	D9_2xBnc100v      ConnectProbeRange = C.PICO_D9_2X_BNC_100V
	D9_2xBnc200v      ConnectProbeRange = C.PICO_D9_2X_BNC_200V
	MaxD9_2xBncRanges ConnectProbeRange = C.PICO_MAX_D9_2X_BNC_RANGES

	Differential10mv      ConnectProbeRange = C.PICO_DIFFERENTIAL_10MV
	Differential20mv      ConnectProbeRange = C.PICO_DIFFERENTIAL_20MV
	Differential50mv      ConnectProbeRange = C.PICO_DIFFERENTIAL_50MV
	Differential100mv     ConnectProbeRange = C.PICO_DIFFERENTIAL_100MV
	Differential200mv     ConnectProbeRange = C.PICO_DIFFERENTIAL_200MV
	Differential500mv     ConnectProbeRange = C.PICO_DIFFERENTIAL_500MV
	Differential1v        ConnectProbeRange = C.PICO_DIFFERENTIAL_1V
	Differential2v        ConnectProbeRange = C.PICO_DIFFERENTIAL_2V
	Differential5v        ConnectProbeRange = C.PICO_DIFFERENTIAL_5V
	Differential10v       ConnectProbeRange = C.PICO_DIFFERENTIAL_10V
	Differential20v       ConnectProbeRange = C.PICO_DIFFERENTIAL_20V
	Differential50v       ConnectProbeRange = C.PICO_DIFFERENTIAL_50V
	Differential100v      ConnectProbeRange = C.PICO_DIFFERENTIAL_100V
	Differential200v      ConnectProbeRange = C.PICO_DIFFERENTIAL_200V
	MaxDifferentialRanges ConnectProbeRange = C.PICO_MAX_DIFFERENTIAL_RANGES

	CurrentClamp200a2ka1a        ConnectProbeRange = C.PICO_CURRENT_CLAMP_200A_2kA_1A
	CurrentClamp200a2ka2a        ConnectProbeRange = C.PICO_CURRENT_CLAMP_200A_2kA_2A
	CurrentClamp200a2ka5a        ConnectProbeRange = C.PICO_CURRENT_CLAMP_200A_2kA_5A
	CurrentClamp200a2ka10a       ConnectProbeRange = C.PICO_CURRENT_CLAMP_200A_2kA_10A
	CurrentClamp200a2ka20a       ConnectProbeRange = C.PICO_CURRENT_CLAMP_200A_2kA_20A
	CurrentClamp200a2ka50a       ConnectProbeRange = C.PICO_CURRENT_CLAMP_200A_2kA_50A
	CurrentClamp200a2ka100a      ConnectProbeRange = C.PICO_CURRENT_CLAMP_200A_2kA_100A
	CurrentClamp200a2ka200a      ConnectProbeRange = C.PICO_CURRENT_CLAMP_200A_2kA_200A
	CurrentClamp200a2ka500a      ConnectProbeRange = C.PICO_CURRENT_CLAMP_200A_2kA_500A
	CurrentClamp200a2ka1000a     ConnectProbeRange = C.PICO_CURRENT_CLAMP_200A_2kA_1000A
	CurrentClamp200a2ka2000a     ConnectProbeRange = C.PICO_CURRENT_CLAMP_200A_2kA_2000A
	MaxCurrentClamp200a2kaRanges ConnectProbeRange = C.PICO_MAX_CURRENT_CLAMP_200A_2kA_RANGES

	CurrentClamp40a100ma     ConnectProbeRange = C.PICO_CURRENT_CLAMP_40A_100mA
	CurrentClamp40a200ma     ConnectProbeRange = C.PICO_CURRENT_CLAMP_40A_200mA
	CurrentClamp40a500ma     ConnectProbeRange = C.PICO_CURRENT_CLAMP_40A_500mA
	CurrentClamp40a1a        ConnectProbeRange = C.PICO_CURRENT_CLAMP_40A_1A
	CurrentClamp40a2a        ConnectProbeRange = C.PICO_CURRENT_CLAMP_40A_2A
	CurrentClamp40a5a        ConnectProbeRange = C.PICO_CURRENT_CLAMP_40A_5A
	CurrentClamp40a10a       ConnectProbeRange = C.PICO_CURRENT_CLAMP_40A_10A
	CurrentClamp40a20a       ConnectProbeRange = C.PICO_CURRENT_CLAMP_40A_20A
	CurrentClamp40a40a       ConnectProbeRange = C.PICO_CURRENT_CLAMP_40A_40A
	MaxCurrentClamp40aRanges ConnectProbeRange = C.PICO_MAX_CURRENT_CLAMP_40A_RANGES

	Pico1kv2_5v  ConnectProbeRange = C.PICO_1KV_2_5V
	Pico1kv5v    ConnectProbeRange = C.PICO_1KV_5V
	Pico1kv12_5v ConnectProbeRange = C.PICO_1KV_12_5V
	Pico1kv25v   ConnectProbeRange = C.PICO_1KV_25V
	Pico1kv50v   ConnectProbeRange = C.PICO_1KV_50V
	Pico1kv125v  ConnectProbeRange = C.PICO_1KV_125V
	Pico1kv250v  ConnectProbeRange = C.PICO_1KV_250V
	Pico1kv500v  ConnectProbeRange = C.PICO_1KV_500V
	Pico1kv1000v ConnectProbeRange = C.PICO_1KV_1000V
	Max1kvRanges ConnectProbeRange = C.PICO_MAX_1KV_RANGES

	CurrentClamp2000arms10a    ConnectProbeRange = C.PICO_CURRENT_CLAMP_2000ARMS_10A
	CurrentClamp2000arms20a    ConnectProbeRange = C.PICO_CURRENT_CLAMP_2000ARMS_20A
	CurrentClamp2000arms50a    ConnectProbeRange = C.PICO_CURRENT_CLAMP_2000ARMS_50A
	CurrentClamp2000arms100a   ConnectProbeRange = C.PICO_CURRENT_CLAMP_2000ARMS_100A
	CurrentClamp2000arms200a   ConnectProbeRange = C.PICO_CURRENT_CLAMP_2000ARMS_200A
	CurrentClamp2000arms500a   ConnectProbeRange = C.PICO_CURRENT_CLAMP_2000ARMS_500A
	CurrentClamp2000arms1000a  ConnectProbeRange = C.PICO_CURRENT_CLAMP_2000ARMS_1000A
	CurrentClamp2000arms2000a  ConnectProbeRange = C.PICO_CURRENT_CLAMP_2000ARMS_2000A
	CurrentClamp2000arms5000a  ConnectProbeRange = C.PICO_CURRENT_CLAMP_2000ARMS_5000A
	CurrentClamp2000armsRanges ConnectProbeRange = C.PICO_CURRENT_CLAMP_2000ARMS_RANGES

	ResistanceLeadNeg5To20ohm     ConnectProbeRange = C.PICO_RESISTANCE_LEAD_NEG5_TO_20OHM
	ResistanceLeadNeg50To200ohm   ConnectProbeRange = C.PICO_RESISTANCE_LEAD_NEG50_TO_200OHM
	ResistanceLeadNeg500To2kohm   ConnectProbeRange = C.PICO_RESISTANCE_LEAD_NEG500_TO_2KOHM
	ResistanceLeadNeg5kTo20kohm   ConnectProbeRange = C.PICO_RESISTANCE_LEAD_NEG5K_TO_20KOHM
	ResistanceLeadNeg50kTo200kohm ConnectProbeRange = C.PICO_RESISTANCE_LEAD_NEG50K_TO_LEAD_200KOHM
	ResistanceLeadNeg500kTo2mohm  ConnectProbeRange = C.PICO_RESISTANCE_LEAD_NEG500K_TO_LEAD_2MOHM
	ResistanceLeadDiodeTest       ConnectProbeRange = C.PICO_RESISTANCE_LEAD_DIODE_TEST
	MaxResistanceLeadRanges       ConnectProbeRange = C.PICO_MAX_RESISTANCE_LEAD_RANGES

	HtNeg3To5kv           ConnectProbeRange = C.PICO_HT_NEG3_TO_5KV
	HtNeg3To10kv          ConnectProbeRange = C.PICO_HT_NEG3_TO_10KV
	HtNeg5To20kv          ConnectProbeRange = C.PICO_HT_NEG5_TO_20KV
	HtNeg5To50kv          ConnectProbeRange = C.PICO_HT_NEG5_TO_50KV
	HtNeg5To100kv         ConnectProbeRange = C.PICO_HT_NEG5_TO_100KV
	HtNeg3To5kvInverted   ConnectProbeRange = C.PICO_HT_NEG3_TO_5KV_INVERTED
	HtNeg3To10kvInverted  ConnectProbeRange = C.PICO_HT_NEG3_TO_10KV_INVERTED
	HtNeg5To20kvInverted  ConnectProbeRange = C.PICO_HT_NEG5_TO_20KV_INVERTED
	HtNeg5To50kvInverted  ConnectProbeRange = C.PICO_HT_NEG5_TO_50KV_INVERTED
	HtNeg5To100kvInverted ConnectProbeRange = C.PICO_HT_NEG5_TO_100KV_INVERTED
	MaxHtRanges           ConnectProbeRange = C.PICO_MAX_HT_RANGES

	TemperatureNeg50To150degC ConnectProbeRange = C.PICO_TEMPERATURE_NEG50_TO_150DEGC
	TemperatureNeg4To125degC  ConnectProbeRange = C.PICO_TEMPERATURE_NEG4_TO_125DEGC

	PressureSensorNeg100000To150000Pascals  ConnectProbeRange = C.PICO_PRESSURE_SENSOR_NEG100000_TO_150000_PASCALS
	PressureSensorNeg100000To400000Pascals  ConnectProbeRange = C.PICO_PRESSURE_SENSOR_NEG100000_TO_400000_PASCALS
	PressureSensorNeg200000To800000Pascals  ConnectProbeRange = C.PICO_PRESSURE_SENSOR_NEG200000_TO_800000_PASCALS
	PressureSensorNeg400000To1600000Pascals ConnectProbeRange = C.PICO_PRESSURE_SENSOR_NEG400000_TO_1600000_PASCALS
	PressureSensorNeg400000To3400000Pascals ConnectProbeRange = C.PICO_PRESSURE_SENSOR_NEG400000_TO_3400000_PASCALS
	PressureSensorNeg150000To1350000Pascals ConnectProbeRange = C.PICO_PRESSURE_SENSOR_NEG150000_TO_1350000_PASCALS

	CurrentClamp100a2_5a ConnectProbeRange = C.PICO_CURRENT_CLAMP_100A_2_5A
	CurrentClamp100a5a   ConnectProbeRange = C.PICO_CURRENT_CLAMP_100A_5A
	CurrentClamp100a10a  ConnectProbeRange = C.PICO_CURRENT_CLAMP_100A_10A
	CurrentClamp100a25a  ConnectProbeRange = C.PICO_CURRENT_CLAMP_100A_25A
	CurrentClamp100a50a  ConnectProbeRange = C.PICO_CURRENT_CLAMP_100A_50A
	CurrentClamp100a100a ConnectProbeRange = C.PICO_CURRENT_CLAMP_100A_100A

	CurrentClamp60a2a  ConnectProbeRange = C.PICO_CURRENT_CLAMP_60A_2A
	CurrentClamp60a5a  ConnectProbeRange = C.PICO_CURRENT_CLAMP_60A_5A
	CurrentClamp60a10a ConnectProbeRange = C.PICO_CURRENT_CLAMP_60A_10A
	CurrentClamp60a20a ConnectProbeRange = C.PICO_CURRENT_CLAMP_60A_20A
	CurrentClamp60a50a ConnectProbeRange = C.PICO_CURRENT_CLAMP_60A_50A
	CurrentClamp60a60a ConnectProbeRange = C.PICO_CURRENT_CLAMP_60A_60A

	OpticalSensor10v ConnectProbeRange = C.PICO_OPTICAL_SENSOR_10V

	CurrentClamp60aV2_0_5a ConnectProbeRange = C.PICO_CURRENT_CLAMP_60A_V2_0_5A
	CurrentClamp60aV2_1a   ConnectProbeRange = C.PICO_CURRENT_CLAMP_60A_V2_1A
	CurrentClamp60aV2_2a   ConnectProbeRange = C.PICO_CURRENT_CLAMP_60A_V2_2A
	CurrentClamp60aV2_5a   ConnectProbeRange = C.PICO_CURRENT_CLAMP_60A_V2_5A
	CurrentClamp60aV2_10a  ConnectProbeRange = C.PICO_CURRENT_CLAMP_60A_V2_10A
	CurrentClamp60aV2_20a  ConnectProbeRange = C.PICO_CURRENT_CLAMP_60A_V2_20A
	CurrentClamp60aV2_50a  ConnectProbeRange = C.PICO_CURRENT_CLAMP_60A_V2_50A
	CurrentClamp60aV2_60a  ConnectProbeRange = C.PICO_CURRENT_CLAMP_60A_V2_60A

	HighVoltageDifferential5v    ConnectProbeRange = C.PICO_HIGH_VOLTAGE_DIFFERENTIAL_5V
	HighVoltageDifferential10v   ConnectProbeRange = C.PICO_HIGH_VOLTAGE_DIFFERENTIAL_10V
	HighVoltageDifferential20v   ConnectProbeRange = C.PICO_HIGH_VOLTAGE_DIFFERENTIAL_20V
	HighVoltageDifferential50v   ConnectProbeRange = C.PICO_HIGH_VOLTAGE_DIFFERENTIAL_50V
	HighVoltageDifferential100v  ConnectProbeRange = C.PICO_HIGH_VOLTAGE_DIFFERENTIAL_100V
	HighVoltageDifferential200v  ConnectProbeRange = C.PICO_HIGH_VOLTAGE_DIFFERENTIAL_200V
	HighVoltageDifferential500v  ConnectProbeRange = C.PICO_HIGH_VOLTAGE_DIFFERENTIAL_500V
	HighVoltageDifferential1000v ConnectProbeRange = C.PICO_HIGH_VOLTAGE_DIFFERENTIAL_1000V
	HighVoltageDifferential1400v ConnectProbeRange = C.PICO_HIGH_VOLTAGE_DIFFERENTIAL_1400V

	X10ActiveProbe100mv ConnectProbeRange = C.PICO_X10_ACTIVE_PROBE_100MV
	X10ActiveProbe200mv ConnectProbeRange = C.PICO_X10_ACTIVE_PROBE_200MV
	X10ActiveProbe500mv ConnectProbeRange = C.PICO_X10_ACTIVE_PROBE_500MV
	X10ActiveProbe1v    ConnectProbeRange = C.PICO_X10_ACTIVE_PROBE_1V
	X10ActiveProbe2v    ConnectProbeRange = C.PICO_X10_ACTIVE_PROBE_2V
	X10ActiveProbe5v    ConnectProbeRange = C.PICO_X10_ACTIVE_PROBE_5V

	TriaxialAccelerometer5g       ConnectProbeRange = C.PICO_TRIAXIAL_ACCELEROMETER_5G
	TriaxialAccelerometer9g       ConnectProbeRange = C.PICO_TRIAXIAL_ACCELEROMETER_9G
	SingleaxisAccelerometer150mg  ConnectProbeRange = C.PICO_SINGLEAXIS_ACCELEROMETER_150MG
	SingleaxisAccelerometer300mg  ConnectProbeRange = C.PICO_SINGLEAXIS_ACCELEROMETER_300MG
	SingleaxisAccelerometer600mg  ConnectProbeRange = C.PICO_SINGLEAXIS_ACCELEROMETER_600MG
	SingleaxisAccelerometer1500mg ConnectProbeRange = C.PICO_SINGLEAXIS_ACCELEROMETER_1500MG
	SingleaxisAccelerometer3g     ConnectProbeRange = C.PICO_SINGLEAXIS_ACCELEROMETER_3G
	SingleaxisAccelerometer6g     ConnectProbeRange = C.PICO_SINGLEAXIS_ACCELEROMETER_6G
	SingleaxisAccelerometer9g     ConnectProbeRange = C.PICO_SINGLEAXIS_ACCELEROMETER_9G

	Microphone800MilliPascal   ConnectProbeRange = C.PICO_MICROPHONE_800_MILLI_PASCAL
	Microphone1600MilliPascal  ConnectProbeRange = C.PICO_MICROPHONE_1600_MILLI_PASCAL
	Microphone3200MilliPascal  ConnectProbeRange = C.PICO_MICROPHONE_3200_MILLI_PASCAL
	Microphone8000MilliPascal  ConnectProbeRange = C.PICO_MICROPHONE_8000_MILLI_PASCAL
	Microphone16000MilliPascal ConnectProbeRange = C.PICO_MICROPHONE_16000_MILLI_PASCAL
)

type ProbeRangeInfo int32

const (
	ProbeNoneNv ProbeRangeInfo = C.PICO_PROBE_NONE_NV
	X1ProbeNv   ProbeRangeInfo = C.PICO_X1_PROBE_NV
	X10ProbeNv  ProbeRangeInfo = C.PICO_X10_PROBE_NV
)
