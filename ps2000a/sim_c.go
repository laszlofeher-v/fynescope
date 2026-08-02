//go:build sim

package ps2000a

/*
#include <stdlib.h>
#include "/opt/picoscope/include/libps2000/ps2000.h"
#include "/opt/picoscope/include/libps2000a/PicoStatus.h"
#include "/opt/picoscope/include/libps2000a/ps2000aApi.h"

// Forward declarations of Go exports (defined by CGO from the //export directives below).
extern uint32_t Gops2000aEnumerateUnits(int16_t *count, int8_t *serials, int16_t *serialLth);
extern uint32_t Gops2000aOpenUnit(int16_t *handle, int8_t *serial);
extern uint32_t Gops2000aOpenUnitAsync(int16_t *status, int8_t *serial);
extern uint32_t Gops2000aCloseUnit(int16_t handle);
extern uint32_t Gops2000aGetUnitInfo(int16_t handle, int8_t *stringData, int16_t stringLength, int16_t *requiredSize, uint32_t info);
extern uint32_t Gops2000aSetChannel(int16_t handle, int32_t channel, int16_t enabled, int32_t type, int32_t range, float analogOffset);
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

// Static helper to invoke a ps2000aBlockReady function pointer from Go.
static inline void call_ps2000aBlockReady(ps2000aBlockReady fp, int16_t handle, PICO_STATUS status, void *pParameter) {
    if (fp != NULL) {
        fp(handle, status, pParameter);
    }
}
*/
import "C"
import (
	"fynescope/demo"
	"unsafe"
)

//export Gops2000aMemorySegments
func Gops2000aMemorySegments(handle C.int16_t, nSegments C.uint32_t, nMaxSamples *C.int32_t) C.uint32_t {
	*nMaxSamples = 64000000
	return 0
}

//export Gops2000aEnumerateUnits
func Gops2000aEnumerateUnits(count *C.int16_t, serials *C.int8_t, serialLth *C.int16_t) C.uint32_t {
	*count = 1
	*serials = '2'
	return 0
}

//export Gops2000aOpenUnit
func Gops2000aOpenUnit(handle *C.int16_t, serial *C.int8_t) C.uint32_t {
	*handle = 1
	*serial = '2'
	simOpenUnit(int16(*handle))
	return 0
}

//export Gops2000aOpenUnitAsync
func Gops2000aOpenUnitAsync(status *C.int16_t, serial *C.int8_t) C.uint32_t {
	*status = 0
	return 0
}

//export Gops2000aCloseUnit
func Gops2000aCloseUnit(handle C.int16_t) C.uint32_t {
	return 0
}

//export Gops2000aGetUnitInfo
func Gops2000aGetUnitInfo(handle C.int16_t, stringData *C.int8_t, stringLength C.int16_t, requiredSize *C.int16_t, info C.uint32_t) C.uint32_t {
	if requiredSize != nil {
		*requiredSize = 8
	}
	if stringData != nil && stringLength >= 8 {
		str := "2407SIM"
		ptr := (*[1 << 20]C.int8_t)(unsafe.Pointer(stringData))
		for i := 0; i < len(str); i++ {
			ptr[i] = C.int8_t(str[i])
		}
		ptr[len(str)] = 0
	}
	return 0
}

//export Gops2000aSetChannel
func Gops2000aSetChannel(handle C.int16_t, channel C.int32_t, enabled C.int16_t, dc C.int32_t, rangeEnum C.int32_t, analogOffset C.float) C.uint32_t {
	simSetChannel(int16(handle), int(channel), enabled != 0, int(dc), int(rangeEnum), float32(analogOffset))
	return 0
}

//export Gops2000aSetSimpleTrigger
func Gops2000aSetSimpleTrigger(handle C.int16_t, enable C.int16_t, source C.int32_t, threshold C.int16_t, direction C.int32_t, delay C.uint32_t, autoTriggerMs C.int16_t) C.uint32_t {
	simSetSimpleTrigger(int16(handle), enable != 0, int(source), int16(threshold), int(direction), uint32(delay), int16(autoTriggerMs))
	return 0
}

