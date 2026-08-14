package control

import (
	"log/slog"
	"math"
)

// DecodeResult represents a decoded byte or token
type DecodeResult struct {
	StartTime float64 // in seconds relative to trigger
	EndTime   float64 // in seconds relative to trigger
	Value     uint16
	Error     bool // Framing error or parity error
}

// DecodeBit represents a single decoded bit
type DecodeBit struct {
	StartTime float64
	EndTime   float64
	Bit       bool
}

// DecoderState maintains state across blocks if necessary
type DecoderState struct {
	Bytes []DecodeResult
	Bits  []DecodeBit
}

// DecodeUART decodes a UART stream from an analog buffer.
func DecodeUART(buffer []int16, samplingTimeInterval float64, triggerTimeOffset int64,
	baudRate int, dataBits int, stopBits string, parity string, bitOrder string, threshold int16,
	hysteresis int32, invert bool) DecoderState {
	slog.Debug("DecodeUART", "samplingTimeInterval", samplingTimeInterval, "triggerTimeOffset", triggerTimeOffset,
		"triggerTimeOffset", triggerTimeOffset, "baudRate", baudRate,
		"dataBits", dataBits, "stopBits", stopBits, "parity", parity, "threshold", threshold,
		"hysteresis", hysteresis, "invert", invert)
	var state DecoderState
	if len(buffer) == 0 || baudRate <= 0 {
		return state
	}

	upper := threshold + int16(hysteresis/2)
	lower := threshold - int16(hysteresis/2)

	digitalBuffer := make([]bool, len(buffer))
	currentState := buffer[0] >= threshold
	for i, val := range buffer {
		if val > upper {
			currentState = true
		} else if val < lower {
			currentState = false
		}
		digitalBuffer[i] = currentState
	}
	getDigital := func(idx int) bool {
		bit := digitalBuffer[idx]
		if invert {
			return !bit
		}
		return bit
	}

	// 2. UART State Machine
	samplesPerBit := (1.0 / float64(baudRate)) / samplingTimeInterval
	if samplesPerBit < 1.0 {
		slog.Debug("DecodeUART", "samplesPerBit<1", samplesPerBit)
		return state
	}

	i := 0
	slog.Debug("DecodeUART", "samplingTimeInterval", samplingTimeInterval, "samplesPerBit", samplesPerBit, "len", len(buffer))
	// Find first start bit
	for i < len(buffer) {
		if !getDigital(i) {
			startOfStartBit := float64(i)
			startBitIn := startOfStartBit
			for startBitIn < float64(len(buffer)) && !getDigital(int(startBitIn)) && (startBitIn-startOfStartBit) < samplesPerBit {
				startBitIn += 1.0
			}
			if (startBitIn - startOfStartBit) >= samplesPerBit*0.75 {
				i = int(math.Round(startOfStartBit))
				slog.Debug("DecodeUART", "start at i", i)
				break
			} else {
				i = int(startBitIn)
				continue
			}
		}
		i++
	}
	for i < len(buffer) {
		// Look for start bit (transition from 1 to 0)
		if !getDigital(i) {
			// Found start bit, sample at middle of bit
			startSample := float64(i)
			midSample := startSample + (samplesPerBit / 2.0)

			if int(midSample) >= len(buffer) {
				break
			}
			if getDigital(int(midSample)) { // false start
				slog.Debug("DecodeUART false start", "i", i)
				i++
				continue
			}
			state.Bits = append(state.Bits, DecodeBit{
				StartTime: (midSample * samplingTimeInterval) + (float64(triggerTimeOffset) / 1e15),
				Bit:       false,
			})

			// Read data bits (LSB first)
			var val uint16
			bitIndex := 0
			sampleIdx := midSample + samplesPerBit

			for bitIndex < dataBits {
				if int(sampleIdx) >= len(buffer) {
					break
				}
				bitVal := getDigital(int(sampleIdx))
				if bitVal {
					if bitOrder == "MSB First" {
						val |= (1 << (dataBits - 1 - bitIndex))
					} else {
						val |= (1 << bitIndex)
					}
				}
				state.Bits = append(state.Bits, DecodeBit{
					StartTime: (sampleIdx * samplingTimeInterval) + (float64(triggerTimeOffset) / 1e15),
					Bit:       bitVal,
				})
				sampleIdx += samplesPerBit
				bitIndex++
			}
			slog.Debug("DecodeUART", "val", val)
			if bitIndex < dataBits {
				slog.Debug("DecodeUART Not enough data")
				break // Not enough data
			}

			parityError := false
			if parity != "None" {
				if int(sampleIdx) >= len(buffer) {
					break
				}
				parityBit := getDigital(int(sampleIdx))
				state.Bits = append(state.Bits, DecodeBit{
					StartTime: (sampleIdx * samplingTimeInterval) + (float64(triggerTimeOffset) / 1e15),
					Bit:       parityBit,
				})
				sampleIdx += samplesPerBit

				ones := 0
				for j := 0; j < dataBits; j++ {
					if (val & (1 << j)) != 0 {
						ones++
					}
				}
				switch parity {
				case "Even":
					if (ones%2 == 0 && parityBit) || (ones%2 != 0 && !parityBit) {
						parityError = true
					}
				case "Odd":
					if (ones%2 == 0 && !parityBit) || (ones%2 != 0 && parityBit) {
						parityError = true
					}
				case "Mark":
					if !parityBit {
						parityError = true
					}
				case "Space":
					if parityBit {
						parityError = true
					}
				}
			}

			// Read Stop bit
			if int(sampleIdx) < len(buffer) {
				stopBit := getDigital(int(sampleIdx))
				state.Bits = append(state.Bits, DecodeBit{
					StartTime: (sampleIdx * samplingTimeInterval) + (float64(triggerTimeOffset) / 1e15),
					Bit:       stopBit,
				})

				timeStart := (float64(i) * samplingTimeInterval) + (float64(triggerTimeOffset) / 1e15)
				timeEnd := (sampleIdx * samplingTimeInterval) + (float64(triggerTimeOffset) / 1e15)

				state.Bytes = append(state.Bytes, DecodeResult{
					StartTime: timeStart,
					EndTime:   timeEnd,
					Value:     val,
					Error:     !stopBit || parityError,
				})
			}

			// Advance past stop bit(s)
			stopBitsLen := 1.0
			if stopBits == "1.5" {
				stopBitsLen = 1.5
			} else if stopBits == "2" {
				stopBitsLen = 2.0
			}
			// sampleIdx is currently in the middle of the first stop bit.
			// Advance to the end of the stop bits, minus a small margin so we 
			// land securely in the idle state before the next start bit, 
			// rather than accidentally overshooting into the next start bit.
			i = int(sampleIdx - (samplesPerBit / 2.0) + stopBitsLen*samplesPerBit)
		} else {
			i++
		}
	}

	return state
}

