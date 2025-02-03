package configuration

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// Helper function to load test configuration
func loadTestConfiguration(t *testing.T) *TaggyScanConfig {
	// Use the example configuration from tag-compliance.yaml
	configContent := `
version: "1.0"
aws:
  regions:
    mode: all
resources:
  s3:
    enabled: true
    tag_criteria:
      minimum_required_tags: 4
  ec2:
    enabled: true
compliance_levels:
  high:
    required_tags:
      - SecurityLevel
      - DataClassification
tag_validation:
  allowed_values:
    Environment:
      - production
      - staging
notifications:
  slack:
    enabled: true
    channels:
      high_priority: "compliance-alerts"
`

	cfg := &TaggyScanConfig{}
	err := yaml.Unmarshal([]byte(configContent), cfg)
	require.NoError(t, err, "Failed to unmarshal test configuration")

	return cfg
}

func TestConfigQuerier(t *testing.T) {
	cfg := loadTestConfiguration(t)
	querier, err := NewConfigQuerier(cfg)
	require.NoError(t, err, "Failed to create config querier")

	t.Run("GetResources", func(t *testing.T) {
		resources, err := querier.GetResources()
		assert.NoError(t, err)
		assert.Len(t, resources, 2)
		assert.Contains(t, resources, "s3")
		assert.Contains(t, resources, "ec2")
	})

	t.Run("GetAWSConfig", func(t *testing.T) {
		awsConfig, err := querier.GetAWSConfig()
		assert.NoError(t, err)
		assert.Equal(t, "all", awsConfig.Regions.Mode)
	})

	t.Run("GetComplianceLevels", func(t *testing.T) {
		levels, err := querier.GetComplianceLevels()
		assert.NoError(t, err)
		assert.Len(t, levels, 1)
		assert.Contains(t, levels, "high")
	})

	t.Run("GetTagValidationConfig", func(t *testing.T) {
		tagValidation, err := querier.GetTagValidationConfig()
		assert.NoError(t, err)
		assert.Len(t, tagValidation.AllowedValues["Environment"], 2)
	})

	t.Run("GetNotificationsConfig", func(t *testing.T) {
		notifications, err := querier.GetNotificationsConfig()
		assert.NoError(t, err)
		assert.True(t, notifications.Slack.Enabled)
	})

	t.Run("GetResourceByType", func(t *testing.T) {
		s3Resource, err := querier.GetResourceByType("s3")
		assert.NoError(t, err)
		assert.True(t, s3Resource.Enabled)
		assert.Equal(t, 4, s3Resource.TagCriteria.MinimumRequiredTags)
	})

	t.Run("GetComplianceLevelByName", func(t *testing.T) {
		highLevel, err := querier.GetComplianceLevelByName("high")
		assert.NoError(t, err)
		assert.Contains(t, highLevel.RequiredTags, "SecurityLevel")
		assert.Contains(t, highLevel.RequiredTags, "DataClassification")
	})

	t.Run("Error Scenarios", func(t *testing.T) {
		t.Run("Empty Configuration", func(t *testing.T) {
			emptyQuerier, err := NewConfigQuerier(&TaggyScanConfig{})
			require.NoError(t, err)

			_, err = emptyQuerier.GetResources()
			assert.Error(t, err)

			_, err = emptyQuerier.GetAWSConfig()
			assert.Error(t, err)

			_, err = emptyQuerier.GetComplianceLevels()
			assert.Error(t, err)

			_, err = emptyQuerier.GetTagValidationConfig()
			assert.Error(t, err)

			_, err = emptyQuerier.GetNotificationsConfig()
			assert.Error(t, err)
		})

		t.Run("Non-Existent Resource", func(t *testing.T) {
			_, err := querier.GetResourceByType("rds")
			assert.Error(t, err)
		})

		t.Run("Non-Existent Compliance Level", func(t *testing.T) {
			_, err := querier.GetComplianceLevelByName("low")
			assert.Error(t, err)
		})
	})
}
