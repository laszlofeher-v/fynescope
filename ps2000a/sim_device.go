//go:build sim

package ps2000a

import (
	"fynescope/demo"
)

// SimDeviceConfig holds device-specific parameters and behaviors for the simulator.
type SimDeviceConfig struct {
	MaxSamples            int32
	MaxMemorySamples      int32
	MinRange              int32
	MaxRange              int32
	AwgBandwidth          float64
	AwgUpdateRate         float64
	AwgBufferSize         int32
	TimebaseToNs          func(timebase uint32) float32
	GetAnalogueOffsetVals func(rangeEnum int32, coupling Coupling) (maxOffset, minOffset float32)
}

// defaultTimebaseToNs matches the logic that was originally hardcoded.
func defaultTimebaseToNs(timebase uint32) float32 {
	return float32(simTimebaseToNs(timebase))
}

// defaultGetAnalogueOffsetVals returns the analogue offset limits for the
// simulator. The real ps2000aGetAnalogueOffset SDK call is not available in
// sim mode, so we return the fixed ±20 V limits that simGetAnalogueOffset
// would produce for a valid handle.
func defaultGetAnalogueOffsetVals(rangeEnum int32, coupling Coupling) (maxOffset, minOffset float32) {
	return 20.0, -20.0
}

var defaultSimDeviceConfig = &SimDeviceConfig{
	MaxSamples:            1000000,
	MaxMemorySamples:      64000000,
	MinRange:              0,     // 10mV
	MaxRange:              11,    // 50V
	AwgBandwidth:          1e6,   // 1 MHz
	AwgUpdateRate:         20e6,  // 20 MS/s
	AwgBufferSize:         32768, // 32 kS
	TimebaseToNs:          defaultTimebaseToNs,
	GetAnalogueOffsetVals: defaultGetAnalogueOffsetVals,
}

