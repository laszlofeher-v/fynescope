//go:build !sim

package main

func registerSimFlag() *bool {
	b := false
	return &b
}
