package disp7

import (
	"image/color"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
	"github.com/stretchr/testify/assert"
)

const (
	segmentA = iota
	segmentB
	segmentC
	segmentD
	segmentE
	segmentF
	segmentG
)

func assertDpSegment(t *testing.T, disp *DigitArray) {
}

func assertSegments(t *testing.T, disp *DigitArray) {
	if disp.dpPos > 0 {
		assertDpSegment(t, disp)
	}
	for i := range disp.digits {
		d := disp.digits[i]
		switch d.val {
		case 0:
			assert.True(t, d.segmentStates[segmentA])
			assert.True(t, d.segmentStates[segmentB])
			assert.True(t, d.segmentStates[segmentC])
			assert.True(t, d.segmentStates[segmentD])
			assert.True(t, d.segmentStates[segmentE])
			assert.True(t, d.segmentStates[segmentF])
		case 1:
			assert.True(t, d.segmentStates[segmentB])
			assert.True(t, d.segmentStates[segmentC])
		case 2:
			assert.True(t, d.segmentStates[segmentA])
			assert.True(t, d.segmentStates[segmentB])
			assert.True(t, d.segmentStates[segmentG])
			assert.True(t, d.segmentStates[segmentE])
			assert.True(t, d.segmentStates[segmentD])
		case 3:
			assert.True(t, d.segmentStates[segmentA])
			assert.True(t, d.segmentStates[segmentB])
			assert.True(t, d.segmentStates[segmentG])
			assert.True(t, d.segmentStates[segmentC])
			assert.True(t, d.segmentStates[segmentD])
		case 4:
			assert.True(t, d.segmentStates[segmentF])
			assert.True(t, d.segmentStates[segmentG])
			assert.True(t, d.segmentStates[segmentB])
			assert.True(t, d.segmentStates[segmentC])
		case 5:
			assert.True(t, d.segmentStates[segmentA])
			assert.True(t, d.segmentStates[segmentF])
			assert.True(t, d.segmentStates[segmentG])
			assert.True(t, d.segmentStates[segmentC])
			assert.True(t, d.segmentStates[segmentD])
		case 6:
			assert.True(t, d.segmentStates[segmentA])
			assert.True(t, d.segmentStates[segmentG])
			assert.True(t, d.segmentStates[segmentC])
			assert.True(t, d.segmentStates[segmentD])
			assert.True(t, d.segmentStates[segmentE])
			assert.True(t, d.segmentStates[segmentF])
		case 7:
			assert.True(t, d.segmentStates[segmentA])
			assert.True(t, d.segmentStates[segmentB])
			assert.True(t, d.segmentStates[segmentC])
		case 8:
			assert.True(t, d.segmentStates[segmentA])
			assert.True(t, d.segmentStates[segmentB])
			assert.True(t, d.segmentStates[segmentC])
			assert.True(t, d.segmentStates[segmentD])
			assert.True(t, d.segmentStates[segmentE])
			assert.True(t, d.segmentStates[segmentF])
			assert.True(t, d.segmentStates[segmentG])
		case 9:
			assert.True(t, d.segmentStates[segmentA])
			assert.True(t, d.segmentStates[segmentB])
			assert.True(t, d.segmentStates[segmentC])
			assert.True(t, d.segmentStates[segmentD])
			assert.True(t, d.segmentStates[segmentF])
			assert.True(t, d.segmentStates[segmentG])
		}
	}
}

// Helper function to create a default display for testing
func createTestDisplay(t *testing.T, numOfDigits, dpPos, max, min int, signed signType) *DigitArray {
	// We need a Fyne app and window for some widget functionalities,
	// even if we don't render visually in tests.
	// test.NewApp() creates a headless app suitable for testing.
	// test.NewWindow(nil) creates a window without content initially.
	// Note: Direct interaction testing (mouse/key) might still be limited.
	a := test.NewApp()
	w := a.NewWindow("Test")
	disp, err := NewCustomDisp7Array(numOfDigits, dpPos, max, min, signed, TrailingZeroes, w, color.NRGBA{R: 255, A: 255}, ReadWrite, DefaultDigitWidth, DeafultDigitHeight, DefaultSkew, DefaultVCursorSpace, "TestLabel", "TU")
	assert.NoError(t, err, "Failed to create test display")
	assert.NotNil(t, disp, "Created display should not be nil")
	return disp
}

