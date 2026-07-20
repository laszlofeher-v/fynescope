package settings

import (
	"crypto/sha256"
	"fmt"
	"fynescope/genericps"
	"image/color"
	"log/slog"
	"os"

	"gopkg.in/yaml.v3"
)

const warningString = "# If you edit this file the checksum will fail.\n# As a result, the system will discard your edits,\n# and revert to safe default settings.\n"

type ChannelColorIndexType int

const (
	DarkChannel ChannelColorIndexType = iota
	LightChannel
)

type FvMode int

const (
	FvDisabled FvMode = iota
	FvArgument        // X
	FvValue           // Y
)

type (
	InterpolationType int
	WindowSettings    struct {
		Width            float32 `yaml:"width"`
		Height           float32 `yaml:"height"`
		Function         int     `yaml:"function"`
		LastDispFunction int     `yaml:"lastdispfunction"`
		Fullscreen       bool    `yaml:"fullscreen"`
		LeftControl      bool    `yaml:"leftsignalscreen"`
		FilterActiveTab  int     `yaml:"filteractivetab"`
		SimGenActiveTab  int     `yaml:"simgenactivetab"`
	}
	ChTriggerSettings struct {
		Type              string                       `yaml:"type"` // "Simple", "Advanced", "Window", "Complex", "Interval", "Pulse Width"
		Condition         genericps.TriggerRespBase    `yaml:"condition"`
		TriggerDirection  genericps.ThresholdDirection `yaml:"triggerdirection"`
		ThresholdMode     genericps.ThresholdModeId    `yaml:"thresholdmode"`
		Mv                int32                        `yaml:"mv"`
		Hysteresis        int32                        `yaml:"hysteresis"`
		LowerMv           int32                        `yaml:"lowermv"`
		LowerHysteresis   int32                        `yaml:"lowerhysteresis"`
		IntervalType      genericps.PulseWidthType     `yaml:"intervaltype"`
		IntervalTimeLower float64                      `yaml:"intervaltimelower"`
		IntervalTimeUpper float64                      `yaml:"intervaltimeupper"`
	}

	ChSettings struct {
		ID                 genericps.ChannelId   `yaml:"id"`
		Inverted           bool                  `yaml:"inverted"`
		X10                bool                  `yaml:"x10"`
		DisplayVOffset     int                   `yaml:"displayvoffset"`
		DftDisplayVOffset  int                   `yaml:"dftdisplayvoffset"`
		Col                [2]color.NRGBA        `yaml:"color"`
		VRange             genericps.RangeEnum   `yaml:"range"`
		CoupleType         genericps.Coupling    `yaml:"couple"`
		Enabled            bool                  `yaml:"enabled"`
		Offset             float32               `yaml:"offset"`
		FvMode             FvMode                `yaml:"fvmode"`
		TriggerSource      bool                  `yaml:"triggersource"`
		Persistence        bool                  `yaml:"persistence"`
		DftPersistence     bool                  `yaml:"dftpersistence"`
		Trigger            ChTriggerSettings     `yaml:"triggersettings"`
		FfAmplitudeEnabled bool                  `yaml:"ffamplitudeenabled"`
		FfPhaseEnabled     bool                  `yaml:"ffphaseenabled"`
		FfDisplayVOffset   int                   `yaml:"ffdisplayvoffset"`
		RlcFilter          RlcFilterSettings     `yaml:"rlcfilter"`
		DigitalFilter      DigitalFilterSettings `yaml:"digitalfilter"`
	}
	DigitalFilterSettings struct {
		LowpassEnabled  bool    `yaml:"lowpassenabled"`
		LowpassFc       float64 `yaml:"lowpassfc"`
		HighpassEnabled bool    `yaml:"highpassenabled"`
		HighpassFc      float64 `yaml:"highpassfc"`
		BandpassEnabled bool    `yaml:"bandpassenabled"`
		BandpassFc1     float64 `yaml:"bandpassfc1"`
		BandpassFc2     float64 `yaml:"bandpassfc2"`
		BandstopEnabled bool    `yaml:"bandstopenabled"`
		BandstopFc1     float64 `yaml:"bandstopfc1"`
		BandstopFc2     float64 `yaml:"bandstopfc2"`
	}
	RlcFilterSettings struct {
		GeneratorSource genericps.ChannelId `yaml:"generatorsource"`
		Enabled         bool                `yaml:"enabled"`
		Type            string              `yaml:"type"` // "Lowpass RC", etc.
		R               float64             `yaml:"r"`
		RUnit           string              `yaml:"runit"`
		L               float64             `yaml:"l"`
		LUnit           string              `yaml:"lunit"`
		C               float64             `yaml:"c"`
		CUnit           string              `yaml:"cunit"`
	}
	TimeSettings struct {
		TimeDiv           string            `yaml:"timediv"`
		Unit              string            `yaml:"unit"`
		Interpolation     InterpolationType `yaml:"interpolation"`
		TriggerTimeOffset float64           `yaml:"triggertimeoffset"` // trigger distance in sec
		SampleRate        string            `yaml:"samplerate"`
		SampleRateUnit    string            `yaml:"samplerateunits"`
	}
	TriggerSettings struct {
		Mode            string `yaml:"triggermode"`
		Type            string `yaml:"triggertype"`
		CalculationMode int    `yaml:"calculationmode"`
		ComplexEnabled  bool   `yaml:"complexenabled"`
	}
	GeneratorSettings struct {
		On                   bool                      `yaml:"on"`
		Digital              bool                      `yaml:"digital"`
		Frequency            float64                   `yaml:"frequency"`
		StartFrequency       float64                   `yaml:"startfrequency"`
		StopFrequency        float64                   `yaml:"stopfrequency"`
		Increment            float64                   `yaml:"increment"`
		Sweep                genericps.SweepTypeEnum   `yaml:"sweep"`
		Dwelltime            float64                   `yaml:"dwelltime"`
		OffsetVoltage        int32                     `yaml:"offsetvoltage"`
		Amplitude            uint32                    `yaml:"amplitude"`
		WaveType             genericps.WaveTypeEnum    `yaml:"wavetype"`
		Operation            genericps.ExtraOperations `yaml:"operation"`
		RaiseFallTimePercent float64                   `yaml:"raisefalltimepercent"`
		TriggerTimeOffset    float64                   `yaml:"triggertimeoffset"`
		NoiseAmplitude       float64                   `yaml:"noise_amplitude"`
		PhaseNoiseDegree     float64                   `yaml:"phase_noise_degree"`
		Phase                float64                   `yaml:"phase"`
		ImpedanceMode        string                    `yaml:"impedance_mode"` // "ohms", "INFinity", "MINimum", "MAXimum"
		ImpedanceOhms        int                       `yaml:"impedance_ohms"` // 1–10000, used when ImpedanceMode == "ohms"
	}
	DftSettings struct {
		MaxFreq        float64 `yaml:"maxfreq"`
		MinFreq        float64 `yaml:"minfreq"`
		Window         string  `yaml:"window"`
		DisplayMode    string  `yaml:"displaymode"`
		Bins           int     `yaml:"bins"`
		SampleRate     string  `yaml:"samplerate"`
		SampleRateUnit string  `yaml:"samplerateunits"`
	}
	FfSettings struct {
		ReferenceChannel  int     `yaml:"referencechannel"`
		MinFreq           float64 `yaml:"minfreq"`
		MaxFreq           float64 `yaml:"maxfreq"`
		DisplayMode       string  `yaml:"displaymode"`
		PtsDec            float64 `yaml:"deltafreq"`
		UseExternalGen    bool    `yaml:"useexternalgen"`
		ExternalGenPort   string  `yaml:"externalgenport"`
		ExternalGenUsbVid string  `yaml:"externalgenusbvid"`
		ExternalGenUsbPid string  `yaml:"externalgenusbpid"`
		DeltaT            float64 `yaml:"deltat"`
		Amplitude         uint32  `yaml:"amplitude"`
	}
	PsSettings struct {
		Theme             ThemeType             `yaml:"theme"`
		ChannelColorIndex ChannelColorIndexType `yaml:"channelcolorindex"`
		Window            WindowSettings        `yaml:"window"`
		ScreenSize        string                `yaml:"screensize,omitempty"`
		Channels          []ChSettings          `yaml:"channels"`
		Trigger           TriggerSettings       `yaml:"trigger"`
		Time              TimeSettings          `yaml:"time"`
		GenPanel          GeneratorSettings     `yaml:"genpanel"`
		SimGenPanel       []GeneratorSettings   `yaml:"simgenpanel"`
		FfGen             GeneratorSettings     `yaml:"ffgen"`  // Simulator and real hw has the same functionality
		ExtGen            [2]GeneratorSettings  `yaml:"extgen"` // External generator explicit settings
		Dft               DftSettings           `yaml:"dft"`
		Ff                FfSettings            `yaml:"ff"`
		StreamEnabled     *bool                 `yaml:"streamenabled,omitempty"`
	}
)

