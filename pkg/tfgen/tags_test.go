package tfgen

import (
	"regexp"
	"testing"

	"github.com/Excoriate/aws-taggy/pkg/configuration"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// createTestConfig generates a sample configuration for testing
func createTestConfig() *configuration.TaggyScanConfig {
	return &configuration.TaggyScanConfig{
		Version: "1.0",
		Global: configuration.GlobalConfig{
			Enabled: true,
			TagCriteria: configuration.TagCriteria{
				MinimumRequiredTags: 2,
				ComplianceLevel:     "standard",
				RequiredTags:        []string{"Environment", "Project"},
			},
		},
		ComplianceLevels: map[string]configuration.ComplianceLevel{
			"standard": {
				RequiredTags: []string{"Environment", "Project"},
				SpecificTags: map[string]string{
					"ManagedBy": "aws-taggy",
				},
			},
		},
		Resources: map[string]configuration.ResourceConfig{
			"ec2": {
				Enabled: true,
				TagCriteria: configuration.TagCriteria{
					ComplianceLevel: "standard",
					RequiredTags:    []string{"Name"},
				},
			},
		},
		TagValidation: configuration.TagValidation{
			AllowedValues: map[string][]string{
				"Environment": {"dev", "staging", "prod"},
			},
			LengthRules: map[string]configuration.LengthRule{
				"Project": {
					MinLength: intPtr(3),
					MaxLength: intPtr(10),
				},
			},
			CaseRules: map[string]configuration.CaseRule{
				"Environment": {
					Case: configuration.CaseLowercase,
				},
			},
			PatternRules: map[string]string{
				"CostCenter": "^[A-Z]{2}-[0-9]{4}$",
			},
		},
	}
}

// intPtr is a helper function to create an integer pointer
func intPtr(i int) *int {
	return &i
}

// TestNewTagGenerator tests the creation of a new TagGenerator
func TestNewTagGenerator(t *testing.T) {
	config := createTestConfig()
	generator, err := NewTagGenerator(config)

	assert.NoError(t, err)
	assert.NotNil(t, generator)
}

// TestNewTagGenerator_NilConfig tests creating a TagGenerator with nil config
func TestNewTagGenerator_NilConfig(t *testing.T) {
	generator, err := NewTagGenerator(nil)

	assert.Error(t, err)
	assert.Nil(t, generator)
}

// TestGenerateTagValue tests the tag value generation logic
func TestGenerateTagValue(t *testing.T) {
	config := createTestConfig()
	generator, _ := NewTagGenerator(config)

	testCases := []struct {
		name           string
		tagName        string
		expectedPrefix string
		validateFunc   func(string) bool
	}{
		{
			name:           "Environment Tag",
			tagName:        "Environment",
			expectedPrefix: "",
			validateFunc: func(value string) bool {
				return value == "dev" || value == "staging" || value == "prod"
			},
		},
		{
			name:           "Project Tag",
			tagName:        "Project",
			expectedPrefix: "default-",
			validateFunc: func(value string) bool {
				return len(value) >= 3 && len(value) <= 10
			},
		},
		{
			name:           "CostCenter Tag",
			tagName:        "CostCenter",
			expectedPrefix: "",
			validateFunc: func(value string) bool {
				match, _ := regexp.MatchString(`^[A-Z]{2}-[0-9]{4}$`, value)
				return match
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			value := generator.generateTagValue(tc.name)

			if tc.expectedPrefix != "" {
				assert.Contains(t, value, tc.expectedPrefix)
			}

			assert.True(t, tc.validateFunc(value),
				"Generated value %s does not meet validation criteria", value)
		})
	}
}

// TestGenerateTags tests the generation of tags for a specific resource type
func TestGenerateTags(t *testing.T) {
	config := createTestConfig()
	generator, _ := NewTagGenerator(config)

	file, err := generator.GenerateTags("ec2")
	require.NoError(t, err)
	require.NotNil(t, file)

	// Convert file to string for inspection
	fileContent := string(file.Bytes())

	// Verify key components
	assert.Contains(t, fileContent, "resource \"ec2\" \"example\"")
	assert.Contains(t, fileContent, "tags")
	assert.Contains(t, fileContent, "Environment")
	assert.Contains(t, fileContent, "Project")
	assert.Contains(t, fileContent, "Name")
}

// TestApplyTagConstraints tests the application of tag constraints
func TestApplyTagConstraints(t *testing.T) {
	config := createTestConfig()
	generator, _ := NewTagGenerator(config)

	testCases := []struct {
		name          string
		tagName       string
		inputValue    string
		expectedValue string
	}{
		{
			name:          "Lowercase Environment",
			tagName:       "Environment",
			inputValue:    "PROD",
			expectedValue: "prod",
		},
		{
			name:          "Project Length Constraint",
			tagName:       "Project",
			inputValue:    "very-long-project-name",
			expectedValue: "very-long-p",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			value := generator.applyTagConstraints(tc.tagName, tc.inputValue)
			assert.Equal(t, tc.expectedValue, value)
		})
	}
}

// TestGenerateFileHeader tests the file header generation
func TestGenerateFileHeader(t *testing.T) {
	config := createTestConfig()
	generator, _ := NewTagGenerator(config)

	header := generator.generateFileHeader("ec2")

	assert.Contains(t, header, "AWS Taggy - Automated Tag Compliance Generator")
	assert.Contains(t, header, "Resource Type:    ec2")
	assert.Contains(t, header, "Do not manually edit this file")
}