func TestNewCustomDisp7Array(t *testing.T) {
	a := test.NewApp()
	w := a.NewWindow("Test")

	tests := []struct {
		name                string
		numOfDigits         int
		numOfFractionDigits int
		maxValue            int
		minValue            int
		signed              signType
		trailingZeroes      trailingZeroesType
		readOnly            accessModeType
		label               string
		unit                string
		expectError         bool
		checkFunc           func(t *testing.T, disp *DigitArray) // Optional check function
	}{
		{
			name:                "Valid Unsigned",
			numOfDigits:         5,
			numOfFractionDigits: 2,
			maxValue:            99999,
			minValue:            0,
			signed:              UnSigned,
			trailingZeroes:      TrailingZeroes,
			readOnly:            ReadWrite,
			label:               "Value",
			unit:                "V",
			expectError:         false,
			checkFunc: func(t *testing.T, disp *DigitArray) {
				assert.Equal(t, 5, len(disp.digits))
				assert.Equal(t, 2, disp.dpPos)
				assert.Equal(t, 99999, disp.maxValue)
				assert.Equal(t, 0, disp.minValue)
				assert.Equal(t, UnSigned, disp.signed)
				assert.True(t, disp.trailingZeroesOn)
				assert.False(t, disp.Readonly)
				assert.Equal(t, "Value", disp.label.Text)
				assert.Equal(t, "V", disp.unit.Text)
			},
		},
		{
			name:                "Valid Signed",
			numOfDigits:         4,
			numOfFractionDigits: 0,
			maxValue:            9999,
			minValue:            -9999,
			signed:              Signed,
			trailingZeroes:      NoTrailingZeroes,
			readOnly:            ReaOnly,
			label:               "",
			unit:                "",
			expectError:         false,
			checkFunc: func(t *testing.T, disp *DigitArray) {
				assert.Equal(t, 4, len(disp.digits))
				assert.Equal(t, 0, disp.dpPos)
				assert.Equal(t, 9999, disp.maxValue)
				assert.Equal(t, -9999, disp.minValue)
				assert.Equal(t, Signed, disp.signed)
				assert.False(t, disp.trailingZeroesOn)
				assert.True(t, disp.Readonly)
			},
		},
		{
			name:        "Invalid NumOfDigits",
			numOfDigits: 0, // Invalid
			expectError: true,
		},
		{
			name:        "Invalid Unsigned MinValue",
			numOfDigits: 3,
			minValue:    -10, // Invalid for Unsigned
			signed:      UnSigned,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			disp, err := NewCustomDisp7Array(tt.numOfDigits, tt.numOfFractionDigits, tt.maxValue,
				tt.minValue, tt.signed, tt.trailingZeroes, w, color.White,
				tt.readOnly, DefaultDigitWidth, DeafultDigitHeight, DefaultSkew, DefaultVCursorSpace,
				tt.label, tt.unit)

			if tt.expectError {
				assert.Error(t, err, "Expected an error but got none")
				assert.Nil(t, disp, "Display should be nil on error")
			} else {
				assert.NoError(t, err, "Did not expect an error but got one")
				assert.NotNil(t, disp, "Expected a valid display object")
				if tt.checkFunc != nil {
					tt.checkFunc(t, disp)
				}
			}
		})
	}
}

func TestNewDisp7Array(t *testing.T) {
	a := test.NewApp()
	w := a.NewWindow("Test")
	disp, err := NewDisp7Array(4, 1, 1000, -1000, Signed, w, color.Black, ReaOnly)
	assert.NoError(t, err)
	assert.NotNil(t, disp)
	assert.Equal(t, 4, len(disp.digits))
	assert.Equal(t, 1, disp.dpPos)
	assert.Equal(t, 1000, disp.maxValue)
	assert.Equal(t, -1000, disp.minValue)
	assert.Equal(t, Signed, disp.signed)
	assert.True(t, disp.Readonly)
	// Check defaults from NewCustomDisp7Array that NewDisp7Array uses
	assert.True(t, disp.trailingZeroesOn)
	assert.Equal(t, DefaultDigitWidth, disp.digitWidth)
	assert.Equal(t, DeafultDigitHeight, disp.size.Height)
	assert.Equal(t, DefaultSkew, disp.skew)
	assert.Equal(t, DefaultVCursorSpace, disp.cursorVSpace)
	assert.Equal(t, "", disp.label.Text)
	assert.Equal(t, "", disp.unit.Text)
}

