package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"fynescope/genericps"
	"strings"
	"testing"
)

// No MockConnection needed as we use genericps directly for tests here or mock at lower level.

func TestOpenSimulator(t *testing.T) {
	tests := []struct {
		name           string
		id             string
		expectedHandle int16
		expectedError  string
	}{
		{
			name:           "Successful Open Sim",
			id:             genericps.SimId,
			expectedHandle: 1, // first handle is usually 1 from uniqueHandle()
			expectedError:  "",
		},
		{
			name:           "Simulator Not Found",
			id:             "wrong_id",
			expectedHandle: 0,
			expectedError:  "Simulator not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			con, err := openSimulator(tt.id)

			if tt.expectedError != "" {
				if err == nil {
					t.Fatalf("Expected error: %v, got nil", tt.expectedError)
				}
				if !strings.Contains(err.Error(), tt.expectedError) {
					t.Fatalf("Expected error to contain: %s, got: %s", tt.expectedError, err.Error())
				}
			} else {
				if err != nil {
					t.Fatalf("Expected no error, got: %v", err)
				}
			}

			if con != nil {
				if tt.expectedHandle > 0 && con.Handle <= 0 {
					t.Errorf("Expected non-zero handle, got: %v", con.Handle)
				} else if tt.expectedHandle == 0 && con.Handle != 0 {
					t.Errorf("Expected zero handle, got: %v", con.Handle)
				}
			}
		})
	}
}

func TestSubpackages(t *testing.T) {
	// Find subdirectories with test files
	var subdirs []string
	err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			// Skip hidden directories and vendor directories
			if path != "." && (strings.HasPrefix(path, ".") || strings.Contains(path, "/.")) {
				return filepath.SkipDir
			}
			return nil
		}
		// If it's a test file in a subdirectory, add the directory to subdirs
		if strings.HasSuffix(path, "_test.go") {
			dir := filepath.Dir(path)
			if dir != "." {
				// Avoid duplicate entries
				alreadyAdded := false
				for _, d := range subdirs {
					if d == dir {
						alreadyAdded = true
						break
					}
				}
				if !alreadyAdded {
					subdirs = append(subdirs, "./"+dir)
				}
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("Failed to walk directory: %v", err)
	}

	if len(subdirs) == 0 {
		t.Log("No subpackages with tests found")
		return
	}

	// Run go test for each subdirectory as a subtest
	for _, subdir := range subdirs {
		subdir := subdir // capture loop variable
		t.Run(subdir, func(t *testing.T) {
			cmd := exec.Command("go", "test", "-count=1", subdir)
			output, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("Test failed for %s:\n%s", subdir, string(output))
			}
			t.Logf("Test passed for %s:\n%s", subdir, string(output))
		})
	}
}

