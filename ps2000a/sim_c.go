//go:build sim

package ps2000a

/*
#cgo linux CFLAGS: -I/opt/picoscope/include/libps2000a
#cgo windows CFLAGS: -I"C:/Program Files/Pico Technology/SDK/inc"
#include <stdlib.h>
#include <PicoStatus.h>
#include <ps2000aApi.h>

// Forward declarations of Go exports (defined by CGO from the //export directives below).
extern uint32_t Gops2000aEnumerateUnits(int16_t *count, int8_t *serials, int16_t *serialLth);
extern uint32_t Gops2000aOpenUnit(int16_t *handle, int8_t *serial);
extern uint32_t Gops2000aOpenUnitAsync(int16_t *status, int8_t *serial);
extern uint32_t Gops2000aCloseUnit(int16_t handle);
extern uint32_t Gops2000aGetUnitInfo(int16_t handle, int8_t *stringData, int16_t stringLength, int16_t *requiredSize, uint32_t info);
extern uint32_t Gops2000aSetChannel(int16_t handle, int32_t channel, int16_t enabled, int32_t type, int32_t range, float analogOffset);
extern uint32_t Gops2000aGetAnalogueOffset(int16_t handle, int32_t range, int32_t coupling, float *maximumOffset, float *minimumOffset);
extern uint32_t Gops2000aSetSimpleTrigger(int16_t handle, int16_t enable, int32_t source, int16_t threshold, int32_t direction, uint32_t delay, int16_t autoTrigger_ms);
extern uint32_t Gops2000aSetDataBuffer(int16_t handle, int32_t channel, int16_t *buffer, int32_t bufferLth, uint32_t segmentIndex, int32_t mode);
extern uint32_t Gops2000aRunBlock(int16_t handle, int32_t noOfPreTriggerSamples, int32_t noOfPostTriggerSamples, uint32_t timebase, int16_t oversample, int32_t *timeIndisposedMs, uint32_t segmentIndex, ps2000aBlockReady lpReady, void *pParameter);
extern uint32_t Gops2000aStop(int16_t handle);
extern uint32_t Gops2000aSetSigGenBuiltIn(int16_t handle, int32_t offsetVoltage, uint32_t pkToPk, int16_t waveType, float startFrequency, float stopFrequency, float increment, float dwellTime, int32_t sweepType, int32_t operation, uint32_t shots, uint32_t sweeps, int32_t triggerType, int32_t triggerSource, int16_t extInThreshold);
extern uint32_t Gops2000aSetSigGenBuiltInV2(int16_t handle, int32_t offsetVoltage, uint32_t pkToPk, int16_t waveType, double startFrequency, double stopFrequency, double increment, double dwellTime, int32_t sweepType, int32_t operation, uint32_t shots, uint32_t sweeps, int32_t triggerType, int32_t triggerSource, int16_t extInThreshold);
extern uint32_t Gops2000aIsReady(int16_t handle, int16_t *ready);
extern uint32_t Gops2000aGetValues(int16_t handle, uint32_t startIndex, uint32_t *noOfSamples, uint32_t downSampleRatio, int32_t downSampleRatioMode, uint32_t segmentIndex, int16_t *overflow);
extern uint32_t Gops2000aMaximumValue(int16_t handle, int16_t *value);
extern uint32_t Gops2000aMinimumValue(int16_t handle, int16_t *value);
extern uint32_t Gops2000aGetTimebase(int16_t handle, uint32_t timebase, int32_t noSamples, int32_t *timeIntervalNanoseconds, int16_t oversample, int32_t *maxSamples, uint32_t segmentIndex);
extern uint32_t Gops2000aGetTimebase2(int16_t handle, uint32_t timebase, int32_t noSamples, float *timeIntervalNanoseconds, int16_t oversample, int32_t *maxSamples, uint32_t segmentIndex);
extern uint32_t Gops2000aFlashLed(int16_t handle, int16_t start);
extern uint32_t Gops2000aGetChannelInformation(int16_t handle, int32_t info, int32_t probe, int32_t *ranges, int32_t *length, int32_t channels);
extern uint32_t Gops2000aSetEts(int16_t handle, int32_t mode, int16_t etsCycles, int16_t etsInterleave, int32_t *sampleTimePicoseconds);
extern uint32_t Gops2000aSetEtsTimeBuffer(int16_t handle, int64_t *buffer, int32_t bufferLth);
extern uint32_t Gops2000aSetEtsTimeBuffers(int16_t handle, uint32_t *timeUpper, uint32_t *timeLower, int32_t bufferLth);
extern uint32_t Gops2000aGetMaxEtsValues(int16_t handle, int16_t *etsCycles, int16_t *etsInterleave);
extern uint32_t Gops2000aSetTriggerChannelProperties(int16_t handle, void *channelProperties, int16_t nChannelProperties, int16_t auxOutputEnable, int32_t autoTriggerMilliseconds);
extern uint32_t Gops2000aSetTriggerChannelConditions(int16_t handle, void *conditions, int16_t nConditions);
extern uint32_t Gops2000aSetTriggerChannelDirections(int16_t handle, int32_t channelA, int32_t channelB, int32_t channelC, int32_t channelD, int32_t ext, int32_t aux);
extern uint32_t Gops2000aSetTriggerDelay(int16_t handle, uint32_t delay);
extern uint32_t Gops2000aSetDataBuffers(int16_t handle, int32_t channelOrPort, int16_t *bufferMax, int16_t *bufferMin, int32_t bufferLth, uint32_t segmentIndex, int32_t mode);
extern uint32_t Gops2000aIsTriggerOrPulseWidthQualifierEnabled(int16_t handle, int16_t *triggerEnabled, int16_t *pulseWidthQualifierEnabled);
extern uint32_t Gops2000aSetPulseWidthQualifier(int16_t handle, void *conditions, int16_t nConditions, int32_t direction, uint32_t lower, uint32_t upper, int32_t type);
extern uint32_t Gops2000aGetTriggerTimeOffset64(int16_t handle, int64_t *time, int32_t *timeUnits, uint32_t segmentIndex);
extern uint32_t Gops2000aGetMaxDownSampleRatio(int16_t handle, uint32_t noOfUnaggregatedSamples, uint32_t *maxDownSampleRatio, int32_t downSampleRatioMode, uint32_t segmentIndex);
extern uint32_t Gops2000aSetDigitalPort(int16_t handle, int32_t port, int16_t enabled, int16_t logicLevel);
extern uint32_t Gops2000aSetTriggerDigitalPortProperties(int16_t handle, void *directions, int16_t nDirections);
extern uint32_t Gops2000aSetDigitalAnalogTriggerOperand(int16_t handle, int32_t operand);
extern uint32_t Gops2000aSetPulseWidthDigitalPortProperties(int16_t handle, void *directions, int16_t nDirections);
extern uint32_t Gops2000aRunStreaming(int16_t handle, uint32_t *sampleInterval, int32_t timeUnits, uint32_t maxPreTriggerSamples, uint32_t maxPostTriggerSamples, int16_t autoStop, uint32_t downSampleRatio, int32_t downSampleRatioMode, uint32_t overviewBufferSize);
extern uint32_t Gops2000aGetStreamingLatestValues(int16_t handle, void *lpDataReady, void *pParameter);
extern uint32_t Gops2000aNoOfStreamingValues(int16_t handle, uint32_t *noOfValues);

// Static helper to invoke a ps2000aBlockReady function pointer from Go.
static inline void call_ps2000aBlockReady(ps2000aBlockReady fp, int16_t handle, PICO_STATUS status, void *pParameter) {
    if (fp != NULL) {
        fp(handle, status, pParameter);
    }
}

static inline void call_ps2000aStreamingReady(ps2000aStreamingReady fp, int16_t handle, int32_t noOfSamples, uint32_t startIndex, int16_t overflow, uint32_t triggerAt, int16_t triggered, int16_t autoStop, void *pParameter) {
    if (fp != NULL) {
        fp(handle, noOfSamples, startIndex, overflow, triggerAt, triggered, autoStop, pParameter);
    }
}
*/
import "C"
import (
	"fynescope/demo"
	"math"
	"unsafe"
)

