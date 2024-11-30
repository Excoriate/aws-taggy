package inspector

import "github.com/Excoriate/aws-taggy/pkg/o11y"

// InspectorConfig holds configuration for the scanning process
// InspectorConfig represents the comprehensive configuration settings for the inspection process.
// It provides fine-grained control over how resources are scanned, processed, and logged.
//
// The configuration allows customization of:
// - Logging: A custom logger for capturing inspection-related events and diagnostics
// - Concurrency: Number of workers to parallelize the scanning process
// - Batch Processing: Size of batches for efficient resource scanning
type InspectorConfig struct {
	// Logger is a pointer to a custom logger from the o11y package,
	// used for capturing detailed logs during the inspection process.
	Logger *o11y.Logger

	// NumWorkers defines the number of concurrent workers used for parallel scanning.
	// This allows for improved performance by distributing the scanning workload.
	NumWorkers int

	// BatchSize determines the number of resources processed in a single batch.
	// Helps in managing memory and processing efficiency during large-scale inspections.
	BatchSize int
}

// DefaultInspectorConfig returns a default scan configuration
// DefaultInspectorConfig provides a pre-configured default configuration for the inspector.
//
// This function returns an InspectorConfig with sensible default settings that are suitable
// for most general-purpose resource scanning scenarios. The defaults are designed to balance
// performance and resource utilization:
//   - Logger: Uses the default logger from the o11y package for standard logging
//   - NumWorkers: Sets 10 concurrent workers to enable parallel processing
//   - BatchSize: Configures batch processing of 100 resources per batch
//
// The default configuration can be easily modified after creation to suit specific
// inspection requirements. It serves as a convenient starting point for most use cases.
//
// Returns:
//   - InspectorConfig: A fully initialized configuration with default settings
func DefaultInspectorConfig() InspectorConfig {
	return InspectorConfig{
		Logger:     o11y.DefaultLogger(),
		NumWorkers: 10,
		BatchSize:  100,
	}
}
