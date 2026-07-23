package gui

import (
	"math"

	"github.com/antonmedv/expr"
	"github.com/antonmedv/expr/vm"
)

type VirtualEnv struct {
	A float64
	B float64
	C float64
	D float64

	// Math functions
	Sin   func(float64) float64
	Cos   func(float64) float64
	Tan   func(float64) float64
	Asin  func(float64) float64
	Acos  func(float64) float64
	Atan  func(float64) float64
	Sinh  func(float64) float64
	Cosh  func(float64) float64
	Tanh  func(float64) float64
	Abs   func(float64) float64
	Sqrt  func(float64) float64
	Pow   func(float64, float64) float64
	Log   func(float64) float64
	Log10 func(float64) float64
	Exp   func(float64) float64
}

type VirtualChannelEngine struct {
	program *vm.Program
	env     *VirtualEnv
}

func CompileVirtualChannel(expression string) (*VirtualChannelEngine, error) {
	env := &VirtualEnv{
		Sin:   math.Sin,
		Cos:   math.Cos,
		Tan:   math.Tan,
		Asin:  math.Asin,
		Acos:  math.Acos,
		Atan:  math.Atan,
		Sinh:  math.Sinh,
		Cosh:  math.Cosh,
		Tanh:  math.Tanh,
		Abs:   math.Abs,
		Sqrt:  math.Sqrt,
		Pow:   math.Pow,
		Log:   math.Log,
		Log10: math.Log10,
		Exp:   math.Exp,
	}
	program, err := expr.Compile(expression, expr.Env(env))
	if err != nil {
		return nil, err
	}
	return &VirtualChannelEngine{
		program: program,
		env:     env,
	}, nil
}

func (e *VirtualChannelEngine) Evaluate(A, B, C, D float32) float32 {
	e.env.A = float64(A) / 1000.0
	e.env.B = float64(B) / 1000.0
	e.env.C = float64(C) / 1000.0
	e.env.D = float64(D) / 1000.0

	out, err := expr.Run(e.program, e.env)
	if err != nil {
		return 0
	}

	var res float32
	switch v := out.(type) {
	case float64:
		res = float32(v)
	case float32:
		res = v
	case int:
		res = float32(v)
	case int64:
		res = float32(v)
	default:
		res = 0
	}
	return res * 1000.0
}

// EvaluateBuffer evaluates the virtual channel for a full buffer length.
// It takes slices of physical channel buffers.
func (e *VirtualChannelEngine) EvaluateBuffer(dest []float32, A, B, C, D []float32, size int) {
	for i := 0; i < size; i++ {
		var a, b, c, d float32
		if len(A) > i {
			a = A[i]
		}
		if len(B) > i {
			b = B[i]
		}
		if len(C) > i {
			c = C[i]
		}
		if len(D) > i {
			d = D[i]
		}

		dest[i] = e.Evaluate(a, b, c, d)
	}
}