const (
	picoOk                             = 0x00000000
	picoMaxUnitsOpened                 = 0x00000001
	picoMemoryFail                     = 0x00000002
	picoNotFound                       = 0x00000003
	picoNotResponding                  = 0x00000007
	picoInvalidHandle                  = 0x0000000C
	picoInvalidParameter               = 0x0000000D
	picoInvalidTimebase                = 0x0000000E
	picoInvalidVoltageRange            = 0x0000000F
	picoInvalidChannel                 = 0x00000010
	picoNullParameter                  = 0x00000016
	picoTooManySamples                 = 0x0000001D
	picoTooManySegments                = 0x0000001E
	picoPulseWidthQualifier            = 0x0000001F
	picoSegmentOutOfRange              = 0x00000026
	picoInvalidInfo                    = 0x00000029
	picoInfoUnavailable                = 0x0000002A
	picoInvalidSampleInterval          = 0x0000002B
	picoSigGenParam                    = 0x0000002E
	picoSiggenOutputOverVoltage        = 0x00000035
	picoInvalidBuffer                  = 0x00000037
	picoInvalidCoupling                = 0x00000045
	picoAnalogueOffsetOutOfRange       = 0x00000059
	picoInvalidDigitalPort             = 0x00000113
	picoInvalidTriggerProperty         = 0x0000000D
)

//export Gops2000aMemorySegments
func Gops2000aMemorySegments(handle C.int16_t, nSegments C.uint32_t, nMaxSamples *C.int32_t) C.uint32_t {
	if handle <= 0 {
		return picoInvalidHandle
	}
	if nMaxSamples == nil {
		return picoNullParameter
	}
	if nSegments == 0 {
		return picoInvalidParameter
	}
	if nSegments > 128 {
		return picoTooManySegments
	}
	cfg := GetSimConfig()
	*nMaxSamples = C.int32_t(cfg.MaxMemorySamples)
	return picoOk
}

//export Gops2000aEnumerateUnits
func Gops2000aEnumerateUnits(count *C.int16_t, serials *C.int8_t, serialLth *C.int16_t) C.uint32_t {
	if count == nil || serials == nil || serialLth == nil {
		return picoNullParameter
	}
	*count = 1
	*serials = '2'
	return picoOk
}

//export Gops2000aOpenUnit
func Gops2000aOpenUnit(handle *C.int16_t, serial *C.int8_t) C.uint32_t {
	if handle == nil {
		return picoNullParameter
	}
	*handle = 1
	if serial != nil {
		*serial = '2'
	}
	simOpenUnit(int16(*handle))
	return picoOk
}

//export Gops2000aOpenUnitAsync
func Gops2000aOpenUnitAsync(status *C.int16_t, serial *C.int8_t) C.uint32_t {
	if status == nil {
		return picoNullParameter
	}
	*status = 0
	return picoOk
}

//export Gops2000aCloseUnit
func Gops2000aCloseUnit(handle C.int16_t) C.uint32_t {
	if handle <= 0 {
		return picoInvalidHandle
	}
	return picoOk
}

