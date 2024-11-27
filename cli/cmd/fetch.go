package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/Excoriate/aws-taggy/cli/internal/output"
	"github.com/Excoriate/aws-taggy/cli/internal/tui"
	"github.com/Excoriate/aws-taggy/pkg/configuration"
	"github.com/Excoriate/aws-taggy/pkg/o11y"
	"github.com/Excoriate/aws-taggy/pkg/scanner"
	"github.com/Excoriate/aws-taggy/pkg/taggy"
)

// FetchCmd represents the fetch command and its subcommands
type FetchCmd struct {
	Tags TagsCmd `cmd:"" help:"Fetch tags for a specific AWS resource"`
	Info InfoCmd `cmd:"" help:"Fetch detailed information about a specific AWS resource"`
}

// TagsCmd represents the fetch tags subcommand
type TagsCmd struct {
	ARN       string `help:"ARN of the resource to fetch tags for" required:"true"`
	Service   string `help:"AWS service type (e.g., s3, ec2)" required:"true"`
	Output    string `help:"Output format (table|json|yaml)" default:"table" enum:"table,json,yaml"`
	Clipboard bool   `help:"Copy output to clipboard" default:"false"`
}

// InfoCmd represents the fetch info subcommand
type InfoCmd struct {
	ARN       string `help:"ARN of the resource to fetch information for" required:"true"`
	Service   string `help:"AWS service type (e.g., s3, ec2)" required:"true"`
	Output    string `help:"Output format (table|json|yaml)" default:"table" enum:"table,json,yaml"`
	Clipboard bool   `help:"Copy output to clipboard" default:"false"`
}

// Run implements the tags fetch logic
func (t *TagsCmd) Run() error {
	logger := o11y.DefaultLogger()
	logger.Info(fmt.Sprintf("🔍 Fetching tags for resource: %s", t.ARN))

	// Create minimal config for the specific service
	config := configuration.TaggyScanConfig{
		AWS: configuration.AWSConfig{
			Regions: configuration.RegionsConfig{
				Mode: "specific",
				List: []string{extractRegionFromARN(t.ARN)},
			},
		},
		Resources: map[string]configuration.ResourceConfig{
			t.Service: {
				Enabled: true,
			},
		},
	}

	// Initialize client
	client, err := taggy.NewWithConfig(&config)
	if err != nil {
		return fmt.Errorf("failed to initialize client: %w", err)
	}

	// Create scanner for the specific service
	scanner, err := scanner.NewScanner(t.Service, *client.Config())
	if err != nil {
		return fmt.Errorf("failed to create scanner: %w", err)
	}

	// Fetch resource details
	ctx := context.Background()
	resource, err := scanner.Fetch(ctx, t.ARN, *client.Config())
	if err != nil {
		return fmt.Errorf("failed to fetch resource: %w", err)
	}

	// Create output formatter
	formatter := output.NewFormatter(t.Output)

	if formatter.IsStructured() {
		type TagsResult struct {
			Resource string            `json:"resource" yaml:"resource"`
			ARN      string            `json:"arn" yaml:"arn"`
			Tags     map[string]string `json:"tags" yaml:"tags"`
		}

		result := TagsResult{
			Resource: resource.ID,
			ARN:      t.ARN,
			Tags:     resource.Tags,
		}

		// If clipboard flag is set, copy to clipboard
		if t.Clipboard {
			if err := output.CopyToClipboard(result); err != nil {
				return fmt.Errorf("failed to copy to clipboard: %w", err)
			}
			fmt.Println("✅ Resource tags copied to clipboard!")
			return nil
		}

		return formatter.Output(result)
	}

	// Prepare table data
	type TagRow struct {
		Key   string
		Value string
	}

	var tagRows []TagRow
	for key, value := range resource.Tags {
		tagRows = append(tagRows, TagRow{
			Key:   key,
			Value: value,
		})
	}

	// Create and render table
	tableOpts := tui.TableOptions{
		Title: fmt.Sprintf("🏷️  Tags for %s", shortenARN(t.ARN)),
		Columns: []tui.Column{
			{Title: "Key", Key: "Key", Width: 30, Flexible: true},
			{Title: "Value", Key: "Value", Width: 50, Flexible: true},
		},
		FlexibleColumns: true,
	}

	tableModel := tui.NewTableModel(tableOpts, tagRows)
	return tableModel.Render()
}

// Run implements the info fetch logic
func (i *InfoCmd) Run() error {
	logger := o11y.DefaultLogger()
	logger.Info(fmt.Sprintf("🔍 Fetching information for resource: %s", i.ARN))

	// Similar initialization as TagsCmd
	config := configuration.TaggyScanConfig{
		AWS: configuration.AWSConfig{
			Regions: configuration.RegionsConfig{
				Mode: "specific",
				List: []string{extractRegionFromARN(i.ARN)},
			},
		},
		Resources: map[string]configuration.ResourceConfig{
			i.Service: {
				Enabled: true,
			},
		},
	}

	client, err := taggy.NewWithConfig(&config)
	if err != nil {
		return fmt.Errorf("failed to initialize client: %w", err)
	}

	scanner, err := scanner.NewScanner(i.Service, *client.Config())
	if err != nil {
		return fmt.Errorf("failed to create scanner: %w", err)
	}

	ctx := context.Background()
	resource, err := scanner.Fetch(ctx, i.ARN, *client.Config())
	if err != nil {
		return fmt.Errorf("failed to fetch resource: %w", err)
	}

	// Create output formatter
	formatter := output.NewFormatter(i.Output)

	if formatter.IsStructured() {
		// If clipboard flag is set, copy to clipboard
		if i.Clipboard {
			if err := output.CopyToClipboard(resource); err != nil {
				return fmt.Errorf("failed to copy to clipboard: %w", err)
			}
			fmt.Println("✅ Resource information copied to clipboard!")
			return nil
		}

		return formatter.Output(resource)
	}

	// Prepare table data for resource details
	type DetailRow struct {
		Property string
		Value    string
	}

	var detailRows []DetailRow
	detailRows = append(detailRows,
		DetailRow{Property: "ID", Value: resource.ID},
		DetailRow{Property: "Type", Value: resource.Type},
		DetailRow{Property: "Region", Value: resource.Region},
		DetailRow{Property: "Provider", Value: resource.Provider},
		DetailRow{Property: "Tag Count", Value: fmt.Sprintf("%d", len(resource.Tags))},
		DetailRow{Property: "ARN", Value: resource.Details.ARN},
	)

	// Add any additional properties from Details.Properties
	for k, v := range resource.Details.Properties {
		detailRows = append(detailRows, DetailRow{
			Property: k,
			Value:    fmt.Sprintf("%v", v),
		})
	}

	// Create and render table
	tableOpts := tui.TableOptions{
		Title: fmt.Sprintf("ℹ️  Resource Details for %s", shortenARN(i.ARN)),
		Columns: []tui.Column{
			{Title: "Property", Key: "Property", Width: 20},
			{Title: "Value", Key: "Value", Width: 60, Flexible: true},
		},
		FlexibleColumns: true,
	}

	tableModel := tui.NewTableModel(tableOpts, detailRows)
	return tableModel.Render()
}

// Helper functions
func extractRegionFromARN(arn string) string {
	parts := strings.Split(arn, ":")
	if len(parts) >= 4 {
		return parts[3]
	}
	return "us-east-1" // Default region if ARN parsing fails
}

func shortenARN(arn string) string {
	parts := strings.Split(arn, "/")
	if len(parts) > 1 {
		return "..." + parts[len(parts)-1]
	}
	return arn
}