func TestSetValue(t *testing.T) {
	disp := createTestDisplay(t, 5, 0, 99999, -99999, Signed)
	changed := false
	disp.OnChanged = func(v float64) {
		changed = true
	}

	tests := []struct {
		name          string
		valueToSet    int
		expectedValue int
		expectChange  bool
	}{
		{"Set Positive", 12345, 12345, true},
		{"Set Negative", -5432, -5432, true},
		{"Set Zero", 0, 0, true},
		{"Set Max Value", 99999, 99999, true},
		{"Set Min Value", -99999, -99999, true},
		{"Set Above Max", 100000, 99999, true},   // Should clamp
		{"Set Below Min", -100000, -99999, true}, // Should clamp
		{"Set Same Value", -99999, -99999, true}, // OnChanged might still trigger depending on exact SetValue logic
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			changed = false // Reset flag
			disp.SetValue(tt.valueToSet)
			assert.Equal(t, tt.expectedValue, disp.Value, "Value mismatch")
			assert.Equal(t, tt.expectChange, changed, "OnChanged trigger mismatch")

			// Verify internal digits representation for a simple case
			if tt.valueToSet == 12345 && tt.expectedValue == 12345 {
				assert.Equal(t, 5, disp.digits[0].val) // LSB
				assert.Equal(t, 4, disp.digits[1].val)
				assert.Equal(t, 3, disp.digits[2].val)
				assert.Equal(t, 2, disp.digits[3].val)
				assert.Equal(t, 1, disp.digits[4].val) // MSB
			}
			if tt.valueToSet == -5432 && tt.expectedValue == -5432 {
				assert.Equal(t, 2, disp.digits[0].val) // LSB
				assert.Equal(t, 3, disp.digits[1].val)
				assert.Equal(t, 4, disp.digits[2].val)
				assert.Equal(t, 5, disp.digits[3].val)
				assert.Equal(t, 0, disp.digits[4].val) // MSB (sign handled separately)
			}
		})
	}
}

func TestSilentSetValue(t *testing.T) {
	disp := createTestDisplay(t, 4, 0, 1000, -1000, Signed)
	changed := false
	disp.OnChanged = func(v float64) {
		changed = true
	}

	disp.SilentSetValue(500)
	assert.Equal(t, 500, disp.Value)
	assert.False(t, changed, "OnChanged should not trigger on SilentSetValue")

	disp.SilentSetValue(-2000) // Below min
	assert.Equal(t, -1000, disp.Value)
	assert.False(t, changed)

	disp.SilentSetValue(2000) // Above max
	assert.Equal(t, 1000, disp.Value)
	assert.False(t, changed)
}