//export Gops2000aGetUnitInfo
func Gops2000aGetUnitInfo(handle C.int16_t, stringData *C.int8_t, stringLength C.int16_t, requiredSize *C.int16_t, info C.uint32_t) C.uint32_t {
	if handle <= 0 && info != 0 {
		return picoInvalidHandle
	}
	if info > 0x10 && info != 0x10000001 {
		return picoInvalidInfo
	}
	if stringData == nil && requiredSize == nil {
		return picoNullParameter
	}
	str := demo.ScopeSimVariantInfo
	if requiredSize != nil {
		*requiredSize = C.int16_t(len(str)) + 1
	}
	if stringData != nil {
		if stringLength < C.int16_t(len(str))+1 {
			return picoInvalidParameter
		}
		ptr := (*[1 << 20]C.int8_t)(unsafe.Pointer(stringData))
		for i := 0; i < len(str); i++ {
			ptr[i] = C.int8_t(str[i])
		}
		ptr[len(str)] = 0
	}
	return picoOk
}

//export Gops2000aSetChannel
func Gops2000aSetChannel(handle C.int16_t, channel C.int32_t, enabled C.int16_t, dc C.int32_t, rangeEnum C.int32_t, analogOffset C.float) C.uint32_t {
	if handle <= 0 {
		return picoInvalidHandle
	}
	if channel < 0 || channel > 3 {
		return picoInvalidChannel
	}
	if dc < 0 || dc > 1 {
		return picoInvalidCoupling
	}
	cfg := GetSimConfig()
	if int32(rangeEnum) < cfg.MinRange || int32(rangeEnum) > cfg.MaxRange {
		return picoInvalidVoltageRange
	}
	maxOff, minOff := cfg.GetAnalogueOffsetVals(int32(rangeEnum), Coupling(dc))
	if float32(analogOffset) > maxOff || float32(analogOffset) < minOff {
		return picoAnalogueOffsetOutOfRange
	}
	simSetChannel(int16(handle), int(channel), enabled != 0, int(dc), int(rangeEnum), float32(analogOffset))
	return picoOk
}

//export Gops2000aGetAnalogueOffset
func Gops2000aGetAnalogueOffset(handle C.int16_t, rangeEnum C.int32_t, coupling C.int32_t, maximumOffset *C.float, minimumOffset *C.float) C.uint32_t {
	if handle <= 0 {
		return picoInvalidHandle
	}
	if maximumOffset == nil && minimumOffset == nil {
		return picoNullParameter
	}
	if coupling < 0 || coupling > 1 {
		return picoInvalidCoupling
	}
	cfg := GetSimConfig()
	if int32(rangeEnum) < cfg.MinRange || int32(rangeEnum) > cfg.MaxRange {
		return picoInvalidVoltageRange
	}
	maxOff, minOff := cfg.GetAnalogueOffsetVals(int32(rangeEnum), Coupling(coupling))
	if maximumOffset != nil {
		*maximumOffset = C.float(maxOff)
	}
	if minimumOffset != nil {
		*minimumOffset = C.float(minOff)
	}
	return picoOk
}

//export Gops2000aSetSimpleTrigger
func Gops2000aSetSimpleTrigger(handle C.int16_t, enable C.int16_t, source C.int32_t, threshold C.int16_t, direction C.int32_t, delay C.uint32_t, autoTriggerMs C.int16_t) C.uint32_t {
	if handle <= 0 {
		return picoInvalidHandle
	}
	if source < 0 || (source > 3 && source != 4 && source != 5) {
		return picoInvalidChannel
	}
	if direction < 0 || direction > 4 {
		return picoInvalidParameter
	}
	if autoTriggerMs < 0 {
		return picoInvalidParameter
	}
	simSetSimpleTrigger(int16(handle), enable != 0, int(source), int16(threshold), int(direction), uint32(delay), int16(autoTriggerMs))
	return picoOk
}

//export Gops2000aSetDataBuffer
func Gops2000aSetDataBuffer(handle C.int16_t, channel C.int32_t, buffer *C.int16_t, bufferLth C.int32_t, segmentIndex C.uint32_t, mode C.int32_t) C.uint32_t {
	if handle <= 0 {
		return picoInvalidHandle
	}
	if buffer == nil {
		return picoNullParameter
	}
	if bufferLth <= 0 {
		return picoInvalidParameter
	}
	if channel < 0 || (channel > 3 && channel != 0x80 && channel != 0x81 && channel != 128 && channel != 129) {
		return picoInvalidChannel
	}
	if segmentIndex != 0 {
		return picoSegmentOutOfRange
	}
	slice := unsafe.Slice((*int16)(unsafe.Pointer(buffer)), int(bufferLth))
	simSetDataBufferWithMode(int16(handle), int(channel), slice, uint32(segmentIndex), int32(mode))
	return picoOk
}

//export Gops2000aRunBlock
func Gops2000aRunBlock(handle C.int16_t, noOfPreTriggerSamples C.int32_t, noOfPostTriggerSamples C.int32_t, timebase C.uint32_t, oversample C.int16_t, timeIndisposedMs *C.int32_t, segmentIndex C.uint32_t, lpReady C.ps2000aBlockReady, pParameter unsafe.Pointer) C.uint32_t {
	if handle <= 0 {
		return picoInvalidHandle
	}
	if noOfPreTriggerSamples < 0 || noOfPostTriggerSamples < 0 || (noOfPreTriggerSamples == 0 && noOfPostTriggerSamples == 0) {
		return picoInvalidParameter
	}
	cfg := GetSimConfig()
	if int64(noOfPreTriggerSamples)+int64(noOfPostTriggerSamples) > int64(cfg.MaxSamples) {
		return picoTooManySamples
	}
	if segmentIndex != 0 {
		return picoSegmentOutOfRange
	}
	if timeIndisposedMs != nil {
		*timeIndisposedMs = 0
	}
	simRunBlock(int16(handle), int32(noOfPreTriggerSamples), int32(noOfPostTriggerSamples), uint32(timebase), func(h int16, status int32) {
		C.call_ps2000aBlockReady(lpReady, C.int16_t(h), C.uint32_t(status), pParameter)
	})
	return picoOk
}