const (
	// Dot no post processing
	Dot InterpolationType = iota
	// Raw horizontal and vertical lines
	Raw
	// Linear straight lines between sample points
	Linear
	// Sinc Sin(x)/x interpolation, needs more samples
	Sinc
)

type ThemeType int

const (
	DarkTheme ThemeType = iota
	LightTheme
)

const (
	defaultRange                = genericps.RangeEnum(8) // 5V
	defaultCouple               = genericps.Coupling(1)  // DC
	defaultTimeUnit             = "ms/div"
	defaultTime                 = "1"
	defaultInterpolation        = Raw
	defaultFrequency            = float64(1000)
	defaultAmplitude            = uint32(1000000)
	defaultRaiseFallTimePercent = float64(1.0)
	defaultTheme                = DarkTheme
	defaultSamplerate           = "1"
	defaultSampleRateUnit       = "ks/s"
)

func NewDefaultSettings() *PsSettings {
	streamDefault := true
	return &PsSettings{
		Window: WindowSettings{Width: 1366, Height: 768, LeftControl: false,
			Function: 0},
		ScreenSize: ScreenSize1920x1080,
		Time: TimeSettings{Unit: defaultTimeUnit, TriggerTimeOffset: 0, TimeDiv: defaultTime,
			Interpolation: Raw, SampleRate: defaultSamplerate, SampleRateUnit: defaultSampleRateUnit},
		Trigger: TriggerSettings{Mode: TriggerModeAuto, Type: TriggerTypeSimple, CalculationMode: 0, ComplexEnabled: false},
		Channels: []ChSettings{
			{ID: genericps.ChA, Col: [2]color.NRGBA{{100, 200, 255, 255},
				{1, 5, 191, 255}}, VRange: defaultRange,
				CoupleType: defaultCouple, Enabled: true, TriggerSource: true,
				Trigger:       ChTriggerSettings{TriggerDirection: genericps.TriggerRising, Mv: 0},
				RlcFilter:     RlcFilterSettings{GeneratorSource: genericps.ChA, Enabled: false, Type: RlcFilterTypeDisabled, R: 1, RUnit: "kΩ", L: 1, LUnit: "mH", C: 1, CUnit: "µF"},
				DigitalFilter: DigitalFilterSettings{LowpassFc: 10000.0, HighpassFc: 100.0, BandpassFc1: 500.0, BandpassFc2: 5000.0, BandstopFc1: 900.0, BandstopFc2: 1100.0}},
			{ID: genericps.ChB, Col: [2]color.NRGBA{{255, 0, 0, 255},
				{255, 0, 0, 255}}, VRange: defaultRange,
				CoupleType: defaultCouple, Enabled: false, TriggerSource: false,
				Trigger:       ChTriggerSettings{TriggerDirection: genericps.TriggerRising, Mv: 0},
				RlcFilter:     RlcFilterSettings{GeneratorSource: genericps.ChB, Enabled: false, Type: RlcFilterTypeDisabled, R: 1, RUnit: "kΩ", L: 1, LUnit: "mH", C: 1, CUnit: "µF"},
				DigitalFilter: DigitalFilterSettings{LowpassFc: 10000.0, HighpassFc: 100.0, BandpassFc1: 500.0, BandpassFc2: 5000.0, BandstopFc1: 900.0, BandstopFc2: 1100.0}},
			{ID: genericps.ChC, Col: [2]color.NRGBA{{0, 255, 0, 255},
				{0, 135, 0, 255}}, VRange: defaultRange,
				CoupleType: defaultCouple, Enabled: false, TriggerSource: false,
				Trigger:       ChTriggerSettings{TriggerDirection: genericps.TriggerRising, Mv: 0},
				RlcFilter:     RlcFilterSettings{GeneratorSource: genericps.ChC, Enabled: false, Type: RlcFilterTypeDisabled, R: 1, RUnit: "kΩ", L: 1, LUnit: "mH", C: 1, CUnit: "µF"},
				DigitalFilter: DigitalFilterSettings{LowpassFc: 10000.0, HighpassFc: 100.0, BandpassFc1: 500.0, BandpassFc2: 5000.0, BandstopFc1: 900.0, BandstopFc2: 1100.0}},
			{ID: genericps.ChD, Col: [2]color.NRGBA{{255, 255, 0, 255},
				{100, 100, 0, 255}}, VRange: defaultRange,
				CoupleType: defaultCouple, Enabled: false, TriggerSource: false,
				Trigger:       ChTriggerSettings{TriggerDirection: genericps.TriggerRising, Mv: 0},
				RlcFilter:     RlcFilterSettings{GeneratorSource: genericps.ChD, Enabled: false, Type: RlcFilterTypeDisabled, R: 1, RUnit: "kΩ", L: 1, LUnit: "mH", C: 1, CUnit: "µF"},
				DigitalFilter: DigitalFilterSettings{LowpassFc: 10000.0, HighpassFc: 100.0, BandpassFc1: 500.0, BandpassFc2: 5000.0, BandstopFc1: 900.0, BandstopFc2: 1100.0}},
		},
		GenPanel: GeneratorSettings{Frequency: defaultFrequency, StartFrequency: defaultFrequency,
			StopFrequency: defaultFrequency, Sweep: genericps.NoSweep, Digital: true,
			Amplitude: defaultAmplitude, RaiseFallTimePercent: defaultRaiseFallTimePercent,
			TriggerTimeOffset: 0, NoiseAmplitude: 0, PhaseNoiseDegree: 0, Phase: 0},
		FfGen: GeneratorSettings{Frequency: defaultFrequency, StartFrequency: defaultFrequency,
			StopFrequency: defaultFrequency, Sweep: genericps.SweepUp, Digital: true,
			Amplitude: defaultAmplitude, RaiseFallTimePercent: defaultRaiseFallTimePercent,
			TriggerTimeOffset: 0, NoiseAmplitude: 0, PhaseNoiseDegree: 0, Phase: 0, WaveType: genericps.Sine, On: true},
		ExtGen: [2]GeneratorSettings{
			{Frequency: defaultFrequency, StartFrequency: defaultFrequency, StopFrequency: defaultFrequency, Amplitude: defaultAmplitude, WaveType: genericps.Sine, On: false},
			{Frequency: defaultFrequency, StartFrequency: defaultFrequency, StopFrequency: defaultFrequency, Amplitude: defaultAmplitude, WaveType: genericps.Sine, On: false},
		},
		SimGenPanel: []GeneratorSettings{
			{Frequency: defaultFrequency, StartFrequency: defaultFrequency,
				StopFrequency: defaultFrequency, Sweep: genericps.NoSweep, Digital: true,
				Amplitude: defaultAmplitude, RaiseFallTimePercent: defaultRaiseFallTimePercent,
				TriggerTimeOffset: 0, NoiseAmplitude: 0, PhaseNoiseDegree: 0, Phase: 0},
			{Frequency: defaultFrequency, StartFrequency: defaultFrequency,
				StopFrequency: defaultFrequency, Sweep: genericps.NoSweep, Digital: true,
				Amplitude: defaultAmplitude, RaiseFallTimePercent: defaultRaiseFallTimePercent,
				TriggerTimeOffset: 0, NoiseAmplitude: 0, PhaseNoiseDegree: 0, Phase: 0},
			{Frequency: defaultFrequency, StartFrequency: defaultFrequency,
				StopFrequency: defaultFrequency, Sweep: genericps.NoSweep, Digital: true,
				Amplitude: defaultAmplitude, RaiseFallTimePercent: defaultRaiseFallTimePercent,
				TriggerTimeOffset: 0, NoiseAmplitude: 0, PhaseNoiseDegree: 0, Phase: 0},
			{Frequency: defaultFrequency, StartFrequency: defaultFrequency,
				StopFrequency: defaultFrequency, Sweep: genericps.NoSweep, Digital: true,
				Amplitude: defaultAmplitude, RaiseFallTimePercent: defaultRaiseFallTimePercent,
				TriggerTimeOffset: 0, NoiseAmplitude: 0, PhaseNoiseDegree: 0, Phase: 0},
		},

		Dft:           DftSettings{MaxFreq: 1000000.0, MinFreq: 0, Window: WindowRectangular, DisplayMode: ModeDB, Bins: 1024, SampleRate: "100", SampleRateUnit: "MS/s"},
		Ff:            FfSettings{ReferenceChannel: 0, MinFreq: 1000, MaxFreq: 10000, DisplayMode: ModeDB, PtsDec: 100, DeltaT: 0.1, Amplitude: defaultAmplitude},
		Theme:         DarkTheme,
		StreamEnabled: &streamDefault,
	}
}

