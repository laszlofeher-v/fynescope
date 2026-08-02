//go:build sim

package ps2000a

import (
	"fynescope/demo"

	"math"
	"math/rand"
	"time"
)

type channelDesc struct {
	enabled  bool
	vrange   int
	offset   float64
	coupling int
}

var (
	channels         [4]channelDesc
	triggerDetector  *demo.TriggerDetector
	buffers          [4][]int16
	running          bool
	isReady          bool
	timeBaseSet      uint32
	nOfPreTrSamples  int32
	nOfPostTrSamples int32

	// Single channel generator
	genOn            bool
	genPkToPk        uint32
	genOffsetVoltage int32
	genWaveFunction  demo.WaveformGenerator
	sweepController  *demo.SweepController

	etsEnabled              bool
	etsTimeBuffer           []int64
	timeIntervalPicoSeconds float64
)

const maxValue = 32767

func simOpenUnit(handle int16) {
	triggerDetector = demo.NewTriggerDetector(false, 0, 0, 0, 0)
}

func simSetChannel(handle int16, channel int, enabled bool, dc int, rangeEnum int, analogOffset float32) {
	if channel >= 0 && channel < 4 {
		channels[channel].enabled = enabled
		channels[channel].coupling = dc
		channels[channel].vrange = rangeEnum
		channels[channel].offset = float64(analogOffset)
	}
}

func simSetSimpleTrigger(handle int16, enable bool, source int, threshold int16, direction int, delay uint32, autoTriggerMs int16) {
	triggerDetector = demo.NewTriggerDetector(enable, threshold, 0, demo.ThresholdDirection(direction), demo.ChannelId(source))
}

func simSetTriggerChannelProperties(handle int16, props []demo.TriggerChannelProperties, auxOutputEnable bool, autoTriggerMs int32) {
	if triggerDetector == nil {
		triggerDetector = demo.NewTriggerDetector(true, 0, 0, demo.TriggerNone, demo.ChA)
	}
	triggerDetector.SetChannelProperties(props)
}

func simSetTriggerChannelConditions(handle int16, conds []demo.TriggerConditions) {
	if triggerDetector == nil {
		triggerDetector = demo.NewTriggerDetector(true, 0, 0, demo.TriggerNone, demo.ChA)
	}
	triggerDetector.SetChannelConditions(conds)
}

func simSetTriggerChannelDirections(handle int16, dirA, dirB, dirC, dirD demo.ThresholdDirection) {
	if triggerDetector == nil {
		triggerDetector = demo.NewTriggerDetector(true, 0, 0, demo.TriggerNone, demo.ChA)
	}
	triggerDetector.SetChannelDirections(dirA, dirB, dirC, dirD)
}

func simSetPulseWidthQualifier(handle int16, conds []demo.PwqConditions, direction demo.ThresholdDirection, lower, upper uint32, type_ demo.PulseWidthType) {
	if triggerDetector == nil {
		triggerDetector = demo.NewTriggerDetector(true, 0, 0, demo.TriggerNone, demo.ChA)
	}
	triggerDetector.SetPulseWidthQualifier(conds, direction, lower, upper, type_)
}

func simSetDataBuffer(handle int16, channel int, buffer []int16, segmentIndex uint32) {
	if channel >= 0 && channel < 4 {
		buffers[channel] = buffer
	}
}

func calculateSampleLevelAtTime(t float64, ch int) float64 {
	chDesc := &channels[ch]

	freq := float64(0)
	if sweepController != nil {
		freq = sweepController.GetCurrentFrequency()
	}

	phase := (t * freq) * math.Pi * 2

	if pnd := demo.GetPhaseNoiseDegree(ch); pnd > 0 {
		phase += (rand.Float64()*2 - 1) * pnd * math.Pi / 180.0
	}

	signal := float64(0)
	if genWaveFunction != nil && genOn {
		signal = genWaveFunction(phase, freq)
	}

	rangeMv := float64(10) // Mock ranges: map vrange to mV. E.g. PS2000A_10V = 10000. Just mock 10000.
	switch chDesc.vrange {
	case 1:
		rangeMv = 20
	case 2:
		rangeMv = 50
	case 3:
		rangeMv = 100
	case 4:
		rangeMv = 200
	case 5:
		rangeMv = 500
	case 6:
		rangeMv = 1000
	case 7:
		rangeMv = 2000
	case 8:
		rangeMv = 5000
	case 9:
		rangeMv = 10000
	case 10:
		rangeMv = 20000
	}

	a := float64(0)
	genOffset := float64(0)

	if genOn {
		a = float64(genPkToPk/2000) / rangeMv
		genOffset = float64(genOffsetVoltage/1000) / rangeMv
	}

	chOffset := (chDesc.offset * 1000.0) / rangeMv

	noise_offset := float64(0)
	if na := demo.GetNoiseAmplitude(ch); na > 0 && genOn {
		rnd := rand.Intn(100)
		if rnd < 2 {
			noise_offset = (rand.Float64()*2 - 1) * na / rangeMv
		}
	}

	levelFloat := (signal*a + genOffset + chOffset + noise_offset) * float64(maxValue)
	return levelFloat
}

