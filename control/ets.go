package control

import (
	"fmt"
	"fynescope/genericps"
	"log/slog"
	"math"

	"time"
)

const minEtsRefreshTime = 100 * time.Millisecond

func (psControl *PscDesc) etsTimes(sampleTimeInPicoSeconds int32) (EtsCycles, EtsInterleave int16, err error) {
	switch psControl.Info {
	case "2407B", "2407SIM":
		// Specification for 2407B:
		// Sample time = 2000 / EtsInterleave
		// EtsCycles >= EtsInterleave
		// EtsCycles <= EtsInterleave * 10 + 9
		// EtsInterleave <= 40
		if sampleTimeInPicoSeconds >= 50 && sampleTimeInPicoSeconds <= 1000 {
			EtsInterleave = int16(math.Round(float64(2000 / sampleTimeInPicoSeconds)))
			EtsCycles = 2 * EtsInterleave
		} else {
			err = fmt.Errorf("etsTimes: sampleTimeInPicoSeconds %d must be between 50 and 1000 for %s", sampleTimeInPicoSeconds, psControl.Info)
		}
	default:
		err = fmt.Errorf("etsTimes: not implemented for variant %s", psControl.Info)
	}
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

		psControl.refreshTime = time.Now()
		psControl.SampleCountRequired = int32(math.Round(psControl.scopeScreenWidth))
		slog.Debug("prepare", "psControl.scopeScreenWidth", psControl.scopeScreenWidth)

		sampleCount, err := psControl.memorySegments(1)
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

		etsCycles, etsInterleave, err := psControl.etsTimes(int32(minSampleTimeInPicoseconds))
		slog.Debug("prepare", "etsCycles", etsCycles, "etsInterleave", etsInterleave, "err", err)
		if err != nil {
			slog.Error("ETS prepare: etsTimes failed", "error", err)
			return err
		}

		// Calculate valid timeBase for RunBlock
		psControl.overSample = 1
		tbInput := uint32(float64(psControl.maxScreenTime*1e9) / float64(psControl.SampleCountRequired))
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
		psControl.SampleCountRequired = int32(math.Round(float64(psControl.maxScreenTime) / psControl.SamplingTimeInterval))

		const maxEtsSamples = 250000
		if psControl.SampleCountRequired > maxEtsSamples {
			slog.Debug("ETS sample count clamped to safe limit", "original", psControl.SampleCountRequired, "max", maxEtsSamples)
			psControl.SampleCountRequired = maxEtsSamples
		}
		if psControl.SampleCountRequired > sampleCount {
			slog.Debug("ETS sample count clamped to memory segment limit", "original", psControl.SampleCountRequired, "max", sampleCount)
			psControl.SampleCountRequired = sampleCount
		}

		if psControl.SampleCountRequired <= 0 {
			err = fmt.Errorf("invalid sample count: %d (TimeInterval=%f, maxScreenTime=%f)",
				psControl.SampleCountRequired, psControl.SamplingTimeInterval, psControl.maxScreenTime)
			slog.Error("ETS prepare", "error", err)
			return err
		}
		slog.Debug("ETS", "SampleCount", psControl.SampleCountRequired)
		if err := psControl.setEtsBuffer(psControl.SampleCountRequired, 0); err != nil {
			slog.Error("ETS prepare: setEtsBuffers failed", "error", err)
			return err
		}
		// here we need the npre and npost values
		// and have to modify triggerTimeOffset
		// etscallback also needs triggerTimeOffset
		psControl.NPre = int32(math.Round(psControl.triggerSetting.XOffset / psControl.SamplingTimeInterval))
		psControl.NPro = psControl.SampleCountRequired - psControl.NPre
		if psControl.NPro < 0 {
			psControl.NPro = 0
		}
		slog.Debug("pre", "SamplingTimeInterval", psControl.SamplingTimeInterval)
		psControl.XRoundError = psControl.triggerSetting.XOffset - psControl.SamplingTimeInterval*float64(psControl.NPre-1)
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
			if err := runBlock(); err != nil {
				psControl.DisplayStatus(err.Error(), Fatal)
				return nil
			}
			return get
		}

		get = func() eventHandlerFunc {
			select {
			case <-callbackChannel:
				// Data acquisition finished
			case <-psControl.restartChannel:
				psControl.quit()
				return start
			case <-psControl.stopChannel:
				return nil
			}
			// Throttling to avoid canvas cache issues
			if dt := minEtsRefreshTime - time.Since(psControl.refreshTime); dt > 0 {
				time.Sleep(dt)
			}
			psControl.refreshTime = time.Now()

			if err := psControl.getData(psControl.SampleCountRequired, 0, true); err != nil {
				slog.Error("ETS get data failed", "error", err)
				psControl.DisplayStatus(err.Error(), Fatal)
				return nil
			}
			return get
		}

		for handler := start; handler != nil; {
			handler = handler()
		}

		slog.Debug("ETS quit")
		psControl.quit()
	}

	stateMachine()
	return idle
}
