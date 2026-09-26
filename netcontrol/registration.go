package netcontrol

import (
	"encoding/gob"
	"image/color"
	"sync"

	"fynescope/control"
	"fynescope/genericps"
	"fynescope/settings"
)

var registerOnce sync.Once

func init() {
	RegisterGobTypes()
}

// RegisterGobTypes registers all message and payload types used in the netcontrol protocol.
// It is called automatically during package initialization and is idempotent.
func RegisterGobTypes() {
	registerOnce.Do(func() {
		// Envelopes & Response
		gob.Register(&CommandMessage{})
		gob.Register(&ServerMessage{})
		gob.Register(&ResponseMessage{})

		// Command Payloads
		gob.Register(&SetTriggerCmd{})
		gob.Register(&SetChannelCmd{})
		gob.Register(&SetDigitalPortCmd{})
		gob.Register(&SetGeneratorCmd{})
		gob.Register(&SetDemoGenCmd{})
		gob.Register(&SetInterpolationModeCmd{})
		gob.Register(&SetScopeScreenWidthCmd{})
		gob.Register(&SetMaxScreenTimeCmd{})
		gob.Register(&SuggestSampleCountCmd{})
		gob.Register(&SetETSModeCmd{})
		gob.Register(&SetBlockModeCmd{})
		gob.Register(&StopCmd{})
		gob.Register(&ShutdownCmd{})
		gob.Register(&SetResolutionModeCmd{})
		gob.Register(&SetDigitalPortEnabledCmd{})
		gob.Register(&SetMaxSamplingRateCmd{})
		gob.Register(&RequestRestartCmd{})
		gob.Register(&SetStreamEnabledCmd{})
		gob.Register(&NewChannelsCmd{})
		gob.Register(&SetInfoCmd{})
		gob.Register(&SetScopeModelCmd{})
		gob.Register(&SetDemoDigitalGenCmd{})
		gob.Register(&SetDemoRlcFilterCmd{})

		// Query Request Payloads
		gob.Register(&GetAnalogueOffsetReq{})
		gob.Register(&ChannelRangesReq{})
		gob.Register(&MinMaxValuesReq{})
		gob.Register(&UnitVariantInfoReq{})
		gob.Register(&UnitBatchAndSerialInfoReq{})
		gob.Register(&GetInfoReq{})
		gob.Register(&GetScopeModelReq{})
		gob.Register(&GetMaxSamplingRateReq{})
		gob.Register(&GetStreamEnabledReq{})
		gob.Register(&GetSamplingTimeIntervalReq{})
		gob.Register(&GetTimeBaseReq{})
		gob.Register(&NumberOfEnabledAnalogChannelsReq{})
		gob.Register(&GetEtsLimitsReq{})
		gob.Register(&IsDemoReq{})

		// Query Response Payloads
		gob.Register(&GetAnalogueOffsetRsp{})
		gob.Register(&ChannelRangesRsp{})
		gob.Register(&MinMaxValuesRsp{})
		gob.Register(&UnitVariantInfoRsp{})
		gob.Register(&UnitBatchAndSerialInfoRsp{})
		gob.Register(&GetInfoRsp{})
		gob.Register(&GetScopeModelRsp{})
		gob.Register(&GetMaxSamplingRateRsp{})
		gob.Register(&GetStreamEnabledRsp{})
		gob.Register(&GetSamplingTimeIntervalRsp{})
		gob.Register(&GetTimeBaseRsp{})
		gob.Register(&NumberOfEnabledAnalogChannelsRsp{})
		gob.Register(&GetEtsLimitsRsp{})
		gob.Register(&IsDemoRsp{})

		// Telemetry Payloads
		gob.Register(&DataMessage{})
		gob.Register(&EtsDataMessage{})
		gob.Register(&BufferResizeMessage{})
		gob.Register(&EtsBufferResizeMessage{})
		gob.Register(&StatusMessage{})

		// Nested / Domain Types
		gob.Register(control.TriggerDesc{})
		gob.Register(control.TriggerDirections{})
		gob.Register(control.TriggerModes(0))
		gob.Register(control.TriggerTypes(0))
		gob.Register(control.ScopeType(0))
		gob.Register(control.ScopeError(0))
		gob.Register(control.DigitalPortMsg{})
		gob.Register(control.GeneratorDesc{})

		gob.Register(settings.ChSettings{})
		gob.Register(settings.DigitalFilterSettings{})
		gob.Register(settings.RlcFilterSettings{})
		gob.Register(settings.ChTriggerSettings{})
		gob.Register(settings.DigitalPortSettings{})
		gob.Register(settings.InterpolationType(0))
		gob.Register(settings.FvMode(0))

		gob.Register(genericps.ChannelId(0))
		gob.Register(genericps.Coupling(0))
		gob.Register(genericps.RangeEnum(0))
		gob.Register(genericps.ThresholdDirection(0))
		gob.Register(genericps.ThresholdModeId(0))
		gob.Register(genericps.TriggerChannelProperties{})
		gob.Register(genericps.TriggerConditions{})
		gob.Register(genericps.PwqConditions{})
		gob.Register(genericps.TriggerRespBase(0))
		gob.Register(genericps.DigitalChannelDirections{})
		gob.Register(genericps.DigitalChannel(0))
		gob.Register(genericps.DigitalDirection(0))
		gob.Register(genericps.TriggerOperand(0))
		gob.Register(genericps.PulseWidthType(0))
		gob.Register(genericps.WaveTypeEnum(0))
		gob.Register(genericps.SweepTypeEnum(0))
		gob.Register(genericps.ExtraOperations(0))
		gob.Register(genericps.SigGenTrigType(0))
		gob.Register(genericps.SigGenTrigSource(0))
		gob.Register(genericps.IndexMode(0))
		gob.Register(genericps.RatioMode(0))
		gob.Register(genericps.DigitalDemoGenDirection(0))
		gob.Register(genericps.DigitalDemoGenEncoding(0))
		gob.Register(genericps.DigitalDemoGenMode(0))
		gob.Register(color.NRGBA{})
	})
}