func simRunBlock(handle int16, pre int32, post int32, timebase uint32, readyCallback func(handle int16, status int32)) {
	running = true
	isReady = false
	nOfPreTrSamples = pre
	nOfPostTrSamples = post
	timeBaseSet = timebase
	demo.AdvancePRBS()

	go func() {
		// Fill buffers
		var dt float64
		switch {
		case etsEnabled:
			dt = timeIntervalPicoSeconds / 1e12
		case timebase == 0:
			dt = 1e-9
		case timebase == 1:
			dt = 2e-9
		case timebase == 2:
			dt = 4e-9
		default:
			dt = (1000.0 * (float64(timebase) - 2.0) / 125.0) * 1e-9
		}

		signalFunc := func(t float64, ch demo.ChannelId) float64 {
			return calculateSampleLevelAtTime(t, int(ch))
		}

		reqSamples := uint32(pre + post)
		triggerTime := triggerDetector.FindTriggerPoint(signalFunc, reqSamples, 1.0, dt)

		for ch := 0; ch < 4; ch++ {
			buf := buffers[ch]
			if channels[ch].enabled && buf != nil {
				length := int(pre + post)
				if length > len(buf) {
					length = len(buf)
				}
				for i := 0; i < length; i++ {
					rt := (float64(i)-float64(pre))*dt + triggerTime

					if etsEnabled && ch == 0 {
						t0Fs := 1e15 * float64(pre) * dt
						rteFs := float64(i) * dt * 1e15
						if i < len(etsTimeBuffer) {
							etsTimeBuffer[i] = int64(rteFs - t0Fs)
						}
					}

					val := calculateSampleLevelAtTime(rt, ch)

					var level int16
					if val > maxValue {
						level = maxValue
					} else if val < -maxValue {
						level = -maxValue
					} else {
						level = int16(math.Round(val))
					}
					buf[i] = level
				}
			}
			if sweepController != nil {
				sweepController.Update()
			}
		}

		time.Sleep(10 * time.Millisecond)
		isReady = true
		if readyCallback != nil {
			readyCallback(handle, 0)
		}
	}()
}

func simStop(handle int16) {
	running = false
	isReady = false
}

func simSetSigGenBuiltIn(handle int16, offsetVoltage int32, pkToPk uint32, waveType int, startFreq float64, stopFreq float64, increment float64, dwellTime float64, sweepType int, operation int) {
	genOn = true
	genOffsetVoltage = offsetVoltage
	genPkToPk = pkToPk
	if operation == int(Prbs) {
		genWaveFunction = demo.NewPrbsGenerator()
	} else {
		genWaveFunction = demo.NewWaveformGenerator(demo.WaveTypeEnum(waveType))
	}
	dwellDuration := time.Duration(dwellTime*1000000000) * time.Nanosecond
	sweepController = demo.NewSweepController(startFreq, stopFreq, increment, demo.SweepTypeEnum(sweepType), dwellDuration)
}

func simIsReady(handle int16) bool {
	return isReady
}

func simGetValues(handle int16, startIndex uint32, noOfSamples uint32) (uint32, int16) {
	return noOfSamples, 0
}

func simSetEts(handle int16, mode int, etsCycles int16, etsInterLeave int16, sampleTimePicoseconds *int32) {
	etsEnabled = mode != 0
	if etsCycles < etsInterLeave {
		return
	}
	if etsCycles > etsInterLeave*10+9 {
		return
	}
	timeIntervalPicoSeconds = 2000.0 / float64(etsInterLeave)
	if sampleTimePicoseconds != nil {
		*sampleTimePicoseconds = int32(timeIntervalPicoSeconds)
	}
}

func simSetEtsTimeBuffer(handle int16, buffer []int64) {
	etsTimeBuffer = buffer
}

func simSetEtsTimeBuffers(handle int16, timeUpper, timeLower []uint32) {
}
