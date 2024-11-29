package compliance

import (
	"fmt"
	"strings"
)

// Violation represents a specific tag compliance violation
type Violation struct {
	// Type of violation
	Type ViolationType

	// Detailed message explaining the violation
	Message string

	// Tag key associated with the violation (if applicable)
	TagKey string

	// Suggested fix or correction (optional)
	SuggestedFix string
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

	// Resource type (e.g., s3, ec2)
	ResourceType string
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

	// Compliance level distribution
	ComplianceLevelDistribution map[ComplianceLevel]int

	// Resource type compliance summary
	ResourceTypeCompliance map[string]float64
}

// GenerateSummary creates a summary from multiple compliance results
func GenerateSummary(results []*ComplianceResult) *Summary {
	summary := &Summary{
		TotalResources:              len(results),
		GlobalViolations:            make(map[ViolationType]int),
		ComplianceLevelDistribution: make(map[ComplianceLevel]int),
		ResourceTypeCompliance:      make(map[string]float64),
	}

	resourceTypeCount := make(map[string]int)

	for _, result := range results {
		// Track compliance levels
		summary.ComplianceLevelDistribution[result.ComplianceLevel]++

		// Track resource type compliance
		resourceTypeCount[result.ResourceType]++
		if result.IsCompliant {
			summary.CompliantResources++
		} else {
			summary.NonCompliantResources++

			// Track global violations
			for _, violation := range result.Violations {
				summary.GlobalViolations[violation.Type]++
			}
		}
	}

	// Calculate resource type compliance percentages
	for resourceType, count := range resourceTypeCount {
		compliantCount := 0
		for _, result := range results {
			if result.ResourceType == resourceType && result.IsCompliant {
				compliantCount++
			}
		}
		summary.ResourceTypeCompliance[resourceType] = float64(compliantCount) / float64(count) * 100
	}

	return summary
}

// String provides a human-readable summary of the compliance result
func (cr *ComplianceResult) String() string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Compliance Status: %v\n", cr.IsCompliant))
	sb.WriteString(fmt.Sprintf("Compliance Level: %s\n", cr.ComplianceLevel))
	sb.WriteString(fmt.Sprintf("Resource Type: %s\n", cr.ResourceType))

	if !cr.IsCompliant {
		sb.WriteString("Violations:\n")
		for _, violation := range cr.Violations {
			sb.WriteString(fmt.Sprintf("- %s: %s\n", violation.Type, violation.Message))
			if violation.SuggestedFix != "" {
				sb.WriteString(fmt.Sprintf("  Suggested Fix: %s\n", violation.SuggestedFix))
			}
		}
	}

	return sb.String()
}

// ToJSON converts the ComplianceResult to a JSON-friendly map
func (cr *ComplianceResult) ToJSON() map[string]interface{} {
	return map[string]interface{}{
		"is_compliant":     cr.IsCompliant,
		"compliance_level": cr.ComplianceLevel,
		"resource_type":    cr.ResourceType,
		"resource_tags":    cr.ResourceTags,
		"violations":       cr.Violations,
	}
}

// Merge combines multiple compliance results into a single result
func Merge(results []*ComplianceResult) *ComplianceResult {
	if len(results) == 0 {
		return nil
	}

	mergedResult := &ComplianceResult{
		IsCompliant:     true,
		ResourceTags:    make(map[string]string),
		Violations:      []Violation{},
		ComplianceLevel: ComplianceLevelLow,
		ResourceType:    results[0].ResourceType,
	}

	// Determine the lowest compliance level
	for _, result := range results {
		if !result.IsCompliant {
			mergedResult.IsCompliant = false
		}

		// Merge tags
		for k, v := range result.ResourceTags {
			mergedResult.ResourceTags[k] = v
		}

		// Merge violations
		mergedResult.Violations = append(mergedResult.Violations, result.Violations...)

		// Set the most stringent compliance level
		if result.ComplianceLevel == ComplianceLevelHigh {
			mergedResult.ComplianceLevel = ComplianceLevelHigh
		} else if result.ComplianceLevel == ComplianceLevelStandard &&
			mergedResult.ComplianceLevel != ComplianceLevelHigh {
			mergedResult.ComplianceLevel = ComplianceLevelStandard
		}
	}

	return mergedResult
}
