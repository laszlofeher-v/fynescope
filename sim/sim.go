// Package sim provides a PicoScope simulator for testing and development.
// It simulates the behavior of a PicoScope oscilloscope including signal generation,
// triggering, and data acquisition without requiring physical hardware.
package sim

import (
	"fmt"
	"log/slog"
	"math"

	"math/rand"
	"fynescope/genericps"
	"time"
)

type (
	returnStatus int
	channelDesc  struct {
		enabled  bool
		vrange   RangeEnum
		offset   float64
		coupling Coupling
		// Generator settings for this channel
		genOn            bool
		genPkToPk        uint32
		genOffsetVoltage int32
		genWaveFunction  WaveformGenerator
		sweepController  *SweepController
		phase            float64
		// RLC Filter settings
		genSource  ChannelId
		rlcEnabled bool
		rlcType    string
		rlcR       float64
		rlcRUnit   string
		rlcL       float64
		rlcLUnit   string
		rlcC       float64
		rlcCUnit   string

	}
)

// Singleton simulator
var (
	// handle                  int16
	channels                [MaxChannels]channelDesc
	behaviour               returnStatus
	timeBaseSet             uint32
	timeIntervalPicoSeconds float64
	nOfPreTrSamples         int32
	nOfPostTrSamples        int32
	triggerDetector         *TriggerDetector
	// triggerDelay            uint32
	TtToPercent float64
	// NoiseAmplitude, PhaseNoiseDegree, TriggerTimeOffset are now accessed via
	// the atomic getter/setter functions in params.go to avoid data races.
	// Keep these exported names as thin aliases so existing external callers compile.
	NoiseAmplitude       float64
	PhaseNoiseDegree     float64
	complexTrigger       bool
	autoTrigger          int32
	simChannelProperties [MaxChannels]TriggerChannelProperties
	simTriggerConditions [MaxChannels]TriggerConditions
	channelAThresholdDirection, channelBThresholdDirection,
	channelCThresholdDirection, channelDThresholdDirection ThresholdDirection
	etsEnbaled             bool
	etsTimeBuffer          []int64
	running                bool
	TriggerCalculationMode int
	streamingRunning       bool
	streamingIntervalNs    float64
	streamingStartTime     time.Time
	streamingLastReadTime  time.Time
	streamingWriteIndex    int32
	totalSamplesGenerated  int64
)

type (
	SimDesc struct {
		handle int16
	}
)

func boolToint16(b bool) int16 {
	if b {
		return int16(1)
	}
	return int16(0)
}

func nextInt16() func() int16 {
	i := int16(0)
	return func() int16 {
		i++
		return i
	}
}

var uniqueHandle = nextInt16()

func EnumerateUnits(bufferLen int16) (count int16, serials string, serialLth int16, err error) {
	nf := func() {
		count = 1
		serials = scopeBathAndSerialInfo
		serialLth = int16(len(serials))
		err = nil
	}
	ff := func() {
		err = fmt.Errorf("EnumerateUnits error")
	}
	switch behaviour {
	case normal:
		nf()
	case faulty:
		ff()
	case timeoutNormal:
		time.Sleep(timeout)
		nf()
	case timeoutfaulty:
		time.Sleep(timeout)
		ff()
	}
	return
}
func openUnit(serial string) (handle int16, err error) {
	slog.Debug("openUnit called")
	nf := func() {
		handle = uniqueHandle()
		triggerDetector = NewTriggerDetector(false, 0, 0, 0, 0)
		triggerDetector.SetTriggerCalculationMode(TriggerCalculationMode)
		err = nil
	}
	ff := func() {
		err = fmt.Errorf("OpenUnit error")
	}
	switch behaviour {
	case normal:
		nf()
	case faulty:
		ff()
	case timeoutNormal:
		time.Sleep(timeout)
		nf()
	case timeoutfaulty:
		time.Sleep(timeout)
		ff()
	}
	loadConstants()
	return
}
func openUnitAsync(serial string) (status int16, err error) {
	nf := func() {
		status = 0
		triggerDetector = NewTriggerDetector(false, 0, 0, 0, 0)
		triggerDetector.SetTriggerCalculationMode(TriggerCalculationMode)
		err = nil
	}
	ff := func() {
		err = fmt.Errorf("OpenUnit error")
	}
	switch behaviour {
	case normal:
		nf()
	case faulty:
		ff()
	case timeoutNormal:
		time.Sleep(timeout)
		nf()
	case timeoutfaulty:
		time.Sleep(timeout)
		ff()
	}
	return
}
func openUnitProgress() (handle int16, progressPercent, complete int16, err error) {
	nf := func() {
		handle = uniqueHandle()
		progressPercent = 100
		complete = 1
		triggerDetector = NewTriggerDetector(false, 0, 0, 0, 0)
		triggerDetector.SetTriggerCalculationMode(TriggerCalculationMode)
		err = nil
	}
	ff := func() {
		err = fmt.Errorf("OpenUnit error")
	}
	switch behaviour {
	case normal:
		nf()
	case faulty:
		ff()
	case timeoutNormal:
		time.Sleep(timeout)
		nf()
	case timeoutfaulty:
		time.Sleep(timeout)
		ff()
	}
	return
}

func simCloseUnit(handle int16) (err error) {
	return
}

func simGetUnitInfo(handle int16, info PicoInfo) (infoString string, err error) {
	switch info {
	case PicoVariantInfo:
		infoString = scopeVariantInfo
	case PicoBatchAndSerial:
		infoString = scopeBathAndSerialInfo
	default:
		infoString = "?"
	}
	return
}

func simFlashLed(handle int16, start int16) (err error) {
	err = fmt.Errorf(notImplemented)
	return
}

func lpDataReadyGo(handle int16, status int, noOfSamples uint32, overflow int16, param any) {
}

func simGetValuesAsync(handle int16, startIndex, noOfSamples, downSampleRatio uint32,
	downSampleRatioMode RatioMode, lpDataReadyGoPar DataReady, segmentIndex uint32,
	param any) (err error) {
	return
}

// simGetValues retrieves simulated sample data for the specified channels.
// It generates waveform data, applies triggering, and handles frequency sweeps.

