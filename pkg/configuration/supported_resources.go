package configuration

import (
	"fmt"
	"strings"

	"github.com/Excoriate/aws-taggy/pkg/constants"
)

var SupportedAWSResources = map[string]bool{
	constants.ResourceTypeS3:             true,
	constants.ResourceTypeEC2:            true,
	constants.ResourceTypeVPC:            true,
	constants.ResourceTypeCloudWatchLogs: true,
	constants.ResourceTypeRoute53:        true,
	constants.ResourceTypeSNS:            true,
	constants.ResourceTypeRDS:            true,
	constants.ResourceTypeSQS:            true,
	constants.ResourceTypeLambda:         false,
	constants.ResourceTypeEKS:            false,
	constants.ResourceTypeECR:            false,
	constants.ResourceTypeCloudfront:     false,
}

var SupportedAWSRegions = map[string]bool{
	"us-east-1":      true,
	"us-east-2":      true,
	"us-west-1":      true,
	"us-west-2":      true,
	"ca-central-1":   true,
	"eu-central-1":   true,
	"eu-west-1":      true,
	"eu-west-2":      true,
	"eu-west-3":      true,
	"eu-north-1":     true,
	"ap-northeast-1": true,
	"ap-northeast-2": true,
	"ap-southeast-1": true,
	"ap-southeast-2": true,
	"ap-south-1":     true,
	"sa-east-1":      true,
	"me-south-1":     true,
	"af-south-1":     true,
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
	switch normalized {
	case "virtual-private-cloud", "vpc":
		return constants.ResourceTypeVPC
	case "elastic-compute-cloud", "ec2":
		return constants.ResourceTypeEC2
	case "simple-storage-service", "s3":
		return constants.ResourceTypeS3
	case "simple-notification-service", "sns":
		return constants.ResourceTypeSNS
	case "relational-database-service", "rds":
		return constants.ResourceTypeRDS
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
