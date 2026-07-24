package gui

import (
	"fynescope/control"
	"fynescope/genericps"

	"fyne.io/fyne/v2/container"
)

func (scp *ScpDesc) handleTabTransition(prevTab, newTab int) {
	if prevTab == newTab {
		return
	}

	// Transitioning from non-f(f) to f(f)
	if prevTab != ffTabIndex && newTab == ffTabIndex {
		// Force Sine wave for Bode plots across all generator configurations
		scp.Settings.GenPanel.WaveType = genericps.Sine
		for i := 0; i < len(scp.Settings.ExtGen); i++ {
			scp.Settings.ExtGen[i].WaveType = genericps.Sine
		}
		for i := 0; i < len(scp.Settings.SimGenPanel); i++ {
			scp.Settings.SimGenPanel[i].WaveType = genericps.Sine
		}

		scp.SaveSettings()

		// Refresh generator panels to reflect Sine wave selection in UI
		if scp.genLayout != nil {
			scp.genLayout.RemoveAll()
			if scp.psControl != nil && scp.psControl.Con != nil && scp.psControl.Con.ID == genericps.SimId {
				scp.newSimGenPanel(scp.genLayout, true)
			} else {
				scp.newGenPanel(scp.genLayout)
			}
		}

		if scp.extgenLayout != nil && scp.ExtGenEnabled {
			scp.extgenLayout.RemoveAll()
			scp.extgenLayout.Add(scp.newExtGenTab(true))
		}

		if scp.ffAmpDisp != nil {
			scp.ffAmpDisp.SetValue(int(scp.Settings.FfGen.Amplitude))
			scp.ffAmpDisp.Refresh()
		}

		if scp.psControl != nil {
			if scp.Settings.Ff.UseExternalGen && scp.extGen.Connected() {
				scp.syncExtGenSettings()
			} else if scp.psControl.Con != nil && scp.psControl.Con.ID == genericps.SimId {
				scp.applyFfSimGenSettings(false)
				scp.applyFfSimGenSettings(scp.Settings.FfGen.On)
			} else {
				scp.applyFfGenSettings(false)
				scp.applyFfGenSettings(scp.Settings.FfGen.On)
			}
		}
		if scp.running {
			scp.ResetFfSweep()
			scp.startFfSweep()
		}
	}

	// Transitioning from f(f) to non-f(f)
	if prevTab == ffTabIndex && newTab != ffTabIndex {
		if scp.status.Code() == StatusWrongFfTrigger {
			scp.psControl.DisplayStatus("", control.Info)
		}
		scp.stopFfSweep() // stop any running Bode sweep
		if scp.psControl != nil && scp.psControl.Con.ID == genericps.SimId {
			for i := 0; i < int(scp.channelCount); i++ {
				scp.applySimGenSettings(genericps.ChannelId(i), &scp.Settings.SimGenPanel[i])
			}
		} else {
			scp.applyInternalGenSettings(scp.Settings.GenPanel.On)
		}
	}
}

func (scp *ScpDesc) shouldDrawRaster(targetTabIndex int) bool {
	if scp.controlTab == nil {
		return false
	}
	selectedIndex := scp.controlTab.SelectedIndex()
	if selectedIndex == targetTabIndex {
		return true
	}
	if targetTabIndex == ftTabIndex && selectedIndex == rlcTabIndex {
		return true
	}
	if scp.Settings.Window.LastDispFunction == targetTabIndex {
		sel := scp.controlTab.Selected()
		if sel == scp.genTab || sel == scp.filterTab || sel == scp.extgenTab || sel == scp.vchTab {
			return true
		}
	}
	return false
}

func (scp *ScpDesc) dockTab(tab *container.TabItem) {
	if tab == nil || scp.controlTab == nil {
		return
	}
	// ensure tab is not already in Items
	for _, t := range scp.controlTab.Items {
		if t == tab {
			scp.controlTab.Select(tab)
			return
		}
	}
	// Global ordered list of all possible tabs
	allTabs := []*container.TabItem{
		scp.ftTab, scp.fvTab, scp.dftTab, scp.ffTab, scp.rlcTab, scp.filterTab, scp.genTab, scp.extgenTab, scp.vchTab,
	}

	var newItems []*container.TabItem
	for _, t := range allTabs {
		if t == nil {
			continue
		}
		if t == tab {
			newItems = append(newItems, t)
			continue
		}
		for _, existing := range scp.controlTab.Items {
			if existing == t {
				newItems = append(newItems, t)
				break
			}
		}
	}

	scp.controlTab.Items = newItems
	scp.controlTab.Refresh()
	scp.controlTab.Select(tab)
}
