package control

import (
	"sync"
	"testing"
	"time"

	_ "fynescope/demo"
	"fynescope/genericps"
)

func TestPscDesc_getValidTriggerProperties(t *testing.T) {
	psControl := &PscDesc{}

	// Set up a basic trigger setting
	psControl.triggerSetting = TriggerDesc{
		TriggerADC:         1000,
		LowerTriggerADC:    -1000,
		HysteresisADC:      100,
		LowerHysteresisADC: 50,
		Source:             genericps.ChA,
		ThresholdMode:      genericps.Level,
	}

	props := psControl.getValidTriggerProperties()

	if len(props) != 1 {
		t.Fatalf("Expected 1 property, got %d", len(props))
	}

	p := props[0]
	if p.ThresholdUpper != 1000 {
		t.Errorf("Expected ThresholdUpper 1000, got %v", p.ThresholdUpper)
	}
	if p.ThresholdUpperHysteresis != 100 {
		t.Errorf("Expected ThresholdUpperHysteresis 100, got %v", p.ThresholdUpperHysteresis)
	}
	if p.ThresholdLower != -1000 {
		t.Errorf("Expected ThresholdLower -1000, got %v", p.ThresholdLower)
	}
	if p.ThresholdLowerHysteresis != 50 {
		t.Errorf("Expected ThresholdLowerHysteresis 50, got %v", p.ThresholdLowerHysteresis)
	}
	if p.Channel != genericps.ChA {
		t.Errorf("Expected Channel ChA, got %v", p.Channel)
	}
	if p.ThresholdMode != genericps.Level {
		t.Errorf("Expected ThresholdMode Level, got %v", p.ThresholdMode)
	}
}

func TestPscDesc_getValidTriggerProperties_WindowCorrection(t *testing.T) {
	psControl := &PscDesc{}

	// Case 1: lower > upper (should swap)
	psControl.triggerSetting = TriggerDesc{
		TriggerADC:      500,
		LowerTriggerADC: 1500,
		ThresholdMode:   genericps.Window,
	}

	props := psControl.getValidTriggerProperties()
	if props[0].ThresholdUpper != 1500 || props[0].ThresholdLower != 500 {
		t.Errorf("Expected bounds to be swapped to 1500 and 500, got upper: %v, lower: %v", props[0].ThresholdUpper, props[0].ThresholdLower)
	}

	// Case 2: lower == upper (should expand window to minimum 512 counts)
	psControl.triggerSetting = TriggerDesc{
		TriggerADC:      500,
		LowerTriggerADC: 500,
		ThresholdMode:   genericps.Window,
	}

	props = psControl.getValidTriggerProperties()
	if props[0].ThresholdUpper != 1012 || props[0].ThresholdLower != 500 {
		t.Errorf("Expected upper bound to be expanded to 1012, got upper: %v, lower: %v", props[0].ThresholdUpper, props[0].ThresholdLower)
	}

	// Case 3: lower == upper near max int16 (should decrement lower by 512)
	psControl.triggerSetting = TriggerDesc{
		TriggerADC:      32700,
		LowerTriggerADC: 32700,
		ThresholdMode:   genericps.Window,
	}

	props = psControl.getValidTriggerProperties()
	if props[0].ThresholdUpper != 32700 || props[0].ThresholdLower != 32188 {
		t.Errorf("Expected lower bound to be decremented to 32188, got upper: %v, lower: %v", props[0].ThresholdUpper, props[0].ThresholdLower)
	}

	// Case 4: window < 512 counts (e.g. upper=256, lower=0) should expand to at least 512 counts
	psControl.triggerSetting = TriggerDesc{
		TriggerADC:      256,
		LowerTriggerADC: 0,
		ThresholdMode:   genericps.Window,
	}
	props = psControl.getValidTriggerProperties()
	if props[0].ThresholdUpper != 512 || props[0].ThresholdLower != 0 {
		t.Errorf("Expected window to expand to upper: 512, lower: 0, got upper: %v, lower: %v", props[0].ThresholdUpper, props[0].ThresholdLower)
	}
}

