package cloud

import (
	"context"
	"os"
	"testing"

	"github.com/Excoriate/aws-taggy/pkg/constants"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAWSClientConfig(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		inputRegion    string
		expectedRegion string
	}{
		{
			name:           "Empty region uses default",
			inputRegion:    "",
			expectedRegion: constants.DefaultAWSRegion,
		},
		{
			name:           "Specified region is used",
			inputRegion:    "us-west-2",
			expectedRegion: "us-west-2",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := NewAWSClientConfig(tc.inputRegion)
			require.NotNil(t, cfg)
			assert.Equal(t, tc.expectedRegion, cfg.GetRegion())
		})
	}
}

func TestAWSClientConfigOptions_Validate(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		region         string
		preSetEnvVar   string
		expectedResult bool
	}{
		{
			name:           "Valid region",
			region:         "us-east-1",
			expectedResult: true,
		},
		{
			name:           "Empty region with AWS_REGION set",
			region:         "",
			preSetEnvVar:   "eu-west-1",
			expectedResult: true,
		},
		{
			name:           "Empty region without AWS_REGION",
			region:         "",
			expectedResult: true, // Falls back to default region
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Backup and restore env var
			originalRegion := os.Getenv("AWS_REGION")
			if tc.preSetEnvVar != "" {
				os.Setenv("AWS_REGION", tc.preSetEnvVar)
			} else {
				os.Unsetenv("AWS_REGION")
			}
			defer os.Setenv("AWS_REGION", originalRegion)

			cfg := &AWSClientConfigOptions{Region: tc.region}
			err := cfg.Validate()

			if tc.expectedResult {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestAWSClientConfigOptions_LoadConfig(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		region      string
		expectError bool
	}{
		{
			name:        "Valid region configuration",
			region:      "us-east-1",
			expectError: false,
		},
		{
			name:        "Empty region falls back to default",
			region:      "",
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := NewAWSClientConfig(tc.region)
			awsCfg, err := cfg.LoadConfig(context.Background())

			if tc.expectError {
				assert.Error(t, err)
				assert.Nil(t, awsCfg)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, awsCfg)
				assert.NotEmpty(t, awsCfg.Region)
			}
		})
	}
}
