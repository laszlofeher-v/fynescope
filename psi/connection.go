package psi

// Receives messages from the GUI
// Calls go ps2000a API wrapper

import (
	"fmt"
	"psgo/ps2000a"

	//	"psgo/ps2000asim"
	"time"
)

const (
	cmdSendTimeout         = 1000 * time.Millisecond
	responseReceiveTimeout = 1000 * time.Millisecond
)

type (
	MessengerCh chan Messenger
	ResponderCh chan Responder
	Responder   interface {
		Status() error
	}
	Messenger interface {
		setCmdCh(cmdCh MessengerCh)
		setRespCh(RespCh ResponderCh)
		setPs(ps ps2000a.Picoscope2000Handler)
	}
	ConnectionDesc struct {
		cmdCh  MessengerCh                  // command channel
		ps     ps2000a.Picoscope2000Handler // scope API
		RespCh ResponderCh                  // respond channel
	}
	Unit struct {
		timeStamp time.Time
		ps        ps2000a.Picoscope2000Handler
		receive   func()
		cmdCh     MessengerCh // command channel
		respCh    ResponderCh // respond channel
		err       error
	}
	NullMsg struct {
		Unit
	}
)

func (s *ConnectionDesc) Receive() (resp Responder, err error) {
	select {
	case resp = <-s.RespCh:
	case <-time.After(responseReceiveTimeout):
		err = fmt.Errorf("Could not receive on %v\n", s.RespCh)
	}
	return

}
func (s *ConnectionDesc) Send(m Messenger) (err error) {
	m.setCmdCh(s.cmdCh)
	m.setRespCh(s.RespCh)
	m.setPs(s.ps)
	select {
	case s.cmdCh <- m:
	case <-time.After(cmdSendTimeout):
		err = fmt.Errorf("Could not send %v\n", m)
	}
	return
}

func (msg *Unit) setRespCh(respCh ResponderCh) {
	msg.respCh = respCh
}

func (msg *Unit) setCmdCh(cmdCh MessengerCh) {
	msg.cmdCh = cmdCh
}

func (msg *Unit) setPs(ps ps2000a.Picoscope2000Handler) {
	msg.ps = ps
}

func (msg Stat) Status() error {
	return msg.err
}