func TestPscDesc_getValidTriggerProperties_RiseFall(t *testing.T) {
	psControl := &PscDesc{}

	// RiseFall with initial Level mode should automatically become Window mode in properties
	psControl.triggerSetting = TriggerDesc{
		Type:            RiseFall,
		TriggerADC:      2000,
		LowerTriggerADC: 500,
		ThresholdMode:   genericps.Level,
	}

	props := psControl.getValidTriggerProperties()
	if len(props) != 1 {
		t.Fatalf("Expected 1 property, got %d", len(props))
	}
	if props[0].ThresholdMode != genericps.Window {
		t.Errorf("Expected ThresholdMode Window for RiseFall, got %v", props[0].ThresholdMode)
	}
	if props[0].ThresholdUpper != 2000 || props[0].ThresholdLower != 500 {
		t.Errorf("Expected upper: 2000, lower: 500, got upper: %v, lower: %v", props[0].ThresholdUpper, props[0].ThresholdLower)
	}
}

func TestPscDesc_getValidTriggerProperties_HysteresisClamping(t *testing.T) {
	psControl := &PscDesc{}

	// Case 1: Window mode with upper hysteresis = 32767 (excessive hysteresis)
	psControl.triggerSetting = TriggerDesc{
		TriggerADC:         1000,
		LowerTriggerADC:    0,
		HysteresisADC:      32767,
		LowerHysteresisADC: 0,
		Source:             genericps.ChA,
		ThresholdMode:      genericps.Window,
	}

	props := psControl.getValidTriggerProperties()
	p := props[0]
	if p.ThresholdMode != genericps.Window {
		t.Errorf("Expected ThresholdMode Window, got %v", p.ThresholdMode)
	}
	if p.ThresholdUpper != 1000 || p.ThresholdLower != 0 {
		t.Errorf("Expected upper: 1000, lower: 0, got upper: %v, lower: %v", p.ThresholdUpper, p.ThresholdLower)
	}
	// Verify that upper - upperHyst >= lower + lowerHyst (no crossover)
	if int32(p.ThresholdUpper)-int32(p.ThresholdUpperHysteresis) < int32(p.ThresholdLower)+int32(p.ThresholdLowerHysteresis) {
		t.Errorf("Upper hysteresis overlaps lower threshold: upper=%d, uh=%d, lower=%d, lh=%d",
			p.ThresholdUpper, p.ThresholdUpperHysteresis, p.ThresholdLower, p.ThresholdLowerHysteresis)
	}
	if p.ThresholdUpperHysteresis == 32767 {
		t.Errorf("Expected ThresholdUpperHysteresis to be clamped from 32767, got %v", p.ThresholdUpperHysteresis)
	}

	// Case 2: Level mode with excessive hysteresis
	psControl.triggerSetting = TriggerDesc{
		Type:               Advanced,
		TriggerADC:         100,
		LowerTriggerADC:    100,
		HysteresisADC:      50000,
		LowerHysteresisADC: 0,
		Source:             genericps.ChA,
		ThresholdMode:      genericps.Level,
	}
	props = psControl.getValidTriggerProperties()
	p = props[0]
	if p.ThresholdUpperHysteresis > 32767 {
		t.Errorf("Expected upper hysteresis in level mode to be <= 32767, got %v", p.ThresholdUpperHysteresis)
	}
}

