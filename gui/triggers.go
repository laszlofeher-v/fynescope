package gui

import (
	"fmt"
	"psgo/ps2000a"
	"psgo/psi"
)

const autoTriggerMs = 1000

func (ps *PsDesc) setSimpleTrigger() (err error) {
	err = ps.Con.Send(&psi.SetSimpleTriggerMsg{Eable: true, Source: ps2000a.ChA, Threshold: 0,
		Direction: ps2000a.TriggerRaising, Delay: 10, AutoTriggerMs: autoTriggerMs})
	if err != nil {
		fmt.Println(err)
		return
	}
	resp, err := ps.Con.Receive()
	if err != nil || resp.Status() != nil {
		fmt.Println(err, resp.Status())
		return
	}
	return
}
func (ps *PsDesc) setTrigger() (err error) {
	propMsg := &psi.SetTriggerChannelPropertiesMsg{ChannelProperties: nil, AuxOutputEnable: false, AutoTriggerMs: autoTriggerMs}
	err = ps.Con.Send(propMsg)
	if err != nil {
		fmt.Println("runblock SetTriggerChannelProperties:", err)
		return
	}
	resp, err := ps.Con.Receive()
	if err != nil || resp.Status() != nil {
		fmt.Println(err, resp.Status())
		return
	}
	triggerVoltage := int16(ps.mvToAdc(0, ps.ChannelSettings[ps2000a.ChA].Range))
	propMsg.ChannelProperties = []ps2000a.TriggerChannelProperties{{ThresholdUpper: triggerVoltage,
		ThresholdUpperHysteresis: uint16(1024), ThresholdLower: triggerVoltage,
		ThresholdLowerHysteresis: uint16(1024), Channel: ps2000a.ChA, ThresholdMode: ps2000a.Level}}
	// err = ps.SetTriggerChannelProperties(channelProperties, false, autoTriggerMs)
	err = ps.Con.Send(propMsg)
	if err != nil {
		fmt.Println("runblock SetTriggerChannelProperties:", err)
		fmt.Println("channelProperties:", propMsg.ChannelProperties)
		return
	}
	resp, err = ps.Con.Receive()
	if err != nil || resp.Status() != nil {
		fmt.Println(err, resp.Status())
		return
	}
	condMsg := &psi.SetTriggerChannelConditionsMsg{}
	condMsg.TriggerConditions = []ps2000a.TriggerConditions{{ps2000a.CondTrue, ps2000a.CondDontCare, ps2000a.CondDontCare,
		ps2000a.CondDontCare, ps2000a.CondDontCare, ps2000a.CondDontCare, ps2000a.CondDontCare, ps2000a.CondDontCare}}
	err = ps.Con.Send(condMsg)
	if err != nil {
		fmt.Println("runblock SetTriggerChannelProperties:", err)
		return
	}
	resp, err = ps.Con.Receive()
	if err != nil || resp.Status() != nil {
		fmt.Println(err, resp.Status())
		return
	}
	//		err = ps.SetTriggerChannelConditions(conditions)
	// if err != nil {
	// 	fmt.Println("runblock SetTriggerChannelConditions:", err)
	// 	fmt.Println("condMsg.TriggerConditions:", condMsg.TriggerConditions)
	// 	return
	// }
	dirMsg := &psi.SetTriggerChannelDirectionsMsg{ChannelA: ps2000a.TriggerRaising,
		ChannelB: ps2000a.TriggerNone, ChannelC: ps2000a.TriggerNone, ChannelD: ps2000a.TriggerNone,
		Ext: ps2000a.TriggerNone, Aux: ps2000a.TriggerNone}
	// err = ps.SetTriggerChannelDirections(ps2000a.TriggerRaising, ps2000a.TriggerNone,
	// 	ps2000a.TriggerNone, ps2000a.TriggerNone, ps2000a.TriggerNone, ps2000a.TriggerNone)
	// if err != nil {
	// 	fmt.Println("runblock SetTriggerChannelDirections:", err)
	// 	return
	// }
	err = ps.Con.Send(dirMsg)
	if err != nil {
		fmt.Println("runblock SetTriggerChannelProperties:", err)
		return
	}
	resp, err = ps.Con.Receive()
	if err != nil || resp.Status() != nil {
		fmt.Println(err, resp.Status())
		return
	}
	return
} //end setTrigger
