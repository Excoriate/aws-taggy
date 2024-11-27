package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestComplianceResult(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		result   ComplianceResult
		expected bool
	}{
		{
			name: "Fully Compliant Result",
			result: ComplianceResult{
				IsCompliant:     true,
				ResourceTags:    map[string]string{"environment": "production"},
				Violations:      []Violation{},
				ComplianceLevel: "high",
			},
			expected: true,
		},
		{
			name: "Non-Compliant Result",
			result: ComplianceResult{
				IsCompliant: false,
				ResourceTags: map[string]string{
					"environment": "staging",
				},
				Violations: []Violation{
					{
						Type:    "missing_tags",
						Message: "Missing required tags",
					},
				},
				ComplianceLevel: "low",
			},
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.result.IsCompliant)
			assert.Equal(t, tc.expected, len(tc.result.Violations) == 0)
		})
	}
}

func TestComplianceSummary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		summary  ComplianceSummary
		expected int
	}{
		{
			name: "Mixed Compliance Summary",
			summary: ComplianceSummary{
				TotalResources:        10,
				CompliantResources:    6,
				NonCompliantResources: 4,
				GlobalViolations: map[string]int{
					"missing_tags": 3,
					"invalid_case": 1,
				},
			},
			expected: 10,
		},
		{
			name:     "Empty Summary",
			summary:  ComplianceSummary{},
			expected: 0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.summary.TotalResources)
		})
	}
}

func TestValidationResult(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		result   ValidationResult
		expected bool
	}{
		{
			name: "Valid Validation Result",
			result: ValidationResult{
				File:    "config.yaml",
				Valid:   true,
				Status:  "passed",
				Version: "1.0",
				ComplianceResults: []*ComplianceResult{
					{
						IsCompliant:     true,
						ResourceTags:    map[string]string{"environment": "production"},
						ComplianceLevel: "high",
					},
				},
			},
			expected: true,
		},
		{
			name: "Invalid Validation Result",
			result: ValidationResult{
				File:    "config.yaml",
				Valid:   false,
				Status:  "failed",
				Version: "1.0",
				ComplianceResults: []*ComplianceResult{
					{
						IsCompliant: false,
						Violations: []Violation{
							{
								Type:    "missing_tags",
								Message: "Missing required tags",
							},
						},
						ComplianceLevel: "low",
					},
				},
			},
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.result.Valid)
		})
	}
}