func TestPscDesc_buildTriggerConditions(t *testing.T) {
	channels := []genericps.ChannelId{genericps.ChA, genericps.ChB, genericps.ChC, genericps.ChD}
	for _, ch := range channels {
		psControl := &PscDesc{}
		psControl.triggerSetting.Source = ch
		psControl.triggerSetting.DigitalTriggerEnabled = false

		conds := psControl.buildTriggerConditions(genericps.CondTrue, genericps.CondFalse)
		if len(conds) != 1 {
			t.Fatalf("expected 1 condition, got %d", len(conds))
		}
		c := conds[0]
		if c.PulseWidthQualifier != genericps.CondFalse {
			t.Errorf("expected pwqCond false, got %v", c.PulseWidthQualifier)
		}
		if c.Digital != genericps.CondDontCare {
			t.Errorf("expected digital CondDontCare, got %v", c.Digital)
		}
		switch ch {
		case genericps.ChA:
			if c.ChannelA != genericps.CondTrue || c.ChannelB != genericps.CondDontCare {
				t.Errorf("ChA condition mismatch: %+v", c)
			}
		case genericps.ChB:
			if c.ChannelB != genericps.CondTrue || c.ChannelA != genericps.CondDontCare {
				t.Errorf("ChB condition mismatch: %+v", c)
			}
		case genericps.ChC:
			if c.ChannelC != genericps.CondTrue || c.ChannelA != genericps.CondDontCare {
				t.Errorf("ChC condition mismatch: %+v", c)
			}
		case genericps.ChD:
			if c.ChannelD != genericps.CondTrue || c.ChannelA != genericps.CondDontCare {
				t.Errorf("ChD condition mismatch: %+v", c)
			}
		}

		// With digital enabled
		psControl.triggerSetting.DigitalTriggerEnabled = true
		psControl.triggerSetting.DigitalDirections = []genericps.DigitalChannelDirections{
			{Channel: genericps.Dch0, Direction: genericps.DigitalDirectionHigh},
		}
		condsDig := psControl.buildTriggerConditions(genericps.CondTrue, genericps.CondTrue)
		if condsDig[0].Digital != genericps.CondTrue {
			t.Errorf("expected digital CondTrue when enabled, got %v", condsDig[0].Digital)
		}
	}
}

func TestPscDesc_buildPwqConditions(t *testing.T) {
	channels := []genericps.ChannelId{genericps.ChA, genericps.ChB, genericps.ChC, genericps.ChD}
	for _, ch := range channels {
		psControl := &PscDesc{}
		psControl.triggerSetting.Source = ch

		pwqConds := psControl.buildPwqConditions(genericps.CondTrue)
		if len(pwqConds) != 1 {
			t.Fatalf("expected 1 pwq condition, got %d", len(pwqConds))
		}
		p := pwqConds[0]
		switch ch {
		case genericps.ChA:
			if p.ChannelA != genericps.CondTrue || p.ChannelB != genericps.CondDontCare {
				t.Errorf("ChA pwq mismatch: %+v", p)
			}
		case genericps.ChB:
			if p.ChannelB != genericps.CondTrue || p.ChannelA != genericps.CondDontCare {
				t.Errorf("ChB pwq mismatch: %+v", p)
			}
		case genericps.ChC:
			if p.ChannelC != genericps.CondTrue || p.ChannelA != genericps.CondDontCare {
				t.Errorf("ChC pwq mismatch: %+v", p)
			}
		case genericps.ChD:
			if p.ChannelD != genericps.CondTrue || p.ChannelA != genericps.CondDontCare {
				t.Errorf("ChD pwq mismatch: %+v", p)
			}
		}
	}
}

