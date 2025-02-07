package cmd

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/Excoriate/aws-taggy/cli/internal/normaliser"
	"github.com/Excoriate/aws-taggy/cli/internal/tui"
	"github.com/Excoriate/aws-taggy/pkg/configuration"
	"github.com/Excoriate/aws-taggy/pkg/inspector"
	"github.com/Excoriate/aws-taggy/pkg/o11y"
	"github.com/Excoriate/aws-taggy/pkg/output"
	"github.com/Excoriate/aws-taggy/pkg/taggy"
)

// DiscoverCmd represents the discover subcommand
type DiscoverCmd struct {
	Service   string `help:"AWS service to discover (e.g., s3, ec2)" required:"true"`
	Region    string `help:"AWS region to discover resources in" default:"us-east-1"`
	WithARN   bool   `help:"Include ARN in the output"`
	Output    string `help:"Output format (table|json|yaml|yml)" default:"table" enum:"table,json,yaml,yml,TABLE,JSON,YAML,YML"`
	Untagged  bool   `help:"Only show resources without tags"`
	Clipboard bool   `help:"Copy the output to the clipboard"`
}

// Run method for DiscoverCmd implements the resource discovery logic
func (d *DiscoverCmd) Run() error {
	// Initialize logger
	logger := o11y.DefaultLogger()

	// Normalize service name
	d.Service = normaliser.NormalizeServiceName(d.Service)

	// Normalize output format to lowercase
	d.Output = normaliser.NormalizeOutputFormat(d.Output)

	// Validate service
	if err := configuration.IsSupportedAWSResource(d.Service); err != nil {
		return fmt.Errorf("service %s is not supported: %w", d.Service, err)
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
		return fmt.Errorf("failed to create Taggy client with custom configuration for service %s in region %s: %w", d.Service, d.Region, err)
	}

	// Perform resource discovery
	return d.discoverResources(client, logger)
}