func TestSetFloatValue(t *testing.T) {
	// 5 digits total, 2 fraction digits. Range: -999.99 to 999.99 (implies internal range -99999 to 99999)
	disp := createTestDisplay(t, 5, 2, 99999, -99999, Signed)
	changed := false
	disp.OnChanged = func(v float64) {
		changed = true
	}

	tests := []struct {
		name          string
		valueToSet    float64
		dpPosToSet    int
		expectedValue int // Expected internal integer value
		expectedDpPos int
		expectChange  bool
		expectError   bool // SetFloatValue doesn't return error, but checks bounds
	}{
		{"Set Positive Float", 123.45, 2, 12345, 2, true, false},
		{"Set Negative Float", -98.76, 2, -9876, 2, true, false},
		{"Set Float needing rounding", 12.345, 2, 1235, 2, true, false}, // Rounds up
		{"Set Float needing rounding down", 12.344, 2, 1234, 2, true, false},
		{"Set Integer as Float", 500.0, 2, 50000, 2, true, false},
		{"Set Float with different dpPos", 67.8, 1, 678, 1, true, false}, // Change dpPos
		{"Set Float Above Max", 1000.00, 2, 99999, 2, true, false},       // Clamps
		{"Set Float Below Min", -1000.00, 2, -99999, 2, true, false},     // Clamps
		{"Invalid dpPos (negative)", 10.0, -1, -99999, 2, false, true},   // Should not change value or dpPos
		{"Invalid dpPos (too large)", 10.0, 5, -99999, 2, false, true},   // Should not change value or dpPos
	}

	initialValue := disp.Value
	initialDpPos := disp.dpPos

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			changed = false // Reset flag
			currentValue := disp.Value
			currentDpPos := disp.dpPos

			disp.SetFloatValue(tt.valueToSet, tt.dpPosToSet)

			if tt.expectError {
				// If an error condition is met, value and dpPos shouldn't change
				assert.Equal(t, currentValue, disp.Value, "Value should not change on invalid input")
				assert.Equal(t, currentDpPos, disp.dpPos, "dpPos should not change on invalid input")
				assert.False(t, changed, "OnChanged should not trigger on invalid input")
			} else {
				assert.Equal(t, tt.expectedValue, disp.Value, "Internal value mismatch")
				assert.Equal(t, tt.expectedDpPos, disp.dpPos, "dpPos mismatch")
				assert.Equal(t, tt.expectChange, changed, "OnChanged trigger mismatch")
			}
		})
	}
	// Restore initial state for next test if needed, though SetFloatValue should handle bounds.
	disp.Value = initialValue
	disp.dpPos = initialDpPos
}

func TestSilentSetFloatValue(t *testing.T) {
	max := 99999
	min := -99999
	disp := createTestDisplay(t, 5, 2, max, min, Signed)
	changed := false
	disp.OnChanged = func(v float64) {
		changed = true
	}

	disp.SilentSetFloatValue(12.34, 2)
	assert.Equal(t, 1234, disp.Value)
	assert.Equal(t, 2, disp.dpPos)
	assert.False(t, changed)
	assertSegments(t, disp)

	disp.SilentSetFloatValue(56.789, 3) // Change dpPos silently
	assert.Equal(t, 56789, disp.Value)
	assert.Equal(t, 3, disp.dpPos)
	assert.False(t, changed)
	assertSegments(t, disp)

	// Test bounds (should not change value if out of bounds)
	// currentValue := disp.Value
	currentDpPos := 2
	disp.SilentSetFloatValue(1000.0, currentDpPos) // Above max
	assert.Equal(t, max, disp.Value)
	assert.Equal(t, currentDpPos, disp.dpPos)
	assert.False(t, changed)
	assertSegments(t, disp)

	disp.SilentSetFloatValue(-1000.0, currentDpPos) // Below min
	assert.Equal(t, min, disp.Value)
	assert.Equal(t, currentDpPos, disp.dpPos)
	assert.False(t, changed)
	assertSegments(t, disp)

	disp.SilentSetFloatValue(10.0, -1) // Invalid dpPos
	assert.Equal(t, min, disp.Value)
	assert.Equal(t, currentDpPos, disp.dpPos)
	assert.False(t, changed)
	assertSegments(t, disp)
}

func TestSetMinMax(t *testing.T) {
	disp := createTestDisplay(t, 4, 0, 1000, -1000, Signed)

	// 1. Set value, then change bounds so value becomes out of bounds
	disp.SetValue(500)
	disp.SetMinMax(-100, 100) // New bounds
	assert.Equal(t, 100, disp.Value, "Value should clamp to new max")
	assert.Equal(t, 100, disp.maxValue)
	assert.Equal(t, -100, disp.minValue)

	disp.SetValue(-50)
	disp.SetMinMax(0, 50) // New bounds, making -50 invalid
	assert.Equal(t, 0, disp.Value, "Value should clamp to new min")
	assert.Equal(t, 50, disp.maxValue)
	assert.Equal(t, 0, disp.minValue)
	assertSegments(t, disp)

	// 2. Set bounds that are valid
	disp.SetValue(25)
	disp.SetMinMax(-1000, 1000)
	assert.Equal(t, 25, disp.Value, "Value should remain unchanged")
	assert.Equal(t, 1000, disp.maxValue)
	assert.Equal(t, -1000, disp.minValue)
	assertSegments(t, disp)
}

