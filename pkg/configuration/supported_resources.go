package configuration

import (
	"fmt"
	"strings"

	"github.com/Excoriate/aws-taggy/pkg/constants"
)

var SupportedAWSResources = map[string]bool{
	constants.ResourceTypeS3:         true,
	constants.ResourceTypeEC2:        true,
	constants.ResourceTypeRDS:        false,
	constants.ResourceTypeLambda:     false,
	constants.ResourceTypeEKS:        false,
	constants.ResourceTypeECR:        false,
	constants.ResourceTypeCloudfront: false,
	constants.ResourceTypeRoute53:    false,
}

// IsSupportedAWSResource checks if the given AWS resource type is supported by the application.
// It performs the following validations:
// 1. Trims any leading or trailing whitespace from the resource type
// 2. Checks if the resource type exists in the predefined SupportedAWSResources map
//
// Parameters:
//   - resource: A string representing the AWS resource type to validate
//
// Returns:
//   - An error if the resource type is not supported, nil otherwise
//
// Example usage:
//
//	err := IsSupportedAWSResource("s3")
//	if err != nil {
//	    // Handle unsupported resource type
//	}
func IsSupportedAWSResource(resource string) error {
	resource = strings.TrimSpace(resource)

	if !SupportedAWSResources[resource] {
		return fmt.Errorf("unsupported resource type: %s", resource)
	}

	return nil
}
