//go:build sim

package main

import "flag"

func registerSimFlag() *string {
	return flag.String("sim", "", "-sim=NAME (use simulator with scope name instead of demo mode, name is extended with SIM)")
}