func TestSetters(t *testing.T) {
	disp := createTestDisplay(t, 3, 0, 100, 0, UnSigned)

	// SetOncolor
	newColor := color.NRGBA{B: 255, A: 255}
	disp.SetOncolor(newColor)
	assert.Equal(t, newColor, disp.onColor)
	// Note: We can't easily test if Refresh() was called or if the rendered output changed

	// SetUnit
	disp.SetUnit("mA")
	assert.NotNil(t, disp.unit)
	assert.Equal(t, "mA", disp.unit.Text)
	assert.Equal(t, newColor, disp.unit.Color, "Unit color should match onColor")

	// SetLabel
	disp.SetLabel("Current")
	assert.NotNil(t, disp.label)
	assert.Equal(t, "Current", disp.label.Text)
	assert.Equal(t, newColor, disp.label.Color, "Label color should match onColor")
}

func TestSetDigitAtIndex(t *testing.T) {
	disp := createTestDisplay(t, 4, 0, 9999, -9999, Signed)

	// Positive value
	disp.SetValue(1234)
	assertSegments(t, disp)
	disp.setDigitAtIndex(9, 0) // Change LSB (4 -> 9)
	assert.Equal(t, 1239, disp.Value)
	assertSegments(t, disp)
	disp.setDigitAtIndex(0, 2) // Change hundreds (2 -> 0)
	assert.Equal(t, 1039, disp.Value)
	assertSegments(t, disp)
	disp.setDigitAtIndex(5, 3) // Change MSB (1 -> 5)
	assert.Equal(t, 5039, disp.Value)
	assertSegments(t, disp)

	// Negative value
	disp.SetValue(-1234)
	disp.setDigitAtIndex(9, 0) // Change LSB (4 -> 9)
	assert.Equal(t, -1239, disp.Value)
	assertSegments(t, disp)
	disp.setDigitAtIndex(0, 1) // Change tens (3 -> 0)
	assert.Equal(t, -1209, disp.Value)
	assertSegments(t, disp)
	disp.setDigitAtIndex(8, 3) // Change MSB (1 -> 8)
	assert.Equal(t, -8209, disp.Value)
	assertSegments(t, disp)

	// Test clamping
	disp.SetValue(9990)
	disp.setDigitAtIndex(9, 0) // 9990 -> 9999 (max)
	assert.Equal(t, 9999, disp.Value)
	assertSegments(t, disp)
	disp.setDigitAtIndex(8, 0) // 9999 -> 9998
	assert.Equal(t, 9998, disp.Value)
	assertSegments(t, disp)

	disp.SetValue(-9990)
	disp.setDigitAtIndex(9, 0) // -9990 -> -9999 (min)
	assert.Equal(t, -9999, disp.Value)
	assertSegments(t, disp)
	disp.setDigitAtIndex(8, 0) // -9999 -> -9998
	assert.Equal(t, -9998, disp.Value)
	assertSegments(t, disp)
}

func TestDownLogic(t *testing.T) {
	disp := createTestDisplay(t, 4, 0, 9999, 0, UnSigned) // Unsigned for simpler borrow test

	// Simple decrement
	disp.SetValue(1234)
	disp.down(0) // Decrement LSB
	assert.Equal(t, 1233, disp.Value)
	assertSegments(t, disp)
	disp.down(1) // Decrement tens
	assert.Equal(t, 1223, disp.Value)
	assertSegments(t, disp)

	// Borrowing
	disp.SetValue(1200)
	disp.down(0) // Decrement LSB (0 -> 9, borrow from tens) -> tens becomes 9, borrow from hundreds -> hundreds becomes 1
	assert.Equal(t, 1199, disp.Value)
	assertSegments(t, disp)

	disp.SetValue(1000)
	disp.down(0) // 1000 -> 0999
	assert.Equal(t, 999, disp.Value)
	assertSegments(t, disp)

	disp.SetValue(2030)
	disp.down(1) // Decrement tens (3 -> 2)
	assert.Equal(t, 2020, disp.Value)
	assertSegments(t, disp)
	disp.down(1) // Decrement tens (2 -> 1)
	assert.Equal(t, 2010, disp.Value)
	assertSegments(t, disp)
	disp.down(1) // Decrement tens (1 -> 0)
	assert.Equal(t, 2000, disp.Value)
	assertSegments(t, disp)
	disp.down(1) // Decrement tens (0 -> 9, borrow from hundreds) -> hundreds becomes 9, borrow from thousands -> thousands becomes 1
	assert.Equal(t, 1990, disp.Value)
	assertSegments(t, disp)

	// Edge cases
	disp.SetValue(0)
	disp.down(0) // Try to decrement below min
	assert.Equal(t, 0, disp.Value)
	assertSegments(t, disp)
	disp.down(3) // Try to decrement below min (MSB)
	assert.Equal(t, 0, disp.Value)
	assertSegments(t, disp)

	disp.SetValue(1)
	disp.down(1) // Try decrementing a higher digit when lower is non-zero
	assert.Equal(t, 0, disp.Value)
	assertSegments(t, disp)
	disp.SetValue(10)
	disp.down(0) // 10 -> 9
	assert.Equal(t, 9, disp.Value)
	assertSegments(t, disp)
	disp.SetValue(10)
	disp.down(1) // 10 -> 0 (decrementing the '1')
	assert.Equal(t, 0, disp.Value)
	assertSegments(t, disp)
}

