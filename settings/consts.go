package settings

const (
	UnitHz  = "Hz"
	UnitKHz = "kHz"
	UnitMHz = "MHz"

	ModeVoltage = "Voltage"
	ModeDB      = "dB"

	WindowBartlettHann    = "BartlettHann"
	WindowBlackman        = "Blackman"
	WindowBlackmanHarris  = "BlackmanHarris"
	WindowBlackmanNuttall = "BlackmanNuttall"
	WindowFlatTop         = "FlatTop"
	WindowHamming         = "Hamming"
	WindowHann            = "Hann"
	WindowLanczos         = "Lanczos"
	WindowNuttall         = "Nuttall"
	WindowTriangular      = "Triangular"
	WindowRectangular     = "Rectangular"
)

// Trigger type name constants (stored as strings in settings YAML).
const (
	TriggerTypeSimple   = "Simple"
	TriggerTypeAdvanced = "Advanced"
	TriggerTypeWindow   = "Window"
	TriggerTypeInterval = "Interval"
	TriggerTypeComplex  = "Complex"
)

// Trigger mode name constants (stored as strings in settings YAML).
const (
	TriggerModeAuto   = "Auto"
	TriggerModeETS    = "ETS"
	TriggerModeRepeat = "Repeat"
	TriggerModeSingle = "Single"
)

// Screen-size preset strings.
const (
	ScreenSize1920x1080 = "1920x1080"
	ScreenSize1366x768  = "1366x768"
	ScreenSize1280x720  = "1280x720"
	ScreenSize1024x768  = "1024x768"
)

// RlcFilterTypeDisabled is the sentinel value for a disabled RLC filter.
const RlcFilterTypeDisabled = "Disabled"