// calculateSampleLevelAtTime calculates the signal level for a specific channel at a given time.
func calculateSampleLevelAtTime(t float64, ch ChannelId) float64 {
	chDesc := &channels[ch]

	genSrc := chDesc.genSource
	if genSrc < 0 || int(genSrc) >= MaxChannels {
		genSrc = ch
	}
	genChDesc := &channels[genSrc]

	// Get current frequency
	freq := float64(0)
	if genChDesc.sweepController != nil {
		freq = genChDesc.sweepController.GetCurrentFrequency()
	}

	// Generate base waveform
	// t is time in seconds, freq is Hz
	phase := (t*freq + genChDesc.phase/360.0) * math.Pi * 2

	// Add phase noise (read atomically — written by GUI goroutine)
	if pnd := GetPhaseNoiseDegree(int(genSrc)); pnd > 0 {
		phase += (rand.Float64()*2 - 1) * pnd * math.Pi / 180.0
	}

	// Get base signal from generator
	signal := float64(0)
	if genChDesc.genWaveFunction != nil && genChDesc.genOn {
		signal = genChDesc.genWaveFunction(phase, freq)
	}

	// Signal generated from waveform function (typically in range [-1, 1])
	rangeMv := float64(InputRanges(chDesc.vrange))

	// Default to no generator output
	a := float64(0)
	genOffset := float64(0)

	if genChDesc.genOn {
		// Amplitude scaling
		// Note: integer division for genPkToPk/2000 is intentional to match original behavior
		a = float64(genChDesc.genPkToPk/2000) / rangeMv
		// Offset calculation
		// Note: integer division for genOffsetVoltage/1000 is intentional to match original behavior
		genOffset = float64(genChDesc.genOffsetVoltage/1000) / rangeMv
	}

	chOffset := (chDesc.offset * 1000.0) / rangeMv

	noise_offset := float64(0)
	if na := GetNoiseAmplitude(int(genSrc)); na > 0 && genChDesc.genOn {
		rnd := rand.Intn(100)
		if rnd < 2 {
			noise_offset = (rand.Float64()*2 - 1) * na / rangeMv
		}
	}

	levelFloat := (signal*a + genOffset + chOffset + noise_offset) * float64(maxValue)
	return levelFloat
}

// // calculateSampleLevel calculates signal level from sample index.
// func calculateSampleLevel(sampleIndex float64, timeIntervalNs float64, ch ChannelId) float64 {
// 	// Convert sample index to time in seconds
// 	t := sampleIndex * timeIntervalNs / 1e9
// 	return calculateSampleLevelAtTime(t, ch)
// }

