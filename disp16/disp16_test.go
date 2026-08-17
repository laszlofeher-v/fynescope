package disp16

import (
	"image/color"
	"testing"

	"fyne.io/fyne/v2/test"
	"github.com/stretchr/testify/assert"
)

func TestNewHexArray(t *testing.T) {
	test.NewApp()
	
	_, err := NewHexArray(0, nil, color.White, false, "Test")
	assert.Error(t, err)

	disp, err := NewHexArray(4, nil, color.White, false, "Test")
	assert.NoError(t, err)
	assert.NotNil(t, disp)
	assert.Equal(t, 4, disp.numOfDigits)
	assert.Equal(t, 4, len(disp.digits))
	assert.Equal(t, uint64(0), disp.GetValue())
}

func TestSetValue(t *testing.T) {
	test.NewApp()
	disp, _ := NewHexArray(2, nil, color.White, false, "")
	
	// Value within bounds (2 digits = max 0xFF)
	disp.SetValue(0x42)
	assert.Equal(t, uint64(0x42), disp.GetValue())
	assert.Equal(t, 4, disp.digits[0].val)
	assert.Equal(t, 2, disp.digits[1].val)

	// Value out of bounds (should truncate/clamp to maxVal which is 0xFF)
	disp.SetValue(0x100)
	assert.Equal(t, uint64(0xFF), disp.GetValue())
	assert.Equal(t, 0xF, disp.digits[0].val)
	assert.Equal(t, 0xF, disp.digits[1].val)
}

func TestOnChanged(t *testing.T) {
	test.NewApp()
	disp, _ := NewHexArray(4, nil, color.White, false, "")
	
	var changedVal uint64
	disp.OnChanged = func(v uint64) {
		changedVal = v
	}

	disp.SetValue(0x1234)
	assert.Equal(t, uint64(0x1234), changedVal)
}
