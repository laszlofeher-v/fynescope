//go:build emu && ps6000 && !demo

package ps6000

/*
#cgo CFLAGS: -I/opt/picoscope/include/libps6000

#include <stdint.h>
#include <string.h>
#include "/opt/picoscope/include/libps6000/ps6000Api.h"

PICO_STATUS ps6000EnumerateUnits(int16_t *count, int8_t *serials, int16_t *serialLth) { return 0; }
PICO_STATUS ps6000OpenUnit(int16_t *handle, int8_t *serial) { return 0; }
PICO_STATUS ps6000OpenUnitAsync(int16_t *status, int8_t *serial) { return 0; }
PICO_STATUS ps6000OpenUnitProgress(int16_t *handle, int16_t *progressPercent, int16_t *complete) { return 0; }
PICO_STATUS ps6000CloseUnit(int16_t handle) { return 0; }
PICO_STATUS ps6000GetUnitInfo(int16_t handle, int8_t *string, int16_t stringLength, int16_t *requiredSize, PICO_INFO info) {
    if (string && stringLength > 0) { strcpy((char *)string, "66EMU"); }
    return 0;
}
PICO_STATUS ps6000FlashLed(int16_t handle, int16_t start) { return 0; }
PICO_STATUS ps6000GetValuesAsync(int16_t handle, uint32_t startIndex, uint32_t noOfSamples, uint32_t downSampleRatio, PS6000_RATIO_MODE downSampleRatioMode, uint32_t segmentIndex, void *lpDataReady, void *pParameter) { return 0; }
PICO_STATUS ps6000GetValues(int16_t handle, uint32_t startIndex, uint32_t *noOfSamples, uint32_t downSampleRatio, PS6000_RATIO_MODE downSampleRatioMode, uint32_t segmentIndex, int16_t *overflow) { return 0; }
PICO_STATUS ps6000GetValuesBulk(int16_t handle, uint32_t *noOfSamples, uint32_t fromSegmentIndex, uint32_t toSegmentIndex, int16_t *overflow) { return 0; }

PICO_STATUS ps6000GetMaxDownSampleRatio(int16_t handle, uint32_t noOfUnaggreatedSamples, uint32_t *maxDownSampleRatio, PS6000_RATIO_MODE downSampleRatioMode, uint32_t segmentIndex) { return 0; }
PICO_STATUS ps6000GetNoOfCaptures(int16_t handle, uint32_t *nCaptures) { return 0; }
PICO_STATUS ps6000GetNoOfProcessedCaptures(int16_t handle, uint32_t *nCaptures) { return 0; }
PICO_STATUS ps6000GetStreamingLatestValues(int16_t handle, ps6000StreamingReady lpPs6000Ready, void *pParameter) { return 0; }
PICO_STATUS ps6000GetTimebase(int16_t handle, uint32_t timebase, uint32_t noSamples, int32_t *timeIntervalNanoseconds, int16_t oversample, uint32_t *maxSamples, uint32_t segmentIndex) {
    if (timeIntervalNanoseconds) *timeIntervalNanoseconds = 10;
    if (maxSamples) *maxSamples = noSamples;
    return 0;
}
PICO_STATUS ps6000GetTimebase2(int16_t handle, uint32_t timebase, uint32_t noSamples, float *timeIntervalNanoseconds, int16_t oversample, uint32_t *maxSamples, uint32_t segmentIndex) { return 0; }
PICO_STATUS ps6000SetChannel(int16_t handle, PS6000_CHANNEL channel, int16_t enabled, PS6000_COUPLING type, PS6000_RANGE range, float analogueOffset, PS6000_BANDWIDTH_LIMITER bandwidth) { return 0; }
PICO_STATUS ps6000SetSimpleTrigger(int16_t handle, int16_t enable, PS6000_CHANNEL source, int16_t threshold, THRESHOLD_DIRECTION direction, uint32_t delay, int16_t autoTrigger_ms) { return 0; }
PICO_STATUS ps6000SetDataBufferBulk(int16_t handle, PS6000_CHANNEL channel, int16_t *buffer, uint32_t bufferLth, uint32_t waveform, PS6000_RATIO_MODE downSampleRatioMode) { return 0; }
PICO_STATUS ps6000SetDataBuffersBulk(int16_t handle, PS6000_CHANNEL channel, int16_t *bufferMax, int16_t *bufferMin, uint32_t bufferLth, uint32_t waveform, PS6000_RATIO_MODE downSampleRatioMode) { return 0; }
PICO_STATUS ps6000SetEtsTimeBuffer(int16_t handle, int64_t *buffer, int32_t bufferLth) { return 0; }
PICO_STATUS ps6000SetEtsTimeBuffers(int16_t handle, uint32_t *timeUpper, uint32_t *timeLower, int32_t bufferLth) { return 0; }
PICO_STATUS ps6000SetEts(int16_t handle, PS6000_ETS_MODE mode, int16_t etsCycles, int16_t etsInterleave, int32_t *sampleTimePicoseconds) { return 0; }
PICO_STATUS ps6000RunStreaming(int16_t handle, uint32_t *sampleInterval, PS6000_TIME_UNITS sampleIntervalTimeUnits, uint32_t maxPreTriggerSamples, uint32_t maxPostPreTriggerSamples, int16_t autoStop, uint32_t downSampleRatio, uint32_t overviewBufferSize) { return 0; }
PICO_STATUS ps6000RunBlock(int16_t handle, uint32_t noOfPreTriggerSamples, uint32_t noOfPostTriggerSamples, uint32_t timebase, int16_t oversample, int32_t *timeIndisposedMs, uint32_t segmentIndex, ps6000BlockReady lpReady, void *pParameter) { return 0; }
PICO_STATUS ps6000SetTriggerChannelProperties(int16_t handle, TRIGGER_CHANNEL_PROPERTIES *channelProperties, int16_t nChannelProperties, int16_t auxOutputEnable, int32_t autoTriggerMilliseconds) { return 0; }
PICO_STATUS ps6000SetTriggerChannelConditions(int16_t handle, TRIGGER_CONDITIONS *conditions, int16_t nConditions) { return 0; }
PICO_STATUS ps6000SetTriggerChannelDirections(int16_t handle, THRESHOLD_DIRECTION channelA, THRESHOLD_DIRECTION channelB, THRESHOLD_DIRECTION channelC, THRESHOLD_DIRECTION channelD, THRESHOLD_DIRECTION ext, THRESHOLD_DIRECTION aux) { return 0; }
PICO_STATUS ps6000SetTriggerDelay(int16_t handle, uint32_t delay) { return 0; }
PICO_STATUS ps6000SetPulseWidthQualifier(int16_t handle, PWQ_CONDITIONS *conditions, int16_t nConditions, THRESHOLD_DIRECTION direction, uint32_t lower, uint32_t upper, PULSE_WIDTH_TYPE type) { return 0; }
PICO_STATUS ps6000Stop(int16_t handle) { return 0; }
PICO_STATUS ps6000SetSigGenBuiltIn(int16_t handle, int32_t offsetVoltage, uint32_t pkToPk, int16_t waveType, float startFrequency, float stopFrequency, float increment, float dwellTime, SWEEP_TYPE sweepType, int16_t operationType, uint32_t shots, uint32_t sweeps, SIGGEN_TRIG_TYPE triggerType, SIGGEN_TRIG_SOURCE triggerSource, int16_t extInThreshold) { return 0; }
PICO_STATUS ps6000SigGenFrequencyToPhase(int16_t handle, double frequency, INDEX_MODE indexMode, uint32_t bufferLength, uint32_t *phase) { return 0; }
PICO_STATUS ps6000SetNoOfCaptures(int16_t handle, uint32_t nCaptures) { return 0; }
PICO_STATUS ps6000GetTriggerTimeOffset(int16_t handle, uint32_t *timeUpper, uint32_t *timeLower, PS6000_TIME_UNITS *timeUnits, uint32_t segmentIndex) { return 0; }
PICO_STATUS ps6000GetTriggerTimeOffset64(int16_t handle, int64_t *time, PS6000_TIME_UNITS *timeUnits, uint32_t segmentIndex) { return 0; }
PICO_STATUS ps6000GetValuesTriggerTimeOffsetBulk(int16_t handle, uint32_t *timesUpper, uint32_t *timesLower, PS6000_TIME_UNITS *timeUnits, uint32_t fromSegmentIndex, uint32_t toSegmentIndex) { return 0; }
PICO_STATUS ps6000GetValuesTriggerTimeOffsetBulk64(int16_t handle, int64_t *times, PS6000_TIME_UNITS *timeUnits, uint32_t fromSegmentIndex, uint32_t toSegmentIndex) { return 0; }

PICO_STATUS ps6000IsReady(int16_t handle, int16_t *ready) { return 0; }
PICO_STATUS ps6000IsTriggerOrPulseWidthQualifierEnabled(int16_t handle, int16_t *triggerEnabled, int16_t *pulseWidthQualifierEnabled) { return 0; }
PICO_STATUS ps6000MemorySegments(int16_t handle, uint32_t nSegments, uint32_t *nMaxSamples) { return 0; }
PICO_STATUS ps6000NoOfStreamingValues(int16_t handle, uint32_t *noOfValues) { return 0; }
PICO_STATUS ps6000PingUnit(int16_t handle) { return 0; }
PICO_STATUS ps6000SetSigGenArbitrary(int16_t handle, int32_t offsetVoltage, uint32_t pkToPk, uint32_t startDeltaPhase, uint32_t stopDeltaPhase, uint32_t deltaPhaseIncrement, uint32_t dwellCount, int16_t *arbitraryWaveform, int32_t arbitraryWaveformSize, SWEEP_TYPE sweepType, int16_t operationType, INDEX_MODE indexMode, uint32_t shots, uint32_t sweeps, SIGGEN_TRIG_TYPE triggerType, SIGGEN_TRIG_SOURCE triggerSource, int16_t extInThreshold) { return 0; }
PICO_STATUS ps6000SigGenArbitraryMinMaxValues(int16_t handle, int16_t *minArbitraryWaveformValue, int16_t *maxArbitraryWaveformValue, uint32_t *minArbitraryWaveformSize, uint32_t *maxArbitraryWaveformSize) { return 0; }
PICO_STATUS ps6000SigGenSoftwareControl(int16_t handle, int16_t state) { return 0; }

*/
import "C"