// --- Tests for Interaction Logic (Focus, Keyboard, Mouse) ---
// These tests verify state changes, assuming Fyne calls the methods correctly.

func TestFocus(t *testing.T) {
	disp := createTestDisplay(t, 3, 0, 100, 0, UnSigned)
	disp.Readonly = false

	assert.Equal(t, digitCursorOut, disp.digitCursor, "Cursor should be out initially")

	// Gain Focus
	disp.FocusGained()
	assert.Equal(t, len(disp.digits)-1, disp.digitCursor, "Cursor should be at MSB on focus gain")

	// Lose Focus
	disp.FocusLost()
	assert.Equal(t, digitCursorOut, disp.digitCursor, "Cursor should be out on focus lost")

	// Readonly test
	disp.Readonly = true
	disp.FocusGained()
	assert.Equal(t, digitCursorOut, disp.digitCursor, "Cursor should remain out when readonly")
	disp.FocusLost() // Should have no effect
	assert.Equal(t, digitCursorOut, disp.digitCursor)
}

func TestTypedKey(t *testing.T) {
	disp := createTestDisplay(t, 4, 1, 9999, -9999, Signed) // 4 digits, 1 decimal => xxx.y
	disp.Readonly = false
	disp.FocusGained() // Need focus for keyboard input

	initialValue := 1234 // Represents 123.4
	disp.SetValue(initialValue)
	assertSegments(t, disp)
	disp.digitCursor = 0 // Start at LSB (tenths place)

	// Move cursor
	disp.TypedKey(&fyne.KeyEvent{Name: fyne.KeyLeft})
	assert.Equal(t, 1, disp.digitCursor) // Moved left
	assertSegments(t, disp)
	disp.TypedKey(&fyne.KeyEvent{Name: fyne.KeyLeft})
	assert.Equal(t, 2, disp.digitCursor) // Moved left
	assertSegments(t, disp)
	disp.TypedKey(&fyne.KeyEvent{Name: fyne.KeyRight})
	assert.Equal(t, 1, disp.digitCursor) // Moved right
	assertSegments(t, disp)

	// Increment/Decrement
	disp.digitCursor = 0                            // Tenths place (value 4)
	disp.TypedKey(&fyne.KeyEvent{Name: fyne.KeyUp}) // 123.4 -> 123.5 (internal 1234 -> 1235)
	assert.Equal(t, 1235, disp.Value)
	assertSegments(t, disp)
	disp.TypedKey(&fyne.KeyEvent{Name: fyne.KeyDown}) // 123.5 -> 123.4 (internal 1235 -> 1234)
	assert.Equal(t, 1234, disp.Value)
	assertSegments(t, disp)

	disp.digitCursor = 2                            // Tens place (value 2)
	disp.TypedKey(&fyne.KeyEvent{Name: fyne.KeyUp}) // 123.4 -> 133.4 (internal 1234 -> 1334)
	assert.Equal(t, 1334, disp.Value)
	assertSegments(t, disp)
	disp.TypedKey(&fyne.KeyEvent{Name: fyne.KeyDown}) // 133.4 -> 123.4 (internal 1334 -> 1234)
	assert.Equal(t, 1234, disp.Value)
	assertSegments(t, disp)

	// Set digits
	disp.digitCursor = 1                           // Units place (value 3)
	disp.TypedKey(&fyne.KeyEvent{Name: fyne.Key9}) // 123.4 -> 129.4 (internal 1234 -> 1294), cursor moves right
	assert.Equal(t, 1294, disp.Value)
	assert.Equal(t, 0, disp.digitCursor) // Cursor moved right after typing digit
	assertSegments(t, disp)
	disp.TypedKey(&fyne.KeyEvent{Name: fyne.Key0}) // 129.4 -> 129.0 (internal 1294 -> 1290), cursor moves right (stays at 0)
	assert.Equal(t, 1290, disp.Value)
	assert.Equal(t, 0, disp.digitCursor) // Cursor cannot move further right
	assertSegments(t, disp)

	// Delete / Backspace
	disp.SetValue(5678) // 567.8
	assertSegments(t, disp)
	disp.digitCursor = 1                                // Units place (value 7)
	disp.TypedKey(&fyne.KeyEvent{Name: fyne.KeyDelete}) // 567.8 -> 560.8 (internal 5678 -> 5608), cursor moves right
	assert.Equal(t, 5608, disp.Value)
	assert.Equal(t, 0, disp.digitCursor)
	assertSegments(t, disp)

	disp.SetValue(5678) // 567.8
	assertSegments(t, disp)
	disp.digitCursor = 2                                   // Tens place (value 6)
	disp.TypedKey(&fyne.KeyEvent{Name: fyne.KeyBackspace}) // 567.8 -> 507.8 (internal 5678 -> 5078), cursor moves left
	assert.Equal(t, 5078, disp.Value)
	assert.Equal(t, 3, disp.digitCursor)
	assertSegments(t, disp)

	// Readonly check
	disp.Readonly = true
	disp.SetValue(1111)
	assertSegments(t, disp)
	disp.digitCursor = 1
	disp.TypedKey(&fyne.KeyEvent{Name: fyne.KeyUp})
	assert.Equal(t, 1111, disp.Value, "Value should not change when readonly")
	disp.TypedKey(&fyne.KeyEvent{Name: fyne.Key5})
	assert.Equal(t, 1111, disp.Value, "Value should not change when readonly")
	assert.Equal(t, 1, disp.digitCursor, "Cursor should not move when readonly")
	assertSegments(t, disp)
}

