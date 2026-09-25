//go:build multi

package main

import "flag"

func registerMultiFlag() *bool {
	return flag.Bool("multi", false, "-multi=true (enables multiscope support)")
}
