package control

import (
	"log/slog"
	"math"
	"fynescope/genericps"
	"time"
)

func streamMode(psControl *PscDesc) state {
	var (
		callbackChannel chan struct{}
		nextState       state = idle
		driverBuffer    [][]int16
		rollBuffer      [][]int16
	)

	callbackStream := func(handle int16, noOfSamples int32, startIndex uint32, overflow int16, triggeredAt uint32, triggered, autoStop int16, param any) (err error) {
		if noOfSamples > 0 {
			for chIndex := range psControl.receiveBuffer {
				if !psControl.chEnabled[chIndex].Load() {
					continue
				}

				// The driver writes new samples sequentially into driverBuffer[chIndex] starting at startIndex
				newSamples := driverBuffer[chIndex][startIndex : startIndex+uint32(noOfSamples)]

				// Append new samples to rolling buffer
				buf := rollBuffer[chIndex]
				n := int(noOfSamples)
				if n >= len(buf) {
					copy(buf, newSamples[n-len(buf):])
				} else {
					copy(buf, buf[n:])
					copy(buf[len(buf)-n:], newSamples)
				}

				// Copy back to display/receive buffer for the GUI/RefreshCallback
				copy(psControl.receiveBuffer[chIndex], buf)
			}

			psControl.checkOverflow(overflow)
			psControl.RefreshCallback(psControl.receiveBuffer, 0, 0, psControl.SamplingTimeInterval)
		}

		select {
		case callbackChannel <- struct{}{}:
		default:
		}
		return nil
	}

	prepare := func() (err error) {
		err = psControl.setEverything()
		if err != nil {
			return
		}

		psControl.refreshTime = time.Now()
		psControl.SampleCountRequired = int32(math.Round(psControl.scopeScreenWidth))
		if psControl.SampleCountRequired <= 0 {
			psControl.SampleCountRequired = 1000
		}

		// Setup scrolling display buffers
		rollBuffer = make([][]int16, len(psControl.receiveBuffer))
		for i := range rollBuffer {
			rollBuffer[i] = make([]int16, psControl.SampleCountRequired)
		}

		// Adjust receiveBuffer size
		for chIndex := range psControl.receiveBuffer {
			if len(psControl.receiveBuffer[chIndex]) < int(psControl.SampleCountRequired) {
				psControl.receiveBuffer[chIndex] = make([]int16, psControl.SampleCountRequired)
			} else {
				psControl.receiveBuffer[chIndex] = psControl.receiveBuffer[chIndex][:psControl.SampleCountRequired]
			}
		}

		// Setup separate driverBuffer and register it
		driverBuffer = make([][]int16, len(psControl.receiveBuffer))
		for chIndex := range driverBuffer {
			driverBuffer[chIndex] = make([]int16, psControl.SampleCountRequired)
			err = psControl.Con.SetDataBuffer(genericps.ChannelId(chIndex),
				driverBuffer[chIndex], 0, psControl.downSampleRatioMode)
			if err != nil {
				slog.Error("stream prepare SetDataBuffer", "err", err)
				return
			}
		}

		psControl.BufferCallback(int(psControl.SampleCountRequired))

		// Calculate sampling interval
		timeInterval := uint32(math.Round(psControl.maxScreenTime * 1e9 / float64(psControl.SampleCountRequired)))
		if timeInterval == 0 {
			timeInterval = 1
		}
		psControl.SamplingTimeInterval = float64(timeInterval) * 1e-9

		err = psControl.setTrigger()
		if err != nil {
			slog.Error("stream prepare SetTrigger", "error", err)
			return
		}

		callbackChannel = make(chan struct{}, 1)
		return
	}

	runStreaming := func() (err error) {
		timeInterval := uint32(math.Round(psControl.maxScreenTime * 1e9 / float64(psControl.SampleCountRequired)))
		if timeInterval == 0 {
			timeInterval = 1
		}

		_, err = psControl.Con.RunStreaming(timeInterval, genericps.TuNs, 0, uint32(psControl.SampleCountRequired),
			false, 1, psControl.downSampleRatioMode, uint32(psControl.SampleCountRequired))
		if err != nil {
			slog.Error("RunStreaming failed", "error", err)
			psControl.DisplayStatus(err.Error(), Fatal)
			return
		}
		return
	}

	stateMachine := func() {
		type eventHandlerFunc func() (nextFunc eventHandlerFunc)
		var (
			eventHandler eventHandlerFunc
			start        eventHandlerFunc
			run          eventHandlerFunc
		)

		start = func() eventHandlerFunc {
			for n := psControl.numberOfEnabledChannels(); n == 0; {
				select {
				case <-psControl.restartChannel:
					psControl.quit()
					return start
				case <-psControl.stopChannel:
					slog.Debug("stream start stop received")
					return nil
				}
			}
			err := prepare()
			if err != nil {
				psControl.DisplayStatus(err.Error(), Fatal)
				return nil
			}

			// Switch back to blockMode if screen time is fast (high frequency) or stream is disabled
			if psControl.maxScreenTime < StreamThreshold || !psControl.StreamEnabled.Load() {
				nextState = blockMode
				return nil
			}
			return run
		}

		run = func() eventHandlerFunc {
			err := runStreaming()
			if err != nil {
				psControl.DisplayStatus(err.Error(), Fatal)
				return nil
			}

			for {
				select {
				case <-psControl.restartChannel:
					slog.Debug("stream run restart received")
					psControl.quit()
					return start
				case <-psControl.stopChannel:
					slog.Debug("stream run stop received")
					psControl.quit()
					return nil
				default:
					err := psControl.Con.GetStreamingLatestValues(callbackStream, nil)
					if err != nil {
						slog.Error("GetStreamingLatestValues failed", "err", err)
						psControl.DisplayStatus(err.Error(), Fatal)
						return nil
					}
					time.Sleep(20 * time.Millisecond)
				}
			}
		}

		eventHandler = start
		for eventHandler != nil {
			eventHandler = eventHandler()
		}
		psControl.quit()
	}

	stateMachine()
	return nextState
}