var simDeviceRegistry = map[string]*SimDeviceConfig{
	"DEFAULT": defaultSimDeviceConfig,
	"2103SIM": {
		MaxSamples:            8000,
		MaxMemorySamples:      8000,
		MinRange:              2,
		MaxRange:              10,
		AwgBandwidth:          100e3,
		AwgUpdateRate:         1.548e6,
		AwgBufferSize:         4096,
		TimebaseToNs:          defaultTimebaseToNs,
		GetAnalogueOffsetVals: defaultGetAnalogueOffsetVals,
	},
	"2104SIM": {
		MaxSamples:            8000,
		MaxMemorySamples:      8000,
		MinRange:              2,
		MaxRange:              10,
		AwgBandwidth:          100e3,
		AwgUpdateRate:         1.548e6,
		AwgBufferSize:         4096,
		TimebaseToNs:          defaultTimebaseToNs,
		GetAnalogueOffsetVals: defaultGetAnalogueOffsetVals,
	},
	"2203SIM": {
		MaxSamples:            8000,
		MaxMemorySamples:      8000,
		MinRange:              2,
		MaxRange:              10,
		AwgBandwidth:          100e3,
		AwgUpdateRate:         1.548e6,
		AwgBufferSize:         4096,
		TimebaseToNs:          defaultTimebaseToNs,
		GetAnalogueOffsetVals: defaultGetAnalogueOffsetVals,
	},
	"2204SIM": {
		MaxSamples:            8000,
		MaxMemorySamples:      8000,
		MinRange:              2,
		MaxRange:              10,
		AwgBandwidth:          100e3,
		AwgUpdateRate:         1.548e6,
		AwgBufferSize:         4096,
		TimebaseToNs:          defaultTimebaseToNs,
		GetAnalogueOffsetVals: defaultGetAnalogueOffsetVals,
	},
	"2205SIM": {
		MaxSamples:            16000,
		MaxMemorySamples:      16000,
		MinRange:              2,
		MaxRange:              10,
		AwgBandwidth:          100e3,
		AwgUpdateRate:         1.548e6,
		AwgBufferSize:         4096,
		TimebaseToNs:          defaultTimebaseToNs,
		GetAnalogueOffsetVals: defaultGetAnalogueOffsetVals,
	},
	"2204ASIM": {
		MaxSamples:            8000,
		MaxMemorySamples:      8000,
		MinRange:              2,  // 50mV
		MaxRange:              10, // 20V
		AwgBandwidth:          100e3,
		AwgUpdateRate:         1.548e6,
		AwgBufferSize:         4096,
		TimebaseToNs:          defaultTimebaseToNs,
		GetAnalogueOffsetVals: defaultGetAnalogueOffsetVals,
	},
	"2205ASIM": {
		MaxSamples:            16000,
		MaxMemorySamples:      16000,
		MinRange:              2,  // 50mV
		MaxRange:              10, // 20V
		AwgBandwidth:          100e3,
		AwgUpdateRate:         1.548e6,
		AwgBufferSize:         4096,
		TimebaseToNs:          defaultTimebaseToNs,
		GetAnalogueOffsetVals: defaultGetAnalogueOffsetVals,
	},
	"2206BSIM": {
		MaxSamples:            32000000,
		MaxMemorySamples:      32000000,
		MinRange:              1,  // 20mV
		MaxRange:              10, // 20V
		AwgBandwidth:          1e6,
		AwgUpdateRate:         20e6,
		AwgBufferSize:         32768,
		TimebaseToNs:          defaultTimebaseToNs,
		GetAnalogueOffsetVals: defaultGetAnalogueOffsetVals,
	},
	"2207BSIM": {
		MaxSamples:            64000000,
		MaxMemorySamples:      64000000,
		MinRange:              1,  // 20mV
		MaxRange:              10, // 20V
		AwgBandwidth:          1e6,
		AwgUpdateRate:         20e6,
		AwgBufferSize:         32768,
		TimebaseToNs:          defaultTimebaseToNs,
		GetAnalogueOffsetVals: defaultGetAnalogueOffsetVals,
	},
	"2208BSIM": {
		MaxSamples:            128000000,
		MaxMemorySamples:      128000000,
		MinRange:              1,  // 20mV
		MaxRange:              10, // 20V
		AwgBandwidth:          1e6,
		AwgUpdateRate:         20e6,
		AwgBufferSize:         32768,
		TimebaseToNs:          defaultTimebaseToNs,
		GetAnalogueOffsetVals: defaultGetAnalogueOffsetVals,
	},
	"2405ASIM": {
		MaxSamples:            48000,
		MaxMemorySamples:      48000,
		MinRange:              1,  // 20mV
		MaxRange:              10, // 20V
		AwgBandwidth:          1e6,
		AwgUpdateRate:         20e6,
		AwgBufferSize:         8192,
		TimebaseToNs:          defaultTimebaseToNs,
		GetAnalogueOffsetVals: defaultGetAnalogueOffsetVals,
	},
	"2406BSIM": {
		MaxSamples:            32000000,
		MaxMemorySamples:      32000000,
		MinRange:              1,  // 20mV
		MaxRange:              10, // 20V
		AwgBandwidth:          1e6,
		AwgUpdateRate:         20e6,
		AwgBufferSize:         32768,
		TimebaseToNs:          defaultTimebaseToNs,
		GetAnalogueOffsetVals: defaultGetAnalogueOffsetVals,
	},
	"2407BSIM": {
		MaxSamples:            64000000,
		MaxMemorySamples:      64000000,
		MinRange:              1,  // 20mV
		MaxRange:              10, // 20V
		AwgBandwidth:          1e6,
		AwgUpdateRate:         20e6,
		AwgBufferSize:         32768,
		TimebaseToNs:          defaultTimebaseToNs,
		GetAnalogueOffsetVals: defaultGetAnalogueOffsetVals,
	},
	"2408BSIM": {
		MaxSamples:            128000000,
		MaxMemorySamples:      128000000,
		MinRange:              1,  // 20mV
		MaxRange:              10, // 20V
		AwgBandwidth:          1e6,
		AwgUpdateRate:         20e6,
		AwgBufferSize:         32768,
		TimebaseToNs:          defaultTimebaseToNs,
		GetAnalogueOffsetVals: defaultGetAnalogueOffsetVals,
	},
	"2205AMSOSIM": {
		MaxSamples:            48000,
		MaxMemorySamples:      48000,
		MinRange:              1,  // 20mV
		MaxRange:              10, // 20V
		AwgBandwidth:          1e6,
		AwgUpdateRate:         20e6,
		AwgBufferSize:         8192,
		TimebaseToNs:          defaultTimebaseToNs,
		GetAnalogueOffsetVals: defaultGetAnalogueOffsetVals,
	},
	"2206BMSOSIM": {
		MaxSamples:            32000000,
		MaxMemorySamples:      32000000,
		MinRange:              1,  // 20mV
		MaxRange:              10, // 20V
		AwgBandwidth:          1e6,
		AwgUpdateRate:         20e6,
		AwgBufferSize:         32768,
		TimebaseToNs:          defaultTimebaseToNs,
		GetAnalogueOffsetVals: defaultGetAnalogueOffsetVals,
	},
	"2207BMSOSIM": {
		MaxSamples:            64000000,
		MaxMemorySamples:      64000000,
		MinRange:              1,  // 20mV
		MaxRange:              10, // 20V
		AwgBandwidth:          1e6,
		AwgUpdateRate:         20e6,
		AwgBufferSize:         32768,
		TimebaseToNs:          defaultTimebaseToNs,
		GetAnalogueOffsetVals: defaultGetAnalogueOffsetVals,
	},
	"2208BMSOSIM": {
		MaxSamples:            128000000,
		MaxMemorySamples:      128000000,
		MinRange:              1,  // 20mV
		MaxRange:              10, // 20V
		AwgBandwidth:          1e6,
		AwgUpdateRate:         20e6,
		AwgBufferSize:         32768,
		TimebaseToNs:          defaultTimebaseToNs,
		GetAnalogueOffsetVals: defaultGetAnalogueOffsetVals,
	},
}

// GetSimConfig returns the simulator config based on the scope simulated variant info.
func GetSimConfig() *SimDeviceConfig {
	variant := demo.ScopeSimVariantInfo
	if cfg, ok := simDeviceRegistry[variant]; ok {
		return cfg
	}
	return defaultSimDeviceConfig
}