func TestPscDesc_computePwqSamples(t *testing.T) {
	ps := &PscDesc{}
	ps.scopeScreenWidth = 10000
	ps.SamplingTimeInterval = 10 // maxPwqSamples = 1000

	ps.triggerSetting.IntervalTimeLower = 500  // 50 samples
	ps.triggerSetting.IntervalTimeUpper = 2000 // 200 samples
	ps.triggerSetting.IntervalType = genericps.PwTypeNone

	low, up := ps.computePwqSamples()
	if low != 50 || up != 200 {
		t.Errorf("expected low=50, up=200, got low=%d, up=%d", low, up)
	}

	// Test InRange swap/adjust when lower >= upper
	ps.triggerSetting.IntervalTimeLower = 2000 // 200 samples
	ps.triggerSetting.IntervalTimeUpper = 500  // 50 samples
	ps.triggerSetting.IntervalType = genericps.PwTypeInRange

	low, up = ps.computePwqSamples()
	if low >= up {
		t.Errorf("expected low < up for InRange, got low=%d, up=%d", low, up)
	}

	// Test clamping to maxPwqSamples
	ps.triggerSetting.IntervalType = genericps.PwTypeNone
	ps.triggerSetting.IntervalTimeLower = 500000
	ps.triggerSetting.IntervalTimeUpper = 500000
	low, up = ps.computePwqSamples()
	if low != 1000 || up != 1000 {
		t.Errorf("expected samples clamped to 1000, got low=%d, up=%d", low, up)
	}
}

func TestPscDesc_autoTriggerMilliseconds32(t *testing.T) {
	ps := &PscDesc{}
	ps.triggerSetting.Mode = Auto
	if ps.autoTriggerMilliseconds32() != autoTriggerMs {
		t.Errorf("expected %d, got %d", autoTriggerMs, ps.autoTriggerMilliseconds32())
	}
	ps.triggerSetting.Mode = Repeat
	if ps.autoTriggerMilliseconds32() != 0 {
		t.Errorf("expected 0 for Repeat mode, got %d", ps.autoTriggerMilliseconds32())
	}
}

func TestWindowChannelDir(t *testing.T) {
	tests := []struct {
		in   genericps.ThresholdDirection
		want genericps.ThresholdDirection
	}{
		{genericps.TriggerEnter, genericps.TriggerOutside},
		{genericps.TriggerEnterOrExit, genericps.TriggerOutside},
		{genericps.TriggerExit, genericps.TriggerInside},
		{genericps.TriggerRising, genericps.TriggerRising},
		{genericps.TriggerFalling, genericps.TriggerFalling},
	}
	for _, tt := range tests {
		if got := windowChannelDir(tt.in); got != tt.want {
			t.Errorf("windowChannelDir(%v) = %v, want %v", tt.in, got, tt.want)
		}
	}
}

func TestAdjustPwqSamples(t *testing.T) {
	// LessThan
	it, low, up := adjustPwqSamples(genericps.PwTypeLessThan, 10, 50)
	if it != genericps.PwTypeLessThan || low != 50 || up != 0 {
		t.Errorf("LessThan: got it=%v, low=%d, up=%d", it, low, up)
	}

	// GreaterThan
	it, low, up = adjustPwqSamples(genericps.PwTypeGreaterThan, 10, 50)
	if it != genericps.PwTypeGreaterThan || low != 10 || up != 0 {
		t.Errorf("GreaterThan: got it=%v, low=%d, up=%d", it, low, up)
	}

	// InRange
	it, low, up = adjustPwqSamples(genericps.PwTypeInRange, 10, 50)
	if it != genericps.PwTypeInRange || low != 10 || up != 50 {
		t.Errorf("InRange: got it=%v, low=%d, up=%d", it, low, up)
	}
}

func TestPscDesc_adjustPwqSamplesRiseFall(t *testing.T) {
	ps := &PscDesc{}

	// LessThan
	it, low, up := ps.adjustPwqSamplesRiseFall(genericps.PwTypeLessThan, 10, 50)
	if it != genericps.PwTypeLessThan || low != 50 || up != 0 {
		t.Errorf("LessThan: got it=%v, low=%d, up=%d", it, low, up)
	}

	// GreaterThan converts to InRange with upper = maxPwqSamples - 1 (32766)
	it, low, up = ps.adjustPwqSamplesRiseFall(genericps.PwTypeGreaterThan, 20, 100)
	if it != genericps.PwTypeInRange || low != 20 || up != 32766 {
		t.Errorf("GreaterThan: got it=%v, low=%d, up=%d", it, low, up)
	}
}

