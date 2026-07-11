//go:build scpi

package main

import "flag"

func registerExtGenFlag(inTestMode bool) *bool {
	return flag.Bool("extgen", inTestMode, "enable external generator (-extgen=true/false)")
}
