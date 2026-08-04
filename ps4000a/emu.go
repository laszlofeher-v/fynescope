//go:build emu && ps4000a && !demo

package ps4000a

/*
#cgo CFLAGS: -I/opt/picoscope/include/libps4000a
#cgo LDFLAGS: -L/opt/picoscope/lib/ -lps2000a

#include <stdint.h>
#include <string.h>
#include "/opt/picoscope/include/libps4000a/PicoStatus.h"
#include "/opt/picoscope/include/libps4000a/ps4000aApi.h"

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
uint32_t ps2000aGetAnalogueOffset(int16_t handle, uint32_t range, uint32_t coupling, float *maximumOffset, float *minimumOffset);
uint32_t ps2000aGetChannelInformation(int16_t handle, uint32_t info, int32_t probe, int32_t *ranges, int32_t *length, int32_t channels);
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
uint32_t ps2000aSetDataBuffer(int16_t handle, uint32_t channel, int16_t *buffer, int32_t bufferLth, uint32_t segmentIndex, uint32_t mode);
uint32_t ps2000aSetDataBuffers(int16_t handle, uint32_t channel, int16_t *bufferMax, int16_t *bufferMin, int32_t bufferLth, uint32_t segmentIndex, uint32_t mode);
uint32_t ps2000aSetEtsTimeBuffer(int16_t handle, int64_t *buffer, int32_t bufferLth);
uint32_t ps2000aSetEtsTimeBuffers(int16_t handle, uint32_t *timeUpper, uint32_t *timeLower, int32_t bufferLth);
uint32_t ps2000aSetEts(int16_t handle, uint32_t mode, int16_t etsCycles, int16_t etsInterleave, int32_t *sampleTimePicoseconds);
uint32_t ps2000aRunStreaming(int16_t handle, uint32_t *sampleInterval, uint32_t sampleIntervalTimeUnits, uint32_t maxPreTriggerSamples, uint32_t maxPostTriggerSamples, int16_t autoStop, uint32_t downSampleRatio, uint32_t downSampleRatioMode, uint32_t overviewBufferSize);
uint32_t ps2000aRunBlock(int16_t handle, int32_t noOfPreTriggerSamples, int32_t noOfPostTriggerSamples, uint32_t timebase, int16_t oversample, int32_t *timeIndisposedMs, uint32_t segmentIndex, void *lpReady, void *pParameter);
uint32_t ps2000aSetTriggerChannelProperties(int16_t handle, void *channelProperties, int16_t nChannelProperties, int16_t auxOutputEnable, int32_t autoTriggerMilliseconds);
uint32_t ps2000aSetTriggerChannelConditions(int16_t handle, void *conditions, int16_t nConditions);
uint32_t ps2000aSetTriggerChannelDirections(int16_t handle, uint32_t channelA, uint32_t channelB, uint32_t channelC, uint32_t channelD, uint32_t ext, uint32_t aux);
uint32_t ps2000aSetTriggerDelay(int16_t handle, uint32_t delay);
uint32_t ps2000aStop(int16_t handle);
uint32_t ps2000aSetSigGenBuiltIn(int16_t handle, int32_t offsetVoltage, uint32_t pkToPk, int16_t waveType, float startFrequency, float stopFrequency, float increment, float dwellTime, uint32_t sweepType, uint32_t operation, uint32_t shots, uint32_t sweeps, uint32_t triggerType, uint32_t triggerSource, int16_t extInThreshold);
uint32_t ps2000aSetSigGenBuiltInV2(int16_t handle, int32_t offsetVoltage, uint32_t pkToPk, int16_t waveType, double startFrequency, double stopFrequency, double increment, double dwellTime, uint32_t sweepType, uint32_t operation, uint32_t shots, uint32_t sweeps, uint32_t triggerType, uint32_t triggerSource, int16_t extInThreshold);
uint32_t ps2000aSigGenFrequencyToPhase(int16_t handle, double frequency, uint32_t indexMode, uint32_t bufferLength, uint32_t *phase);
uint32_t ps2000aSetNoOfCaptures(int16_t handle, uint32_t nCaptures);
uint32_t ps2000aGetTriggerTimeOffset(int16_t handle, uint32_t *timeUpper, uint32_t *timeLower, void *timeUnits, uint32_t segmentIndex);
uint32_t ps2000aGetTriggerTimeOffset64(int16_t handle, int64_t *time, void *timeUnits, uint32_t segmentIndex);
uint32_t ps2000aGetValuesTriggerTimeOffsetBulk(int16_t handle, uint32_t *timesUpper, uint32_t *timesLower, void *timeUnits, uint32_t fromSegmentIndex, uint32_t toSegmentIndex);
uint32_t ps2000aGetValuesTriggerTimeOffsetBulk64(int16_t handle, int64_t *times, void *timeUnits, uint32_t fromSegmentIndex, uint32_t toSegmentIndex);
uint32_t ps2000aIsReady(int16_t handle, int16_t *ready);
uint32_t ps2000aMemorySegments(int16_t handle, uint32_t nSegments, int32_t *nMaxSamples);
uint32_t ps2000aPingUnit(int16_t handle);

PICO_STATUS ps4000aEnumerateUnits(int16_t *count, int8_t *serials, int16_t *serialLth) {
    return ps2000aEnumerateUnits(count, serials, serialLth);
}
PICO_STATUS ps4000aOpenUnit(int16_t *handle, int8_t *serial) { return ps2000aOpenUnit(handle, serial); }
PICO_STATUS ps4000aOpenUnitAsync(int16_t *status, int8_t *serial) { return ps2000aOpenUnitAsync(status, serial); }
PICO_STATUS ps4000aOpenUnitProgress(int16_t *handle, int16_t *progressPercent, int16_t *complete) { return ps2000aOpenUnitProgress(handle, progressPercent, complete); }
PICO_STATUS ps4000aCloseUnit(int16_t handle) { return ps2000aCloseUnit(handle); }
PICO_STATUS ps4000aGetUnitInfo(int16_t handle, int8_t *string, int16_t stringLength, int16_t *requiredSize, PICO_INFO info) {
    uint32_t status = ps2000aGetUnitInfo(handle, string, stringLength, requiredSize, (uint32_t)info);
    if (status == 0 && string != 0 && stringLength > 0) {
        if (info == 3) strcpy((char *)string, "44AEM");
    }
    return status;
}
PICO_STATUS ps4000aFlashLed(int16_t handle, int16_t start) { return ps2000aFlashLed(handle, start); }
PICO_STATUS ps4000aGetValuesAsync(int16_t handle, uint32_t startIndex, uint32_t noOfSamples, uint32_t downSampleRatio, PS4000A_RATIO_MODE downSampleRatioMode, uint32_t segmentIndex, void *lpDataReady, void *pParameter) { return ps2000aGetValuesAsync(handle, startIndex, noOfSamples, downSampleRatio, (uint32_t)downSampleRatioMode, segmentIndex, lpDataReady, pParameter); }
PICO_STATUS ps4000aGetValues(int16_t handle, uint32_t startIndex, uint32_t *noOfSamples, uint32_t downSampleRatio, PS4000A_RATIO_MODE downSampleRatioMode, uint32_t segmentIndex, int16_t *overflow) { return ps2000aGetValues(handle, startIndex, noOfSamples, downSampleRatio, (uint32_t)downSampleRatioMode, segmentIndex, overflow); }
PICO_STATUS ps4000aGetValuesBulk(int16_t handle, uint32_t *noOfSamples, uint32_t fromSegmentIndex, uint32_t toSegmentIndex, uint32_t downSampleRatio, PS4000A_RATIO_MODE downSampleRatioMode, int16_t *overflow) { return ps2000aGetValuesBulk(handle, noOfSamples, fromSegmentIndex, toSegmentIndex, downSampleRatio, (uint32_t)downSampleRatioMode, overflow); }
PICO_STATUS ps4000aGetValuesOverlapped(int16_t handle, uint32_t startIndex, uint32_t *noOfSamples, uint32_t downSampleRatio, PS4000A_RATIO_MODE downSampleRatioMode, uint32_t segmentIndex, int16_t *overflow) { return ps2000aGetValuesOverlapped(handle, startIndex, noOfSamples, downSampleRatio, (uint32_t)downSampleRatioMode, segmentIndex, overflow); }
PICO_STATUS ps4000aGetValuesOverlappedBulk(int16_t handle, uint32_t startIndex, uint32_t *noOfSamples, uint32_t downSampleRatio, PS4000A_RATIO_MODE downSampleRatioMode, uint32_t fromSegmentIndex, uint32_t toSegmentIndex, int16_t *overflow) { return ps2000aGetValuesOverlappedBulk(handle, startIndex, noOfSamples, downSampleRatio, (uint32_t)downSampleRatioMode, fromSegmentIndex, toSegmentIndex, overflow); }
PICO_STATUS ps4000aGetAnalogueOffset(int16_t handle, PICO_CONNECT_PROBE_RANGE range, PS4000A_COUPLING coupling, float *maximumOffset, float *minimumOffset) { return ps2000aGetAnalogueOffset(handle, (uint32_t)range, (uint32_t)coupling, maximumOffset, minimumOffset); }
PICO_STATUS ps4000aGetChannelInformation(int16_t handle, PS4000A_CHANNEL_INFO info, int32_t probe, int32_t *ranges, int32_t *length, int32_t channels) { return ps2000aGetChannelInformation(handle, (uint32_t)info, probe, ranges, length, channels); }
PICO_STATUS ps4000aGetMaxDownSampleRatio(int16_t handle, uint32_t noOfUnaggregatedSamples, uint32_t *maxDownSampleRatio, PS4000A_RATIO_MODE downSampleRatioMode, uint32_t segmentIndex) { return ps2000aGetMaxDownSampleRatio(handle, noOfUnaggregatedSamples, maxDownSampleRatio, (uint32_t)downSampleRatioMode, segmentIndex); }
PICO_STATUS ps4000aGetMaxSegments(int16_t handle, uint32_t *maxSegments) { return ps2000aGetMaxSegments(handle, maxSegments); }
PICO_STATUS ps4000aGetNoOfCaptures(int16_t handle, uint32_t *nCaptures) { return ps2000aGetNoOfCaptures(handle, nCaptures); }
PICO_STATUS ps4000aGetNoOfProcessedCaptures(int16_t handle, uint32_t *nProcessedCaptures) { return ps2000aGetNoOfProcessedCaptures(handle, nProcessedCaptures); }
PICO_STATUS ps4000aGetStreamingLatestValues(int16_t handle, ps4000aStreamingReady lpPs4000aReady, void *pParameter) { return ps2000aGetStreamingLatestValues(handle, (void*)lpPs4000aReady, pParameter); }
PICO_STATUS ps4000aGetTimebase(int16_t handle, uint32_t timebase, int32_t noSamples, int32_t *timeIntervalNanoseconds, int32_t *maxSamples, uint32_t segmentIndex) { return ps2000aGetTimebase(handle, timebase, noSamples, timeIntervalNanoseconds, 0, maxSamples, segmentIndex); }
PICO_STATUS ps4000aGetTimebase2(int16_t handle, uint32_t timebase, int32_t noSamples, float *timeIntervalNanoseconds, int32_t *maxSamples, uint32_t segmentIndex) { return ps2000aGetTimebase2(handle, timebase, noSamples, timeIntervalNanoseconds, 0, maxSamples, segmentIndex); }
PICO_STATUS ps4000aSetChannel(int16_t handle, PS4000A_CHANNEL channel, int16_t enabled, PS4000A_COUPLING type, PICO_CONNECT_PROBE_RANGE range, float analogOffset) { return ps2000aSetChannel(handle, (uint32_t)channel, enabled, (uint32_t)type, (uint32_t)range, analogOffset); }
PICO_STATUS ps4000aMaximumValue(int16_t handle, int16_t *value) { return ps2000aMaximumValue(handle, value); }
PICO_STATUS ps4000aMinimumValue(int16_t handle, int16_t *value) { return ps2000aMinimumValue(handle, value); }
PICO_STATUS ps4000aSetSimpleTrigger(int16_t handle, int16_t enable, PS4000A_CHANNEL source, int16_t threshold, PS4000A_THRESHOLD_DIRECTION direction, uint32_t delay, int16_t autoTrigger_ms) { return ps2000aSetSimpleTrigger(handle, enable, (uint32_t)source, threshold, (uint32_t)direction, delay, autoTrigger_ms); }
PICO_STATUS ps4000aSetDataBuffer(int16_t handle, PS4000A_CHANNEL channel, int16_t *buffer, int32_t bufferLth, uint32_t segmentIndex, PS4000A_RATIO_MODE mode) { return ps2000aSetDataBuffer(handle, (uint32_t)channel, buffer, bufferLth, segmentIndex, (uint32_t)mode); }
PICO_STATUS ps4000aSetDataBuffers(int16_t handle, PS4000A_CHANNEL channel, int16_t *bufferMax, int16_t *bufferMin, int32_t bufferLth, uint32_t segmentIndex, PS4000A_RATIO_MODE mode) { return ps2000aSetDataBuffers(handle, (uint32_t)channel, bufferMax, bufferMin, bufferLth, segmentIndex, (uint32_t)mode); }
PICO_STATUS ps4000aSetEtsTimeBuffer(int16_t handle, int64_t *buffer, int32_t bufferLth) { return ps2000aSetEtsTimeBuffer(handle, buffer, bufferLth); }
PICO_STATUS ps4000aSetEtsTimeBuffers(int16_t handle, uint32_t *timeUpper, uint32_t *timeLower, int32_t bufferLth) { return ps2000aSetEtsTimeBuffers(handle, timeUpper, timeLower, bufferLth); }
PICO_STATUS ps4000aSetEts(int16_t handle, PS4000A_ETS_MODE mode, int16_t etsCycles, int16_t etsInterleave, int32_t *sampleTimePicoseconds) { return ps2000aSetEts(handle, (uint32_t)mode, etsCycles, etsInterleave, sampleTimePicoseconds); }
PICO_STATUS ps4000aRunStreaming(int16_t handle, uint32_t *sampleInterval, PS4000A_TIME_UNITS sampleIntervalTimeUnits, uint32_t maxPreTriggerSamples, uint32_t maxPostTriggerSamples, int16_t autoStop, uint32_t downSampleRatio, PS4000A_RATIO_MODE downSampleRatioMode, uint32_t overviewBufferSize) { return ps2000aRunStreaming(handle, sampleInterval, (uint32_t)sampleIntervalTimeUnits, maxPreTriggerSamples, maxPostTriggerSamples, autoStop, downSampleRatio, (uint32_t)downSampleRatioMode, overviewBufferSize); }
PICO_STATUS ps4000aRunBlock(int16_t handle, int32_t noOfPreTriggerSamples, int32_t noOfPostTriggerSamples, uint32_t timebase, int32_t *timeIndisposedMs, uint32_t segmentIndex, ps4000aBlockReady lpReady, void *pParameter) { return ps2000aRunBlock(handle, noOfPreTriggerSamples, noOfPostTriggerSamples, timebase, 0, timeIndisposedMs, segmentIndex, (void*)lpReady, pParameter); }
typedef enum {
  PS2000A_CONDITION_DONT_CARE,
  PS2000A_CONDITION_TRUE,
  PS2000A_CONDITION_FALSE,
  PS2000A_CONDITION_MAX
} PS2000A_TRIGGER_STATE_LOCAL;

typedef struct tPS2000ATriggerConditionsLocal
{
  PS2000A_TRIGGER_STATE_LOCAL channelA;
  PS2000A_TRIGGER_STATE_LOCAL channelB;
  PS2000A_TRIGGER_STATE_LOCAL channelC;
  PS2000A_TRIGGER_STATE_LOCAL channelD;
  PS2000A_TRIGGER_STATE_LOCAL external;
  PS2000A_TRIGGER_STATE_LOCAL aux;
  PS2000A_TRIGGER_STATE_LOCAL pulseWidthQualifier;
  PS2000A_TRIGGER_STATE_LOCAL digital;
} PS2000A_TRIGGER_CONDITIONS_LOCAL;

PICO_STATUS ps4000aSetTriggerChannelProperties(int16_t handle, PS4000A_TRIGGER_CHANNEL_PROPERTIES *channelProperties, int16_t nChannelProperties, int16_t auxOutputEnable, int32_t autoTriggerMilliseconds) { return ps2000aSetTriggerChannelProperties(handle, (void*)channelProperties, nChannelProperties, auxOutputEnable, autoTriggerMilliseconds); }

PICO_STATUS ps4000aSetTriggerChannelConditions(int16_t handle, PS4000A_CONDITION *conditions, int16_t nConditions, PS4000A_CONDITIONS_INFO info) {
    PS2000A_TRIGGER_CONDITIONS_LOCAL tc;
    tc.channelA = PS2000A_CONDITION_DONT_CARE;
    tc.channelB = PS2000A_CONDITION_DONT_CARE;
    tc.channelC = PS2000A_CONDITION_DONT_CARE;
    tc.channelD = PS2000A_CONDITION_DONT_CARE;
    tc.external = PS2000A_CONDITION_DONT_CARE;
    tc.aux = PS2000A_CONDITION_DONT_CARE;
    tc.pulseWidthQualifier = PS2000A_CONDITION_DONT_CARE;
    tc.digital = PS2000A_CONDITION_DONT_CARE;

    for (int16_t i = 0; i < nConditions; i++) {
        switch (conditions[i].source) {
            case PS4000A_CHANNEL_A: tc.channelA = (PS2000A_TRIGGER_STATE_LOCAL)conditions[i].condition; break;
            case PS4000A_CHANNEL_B: tc.channelB = (PS2000A_TRIGGER_STATE_LOCAL)conditions[i].condition; break;
            case PS4000A_CHANNEL_C: tc.channelC = (PS2000A_TRIGGER_STATE_LOCAL)conditions[i].condition; break;
            case PS4000A_CHANNEL_D: tc.channelD = (PS2000A_TRIGGER_STATE_LOCAL)conditions[i].condition; break;
            case PS4000A_EXTERNAL: tc.external = (PS2000A_TRIGGER_STATE_LOCAL)conditions[i].condition; break;
            case PS4000A_TRIGGER_AUX: tc.aux = (PS2000A_TRIGGER_STATE_LOCAL)conditions[i].condition; break;
        }
    }
    return ps2000aSetTriggerChannelConditions(handle, (void*)&tc, 1);
}

PICO_STATUS ps4000aSetTriggerChannelDirections(int16_t handle, PS4000A_DIRECTION *directions, int16_t nDirections) {
    uint32_t dirA = 2, dirB = 2, dirC = 2, dirD = 2, dirExt = 2, dirAux = 2;
    for (int16_t i = 0; i < nDirections; i++) {
        switch (directions[i].channel) {
            case PS4000A_CHANNEL_A: dirA = (uint32_t)directions[i].direction; break;
            case PS4000A_CHANNEL_B: dirB = (uint32_t)directions[i].direction; break;
            case PS4000A_CHANNEL_C: dirC = (uint32_t)directions[i].direction; break;
            case PS4000A_CHANNEL_D: dirD = (uint32_t)directions[i].direction; break;
            case PS4000A_EXTERNAL: dirExt = (uint32_t)directions[i].direction; break;
            case PS4000A_TRIGGER_AUX: dirAux = (uint32_t)directions[i].direction; break;
        }
    }
    return ps2000aSetTriggerChannelDirections(handle, dirA, dirB, dirC, dirD, dirExt, dirAux);
}
PICO_STATUS ps4000aSetTriggerDelay(int16_t handle, uint32_t delay) { return ps2000aSetTriggerDelay(handle, delay); }
PICO_STATUS ps4000aStop(int16_t handle) { return ps2000aStop(handle); }
PICO_STATUS ps4000aSetSigGenBuiltIn(int16_t handle, int32_t offsetVoltage, uint32_t pkToPk, PS4000A_WAVE_TYPE waveType, double startFrequency, double stopFrequency, double increment, double dwellTime, PS4000A_SWEEP_TYPE sweepType, PS4000A_EXTRA_OPERATIONS operation, uint32_t shots, uint32_t sweeps, PS4000A_SIGGEN_TRIG_TYPE triggerType, PS4000A_SIGGEN_TRIG_SOURCE triggerSource, int16_t extInThreshold) { return ps2000aSetSigGenBuiltInV2(handle, offsetVoltage, pkToPk, (int16_t)waveType, startFrequency, stopFrequency, increment, dwellTime, (uint32_t)sweepType, (uint32_t)operation, shots, sweeps, (uint32_t)triggerType, (uint32_t)triggerSource, extInThreshold); }
PICO_STATUS ps4000aSetSigGenBuiltInV2(int16_t handle, int32_t offsetVoltage, uint32_t pkToPk, PS4000A_WAVE_TYPE waveType, double startFrequency, double stopFrequency, double increment, double dwellTime, PS4000A_SWEEP_TYPE sweepType, PS4000A_EXTRA_OPERATIONS operation, uint64_t shots, uint64_t sweeps, PS4000A_SIGGEN_TRIG_TYPE triggerType, PS4000A_SIGGEN_TRIG_SOURCE triggerSource, int16_t extInThreshold) { return ps2000aSetSigGenBuiltInV2(handle, offsetVoltage, pkToPk, (int16_t)waveType, startFrequency, stopFrequency, increment, dwellTime, (uint32_t)sweepType, (uint32_t)operation, (uint32_t)shots, (uint32_t)sweeps, (uint32_t)triggerType, (uint32_t)triggerSource, extInThreshold); }
PICO_STATUS ps4000aSigGenFrequencyToPhase(int16_t handle, double frequency, PS4000A_INDEX_MODE indexMode, uint32_t bufferLength, uint32_t *phase) { return ps2000aSigGenFrequencyToPhase(handle, frequency, (uint32_t)indexMode, bufferLength, phase); }
PICO_STATUS ps4000aSetNoOfCaptures(int16_t handle, uint32_t nCaptures) { return ps2000aSetNoOfCaptures(handle, nCaptures); }
PICO_STATUS ps4000aGetTriggerTimeOffset(int16_t handle, uint32_t *timeUpper, uint32_t *timeLower, PS4000A_TIME_UNITS *timeUnits, uint32_t segmentIndex) { return ps2000aGetTriggerTimeOffset(handle, timeUpper, timeLower, (void*)timeUnits, segmentIndex); }
PICO_STATUS ps4000aGetTriggerTimeOffset64(int16_t handle, int64_t *time, PS4000A_TIME_UNITS *timeUnits, uint32_t segmentIndex) { return ps2000aGetTriggerTimeOffset64(handle, time, (void*)timeUnits, segmentIndex); }
PICO_STATUS ps4000aGetValuesTriggerTimeOffsetBulk(int16_t handle, uint32_t *timesUpper, uint32_t *timesLower, PS4000A_TIME_UNITS *timeUnits, uint32_t fromSegmentIndex, uint32_t toSegmentIndex) { return ps2000aGetValuesTriggerTimeOffsetBulk(handle, timesUpper, timesLower, (void*)timeUnits, fromSegmentIndex, toSegmentIndex); }
PICO_STATUS ps4000aGetValuesTriggerTimeOffsetBulk64(int16_t handle, int64_t *times, PS4000A_TIME_UNITS *timeUnits, uint32_t fromSegmentIndex, uint32_t toSegmentIndex) { return ps2000aGetValuesTriggerTimeOffsetBulk64(handle, times, (void*)timeUnits, fromSegmentIndex, toSegmentIndex); }
PICO_STATUS ps4000aIsReady(int16_t handle, int16_t *ready) { return ps2000aIsReady(handle, ready); }
PICO_STATUS ps4000aMemorySegments(int16_t handle, uint32_t nSegments, int32_t *nMaxSamples) { return ps2000aMemorySegments(handle, nSegments, nMaxSamples); }
PICO_STATUS ps4000aPingUnit(int16_t handle) { return ps2000aPingUnit(handle); }
*/
import "C"