func TestSanitizeTriggerChannelProperties_Limits(t *testing.T) {
	props := []genericps.TriggerChannelProperties{
		{
			ThresholdUpper: 32767,
			ThresholdLower: -32768,
			ThresholdMode:  genericps.Level,
		},
	}
	san := sanitizeTriggerChannelProperties(props)
	if san[0].ThresholdUpper != 32767 || san[0].ThresholdLower != -32767 {
		t.Errorf("expected 32767 / -32767, got upper=%d, lower=%d", san[0].ThresholdUpper, san[0].ThresholdLower)
	}
}

func TestPscDesc_setTriggerChannelDirections(t *testing.T) {
	con := genericps.NewConnection()
	handle, err := genericps.OpenDemo(con, genericps.DemoId)
	if err != nil {
		t.Fatalf("failed to open demo: %v", err)
	}
	con.Handle = handle
	defer con.CloseUnit()

	channels := []genericps.ChannelId{genericps.ChA, genericps.ChB, genericps.ChC, genericps.ChD}
	for _, ch := range channels {
		ps := &PscDesc{Con: con}
		ps.triggerSetting.Source = ch
		err := ps.setTriggerChannelDirections(genericps.TriggerRising)
		if err != nil {
			t.Errorf("setTriggerChannelDirections failed for %v: %v", ch, err)
		}
	}
}

func TestPscDesc_disablePwq(t *testing.T) {
	con := genericps.NewConnection()
	handle, err := genericps.OpenDemo(con, genericps.DemoId)
	if err != nil {
		t.Fatalf("failed to open demo: %v", err)
	}
	con.Handle = handle
	defer con.CloseUnit()

	ps := &PscDesc{Con: con}
	err = ps.disablePwq(genericps.TriggerRising)
	if err != nil {
		t.Errorf("disablePwq failed: %v", err)
	}
}

