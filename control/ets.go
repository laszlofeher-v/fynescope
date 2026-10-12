package control

import (
	"fmt"
	"fynescope/genericps"
	"log/slog"
	"math"
	"time"
)

// minEtsRefreshTime is the minimum interval between consecutive ETS waveform refreshes
// to prevent overloading the GUI canvas and rendering thread.
const minEtsRefreshTime = 100 * time.Millisecond

// EtsTimes calculates the suggested ETS cycles and interleave factor based on the desired
// effective sampling interval in picoseconds (between 50 ps and 1000 ps).
//
// In Equivalent Time Sampling (ETS), repetitive waveforms are reconstructed by capturing
// samples across multiple trigger cycles with subtle time offsets (sub-sampling jitter).
// Interleave defines how many acquisition passes are interleaved to achieve the effective
// sampling rate, and Cycles defines the maximum number of trigger events allowed to collect
// those interleaved points.
func (psControl *PscDesc) EtsTimes(sampleTimeInPicoSeconds int32) (EtsCycles, EtsInterleave int16, err error) {
	// Hardware specification rules (e.g. for PS2000A / PS2407B / PS2207B):
	// - Effective sample time = BaseSampleTime / EtsInterleave
	// - EtsCycles >= EtsInterleave
	// - EtsCycles <= EtsInterleave * 10 + 9
	// - EtsInterleave <= 40 (or device specific limit)
	if sampleTimeInPicoSeconds >= 50 && sampleTimeInPicoSeconds <= 1000 {
		desiredEffectiveRate := 1e12 / float64(sampleTimeInPicoSeconds)
		interleaveFloat := desiredEffectiveRate / float64(psControl.MaxSamplingRate)
		EtsInterleave = int16(math.Round(interleaveFloat))

		maxInterleave, maxCycles := psControl.GetEtsLimits()
		if EtsInterleave > maxInterleave {
			EtsInterleave = maxInterleave
		}
		if EtsInterleave < 1 {
			EtsInterleave = 1
		}

		// Use 2x interleave as a robust default for cycles to ensure sufficient trigger
		// events are captured to fill all interleave bins without excessive waiting.
		EtsCycles = 2 * EtsInterleave

		if EtsCycles > maxCycles {
			EtsCycles = maxCycles
		}
	} else {
		err = fmt.Errorf("etsTimes: sampleTimeInPicoSeconds %d must be between 50 and 1000 for %s", sampleTimeInPicoSeconds, psControl.Info)
	}
	return
}

// GetEtsLimits returns the hardware upper bounds for ETS interleave and cycles.
func (psControl *PscDesc) GetEtsLimits() (maxInterleave, maxCycles int16) {
	maxInterleave = 40
	maxCycles = 500
	return
}