func simGetValues(handle int16, startIndex, reqNoOfSamples, downSampleRatio uint32,
	downSampleRatioMode RatioMode, segmentIndex uint32) (noOfSamples uint32, overflow int16, err error) {
	var timeIntervalNanoseconds float64

	// Calculate time interval based on ETS mode or normal timebase
	nec := numberOfEnabledChannels()
	if etsEnbaled {
		timeIntervalNanoseconds = timeIntervalPicoSeconds / 1e3
	} else {
		timeIntervalNanoseconds, err = simGetTimeInterval2(timeBaseSet, nec)
		if err != nil {
			return
		}
	}

	// Find trigger point using trigger detector

	var triggerTime float64
	freq := float64(0)
	sourceCh := triggerDetector.GetSource()
	if sourceCh >= 0 && int(sourceCh) < MaxChannels {
		genSrc := channels[sourceCh].genSource
		if genSrc < 0 || int(genSrc) >= MaxChannels {
			genSrc = sourceCh
		}
		if channels[genSrc].sweepController != nil {
			freq = channels[genSrc].sweepController.GetCurrentFrequency()
		}
	}

	minIter := int(reqNoOfSamples)
	if minIter < 100 {
		minIter = 100
	}
	maxIter := minIter

	if freq > 0 && timeIntervalNanoseconds > 0 {
		samplesPerPeriod := 1.0 / (freq * timeIntervalNanoseconds * 1e-9)
		calcIter := int(samplesPerPeriod * 2.5)
		if calcIter > maxIter {
			maxIter = calcIter
		}
	}
	if maxIter > maxTriggerTest {
		maxIter = maxTriggerTest
	}
	triggerDetector.SetMaxIterations(maxIter)

	dt := timeIntervalNanoseconds / 1e9

	triggerFilters := make([]*RlcFilter, MaxChannels)
	triggerAcFilters := make([]*RlcFilter, MaxChannels)

	lastT := make([]float64, MaxChannels)
	prevT := make([]float64, MaxChannels)
	lastVal := make([]float64, MaxChannels)
	prevVal := make([]float64, MaxChannels)
	for i := 0; i < MaxChannels; i++ {
		lastT[i] = -1
		prevT[i] = -1
	}

	for ch := 0; ch < MaxChannels; ch++ {
		if !channels[ch].enabled {
			continue
		}

		if channels[ch].rlcEnabled {
			triggerFilters[ch] = NewRlcFilter(channels[ch].rlcType, channels[ch].rlcR, channels[ch].rlcRUnit,
				channels[ch].rlcL, channels[ch].rlcLUnit, channels[ch].rlcC, channels[ch].rlcCUnit, dt)
		}

		// AC-coupling: analogue 1 Hz highpass filter (models the input capacitor)
		if channels[ch].coupling == Ac {
			triggerAcFilters[ch] = NewAcCouplingFilter(dt)
		}

		// Preroll all trigger filters to reach steady state before t=0
		prerollSamples := 1000
		prerollStart := -float64(prerollSamples) * dt
		firstVal := calculateSampleLevelAtTime(prerollStart, ChannelId(ch))
		if triggerFilters[ch] != nil {
			firstVal = triggerFilters[ch].Step(firstVal)
		}
		if triggerAcFilters[ch] != nil {
			firstVal = triggerAcFilters[ch].Step(firstVal)
		}

		for i := 1; i < prerollSamples; i++ {
			pt := prerollStart + float64(i)*dt
			raw := calculateSampleLevelAtTime(pt, ChannelId(ch))
			if triggerFilters[ch] != nil {
				raw = triggerFilters[ch].Step(raw)
			}
			if triggerAcFilters[ch] != nil {
				raw = triggerAcFilters[ch].Step(raw)
			}

		}
	}

	// Create signal function for trigger source channel
	signalFunc := func(t float64, ch ChannelId) float64 {
		i := int(ch)
		if i < 0 || i >= MaxChannels {
			return 0
		}
		if t == lastT[i] {
			return lastVal[i]
		}
		if t == prevT[i] {
			return prevVal[i]
		}
		if t > prevT[i] && t < lastT[i] && prevT[i] != -1 {
			// Interpolate for FineGrainedTrigger
			fraction := (t - prevT[i]) / (lastT[i] - prevT[i])
			return prevVal[i] + fraction*(lastVal[i] - prevVal[i])
		}

		raw := calculateSampleLevelAtTime(t, ch)
		val := raw
		if triggerFilters[i] != nil {
			val = triggerFilters[i].Step(val)
		}
		if triggerAcFilters[i] != nil {
			val = triggerAcFilters[i].Step(val)
		}


		prevT[i] = lastT[i]
		prevVal[i] = lastVal[i]
		lastT[i] = t
		lastVal[i] = val
		return val
	}

	maxTime := float64(maxIter) * dt
	triggerTime = triggerDetector.FindTriggerPoint(signalFunc, reqNoOfSamples, maxTime, dt)

	// Initialize and pre-roll filters for enabled channels
	filters := make([]*RlcFilter, MaxChannels)
	acFilters := make([]*RlcFilter, MaxChannels)

	prerollSamples := 1000
	for ch := range buffers {
		if buffers[ch] != nil && channels[ch].enabled {
			prerollStart := triggerTime - float64(nOfPreTrSamples)*dt - float64(prerollSamples)*dt

			if channels[ch].rlcEnabled {
				// Initialize RLC filter with current dt and pre-roll
				filters[ch] = NewRlcFilter(channels[ch].rlcType, channels[ch].rlcR, channels[ch].rlcRUnit,
					channels[ch].rlcL, channels[ch].rlcLUnit, channels[ch].rlcC, channels[ch].rlcCUnit, dt)
			}

			// AC-coupling: analogue 1 Hz highpass filter (models the input capacitor)
			if channels[ch].coupling == Ac {
				acFilters[ch] = NewAcCouplingFilter(dt)
			}

			// Pre-roll RLC and AC coupling together in a single pass
			if filters[ch] != nil || acFilters[ch] != nil {
				for i := 0; i < prerollSamples; i++ {
					rt := prerollStart + float64(i)*dt
					raw := calculateSampleLevelAtTime(rt, ChannelId(ch))
					if filters[ch] != nil {
						raw = filters[ch].Step(raw)
					}
					if acFilters[ch] != nil {
						acFilters[ch].Step(raw)
					}
				}
			}

		}
	}

	// Fill sample buffers for all enabled channels
	for ch := range buffers {
		if buffers[ch] != nil && channels[ch].enabled {
			if noOfSamples < uint32(len(buffers[ch])) {
				noOfSamples = uint32(len(buffers[ch]))
			}
			for t := range buffers[ch] {
				// Calculate real time for this sample
				dt := timeIntervalNanoseconds / 1e9
				rt := (float64(t)-float64(nOfPreTrSamples))*dt + triggerTime
				// Store ETS time if enabled (fs)
				if etsEnbaled && ch == int(ChA) {
					t0Fs := 1e15 * float64(nOfPreTrSamples) * timeIntervalNanoseconds / 1e9
					rteFs := (float64(t) * timeIntervalNanoseconds) / 1e9 * 1e15
					etsTimeBuffer[t] = int64(rteFs - t0Fs)
					// slog.Debug("sim", "b", etsTimeBuffer[t], "rteFs", rteFs, "t0Fs", t0Fs)
				}

				// Calculate sample level for this channel
				levelFloat := calculateSampleLevelAtTime(rt, ChannelId(ch))

				// Apply RLC Filter if enabled
				if filters[ch] != nil {
					levelFloat = filters[ch].Step(levelFloat)
				}

				// Apply AC coupling filter (analogue 1 Hz highpass) if channel is AC-coupled
				if acFilters[ch] != nil {
					levelFloat = acFilters[ch].Step(levelFloat)
				}



				// Clamp to valid range and detect overflow
				var level int16
				if levelFloat > maxValue {
					overflow |= int16(1 << ch)
					level = maxValue
				} else if levelFloat < -maxValue {
					overflow |= int16(1 << ch)
					level = -maxValue
				} else {
					level = int16(math.Round(levelFloat))
				}
				buffers[ch][t] = level
			}
		}

		// Update sweep controller for each channel after each buffer fill
		for i := range channels {
			if channels[i].sweepController != nil {
				channels[i].sweepController.Update()
			}
		}
	}
	if etsEnbaled && running {
		go delayedCall(handle, regLpBlockReadyGo)
	}
	return
}

func simGetValuesBulk(handle int16, reqNoOfSamples uint32, fromSegmentIndex, toSegmentIndex, downSampleRatio uint32,
	downSampleRatioMode RatioMode, overflow []int16) (noSamples uint32, err error) {
	err = fmt.Errorf(notImplemented)
	return
}

func simGetValuesOverlapped(handle int16, startIndex, reqNoOfSamples, downSampleRatio uint32,
	downSampleRatioMode RatioMode, segmentIndex uint32, overflow []int16) (noSamples uint32, err error) {
	err = fmt.Errorf(notImplemented)
	return
}

func simGetValuesOverlappedBulk(handle int16, startIndex, reqNoOfSamples, downSampleRatio uint32,
	downSampleRatioMode RatioMode, fromSegmentIndex, toSegmentIndex uint32, overflow []int16) (noSamples uint32, err error) {
	err = fmt.Errorf(notImplemented)
	return
}

func simGetAnalogueOffset(handle int16, voltageRange int, coupling Coupling) (maximumVoltage, minimumVoltage float32, err error) {
	maximumVoltage, minimumVoltage = 20, -20
	return
}

