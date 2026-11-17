//go:build !sim

package main

const isSimBuild = false

func registerSimFlag() *string {
	s := ""
	return &s
}
