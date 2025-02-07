package cmd

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/Excoriate/aws-taggy/cli/internal/output"
	"github.com/Excoriate/aws-taggy/pkg/o11y"
	"github.com/alecthomas/kong"
)

// ConfigCmd represents the config command with subcommands
type ConfigCmd struct {
	Validate ValidateCmd `cmd:"" help:"Validate the tag compliance configuration file"`
	Generate GenerateCmd `cmd:"" help:"Generate a sample configuration file"`
}

// BeforeApply is a Kong hook to perform any pre-processing before the command is run
func (c *ConfigCmd) BeforeApply(kongCtx *kong.Context) error {
	// Currently a no-op method, but we'll add some basic error handling and logging
	logger := o11y.DefaultLogger()
	logger.Info("Preparing to execute configuration command")

	// Perform any necessary pre-processing or validation
	if kongCtx == nil {
		return fmt.Errorf("invalid Kong context: context is nil")
	}

	return nil
}

// GenerateCmd represents the command to generate a sample configuration file
type GenerateCmd struct {
	Output       string `short:"o" help:"Output file path for the generated configuration" default:"aws-taggy-config.yaml"`
	Overwrite    bool   `short:"f" help:"Force overwrite if the configuration file already exists"`
	GenerateDocs bool   `short:"d" help:"Generate markdown documentation alongside the configuration file"`
}

// Run implements the logic for generating a sample configuration file
func (c *GenerateCmd) Run() error {
	// Ensure the file has a .yaml or .yml extension
	outputFile := c.Output
	ext := filepath.Ext(outputFile)
	if ext == "" {
		outputFile += ".yaml"
	} else if ext != ".yaml" && ext != ".yml" {
		outputFile = strings.TrimSuffix(outputFile, ext) + ".yaml"
	}

	configWriter := output.NewConfigurationWriter()
	if err := configWriter.WriteConfiguration(outputFile, c.Overwrite); err != nil {
		return err
	}

	// Generate documentation if requested
	if c.GenerateDocs {
		docWriter := output.NewDocumentationWriter()
		if err := docWriter.WriteDocumentation(outputFile); err != nil {
			return fmt.Errorf("failed to generate documentation: %w", err)
		}
	}

	return nil
}
