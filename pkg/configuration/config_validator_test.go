package configuration

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// createTestConfig creates a valid base configuration for testing
func createTestConfig() *TaggyScanConfig {
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
			CaseRules: map[string]CaseRule{
				"Environment": {
					Case:    CaseLowercase,
					Message: "Environment tag must be lowercase",
				},
			},
		},
		Notifications: NotificationConfig{
			Slack: SlackNotificationConfig{
				Enabled: true,
				Channels: map[string]string{
					"high_priority": "compliance-alerts",
				},
			},
			Email: EmailNotificationConfig{
				Enabled:    true,
				Recipients: []string{"alerts@company.com"},
				Frequency:  "daily",
			},
		},
	}
}

func TestFileValidator_Validate(t *testing.T) {
	tests := []struct {
		name    string
		cfgPath string
		wantErr bool
	}{
		{
			name:    "Empty Path",
			cfgPath: "",
			wantErr: true,
		},
		// Add more test cases for file validation
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator, err := NewFileValidator(tt.cfgPath)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)

			err = validator.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestContentValidator_ValidateContent(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(*TaggyScanConfig)
		wantErr bool
	}{
		{
			name:    "Valid Configuration",
			setup:   func(cfg *TaggyScanConfig) {},
			wantErr: false,
		},
		{
			name: "Invalid Version",
			setup: func(cfg *TaggyScanConfig) {
				cfg.Version = "invalid"
			},
			wantErr: true,
		},
		{
			name: "Invalid AWS Region Mode",
			setup: func(cfg *TaggyScanConfig) {
				cfg.AWS.Regions.Mode = "invalid"
			},
			wantErr: true,
		},
		{
			name: "Invalid Global Batch Size",
			setup: func(cfg *TaggyScanConfig) {
				negativeSize := -1
				cfg.Global.BatchSize = &negativeSize
			},
			wantErr: true,
		},
		{
			name: "Invalid Tag Criteria",
			setup: func(cfg *TaggyScanConfig) {
				cfg.Global.TagCriteria.MinimumRequiredTags = -1
			},
			wantErr: true,
		},
		{
			name: "Invalid Compliance Level",
			setup: func(cfg *TaggyScanConfig) {
				cfg.ComplianceLevels["invalid"] = ComplianceLevel{}
			},
			wantErr: true,
		},
		{
			name: "Invalid Case Rule",
			setup: func(cfg *TaggyScanConfig) {
				cfg.TagValidation.CaseRules["Test"] = CaseRule{
					Case: "",
				}
			},
			wantErr: true,
		},
		{
			name: "Invalid Pattern Rule",
			setup: func(cfg *TaggyScanConfig) {
				cfg.TagValidation.PatternRules["Test"] = "["
			},
			wantErr: true,
		},
		{
			name: "Invalid Email",
			setup: func(cfg *TaggyScanConfig) {
				cfg.Notifications.Email.Recipients = []string{"invalid"}
			},
			wantErr: true,
		},
		{
			name: "Invalid Email Frequency",
			setup: func(cfg *TaggyScanConfig) {
				cfg.Notifications.Email.Frequency = "invalid"
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := createTestConfig()
			tt.setup(cfg)

			validator, err := NewContentValidator(cfg)
			require.NoError(t, err)

			err = validator.ValidateContent()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestContentValidator_ValidateAWSConfig(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(*TaggyScanConfig)
		wantErr bool
	}{
		{
			name:    "Valid AWS Config",
			setup:   func(cfg *TaggyScanConfig) {},
			wantErr: false,
		},
		{
			name: "Empty Region Mode",
			setup: func(cfg *TaggyScanConfig) {
				cfg.AWS.Regions.Mode = ""
			},
			wantErr: true,
		},
		{
			name: "Invalid Region Mode",
			setup: func(cfg *TaggyScanConfig) {
				cfg.AWS.Regions.Mode = "invalid"
			},
			wantErr: true,
		},
		{
			name: "Specific Mode Without Regions",
			setup: func(cfg *TaggyScanConfig) {
				cfg.AWS.Regions.Mode = "specific"
				cfg.AWS.Regions.List = nil
			},
			wantErr: true,
		},
		{
			name: "Invalid Batch Size",
			setup: func(cfg *TaggyScanConfig) {
				size := 0
				cfg.AWS.BatchSize = &size
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := createTestConfig()
			tt.setup(cfg)

			validator, err := NewContentValidator(cfg)
			require.NoError(t, err)

			err = validator.validateAWSConfig()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestContentValidator_ValidateTagValidation(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(*TaggyScanConfig)
		wantErr bool
	}{
		{
			name:    "Valid Tag Validation",
			setup:   func(cfg *TaggyScanConfig) {},
			wantErr: false,
		},
		{
			name: "Empty Case Type",
			setup: func(cfg *TaggyScanConfig) {
				cfg.TagValidation.CaseRules["Test"] = CaseRule{
					Case: "",
				}
			},
			wantErr: true,
		},
		{
			name: "Invalid Pattern",
			setup: func(cfg *TaggyScanConfig) {
				cfg.TagValidation.CaseRules["Test"] = CaseRule{
					Case:    CaseMixed,
					Pattern: "[",
				}
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := createTestConfig()
			tt.setup(cfg)

			validator, err := NewContentValidator(cfg)
			require.NoError(t, err)

			err = validator.validateTagValidation()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestContentValidator_ValidateNotifications(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(*TaggyScanConfig)
		wantErr bool
	}{
		{
			name:    "Valid Notifications",
			setup:   func(cfg *TaggyScanConfig) {},
			wantErr: false,
		},
		{
			name: "Slack Enabled Without Channels",
			setup: func(cfg *TaggyScanConfig) {
				cfg.Notifications.Slack.Enabled = true
				cfg.Notifications.Slack.Channels = nil
			},
			wantErr: true,
		},
		{
			name: "Email Enabled Without Recipients",
			setup: func(cfg *TaggyScanConfig) {
				cfg.Notifications.Email.Enabled = true
				cfg.Notifications.Email.Recipients = nil
			},
			wantErr: true,
		},
		{
			name: "Invalid Email Address",
			setup: func(cfg *TaggyScanConfig) {
				cfg.Notifications.Email.Enabled = true
				cfg.Notifications.Email.Recipients = []string{"invalid"}
			},
			wantErr: true,
		},
		{
			name: "Invalid Email Frequency",
			setup: func(cfg *TaggyScanConfig) {
				cfg.Notifications.Email.Enabled = true
				cfg.Notifications.Email.Frequency = "invalid"
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := createTestConfig()
			tt.setup(cfg)

			validator, err := NewContentValidator(cfg)
			require.NoError(t, err)

			err = validator.validateNotifications()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
