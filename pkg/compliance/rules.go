package compliance

// ViolationType represents different types of tag compliance violations
type ViolationType string

const (
	// ViolationTypeMissingTags indicates missing required tags
	ViolationTypeMissingTags ViolationType = "missing_tags"

	// ViolationTypeCaseViolation indicates a tag that violates case rules
	ViolationTypeCaseViolation ViolationType = "case_violation"

	// ViolationTypeInvalidValue indicates a tag with a value not in the allowed list
	ViolationTypeInvalidValue ViolationType = "invalid_value"

	// ViolationTypePatternViolation indicates a tag that doesn't match the required pattern
	ViolationTypePatternViolation ViolationType = "pattern_violation"
)

// ComplianceLevel defines the strictness of tag compliance
type ComplianceLevel string

const (
	// ComplianceLevelHigh represents the strictest compliance level
	ComplianceLevelHigh ComplianceLevel = "high"

	// ComplianceLevelStandard represents a standard compliance level
	ComplianceLevelStandard ComplianceLevel = "standard"

	// ComplianceLevelLow represents a relaxed compliance level
	ComplianceLevelLow ComplianceLevel = "low"
)

// Rule represents a single tag validation rule
type Rule struct {
	// Type of rule (case, value, pattern)
	Type string

	// Specific validation parameters
	Parameters map[string]interface{}
}

// RuleSet represents a collection of rules for tag validation
type RuleSet struct {
	// Rules for different tag keys
	Rules map[string]Rule
}
