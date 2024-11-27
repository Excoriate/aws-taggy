package compliance

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestComplianceLevels(t *testing.T) {
	testCases := []struct {
		name     string
		level    ComplianceLevel
		expected string
	}{
		{
			name:     "High Compliance Level",
			level:    ComplianceLevelHigh,
			expected: "high",
		},
		{
			name:     "Standard Compliance Level",
			level:    ComplianceLevelStandard,
			expected: "standard",
		},
		{
			name:     "Low Compliance Level",
			level:    ComplianceLevelLow,
			expected: "low",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, string(tc.level))
		})
	}
}

func TestViolationTypes(t *testing.T) {
	testCases := []struct {
		name     string
		vType    ViolationType
		expected string
	}{
		{
			name:     "Missing Tags Violation",
			vType:    ViolationTypeMissingTags,
			expected: "missing_tags",
		},
		{
			name:     "Case Violation",
			vType:    ViolationTypeCaseViolation,
			expected: "case_violation",
		},
		{
			name:     "Invalid Value Violation",
			vType:    ViolationTypeInvalidValue,
			expected: "invalid_value",
		},
		{
			name:     "Pattern Violation",
			vType:    ViolationTypePatternViolation,
			expected: "pattern_violation",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, string(tc.vType))
		})
	}
}

func TestRuleSet(t *testing.T) {
	ruleSet := RuleSet{
		Rules: map[string]Rule{
			"Environment": {
				Type: "case",
				Parameters: map[string]interface{}{
					"case": "lowercase",
				},
			},
			"Owner": {
				Type: "pattern",
				Parameters: map[string]interface{}{
					"pattern": `^[a-z0-9._%+-]+@company\.com$`,
				},
			},
		},
	}

	assert.Len(t, ruleSet.Rules, 2)
	assert.Contains(t, ruleSet.Rules, "Environment")
	assert.Contains(t, ruleSet.Rules, "Owner")

	envRule := ruleSet.Rules["Environment"]
	assert.Equal(t, "case", envRule.Type)
	assert.Equal(t, "lowercase", envRule.Parameters["case"])

	ownerRule := ruleSet.Rules["Owner"]
	assert.Equal(t, "pattern", ownerRule.Type)
	assert.Equal(t, `^[a-z0-9._%+-]+@company\.com$`, ownerRule.Parameters["pattern"])
}
