package gui

import (
	"fmt"
	"strconv"

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
		sampleCount       int32
		timeIndisposed    int32
		segmentIndex      uint32
		ratioMode         = ps2000a.RatioModeNone
		callbackChannel   chan bool
		buffers           [ps2000a.MaxChannelBuffers][]int16
		err               error
	)
	sampleCount, err = ps.MemorySegments(1)
	fmt.Println("sampleCount:", sampleCount)
	sampleCount = sampleCount / 32
	fmt.Println("sampleCount:", sampleCount)
	if err != nil {
		fmt.Println("vna MemorySegments:", err)
		return
	}
	ps.signals = &signalDesc{time: 0, length: sampleCount}
	ps.signals.signal = make([][]int32, 2)
	ps.signals.signal[0] = make([]int32, sampleCount)
	ps.signals.signal[1] = make([]int32, sampleCount)
	callbackChannel = make(chan bool)
	// func callbackBlock
	callbackBlock := func(handle int16, status int, param interface{}) {
		callbackChannel <- true
	}
	//Get data
	getData := func() {
		for {
			_, ok := <-callbackChannel
			if !ok {
				return
			}
			max := int32(0)
			// defer ps.genOff()
			// fmt.Println("Callback status:", status)
			ps.screenLocker.Lock()
			var err error
			ps.MaxValue, err = ps.MaximumValue()
			fmt.Println("........................................maxValue=", ps.MaxValue)
			if err != nil {
				fmt.Println("VNA getData:", err, " ps.MaxValue:", ps.MaxValue)
			}
			for channel := ps2000a.ChA; channel <= ps2000a.ChB; channel++ {
				if ps.ChannelSettings[int(channel)].Enabled {
					buffers[channel] = make([]int16, sampleCount)
					buffer := make([]int16, sampleCount)
					//	err = ps.SetDataBuffer(channel, buffers[channel], segmentIndex, ratioMode)
					err = ps.SetUnscaledDataBuffers(channel, buffers[channel], buffer, segmentIndex, ratioMode)
					if err != nil {
						fmt.Println("vna SetDataBuffers:", channel, err)
						return
					}
				}
			}
			noSamples, overflow, err := ps.GetValues(0, uint32(sampleCount), 1, ratioMode, 0)
			if err != nil {
				fmt.Println("VNA MaxValue:", err, " overflow:", overflow)
			}
			ps.Stop()
			// fmt.Println("noSamples=", noSamples, overflow, err)
			//		callbackChannel <- true
			//		fmt.Println("Callback cha range:", ps.ChannelSettings[ps2000a.ChA].Range)
			for channel := ps2000a.ChA; channel <= ps2000a.ChB; channel++ {
				if ps.ChannelSettings[channel].Enabled {
					// fmt.Println("Copy ch:", channel)
					// for i, v := range buffers[channel*2] {
					for i := 0; i < int(noSamples); i++ {
						v := int32(buffers[channel][i])
						if v > max {
							max = v
						}
						ps.signals.signal[channel][i] = ps.adcToMv(int32(buffers[channel][i]), ps.ChannelSettings[channel].Range)
						// ps.signals.signal[channel][i] = int32(buffers[channel][i])
					}
				}
			}
			fmt.Println("raw max=", max)
			ps.screenLocker.Unlock()
			ps.raster.Refresh()
			// TODO change settings if needed
			// timeIndisposed, err = ps.RunBlock(0, sampleCount, ps.timeBase, ps.overSample, 0, callbackBlock, nil)
			// if err != nil {
			// 	fmt.Println("vna RunBlock:", err)
			// 	fmt.Println("timeIndisposed:", timeIndisposed)
			// 	return
			// }
		}
	}
	go getData()
	setTrigger := func() (err error) {

		err = ps.SetTriggerChannelProperties(nil, false, 0)
		if err != nil {
			fmt.Println("vna SetTriggerChannelProperties:", err)
			return
		}
		triggerVoltage := int16(ps.mvToAdc(0, ps.ChannelSettings[ps2000a.ChA].Range))
		channelProperties = []ps2000a.TriggerChannelProperties{{triggerVoltage, uint16(1024),
			triggerVoltage, uint16(1024), ps2000a.ChA, ps2000a.Level}}
		err = ps.SetTriggerChannelProperties(channelProperties, false, 0)
		if err != nil {
			fmt.Println("vna SetTriggerChannelProperties:", err)
			fmt.Println("channelProperties:", channelProperties)
			return
		}
		conditions := []ps2000a.TriggerConditions{{ps2000a.CondTrue, ps2000a.CondDontCare, ps2000a.CondDontCare,
			ps2000a.CondDontCare, ps2000a.CondDontCare, ps2000a.CondDontCare, ps2000a.CondDontCare, ps2000a.CondDontCare}}
		err = ps.SetTriggerChannelConditions(conditions)
		if err != nil {
			fmt.Println("vna SetTriggerChannelConditions:", err)
			fmt.Println("conditions:", conditions)
			return
		}
		err = ps.SetTriggerChannelDirections(ps2000a.TriggerRaising, ps2000a.TriggerNone,
			ps2000a.TriggerNone, ps2000a.TriggerNone, ps2000a.TriggerNone, ps2000a.TriggerNone)
		if err != nil {
			fmt.Println("vna SetTriggerChannelDirections:", err)
			return
		}
		return
	}
	//	ps.SetDefaults()
	simpleTrigger := true
	//	ps.GenFreq = 1000000.0
	//ps.genOn()
	err = ps.SetChannel(ps2000a.ChA, true, ps.ChannelSettings[ps2000a.ChA].CoupleType, ps.ChannelSettings[ps2000a.ChA].Range, 0.0)
	ranges := make([]int32, 32)
	lengthOfRanges, err := ps.GetChannelInformation(ps2000a.ChannelInfoRanges, 0, ranges, ps2000a.ChA)
	if err != nil {
		fmt.Println("vna GetChannelInformation:", err)
		return
	}
	fmt.Println("ranges=", ranges, "length=", lengthOfRanges)

	err = ps.SetChannel(ps2000a.ChA, ps.ChannelSettings[ps2000a.ChA].Enabled, ps.ChannelSettings[ps2000a.ChA].CoupleType, ps.ChannelSettings[ps2000a.ChA].Range, 0.0)
	if err != nil {
		fmt.Println("vna SetChannel:", err)
		return
	}
	err = ps.SetChannel(ps2000a.ChB, ps.ChannelSettings[ps2000a.ChB].Enabled, ps.ChannelSettings[ps2000a.ChB].CoupleType, ps.ChannelSettings[ps2000a.ChB].Range, 0.0)
	if err != nil {
		fmt.Println("vna SetChannel:", err)
		return
	}
	_, err = ps.SetEts(ps2000a.EtsOff, 20, 4)
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
	if ps.EtsMode {
		if ps.etsInterleave.Text == "" {
			ps.etsInterleave.Text = "10"
		}
		if ps.etsSampleCountEntry.Text == "" {
			ps.etsSampleCountEntry.Text = "4096"
		}
		if ps.etsInterleave.Text == "" {
			ps.etsInterleave.Text = "10"
		}
		if ps.etsCycles.Text == "" {
			ps.etsCycles.Text = "40"
		}
		sc, err := strconv.Atoi(ps.etsSampleCountEntry.Text)
		if err != nil {
			ps.etsSampleCountEntry.Text = "4096"
			sc, err = strconv.Atoi(ps.etsSampleCountEntry.Text)
		}
		sampleCount = int32(sc)
		cycles, err := strconv.Atoi(ps.etsCycles.Text)
		if err != nil {
			ps.etsCycles.Text = "40"
			cycles, err = strconv.Atoi(ps.etsCycles.Text)
		}
		interleave, err := strconv.Atoi(ps.etsInterleave.Text)
		if err != nil {
			ps.etsInterleave.Text = "10"
			interleave, err = strconv.Atoi(ps.etsInterleave.Text)
		}
		timeInterval, err = ps.SetEts(ps2000a.EtsFast, int16(cycles), int16(interleave)) //max 109 40 ?
		if err != nil {
			fmt.Println("vna SetEts:", err)
			fmt.Println("timeInterval:", timeInterval)
			return
		}
		fmt.Println(" ETStimeInterval:", timeInterval)
		//TODO smapleCount may be wrong here
		ps.signals.time = float64(timeInterval) * float64(sampleCount)
		ps.signals.timeInterval = float64(timeInterval) / 1000.0
		etsTime := make([]int64, sampleCount)
		err = ps.SetEtsTimeBuffer(etsTime)
		ps.timeBase = 2
		if err != nil {
			fmt.Println("vna SetEtsTimeBuffer:", err)
			return
		}
		// } else {
		// _, err = ps.SetEts(ps2000a.EtsOff, 20, 4)
		// ps.timeBase = 0
		// for {
		// 	timeInterval, maxSamples, err = ps.GetTimebase(ps.timeBase, sampleCount, ps.overSample, 0)
		// 	if err == nil {
		// 		break
		// 	}
		// 	fmt.Println(err)
		// 	ps.timeBase++
		// }
		// fmt.Println("timeInterval:", timeInterval, " maxSamples:", maxSamples, "ps.timeBase=", ps.timeBase, " ps.overSample=", ps.overSample)
		// ps.signals.time = float64(timeInterval) * float64(maxSamples)
		// ps.signals.timeInterval = float64(timeInterval)
	}
	if simpleTrigger {
		err = ps.SetSimpleTrigger(true, ps2000a.ChA, 0, ps2000a.TriggerRaising, 0, 10)
	} else {
		err = setTrigger()
	}
	if err != nil {
		fmt.Println("vna SetTriggerChannelProperties:", err)
		return
	}

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
