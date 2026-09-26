package netcontrol

import (
	"fynescope/control"
	"fynescope/genericps"
	"fynescope/settings"
)

// Command Message Types (Client -> Server)
const (
	TypeSetTrigger                   = "SetTrigger"
	TypeSetChannel                   = "SetChannel"
	TypeSetDigitalPort               = "SetDigitalPort"
	TypeSetGenerator                 = "SetGenerator"
	TypeSetDemoGen                   = "SetDemoGen"
	TypeSetInterpolationMode         = "SetInterpolationMode"
	TypeSetScopeScreenWidth         = "SetScopeScreenWidth"
	TypeSetMaxScreenTime             = "SetMaxScreenTime"
	TypeSuggestSampleCount           = "SuggestSampleCount"
	TypeSetETSMode                   = "SetETSMode"
	TypeSetBlockMode                 = "SetBlockMode"
	TypeStop                         = "Stop"
	TypeShutdown                     = "Shutdown"
	TypeSetResolutionMode             = "SetResolutionMode"
	TypeSetDigitalPortEnabled        = "SetDigitalPortEnabled"
	TypeSetMaxSamplingRate           = "SetMaxSamplingRate"
	TypeRequestRestart               = "RequestRestart"
	TypeSetStreamEnabled             = "SetStreamEnabled"
	TypeNewChannels                  = "NewChannels"
	TypeSetInfo                      = "SetInfo"
	TypeSetScopeModel                = "SetScopeModel"
	TypeSetDemoDigitalGen            = "SetDemoDigitalGen"
	TypeSetDemoRlcFilter             = "SetDemoRlcFilter"
	TypeGetAnalogueOffset            = "GetAnalogueOffset"
	TypeChannelRanges                = "ChannelRanges"
	TypeMinMaxValues                 = "MinMaxValues"
	TypeUnitVariantInfo              = "UnitVariantInfo"
	TypeUnitBatchAndSerialInfo       = "UnitBatchAndSerialInfo"
	TypeGetInfo                      = "GetInfo"
	TypeGetScopeModel                = "GetScopeModel"
	TypeGetMaxSamplingRate           = "GetMaxSamplingRate"
	TypeGetStreamEnabled             = "GetStreamEnabled"
	TypeGetSamplingTimeInterval      = "GetSamplingTimeInterval"
	TypeGetTimeBase                  = "GetTimeBase"
	TypeNumberOfEnabledAnalogChannels = "NumberOfEnabledAnalogChannels"
	TypeGetEtsLimits                 = "GetEtsLimits"
	TypeIsDemo                       = "IsDemo"
)

// Server Message Types (Server -> Client)
const (
	TypeRefresh         = "Refresh"
	TypeRefreshEts      = "RefreshEts"
	TypeBufferResize    = "BufferResize"
	TypeEtsBufferResize = "EtsBufferResize"
	TypeStatus          = "Status"
	TypeResponse        = "Response"
)

// CommandMessage is the top-level envelope for commands and requests sent from client to server.
type CommandMessage struct {
	ID      uint64      // Request ID (0 for fire-and-forget commands, >0 for query/synchronous RPCs)
	Type    string      // Command type identifier
	Payload interface{} // Concrete command or query payload
}

// ServerMessage is the top-level envelope for telemetry, callbacks, and responses sent from server to client.
type ServerMessage struct {
	Type    string      // Server message type identifier
	Payload interface{} // Concrete telemetry or response payload
}

// ResponseMessage carries the result of a synchronous RPC or acknowledged command.
type ResponseMessage struct {
	ID      uint64      // Matches CommandMessage.ID
	Success bool        // True if the command/query succeeded
	Error   string      // Error message if Success is false
	Result  interface{} // Result payload for queries, or nil
}

// --- Command Payloads (Fire-and-forget or state modifications) ---

type SetTriggerCmd struct {
	Desc control.TriggerDesc
}

type SetChannelCmd struct {
	Settings settings.ChSettings
}

type SetDigitalPortCmd struct {
	Msg control.DigitalPortMsg
}

type SetGeneratorCmd struct {
	Desc control.GeneratorDesc
}

type SetDemoGenCmd struct {
	Desc control.GeneratorDesc
}

type SetInterpolationModeCmd struct {
	Mode settings.InterpolationType
}

type SetScopeScreenWidthCmd struct {
	Width float64
}

