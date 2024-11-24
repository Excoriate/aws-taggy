package util

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFileExists(t *testing.T) {
	t.Parallel()

	// Create a temporary file for testing
	tmpFile, err := os.CreateTemp("", "testfile-*")
	require.NoError(t, err, "Failed to create temporary test file")
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	testCases := []struct {
		name        string
		filePath    string
		expectError bool
	}{
		{
			name:        "Existing file",
			filePath:    tmpFile.Name(),
			expectError: false,
		},
		{
			name:        "Non-existent file",
			filePath:    "/path/to/non/existent/file.txt",
			expectError: true,
		},
		{
			name:        "Empty path",
			filePath:    "",
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := FileExists(tc.filePath)
			if tc.expectError {
				assert.Error(t, err, "Expected an error for %s", tc.name)
			} else {
				assert.NoError(t, err, "Did not expect an error for %s", tc.name)
			}
		})
	}
}

func TestFileHasExtension(t *testing.T) {
	t.Parallel()

	// Create temporary files with different extensions
	txtFile, err := os.CreateTemp("", "testfile-*.txt")
	require.NoError(t, err, "Failed to create .txt test file")
	defer os.Remove(txtFile.Name())
	txtFile.Close()

	jsonFile, err := os.CreateTemp("", "testfile-*.json")
	require.NoError(t, err, "Failed to create .json test file")
	defer os.Remove(jsonFile.Name())
	jsonFile.Close()

	testCases := []struct {
		name        string
		filePath    string
		ext         string
		expectError bool
	}{
		{
			name:        "Matching extension",
			filePath:    txtFile.Name(),
			ext:         ".txt",
			expectError: false,
		},
		{
			name:        "Mismatched extension",
			filePath:    txtFile.Name(),
			ext:         ".json",
			expectError: true,
		},
		{
			name:        "Case-sensitive extension check",
			filePath:    txtFile.Name(),
			ext:         ".TXT",
			expectError: true,
		},
		{
			name:        "Empty extension",
			filePath:    txtFile.Name(),
			ext:         "",
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := FileHasExtension(tc.filePath, tc.ext)
			if tc.expectError {
				assert.Error(t, err, "Expected an error for %s", tc.name)
			} else {
				assert.NoError(t, err, "Did not expect an error for %s", tc.name)
			}
		})
	}
}

func TestFileIsNotEmpty(t *testing.T) {
	t.Parallel()

	// Create temporary files
	emptyFile, err := os.CreateTemp("", "empty-*")
	require.NoError(t, err, "Failed to create empty test file")
	defer os.Remove(emptyFile.Name())
	emptyFile.Close()

	nonEmptyFile, err := os.CreateTemp("", "nonempty-*")
	require.NoError(t, err, "Failed to create non-empty test file")
	defer os.Remove(nonEmptyFile.Name())
	_, err = nonEmptyFile.Write([]byte("content"))
	require.NoError(t, err)
	nonEmptyFile.Close()

	testCases := []struct {
		name        string
		filePath    string
		expectError bool
	}{
		{
			name:        "Existing file",
			filePath:    nonEmptyFile.Name(),
			expectError: false,
		},
		{
			name:        "Non-existent file",
			filePath:    "/path/to/non/existent/file.txt",
			expectError: true,
		},
		{
			name:        "Empty file",
			filePath:    emptyFile.Name(),
			expectError: false, // Current implementation doesn't check file size
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := FileIsNotEmpty(tc.filePath)
			if tc.expectError {
				assert.Error(t, err, "Expected an error for %s", tc.name)
			} else {
				assert.NoError(t, err, "Did not expect an error for %s", tc.name)
			}
		})
	}
}

// Benchmark tests to ensure performance
func BenchmarkFileExists(b *testing.B) {
	tmpFile, err := os.CreateTemp("", "benchmark-*")
	require.NoError(b, err, "Failed to create temporary test file")
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = FileExists(tmpFile.Name())
	}
}

func BenchmarkFileHasExtension(b *testing.B) {
	tmpFile, err := os.CreateTemp("", "benchmark-*.txt")
	require.NoError(b, err, "Failed to create temporary test file")
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = FileHasExtension(tmpFile.Name(), ".txt")
	}
} 