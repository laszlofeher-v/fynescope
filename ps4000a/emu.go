//go:build !noscope && ps4000a && emu

package ps4000a

/*
#cgo CFLAGS: -I/opt/picoscope/include/libps4000a

#include <stdint.h>
#include <string.h>
#include "/opt/picoscope/include/libps4000a/PicoStatus.h"
#include "/opt/picoscope/include/libps4000a/ps4000aApi.h"

PICO_STATUS ps4000aEnumerateUnits(int16_t *count, int8_t *serials, int16_t *serialLth) { return 0; }
PICO_STATUS ps4000aOpenUnit(int16_t *handle, int8_t *serial) { return 0; }
PICO_STATUS ps4000aOpenUnitAsync(int16_t *status, int8_t *serial) { return 0; }
PICO_STATUS ps4000aOpenUnitProgress(int16_t *handle, int16_t *progressPercent, int16_t *complete) { return 0; }
PICO_STATUS ps4000aCloseUnit(int16_t handle) { return 0; }
PICO_STATUS ps4000aGetUnitInfo(int16_t handle, int8_t *string, int16_t stringLength, int16_t *requiredSize, PICO_INFO info) {
    if (string && stringLength > 0) { strcpy((char *)string, "4AEMU"); }
    return 0;
}
PICO_STATUS ps4000aFlashLed(int16_t handle, int16_t start) { return 0; }
PICO_STATUS ps4000aGetValuesAsync(int16_t handle, uint32_t startIndex, uint32_t noOfSamples, uint32_t downSampleRatio, PS4000A_RATIO_MODE downSampleRatioMode, uint32_t segmentIndex, void *lpDataReady, void *pParameter) { return 0; }
PICO_STATUS ps4000aGetValues(int16_t handle, uint32_t startIndex, uint32_t *noOfSamples, uint32_t downSampleRatio, PS4000A_RATIO_MODE downSampleRatioMode, uint32_t segmentIndex, int16_t *overflow) { return 0; }
PICO_STATUS ps4000aGetValuesBulk(int16_t handle, uint32_t *noOfSamples, uint32_t fromSegmentIndex, uint32_t toSegmentIndex, uint32_t downSampleRatio, PS4000A_RATIO_MODE downSampleRatioMode, int16_t *overflow) { return 0; }
PICO_STATUS ps4000aGetValuesOverlapped(int16_t handle, uint32_t startIndex, uint32_t *noOfSamples, uint32_t downSampleRatio, PS4000A_RATIO_MODE downSampleRatioMode, uint32_t segmentIndex, int16_t *overflow) { return 0; }
PICO_STATUS ps4000aGetValuesOverlappedBulk(int16_t handle, uint32_t startIndex, uint32_t *noOfSamples, uint32_t downSampleRatio, PS4000A_RATIO_MODE downSampleRatioMode, uint32_t fromSegmentIndex, uint32_t toSegmentIndex, int16_t *overflow) { return 0; }
PICO_STATUS ps4000aGetAnalogueOffset(int16_t handle, PICO_CONNECT_PROBE_RANGE range, PS4000A_COUPLING coupling, float *maximumOffset, float *minimumOffset) { return 0; }
PICO_STATUS ps4000aGetChannelInformation(int16_t handle, PS4000A_CHANNEL_INFO info, int32_t probe, int32_t *ranges, int32_t *length, int32_t channels) { return 0; }
PICO_STATUS ps4000aGetMaxDownSampleRatio(int16_t handle, uint32_t noOfUnaggregatedSamples, uint32_t *maxDownSampleRatio, PS4000A_RATIO_MODE downSampleRatioMode, uint32_t segmentIndex) { return 0; }
PICO_STATUS ps4000aGetMaxSegments(int16_t handle, uint32_t *maxSegments) { return 0; }
PICO_STATUS ps4000aGetNoOfCaptures(int16_t handle, uint32_t *nCaptures) { return 0; }
PICO_STATUS ps4000aGetNoOfProcessedCaptures(int16_t handle, uint32_t *nProcessedCaptures) { return 0; }
PICO_STATUS ps4000aGetStreamingLatestValues(int16_t handle, ps4000aStreamingReady lpPs4000aReady, void *pParameter) { return 0; }
PICO_STATUS ps4000aGetTimebase(int16_t handle, uint32_t timebase, int32_t noSamples, int32_t *timeIntervalNanoseconds, int32_t *maxSamples, uint32_t segmentIndex) {
    if (timeIntervalNanoseconds) *timeIntervalNanoseconds = 10;
    if (maxSamples) *maxSamples = noSamples;
    return 0;
}
PICO_STATUS ps4000aGetTimebase2(int16_t handle, uint32_t timebase, int32_t noSamples, float *timeIntervalNanoseconds, int32_t *maxSamples, uint32_t segmentIndex) { return 0; }
PICO_STATUS ps4000aSetChannel(int16_t handle, PS4000A_CHANNEL channel, int16_t enabled, PS4000A_COUPLING type, PICO_CONNECT_PROBE_RANGE range, float analogOffset) { return 0; }
PICO_STATUS ps4000aMaximumValue(int16_t handle, int16_t *value) { return 0; }
PICO_STATUS ps4000aMinimumValue(int16_t handle, int16_t *value) { return 0; }
PICO_STATUS ps4000aSetSimpleTrigger(int16_t handle, int16_t enable, PS4000A_CHANNEL source, int16_t threshold, PS4000A_THRESHOLD_DIRECTION direction, uint32_t delay, int16_t autoTrigger_ms) { return 0; }
PICO_STATUS ps4000aSetDataBuffer(int16_t handle, PS4000A_CHANNEL channel, int16_t *buffer, int32_t bufferLth, uint32_t segmentIndex, PS4000A_RATIO_MODE mode) { return 0; }
PICO_STATUS ps4000aSetDataBuffers(int16_t handle, PS4000A_CHANNEL channel, int16_t *bufferMax, int16_t *bufferMin, int32_t bufferLth, uint32_t segmentIndex, PS4000A_RATIO_MODE mode) { return 0; }
PICO_STATUS ps4000aSetEtsTimeBuffer(int16_t handle, int64_t *buffer, int32_t bufferLth) { return 0; }
PICO_STATUS ps4000aSetEtsTimeBuffers(int16_t handle, uint32_t *timeUpper, uint32_t *timeLower, int32_t bufferLth) { return 0; }
PICO_STATUS ps4000aSetEts(int16_t handle, PS4000A_ETS_MODE mode, int16_t etsCycles, int16_t etsInterleave, int32_t *sampleTimePicoseconds) { return 0; }
PICO_STATUS ps4000aRunStreaming(int16_t handle, uint32_t *sampleInterval, PS4000A_TIME_UNITS sampleIntervalTimeUnits, uint32_t maxPreTriggerSamples, uint32_t maxPostTriggerSamples, int16_t autoStop, uint32_t downSampleRatio, PS4000A_RATIO_MODE downSampleRatioMode, uint32_t overviewBufferSize) { return 0; }
PICO_STATUS ps4000aRunBlock(int16_t handle, int32_t noOfPreTriggerSamples, int32_t noOfPostTriggerSamples, uint32_t timebase, int32_t *timeIndisposedMs, uint32_t segmentIndex, ps4000aBlockReady lpReady, void *pParameter) { return 0; }
PICO_STATUS ps4000aSetTriggerChannelProperties(int16_t handle, PS4000A_TRIGGER_CHANNEL_PROPERTIES *channelProperties, int16_t nChannelProperties, int16_t auxOutputEnable, int32_t autoTriggerMilliseconds) { return 0; }
PICO_STATUS ps4000aSetTriggerChannelConditions(int16_t handle, PS4000A_CONDITION *conditions, int16_t nConditions, PS4000A_CONDITIONS_INFO info) { return 0; }
PICO_STATUS ps4000aSetTriggerChannelDirections(int16_t handle, PS4000A_DIRECTION *directions, int16_t nDirections) { return 0; }
PICO_STATUS ps4000aSetTriggerDelay(int16_t handle, uint32_t delay) { return 0; }
PICO_STATUS ps4000aStop(int16_t handle) { return 0; }
PICO_STATUS ps4000aSetSigGenBuiltIn(int16_t handle, int32_t offsetVoltage, uint32_t pkToPk, PS4000A_WAVE_TYPE waveType, double startFrequency, double stopFrequency, double increment, double dwellTime, PS4000A_SWEEP_TYPE sweepType, PS4000A_EXTRA_OPERATIONS operation, uint32_t shots, uint32_t sweeps, PS4000A_SIGGEN_TRIG_TYPE triggerType, PS4000A_SIGGEN_TRIG_SOURCE triggerSource, int16_t extInThreshold) { return 0; }
PICO_STATUS ps4000aSetSigGenBuiltInV2(int16_t handle, int32_t offsetVoltage, uint32_t pkToPk, PS4000A_WAVE_TYPE waveType, double startFrequency, double stopFrequency, double increment, double dwellTime, PS4000A_SWEEP_TYPE sweepType, PS4000A_EXTRA_OPERATIONS operation, uint64_t shots, uint64_t sweeps, PS4000A_SIGGEN_TRIG_TYPE triggerType, PS4000A_SIGGEN_TRIG_SOURCE triggerSource, int16_t extInThreshold) { return 0; }
PICO_STATUS ps4000aSigGenFrequencyToPhase(int16_t handle, double frequency, PS4000A_INDEX_MODE indexMode, uint32_t bufferLength, uint32_t *phase) { return 0; }
PICO_STATUS ps4000aSetNoOfCaptures(int16_t handle, uint32_t nCaptures) { return 0; }
PICO_STATUS ps4000aGetTriggerTimeOffset(int16_t handle, uint32_t *timeUpper, uint32_t *timeLower, PS4000A_TIME_UNITS *timeUnits, uint32_t segmentIndex) { return 0; }
PICO_STATUS ps4000aGetTriggerTimeOffset64(int16_t handle, int64_t *time, PS4000A_TIME_UNITS *timeUnits, uint32_t segmentIndex) { return 0; }
PICO_STATUS ps4000aGetValuesTriggerTimeOffsetBulk(int16_t handle, uint32_t *timesUpper, uint32_t *timesLower, PS4000A_TIME_UNITS *timeUnits, uint32_t fromSegmentIndex, uint32_t toSegmentIndex) { return 0; }
PICO_STATUS ps4000aGetValuesTriggerTimeOffsetBulk64(int16_t handle, int64_t *times, PS4000A_TIME_UNITS *timeUnits, uint32_t fromSegmentIndex, uint32_t toSegmentIndex) { return 0; }
PICO_STATUS ps4000aIsReady(int16_t handle, int16_t *ready) { return 0; }
*/
import "C"
