package util

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper function to set and restore environment variables
func setAndRestoreEnv(t *testing.T, key, value string) func() {
	// Store the original value
	originalValue, existed := os.LookupEnv(key)

	// Set the new environment variable
	err := os.Setenv(key, value)
	require.NoError(t, err, "Failed to set environment variable")

	// Return a cleanup function to restore the original state
	return func() {
		if existed {
			err := os.Setenv(key, originalValue)
			require.NoError(t, err, "Failed to restore environment variable")
		} else {
			err := os.Unsetenv(key)
			require.NoError(t, err, "Failed to unset environment variable")
		}
	}
}

func TestScanAWSEnvVars(t *testing.T) {
	testCases := []struct {
		name           string
		setupEnvVars   map[string]string
		expectedResult map[string]string
		expectError    bool
	}{
		{
			name: "Multiple AWS Environment Variables",
			setupEnvVars: map[string]string{
				"AWS_REGION":            "us-west-2",
				"AWS_ACCESS_KEY_ID":     "test-key",
				"AWS_SECRET_ACCESS_KEY": "test-secret",
				"SOME_OTHER_VAR":        "ignored",
			},
			expectedResult: map[string]string{
				"AWS_REGION":            "us-west-2",
				"AWS_ACCESS_KEY_ID":     "test-key",
				"AWS_SECRET_ACCESS_KEY": "test-secret",
			},
			expectError: false,
		},
		{
			name: "No AWS Environment Variables",
			setupEnvVars: map[string]string{
				"SOME_OTHER_VAR": "ignored",
				"AWS_SDK_LOAD_CONFIG": "true", // This should not count as an AWS credential var
			},
			expectedResult: nil,
			expectError:    true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Clean up environment variables after the test
			var cleanupFuncs []func()
			defer func() {
				for _, cleanup := range cleanupFuncs {
					cleanup()
				}
			}()

			// Set up test environment variables
			for key, value := range tc.setupEnvVars {
				cleanupFuncs = append(cleanupFuncs, setAndRestoreEnv(t, key, value))
			}

			// Run the test
			result, err := ScanAWSEnvVars()

			if tc.expectError {
				assert.Error(t, err, "Expected an error when no AWS vars are present")
				assert.Nil(t, result, "Result should be nil when no AWS vars are found")
			} else {
				assert.NoError(t, err, "Did not expect an error")
				assert.NotNil(t, result, "Result should not be nil")
				
				// Check that only AWS-prefixed variables are returned
				for key, value := range result {
					assert.Truef(t, 
						len(key) > 4 && key[:4] == "AWS_", 
						"Key %s should start with AWS_", key,
					)
					assert.NotEmptyf(t, value, "Value for key %s should not be empty", key)
				}
			}
		})
	}
}

func TestGetAWSSpecificEnvVars(t *testing.T) {
	testCases := []struct {
		name           string
		setupEnvVars   map[string]string
		getEnvVarFunc  func() (string, error)
		expectedValue  string
		expectError    bool
	}{
		{
			name: "Get AWS Region",
			setupEnvVars: map[string]string{
				"AWS_REGION": "us-west-2",
			},
			getEnvVarFunc: GetAWSRegionEnvVar,
			expectedValue: "us-west-2",
			expectError:   false,
		},
		{
			name: "Get AWS Default Region",
			setupEnvVars: map[string]string{
				"AWS_DEFAULT_REGION": "us-east-1",
			},
			getEnvVarFunc: GetAWSRegionDefaultEnvVar,
			expectedValue: "us-east-1",
			expectError:   false,
		},
		{
			name: "Get AWS Access Key ID",
			setupEnvVars: map[string]string{
				"AWS_ACCESS_KEY_ID": "test-access-key",
			},
			getEnvVarFunc: GetAWSAccessKeyIDEnvVar,
			expectedValue: "test-access-key",
			expectError:   false,
		},
		{
			name: "Get AWS Secret Access Key",
			setupEnvVars: map[string]string{
				"AWS_SECRET_ACCESS_KEY": "test-secret-key",
			},
			getEnvVarFunc: GetAWSSecretAccessKeyEnvVar,
			expectedValue: "test-secret-key",
			expectError:   false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Clean up environment variables after the test
			var cleanupFuncs []func()
			defer func() {
				for _, cleanup := range cleanupFuncs {
					cleanup()
				}
			}()

			// Set up test environment variables
			for key, value := range tc.setupEnvVars {
				cleanupFuncs = append(cleanupFuncs, setAndRestoreEnv(t, key, value))
			}

			// Run the test
			result, err := tc.getEnvVarFunc()

			if tc.expectError {
				assert.Error(t, err, "Expected an error")
				assert.Empty(t, result, "Result should be empty")
			} else {
				assert.NoError(t, err, "Did not expect an error")
				assert.Equal(t, tc.expectedValue, result, "Returned value should match")
			}
		})
	}
}

// Benchmark tests to ensure performance
func BenchmarkScanAWSEnvVars(b *testing.B) {
	// Set up some test environment variables
	os.Setenv("AWS_REGION", "us-west-2")
	os.Setenv("AWS_ACCESS_KEY_ID", "test-key")
	defer os.Unsetenv("AWS_REGION")
	defer os.Unsetenv("AWS_ACCESS_KEY_ID")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ScanAWSEnvVars()
	}
}

func BenchmarkGetAWSRegionEnvVar(b *testing.B) {
	// Set up test environment variable
	os.Setenv("AWS_REGION", "us-west-2")
	defer os.Unsetenv("AWS_REGION")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = GetAWSRegionEnvVar()
	}
}