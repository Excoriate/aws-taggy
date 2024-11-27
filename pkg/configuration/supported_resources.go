package configuration

import (
	"fmt"
	"strings"

	"github.com/Excoriate/aws-taggy/pkg/constants"
)

var SupportedAWSResources = map[string]bool{
	constants.ResourceTypeS3:         true,
	constants.ResourceTypeEC2:        true,
	constants.ResourceTypeVPC:        true,
	constants.ResourceTypeRDS:        false,
	constants.ResourceTypeLambda:     false,
	constants.ResourceTypeEKS:        false,
	constants.ResourceTypeECR:        false,
	constants.ResourceTypeCloudfront: false,
	constants.ResourceTypeRoute53:    false,
}

// NormalizeResourceType normalizes the resource type string by:
// 1. Converting to lowercase
// 2. Trimming whitespace
// 3. Handling common aliases (e.g., "vpc" and "VPC" both map to constants.ResourceTypeVPC)
//
// Parameters:
//   - resource: A string representing the AWS resource type to normalize
//
// Returns:
//   - The normalized resource type string
func NormalizeResourceType(resource string) string {
	// Convert to lowercase and trim spaces
	normalized := strings.ToLower(strings.TrimSpace(resource))

	// Add any specific normalizations here if needed
	// For example, if we want to support "virtual-private-cloud" as an alias for "vpc"
	switch normalized {
	case "virtual-private-cloud", "vpc":
		return constants.ResourceTypeVPC
	case "elastic-compute-cloud", "ec2":
		return constants.ResourceTypeEC2
	case "simple-storage-service", "s3":
		return constants.ResourceTypeS3
	default:
		return normalized
	}
}

// IsSupportedAWSResource checks if the given AWS resource type is supported by the application.
// It performs the following validations:
// 1. Normalizes the resource type (case-insensitive, trimmed, aliases handled)
// 2. Checks if the normalized resource type exists in the predefined SupportedAWSResources map
// 3. Verifies that the resource type is enabled (value is true in the map)
//
// Parameters:
//   - resource: A string representing the AWS resource type to validate
//
// Returns:
//   - An error if the resource type is not supported or not enabled, nil otherwise
//
// Example usage:
//
//	err := IsSupportedAWSResource("vpc")  // Supports "vpc", "VPC", etc.
//	if err != nil {
//	    // Handle unsupported resource type
//	}
func IsSupportedAWSResource(resource string) error {
	normalized := NormalizeResourceType(resource)

	supported, exists := SupportedAWSResources[normalized]
	if !exists {
		return fmt.Errorf("unsupported resource type: %s", resource)
	}

	if !supported {
		return fmt.Errorf("resource type %s is not currently enabled", resource)
	}

	return nil
}
