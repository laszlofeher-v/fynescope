//go:build !noscope && ps3000a && emu

package ps3000a

/*
#cgo CFLAGS: -I/opt/picoscope/include/libps2000a -I/opt/picoscope/include/libps2000
#cgo LDFLAGS: -L/opt/picoscope/lib/ -lps2000a

#include <stdint.h>
#include "/opt/picoscope/include/libps2000/ps2000.h"
#include "/opt/picoscope/include/libps2000a/PicoStatus.h"
#include "/opt/picoscope/include/libps2000a/ps2000aApi.h"

// ps3000a type aliases mapped to ps2000a equivalents
typedef PS2000A_CHANNEL         PS3000A_CHANNEL;
typedef PS2000A_COUPLING        PS3000A_COUPLING;
typedef PS2000A_RANGE           PS3000A_RANGE;
typedef PS2000A_RATIO_MODE      PS3000A_RATIO_MODE;
typedef PS2000A_TIME_UNITS      PS3000A_TIME_UNITS;
typedef PS2000A_ETS_MODE        PS3000A_ETS_MODE;
typedef PS2000A_SWEEP_TYPE      PS3000A_SWEEP_TYPE;
typedef PS2000A_EXTRA_OPERATIONS PS3000A_EXTRA_OPERATIONS;
typedef PS2000A_SIGGEN_TRIG_TYPE PS3000A_SIGGEN_TRIG_TYPE;
typedef PS2000A_SIGGEN_TRIG_SOURCE PS3000A_SIGGEN_TRIG_SOURCE;
typedef PS2000A_INDEX_MODE      PS3000A_INDEX_MODE;
typedef PS2000A_HOLDOFF_TYPE    PS3000A_HOLDOFF_TYPE;
typedef PS2000A_CHANNEL_INFO    PS3000A_CHANNEL_INFO;
typedef PS2000A_THRESHOLD_DIRECTION PS3000A_THRESHOLD_DIRECTION;
typedef PS2000A_PULSE_WIDTH_TYPE PS3000A_PULSE_WIDTH_TYPE;
typedef PS2000A_DIGITAL_PORT    PS3000A_DIGITAL_PORT;
typedef PS2000A_TRIGGER_CHANNEL_PROPERTIES PS3000A_TRIGGER_CHANNEL_PROPERTIES;
typedef PS2000A_TRIGGER_CONDITIONS PS3000A_TRIGGER_CONDITIONS;
typedef PS2000A_PWQ_CONDITIONS  PS3000A_PWQ_CONDITIONS;
typedef PS2000A_DIGITAL_CHANNEL_DIRECTIONS PS3000A_DIGITAL_CHANNEL_DIRECTIONS;

typedef void (*ps3000aBlockReady)(int16_t handle, PICO_STATUS status, void *pParameter);
typedef void (*ps3000aStreamingReady)(
    int16_t handle, int32_t noOfSamples, uint32_t startIndex,
    int16_t overflow, uint32_t triggerAt, int16_t triggered,
    int16_t autoStop, void *pParameter);


PICO_STATUS ps3000aEnumerateUnits(int16_t *count, int8_t *serials, int16_t *serialLth) {
    return ps2000aEnumerateUnits(count, serials, serialLth);
}
PICO_STATUS ps3000aOpenUnit(int16_t *handle, int8_t *serial) {
    return ps2000aOpenUnit(handle, serial);
}
PICO_STATUS ps3000aOpenUnitAsync(int16_t *status, int8_t *serial) {
    return ps2000aOpenUnitAsync(status, serial);
}
PICO_STATUS ps3000aOpenUnitProgress(int16_t *handle, int16_t *progressPercent, int16_t *complete) {
    return ps2000aOpenUnitProgress(handle, progressPercent, complete);
}
PICO_STATUS ps3000aCloseUnit(int16_t handle) {
    return ps2000aCloseUnit(handle);
}
PICO_STATUS ps3000aGetUnitInfo(int16_t handle, int8_t *str, int16_t stringLength, int16_t *requiredSize, PICO_INFO info) {
    PICO_STATUS status = ps2000aGetUnitInfo(handle, str, stringLength, requiredSize, info);
    if (status == 0 && str != 0 && stringLength > 0) {
        if (info == 3) {
            if (str[0] == '2') { str[0] = '3'; }
        } else if (info == 0 || info == 6) {
            if (stringLength >= 7 && str[0] == 'P' && str[1] == 'S' && str[2] == '2') {
                str[2] = '3';
            }
        }
    }
    return status;
}
PICO_STATUS ps3000aFlashLed(int16_t handle, int16_t start) {
    return ps2000aFlashLed(handle, start);
}
PICO_STATUS ps3000aGetValuesAsync(int16_t handle, uint32_t startIndex, uint32_t noOfSamples, uint32_t downSampleRatio, PS3000A_RATIO_MODE downSampleRatioMode, uint32_t segmentIndex, void *lpDataReady, void *pParameter) {
    return ps2000aGetValuesAsync(handle, startIndex, noOfSamples, downSampleRatio, (PS2000A_RATIO_MODE)downSampleRatioMode, segmentIndex, lpDataReady, pParameter);
}
PICO_STATUS ps3000aGetValues(int16_t handle, uint32_t startIndex, uint32_t *noOfSamples, uint32_t downSampleRatio, PS3000A_RATIO_MODE downSampleRatioMode, uint32_t segmentIndex, int16_t *overflow) {
    return ps2000aGetValues(handle, startIndex, noOfSamples, downSampleRatio, (PS2000A_RATIO_MODE)downSampleRatioMode, segmentIndex, overflow);
}
PICO_STATUS ps3000aGetValuesBulk(int16_t handle, uint32_t *noOfSamples, uint32_t fromSegmentIndex, uint32_t toSegmentIndex, uint32_t downSampleRatio, PS3000A_RATIO_MODE downSampleRatioMode, int16_t *overflow) {
    return ps2000aGetValuesBulk(handle, noOfSamples, fromSegmentIndex, toSegmentIndex, downSampleRatio, (PS2000A_RATIO_MODE)downSampleRatioMode, overflow);
}
PICO_STATUS ps3000aGetValuesOverlapped(int16_t handle, uint32_t startIndex, uint32_t *noOfSamples, uint32_t downSampleRatio, PS3000A_RATIO_MODE downSampleRatioMode, uint32_t segmentIndex, int16_t *overflow) {
    return ps2000aGetValuesOverlapped(handle, startIndex, noOfSamples, downSampleRatio, (PS2000A_RATIO_MODE)downSampleRatioMode, segmentIndex, overflow);
}
PICO_STATUS ps3000aGetValuesOverlappedBulk(int16_t handle, uint32_t startIndex, uint32_t *noOfSamples, uint32_t downSampleRatio, PS3000A_RATIO_MODE downSampleRatioMode, uint32_t fromSegmentIndex, uint32_t toSegmentIndex, int16_t *overflow) {
    return ps2000aGetValuesOverlappedBulk(handle, startIndex, noOfSamples, downSampleRatio, (PS2000A_RATIO_MODE)downSampleRatioMode, fromSegmentIndex, toSegmentIndex, overflow);
}
PICO_STATUS ps3000aGetAnalogueOffset(int16_t handle, PS3000A_RANGE range, PS3000A_COUPLING coupling, float *maximumVoltage, float *minimumVoltage) {
    return ps2000aGetAnalogueOffset(handle, (PS2000A_RANGE)range, (PS2000A_COUPLING)coupling, maximumVoltage, minimumVoltage);
}
PICO_STATUS ps3000aGetChannelInformation(int16_t handle, PS3000A_CHANNEL_INFO info, int32_t probe, int32_t *ranges, int32_t *length, int32_t channels) {
    return ps2000aGetChannelInformation(handle, (PS2000A_CHANNEL_INFO)info, probe, ranges, length, channels);
}
PICO_STATUS ps3000aGetMaxDownSampleRatio(int16_t handle, uint32_t noOfUnaggregatedSamples, uint32_t *maxDownSampleRatio, PS3000A_RATIO_MODE downSampleRatioMode, uint32_t segmentIndex) {
    return ps2000aGetMaxDownSampleRatio(handle, noOfUnaggregatedSamples, maxDownSampleRatio, (PS2000A_RATIO_MODE)downSampleRatioMode, segmentIndex);
}
PICO_STATUS ps3000aGetMaxSegments(int16_t handle, uint32_t *maxSegments) {
    return ps2000aGetMaxSegments(handle, maxSegments);
}
PICO_STATUS ps3000aGetNoOfCaptures(int16_t handle, uint32_t *nCaptures) {
    return ps2000aGetNoOfCaptures(handle, nCaptures);
}
PICO_STATUS ps3000aGetNoOfProcessedCaptures(int16_t handle, uint32_t *nCaptures) {
    return ps2000aGetNoOfProcessedCaptures(handle, nCaptures);
}
PICO_STATUS ps3000aGetStreamingLatestValues(int16_t handle, ps3000aStreamingReady lpStreamingReady, void *pParameter) {
    return ps2000aGetStreamingLatestValues(handle, (ps2000aStreamingReady)lpStreamingReady, pParameter);
}
PICO_STATUS ps3000aGetTimebase(int16_t handle, uint32_t timebase, int32_t noSamples, int32_t *timeIntervalNanoseconds, int16_t oversample, int32_t *maxSamples, uint32_t segmentIndex) {
    return ps2000aGetTimebase(handle, timebase, noSamples, timeIntervalNanoseconds, oversample, maxSamples, segmentIndex);
}
PICO_STATUS ps3000aGetTimebase2(int16_t handle, uint32_t timebase, int32_t noSamples, float *timeIntervalNanoseconds, int16_t oversample, int32_t *maxSamples, uint32_t segmentIndex) {
    return ps2000aGetTimebase2(handle, timebase, noSamples, timeIntervalNanoseconds, oversample, maxSamples, segmentIndex);
}
PICO_STATUS ps3000aSetChannel(int16_t handle, PS3000A_CHANNEL channel, int16_t enabled, PS3000A_COUPLING type, PS3000A_RANGE range, float analogueOffset) {
    return ps2000aSetChannel(handle, (PS2000A_CHANNEL)channel, enabled, (PS2000A_COUPLING)type, (PS2000A_RANGE)range, analogueOffset);
}
PICO_STATUS ps3000aMaximumValue(int16_t handle, int16_t *value) {
    return ps2000aMaximumValue(handle, value);
}
PICO_STATUS ps3000aMinimumValue(int16_t handle, int16_t *value) {
    return ps2000aMinimumValue(handle, value);
}
PICO_STATUS ps3000aSetSimpleTrigger(int16_t handle, int16_t enable, PS3000A_CHANNEL source, int16_t threshold, PS3000A_THRESHOLD_DIRECTION direction, uint32_t delay, int16_t autoTriggerMs) {
    return ps2000aSetSimpleTrigger(handle, enable, (PS2000A_CHANNEL)source, threshold, (PS2000A_THRESHOLD_DIRECTION)direction, delay, autoTriggerMs);
}
PICO_STATUS ps3000aSetDataBuffer(int16_t handle, PS3000A_CHANNEL channel, int16_t *buffer, int32_t bufferLth, uint32_t segmentIndex, PS3000A_RATIO_MODE mode) {
    return ps2000aSetDataBuffer(handle, (PS2000A_CHANNEL)channel, buffer, bufferLth, segmentIndex, (PS2000A_RATIO_MODE)mode);
}
PICO_STATUS ps3000aSetDataBuffers(int16_t handle, PS3000A_CHANNEL channel, int16_t *bufferMax, int16_t *bufferMin, int32_t bufferLth, uint32_t segmentIndex, PS3000A_RATIO_MODE mode) {
    return ps2000aSetDataBuffers(handle, (PS2000A_CHANNEL)channel, bufferMax, bufferMin, bufferLth, segmentIndex, (PS2000A_RATIO_MODE)mode);
}
PICO_STATUS ps3000aSetEtsTimeBuffer(int16_t handle, int64_t *buffer, int32_t bufferLth) {
    return ps2000aSetEtsTimeBuffer(handle, buffer, bufferLth);
}
PICO_STATUS ps3000aSetEtsTimeBuffers(int16_t handle, uint32_t *timeUpper, uint32_t *timeLower, int32_t bufferLth) {
    return ps2000aSetEtsTimeBuffers(handle, timeUpper, timeLower, bufferLth);
}
PICO_STATUS ps3000aSetEts(int16_t handle, PS3000A_ETS_MODE mode, int16_t etsCycles, int16_t etsInterleave, int32_t *sampleTimePicoseconds) {
    return ps2000aSetEts(handle, (PS2000A_ETS_MODE)mode, etsCycles, etsInterleave, sampleTimePicoseconds);
}
PICO_STATUS ps3000aRunStreaming(int16_t handle, uint32_t *sampleInterval, PS3000A_TIME_UNITS sampleIntervalTimeUnits, uint32_t maxPreTriggerSamples, uint32_t maxPostTriggerSamples, int16_t autoStop, uint32_t downSampleRatio, PS3000A_RATIO_MODE downSampleRatioMode, uint32_t overviewBufferSize) {
    return ps2000aRunStreaming(handle, sampleInterval, (PS2000A_TIME_UNITS)sampleIntervalTimeUnits, maxPreTriggerSamples, maxPostTriggerSamples, autoStop, downSampleRatio, (PS2000A_RATIO_MODE)downSampleRatioMode, overviewBufferSize);
}
PICO_STATUS ps3000aRunBlock(int16_t handle, int32_t noOfPreTriggerSamples, int32_t noOfPostTriggerSamples, uint32_t timebase, int16_t oversample, int32_t *timeIndisposedMs, uint32_t segmentIndex, ps3000aBlockReady lpReady, void *pParameter) {
    return ps2000aRunBlock(handle, noOfPreTriggerSamples, noOfPostTriggerSamples, timebase, oversample, timeIndisposedMs, segmentIndex, (ps2000aBlockReady)lpReady, pParameter);
}
PICO_STATUS ps3000aSetTriggerChannelProperties(int16_t handle, PS3000A_TRIGGER_CHANNEL_PROPERTIES *channelProperties, int16_t nChannelProperties, int16_t auxOutputEnable, int32_t autoTriggerMilliseconds) {
    return ps2000aSetTriggerChannelProperties(handle, (PS2000A_TRIGGER_CHANNEL_PROPERTIES*)channelProperties, nChannelProperties, auxOutputEnable, autoTriggerMilliseconds);
}
PICO_STATUS ps3000aSetTriggerChannelConditions(int16_t handle, PS3000A_TRIGGER_CONDITIONS *conditions, int16_t nConditions) {
    return ps2000aSetTriggerChannelConditions(handle, (PS2000A_TRIGGER_CONDITIONS*)conditions, nConditions);
}
PICO_STATUS ps3000aSetTriggerChannelDirections(int16_t handle, PS3000A_THRESHOLD_DIRECTION channelA, PS3000A_THRESHOLD_DIRECTION channelB, PS3000A_THRESHOLD_DIRECTION channelC, PS3000A_THRESHOLD_DIRECTION channelD, PS3000A_THRESHOLD_DIRECTION ext, PS3000A_THRESHOLD_DIRECTION aux) {
    return ps2000aSetTriggerChannelDirections(handle, (PS2000A_THRESHOLD_DIRECTION)channelA, (PS2000A_THRESHOLD_DIRECTION)channelB, (PS2000A_THRESHOLD_DIRECTION)channelC, (PS2000A_THRESHOLD_DIRECTION)channelD, (PS2000A_THRESHOLD_DIRECTION)ext, (PS2000A_THRESHOLD_DIRECTION)aux);
}
PICO_STATUS ps3000aSetTriggerDelay(int16_t handle, uint32_t delay) {
    return ps2000aSetTriggerDelay(handle, delay);
}
PICO_STATUS ps3000aSetPulseWidthQualifier(int16_t handle, PS3000A_PWQ_CONDITIONS *conditions, int16_t nConditions, PS3000A_THRESHOLD_DIRECTION direction, uint32_t lower, uint32_t upper, PS3000A_PULSE_WIDTH_TYPE type) {
    return ps2000aSetPulseWidthQualifier(handle, (PS2000A_PWQ_CONDITIONS*)conditions, nConditions, (PS2000A_THRESHOLD_DIRECTION)direction, lower, upper, (PS2000A_PULSE_WIDTH_TYPE)type);
}
PICO_STATUS ps3000aSetTriggerDigitalPortProperties(int16_t handle, PS3000A_DIGITAL_CHANNEL_DIRECTIONS *directions, int16_t nDirections) {
    return ps2000aSetTriggerDigitalPortProperties(handle, (PS2000A_DIGITAL_CHANNEL_DIRECTIONS*)directions, nDirections);
}
PICO_STATUS ps3000aStop(int16_t handle) {
    return ps2000aStop(handle);
}
PICO_STATUS ps3000aSetSigGenBuiltIn(int16_t handle, int32_t offsetVoltage, uint32_t pkToPk, int16_t waveType, float startFrequency, float stopFrequency, float increment, float dwellTime, PS3000A_SWEEP_TYPE sweepType, PS3000A_EXTRA_OPERATIONS operation, uint32_t shots, uint32_t sweeps, PS3000A_SIGGEN_TRIG_TYPE triggerType, PS3000A_SIGGEN_TRIG_SOURCE triggerSource, int16_t extInThreshold) {
    return ps2000aSetSigGenBuiltIn(handle, offsetVoltage, pkToPk, waveType, startFrequency, stopFrequency, increment, dwellTime, (PS2000A_SWEEP_TYPE)sweepType, (PS2000A_EXTRA_OPERATIONS)operation, shots, sweeps, (PS2000A_SIGGEN_TRIG_TYPE)triggerType, (PS2000A_SIGGEN_TRIG_SOURCE)triggerSource, extInThreshold);
}
PICO_STATUS ps3000aSetSigGenBuiltInV2(int16_t handle, int32_t offsetVoltage, uint32_t pkToPk, int16_t waveType, double startFrequency, double stopFrequency, double increment, double dwellTime, PS3000A_SWEEP_TYPE sweepType, PS3000A_EXTRA_OPERATIONS operation, uint32_t shots, uint32_t sweeps, PS3000A_SIGGEN_TRIG_TYPE triggerType, PS3000A_SIGGEN_TRIG_SOURCE triggerSource, int16_t extInThreshold) {
    return ps2000aSetSigGenBuiltInV2(handle, offsetVoltage, pkToPk, waveType, startFrequency, stopFrequency, increment, dwellTime, (PS2000A_SWEEP_TYPE)sweepType, (PS2000A_EXTRA_OPERATIONS)operation, shots, sweeps, (PS2000A_SIGGEN_TRIG_TYPE)triggerType, (PS2000A_SIGGEN_TRIG_SOURCE)triggerSource, extInThreshold);
}
PICO_STATUS ps3000aSigGenFrequencyToPhase(int16_t handle, double frequency, PS3000A_INDEX_MODE indexMode, uint32_t bufferLength, uint32_t *phase) {
    return ps2000aSigGenFrequencyToPhase(handle, frequency, (PS2000A_INDEX_MODE)indexMode, bufferLength, phase);
}
PICO_STATUS ps3000aSetNoOfCaptures(int16_t handle, uint32_t nCaptures) {
    return ps2000aSetNoOfCaptures(handle, nCaptures);
}
PICO_STATUS ps3000aGetTriggerTimeOffset(int16_t handle, uint32_t *timeUpper, uint32_t *timeLower, PS3000A_TIME_UNITS *timeUnits, uint32_t segmentIndex) {
    return ps2000aGetTriggerTimeOffset(handle, timeUpper, timeLower, (PS2000A_TIME_UNITS*)timeUnits, segmentIndex);
}
PICO_STATUS ps3000aGetTriggerTimeOffset64(int16_t handle, int64_t *time, PS3000A_TIME_UNITS *timeUnits, uint32_t segmentIndex) {
    return ps2000aGetTriggerTimeOffset64(handle, time, (PS2000A_TIME_UNITS*)timeUnits, segmentIndex);
}
PICO_STATUS ps3000aGetValuesTriggerTimeOffsetBulk(int16_t handle, uint32_t *timesUpper, uint32_t *timesLower, PS3000A_TIME_UNITS *timeUnits, uint32_t fromSegmentIndex, uint32_t toSegmentIndex) {
    return ps2000aGetValuesTriggerTimeOffsetBulk(handle, timesUpper, timesLower, (PS2000A_TIME_UNITS*)timeUnits, fromSegmentIndex, toSegmentIndex);
}
PICO_STATUS ps3000aGetValuesTriggerTimeOffsetBulk64(int16_t handle, int64_t *times, PS3000A_TIME_UNITS *timeUnits, uint32_t fromSegmentIndex, uint32_t toSegmentIndex) {
    return ps2000aGetValuesTriggerTimeOffsetBulk64(handle, times, (PS2000A_TIME_UNITS*)timeUnits, fromSegmentIndex, toSegmentIndex);
}
PICO_STATUS ps3000aHoldOff(int16_t handle, uint64_t holdoff, PS3000A_HOLDOFF_TYPE type) {
    return ps2000aHoldOff(handle, holdoff, (PS2000A_HOLDOFF_TYPE)type);
}
PICO_STATUS ps3000aIsReady(int16_t handle, int16_t *ready) {
    return ps2000aIsReady(handle, ready);
}
PICO_STATUS ps3000aIsTriggerOrPulseWidthQualifierEnabled(int16_t handle, int16_t *triggerEnabled, int16_t *pulseWidthQualifierEnabled) {
    return ps2000aIsTriggerOrPulseWidthQualifierEnabled(handle, triggerEnabled, pulseWidthQualifierEnabled);
}
PICO_STATUS ps3000aMemorySegments(int16_t handle, uint32_t nSegments, int32_t *nMaxSamples) {
    return ps2000aMemorySegments(handle, nSegments, nMaxSamples);
}
PICO_STATUS ps3000aNoOfStreamingValues(int16_t handle, uint32_t *noOfValues) {
    return ps2000aNoOfStreamingValues(handle, noOfValues);
}
PICO_STATUS ps3000aPingUnit(int16_t handle) {
    return ps2000aPingUnit(handle);
}
PICO_STATUS ps3000aQueryOutputEdgeDetect(int16_t handle, int16_t *state) {
    return ps2000aQueryOutputEdgeDetect(handle, state);
}
PICO_STATUS ps3000aSetDigitalPort(int16_t handle, PS3000A_DIGITAL_PORT port, int16_t enabled, int16_t logicLevel) {
    return ps2000aSetDigitalPort(handle, (PS2000A_DIGITAL_PORT)port, enabled, logicLevel);
}
PICO_STATUS ps3000aSetOutputEdgeDetect(int16_t handle, int16_t state) {
    return ps2000aSetOutputEdgeDetect(handle, state);
}
PICO_STATUS ps3000aSetPulseWidthDigitalPortProperties(int16_t handle, PS3000A_DIGITAL_CHANNEL_DIRECTIONS *directions, int16_t nDirections) {
    return ps2000aSetPulseWidthDigitalPortProperties(handle, (PS2000A_DIGITAL_CHANNEL_DIRECTIONS*)directions, nDirections);
}
PICO_STATUS ps3000aSetSigGenArbitrary(int16_t handle, int32_t offsetVoltage, uint32_t pkToPk, uint32_t startDeltaPhase, uint32_t stopDeltaPhase, uint32_t deltaPhaseIncrement, uint32_t dwellCount, int16_t *arbitraryWaveform, int32_t arbitraryWaveformSize, PS3000A_SWEEP_TYPE sweepType, PS3000A_EXTRA_OPERATIONS operation, PS3000A_INDEX_MODE indexMode, uint32_t shots, uint32_t sweeps, PS3000A_SIGGEN_TRIG_TYPE triggerType, PS3000A_SIGGEN_TRIG_SOURCE triggerSource, int16_t extInThreshold) {
    return ps2000aSetSigGenArbitrary(handle, offsetVoltage, pkToPk, startDeltaPhase, stopDeltaPhase, deltaPhaseIncrement, dwellCount, arbitraryWaveform, arbitraryWaveformSize, (PS2000A_SWEEP_TYPE)sweepType, (PS2000A_EXTRA_OPERATIONS)operation, (PS2000A_INDEX_MODE)indexMode, shots, sweeps, (PS2000A_SIGGEN_TRIG_TYPE)triggerType, (PS2000A_SIGGEN_TRIG_SOURCE)triggerSource, extInThreshold);
}
PICO_STATUS ps3000aSetSigGenPropertiesArbitrary(int16_t handle, uint32_t startDeltaPhase, uint32_t stopDeltaPhase, uint32_t deltaPhaseIncrement, uint32_t dwellCount, PS3000A_SWEEP_TYPE sweepType, uint32_t shots, uint32_t sweeps, PS3000A_SIGGEN_TRIG_TYPE triggerType, PS3000A_SIGGEN_TRIG_SOURCE triggerSource, int16_t extInThreshold) {
    return ps2000aSetSigGenPropertiesArbitrary(handle, startDeltaPhase, stopDeltaPhase, deltaPhaseIncrement, dwellCount, (PS2000A_SWEEP_TYPE)sweepType, shots, sweeps, (PS2000A_SIGGEN_TRIG_TYPE)triggerType, (PS2000A_SIGGEN_TRIG_SOURCE)triggerSource, extInThreshold);
}
PICO_STATUS ps3000aSetSigGenPropertiesBuiltIn(int16_t handle, double startFrequency, double stopFrequency, double increment, double dwellTime, PS3000A_SWEEP_TYPE sweepType, uint32_t shots, uint32_t sweeps, PS3000A_SIGGEN_TRIG_TYPE triggerType, PS3000A_SIGGEN_TRIG_SOURCE triggerSource, int16_t extInThreshold) {
    return ps2000aSetSigGenPropertiesBuiltIn(handle, startFrequency, stopFrequency, increment, dwellTime, (PS2000A_SWEEP_TYPE)sweepType, shots, sweeps, (PS2000A_SIGGEN_TRIG_TYPE)triggerType, (PS2000A_SIGGEN_TRIG_SOURCE)triggerSource, extInThreshold);
}
PICO_STATUS ps3000aSigGenArbitraryMinMaxValues(int16_t handle, int16_t *minArbitraryWaveformValue, int16_t *maxArbitraryWaveformValue, uint32_t *minArbitraryWaveformSize, uint32_t *maxArbitraryWaveformSize) {
    return ps2000aSigGenArbitraryMinMaxValues(handle, minArbitraryWaveformValue, maxArbitraryWaveformValue, minArbitraryWaveformSize, maxArbitraryWaveformSize);
}
PICO_STATUS ps3000aSigGenSoftwareControl(int16_t handle, int16_t state) {
    return ps2000aSigGenSoftwareControl(handle, state);
}
*/
import "C"
