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
			RequiredTags:  make([]string, len(level.RequiredTags)),
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
				cfg.Version = ""
			},
			wantErr: true,
		},
		{
			name: "Invalid AWS Configuration",
			setup: func(cfg *TaggyScanConfig) {
				cfg.AWS.Regions.Mode = "invalid"
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