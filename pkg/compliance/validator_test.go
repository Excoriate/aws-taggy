package compliance

import (
	"testing"

	"github.com/Excoriate/aws-taggy/pkg/configuration"
	"github.com/stretchr/testify/assert"
)

func createTestConfig() *configuration.TaggyScanConfig {
	return &configuration.TaggyScanConfig{
		Global: configuration.GlobalConfig{
			Enabled: true,
			TagCriteria: configuration.TagCriteria{
				MinimumRequiredTags: 2,
				RequiredTags:        []string{"Environment", "Owner"},
			},
		},
		TagValidation: configuration.TagValidation{
			AllowedValues: map[string][]string{
				"Environment": {"production", "staging", "development"},
			},
			CaseRules: map[string]configuration.CaseRule{
				"Environment": {
					Case:    "lowercase",
					Message: "Environment must be lowercase",
				},
				"Owner": {
					Case:    "lowercase",
					Message: "Owner must be lowercase",
				},
			},
			PatternRules: map[string]string{
				"Owner": `^[a-z0-9._%+-]+@company\.com$`,
			},
		},
	}
}

func TestValidateTags_RequiredTags(t *testing.T) {
	testCases := []struct {
		name               string
		tags               map[string]string
		expectedResult     bool
		expectedViolations int
	}{
		{
			name: "Valid tags with all required tags",
			tags: map[string]string{
				"Environment": "production",
				"Owner":       "team@company.com",
			},
			expectedResult:     true,
			expectedViolations: 0,
		},
		{
			name: "Missing required tags",
			tags: map[string]string{
				"Environment": "production",
			},
			expectedResult:     false,
			expectedViolations: 1,
		},
	}

	config := createTestConfig()
	validator := NewTagValidator(config)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := validator.ValidateTags(tc.tags)
			assert.Equal(t, tc.expectedResult, result.IsCompliant)
			assert.Len(t, result.Violations, tc.expectedViolations)
		})
	}
}

func TestValidateTags_AllowedValues(t *testing.T) {
	testCases := []struct {
		name               string
		tags               map[string]string
		expectedResult     bool
		expectedViolations int
	}{
		{
			name: "Valid environment tag",
			tags: map[string]string{
				"Environment": "production",
				"Owner":       "team@company.com",
			},
			expectedResult:     true,
			expectedViolations: 0,
		},
		{
			name: "Invalid environment tag",
			tags: map[string]string{
				"Environment": "invalid-env",
				"Owner":       "team@company.com",
			},
			expectedResult:     false,
			expectedViolations: 1,
		},
	}

	config := createTestConfig()
	validator := NewTagValidator(config)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := validator.ValidateTags(tc.tags)
			assert.Equal(t, tc.expectedResult, result.IsCompliant)
			assert.Len(t, result.Violations, tc.expectedViolations)
		})
	}
}

func TestValidateTags_CaseRules(t *testing.T) {
	testCases := []struct {
		name               string
		tags               map[string]string
		expectedResult     bool
		expectedViolations int
	}{
		{
			name: "Valid lowercase tags",
			tags: map[string]string{
				"Environment": "production",
				"Owner":       "team@company.com",
			},
			expectedResult:     true,
			expectedViolations: 0,
		},
		{
			name: "Invalid case for environment tag",
			tags: map[string]string{
				"Environment": "PRODUCTION",
				"Owner":       "team@company.com",
			},
			expectedResult:     false,
			expectedViolations: 1,
		},
	}

	config := createTestConfig()
	validator := NewTagValidator(config)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := validator.ValidateTags(tc.tags)
			assert.Equal(t, tc.expectedResult, result.IsCompliant)
			assert.Len(t, result.Violations, tc.expectedViolations)
		})
	}
}

func TestValidateTags_PatternRules(t *testing.T) {
	testCases := []struct {
		name               string
		tags               map[string]string
		expectedResult     bool
		expectedViolations int
	}{
		{
			name: "Valid owner email",
			tags: map[string]string{
				"Environment": "production",
				"Owner":       "team@company.com",
			},
			expectedResult:     true,
			expectedViolations: 0,
		},
		{
			name: "Invalid owner email",
			tags: map[string]string{
				"Environment": "production",
				"Owner":       "invalid-email",
			},
			expectedResult:     false,
			expectedViolations: 1,
		},
	}

	config := createTestConfig()
	validator := NewTagValidator(config)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := validator.ValidateTags(tc.tags)
			assert.Equal(t, tc.expectedResult, result.IsCompliant)
			assert.Len(t, result.Violations, tc.expectedViolations)
		})
	}
}

func TestValidateTags_MultipleViolations(t *testing.T) {
	config := createTestConfig()
	validator := NewTagValidator(config)

	tags := map[string]string{
		"Environment": "INVALID-ENV",
		"Owner":       "invalid-email",
	}

	result := validator.ValidateTags(tags)
	assert.False(t, result.IsCompliant)
	assert.Len(t, result.Violations, 3) // Invalid environment, invalid email, case violation
}
