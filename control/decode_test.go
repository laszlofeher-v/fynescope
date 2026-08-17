package control

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func generateSamples(bits []bool, samplesPerBit int, highVal int16, lowVal int16) []int16 {
	buffer := make([]int16, 0, len(bits)*samplesPerBit)
	for _, bit := range bits {
		val := lowVal
		if bit {
			val = highVal
		}
		for i := 0; i < samplesPerBit; i++ {
			buffer = append(buffer, val)
		}
	}
	return buffer
}

func TestDecodeUART(t *testing.T) {
	baudRate := 9600
	samplesPerBit := 10
	samplingTimeInterval := 1.0 / (float64(baudRate) * float64(samplesPerBit))
	
	// Construct 8N1 UART frame for value 0x55 (01010101) LSB first
	// Idle = 1, Start = 0, D0=1, D1=0, D2=1, D3=0, D4=1, D5=0, D6=1, D7=0, Stop = 1, Idle = 1
	bits := []bool{
		true, true, // Idle
		false,      // Start
		true, false, true, false, true, false, true, false, // Data (0x55)
		true,       // Stop
		true, true, // Idle
	}
	
	buffer := generateSamples(bits, samplesPerBit, 100, -100)
	
	state := DecodeUART(buffer, samplingTimeInterval, 0, baudRate, 8, "1", "None", "LSB First", 0, 10, false)
	
	assert.Len(t, state.Bytes, 1)
	if len(state.Bytes) == 1 {
		assert.Equal(t, uint16(0x55), state.Bytes[0].Value)
		assert.False(t, state.Bytes[0].Error)
	}
	assert.True(t, len(state.Bits) >= 9) // 8 data + possibly stop/parity depending on logic
}

func TestDecodeUART_Parity(t *testing.T) {
	baudRate := 9600
	samplesPerBit := 10
	samplingTimeInterval := 1.0 / (float64(baudRate) * float64(samplesPerBit))
	
	// Construct 8E1 UART frame for value 0x55 (01010101) - 4 ones, so Even parity bit = 0
	bits := []bool{
		true, true, // Idle
		false,      // Start
		true, false, true, false, true, false, true, false, // Data (0x55)
		false,      // Parity (Even)
		true,       // Stop
		true, true, // Idle
	}
	
	buffer := generateSamples(bits, samplesPerBit, 100, -100)
	
	state := DecodeUART(buffer, samplingTimeInterval, 0, baudRate, 8, "1", "Even", "LSB First", 0, 10, false)
	
	assert.Len(t, state.Bytes, 1)
	if len(state.Bytes) == 1 {
		assert.Equal(t, uint16(0x55), state.Bytes[0].Value)
		assert.False(t, state.Bytes[0].Error)
	}
}

func TestDecodeSPI(t *testing.T) {
	// 1 byte transmission, 0xA3 (10100011) MSB first, CPOL=0, CPHA=0
	// CPOL=0 means clock idle low. Data sampled on rising edge.
	
	// Let's create an analog buffer manually.
	// We'll use 4 samples per clock cycle (2 low, 2 high)
	
	val := byte(0xA3) // 10100011
	clkBuffer := make([]int16, 0, 8*4+8)
	mosiBuffer := make([]int16, 0, 8*4+8)
	
	// Idle
	for i := 0; i < 4; i++ {
		clkBuffer = append(clkBuffer, -100)
		mosiBuffer = append(mosiBuffer, -100)
	}
	
	for i := 7; i >= 0; i-- {
		bit := (val >> i) & 1
		mosiVal := int16(-100)
		if bit == 1 {
			mosiVal = 100
		}
		
		// Clock Low (MOSI set up)
		clkBuffer = append(clkBuffer, -100)
		mosiBuffer = append(mosiBuffer, mosiVal)
		clkBuffer = append(clkBuffer, -100)
		mosiBuffer = append(mosiBuffer, mosiVal)
		
		// Clock High (Sampled on rising edge)
		clkBuffer = append(clkBuffer, 100)
		mosiBuffer = append(mosiBuffer, mosiVal)
		clkBuffer = append(clkBuffer, 100)
		mosiBuffer = append(mosiBuffer, mosiVal)
	}
	
	// Idle
	for i := 0; i < 4; i++ {
		clkBuffer = append(clkBuffer, -100)
		mosiBuffer = append(mosiBuffer, -100)
	}
	
	state := DecodeSPI(clkBuffer, mosiBuffer, 1e-6, 0, 0, 10, false)
	
	assert.Len(t, state.Bytes, 1)
	if len(state.Bytes) == 1 {
		assert.Equal(t, uint16(0xA3), state.Bytes[0].Value)
		assert.False(t, state.Bytes[0].Error)
	}
}