//export Gops2000aStop
func Gops2000aStop(handle C.int16_t) C.uint32_t {
	if handle <= 0 {
		return picoInvalidHandle
	}
	simStop(int16(handle))
	return picoOk
}

//export Gops2000aSetSigGenBuiltIn
func Gops2000aSetSigGenBuiltIn(handle C.int16_t, offsetVoltage C.int32_t, pkToPk C.uint32_t, waveType C.int16_t, startFrequency C.float, stopFrequency C.float, increment C.float, dwellTime C.float, sweepType C.int32_t, operation C.int32_t, shots C.uint32_t, sweeps C.uint32_t, triggerType C.int32_t, triggerSource C.int32_t, extInThreshold C.int16_t) C.uint32_t {
	if handle <= 0 {
		return picoInvalidHandle
	}
	if waveType < 0 || waveType > 9 {
		return picoSigGenParam
	}
	totalV := uint64(pkToPk) + uint64(2*math.Abs(float64(offsetVoltage)))
	if totalV > 4000000 {
		return picoSiggenOutputOverVoltage
	}
	if startFrequency <= 0 || stopFrequency <= 0 {
		return picoSigGenParam
	}
	simSetSigGenBuiltIn(int16(handle), int32(offsetVoltage), uint32(pkToPk), int(waveType), float64(startFrequency), float64(stopFrequency), float64(increment), float64(dwellTime), int(sweepType), int(operation))
	return picoOk
}

//export Gops2000aSetSigGenBuiltInV2
func Gops2000aSetSigGenBuiltInV2(handle C.int16_t, offsetVoltage C.int32_t, pkToPk C.uint32_t, waveType C.int16_t, startFrequency C.double, stopFrequency C.double, increment C.double, dwellTime C.double, sweepType C.int32_t, operation C.int32_t, shots C.uint32_t, sweeps C.uint32_t, triggerType C.int32_t, triggerSource C.int32_t, extInThreshold C.int16_t) C.uint32_t {
	if handle <= 0 {
		return picoInvalidHandle
	}
	if waveType < 0 || waveType > 9 {
		return picoSigGenParam
	}
	totalV := uint64(pkToPk) + uint64(2*math.Abs(float64(offsetVoltage)))
	if totalV > 4000000 {
		return picoSiggenOutputOverVoltage
	}
	if startFrequency <= 0 || stopFrequency <= 0 {
		return picoSigGenParam
	}
	simSetSigGenBuiltIn(int16(handle), int32(offsetVoltage), uint32(pkToPk), int(waveType), float64(startFrequency), float64(stopFrequency), float64(increment), float64(dwellTime), int(sweepType), int(operation))
	return picoOk
}

//export Gops2000aSetSigGenArbitrary
func Gops2000aSetSigGenArbitrary(handle C.int16_t, offsetVoltage C.int32_t, pkToPk C.uint32_t, startDeltaPhase C.uint32_t, stopDeltaPhase C.uint32_t, deltaPhaseIncrement C.uint32_t, dwellCount C.uint32_t, arbitraryWaveform *C.int16_t, arbitraryWaveformSize C.int32_t, sweepType C.int32_t, operation C.int32_t, indexMode C.int32_t, shots C.uint32_t, sweeps C.uint32_t, triggerType C.int32_t, triggerSource C.int32_t, extInThreshold C.int16_t) C.uint32_t {
	if handle <= 0 {
		return picoInvalidHandle
	}
	if arbitraryWaveform == nil {
		return picoNullParameter
	}
	if arbitraryWaveformSize < 10 || arbitraryWaveformSize > 32768 {
		return picoSigGenParam
	}
	totalV := uint64(pkToPk) + uint64(2*math.Abs(float64(offsetVoltage)))
	if totalV > 4000000 {
		return picoSiggenOutputOverVoltage
	}
	slice := unsafe.Slice((*int16)(arbitraryWaveform), int(arbitraryWaveformSize))
	wfCopy := make([]int16, len(slice))
	copy(wfCopy, slice)
	simSetSigGenArbitrary(int16(handle), int32(offsetVoltage), uint32(pkToPk), uint32(startDeltaPhase), uint32(stopDeltaPhase), uint32(deltaPhaseIncrement), uint32(dwellCount), wfCopy, int(sweepType), int(operation), int(indexMode), uint32(shots), uint32(sweeps), int(triggerType), int(triggerSource), int16(extInThreshold))
	return picoOk
}

//export Gops2000aSigGenFrequencyToPhase
func Gops2000aSigGenFrequencyToPhase(handle C.int16_t, frequency C.double, indexMode C.int32_t, bufferLength C.uint32_t, phase *C.uint32_t) C.uint32_t {
	if handle <= 0 {
		return picoInvalidHandle
	}
	if phase == nil {
		return picoNullParameter
	}
	if frequency <= 0 || bufferLength == 0 {
		return picoSigGenParam
	}
	p := simSigGenFrequencyToPhase(int16(handle), float64(frequency), int(indexMode), uint32(bufferLength))
	*phase = C.uint32_t(p)
	return picoOk
}

//export Gops2000aIsReady
func Gops2000aIsReady(handle C.int16_t, ready *C.int16_t) C.uint32_t {
	if handle <= 0 {
		return picoInvalidHandle
	}
	if ready == nil {
		return picoNullParameter
	}
	r := simIsReady(int16(handle))
	if r {
		*ready = 1
	} else {
		*ready = 0
	}
	return picoOk
}

