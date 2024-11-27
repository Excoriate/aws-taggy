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

	// ViolationTypeInvalidKeyFormat indicates a tag key that doesn't follow the required format
	ViolationTypeInvalidKeyFormat ViolationType = "invalid_key_format"

	// ViolationTypeValueLength indicates a tag value that violates length constraints
	ViolationTypeValueLength ViolationType = "value_length_violation"

	// ViolationTypeProhibitedTag indicates use of a prohibited tag
	ViolationTypeProhibitedTag ViolationType = "prohibited_tag"

	// ViolationTypeExcessTags indicates exceeding the maximum number of allowed tags
	ViolationTypeExcessTags ViolationType = "excess_tags"
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
	// Type of rule (case, value, pattern, key_format, length, prohibited)
	Type string

	// Specific validation parameters
	Parameters map[string]interface{}

	// MinLength specifies minimum length for tag values
	MinLength *int `json:"min_length,omitempty"`

	// MaxLength specifies maximum length for tag values
	MaxLength *int `json:"max_length,omitempty"`

	// KeyPattern specifies regex pattern for tag keys
	KeyPattern string `json:"key_pattern,omitempty"`

	// Message provides a custom message for violations
	Message string `json:"message,omitempty"`
}

// RuleSet represents a collection of rules for tag validation
type RuleSet struct {
	// Rules for different tag keys
	Rules map[string]Rule
}
