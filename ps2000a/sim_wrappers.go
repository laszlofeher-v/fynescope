//go:build sim

package ps2000a

/*
#include <stdlib.h>
#include "/opt/picoscope/include/libps2000/ps2000.h"
#include "/opt/picoscope/include/libps2000a/PicoStatus.h"
#include "/opt/picoscope/include/libps2000a/ps2000aApi.h"

// Forward declarations of Go exports
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
extern uint32_t Gops2000aMemorySegments(int16_t handle, uint32_t nSegments, int32_t * nMaxSamples);
extern uint32_t Gops2000aGetTriggerTimeOffset64(int16_t handle, int64_t *time, int32_t *timeUnits, uint32_t segmentIndex);

// C implementations - each ps2000a API function delegates to its Go counterpart.
PICO_STATUS ps2000aMemorySegments(int16_t handle, uint32_t nSegments, int32_t * nMaxSamples){
    return Gops2000aMemorySegments(handle, nSegments, nMaxSamples);
}
PICO_STATUS ps2000aEnumerateUnits(int16_t *count, int8_t *serials, int16_t *serialLth) {
    return Gops2000aEnumerateUnits(count, serials, serialLth);
}

PICO_STATUS ps2000aOpenUnit(int16_t *handle, int8_t *serial) {
    return Gops2000aOpenUnit(handle, serial);
}

PICO_STATUS ps2000aOpenUnitAsync(int16_t *status, int8_t *serial) {
    return Gops2000aOpenUnitAsync(status, serial);
}

PICO_STATUS ps2000aCloseUnit(int16_t handle) {
    return Gops2000aCloseUnit(handle);
}

PICO_STATUS ps2000aGetUnitInfo(int16_t handle, int8_t *stringData, int16_t stringLength, int16_t *requiredSize, PICO_INFO info) {
    return Gops2000aGetUnitInfo(handle, stringData, stringLength, requiredSize, info);
}

PICO_STATUS ps2000aSetChannel(int16_t handle, PS2000A_CHANNEL channel, int16_t enabled, PS2000A_COUPLING type, PS2000A_RANGE range, float analogOffset) {
    return Gops2000aSetChannel(handle, channel, enabled, type, range, analogOffset);
}

PICO_STATUS ps2000aSetSimpleTrigger(int16_t handle, int16_t enable, PS2000A_CHANNEL source, int16_t threshold, PS2000A_THRESHOLD_DIRECTION direction, uint32_t delay, int16_t autoTrigger_ms) {
    return Gops2000aSetSimpleTrigger(handle, enable, source, threshold, direction, delay, autoTrigger_ms);
}

PICO_STATUS ps2000aSetDataBuffer(int16_t handle, int32_t channel, int16_t *buffer, int32_t bufferLth, uint32_t segmentIndex, PS2000A_RATIO_MODE mode) {
    return Gops2000aSetDataBuffer(handle, channel, buffer, bufferLth, segmentIndex, mode);
}

PICO_STATUS ps2000aRunBlock(int16_t handle, int32_t noOfPreTriggerSamples, int32_t noOfPostTriggerSamples, uint32_t timebase, int16_t oversample, int32_t *timeIndisposedMs, uint32_t segmentIndex, ps2000aBlockReady lpReady, void *pParameter) {
    return Gops2000aRunBlock(handle, noOfPreTriggerSamples, noOfPostTriggerSamples, timebase, oversample, timeIndisposedMs, segmentIndex, lpReady, pParameter);
}

PICO_STATUS ps2000aStop(int16_t handle) {
    return Gops2000aStop(handle);
}

PICO_STATUS ps2000aSetSigGenBuiltIn(int16_t handle, int32_t offsetVoltage, uint32_t pkToPk, int16_t waveType, float startFrequency, float stopFrequency, float increment, float dwellTime, PS2000A_SWEEP_TYPE sweepType, PS2000A_EXTRA_OPERATIONS operation, uint32_t shots, uint32_t sweeps, PS2000A_SIGGEN_TRIG_TYPE triggerType, PS2000A_SIGGEN_TRIG_SOURCE triggerSource, int16_t extInThreshold) {
    return Gops2000aSetSigGenBuiltIn(handle, offsetVoltage, pkToPk, waveType, startFrequency, stopFrequency, increment, dwellTime, sweepType, operation, shots, sweeps, triggerType, triggerSource, extInThreshold);
}

PICO_STATUS ps2000aSetSigGenBuiltInV2(int16_t handle, int32_t offsetVoltage, uint32_t pkToPk, int16_t waveType, double startFrequency, double stopFrequency, double increment, double dwellTime, PS2000A_SWEEP_TYPE sweepType, PS2000A_EXTRA_OPERATIONS operation, uint32_t shots, uint32_t sweeps, PS2000A_SIGGEN_TRIG_TYPE triggerType, PS2000A_SIGGEN_TRIG_SOURCE triggerSource, int16_t extInThreshold) {
    return Gops2000aSetSigGenBuiltInV2(handle, offsetVoltage, pkToPk, waveType, startFrequency, stopFrequency, increment, dwellTime, sweepType, operation, shots, sweeps, triggerType, triggerSource, extInThreshold);
}

PICO_STATUS ps2000aIsReady(int16_t handle, int16_t *ready) {
    return Gops2000aIsReady(handle, ready);
}

PICO_STATUS ps2000aGetValues(int16_t handle, uint32_t startIndex, uint32_t *noOfSamples, uint32_t downSampleRatio, PS2000A_RATIO_MODE downSampleRatioMode, uint32_t segmentIndex, int16_t *overflow) {
    return Gops2000aGetValues(handle, startIndex, noOfSamples, downSampleRatio, downSampleRatioMode, segmentIndex, overflow);
}

PICO_STATUS ps2000aMaximumValue(int16_t handle, int16_t *value) {
    return Gops2000aMaximumValue(handle, value);
}

PICO_STATUS ps2000aMinimumValue(int16_t handle, int16_t *value) {
    return Gops2000aMinimumValue(handle, value);
}

PICO_STATUS ps2000aGetTimebase2(int16_t handle, uint32_t timebase, int32_t noSamples, float *timeIntervalNanoseconds, int16_t oversample, int32_t *maxSamples, uint32_t segmentIndex) {
    return Gops2000aGetTimebase2(handle, timebase, noSamples, timeIntervalNanoseconds, oversample, maxSamples, segmentIndex);
}

PICO_STATUS ps2000aFlashLed(int16_t handle, int16_t start) {
    return Gops2000aFlashLed(handle, start);
}

PICO_STATUS ps2000aGetChannelInformation(int16_t handle, PS2000A_CHANNEL_INFO info, int32_t probe, int32_t *ranges, int32_t *length, int32_t channels) {
    return Gops2000aGetChannelInformation(handle, (int32_t)info, probe, ranges, length, channels);
}

PICO_STATUS ps2000aSetEts(int16_t handle, PS2000A_ETS_MODE mode, int16_t etsCycles, int16_t etsInterleave, int32_t *sampleTimePicoseconds) {
    return Gops2000aSetEts(handle, (int32_t)mode, etsCycles, etsInterleave, sampleTimePicoseconds);
}

PICO_STATUS ps2000aSetEtsTimeBuffer(int16_t handle, int64_t *buffer, int32_t bufferLth) {
    return Gops2000aSetEtsTimeBuffer(handle, buffer, bufferLth);
}

PICO_STATUS ps2000aSetEtsTimeBuffers(int16_t handle, uint32_t *timeUpper, uint32_t *timeLower, int32_t bufferLth) {
    return Gops2000aSetEtsTimeBuffers(handle, timeUpper, timeLower, bufferLth);
}

PICO_STATUS ps2000aGetMaxEtsValues(int16_t handle, int16_t *etsCycles, int16_t *etsInterleave) {
    return Gops2000aGetMaxEtsValues(handle, etsCycles, etsInterleave);
}

PICO_STATUS ps2000aSetTriggerChannelProperties(int16_t handle, PS2000A_TRIGGER_CHANNEL_PROPERTIES *channelProperties, int16_t nChannelProperties, int16_t auxOutputEnable, int32_t autoTriggerMilliseconds) {
    return Gops2000aSetTriggerChannelProperties(handle, channelProperties, nChannelProperties, auxOutputEnable, autoTriggerMilliseconds);
}

PICO_STATUS ps2000aSetTriggerChannelConditions(int16_t handle, PS2000A_TRIGGER_CONDITIONS *conditions, int16_t nConditions) {
    return Gops2000aSetTriggerChannelConditions(handle, conditions, nConditions);
}

PICO_STATUS ps2000aSetTriggerChannelDirections(int16_t handle, PS2000A_THRESHOLD_DIRECTION channelA, PS2000A_THRESHOLD_DIRECTION channelB, PS2000A_THRESHOLD_DIRECTION channelC, PS2000A_THRESHOLD_DIRECTION channelD, PS2000A_THRESHOLD_DIRECTION ext, PS2000A_THRESHOLD_DIRECTION aux) {
    return Gops2000aSetTriggerChannelDirections(handle, (int32_t)channelA, (int32_t)channelB, (int32_t)channelC, (int32_t)channelD, (int32_t)ext, (int32_t)aux);
}

PICO_STATUS ps2000aSetTriggerDelay(int16_t handle, uint32_t delay) {
    return Gops2000aSetTriggerDelay(handle, delay);
}

PICO_STATUS ps2000aSetDataBuffers(int16_t handle, int32_t channelOrPort, int16_t *bufferMax, int16_t *bufferMin, int32_t bufferLth, uint32_t segmentIndex, PS2000A_RATIO_MODE mode) {
    return Gops2000aSetDataBuffers(handle, channelOrPort, bufferMax, bufferMin, bufferLth, segmentIndex, (int32_t)mode);
}

PICO_STATUS ps2000aIsTriggerOrPulseWidthQualifierEnabled(int16_t handle, int16_t *triggerEnabled, int16_t *pulseWidthQualifierEnabled) {
    return Gops2000aIsTriggerOrPulseWidthQualifierEnabled(handle, triggerEnabled, pulseWidthQualifierEnabled);
}

PICO_STATUS ps2000aGetTimebase(int16_t handle, uint32_t timebase, int32_t noSamples, int32_t *timeIntervalNanoseconds, int16_t oversample, int32_t *maxSamples, uint32_t segmentIndex) {
    return Gops2000aGetTimebase(handle, timebase, noSamples, timeIntervalNanoseconds, oversample, maxSamples, segmentIndex);
}

PICO_STATUS ps2000aSetPulseWidthQualifier(int16_t handle, PS2000A_PWQ_CONDITIONS *conditions, int16_t nConditions, PS2000A_THRESHOLD_DIRECTION direction, uint32_t lower, uint32_t upper, PS2000A_PULSE_WIDTH_TYPE type) {
    return Gops2000aSetPulseWidthQualifier(handle, conditions, nConditions, (int32_t)direction, lower, upper, (int32_t)type);
}

PICO_STATUS ps2000aGetTriggerTimeOffset64(int16_t handle, int64_t *time, PS2000A_TIME_UNITS *timeUnits, uint32_t segmentIndex) {
    return Gops2000aGetTriggerTimeOffset64(handle, time, (int32_t *)timeUnits, segmentIndex);
}
*/
import "C"
