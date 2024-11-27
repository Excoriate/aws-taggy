package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/Excoriate/aws-taggy/cli/internal/output"
	"github.com/Excoriate/aws-taggy/cli/internal/tui"
	"github.com/Excoriate/aws-taggy/pkg/configuration"
	"github.com/Excoriate/aws-taggy/pkg/inspector"
	"github.com/Excoriate/aws-taggy/pkg/o11y"
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

	// Create inspector for the specific service
	inspectorClient, err := inspector.New(t.Service, config)
	if err != nil {
		return fmt.Errorf("failed to create inspector: %w", err)
	}

	// Fetch resource details
	ctx := context.Background()
	resource, err := inspectorClient.Fetch(ctx, t.ARN, config)
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
			if err := output.WriteToClipboard(result); err != nil {
				return fmt.Errorf("failed to copy to clipboard: %w", err)
			}
			fmt.Println("✅ Resource tags copied to clipboard!")
			return nil
		}

		return formatter.Output(result)
	}

	// Prepare table data
	tableData := make([][]string, 0, len(resource.Tags))
	for key, value := range resource.Tags {
		tableData = append(tableData, []string{key, value})
	}

	// Create and render table for tags
	tableOpts := tui.TableOptions{
		Title: fmt.Sprintf("🏷️  Tags for %s", shortenARN(t.ARN)),
		Columns: []tui.Column{
			{Title: "Key", Width: 30, Flexible: true},
			{Title: "Value", Width: 50, Flexible: true},
		},
		FlexibleColumns: true,
		AutoWidth:       true,
	}

	return tui.RenderTable(tableOpts, tableData)
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

	inspectorClient, err := inspector.New(i.Service, config)
	if err != nil {
		return fmt.Errorf("failed to create inspector: %w", err)
	}

	ctx := context.Background()
	resource, err := inspectorClient.Fetch(ctx, i.ARN, config)
	if err != nil {
		return fmt.Errorf("failed to fetch resource: %w", err)
	}

	// Create output formatter
	formatter := output.NewFormatter(i.Output)

	if formatter.IsStructured() {
		// If clipboard flag is set, copy to clipboard
		if i.Clipboard {
			if err := output.WriteToClipboard(resource); err != nil {
				return fmt.Errorf("failed to copy to clipboard: %w", err)
			}
			fmt.Println("✅ Resource information copied to clipboard!")
			return nil
		}

		return formatter.Output(resource)
	}

	// Prepare table data for resource details
	tableData := [][]string{
		{"ID", resource.ID},
		{"Type", resource.Type},
		{"Region", resource.Region},
		{"Provider", resource.Provider},
		{"Tag Count", fmt.Sprintf("%d", len(resource.Tags))},
		{"ARN", resource.Details.ARN},
	}

	// Add any additional properties from Details.Properties
	for k, v := range resource.Details.Properties {
		tableData = append(tableData, []string{k, fmt.Sprintf("%v", v)})
	}

	// Create and render table for resource details
	tableOpts := tui.TableOptions{
		Title: fmt.Sprintf("ℹ️  Resource Details for %s", shortenARN(i.ARN)),
		Columns: []tui.Column{
			{Title: "Property", Width: 20},
			{Title: "Value", Width: 60, Flexible: true},
		},
		FlexibleColumns: true,
		AutoWidth:       true,
	}

	return tui.RenderTable(tableOpts, tableData)
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
