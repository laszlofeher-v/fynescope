//go:build !noscope && ps3000a && emu

package ps3000a

// #cgo LDFLAGS: -L. -lemu3000a
import "C"
