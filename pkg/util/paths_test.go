package util

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveAbsolutePath(t *testing.T) {
	// Get current working directory for creating cross-platform absolute paths
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current working directory: %v", err)
	}

	testCases := []struct {
		name        string
		inputPath   string
		wantErr     bool
		pathChecker func(string) bool
	}{
		{
			name:      "Relative path with current directory",
			inputPath: "./test_config.yaml",
			wantErr:   false,
			pathChecker: func(absPath string) bool {
				return filepath.IsAbs(absPath) &&
					filepath.Base(absPath) == "test_config.yaml"
			},
		},
		{
			name:      "Relative path with parent directory",
			inputPath: "../config/app.yaml",
			wantErr:   false,
			pathChecker: func(absPath string) bool {
				return filepath.IsAbs(absPath) &&
					filepath.Base(absPath) == "app.yaml"
			},
		},
		{
			name:      "Absolute path",
			inputPath: filepath.Join(cwd, "etc", "myapp", "config.yaml"),
			wantErr:   false,
			pathChecker: func(absPath string) bool {
				return filepath.IsAbs(absPath) &&
					filepath.Base(absPath) == "config.yaml"
			},
		},
		{
			name:      "Empty path",
			inputPath: "",
			wantErr:   false,
			pathChecker: func(absPath string) bool {
				return filepath.IsAbs(absPath)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			absPath, err := ResolveAbsolutePath(tc.inputPath)

			// Check for unexpected errors
			if tc.wantErr && err == nil {
				t.Errorf("Expected an error, but got none")
			}

			// Check for unexpected lack of error
			if !tc.wantErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			// If no error, perform additional path checks
			if err == nil {
				if !tc.pathChecker(absPath) {
					t.Errorf("Path check failed for input %q. Got: %q",
						tc.inputPath, absPath)
				}
			}
		})
	}
}
