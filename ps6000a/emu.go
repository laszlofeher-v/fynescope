//go:build emu && ps6000a && !demo

package ps6000a

/*
#cgo CFLAGS: -I/opt/picoscope/include/libps6000a

#include <stdint.h>
#include <string.h>
#include "/opt/picoscope/include/libps6000a/ps6000aApi.h"

PICO_STATUS ps6000aEnumerateUnits(int16_t *count, int8_t *serials, int16_t *serialLth) { return 0; }
PICO_STATUS ps6000aOpenUnit(int16_t *handle, int8_t *serial, PICO_DEVICE_RESOLUTION resolution) { return 0; }
PICO_STATUS ps6000aOpenUnitAsync(int16_t *status, int8_t *serial, PICO_DEVICE_RESOLUTION resolution) { return 0; }
PICO_STATUS ps6000aOpenUnitProgress(int16_t *handle, int16_t *progressPercent, int16_t *complete) { return 0; }
PICO_STATUS ps6000aCloseUnit(int16_t handle) { return 0; }
PICO_STATUS ps6000aGetUnitInfo(int16_t handle, int8_t *string, int16_t stringLength, int16_t *requiredSize, PICO_INFO info) {
    if (string && stringLength > 0) { strcpy((char *)string, "66AEMU"); }
    return 0;
}
PICO_STATUS ps6000aFlashLed(int16_t handle, int16_t start) { return 0; }
PICO_STATUS ps6000aGetValuesAsync(int16_t handle, uint64_t startIndex, uint64_t noOfSamples, uint64_t downSampleRatio, PICO_RATIO_MODE downSampleRatioMode, uint64_t segmentIndex, PICO_POINTER lpDataReady, PICO_POINTER pParameter) { return 0; }
PICO_STATUS ps6000aGetValuesBulkAsync(int16_t handle, uint64_t startIndex, uint64_t noOfSamples, uint64_t fromSegmentIndex, uint64_t toSegmentIndex, uint64_t downSampleRatio, PICO_RATIO_MODE downSampleRatioMode, PICO_POINTER lpDataReady, PICO_POINTER pParameter) { return 0; }
PICO_STATUS ps6000aGetValues(int16_t handle, uint64_t startIndex, uint64_t *noOfSamples, uint64_t downSampleRatio, PICO_RATIO_MODE downSampleRatioMode, uint64_t segmentIndex, int16_t *overflow) { return 0; }
PICO_STATUS ps6000aGetValuesBulk(int16_t handle, uint64_t startIndex, uint64_t *noOfSamples, uint64_t fromSegmentIndex, uint64_t toSegmentIndex, uint64_t downSampleRatio, PICO_RATIO_MODE downSampleRatioMode, int16_t *overflow) { return 0; }
PICO_STATUS ps6000aGetValuesOverlapped(int16_t handle, uint64_t startIndex, uint64_t *noOfSamples, uint64_t downSampleRatio, PICO_RATIO_MODE downSampleRatioMode, uint64_t fromSegmentIndex, uint64_t toSegmentIndex, int16_t *overflow) { return 0; }
PICO_STATUS ps6000aGetNoOfCaptures(int16_t handle, uint64_t *nCaptures) { return 0; }
PICO_STATUS ps6000aGetNoOfProcessedCaptures(int16_t handle, uint64_t *nCaptures) { return 0; }
PICO_STATUS ps6000aGetTimebase(int16_t handle, uint32_t timebase, uint64_t noSamples, double *timeIntervalNanoseconds, uint64_t *maxSamples, uint64_t segmentIndex) {
    if (timeIntervalNanoseconds) *timeIntervalNanoseconds = 10;
    if (maxSamples) *maxSamples = noSamples;
    return 0;
}
PICO_STATUS ps6000aSetChannelOn(int16_t handle, PICO_CHANNEL channel, PICO_COUPLING coupling, PICO_CONNECT_PROBE_RANGE range, double analogueOffset, PICO_BANDWIDTH_LIMITER bandwidth) { return 0; }
PICO_STATUS ps6000aSetChannelOff(int16_t handle, PICO_CHANNEL channel) { return 0; }
PICO_STATUS ps6000aGetAdcLimits(int16_t handle, PICO_DEVICE_RESOLUTION resolution, int16_t *minValue, int16_t *maxValue) { return 0; }
PICO_STATUS ps6000aSetSimpleTrigger(int16_t handle, int16_t enable, PICO_CHANNEL source, int16_t threshold, PICO_THRESHOLD_DIRECTION direction, uint64_t delay, uint32_t autoTrigger_us) { return 0; }
PICO_STATUS ps6000aSetDataBuffer(int16_t handle, PICO_CHANNEL channel, void *buffer, int32_t bufferLth, PICO_DATA_TYPE dataType, uint64_t waveform, PICO_RATIO_MODE downSampleRatioMode, PICO_ACTION action) { return 0; }
PICO_STATUS ps6000aSetDataBuffers(int16_t handle, PICO_CHANNEL channel, void *bufferMax, void *bufferMin, int32_t bufferLth, PICO_DATA_TYPE dataType, uint64_t waveform, PICO_RATIO_MODE downSampleRatioMode, PICO_ACTION action) { return 0; }
PICO_STATUS ps6000aRunStreaming(int16_t handle, double *sampleInterval, PICO_TIME_UNITS sampleIntervalTimeUnits, uint64_t maxPreTriggerSamples, uint64_t maxPostPreTriggerSamples, int16_t autoStop, uint64_t downSampleRatio, PICO_RATIO_MODE downSampleRatioMode) { return 0; }
PICO_STATUS ps6000aRunBlock(int16_t handle, uint64_t noOfPreTriggerSamples, uint64_t noOfPostTriggerSamples, uint32_t timebase, double *timeIndisposedMs, uint64_t segmentIndex, ps6000aBlockReady lpReady, PICO_POINTER pParameter) { return 0; }
PICO_STATUS ps6000aSetTriggerChannelProperties(int16_t handle, PICO_TRIGGER_CHANNEL_PROPERTIES *channelProperties, int16_t nChannelProperties, int16_t auxOutputEnable, uint32_t autoTriggerMicroSeconds) { return 0; }
PICO_STATUS ps6000aSetTriggerChannelConditions(int16_t handle, PICO_CONDITION *conditions, int16_t nConditions, PICO_ACTION action) { return 0; }
PICO_STATUS ps6000aSetTriggerChannelDirections(int16_t handle, PICO_DIRECTION *directions, int16_t nDirections) { return 0; }
PICO_STATUS ps6000aSetTriggerDelay(int16_t handle, uint64_t delay) { return 0; }
PICO_STATUS ps6000aSetTriggerDigitalPortProperties(int16_t handle, PICO_CHANNEL port, PICO_DIGITAL_CHANNEL_DIRECTIONS *directions, int16_t nDirections) { return 0; }
PICO_STATUS ps6000aStop(int16_t handle) { return 0; }
PICO_STATUS ps6000aSetNoOfCaptures(int16_t handle, uint64_t nCaptures) { return 0; }
PICO_STATUS ps6000aGetTriggerTimeOffset(int16_t handle, int64_t *time, PICO_TIME_UNITS *timeUnits, uint64_t segmentIndex) { return 0; }
PICO_STATUS ps6000aGetValuesTriggerTimeOffsetBulk(int16_t handle, int64_t *times, PICO_TIME_UNITS *timeUnits, uint64_t fromSegmentIndex, uint64_t toSegmentIndex) { return 0; }
PICO_STATUS ps6000aIsReady(int16_t handle, int16_t *ready) { return 0; }
PICO_STATUS ps6000aMemorySegments(int16_t handle, uint64_t nSegments, uint64_t *nMaxSamples) { return 0; }
PICO_STATUS ps6000aNoOfStreamingValues(int16_t handle, uint64_t *noOfValues) { return 0; }
PICO_STATUS ps6000aPingUnit(int16_t handle) { return 0; }
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
