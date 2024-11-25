package scanner

import (
	"context"
	"fmt"
	"time"
)

// ResourceType constants matching those in scanconfig.go
const (
	ResourceTypeS3         = "s3"
	ResourceTypeEC2        = "ec2"
	ResourceTypeRDS        = "rds"
	ResourceTypeLambda     = "lambda"
	ResourceTypeEKS        = "eks"
	ResourceTypeECR        = "ecr"
	ResourceTypeCloudfront = "cloudfront"
	ResourceTypeRoute53    = "route53"
)

// ResourceMetadata represents the core information about a cloud resource
type ResourceMetadata struct {
	// Unique identifier for the resource
	ID string `json:"id"`
	
	// Type of the resource (e.g., "s3", "ec2")
	Type string `json:"type"`
	
	// Cloud provider specific details
	Provider string `json:"provider"`
	
	// Region where the resource exists
	Region string `json:"region"`
	
	// Account or subscription ID
	AccountID string `json:"account_id"`
	
	// Raw tags associated with the resource
	Tags map[string]string `json:"tags"`
	
	// Timestamp of resource discovery
	DiscoveredAt time.Time `json:"discovered_at"`
	
	// Additional provider-specific metadata
	RawMetadata interface{} `json:"raw_metadata,omitempty"`
}

// ScannerConfig defines the configuration for resource scanning
type ScannerConfig interface {
	// Validate checks the configuration for correctness
	Validate() error
	
	// GetRegions returns the regions to be scanned
	GetRegions() []string
	
	// GetResourceTypes returns the types of resources to scan
	GetResourceTypes() []string
}

// ScanResult represents the outcome of a resource scanning operation
type ScanResult struct {
	// Total number of resources discovered
	TotalResources int `json:"total_resources"`
	
	// Discovered resources
	Resources []ResourceMetadata `json:"resources"`
	
	// Scan metadata
	StartTime   time.Time     `json:"start_time"`
	EndTime     time.Time     `json:"end_time"`
	Duration    time.Duration `json:"duration"`
	
	// Scanned region
	Region string `json:"region"`
	
	// Any errors encountered during scanning
	Errors []string `json:"errors,omitempty"`
}

// Scanner defines the interface for cloud resource discovery
type Scanner interface {
	// Scan discovers resources based on the provided configuration
	//
	// Parameters:
	//   - ctx: Context for cancellation and timeout
	//   - region: Specific region to scan
	//
	// Returns:
	//   - ScanResult containing discovered resources
	//   - Any error encountered during scanning
	Scan(ctx context.Context, region string) (*ScanResult, error)
	
	// GetConfig returns the current scanner configuration
	GetConfig() ScannerConfig
	
	// SetConfig updates the scanner's configuration
	SetConfig(cfg ScannerConfig) error
}

// BaseScannerConfig provides a default implementation of ScannerConfig
type BaseScannerConfig struct {
	// Cloud provider (e.g., "aws", "azure", "gcp")
	Provider string `json:"provider"`
	
	// Regions to scan
	Regions []string `json:"regions"`
	
	// Resource types to scan (matching constants defined above)
	ResourceTypes []string `json:"resource_types"`
	
	// Optional batch size for scanning
	BatchSize int `json:"batch_size,omitempty"`
}

// Validate provides basic validation for the base scanner configuration
func (c *BaseScannerConfig) Validate() error {
	if c.Provider == "" {
		return fmt.Errorf("cloud provider must be specified")
	}
	
	if len(c.Regions) == 0 {
		return fmt.Errorf("at least one region must be specified")
	}
	
	if len(c.ResourceTypes) == 0 {
		return fmt.Errorf("at least one resource type must be specified")
	}
	
	// Validate resource types against known types
	validResourceTypes := map[string]bool{
		ResourceTypeS3:         true,
		ResourceTypeEC2:        true,
		ResourceTypeRDS:        true,
		ResourceTypeLambda:     true,
		ResourceTypeEKS:        true,
		ResourceTypeECR:        true,
		ResourceTypeCloudfront: true,
		ResourceTypeRoute53:    true,
	}
	
	for _, resourceType := range c.ResourceTypes {
		if !validResourceTypes[resourceType] {
			return fmt.Errorf("invalid resource type: %s", resourceType)
		}
	}
	
	return nil
}

// GetRegions returns the configured regions
func (c *BaseScannerConfig) GetRegions() []string {
	return c.Regions
}

// GetResourceTypes returns the configured resource types
func (c *BaseScannerConfig) GetResourceTypes() []string {
	return c.ResourceTypes
}
