package util

import (
	"fmt"
	"os"
	"path/filepath"
)

// FileExists checks if a file exists at the specified file path.
// It returns an error if the file does not exist, otherwise it returns nil.
// The function uses os.Stat to check the file's existence.
//
// Parameters:
//   - filePath: The full path to the file to be checked
//
// Returns:
//   - error: Nil if the file exists, otherwise an error describing the non-existence
func FileExists(filePath string) error {
	_, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		return fmt.Errorf("file does not exist: %s", filePath)
	}

	return nil
}

// FileHasExtension verifies if a file has the expected file extension.
// It compares the file's extension against the provided expected extension.
// The comparison is case-sensitive.
//
// Parameters:
//   - filePath: The full path to the file to be checked
//   - ext: The expected file extension (including the dot, e.g., ".txt")
//
// Returns:
//   - error: Nil if the file extension matches, otherwise an error describing the mismatch
func FileHasExtension(filePath string, ext string) error {
	if filepath.Ext(filePath) != ext {
		return fmt.Errorf("invalid file extension. Expected %s, got %s", ext, filepath.Ext(filePath))
	}
	return nil
}

// FileIsNotEmpty checks if a file exists and is not empty.
// Currently, this implementation only checks for file existence and does not verify file size.
//
// Parameters:
//   - filePath: The full path to the file to be checked
//
// Returns:
//   - error: Nil if the file exists, otherwise an error describing the non-existence
func FileIsNotEmpty(filePath string) error {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("file does not exist: %s", filePath)
	}

	return nil
}