package gui

import (
	"fmt"
	"psgo/psi"
)

func (ps *PsDesc) setTimeBase(sampleCount int32) (maxSamples, timeInterval int32, err error) {
	var resp psi.Responder
	ps.timeBase = 0
	for {
		err = ps.Con.Send(&psi.GetTimebaseMsg{TimeBase: ps.timeBase, NumOfSamples: sampleCount, OverSample: ps.overSample, SegmentIndex: 0})
		if err != nil {
			fmt.Println("setTimeBase:", err)
			return
		}
		resp, err = ps.Con.Receive()
		if err != nil {
			fmt.Println("setTimeBase:", err, resp.Status())
			return
		}
		switch m := resp.(type) {
		case *psi.GetTimebaseRsp:
			maxSamples = m.MaxSamples
			timeInterval = m.TimeIntervalNanoseconds
			sampleCount = maxSamples / 32
		default:
			fmt.Println("Bad response type:", resp)
		}
		if resp.Status() == nil {
			break
		}
		ps.timeBase++
	}
	return
} //end setTimeBase
