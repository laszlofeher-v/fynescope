//go:build !noscope && ps2000 && emu

package ps2000

/*
#cgo CFLAGS: -I/opt/picoscope/include/libps2000 -I/opt/picoscope/include/libps2000a
#cgo LDFLAGS: -L/opt/picoscope/lib/ -lps2000a

#include <stdint.h>
#include <string.h>
#include "/opt/picoscope/include/libps2000/ps2000.h"
#include "/opt/picoscope/include/libps2000a/PicoStatus.h"
#include "/opt/picoscope/include/libps2000a/ps2000aApi.h"
// ps2000a type aliases mapped to ps2000a equivalents
typedef PS2000A_COUPLING        PS2000_COUPLING;
typedef PS2000A_RATIO_MODE      PS2000_RATIO_MODE;
typedef PS2000A_EXTRA_OPERATIONS PS2000_EXTRA_OPERATIONS;
typedef PS2000A_SIGGEN_TRIG_TYPE PS2000_SIGGEN_TRIG_TYPE;
typedef PS2000A_SIGGEN_TRIG_SOURCE PS2000_SIGGEN_TRIG_SOURCE;
typedef PS2000A_INDEX_MODE      PS2000_INDEX_MODE;
typedef PS2000A_HOLDOFF_TYPE    PS2000_HOLDOFF_TYPE;
typedef PS2000A_CHANNEL_INFO    PS2000_CHANNEL_INFO;
typedef PS2000A_DIGITAL_PORT    PS2000_DIGITAL_PORT;
typedef PS2000A_TRIGGER_CONDITIONS P2000_TRIGGER_CONDITIONS;

typedef void (*ps2000aBlockReady)(int16_t handle, PICO_STATUS status, void *pParameter);
typedef void (*ps2000aStreamingReady)(
    int16_t handle, int32_t noOfSamples, uint32_t startIndex,
    int16_t overflow, uint32_t triggerAt, int16_t triggered,
    int16_t autoStop, void *pParameter);

int16_t ps2000_open_unit(void) {
    int16_t handle= 0;
    PICO_STATUS status = ps2000aOpenUnit(&handle, NULL);
    if (status == 0) return handle;
    return 0;
}

int16_t ps2000_get_unit_info(int16_t handle, int8_t *string, int16_t string_length, int16_t line) {
    int16_t reqSize = 0;
    PICO_STATUS status = ps2000aGetUnitInfo(handle, string, string_length, &reqSize, line);
    if (status == 0) {
        if (string && string_length > 0 && line == 3) {
            strcpy((char *)&string[0], "24EMU");
        }
        return reqSize;
    }
    return 0;
}

int16_t ps2000_flash_led(int16_t handle) {
    return (ps2000aFlashLed(handle, 5) == 0) ? 1 : 0;
}

int16_t ps2000_close_unit(int16_t handle) {
    return (ps2000aCloseUnit(handle) == 0) ? 1 : 0;
}

int16_t ps2000_set_channel(int16_t handle, int16_t channel, int16_t enabled, int16_t dc, int16_t range) {
    return (ps2000aSetChannel(handle, (PS2000A_CHANNEL)channel, enabled, (PS2000A_COUPLING)dc, (PS2000A_RANGE)range, 0.0f) == 0) ? 1 : 0;
}

int16_t ps2000_get_timebase(int16_t handle, int16_t timebase, int32_t no_of_samples, int32_t *time_interval, int16_t *time_units, int16_t oversample, int32_t *max_samples) {
    return ps2000aGetTimebase(handle, timebase, no_of_samples, time_interval, oversample, max_samples, 0);
}

int16_t ps2000_set_trigger(int16_t handle, int16_t source, int16_t threshold, int16_t direction, int16_t delay, int16_t auto_trigger_ms) {
    return (ps2000aSetSimpleTrigger(handle, 1, (PS2000A_CHANNEL)source, threshold, (PS2000A_THRESHOLD_DIRECTION)direction, delay, auto_trigger_ms) == 0) ? 1 : 0;
}

int16_t ps2000_set_trigger2(int16_t handle, int16_t source, int16_t threshold, int16_t direction, float delay, int16_t auto_trigger_ms) {
    return (ps2000aSetSimpleTrigger(handle, 1, (PS2000A_CHANNEL)source, threshold, (PS2000A_THRESHOLD_DIRECTION)direction, (uint32_t)(delay * 100000.0f), auto_trigger_ms) == 0) ? 1 : 0;
}

int16_t ps2000_run_block(int16_t handle, int32_t no_of_values, int16_t timebase, int16_t oversample, int32_t * time_indisposed_ms) {
    return (ps2000aRunBlock(handle, 0, no_of_values, timebase, oversample,  time_indisposed_ms, 0, NULL, NULL) == 0) ? 1 : 0;
}

int16_t ps2000_run_streaming(int16_t handle, int16_t sample_interval_ms, int32_t max_samples, int16_t windowed) {
    uint32_t interval = sample_interval_ms;
    return (ps2000aRunStreaming(handle, &interval, PS2000A_MS, 0, max_samples, 0, 1, PS2000A_RATIO_MODE_NONE, 100000) == 0) ? 1 : 0;
}

int16_t ps2000_run_streaming_ns(int16_t handle, uint32_t sample_interval, PS2000_TIME_UNITS time_units, uint32_t max_samples, int16_t auto_stop, uint32_t noOfSamplesPerAggregate, uint32_t overview_buffer_size) {
    return (ps2000aRunStreaming(handle, &sample_interval, (PS2000A_TIME_UNITS)time_units, 0, max_samples, auto_stop, 1, PS2000A_RATIO_MODE_NONE, overview_buffer_size) == 0) ? 1 : 0;
}

int16_t ps2000_ready(int16_t handle) {
    int16_t ready = 0;
    ps2000aIsReady(handle, &ready);
    return ready;
}

int16_t ps2000_stop(int16_t handle) {
    return (ps2000aStop(handle) == 0) ? 1 : 0;
}

int32_t ps2000_get_values(int16_t handle, int16_t *buffer_a, int16_t *buffer_b, int16_t *buffer_c, int16_t *buffer_d, int16_t *overflow, int32_t no_of_values) {
    if (buffer_a) ps2000aSetDataBuffer(handle, PS2000A_CHANNEL_A, buffer_a, no_of_values, 0, PS2000A_RATIO_MODE_NONE);
    if (buffer_b) ps2000aSetDataBuffer(handle, PS2000A_CHANNEL_B, buffer_b, no_of_values, 0, PS2000A_RATIO_MODE_NONE);
    if (buffer_c) ps2000aSetDataBuffer(handle, PS2000A_CHANNEL_C, buffer_c, no_of_values, 0, PS2000A_RATIO_MODE_NONE);
    if (buffer_d) ps2000aSetDataBuffer(handle, PS2000A_CHANNEL_D, buffer_d, no_of_values, 0, PS2000A_RATIO_MODE_NONE);

    uint32_t samples = no_of_values;
    int16_t ov = 0;
    ps2000aGetValues(handle, 0, &samples, 1, PS2000A_RATIO_MODE_NONE, 0, &ov);
    if (overflow) *overflow = ov;
    return samples;
}

int32_t ps2000_get_times_and_values(int16_t handle, int32_t *times, int16_t *buffer_a, int16_t *buffer_b, int16_t *buffer_c, int16_t *buffer_d, int16_t *overflow, int16_t time_units, int32_t no_of_values) {
    return ps2000_get_values(handle, buffer_a, buffer_b, buffer_c, buffer_d, overflow, no_of_values);
}

int16_t ps2000_last_button_press(int16_t handle) {
    return 0;
}

int32_t ps2000_set_ets(int16_t handle, int16_t mode, int16_t ets_cycles, int16_t ets_interleave) {
    int32_t sampleTime = 0;
    if (ps2000aSetEts(handle, (PS2000A_ETS_MODE)mode, ets_cycles, ets_interleave, &sampleTime) == 0) return sampleTime;
    return 0;
}

int16_t ps2000_set_led(int16_t handle, int16_t state) {
    return (ps2000aFlashLed(handle, state) == 0) ? 1 : 0;
}

int16_t ps2000_open_unit_async(void) {
    int16_t status = 0;
    ps2000aOpenUnitAsync(&status, NULL);
    return (status == 0) ? 1 : 0;
}

int16_t ps2000_open_unit_progress(int16_t *handle, int16_t *progress_percent) {
    int16_t complete = 0;
    ps2000aOpenUnitProgress(handle, progress_percent, &complete);
    return complete;
}

int16_t ps2000_get_streaming_last_values(int16_t handle, GetOverviewBuffersMaxMin getOverviewBuffersMaxMin) {
    return 1;
}

int16_t ps2000_overview_buffer_status(int16_t handle, int16_t *previous_buffer_overrun) {
    return 1;
}

uint32_t ps2000_get_streaming_values(int16_t handle, double *start_time, int16_t *pbuffer_a_max, int16_t *pbuffer_a_min, int16_t *pbuffer_b_max, int16_t *pbuffer_b_min, int16_t *pbuffer_c_max, int16_t *pbuffer_c_min, int16_t *pbuffer_d_max, int16_t *pbuffer_d_min, int16_t *overflow, uint32_t *triggerAt, int16_t *triggered, uint32_t no_of_values, uint32_t noOfSamplesPerAggregate) {
    return 0;
}

uint32_t ps2000_get_streaming_values_no_aggregation(int16_t handle, double *start_time, int16_t * pbuffer_a, int16_t * pbuffer_b, int16_t * pbuffer_c, int16_t * pbuffer_d, int16_t * overflow, uint32_t * triggerAt, int16_t * trigger, uint32_t no_of_values) {
    return 0;
}

int16_t ps2000_set_light(int16_t handle, int16_t state) {
    return 1;
}

int16_t ps2000_set_sig_gen_arbitrary(int16_t handle, int32_t offsetVoltage, uint32_t pkToPk, uint32_t startDeltaPhase, uint32_t stopDeltaPhase, uint32_t deltaPhaseIncrement, uint32_t dwellCount, uint8_t *arbitraryWaveform, int32_t arbitraryWaveformSize, PS2000_SWEEP_TYPE sweepType, uint32_t sweeps) {
    return (ps2000aSetSigGenArbitrary(handle, offsetVoltage, pkToPk, startDeltaPhase, stopDeltaPhase, deltaPhaseIncrement, dwellCount, (int16_t*)arbitraryWaveform, arbitraryWaveformSize, (PS2000A_SWEEP_TYPE)sweepType, 0, PS2000A_SINGLE, 0, sweeps, PS2000A_SIGGEN_RISING, PS2000A_SIGGEN_NONE, 0) == 0) ? 1 : 0;
}

int16_t ps2000_set_sig_gen_built_in(int16_t handle, int32_t offsetVoltage, uint32_t pkToPk, PS2000_WAVE_TYPE waveType, float startFrequency, float stopFrequency, float increment, float dwellTime, PS2000_SWEEP_TYPE sweepType, uint32_t sweeps) {
    return (ps2000aSetSigGenBuiltIn(handle, offsetVoltage, pkToPk, (PS2000A_WAVE_TYPE)waveType, startFrequency, stopFrequency, increment, dwellTime, (PS2000A_SWEEP_TYPE)sweepType, 0, 0, sweeps, PS2000A_SIGGEN_RISING, PS2000A_SIGGEN_NONE, 0) == 0) ? 1 : 0;
}

int16_t ps2000SetAdvTriggerChannelProperties(int16_t handle, PS2000_TRIGGER_CHANNEL_PROPERTIES *channelProperties, int16_t nChannelProperties, int32_t autoTriggerMilliseconds) {
    return (ps2000aSetTriggerChannelProperties(handle, (PS2000A_TRIGGER_CHANNEL_PROPERTIES*)channelProperties, nChannelProperties, 0, autoTriggerMilliseconds) == 0) ? 1 : 0;
}

int16_t ps2000SetAdvTriggerChannelConditions(int16_t handle, PS2000_TRIGGER_CONDITIONS *conditions, int16_t nConditions) {
    return ps2000aSetTriggerChannelConditions(handle, (PS2000A_TRIGGER_CONDITIONS*)conditions, nConditions);
}

int16_t ps2000SetAdvTriggerChannelDirections(int16_t handle, PS2000_THRESHOLD_DIRECTION channelA, PS2000_THRESHOLD_DIRECTION channelB, PS2000_THRESHOLD_DIRECTION channelC, PS2000_THRESHOLD_DIRECTION channelD, PS2000_THRESHOLD_DIRECTION ext) {
    return ps2000aSetTriggerChannelDirections(handle, (PS2000A_THRESHOLD_DIRECTION)channelA, (PS2000A_THRESHOLD_DIRECTION)channelB, (PS2000A_THRESHOLD_DIRECTION)channelC, (PS2000A_THRESHOLD_DIRECTION)channelD, (PS2000A_THRESHOLD_DIRECTION)ext, (PS2000A_THRESHOLD_DIRECTION)0);
}

int16_t ps2000SetPulseWidthQualifier(int16_t handle, PS2000_PWQ_CONDITIONS *conditions, int16_t nConditions, PS2000_THRESHOLD_DIRECTION direction, uint32_t lower, uint32_t upper, PS2000_PULSE_WIDTH_TYPE type) {
    return ps2000aSetPulseWidthQualifier(handle, (PS2000A_PWQ_CONDITIONS*)conditions, nConditions, (PS2000A_THRESHOLD_DIRECTION)direction, lower, upper, (PS2000A_PULSE_WIDTH_TYPE)type);
}

int16_t ps2000SetAdvTriggerDelay(int16_t handle, uint32_t delay, float preTriggerDelay) {
    return (ps2000aSetTriggerDelay(handle, delay) == 0) ? 1 : 0;
}

int16_t ps2000PingUnit(int16_t handle) {
    return (ps2000aPingUnit(handle) == 0) ? 1 : 0;
}

*/
import "C"