// etsBlockMode implements the acquisition lifecycle and state machine for Equivalent Time
// Sampling (ETS) block mode.
//
// ETS mode is designed for repetitive high-frequency signals exceeding the real-time
// analog-to-digital converter (ADC) sampling rate. The scope hardware captures sample points
// across multiple trigger events at varying small phase offsets and interleaves them.
//
// Key operational characteristics:
//   - ETS requires real hardware triggering; hardware auto-trigger is unsupported.
//   - An internal timeout (1.5 s) is used so that if a trigger event does not occur (e.g.,
//     signal removed or trigger threshold out of range), the capture is aborted and restarted,
//     keeping the UI responsive.
//   - Trigger status is reported as "Wait" (false) if armed for longer than
//     triggerWaitIndicatorDelay (250 ms) without a trigger, or upon timing out.
//   - Successful ETS block completion reports trigger status as "Trig'd" (true) right before
//     waveform buffers are dispatched to the GUI.
func etsBlockMode(psControl *PscDesc) state {
	// Buffered channel to receive hardware block-ready notifications from the driver callback.
	callbackChannel := make(chan struct{}, 1)

	// callbackBlock is registered with the driver RunBlock call.
	callbackBlock := func(handle int16, status int, param any) {
		select {
		case callbackChannel <- struct{}{}:
		default:
			slog.Debug("ETS callback dropped")
		}
	}

	// prepare sets up hardware channel parameters, sample counts, ETS interleave/cycles,
	// timebase, trigger settings, and pre/post trigger sample partitions.
	prepare := func() error {
		slog.Debug("ETS prepare")
		if err := psControl.setEverything(); err != nil {
			return err
		}
		// ETS relies on analog trigger inputs; at least one analog channel must be enabled.
		if psControl.numberOfEnabledAnalogChannels() == 0 {
			return fmt.Errorf("ETS mode requires at least one enabled analog channel")
		}

		// Fetch the latest trigger settings from triggerMonitor so we have up-to-date EtsInterleave/EtsCycles
		psControl.getTriggerCh <- &psControl.getTrigger
		newSettings := <-psControl.getTrigger.newSettings

		// Maximum samples to prevent memory exhaustion and excessive transfer latency in ETS mode.
		const maxEtsSamples = 250000

		psControl.refreshTime = time.Now()
		// Base sample count on screen width (at least 1024 samples for sufficient waveform detail).
		minSampleCount := uint64(math.Round(psControl.scopeScreenWidth))
		if minSampleCount < 1024 {
			minSampleCount = 1024
		}
		slog.Debug("prepare", "psControl.scopeScreenWidth", psControl.scopeScreenWidth)

		// Query available samples for 1 memory segment in current hardware configuration.
		sampleCount, err := psControl.memorySegments(uint64(1))
		if err != nil {
			slog.Error("ETS prepare: memorySegments failed", "error", err)
			return err
		}
		if sampleCount > 0 && minSampleCount > sampleCount {
			minSampleCount = sampleCount
		}
		if minSampleCount > maxEtsSamples {
			minSampleCount = maxEtsSamples
		}
		psControl.SampleCountRequired = minSampleCount
		etsDx := psControl.scopeScreenWidth / (psControl.maxScreenTime * 1e15)
		slog.Debug("prepare", "etsDx", etsDx)

		// Calculate desired effective sampling interval in picoseconds based on time span and sample count.
		minSampleTimeInPicoseconds := psControl.maxScreenTime * 1e12 / float64(psControl.SampleCountRequired)
		// Clamp between hardware-supported ETS limits (50 ps to 1000 ps).
		if minSampleTimeInPicoseconds > 1000 {
			slog.Debug("Clamp ", "minSampleTimeInPicoseconds", minSampleTimeInPicoseconds)
			minSampleTimeInPicoseconds = 1000
		} else if minSampleTimeInPicoseconds < 50 {
			slog.Debug("Clamp ", "minSampleTimeInPicoseconds", minSampleTimeInPicoseconds)
			minSampleTimeInPicoseconds = 50
		}

		sugCycles, sugInterleave, err := psControl.EtsTimes(int32(minSampleTimeInPicoseconds))

		etsInterleave := sugInterleave
		etsCycles := sugCycles

		userInterleave := psControl.triggerSetting.EtsInterleave
		userCycles := psControl.triggerSetting.EtsCycles

		// If user specified custom ETS parameters, apply them within hardware constraints.
		if userInterleave > 0 && userCycles > 0 {
			maxInterleave, maxCycles := psControl.GetEtsLimits()

			// Clamp interleave to hardware maximum.
			etsInterleave = userInterleave
			if etsInterleave > maxInterleave {
				etsInterleave = maxInterleave
			}

			// Re-evaluate maximum cycles based on interleave ratio.
			dynMaxCycles := etsInterleave * 5
			if maxCycles != 1000 { // If not the generic default
				if dynMaxCycles < maxCycles {
					maxCycles = dynMaxCycles
				}
			}

			// Clamp cycles between minimum required (2 * interleave) and maximum allowed.
			etsCycles = userCycles
			if etsCycles < etsInterleave*2 {
				etsCycles = etsInterleave * 2
			} else if etsCycles > maxCycles {
				etsCycles = maxCycles
			}
		}

		slog.Debug("prepare", "etsCycles", etsCycles, "etsInterleave", etsInterleave, "err", err)
		if err != nil {
			slog.Error("ETS prepare: etsTimes failed", "error", err)
			return err
		}

		// Calculate valid hardware timebase for RunBlock.
		psControl.overSample = 1
		tbInput := uint64(float64(psControl.maxScreenTime*1e9) / float64(psControl.SampleCountRequired))
		switch psControl.MaxSamplingRate {
		case MaxSampling100M:
			psControl.timeBase = timeBase100M(tbInput)
		case MaxSampling200M:
			psControl.timeBase = timeBase200M(tbInput)
		case MaxSampling500M:
			psControl.timeBase = timeBase500M(tbInput)
		default:
			psControl.timeBase = timeBase1G(tbInput)
		}
		if psControl.timeBase > psControl.timeBaseDec {
			psControl.timeBase -= psControl.timeBaseDec
		} else {
			psControl.timeBase = 0
		}
		maxSampleCount, _, err := psControl.getTimeBase(psControl.SampleCountRequired)
		if err != nil {
			slog.Error("ETS prepare: getTimeBase failed", "error", err)
			return err
		}

		// Arm ETS mode in the driver with fast ETS sampling.
		sampleTimePicoseconds, err := psControl.Con.SetEts(genericps.EtsFast, etsCycles, etsInterleave)
		if err != nil {
			slog.Error("ETS prepare: SetEts failed", "error", err)
			return err
		}
		if sampleTimePicoseconds <= 0 {
			err = fmt.Errorf("invalid ETS sample time achieved: %d ps", sampleTimePicoseconds)
			slog.Error("ETS prepare", "error", err)
			return err
		}

		// Update actual effective sampling interval resulting from driver configuration.
		psControl.SamplingTimeInterval = float64(sampleTimePicoseconds) * 1e-12
		slog.Debug("ETS", "TimeInterval", psControl.SamplingTimeInterval, "psControl.maxScreenTime", psControl.maxScreenTime)

		samplingIntervalChanged := psControl.SamplingTimeInterval != psControl.lastTriggerSamplingInterval

		// In ETS mode, triggering MUST be enabled in hardware. If the user has not explicitly
		// enabled a trigger, force a default simple rising-edge trigger on the selected source.
		if !psControl.triggerSetting.Enabled {
			dir := genericps.TriggerRising
			err = psControl.Con.SetSimpleTrigger(true,
				psControl.triggerSetting.Source, psControl.triggerSetting.TriggerADC,
				dir, 0, 0)
			if err != nil {
				slog.Error("ETS prepare default trigger failed", "error", err)
				return err
			}
			psControl.initialTriggerSet = true
		} else if newSettings || samplingIntervalChanged || !psControl.initialTriggerSet {
			err = psControl.sendTrigger()
			if err != nil {
				slog.Error("ETS prepare sendTrigger failed", "error", err)
				return err
			}
			psControl.initialTriggerSet = true
			psControl.lastTriggerSamplingInterval = psControl.SamplingTimeInterval
		}

		// Compute final required sample count clamped to driver and memory limits.
		rawSampleCount := uint64(math.Round(float64(psControl.maxScreenTime) / psControl.SamplingTimeInterval))

		limit := uint64(maxEtsSamples)
		if maxSampleCount > 0 && maxSampleCount < limit {
			limit = maxSampleCount
		}
		if sampleCount > 0 && sampleCount < limit {
			limit = sampleCount
		}

		if rawSampleCount > limit {
			slog.Debug("ETS sample count clamped to safe limit", "original", rawSampleCount, "max", limit)
			rawSampleCount = limit
		}
		if rawSampleCount < 2 {
			slog.Debug("ETS sample count clamped to minimum", "original", rawSampleCount, "min", 2)
			rawSampleCount = 2
		}
		psControl.SampleCountRequired = rawSampleCount
		if psControl.SampleCountRequired <= 0 {
			err = fmt.Errorf("invalid sample count: %d (TimeInterval=%g, maxScreenTime=%g)",
				psControl.SampleCountRequired, psControl.SamplingTimeInterval, psControl.maxScreenTime)
			slog.Error("ETS prepare", "error", err)
			return err
		}
		slog.Debug("ETS", "SampleCount", psControl.SampleCountRequired)

		// Calculate pre-trigger (NPre) and post-trigger (NPro) sample distributions based on horizontal trigger offset.
		psControl.NPre = uint64(math.Round(psControl.triggerSetting.XOffset / psControl.SamplingTimeInterval))
		if psControl.NPre > psControl.SampleCountRequired {
			psControl.NPre = psControl.SampleCountRequired
		}
		psControl.NPro = psControl.SampleCountRequired - psControl.NPre
		slog.Debug("pre", "SamplingTimeInterval", psControl.SamplingTimeInterval)
		// XRoundError captures sub-sample rounding error between exact trigger position and discrete sample points.
		psControl.XRoundError = psControl.triggerSetting.XOffset - psControl.SamplingTimeInterval*(float64(psControl.NPre)-1.0)
		slog.Debug("ets pre", "XRoundError", psControl.XRoundError)

		// Notify controller and buffers of the required sample size.
		psControl.EtsBufferCallback(int(psControl.SampleCountRequired))
		slog.Debug("ets pre", "SampleCount", psControl.SampleCountRequired, "scopeScreenWidth", psControl.scopeScreenWidth)
		return nil
	}

	// runBlock arms the oscilloscope hardware to start acquiring the ETS block capture.
	runBlock := func() error {
		slog.Debug("run", "psControl.NPre", psControl.NPre, "psControl.NPro", psControl.NPro,
			"psControl.timeBase", psControl.timeBase, "psControl.overSample", psControl.overSample)
		// Record timestamp when the hardware capture was armed.
		psControl.blockRunStart = time.Now()
		_, err := psControl.Con.RunBlock(psControl.NPre, psControl.NPro, psControl.timeBase, psControl.overSample, 0, callbackBlock, nil)
		if err != nil {
			slog.Error("ETS runBlock failed", "error", err)
			return err
		}
		return nil
	}

	stateMachine := func() {
		type eventHandlerFunc func() eventHandlerFunc

		var start, run, get eventHandlerFunc

		// start initializes the acquisition, waiting if no analog channels are enabled.
		start = func() eventHandlerFunc {
			// Ensure hardware is stopped before prepare/memorySegments.
			_ = psControl.Con.Stop()
			for psControl.numberOfEnabledAnalogChannels() == 0 {
				select {
				case <-psControl.restartChannel:
					return start
				case <-psControl.stopChannel:
					return nil
				}
			}

			if err := prepare(); err != nil {
				psControl.DisplayStatus(err.Error(), Fatal)
				return nil
			}
			return run
		}

		// run initiates hardware capture and transitions to get.
		run = func() eventHandlerFunc {
			// Drain any stale callback before starting a new acquisition.
			select {
			case <-callbackChannel:
			default:
			}
			if err := runBlock(); err != nil {
				psControl.DisplayStatus(err.Error(), Fatal)
				return nil
			}
			return get
		}

		// get waits for trigger events and capture completion, polling readiness and handling timeouts.
		get = func() eventHandlerFunc {
			ticker := time.NewTicker(20 * time.Millisecond)
			defer ticker.Stop()
			timeout := time.After(1500 * time.Millisecond)

			// waitTimer signals that the scope is waiting for a trigger event if
			// the capture has not completed within triggerWaitIndicatorDelay (250 ms).
			// This prevents rapid UI indicator flickering when triggers occur at a high rate.
			waitTimer := time.NewTimer(triggerWaitIndicatorDelay)
			defer waitTimer.Stop()

			for {
				select {
				case <-callbackChannel:
					// Hardware finished via driver callback.
					goto ready
				case <-ticker.C:
					// Polling fallback: check if hardware reports block ready.
					if ready, err := psControl.Con.LsReady(); err == nil && ready != 0 {
						goto ready
					}
				case <-waitTimer.C:
					// Armed for longer than delay without a trigger event: notify GUI "Wait".
					psControl.notifyTriggerState(false)
				case <-timeout:
					// In ETS mode, auto-trigger is not supported by hardware.
					// If no trigger event occurred within 1.5 s (e.g. trigger level out of signal range),
					// stop the acquisition, notify "Wait", and restart so the UI remains responsive.
					_ = psControl.Con.Stop()
					psControl.notifyTriggerState(false)
					return run
				case <-psControl.restartChannel:
					// Settings changed (e.g., timebase, channel, trigger level):
					// Stop hardware to cancel in-progress capture and re-initialize.
					_ = psControl.Con.Stop()
					select {
					case <-callbackChannel:
					case <-time.After(200 * time.Millisecond):
					}
					return start
				case <-psControl.stopChannel:
					// User stopped acquisition.
					return nil
				}
			}

		ready:
			// Enforce minimum refresh rate to prevent canvas overload.
			if dt := minEtsRefreshTime - time.Since(psControl.refreshTime); dt > 0 {
				time.Sleep(dt)
			}
			// Allocate and bind timestamp buffer for ETS sample interleaving.
			if err := psControl.setEtsBuffer(psControl.SampleCountRequired, 0); err != nil {
				slog.Error("ETS prepare: setEtsBuffers failed", "error", err)
				psControl.DisplayStatus(err.Error(), Fatal)
				return nil
			}
			psControl.refreshTime = time.Now()
			// Retrieve waveform data and timestamps from the driver (ets = true).
			// getData notifies triggerState(true) before invoking RefreshEtsCallback.
			if err := psControl.getData(psControl.SampleCountRequired, 0, true); err != nil {
				slog.Error("ETS get data failed", "error", err)
				psControl.DisplayStatus(err.Error(), Fatal)
				return nil
			}
			// Stop the hardware between block captures so the next RunBlock
			// starts from a clean, idle hardware state machine.
			_ = psControl.Con.Stop()
			time.Sleep(20 * time.Millisecond)
			return run
		}

		for handler := start; handler != nil; {
			handler = handler()
		}

		slog.Debug("ETS quit")
		_ = psControl.stopHardware()
		if psControl.Con != nil {
			// Disable ETS mode in hardware upon exiting ETS state machine.
			_, _ = psControl.Con.SetEts(genericps.EtsOff, 40, 4)
		}
	}

	stateMachine()
	return idle
}
