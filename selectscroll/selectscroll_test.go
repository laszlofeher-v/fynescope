package selectscroll

import (
	"math/rand"
	"os"
	"testing"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/test"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	// Create a real application on the main thread so tests can be visible
	a := app.New()
	go func() {
		os.Exit(m.Run())
	}()
	a.Run()
}

func TestParseOptionToValue(t *testing.T) {
	tests := []struct {
		input    string
		expected float64
		ok       bool
	}{
		{"±50V", 50.0, true},
		{"±10mV", 0.01, true},
		{"5000", 5000.0, true},
		{"ms/div", 0.001, true},
		{"ps/div", 1e-12, true},
		{"Hz", 1.0, true},
		{"MHz", 1e6, true},
		{"S/s", 1.0, true},
		{"GS/s", 1e9, true},
		{"AC", 0.0, false},
		{"DC", 0.0, false},

		// RLC units
		{"mΩ", 1e-3, true},
		{"Ω", 1.0, true},
		{"kΩ", 1e3, true},
		{"MΩ", 1e6, true},
		{"µH", 1e-6, true},
		{"mH", 1e-3, true},
		{"H", 1.0, true},
		{"pF", 1e-12, true},
		{"nF", 1e-9, true},
		{"µF", 1e-6, true},
		{"mF", 1e-3, true},
		{"10 kΩ", 1e4, true},
		{"1.5 uF", 1.5e-6, true},
		{"100pF", 1e-10, true},
	}

	for _, tc := range tests {
		val, ok := parseOptionToValue(tc.input)
		assert.Equal(t, tc.ok, ok, "Input: %s", tc.input)
		if tc.ok {
			assert.InDelta(t, tc.expected, val, 1e-15, "Input: %s", tc.input)
		}
	}
}

func TestIsAscending(t *testing.T) {
	// Ascending lists
	assert.True(t, isAscending([]string{"1", "2", "5", "10", "20"}))
	assert.True(t, isAscending([]string{"Hz", "kHz", "MHz"}))
	assert.True(t, isAscending([]string{"S/s", "kS/s", "MS/s", "GS/s"}))
	assert.True(t, isAscending([]string{"mΩ", "Ω", "kΩ", "MΩ"}))
	assert.True(t, isAscending([]string{"µH", "mH", "H"}))
	assert.True(t, isAscending([]string{"pF", "nF", "µF", "mF"}))

	// Descending lists
	assert.False(t, isAscending([]string{"5000", "2000", "1000", "500"}))
	assert.False(t, isAscending([]string{"±50V", "±20V", "±10V"}))
	assert.False(t, isAscending([]string{"s/div", "ms/div", "µs/div"}))

	// Non-parseable or single-item lists
	assert.False(t, isAscending([]string{"AC", "DC"}))
	assert.False(t, isAscending([]string{"Sine"}))
}

func TestScrolled(t *testing.T) {
	// Real app is provided by TestMain

	// Test descending options (original style)
	descOptions := []string{"20", "10", "5"}
	var descChangedVal string
	var descChangedExc Exception
	var descSelect *SelectScroll
	
	fyne.DoAndWait(func() {
		descSelect = NewSelectScroll(descOptions, func(opt string, ex Exception) {
			descChangedVal = opt
			descChangedExc = ex
		}, "5")
		descSelect.SetSelected("10") // index 1
	})

	// Scroll up on descending list should select smaller index (larger value) => "20" (index 0)
	fyne.DoAndWait(func() {
		descSelect.Scrolled(&fyne.ScrollEvent{Scrolled: fyne.NewDelta(0, 1.0)})
	})
	assert.Equal(t, "20", descSelect.Selected)
	assert.Equal(t, "20", descChangedVal)
	assert.Equal(t, None, descChangedExc)

	// Test ascending options
	ascOptions := []string{"5", "10", "20"}
	var ascChangedVal string
	var ascChangedExc Exception
	var ascSelect *SelectScroll
	
	fyne.DoAndWait(func() {
		ascSelect = NewSelectScroll(ascOptions, func(opt string, ex Exception) {
			ascChangedVal = opt
			ascChangedExc = ex
		}, "10")
		ascSelect.SetSelected("10") // index 1
	})

	// Scroll up on ascending list should select larger index (larger value) => "20" (index 2)
	fyne.DoAndWait(func() {
		ascSelect.Scrolled(&fyne.ScrollEvent{Scrolled: fyne.NewDelta(0, 1.0)})
	})
	assert.Equal(t, "20", ascSelect.Selected)
	assert.Equal(t, "20", ascChangedVal)
	assert.Equal(t, None, ascChangedExc)
}

func TestSelectScroll_Stress(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping 10-minute stress test in short mode")
	}

	a := fyne.CurrentApp()
	if a == nil {
		t.Fatal("No current Fyne app found")
	}

	var w fyne.Window
	var ss1, ss2, ss3 *SelectScroll

	fyne.DoAndWait(func() {
		w = a.NewWindow("Stress Test")
		ss1 = NewSelectScroll([]string{"1", "2", "3", "4", "5"}, func(string, Exception) {}, "1")
		ss2 = NewSelectScroll([]string{"10", "20", "30"}, func(string, Exception) {}, "10")
		ss3 = NewSelectScroll([]string{"100", "200", "300"}, func(string, Exception) {}, "100")
		
		panel := container.NewVBox(ss1, ss2, ss3)
		w.SetContent(panel)
		w.Resize(fyne.NewSize(300, 300))
		w.Show()
	})

	// Run for 10 minutes
	timeout := time.After(10 * time.Minute)
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	widgets := []*SelectScroll{ss1, ss2, ss3}

	for {
		select {
		case <-timeout:
			fyne.DoAndWait(func() {
				w.Close()
			})
			return
		default:
			widx := rng.Intn(len(widgets))
			widget := widgets[widx]

			action := rng.Intn(3)
			fyne.DoAndWait(func() {
				switch action {
				case 0:
					// Push / Tap
					test.Tap(widget)
				case 1:
					// Drag
					if draggable, ok := interface{}(widget).(fyne.Draggable); ok {
						draggable.Dragged(&fyne.DragEvent{
							PointEvent: fyne.PointEvent{Position: widget.Position()},
							Dragged:    fyne.NewDelta(float32(rng.Intn(5)-2), float32(rng.Intn(5)-2)),
						})
					}
				case 2:
					// Scroll
					widget.Scrolled(&fyne.ScrollEvent{
						Scrolled: fyne.NewDelta(float32(rng.Intn(5)-2), float32(rng.Intn(5)-2)),
					})
				}
			})
			time.Sleep(1 * time.Millisecond) // Yield slightly so the UI actually paints
		}
	}
}
