package gui

import (
	"math"
	"testing"
	"github.com/antonmedv/expr"
)

func TestExprNestedSin(t *testing.T) {
	env := &VirtualEnv{A: 1.0}
	program, err := expr.Compile("Sin(Sin(A))", expr.Env(env))
	if err != nil {
		t.Fatalf("Compile error: %v", err)
	}
	
	output, err := expr.Run(program, env)
	if err != nil {
		t.Fatalf("Run error: %v", err)
	}
	
	expected := math.Sin(math.Sin(1.0))
	if output != expected {
		t.Fatalf("Expected %v, got %v", expected, output)
	}
}
