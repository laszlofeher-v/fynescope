package ps2000acon

import (
	"bytes"
	"fmt"
	"os"
	"psgo/ps2000a"
	"strings"

	"golang.org/x/term"
)

const (
	blockFile     = "block.txt"
	digiBlockFile = "digiblock.txt"
	streamFile    = "stream.txt"
	bufferSize    = 1024
)

type Mode int

const (
	Analogue Mode = iota
	Digital
	Aggregated
	Mixed
)

var (
	terminal chan byte
)

// var (
// 	scaleVoltages bool   = true
// 	timeBase      uint32 = 8
// 	overSample    int16  = 1
// )

type (
	psHandler interface {
		ps2000a.Picoscope2000Handler
	}
	PsDesc struct {
		ps2000a.PsDesc
		FirstRange      ps2000a.RangeEnum
		LastRange       ps2000a.RangeEnum
		SignalGenerator bool
		Ets             bool
		ChannelCount    ps2000a.NumOfChannelEnum
		MaxChannel      ps2000a.ChannelId
		MaxValue        int16
		ChannelSettings [ps2000a.MaxChannels]ps2000a.ChannelDesc
		DigitalPorts    int16
		AwgBufferSize   int16
		AwgDACFrequency float64
		scaleVoltages   bool
		timeBase        uint32
		overSample      int16
		GenFreq         float64
		GenFreqStep     float64
		GenFreqDelay    float64
	}
)

var (
	inputRanges []int32 = []int32{10,
		20,
		50,
		100,
		200,
		500,
		1000,
		2000,
		5000,
		10000,
		20000,
		50000}
)

func (ps *PsDesc) SetDefaults() (err error) {
	_, err = ps.SetEts(ps2000a.EtsOff, 0, 0)
	if err != nil {
		return
	}
	for channel := ps2000a.ChA; channel < ps.MaxChannel; channel++ {
		err = ps.SetChannel(channel,
			ps.ChannelSettings[channel].Enabled,
			ps.ChannelSettings[channel].CoupleType,
			ps.ChannelSettings[channel].Range, 0.0)
		if err != nil {
			return
		}
	}
	return
}

func (ps *PsDesc) SetTrigger(channelProperties []ps2000a.TriggerChannelProperties,
	triggerConditions []ps2000a.TriggerConditions,
	directions *ps2000a.TriggerDirections,
	pwq *ps2000a.Pwq,
	delay uint32,
	auxOutputEnabled bool,
	autoTriggerMs int32,
	digitalDirections []ps2000a.DigitalChannelDirections) (err error) {

	err = ps.SetTriggerChannelProperties(channelProperties, auxOutputEnabled, autoTriggerMs)
	if err != nil {
		return
	}
	err = ps.SetTriggerChannelConditions(triggerConditions)
	if err != nil {
		return
	}
	err = ps.SetTriggerChannelDirections(directions.ChannelA,
		directions.ChannelB, directions.ChannelC, directions.ChannelD,
		directions.Ext, directions.Aux)
	if err != nil {
		return
	}
	err = ps.SetTriggerDelay(delay)
	if err != nil {
		return
	}
	if ps.DigitalPorts > 0 {
		err = ps.SetTriggerDigitalPortProperties(digitalDirections)
		if err != nil {
			return
		}

	}
	return
}

func (ps *PsDesc) adcToMv(raw int32, ch ps2000a.RangeEnum) int32 {
	return raw * int32(inputRanges[int(ch)]) / int32(ps.MaxValue)
}

