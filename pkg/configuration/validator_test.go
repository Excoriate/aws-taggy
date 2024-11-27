package configuration

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper function to create a valid base configuration for testing
func createBaseValidConfig() *TaggyScanConfig {
	return &TaggyScanConfig{
		Version: "1.0",
		AWS: AWSConfig{
			Regions: RegionsConfig{
				Mode: "all",
			},
		},
		Global: GlobalConfig{
			Enabled: true,
			TagCriteria: TagCriteria{
				MinimumRequiredTags: 2,
				RequiredTags:        []string{"Environment", "Owner"},
			},
		},
		Resources: map[string]ResourceConfig{
			"s3": {
				Enabled: true,
				TagCriteria: TagCriteria{
					MinimumRequiredTags: 2,
					RequiredTags:        []string{"DataClassification", "BackupPolicy"},
				},
			},
		},
		ComplianceLevels: map[string]ComplianceLevel{
			"high": {
				RequiredTags: []string{"SecurityLevel", "DataClassification"},
			},
		},
		TagValidation: TagValidation{
			AllowedValues: map[string][]string{
				"Environment": {"production", "staging"},
			},
			PatternRules: map[string]string{
				"CostCenter": `^[A-Z]{2}-[0-9]{4}$`,
			},
		},
		Notifications: NotificationConfig{
			Slack: SlackNotificationConfig{
				Enabled: true,
				Channels: map[string]string{
					"high_priority": "compliance-alerts",
				},
			},
		},
	}
}

// deepCopyConfig creates a deep copy of the configuration to allow modifications
func deepCopyConfig(cfg *TaggyScanConfig) *TaggyScanConfig {
	// Create a new configuration with the same base values
	newCfg := &TaggyScanConfig{
		Version: cfg.Version,
		AWS:     cfg.AWS,
		Global:  cfg.Global,

		// Deep copy resources
		Resources: make(map[string]ResourceConfig),

		// Deep copy compliance levels
		ComplianceLevels: make(map[string]ComplianceLevel),

		TagValidation: cfg.TagValidation,
		Notifications: cfg.Notifications,
	}

	// Copy resources
	for resourceType, resource := range cfg.Resources {
		newResource := resource
		newCfg.Resources[resourceType] = newResource
	}

	// Copy compliance levels
	for levelName, level := range cfg.ComplianceLevels {
		newLevel := ComplianceLevel{
			RequiredTags: make([]string, len(level.RequiredTags)),
			SpecificTags: make(map[string]string),
		}
		copy(newLevel.RequiredTags, level.RequiredTags)

		for k, v := range level.SpecificTags {
			newLevel.SpecificTags[k] = v
		}

		newCfg.ComplianceLevels[levelName] = newLevel
	}

	return newCfg
}

