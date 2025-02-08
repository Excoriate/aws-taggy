package configuration

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultConfiguration(t *testing.T) {
	config := DefaultConfiguration()

	// Test basic configuration values
	assert.Equal(t, "1.0", config.Version)
	assert.Equal(t, "all", config.AWS.Regions.Mode)
	assert.Equal(t, 20, *config.AWS.BatchSize)

	// Test global configuration
	assert.True(t, config.Global.Enabled)
	assert.Equal(t, 3, config.Global.TagCriteria.MinimumRequiredTags)
	assert.Equal(t, 50, config.Global.TagCriteria.MaxTags)
	assert.Equal(t, "high", config.Global.TagCriteria.ComplianceLevel)

	// Test required global tags
	expectedGlobalTags := []string{"Environment", "Owner", "Project"}
	assert.ElementsMatch(t, expectedGlobalTags, config.Global.TagCriteria.RequiredTags)

	// Test forbidden global tags
	expectedForbiddenTags := []string{"Temporary", "Test"}
	assert.ElementsMatch(t, expectedForbiddenTags, config.Global.TagCriteria.ForbiddenTags)

	// Test specific global tags
	expectedSpecificTags := map[string]string{
		"ComplianceLevel": "high",
		"ManagedBy":       "terraform",
	}
	assert.Equal(t, expectedSpecificTags, config.Global.TagCriteria.SpecificTags)

	// Test S3 resource configuration
	s3Config, exists := config.Resources["s3"]
	assert.True(t, exists)
	assert.True(t, s3Config.Enabled)
	assert.Equal(t, 4, s3Config.TagCriteria.MinimumRequiredTags)
	assert.Equal(t, "high", s3Config.TagCriteria.ComplianceLevel)

	// Test EC2 resource configuration
	ec2Config, exists := config.Resources["ec2"]
	assert.True(t, exists)
	assert.True(t, ec2Config.Enabled)
	assert.Equal(t, 3, ec2Config.TagCriteria.MinimumRequiredTags)
	assert.Equal(t, "standard", ec2Config.TagCriteria.ComplianceLevel)
}
