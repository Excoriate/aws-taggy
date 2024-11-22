package cmd

import (
	"fmt"
	"os"

	"github.com/Excoriate/aws-taggy/cli/internal/configuration"
	"github.com/Excoriate/aws-taggy/cli/internal/tui"
	"github.com/alecthomas/kong"
)

// Version will be populated during build
var version = "dev"

// RootCmd represents the base command structure for aws-taggy
type RootCmd struct {
	Version bool `short:"v" help:"Display version information"`
	Debug   bool `help:"Enable debug mode"`

	// Subcommands will be added here
	Scan    ScanCmd    `cmd:"" help:"Scan AWS resources for tag compliance"`
	Validate ValidateCmd `cmd:"" help:"Validate tag configurations"`
}

// ScanCmd represents the scan subcommand
type ScanCmd struct {
	All     bool   `help:"Scan all supported AWS resources"`
	Service string `help:"Specify a specific AWS service to scan"`
	Config  string `help:"Path to the tag compliance configuration file" required:"true"`
}

// ValidateCmd represents the validate subcommand
type ValidateCmd struct {
	Config string `help:"Path to the tag validation configuration file" required:"true"`
}

// Run implements the main logic for the root command
func (r *RootCmd) Run() error {
	if r.Version {
		fmt.Printf("aws-taggy version %s\n", version)
		return nil
	}

	// Default behavior if no subcommand is specified
	fmt.Println("No command specified. Use --help to see available commands.")
	fmt.Println("Recommended: use 'scan' or 'validate' commands.")
	return nil
}

// Run method for ScanCmd
func (s *ScanCmd) Run() error {
	if s.Config == "" {
		return fmt.Errorf("configuration file path is required for scanning")
	}

	if s.All {
		fmt.Printf("Scanning all AWS resources using configuration from: %s\n", s.Config)
		return nil
	}
	
	if s.Service != "" {
		fmt.Printf("Scanning AWS service: %s using configuration from: %s\n", s.Service, s.Config)
		return nil
	}

	return fmt.Errorf("no scan parameters specified. Use --help for more information")
}

// Run method for ValidateCmd
func (v *ValidateCmd) Run() error {
	if v.Config == "" {
		return fmt.Errorf("please specify a configuration file path")
	}

	fmt.Printf("Validating tag configuration from: %s\n", v.Config)
	return nil
}

// NewRootCommand creates and configures the root command
func NewRootCommand() *kong.Kong {
	cli := &RootCmd{}

	banner := tui.GetBanner()
	fmt.Println(banner)

	kongOptions := []kong.Option{
		kong.Name(configuration.AppName),
		kong.Description(configuration.AppDescription),
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
		parser.FatalIfErrorf(err)
	}

	return ctx.Run()
}
