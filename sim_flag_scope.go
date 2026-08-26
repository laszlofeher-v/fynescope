//go:build !sim

package main

func registerSimFlag() *string {
	s := ""
	return &s
}
