package inspector

import (
	"context"
	"time"
)

// ResourceCost represents the financial aspects of a cloud resource
type ResourceCost struct {
	// Estimated monthly cost of the resource
	MonthlyCost float64 `json:"monthly_cost"`

	// Currency of the cost estimation
	Currency string `json:"currency"`

	// Detailed breakdown of costs by category
	CostBreakdown map[string]float64 `json:"cost_breakdown,omitempty"`

	// Metadata about the cost estimation
	Metadata struct {
		// Timestamp when cost was estimated
		EstimatedAt time.Time `json:"estimated_at"`

		// Source system for cost estimation
		SourceSystem string `json:"source_system"`

		// Confidence level of the cost estimation (0.0 to 1.0)
		Confidence float64 `json:"confidence"`
	} `json:"metadata"`
}

// ResourceUsage captures comprehensive usage metrics for a cloud resource
type ResourceUsage struct {
	// Total number of requests or interactions
	TotalRequests int64 `json:"total_requests"`

	// Duration of active resource utilization
	ActiveDuration time.Duration `json:"active_duration"`

	// Flexible map for resource-type specific usage metrics
	TypeSpecificMetrics map[string]interface{} `json:"type_specific_metrics,omitempty"`

	// Metadata about the usage collection
	Metadata struct {
		// Timestamp when usage was collected
		CollectedAt time.Time `json:"collected_at"`

		// Source system for usage metrics
		SourceSystem string `json:"source_system"`

		// Confidence level of the usage metrics (0.0 to 1.0)
		Confidence float64 `json:"confidence"`
	} `json:"metadata"`
}

// ResourceCostProvider is an interface for retrieving cost information
type ResourceCostProvider interface {
	// GetResourceCost retrieves the cost information for a resource.
	//
	// Parameters:
	//   - ctx: A context for managing request cancellation and timeouts
	//
	// Returns:
	//   - *ResourceCost: Detailed cost information about the resource
	//   - error: Any error encountered during cost retrieval
	GetResourceCost(ctx context.Context) (*ResourceCost, error)
}

// ResourceUsageProvider is an interface for retrieving usage metrics
type ResourceUsageProvider interface {
	// GetResourceUsage retrieves the usage metrics for a resource.
	//
	// Parameters:
	//   - ctx: A context for managing request cancellation and timeouts
	//
	// Returns:
	//   - *ResourceUsage: Comprehensive usage metrics for the resource
	//   - error: Any error encountered during usage metrics collection
	GetResourceUsage(ctx context.Context) (*ResourceUsage, error)
}

// ResourceInsightsAggregator is an optional higher-level interface
// that can combine cost and usage information
type ResourceInsightsAggregator interface {
	// GetResourceInsights provides a comprehensive view of resource information
	//
	// Parameters:
	//   - ctx: A context for managing request cancellation and timeouts
	//
	// Returns:
	//   - *ResourceCost: Cost information for the resource
	//   - *ResourceUsage: Usage metrics for the resource
	//   - error: Any error encountered during insights collection
	GetResourceInsights(ctx context.Context) (*ResourceCost, *ResourceUsage, error)
}
