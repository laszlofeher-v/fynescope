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
	
	valMOSI := byte(0xA3) // 10100011
	valMISO := byte(0x5A) // 01011010
	clkBuffer := make([]int16, 0, 8*4+8)
	mosiBuffer := make([]int16, 0, 8*4+8)
	misoBuffer := make([]int16, 0, 8*4+8)
	csBuffer := make([]int16, 0, 8*4+8)
	
	// Idle (CS high/inactive)
	for i := 0; i < 4; i++ {
		clkBuffer = append(clkBuffer, -100)
		mosiBuffer = append(mosiBuffer, -100)
		misoBuffer = append(misoBuffer, -100)
		csBuffer = append(csBuffer, 100)
	}
	
	for i := 7; i >= 0; i-- {
		bitMOSI := (valMOSI >> i) & 1
		mosiVal := int16(-100)
		if bitMOSI == 1 {
			mosiVal = 100
		}

		bitMISO := (valMISO >> i) & 1
		misoVal := int16(-100)
		if bitMISO == 1 {
			misoVal = 100
		}
		
		// Clock Low (MOSI set up)
		clkBuffer = append(clkBuffer, -100)
		mosiBuffer = append(mosiBuffer, mosiVal)
		misoBuffer = append(misoBuffer, misoVal)
		csBuffer = append(csBuffer, -100)
		clkBuffer = append(clkBuffer, -100)
		mosiBuffer = append(mosiBuffer, mosiVal)
		misoBuffer = append(misoBuffer, misoVal)
		csBuffer = append(csBuffer, -100)
		
		// Clock High (Sampled on rising edge)
		clkBuffer = append(clkBuffer, 100)
		mosiBuffer = append(mosiBuffer, mosiVal)
		misoBuffer = append(misoBuffer, misoVal)
		csBuffer = append(csBuffer, -100)
		clkBuffer = append(clkBuffer, 100)
		mosiBuffer = append(mosiBuffer, mosiVal)
		misoBuffer = append(misoBuffer, misoVal)
		csBuffer = append(csBuffer, -100)
	}
	
	// Idle
	for i := 0; i < 4; i++ {
		clkBuffer = append(clkBuffer, -100)
		mosiBuffer = append(mosiBuffer, -100)
		misoBuffer = append(misoBuffer, -100)
		csBuffer = append(csBuffer, 100)
	}
	
	state := DecodeSPI(clkBuffer, mosiBuffer, misoBuffer, csBuffer, false, 1e-6, 0, 0, 10, false)
	
	assert.Len(t, state.Bytes, 1)
	if len(state.Bytes) == 1 {
		assert.Equal(t, uint16(0xA3), state.Bytes[0].Value)
		assert.Equal(t, uint16(0x5A), state.Bytes[0].Value2)
		assert.True(t, state.Bytes[0].HasValue2)
		assert.False(t, state.Bytes[0].Error)
	}

	// Test without MISO and without CS
	stateNoMisoCs := DecodeSPI(clkBuffer, mosiBuffer, nil, nil, false, 1e-6, 0, 0, 10, false)
	assert.Len(t, stateNoMisoCs.Bytes, 1)
	if len(stateNoMisoCs.Bytes) == 1 {
		assert.Equal(t, uint16(0xA3), stateNoMisoCs.Bytes[0].Value)
		assert.False(t, stateNoMisoCs.Bytes[0].HasValue2)
	}
}

func TestDecodeI2C(t *testing.T) {
	sclBuffer := make([]int16, 200)
	sdaBuffer := make([]int16, 200)

	// idle high
	for i := range sclBuffer {
		sclBuffer[i] = 1000
		sdaBuffer[i] = 1000
	}

	// START: SCL high, SDA falls
	sdaBuffer[10] = 0

	// Write Addr byte: 0x50 (write, so byte is 0xA0)
	// Addr byte = 10100000 -> bit7=1, bit6=0, bit5=1, bit4=0, bit3=0, bit2=0, bit1=0, bit0=0
	bits := []int{1, 0, 1, 0, 0, 0, 0, 0}
	
	idx := 15
	for _, bit := range bits {
		sclBuffer[idx] = 0 // scl low
		if bit == 1 {
			sdaBuffer[idx] = 1000
		} else {
			sdaBuffer[idx] = 0
		}
		idx += 2
		sclBuffer[idx] = 1000 // scl high (sample)
		if bit == 1 {
			sdaBuffer[idx] = 1000
		} else {
			sdaBuffer[idx] = 0
		}
		idx += 2
	}

	// ACK
	sclBuffer[idx] = 0
	sdaBuffer[idx] = 0 // ACK
	idx += 2
	sclBuffer[idx] = 1000
	sdaBuffer[idx] = 0
	idx += 2

	// Data byte: 0x5A -> 01011010
	dataBits := []int{0, 1, 0, 1, 1, 0, 1, 0}
	for _, bit := range dataBits {
		sclBuffer[idx] = 0 // scl low
		if bit == 1 {
			sdaBuffer[idx] = 1000
		} else {
			sdaBuffer[idx] = 0
		}
		idx += 2
		sclBuffer[idx] = 1000 // scl high (sample)
		if bit == 1 {
			sdaBuffer[idx] = 1000
		} else {
			sdaBuffer[idx] = 0
		}
		idx += 2
	}

	// ACK
	sclBuffer[idx] = 0
	sdaBuffer[idx] = 0 // ACK
	idx += 2
	sclBuffer[idx] = 1000
	sdaBuffer[idx] = 0
	idx += 2

	// STOP: scl high, sda rises
	sclBuffer[idx] = 0
	sdaBuffer[idx] = 0
	idx += 2
	sclBuffer[idx] = 1000
	sdaBuffer[idx] = 0
	idx += 2
	sclBuffer[idx] = 1000
	sdaBuffer[idx] = 1000

	state := DecodeI2C(sclBuffer, sdaBuffer, 1e-6, 0, 500, 100, false)
	assert.Len(t, state.Bytes, 2)
	if len(state.Bytes) == 2 {
		assert.Equal(t, uint16(0xA0), state.Bytes[0].Value)
		assert.Equal(t, "A:A0(W)", state.Bytes[0].Label)
		assert.False(t, state.Bytes[0].Error)

		assert.Equal(t, uint16(0x5A), state.Bytes[1].Value)
		assert.Equal(t, "D:5A", state.Bytes[1].Label)
		assert.False(t, state.Bytes[1].Error)
	}
}
