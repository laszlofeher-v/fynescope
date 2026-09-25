package control

import (
	"fynescope/genericps"
	"fynescope/settings"
)

func (ps *PscDesc) SetTrigger(msg *TriggerDescMsg) {
	if ps.SetTriggerCh != nil {
		ps.SetTriggerCh <- msg
	}
}

func (ps *PscDesc) SetChannel(msg *settings.ChSettings) {
	if ps.SetChannelCh != nil {
		ps.SetChannelCh <- msg
	}
}

func (ps *PscDesc) SetDigitalPort(msg *DigitalPortMsg) {
	if ps.SetDigitalPortCh != nil {
		ps.SetDigitalPortCh <- msg
	}
}

func (ps *PscDesc) SetGenerator(msg *GeneratorDescMsg) {
	if ps.SetGeneratorCh != nil {
		ps.SetGeneratorCh <- msg
	}
}

func (ps *PscDesc) SetDemoGen(msg *GeneratorDescMsg) {
	if ps.SetDemoGenCh != nil {
		ps.SetDemoGenCh <- msg
	}
}

func (ps *PscDesc) SetInterpolationMode(mode settings.InterpolationType) {
	if ps.SetInterpolationModeCh != nil {
		ps.SetInterpolationModeCh <- mode
	}
}

func (ps *PscDesc) GetCon() *genericps.Connection {
	return ps.Con
}

func (ps *PscDesc) GetInfo() string {
	return ps.Info
}

func (ps *PscDesc) GetScopeModel() ScopeType {
	return ps.ScopeModel
}

func (ps *PscDesc) GetMaxSamplingRate() uint32 {
	return ps.MaxSamplingRate
}

func (ps *PscDesc) GetStreamEnabled() bool {
	return ps.StreamEnabled.Load()
}

func (ps *PscDesc) SetStreamEnabled(b bool) {
	ps.StreamEnabled.Store(b)
}

func (ps *PscDesc) GetSamplingTimeInterval() float64 {
	return ps.SamplingTimeInterval
}

func (ps *PscDesc) GetTimeBase() uint64 {
	return ps.timeBase
}

func (ps *PscDesc) SetRefreshCallback(cb func(buffers [][]int16, buffersMin [][]int16, digitalBuffers [][]int16, startTimeOffset int64, xRoundError float64, samplingTimeInterval float64)) {
	ps.RefreshCallback = cb
}

func (ps *PscDesc) SetRefreshEtsCallback(cb func(buffers [][]int16, etsOutBuffer []int64, xRoundError float64, samplingTimeInterval float64)) {
	ps.RefreshEtsCallback = cb
}

func (ps *PscDesc) SetBufferCallback(cb func(size int)) {
	ps.BufferCallback = cb
}

func (ps *PscDesc) SetEtsBufferCallback(cb func(size int)) {
	ps.EtsBufferCallback = cb
}

func (ps *PscDesc) SetDisplayStatus(cb func(s string, errorType ScopeError)) {
	ps.DisplayStatus = cb
}

func (ps *PscDesc) ShowDisplayStatus(s string, errorType ScopeError) {
	if ps.DisplayStatus != nil {
		ps.DisplayStatus(s, errorType)
	}
}

func (ps *PscDesc) SetMaxSamplingRate(rate uint32) {
	ps.MaxSamplingRate = rate
}

func (ps *PscDesc) SetResolutionMode(mode genericps.RatioMode) {
	ps.ResolutionMode.Store(int32(mode))
}

func (ps *PscDesc) CallBufferCallback(size int) {
	if ps.BufferCallback != nil {
		ps.BufferCallback(size)
	}
}

func (ps *PscDesc) CallRefreshCallback(buffers [][]int16, buffersMin [][]int16, digitalBuffers [][]int16, startTimeOffset int64, xRoundError float64, samplingTimeInterval float64) {
	if ps.RefreshCallback != nil {
		ps.RefreshCallback(buffers, buffersMin, digitalBuffers, startTimeOffset, xRoundError, samplingTimeInterval)
	}
}

func (ps *PscDesc) SetInfo(info string) {
	ps.Info = info
}

func (ps *PscDesc) SetScopeModel(model ScopeType) {
	ps.ScopeModel = model
}