func (ps *PsDesc) BlockDataHandler(text string, offset int32,
	mode Mode, etsModeSet bool) (err error) {
	var (
		buffers [ps2000a.MaxChannelBuffers][]int16
		// digiBuffers [ps2000a.MaxDigitalPorts][]int16
		sampleCount     int32 = bufferSize
		segmentIndex    uint32
		ratioMode       = ps2000a.RatioModeNone
		etsTime         []int64
		timeInterval    int32
		maxSamples      int32
		overflow        int16
		noSamples       uint32
		callbackChannel chan bool
		timeIndisposed  int32
	)

	// func callbackBlock
	callbackBlock := func(handle int16, status int, param interface{}) {
		callbackChannel <- true
	}

	// func setBuffers
	setBuffers := func() (err error) {
		if mode == Analogue || mode == Mixed {
			for channel := ps2000a.ChA; channel <= ps.MaxChannel; channel++ {
				if ps.ChannelSettings[int(channel)].Enabled {
					buffers[channel*2] = make([]int16, sampleCount)
					buffers[channel*2+1] = make([]int16, sampleCount)
					err = ps.SetDataBuffers(channel, buffers[channel*2], buffers[channel*2+1], segmentIndex, ratioMode)
					if err != nil {
						fmt.Printf("ps2000aSetDataBuffers(channel %d) ------ %v \n \r", channel, err)
						return
					}
					fmt.Printf("ps2000aSetDataBuffers(channel %d) ------  \n \r", channel)
				}
			}
		}
		return
	}

	// func setEtsTimeBuffer
	setTime := func() (err error) {
		if mode == Analogue && etsModeSet {
			etsTime = make([]int64, sampleCount)
			err = ps.SetEtsTimeBuffer(etsTime)
			if err == nil {
				return
			}
		}
		for {
			timeInterval, maxSamples, err = ps.GetTimebase(ps.timeBase, sampleCount, ps.overSample, 0)
			if err == nil {
				break
			}
			ps.timeBase++
		}
		if !etsModeSet {
			fmt.Print("\n\rTimebase:", ps.timeBase, " SampleInterval:", timeInterval, " Oversample:", ps.overSample, " Maxsamples:", maxSamples, "\n\r")
		}
		return
	}

	// func printResult
	printResult := func() (err error) {
		fmt.Print("\n\r timIndisposed:", timeIndisposed, "\n\r")
		noSamples, overflow, err = ps.GetValues(0, uint32(sampleCount), 10, ratioMode, 0)
		if err != nil {
			return
		}
		if mode == Analogue || mode == Mixed {
			if ps.scaleVoltages {
				fmt.Printf("Channels are in mv\n\n")
			} else {
				fmt.Printf("Channels are in ADC Counts\n\n")
			}
			for channel := ps2000a.ChA; channel <= ps.MaxChannel; channel++ {
				if ps.ChannelSettings[channel].Enabled {
					fmt.Print(channel)
				}
			}
			fmt.Println("\r")
		}
		for o := offset; o < offset+10; o++ {
			if mode == Analogue || mode == Mixed {
				for channel := ps2000a.ChA; channel <= ps.MaxChannel; channel++ {
					if ps.ChannelSettings[channel].Enabled {
						if ps.scaleVoltages {
							fmt.Printf("  %6d        \n\r", ps.adcToMv(int32(buffers[channel*2][o]), ps.ChannelSettings[channel].Range))
						} else {
							fmt.Printf("Channels are in ADC Counts\n\n")
						}
					}
				}
				fmt.Println()
			}

		}
		fmt.Print("\n\r noSamples:", noSamples, "\n\r")
		fmt.Print("\n\r overflow:", overflow, "\n\r")
		return
	}

	// func BlockDataHandler
	err = setBuffers()
	if err != nil {
		return
	}
	err = setTime()
	if err != nil {
		return
	}
	// err = setDigitalBuffer()
	// if err != nil {
	// 	return
	// }
	defer ps.Stop()
	callbackChannel = make(chan bool)
	timeIndisposed, err = ps.RunBlock(0, sampleCount, ps.timeBase, ps.overSample, 0, callbackBlock, nil)
	select {
	case <-terminal:
		fmt.Println("data collection aborted, press a key\n")
		<-terminal
	case <-callbackChannel:
		printResult()
	}
	return
}

func (ps *PsDesc) CollectBlockImmediate() (err error) {
	var (
		pulseWidth ps2000a.Pwq
		directions ps2000a.TriggerDirections
	)
	fmt.Print("Collect block immediate\n\r")
	fmt.Printf("Data is written to disk file (%v)\n \r", blockFile)
	fmt.Print("Press a key to start...\n \r")
	<-terminal
	err = ps.SetDefaults()
	if err != nil {
		return
	}
	// /* Trigger disabled	*/
	err = ps.SetTrigger(nil, nil, &directions, &pulseWidth, 0, false, 0, nil)
	if err != nil {
		return
	}
	err = ps.BlockDataHandler("\nFirst 10 readings:\n", 0, Analogue, false)
	if err != nil {
		return
	}
	return
}

