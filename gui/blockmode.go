package gui

import (
	"fmt"
	"psgo/ps2000a"
	"psgo/psi"

	//	"strconv"
	//	"sync"
	"time"
)

var loopMutex chan struct{} = make(chan struct{}, 1)

func (ps *PsDesc) BlockMode() {
	var (
		err             error
		callbackChannel chan struct{}
	)

	callbackBlock := func(handle int16, status int, param interface{}) {
		// Must not call ps2000 function from here
		// fmt.Println("Callback ")
		select {
		case callbackChannel <- struct{}{}:
		// fmt.Println("Callback message sent ")
		case <-time.After(1000 * time.Millisecond):
			fmt.Println("Timeout send Callback msg")
		default:
			fmt.Println("Could not send Callback msg")
		}
	} //end callbackBlock

	setChannels := func() {
		// ranges := make([]int32, 32)
		// lengthOfRanges, err := ps.GetChannelInformation(ps2000a.ChannelInfoRanges, 0, ranges, ps2000a.ChA)
		// if err != nil {
		// 	fmt.Println("runblock GetChannelInformation:", err)
		// 	fmt.Println("ranges=", ranges, "length=", lengthOfRanges)
		// 	return
		// }
		// time.Sleep(1 * time.Second)
		//		fmt.Println("ranges=", ranges, "length=", lengthOfRanges)
		//		ps.Stop()
		for ch := ps2000a.ChA; ch <= ps.MaxChannel; ch++ {
			// err = ps.SetChannel(ch, ps.ChannelSettings[ch].Enabled, ps.ChannelSettings[ch].CoupleType, ps.ChannelSettings[ch].Range, ps.ChannelSettings[ch].Offset)
			err = ps.Con.Send(&psi.SetChannelMsg{Channel: ch, Enabled: ps.ChannelSettings[ch].Enabled,
				CouplingType: ps.ChannelSettings[ch].CoupleType, VoltageRange: ps.ChannelSettings[ch].Range,
				AnalogOffset: ps.ChannelSettings[ch].Offset})
			if err != nil {
				fmt.Println("runblock SetChannel:", err)
				return
			}
			resp, err := ps.Con.Receive()
			if err != nil || resp.Status() != nil {
				fmt.Println(err, resp.Status())
				return
			}
		}
		// time.Sleep(1 * time.Second)
	} // end setChannels

	prepare := func() (sampleCount int32, err error) {
		sampleCount, err = ps.memorySegments()
		if err != nil {
			fmt.Println("runblock MemorySegmentsRsp:", sampleCount)
			return
		}
		fmt.Println("runblock MemorySegmentsRsp:", sampleCount)
		//sampleCount = sampleCount / 32
		sampleCount = 65536
		ps.overSample = 1
		setChannels()
		maxSamples, timeInterval, err := ps.setTimeBase(sampleCount)
		fmt.Println("timeInterval:", timeInterval, " maxSamples:", maxSamples, "ps.timeBase=", ps.timeBase, " ps.overSample=", ps.overSample)
		// ps.signals = &signalDesc{time: 0}
		ps.signals.time = float64(timeInterval) * float64(maxSamples)
		ps.signals.timeInterval = float64(timeInterval)
		simpleTrigger := false
		if simpleTrigger {
			err = ps.setSimpleTrigger()
		} else {
			err = ps.setTrigger()
		}
		if err != nil {
			fmt.Println("runblock SetTrigger:", err)
			return
		}
		fmt.Println("RunBlockMsg....... ps.signals.time=", ps.signals.time, "samplecount:", sampleCount)
		ps.setBuffers(sampleCount)
		callbackChannel = make(chan struct{})
		return
	} //end prepare

	stateMachine := func(sampleCount int32) {
		type (
			eventHandlerFunc func() eventHandlerFunc
		)
		var (
			eventHandler eventHandlerFunc
			init         eventHandlerFunc
		)
		init = func() eventHandlerFunc {
			select {
			// wait for events possible timeouts
			}
			return init
		}

		eventHandler = init
	exitLoop:
		for {
			eventHandler = eventHandler()
		}
	}

	loop := func(sampleCount int32) {
		defer func() { <-loopMutex }()
		defer func() { <-ps.runLock }()
	exitLoop:
		for {
			select {
			case ps.runLock <- struct{}{}:
			default:
				fmt.Println("Wait for runLock")
				ps.runLock <- struct{}{}
				fmt.Println("Got runLock")
				prepare()
			}
			err = ps.Con.Send(&psi.RunBlockMsg{NumOfPreTriggerSamples: 0, NumOfPostTriggerSamples: sampleCount, TimeBase: ps.timeBase, OverSample: ps.overSample,
				SegmentIndex: 0, LpBlockReadyGoPar: callbackBlock, Param: nil})
			if err != nil {
				fmt.Println(err)
				break exitLoop
			}
			resp, err := ps.Con.Receive()
			if err != nil || resp.Status() != nil {
				fmt.Println(err, resp.Status())
				break exitLoop
			}
			select {
			case <-ps.stopChannel:
				fmt.Println("Stop received")
				break exitLoop
			case <-callbackChannel:
			}
			ps.getData(sampleCount)
			<-ps.runLock
			// select {
			// case <-ps.stopChannel:
			// 	fmt.Println("Stop received")
			// 	break exitLoop
			// default:
			// }
		}
		fmt.Println("Exit from loop")
	} // end loop

	// begin BlockMode
	enabledChCount := 0
	for channel := ps2000a.ChA; channel <= ps.MaxChannel; channel++ {
		if ps.ChannelSettings[channel].Enabled {
			enabledChCount++
		}
	}
	if enabledChCount == 0 {
		//		return
	}
	select {
	case loopMutex <- struct{}{}:
	default:
		return
	}
	sampleCount, err := prepare()
	if err != nil {
		<-loopMutex
		return
	}
	go loop(sampleCount)
}
