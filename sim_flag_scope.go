//go:build !demo

package main

func registerSimFlag() *bool {
	b := false
	return &b
}
