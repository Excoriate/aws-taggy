package taggy

import (
	"fmt"

	"github.com/Excoriate/aws-taggy/pkg/configuration"
)

// ScannerEngine is the primary implementation of the Scanner interface.
// It encapsulates the configuration and provides methods to interact with
// the scanner's settings.
type TaggyClient struct {
	// config holds the configuration for the scanner, defining its behavior
	// and parameters.
	config *configuration.TaggyScanConfig
}

// New creates and initializes a new TaggyClient instance with the specified configuration.
//
// This function takes a configuration file path and optional scanner options as input.
// It performs the following steps:
// 1. Loads the configuration using the TaggyScanConfigLoader
// 2. Creates a new Scanner instance with the loaded configuration
// 3. Applies any provided functional options to customize the scanner
//
// Parameters:
//   - cfgFilePath: The file path to the configuration file that defines scanner settings
//   - opts: Optional variadic functional options to configure the scanner's behavior
//
// Returns:
//   - A pointer to the initialized Scanner instance
//   - An error if configuration loading or option application fails
//
// Example usage:
//   scanner, err := NewTaggyClient("/path/to/config.yaml", WithAWS())
func New(cfgFilePath string) (*TaggyClient, error) {
	loader := configuration.NewTaggyScanConfigLoader()
	config, err := loader.LoadConfig(cfgFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create a new taggy client: %w", err)
	}

	s := &TaggyClient{config: config}

	return s, nil
}

// Config returns the configuration associated with the Taggy client
func (s *TaggyClient) Config() *configuration.TaggyScanConfig {
	return s.config
}