func Load(fileName string) (*PsSettings, error) {
	f, err := os.ReadFile(fileName)
	if err != nil {
		return nil, fmt.Errorf("read settings file: %w", err)
	}
	if len(f) < 65 {
		return nil, fmt.Errorf("settings file too short")
	}
	savedSum := string(f[0:64])
	f = f[65:]
	h := sha256.New()
	if _, err := h.Write(f); err != nil {
		return nil, fmt.Errorf("calculate hash: %w", err)
	}
	sum := fmt.Sprintf("%x", h.Sum(nil))
	if sum != savedSum {
		return nil, fmt.Errorf("checksum mismatch: expected %s, got %s", savedSum, sum)
	}
	settings := &PsSettings{}
	if err := yaml.Unmarshal(f, settings); err != nil {
		return nil, fmt.Errorf("unmarshal settings: %w", err)
	}

	// Ensure SimGenPanel has at least 4 elements (matching maximum possible channels)
	const minSimGenElements = 4
	for len(settings.SimGenPanel) < minSimGenElements {
		settings.SimGenPanel = append(settings.SimGenPanel, GeneratorSettings{
			Frequency:            defaultFrequency,
			StartFrequency:       defaultFrequency,
			StopFrequency:        defaultFrequency,
			Sweep:                genericps.NoSweep,
			Digital:              true,
			Amplitude:            defaultAmplitude,
			RaiseFallTimePercent: defaultRaiseFallTimePercent,
			TriggerTimeOffset:    0,
			NoiseAmplitude:       0,
			PhaseNoiseDegree:     0,
			Phase:                270,
		})
	}

	// Ensure digital filter settings are initialized if missing
	for i := range settings.Channels {
		df := &settings.Channels[i].DigitalFilter
		if df.LowpassFc == 0 {
			df.LowpassFc = 10000.0
		}
		if df.HighpassFc == 0 {
			df.HighpassFc = 100.0
		}
		if df.BandpassFc1 == 0 {
			df.BandpassFc1 = 500.0
		}
		if df.BandpassFc2 == 0 {
			df.BandpassFc2 = 5000.0
		}
		if df.BandstopFc1 == 0 {
			df.BandstopFc1 = 900.0
		}
		if df.BandstopFc2 == 0 {
			df.BandstopFc2 = 1100.0
		}
	}

	if settings.Trigger.Type == TriggerTypeComplex {
		settings.Trigger.ComplexEnabled = true
		settings.Trigger.Type = TriggerTypeAdvanced
	}

	if settings.StreamEnabled == nil {
		streamDefault := true
		settings.StreamEnabled = &streamDefault
	}

	if settings.ScreenSize == "" {
		settings.ScreenSize = ScreenSize1920x1080
	}
	return settings, nil
}

func Save(fileName string, settings *PsSettings) error {
	d, err := yaml.Marshal(settings)
	if err != nil {
		return fmt.Errorf("marshal settings: %w", err)
	}
	h := sha256.New()
	var buf []byte
	buf = append(buf, warningString...)
	d = append(buf, d...)
	if _, err := h.Write(d); err != nil {
		return fmt.Errorf("calculate hash: %w", err)
	}
	sum := fmt.Sprintf("%x\n", h.Sum(nil))
	d = append([]byte(sum), d...)
	if err := os.WriteFile(fileName, d, 0644); err != nil {
		slog.Debug("os.WriteFile 0644", "filename", fileName, "err", err)
		return fmt.Errorf("write settings file: %w", err)
	}
	return nil
}