//export Gops2000aGetValues
func Gops2000aGetValues(handle C.int16_t, startIndex C.uint32_t, noOfSamples *C.uint32_t, downSampleRatio C.uint32_t, downSampleRatioMode C.int32_t, segmentIndex C.uint32_t, overflow *C.int16_t) C.uint32_t {
	if handle <= 0 {
		return picoInvalidHandle
	}
	if noOfSamples == nil {
		return picoNullParameter
	}
	if segmentIndex != 0 {
		return picoSegmentOutOfRange
	}
	samples, ov := simGetValues(int16(handle), uint32(startIndex), uint32(*noOfSamples))
	*noOfSamples = C.uint32_t(samples)
	if overflow != nil {
		*overflow = C.int16_t(ov)
	}
	return picoOk
}

//export Gops2000aMaximumValue
func Gops2000aMaximumValue(handle C.int16_t, value *C.int16_t) C.uint32_t {
	if handle <= 0 {
		return picoInvalidHandle
	}
	if value == nil {
		return picoNullParameter
	}
	*value = 32767
	return picoOk
}

//export Gops2000aMinimumValue
func Gops2000aMinimumValue(handle C.int16_t, value *C.int16_t) C.uint32_t {
	if handle <= 0 {
		return picoInvalidHandle
	}
	if value == nil {
		return picoNullParameter
	}
	*value = -32767
	return picoOk
}

//export Gops2000aGetTimebase2
func Gops2000aGetTimebase2(handle C.int16_t, timebase C.uint32_t, noSamples C.int32_t, timeIntervalNanoseconds *C.float, oversample C.int16_t, maxSamples *C.int32_t, segmentIndex C.uint32_t) C.uint32_t {
	if handle <= 0 {
		return picoInvalidHandle
	}
	if timeIntervalNanoseconds == nil && maxSamples == nil {
		return picoNullParameter
	}
	if segmentIndex != 0 {
		return picoSegmentOutOfRange
	}
	cfg := GetSimConfig()
	if timeIntervalNanoseconds != nil {
		*timeIntervalNanoseconds = C.float(cfg.TimebaseToNs(uint32(timebase)))
	}
	if maxSamples != nil {
		*maxSamples = C.int32_t(cfg.MaxSamples)
	}
	return picoOk
}

//export Gops2000aFlashLed
func Gops2000aFlashLed(handle C.int16_t, start C.int16_t) C.uint32_t {
	if handle <= 0 {
		return picoInvalidHandle
	}
	return picoOk
}

//export Gops2000aGetChannelInformation
func Gops2000aGetChannelInformation(handle C.int16_t, info C.int32_t, probe C.int32_t, ranges *C.int32_t, length *C.int32_t, channels C.int32_t) C.uint32_t {
	if handle <= 0 {
		return picoInvalidHandle
	}
	if length == nil {
		return picoNullParameter
	}
	if channels < 0 || channels > 3 {
		return picoInvalidChannel
	}
	if info != 0 {
		return picoInvalidInfo
	}
	cfg := GetSimConfig()
	numRanges := C.int32_t(cfg.MaxRange - cfg.MinRange + 1)
	if ranges == nil {
		// Just report how many ranges are available
		*length = numRanges
		return picoOk
	}
	// Fill the ranges array
	if *length > numRanges {
		*length = numRanges
	}
	slice := unsafe.Slice(ranges, int(*length))
	for i := C.int32_t(0); i < *length; i++ {
		slice[i] = C.int32_t(cfg.MinRange + int32(i))
	}
	return picoOk
}

//export Gops2000aSetEts
func Gops2000aSetEts(handle C.int16_t, mode C.int32_t, etsCycles C.int16_t, etsInterleave C.int16_t, sampleTimePicoseconds *C.int32_t) C.uint32_t {
	if handle <= 0 {
		return picoInvalidHandle
	}
	if mode < 0 || mode > 2 {
		return picoInvalidParameter
	}
	simSetEts(int16(handle), int(mode), int16(etsCycles), int16(etsInterleave), (*int32)(unsafe.Pointer(sampleTimePicoseconds)))
	return picoOk
}

//export Gops2000aSetEtsTimeBuffer
func Gops2000aSetEtsTimeBuffer(handle C.int16_t, buffer *C.int64_t, bufferLth C.int32_t) C.uint32_t {
	if handle <= 0 {
		return picoInvalidHandle
	}
	if buffer == nil {
		return picoNullParameter
	}
	if bufferLth <= 0 {
		return picoInvalidParameter
	}
	slice := unsafe.Slice((*int64)(unsafe.Pointer(buffer)), int(bufferLth))
	simSetEtsTimeBuffer(int16(handle), slice)
	return picoOk
}

//export Gops2000aSetEtsTimeBuffers
func Gops2000aSetEtsTimeBuffers(handle C.int16_t, timeUpper *C.uint32_t, timeLower *C.uint32_t, bufferLth C.int32_t) C.uint32_t {
	if handle <= 0 {
		return picoInvalidHandle
	}
	if timeUpper == nil || timeLower == nil {
		return picoNullParameter
	}
	if bufferLth <= 0 {
		return picoInvalidParameter
	}
	sliceUpper := unsafe.Slice((*uint32)(unsafe.Pointer(timeUpper)), int(bufferLth))
	sliceLower := unsafe.Slice((*uint32)(unsafe.Pointer(timeLower)), int(bufferLth))
	simSetEtsTimeBuffers(int16(handle), sliceUpper, sliceLower)
	return picoOk
}

