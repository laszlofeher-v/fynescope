//go:build !noscope && ps2000 && emu

package ps2000

// #cgo LDFLAGS: -L. -lemu2000
import "C"
