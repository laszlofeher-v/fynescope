package gui

import (
	"fynescope/genericps"
	"image"
	"time"

	"fyne.io/fyne/v2"
)

func (scp *ScpDesc) addFtDrawer(d drawer) {
	scp.ftDrawers = append(scp.ftDrawers, d)
}

func (scp *ScpDesc) deleteFtDrawer(d drawer) {
	if len(scp.ftDrawers) == 0 {
		return
	}
	i := 0
	for i < len(scp.ftDrawers) && scp.ftDrawers[i] != d {
		i++
		if i == len(scp.ftDrawers) {
			return
		}
	}
	scp.ftDrawers = scp.ftDrawers[:i+copy(scp.ftDrawers[i:], scp.ftDrawers[i+1:])]
}

func clearRGBA(img *image.RGBA) {
	if img == nil {
		return
	}
	for i := range img.Pix {
		img.Pix[i] = 0
	}
}

func (scp *ScpDesc) scheduleClearAllPersistentLayers() {
	if scp.delayedClearTimer != nil {
		scp.delayedClearTimer.Stop()
	}
	scp.delayedClearTimer = time.AfterFunc(500*time.Millisecond, func() {
		fyne.Do(func() {
			for i := range scp.ftPersistentLayers {
				clearRGBA(scp.ftPersistentLayers[i])
			}
			for i := range scp.dftPersistentLayers {
				clearRGBA(scp.dftPersistentLayers[i])
			}
		})
	})
}

func (scp *ScpDesc) clearFtPersistentLayer(chIndex genericps.ChannelId) {
	if int(chIndex) < len(scp.ftPersistentLayers) {
		clearRGBA(scp.ftPersistentLayers[chIndex])
	}
	scp.scheduleClearAllPersistentLayers()
}

func (scp *ScpDesc) clearAllFtPersistentLayers() {
	for i := range scp.ftPersistentLayers {
		clearRGBA(scp.ftPersistentLayers[i])
	}
	scp.scheduleClearAllPersistentLayers()
}

func (scp *ScpDesc) clearDftPersistentLayer(chIndex genericps.ChannelId) {
	if int(chIndex) < len(scp.dftPersistentLayers) {
		clearRGBA(scp.dftPersistentLayers[chIndex])
	}
	scp.scheduleClearAllPersistentLayers()
}

func (scp *ScpDesc) clearAllDftPersistentLayers() {
	for i := range scp.dftPersistentLayers {
		clearRGBA(scp.dftPersistentLayers[i])
	}
	scp.scheduleClearAllPersistentLayers()
}

func (scp *ScpDesc) addDftDrawer(d drawer) {
	scp.dftDrawers = append(scp.dftDrawers, d)
}

func (scp *ScpDesc) deleteDftDrawer(d drawer) {
	if len(scp.dftDrawers) == 0 {
		return
	}
	i := 0
	for i < len(scp.dftDrawers) && scp.dftDrawers[i] != d {
		i++
		if i == len(scp.dftDrawers) {
			return
		}
	}
	scp.dftDrawers = scp.dftDrawers[:i+copy(scp.dftDrawers[i:], scp.dftDrawers[i+1:])]
}

func (scp *ScpDesc) addFvDrawer(d drawer) {
	scp.fvDrawers = append(scp.fvDrawers, d)
}

func (scp *ScpDesc) deleteFvDrawer(d drawer) {
	if len(scp.fvDrawers) == 0 {
		return
	}
	i := 0
	for i < len(scp.fvDrawers) && scp.fvDrawers[i] != d {
		i++
		if i == len(scp.fvDrawers) {
			return
		}
	}
	scp.fvDrawers = scp.fvDrawers[:i+copy(scp.fvDrawers[i:], scp.fvDrawers[i+1:])]
}
