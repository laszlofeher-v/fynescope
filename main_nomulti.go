//go:build !multi

package main

func registerMultiFlag() *bool {
	val := false
	return &val
}