func dispatch(msg Messenger) {
	switch m := msg.(type) {
	case *NullMsg:
		m.respCh <- &Stat{}
	case *SetChannelMsg:
		setChannel(m)
	case *GetUnitInfoMsg:
		getUnitInfo(m)
	case *CloseUnitMsg:
		closeUnit(m)
	case *FlashLedMsg:
		flashLed(m)
	case *GetValuesAsyncMsg:
		getValuesAsync(m)
	case *GetValuesBulkMsg:
		getValuesBulk(m)
	case *GetValuesOverlappedBulkMsg:
		getValuesOverlappedBulk(m)
	case *GetValuesOverlappedMsg:
		getValuesOverlapped(m)
	case *GetValuesMsg:
		getValues(m)
	case *GetAnalogueOffsetMsg:
		getAnalogueOffsetMsg(m)
	case *GetChannelInformationMsg:
		getChannelInformation(m)
	case *GetMaxDownSampleRatioMsg:
		getMaxDownSampleRatio(m)
	case *GetMaxSegmentsMsg:
		getMaxSegments(m)
	case *GetNumOfCapturesMsg:
		getNumberOfCaptures(m)
	case *GetNumOfProcessedCapturesMsg:
		getNumberOfProcessedCaptures(m)
	case *GetTimebaseMsg:
		getTimebase(m)
	case *GetTimebase2Msg:
		getTimebase2(m)
	case *MaximumValueMsg:
		maximumValue(m)
	case *MinimumValueMsg:
		minimumValue(m)
	case *SetSimpleTriggerMsg:
		setSimpleTrigger(m)
	case *SetDataBufferMsg:
		setDataBuffer(m)
	case *SetDataBuffersMsg:
		setDataBuffers(m)
	case *SetUnscaledDataBuffersMsg:
		setUnscaledDataBuffers(m)
	case *SetEtsTimeBufferMsg:
		setEtsTimeBuffer(m)
	case *SetEtsTimeBuffersMsg:
		setEtsTimeBuffers(m)
	case *SetEtsMsg:
		setEts(m)
	case *RunStreamingMsg:
		runStreaming(m)
	case *RunBlockMsg:
		runBlock(m)
	case *SetTriggerChannelPropertiesMsg:
		setTriggerChannelProperties(m)
	case *SetTriggerChannelConditionsMsg:
		setTriggerChannelConditions(m)
	case *SetTriggerChannelDirectionsMsg:
		setTriggerChannelDirections(m)
	case *SetTriggerDelayMsg:
		setTriggerDelay(m)
	case *SetPulseWidthQualifierMsg:
		setPulseWidthQualifier(m)
	case *SetTriggerDigitalPortPropertiesMsg:
		setTriggerDigitalPortProperties(m)
	case *StopMsg:
		stop(m)
	case *SetSigGenBuiltInMsg:
		setSigGenBuiltIn(m)
	case *SetSigGenBuiltInV2Msg:
		setSigGenBuiltInV2(m)
	case *SigGenFrequencyToPhasenMsg:
		sigGenFrequencyToPhase(m)
	case *SetNumOfCapturesMsg:
		setNumOfCaptures(m)
	case *GetTriggerTimeOffsetsMsg:
		getTriggerTimeOffset(m)
	case *GetTriggerTimeOffsets64Msg:
		getTriggerTimeOffset64(m)
	case *GetValuesTriggerTimeOffsetBulkMsg:
		getValuesTriggerTimeOffsetBulk(m)
	case *GetValuesTriggerTimeOffsetBulk64Msg:
		getValuesTriggerTimeOffsetBulk64(m)
	case *HoldOffMsg:
		holdOff(m)
	case *LsReadyMsg:
		lsReady(m)
	case *TriggerOrPulseWidthQualifierEnabledMsg:
		triggerOrPulseWidthQualifierEnabled(m)
	case *MemorySegmentsMsg:
		memorySegments(m)
	case *NumOfStreamingValuesMsg:
		numOfStreamingValues(m)
	case *OpenUnitProgressMsg:
		openUnitProgress(m)
	case *PingUnitMsg:
		pingUnit(m)
	case *QueryOutputEdgeDetectMsg:
		queryOutputEdgeDetect(m)
	case *SetDigitalAnalogTriggerOperandMsg:
		setDigitalAnalogTriggerOperand(m)
	case *SetDigitalPortMsg:
		setDigitalPort(m)
	case *SetOutputEdgeDetectMsg:
		setOutputEdgeDetect(m)
	case *SetPulseWidthDigitalPortPropertiesMsg:
		setPulseWidthDigitalPortProperties(m)
	case *SetSigGenArbitraryMsg:
		setSigGenArbitrary(m)
	case *SetSigGenPropertiesArbitraryMsg:
		setSigGenPropertiesArbitrary(m)
	case *SetSigGenPropertiesBuiltInMsg:
		setSigGenPropertiesBuiltIn(m)
	case *SigGenArbitraryMinMaxValuesMsg:
		sigGenArbitraryMinMaxValues(m)
	case *SigGenSoftwareControlMsg:
		sigGenSoftwareControl(m)
	default:
		fmt.Printf("Unhandled type: %T\n", msg)
	}
}

func openUnit(serial string, ps ps2000a.Picoscope2000Handler) (con *ConnectionDesc, err error) {
	var (
		cmdCh  MessengerCh // command receiver channel
		respCh ResponderCh // respond sender channel
	)

	u := &Unit{}               // create unit descriptor
	u.ps = ps                  // add scope descriptor
	cmdCh = make(MessengerCh)  // create receiver channel
	respCh = make(ResponderCh) // create responder channel
	u.respCh = respCh          // add response channel
	u.cmdCh = cmdCh            // add response channel

	r := func() { // receiver function of the opened unit
		for {
			msg, ok := <-cmdCh
			if !ok {
				fmt.Println("Unexpected close of channel CmdCh")
				return
			}
			dispatch(msg)
		}
	}

	u.receive = r               // add receiver
	go u.receive()              // start reciever for the opened unit
	err = u.ps.OpenUnit(serial) // open unit
	if err != nil {
		return
	}
	con = &ConnectionDesc{RespCh: respCh, cmdCh: cmdCh, ps: ps}
	con.cmdCh <- &NullMsg{Unit{respCh: respCh}} // wait for receiver ready
	<-con.RespCh
	return
}

// OpenUnit called from the application or test
func OpenSimUnit(serial string) (s *ConnectionDesc, err error) {
	//s, err = openUnit(serial, &ps2000asim.SimDesc{})
	return
}

// OpenUnit called from the application
func OpenUnit(serial string) (s *ConnectionDesc, err error) {
	s, err = openUnit(serial, &ps2000a.PsDesc{})
	return
}
