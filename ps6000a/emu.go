//go:build emu && ps6000a && !demo

package ps6000a

/*
#cgo CFLAGS: -I/opt/picoscope/include/libps6000a
#cgo LDFLAGS: -L/opt/picoscope/lib/ -lps2000a

#include <stdint.h>
#include <string.h>
#include "/opt/picoscope/include/libps6000a/ps6000aApi.h"

// ps2000a forward declarations
uint32_t ps2000aEnumerateUnits(int16_t *count, int8_t *serials, int16_t *serialLth);
uint32_t ps2000aOpenUnit(int16_t *handle, int8_t *serial);
uint32_t ps2000aOpenUnitAsync(int16_t *status, int8_t *serial);
uint32_t ps2000aOpenUnitProgress(int16_t *handle, int16_t *progressPercent, int16_t *complete);
uint32_t ps2000aCloseUnit(int16_t handle);
uint32_t ps2000aGetUnitInfo(int16_t handle, int8_t *string, int16_t stringLength, int16_t *requiredSize, uint32_t info);
uint32_t ps2000aFlashLed(int16_t handle, int16_t start);
uint32_t ps2000aGetValuesAsync(int16_t handle, uint32_t startIndex, uint32_t noOfSamples, uint32_t downSampleRatio, uint32_t downSampleRatioMode, uint32_t segmentIndex, void *lpDataReady, void *pParameter);
uint32_t ps2000aGetValues(int16_t handle, uint32_t startIndex, uint32_t *noOfSamples, uint32_t downSampleRatio, uint32_t downSampleRatioMode, uint32_t segmentIndex, int16_t *overflow);
uint32_t ps2000aGetValuesBulk(int16_t handle, uint32_t *noOfSamples, uint32_t fromSegmentIndex, uint32_t toSegmentIndex, uint32_t downSampleRatio, uint32_t downSampleRatioMode, int16_t *overflow);
uint32_t ps2000aGetValuesOverlapped(int16_t handle, uint32_t startIndex, uint32_t *noOfSamples, uint32_t downSampleRatio, uint32_t downSampleRatioMode, uint32_t segmentIndex, int16_t *overflow);
uint32_t ps2000aGetValuesOverlappedBulk(int16_t handle, uint32_t startIndex, uint32_t *noOfSamples, uint32_t downSampleRatio, uint32_t downSampleRatioMode, uint32_t fromSegmentIndex, uint32_t toSegmentIndex, int16_t *overflow);
uint32_t ps2000aGetMaxDownSampleRatio(int16_t handle, uint32_t noOfUnaggregatedSamples, uint32_t *maxDownSampleRatio, uint32_t downSampleRatioMode, uint32_t segmentIndex);
uint32_t ps2000aGetMaxSegments(int16_t handle, uint32_t *maxSegments);
uint32_t ps2000aGetNoOfCaptures(int16_t handle, uint32_t *nCaptures);
uint32_t ps2000aGetNoOfProcessedCaptures(int16_t handle, uint32_t *nProcessedCaptures);
uint32_t ps2000aGetStreamingLatestValues(int16_t handle, void *lpPs2000aReady, void *pParameter);
uint32_t ps2000aGetTimebase(int16_t handle, uint32_t timebase, int32_t noSamples, int32_t *timeIntervalNanoseconds, int16_t oversample, int32_t *maxSamples, uint32_t segmentIndex);
uint32_t ps2000aGetTimebase2(int16_t handle, uint32_t timebase, int32_t noSamples, float *timeIntervalNanoseconds, int16_t oversample, int32_t *maxSamples, uint32_t segmentIndex);
uint32_t ps2000aSetChannel(int16_t handle, uint32_t channel, int16_t enabled, uint32_t type, uint32_t range, float analogOffset);
uint32_t ps2000aMaximumValue(int16_t handle, int16_t *value);
uint32_t ps2000aMinimumValue(int16_t handle, int16_t *value);
uint32_t ps2000aSetSimpleTrigger(int16_t handle, int16_t enable, uint32_t source, int16_t threshold, uint32_t direction, uint32_t delay, int16_t autoTrigger_ms);
uint32_t ps2000aSetDataBuffer(int16_t handle, uint32_t channel, void *buffer, int32_t bufferLth, uint32_t segmentIndex, uint32_t mode);
uint32_t ps2000aSetDataBuffers(int16_t handle, uint32_t channel, void *bufferMax, void *bufferMin, int32_t bufferLth, uint32_t segmentIndex, uint32_t mode);
uint32_t ps2000aRunStreaming(int16_t handle, uint32_t *sampleInterval, uint32_t sampleIntervalTimeUnits, uint32_t maxPreTriggerSamples, uint32_t maxPostTriggerSamples, int16_t autoStop, uint32_t downSampleRatio, uint32_t downSampleRatioMode, uint32_t overviewBufferSize);
uint32_t ps2000aRunBlock(int16_t handle, int32_t noOfPreTriggerSamples, int32_t noOfPostTriggerSamples, uint32_t timebase, int16_t oversample, int32_t *timeIndisposedMs, uint32_t segmentIndex, void *lpReady, void *pParameter);
uint32_t ps2000aSetTriggerChannelProperties(int16_t handle, void *channelProperties, int16_t nChannelProperties, int16_t auxOutputEnable, int32_t autoTriggerMilliseconds);
uint32_t ps2000aSetTriggerChannelConditions(int16_t handle, void *conditions, int16_t nConditions);
uint32_t ps2000aSetTriggerDelay(int16_t handle, uint32_t delay);
uint32_t ps2000aStop(int16_t handle);
uint32_t ps2000aSetNoOfCaptures(int16_t handle, uint32_t nCaptures);
uint32_t ps2000aGetTriggerTimeOffset(int16_t handle, uint32_t *timeUpper, uint32_t *timeLower, void *timeUnits, uint32_t segmentIndex);
uint32_t ps2000aGetTriggerTimeOffset64(int16_t handle, int64_t *time, void *timeUnits, uint32_t segmentIndex);
uint32_t ps2000aGetValuesTriggerTimeOffsetBulk(int16_t handle, uint32_t *timesUpper, uint32_t *timesLower, void *timeUnits, uint32_t fromSegmentIndex, uint32_t toSegmentIndex);
uint32_t ps2000aGetValuesTriggerTimeOffsetBulk64(int16_t handle, int64_t *times, void *timeUnits, uint32_t fromSegmentIndex, uint32_t toSegmentIndex);
uint32_t ps2000aIsReady(int16_t handle, int16_t *ready);
uint32_t ps2000aPingUnit(int16_t handle);
uint32_t ps2000aNoOfStreamingValues(int16_t handle, uint32_t *noOfValues);
uint32_t ps2000aMemorySegments(int16_t handle, uint32_t nSegments, int32_t *nMaxSamples);

PICO_STATUS ps6000aEnumerateUnits(int16_t *count, int8_t *serials, int16_t *serialLth) {
    uint32_t status = ps2000aEnumerateUnits(count, serials, serialLth);
    if (status == 0 && serials != 0 && *serialLth > 0) {
        strcpy((char *)serials, "66AEMU");
    }
    return status;
}
PICO_STATUS ps6000aOpenUnit(int16_t *handle, int8_t *serial, PICO_DEVICE_RESOLUTION resolution) { return ps2000aOpenUnit(handle, serial); }
PICO_STATUS ps6000aOpenUnitAsync(int16_t *status, int8_t *serial, PICO_DEVICE_RESOLUTION resolution) { return ps2000aOpenUnitAsync(status, serial); }
PICO_STATUS ps6000aOpenUnitProgress(int16_t *handle, int16_t *progressPercent, int16_t *complete) { return ps2000aOpenUnitProgress(handle, progressPercent, complete); }
PICO_STATUS ps6000aCloseUnit(int16_t handle) { return ps2000aCloseUnit(handle); }
PICO_STATUS ps6000aGetUnitInfo(int16_t handle, int8_t *string, int16_t stringLength, int16_t *requiredSize, PICO_INFO info) {
    uint32_t status = ps2000aGetUnitInfo(handle, string, stringLength, requiredSize, (uint32_t)info);
    if (status == 0 && string != 0 && stringLength > 0) {
        if (info == 3) strcpy((char *)string, "66AEMU");
    }
    return status;
}
PICO_STATUS ps6000aFlashLed(int16_t handle, int16_t start) { return ps2000aFlashLed(handle, start); }
PICO_STATUS ps6000aGetValuesAsync(int16_t handle, uint64_t startIndex, uint64_t noOfSamples, uint64_t downSampleRatio, PICO_RATIO_MODE downSampleRatioMode, uint64_t segmentIndex, PICO_POINTER lpDataReady, PICO_POINTER pParameter) { return ps2000aGetValuesAsync(handle, (uint32_t)startIndex, (uint32_t)noOfSamples, (uint32_t)downSampleRatio, (uint32_t)downSampleRatioMode, (uint32_t)segmentIndex, lpDataReady, pParameter); }
PICO_STATUS ps6000aGetValuesBulkAsync(int16_t handle, uint64_t startIndex, uint64_t noOfSamples, uint64_t fromSegmentIndex, uint64_t toSegmentIndex, uint64_t downSampleRatio, PICO_RATIO_MODE downSampleRatioMode, PICO_POINTER lpDataReady, PICO_POINTER pParameter) { return 0; } // not direct map
PICO_STATUS ps6000aGetValues(int16_t handle, uint64_t startIndex, uint64_t *noOfSamples, uint64_t downSampleRatio, PICO_RATIO_MODE downSampleRatioMode, uint64_t segmentIndex, int16_t *overflow) { 
    uint32_t samples32 = (uint32_t)*noOfSamples;
    uint32_t status = ps2000aGetValues(handle, (uint32_t)startIndex, &samples32, (uint32_t)downSampleRatio, (uint32_t)downSampleRatioMode, (uint32_t)segmentIndex, overflow);
    *noOfSamples = samples32;
    return status;
}
PICO_STATUS ps6000aGetValuesBulk(int16_t handle, uint64_t startIndex, uint64_t *noOfSamples, uint64_t fromSegmentIndex, uint64_t toSegmentIndex, uint64_t downSampleRatio, PICO_RATIO_MODE downSampleRatioMode, int16_t *overflow) { 
    uint32_t samples32 = (uint32_t)*noOfSamples;
    uint32_t status = ps2000aGetValuesBulk(handle, &samples32, (uint32_t)fromSegmentIndex, (uint32_t)toSegmentIndex, (uint32_t)downSampleRatio, (uint32_t)downSampleRatioMode, overflow);
    *noOfSamples = samples32;
    return status;
}
PICO_STATUS ps6000aGetValuesOverlapped(int16_t handle, uint64_t startIndex, uint64_t *noOfSamples, uint64_t downSampleRatio, PICO_RATIO_MODE downSampleRatioMode, uint64_t fromSegmentIndex, uint64_t toSegmentIndex, int16_t *overflow) {
    uint32_t samples32 = (uint32_t)*noOfSamples;
    uint32_t status = ps2000aGetValuesOverlapped(handle, (uint32_t)startIndex, &samples32, (uint32_t)downSampleRatio, (uint32_t)downSampleRatioMode, (uint32_t)fromSegmentIndex, overflow);
    *noOfSamples = samples32;
    return status;
}
PICO_STATUS ps6000aGetNoOfCaptures(int16_t handle, uint64_t *nCaptures) { 
    uint32_t cap32;
    uint32_t status = ps2000aGetNoOfCaptures(handle, &cap32);
    *nCaptures = cap32;
    return status;
}
PICO_STATUS ps6000aGetNoOfProcessedCaptures(int16_t handle, uint64_t *nCaptures) { 
    uint32_t cap32;
    uint32_t status = ps2000aGetNoOfProcessedCaptures(handle, &cap32);
    *nCaptures = cap32;
    return status;
}
PICO_STATUS ps6000aGetTimebase(int16_t handle, uint32_t timebase, uint64_t noSamples, double *timeIntervalNanoseconds, uint64_t *maxSamples, uint64_t segmentIndex) {
    float intervalFloat;
    int32_t maxSamples32;
    uint32_t status = ps2000aGetTimebase2(handle, timebase, (int32_t)noSamples, &intervalFloat, 0, &maxSamples32, (uint32_t)segmentIndex);
    if (timeIntervalNanoseconds) *timeIntervalNanoseconds = intervalFloat;
    if (maxSamples) *maxSamples = maxSamples32;
    return status;
}
PICO_STATUS ps6000aSetChannelOn(int16_t handle, PICO_CHANNEL channel, PICO_COUPLING coupling, PICO_CONNECT_PROBE_RANGE range, double analogueOffset, PICO_BANDWIDTH_LIMITER bandwidth) { return ps2000aSetChannel(handle, (uint32_t)channel, 1, (uint32_t)coupling, (uint32_t)range, (float)analogueOffset); }
PICO_STATUS ps6000aSetChannelOff(int16_t handle, PICO_CHANNEL channel) { return ps2000aSetChannel(handle, (uint32_t)channel, 0, 0, 0, 0.0f); }
PICO_STATUS ps6000aGetAdcLimits(int16_t handle, PICO_DEVICE_RESOLUTION resolution, int16_t *minValue, int16_t *maxValue) { 
    ps2000aMinimumValue(handle, minValue);
    return ps2000aMaximumValue(handle, maxValue);
}
PICO_STATUS ps6000aSetSimpleTrigger(int16_t handle, int16_t enable, PICO_CHANNEL source, int16_t threshold, PICO_THRESHOLD_DIRECTION direction, uint64_t delay, uint32_t autoTrigger_us) { return ps2000aSetSimpleTrigger(handle, enable, (uint32_t)source, threshold, (uint32_t)direction, (uint32_t)delay, autoTrigger_us/1000); }
PICO_STATUS ps6000aSetDataBuffer(int16_t handle, PICO_CHANNEL channel, void *buffer, int32_t bufferLth, PICO_DATA_TYPE dataType, uint64_t waveform, PICO_RATIO_MODE downSampleRatioMode, PICO_ACTION action) { return ps2000aSetDataBuffer(handle, (uint32_t)channel, buffer, bufferLth, (uint32_t)waveform, (uint32_t)downSampleRatioMode); }
PICO_STATUS ps6000aSetDataBuffers(int16_t handle, PICO_CHANNEL channel, void *bufferMax, void *bufferMin, int32_t bufferLth, PICO_DATA_TYPE dataType, uint64_t waveform, PICO_RATIO_MODE downSampleRatioMode, PICO_ACTION action) { return ps2000aSetDataBuffers(handle, (uint32_t)channel, bufferMax, bufferMin, bufferLth, (uint32_t)waveform, (uint32_t)downSampleRatioMode); }
PICO_STATUS ps6000aRunStreaming(int16_t handle, double *sampleInterval, PICO_TIME_UNITS sampleIntervalTimeUnits, uint64_t maxPreTriggerSamples, uint64_t maxPostPreTriggerSamples, int16_t autoStop, uint64_t downSampleRatio, PICO_RATIO_MODE downSampleRatioMode) {
    uint32_t interval32 = (uint32_t)*sampleInterval;
    uint32_t status = ps2000aRunStreaming(handle, &interval32, (uint32_t)sampleIntervalTimeUnits, (uint32_t)maxPreTriggerSamples, (uint32_t)maxPostPreTriggerSamples, autoStop, (uint32_t)downSampleRatio, (uint32_t)downSampleRatioMode, 0);
    *sampleInterval = interval32;
    return status;
}
PICO_STATUS ps6000aRunBlock(int16_t handle, uint64_t noOfPreTriggerSamples, uint64_t noOfPostTriggerSamples, uint32_t timebase, double *timeIndisposedMs, uint64_t segmentIndex, ps6000aBlockReady lpReady, PICO_POINTER pParameter) {
    int32_t indisposed;
    uint32_t status = ps2000aRunBlock(handle, (int32_t)noOfPreTriggerSamples, (int32_t)noOfPostTriggerSamples, timebase, 0, &indisposed, (uint32_t)segmentIndex, (void*)lpReady, pParameter);
    if (timeIndisposedMs) *timeIndisposedMs = indisposed;
    return status;
}
PICO_STATUS ps6000aSetTriggerChannelProperties(int16_t handle, PICO_TRIGGER_CHANNEL_PROPERTIES *channelProperties, int16_t nChannelProperties, int16_t auxOutputEnable, uint32_t autoTriggerMicroSeconds) { return ps2000aSetTriggerChannelProperties(handle, (void*)channelProperties, nChannelProperties, auxOutputEnable, autoTriggerMicroSeconds/1000); }
PICO_STATUS ps6000aSetTriggerChannelConditions(int16_t handle, PICO_CONDITION *conditions, int16_t nConditions, PICO_ACTION action) { return ps2000aSetTriggerChannelConditions(handle, (void*)conditions, nConditions); }
PICO_STATUS ps6000aSetTriggerChannelDirections(int16_t handle, PICO_DIRECTION *directions, int16_t nDirections) { return 0; } // Too different
PICO_STATUS ps6000aSetTriggerDelay(int16_t handle, uint64_t delay) { return ps2000aSetTriggerDelay(handle, (uint32_t)delay); }
PICO_STATUS ps6000aSetTriggerDigitalPortProperties(int16_t handle, PICO_CHANNEL port, PICO_DIGITAL_CHANNEL_DIRECTIONS *directions, int16_t nDirections) { return 0; }
PICO_STATUS ps6000aStop(int16_t handle) { return ps2000aStop(handle); }
PICO_STATUS ps6000aSetNoOfCaptures(int16_t handle, uint64_t nCaptures) { return ps2000aSetNoOfCaptures(handle, (uint32_t)nCaptures); }
PICO_STATUS ps6000aGetTriggerTimeOffset(int16_t handle, int64_t *time, PICO_TIME_UNITS *timeUnits, uint64_t segmentIndex) { return ps2000aGetTriggerTimeOffset64(handle, time, (void*)timeUnits, (uint32_t)segmentIndex); }
PICO_STATUS ps6000aGetValuesTriggerTimeOffsetBulk(int16_t handle, int64_t *times, PICO_TIME_UNITS *timeUnits, uint64_t fromSegmentIndex, uint64_t toSegmentIndex) { return ps2000aGetValuesTriggerTimeOffsetBulk64(handle, times, (void*)timeUnits, (uint32_t)fromSegmentIndex, (uint32_t)toSegmentIndex); }
PICO_STATUS ps6000aIsReady(int16_t handle, int16_t *ready) { return ps2000aIsReady(handle, ready); }
PICO_STATUS ps6000aMemorySegments(int16_t handle, uint64_t nSegments, uint64_t *nMaxSamples) { 
    int32_t maxSamples32;
    uint32_t status = ps2000aMemorySegments(handle, (uint32_t)nSegments, &maxSamples32);
    if (nMaxSamples) *nMaxSamples = maxSamples32;
    return status;
}
PICO_STATUS ps6000aNoOfStreamingValues(int16_t handle, uint64_t *noOfValues) { 
    uint32_t vals32;
    uint32_t status = ps2000aNoOfStreamingValues(handle, &vals32);
    if (noOfValues) *noOfValues = vals32;
    return status;
}
PICO_STATUS ps6000aPingUnit(int16_t handle) { return ps2000aPingUnit(handle); }
PICO_STATUS ps6000aQueryOutputEdgeDetect(int16_t handle, int16_t *state) { return 0; }
PICO_STATUS ps6000aSetOutputEdgeDetect(int16_t handle, int16_t state) { return 0; }
PICO_STATUS ps6000aSetPulseWidthDigitalPortProperties(int16_t handle, PICO_CHANNEL port, PICO_DIGITAL_CHANNEL_DIRECTIONS *directions, int16_t nDirections) { return 0; }
PICO_STATUS ps6000aGetAccessoryInfo(int16_t handle, PICO_CHANNEL channel, int8_t *string, int16_t stringLength, int16_t *requiredSize, PICO_INFO info) { return 0; }
PICO_STATUS ps6000aSetAuxIoMode(int16_t handle, PICO_AUXIO_MODE auxIoMode) { return 0; }
PICO_STATUS ps6000aSetTriggerHoldoffCounterBySamples(int16_t handle, uint64_t samples) { return 0; }
PICO_STATUS ps6000aMemorySegmentsBySamples(int16_t handle, uint64_t nSamples, uint64_t *nMaxSegments) { return 0; }
PICO_STATUS ps6000aGetMaximumAvailableMemory(int16_t handle, uint64_t *nMaxSamples, PICO_DEVICE_RESOLUTION resolution) { return 0; }
PICO_STATUS ps6000aQueryMaxSegmentsBySamples(int16_t handle, uint64_t nSamples, uint32_t nChannelEnabled, uint64_t *nMaxSegments, PICO_DEVICE_RESOLUTION resolution) { return 0; }
PICO_STATUS ps6000aGetScopeState(int16_t handle, PICO_SCOPE_STATE *state) { return 0; }
PICO_STATUS ps6000aSetDigitalPortOn(int16_t handle, PICO_CHANNEL port, int16_t *logicThresholdLevel, int16_t logicThresholdLevelLength, PICO_DIGITAL_PORT_HYSTERESIS hysteresis) { return 0; }
PICO_STATUS ps6000aSetDigitalPortOff(int16_t handle, PICO_CHANNEL port) { return 0; }
PICO_STATUS ps6000aSigGenWaveform(int16_t handle, PICO_WAVE_TYPE waveType, int16_t *buffer, uint64_t bufferLength) { return 0; }
PICO_STATUS ps6000aSigGenRange(int16_t handle, double pkToPk, double offsetVoltage) { return 0; }
*/
import "C"
