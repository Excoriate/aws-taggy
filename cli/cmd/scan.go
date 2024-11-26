package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/Excoriate/aws-taggy/cli/internal/tui"
	"github.com/Excoriate/aws-taggy/pkg/scanner"
	"github.com/Excoriate/aws-taggy/pkg/taggy"
)

// ScanCmd represents the scan subcommand
type ScanCmd struct {
	Config string `help:"Path to the tag compliance configuration file" required:"true"`
}

// Run method for ScanCmd implements the scanning logic
func (s *ScanCmd) Run() error {
	// Initialize Taggy client with the configuration file
	client, err := taggy.New(s.Config)
	if err != nil {
		return fmt.Errorf("failed to create Taggy client: %w", err)
	}

	// Perform full resource scanning
	return s.startScan(client)
}

// startScan performs a comprehensive scan across all supported AWS resources
func (s *ScanCmd) startScan(client *taggy.TaggyClient) error {
	ctx := context.Background()
	startTime := time.Now()

	fmt.Println("🔍 Performing full AWS resource tag compliance scan...")

	// Create a scanner manager
	scannerManager, err := scanner.NewScannerManager(*client.Config())
	if err != nil {
		return fmt.Errorf("failed to create scanner manager: %w", err)
	}

	// Perform the scan
	if err := scannerManager.Scan(ctx); err != nil {
		return fmt.Errorf("scanning encountered errors: %v", err)
	}

	// Process scan results
	results := scannerManager.GetResults()

	// Prepare table data
	type ComplianceRow struct {
		Region            string
		TotalResources    int
		CompliantCount    int
		NonCompliantCount int
		CompliancePercent float64
	}

	var totalResources, totalCompliant int
	var complianceRows []ComplianceRow
	for region, result := range results {
		compliantResources := 0
		for _, resource := range result.Resources {
			if len(resource.Tags) > 0 {
				compliantResources++
			}
		}

		totalResources += result.TotalResources
		totalCompliant += compliantResources

		compliancePercentage := 0.0
		if result.TotalResources > 0 {
			compliancePercentage = (float64(compliantResources) / float64(result.TotalResources)) * 100
		}

		complianceRows = append(complianceRows, ComplianceRow{
			Region:            region,
			TotalResources:    result.TotalResources,
			CompliantCount:    compliantResources,
			NonCompliantCount: result.TotalResources - compliantResources,
			CompliancePercent: compliancePercentage,
		})
	}

	// Calculate overall compliance
	overallCompliancePercentage := 0.0
	if totalResources > 0 {
		overallCompliancePercentage = (float64(totalCompliant) / float64(totalResources)) * 100
	}

	// Create and render TUI table
	tableOpts := tui.TableOptions{
		Title: fmt.Sprintf("🏷️  AWS Tag Compliance Scan Results (Overall: %.2f%%)", overallCompliancePercentage),
		Columns: []tui.Column{
			{Title: "Region", Key: "Region", Width: 15, Flexible: true},
			{Title: "Total Resources", Key: "TotalResources", Width: 15},
			{Title: "Compliant", Key: "CompliantCount", Width: 12},
			{Title: "Non-Compliant", Key: "NonCompliantCount", Width: 15},
			{Title: "Compliance %", Key: "CompliancePercent", Width: 12},
		},
		FlexibleColumns: true,
	}

	tableModel := tui.NewTableModel(tableOpts, complianceRows)

	// Render table to console
	if err := tableModel.Render(); err != nil {
		return fmt.Errorf("failed to render results table: %w", err)
	}

	// Check for any errors
	if errors := scannerManager.GetErrors(); len(errors) > 0 {
		fmt.Println("\n🚨 Scanning completed with the following errors:")
		for _, err := range errors {
			fmt.Println(err)
		}
	}

	duration := time.Since(startTime)
	fmt.Printf("\n⏱️  Total scan duration: %v\n", duration)

	return nil
}
