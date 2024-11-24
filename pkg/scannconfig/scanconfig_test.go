package scannconfig

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsValidComplianceLevel(t *testing.T) {
	testCases := []struct {
		name     string
		level    string
		expected bool
	}{
		{"Valid High Level", "high", true},
		{"Valid Medium Level", "medium", true},
		{"Valid Low Level", "low", true},
		{"Valid Standard Level", "standard", true},
		{"Invalid Level", "invalid", false},
		{"Empty Level", "", false},
		{"Case Sensitive", "High", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := IsValidComplianceLevel(tc.level)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestValidAWSRegions(t *testing.T) {
	regions := ValidAWSRegions()
	t.Run("Region List Validation", func(t *testing.T) {
		assert.NotEmpty(t, regions, "AWS regions list should not be empty")
		assert.Contains(t, regions, "us-east-1", "us-east-1 should be in the list")
		assert.Contains(t, regions, "eu-west-1", "eu-west-1 should be in the list")
		assert.Contains(t, regions, "ap-southeast-1", "ap-southeast-1 should be in the list")
	})

	t.Run("Minimum Region Coverage", func(t *testing.T) {
		minimumRegions := []string{
			"us-east-1", "us-west-2", 
			"eu-west-1", "ap-southeast-1", 
			"sa-east-1", "af-south-1",
		}

		for _, region := range minimumRegions {
			assert.Contains(t, regions, region, 
				"Region %s should be in the list of valid AWS regions", region)
		}
	})
}

func TestIsValidRegion(t *testing.T) {
	testCases := []struct {
		name     string
		region   string
		expected bool
	}{
		{"Valid US East Region", "us-east-1", true},
		{"Valid EU West Region", "eu-west-1", true},
		{"Valid Asia Pacific Region", "ap-southeast-1", true},
		{"Invalid Region", "invalid-region", false},
		{"Empty Region", "", false},
		{"Case Sensitive Region", "US-EAST-1", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := IsValidRegion(tc.region)
			assert.Equal(t, tc.expected, result)
		})
	}
}
func TestNormalizeAWSConfig(t *testing.T) {
	t.Run("Empty Configuration", func(t *testing.T) {
		cfg := &AWSConfig{}
		NormalizeAWSConfig(cfg)

		assert.Equal(t, "specific", cfg.Regions.Mode)
		assert.Len(t, cfg.Regions.List, 1)
		assert.Equal(t, DefaultAWSRegion, cfg.Regions.List[0])
	})
	t.Run("Partial Configuration", func(t *testing.T) {
		cfg := &AWSConfig{
			Regions: RegionsConfig{
				Mode: "all",
			},
		}
		NormalizeAWSConfig(cfg)

		assert.Equal(t, "all", cfg.Regions.Mode)
	})

	t.Run("Default Batch Size", func(t *testing.T) {
		cfg := &AWSConfig{}
		NormalizeAWSConfig(cfg)

		assert.NotNil(t, cfg.BatchSize)
		assert.Equal(t, 20, *cfg.BatchSize)
	})

	t.Run("Existing Batch Size Preserved", func(t *testing.T) {
		batchSize := 50
		cfg := &AWSConfig{
			BatchSize: &batchSize,
		}
		NormalizeAWSConfig(cfg)

		assert.NotNil(t, cfg.BatchSize)
		assert.Equal(t, 50, *cfg.BatchSize)
	})
}

func TestAWSConfigScenarios(t *testing.T) {
	t.Run("All Regions Mode", func(t *testing.T) {
		cfg := &AWSConfig{
			Regions: RegionsConfig{
				Mode: "all",
			},
		}
		NormalizeAWSConfig(cfg)

		assert.Equal(t, "all", cfg.Regions.Mode)
	})

	t.Run("Specific Regions Mode", func(t *testing.T) {
		cfg := &AWSConfig{
			Regions: RegionsConfig{
				Mode: "specific",
				List: []string{"us-west-2", "eu-west-1"},
			},
		}
		NormalizeAWSConfig(cfg)

		assert.Equal(t, "specific", cfg.Regions.Mode)
		assert.Len(t, cfg.Regions.List, 2)
		assert.Contains(t, cfg.Regions.List, "us-west-2")
		assert.Contains(t, cfg.Regions.List, "eu-west-1")
	})
}