type SetMaxScreenTimeCmd struct {
	Time float64
}

type SuggestSampleCountCmd struct {
	Count uint64
}

type SetETSModeCmd struct{}

type SetBlockModeCmd struct{}

type StopCmd struct{}

type ShutdownCmd struct{}

type SetResolutionModeCmd struct {
	Mode genericps.RatioMode
}

type SetDigitalPortEnabledCmd struct {
	Port    int
	Enabled bool
}

type SetMaxSamplingRateCmd struct {
	Rate uint32
}

type RequestRestartCmd struct{}

type SetStreamEnabledCmd struct {
	Enabled bool
}

type NewChannelsCmd struct {
	NumberOfChannels int
}

type SetInfoCmd struct {
	Info string
}

type SetScopeModelCmd struct {
	Model control.ScopeType
}

type SetDemoDigitalGenCmd struct {
	Port0Enabled bool
	Port1Enabled bool
	Frequency    float64
	Direction    genericps.DigitalDemoGenDirection
	Encoding     genericps.DigitalDemoGenEncoding
	Mode         genericps.DigitalDemoGenMode
	BitDelay     float64
}

type SetDemoRlcFilterCmd struct {
	Channel    genericps.ChannelId
	GenSource  genericps.ChannelId
	Enabled    bool
	FilterType string
	R          float64
	RUnit      string
	L          float64
	LUnit      string
	C          float64
	CUnit      string
}

// --- Query Payloads (Client -> Server, expecting a ResponseMessage) ---

type GetAnalogueOffsetReq struct {
	VoltageRange int
	Coupling     genericps.Coupling
}

type ChannelRangesReq struct {
	Channel genericps.ChannelId
}

type MinMaxValuesReq struct{}

type UnitVariantInfoReq struct{}

type UnitBatchAndSerialInfoReq struct{}

type GetInfoReq struct{}

type GetScopeModelReq struct{}

type GetMaxSamplingRateReq struct{}

type GetStreamEnabledReq struct{}

type GetSamplingTimeIntervalReq struct{}

type GetTimeBaseReq struct{}

type NumberOfEnabledAnalogChannelsReq struct{}

type GetEtsLimitsReq struct{}

type IsDemoReq struct{}

// --- Response Payloads (Carried inside ResponseMessage.Result) ---

type GetAnalogueOffsetRsp struct {
	Max float32
	Min float32
}

type ChannelRangesRsp struct {
	Ranges []int32
}

type MinMaxValuesRsp struct {
	Min int32
	Max int32
}

type UnitVariantInfoRsp struct {
	Info string
}

type UnitBatchAndSerialInfoRsp struct {
	Info string
}

type GetInfoRsp struct {
	Info string
}

type GetScopeModelRsp struct {
	Model control.ScopeType
}

type GetMaxSamplingRateRsp struct {
	Rate uint32
}

type GetStreamEnabledRsp struct {
	Enabled bool
}

type GetSamplingTimeIntervalRsp struct {
	Interval float64
}

type GetTimeBaseRsp struct {
	TimeBase uint64
}

type NumberOfEnabledAnalogChannelsRsp struct {
	Count int
}

type GetEtsLimitsRsp struct {
	MaxInterleave int16
	MaxCycles     int16
}

type IsDemoRsp struct {
	IsDemo bool
}

// --- Telemetry Payloads (Server -> Client callbacks) ---

// DataMessage carries analog and digital waveform buffers from the server's RefreshCallback.
type DataMessage struct {
	Buffers              [][]int16
	BuffersMin           [][]int16
	DigitalBuffers       [][]int16
	StartTimeOffset      int64
	XRoundError          float64
	SamplingTimeInterval float64
}

// EtsDataMessage carries waveform data from the server's RefreshEtsCallback.
type EtsDataMessage struct {
	Buffers              [][]int16
	EtsOutBuffer         []int64
	XRoundError          float64
	SamplingTimeInterval float64
}

// BufferResizeMessage notifies the client of buffer resizing (BufferCallback).
type BufferResizeMessage struct {
	Size int
}

// EtsBufferResizeMessage notifies the client of ETS buffer resizing (EtsBufferCallback).
type EtsBufferResizeMessage struct {
	Size int
}

// StatusMessage relays scope status messages to the client (DisplayStatus callback).
type StatusMessage struct {
	Text      string
	ErrorType control.ScopeError
}
