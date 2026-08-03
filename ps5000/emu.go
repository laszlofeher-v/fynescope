//go:build emu && ps5000 && !demo

package ps5000

/*
#cgo CFLAGS: -I/opt/picoscope/include/libps5000

#include <stdint.h>
#include <string.h>
#include "/opt/picoscope/include/libps5000/ps5000Api.h"


PICO_STATUS ps5000OpenUnit(int16_t *handle) { return 0; }
PICO_STATUS ps5000OpenUnitAsync(int16_t *status) { return 0; }
PICO_STATUS ps5000OpenUnitProgress(int16_t *handle, int16_t *progressPercent, int16_t *complete) { return 0; }
PICO_STATUS ps5000CloseUnit(int16_t handle) { return 0; }
PICO_STATUS ps5000GetUnitInfo(int16_t handle, int8_t *string, int16_t stringLength, int16_t *requiredSize, PICO_INFO info) {
    if (string && stringLength > 0) { strcpy((char *)string, "55EMU"); }
    return 0;
}
PICO_STATUS ps5000FlashLed(int16_t handle, int16_t start) { return 0; }
PICO_STATUS ps5000GetValuesAsync(int16_t handle, uint32_t startIndex, uint32_t noOfSamples, uint32_t downSampleRatio, int16_t downSampleRatioMode, uint16_t segmentIndex, void *lpDataReady, void *pParameter) { return 0; }
PICO_STATUS ps5000GetValues(int16_t handle, uint32_t startIndex, uint32_t *noOfSamples, uint32_t downSampleRatio, int16_t downSampleRatioMode, uint16_t segmentIndex, int16_t *overflow) { return 0; }
PICO_STATUS ps5000GetValuesBulk(int16_t handle, uint32_t *noOfSamples, uint16_t fromSegmentIndex, uint16_t toSegmentIndex, int16_t *overflow) { return 0; }

PICO_STATUS ps5000GetMaxDownSampleRatio(int16_t handle, uint32_t noOfUnaggreatedSamples, uint32_t *maxDownSampleRatio, int16_t downSampleRatioMode, uint16_t segmentIndex) { return 0; }

PICO_STATUS ps5000GetStreamingLatestValues(int16_t handle, ps5000StreamingReady lpPs5000Ready, void *pParameter) { return 0; }
PICO_STATUS ps5000GetTimebase(int16_t handle, uint32_t timebase, int32_t noSamples, int32_t *timeIntervalNanoseconds, int16_t oversample, int32_t *maxSamples, uint16_t segmentIndex) {
    if (timeIntervalNanoseconds) *timeIntervalNanoseconds = 10;
    if (maxSamples) *maxSamples = noSamples;
    return 0;
}
PICO_STATUS ps5000GetTimebase2(int16_t handle, uint32_t timebase, int32_t noSamples, float *timeIntervalNanoseconds, int16_t oversample, int32_t *maxSamples, uint16_t segmentIndex) { return 0; }
PICO_STATUS ps5000SetChannel(int16_t handle, PS5000_CHANNEL channel, int16_t enabled, int16_t dc, PS5000_RANGE range) { return 0; }
PICO_STATUS ps5000SetSimpleTrigger(int16_t handle, int16_t enable, PS5000_CHANNEL source, int16_t threshold, THRESHOLD_DIRECTION direction, uint32_t delay, int16_t autoTrigger_ms) { return 0; }
PICO_STATUS ps5000SetDataBuffer(int16_t handle, PS5000_CHANNEL channel, int16_t *buffer, int32_t bufferLth) { return 0; }
PICO_STATUS ps5000SetDataBuffers(int16_t handle, PS5000_CHANNEL channel, int16_t *bufferMax, int16_t *bufferMin, int32_t bufferLth) { return 0; }
PICO_STATUS ps5000SetEtsTimeBuffer(int16_t handle, int64_t *buffer, int32_t bufferLth) { return 0; }
PICO_STATUS ps5000SetEtsTimeBuffers(int16_t handle, uint32_t *timeUpper, uint32_t *timeLower, int32_t bufferLth) { return 0; }
PICO_STATUS ps5000SetEts(int16_t handle, PS5000_ETS_MODE mode, int16_t etsCycles, int16_t etsInterleave, int32_t *sampleTimePicoseconds) { return 0; }
PICO_STATUS ps5000RunStreaming(int16_t handle, uint32_t *sampleInterval, PS5000_TIME_UNITS sampleIntervalTimeUnits, uint32_t maxPreTriggerSamples, uint32_t maxPostPreTriggerSamples, int16_t autoStop, uint32_t downSampleRatio, uint32_t overviewBufferSize) { return 0; }
PICO_STATUS ps5000RunBlock(int16_t handle, int32_t noOfPreTriggerSamples, int32_t noOfPostTriggerSamples, uint32_t timebase, int16_t oversample, int32_t *timeIndisposedMs, uint16_t segmentIndex, ps5000BlockReady lpReady, void *pParameter) { return 0; }
PICO_STATUS ps5000SetTriggerChannelProperties(int16_t handle, TRIGGER_CHANNEL_PROPERTIES *channelProperties, int16_t nChannelProperties, int16_t auxOutputEnable, int32_t autoTriggerMilliseconds) { return 0; }
PICO_STATUS ps5000SetTriggerChannelConditions(int16_t handle, TRIGGER_CONDITIONS *conditions, int16_t nConditions) { return 0; }
PICO_STATUS ps5000SetTriggerChannelDirections(int16_t handle, THRESHOLD_DIRECTION channelA, THRESHOLD_DIRECTION channelB, THRESHOLD_DIRECTION channelC, THRESHOLD_DIRECTION channelD, THRESHOLD_DIRECTION ext, THRESHOLD_DIRECTION aux) { return 0; }
PICO_STATUS ps5000SetTriggerDelay(int16_t handle, uint32_t delay) { return 0; }
PICO_STATUS ps5000SetPulseWidthQualifier(int16_t handle, PWQ_CONDITIONS *conditions, int16_t nConditions, THRESHOLD_DIRECTION direction, uint32_t lower, uint32_t upper, PULSE_WIDTH_TYPE type) { return 0; }
PICO_STATUS ps5000Stop(int16_t handle) { return 0; }
PICO_STATUS ps5000SetSigGenBuiltIn(int16_t handle, int32_t offsetVoltage, uint32_t pkToPk, int16_t waveType, float startFrequency, float stopFrequency, float increment, float dwellTime, SWEEP_TYPE sweepType, int16_t operationType, uint32_t shots, uint32_t sweeps, SIGGEN_TRIG_TYPE triggerType, SIGGEN_TRIG_SOURCE triggerSource, int16_t extInThreshold) { return 0; }
PICO_STATUS ps5000SigGenFrequencyToPhase(int16_t handle, double frequency, INDEX_MODE indexMode, uint32_t bufferLength, uint32_t *phase) { return 0; }
PICO_STATUS ps5000SetNoOfCaptures(int16_t handle, uint16_t nCaptures) { return 0; }
PICO_STATUS ps5000GetTriggerTimeOffset(int16_t handle, uint32_t *timeUpper, uint32_t *timeLower, PS5000_TIME_UNITS *timeUnits, uint16_t segmentIndex) { return 0; }
PICO_STATUS ps5000GetTriggerTimeOffset64(int16_t handle, int64_t *time, PS5000_TIME_UNITS *timeUnits, uint16_t segmentIndex) { return 0; }
PICO_STATUS ps5000GetValuesTriggerTimeOffsetBulk(int16_t handle, uint32_t *timesUpper, uint32_t *timesLower, PS5000_TIME_UNITS *timeUnits, uint16_t fromSegmentIndex, uint16_t toSegmentIndex) { return 0; }
PICO_STATUS ps5000GetValuesTriggerTimeOffsetBulk64(int16_t handle, int64_t *times, PS5000_TIME_UNITS *timeUnits, uint16_t fromSegmentIndex, uint16_t toSegmentIndex) { return 0; }

PICO_STATUS ps5000IsReady(int16_t handle, int16_t *ready) { return 0; }
PICO_STATUS ps5000IsTriggerOrPulseWidthQualifierEnabled(int16_t handle, int16_t *triggerEnabled, int16_t *pulseWidthQualifierEnabled) { return 0; }
PICO_STATUS ps5000MemorySegments(int16_t handle, uint16_t nSegments, int32_t *nMaxSamples) { return 0; }
PICO_STATUS ps5000NoOfStreamingValues(int16_t handle, uint32_t *noOfValues) { return 0; }
PICO_STATUS ps5000PingUnit(int16_t handle) { return 0; }
PICO_STATUS ps5000SetSigGenArbitrary(int16_t handle, int32_t offsetVoltage, uint32_t pkToPk, uint32_t startDeltaPhase, uint32_t stopDeltaPhase, uint32_t deltaPhaseIncrement, uint32_t dwellCount, int16_t *arbitraryWaveform, int32_t arbitraryWaveformSize, SWEEP_TYPE sweepType, int16_t operationType, INDEX_MODE indexMode, uint32_t shots, uint32_t sweeps, SIGGEN_TRIG_TYPE triggerType, SIGGEN_TRIG_SOURCE triggerSource, int16_t extInThreshold) { return 0; }
PICO_STATUS ps5000SigGenArbitraryMinMaxValues(int16_t handle, int16_t *minArbitraryWaveformValue, int16_t *maxArbitraryWaveformValue, uint32_t *minArbitraryWaveformSize, uint32_t *maxArbitraryWaveformSize) { return 0; }
PICO_STATUS ps5000SigGenSoftwareControl(int16_t handle, int16_t state) { return 0; }

*/
import "C"
