package cmd

import (
	"fmt"

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
