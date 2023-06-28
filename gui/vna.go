//go:build notnow

package gui

import (
	"fmt"
	"strconv"

	"time"

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
	const autoTriggerMs = 1000
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
	ps.signals.signal = make([][]int32, 4)
	// ps.signals.signal[0] = make([]int32, sampleCount)
	// ps.signals.signal[1] = make([]int32, sampleCount)
	callbackChannel = make(chan bool)
	// func callbackBlock
	callbackBlock := func(handle int16, status int, param interface{}) {
		// Must not call ps2000 function from here
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
			ps.screenLocker.Lock()
			var err error
			for channel := ps2000a.ChA; channel <= ps2000a.ChD; channel++ {
				//	if ps.ChannelSettings[int(channel)].Enabled {
				if buffers[channel] == nil || len(buffers[channel]) != int(sampleCount) {
					buffers[channel] = make([]int16, sampleCount)
					ps.signals.signal[channel] = make([]int32, sampleCount)
					ps.signals.signal[channel] = make([]int32, sampleCount)
				}
				if ps.signals.signal[channel] == nil || len(ps.signals.signal[channel]) != int(sampleCount) {
					ps.signals.signal[channel] = make([]int32, sampleCount)
				}
				err = ps.SetDataBuffer(channel, buffers[channel], segmentIndex, ratioMode)
				if err != nil {
					fmt.Println("vna SetDataBuffers:", channel, err)
					return
				}
				//	}
			}
			noSamples, overflow, err := ps.GetValues(0, uint32(sampleCount), 1, ratioMode, 0)
			if err != nil {
				fmt.Println("VNA MaxValue:", err, " overflow:", overflow)
			}
			ps.Stop()

			for channel := ps2000a.ChA; channel <= ps2000a.ChD; channel++ {
				if ps.ChannelSettings[channel].Enabled {
					for i := 0; i < int(noSamples); i++ {
						v := int32(buffers[channel][i])
						if v > max {
							max = v
						}
						ps.signals.signal[channel][i] = ps.adcToMv(int32(buffers[channel][i]), ps.ChannelSettings[channel].Range)
					}
				}
			}
			fmt.Println("raw max=", max)
			ps.screenLocker.Unlock()
			since := time.Since(ps.refreshTime)
			if since > 10*time.Millisecond {
				ps.Window.Canvas().Refresh(ps.raster)
				ps.refreshTime = time.Now()
			}
			// TODO change settings if needed
			// timeIndisposed, err = ps.RunBlock(0, sampleCount, ps.timeBase, ps.overSample, 0, callbackBlock, nil)
			// if err != nil {
			// 	fmt.Println("vna RunBlock:", err)
			// 	fmt.Println("timeIndisposed:", timeIndisposed)
			// 	return
			// }
			timeIndisposed, err = ps.RunBlock(0, sampleCount, ps.timeBase, ps.overSample, 0, callbackBlock, nil)

		}
	}
	setChannels := func() {
		ranges := make([]int32, 32)
		lengthOfRanges, err := ps.GetChannelInformation(ps2000a.ChannelInfoRanges, 0, ranges, ps2000a.ChA)
		if err != nil {
			fmt.Println("vna GetChannelInformation:", err)
			return
		}
		// time.Sleep(1 * time.Second)
		fmt.Println("ranges=", ranges, "length=", lengthOfRanges)
		//		ps.Stop()
		err = ps.SetChannel(ps2000a.ChA, ps.ChannelSettings[ps2000a.ChA].Enabled, ps.ChannelSettings[ps2000a.ChA].CoupleType, ps.ChannelSettings[ps2000a.ChA].Range, 0.0)
		if err != nil {
			fmt.Println("vna SetChannel:", err)
			return
		}
		// time.Sleep(1 * time.Second)
		err = ps.SetChannel(ps2000a.ChB, ps.ChannelSettings[ps2000a.ChB].Enabled, ps.ChannelSettings[ps2000a.ChB].CoupleType, ps.ChannelSettings[ps2000a.ChB].Range, 0.0)
		if err != nil {
			fmt.Println("vna SetChannel:", err)
			return
		}
		// time.Sleep(1 * time.Second)
		err = ps.SetChannel(ps2000a.ChC, ps.ChannelSettings[ps2000a.ChC].Enabled, ps.ChannelSettings[ps2000a.ChC].CoupleType, ps.ChannelSettings[ps2000a.ChC].Range, 0.0)
		if err != nil {
			fmt.Println("vna SetChannel:", err)
			return
		}
		// time.Sleep(1 * time.Second)
		err = ps.SetChannel(ps2000a.ChD, ps.ChannelSettings[ps2000a.ChD].Enabled, ps.ChannelSettings[ps2000a.ChD].CoupleType, ps.ChannelSettings[ps2000a.ChD].Range, 0.0)
		if err != nil {
			fmt.Println("vna SetChannel:", err)
			return
		}
	} // end setChannels

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
		err = ps.SetTriggerChannelProperties(channelProperties, false, autoTriggerMs)
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
	} //end getData
	//	ps.SetDefaults()
	if ps.EtsMode {
		ps.overSample = 0
	} else {
		ps.overSample = 1
	}
	simpleTrigger := false
	//	ps.GenFreq = 1000000.0
	//ps.genOn()
	//	setChannels()
	//	_, err = ps.SetEts(ps2000a.EtsOff, 20, 4)
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
		if ps.etsSampleCountEntry.Text == "" {
			ps.etsSampleCountEntry.Text = "4096"
		}
		if ps.etsInterleave.Text == "" {
			ps.etsInterleave.Text = "4"
		}
		if ps.etsCycles.Text == "" {
			ps.etsCycles.Text = "20"
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
		timeInterval, err = ps.SetEts(ps2000a.EtsSlow, int16(cycles), int16(interleave)) //max 109 40 ?
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
		setChannels()
	}
	if simpleTrigger {
		err = ps.SetSimpleTrigger(true, ps2000a.ChA, 0, ps2000a.TriggerRaising, 0, autoTriggerMs)
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
	//ps.raster.Refresh()
}
