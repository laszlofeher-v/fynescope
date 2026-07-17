// SPDX-License-Identifier: MIT

package scpi

import (
	"strconv"
	"strings"
)

// parseHex parses a hex string (with or without 0x/0X prefix) into a uint16.
// Returns 0 for empty or invalid input.
func parseHex(s string) uint16 {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}
	s = strings.TrimPrefix(s, "0x")
	s = strings.TrimPrefix(s, "0X")
	v, err := strconv.ParseUint(s, 16, 16)
	if err != nil {
		return 0
	}
	return uint16(v)
}
