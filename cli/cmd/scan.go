package cmd

import (
	"fmt"

	"github.com/Excoriate/aws-taggy/pkg/taggy"
)

// ScanCmd represents the scan subcommand
type ScanCmd struct {
	Config string `help:"Path to the tag compliance configuration file" required:"true"`
}

// Run method for ScanCmd implements the scanning logic
func (s *ScanCmd) Run() error {
	// Initialize Taggy client with the configuration file
	client, err := taggy.New(s.Config)
	if err != nil {
		return fmt.Errorf("failed to create Taggy client: %w", err)
	}

	// Perform full resource scanning
	return s.startScan(client)
}

// scanAllResources performs a comprehensive scan across all supported AWS resources
func (s *ScanCmd) startScan(client *taggy.TaggyClient) error {
	// Placeholder for full resource scanning implementation
	fmt.Println("Performing full AWS resource tag compliance scan...")

	// Iterate through enabled resources in the configuration
	for resourceType, resourceConfig := range client.Config().Resources {
		if resourceConfig.Enabled {
			fmt.Printf("Scanning resource type: %s\n", resourceType)
			// TODO: Implement actual scanning logic for each resource type
		}
	}

	return nil
}
