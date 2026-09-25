package control

import (
	"fynescope/genericps"
	"fynescope/settings"
)

type ScopeController interface {
	SetTrigger(msg *TriggerDescMsg)
	SetChannel(msg *settings.ChSettings)
	SetDigitalPort(msg *DigitalPortMsg)
	SetGenerator(msg *GeneratorDescMsg)
	SetDemoGen(msg *GeneratorDescMsg)
	SetInterpolationMode(mode settings.InterpolationType)
	SetScopeScreenWidth(width float64)
	SetMaxScreenTime(t float64)
	SuggestSampleCount(sc uint64)
	SetETSMode() error
	SetBlockMode() error
	Shutdown()
	Stop() error

	// Connection bypass (for now, to ease transition)
	GetCon() *genericps.Connection

	// Accessors
	GetInfo() string
	GetScopeModel() ScopeType
	GetMaxSamplingRate() uint32
	GetStreamEnabled() bool
	SetStreamEnabled(b bool)
	GetSamplingTimeInterval() float64
	GetTimeBase() uint64

	// Methods already implemented by PscDesc
	ChannelRanges(chIndex genericps.ChannelId) ([]int32, error)
	NumberOfEnabledAnalogChannels() int
	MinMaxValues() (min, max int32, err error)
	UnitVariantInfo() (info string, err error)
	UnitBatchAndSerialInfo() (info string, err error)
	SetDigitalPortEnabled(port int, enabled bool)

	// Callbacks
	SetRefreshCallback(cb func(buffers [][]int16, buffersMin [][]int16, digitalBuffers [][]int16, startTimeOffset int64, xRoundError float64, samplingTimeInterval float64))
	SetRefreshEtsCallback(cb func(buffers [][]int16, etsOutBuffer []int64, xRoundError float64, samplingTimeInterval float64))
	SetBufferCallback(cb func(size int))
	SetEtsBufferCallback(cb func(size int))
	SetDisplayStatus(cb func(s string, errorType ScopeError))
	ShowDisplayStatus(s string, errorType ScopeError)
	NewChannels(numberOfChannels int)
	SetMaxSamplingRate(rate uint32)
	RequestRestart()
	SetResolutionMode(mode genericps.RatioMode)
	CallBufferCallback(size int)
	CallRefreshCallback(buffers [][]int16, buffersMin [][]int16, digitalBuffers [][]int16, startTimeOffset int64, xRoundError float64, samplingTimeInterval float64)
	SetInfo(info string)
	SetScopeModel(model ScopeType)
	GetEtsLimits() (maxInterleave, maxCycles int16)
}
