package psi

import (
	"fmt"
	"psgo/ps2000a"
	"testing"
	"time"
)

func TestOpenSimUnit(t *testing.T) {
	con, err := OpenSimUnit("")
	if err != nil {
		fmt.Println(err)
		return
	}
	err = con.Send(&SetChannelMsg{Channel: ps2000a.ChB, Enabled: true})
	if err != nil {
		fmt.Println(err)
		return
	}
	stat := <-con.RespCh
	fmt.Println(stat)
	err = con.Send(&CloseUnitMsg{})
	if err != nil {
		fmt.Println(err)
		return
	}
	stat = <-con.RespCh
	fmt.Println(stat)
	time.Sleep(2 * time.Second)
}