func (ps *PsDesc) genOn() {
	var (
		offsetVoltage  int32                   = 0
		pkToPK         uint32                  = 1000000 // microvolts = 1V
		waveType       ps2000a.WaveTypeEnum    = ps2000a.Sine
		sweepType      ps2000a.SweepTypeEnum   = ps2000a.SweepUp
		operation      ps2000a.ExtraOperations = ps2000a.EsOff
		shots          uint32                  = 0
		sweeps         uint32                  = 0
		triggerType    ps2000a.SigGenTrigType
		triggerSource  ps2000a.SigGenTrigSource = ps2000a.SigGenNone
		extInThreshold int16
	)
	startFrequency := ps.GenFreq
	stopFrequency := startFrequency
	increment := ps.GenFreqStep
	dwellTime := ps.GenFreqDelay
	fmt.Println(offsetVoltage, pkToPK, waveType, startFrequency, stopFrequency, increment, dwellTime,
		sweepType, operation, shots, sweeps, triggerType, triggerSource, extInThreshold)
	err := ps.SetSigGenBuiltInV2(offsetVoltage, pkToPK, waveType, startFrequency, stopFrequency, increment, dwellTime,
		sweepType, operation, shots, sweeps, triggerType, triggerSource, extInThreshold)
	if err != nil {
		fmt.Println("genOn", err)
	}
}

func (ps *PsDesc) GetInfo() (err error) {
	ps.SignalGenerator = true
	ps.Ets = false
	ps.FirstRange = ps2000a.Range_20mv
	ps.LastRange = ps2000a.Range_20v
	ps.ChannelCount = ps2000a.DualScope
	ps.DigitalPorts = 0
	ps.AwgBufferSize = ps2000a.AwgMaxSigGenBufferSize
	ps.scaleVoltages = true
	ps.timeBase = 8
	ps.overSample = 1
	info, err := ps.GetUnitInfo(ps2000a.PicoVariantInfo)
	if err != nil {
		return
	}
	switch string(info[1]) {
	case "2":
		if len(info) == 4 || (len(info) == 5 && info[4] == 'A' || info == "2205MSO") {
			ps.FirstRange = ps2000a.Range_50mv
		}
		ps.MaxChannel = ps2000a.ChB
	case "4":
		ps.ChannelCount = ps2000a.QuadScope
		ps.MaxChannel = ps2000a.ChD
	default:
		fmt.Println(info)
		fmt.Println(byte(info[2]))
		err = fmt.Errorf("OpenUnit: unknown variant info %s", info)
		return
	}
	if strings.Contains(info, "MSO") {
		ps.DigitalPorts = 2
	}
	for ch := ps2000a.ChA; ch <= ps.MaxChannel; ch++ {
		ps.ChannelSettings[ch].Enabled = true
		ps.ChannelSettings[ch].CoupleType = ps2000a.Dc
		ps.ChannelSettings[ch].Range = ps2000a.Range_2v
	}
	ps.MaxValue, err = ps.MaximumValue()
	fmt.Println("........................................maxValue=", ps.MaxValue)
	if err != nil {
		return
	}
	return
}

func (ps *PsDesc) OpenDevice() (err error) {
	err = ps.OpenUnit("")
	if err != nil {
		fmt.Println("main open err=", err)
		return
	}
	err = ps.GetInfo()
	if err != nil {
		return
	}
	return
}

func terminalHandler() {
	b := []byte(" ")
	for {
		_, _ = os.Stdin.Read(b)
		// if err != nil {
		// 	fmt.Println("stdin read error ", err)
		// 	os.Exit(-1)
		// }
		b = bytes.ToUpper(b)
		terminal <- b[0]
	}
}

func (ps *PsDesc) Menu() (err error) {
	err = ps.OpenDevice()
	if err != nil {
		fmt.Println(err)
		return
	}
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		fmt.Println(err)
		return
	}
	defer term.Restore(int(os.Stdin.Fd()), oldState)
	terminal = make(chan byte)
	go terminalHandler()
	for {
		fmt.Print("\n")
		fmt.Print("B - Immediate block                           V - Set voltages\n \r")
		fmt.Print("T - Triggered block                           I - Set timebase\n \r")
		fmt.Print("E - Collect a block of data using ETS         A - ADC counts/mV\n \r")
		fmt.Print("R - Collect set of rapid captures             G - Signal generator\n \r")
		fmt.Print("S - Immediate streaming\n")
		fmt.Print("W - Triggered streaming\n")
		if ps.DigitalPorts > 0 {
			fmt.Print("D - Digital Ports menu\n \r")
		}
		fmt.Print("                                              X - Exit\n\n \r")
		fmt.Print("Operation:")
		b := <-terminal
		switch b {
		case 'A':
			ps.scaleVoltages = !ps.scaleVoltages
		case 'B':
			err = ps.CollectBlockImmediate()
		case 'G':
			ps.GenFreq = 1000000.0
			ps.genOn()
		case 'X':
			return
		default:
			fmt.Printf("invalid operation: %s\n\r", string(b))
		}
		if err != nil {
			return
		}
	}
}
