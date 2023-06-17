package gui

import (
	"fmt"
	//	"math"
	"psgo/ps2000a"
)

func (ps *PsDesc) vna() {
	// Disable trigger on CHA and CHB
	// Set generator starting frequency
	// Set time base  (sample rate)
	//TODO if sample interval is not short enough for phase calculation set
	// ETS mode
	// Run stream data handler on CHA and CHB
	// for frequency interval
	// set optimal V/divs
	// measure frequency
	//    calculate bode points
	//    step frequency
	var (
		channelProperties []ps2000a.TriggerChannelProperties
		timeInterval      int32
		maxSamples        int32
		sampleCount       int32 = 1024 * 1024 * 2
		timeIndisposed    int32
		segmentIndex      uint32
		ratioMode         = ps2000a.RatioModeNone
		//		callbackChannel   chan bool
		buffers [ps2000a.MaxChannelBuffers][]int16
	)

	// func callbackBlock
	//mv := make([]int32, sampleCount)
	ps.signals = &signalDesc{time: 0, length: sampleCount}
	ps.signals.signal = make([][]int32, 2)
	ps.signals.signal[0] = make([]int32, sampleCount)
	ps.signals.signal[1] = make([]int32, sampleCount)
	callbackBlock := func(handle int16, status int, param interface{}) {
		// defer ps.genOff()
		fmt.Println("Callback status:", status)
		noSamples, overflow, err := ps.GetValues(0, uint32(sampleCount), 1, ratioMode, 0)
		fmt.Println("noSamples=", noSamples, overflow, err)
		//		callbackChannel <- true
		for channel := ps2000a.ChA; channel <= ps2000a.ChB; channel++ {
			if ps.ChannelSettings[channel].Enabled {
				fmt.Println("Copy ch:", channel)
				for i, v := range buffers[channel*2] {
					// //fmt.Println("Hahó copy!")
					// v = 1000
					// if v != 0 {
					// 	fmt.Println(v)
					// } else {
					// 	v = 1000
					// }
					// ps.signals.signal[channel][i] = int32(v)
					ps.signals.signal[channel][i] = ps.adcToMv(int32(v), ps.ChannelSettings[channel].Range)
				}
				// fmt.Println()
				// fmt.Println(ps.signals.signal[channel][100:1000])
			}
			//			ps.drawSignal(channel, mv, timeInterval)
		}
		ps.raster.Refresh()
	}
	//	ps.SetDefaults()
	channelProperties = make([]ps2000a.TriggerChannelProperties, 2)
	channelProperties[1].Channel = 1
	err := ps.SetTriggerChannelProperties(channelProperties, false, 0)
	if err != nil {
		fmt.Println("vna SetTriggerChannelProperties:", err)
		return
	}
	ps.GenFreq = 1000.0
	ps.genOn()
	err = ps.SetChannel(ps2000a.ChA, true, ps.ChannelSettings[ps2000a.ChA].CoupleType, ps.ChannelSettings[ps2000a.ChA].Range, 0.0)
	if err != nil {
		fmt.Println("vna SetChannel:", err)
		return
	}
	err = ps.SetChannel(ps2000a.ChB, true, ps.ChannelSettings[ps2000a.ChB].CoupleType, ps.ChannelSettings[ps2000a.ChB].Range, 0.0)
	if err != nil {
		fmt.Println("vna SetChannel:", err)
		return
	}
	// ranges := make([]int32, 32)
	// lengthOfRanges, err := ps.GetChannelInformation(ps2000a.ChannelInfoRanges, 0, ranges, ps2000a.ChA)
	// if err != nil {
	// 	fmt.Println("vna GetChannelInformation:", err)
	// 	return
	// }
	// fmt.Println("ranges=", ranges, "length=", lengthOfRanges)
	for channel := ps2000a.ChA; channel <= ps2000a.ChB; channel++ {
		if ps.ChannelSettings[int(channel)].Enabled {
			buffers[channel*2] = make([]int16, sampleCount)
			buffers[channel*2+1] = make([]int16, sampleCount)
			err = ps.SetDataBuffers(channel, buffers[channel*2], buffers[channel*2+1], segmentIndex, ratioMode)
			if err != nil {
				fmt.Println("vna SetDataBuffers:", channel, err)
				return
			}
		}
	}
	ps.timeBase = 0
	for {
		timeInterval, maxSamples, err = ps.GetTimebase(ps.timeBase, sampleCount, ps.overSample, 0)
		if err == nil {
			break
		}
		fmt.Println(err)
		ps.timeBase++
	}
	fmt.Println("timeInterval:", timeInterval, " maxSamples:", maxSamples, "ps.timeBase=", ps.timeBase, " ps.overSample=", ps.overSample)
	ps.signals.time = float64(timeInterval) * float64(maxSamples)
	ps.signals.timeInterval = float64(timeInterval)
	fmt.Println("ps.signals.time=", ps.signals.time)
	timeIndisposed, err = ps.RunBlock(0, sampleCount, ps.timeBase, ps.overSample, 0, callbackBlock, nil)
	if err != nil {
		fmt.Println("vna RunBlock:", err)
		fmt.Println("timeIndisposed:", timeIndisposed)
		return
	}
	fmt.Println("timeIndisposed:", timeIndisposed)
	// func printResult
	ps.raster.Refresh()
}
