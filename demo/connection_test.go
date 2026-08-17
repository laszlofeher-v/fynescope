package demo

import (
	"fynescope/genericps"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCheckHandle(t *testing.T) {
	assert.NoError(t, checkHandle(1))
	assert.NoError(t, checkHandle(100))
	assert.Error(t, checkHandle(0))
	assert.Error(t, checkHandle(-1))
}

func TestStop(t *testing.T) {
	err := Stop()
	assert.NoError(t, err)
}

func TestCloseUnit(t *testing.T) {
	rspCh := make(chan struct{}, 1)
	rsp := &genericps.CloseUnitRsp{}
	
	msg := &genericps.CloseUnitMsg{}
	msg.SetHandle(1)
	msg.SetRsp(rsp)
	msg.SetRspCh(rspCh)

	// In test environment, simCloseUnit might return an error if unit isn't open,
	// but the handler should not panic and should send to the channel.
	closeUnit(msg)

	// Wait for response channel
	<-rspCh

	// If the simulator isn't properly initialized in the test, it might return an error,
	// but we just verify the response struct gets populated.
	assert.NoError(t, rsp.Status())
}

func TestSetChannel(t *testing.T) {
	rspCh := make(chan struct{}, 1)
	rsp := &genericps.SetChannelRsp{}
	
	msg := &genericps.SetChannelMsg{
		Channel:      0,
		Enabled:      true,
		CouplingType: 1, // DC
		VoltageRange: 5, // 5V
		AnalogOffset: 0.0,
	}
	msg.SetHandle(1)
	msg.SetRsp(rsp)
	msg.SetRspCh(rspCh)

	setChannel(msg)

	<-rspCh
	assert.NoError(t, rsp.Status())
}

func TestPingUnit(t *testing.T) {
	rspCh := make(chan struct{}, 1)
	rsp := &genericps.PingUnitRsp{}

	msg := &genericps.PingUnitMsg{}
	msg.SetHandle(1)
	msg.SetRsp(rsp)
	msg.SetRspCh(rspCh)

	pingUnit(msg)

	<-rspCh
	// The demo simulator always returns "Not implemented" for PingUnit.
	assert.Error(t, rsp.Status())
	assert.Contains(t, rsp.Status().Error(), "Not implemented")
}
