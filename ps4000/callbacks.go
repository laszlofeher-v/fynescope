//go:build !noscope && ps4000

package ps4000

/*
#include <stdio.h>
#include "/opt/picoscope/include/libps4000/PicoStatus.h"
// C callback function
int ps4000LpDataReady(int16_t handle, PICO_STATUS status, uint32_t noOfSamples,
				int16_t overflow, void * pParameter)
{
	int ps4000LpDataReadyGo(int16_t handle, PICO_STATUS status, uint32_t noOfSamples,
				int16_t overflow, void * pParameter);
	return ps4000LpDataReadyGo(handle, status, noOfSamples, overflow, pParameter);
}

int ps4000LpBlockReady(int16_t handle, PICO_STATUS status, void * pParameter)
{
	int ps4000LpBlockReadyGo(int16_t handle, PICO_STATUS status, void * pParameter);
	return ps4000LpBlockReadyGo(handle, status,  pParameter);
}

int ps4000LpStreamingReady(int16_t handle, int32_t noOfSamples, uint32_t startIndex,
                int16_t overflow, uint32_t triggerAt, int16_t triggered,
                int16_t autoStop, void * pParameter)
{
	int ps4000LpStreamingReadyGo(int16_t handle, int32_t noOfSamples, uint32_t startIndex,
                int16_t overflow, uint32_t triggerAt, int16_t triggered,
                int16_t autoStop, void * pParameter);
	return ps4000LpStreamingReadyGo(handle,noOfSamples, startIndex,overflow,triggerAt, triggered,autoStop,pParameter);
}
*/
import "C"