func TestSetDigitCursor(t *testing.T) {
	disp := createTestDisplay(t, 5, 0, 100, 0, UnSigned)
	// Need to simulate layout to get relative positions
	disp.CreateRenderer().Layout(disp.MinSize()) // This calculates relPos

	// Simulate X positions relative to the calculated digit positions
	// These X values are just estimates for testing the logic.
	// The exact values depend on digitWidth, skew, spacing etc.
	// We assume digits are laid out right-to-left visually, but indexed left-to-right internally.
	// relPos[0] is the leftmost digit visually (MSB), relPos[4] is the rightmost (LSB)

	// Check positions relative to calculated relPos
	// Note: The test assumes X decreases from left to right on screen,
	// matching how relPos seems to be calculated (higher index = further right = smaller X?)
	// Let's verify relPos calculation direction first.
	// Based on digitSetRelPos, higher index seems to have *smaller* X.
	// So X > pos.X means we are to the *left* of that digit's start.

	// Example: Assume relPos are roughly [4: 20, 3: 60, 2: 100, 1: 140, 0: 180] (higher index = smaller X)
	// Let's mock these for predictability
	mockPositions := []fyne.Position{
		{X: 180, Y: 5}, // Index 0 (LSB, rightmost)
		{X: 140, Y: 5}, // Index 1
		{X: 100, Y: 5}, // Index 2
		{X: 60, Y: 5},  // Index 3
		{X: 20, Y: 5},  // Index 4 (MSB, leftmost)
	}
	for i := range disp.digits {
		disp.digits[i].relPos = mockPositions[i]
	}

	tests := []struct {
		name          string
		xPos          float32
		expectedIndex int
	}{
		{"Far Right (LSB)", 190, 0},
		{"Between 0 and 1", 150, 1},
		{"Between 1 and 2", 110, 2},
		{"Between 2 and 3", 70, 3},
		{"Between 3 and 4", 30, 4},
		{"Far Left (MSB)", 10, 4}, // Should still select the leftmost digit
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ok := disp.setDigitCursor(tt.xPos)
			assert.True(t, ok) // Should always return true in this setup
			assert.Equal(t, tt.expectedIndex, disp.digitCursor)
		})
	}
}

