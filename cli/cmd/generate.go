package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Excoriate/aws-taggy/cli/internal/output"
	"github.com/Excoriate/aws-taggy/pkg/configuration"
	"github.com/Excoriate/aws-taggy/pkg/o11y"
	"gopkg.in/yaml.v3"
)

// GenerateCmd represents the generate subcommand for creating configuration files
type GenerateCmd struct {
	Output    string `help:"Output format (file|clipboard)" default:"file" enum:"file,clipboard"`
	Directory string `help:"Directory to save the generated configuration file" default:"."`
	Filename  string `help:"Filename for the generated configuration file" default:"tag-compliance.yaml"`
	Overwrite bool   `help:"Overwrite existing file if it exists" default:"false"`
}

// Run method for GenerateCmd implements the configuration file generation logic
func (g *GenerateCmd) Run() error {
	logger := o11y.DefaultLogger()
	logger.Info("🚀 Generating tag compliance configuration file")

	// Create a sample configuration based on the tag-compliance.yaml template
	sampleConfig := configuration.TaggyScanConfig{
		Version: "v1alpha1",
		Global: configuration.GlobalConfig{
			Enabled:   true,
			BatchSize: intPtr(20),
			TagCriteria: configuration.TagCriteria{
				MinimumRequiredTags: 2,
				RequiredTags:        []string{"Environment", "Project"},
				ComplianceLevel:     "standard",
			},
		},
		AWS: configuration.AWSConfig{
			Regions: configuration.RegionsConfig{
				Mode: "all",
			},
			BatchSize: intPtr(20),
		},
		ComplianceLevels: map[string]configuration.ComplianceLevel{
			"standard": {
				RequiredTags: []string{"Environment", "Project"},
				SpecificTags: map[string]string{
					"Environment": "dev",
					"Project":     "aws-taggy",
				},
			},
		},
		Resources: map[string]configuration.ResourceConfig{
			"ec2": {
				Enabled: true,
				TagCriteria: configuration.TagCriteria{
					ComplianceLevel: "standard",
				},
			},
			"s3": {
				Enabled: true,
				TagCriteria: configuration.TagCriteria{
					ComplianceLevel: "standard",
				},
			},
			"vpc": {
				Enabled: true,
				TagCriteria: configuration.TagCriteria{
					ComplianceLevel: "standard",
				},
			},
		},
		TagValidation: configuration.TagValidation{
			AllowedValues: map[string][]string{
				"Environment": {"dev", "staging", "prod"},
				"Project":     {"taggy", "aws-taggy"},
			},
			CaseRules: map[string]configuration.CaseRule{
				"Environment": {
					Case:    configuration.CaseLowercase,
					Message: "Environment tag must be lowercase",
				},
				"Project": {
					Case:    configuration.CaseLowercase,
					Message: "Project tag must be lowercase",
				},
			},
		},
		Notifications: configuration.NotificationConfig{
			Slack: configuration.SlackNotificationConfig{
				Enabled: false,
			},
			Email: configuration.EmailNotificationConfig{
				Enabled: false,
			},
			Frequency: "daily",
		},
	}

	// Handle output based on the selected method
	switch g.Output {
	case "clipboard":
		return output.CopyToClipboard(sampleConfig)
	case "file":
		return generateConfigFile(g, sampleConfig)
	default:
		return fmt.Errorf("unsupported output format: %s", g.Output)
	}
}

func generateConfigFile(g *GenerateCmd, config configuration.TaggyScanConfig) error {
	// Resolve absolute path for the output directory
	absDir, err := filepath.Abs(g.Directory)
	if err != nil {
		return fmt.Errorf("failed to resolve absolute path: %w", err)
	}

	// Create the full file path
	filePath := filepath.Join(absDir, g.Filename)

	// Check if file exists and handle overwrite
	if _, err := os.Stat(filePath); err == nil && !g.Overwrite {
		return fmt.Errorf("file %s already exists. Use --overwrite to replace", filePath)
	}

	// Create the file
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	// Create YAML encoder with comments
	encoder := yaml.NewEncoder(file)
	encoder.SetIndent(2)

	// Write configuration with comments
	if err := encoder.Encode(config); err != nil {
		return fmt.Errorf("failed to encode configuration: %w", err)
	}

	fmt.Printf("✅ Configuration file generated successfully at: %s\n", filePath)
	return nil
}

// Helper function to create an integer pointer
func intPtr(i int) *int {
	return &i
}
