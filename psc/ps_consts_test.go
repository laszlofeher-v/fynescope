package psc

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStatStr(t *testing.T) {
	assert.Contains(t, StatStr(PICO_OK), "PICO_OK")
	assert.Contains(t, StatStr(PICO_NOT_FOUND), "PICO_NOT_FOUND")
	assert.Contains(t, StatStr(PICO_INVALID_HANDLE), "PICO_INVALID_HANDLE")
	assert.Equal(t, 0, PICO_OK)
	assert.Equal(t, 0, PicoOk)
	assert.True(t, strings.HasPrefix(StatStr(PICO_INVALID_TIMEBASE), "PICO_INVALID_TIMEBASE"))
}
