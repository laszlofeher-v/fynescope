//go:build !demo && ps3000a && !emu

package ps3000a

// #cgo LDFLAGS: -L/opt/picoscope/lib/ -lps3000a
import "C"
