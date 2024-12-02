package inspector

import (
	"fmt"

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
