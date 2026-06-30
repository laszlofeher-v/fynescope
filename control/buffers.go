package control

import (
	"fynescope/genericps"
	// "fmt"
	"log/slog"
	// "fynescope/psi"
)

func (psControl *PscDesc) setEtsBuffer(sampleCount int32, segmentIndex uint32) (err error) {
	err = psControl.setBuffers(sampleCount, segmentIndex) // set data buffers
	if err != nil {
		slog.Error("SetEtsBuffers", "error:", err)
		return
	}
	if psControl.EtsInBuffer == nil {
		psControl.EtsInBuffer = make([]int64, sampleCount)
	} else if len(psControl.EtsInBuffer) < int(sampleCount) {
		if cap(psControl.EtsInBuffer) < int(sampleCount) {
			psControl.EtsInBuffer = make([]int64, sampleCount)
		} else {
			psControl.EtsInBuffer = psControl.EtsInBuffer[:sampleCount]
		}
	} else {
		psControl.EtsInBuffer = psControl.EtsInBuffer[:sampleCount]
	}
	// _, err = psControl.Con.Send(&genericps.SetEtsTimeBufferMsg{Buffer: psControl.EtsInBuffer})
	err = psControl.Con.SetEtsTimeBuffer(psControl.EtsInBuffer)
	if err != nil {
		slog.Error("SetEtsTimeBufferMsg", "error:", err)
		return
	}
	return
}

func (psControl *PscDesc) setBuffers(sampleCount int32, segmentIndex uint32) (err error) {
	// var resp psi.Responder
	for chIndex := range psControl.receiveBuffer {
		if len(psControl.receiveBuffer[chIndex]) < int(sampleCount) {
			if cap(psControl.receiveBuffer[chIndex]) < int(sampleCount) {
				// log.Println("*************Allocate r buffer 2 sampleCount:", sampleCount, cap(channel.ReceiveBuffer), len(channel.ReceiveBuffer))
				psControl.receiveBuffer[chIndex] = make([]int16, sampleCount)
			} else {
				psControl.receiveBuffer[chIndex] = psControl.receiveBuffer[chIndex][:sampleCount]
			}
		} else {
			psControl.receiveBuffer[chIndex] = psControl.receiveBuffer[chIndex][:sampleCount]
		}
		if len(psControl.displayBuffer[chIndex]) < int(sampleCount) {
			if cap(psControl.displayBuffer[chIndex]) < int(sampleCount) {
				// log.Println("*************Allocate d buffer 2 sampleCount:", sampleCount, cap(channel.ReceiveBuffer), len(channel.ReceiveBuffer))
				psControl.displayBuffer[chIndex] = make([]float32, sampleCount)
			} else {
				psControl.displayBuffer[chIndex] = psControl.displayBuffer[chIndex][:sampleCount]
			}
		} else {
			psControl.displayBuffer[chIndex] = psControl.displayBuffer[chIndex][:sampleCount]
		}

		// _, err = psControl.Con.Send(&genericps.SetDataBufferMsg{Ch: genericps.ChannelId(chIndex), BufferIn: psControl.receiveBuffer[chIndex][:sampleCount], SegmentIndex: segmentIndex, Mode: psControl.downSampleRatioMode})
		err = psControl.Con.SetDataBuffer(genericps.ChannelId(chIndex),
			psControl.receiveBuffer[chIndex][:sampleCount], segmentIndex,
			psControl.downSampleRatioMode)
		if err != nil {
			slog.Error("SetDataBuffers", "buffers:", psControl.receiveBuffer[chIndex], "error:", err)
			return
		}
	}
	return
}

func (psControl *PscDesc) memorySegments(numberOfSegments uint32) (sampleCount int32, err error) {
	// resp, err = psControl.Con.Send(&genericps.MemorySegmentsMsg{NSegments: numberOfSegments})
	sampleCount, err = psControl.Con.MemorySegments(numberOfSegments)
	if err != nil {
		slog.Error("memorySegments", "error:", err)
		return
	}
	// sampleCount = resp.(*genericps.MemorySegmentsRsp).NMaxSamples
	// if n, _ := scp.numberOfEnabledChannels(); sampleCount == 0 && n == 0 {
	// 	sampleCount = minSampleCount
	// }
	return
}

func (psControl *PscDesc) checkOverflow(overflow int16) {
	// Avoid spurious overflow on inactive channels
	chNames := ""
	if overflow != 0 {
		for i, ch := 0, 'A'; i < len(psControl.chEnabled); i, ch = i+1, ch+1 {
			if overflow&(1<<i) != 0 && psControl.chEnabled[i].Load() {
				chNames = chNames + string(ch) + " "
				// slog.Debug("chk", "overflow", overflow, "i", i, "ch", ch, "chNames", chNames)
			}
		}
		s := "Overflow error on channel"
		switch {
		case len(chNames) > 2:
			s = s + "s"
		case len(chNames) == 0:
			return
		}
		s = s + ":"
		psControl.DisplayStatus(s+chNames, Warning)
	}
}

func (psControl *PscDesc) getData(sampleCount int32, segmentIndex uint32, ets bool) (err error) {
	var overflow int16
	var downSampleRatio uint32 = 1
	var downSampleRatioMode = genericps.RatioModeNone

	if ets {
		downSampleRatio = psControl.downSampleRatio
		downSampleRatioMode = psControl.downSampleRatioMode
	}
	psControl.numOfSamplesAcquired, overflow, err = psControl.Con.GetValues(0, uint32(sampleCount), downSampleRatio, downSampleRatioMode, segmentIndex)
	if err != nil {
		return
	}
	psControl.checkOverflow(overflow)

	if ets {
		psControl.triggerTimeOffset = 0
		psControl.RefreshEtsCallback(psControl.receiveBuffer, psControl.EtsInBuffer[:psControl.numOfSamplesAcquired], psControl.XRoundError)
	} else {
		triggerTimeOffset, timeUnits, err := psControl.Con.GetTriggerTimeOffset64(segmentIndex)
		if err != nil {
			slog.Error("getData trigger offset:", "error:", err)
			return err
		}
		psControl.triggerTimeOffset = int64(float64(triggerTimeOffset) *
			(genericps.TimeUnitToVal(timeUnits) / genericps.TimeUnitToVal(genericps.TuFs)))
		// slog.Error("getData", "timeUnits", timeUnits, "v", genericps.TimeUnitToVal(timeUnits))
		// slog.Debug("getData", "XRoundError", float64(psControl.XRoundError))
		// slog.Debug("getData", "psControl.triggerTimeOffset:", psControl.triggerTimeOffset)
		psControl.RefreshCallback(psControl.receiveBuffer, psControl.triggerTimeOffset, psControl.XRoundError, psControl.SamplingTimeInterval)
		// psControl.RefreshCallback(psControl.receiveBuffer, 0, psControl.SamplingTimeInterval)
	}
	return
}
