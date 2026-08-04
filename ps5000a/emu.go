//go:build emu && ps5000a && !demo

package ps5000a

/*
#cgo CFLAGS: -I/opt/picoscope/include/libps5000a

#include <stdint.h>
#include <string.h>
#include "/opt/picoscope/include/libps5000a/PicoStatus.h"
#include "/opt/picoscope/include/libps5000a/ps5000aApi.h"

PICO_STATUS ps5000aEnumerateUnits(int16_t *count, int8_t *serials, int16_t *serialLth) {
    if (count) *count = 1;
    if (serials && *serialLth > 6) {
        strcpy((char *)serials, "55AEMU");
        *serialLth = 7;
    }
    return 0;
}
PICO_STATUS ps5000aOpenUnit(int16_t *handle, int8_t *serial, PS5000A_DEVICE_RESOLUTION resolution) { return 0; }
PICO_STATUS ps5000aOpenUnitAsync(int16_t *status, int8_t *serial, PS5000A_DEVICE_RESOLUTION resolution) { return 0; }
PICO_STATUS ps5000aOpenUnitProgress(int16_t *handle, int16_t *progressPercent, int16_t *complete) { return 0; }
PICO_STATUS ps5000aCloseUnit(int16_t handle) { return 0; }
PICO_STATUS ps5000aGetUnitInfo(int16_t handle, int8_t *string, int16_t stringLength, int16_t *requiredSize, PICO_INFO info) {
    if (string && stringLength > 0) { strcpy((char *)string, "55AEMU"); }
    return 0;
}
PICO_STATUS ps5000aFlashLed(int16_t handle, int16_t start) { return 0; }
PICO_STATUS ps5000aGetValuesAsync(int16_t handle, uint32_t startIndex, uint32_t noOfSamples, uint32_t downSampleRatio, PS5000A_RATIO_MODE downSampleRatioMode, uint32_t segmentIndex, void *lpDataReady, void *pParameter) { return 0; }
PICO_STATUS ps5000aGetValues(int16_t handle, uint32_t startIndex, uint32_t *noOfSamples, uint32_t downSampleRatio, PS5000A_RATIO_MODE downSampleRatioMode, uint32_t segmentIndex, int16_t *overflow) { return 0; }
PICO_STATUS ps5000aGetValuesBulk(int16_t handle, uint32_t *noOfSamples, uint32_t fromSegmentIndex, uint32_t toSegmentIndex, uint32_t downSampleRatio, PS5000A_RATIO_MODE downSampleRatioMode, int16_t *overflow) { return 0; }
PICO_STATUS ps5000aGetValuesOverlapped(int16_t handle, uint32_t startIndex, uint32_t *noOfSamples, uint32_t downSampleRatio, PS5000A_RATIO_MODE downSampleRatioMode, uint32_t segmentIndex, int16_t *overflow) { return 0; }
PICO_STATUS ps5000aGetValuesOverlappedBulk(int16_t handle, uint32_t startIndex, uint32_t *noOfSamples, uint32_t downSampleRatio, PS5000A_RATIO_MODE downSampleRatioMode, uint32_t fromSegmentIndex, uint32_t toSegmentIndex, int16_t *overflow) { return 0; }
PICO_STATUS ps5000aGetAnalogueOffset(int16_t handle, PS5000A_RANGE range, PS5000A_COUPLING coupling, float *maximumOffset, float *minimumOffset) { return 0; }
PICO_STATUS ps5000aGetChannelInformation(int16_t handle, PS5000A_CHANNEL_INFO info, int32_t probe, int32_t *ranges, int32_t *length, int32_t channels) { return 0; }
PICO_STATUS ps5000aGetMaxDownSampleRatio(int16_t handle, uint32_t noOfUnaggregatedSamples, uint32_t *maxDownSampleRatio, PS5000A_RATIO_MODE downSampleRatioMode, uint32_t segmentIndex) { return 0; }
PICO_STATUS ps5000aGetMaxSegments(int16_t handle, uint32_t *maxSegments) { return 0; }
PICO_STATUS ps5000aGetNoOfCaptures(int16_t handle, uint32_t *nCaptures) { return 0; }
PICO_STATUS ps5000aGetNoOfProcessedCaptures(int16_t handle, uint32_t *nProcessedCaptures) { return 0; }
PICO_STATUS ps5000aGetStreamingLatestValues(int16_t handle, ps5000aStreamingReady lpPs5000aReady, void *pParameter) { return 0; }
PICO_STATUS ps5000aGetTimebase(int16_t handle, uint32_t timebase, int32_t noSamples, int32_t *timeIntervalNanoseconds, int32_t *maxSamples, uint32_t segmentIndex) {
    if (timeIntervalNanoseconds) *timeIntervalNanoseconds = 10;
    if (maxSamples) *maxSamples = noSamples;
    return 0;
}
PICO_STATUS ps5000aGetTimebase2(int16_t handle, uint32_t timebase, int32_t noSamples, float *timeIntervalNanoseconds, int32_t *maxSamples, uint32_t segmentIndex) { return 0; }
PICO_STATUS ps5000aSetChannel(int16_t handle, PS5000A_CHANNEL channel, int16_t enabled, PS5000A_COUPLING type, PS5000A_RANGE range, float analogOffset) { return 0; }
PICO_STATUS ps5000aMaximumValue(int16_t handle, int16_t *value) { return 0; }
PICO_STATUS ps5000aMinimumValue(int16_t handle, int16_t *value) { return 0; }
PICO_STATUS ps5000aSetSimpleTrigger(int16_t handle, int16_t enable, PS5000A_CHANNEL source, int16_t threshold, PS5000A_THRESHOLD_DIRECTION direction, uint32_t delay, int16_t autoTrigger_ms) { return 0; }
PICO_STATUS ps5000aSetDataBuffer(int16_t handle, PS5000A_CHANNEL channel, int16_t *buffer, int32_t bufferLth, uint32_t segmentIndex, PS5000A_RATIO_MODE mode) { return 0; }
PICO_STATUS ps5000aSetDataBuffers(int16_t handle, PS5000A_CHANNEL channel, int16_t *bufferMax, int16_t *bufferMin, int32_t bufferLth, uint32_t segmentIndex, PS5000A_RATIO_MODE mode) { return 0; }
PICO_STATUS ps5000aSetEtsTimeBuffer(int16_t handle, int64_t *buffer, int32_t bufferLth) { return 0; }
PICO_STATUS ps5000aSetEtsTimeBuffers(int16_t handle, uint32_t *timeUpper, uint32_t *timeLower, int32_t bufferLth) { return 0; }
PICO_STATUS ps5000aSetEts(int16_t handle, PS5000A_ETS_MODE mode, int16_t etsCycles, int16_t etsInterleave, int32_t *sampleTimePicoseconds) { return 0; }
PICO_STATUS ps5000aRunStreaming(int16_t handle, uint32_t *sampleInterval, PS5000A_TIME_UNITS sampleIntervalTimeUnits, uint32_t maxPreTriggerSamples, uint32_t maxPostTriggerSamples, int16_t autoStop, uint32_t downSampleRatio, PS5000A_RATIO_MODE downSampleRatioMode, uint32_t overviewBufferSize) { return 0; }
PICO_STATUS ps5000aRunBlock(int16_t handle, int32_t noOfPreTriggerSamples, int32_t noOfPostTriggerSamples, uint32_t timebase, int32_t *timeIndisposedMs, uint32_t segmentIndex, ps5000aBlockReady lpReady, void *pParameter) { return 0; }
PICO_STATUS ps5000aSetTriggerChannelProperties(int16_t handle, PS5000A_TRIGGER_CHANNEL_PROPERTIES *channelProperties, int16_t nChannelProperties, int16_t auxOutputEnable, int32_t autoTriggerMilliseconds) { return 0; }
PICO_STATUS ps5000aSetTriggerChannelConditions(int16_t handle, PS5000A_TRIGGER_CONDITIONS *conditions, int16_t nConditions) { return 0; }
PICO_STATUS ps5000aSetTriggerChannelDirections(int16_t handle, PS5000A_THRESHOLD_DIRECTION channelA, PS5000A_THRESHOLD_DIRECTION channelB, PS5000A_THRESHOLD_DIRECTION channelC, PS5000A_THRESHOLD_DIRECTION channelD, PS5000A_THRESHOLD_DIRECTION ext, PS5000A_THRESHOLD_DIRECTION aux) { return 0; }
PICO_STATUS ps5000aSetTriggerDelay(int16_t handle, uint32_t delay) { return 0; }
PICO_STATUS ps5000aStop(int16_t handle) { return 0; }
PICO_STATUS ps5000aSetSigGenBuiltIn(int16_t handle, int32_t offsetVoltage, uint32_t pkToPk, PS5000A_WAVE_TYPE waveType, float startFrequency, float stopFrequency, float increment, float dwellTime, PS5000A_SWEEP_TYPE sweepType, PS5000A_EXTRA_OPERATIONS operation, uint32_t shots, uint32_t sweeps, PS5000A_SIGGEN_TRIG_TYPE triggerType, PS5000A_SIGGEN_TRIG_SOURCE triggerSource, int16_t extInThreshold) { return 0; }
PICO_STATUS ps5000aSetSigGenBuiltInV2(int16_t handle, int32_t offsetVoltage, uint32_t pkToPk, PS5000A_WAVE_TYPE waveType, double startFrequency, double stopFrequency, double increment, double dwellTime, PS5000A_SWEEP_TYPE sweepType, PS5000A_EXTRA_OPERATIONS operation, uint32_t shots, uint32_t sweeps, PS5000A_SIGGEN_TRIG_TYPE triggerType, PS5000A_SIGGEN_TRIG_SOURCE triggerSource, int16_t extInThreshold) { return 0; }
PICO_STATUS ps5000aSigGenFrequencyToPhase(int16_t handle, double frequency, PS5000A_INDEX_MODE indexMode, uint32_t bufferLength, uint32_t *phase) { return 0; }
PICO_STATUS ps5000aSetNoOfCaptures(int16_t handle, uint32_t nCaptures) { return 0; }
PICO_STATUS ps5000aGetTriggerTimeOffset(int16_t handle, uint32_t *timeUpper, uint32_t *timeLower, PS5000A_TIME_UNITS *timeUnits, uint32_t segmentIndex) { return 0; }
PICO_STATUS ps5000aGetTriggerTimeOffset64(int16_t handle, int64_t *time, PS5000A_TIME_UNITS *timeUnits, uint32_t segmentIndex) { return 0; }
PICO_STATUS ps5000aGetValuesTriggerTimeOffsetBulk(int16_t handle, uint32_t *timesUpper, uint32_t *timesLower, PS5000A_TIME_UNITS *timeUnits, uint32_t fromSegmentIndex, uint32_t toSegmentIndex) { return 0; }
PICO_STATUS ps5000aGetValuesTriggerTimeOffsetBulk64(int16_t handle, int64_t *times, PS5000A_TIME_UNITS *timeUnits, uint32_t fromSegmentIndex, uint32_t toSegmentIndex) { return 0; }
PICO_STATUS ps5000aIsReady(int16_t handle, int16_t *ready) { return 0; }
*/
import "C"