//export Gops2000aSetDataBuffer
func Gops2000aSetDataBuffer(handle C.int16_t, channel C.int32_t, buffer *C.int16_t, bufferLth C.int32_t, segmentIndex C.uint32_t, mode C.int32_t) C.uint32_t {
	slice := unsafe.Slice((*int16)(unsafe.Pointer(buffer)), int(bufferLth))
	simSetDataBuffer(int16(handle), int(channel), slice, uint32(segmentIndex))
	return 0
}

//export Gops2000aRunBlock
func Gops2000aRunBlock(handle C.int16_t, noOfPreTriggerSamples C.int32_t, noOfPostTriggerSamples C.int32_t, timebase C.uint32_t, oversample C.int16_t, timeIndisposedMs *C.int32_t, segmentIndex C.uint32_t, lpReady C.ps2000aBlockReady, pParameter unsafe.Pointer) C.uint32_t {
	simRunBlock(int16(handle), int32(noOfPreTriggerSamples), int32(noOfPostTriggerSamples), uint32(timebase), func(h int16, status int32) {
		C.call_ps2000aBlockReady(lpReady, C.int16_t(h), C.uint32_t(status), pParameter)
	})
	return 0
}

//export Gops2000aStop
func Gops2000aStop(handle C.int16_t) C.uint32_t {
	simStop(int16(handle))
	return 0
}

//export Gops2000aSetSigGenBuiltIn
func Gops2000aSetSigGenBuiltIn(handle C.int16_t, offsetVoltage C.int32_t, pkToPk C.uint32_t, waveType C.int16_t, startFrequency C.float, stopFrequency C.float, increment C.float, dwellTime C.float, sweepType C.int32_t, operation C.int32_t, shots C.uint32_t, sweeps C.uint32_t, triggerType C.int32_t, triggerSource C.int32_t, extInThreshold C.int16_t) C.uint32_t {
	simSetSigGenBuiltIn(int16(handle), int32(offsetVoltage), uint32(pkToPk), int(waveType), float64(startFrequency), float64(stopFrequency), float64(increment), float64(dwellTime), int(sweepType), int(operation))
	return 0
}

//export Gops2000aSetSigGenBuiltInV2
func Gops2000aSetSigGenBuiltInV2(handle C.int16_t, offsetVoltage C.int32_t, pkToPk C.uint32_t, waveType C.int16_t, startFrequency C.double, stopFrequency C.double, increment C.double, dwellTime C.double, sweepType C.int32_t, operation C.int32_t, shots C.uint32_t, sweeps C.uint32_t, triggerType C.int32_t, triggerSource C.int32_t, extInThreshold C.int16_t) C.uint32_t {
	simSetSigGenBuiltIn(int16(handle), int32(offsetVoltage), uint32(pkToPk), int(waveType), float64(startFrequency), float64(stopFrequency), float64(increment), float64(dwellTime), int(sweepType), int(operation))
	return 0
}

//export Gops2000aIsReady
func Gops2000aIsReady(handle C.int16_t, ready *C.int16_t) C.uint32_t {
	r := simIsReady(int16(handle))
	if r {
		*ready = 1
	} else {
		*ready = 0
	}
	return 0
}

//export Gops2000aGetValues
func Gops2000aGetValues(handle C.int16_t, startIndex C.uint32_t, noOfSamples *C.uint32_t, downSampleRatio C.uint32_t, downSampleRatioMode C.int32_t, segmentIndex C.uint32_t, overflow *C.int16_t) C.uint32_t {
	samples, ov := simGetValues(int16(handle), uint32(startIndex), uint32(*noOfSamples))
	*noOfSamples = C.uint32_t(samples)
	*overflow = C.int16_t(ov)
	return 0
}

//export Gops2000aMaximumValue
func Gops2000aMaximumValue(handle C.int16_t, value *C.int16_t) C.uint32_t {
	*value = 32767
	return 0
}

