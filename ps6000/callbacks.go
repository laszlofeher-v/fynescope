//go:build !noscope && ps6000

package ps6000

/*
#include <stdio.h>
#include "/opt/picoscope/include/libps6000/PicoStatus.h"
// C callback function
int ps6000LpDataReady(int16_t handle, PICO_STATUS status, uint32_t noOfSamples,
				int16_t overflow, void * pParameter)
{
	int ps6000LpDataReadyGo(int16_t handle, PICO_STATUS status, uint32_t noOfSamples,
				int16_t overflow, void * pParameter);
	return ps6000LpDataReadyGo(handle, status, noOfSamples, overflow, pParameter);
}

int ps6000LpBlockReady(int16_t handle, PICO_STATUS status, void * pParameter)
{
	int ps6000LpBlockReadyGo(int16_t handle, PICO_STATUS status, void * pParameter);
	return ps6000LpBlockReadyGo(handle, status,  pParameter);
}

int ps6000LpStreamingReady(int16_t handle, int32_t noOfSamples, uint32_t startIndex,
                int16_t overflow, uint32_t triggerAt, int16_t triggered,
                int16_t autoStop, void * pParameter)
{
	int ps6000LpStreamingReadyGo(int16_t handle, int32_t noOfSamples, uint32_t startIndex,
                int16_t overflow, uint32_t triggerAt, int16_t triggered,
                int16_t autoStop, void * pParameter);
	return ps6000LpStreamingReadyGo(handle,noOfSamples, startIndex,overflow,triggerAt, triggered,autoStop,pParameter);
}
*/
import "C"
