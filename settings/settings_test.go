package settings

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewDefaultSettings(t *testing.T) {
	s := NewDefaultSettings()
	assert.NotNil(t, s)
	assert.Len(t, s.Channels, 4)
	assert.True(t, s.Channels[0].Enabled)
	assert.False(t, s.Channels[1].Enabled)
	assert.Equal(t, float64(1000), s.GenPanel.Frequency)
	assert.NotNil(t, s.StreamEnabled)
	assert.True(t, *s.StreamEnabled)
}

func TestSaveAndLoadSettings(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "settings.yaml")

	s := NewDefaultSettings()
	s.Time.TimeDiv = "10"
	s.GenPanel.Frequency = 2500
	s.Window.FilterActiveTab = 2
	s.Window.DemoGenActiveTab = 3
	streamVal := false
	s.StreamEnabled = &streamVal

	err := Save(filePath, s)
	assert.NoError(t, err)

	loaded, err := Load(filePath)
	assert.NoError(t, err)
	assert.NotNil(t, loaded)
	assert.Equal(t, "10", loaded.Time.TimeDiv)
	assert.Equal(t, 2500.0, loaded.GenPanel.Frequency)
	assert.Equal(t, 2, loaded.Window.FilterActiveTab)
	assert.Equal(t, 3, loaded.Window.DemoGenActiveTab)
	assert.Len(t, loaded.Channels, 4)
	assert.NotNil(t, loaded.StreamEnabled)
	assert.False(t, *loaded.StreamEnabled)
}

func TestLoad_InvalidChecksum(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "settings.yaml")

	s := NewDefaultSettings()
	err := Save(filePath, s)
	assert.NoError(t, err)

	// Read file, corrupt one byte of YAML content, and write it back
	content, err := os.ReadFile(filePath)
	assert.NoError(t, err)
	assert.True(t, len(content) > 100)

	// Corrupt a byte in the YAML section (after the first 65 bytes)
	content[len(content)-10] = 'X'

	// Temporarily make it writable to overwrite
	_ = os.Chmod(filePath, 0644)
	err = os.WriteFile(filePath, content, 0644)
	assert.NoError(t, err)

	_, err = Load(filePath)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "checksum mismatch")
}

func TestLoad_FileTooShort(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "settings.yaml")

	err := os.WriteFile(filePath, []byte("short"), 0644)
	assert.NoError(t, err)

	_, err = Load(filePath)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "settings file too short")
}

func TestLoad_NonExistentFile(t *testing.T) {
	_, err := Load("/path/to/nonexistent/file.yaml")
	assert.Error(t, err)
}

func TestSaveAndLoad_Waveforms(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "scopesettings.yaml")

	s := NewDefaultSettings()
	s.GenPanel.ArbitraryWaveform = []int16{100, -200, 300, -400}
	s.DemoGenPanel[0].ArbitraryWaveform = []int16{50, 60, 70}

	err := Save(filePath, s)
	assert.NoError(t, err)

	loaded, err := Load(filePath)
	assert.NoError(t, err)
	assert.NotNil(t, loaded)
	assert.Equal(t, []int16{100, -200, 300, -400}, loaded.GenPanel.ArbitraryWaveform)
	assert.Equal(t, []int16{50, 60, 70}, loaded.DemoGenPanel[0].ArbitraryWaveform)

	// Now save with empty waveforms to trigger removal
	s.GenPanel.ArbitraryWaveform = nil
	s.DemoGenPanel[0].ArbitraryWaveform = nil
	err = Save(filePath, s)
	assert.NoError(t, err)

	wfFile := WaveformFileName(filePath)
	_, err = os.Stat(wfFile)
	assert.True(t, os.IsNotExist(err))

	demoWfFile := WaveformDemoFileName(filePath, 0)
	_, err = os.Stat(demoWfFile)
	assert.True(t, os.IsNotExist(err))
}

func TestLoad_DefaultsAndFallbacks(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "scopesettings.yaml")

	s := &PsSettings{
		Trigger: TriggerSettings{
			Type: TriggerTypeComplex,
		},
		Channels: []ChSettings{
			{}, // all digital filter values 0
		},
		Digital: DigitalSettings{
			ChannelsEnabled: [16]bool{}, // all false
		},
		StreamEnabled: nil,
		DigitalDemoGenPanel: DigitalDemoGenSettings{
			Frequency: 0,
		},
		ScreenSize: "",
		Decode: DecodeSettings{
			DataBits: 0,
			StopBits: "",
			Parity:   "",
		},
	}

	err := Save(filePath, s)
	assert.NoError(t, err)

	loaded, err := Load(filePath)
	assert.NoError(t, err)
	assert.NotNil(t, loaded)

	// Verify fallback corrections
	assert.Equal(t, TriggerTypeAdvanced, loaded.Trigger.Type)
	assert.True(t, loaded.Trigger.ComplexEnabled)

	assert.Equal(t, 10000.0, loaded.Channels[0].DigitalFilter.LowpassFc)
	assert.Equal(t, 100.0, loaded.Channels[0].DigitalFilter.HighpassFc)
	assert.Equal(t, 500.0, loaded.Channels[0].DigitalFilter.BandpassFc1)
	assert.Equal(t, 5000.0, loaded.Channels[0].DigitalFilter.BandpassFc2)
	assert.Equal(t, 900.0, loaded.Channels[0].DigitalFilter.BandstopFc1)
	assert.Equal(t, 1100.0, loaded.Channels[0].DigitalFilter.BandstopFc2)

	for i := range loaded.Digital.ChannelsEnabled {
		assert.True(t, loaded.Digital.ChannelsEnabled[i])
	}

	assert.NotNil(t, loaded.StreamEnabled)
	assert.True(t, *loaded.StreamEnabled)
	assert.Equal(t, defaultFrequency, loaded.DigitalDemoGenPanel.Frequency)
	assert.Equal(t, ScreenSize1920x1080, loaded.ScreenSize)
	assert.Equal(t, 8, loaded.Decode.DataBits)
	assert.Equal(t, "1", loaded.Decode.StopBits)
	assert.Equal(t, "None", loaded.Decode.Parity)
	assert.GreaterOrEqual(t, len(loaded.DemoGenPanel), 4)
}

func TestWaveformFileNames(t *testing.T) {
	assert.Equal(t, "waveform.bin", filepath.Base(WaveformFileName("scopesettings.yaml")))
	assert.Equal(t, "waveform_custom.bin", filepath.Base(WaveformFileName("custom.yaml")))
	assert.Equal(t, "waveform_demo0.bin", filepath.Base(WaveformDemoFileName("scopesettings.yaml", 0)))
	assert.Equal(t, "waveform_demo2_1_1.bin", filepath.Base(WaveformDemoFileName("scopesettings_1_1.yaml", 2)))
	assert.Equal(t, "waveform_demo1.bin", filepath.Base(WaveformDemoFileName("custom.yaml", 1)))
}