//export Gops2000aGetMaxEtsValues
func Gops2000aGetMaxEtsValues(handle C.int16_t, etsCycles *C.int16_t, etsInterleave *C.int16_t) C.uint32_t {
	if handle <= 0 {
		return picoInvalidHandle
	}
	if etsCycles == nil && etsInterleave == nil {
		return picoNullParameter
	}
	if etsCycles != nil {
		*etsCycles = 250
	}
	if etsInterleave != nil {
		*etsInterleave = 50
	}
	return picoOk
}

//export Gops2000aSetTriggerChannelProperties
func Gops2000aSetTriggerChannelProperties(handle C.int16_t, channelProperties unsafe.Pointer, nChannelProperties C.int16_t, auxOutputEnable C.int16_t, autoTriggerMilliseconds C.int32_t) C.uint32_t {
	if handle <= 0 {
		return picoInvalidHandle
	}
	if channelProperties == nil && nChannelProperties > 0 {
		return picoNullParameter
	}
	if nChannelProperties < 0 {
		return picoInvalidParameter
	}
	cProps := unsafe.Slice((*C.PS2000A_TRIGGER_CHANNEL_PROPERTIES)(channelProperties), int(nChannelProperties))
	var props []demo.TriggerChannelProperties
	for _, p := range cProps {
		if int(p.channel) < 0 || int(p.channel) > 3 {
			return picoInvalidChannel
		}
		if p.thresholdMode != 0 && p.thresholdMode != 1 {
			return picoInvalidParameter
		}
		if p.thresholdMode == 1 {
			if int16(p.thresholdLower) > int16(p.thresholdUpper) {
				return picoInvalidTriggerProperty
			}
			// In window mode, hardware requires window height (upper - lower) to be at least 512 ADC counts
			if int32(p.thresholdUpper)-int32(p.thresholdLower) < 512 {
				return picoInvalidTriggerProperty
			}
			// In window mode, upper threshold must be greater than or equal to lower threshold + upper hysteresis + lower hysteresis
			// To avoid overflow, we can do it in int32:
			if int32(p.thresholdUpper)-int32(p.thresholdUpperHysteresis) < int32(p.thresholdLower)+int32(p.thresholdLowerHysteresis) {
				return picoInvalidTriggerProperty
			}
		}
		props = append(props, demo.TriggerChannelProperties{
			ThresholdUpper:           int16(p.thresholdUpper),
			ThresholdUpperHysteresis: uint16(p.thresholdUpperHysteresis),
			ThresholdLower:           int16(p.thresholdLower),
			ThresholdLowerHysteresis: uint16(p.thresholdLowerHysteresis),
			Channel:                  demo.ChannelId(p.channel),
			ThresholdMode:            demo.ThresholdModeId(p.thresholdMode),
		})
	}
	simSetTriggerChannelProperties(int16(handle), props, auxOutputEnable != 0, int32(autoTriggerMilliseconds))
	return picoOk
}

//export Gops2000aSetTriggerChannelConditions
func Gops2000aSetTriggerChannelConditions(handle C.int16_t, conditions unsafe.Pointer, nConditions C.int16_t) C.uint32_t {
	if handle <= 0 {
		return picoInvalidHandle
	}
	if conditions == nil && nConditions > 0 {
		return picoNullParameter
	}
	if nConditions < 0 {
		return picoInvalidParameter
	}
	cConds := unsafe.Slice((*C.PS2000A_TRIGGER_CONDITIONS)(conditions), int(nConditions))
	var conds []demo.TriggerConditions
	for _, c := range cConds {
		conds = append(conds, demo.TriggerConditions{
			ChannelA:            demo.TriggerState(c.channelA),
			ChannelB:            demo.TriggerState(c.channelB),
			ChannelC:            demo.TriggerState(c.channelC),
			ChannelD:            demo.TriggerState(c.channelD),
			External:            demo.TriggerState(c.external),
			Aux:                 demo.TriggerState(c.aux),
			PulseWidthQualifier: demo.TriggerState(c.pulseWidthQualifier),
			Digital:             demo.TriggerState(c.digital),
		})
	}
	simSetTriggerChannelConditions(int16(handle), conds)
	return picoOk
}

func cDirToDemoDir(dir C.int32_t) demo.ThresholdDirection {
	switch dir {
	case C.PS2000A_ABOVE:
		return demo.TriggerAbove
	case C.PS2000A_BELOW:
		return demo.TriggerBelow
	case C.PS2000A_RISING:
		return demo.TriggerRising
	case C.PS2000A_FALLING:
		return demo.TriggerFalling
	case C.PS2000A_RISING_OR_FALLING:
		return demo.TriggerRisingOrFalling
	case C.PS2000A_ABOVE_LOWER:
		return demo.TriggerAboveLower
	case C.PS2000A_BELOW_LOWER:
		return demo.TriggerBelowLower
	case C.PS2000A_RISING_LOWER:
		return demo.TriggerRisingLower
	case C.PS2000A_FALLING_LOWER:
		return demo.TriggerFallingLower
	case C.PS2000A_POSITIVE_RUNT:
		return demo.TriggerPositiveRunt
	case C.PS2000A_NEGATIVE_RUNT:
		return demo.TriggerNegativeRunt
	default:
		return demo.TriggerNone
	}
}

//export Gops2000aSetTriggerChannelDirections
func Gops2000aSetTriggerChannelDirections(handle C.int16_t, channelA C.int32_t, channelB C.int32_t, channelC C.int32_t, channelD C.int32_t, ext C.int32_t, aux C.int32_t) C.uint32_t {
	if handle <= 0 {
		return picoInvalidHandle
	}
	simSetTriggerChannelDirections(int16(handle), cDirToDemoDir(channelA), cDirToDemoDir(channelB), cDirToDemoDir(channelC), cDirToDemoDir(channelD))
	return picoOk
}

