package compliance

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestComplianceResult(t *testing.T) {
	testCases := []struct {
		name           string
		tags           map[string]string
		violations     []Violation
		expectedResult bool
	}{
		{
			name: "Fully compliant result",
			tags: map[string]string{
				"Environment": "production",
				"Owner":       "team@company.com",
			},
			violations:     []Violation{},
			expectedResult: true,
		},
		{
			name: "Non-compliant result with violations",
			tags: map[string]string{
				"Environment": "production",
			},
			violations: []Violation{
				{
					Type:    ViolationTypeMissingTags,
					Message: "Missing required tags",
				},
			},
			expectedResult: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := &ComplianceResult{
				IsCompliant:     len(tc.violations) == 0,
				Violations:      tc.violations,
				ResourceTags:    tc.tags,
				ComplianceLevel: ComplianceLevelStandard,
			}

			assert.Equal(t, tc.expectedResult, result.IsCompliant)
			assert.Equal(t, tc.tags, result.ResourceTags)
			assert.Len(t, result.Violations, len(tc.violations))
		})
	}
}

func TestGenerateSummary(t *testing.T) {
	testResults := []*ComplianceResult{
		{
			IsCompliant: true,
			ResourceTags: map[string]string{
				"Environment": "production",
				"Owner":       "team@company.com",
			},
		},
		{
			IsCompliant: false,
			Violations: []Violation{
				{
					Type:    ViolationTypeMissingTags,
					Message: "Missing required tags",
				},
			},
		},
		{
			IsCompliant: false,
			Violations: []Violation{
				{
					Type:    ViolationTypeInvalidValue,
					Message: "Invalid tag value",
				},
			},
		},
	}

	summary := GenerateSummary(testResults)

	assert.Equal(t, 3, summary.TotalResources)
	assert.Equal(t, 1, summary.CompliantResources)
	assert.Equal(t, 2, summary.NonCompliantResources)
	assert.Len(t, summary.GlobalViolations, 2)
	assert.Equal(t, 1, summary.GlobalViolations[ViolationTypeMissingTags])
	assert.Equal(t, 1, summary.GlobalViolations[ViolationTypeInvalidValue])
}
