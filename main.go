package main

import (
	"fmt"
	//	"image/color"
	"psgo/ps2000a"

	"psgo/gui"
	"psgo/gui/inter"
	"psgo/ps2000acon"
	"psgo/psi"
	//	"time"
)

func main() {
	console := false
	if !console {
		con, err := psi.OpenUnit("")
		if err != nil {
			fmt.Println(err)
			return
		}
		ps := &gui.PsDesc{Con: con}
		// if err := con.Send(&psi.PingUnitMsg{}); err != nil {
		// 	fmt.Println(err)
		// } else {
		// 	resp, ok := <-con.RespCh
		// 	fmt.Println(resp, ok)
		// }
		// if err := con.Send(&psi.FlashLedMsg{Start: 7}); err != nil {
		// 	fmt.Println(err)
		// } else {
		// 	resp, ok := <-con.RespCh
		// 	fmt.Println(resp, ok)
		// }

		ps.ChannelCount = 4
		ps.Language = inter.Eng
		ps.GenFreq = ps2000a.MinFrequency
		//_ = ps2000a.Picoscope2000Handler(ps) //check if object implements the interface

		// sim := &gui.DemoDesc{}
		// _ = ps2000a.Picoscope2000Handler(sim) //check if object implements the interface

		err = ps.Menu()
		//err = ps.CloseUnit()
		if err != nil {
			fmt.Println("main close err=", err)
		}
	} else {
		ps := &ps2000acon.PsDesc{}
		_ = ps2000a.Picoscope2000Handler(ps) //check if object implements the interface
		count, serials, serialTh, err := ps2000a.EnumerateUnits()
		fmt.Println("main enum ", count, serials, serialTh, err)
		if err != nil {
			return
		}
		err = ps.Menu()
		if err != nil {
			fmt.Println("main console menu err=", err)
			return
		}
		// for i := ps2000a.PicoInfo(0); i < 10; i++ {
		// 	str, err := ps.GetUnitInfo(i)
		// 	fmt.Println(i, str, err)
		// }
		// err = ps.FlashLed(4)
		// time.Sleep(5 * time.Second)
		// fmt.Println("main flash err=", err)
		// maxSegments, err := ps.GetMaxSegments()
		// fmt.Println("maxSegments=", maxSegments, " err=", err)
		// minimumValue, err := ps.MinimumValue()
		// fmt.Println("minimumValue=", minimumValue, " err=", err)
		// maximumValue, err := ps.MaximumValue()
		// fmt.Println("maximumValue=", maximumValue, " err=", err)

		// err = ps.SetChannel(ps2000a.ChA, true, ps2000a.Ac, ps2000a.Range_10v, 0.0)
		// if err != nil {
		// 	fmt.Println(" err=", err)
		// 	return
		// }
		// err = ps.SetSimpleTrigger(true, ps2000a.ChA, int16(100),
		// 	ps2000a.TriggerRaising, uint32(0), int16(100))
		// if err != nil {
		// 	fmt.Println(" err=", err)
		// 	return
		// }
		// // func (ps *Ps2000) PsSetDataBuffer(ch int32, bufferIn []int16, segmentIndex uint32,
		// // 	mode int16) (buffer []int16, err error) {
		// buffer := make([]int16, 8192)
		// buffer, err = ps.SetDataBuffer(ps2000a.ChA, buffer, 0, ps2000a.RatioModeNone)
		// sampleInterval, err := ps.RunStreaming(1, ps2000a.TuMs, 1, 1, false, 1, ps2000a.RatioModeNone, uint32(len(buffer)))
		// if err != nil {
		// 	fmt.Println(" err=", err)
		// 	return
		// }
		// fmt.Println("sampleInterval=", sampleInterval)
		// noOfCaptures, err := ps.GetNoOfCaptures()
		// if err != nil {
		// 	fmt.Println(" err=", err)
		// 	return
		// }
		// fmt.Println("noOfCaptures=", noOfCaptures, " err=", err)
		err = ps.CloseUnit()
		if err != nil {
			fmt.Println("main close err=", err)
		}
		fmt.Print("\n\r")
	}
}
