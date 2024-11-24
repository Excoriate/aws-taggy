package util

import (
	"fmt"
	"path/filepath"
)

// ResolveAbsolutePath resolves the absolute path of the configuration file
// ResolveAbsolutePath converts a potentially relative file path to an absolute path.
//
// This function takes a configuration file path as input and returns its absolute path.
// It handles both relative and absolute input paths, converting them to a fully
// qualified absolute path that can be used consistently across different working directories.
//
// Parameters:
//   - configPath: A string representing the file path to be resolved. This can be
//     a relative or absolute path.
//
// Returns:
//   - A string containing the fully resolved absolute path.
//   - An error if the path cannot be resolved, which could occur due to filesystem
//     access issues or invalid path formats.
//
// Example:
//   absPath, err := ResolveAbsolutePath("./config/app.yaml")
//   if err != nil {
//       // Handle error
//   }
//   // absPath now contains the full absolute path
func ResolveAbsolutePath(configPath string) (string, error) {
	absPath, err := filepath.Abs(configPath)
	if err != nil {
		return "", fmt.Errorf("failed to resolve absolute path: %w", err)
	}

	return absPath, nil
}