//export Gops2000aMinimumValue
func Gops2000aMinimumValue(handle C.int16_t, value *C.int16_t) C.uint32_t {
	*value = -32767
	return 0
}

//export Gops2000aGetTimebase2
func Gops2000aGetTimebase2(handle C.int16_t, timebase C.uint32_t, noSamples C.int32_t, timeIntervalNanoseconds *C.float, oversample C.int16_t, maxSamples *C.int32_t, segmentIndex C.uint32_t) C.uint32_t {
	if timeIntervalNanoseconds != nil {
		*timeIntervalNanoseconds = C.float(timebase + 1)
	}
	if maxSamples != nil {
		*maxSamples = 1000000
	}
	return 0
}

//export Gops2000aFlashLed
func Gops2000aFlashLed(handle C.int16_t, start C.int16_t) C.uint32_t {
	return 0
}

//export Gops2000aGetChannelInformation
func Gops2000aGetChannelInformation(handle C.int16_t, info C.int32_t, probe C.int32_t, ranges *C.int32_t, length *C.int32_t, channels C.int32_t) C.uint32_t {
	// PS2407B supports all 12 ranges: 10mV(0) through 50V(11)
	numRanges := C.int32_t(12)
	if ranges == nil {
		// Just report how many ranges are available
		*length = numRanges
		return 0
	}
	// Fill the ranges array
	if *length > numRanges {
		*length = numRanges
	}
	slice := unsafe.Slice(ranges, int(*length))
	for i := C.int32_t(0); i < *length; i++ {
		slice[i] = C.int32_t(i) // PS2000A_10MV=0 .. PS2000A_50V=11
	}
	return 0
}

//export Gops2000aSetEts
func Gops2000aSetEts(handle C.int16_t, mode C.int32_t, etsCycles C.int16_t, etsInterleave C.int16_t, sampleTimePicoseconds *C.int32_t) C.uint32_t {
	simSetEts(int16(handle), int(mode), int16(etsCycles), int16(etsInterleave), (*int32)(unsafe.Pointer(sampleTimePicoseconds)))
	return 0
}

//export Gops2000aSetEtsTimeBuffer
func Gops2000aSetEtsTimeBuffer(handle C.int16_t, buffer *C.int64_t, bufferLth C.int32_t) C.uint32_t {
	slice := unsafe.Slice((*int64)(unsafe.Pointer(buffer)), int(bufferLth))
	simSetEtsTimeBuffer(int16(handle), slice)
	return 0
}

//export Gops2000aSetEtsTimeBuffers
func Gops2000aSetEtsTimeBuffers(handle C.int16_t, timeUpper *C.uint32_t, timeLower *C.uint32_t, bufferLth C.int32_t) C.uint32_t {
	sliceUpper := unsafe.Slice((*uint32)(unsafe.Pointer(timeUpper)), int(bufferLth))
	sliceLower := unsafe.Slice((*uint32)(unsafe.Pointer(timeLower)), int(bufferLth))
	simSetEtsTimeBuffers(int16(handle), sliceUpper, sliceLower)
	return 0
}

//export Gops2000aGetMaxEtsValues
func Gops2000aGetMaxEtsValues(handle C.int16_t, etsCycles *C.int16_t, etsInterleave *C.int16_t) C.uint32_t {
	if etsCycles != nil {
		*etsCycles = 250
	}
	if etsInterleave != nil {
		*etsInterleave = 50
	}
	return 0
}