func TestTriggerMonitor(t *testing.T) {
	psControl := &PscDesc{
		shutdownCh:     make(chan struct{}),
		restartChannel: make(chan struct{}, 10),
		SetTriggerCh:   make(chan *TriggerDescMsg, 10),
		getTriggerCh:   make(chan *getTriggerMsg, 10),
	}

	go psControl.triggerMonitor()

	// Initial getTrigger check (should report newSettings = false)
	var trigSetting TriggerDesc
	getMsg := &getTriggerMsg{
		triggerSettings: &trigSetting,
		newSettings:     make(chan bool, 1),
	}
	psControl.getTriggerCh <- getMsg
	select {
	case changed := <-getMsg.newSettings:
		if changed {
			t.Errorf("expected no new settings initially")
		}
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for getTrigger")
	}

	// Send new trigger settings
	setMsg := &TriggerDescMsg{
		TriggerDesc: TriggerDesc{
			Enabled:    true,
			TriggerADC: 1234,
		},
		Done: make(chan struct{}, 1),
	}
	psControl.SetTriggerCh <- setMsg
	select {
	case <-setMsg.Done:
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for SetTriggerCh done")
	}

	// Query updated trigger settings
	getMsg = &getTriggerMsg{
		triggerSettings: &trigSetting,
		newSettings:     make(chan bool, 1),
	}
	psControl.getTriggerCh <- getMsg
	select {
	case changed := <-getMsg.newSettings:
		if !changed {
			t.Errorf("expected newSettings = true after setting update")
		}
		if trigSetting.TriggerADC != 1234 {
			t.Errorf("expected TriggerADC 1234, got %d", trigSetting.TriggerADC)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for getTrigger after update")
	}

	// Clean shutdown
	close(psControl.shutdownCh)
}

func TestPscDesc_applyDigitalTrigger(t *testing.T) {
	con := genericps.NewConnection()
	handle, err := genericps.OpenDemo(con, genericps.DemoId)
	if err != nil {
		t.Fatalf("failed to open demo: %v", err)
	}
	con.Handle = handle
	defer con.CloseUnit()

	ps := &PscDesc{Con: con}
	ps.triggerSetting.DigitalTriggerEnabled = true
	ps.triggerSetting.DigitalAnalogOperand = genericps.OperandOr
	ps.triggerSetting.DigitalDirections = []genericps.DigitalChannelDirections{
		{Channel: genericps.Dch0, Direction: genericps.DigitalDirectionRising},
	}

	err = ps.applyDigitalTrigger()
	if err != nil {
		t.Errorf("applyDigitalTrigger failed: %v", err)
	}

	// Disabled digital trigger
	ps.triggerSetting.DigitalTriggerEnabled = false
	err = ps.applyDigitalTrigger()
	if err != nil {
		t.Errorf("applyDigitalTrigger (disabled) failed: %v", err)
	}
}

func captureDriverMessages(t *testing.T, id string) (*genericps.Connection, func() []genericps.Message) {
	t.Helper()
	con := genericps.NewConnection()
	con.Handle = 1
	con.ID = id
	var mu sync.Mutex
	var msgs []genericps.Message
	done := make(chan struct{})
	t.Cleanup(func() {
		close(done)
	})
	go func() {
		for {
			select {
			case <-done:
				return
			case msg := <-con.MsgCh:
				mu.Lock()
				msgs = append(msgs, msg)
				mu.Unlock()
				msg.RspCh() <- struct{}{}
			}
		}
	}()
	return con, func() []genericps.Message {
		mu.Lock()
		defer mu.Unlock()
		out := make([]genericps.Message, len(msgs))
		copy(out, msgs)
		return out
	}
}

func TestSendSimpleTrigger_ETSFallingUsesLevelWithAlignedLower(t *testing.T) {
	tests := []struct {
		name           string
		driverID       string
		mode           TriggerModes
		enabled        bool
		dir            genericps.ThresholdDirection
		triggerADC     int16
		wantSimple     bool
		wantUpperHyst  uint16
	}{
		{
			name:          "ETS falling non-demo negative threshold uses Level props with ThresholdLower==ThresholdUpper",
			driverID:      "ps2000a",
			mode:          ETS,
			enabled:       true,
			dir:           genericps.TriggerFalling,
			triggerADC:    -3000,
			wantSimple:    false,
			wantUpperHyst: defaultSimpleTriggerHysteresisADC,
		},
		{
			name:          "ETS falling near upper rail clamps hysteresis cleanly",
			driverID:      "ps2000a",
			mode:          ETS,
			enabled:       true,
			dir:           genericps.TriggerFalling,
			triggerADC:    32700,
			wantSimple:    false,
			wantUpperHyst: 67,
		},
		{
			name:          "ETS falling demo keeps zero hysteresis while aligning ThresholdLower",
			driverID:      genericps.DemoId,
			mode:          ETS,
			enabled:       true,
			dir:           genericps.TriggerFalling,
			triggerADC:    -3000,
			wantSimple:    false,
			wantUpperHyst: 0,
		},
		{
			name:       "Repeat mode still uses SetSimpleTrigger",
			driverID:   "ps2000a",
			mode:       Repeat,
			enabled:    true,
			dir:        genericps.TriggerFalling,
			triggerADC: -3000,
			wantSimple: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			con, collected := captureDriverMessages(t, tt.driverID)
			ps := &PscDesc{
				Con:        con,
				ScopeModel: ScopeUnknown,
			}
			ps.triggerSetting = TriggerDesc{
				Enabled:            tt.enabled,
				Source:             genericps.ChA,
				TriggerADC:         tt.triggerADC,
				ThresholdDirection: tt.dir,
				Mode:               tt.mode,
				Type:               Simple,
			}

			if err := ps.sendSimpleTrigger(); err != nil {
				t.Fatalf("sendSimpleTrigger failed: %v", err)
			}

			msgs := collected()
			var gotSimple []*genericps.SetSimpleTriggerMsg
			var gotProps []*genericps.SetTriggerChannelPropertiesMsg
			var gotConds []*genericps.SetTriggerChannelConditionsMsg
			var gotDirs []*genericps.SetTriggerChannelDirectionsMsg
			var gotPwq []*genericps.SetPulseWidthQualifierMsg
			for _, m := range msgs {
				switch v := m.(type) {
				case *genericps.SetSimpleTriggerMsg:
					gotSimple = append(gotSimple, v)
				case *genericps.SetTriggerChannelPropertiesMsg:
					gotProps = append(gotProps, v)
				case *genericps.SetTriggerChannelConditionsMsg:
					gotConds = append(gotConds, v)
				case *genericps.SetTriggerChannelDirectionsMsg:
					gotDirs = append(gotDirs, v)
				case *genericps.SetPulseWidthQualifierMsg:
					gotPwq = append(gotPwq, v)
				}
			}

			if tt.wantSimple {
				if len(gotSimple) != 1 || len(gotProps) != 0 {
					t.Fatalf("expected 1 SetSimpleTrigger and 0 SetTriggerChannelProperties, got %d simple, %d props", len(gotSimple), len(gotProps))
				}
				return
			}

			if len(gotSimple) != 0 {
				t.Fatalf("expected 0 SetSimpleTrigger calls in ETS enabled mode, got %d", len(gotSimple))
			}
			if len(gotProps) != 1 || len(gotProps[0].ChannelProperties) != 1 {
				t.Fatalf("expected 1 SetTriggerChannelProperties with 1 entry, got %+v", gotProps)
			}
			p := gotProps[0].ChannelProperties[0]
			if p.ThresholdMode != genericps.Level {
				t.Errorf("expected Level threshold mode, got %v", p.ThresholdMode)
			}
			if p.ThresholdUpper != tt.triggerADC || p.ThresholdLower != tt.triggerADC {
				t.Errorf("expected ThresholdUpper=ThresholdLower=%d, got upper=%d lower=%d", tt.triggerADC, p.ThresholdUpper, p.ThresholdLower)
			}
			if p.ThresholdUpperHysteresis != tt.wantUpperHyst || p.ThresholdLowerHysteresis != 0 {
				t.Errorf("expected hyst=(%d, 0), got (%d, %d)", tt.wantUpperHyst, p.ThresholdUpperHysteresis, p.ThresholdLowerHysteresis)
			}
			if gotProps[0].AutoTriggerMs != 0 {
				t.Errorf("expected AutoTriggerMs=0 in ETS, got %d", gotProps[0].AutoTriggerMs)
			}
			if len(gotConds) != 1 || len(gotConds[0].TriggerConditions) != 1 || gotConds[0].TriggerConditions[0].ChannelA != genericps.CondTrue {
				t.Errorf("expected ChannelA CondTrue trigger condition, got %+v", gotConds)
			}
			if len(gotDirs) != 1 || gotDirs[0].ChannelA != tt.dir {
				t.Errorf("expected ChannelA direction %v, got %+v", tt.dir, gotDirs)
			}
			if len(gotPwq) != 1 || gotPwq[0].PwType != genericps.PwTypeNone {
				t.Errorf("expected PWQ disabled with PwTypeNone, got %+v", gotPwq)
			}
		})
	}
}

