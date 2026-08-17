package gui

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCompileVirtualChannel(t *testing.T) {
	// Valid expression
	engine, err := CompileVirtualChannel("A + B")
	assert.NoError(t, err)
	assert.NotNil(t, engine)

	// Invalid expression
	engine, err = CompileVirtualChannel("A + ")
	assert.Error(t, err)
	assert.Nil(t, engine)
}

func TestVirtualChannelEngine_Evaluate(t *testing.T) {
	engine, _ := CompileVirtualChannel("A + B * 2")

	// Evaluate takes inputs in mV (so 1000 == 1V). The output is also in mV.
	// Internally, A = 1000/1000 = 1, B = 2000/1000 = 2.
	// 1 + 2 * 2 = 5V = 5000mV.
	res := engine.Evaluate(1000, 2000, 0, 0)
	assert.Equal(t, float32(5000), res)

	// Math functions
	engine2, _ := CompileVirtualChannel("Sin(A)")
	// 0 -> 0
	res = engine2.Evaluate(0, 0, 0, 0)
	assert.Equal(t, float32(0), res)

	// pi / 2 = 1.5707963V -> Sin(pi/2) = 1V = 1000mV
	// 1.5707963V is 1570.7963mV
	res = engine2.Evaluate(float32(math.Pi/2*1000), 0, 0, 0)
	assert.InDelta(t, float32(1000), res, 0.1)

	// Division by zero (Inf) handling
	engine3, _ := CompileVirtualChannel("A / B")
	res = engine3.Evaluate(1000, 0, 0, 0) // Should be Inf, which Evaluate maps to 0
	assert.Equal(t, float32(0), res)
}

func TestVirtualChannelEngine_EvaluateBuffer(t *testing.T) {
	engine, _ := CompileVirtualChannel("A + B")

	A := []float32{1000, 2000, 3000}
	B := []float32{500, 1500} // shorter buffer to test bounds checking
	C := []float32{}
	D := []float32{}

	dest := make([]float32, 3)

	engine.EvaluateBuffer(dest, A, B, C, D, 3)

	assert.Equal(t, float32(1500), dest[0])
	assert.Equal(t, float32(3500), dest[1])
	assert.Equal(t, float32(3000), dest[2]) // B has no 3rd element, so it defaults to 0
}
