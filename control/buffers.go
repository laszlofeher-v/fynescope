package control

import (
	"fynescope/genericps"
	"log/slog"
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
	err = psControl.Con.SetEtsTimeBuffer(psControl.EtsInBuffer)
	if err != nil {
		slog.Error("SetEtsTimeBufferMsg", "error:", err)
		return
	}
	return
}

func (psControl *PscDesc) setBuffers(sampleCount int32, segmentIndex uint32) (err error) {
	for chIndex := range psControl.receiveBuffer {
		if len(psControl.receiveBuffer[chIndex]) < int(sampleCount) {
			if cap(psControl.receiveBuffer[chIndex]) < int(sampleCount) {
				psControl.receiveBuffer[chIndex] = make([]int16, sampleCount)
			} else {
				psControl.receiveBuffer[chIndex] = psControl.receiveBuffer[chIndex][:sampleCount]
			}
		} else {
			psControl.receiveBuffer[chIndex] = psControl.receiveBuffer[chIndex][:sampleCount]
		}
		if len(psControl.displayBuffer[chIndex]) < int(sampleCount) {
			if cap(psControl.displayBuffer[chIndex]) < int(sampleCount) {
				psControl.displayBuffer[chIndex] = make([]float32, sampleCount)
			} else {
				psControl.displayBuffer[chIndex] = psControl.displayBuffer[chIndex][:sampleCount]
			}
		} else {
			psControl.displayBuffer[chIndex] = psControl.displayBuffer[chIndex][:sampleCount]
		}
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
	sampleCount, err = psControl.Con.MemorySegments(numberOfSegments)
	if err != nil {
		slog.Error("memorySegments", "error:", err)
		return
	}
	return
}

func (psControl *PscDesc) checkOverflow(overflow int16) {
	// Avoid spurious overflow on inactive channels
	chNames := ""
	if overflow != 0 {
		for i, ch := 0, 'A'; i < len(psControl.chEnabled); i, ch = i+1, ch+1 {
			if overflow&(1<<i) != 0 && psControl.chEnabled[i].Load() {
				chNames = chNames + string(ch) + " "
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
	downSampleRatio := psControl.downSampleRatio
	if downSampleRatio < 1 {
		downSampleRatio = 1
	}
	downSampleRatioMode := psControl.downSampleRatioMode
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
		psControl.RefreshCallback(psControl.receiveBuffer, psControl.triggerTimeOffset, psControl.XRoundError, psControl.SamplingTimeInterval)
	}
	return
}