// discoverResources performs resource discovery for a specific service and region
func (d *DiscoverCmd) discoverResources(client *taggy.TaggyClient, logger *o11y.Logger) error {
	ctx := context.Background()

	logger.Info(fmt.Sprintf("🔍 Discovering %s resources in region %s", d.Service, d.Region))

	// Create a inspector manager
	inspectorManager, err := inspector.NewInspectorManagerFromConfig(*client.Config())
	if err != nil {
		return fmt.Errorf("failed to create inspector manager for service %s in region %s: %w", d.Service, d.Region, err)
	}

	// Perform the scan
	if err := inspectorManager.Inspect(ctx); err != nil {
		return fmt.Errorf("resource discovery failed for service %s in region %s: %w", d.Service, d.Region, err)
	}

	// Process discovery results
	inspectResults := inspectorManager.GetResults()

	// Prepare table data
	type ResourceRow struct {
		ID       string `json:"id" yaml:"id"`
		Region   string `json:"region" yaml:"region"`
		HasTags  bool   `json:"has_tags" yaml:"has_tags"`
		TagCount int    `json:"tag_count" yaml:"tag_count"`
		ARN      string `json:"arn,omitempty" yaml:"arn,omitempty"`
	}

	var totalResources, resourcesWithTags int
	var resourceRows []ResourceRow

	// Process all resources regardless of region for S3 buckets
	if d.Service == "s3" {
		for _, result := range inspectResults {
			for _, resource := range result.Resources {
				hasTags := len(resource.Tags) > 0

				// Skip if we're only looking for untagged resources and this one has tags
				if d.Untagged && hasTags {
					continue
				}

				resourceRows = append(resourceRows, ResourceRow{
					ID:       resource.ID,
					Region:   resource.Region,
					HasTags:  hasTags,
					TagCount: len(resource.Tags),
					ARN:      resource.Details.ARN,
				})

				if hasTags {
					resourcesWithTags++
				}
				totalResources++
			}
		}
	} else {
		// For non-S3 resources, filter by specified region
		result, exists := inspectResults[d.Region]
		if !exists {
			logger.Info(fmt.Sprintf("No %s resources found in region %s", d.Service, d.Region))
			return nil
		}

		for _, resource := range result.Resources {
			hasTags := len(resource.Tags) > 0

			// Skip if we're only looking for untagged resources and this one has tags
			if d.Untagged && hasTags {
				continue
			}

			resourceRows = append(resourceRows, ResourceRow{
				ID:       resource.ID,
				Region:   d.Region,
				HasTags:  hasTags,
				TagCount: len(resource.Tags),
				ARN:      resource.Details.ARN,
			})

			if hasTags {
				resourcesWithTags++
			}
			totalResources++
		}
	}

	// Check if we found any resources after filtering
	if len(resourceRows) == 0 {
		if d.Untagged {
			logger.Info(fmt.Sprintf("No untagged %s resources found in region %s", d.Service, d.Region))
		} else {
			logger.Info(fmt.Sprintf("No %s resources found in region %s", d.Service, d.Region))
		}
		return nil
	}

	// Prepare clipboard output (always in YAML)
	type DiscoveryResult struct {
		Service           string        `json:"service" yaml:"service"`
		Region            string        `json:"region" yaml:"region"`
		TotalResources    int           `json:"total_resources" yaml:"total_resources"`
		TaggedResources   int           `json:"tagged_resources" yaml:"tagged_resources"`
		UntaggedResources int           `json:"untagged_resources" yaml:"untagged_resources"`
		Resources         []ResourceRow `json:"resources" yaml:"resources"`
	}

	clipboardOutput := DiscoveryResult{
		Service:           d.Service,
		Region:            d.Region,
		TotalResources:    totalResources,
		TaggedResources:   resourcesWithTags,
		UntaggedResources: totalResources - resourcesWithTags,
		Resources:         resourceRows,
	}

	// If clipboard flag is set, copy to clipboard in YAML
	if d.Clipboard {
		yamlFormatter := output.NewYAMLFormatter(false)
		clipboardContent, err := yamlFormatter.Format(clipboardOutput)
		if err != nil {
			return fmt.Errorf("failed to format clipboard output: %w", err)
		}

		// Use system clipboard
		cmd := exec.Command("pbcopy")
		cmd.Stdin = strings.NewReader(clipboardContent)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to copy resource discovery results to clipboard: %w", err)
		}

		logger.Info("✅ Resource discovery results copied to clipboard!")
	}

	// Create output formatter
	var formatter output.Formatter
	switch d.Output {
	case "json":
		formatter = output.NewJSONFormatter(false)
	case "yaml", "yml":
		formatter = output.NewYAMLFormatter(false)
	default:
		formatter = output.NewTableFormatter([]string{"Resource", "Region", "Has Tags", "Tag Count"})
	}

	// If using structured output (JSON/YAML), prepare the data structure
	if d.Output == "json" || d.Output == "yaml" || d.Output == "yml" {
		formattedOutput, err := formatter.Format(clipboardOutput)
		if err != nil {
			return fmt.Errorf("failed to format output: %w", err)
		}
		fmt.Println(formattedOutput)
		return nil
	}

	// Default table output
	columns := []tui.Column{
		{Title: "Resource", Key: "ID", Width: 60, Flexible: true, Align: "left"},
		{Title: "Region", Key: "Region", Width: 15, Align: "center"},
		{Title: "Has Tags", Key: "HasTags", Width: 12, Align: "center"},
		{Title: "Tag Count", Key: "TagCount", Width: 12, Align: "center"},
	}

	if d.WithARN {
		columns = append(columns, tui.Column{
			Title:    "ARN",
			Key:      "ARN",
			Width:    100,
			Flexible: false,
			Align:    "left",
		})
	}

	title := fmt.Sprintf("🏷️  %s Resource Discovery", d.Service)
	if d.Untagged {
		title = fmt.Sprintf("🏷️  Untagged %s Resources", d.Service)
	}
	title = fmt.Sprintf("%s (Total: %d, Tagged: %d, Untagged: %d)",
		title, totalResources, resourcesWithTags, totalResources-resourcesWithTags)

	tableOpts := tui.TableOptions{
		Title:           title,
		Columns:         columns,
		FlexibleColumns: true,
		AutoWidth:       true,
	}

	// Convert resourceRows to [][]string for RenderTable
	tableData := make([][]string, len(resourceRows))
	for i, row := range resourceRows {
		rowData := []string{
			row.ID,
			row.Region,
			fmt.Sprintf("%v", row.HasTags),
			fmt.Sprintf("%d", row.TagCount),
		}
		if d.WithARN {
			rowData = append(rowData, row.ARN)
		}
		tableData[i] = rowData
	}

	return tui.RenderTable(tableOpts, tableData)
}
