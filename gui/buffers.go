package gui

import (
	"fmt"
	"psgo/ps2000a"
	"psgo/psi"
	"time"
)

func (ps *PsDesc) setBuffers(sampleCount int32) (err error) {
	var resp psi.Responder
	ps.signals.length = sampleCount
	//ps.signals.signal = make([][]int32, ps.ChannelCount)
	for channel := ps2000a.ChA; channel <= ps.MaxChannel; channel++ {
		if ps.buffers[channel] == nil || len(ps.buffers[channel]) < int(sampleCount) {
			ps.buffers[channel] = make([]int16, sampleCount)
			// ps.signals.signal[channel] = make([]int32, sampleCount)
			// ps.signals.signal[channel] = make([]int32, sampleCount)
		}
		if ps.signals.signal[channel] == nil || len(ps.signals.signal[channel]) < int(sampleCount) {
			ps.signals.signal[channel] = make([]int32, sampleCount)
		}
		err = ps.Con.Send(&psi.SetDataBufferMsg{Ch: channel, BufferIn: ps.buffers[channel], SegmentIndex: ps.segmentIndex, Mode: ps.ratioMode})
		if err != nil {
			fmt.Println("SetDataBuffers:", channel, err)
			return
		}
		resp, err = ps.Con.Receive()
		if err != nil || resp.Status() != nil {
			fmt.Println("SetDataBuffers:", channel, err, resp.Status())
			return
		}
	}
	return
}

func (ps *PsDesc) memorySegments() (sampleCount int32, err error) {
	err = ps.Con.Send(&psi.MemorySegmentsMsg{NSegments: 1})
	if err != nil {
		fmt.Println(err)
		return
	}
	resp, err := ps.Con.Receive()
	if err != nil || resp.Status() != nil {
		fmt.Println(err, resp.Status())
		return
	}
	sampleCount = resp.(*psi.MemorySegmentsRsp).NMaxSamples
	fmt.Println("runblock MemorySegmentsRsp:", sampleCount)
	return
}

func (ps *PsDesc) getData(sampleCount int32) (err error) {
	err = ps.Con.Send(&psi.GetValuesMsg{StartIndex: 0, ReqNumOfSamples: uint32(sampleCount),
		DownSampleRatio: 1, DownSampleRatioMode: ps.ratioMode, SegmentIndex: ps.segmentIndex})
	if err != nil {
		fmt.Println("runblock GetValues:", err)
		return
	}
	resp, err := ps.Con.Receive()
	if err != nil || resp.Status() != nil {
		fmt.Println(err, resp.Status())
		return
	}
	numOfSamples := resp.(*psi.GetValuesRsp).NumOfSamples
	since := time.Since(ps.refreshTime)
	if since > 10*time.Millisecond { //to avoid canvas cache error
		ps.screenLocker.Lock()
		// max := int32(0)
		// min := int32(0)
		for channel := ps2000a.ChA; channel <= ps.MaxChannel; channel++ {
			if ps.ChannelSettings[channel].Enabled {
				for i := 0; i < int(numOfSamples); i++ {
					v := int32(ps.buffers[channel][i])
					// if v > max {
					// 	max = v
					// }
					// if v < min {
					// 	min = v
					// }
					ps.signals.signal[channel][i] = (ps.signals.signal[channel][i] + ps.adcToMv(v, ps.ChannelSettings[channel].Range)) / 2
					// if i > 0 {
					// 	ps.signals.signal[channel][i] = (ps.signals.signal[channel][i] + ps.signals.signal[channel][i-1]) / 2
					// }
				}
			}
		}
		ps.screenLocker.Unlock()
		ps.Window.Canvas().Refresh(ps.raster)
		ps.refreshTime = time.Now()
	}
	return
} //end getdata
