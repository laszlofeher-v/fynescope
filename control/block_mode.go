package control

import (
	"fynescope/genericps"
	"fynescope/settings"
	"log/slog"
	"math"
	"time"
)

const (
	minRefreshTime    = 50 * time.Millisecond
	LeftOut, RightOut = 4, 4
)

func blockMode(psControl *PscDesc) state {
	var (
		callbackChannel chan struct{}
		nextState       state = idle
	)

	callbackBlock := func(handle int16, status int, param any) {
		// Must not call ps2000 function from here
		select {
		case callbackChannel <- struct{}{}: // inform the state machine
		default:
		}
	} //end callbackBlock

	prepare := func() (err error) {
		// Disable ETS on hardware BEFORE configuring the trigger.
		// The SDK raises PICO_TRIGGER_ERROR if trigger properties are set
		// while ETS is still active (e.g. after returning from ETS mode).
		_, err = psControl.Con.SetEts(genericps.EtsOff, 40, 4)
		if err != nil {
			slog.Error("Set Ets off", "err", err)
			return
		}
		err = psControl.setEverything()
		if err != nil {
			return
		}
		psControl.refreshTime = time.Now()
		psControl.SampleCountRequired = int32(math.Round(psControl.scopeScreenWidth))
		if psControl.ipmode == settings.Sinc {
			psControl.SampleCountRequired = SincWMultiplier * int32(math.Round(psControl.scopeScreenWidth))
		} else {
			psControl.SampleCountRequired = int32(math.Round(psControl.scopeScreenWidth)) + LeftOut + RightOut
		}
		sampleCount, err := psControl.memorySegments(1)
		if sampleCount < psControl.SampleCountRequired {
			psControl.SampleCountRequired = sampleCount
		}
		if err != nil {
			slog.Error("runblock memorySegments", "err", err)
			return
		}
		psControl.overSample = 1 // not used
		psControl.downSampleRatio = 1
		psControl.downSampleRatioMode = genericps.RatioModeNone
		if psControl.hiResEnabled.Load() && psControl.ipmode != settings.Sinc {
			psControl.downSampleRatio = 256
			psControl.downSampleRatioMode = genericps.RatioModeAverage
		}

		tbInput := uint32(float64(psControl.maxScreenTime*1e9) / float64(psControl.SampleCountRequired*int32(psControl.downSampleRatio)))
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

		rawSampleCount := psControl.SampleCountRequired * int32(psControl.downSampleRatio)
		maxSampleCount, timeIntervalNanoseconds, err := psControl.getTimeBase(rawSampleCount)
		
		if rawSampleCount > maxSampleCount {
			psControl.downSampleRatio = uint32(maxSampleCount / psControl.SampleCountRequired)
			if psControl.downSampleRatio < 1 {
				psControl.downSampleRatio = 1
			}
			rawSampleCount = psControl.SampleCountRequired * int32(psControl.downSampleRatio)
		}
		if psControl.downSampleRatio <= 1 {
			psControl.downSampleRatioMode = genericps.RatioModeNone
			psControl.downSampleRatio = 1
		}

		psControl.SampleCountRequired = int32(math.Round(psControl.maxScreenTime/float64(timeIntervalNanoseconds*1e-9*float32(psControl.downSampleRatio)))) + 2
		if psControl.ipmode == settings.Sinc {
			psControl.SampleCountRequired = SincWMultiplier * psControl.SampleCountRequired
		} else {
			psControl.SampleCountRequired = 2 * psControl.SampleCountRequired
		}
		if psControl.SampleCountRequired > maxSampleCount/int32(psControl.downSampleRatio) || psControl.SampleCountRequired <= 0 {
			slog.Debug("samplecount decreased:", "SampleCount", psControl.SampleCountRequired, " to :", maxSampleCount/int32(psControl.downSampleRatio))
			psControl.SampleCountRequired = maxSampleCount / int32(psControl.downSampleRatio)
		}

		minSampleCount := int32(math.Round(psControl.scopeScreenWidth))
		if minSampleCount < 1024 {
			minSampleCount = 1024
		}
		if psControl.SampleCountRequired < minSampleCount {
			slog.Debug("samplecount increased to minimum:", "from", psControl.SampleCountRequired, "to", minSampleCount)
			psControl.SampleCountRequired = minSampleCount
		}
		psControl.SamplingTimeInterval = float64(timeIntervalNanoseconds) * 1e-9 * float64(psControl.downSampleRatio)
		err = psControl.setTrigger()
		if err != nil {
			slog.Error("runblock setTrigger:", "err", err)
			return
		}
		if psControl.SampleCountRequired == 0 {
			slog.Error("runblock setBuffers sampleCount == 0")
			return
		}
		err = psControl.setBuffers(psControl.SampleCountRequired, 0)
		if err != nil {
			slog.Error("runblock setBuffers:", "err", err)
			return
		}
		psControl.BufferCallback(int(psControl.SampleCountRequired)) // set buffer size
		if psControl.ipmode == settings.Sinc {
			displayRange := float64(psControl.SampleCountRequired) / SincWMultiplier
			leftRightRange := (float64(psControl.SampleCountRequired) - displayRange) / 2
			psControl.NPre = int32(math.Round(psControl.triggerSetting.XOffset/
				psControl.SamplingTimeInterval + leftRightRange))
			psControl.NPro = psControl.SampleCountRequired - psControl.NPre
			psControl.XRoundError = psControl.triggerSetting.XOffset -
				psControl.SamplingTimeInterval*(float64(psControl.NPre)-
					float64(leftRightRange))
		} else {
			psControl.NPre = int32(math.Round(psControl.triggerSetting.XOffset/psControl.SamplingTimeInterval))*int32(psControl.downSampleRatio) + LeftOut*int32(psControl.downSampleRatio)
			psControl.NPro = psControl.SampleCountRequired*int32(psControl.downSampleRatio) - psControl.NPre
			psControl.XRoundError = psControl.triggerSetting.XOffset - psControl.SamplingTimeInterval*float64(psControl.NPre/int32(psControl.downSampleRatio)-1-LeftOut)
			slog.Debug("pre", "SampleCount", psControl.SampleCountRequired)
			slog.Debug("pre", "SamplingTimeInterval", psControl.SamplingTimeInterval)
			slog.Debug("pre", "XOffset", psControl.triggerSetting.XOffset)
			slog.Debug("pre", "NPre", psControl.NPre)
			slog.Debug("pre", "NPro", psControl.NPro)
		}
		if psControl.NPro < 0 {
			psControl.NPro = 0
		}
		slog.Debug("pre", "XRoundError", psControl.XRoundError)
		callbackChannel = make(chan struct{}, 1)
		return
	} //end prepare

	runBlock := func() (err error) {
		_, err = psControl.Con.RunBlock(psControl.NPre, psControl.NPro,
			psControl.timeBase, psControl.overSample, 0, callbackBlock, nil)
		if err != nil {
			slog.Error("runblock msg", "error", err)
			slog.Error("runblock", "NPre", psControl.NPre, "NPro", psControl.NPro,
				"timeBase", psControl.timeBase, "overSample", psControl.overSample)
			psControl.DisplayStatus(err.Error(), Fatal)
			return
		}
		return
	}

	stateMachine := func() {
		type (
			eventHandlerFunc func() (nextFunc eventHandlerFunc)
		)
		var (
			eventHandler eventHandlerFunc
			start        eventHandlerFunc
			run          eventHandlerFunc
			get          eventHandlerFunc
		)
		// initialize block mode
		start = func() eventHandlerFunc {
			// if no channel is enabled then just wait
			for n := psControl.numberOfEnabledChannels(); n == 0; {
				select {
				case <-psControl.restartChannel:
					psControl.quit()
					return start
				case <-psControl.stopChannel:
					slog.Debug("runblock start stop received")
					return nil
				}
			}
			err := prepare()
			if err != nil {
				slog.Error("blockMode prepare failed", "err", err)
				psControl.DisplayStatus(err.Error(), Fatal)
				return nil
			}
			if psControl.maxScreenTime >= StreamThreshold && psControl.StreamEnabled.Load() {
				nextState = streamMode
				return nil
			}
			return run
		}

		//		run block mode
		run = func() eventHandlerFunc {
			err := runBlock() //
			if err != nil {
				psControl.DisplayStatus(err.Error(), Fatal)
				return nil
			}
			select {
			case <-psControl.restartChannel:
				psControl.quit()
				return start
			case <-psControl.stopChannel:
				slog.Info("runblock run stop received ")
				return nil
			case <-callbackChannel: // 				scope finished
				return get //						continue with get data
			}
		}

		// get data
		get = func() eventHandlerFunc {
			// slog.Debug("get")
			err := psControl.getData(psControl.SampleCountRequired, 0, false) //get data from the scope and send to the gui
			if err != nil {
				slog.Error("blockMode getData failed", "err", err)
				psControl.DisplayStatus(err.Error(), Fatal)
				return nil
			}
			since := time.Since(psControl.refreshTime)
			const minTime = minRefreshTime
			dt := minTime - since
			if dt > 0 { // to avoid canvas cache error
				time.Sleep(dt)
			}
			psControl.refreshTime = time.Now()
			if psControl.triggerSetting.Mode == Single {
				return nil
			}
			return run
		}

		//state machine
		eventHandler = start
		for eventHandler != nil {
			eventHandler = eventHandler()
		}
		psControl.quit()
	}

	// begin BlockMode
	stateMachine()
	return nextState
}