//export Gops2000aSetTriggerChannelProperties
func Gops2000aSetTriggerChannelProperties(handle C.int16_t, channelProperties unsafe.Pointer, nChannelProperties C.int16_t, auxOutputEnable C.int16_t, autoTriggerMilliseconds C.int32_t) C.uint32_t {
	cProps := unsafe.Slice((*C.PS2000A_TRIGGER_CHANNEL_PROPERTIES)(channelProperties), int(nChannelProperties))
	var props []demo.TriggerChannelProperties
	for _, p := range cProps {
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
	return 0
}

//export Gops2000aSetTriggerChannelConditions
func Gops2000aSetTriggerChannelConditions(handle C.int16_t, conditions unsafe.Pointer, nConditions C.int16_t) C.uint32_t {
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
	return 0
}

func cDirToDemoDir(dir C.int32_t) demo.ThresholdDirection {
	switch dir {
	case C.PS2000A_ABOVE: return demo.TriggerAbove
	case C.PS2000A_BELOW: return demo.TriggerBelow
	case C.PS2000A_RISING: return demo.TriggerRising
	case C.PS2000A_FALLING: return demo.TriggerFalling
	case C.PS2000A_RISING_OR_FALLING: return demo.TriggerRisingOrFalling
	case C.PS2000A_ABOVE_LOWER: return demo.TriggerAboveLower
	case C.PS2000A_BELOW_LOWER: return demo.TriggerBelowLower
	case C.PS2000A_RISING_LOWER: return demo.TriggerRisingLower
	case C.PS2000A_FALLING_LOWER: return demo.TriggerFallingLower
	case C.PS2000A_POSITIVE_RUNT: return demo.TriggerPositiveRunt
	case C.PS2000A_NEGATIVE_RUNT: return demo.TriggerNegativeRunt
	default: return demo.TriggerNone
	}
}

//export Gops2000aSetTriggerChannelDirections
func Gops2000aSetTriggerChannelDirections(handle C.int16_t, channelA C.int32_t, channelB C.int32_t, channelC C.int32_t, channelD C.int32_t, ext C.int32_t, aux C.int32_t) C.uint32_t {
	simSetTriggerChannelDirections(int16(handle), cDirToDemoDir(channelA), cDirToDemoDir(channelB), cDirToDemoDir(channelC), cDirToDemoDir(channelD))
	return 0
}

//export Gops2000aSetTriggerDelay
func Gops2000aSetTriggerDelay(handle C.int16_t, delay C.uint32_t) C.uint32_t {
	return 0
}

//export Gops2000aSetDataBuffers
func Gops2000aSetDataBuffers(handle C.int16_t, channelOrPort C.int32_t, bufferMax *C.int16_t, bufferMin *C.int16_t, bufferLth C.int32_t, segmentIndex C.uint32_t, mode C.int32_t) C.uint32_t {
	if bufferMax != nil {
		slice := unsafe.Slice((*int16)(unsafe.Pointer(bufferMax)), int(bufferLth))
		simSetDataBuffer(int16(handle), int(channelOrPort), slice, uint32(segmentIndex))
	}
	return 0
}

//export Gops2000aIsTriggerOrPulseWidthQualifierEnabled
func Gops2000aIsTriggerOrPulseWidthQualifierEnabled(handle C.int16_t, triggerEnabled *C.int16_t, pulseWidthQualifierEnabled *C.int16_t) C.uint32_t {
	if triggerEnabled != nil {
		*triggerEnabled = 0
	}
	if pulseWidthQualifierEnabled != nil {
		*pulseWidthQualifierEnabled = 0
	}
	return 0
}

//export Gops2000aGetTimebase
func Gops2000aGetTimebase(handle C.int16_t, timebase C.uint32_t, noSamples C.int32_t, timeIntervalNanoseconds *C.int32_t, oversample C.int16_t, maxSamples *C.int32_t, segmentIndex C.uint32_t) C.uint32_t {
	if timeIntervalNanoseconds != nil {
		*timeIntervalNanoseconds = C.int32_t(timebase + 1)
	}
	if maxSamples != nil {
		*maxSamples = 1000000
	}
	return 0
}

//export Gops2000aSetPulseWidthQualifier
func Gops2000aSetPulseWidthQualifier(handle C.int16_t, conditions unsafe.Pointer, nConditions C.int16_t, direction C.int32_t, lower C.uint32_t, upper C.uint32_t, type_ C.int32_t) C.uint32_t {
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
	return 0
}

//export Gops2000aGetTriggerTimeOffset64
func Gops2000aGetTriggerTimeOffset64(handle C.int16_t, time *C.int64_t, timeUnits *C.int32_t, segmentIndex C.uint32_t) C.uint32_t {
	return 0
}
