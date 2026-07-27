//go:build !noscope && ps3000

package ps3000

/*
#include <stdio.h>
#include <stdint.h>
// C callback function
void lpStreamingReady3000(
  int16_t **overviewBuffers,
  int16_t   overflow,
  uint32_t  triggeredAt,
  int16_t   triggered,
  int16_t   auto_stop,
  uint32_t  nValues
)
{
	void lpStreamingReadyGo3000(int16_t **overviewBuffers, int16_t overflow, uint32_t triggeredAt, int16_t triggered, int16_t auto_stop, uint32_t nValues);
	lpStreamingReadyGo3000(overviewBuffers, overflow, triggeredAt, triggered, auto_stop, nValues);
}
*/
import "C"
