package cmd

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/Excoriate/aws-taggy/cli/internal/tui"
	"github.com/Excoriate/aws-taggy/pkg/configuration"
	"github.com/Excoriate/aws-taggy/pkg/inspector"
	"github.com/Excoriate/aws-taggy/pkg/o11y"
	"github.com/Excoriate/aws-taggy/pkg/output"
)

// QueryCmd represents the query command and its subcommands
type QueryCmd struct {
	Tags TagsCmd `cmd:"" help:"Query tags for a specific AWS resource"`
	Info InfoCmd `cmd:"" help:"Query detailed information about a specific AWS resource"`
}

// TagsCmd represents the query tags subcommand
type TagsCmd struct {
	ARN       string `help:"ARN of the resource to query tags for" required:"true"`
	Service   string `help:"AWS service type (e.g., s3, ec2)" required:"true"`
	Output    string `help:"Output format (table|json|yaml)" default:"table" enum:"table,json,yaml,TABLE,JSON,YAML"`
	Clipboard bool   `help:"Copy output to clipboard" default:"false"`
}

// InfoCmd represents the query info subcommand
type InfoCmd struct {
	ARN       string `help:"ARN of the resource to query information for" required:"true"`
	Service   string `help:"AWS service type (e.g., s3, ec2)" required:"true"`
	Output    string `help:"Output format (table|json|yaml)" default:"table" enum:"table,json,yaml,TABLE,JSON,YAML"`
	Clipboard bool   `help:"Copy output to clipboard" default:"false"`
}

// Run is a no-op method to satisfy the Kong command interface
func (q *QueryCmd) Run() error {
	return nil
}

// Run implements the tags query logic
func (t *TagsCmd) Run() error {
	logger := o11y.DefaultLogger()
	logger.Info(fmt.Sprintf("🔍 Querying tags for resource: %s", t.ARN))

	regionOnARN := inspector.ExtractRegionFromARNOrDefault(t.ARN)

	// Create minimal config for the specific service
	config := configuration.TaggyScanConfig{
		AWS: configuration.AWSConfig{
			Regions: configuration.RegionsConfig{
				Mode: "specific",
				List: []string{regionOnARN},
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
		return fmt.Errorf("failed to create inspector for service %s: %w", t.Service, err)
	}

	// Fetch resource details
	ctx := context.Background()
	resource, err := inspectorClient.Fetch(ctx, t.ARN, config)
	if err != nil {
		return fmt.Errorf("failed to fetch resource details for ARN %s in service %s: %w", t.ARN, t.Service, err)
	}

	// Prepare output
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

	// Normalize output format
	outputFormat := strings.ToLower(t.Output)

	// Create output formatter
	var formatter output.Formatter
	switch outputFormat {
	case "json":
		formatter = output.NewJSONFormatter(false)
	case "yaml", "yml":
		formatter = output.NewYAMLFormatter(false)
	default:
		formatter = output.NewTableFormatter([]string{"Key", "Value"})
	}

	// Prepare clipboard output
	clipboardOutput := result

	// If clipboard flag is set, copy to clipboard in YAML
	if t.Clipboard {
		yamlFormatter := output.NewYAMLFormatter(false)
		clipboardContent, err := yamlFormatter.Format(clipboardOutput)
		if err != nil {
			return fmt.Errorf("failed to format clipboard output: %w", err)
		}

		// Use system clipboard
		cmd := exec.Command("pbcopy")
		cmd.Stdin = strings.NewReader(clipboardContent)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to copy resource tags to clipboard for ARN %s: %w", t.ARN, err)
		}

		logger.Info("✅ Resource tags copied to clipboard!")
	}

	// Check if output should be structured
	if outputFormat == "json" || outputFormat == "yaml" || outputFormat == "yml" {
		formattedOutput, err := formatter.Format(result)
		if err != nil {
			return fmt.Errorf("failed to format output: %w", err)
		}
		fmt.Println(formattedOutput)
		return nil
	}

	// Default table output
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

// Run implements the info query logic
func (i *InfoCmd) Run() error {
	logger := o11y.DefaultLogger()
	logger.Info(fmt.Sprintf("🔍 Querying information for resource: %s", i.ARN))

	regionOnARN := inspector.ExtractRegionFromARNOrDefault(i.ARN)

	// Similar initialization as TagsCmd
	config := configuration.TaggyScanConfig{
		AWS: configuration.AWSConfig{
			Regions: configuration.RegionsConfig{
				Mode: "specific",
				List: []string{regionOnARN},
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
		return fmt.Errorf("failed to create inspector for service %s: %w", i.Service, err)
	}

	ctx := context.Background()
	resource, err := inspectorClient.Fetch(ctx, i.ARN, config)
	if err != nil {
		return fmt.Errorf("failed to fetch resource details for ARN %s in service %s: %w", i.ARN, i.Service, err)
	}

	// Normalize output format
	outputFormat := strings.ToLower(i.Output)

	// Prepare clipboard output
	clipboardOutput := struct {
		Service           string                 `json:"service" yaml:"service"`
		Region            string                 `json:"region" yaml:"region"`
		ResourceID        string                 `json:"resource_id" yaml:"resource_id"`
		ResourceType      string                 `json:"resource_type" yaml:"resource_type"`
		ARN               string                 `json:"arn" yaml:"arn"`
		TagCount          int                    `json:"tag_count" yaml:"tag_count"`
		Tags              map[string]string      `json:"tags" yaml:"tags"`
		AdditionalDetails map[string]interface{} `json:"additional_details,omitempty" yaml:"additional_details,omitempty"`
	}{
		Service:           i.Service,
		Region:            resource.Region,
		ResourceID:        resource.ID,
		ResourceType:      resource.Type,
		ARN:               resource.Details.ARN,
		TagCount:          len(resource.Tags),
		Tags:              resource.Tags,
		AdditionalDetails: resource.Details.Properties,
	}

	// Create output formatter
	var formatter output.Formatter
	switch outputFormat {
	case "json":
		formatter = output.NewJSONFormatter(false)
	case "yaml", "yml":
		formatter = output.NewYAMLFormatter(false)
	default:
		formatter = output.NewTableFormatter([]string{"Property", "Value"})
	}

	// If clipboard flag is set, copy to clipboard in YAML
	if i.Clipboard {
		yamlFormatter := output.NewYAMLFormatter(false)
		clipboardContent, err := yamlFormatter.Format(clipboardOutput)
		if err != nil {
			return fmt.Errorf("failed to format clipboard output: %w", err)
		}

		// Use system clipboard
		cmd := exec.Command("pbcopy")
		cmd.Stdin = strings.NewReader(clipboardContent)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to copy resource information to clipboard for ARN %s: %w", i.ARN, err)
		}

		logger.Info("✅ Resource information copied to clipboard!")
	}

	// Check if output should be structured
	if outputFormat == "json" || outputFormat == "yaml" || outputFormat == "yml" {
		formattedOutput, err := formatter.Format(clipboardOutput)
		if err != nil {
			return fmt.Errorf("failed to format output: %w", err)
		}
		fmt.Println(formattedOutput)
		return nil
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