func TestSendAdvancedTrigger_LevelAlignsThresholdLower(t *testing.T) {
	con, collected := captureDriverMessages(t, "ps2000a")
	ps := &PscDesc{
		Con:        con,
		ScopeModel: ScopeUnknown,
	}
	ps.triggerSetting = TriggerDesc{
		Enabled:            true,
		Source:             genericps.ChA,
		TriggerADC:         4500,
		LowerTriggerADC:    0,
		HysteresisADC:      200,
		LowerHysteresisADC: 50,
		ThresholdDirection: genericps.TriggerFalling,
		ThresholdMode:      genericps.Level,
		Mode:               ETS,
		Type:               Advanced,
	}

	if err := ps.sendAdvancedTrigger(); err != nil {
		t.Fatalf("sendAdvancedTrigger failed: %v", err)
	}

	var gotProps []*genericps.SetTriggerChannelPropertiesMsg
	for _, m := range collected() {
		if v, ok := m.(*genericps.SetTriggerChannelPropertiesMsg); ok {
			gotProps = append(gotProps, v)
		}
	}
	if len(gotProps) != 1 || len(gotProps[0].ChannelProperties) != 1 {
		t.Fatalf("expected 1 SetTriggerChannelProperties with 1 entry, got %+v", gotProps)
	}
	p := gotProps[0].ChannelProperties[0]
	if p.ThresholdUpper != 4500 || p.ThresholdLower != 4500 {
		t.Errorf("expected ThresholdUpper=ThresholdLower=4500, got upper=%d lower=%d", p.ThresholdUpper, p.ThresholdLower)
	}
	if p.ThresholdUpperHysteresis != 200 || p.ThresholdLowerHysteresis != 0 {
		t.Errorf("expected hyst=(200, 0), got (%d, %d)", p.ThresholdUpperHysteresis, p.ThresholdLowerHysteresis)
	}
}

