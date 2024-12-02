package inspector

import (
	"context"
	"fmt"
	"time"

	"github.com/Excoriate/aws-taggy/pkg/configuration"
	"github.com/Excoriate/aws-taggy/pkg/constants"
)

// InspectResult represents the outcome of a resource inspection operation
// InspectResult represents the comprehensive outcome of a cloud resource inspection operation.
// It encapsulates detailed information about the resources discovered, timing metrics,
// and any potential errors encountered during the scanning process.
type InspectResult struct {
	// Resources is a slice of ResourceMetadata containing detailed information about each discovered resource.
	// Each ResourceMetadata provides specific attributes and metadata for individual cloud resources.
	Resources []ResourceMetadata `json:"resources"`

	// StartTime records the precise moment when the inspection process began.
	// This allows for accurate tracking of the inspection's temporal context.
	StartTime time.Time `json:"start_time"`

	// EndTime captures the exact timestamp when the resource inspection was completed.
	// It enables precise measurement of the inspection duration.
	EndTime time.Time `json:"end_time"`

	// Duration represents the total time taken to complete the resource inspection.
	// It is calculated as the difference between EndTime and StartTime.
	Duration time.Duration `json:"duration"`

	// Region specifies the AWS region in which the resources were scanned.
	// This helps in identifying the geographical context of the discovered resources.
	Region string `json:"region"`

	// TotalResources indicates the total number of resources discovered during the inspection.
	// It provides a quick summary of the scan's scope.
	TotalResources int `json:"total_resources"`

	// Errors is an optional slice of error messages encountered during the inspection process.
	// If any errors occurred during resource discovery or processing, they will be captured here.
	Errors []string `json:"errors,omitempty"`
}

// Inspector defines the interface for cloud resource inspection operations
// Inspector defines an interface for cloud resource inspection and retrieval operations.
// It provides methods to discover and fetch detailed information about cloud resources.
// The interface is designed to be flexible and support various cloud resource types.
type Inspector interface {
	// Inspect performs a comprehensive discovery operation for resources of a specific type.
	// It scans and collects metadata for resources based on the provided configuration.
	//
	// Parameters:
	//   - ctx: A context.Context for managing request cancellation, timeouts, and passing request-scoped values.
	//   - config: A TaggyScanConfig containing scanning parameters and resource discovery settings.
	//
	// Returns:
	//   - *InspectResult: A pointer to an InspectResult containing discovered resources and metadata.
	//   - error: An error if the inspection process fails, otherwise nil.
	Inspect(ctx context.Context, config configuration.TaggyScanConfig) (*InspectResult, error)

	// Fetch retrieves detailed metadata for a specific resource identified by its ARN.
	// This method allows for targeted retrieval of individual resource information.
	//
	// Parameters:
	//   - ctx: A context.Context for managing request cancellation, timeouts, and passing request-scoped values.
	//   - arn: The Amazon Resource Name (ARN) of the specific resource to fetch.
	//   - config: A TaggyScanConfig containing configuration settings for the resource fetch operation.
	//
	// Returns:
	//   - *ResourceMetadata: A pointer to a ResourceMetadata containing detailed information about the resource.
	//   - error: An error if the resource fetch fails, otherwise nil.
	Fetch(ctx context.Context, arn string, config configuration.TaggyScanConfig) (*ResourceMetadata, error)
}

// New creates a new inspector for a specific resource type
// New creates a new Inspector instance for a specific AWS resource type.
//
// This function dynamically selects and initializes the appropriate scanner
// based on the provided resource type. It validates the input configuration
// and ensures that at least one AWS region is specified for the scan.
//
// Parameters:
//   - resourceType: A string representing the type of AWS resource to inspect
//     (e.g., "s3", "ec2", "vpc"). The type is case-insensitive and will be
//     normalized internally.
//   - cfg: A TaggyScanConfig containing AWS-specific configuration settings,
//     including the list of regions to scan.
//
// Returns:
//   - Inspector: An initialized inspector implementation specific to the
//     requested resource type, ready to perform resource discovery and fetching.
//   - error: An error if the resource type is unsupported or no regions are
//     specified. Returns nil if the inspector is successfully created.
//
// Supported resource types include:
//   - S3 (Simple Storage Service)
//   - EC2 (Elastic Compute Cloud)
//   - VPC (Virtual Private Cloud)
//   - Route 53 (AWS Route 53)
//
// Example usage:
//
//	inspector, err := New("s3", config)
//	if err != nil {
//	    // Handle error
//	}
func New(resourceType string, cfg configuration.TaggyScanConfig) (Inspector, error) {
	// Determine regions to use
	regions, err := GetEffectiveRegions(cfg)
	if err != nil {
		return nil, fmt.Errorf("error getting effective regions: %w", err)
	}

	switch resourceType {
	case constants.ResourceTypeS3:
		return NewS3Inspector(regions)
	case constants.ResourceTypeEC2:
		return NewEC2Scanner(regions)
	case constants.ResourceTypeVPC:
		return NewVPCInspector(regions)
	case constants.ResourceTypeCloudWatchLogs:
		return NewCloudWatchLogsInspector(regions)
	case constants.ResourceTypeRoute53:
		return NewRoute53Inspector(regions)
	default:
		return nil, fmt.Errorf("unsupported resource type: %s", resourceType)
	}
}
