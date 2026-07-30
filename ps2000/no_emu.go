//go:build !noscope && ps2000 && !emu

package ps2000

// #cgo LDFLAGS: -L/opt/picoscope/lib/ -lps2000
import "C"
