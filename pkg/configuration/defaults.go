package configuration

// DefaultConfiguration returns the default TaggyScanConfig with pre-configured values
// that follow best practices for AWS resource tagging.
func DefaultConfiguration() *TaggyScanConfig {
	batchSize := 20
	return &TaggyScanConfig{
		Version: "1.0",
		AWS: AWSConfig{
			Regions: RegionsConfig{
				Mode: "all",
			},
			BatchSize: &batchSize,
		},
		Global: GlobalConfig{
			Enabled: true,
			TagCriteria: TagCriteria{
				MinimumRequiredTags: 3,
				MaxTags:             50,
				RequiredTags: []string{
					"Environment",
					"Owner",
					"Project",
				},
				ForbiddenTags: []string{
					"Temporary",
					"Test",
				},
				SpecificTags: map[string]string{
					"ComplianceLevel": "high",
					"ManagedBy":       "terraform",
				},
				ComplianceLevel: "high",
			},
		},
		Resources: map[string]ResourceConfig{
			"s3": {
				Enabled: true,
				TagCriteria: TagCriteria{
					MinimumRequiredTags: 4,
					RequiredTags: []string{
						"DataClassification",
						"BackupPolicy",
						"Environment",
						"Owner",
					},
					ForbiddenTags: []string{
						"Temporary",
						"Test",
					},
					SpecificTags: map[string]string{
						"EncryptionRequired": "true",
					},
					ComplianceLevel: "high",
				},
				ExcludedResources: []ExcludedResource{
					{
						Pattern: "terraform-state-*",
						Reason:  "Terraform state buckets managed separately",
					},
					{
						Pattern: "log-archive-*",
						Reason:  "Logging buckets excluded from standard compliance",
					},
				},
			},
			"ec2": {
				Enabled: true,
				TagCriteria: TagCriteria{
					MinimumRequiredTags: 3,
					RequiredTags: []string{
						"Application",
						"PatchGroup",
						"Environment",
					},
					ForbiddenTags: []string{
						"Temporary",
						"Test",
					},
					SpecificTags: map[string]string{
						"AutoStop": "enabled",
					},
					ComplianceLevel: "standard",
				},
				ExcludedResources: []ExcludedResource{
					{
						Pattern: "bastion-*",
						Reason:  "Bastion hosts managed by security team",
					},
				},
			},
		},
		ComplianceLevels: map[string]ComplianceLevel{
			"high": {
				RequiredTags: []string{
					"DataClassification",
					"BackupPolicy",
					"Environment",
					"Owner",
				},
				SpecificTags: map[string]string{
					"EncryptionRequired": "true",
				},
			},
			"standard": {
				RequiredTags: []string{
					"Application",
					"PatchGroup",
					"Environment",
				},
				SpecificTags: map[string]string{
					"AutoStop": "enabled",
				},
			},
		},
		TagValidation: TagValidation{
			KeyValidation: KeyValidation{
				AllowedPrefixes: []string{"project-", "env-", "cost-"},
				AllowedSuffixes: []string{"-prod", "-dev", "-test"},
				MaxLength:       128,
			},
			ProhibitedTags: []string{"aws:", "internal:", "temp:", "test:"},
			KeyFormatRules: []KeyFormatRule{
				{
					Pattern: "^[a-z][a-z0-9_-]*$",
					Message: "Tag keys must start with lowercase letter and contain only letters, numbers, underscores, and hyphens",
				},
			},
			AllowedValues: map[string][]string{
				"Environment":        {"production", "staging", "development", "sandbox"},
				"DataClassification": {"public", "private", "confidential", "restricted"},
			},
			CaseRules: map[string]CaseRule{
				"Environment": {
					Case:    CaseLowercase,
					Message: "Environment tag must be lowercase",
				},
			},
			ValueValidation: ValueValidation{
				AllowedCharacters: "a-zA-Z0-9._-",
				DisallowedValues:  []string{"undefined", "null", "none", "n/a"},
			},
		},
	}
}
