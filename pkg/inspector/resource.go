package inspector

import (
	"time"

	"github.com/Excoriate/aws-taggy/pkg/constants"
)

// Resource is a core interface that defines the basic contract for any resource
// managed by the inspector package. It provides essential methods for retrieving
// key metadata about a resource.
//
// The interface is designed to be flexible and generic, allowing different types
// of resources (such as AWS, Azure, GCP resources) to implement these core methods.
//
// Key methods include:
//   - GetTags(): Retrieves all tags associated with the resource
//   - GetRegion(): Determines the geographical region of the resource
//   - GetType(): Identifies the specific type of the resource
type Resource interface {
	// GetTags returns a map of key-value pairs representing the resource's tags.
	// If no tags are present, it should return an empty map.
	GetTags() map[string]string

	// GetRegion returns the geographical region where the resource is located.
	// This could be an AWS region, Azure region, or any other cloud provider's region.
	// If no region is specified, it may return a default or empty string.
	GetRegion() string

	// GetType returns a string representing the specific type of the resource.
	// This could be something like "EC2Instance", "S3Bucket", "RDSDatabase", etc.
	GetType() string
}

// ResourceMetadata is a comprehensive struct that encapsulates detailed information about a cloud resource.
// It provides a rich, structured representation of a resource's metadata, including identification,
// discovery, and compliance-related information.
//
// The struct is designed to be flexible and support various cloud providers and resource types.
// It includes fields for basic resource identification, extended details, and the ability to store
// the original API response.
//
// Key features:
//   - Supports multiple cloud providers and resource types
//   - Captures resource identification details (ID, type, provider, region)
//   - Tracks resource tags and discovery timestamp
//   - Provides extended details including ARN, name, status, and custom properties
//   - Includes compliance tracking with violations and last check timestamp
//   - Allows storing the raw API response for further inspection
type ResourceMetadata struct {
	// Basic resource identification
	ID           string            `json:"id"`            // Unique identifier for the resource
	Type         string            `json:"type"`          // Type of the resource (e.g., EC2Instance, S3Bucket)
	Provider     string            `json:"provider"`      // Cloud provider (e.g., AWS, Azure, GCP)
	Region       string            `json:"region"`        // Geographical region of the resource
	AccountID    string            `json:"account_id"`    // Cloud account or subscription ID
	Tags         map[string]string `json:"tags"`          // Key-value pairs of resource tags
	DiscoveredAt time.Time         `json:"discovered_at"` // Timestamp when the resource was discovered

	// Extended information about the resource
	Details struct {
		ARN        string                 `json:"arn,omitempty"`        // Amazon Resource Name or equivalent
		Name       string                 `json:"name,omitempty"`       // Human-readable name of the resource
		Status     string                 `json:"status,omitempty"`     // Current status of the resource
		Properties map[string]interface{} `json:"properties,omitempty"` // Additional custom properties

		// Compliance information for the resource
		Compliance struct {
			IsCompliant bool      `json:"is_compliant"`         // Whether the resource meets defined compliance standards
			Violations  []string  `json:"violations,omitempty"` // List of compliance violations
			LastCheck   time.Time `json:"last_check"`           // Timestamp of the last compliance check
		} `json:"compliance"`
	} `json:"details"`

	// RawResponse stores the complete, unmodified API response
	// This can be useful for debugging or additional custom processing
	RawResponse interface{} `json:"raw_response,omitempty"`
}

// BaseResource is a fundamental implementation of the Resource interface that provides
// basic functionality for representing cloud resources across different providers.
//
// This struct serves as a common base for various resource types, offering standard
// properties like resource type, region, and tags. It can be embedded or used directly
// to quickly create resource representations with minimal configuration.
//
// The struct is designed to be flexible and can be used with different cloud providers
// and resource types, providing a consistent interface for resource metadata.
//
// Key features:
//   - Stores the resource type as a string
//   - Captures the region where the resource is located
//   - Maintains a map of tags associated with the resource
type BaseResource struct {
	// Type represents the specific type of the resource (e.g., "EC2Instance", "S3Bucket")
	Type string

	// Region indicates the geographical region or zone where the resource is deployed
	Region string

	// Tags are key-value pairs that provide additional metadata or classification for the resource
	Tags map[string]string
}