// DecodeSPI decodes an SPI stream (CPOL=0, CPHA=0) from clock and data analog buffers.
func DecodeSPI(clkBuffer []int16, mosiBuffer []int16, samplingTimeInterval float64,
	triggerTimeOffset int64, threshold int16, hysteresis int32, invert bool) DecoderState {
	var state DecoderState
	if len(clkBuffer) == 0 || len(clkBuffer) != len(mosiBuffer) {
		return state
	}

	getClk := func(idx int) bool {
		val := clkBuffer[idx]
		bit := val >= threshold
		if invert {
			return !bit
		}
		return bit
	}
	getMosi := func(idx int) bool {
		val := mosiBuffer[idx]
		bit := val >= threshold
		if invert {
			return !bit
		}
		return bit
	}

	// 2. SPI State Machine (Sample MOSI on rising edge of CLK)
	var currentByte byte
	bitCount := 0
	var startIdx int

	for i := 1; i < len(clkBuffer); i++ {
		// Detect rising edge on CLK
		if !getClk(i-1) && getClk(i) {
			if bitCount == 0 {
				startIdx = i
			}
			// Sample MOSI
			currentByte <<= 1
			if getMosi(i) {
				currentByte |= 1
			}
			bitCount++

			if bitCount == 8 {
				timeStart := (float64(startIdx) * samplingTimeInterval) + (float64(triggerTimeOffset) / 1e15)
				timeEnd := (float64(i) * samplingTimeInterval) + (float64(triggerTimeOffset) / 1e15)

				state.Bytes = append(state.Bytes, DecodeResult{
					StartTime: timeStart,
					EndTime:   timeEnd,
					Value:     uint16(currentByte),
					Error:     false,
				})
				bitCount = 0
				currentByte = 0
			}
		}
	}

	return state
}
