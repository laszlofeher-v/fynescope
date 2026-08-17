package gui

import (
	"fynescope/control"
	"testing"

	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"
	"fynescope/selectscroll"
	"github.com/stretchr/testify/assert"
)

func TestTimeDiv_GetTimeUnitName(t *testing.T) {
	initTimeMaps()
	assert.Equal(t, "ps", getTimeUnitName(-12))
	assert.Equal(t, "ns", getTimeUnitName(-9))
	assert.Equal(t, "µs", getTimeUnitName(-6))
	assert.Equal(t, "ms", getTimeUnitName(-3))
}

func TestTimeDiv_SwitchUpTimeUnit(t *testing.T) {
	initTimeMaps()
	dt, unitName := switchUpTimeUnit(2000, -9)
	assert.Equal(t, float32(2), dt)
	assert.Equal(t, "µs", unitName)

	dt, unitName = switchUpTimeUnit(100, -6)
	assert.Equal(t, float32(0.1), dt)
	assert.Equal(t, "ms", unitName)
}

func TestTimeDiv_CheckTimeZoomConstraint(t *testing.T) {
	test.NewApp()
	scp := &ScpDesc{
		psControl: &control.PscDesc{},
	}
	scp.initStatus()

	scp.timeZoomWindow = nil
	assert.True(t, scp.checkTimeZoomConstraint(500, -9))

	scp.timeZoomWindow = test.NewWindow(widget.NewLabel(""))
	scp.timeZoomMaxScreenTime = 1000 * 1e-9 * 10
	// 5000ns < 10000ns original
	assert.True(t, scp.checkTimeZoomConstraint(500, -9))
	assert.False(t, scp.checkTimeZoomConstraint(5000, -9))
}

func TestTimeDiv_SampleUnitUp(t *testing.T) {
	test.NewApp()
	scp := &ScpDesc{
		sampleRateSelect: selectscroll.NewSelectScroll([]string{"1", "2", "3"}, func(s string, e selectscroll.Exception) {}, ""),
		sampleUnitSelect: selectscroll.NewSelectScroll([]string{"Hz", "kHz", "MHz"}, func(s string, e selectscroll.Exception) {}, ""),
	}
	
	scp.sampleUnitSelect.SetSelectedIndex(1) // kHz
	scp.sampleUnitUp()
	
	assert.Equal(t, 2, scp.sampleRateSelect.SelectedIndex())
	assert.Equal(t, 0, scp.sampleUnitSelect.SelectedIndex()) // Hz
}

func TestTimeDiv_SampleUnitDown(t *testing.T) {
	test.NewApp()
	scp := &ScpDesc{
		sampleRateSelect: selectscroll.NewSelectScroll([]string{"1", "2", "3"}, func(s string, e selectscroll.Exception) {}, ""),
		sampleUnitSelect: selectscroll.NewSelectScroll([]string{"Hz", "kHz", "MHz"}, func(s string, e selectscroll.Exception) {}, ""),
	}
	
	scp.sampleUnitSelect.SetSelectedIndex(1) // kHz
	scp.sampleUnitDown()
	
	assert.Equal(t, 0, scp.sampleRateSelect.SelectedIndex())
	assert.Equal(t, 2, scp.sampleUnitSelect.SelectedIndex()) // MHz
}