// GetType returns the type of the resource as a string.
//
// This method is part of the Resource interface implementation for BaseResource.
// It provides a simple way to retrieve the resource type, which can be useful
// for identification, filtering, or logging purposes.
//
// Returns:
//   - A string representing the specific type of the resource (e.g., "EC2Instance", "S3Bucket")
//
// Example:
//
//	baseResource := &BaseResource{Type: "EC2Instance"}
//	resourceType := baseResource.GetType() // Returns "EC2Instance"
func (r *BaseResource) GetType() string {
	return r.Type
}

// GetRegion returns the region of the resource, with a fallback to a default AWS region if not specified.
//
// This method is part of the Resource interface implementation for BaseResource.
// It provides a way to retrieve the geographical region or zone where the resource is deployed.
// If no region is explicitly set, it defaults to a predefined default AWS region.
//
// The method ensures that there's always a valid region associated with the resource,
// which can be crucial for operations that require regional context.
//
// Returns:
//   - A string representing the resource's region
//   - If no region is set, returns the default AWS region from constants
//
// Example:
//
//	baseResource := &BaseResource{Region: "us-west-2"}
//	region := baseResource.GetRegion() // Returns "us-west-2"
//
//	emptyResource := &BaseResource{}
//	defaultRegion := emptyResource.GetRegion() // Returns constants.DefaultAWSRegion
func (r *BaseResource) GetRegion() string {
	if r.Region == "" {
		return constants.DefaultAWSRegion
	}
	return r.Region
}

// GetTags retrieves the tags associated with the resource, ensuring a non-nil map is always returned.
//
// This method is part of the Resource interface implementation for BaseResource.
// It provides a way to access the resource's tags, which are key-value pairs that offer additional
// metadata or classification information about the resource.
//
// The method performs a lazy initialization of the Tags map if it is currently nil. This ensures
// that callers can always safely add or access tags without worrying about nil pointer exceptions.
//
// Returns:
//   - A map[string]string containing the resource's tags
//   - If no tags have been set, returns an empty map
//
// Behavior:
//   - If r.Tags is nil, it initializes a new empty map
//   - Subsequent calls will return the same map, allowing for tag modifications
//
// Example:
//
//	baseResource := &BaseResource{}
//	tags := baseResource.GetTags()  // Returns an empty map
//	tags["Environment"] = "Production"  // Adds a tag to the resource
//
//	anotherResource := &BaseResource{Tags: map[string]string{"Project": "CloudInspector"}}
//	existingTags := anotherResource.GetTags()  // Returns the existing tags
func (r *BaseResource) GetTags() map[string]string {
	if r.Tags == nil {
		r.Tags = make(map[string]string)
	}
	return r.Tags
}

// NewResourceType creates a new Resource instance with a specified resource type.
//
// This function is a factory method for creating a new BaseResource with a given type.
// It initializes a new resource with the provided type and an empty tags map, ensuring
// that the resource is ready for immediate use without additional initialization.
//
// Parameters:
//   - resourceType: A string representing the type of the resource (e.g., "EC2", "S3", "RDS")
//
// Returns:
//   - A Resource interface implementation (BaseResource) with the specified type
//
// The returned resource has:
//   - The Type field set to the provided resourceType
//   - An initialized, empty Tags map to prevent nil pointer issues
//
// Example:
//
//	ec2Resource := NewResourceType("EC2")
//	s3Resource := NewResourceType("S3")
//	ec2Resource.GetTags()["Name"] = "MyInstance"  // Can immediately add tags
func NewResourceType(resourceType string) Resource {
	return &BaseResource{
		Type: resourceType,
		Tags: make(map[string]string),
	}
}
