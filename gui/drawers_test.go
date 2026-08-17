package gui

import (
	"fynescope/genericps"
	"image"
	"testing"

	"fyne.io/fyne/v2/test"
	"github.com/stretchr/testify/assert"
)

type dummyDrawer struct{}

func (d *dummyDrawer) setRect(imgRect image.Rectangle) {}
func (d *dummyDrawer) rect() image.Rectangle           { return image.Rectangle{} }
func (d *dummyDrawer) enableRefresh()                  {}
func (d *dummyDrawer) disableRefresh()                 {}
func (d *dummyDrawer) refresh() bool                   { return false }
func (d *dummyDrawer) draw()                           {}

func TestDrawers_FtDrawers(t *testing.T) {
	scp := &ScpDesc{}
	d1 := &dummyDrawer{}
	d2 := &dummyDrawer{}

	scp.addFtDrawer(d1)
	assert.Len(t, scp.ftDrawers, 1)
	assert.Equal(t, d1, scp.ftDrawers[0])

	scp.addFtDrawer(d2)
	assert.Len(t, scp.ftDrawers, 2)

	// Delete non-existent
	scp.deleteFtDrawer(&dummyDrawer{})
	assert.Len(t, scp.ftDrawers, 2)

	// Delete existing
	scp.deleteFtDrawer(d1)
	assert.Len(t, scp.ftDrawers, 1)
	assert.Equal(t, d2, scp.ftDrawers[0])

	scp.deleteFtDrawer(d2)
	assert.Len(t, scp.ftDrawers, 0)

	// Delete from empty
	scp.deleteFtDrawer(d1)
	assert.Len(t, scp.ftDrawers, 0)
}

func TestDrawers_DftDrawers(t *testing.T) {
	scp := &ScpDesc{}
	d1 := &dummyDrawer{}

	scp.addDftDrawer(d1)
	assert.Len(t, scp.dftDrawers, 1)

	scp.deleteDftDrawer(d1)
	assert.Len(t, scp.dftDrawers, 0)
}

func TestDrawers_FvDrawers(t *testing.T) {
	scp := &ScpDesc{}
	d1 := &dummyDrawer{}

	scp.addFvDrawer(d1)
	assert.Len(t, scp.fvDrawers, 1)

	scp.deleteFvDrawer(d1)
	assert.Len(t, scp.fvDrawers, 0)
}

func TestDrawers_ClearRGBA(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 2, 2))
	for i := range img.Pix {
		img.Pix[i] = 255
	}
	clearRGBA(img)
	for i := range img.Pix {
		assert.Equal(t, uint8(0), img.Pix[i])
	}
	
	// Ensure no panic on nil
	clearRGBA(nil)
}

func TestDrawers_ClearPersistentLayers(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()
	
	imgFt := image.NewRGBA(image.Rect(0, 0, 2, 2))
	imgFt.Pix[0] = 255
	
	imgDft := image.NewRGBA(image.Rect(0, 0, 2, 2))
	imgDft.Pix[0] = 255

	scp := &ScpDesc{
		ftPersistentLayers:  []*image.RGBA{imgFt},
		dftPersistentLayers: []*image.RGBA{imgDft},
	}

	scp.clearFtPersistentLayer(0)
	assert.Equal(t, uint8(0), imgFt.Pix[0])

	imgFt.Pix[0] = 255
	scp.clearAllFtPersistentLayers()
	assert.Equal(t, uint8(0), imgFt.Pix[0])

	scp.clearDftPersistentLayer(0)
	assert.Equal(t, uint8(0), imgDft.Pix[0])

	imgDft.Pix[0] = 255
	scp.clearAllDftPersistentLayers()
	assert.Equal(t, uint8(0), imgDft.Pix[0])
	
	// Verify out of bounds doesn't panic
	scp.clearFtPersistentLayer(genericps.ChannelId(99))
	scp.clearDftPersistentLayer(genericps.ChannelId(99))
}
