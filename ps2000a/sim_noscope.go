//go:build sim

package ps2000a

import (
	"fmt"
	"fynescope/demo"
	"fynescope/genericps"
	"log/slog"

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
	buffersMin       [4][]int16
	digitalBuffers   [256][]int16
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

	// Set by SetDataBuffer/SetDataBuffers; used by simRunBlock for aggregation
	downSampleMode int32 // 0=None, 1=Aggregate(ED), 2=Decimate, 4=Average(High-Res)

	streamingRunning      bool
	streamingIntervalNs   float64
	streamingStartTime    time.Time
	streamingLastReadTime time.Time
	streamingWriteIndex   int32
	totalSamplesGenerated int64
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

func simGetAnalogueOffset(handle int16, voltageRange int, coupling Coupling) (maximumVoltage, minimumVoltage float32, err error) {
	maximumVoltage, minimumVoltage = 20, -20
	if handle <= 0 {
		err = fmt.Errorf("invalid handle")
	}
	return
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
	} else {
		digitalBuffers[channel] = buffer
	}
}

func simSetDataBufferWithMode(handle int16, channel int, buffer []int16, segmentIndex uint32, mode int32) {
	simSetDataBuffer(handle, channel, buffer, segmentIndex)
	downSampleMode = mode
}

func simSetDataBufferMinWithMode(handle int16, channel int, bufMin []int16, segmentIndex uint32, mode int32) {
	if channel >= 0 && channel < 4 {
		buffersMin[channel] = bufMin
	}
	downSampleMode = mode
}

func simSetDigitalPort(handle int16, port int, enabled bool, logicLevel int16) {
	// No-op for sim_noscope; simRunBlock relies on digitalBuffers being non-nil.
}

func simSetTriggerDigitalPortProperties(handle int16, dirs []demo.DigitalChannelDirections) {
	if triggerDetector == nil {
		triggerDetector = demo.NewTriggerDetector(true, 0, 0, demo.TriggerNone, demo.ChA)
	}
	triggerDetector.SetDigitalPortProperties(dirs)
}

