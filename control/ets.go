package control

import (
	"fmt"
	"fynescope/genericps"
	"log/slog"
	"math"
	"time"
)

const minEtsRefreshTime = 100 * time.Millisecond

func (psControl *PscDesc) EtsTimes(sampleTimeInPicoSeconds int32) (EtsCycles, EtsInterleave int16, err error) {
	// Specification for 2407B / 2207B / 2000a:
	// Sample time = 2000 / EtsInterleave
	// EtsCycles >= EtsInterleave
	// EtsCycles <= EtsInterleave * 10 + 9
	// EtsInterleave <= 40
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

		EtsCycles = 2 * EtsInterleave

		if EtsCycles > maxCycles {
			EtsCycles = maxCycles
		}
	} else {
		err = fmt.Errorf("etsTimes: sampleTimeInPicoSeconds %d must be between 50 and 1000 for %s", sampleTimeInPicoSeconds, psControl.Info)
	}
	return
}

func (psControl *PscDesc) GetEtsLimits() (maxInterleave, maxCycles int16) {
	maxInterleave = 40
	maxCycles = 500
	return
}

func etsBlockMode(psControl *PscDesc) state {
	callbackChannel := make(chan struct{}, 1)

	callbackBlock := func(handle int16, status int, param any) {
		select {
		case callbackChannel <- struct{}{}:
		default:
			slog.Debug("ETS callback dropped")
		}
	}

	prepare := func() error {
		slog.Debug("ETS prepare")
		if err := psControl.setEverything(); err != nil {
			return err
		}

		// Fetch the latest trigger settings from triggerMonitor so we have up-to-date EtsInterleave/EtsCycles
		psControl.getTriggerCh <- &psControl.getTrigger
		newSettings := <-psControl.getTrigger.newSettings

		psControl.refreshTime = time.Now()
		psControl.SampleCountRequired = uint64(math.Round(psControl.scopeScreenWidth))
		slog.Debug("prepare", "psControl.scopeScreenWidth", psControl.scopeScreenWidth)

		sampleCount, err := psControl.memorySegments(uint64(1))
		if err != nil {
			slog.Error("ETS prepare: memorySegments failed", "error", err)
			return err
		}
		if sampleCount < psControl.SampleCountRequired {
			psControl.SampleCountRequired = sampleCount
		}
		etsDx := psControl.scopeScreenWidth / (psControl.maxScreenTime * 1e15)
		slog.Debug("draw", "etsDx", etsDx)

		minSampleTimeInPicoseconds := psControl.maxScreenTime * 1e12 / float64(psControl.SampleCountRequired)
		// Clamp between 50 and 1000 ps
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

		if userInterleave > 0 && userCycles > 0 {
			maxInterleave, maxCycles := psControl.GetEtsLimits()

			// Clamp interleave
			etsInterleave = userInterleave
			if etsInterleave > maxInterleave {
				etsInterleave = maxInterleave
			}

			// Re-evaluate max cycles based on the rule
			dynMaxCycles := etsInterleave * 5
			if maxCycles != 1000 { // If not the generic default
				if dynMaxCycles < maxCycles {
					maxCycles = dynMaxCycles
				}
			}

			// Clamp cycles
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

		// Calculate valid timeBase for RunBlock
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
		_, _, err = psControl.getTimeBase(psControl.SampleCountRequired)
		if err != nil {
			slog.Error("ETS prepare: getTimeBase failed", "error", err)
			return err
		}

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

		psControl.SamplingTimeInterval = float64(sampleTimePicoseconds) * 1e-12
		slog.Debug("ETS", "TimeInterval", psControl.SamplingTimeInterval, "psControl.maxScreenTime", psControl.maxScreenTime)

		samplingIntervalChanged := psControl.SamplingTimeInterval != psControl.lastTriggerSamplingInterval
		timeDependentTrigger := psControl.triggerSetting.Type == Interval || psControl.triggerSetting.Type == PulseWidth || psControl.triggerSetting.Type == Dropout || psControl.triggerSetting.Type == WindowDropout || psControl.triggerSetting.Type == WindowPulseWidth || psControl.triggerSetting.Type == RiseFall

		// In ETS mode, triggering MUST be enabled. If the user hasn't explicitly
		// enabled a trigger, force a simple rising-edge trigger on the selected source.
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
		} else if newSettings || (samplingIntervalChanged && timeDependentTrigger) || !psControl.initialTriggerSet {
			err = psControl.sendTrigger()
			if err != nil {
				slog.Error("ETS prepare sendTrigger failed", "error", err)
				return err
			}
			psControl.initialTriggerSet = true
			psControl.lastTriggerSamplingInterval = psControl.SamplingTimeInterval
		}

		rawSampleCount := uint64(math.Round(float64(psControl.maxScreenTime) / psControl.SamplingTimeInterval))

		const maxEtsSamples = 250000
		if rawSampleCount > maxEtsSamples {
			slog.Debug("ETS sample count clamped to safe limit", "original", rawSampleCount, "max", maxEtsSamples)
			rawSampleCount = maxEtsSamples
		}
		if rawSampleCount > sampleCount {
			slog.Debug("ETS sample count clamped to memory segment limit", "original", rawSampleCount, "max", sampleCount)
			rawSampleCount = sampleCount
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
		// if err := psControl.setEtsBuffer(psControl.SampleCountRequired, 0); err != nil {
		// 	slog.Error("ETS prepare: setEtsBuffers failed", "error", err)
		// 	return err
		// }
		// here we need the npre and npost values
		// and have to modify triggerTimeOffset
		// etscallback also needs triggerTimeOffset
		psControl.NPre = uint64(math.Round(psControl.triggerSetting.XOffset / psControl.SamplingTimeInterval))
		if psControl.NPre > psControl.SampleCountRequired {
			psControl.NPre = psControl.SampleCountRequired
		}
		psControl.NPro = psControl.SampleCountRequired - psControl.NPre
		slog.Debug("pre", "SamplingTimeInterval", psControl.SamplingTimeInterval)
		psControl.XRoundError = psControl.triggerSetting.XOffset - psControl.SamplingTimeInterval*(float64(psControl.NPre)-1.0)
		slog.Debug("ets pre", "XRoundError", psControl.XRoundError)

		psControl.EtsBufferCallback(int(psControl.SampleCountRequired))
		slog.Debug("ets pre", "SampleCount", psControl.SampleCountRequired, "scopeScreenWidth", psControl.scopeScreenWidth)
		return nil
	}

	runBlock := func() error {
		slog.Debug("run", "psControl.NPre", psControl.NPre, "psControl.NPro", psControl.NPro,
			"psControl.timeBase", psControl.timeBase, "psControl.overSample", psControl.overSample)
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

		start = func() eventHandlerFunc {
			// Ensure hardware is stopped before prepare/memorySegments
			_ = psControl.Con.Stop()
			for psControl.numberOfEnabledChannels() == 0 {
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

		run = func() eventHandlerFunc {
			// Drain any stale callback before starting a new acquisition
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

		get = func() eventHandlerFunc {
			ticker := time.NewTicker(20 * time.Millisecond)
			defer ticker.Stop()
			timeout := time.After(1500 * time.Millisecond)

			for {
				select {
				case <-callbackChannel:
					goto ready
				case <-ticker.C:
					if ready, err := psControl.Con.LsReady(); err == nil && ready != 0 {
						goto ready
					}
				case <-timeout:
					// In ETS mode, auto-trigger is not supported by hardware.
					// If no trigger event occurred within 1.5s, stop acquisition
					// and restart so the UI remains responsive.
					_ = psControl.Con.Stop()
					return run
				case <-psControl.restartChannel:
					// Stop the hardware to cancel any in-progress acquisition.
					_ = psControl.Con.Stop()
					select {
					case <-callbackChannel:
					case <-time.After(200 * time.Millisecond):
					}
					return start
				case <-psControl.stopChannel:
					return nil
				}
			}

ready:
			if dt := minEtsRefreshTime - time.Since(psControl.refreshTime); dt > 0 {
				time.Sleep(dt)
			}
			if err := psControl.setEtsBuffer(psControl.SampleCountRequired, 0); err != nil {
				slog.Error("ETS prepare: setEtsBuffers failed", "error", err)
				psControl.DisplayStatus(err.Error(), Fatal)
				return nil
			}
			psControl.refreshTime = time.Now()
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
			_, _ = psControl.Con.SetEts(genericps.EtsOff, 40, 4)
		}
	}

	stateMachine()
	return idle
}
