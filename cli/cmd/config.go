package cmd

import (
	"github.com/alecthomas/kong"
)

// ConfigCmd represents the config command with subcommands
type ConfigCmd struct {
	Validate ValidateCmd `cmd:"" help:"Validate the tag compliance configuration file"`
	Generate GenerateCmd `cmd:"" help:"Generate a sample configuration file"`
}

// BeforeApply is a Kong hook to perform any pre-processing before the command is run
func (c *ConfigCmd) BeforeApply(kongCtx *kong.Context) error {
	return nil
}
