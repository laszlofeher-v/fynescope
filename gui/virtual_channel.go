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
}

// Math functions provided to the expressions
func (VirtualEnv) Sin(x float64) float64    { return math.Sin(x) }
func (VirtualEnv) Cos(x float64) float64    { return math.Cos(x) }
func (VirtualEnv) Tan(x float64) float64    { return math.Tan(x) }
func (VirtualEnv) Asin(x float64) float64   { return math.Asin(x) }
func (VirtualEnv) Acos(x float64) float64   { return math.Acos(x) }
func (VirtualEnv) Atan(x float64) float64   { return math.Atan(x) }
func (VirtualEnv) Sinh(x float64) float64   { return math.Sinh(x) }
func (VirtualEnv) Cosh(x float64) float64   { return math.Cosh(x) }
func (VirtualEnv) Tanh(x float64) float64   { return math.Tanh(x) }
func (VirtualEnv) Abs(x float64) float64    { return math.Abs(x) }
func (VirtualEnv) Sqrt(x float64) float64   { return math.Sqrt(x) }
func (VirtualEnv) Pow(x, y float64) float64 { return math.Pow(x, y) }
func (VirtualEnv) Log(x float64) float64    { return math.Log(x) }
func (VirtualEnv) Log10(x float64) float64  { return math.Log10(x) }
func (VirtualEnv) Exp(x float64) float64    { return math.Exp(x) }

type VirtualChannelEngine struct {
	program *vm.Program
	env     *VirtualEnv
}

func CompileVirtualChannel(expression string) (*VirtualChannelEngine, error) {
	env := &VirtualEnv{}
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
	e.env.A = float64(A)
	e.env.B = float64(B)
	e.env.C = float64(C)
	e.env.D = float64(D)

	out, err := expr.Run(e.program, e.env)
	if err != nil {
		return 0
	}

	switch v := out.(type) {
	case float64:
		return float32(v)
	case float32:
		return v
	case int:
		return float32(v)
	case int64:
		return float32(v)
	default:
		return 0
	}
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
