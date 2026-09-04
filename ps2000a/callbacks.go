//go:build !demo

package ps2000a

/*
#cgo linux CFLAGS: -I/opt/picoscope/include/libps2000a
#cgo windows CFLAGS: -I"C:/Program Files/Pico Technology/SDK/inc"
#include <stdio.h>
#include <PicoStatus.h>
#include <ps2000aApi.h>

#ifdef _WIN32
#define CALLBACK_CONV __stdcall
#else
#define CALLBACK_CONV
#endif

// Forward declarations of exported Go functions
void lpDataReadyGo(int16_t handle, PICO_STATUS status, uint32_t noOfSamples,
				int16_t overflow, void * pParameter);
void lpBlockReadyGo(int16_t handle, PICO_STATUS status, void * pParameter);
void lpStreamingReadyGo(int16_t handle, int32_t noOfSamples, uint32_t startIndex,
                int16_t overflow, uint32_t triggerAt, int16_t triggered,
                int16_t autoStop, void * pParameter);

// C callback functions
void CALLBACK_CONV lpDataReady(int16_t handle, PICO_STATUS status, uint32_t noOfSamples,
				int16_t overflow, void * pParameter)
{
	lpDataReadyGo(handle, status, noOfSamples, overflow, pParameter);
}

void CALLBACK_CONV lpBlockReady(int16_t handle, PICO_STATUS status, void * pParameter)
{
	lpBlockReadyGo(handle, status, pParameter);
}

void CALLBACK_CONV lpStreamingReady(int16_t handle, int32_t noOfSamples, uint32_t startIndex,
                int16_t overflow, uint32_t triggerAt, int16_t triggered,
                int16_t autoStop, void * pParameter)
{
	lpStreamingReadyGo(handle, noOfSamples, startIndex, overflow, triggerAt, triggered, autoStop, pParameter);
}
*/
import "C"