func TestMouseInAndOut(t *testing.T) {
	disp := createTestDisplay(t, 3, 0, 100, 0, UnSigned)
	disp.Readonly = false
	disp.CreateRenderer().Layout(disp.MinSize()) // Calculate positions
	// Mock positions for predictability
	mockPositions := []fyne.Position{
		{X: 100, Y: 5}, // Index 0 (LSB)
		{X: 60, Y: 5},  // Index 1
		{X: 20, Y: 5},  // Index 2 (MSB)
	}
	for i := range disp.digits {
		disp.digits[i].relPos = mockPositions[i]
	}

	// Mouse In
	disp.Tapped(&fyne.PointEvent{Position: fyne.Position{100, 0}})
	// disp.MouseIn(&fyne.PointEvent{AbsolutePosition: fyne.Position{c.Position().X, c.Position().Y}, Position: fyne.Position{0, 0}})
	// disp.MouseIn(&desktop.MouseEvent{fyne.PointEvent{AbsolutePosition: fyne.NewPos(70, 10).X, fyne.NewPos(70, 10).Y}}) // Should target index 1
	assert.Equal(t, 1, disp.digitCursor, "MouseIn should set cursor")

	// Mouse Out
	disp.MouseOut() // ??????
	assert.Equal(t, digitCursorOut, disp.digitCursor, "MouseOut should reset cursor")

	// Readonly check
	disp.Readonly = true
	disp.Tapped(&fyne.PointEvent{AbsolutePosition: fyne.Position{0, 0}, Position: fyne.Position{0, 0}})
	// disp.MouseIn(&desktop.MouseEvent{Position: fyne.NewPos(70, 10)})
	assert.Equal(t, digitCursorOut, disp.digitCursor, "MouseIn should do nothing when readonly")
	disp.MouseOut() // Should also do nothing
	assert.Equal(t, digitCursorOut, disp.digitCursor)
}

func TestScrolled(t *testing.T) {
	disp := createTestDisplay(t, 4, 0, 9999, 0, UnSigned)
	disp.Readonly = false
	disp.CreateRenderer().Layout(disp.MinSize())
	// Mock positions
	mockPositions := []fyne.Position{
		{X: 140, Y: 5}, // Index 0
		{X: 100, Y: 5}, // Index 1
		{X: 60, Y: 5},  // Index 2
		{X: 20, Y: 5},  // Index 3
	}
	for i := range disp.digits {
		disp.digits[i].relPos = mockPositions[i]
	}

	disp.SetValue(1234)
	disp.digitCursor = 1 // Target the '3' (tens place)

	// Scroll Up
	disp.Scrolled(&fyne.ScrollEvent{Scrolled: fyne.NewDelta(0, 1)})                 // Positive DY
	assert.Equal(t, 1244, disp.Value, "Scroll up should increment digit at cursor") // 1234 + 10

	// Scroll Down
	disp.Scrolled(&fyne.ScrollEvent{Scrolled: fyne.NewDelta(0, -1)})                  // Negative DY
	assert.Equal(t, 1234, disp.Value, "Scroll down should decrement digit at cursor") // 1244 - 10

	// Scroll Down again (using down logic)
	disp.Scrolled(&fyne.ScrollEvent{Scrolled: fyne.NewDelta(0, -1)})
	assert.Equal(t, 1224, disp.Value) // 1234 - 10

	// Readonly check
	disp.Readonly = true
	disp.SetValue(5555)
	disp.digitCursor = 2
	disp.Scrolled(&fyne.ScrollEvent{Scrolled: fyne.NewDelta(0, 1)})
	assert.Equal(t, 5555, disp.Value, "Scroll should do nothing when readonly")
	disp.Scrolled(&fyne.ScrollEvent{Scrolled: fyne.NewDelta(0, -1)})
	assert.Equal(t, 5555, disp.Value, "Scroll should do nothing when readonly")
}
