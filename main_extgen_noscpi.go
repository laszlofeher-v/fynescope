//go:build demo || !scpi

package main

func registerExtGenFlag(inTestMode bool) *bool {
	val := false
	return &val
}
