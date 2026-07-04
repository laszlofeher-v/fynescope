package gui

import (
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"fynescope/control"
	"fynescope/genericps"
)

// buildComplexTriggerMessage converts the GUI settings (in mV) into ADC values
// and prepares the arrays for the backend control logic.
func (scp *ScpDesc) buildComplexTriggerMessage() {
	var props []genericps.TriggerChannelProperties
	var dirs []control.TriggerDirections
	var directionA, directionB, directionC, directionD genericps.ThresholdDirection
	directionA = genericps.TriggerNone
	directionB = genericps.TriggerNone
	directionC = genericps.TriggerNone
	directionD = genericps.TriggerNone

	condition := genericps.TriggerConditions{
		ChannelA:            genericps.CondDontCare,
		ChannelB:            genericps.CondDontCare,
		ChannelC:            genericps.CondDontCare,
		ChannelD:            genericps.CondDontCare,
		External:            genericps.CondDontCare,
		Aux:                 genericps.CondDontCare,
		PulseWidthQualifier: genericps.CondDontCare,
		Digital:             genericps.CondDontCare,
	}

	for i, chCfg := range scp.Settings.Trigger.Complex.Channels {
		if chCfg.Condition != genericps.CondDontCare {
			vRange := scp.Settings.Channels[i].VRange

			props = append(props, genericps.TriggerChannelProperties{
				ThresholdUpper:           int16(scp.mvToAdc(chCfg.UpperMv, vRange)),
				ThresholdUpperHysteresis: uint16(scp.mvToUAdc(chCfg.UpperHyst, vRange)),
				ThresholdLower:           int16(scp.mvToAdc(chCfg.LowerMv, vRange)),
				ThresholdLowerHysteresis: uint16(scp.mvToUAdc(chCfg.LowerHyst, vRange)),
				Channel:                  genericps.ChannelId(i),
				ThresholdMode:            chCfg.ThresholdMode,
			})

			switch genericps.ChannelId(i) {
			case genericps.ChA:
				condition.ChannelA = chCfg.Condition
				directionA = chCfg.Direction
			case genericps.ChB:
				condition.ChannelB = chCfg.Condition
				directionB = chCfg.Direction
			case genericps.ChC:
				condition.ChannelC = chCfg.Condition
				directionC = chCfg.Direction
			case genericps.ChD:
				condition.ChannelD = chCfg.Condition
				directionD = chCfg.Direction
			}
		}
	}

	dirs = append(dirs, control.TriggerDirections{
		ChannelA: directionA,
		ChannelB: directionB,
		ChannelC: directionC,
		ChannelD: directionD,
		Ext:      genericps.TriggerNone,
		Aux:      genericps.TriggerNone,
	})

	scp.triggerSettingMsg.ComplexProperties = props
	scp.triggerSettingMsg.ComplexConditions = []genericps.TriggerConditions{condition}
	scp.triggerSettingMsg.ComplexDirections = dirs
}

func (scp *ScpDesc) showComplexTriggerPopup() {
	win := scp.Window
	if win == nil {
		return
	}

	// We only show up to max connected channels
	var channelRows []fyne.CanvasObject
	header := container.NewGridWithColumns(6,
		widget.NewLabel("Channel"),
		widget.NewLabel("Condition"),
		widget.NewLabel("Direction"),
		widget.NewLabel("Threshold (mV)"),
		widget.NewLabel("Hysteresis"),
		widget.NewLabel("Mode"),
	)
	channelRows = append(channelRows, header)

	// Copy config to edit in dialog
	editedConfig := scp.Settings.Trigger.Complex

	for i := 0; i < len(scp.Settings.Channels) && i < 4; i++ {
		chIdx := i
		chName := string(rune('A' + i))
		cfg := editedConfig.Channels[chIdx]

		// Condition Dropdown
		condSelect := widget.NewSelect([]string{"Don't Care", "True", "False"}, func(s string) {
			switch s {
			case "True":
				editedConfig.Channels[chIdx].Condition = genericps.CondTrue
			case "False":
				editedConfig.Channels[chIdx].Condition = genericps.CondFalse
			default:
				editedConfig.Channels[chIdx].Condition = genericps.CondDontCare
			}
		})
		switch cfg.Condition {
		case genericps.CondTrue:
			condSelect.SetSelected("True")
		case genericps.CondFalse:
			condSelect.SetSelected("False")
		default:
			condSelect.SetSelected("Don't Care")
		}

		// Direction Dropdown
		dirSelect := widget.NewSelect([]string{"Rising", "Falling", "RisingOrFalling"}, func(s string) {
			switch s {
			case "Rising":
				editedConfig.Channels[chIdx].Direction = genericps.TriggerRaising
			case "Falling":
				editedConfig.Channels[chIdx].Direction = genericps.TriggerFalling
			case "RisingOrFalling":
				editedConfig.Channels[chIdx].Direction = genericps.TriggerRisingOrFalling
			}
		})
		switch cfg.Direction {
		case genericps.TriggerRaising:
			dirSelect.SetSelected("Rising")
		case genericps.TriggerFalling:
			dirSelect.SetSelected("Falling")
		case genericps.TriggerRisingOrFalling:
			dirSelect.SetSelected("RisingOrFalling")
		default:
			dirSelect.SetSelected("Rising")
		}

		// Threshold Input
		threshInput := widget.NewEntry()
		threshInput.SetText(strconv.Itoa(int(cfg.UpperMv)))
		threshInput.OnChanged = func(s string) {
			if v, err := strconv.Atoi(s); err == nil {
				editedConfig.Channels[chIdx].UpperMv = int32(v)
				// For simple Level mode, use same for LowerMv
				editedConfig.Channels[chIdx].LowerMv = int32(v) 
			}
		}

		// Hysteresis Input
		hystInput := widget.NewEntry()
		hystInput.SetText(strconv.Itoa(int(cfg.UpperHyst)))
		hystInput.OnChanged = func(s string) {
			if v, err := strconv.Atoi(s); err == nil {
				editedConfig.Channels[chIdx].UpperHyst = int32(v)
				editedConfig.Channels[chIdx].LowerHyst = int32(v)
			}
		}

		// Mode Dropdown
		modeSelect := widget.NewSelect([]string{"Level", "Window"}, func(s string) {
			if s == "Level" {
				editedConfig.Channels[chIdx].ThresholdMode = genericps.Level
			} else {
				editedConfig.Channels[chIdx].ThresholdMode = genericps.Window
			}
		})
		if cfg.ThresholdMode == genericps.Window {
			modeSelect.SetSelected("Window")
		} else {
			modeSelect.SetSelected("Level")
		}

		row := container.NewGridWithColumns(6,
			widget.NewLabel("Ch "+chName),
			condSelect,
			dirSelect,
			threshInput,
			hystInput,
			modeSelect,
		)
		channelRows = append(channelRows, row)
	}

	content := container.NewVScroll(container.NewVBox(channelRows...))
	content.SetMinSize(fyne.NewSize(650, 300))

	dlg := dialog.NewCustomConfirm("Complex Trigger Configuration", "Apply", "Cancel", content, func(apply bool) {
		if apply {
			scp.Settings.Trigger.Complex = editedConfig
			scp.SaveSettings()
			scp.buildComplexTriggerMessage()
			scp.psControl.SetTriggerCh <- &scp.triggerSettingMsg
			<-scp.triggerSettingMsg.Done
			scp.clearAllFtPersistentLayers()
			scp.refreshRasters()
		}
	}, win)
	dlg.Resize(fyne.NewSize(700, 400))
	dlg.Show()
}
