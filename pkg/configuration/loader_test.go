package configuration

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigLoader(t *testing.T) {
	tests := []struct {
		name    string
		content string
		wantErr bool
		errMsg  string
	}{
		{
			name: "Valid Full Configuration",
			content: `version: "1.0"
aws:
  regions:
    mode: "all"
  batch_size: 100
global:
  enabled: true
  tag_criteria:
    minimum_required_tags: 2
    required_tags:
      - "Environment"
      - "Owner"
resources:
  s3:
    enabled: true
    tag_criteria:
      minimum_required_tags: 2
      required_tags:
        - "DataClassification"
        - "BackupPolicy"
compliance_levels:
  high:
    required_tags:
      - "SecurityLevel"
      - "DataClassification"
tag_validation:
  key_validation:
    max_length: 128
    allowed_prefixes:
      - "env-"
      - "dept-"
    allowed_suffixes:
      - "-prod"
      - "-dev"
  allowed_values:
    Environment:
      - "production"
      - "staging"
  pattern_rules:
    CostCenter: "^[A-Z]{2}-[0-9]{4}$"
  case_rules:
    Environment:
      case: "lowercase"
      message: "Environment tag must be lowercase"
notifications:
  slack:
    enabled: true
    channels:
      high_priority: "compliance-alerts"
  email:
    enabled: true
    recipients:
      - "alerts@company.com"
    frequency: "daily"`,
			wantErr: false,
		},
		{
			name: "Invalid AWS Regions Mode",
			content: `version: "1.0"
aws:
  regions:
    mode: "invalid"
tag_validation:
  key_validation:
    max_length: 128`,
			wantErr: true,
			errMsg:  "invalid AWS regions mode: invalid",
		},
		{
			name: "Missing Required Fields",
			content: `version: "1.0"
aws:
  regions:
    mode: "all"`,
			wantErr: true,
			errMsg:  "key validation failed: key validation max length must be positive",
		},
		{
			name: "Invalid Tag Validation",
			content: `version: "1.0"
aws:
  regions:
    mode: "all"
tag_validation:
  key_validation:
    max_length: 128
  allowed_values:
    Environment: []`,
			wantErr: true,
			errMsg:  "no allowed values specified for tag Environment",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a temporary file
			tmpfile, err := os.CreateTemp("", "config-*.yaml")
			require.NoError(t, err)
			defer os.Remove(tmpfile.Name())

			// Write the test content to the file
			_, err = tmpfile.WriteString(tt.content)
			require.NoError(t, err)
			err = tmpfile.Close()
			require.NoError(t, err)

			// Create loader and load configuration
			loader := NewTaggyScanConfigLoader()
			cfg, err := loader.LoadConfig(tmpfile.Name())

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				assert.Nil(t, cfg)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, cfg)
				if cfg != nil {
					// Add specific assertions for the loaded configuration
					assert.Equal(t, "1.0", cfg.Version)
					assert.Equal(t, "all", cfg.AWS.Regions.Mode)
					assert.Equal(t, 2, cfg.Global.TagCriteria.MinimumRequiredTags)
					assert.Contains(t, cfg.Global.TagCriteria.RequiredTags, "Environment")
					assert.Contains(t, cfg.Global.TagCriteria.RequiredTags, "Owner")
				}
			}
		})
	}

	t.Run("File Validation Scenarios", func(t *testing.T) {
		testCases := []struct {
			name          string
			setupFile     func() (string, func())
			expectedError string
		}{
			{
				name: "Non-Existent File",
				setupFile: func() (string, func()) {
					return "non-existent-file.yaml", func() {}
				},
				expectedError: "configuration file does not exist",
			},
			{
				name: "Invalid File Extension",
				setupFile: func() (string, func()) {
					tmpfile, err := os.CreateTemp("", "config-*.txt")
					require.NoError(t, err)
					_, err = tmpfile.WriteString("version: 1.0")
					require.NoError(t, err)
					err = tmpfile.Close()
					require.NoError(t, err)
					return tmpfile.Name(), func() { os.Remove(tmpfile.Name()) }
				},
				expectedError: "configuration file has invalid extension",
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				filePath, cleanup := tc.setupFile()
				defer cleanup()

				loader := NewTaggyScanConfigLoader()
				_, err := loader.LoadConfig(filePath)

				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedError)
			})
		}
	})
}
