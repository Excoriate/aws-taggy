package cmd

import (
	"fmt"

	"github.com/Excoriate/aws-taggy/pkg/configuration"
)

// ValidateCmd represents the validate subcommand
type ValidateCmd struct {
	Config string `help:"Path to the tag validation configuration file" required:"true"`
}

// Run method for ValidateCmd implements the configuration validation logic
func (v *ValidateCmd) Run() error {
	// Initialize configuration loader
	loader := configuration.NewTaggyScanConfigLoader()

	// Attempt to load and validate the configuration
	_, err := loader.LoadConfig(v.Config)
	if err != nil {
		return fmt.Errorf("configuration validation failed: %w", err)
	}

	// If we reach this point, the configuration is valid
	fmt.Printf("Configuration file %s is valid.\n", v.Config)

	return nil
}