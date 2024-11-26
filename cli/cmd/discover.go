package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/Excoriate/aws-taggy/cli/internal/tui"
	"github.com/Excoriate/aws-taggy/pkg/configuration"
	"github.com/Excoriate/aws-taggy/pkg/o11y"
	"github.com/Excoriate/aws-taggy/pkg/scanner"
	"github.com/Excoriate/aws-taggy/pkg/taggy"
)

// DiscoverCmd represents the discover subcommand
type DiscoverCmd struct {
	Service string `help:"AWS service to discover (e.g., s3, ec2)" required:"true"`
	Region  string `help:"AWS region to discover resources in" default:"us-east-1"`
	WithARN bool   `help:"Include ARN in the output"`
}

// Run method for DiscoverCmd implements the resource discovery logic
func (d *DiscoverCmd) Run() error {
	// Initialize logger
	logger := o11y.DefaultLogger()

	// Validate service
	if err := configuration.IsSupportedAWSResource(d.Service); err != nil {
		return fmt.Errorf("unsupported service: %s. %w", d.Service, err)
	}

	// Create a custom configuration for the specific service and region
	customConfig := configuration.TaggyScanConfig{
		AWS: configuration.AWSConfig{
			Regions: configuration.RegionsConfig{
				Mode: "specific",
				List: []string{d.Region},
			},
		},
		Resources: map[string]configuration.ResourceConfig{
			d.Service: {
				Enabled: true,
				Regions: []string{d.Region},
			},
		},
	}

	// Create Taggy client with empty config since we'll use our custom config
	client, err := taggy.NewWithConfig(&customConfig)
	if err != nil {
		return fmt.Errorf("failed to create Taggy client: %w", err)
	}

	// Perform resource discovery
	return d.discoverResources(client, logger)
}

// discoverResources performs resource discovery for a specific service and region
func (d *DiscoverCmd) discoverResources(client *taggy.TaggyClient, logger *o11y.Logger) error {
	ctx := context.Background()
	startTime := time.Now()

	logger.Info(fmt.Sprintf("🔍 Discovering %s resources in region %s", d.Service, d.Region))

	// Create a scanner manager
	scannerManager, err := scanner.NewScannerManager(*client.Config())
	if err != nil {
		return fmt.Errorf("failed to create scanner manager: %w", err)
	}

	// Perform the scan
	if err := scannerManager.Scan(ctx); err != nil {
		return fmt.Errorf("discovery encountered errors: %v", err)
	}

	// Process discovery results
	results := scannerManager.GetResults()

	// Prepare table data
	type ResourceRow struct {
		ID       string
		Region   string
		HasTags  bool
		TagCount int
		ARN      string
	}

	var totalResources, resourcesWithTags int
	var resourceRows []ResourceRow
	for region, result := range results {
		if region != d.Region {
			continue
		}

		for _, resource := range result.Resources {
			resourceRows = append(resourceRows, ResourceRow{
				ID:       resource.ID,
				Region:   region,
				HasTags:  len(resource.Tags) > 0,
				TagCount: len(resource.Tags),
				ARN:      resource.Details.ARN,
			})
			if len(resource.Tags) > 0 {
				resourcesWithTags++
			}
			totalResources++
		}
	}

	columns := []tui.Column{
		{Title: "Resource", Key: "ID", Width: 30, Flexible: true},
		{Title: "Region", Key: "Region", Width: 15},
		{Title: "Has Tags", Key: "HasTags", Width: 10},
		{Title: "Tag Count", Key: "TagCount", Width: 10},
	}

	if d.WithARN {
		columns = append(columns, tui.Column{
			Title:    "ARN",
			Key:      "ARN",
			Width:    50,
			Flexible: true,
		})
	}

	tableOpts := tui.TableOptions{
		Title: fmt.Sprintf("🏷️  %s Resource Discovery (Total: %d, Tagged: %d)",
			d.Service, totalResources, resourcesWithTags),
		Columns:         columns,
		FlexibleColumns: true,
	}

	tableModel := tui.NewTableModel(tableOpts, resourceRows)

	// Render table to console
	if err := tableModel.Render(); err != nil {
		return fmt.Errorf("failed to render discovery results table: %w", err)
	}

	// Check for any errors
	if errors := scannerManager.GetErrors(); len(errors) > 0 {
		logger.Warn("Discovery completed with the following errors:")
		for _, err := range errors {
			logger.Error(err.Error())
		}
	}

	duration := time.Since(startTime)
	logger.Info(fmt.Sprintf("⏱️  Total discovery duration: %v", duration))

	return nil
}
