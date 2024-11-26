package taggy

import (
	"fmt"

	"github.com/Excoriate/aws-taggy/pkg/configuration"
)

// TaggyClient represents the main client for AWS resource tagging operations
type TaggyClient struct {
	config *configuration.TaggyScanConfig
}

// Config returns the current configuration
func (c *TaggyClient) Config() *configuration.TaggyScanConfig {
	return c.config
}

// New creates and initializes a new TaggyClient instance with configuration from a file
func New(cfgFilePath string) (*TaggyClient, error) {
	loader := configuration.NewTaggyScanConfigLoader()
	config, err := loader.LoadConfig(cfgFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create a new taggy client: %w", err)
	}

	return &TaggyClient{config: config}, nil
}

// NewWithConfig creates a new TaggyClient with the provided configuration
func NewWithConfig(config *configuration.TaggyScanConfig) (*TaggyClient, error) {
	if config == nil {
		return nil, fmt.Errorf("configuration cannot be nil")
	}

	return &TaggyClient{
		config: config,
	}, nil
}
