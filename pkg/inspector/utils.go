package inspector

import (
	"fmt"
	"regexp"

	"github.com/Excoriate/aws-taggy/pkg/configuration"
	"github.com/Excoriate/aws-taggy/pkg/constants"
)

// GetEffectiveRegions returns the list of regions to scan based on the configuration mode
func GetEffectiveRegions(cfg configuration.TaggyScanConfig) ([]string, error) {
	// If mode is 'all', return all valid AWS regions
	if cfg.AWS.Regions.Mode == "all" {
		regions := make([]string, 0, len(configuration.SupportedAWSRegions))
		for region, supported := range configuration.SupportedAWSRegions {
			if supported {
				regions = append(regions, region)
			}
		}
		return regions, nil
	}

	// If mode is 'specific' and regions are provided, validate and return those
	if len(cfg.AWS.Regions.List) > 0 {
		validRegions := make([]string, 0, len(cfg.AWS.Regions.List))
		invalidRegions := make([]string, 0)

		for _, region := range cfg.AWS.Regions.List {
			if supported, exists := configuration.SupportedAWSRegions[region]; exists && supported {
				validRegions = append(validRegions, region)
			} else {
				invalidRegions = append(invalidRegions, region)
			}
		}

		if len(invalidRegions) > 0 {
			return nil, fmt.Errorf("unsupported or disabled AWS regions: %v", invalidRegions)
		}

		return validRegions, nil
	}

	// Default to us-east-1 if no regions are specified
	return []string{constants.DefaultAWSRegion}, nil
}

// ExtractRegionFromARNOrDefault extracts the region from the ARN or defaults to us-east-1 if the ARN is not provided
func ExtractRegionFromARNOrDefault(arn string) string {
	if arn == "" {
		return constants.DefaultAWSRegion
	}

	extractedRegion, err := ExtractRegionFromARN(arn)
	if err != nil {
		return constants.DefaultAWSRegion
	}

	return extractedRegion
}

// ExtractRegionFromARN attempts to extract the region from a given AWS ARN
// It returns an error if the ARN is invalid or the region cannot be extracted
func ExtractRegionFromARN(arn string) (string, error) {
	if arn == "" {
		return "", fmt.Errorf("empty ARN provided")
	}

	// AWS ARN format: arn:aws:service:region:account-id:resource-type/resource-id
	// Regex pattern to extract region from ARN
	regionRegex := regexp.MustCompile(`arn:aws:[^:]+:([^:]+):`)
	matches := regionRegex.FindStringSubmatch(arn)

	if len(matches) < 2 {
		return "", fmt.Errorf("unable to extract region from ARN: %s", arn)
	}

	extractedRegion := matches[1]

	// Validate extracted region against supported regions
	if supported, exists := configuration.SupportedAWSRegions[extractedRegion]; exists && supported {
		return extractedRegion, nil
	}

	return "", fmt.Errorf("unsupported region extracted from ARN: %s", arn)
}