func simSetDigitalAnalogTriggerOperand(handle int16, operand demo.TriggerOperand) {
	if triggerDetector == nil {
		triggerDetector = demo.NewTriggerDetector(true, 0, 0, demo.TriggerNone, demo.ChA)
	}
	triggerDetector.SetDigitalAnalogTriggerOperand(operand)
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
		a = float64(genPkToPk) / 2000.0 / rangeMv
		genOffset = float64(genOffsetVoltage) / 1000.0 / rangeMv
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

// simTimebaseToNs converts a PS2000A timebase index to the corresponding
// sample interval in nanoseconds, matching the real hardware formula and the
// dt computation used inside simRunBlock.
//
//	timebase 0  → 1 ns   (500 MS/s, 1 channel)
//	timebase 1  → 2 ns   (250 MS/s)
//	timebase 2  → 4 ns   (125 MS/s)
//	timebase n≥3 → 1000*(n-2)/125 ns  (increments of 8 ns)
func simTimebaseToNs(timebase uint32) float64 {
	switch {
	case timebase == 0:
		return 1
	case timebase == 1:
		return 2
	case timebase == 2:
		return 4
	default:
		return 1000.0 * (float64(timebase) - 2.0) / 125.0
	}
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
		default:
			dt = simTimebaseToNs(timebase) * 1e-9
		}

		signalFunc := func(t float64, ch demo.ChannelId) float64 {
			return calculateSampleLevelAtTime(t, int(ch))
		}

		reqSamples := uint32(pre + post)
		found, triggerTime := triggerDetector.FindTriggerPoint(signalFunc, reqSamples, 1.0, dt)
		if found {
			for ch := 0; ch < 4; ch++ {
				buf := buffers[ch]
				if !channels[ch].enabled || buf == nil {
					continue
				}

				bufLen := len(buf)
				if bufLen == 0 {
					continue
				}

				// Calculate the downSampleRatio from raw vs output sample counts
				// pre+post are raw samples; buf is sized for output samples
				rawTotal := int(pre + post)
				downSampleRatio := rawTotal / bufLen
				if downSampleRatio < 1 {
					downSampleRatio = 1
				}

				for i := 0; i < bufLen; i++ {
					// Output sample i corresponds to raw samples [i*ratio .. (i+1)*ratio-1]
					// relative to trigger, pre-samples are to the left
					var aggSum float64
					var aggMin float64 = math.MaxFloat64
					var aggMax float64 = -math.MaxFloat64
					var decimateVal float64

					for r := 0; r < downSampleRatio; r++ {
						rawIdx := i*downSampleRatio + r
						rt := (float64(rawIdx)-float64(pre))*dt + triggerTime

						if etsEnabled && ch == 0 {
							t0Fs := 1e15 * float64(pre) * dt
							rteFs := float64(rawIdx) * dt * 1e15
							if i < len(etsTimeBuffer) {
								etsTimeBuffer[i] = int64(rteFs - t0Fs)
							}
						}

						val := calculateSampleLevelAtTime(rt, ch)
						aggSum += val
						if val < aggMin {
							aggMin = val
						}
						if val > aggMax {
							aggMax = val
						}
						if r == 0 {
							decimateVal = val
						}
					}

					// Apply the correct aggregation based on the downSampleMode stored at SetDataBuffer time:
					//   0 = None (ratio will be 1, avgerage = single sample)
					//   1 = Aggregate / ED: use min AND max; we write max to bufMax here
					//   2 = Decimate: take first raw sample
					//   4 = Average / High-Res: arithmetic mean
					var finalVal float64
					switch downSampleMode {
					case 2: // Decimate
						finalVal = decimateVal
					case 1: // Aggregate / ED – write aggMax to the max buffer (bufMax); aggMin goes to bufMin
						finalVal = aggMax // simSetDataBuffers writes aggMin to bufMin separately if needed
					default: // 0 (None) or 4 (Average/High-Res)
						finalVal = aggSum / float64(downSampleRatio)
					}

					var level int16
					if finalVal > maxValue {
						level = maxValue
					} else if finalVal < -maxValue {
						level = -maxValue
					} else {
						level = int16(math.Round(finalVal))
					}
					buf[i] = level

					// For Aggregate/ED mode, also fill the min buffer with aggMin
					if downSampleMode == 1 && buffersMin[ch] != nil && i < len(buffersMin[ch]) {
						minVal := aggMin
						var levelMin int16
						if minVal > maxValue {
							levelMin = maxValue
						} else if minVal < -maxValue {
							levelMin = -maxValue
						} else {
							levelMin = int16(math.Round(minVal))
						}
						buffersMin[ch][i] = levelMin
					}
				}
			}

			// Digital buffers
			p0Buf := digitalBuffers[128]
			p1Buf := digitalBuffers[129]
			if p0Buf != nil || p1Buf != nil {
				// Determine output length from whichever digital buffer exists
				length := 0
				if p0Buf != nil {
					length = len(p0Buf)
				}
				if p1Buf != nil && (length == 0 || len(p1Buf) < length) {
					length = len(p1Buf)
				}

				// Infer downSampleRatio from raw vs output counts
				rawTotal := int(pre + post)
				downSampleRatio := rawTotal / length
				if downSampleRatio < 1 {
					downSampleRatio = 1
				}

				for i := 0; i < length; i++ {
					rawIdx := i * downSampleRatio
					rt := (float64(rawIdx)-float64(pre))*dt + triggerTime
					p0Val, p1Val, p0En, p1En := demo.GetDemoDigitalGenValue(rt)
					if p0Buf != nil && p0En {
						p0Buf[i] = p0Val
					}
					if p1Buf != nil && p1En {
						p1Buf[i] = p1Val
					}
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
	streamingRunning = false
}

func simRunStreaming(handle int16, reqSampleInterval uint32, sampleIntervalTimeUnits TimeUnits,
	maxPreTriggerSamples, maxPostTriggerSamples uint32,
	autoStop bool, downSampleRatio uint32, downSampleRatioMode RatioMode,
	overviewBufferSize uint32) (sampleInterval uint32, err error) {
	if handle <= 0 {
		return 0, fmt.Errorf("invalid handle")
	}
	streamingIntervalNs = float64(reqSampleInterval) * demo.TimeUnitToVal(demo.TimeUnits(sampleIntervalTimeUnits)) * 1e9
	if streamingIntervalNs <= 0 {
		streamingIntervalNs = 1.0
	}

	streamingRunning = true
	running = true
	streamingStartTime = time.Now()
	streamingLastReadTime = time.Now()
	streamingWriteIndex = 0
	totalSamplesGenerated = 0

	return reqSampleInterval, nil
}

func simGetStreamingLatestValues(handle int16, lpStreamingReadyGoPar func(handle int16, noOfSamples int32, startIndex uint32, overflow int16, triggerAt uint32, triggered, autoStop int16, param any), param any) (err error) {
	if !streamingRunning {
		lpStreamingReadyGoPar(handle, 0, 0, 0, 0, 0, 1, param)
		return nil
	}

	elapsed := time.Since(streamingLastReadTime).Seconds()
	numNew := int32(elapsed / (streamingIntervalNs * 1e-9))
	if numNew <= 0 {
		lpStreamingReadyGoPar(handle, 0, 0, 0, 0, 0, 0, param)
		return nil
	}

	var activeBufLen int32 = 0
	for ch := 0; ch < 4; ch++ {
		if channels[ch].enabled && len(buffers[ch]) > 0 {
			activeBufLen = int32(len(buffers[ch]))
			break
		}
	}

	if activeBufLen <= 0 {
		lpStreamingReadyGoPar(handle, 0, 0, 0, 0, 0, 0, param)
		return nil
	}

	writeCount := numNew
	if streamingWriteIndex+writeCount > activeBufLen {
		writeCount = activeBufLen - streamingWriteIndex
	}

	if writeCount > 0 {
		for ch := 0; ch < 4; ch++ {
			if channels[ch].enabled && len(buffers[ch]) > 0 {
				for i := 0; i < int(writeCount); i++ {
					t := float64(totalSamplesGenerated+int64(i)) * (streamingIntervalNs * 1e-9)
					val := calculateSampleLevelAtTime(t, ch)

					var level int16
					if val > float64(maxValue) {
						level = maxValue
					} else if val < -float64(maxValue) {
						level = -maxValue
					} else {
						level = int16(math.Round(val))
					}
					idx := (streamingWriteIndex + int32(i)) % activeBufLen
					if idx < int32(len(buffers[ch])) {
						buffers[ch][idx] = level
					}
				}
			}
		}

		totalSamplesGenerated += int64(writeCount)
		if sweepController != nil {
			sweepController.Update()
		}
		lpStreamingReadyGoPar(handle, writeCount, uint32(streamingWriteIndex), 0, 0, 0, 0, param)
		streamingWriteIndex = (streamingWriteIndex + writeCount) % activeBufLen
		streamingLastReadTime = streamingLastReadTime.Add(time.Duration(float64(writeCount)*streamingIntervalNs) * time.Nanosecond)
	} else {
		lpStreamingReadyGoPar(handle, 0, 0, 0, 0, 0, 0, param)
	}

	return nil
}

func simNoOfStreamingValues(handle int16) (noOfValues uint64, err error) {
	if handle <= 0 {
		return 0, fmt.Errorf("invalid handle")
	}
	return uint64(totalSamplesGenerated), nil
}

func simSetSigGenBuiltIn(handle int16, offsetVoltage int32, pkToPk uint32, waveType int, startFreq float64, stopFreq float64, increment float64, dwellTime float64, sweepType int, operation int) {
	genOn = true
	genOffsetVoltage = offsetVoltage
	genPkToPk = pkToPk
	if operation == int(genericps.Prbs) {
		genWaveFunction = demo.NewPrbsGenerator()
	} else if operation == int(genericps.WhiteNoise) {
		genWaveFunction = demo.NewWhiteNoiseGenerator()
	} else {
		genWaveFunction = demo.NewWaveformGenerator(demo.WaveTypeEnum(waveType))
	}
	dwellDuration := time.Duration(dwellTime*1000000000) * time.Nanosecond
	slog.Debug("simSetSigGenBuiltIn", "startFreq", startFreq, "stopFreq", stopFreq)
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