func TestSendComplexTrigger_LevelAlignsThresholdLower(t *testing.T) {
	con, collected := captureDriverMessages(t, "ps2000a")
	ps := &PscDesc{
		Con:        con,
		ScopeModel: ScopeUnknown,
	}
	ps.triggerSetting = TriggerDesc{
		Enabled: true,
		Mode:    ETS,
		Type:    Complex,
		ComplexProperties: []genericps.TriggerChannelProperties{
			{
				Channel:                  genericps.ChA,
				ThresholdUpper:           -2500,
				ThresholdUpperHysteresis: 150,
				ThresholdLower:           0,
				ThresholdLowerHysteresis: 50,
				ThresholdMode:            genericps.Level,
			},
			{
				Channel:                  genericps.ChB,
				ThresholdUpper:           1200,
				ThresholdUpperHysteresis: 100,
				ThresholdLower:           -1200,
				ThresholdLowerHysteresis: 80,
				ThresholdMode:            genericps.Level,
			},
		},
		ComplexDirections: []TriggerDirections{
			{
				ChannelA: genericps.TriggerFalling,
				ChannelB: genericps.TriggerFallingLower,
			},
		},
	}

	if err := ps.sendComplexTrigger(); err != nil {
		t.Fatalf("sendComplexTrigger failed: %v", err)
	}

	var gotProps []*genericps.SetTriggerChannelPropertiesMsg
	for _, m := range collected() {
		if v, ok := m.(*genericps.SetTriggerChannelPropertiesMsg); ok {
			gotProps = append(gotProps, v)
		}
	}
	if len(gotProps) != 1 || len(gotProps[0].ChannelProperties) != 2 {
		t.Fatalf("expected 1 SetTriggerChannelProperties with 2 entries, got %+v", gotProps)
	}
	chA := gotProps[0].ChannelProperties[0]
	if chA.ThresholdLower != -2500 || chA.ThresholdLowerHysteresis != 0 {
		t.Errorf("ChA Level Falling expected ThresholdLower=-2500 hyst=0, got lower=%d hyst=%d", chA.ThresholdLower, chA.ThresholdLowerHysteresis)
	}
	chB := gotProps[0].ChannelProperties[1]
	if chB.ThresholdLower != -1200 || chB.ThresholdLowerHysteresis != 80 {
		t.Errorf("ChB FallingLower expected ThresholdLower=-1200 hyst=80 preserved, got lower=%d hyst=%d", chB.ThresholdLower, chB.ThresholdLowerHysteresis)
	}
}
