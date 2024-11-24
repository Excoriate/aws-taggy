package scanner

import (
	"fmt"

	"github.com/Excoriate/aws-taggy/pkg/scannconfig"
)

// ScannerEngine is the primary implementation of the Scanner interface.
// It encapsulates the configuration and provides methods to interact with
// the scanner's settings.
type Scanner struct {
	// config holds the configuration for the scanner, defining its behavior
	// and parameters.
	config *scannconfig.TaggyScanConfig
}

// NewScanner creates and initializes a new Scanner instance with the specified configuration.
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
//   scanner, err := NewScanner("/path/to/config.yaml", WithAWS())
func NewScanner(cfgFilePath string) (*Scanner, error) {
	loader := scannconfig.NewTaggyScanConfigLoader()
	config, err := loader.LoadConfig(cfgFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create a new scanner engine: %w", err)
	}

	s := &Scanner{config: config}

	return s, nil
}
