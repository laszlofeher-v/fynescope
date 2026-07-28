//go:build !noscope && ps5000

package ps5000

/*
#include <stdio.h>
#include "/opt/picoscope/include/libps5000/PicoStatus.h"
// C callback function
int ps5000LpDataReady(int16_t handle, PICO_STATUS status, uint32_t noOfSamples,
				int16_t overflow, void * pParameter)
{
	int ps5000LpDataReadyGo(int16_t handle, PICO_STATUS status, uint32_t noOfSamples,
				int16_t overflow, void * pParameter);
	return ps5000LpDataReadyGo(handle, status, noOfSamples, overflow, pParameter);
}

int ps5000LpBlockReady(int16_t handle, PICO_STATUS status, void * pParameter)
{
	int ps5000LpBlockReadyGo(int16_t handle, PICO_STATUS status, void * pParameter);
	return ps5000LpBlockReadyGo(handle, status,  pParameter);
}

int ps5000LpStreamingReady(int16_t handle, int32_t noOfSamples, uint32_t startIndex,
                int16_t overflow, uint32_t triggerAt, int16_t triggered,
                int16_t autoStop, void * pParameter)
{
	int ps5000LpStreamingReadyGo(int16_t handle, int32_t noOfSamples, uint32_t startIndex,
                int16_t overflow, uint32_t triggerAt, int16_t triggered,
                int16_t autoStop, void * pParameter);
	return ps5000LpStreamingReadyGo(handle,noOfSamples, startIndex,overflow,triggerAt, triggered,autoStop,pParameter);
}
*/
import "C"
