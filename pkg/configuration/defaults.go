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
	}
}
