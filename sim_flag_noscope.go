//go:build demo

package main

import "flag"

func registerSimFlag() *bool {
	return flag.Bool("sim", false, "-sim=true (use ps2000a simulator instead of demo mode)")
}
