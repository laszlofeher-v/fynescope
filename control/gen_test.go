package control

import (
	"fynescope/genericps"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// --- Test Setup ---

func setupPscDescForGenTest() (*PscDesc, *genericps.Connection) {
	con := genericps.NewConnection()

	psControl := &PscDesc{
		Con:            con,
		stateChannel:   make(chan state),
		stopChannel:    make(chan struct{}, 1),
		restartChannel: make(chan struct{}, 1),

		SetGeneratorCh: make(chan *GeneratorDescMsg),
		getGeneratorCh: make(chan *getGeneratorMsg),
		getGenerator: getGeneratorMsg{
			newSetting: make(chan bool),
		},
	}

	return psControl, con
}

// --- Generator Monitor Tests ---

func TestGeneratorMonitor_InitialState(t *testing.T) {
	psControl, _ := setupPscDescForGenTest()
	go psControl.generatorMonitor()
	defer func() { psControl.stopChannel <- struct{}{} }() // Ensure monitor stops

	// Test getting settings when unchanged
	psControl.getGeneratorCh <- &psControl.getGenerator
	select {
	case newData := <-psControl.getGenerator.newSetting:
		assert.False(t, newData, "Should indicate no new data initially")
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Timeout waiting for getGenerator response")
	}
}

func TestGeneratorMonitor_SetNewSettingWhenUnchanged(t *testing.T) {
	psControl, _ := setupPscDescForGenTest()
	go psControl.generatorMonitor()
	defer func() { psControl.stopChannel <- struct{}{} }()

	newSetting := GeneratorDesc{OffsetVoltage: 100, PkToPK: 2000, WaveType: genericps.Sine}
	msg := &GeneratorDescMsg{GeneratorDesc: newSetting}

	// Send new setting
	psControl.SetGeneratorCh <- msg
	time.Sleep(50 * time.Millisecond) // Allow monitor to process

	// Verify restart was requested
	select {
	case <-psControl.restartChannel:
		// Success
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Timeout waiting for restart")
	}

	// Test getting settings when changed
	psControl.getGeneratorCh <- &psControl.getGenerator
	select {
	case newData := <-psControl.getGenerator.newSetting:
		assert.True(t, newData, "Should indicate new data after setting")
		assert.NotNil(t, psControl.getGenerator.generatorSettings, "Generator settings should be populated")
		assert.Equal(t, newSetting, *psControl.getGenerator.generatorSettings, "Stored setting mismatch")
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Timeout waiting for getGenerator response")
	}

	// Test getting settings again (should be unchanged now)
	psControl.getGeneratorCh <- &psControl.getGenerator
	select {
	case newData := <-psControl.getGenerator.newSetting:
		assert.False(t, newData, "Should indicate no new data after getting")
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Timeout waiting for getGenerator response")
	}
}

func TestGeneratorMonitor_SetSameSettingWhenUnchanged(t *testing.T) {
	psControl, _ := setupPscDescForGenTest()
	initialSetting := GeneratorDesc{OffsetVoltage: 0, PkToPK: 1000, WaveType: genericps.Square}

	go psControl.generatorMonitor()
	
	// Send initial setting to populate storedSetting
	psControl.SetGeneratorCh <- &GeneratorDescMsg{GeneratorDesc: initialSetting}
	time.Sleep(50 * time.Millisecond) // Allow processing

	// DRAIN the restartChannel and getGenerator to make the monitor UNCHANGED again
	select {
	case <-psControl.restartChannel:
	default:
	}
	psControl.getGeneratorCh <- &psControl.getGenerator
	select {
	case newData := <-psControl.getGenerator.newSetting:
		assert.True(t, newData, "Initial setup should be changed")
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Timeout waiting for getGenerator response")
	}

	// Now send the same setting again
	msg := &GeneratorDescMsg{GeneratorDesc: initialSetting}
	psControl.SetGeneratorCh <- msg
	time.Sleep(50 * time.Millisecond) // Allow monitor to process

	// Verify requestRestart was NOT called
	select {
	case <-psControl.restartChannel:
		t.Fatal("requestRestart should NOT be called when setting is the same")
	default:
		// Success
	}

	// Test getting settings (should still be unchanged)
	psControl.getGeneratorCh <- &psControl.getGenerator
	select {
	case newData := <-psControl.getGenerator.newSetting:
		assert.False(t, newData, "Should indicate no new data")
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Timeout waiting for getGenerator response")
	}
	psControl.stopChannel <- struct{}{} // Stop monitor
}

func TestGeneratorMonitor_SetNewSettingWhenChanged(t *testing.T) {
	psControl, _ := setupPscDescForGenTest()
	go psControl.generatorMonitor()
	defer func() { psControl.stopChannel <- struct{}{} }()

	setting1 := GeneratorDesc{OffsetVoltage: 100, PkToPK: 2000, WaveType: genericps.Sine}
	setting2 := GeneratorDesc{OffsetVoltage: 200, PkToPK: 3000, WaveType: genericps.Triangle}

	// 1. Make it changed
	psControl.SetGeneratorCh <- &GeneratorDescMsg{GeneratorDesc: setting1}
	time.Sleep(50 * time.Millisecond) // Allow processing

	select {
	case <-psControl.restartChannel:
		// Success
	default:
		t.Fatal("requestRestart should be called for setting1")
	}

	// 2. Send another new setting while still changed
	psControl.SetGeneratorCh <- &GeneratorDescMsg{GeneratorDesc: setting2}
	time.Sleep(50 * time.Millisecond) // Allow processing

	select {
	case <-psControl.restartChannel:
		// Success
	default:
		t.Fatal("requestRestart should be called again for setting2")
	}

	// 3. Get the setting (should be the latest one, setting2)
	psControl.getGeneratorCh <- &psControl.getGenerator
	select {
	case newData := <-psControl.getGenerator.newSetting:
		assert.True(t, newData, "Should indicate new data")
		assert.NotNil(t, psControl.getGenerator.generatorSettings, "Generator settings should be populated")
		assert.Equal(t, setting2, *psControl.getGenerator.generatorSettings, "Stored setting should be the latest (setting2)")
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Timeout waiting for getGenerator response")
	}

	// 4. Get again (should be unchanged now)
	psControl.getGeneratorCh <- &psControl.getGenerator
	select {
	case newData := <-psControl.getGenerator.newSetting:
		assert.False(t, newData, "Should indicate no new data after getting")
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Timeout waiting for getGenerator response")
	}
}

// --- Set Generator Function Tests ---

func TestSetGenerator_WhenChanged(t *testing.T) {
	psControl, con := setupPscDescForGenTest()

	storedSetting := GeneratorDesc{
		OffsetVoltage:  50,
		PkToPK:         1500,
		WaveType:       genericps.DcVoltage,
		StartFrequency: 1000,
		StopFrequency:  1000,
		Increment:      0,
		DwellTime:      1,
		SweepType:      genericps.SweepUp,
		Operation:      genericps.EsOff,
		Shots:          1,
		Sweeps:         0,
		TriggerType:    genericps.SigGenRising,
		TriggerSource:  genericps.SigGenNone,
		ExtInThreshold: 0,
	}

	var wg sync.WaitGroup
	wg.Add(2)

	// Intercept SetSigGenBuiltInV2Msg
	go func() {
		defer wg.Done()
		select {
		case msg := <-con.MsgCh:
			v2Msg, ok := msg.(*genericps.SetSigGenBuiltInV2Msg)
			if !ok {
				t.Errorf("Expected SetSigGenBuiltInV2Msg, got %T", msg)
			} else {
				assert.Equal(t, storedSetting.OffsetVoltage, v2Msg.OffsetVoltage)
				assert.Equal(t, storedSetting.PkToPK, v2Msg.PkToPK)
				assert.Equal(t, storedSetting.WaveType, v2Msg.WaveType)
				assert.Equal(t, storedSetting.StartFrequency, v2Msg.StartFrequency)
				assert.Equal(t, storedSetting.StopFrequency, v2Msg.StopFrequency)
				assert.Equal(t, storedSetting.Increment, v2Msg.Increment)
				assert.Equal(t, storedSetting.DwellTime, v2Msg.DwellTime)
				assert.Equal(t, storedSetting.SweepType, v2Msg.SweepType)
				assert.Equal(t, storedSetting.Operation, v2Msg.Operation)
				assert.Equal(t, storedSetting.Shots, v2Msg.Shots)
				assert.Equal(t, storedSetting.Sweeps, v2Msg.Sweeps)
				assert.Equal(t, storedSetting.TriggerType, v2Msg.TriggerType)
				assert.Equal(t, storedSetting.TriggerSource, v2Msg.TriggerSource)
				assert.Equal(t, storedSetting.ExtInThreshold, v2Msg.ExtInThreshold)
			}
			msg.RspCh() <- struct{}{}
		case <-time.After(1 * time.Second):
			t.Error("Timeout waiting for message on Con.MsgCh")
		}
	}()

	// Simulate the monitor being changed and providing data
	go func() {
		defer wg.Done()
		select {
		case req := <-psControl.getGeneratorCh:
			req.generatorSettings = &storedSetting // Provide the data
			req.newSetting <- true                 // Indicate new data is available
		case <-time.After(200 * time.Millisecond):
			t.Error("Timeout waiting for setGenerator to request data")
		}
	}()

	// Call the function under test
	err := psControl.setGenerator()
	assert.NoError(t, err)

	// Wait for the goroutines to finish
	wg.Wait()
}

func TestSetGenerator_WhenUnchanged(t *testing.T) {
	psControl, con := setupPscDescForGenTest()

	var wg sync.WaitGroup
	wg.Add(1)

	// Intercept SetSigGenBuiltInV2Msg - IT SHOULD NOT BE CALLED
	go func() {
		select {
		case msg := <-con.MsgCh:
			t.Errorf("Unexpected message on MsgCh: %T", msg)
			msg.RspCh() <- struct{}{}
		case <-time.After(100 * time.Millisecond):
			// Success: no message received
		}
	}()

	// Simulate the monitor being unchanged
	go func() {
		defer wg.Done()
		select {
		case req := <-psControl.getGeneratorCh:
			req.generatorSettings = nil // No data to provide
			req.newSetting <- false     // Indicate no new data
		case <-time.After(200 * time.Millisecond):
			t.Error("Timeout waiting for setGenerator to request data")
		}
	}()

	// Call the function under test
	err := psControl.setGenerator()
	assert.NoError(t, err)

	// Wait for the monitor goroutine to finish responding
	wg.Wait()
}
