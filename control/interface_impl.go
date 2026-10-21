package control

import (
	"fynescope/genericps"
	"fynescope/settings"
)

// SetTrigger sends updated trigger settings to the trigger monitor actor channel.
func (ps *PscDesc) SetTrigger(msg *TriggerDescMsg) {
	if ps.SetTriggerCh != nil {
		ps.SetTriggerCh <- msg
	}
}

// SetChannel sends updated analog channel settings to the channel state machine actor channel.
func (ps *PscDesc) SetChannel(msg *settings.ChSettings) {
	if ps.SetChannelCh != nil {
		ps.SetChannelCh <- msg
	}
}

// SetDigitalPort sends digital MSO port configuration updates to the digital port monitor.
func (ps *PscDesc) SetDigitalPort(msg *DigitalPortMsg) {
	if ps.SetDigitalPortCh != nil {
		ps.SetDigitalPortCh <- msg
	}
}

// SetGenerator dispatches signal generator parameters to the hardware generator actor.
func (ps *PscDesc) SetGenerator(msg *GeneratorDescMsg) {
	if ps.SetGeneratorCh != nil {
		ps.SetGeneratorCh <- msg
	}
}

// SetDemoGen dispatches signal generator parameters to the simulated demo generator actor.
func (ps *PscDesc) SetDemoGen(msg *GeneratorDescMsg) {
	if ps.SetDemoGenCh != nil {
		ps.SetDemoGenCh <- msg
	}
}

// SetInterpolationMode updates the active waveform interpolation method (Raw, Linear, or Sinc).
func (ps *PscDesc) SetInterpolationMode(mode settings.InterpolationType) {
	if ps.SetInterpolationModeCh != nil {
		ps.SetInterpolationModeCh <- mode
	}
}

// GetCon returns the underlying hardware Connection descriptor.
func (ps *PscDesc) GetCon() *genericps.Connection {
	return ps.Con
}

// GetInfo returns the oscilloscope device identification and variant information string.
func (ps *PscDesc) GetInfo() string {
	return ps.Info
}

// GetScopeModel returns the classified hardware model family (e.g. Ps2000, Ps3000a, etc.).
func (ps *PscDesc) GetScopeModel() ScopeType {
	return ps.ScopeModel
}

// GetMaxSamplingRate returns the maximum sampling rate in samples per second supported by the hardware.
func (ps *PscDesc) GetMaxSamplingRate() uint32 {
	return ps.MaxSamplingRate
}

// GetStreamEnabled reports whether continuous streaming acquisition mode is currently active.
func (ps *PscDesc) GetStreamEnabled() bool {
	return ps.StreamEnabled.Load()
}

// SetStreamEnabled atomically updates whether streaming acquisition mode is active.
func (ps *PscDesc) SetStreamEnabled(b bool) {
	ps.StreamEnabled.Store(b)
}

// GetSamplingTimeInterval returns the duration between consecutive samples in seconds.
func (ps *PscDesc) GetSamplingTimeInterval() float64 {
	return ps.SamplingTimeInterval
}

// GetTimeBase returns the active hardware timebase index.
func (ps *PscDesc) GetTimeBase() uint64 {
	return ps.timeBase
}

// SetRefreshCallback registers the callback invoked when new block or streaming waveform samples are ready for GUI rendering.
func (ps *PscDesc) SetRefreshCallback(cb func(buffers [][]int16, buffersMin [][]int16, digitalBuffers [][]int16, startTimeOffset int64, xRoundError float64, samplingTimeInterval float64)) {
	ps.RefreshCallback = cb
}

// SetRefreshEtsCallback registers the callback invoked when new Equivalent Time Sampling (ETS) data is ready for display.
func (ps *PscDesc) SetRefreshEtsCallback(cb func(buffers [][]int16, etsOutBuffer []int64, xRoundError float64, samplingTimeInterval float64)) {
	ps.RefreshEtsCallback = cb
}

// SetBufferCallback registers the callback invoked when new acquisition buffer capacities are established.
func (ps *PscDesc) SetBufferCallback(cb func(size int)) {
	ps.BufferCallback = cb
}

// SetEtsBufferCallback registers the callback invoked when ETS buffer sizes change.
func (ps *PscDesc) SetEtsBufferCallback(cb func(size int)) {
	ps.EtsBufferCallback = cb
}

// SetDisplayStatus registers the callback invoked to report errors, warnings, or notifications to the UI.
func (ps *PscDesc) SetDisplayStatus(cb func(s string, errorType ScopeError)) {
	ps.DisplayStatus = cb
}

// ShowDisplayStatus triggers the registered status callback with the provided message and severity.
func (ps *PscDesc) ShowDisplayStatus(s string, errorType ScopeError) {
	if ps.DisplayStatus != nil {
		ps.DisplayStatus(s, errorType)
	}
}

// SetMaxSamplingRate sets the maximum hardware sampling frequency in samples per second.
func (ps *PscDesc) SetMaxSamplingRate(rate uint32) {
	ps.MaxSamplingRate = rate
}

// SetResolutionMode configures hardware sample resolution downsampling / aggregation mode.
func (ps *PscDesc) SetResolutionMode(mode genericps.RatioMode) {
	ps.ResolutionMode.Store(int32(mode))
}

// CallBufferCallback invokes BufferCallback if registered.
func (ps *PscDesc) CallBufferCallback(size int) {
	if ps.BufferCallback != nil {
		ps.BufferCallback(size)
	}
}

// CallRefreshCallback invokes RefreshCallback if registered with newly acquired analog and digital sample data.
func (ps *PscDesc) CallRefreshCallback(buffers [][]int16, buffersMin [][]int16, digitalBuffers [][]int16, startTimeOffset int64, xRoundError float64, samplingTimeInterval float64) {
	if ps.RefreshCallback != nil {
		ps.RefreshCallback(buffers, buffersMin, digitalBuffers, startTimeOffset, xRoundError, samplingTimeInterval)
	}
}

// SetInfo sets the device identification metadata string.
func (ps *PscDesc) SetInfo(info string) {
	ps.Info = info
}

// SetScopeModel sets the detected hardware oscilloscope model type.
func (ps *PscDesc) SetScopeModel(model ScopeType) {
	ps.ScopeModel = model
}

// IsDemo reports whether the connected oscilloscope is the software simulation device.
func (ps *PscDesc) IsDemo() bool {
	return ps.Con != nil && ps.Con.ID == genericps.DemoId
}

// SetDemoDigitalGen passes digital pattern generator parameters directly to the demo backend connection.
func (ps *PscDesc) SetDemoDigitalGen(port0Enabled, port1Enabled bool, freq float64, dir genericps.DigitalDemoGenDirection, enc genericps.DigitalDemoGenEncoding, mode genericps.DigitalDemoGenMode, bitDelay float64) error {
	if ps.Con != nil {
		return ps.Con.SetDemoDigitalGen(port0Enabled, port1Enabled, freq, dir, enc, mode, bitDelay)
	}
	return nil
}

// SetDemoRlcFilter passes RLC filter simulation parameters directly to the demo backend connection.
func (ps *PscDesc) SetDemoRlcFilter(channel genericps.ChannelId, genSource genericps.ChannelId, enabled bool, filterType string, r float64, runit string, l float64, lunit string, cval float64, cunit string, q float64) error {
	if ps.Con != nil {
		return ps.Con.SetDemoRlcFilter(channel, genSource, enabled, filterType, r, runit, l, lunit, cval, cunit, q)
	}
	return nil
}

