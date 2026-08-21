//go:build !demo && !sim

package main

func registerSimFlag() *bool {
	b := false
	return &b
}