func TestConfigValidator_ValidateVersion(t *testing.T) {
	testCases := []struct {
		name    string
		version string
		wantErr bool
	}{
		{"Valid Version", "1.0", false},
		{"Missing Version", "", true},
		{"Unsupported Version", "0.1.0", true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := createBaseValidConfig()
			cfg.Version = tc.version

			validator, err := NewConfigValidator(cfg)
			require.NoError(t, err)

			err = validator.ValidateVersion()
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestConfigValidator_ValidateGlobalConfig(t *testing.T) {
	testCases := []struct {
		name    string
		setup   func(*TaggyScanConfig)
		wantErr bool
	}{
		{
			name: "Valid Global Configuration",
			setup: func(cfg *TaggyScanConfig) {
				// Already valid in base config
			},
			wantErr: false,
		},
		{
			name: "Negative Batch Size",
			setup: func(cfg *TaggyScanConfig) {
				negBatchSize := -1
				cfg.Global.BatchSize = &negBatchSize
			},
			wantErr: true,
		},
		{
			name: "Minimum Required Tags Exceeds Required Tags",
			setup: func(cfg *TaggyScanConfig) {
				cfg.Global.TagCriteria.MinimumRequiredTags = 3
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := createBaseValidConfig()
			tc.setup(cfg)

			validator, err := NewConfigValidator(cfg)
			require.NoError(t, err)

			err = validator.ValidateGlobalConfig()
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestConfigValidator_ValidateResourceConfigs(t *testing.T) {
	testCases := []struct {
		name    string
		setup   func(*TaggyScanConfig)
		wantErr bool
	}{
		{
			name: "Valid Resource Configuration",
			setup: func(cfg *TaggyScanConfig) {
				// Already valid in base config
			},
			wantErr: false,
		},
		{
			name: "Invalid Compliance Level",
			setup: func(cfg *TaggyScanConfig) {
				// Modify a copy of the resource
				s3Resource := cfg.Resources["s3"]
				s3Resource.TagCriteria.ComplianceLevel = "invalid"
				cfg.Resources["s3"] = s3Resource
			},
			wantErr: true,
		},
		{
			name: "Minimum Required Tags Exceeds Required Tags",
			setup: func(cfg *TaggyScanConfig) {
				// Modify a copy of the resource
				s3Resource := cfg.Resources["s3"]
				s3Resource.TagCriteria.MinimumRequiredTags = 3
				cfg.Resources["s3"] = s3Resource
			},
			wantErr: true,
		},
		{
			name: "Empty Excluded Resource Pattern",
			setup: func(cfg *TaggyScanConfig) {
				// Modify a copy of the resource
				s3Resource := cfg.Resources["s3"]
				s3Resource.ExcludedResources = []ExcludedResource{
					{Pattern: "", Reason: "Test"},
				}
				cfg.Resources["s3"] = s3Resource
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := deepCopyConfig(createBaseValidConfig())
			tc.setup(cfg)

			validator, err := NewConfigValidator(cfg)
			require.NoError(t, err)

			err = validator.ValidateResourceConfigs()
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestConfigValidator_ValidateComplianceLevels(t *testing.T) {
	testCases := []struct {
		name    string
		setup   func(*TaggyScanConfig)
		wantErr bool
	}{
		{
			name: "Valid Compliance Levels",
			setup: func(cfg *TaggyScanConfig) {
				// Already valid in base config
			},
			wantErr: false,
		},
		{
			name: "Invalid Compliance Level Name",
			setup: func(cfg *TaggyScanConfig) {
				cfg.ComplianceLevels["invalid"] = ComplianceLevel{
					RequiredTags: []string{"Test"},
				}
			},
			wantErr: true,
		},
		{
			name: "Empty Required Tag",
			setup: func(cfg *TaggyScanConfig) {
				// Create a copy of the high compliance level
				highLevel := cfg.ComplianceLevels["high"]
				highLevel.RequiredTags = []string{""}
				cfg.ComplianceLevels["high"] = highLevel
			},
			wantErr: true,
		},
		{
			name: "Empty Specific Tag Key",
			setup: func(cfg *TaggyScanConfig) {
				// Create a copy of the high compliance level
				highLevel := cfg.ComplianceLevels["high"]
				highLevel.SpecificTags = map[string]string{
					"": "value",
				}
				cfg.ComplianceLevels["high"] = highLevel
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := deepCopyConfig(createBaseValidConfig())
			tc.setup(cfg)

			validator, err := NewConfigValidator(cfg)
			require.NoError(t, err)

			err = validator.ValidateComplianceLevels()
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestConfigValidator_ValidateTagValidationRules(t *testing.T) {
	testCases := []struct {
		name    string
		setup   func(*TaggyScanConfig)
		wantErr bool
	}{
		{
			name: "Valid Tag Validation Rules",
			setup: func(cfg *TaggyScanConfig) {
				// Already valid in base config
			},
			wantErr: false,
		},
		{
			name: "No Allowed Values",
			setup: func(cfg *TaggyScanConfig) {
				cfg.TagValidation.AllowedValues = map[string][]string{
					"Environment": {},
				}
			},
			wantErr: true,
		},
		{
			name: "Empty Allowed Value",
			setup: func(cfg *TaggyScanConfig) {
				cfg.TagValidation.AllowedValues = map[string][]string{
					"Environment": {"", "valid"},
				}
			},
			wantErr: true,
		},
		{
			name: "Empty Pattern Rule",
			setup: func(cfg *TaggyScanConfig) {
				cfg.TagValidation.PatternRules = map[string]string{
					"CostCenter": "",
				}
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := createBaseValidConfig()
			tc.setup(cfg)

			validator, err := NewConfigValidator(cfg)
			require.NoError(t, err)

			err = validator.ValidateTagValidationRules()
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestConfigValidator_ValidateNotifications(t *testing.T) {
	testCases := []struct {
		name    string
		setup   func(*TaggyScanConfig)
		wantErr bool
	}{
		{
			name: "Valid Slack Notifications",
			setup: func(cfg *TaggyScanConfig) {
				// Already valid in base config
			},
			wantErr: false,
		},
		{
			name: "Slack Enabled Without Channels",
			setup: func(cfg *TaggyScanConfig) {
				cfg.Notifications.Slack.Channels = map[string]string{}
			},
			wantErr: true,
		},
		{
			name: "Email Enabled Without Recipients",
			setup: func(cfg *TaggyScanConfig) {
				cfg.Notifications.Email.Enabled = true
				cfg.Notifications.Email.Recipients = []string{}
			},
			wantErr: true,
		},
		{
			name: "Invalid Email Frequency",
			setup: func(cfg *TaggyScanConfig) {
				cfg.Notifications.Email.Enabled = true
				cfg.Notifications.Email.Recipients = []string{"test@example.com"}
				cfg.Notifications.Email.Frequency = "monthly"
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := createBaseValidConfig()
			tc.setup(cfg)

			validator, err := NewConfigValidator(cfg)
			require.NoError(t, err)

			err = validator.ValidateNotifications()
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestConfigValidator_ValidateAWSConfig(t *testing.T) {
	testCases := []struct {
		name    string
		setup   func(*TaggyScanConfig)
		wantErr bool
	}{
		{
			name: "Valid AWS Configuration",
			setup: func(cfg *TaggyScanConfig) {
				// Already valid in base config
			},
			wantErr: false,
		},
		{
			name: "Invalid Regions Mode",
			setup: func(cfg *TaggyScanConfig) {
				cfg.AWS.Regions.Mode = "invalid"
			},
			wantErr: true,
		},
		{
			name: "Negative Batch Size",
			setup: func(cfg *TaggyScanConfig) {
				negBatchSize := -1
				cfg.AWS.BatchSize = &negBatchSize
			},
			wantErr: true,
		},
		{
			name: "Specific Regions Mode with Empty List",
			setup: func(cfg *TaggyScanConfig) {
				cfg.AWS.Regions.Mode = "specific"
				cfg.AWS.Regions.List = []string{}
			},
			wantErr: false, // Should default to us-east-1
		},
		{
			name: "Invalid Region in Specific Regions Mode",
			setup: func(cfg *TaggyScanConfig) {
				cfg.AWS.Regions.Mode = "specific"
				cfg.AWS.Regions.List = []string{"invalid-region"}
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := createBaseValidConfig()
			tc.setup(cfg)

			validator, err := NewConfigValidator(cfg)
			require.NoError(t, err)

			err = validator.ValidateAWSConfig()
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestConfigValidator_Validate(t *testing.T) {
	testCases := []struct {
		name    string
		setup   func(*TaggyScanConfig)
		wantErr bool
	}{
		{
			name: "Fully Valid Configuration",
			setup: func(cfg *TaggyScanConfig) {
				// Already valid in base config
			},
			wantErr: false,
		},
		{
			name: "Invalid Version",
			setup: func(cfg *TaggyScanConfig) {
				cfg.Version = "0.1.0"
			},
			wantErr: true,
		},
		{
			name: "Invalid AWS Configuration - Invalid Region Mode",
			setup: func(cfg *TaggyScanConfig) {
				cfg.AWS.Regions.Mode = "invalid"
			},
			wantErr: true,
		},
		{
			name: "Invalid Global Config - Negative Batch Size",
			setup: func(cfg *TaggyScanConfig) {
				negBatchSize := -1
				cfg.Global.BatchSize = &negBatchSize
			},
			wantErr: true,
		},
		{
			name: "Invalid Tag Criteria - MinimumRequiredTags Greater Than Required Tags",
			setup: func(cfg *TaggyScanConfig) {
				cfg.Global.TagCriteria.MinimumRequiredTags = 3
			},
			wantErr: true,
		},
		{
			name: "Invalid Resource Config - Empty Pattern in Excluded Resources",
			setup: func(cfg *TaggyScanConfig) {
				// Create a copy of the existing s3 resource
				s3Resource := cfg.Resources["s3"]
				s3Resource.ExcludedResources = []ExcludedResource{
					{Pattern: "", Reason: "Test"},
				}

				// Update the map with the modified resource
				cfg.Resources["s3"] = s3Resource
			},
			wantErr: true,
		},
		{
			name: "Invalid Compliance Level Reference",
			setup: func(cfg *TaggyScanConfig) {
				// Create a copy of the existing s3 resource
				s3Resource := cfg.Resources["s3"]
				s3Resource.TagCriteria.ComplianceLevel = "non-existent"

				// Update the map with the modified resource
				cfg.Resources["s3"] = s3Resource
			},
			wantErr: true,
		},
		{
			name: "Invalid Case Rule - Unknown Case Type",
			setup: func(cfg *TaggyScanConfig) {
				cfg.TagValidation.CaseRules = map[string]CaseRule{
					"Environment": {
						Case:    "unknown",
						Message: "test",
					},
				}
			},
			wantErr: true,
		},
		{
			name: "Invalid Pattern Rule - Empty Pattern",
			setup: func(cfg *TaggyScanConfig) {
				cfg.TagValidation.PatternRules = map[string]string{
					"CostCenter": "",
				}
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := createBaseValidConfig()
			tc.setup(cfg)

			validator, err := NewConfigValidator(cfg)
			require.NoError(t, err)

			err = validator.Validate()
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateTagCase(t *testing.T) {
	tests := []struct {
		name      string
		config    *TaggyScanConfig
		tagName   string
		tagValue  string
		wantError bool
	}{
		{
			name: "lowercase validation success",
			config: &TaggyScanConfig{
				TagValidation: TagValidation{
					CaseRules: map[string]CaseRule{
						"environment": {
							Case:    CaseLowercase,
							Message: "must be lowercase",
						},
					},
				},
			},
			tagName:   "environment",
			tagValue:  "production",
			wantError: false,
		},
		{
			name: "lowercase validation failure",
			config: &TaggyScanConfig{
				TagValidation: TagValidation{
					CaseRules: map[string]CaseRule{
						"environment": {
							Case:    CaseLowercase,
							Message: "must be lowercase",
						},
					},
				},
			},
			tagName:   "environment",
			tagValue:  "Production",
			wantError: true,
		},
		{
			name: "uppercase validation success",
			config: &TaggyScanConfig{
				TagValidation: TagValidation{
					CaseRules: map[string]CaseRule{
						"costcenter": {
							Case:    CaseUppercase,
							Message: "must be uppercase",
						},
					},
				},
			},
			tagName:   "costcenter",
			tagValue:  "ABC123",
			wantError: false,
		},
		{
			name: "mixed case validation success",
			config: &TaggyScanConfig{
				TagValidation: TagValidation{
					CaseRules: map[string]CaseRule{
						"projectcode": {
							Case:    CaseMixed,
							Pattern: "^[A-Z]+-[0-9]+$",
							Message: "must follow pattern",
						},
					},
				},
			},
			tagName:   "projectcode",
			tagValue:  "PRJ-123",
			wantError: false,
		},
		{
			name: "mixed case validation failure",
			config: &TaggyScanConfig{
				TagValidation: TagValidation{
					CaseRules: map[string]CaseRule{
						"projectcode": {
							Case:    CaseMixed,
							Pattern: "^[A-Z]+-[0-9]+$",
							Message: "must follow pattern",
						},
					},
				},
			},
			tagName:   "projectcode",
			tagValue:  "prj-123",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator, err := NewConfigValidator(tt.config)
			require.NoError(t, err)

			err = validator.ValidateTagCase(tt.tagName, tt.tagValue)
			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateTagValidation(t *testing.T) {
	t.Run("Valid Case Validation Configuration", func(t *testing.T) {
		cfg := &TaggyScanConfig{
			TagValidation: TagValidation{
				AllowedValues: map[string][]string{
					"Environment":   {"Production", "Staging", "Development"},
					"SecurityLevel": {"High", "Medium", "Low"},
				},
				CaseSensitivity: map[string]CaseSensitivityConfig{
					"Environment":   {Mode: CaseValidationStrict},
					"SecurityLevel": {Mode: CaseValidationRelaxed},
				},
				CaseRules: map[string]CaseRule{
					"Environment": {
						Case:    CaseLowercase,
						Message: "must be lowercase",
					},
					"SecurityLevel": {
						Case:    CaseUppercase,
						Message: "must be uppercase",
					},
				},
			},
		}

		validator, err := NewConfigValidator(cfg)
		require.NoError(t, err)

		err = validator.ValidateTagValidation()
		assert.NoError(t, err)
	})

	t.Run("Invalid Allowed Values", func(t *testing.T) {
		testCases := []struct {
			name        string
			allowedVals map[string][]string
			expectedErr string
		}{
			{
				name: "Empty Allowed Values",
				allowedVals: map[string][]string{
					"Environment": {},
				},
				expectedErr: "no allowed values specified for tag Environment",
			},
			{
				name: "Duplicate Allowed Values",
				allowedVals: map[string][]string{
					"Environment": {"Production", "Production"},
				},
				expectedErr: "duplicate value Production found for tag Environment",
			},
			{
				name: "Empty Allowed Value",
				allowedVals: map[string][]string{
					"Environment": {"", "Production"},
				},
				expectedErr: "empty value not allowed in allowed values for tag Environment",
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				cfg := &TaggyScanConfig{
					TagValidation: TagValidation{
						AllowedValues: tc.allowedVals,
					},
				}

				validator, err := NewConfigValidator(cfg)
				require.NoError(t, err)

				err = validator.ValidateTagValidation()
				assert.EqualError(t, err, tc.expectedErr)
			})
		}
	})

	t.Run("Invalid Case Sensitivity", func(t *testing.T) {
		testCases := []struct {
			name            string
			caseSensitivity map[string]CaseSensitivityConfig
			allowedValues   map[string][]string
			expectedErr     string
		}{
			{
				name: "Invalid Case Validation Mode",
				caseSensitivity: map[string]CaseSensitivityConfig{
					"Environment": {Mode: "invalid"},
				},
				expectedErr: "invalid case validation mode invalid for tag Environment",
			},
			{
				name: "Strict Mode Without Allowed Values",
				caseSensitivity: map[string]CaseSensitivityConfig{
					"Environment": {Mode: CaseValidationStrict},
				},
				expectedErr: "strict case validation requires allowed values for tag Environment",
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				cfg := &TaggyScanConfig{
					TagValidation: TagValidation{
						CaseSensitivity: tc.caseSensitivity,
						AllowedValues:   tc.allowedValues,
					},
				}

				validator, err := NewConfigValidator(cfg)
				require.NoError(t, err)

				err = validator.ValidateTagValidation()
				assert.EqualError(t, err, tc.expectedErr)
			})
		}
	})

	t.Run("Invalid Case Transformation Rules", func(t *testing.T) {
		testCases := []struct {
			name        string
			caseRules   map[string]CaseRule
			expectedErr string
		}{
			{
				name: "Invalid Case Transformation",
				caseRules: map[string]CaseRule{
					"environment": {
						Case:    "invalid",
						Message: "must be lowercase",
					},
				},
				expectedErr: "invalid case transformation invalid for tag environment",
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				cfg := &TaggyScanConfig{
					TagValidation: TagValidation{
						CaseRules: tc.caseRules,
					},
				}

				validator, err := NewConfigValidator(cfg)
				require.NoError(t, err)

				err = validator.ValidateTagValidation()
				assert.EqualError(t, err, tc.expectedErr)
			})
		}
	})
}

func TestValidateTagKey(t *testing.T) {
	tests := []struct {
		name    string
		config  *TaggyScanConfig
		key     string
		wantErr bool
	}{
		{
			name: "Valid key with allowed prefix",
			config: &TaggyScanConfig{
				TagValidation: TagValidation{
					KeyValidation: KeyValidation{
						AllowedPrefixes: []string{"project-", "env-"},
						MaxLength:       128,
					},
				},
			},
			key:     "project-test",
			wantErr: false,
		},
		{
			name: "Invalid key prefix",
			config: &TaggyScanConfig{
				TagValidation: TagValidation{
					KeyValidation: KeyValidation{
						AllowedPrefixes: []string{"project-", "env-"},
					},
				},
			},
			key:     "invalid-test",
			wantErr: true,
		},
		{
			name: "Valid key with allowed suffix",
			config: &TaggyScanConfig{
				TagValidation: TagValidation{
					KeyValidation: KeyValidation{
						AllowedSuffixes: []string{"-prod", "-dev"},
					},
				},
			},
			key:     "service-prod",
			wantErr: false,
		},
		{
			name: "Invalid key suffix",
			config: &TaggyScanConfig{
				TagValidation: TagValidation{
					KeyValidation: KeyValidation{
						AllowedSuffixes: []string{"-prod", "-dev"},
					},
				},
			},
			key:     "service-test",
			wantErr: true,
		},
		{
			name: "Key exceeds max length",
			config: &TaggyScanConfig{
				TagValidation: TagValidation{
					KeyValidation: KeyValidation{
						MaxLength: 10,
					},
				},
			},
			key:     "very-long-key-name",
			wantErr: true,
		},
		{
			name: "Key matches format rule",
			config: &TaggyScanConfig{
				TagValidation: TagValidation{
					KeyFormatRules: []KeyFormatRule{
						{
							Pattern: "^[a-z][a-z0-9-]*$",
							Message: "Must start with lowercase letter",
						},
					},
				},
			},
			key:     "app-123",
			wantErr: false,
		},
		{
			name: "Key violates format rule",
			config: &TaggyScanConfig{
				TagValidation: TagValidation{
					KeyFormatRules: []KeyFormatRule{
						{
							Pattern: "^[a-z][a-z0-9-]*$",
							Message: "Must start with lowercase letter",
						},
					},
				},
			},
			key:     "123-app",
			wantErr: true,
		},
		{
			name: "Key is prohibited",
			config: &TaggyScanConfig{
				TagValidation: TagValidation{
					ProhibitedTags: []string{"aws:", "temp:"},
				},
			},
			key:     "aws:name",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator, err := NewConfigValidator(tt.config)
			require.NoError(t, err)

			err = validator.ValidateTagKey(tt.key)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateTagValue(t *testing.T) {
	tests := []struct {
		name    string
		config  *TaggyScanConfig
		key     string
		value   string
		wantErr bool
	}{
		{
			name: "Value within length constraints",
			config: &TaggyScanConfig{
				TagValidation: TagValidation{
					LengthRules: map[string]LengthRule{
						"environment": {
							MinLength: intPtr(2),
							MaxLength: intPtr(15),
							Message:   "Length must be between 2 and 15",
						},
					},
				},
			},
			key:     "environment",
			value:   "production",
			wantErr: false,
		},
		{
			name: "Value too short",
			config: &TaggyScanConfig{
				TagValidation: TagValidation{
					LengthRules: map[string]LengthRule{
						"environment": {
							MinLength: intPtr(4),
							Message:   "Must be at least 4 characters",
						},
					},
				},
			},
			key:     "environment",
			value:   "dev",
			wantErr: true,
		},
		{
			name: "Value matches allowed characters",
			config: &TaggyScanConfig{
				TagValidation: TagValidation{
					ValueValidation: ValueValidation{
						AllowedCharacters: "a-zA-Z0-9-_",
					},
				},
			},
			key:     "name",
			value:   "app-123",
			wantErr: false,
		},
		{
			name: "Value contains disallowed characters",
			config: &TaggyScanConfig{
				TagValidation: TagValidation{
					ValueValidation: ValueValidation{
						AllowedCharacters: "a-zA-Z0-9-_",
					},
				},
			},
			key:     "name",
			value:   "app@123",
			wantErr: true,
		},
		{
			name: "Value is disallowed",
			config: &TaggyScanConfig{
				TagValidation: TagValidation{
					ValueValidation: ValueValidation{
						DisallowedValues: []string{"none", "null", "undefined"},
					},
				},
			},
			key:     "owner",
			value:   "none",
			wantErr: true,
		},
		{
			name: "Value matches pattern rule",
			config: &TaggyScanConfig{
				TagValidation: TagValidation{
					PatternRules: map[string]string{
						"email": "^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$",
					},
				},
			},
			key:     "email",
			value:   "user@company.com",
			wantErr: false,
		},
		{
			name: "Value violates pattern rule",
			config: &TaggyScanConfig{
				TagValidation: TagValidation{
					PatternRules: map[string]string{
						"email": "^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$",
					},
				},
			},
			key:     "email",
			value:   "invalid-email",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator, err := NewConfigValidator(tt.config)
			require.NoError(t, err)

			err = validator.ValidateTagValue(tt.key, tt.value)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateTagCount(t *testing.T) {
	tests := []struct {
		name    string
		config  *TaggyScanConfig
		tags    map[string]string
		wantErr bool
	}{
		{
			name: "Tags within limit",
			config: &TaggyScanConfig{
				Global: GlobalConfig{
					TagCriteria: TagCriteria{
						MaxTags: 5,
					},
				},
			},
			tags: map[string]string{
				"name":        "test",
				"environment": "prod",
				"owner":       "team",
			},
			wantErr: false,
		},
		{
			name: "Tags exceed limit",
			config: &TaggyScanConfig{
				Global: GlobalConfig{
					TagCriteria: TagCriteria{
						MaxTags: 2,
					},
				},
			},
			tags: map[string]string{
				"name":        "test",
				"environment": "prod",
				"owner":       "team",
			},
			wantErr: true,
		},
		{
			name: "No max tags limit",
			config: &TaggyScanConfig{
				Global: GlobalConfig{
					TagCriteria: TagCriteria{
						MaxTags: 0,
					},
				},
			},
			tags: map[string]string{
				"name":        "test",
				"environment": "prod",
				"owner":       "team",
				"project":     "demo",
				"cost":        "123",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator, err := NewConfigValidator(tt.config)
			require.NoError(t, err)

			err = validator.ValidateTagCount(tt.tags)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestConfigValidator_ValidateTagValidation(t *testing.T) {
	testCases := []struct {
		name    string
		setup   func(*TaggyScanConfig)
		wantErr bool
	}{
		{
			name: "Valid Tag Validation Configuration",
			setup: func(cfg *TaggyScanConfig) {
				// Already valid in base config
			},
			wantErr: false,
		},
		{
			name: "Empty Allowed Values",
			setup: func(cfg *TaggyScanConfig) {
				cfg.TagValidation.AllowedValues = map[string][]string{
					"Environment": {},
				}
			},
			wantErr: true,
		},
		{
			name: "Duplicate Allowed Values",
			setup: func(cfg *TaggyScanConfig) {
				cfg.TagValidation.AllowedValues = map[string][]string{
					"Environment": {"production", "production"},
				}
			},
			wantErr: true,
		},
		{
			name: "Invalid Case Validation Mode",
			setup: func(cfg *TaggyScanConfig) {
				cfg.TagValidation.CaseSensitivity = map[string]CaseSensitivityConfig{
					"Environment": {
						Mode: "invalid_mode",
					},
				}
			},
			wantErr: true,
		},
		{
			name: "Strict Case Validation Without Allowed Values",
			setup: func(cfg *TaggyScanConfig) {
				cfg.TagValidation.CaseSensitivity = map[string]CaseSensitivityConfig{
					"Environment": {
						Mode: CaseValidationStrict,
					},
				}
				cfg.TagValidation.AllowedValues = map[string][]string{}
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := createBaseValidConfig()
			tc.setup(cfg)

			validator, err := NewConfigValidator(cfg)
			require.NoError(t, err)

			err = validator.ValidateTagValidation()
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestConfigValidator_ValidateTagKeys(t *testing.T) {
	testCases := []struct {
		name    string
		setup   func(*TaggyScanConfig)
		wantErr bool
	}{
		{
			name: "Valid Key Validation",
			setup: func(cfg *TaggyScanConfig) {
				cfg.TagValidation.KeyValidation = KeyValidation{
					AllowedPrefixes: []string{"project-", "env-"},
					MaxLength:       128,
				}
				cfg.TagValidation.KeyFormatRules = []KeyFormatRule{
					{
						Pattern: "^[a-z][a-z0-9_-]*$",
						Message: "Invalid key format",
					},
				}
			},
			wantErr: false,
		},
		{
			name: "Key Exceeds Max Length",
			setup: func(cfg *TaggyScanConfig) {
				cfg.TagValidation.KeyValidation = KeyValidation{
					MaxLength: 10,
				}
				cfg.TagValidation.KeyFormatRules = []KeyFormatRule{
					{
						Pattern: "very_long_key_pattern_that_exceeds_max_length",
						Message: "Key too long",
					},
				}
			},
			wantErr: true,
		},
		{
			name: "No Allowed Prefix",
			setup: func(cfg *TaggyScanConfig) {
				cfg.TagValidation.KeyValidation = KeyValidation{
					AllowedPrefixes: []string{"specific-"},
				}
				cfg.TagValidation.KeyFormatRules = []KeyFormatRule{
					{
						Pattern: "unrelated_key_pattern",
						Message: "No matching prefix",
					},
				}
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := createBaseValidConfig()
			tc.setup(cfg)

			validator, err := NewConfigValidator(cfg)
			require.NoError(t, err)

			err = validator.validateTagKeys()
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// Helper function to create pointer to int
func intPtr(i int) *int {
	return &i
}
