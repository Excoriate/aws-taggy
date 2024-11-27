package types

// Violation represents a specific compliance violation
type Violation struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

// ComplianceResult represents a single compliance validation result
type ComplianceResult struct {
	IsCompliant     bool              `json:"is_compliant"`
	ResourceTags    map[string]string `json:"resource_tags"`
	Violations      []Violation       `json:"violations,omitempty"`
	ComplianceLevel string            `json:"compliance_level,omitempty"`
}

// ComplianceSummary provides an overview of compliance results
type ComplianceSummary struct {
	TotalResources        int            `json:"total_resources"`
	CompliantResources    int            `json:"compliant_resources"`
	NonCompliantResources int            `json:"non_compliant_resources"`
	GlobalViolations      map[string]int `json:"global_violations,omitempty"`
}

// ValidationResult represents the comprehensive validation result
type ValidationResult struct {
	File              string              `json:"file"`
	Valid             bool                `json:"valid"`
	Status            string              `json:"status"`
	Version           string              `json:"version"`
	ComplianceResults []*ComplianceResult `json:"compliance_results,omitempty"`
	ComplianceSummary *ComplianceSummary  `json:"compliance_summary,omitempty"`
}
