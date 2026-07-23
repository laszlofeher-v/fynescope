package gui

import (
	"math"
	"testing"
)

func TestExprNestedSin(t *testing.T) {
	eng, err := CompileVirtualChannel("Sin(Sin(A))")
	if err != nil {
		t.Fatalf("Compile error: %v", err)
	}

	output := eng.Evaluate(1000.0, 0, 0, 0)

	expected := float32(math.Sin(math.Sin(1.0)) * 1000.0)
	if math.Abs(float64(output-expected)) > 1e-4 {
		t.Fatalf("Expected %v, got %v", expected, output)
	}
}
