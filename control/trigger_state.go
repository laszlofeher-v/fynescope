package control

import (
	"fynescope/genericps"
	"time"
)

const (
	// triggerWaitIndicatorDelay is how long a block capture may stay armed before
	// the trigger state is reported as "not triggered". It keeps the indicator from
	// flickering when triggers arrive at a high rate.
	triggerWaitIndicatorDelay = 250 * time.Millisecond

	// autoTriggerDetectNumerator/Denominator define the fraction of the auto-trigger
	// timeout after which a finished capture is considered to have been released by
	// the auto-trigger rather than by a real trigger event (the estimated waiting time
	// also contains driver and callback latency).
	autoTriggerDetectNumerator   = 9
	autoTriggerDetectDenominator = 10
)

// TriggerStateReporter is implemented by scope controllers that report whether the
// latest acquisition was released by a real trigger event (true) or not (false), i.e.
// the scope is still waiting, the trigger is disabled or the auto-trigger timed out.
// The callback is invoked before the waveform of that acquisition is delivered.
type TriggerStateReporter interface {
	SetTriggerStateCallback(cb func(triggered bool))
}

var _ TriggerStateReporter = (*PscDesc)(nil)

// SetTriggerStateCallback registers the callback invoked when the trigger state is
// (re)evaluated. Passing nil removes the callback.
func (ps *PscDesc) SetTriggerStateCallback(cb func(triggered bool)) {
	if cb == nil {
		ps.triggerStateCb.Store(nil)
		return
	}
	ps.triggerStateCb.Store(&cb)
}

// notifyTriggerState invokes the registered trigger state callback, if any.
func (ps *PscDesc) notifyTriggerState(triggered bool) {
	if cb := ps.triggerStateCb.Load(); cb != nil {
		(*cb)(triggered)
	}
}

// setTriggerChannelProperties sends the trigger channel properties to the hardware and
// remembers the auto-trigger timeout, which is needed to interpret the block result.
func (ps *PscDesc) setTriggerChannelProperties(channelProperties []genericps.TriggerChannelProperties,
	auxOutputEnable bool, autoTriggerMs int32) error {
	ps.activeAutoTriggerMs.Store(autoTriggerMs)
	return ps.Con.SetTriggerChannelProperties(channelProperties, auxOutputEnable, autoTriggerMs)
}

// setSimpleTrigger sends a simple trigger to the hardware and remembers the
// auto-trigger timeout, which is needed to interpret the block result.
func (ps *PscDesc) setSimpleTrigger(enable bool, source genericps.ChannelId, threshold int16,
	direction genericps.ThresholdDirection, delay uint32, autoTriggerMs int16) error {
	ps.activeAutoTriggerMs.Store(int32(autoTriggerMs))
	return ps.Con.SetSimpleTrigger(enable, source, threshold, direction, delay, autoTriggerMs)
}

// blockTriggered decides whether a finished block capture was released by a real
// trigger event. The block-ready callback of the drivers carries no such information,
// so it is derived from timing: with an auto-trigger timeout configured, a capture that
// waited (almost) as long as the timeout was released by the auto-trigger.
//
// triggerActive tells if any trigger is armed at all, autoTriggerMs is the timeout sent
// to the hardware (0 = none) and waited is the time spent waiting for the trigger, i.e.
// the elapsed time since arming minus the pure sample acquisition time.
func blockTriggered(triggerActive bool, autoTriggerMs int32, waited time.Duration) bool {
	if !triggerActive {
		return false
	}
	if autoTriggerMs <= 0 {
		// without auto-trigger the capture can only complete on a trigger event
		return true
	}
	limit := time.Duration(autoTriggerMs) * time.Millisecond *
		autoTriggerDetectNumerator / autoTriggerDetectDenominator
	return waited < limit
}

// reportBlockTriggerState evaluates the finished block capture and notifies the
// trigger state listener. It must be called before the waveform is delivered so the
// display shows the state that belongs to the displayed data.
func (psControl *PscDesc) reportBlockTriggerState() {
	if psControl.triggerStateCb.Load() == nil || psControl.blockRunStart.IsZero() {
		return
	}
	elapsed := time.Since(psControl.blockRunStart)
	var capture time.Duration
	if !psControl.IsDemo() {
		// the simulator delivers the block immediately, real hardware needs the
		// time to acquire the samples before it can complete the capture
		capture = time.Duration(float64(psControl.SampleCountRequired) *
			psControl.SamplingTimeInterval * float64(time.Second))
	}
	active := psControl.triggerSetting.Enabled || psControl.hasActiveDigitalTrigger()
	psControl.notifyTriggerState(blockTriggered(active, psControl.activeAutoTriggerMs.Load(), elapsed-capture))
}
