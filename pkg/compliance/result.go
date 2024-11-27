package compliance

// Violation represents a specific tag compliance violation
type Violation struct {
	// Type of violation
	Type ViolationType

	// Detailed message explaining the violation
	Message string
}

// ComplianceResult represents the result of tag compliance validation
type ComplianceResult struct {
	// Overall compliance status
	IsCompliant bool

	// List of specific violations
	Violations []Violation

	// Original resource tags
	ResourceTags map[string]string

	// Compliance level of the resource
	ComplianceLevel ComplianceLevel
}

// Summary provides a high-level overview of compliance results
type Summary struct {
	// Total number of resources scanned
	TotalResources int

	// Number of compliant resources
	CompliantResources int

	// Number of non-compliant resources
	NonCompliantResources int

	// Detailed violations across all resources
	GlobalViolations map[ViolationType]int
}

// GenerateSummary creates a summary from multiple compliance results
func GenerateSummary(results []*ComplianceResult) *Summary {
	summary := &Summary{
		TotalResources:   len(results),
		GlobalViolations: make(map[ViolationType]int),
	}

	for _, result := range results {
		if result.IsCompliant {
			summary.CompliantResources++
		} else {
			summary.NonCompliantResources++
			for _, violation := range result.Violations {
				summary.GlobalViolations[violation.Type]++
			}
		}
	}

	return summary
}
