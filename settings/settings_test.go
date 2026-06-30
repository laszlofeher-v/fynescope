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
	s.Window.SimGenActiveTab = 3
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
	assert.Equal(t, 3, loaded.Window.SimGenActiveTab)
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