//export Gops2000aSetTriggerDelay
func Gops2000aSetTriggerDelay(handle C.int16_t, delay C.uint32_t) C.uint32_t {
	if handle <= 0 {
		return picoInvalidHandle
	}
	return picoOk
}

//export Gops2000aSetDataBuffers
func Gops2000aSetDataBuffers(handle C.int16_t, channelOrPort C.int32_t, bufferMax *C.int16_t, bufferMin *C.int16_t, bufferLth C.int32_t, segmentIndex C.uint32_t, mode C.int32_t) C.uint32_t {
	if handle <= 0 {
		return picoInvalidHandle
	}
	if bufferMax == nil && bufferMin == nil {
		return picoNullParameter
	}
	if bufferLth <= 0 {
		return picoInvalidParameter
	}
	if channelOrPort < 0 || (channelOrPort > 3 && channelOrPort != 0x80 && channelOrPort != 0x81 && channelOrPort != 128 && channelOrPort != 129) {
		return picoInvalidChannel
	}
	if segmentIndex != 0 {
		return picoSegmentOutOfRange
	}
	if bufferMax != nil {
		slice := unsafe.Slice((*int16)(unsafe.Pointer(bufferMax)), int(bufferLth))
		simSetDataBufferWithMode(int16(handle), int(channelOrPort), slice, uint32(segmentIndex), int32(mode))
	}
	if bufferMin != nil {
		sliceMin := unsafe.Slice((*int16)(unsafe.Pointer(bufferMin)), int(bufferLth))
		simSetDataBufferMinWithMode(int16(handle), int(channelOrPort), sliceMin, uint32(segmentIndex), int32(mode))
	}
	return picoOk
}

//export Gops2000aIsTriggerOrPulseWidthQualifierEnabled
func Gops2000aIsTriggerOrPulseWidthQualifierEnabled(handle C.int16_t, triggerEnabled *C.int16_t, pulseWidthQualifierEnabled *C.int16_t) C.uint32_t {
	if handle <= 0 {
		return picoInvalidHandle
	}
	if triggerEnabled == nil && pulseWidthQualifierEnabled == nil {
		return picoNullParameter
	}
	if triggerEnabled != nil {
		*triggerEnabled = 0
	}
	if pulseWidthQualifierEnabled != nil {
		*pulseWidthQualifierEnabled = 0
	}
	return picoOk
}

//export Gops2000aGetTimebase
func Gops2000aGetTimebase(handle C.int16_t, timebase C.uint32_t, noSamples C.int32_t, timeIntervalNanoseconds *C.int32_t, oversample C.int16_t, maxSamples *C.int32_t, segmentIndex C.uint32_t) C.uint32_t {
	if handle <= 0 {
		return picoInvalidHandle
	}
	if timeIntervalNanoseconds == nil && maxSamples == nil {
		return picoNullParameter
	}
	if segmentIndex != 0 {
		return picoSegmentOutOfRange
	}
	if timeIntervalNanoseconds != nil {
		*timeIntervalNanoseconds = C.int32_t(simTimebaseToNs(uint32(timebase)))
	}
	if maxSamples != nil {
		*maxSamples = 1000000
	}
	return picoOk
}

//export Gops2000aSetPulseWidthQualifier
func Gops2000aSetPulseWidthQualifier(handle C.int16_t, conditions unsafe.Pointer, nConditions C.int16_t, direction C.int32_t, lower C.uint32_t, upper C.uint32_t, type_ C.int32_t) C.uint32_t {
	if handle <= 0 {
		return picoInvalidHandle
	}
	if conditions == nil && nConditions > 0 {
		return picoNullParameter
	}
	if nConditions < 0 {
		return picoInvalidParameter
	}
	// IN_RANGE (3) or OUT_OF_RANGE (4): lower must be <= upper
	if (type_ == 3 || type_ == 4) && lower > upper {
		return picoPulseWidthQualifier
	}
	if conditions != nil && nConditions > 0 {
		cConds := unsafe.Slice((*C.PS2000A_PWQ_CONDITIONS)(conditions), int(nConditions))
		var conds []demo.PwqConditions
		for _, c := range cConds {
			conds = append(conds, demo.PwqConditions{
				ChannelA: demo.TriggerState(c.channelA),
				ChannelB: demo.TriggerState(c.channelB),
				ChannelC: demo.TriggerState(c.channelC),
				ChannelD: demo.TriggerState(c.channelD),
				External: demo.TriggerState(c.external),
				Aux:      demo.TriggerState(c.aux),
				Digital:  demo.TriggerState(c.digital),
			})
		}
		simSetPulseWidthQualifier(int16(handle), conds, demo.ThresholdDirection(direction), uint32(lower), uint32(upper), demo.PulseWidthType(type_))
	} else {
		simSetPulseWidthQualifier(int16(handle), nil, demo.ThresholdDirection(direction), uint32(lower), uint32(upper), demo.PulseWidthType(type_))
	}
	return picoOk
}

//export Gops2000aGetTriggerTimeOffset64
func Gops2000aGetTriggerTimeOffset64(handle C.int16_t, time *C.int64_t, timeUnits *C.int32_t, segmentIndex C.uint32_t) C.uint32_t {
	if handle <= 0 {
		return picoInvalidHandle
	}
	if time == nil || timeUnits == nil {
		return picoNullParameter
	}
	if segmentIndex != 0 {
		return picoSegmentOutOfRange
	}
	*time = 0
	*timeUnits = 0
	return picoOk
}

