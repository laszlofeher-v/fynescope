package control

import (
	"log/slog"
	"math"
	"fynescope/genericps"
	"fynescope/settings"
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
		// slog.Debug("runblock prepare")
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
		_, err = psControl.Con.SetEts(genericps.EtsOff, 40, 4)
		if err != nil {
			slog.Error("Set Ets off", "err", err)
			return
		}
		psControl.overSample = 1 // not used
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
		maxSampleCount, timeIntervalNanoseconds, err := psControl.getTimeBase(psControl.SampleCountRequired)
		// slog.Debug("blockmode", "timeIntervalNanoseconds", timeIntervalNanoseconds)
		psControl.SampleCountRequired = int32(math.Round(psControl.maxScreenTime/float64(timeIntervalNanoseconds*1e-9))) + 2
		if psControl.ipmode == settings.Sinc {
			psControl.SampleCountRequired = SincWMultiplier * psControl.SampleCountRequired
		} else {
			psControl.SampleCountRequired = 2 * psControl.SampleCountRequired
		}
		// slog.Debug("", "psControl.SampleCount", psControl.SampleCountRequired, "scopeScreenWidth", psControl.scopeScreenWidth)
		if psControl.SampleCountRequired > maxSampleCount || psControl.SampleCountRequired <= 0 {
			slog.Debug("samplecount decreased:", "SampleCount", psControl.SampleCountRequired, " to :", maxSampleCount)
			psControl.SampleCountRequired = maxSampleCount
		}

		minSampleCount := int32(math.Round(psControl.scopeScreenWidth))
		if minSampleCount < 1024 {
			minSampleCount = 1024
		}
		if psControl.SampleCountRequired < minSampleCount {
			slog.Debug("samplecount increased to minimum:", "from", psControl.SampleCountRequired, "to", minSampleCount)
			psControl.SampleCountRequired = minSampleCount
		}
		// slog.Debug("blockmode prepare", "samplecount:", psControl.SampleCount)
		psControl.SamplingTimeInterval = float64(timeIntervalNanoseconds) * 1e-9
		if psControl.SampleCountRequired == 0 {
			slog.Error("runblock setBuffers sampleCount == 0")
			return
		}
		err = psControl.setBuffers(psControl.SampleCountRequired, 0)
		if err != nil {
			slog.Error("runblock setBuffers:", "err", err)
			return
		}
		// delta := float64(psControl.scopeScreenWidth) / float64(psControl.SampleCountRequired)
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
			psControl.NPre = int32(math.Round(psControl.triggerSetting.XOffset/psControl.SamplingTimeInterval)) + LeftOut
			psControl.NPro = psControl.SampleCountRequired - psControl.NPre
			psControl.XRoundError = psControl.triggerSetting.XOffset - psControl.SamplingTimeInterval*float64(psControl.NPre-1-LeftOut)
			slog.Debug("pre", "SampleCount", psControl.SampleCountRequired)
			slog.Debug("pre", "SamplingTimeInterval", psControl.SamplingTimeInterval)
			slog.Debug("pre", "XOffset", psControl.triggerSetting.XOffset)
			slog.Debug("pre", "NPre", psControl.NPre)
			slog.Debug("pre", "NPro", psControl.NPro)
		}
		if psControl.NPro < 0 { //TODO Why?
			psControl.NPro = 0
		}

		// slog.Debug("pre", "delta", delta)
		slog.Debug("pre", "XRoundError", psControl.XRoundError)
		callbackChannel = make(chan struct{}, 1)
		return
	} //end prepare

	runBlock := func() (err error) {
		_, err = psControl.Con.RunBlock(psControl.NPre, psControl.NPro,
			psControl.timeBase, psControl.overSample, 0, callbackBlock, nil)
		if err != nil {
			slog.Error("runblock msg", "error", err)
			//TODO Why psControl.NPro<0 NPre=18474 NPro=-5362 timeBase=1 overSample=1
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
			// slog.Debug("runblock start 1", "goid", goid())
			for n := psControl.numberOfEnabledChannels(); n == 0; {
				select {
				case <-psControl.restartChannel:
					// slog.Debug("runblock start restart received", "goid", goid())
					psControl.quit()
					return start
				case <-psControl.stopChannel:
					slog.Debug("runblock start stop received")
					return nil
				}
			}
			// slog.Debug("runblock start 2", "goid", goid())
			err := prepare()
			if err != nil {
				psControl.DisplayStatus(err.Error(), Fatal)
				return nil
			}
			if psControl.maxScreenTime >= StreamThreshold && psControl.StreamEnabled.Load() {
				nextState = streamMode
				return nil
			}
			// slog.Debug("runblock start 3", "goid", goid())
			return run
		}

		//		run block mode
		run = func() eventHandlerFunc {
			// slog.Debug("run!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!")
			err := runBlock() //
			if err != nil {
				psControl.DisplayStatus(err.Error(), Fatal)
				return nil
			}
			select {
			case <-psControl.restartChannel:
				// slog.Debug("runblock run restart received  ******************", "goid", goid())
				psControl.quit()
				return start
			case <-psControl.stopChannel:
				slog.Info("runblock run stop received ")
				// psControl.quit()
				return nil
			case <-callbackChannel: // 				scope finished
				// slog.Info("callback received ")
				return get //							continue with get data
				// case <-time.After((2*time.Duration(timeIndisposedMs) + 2) * time.Millisecond):
				// 	slog.Error("runblock run timeout", "timeout", timeIndisposedMs*2)
				// 	return start
			}
		}

		// get data
		get = func() eventHandlerFunc {
			// slog.Debug("get")
			err := psControl.getData(psControl.SampleCountRequired, 0, false) //get data from the scope and send to the gui
			if err != nil {
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