func simGetChannelInformation(handle int16, info int16, probe int32, ranges []int32, channels ChannelId) (lengthOfRanges int32, err error) {
	chRanges := []int32{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	lengthOfRanges = int32(len(chRanges))
	copy(ranges, chRanges)
	return
}

func simGetMaxDownSampleRatio(handle int16, noOfUnaggregatedSamples uint32, downSampleRatioMode RatioMode, segmentIndex int32) (maxDownSampleRatio uint32, err error) {
	err = fmt.Errorf(notImplemented)
	return
}

func simGetMaxSegments(handle int16) (maxSegments uint32, err error) {
	err = fmt.Errorf(notImplemented)
	return
}

func simGetNoOfCaptures(handle int16) (nCaptures uint32, err error) {
	err = fmt.Errorf(notImplemented)
	return
}

func simGetNoOfProcessedCaptures(handle int16) (nCaptures uint32, err error) {
	err = fmt.Errorf(notImplemented)
	return
}

func lpStreamingReadyGo(handle int16, noOfSamples int32, startIndex uint32, overflow int16,
	triggeredAt uint32, triggered, autoStop int16, param any) {
}

func simGetStreamingLatestValues(handle int16, lpStreamingReadyGoPar StreamingReady, param any) (err error) {
	if !streamingRunning {
		return fmt.Errorf("streaming not running")
	}

	elapsed := time.Since(streamingLastReadTime).Seconds()
	numNew := int32(elapsed / (streamingIntervalNs * 1e-9))
	if numNew <= 0 {
		lpStreamingReadyGoPar(handle, 0, 0, 0, 0, 0, 0, param)
		return nil
	}

	// Find the buffer length of the first enabled channel
	var activeBufLen int32 = 0
	for ch := 0; ch < MaxChannels; ch++ {
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
		for ch := 0; ch < MaxChannels; ch++ {
			if channels[ch].enabled && len(buffers[ch]) > 0 {
				for i := 0; i < int(writeCount); i++ {
					t := float64(totalSamplesGenerated+int64(i)) * (streamingIntervalNs * 1e-9)
					val := calculateSampleLevelAtTime(t, ChannelId(ch))

					var level int16
					if val > float64(maxValue) {
						level = maxValue
					} else if val < -float64(maxValue) {
						level = -maxValue
					} else {
						level = int16(math.Round(val))
					}
					// Write to buffer, handling bounds safely
					idx := (streamingWriteIndex + int32(i)) % activeBufLen
					if idx < int32(len(buffers[ch])) {
						buffers[ch][idx] = level
					}
				}
			}
		}

		totalSamplesGenerated += int64(writeCount)
		for i := range channels {
			if channels[i].sweepController != nil {
				channels[i].sweepController.Update()
			}
		}
		lpStreamingReadyGoPar(handle, writeCount, uint32(streamingWriteIndex), 0, 0, 0, 0, param)
		streamingWriteIndex = (streamingWriteIndex + writeCount) % activeBufLen
		streamingLastReadTime = streamingLastReadTime.Add(time.Duration(float64(writeCount)*streamingIntervalNs) * time.Nanosecond)
	} else {
		lpStreamingReadyGoPar(handle, 0, 0, 0, 0, 0, 0, param)
	}

	return nil
}
func numberOfEnabledChannels() int {
	count := 0
	for _, v := range channels {
		if v.enabled {
			count++
		}
	}
	return count
}
func simGetTimebaseError() (err error) {
	err = fmt.Errorf("Invalid time base")
	return
}
func simGetTimeInterval(timeBase uint32, nec int) (timeIntervalNanoseconds int32, err error) {
	switch {
	case timeBase == 0:
		if nec == 1 {
			timeIntervalNanoseconds = 1
		} else {
			err = simGetTimebaseError()
			return
		}
	case timeBase == 1:
		if nec <= 2 {
			timeIntervalNanoseconds = 2
		} else {
			err = simGetTimebaseError()
			return
		}
	case timeBase == 2:
		timeIntervalNanoseconds = 4
	default:
		timeIntervalNanoseconds = 1000 * (int32(timeBase) - 2) / 125
	}
	return
}

func simGetTimebase(handle int16, timeBase uint32, noOfSamples int32, overSample int16, segmentIndex uint32) (timeIntervalNanoseconds, maxSamples int32, err error) {
	nec := numberOfEnabledChannels()
	if nec == 0 {
		err = simGetTimebaseError()
		return
	}
	timeIntervalNanoseconds, err = simGetTimeInterval(timeBase, nec)
	if err != nil {
		return
	}
	switch nec {
	case 1:
		maxSamples = 64 * mega
	case 2:
		maxSamples = 32 * mega
	case 3:
		maxSamples = 16 * mega
	case 4:
		maxSamples = 16 * mega
	}
	return
}

func simGetTimeInterval2(timeBase uint32, nec int) (timeIntervalNanoseconds float64, err error) {
	switch {
	case timeBase == 0:
		if nec == 1 {
			timeIntervalNanoseconds = 1
		} else {
			err = simGetTimebaseError()
			return
		}
	case timeBase == 1:
		if nec <= 2 {
			timeIntervalNanoseconds = 2
		} else {
			err = simGetTimebaseError()
			return
		}
	case timeBase == 2:
		timeIntervalNanoseconds = 4
	default:
		timeIntervalNanoseconds = 1000 * (float64(timeBase) - 2) / float64(125)
	}
	return
}

func simGetTimebase2(handle int16, timeBase uint32, noOfSamples int32, overSample int16, segmentIndex uint32) (timeIntervalNanoseconds float32, maxSamples int32, err error) {
	// slog.Debug("simGetTimebase2")
	nec := numberOfEnabledChannels()
	if nec == 0 {
		err = simGetTimebaseError()
		return
	}
	t, err := simGetTimeInterval2(timeBase, nec)
	timeIntervalNanoseconds = float32(t)
	if err != nil {
		return
	}
	switch nec {
	case 1:
		maxSamples = 64 * mega
	case 2:
		maxSamples = 32 * mega
	case 3:
		maxSamples = 16 * mega
	case 4:
		maxSamples = 16 * mega
	}
	return
}

func simSetChannel(handle int16, channel ChannelId, enabled bool, couplingType Coupling, voltageRange RangeEnum, analogOffset float32) (err error) {
	slog.Debug("sim SetChannel", "index", channel, "enabled", enabled,
		"coupling", couplingType, "vrange", voltageRange, "analogOffset", analogOffset)
	if voltageRange < 0 {
		slog.Error("simSetChannel: negative voltageRange, clamping to 0", "vrange", voltageRange)
		voltageRange = 0
	}
	channels[channel].enabled = enabled
	channels[channel].vrange = voltageRange
	channels[channel].offset = float64(analogOffset)
	channels[channel].coupling = couplingType
	return
}

func simMaximumValue(handle int16) (value int16, err error) {
	value = maxValue
	return
}

func simMinimumValue(handle int16) (value int16, err error) {
	value = -maxValue
	return
}

var buffers [4][]int16

func simSetDataBuffer(handle int16, ch ChannelId, bufferIn []int16, segmentIndex uint32,
	mode RatioMode) (err error) {
	buffers[ch] = bufferIn
	return
}

func simSetDataBuffers(handle int16, ch ChannelId, bufferMax, bufferMin []int16, segmentIndex uint32, mode RatioMode) (err error) {
	slog.Error(notImplemented)
	err = fmt.Errorf(notImplemented)
	return
}

func simSetUnscaledDataBuffers(handle int16, ch ChannelId, bufferMax, bufferMin []int16, segmentIndex uint32, mode RatioMode) (err error) {
	slog.Error(notImplemented)
	err = fmt.Errorf(notImplemented)
	return
}

func simSetEtsTimeBuffer(handle int16, buffer []int64) (err error) {
	//64 bit buffers
	etsTimeBuffer = buffer
	return
}

func simSetEtsTimeBuffers(handle int16, timeUpper, timeLower []uint32) (err error) {
	//32 bit buffers
	slog.Error(notImplemented)
	err = fmt.Errorf(notImplemented)
	return
}

func simSetEts(handle int16, mode EtsMode, etsCycles int16, etsInterLeave int16) (sampleTimePicoseconds int32, err error) {
	etsEnbaled = mode != EtsOff
	if etsCycles < etsInterLeave {
		err = fmt.Errorf("etsCycles=%d < etsInterleave=%d", etsCycles, etsInterLeave)
		return
	}
	if etsCycles > etsInterLeave*10+9 {
		err = fmt.Errorf("etsCycles=%d > 10*etsInterleave+9=%d", etsCycles, 10*etsInterLeave+9)
		return
	}
	timeIntervalPicoSeconds = 2000 / float64(etsInterLeave)
	sampleTimePicoseconds = int32(timeIntervalPicoSeconds)
	return
}

func simRunStreaming(handle int16, reqSampleInterval uint32, sampleIntervalTimeUnits TimeUnits,
	maxPreTriggerSamples, maxPostTriggerSamples uint32,
	autoStop bool, downSampleRatio uint32, downSampleRatioMode RatioMode,
	overviewBufferSize uint32) (sampleInterval uint32, err error) {
	
	streamingIntervalNs = float64(reqSampleInterval) * TimeUnitToVal(sampleIntervalTimeUnits) * 1e9
	if streamingIntervalNs <= 0 {
		streamingIntervalNs = 1.0
	}
	
	streamingRunning = true
	running = true
	streamingStartTime = time.Now()
	streamingLastReadTime = time.Now()
	streamingWriteIndex = 0
	totalSamplesGenerated = 0
	
	sampleInterval = reqSampleInterval
	return sampleInterval, nil
}

var regLpBlockReadyGo BlockReady // registered go callback function

func lpBlockReadyGo(handle int16, status int, noOfSamples uint32, overflow int16, param any) {
}
func delayedCall(handle int16, lpBlockReadyGoPar BlockReady) {
	time.Sleep(callDelayMs * time.Millisecond)
	lpBlockReadyGoPar(handle, 0, nil)
}

func simRunBlock(handle int16, noOfPreTriggerSamples, noOfPostTriggerSamples int32,
	timeBase uint32, overSample int16, segmentIndex uint32, lpBlockReadyGoPar BlockReady,
	param any) (timeIndisposedMs int32, err error) {
	regLpBlockReadyGo = lpBlockReadyGoPar
	if !running {
		for i := range channels {
			if channels[i].sweepController != nil {
				channels[i].sweepController.Reset()
			}
		}
	}
	running = true
	timeBaseSet = timeBase
	nOfPreTrSamples = noOfPreTriggerSamples
	nOfPostTrSamples = noOfPostTriggerSamples
	go delayedCall(handle, lpBlockReadyGoPar)
	timeIndisposedMs = callDelayMs
	return
}

// simSetSimpleTrigger configures simple edge triggering.
func simSetSimpleTrigger(handle int16, enable bool, source ChannelId, threshold int16,
	direction ThresholdDirection, delay uint32, autoTriggerMs int16) (err error) {
	slog.Debug("SetSimpleTrigger", "threshold", threshold)
	triggerDetector = NewTriggerDetector(enable, threshold, 0, direction, source)
	triggerDetector.SetTriggerCalculationMode(TriggerCalculationMode)
	// triggerDelay = delay
	autoTrigger = int32(autoTriggerMs)
	complexTrigger = false
	return
}

// simSetTriggerChannelProperties configures advanced trigger channel properties.
func simSetTriggerChannelProperties(handle int16, channelProperties []TriggerChannelProperties, auxOutputEnable bool,
	autoTriggerMs int32) (err error) {
	slog.Debug("sim trigg prop", "channelProperties", channelProperties)
	for i := 0; i < len(simChannelProperties) && i < len(channelProperties); i++ {
		simChannelProperties[i] = channelProperties[i]
	}
	if triggerDetector == nil {
		triggerDetector = NewTriggerDetector(false, 0, 0, 0, 0)
		triggerDetector.SetTriggerCalculationMode(TriggerCalculationMode)
	}
	triggerDetector.SetChannelProperties(channelProperties)
	autoTrigger = autoTriggerMs
	return
}

// simSetTriggerChannelConditions configures trigger channel conditions.
func simSetTriggerChannelConditions(handle int16, triggerConditions []TriggerConditions) (err error) {
	slog.Debug("sim trigg cond", "simTriggerConditions", simTriggerConditions)
	for i := 0; i < len(simTriggerConditions) && i < len(triggerConditions); i++ {
		simTriggerConditions[i] = triggerConditions[i]
	}
	if triggerDetector != nil {
		triggerDetector.SetChannelConditions(triggerConditions)
	}
	complexTrigger = true
	return
}

// simSetTriggerChannelDirections configures trigger directions for each channel.
func simSetTriggerChannelDirections(handle int16, channelA, channelB, channelC, channelD, ext, aux ThresholdDirection) (err error) {
	channelAThresholdDirection, channelBThresholdDirection,
		channelCThresholdDirection, channelDThresholdDirection = channelA, channelB, channelC, channelD
	complexTrigger = true

	if triggerDetector != nil {
		triggerDetector.SetChannelDirections(channelA, channelB, channelC, channelD)
	}
	return
}

func simSetTriggerDelay(handle int16, delay uint32) (err error) {
	slog.Error(notImplemented)
	err = fmt.Errorf(notImplemented)
	return
}

func simSetPulseWidthQualifier(handle int16, conditions []PwqConditions, direction ThresholdDirection, lower, upper uint32,
	pwType PulseWidthType) (err error) {
	if triggerDetector != nil {
		triggerDetector.SetPulseWidthQualifier(conditions, direction, lower, upper, pwType)
	}
	return nil
}
func simSetTriggerDigitalPortProperties(handle int16, digitalDirections []DigitalChannelDirections) (err error) {
	slog.Error(notImplemented)
	err = fmt.Errorf(notImplemented)
	return
}

func simStop(handle int16) (err error) {
	running = false
	streamingRunning = false
	return
}

func SetTriggerCalculationMode(mode int) {
	TriggerCalculationMode = mode
	if triggerDetector != nil {
		triggerDetector.SetTriggerCalculationMode(mode)
	}
}

func simSetSigGenBuiltIn(handle int16, offsetVoltage int32, pkToPK uint32, waveType WaveTypeEnum,
	startFrequency, stopFrequency, increment, dwellTime float32, sweepType SweepTypeEnum,
	operation ExtraOperations, shots, sweeps uint32, triggerType SigGenTrigType,
	triggerSource SigGenTrigSource, extInThreshold int16) (err error) {
	err = fmt.Errorf(notImplemented)
	return
}

// simSetSigGenBuiltInV2 is a no-op for the simulator.
// The simulator manages generators per-channel via SetSimGen, so the global
// hardware path must not overwrite the per-channel sweep controllers that were
// already configured by SetSimGen calls. Clobbering them here would reset the
// sweep state on every block-mode restart (e.g. on a tab switch).
func simSetSigGenBuiltInV2(handle int16, offsetVoltage int32, pkToPK uint32, waveType WaveTypeEnum,
	startFrequency, stopFrequency, increment, dwellTime float64, sweepType SweepTypeEnum,
	operation ExtraOperations, shots, sweeps uint32, triggerType SigGenTrigType,
	triggerSource SigGenTrigSource, extInThreshold int16) (err error) {
	slog.Debug("simSetSigGenBuiltInV2 (no-op: sim uses per-channel SetSimGen)")
	return
}

func (s *SimDesc) SetSimGen(channel genericps.ChannelId, on bool, offsetVoltage int32, pkToPK uint32, waveType genericps.WaveTypeEnum,
	startFrequency, stopFrequency, increment, dwellTime float64, sweepType genericps.SweepTypeEnum,
	operation genericps.ExtraOperations, shots, sweeps uint32, triggerType genericps.SigGenTrigType,
	triggerSource genericps.SigGenTrigSource, extInThreshold int16, phase float64) (err error) {

	ch := int(channel)
	if ch < 0 || ch >= numberOfChannels {
		return fmt.Errorf("invalid channel")
	}

	channels[ch].genOn = on
	switch operation {
	case genericps.Prbs:
		channels[ch].genWaveFunction = NewPrbsGenerator()
	default:
		channels[ch].genWaveFunction = NewWaveformGenerator(WaveTypeEnum(waveType))
	}
	dwellDuration := time.Duration(dwellTime*1000000000) * time.Nanosecond
	channels[ch].sweepController = NewSweepController(startFrequency, stopFrequency, increment, SweepTypeEnum(sweepType), dwellDuration)
	channels[ch].genPkToPk = pkToPK
	channels[ch].genOffsetVoltage = offsetVoltage
	channels[ch].phase = phase

	return nil
}

func (s *SimDesc) SetSimRlcFilter(channel genericps.ChannelId, genSource genericps.ChannelId, enabled bool, filterType string, r float64, runit string, l float64, lunit string, c float64, cunit string) (err error) {
	ch := int(channel)
	if ch < 0 || ch >= MaxChannels {
		return fmt.Errorf("invalid channel")
	}
	channels[ch].genSource = ChannelId(genSource)
	channels[ch].rlcEnabled = enabled
	channels[ch].rlcType = filterType
	channels[ch].rlcR = r
	channels[ch].rlcRUnit = runit
	channels[ch].rlcL = l
	channels[ch].rlcLUnit = lunit
	channels[ch].rlcC = c
	channels[ch].rlcCUnit = cunit
	return nil
}

func simSigGenFrequencyToPhase(handle int16, frequency float64, indexMode IndexMode, bufferLength uint32) (phase uint32, err error) {
	slog.Error(notImplemented)
	err = fmt.Errorf(notImplemented)
	return
}

func simSetNoCaptures(handle int16, nCaptures uint32) (err error) {
	slog.Error(notImplemented)
	err = fmt.Errorf(notImplemented)
	return
}

func simGetTriggerTimeOffset(handle int16, segmentIndex uint32) (timeUpper, timeLower uint32, timeUnits TimeUnits, err error) {
	slog.Error(notImplemented)
	err = fmt.Errorf(notImplemented)
	return
}

func simGetTriggerTimeOffset64(handle int16, segmentIndex uint32) (t int64, timeUnits TimeUnits, err error) {
	t = int64(GetTriggerTimeOffset() * 1e15)
	timeUnits = TuFs
	return
}

func simGetValuesTriggerTimeOffsetBulk(handle int16, timesUpper, timesLower []uint32, timeUnits []TimeUnits,
	fromSegmentIndex, toSegmentIndex uint32) (err error) {
	slog.Error(notImplemented)
	err = fmt.Errorf(notImplemented)
	return
}

func simGetValuesTriggerTimeOffsetBulk64(handle int16, times []int64, timeUnits []TimeUnits,
	fromSegmentIndex, toSegmentIndex uint32) (err error) {
	slog.Error(notImplemented)
	err = fmt.Errorf(notImplemented)
	return
}

func simHoldOff(handle int16, holdOff uint64, holdOffType HoldOffType) (err error) {
	slog.Error(notImplemented)
	err = fmt.Errorf(notImplemented)
	return
}

func simLsReady(handle int16) (ready int16, err error) {
	slog.Error(notImplemented)
	err = fmt.Errorf(notImplemented)
	return
}

func simTriggerOrPulseWidthQualifierEnabled(handle int16) (triggerEnabled, pulseWidthQualifierEnabledint16 int16, err error) {
	slog.Error(notImplemented)
	err = fmt.Errorf(notImplemented)
	return
}

func simMemorySegments(handle int16, nSegments uint32) (nMaxSamples int32, err error) {
	nMaxSamples = 64 * mega
	return
}

func simNoOfStreamingValues(handle int16) (noOfValues uint32, err error) {
	slog.Error(notImplemented)
	err = fmt.Errorf(notImplemented)
	return
}

func OpenUnitProgress() (retHandle, progressPercent, complete int16, err error) {
	slog.Error(notImplemented)
	err = fmt.Errorf(notImplemented)
	return
}

func simPingUnit(handle int16) (err error) {
	slog.Error(notImplemented)
	err = fmt.Errorf(notImplemented)
	return
}

func simQueryOutputEdgeDetect(handle int16) (state int16, err error) {
	slog.Error(notImplemented)
	err = fmt.Errorf(notImplemented)
	return
}

func simSetDigitalAnalogTriggerOperand(handle int16, operand TriggerOperand) (err error) {
	slog.Error(notImplemented)
	err = fmt.Errorf(notImplemented)
	return
}

func simSetDigitalPort(handle int16, port DigitalPort, enabled bool, logiclevel int16) (err error) {
	slog.Error(notImplemented)
	err = fmt.Errorf(notImplemented)
	return
}

func simSetOutputEdgeDetect(handle int16, state int16) (err error) {
	slog.Error(notImplemented)
	err = fmt.Errorf(notImplemented)
	return
}

func simSetPulseWidthDigitalPortProperties(handle int16, digitalDirections []DigitalChannelDirections) (err error) {
	slog.Error(notImplemented)
	err = fmt.Errorf(notImplemented)
	return
}

func simSetSigGenArbitrary(handle int16, offsetVoltage int32, pkToPK uint32,
	startDeltaPhase, stopDeltaPhase, deltaPhaseIncrement, dwellCount uint32,
	arbitraryWaveform []int16, sweepType SweepTypeEnum, operation ExtraOperations,
	indexMode IndexMode, shots, sweeps uint32, triggerType SigGenTrigType,
	triggerSource SigGenTrigSource, extInThreshold int16) (err error) {
	slog.Error(notImplemented)
	err = fmt.Errorf(notImplemented)
	return
}

func simSetSigGenPropertiesArbitrary(handle int16, offsetVoltage int32,
	startDeltaPhase, stopDeltaPhase, deltaPhaseIncrement, dwellCount uint32,
	sweepType SweepTypeEnum, operation ExtraOperations,
	indexMode IndexMode, shots, sweeps uint32, triggerType SigGenTrigType,
	triggerSource SigGenTrigSource, extInThreshold int16) (err error) {
	slog.Error(notImplemented)
	err = fmt.Errorf(notImplemented)
	return
}

func simSetSigGenPropertiesBuiltIn(handle int16, offsetVoltage int32,
	startFrequency, stopFrequency, increment, dwellTime float64,
	sweepType SweepTypeEnum,
	shots, sweeps uint32, triggerType SigGenTrigType,
	triggerSource SigGenTrigSource, extInThreshold int16) (err error) {
	slog.Error(notImplemented)
	err = fmt.Errorf(notImplemented)
	return
}

func simSigGenArbitraryMinMaxValues(handle int16) (minArbitraryWaveformValue, maxArbitraryWaveformValue int16,
	minArbitraryWaveformSize, maxArbitraryWaveformSize uint32, err error) {
	slog.Error(notImplemented)
	err = fmt.Errorf(notImplemented)
	return
}

func simSigGenSoftwareControl(handle int16, state int16) (err error) {
	slog.Error(notImplemented)
	err = fmt.Errorf(notImplemented)
	return
}

func dispatch(msg genericps.Message) {
	switch m := msg.(type) {
	case *genericps.NullMsg:
		m.RspCh() <- struct{}{}
	case *genericps.SetChannelMsg:
		setChannel(m)
	case *genericps.GetUnitInfoMsg:
		getUnitInfo(m)
	case *genericps.CloseUnitMsg:
		closeUnit(m)
	case *genericps.FlashLedMsg:
		flashLed(m)
	case *genericps.GetValuesAsyncMsg:
		getValuesAsync(m)
	case *genericps.GetValuesBulkMsg:
		getValuesBulk(m)
	case *genericps.GetValuesOverlappedBulkMsg:
		getValuesOverlappedBulk(m)
	case *genericps.GetValuesOverlappedMsg:
		getValuesOverlapped(m)
	case *genericps.GetValuesMsg:
		getValues(m)
	case *genericps.GetAnalogueOffsetMsg:
		getAnalogueOffsetMsg(m)
	case *genericps.GetChannelInformationMsg:
		getChannelInformation(m)
	case *genericps.GetMaxDownSampleRatioMsg:
		getMaxDownSampleRatio(m)
	case *genericps.GetMaxSegmentsMsg:
		getMaxSegments(m)
	case *genericps.GetNumOfCapturesMsg:
		getNumberOfCaptures(m)
	case *genericps.GetNumOfProcessedCapturesMsg:
		getNumberOfProcessedCaptures(m)
	case *genericps.GetTimebaseMsg:
		getTimebase(m)
	case *genericps.GetTimebase2Msg:
		getTimebase2(m)
	case *genericps.MaximumValueMsg:
		maximumValue(m)
	case *genericps.MinimumValueMsg:
		minimumValue(m)
	case *genericps.SetSimpleTriggerMsg:
		setSimpleTrigger(m)
	case *genericps.SetDataBufferMsg:
		setDataBuffer(m)
	case *genericps.SetDataBuffersMsg:
		setDataBuffers(m)
	case *genericps.SetUnscaledDataBuffersMsg:
		setUnscaledDataBuffers(m)
	case *genericps.SetEtsTimeBufferMsg:
		setEtsTimeBuffer(m)
	case *genericps.SetEtsTimeBuffersMsg:
		setEtsTimeBuffers(m)
	case *genericps.SetEtsMsg:
		setEts(m)
	case *genericps.RunStreamingMsg:
		runStreaming(m)
	case *genericps.GetStreamingLatestValuesMsg:
		getStreamingLatestValues(m)
	case *genericps.RunBlockMsg:
		runBlock(m)
	case *genericps.SetTriggerChannelPropertiesMsg:
		setTriggerChannelProperties(m)
	case *genericps.SetTriggerChannelConditionsMsg:
		setTriggerChannelConditions(m)
	case *genericps.SetTriggerChannelDirectionsMsg:
		setTriggerChannelDirections(m)
	case *genericps.SetTriggerDelayMsg:
		setTriggerDelay(m)
	case *genericps.SetPulseWidthQualifierMsg:
		setPulseWidthQualifier(m)
	case *genericps.SetTriggerDigitalPortPropertiesMsg:
		setTriggerDigitalPortProperties(m)
	case *genericps.StopMsg:
		stop(m)
	case *genericps.SetSigGenBuiltInMsg:
		setSigGenBuiltIn(m)
	case *genericps.SetSigGenBuiltInV2Msg:
		setSigGenBuiltInV2(m)
	case *genericps.SetSimGenMsg:
		setSimGen(m)
	case *genericps.SetSimRlcFilterMsg:
		setSimRlcFilter(m)

	case *genericps.SigGenFrequencyToPhasenMsg:
		sigGenFrequencyToPhase(m)
	case *genericps.SetNumOfCapturesMsg:
		setNumOfCaptures(m)
	case *genericps.GetTriggerTimeOffsetMsg:
		getTriggerTimeOffset(m)
	case *genericps.GetTriggerTimeOffset64Msg:
		getTriggerTimeOffset64(m)
	case *genericps.GetValuesTriggerTimeOffsetBulkMsg:
		getValuesTriggerTimeOffsetBulk(m)
	case *genericps.GetValuesTriggerTimeOffsetBulk64Msg:
		getValuesTriggerTimeOffsetBulk64(m)
	case *genericps.HoldOffMsg:
		holdOff(m)
	case *genericps.LsReadyMsg:
		lsReady(m)
	case *genericps.TriggerOrPulseWidthQualifierEnabledMsg:
		triggerOrPulseWidthQualifierEnabled(m)
	case *genericps.MemorySegmentsMsg:
		memorySegments(m)
	case *genericps.NumOfStreamingValuesMsg:
		numOfStreamingValues(m)
	// case *genericps.OpenUnitProgressMsg:
	// 	openUnitProgress(m)
	case *genericps.PingUnitMsg:
		pingUnit(m)
	case *genericps.QueryOutputEdgeDetectMsg:
		queryOutputEdgeDetect(m)
	case *genericps.SetDigitalAnalogTriggerOperandMsg:
		setDigitalAnalogTriggerOperand(m)
	case *genericps.SetDigitalPortMsg:
		setDigitalPort(m)
	case *genericps.SetOutputEdgeDetectMsg:
		setOutputEdgeDetect(m)
	case *genericps.SetPulseWidthDigitalPortPropertiesMsg:
		setPulseWidthDigitalPortProperties(m)
	case *genericps.SetSigGenArbitraryMsg:
		setSigGenArbitrary(m)
	case *genericps.SetSigGenPropertiesArbitraryMsg:
		setSigGenPropertiesArbitrary(m)
	case *genericps.SetSigGenPropertiesBuiltInMsg:
		setSigGenPropertiesBuiltIn(m)
	case *genericps.SigGenArbitraryMinMaxValuesMsg:
		sigGenArbitraryMinMaxValues(m)
	case *genericps.SigGenSoftwareControlMsg:
		sigGenSoftwareControl(m)
	default:
		// slog.Error("dispatch unhandled", "type", reflect.TypeOf(msg), "msg", msg)
		slog.Error("dispatch unhandled", "m", m, "msg", msg)
	}
}
func init() {
	var scopeHandler genericps.ScopeHandler
	scopeHandler.Dispatch = dispatch
	scopeHandler.EnumerateUnits = EnumerateUnits
	scopeHandler.OpenUnit = openUnit
	scopeHandler.OpenUnitAsync = openUnitAsync
	scopeHandler.OpenUnitProgress = openUnitProgress
	scopeHandler.Id = genericps.SimId
	genericps.Register(scopeHandler)
}
func SetChannelCount(n int, explicit bool) error {
	if !explicit {
		return nil
	}
	if n < MinChannels || n > MaxChannels {
		return fmt.Errorf("Invalid channel number")
	}
	numberOfChannels = n
	scopeVariantInfo = fmt.Sprintf("2%d07SIM", n)
	scopeBathAndSerialInfo = fmt.Sprintf("SIM/CH%d", n)
	return nil
}
