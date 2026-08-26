//go:build !demo && ps2000 && !emu && !sim

package ps2000

/*
#cgo LDFLAGS: -L/opt/picoscope/lib/ -lps2000
#include <stdint.h>

// Dummy types for functions not supported by PS2000 but needed for emu compatibility
typedef struct tPS2000_DIGITAL_CHANNEL_DIRECTIONS {
    int16_t channel;
    int16_t direction;
} PS2000_DIGITAL_CHANNEL_DIRECTIONS;
typedef int16_t PS2000_DIGITAL_PORT;

int16_t ps2000SetTriggerDigitalPortProperties(int16_t handle, PS2000_DIGITAL_CHANNEL_DIRECTIONS *directions, int16_t nDirections) {
    return 0; // Not supported
}

int16_t ps2000SetDigitalPort(int16_t handle, PS2000_DIGITAL_PORT port, int16_t enabled, int16_t logicLevel) {
    return 0; // Not supported
}
*/
import "C"
