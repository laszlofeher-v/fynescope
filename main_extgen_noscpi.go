//go:build !scpi

package main

func registerExtGenFlag(inTestMode bool) *bool {
	val := false
	return &val
}
