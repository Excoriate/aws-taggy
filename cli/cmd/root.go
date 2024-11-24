package cmd

import (
	"fmt"
	"os"

	"github.com/Excoriate/aws-taggy/cli/internal/tui"
	"github.com/Excoriate/aws-taggy/pkg/constants"
	"github.com/alecthomas/kong"
)

// Version will be populated during build
var version = "dev"

// RootCmd represents the base command structure for aws-taggy
type RootCmd struct {
	Version bool `short:"v" help:"Display version information"`
	Debug   bool `help:"Enable debug mode"`

	// Subcommands
	Scan    ScanCmd    `cmd:"" help:"Scan AWS resources for tag compliance"`
	Validate ValidateCmd `cmd:"" help:"Validate tag configurations"`
}

// Run implements the main logic for the root command
func (r *RootCmd) Run() error {
	if r.Version {
		fmt.Printf("aws-taggy version %s\n", version)
		return nil
	}

	// Default behavior if no subcommand is specified
	fmt.Println("No command specified. Use --help to see available commands.")
	return nil
}

// NewRootCommand creates and configures the root command
func NewRootCommand() *kong.Kong {
	cli := &RootCmd{}

	banner := tui.GetBanner()
	fmt.Println(banner)

	kongOptions := []kong.Option{
		kong.Name(constants.AppName),
		kong.Description(constants.AppDescription),
		kong.UsageOnError(),
		kong.ConfigureHelp(kong.HelpOptions{
			Compact: true,
			Summary: true,
		}),
		kong.Vars{
			"version": version,
		},
	}

	parser, err := kong.New(cli, kongOptions...)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing CLI: %v\n", err)
		os.Exit(1)
	}

	return parser
}

// Execute runs the root command and handles parsing
func Execute() error {
	parser := NewRootCommand()

	ctx, err := parser.Parse(os.Args[1:])
	if err != nil {
		// If no arguments are provided, show help and exit successfully
		if len(os.Args) == 1 {
			parser.FatalIfErrorf(err)
		}

		// For other parsing errors, use Kong's default error handling
		parser.FatalIfErrorf(err)
	}

	return ctx.Run()
}