//export Gops2000aGetMaxDownSampleRatio
func Gops2000aGetMaxDownSampleRatio(handle C.int16_t, noOfUnaggregatedSamples C.uint32_t, maxDownSampleRatio *C.uint32_t, downSampleRatioMode C.int32_t, segmentIndex C.uint32_t) C.uint32_t {
	if handle <= 0 {
		return picoInvalidHandle
	}
	if maxDownSampleRatio == nil {
		return picoNullParameter
	}
	if segmentIndex != 0 {
		return picoSegmentOutOfRange
	}
	*maxDownSampleRatio = noOfUnaggregatedSamples
	return picoOk
}

//export Gops2000aSetDigitalPort
func Gops2000aSetDigitalPort(handle C.int16_t, port C.int32_t, enabled C.int16_t, logicLevel C.int16_t) C.uint32_t {
	if handle <= 0 {
		return picoInvalidHandle
	}
	if port != 0x80 && port != 0x81 && port != 0 && port != 1 && port != 128 && port != 129 {
		return picoInvalidDigitalPort
	}
	if logicLevel < -32767 || logicLevel > 32767 {
		return picoInvalidParameter
	}
	simSetDigitalPort(int16(handle), int(port), enabled != 0, int16(logicLevel))
	return picoOk
}

//export Gops2000aSetTriggerDigitalPortProperties
func Gops2000aSetTriggerDigitalPortProperties(handle C.int16_t, directions unsafe.Pointer, nDirections C.int16_t) C.uint32_t {
	if handle <= 0 {
		return picoInvalidHandle
	}
	if directions == nil && nDirections > 0 {
		return picoNullParameter
	}
	if nDirections < 0 {
		return picoInvalidParameter
	}
	if directions != nil && nDirections > 0 {
		cDirs := unsafe.Slice((*C.PS2000A_DIGITAL_CHANNEL_DIRECTIONS)(directions), int(nDirections))
		var dirs []demo.DigitalChannelDirections
		for _, d := range cDirs {
			dirs = append(dirs, demo.DigitalChannelDirections{
				Channel:   demo.DigitalChannel(d.channel),
				Direction: demo.DigitalDirection(d.direction),
			})
		}
		simSetTriggerDigitalPortProperties(int16(handle), dirs)
	} else {
		simSetTriggerDigitalPortProperties(int16(handle), nil)
	}
	return picoOk
}

//export Gops2000aSetDigitalAnalogTriggerOperand
func Gops2000aSetDigitalAnalogTriggerOperand(handle C.int16_t, operand C.int32_t) C.uint32_t {
	if handle <= 0 {
		return picoInvalidHandle
	}
	if operand < 0 || operand > 3 {
		return picoInvalidParameter
	}
	simSetDigitalAnalogTriggerOperand(int16(handle), demo.TriggerOperand(operand))
	return picoOk
}

//export Gops2000aSetPulseWidthDigitalPortProperties
func Gops2000aSetPulseWidthDigitalPortProperties(handle C.int16_t, directions unsafe.Pointer, nDirections C.int16_t) C.uint32_t {
	if handle <= 0 {
		return picoInvalidHandle
	}
	if directions == nil && nDirections > 0 {
		return picoNullParameter
	}
	if nDirections < 0 {
		return picoInvalidParameter
	}
	return picoOk
}

//export Gops2000aRunStreaming
func Gops2000aRunStreaming(handle C.int16_t, sampleInterval *C.uint32_t, timeUnits C.int32_t, maxPreTriggerSamples, maxPostTriggerSamples C.uint32_t, autoStop C.int16_t, downSampleRatio C.uint32_t, downSampleRatioMode C.int32_t, overviewBufferSize C.uint32_t) C.uint32_t {
	if handle <= 0 {
		return picoInvalidHandle
	}
	if sampleInterval == nil {
		return picoNullParameter
	}
	if *sampleInterval == 0 {
		return picoInvalidSampleInterval
	}
	if timeUnits < 0 || timeUnits > 5 {
		return picoInvalidParameter
	}
	reqSi := uint32(*sampleInterval)
	si, err := simRunStreaming(int16(handle), reqSi, TimeUnits(timeUnits), uint32(maxPreTriggerSamples), uint32(maxPostTriggerSamples), autoStop != 0, uint32(downSampleRatio), RatioMode(downSampleRatioMode), uint32(overviewBufferSize))
	if err != nil {
		return picoInvalidHandle
	}
	*sampleInterval = C.uint32_t(si)
	return picoOk
}

//export Gops2000aGetStreamingLatestValues
func Gops2000aGetStreamingLatestValues(handle C.int16_t, lpDataReady unsafe.Pointer, pParameter unsafe.Pointer) C.uint32_t {
	if handle <= 0 {
		return picoInvalidHandle
	}
	cb := C.ps2000aStreamingReady(lpDataReady)
	goCb := func(h int16, noOfSamples int32, startIndex uint32, overflow int16, triggerAt uint32, triggered, autoStop int16, param any) {
		C.call_ps2000aStreamingReady(cb, C.int16_t(h), C.int32_t(noOfSamples), C.uint32_t(startIndex), C.int16_t(overflow), C.uint32_t(triggerAt), C.int16_t(triggered), C.int16_t(autoStop), pParameter)
	}
	err := simGetStreamingLatestValues(int16(handle), goCb, pParameter)
	if err != nil {
		return picoInvalidHandle
	}
	return picoOk
}

//export Gops2000aNoOfStreamingValues
func Gops2000aNoOfStreamingValues(handle C.int16_t, noOfValues *C.uint32_t) C.uint32_t {
	if handle <= 0 {
		return picoInvalidHandle
	}
	if noOfValues == nil {
		return picoNullParameter
	}
	n, err := simNoOfStreamingValues(int16(handle))
	if err != nil {
		return picoInvalidHandle
	}
	*noOfValues = C.uint32_t(n)
	return picoOk
}
