package scanner

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/Excoriate/aws-taggy/pkg/configuration"
	"github.com/Excoriate/aws-taggy/pkg/constants"
	_ "github.com/aws/aws-sdk-go-v2/config"
)

// ScanResult represents the comprehensive outcome of a resource scanning operation.
// It encapsulates detailed information about the resources discovered, scanning process,
// and any potential errors encountered during the scan.
type ScanResult struct {
	// TotalResources indicates the total number of resources discovered during the scanning process.
	// This provides a quick overview of the scan's scope and coverage.
	TotalResources int `json:"total_resources"`

	// Resources is a slice containing detailed metadata for each discovered resource.
	// Each ResourceMetadata provides specific information about an individual resource.
	Resources []ResourceMetadata `json:"resources"`

	// StartTime records the precise moment when the scanning process began.
	// This timestamp helps in tracking the exact timing of the resource discovery.
	StartTime time.Time `json:"start_time"`

	// EndTime captures the exact moment when the scanning process completed.
	// When combined with StartTime, it allows for precise duration calculations.
	EndTime time.Time `json:"end_time"`

	// Duration represents the total wall-clock time taken for the entire scanning operation.
	// This provides a quick way to understand the overall time efficiency of the scan.
	Duration time.Duration `json:"duration"`

	// TotalScanDuration aggregates the time spent scanning individual resources.
	// This can differ from Duration if scanning occurs concurrently or involves multiple steps.
	TotalScanDuration time.Duration `json:"total_scan_duration"`

	// Region specifies the cloud region where the resources were scanned.
	// This is crucial for understanding the geographical context of the discovered resources.
	Region string `json:"region"`

	// Errors captures any issues encountered during the scanning process.
	// This slice allows for comprehensive error tracking without interrupting the entire scan.
	Errors []string `json:"errors,omitempty"`

	// Scan metadata
	ScanMetadata struct {
		ServiceType     string                 `json:"service_type"`
		APICallsMade    int                    `json:"api_calls_made"`
		RateLimit       int                    `json:"rate_limit"`
		FiltersCriteria map[string]interface{} `json:"filters_criteria,omitempty"`
	} `json:"scan_metadata"`
}

// Scanner defines the interface for cloud resource discovery
type Scanner interface {
	// Scan discovers resources based on the provided configuration
	//
	// Parameters:
	//   - ctx: Context for cancellation and timeout
	//   - resource: The specific resource type to scan
	//   - config: The overall configuration for scanning
	//
	// Returns:
	//   - ScanResult containing discovered resources
	//   - Any error encountered during scanning
	Scan(ctx context.Context, resource Resource, config configuration.TaggyScanConfig) (*ScanResult, error)
}

// NewScanner creates a new scanner for a specific resource type
//
// Parameters:
//   - resourceType: The type of resource to scan (e.g., "s3", "ec2")
//   - config: The overall configuration for scanning
//
// Returns:
//   - A Scanner instance for the specified resource type
//   - An error if the scanner cannot be created
func NewScanner(resourceType string, config configuration.TaggyScanConfig) (Scanner, error) {
	if err := configuration.IsSupportedAWSResource(resourceType); err != nil {
		return nil, fmt.Errorf("failed to initialize scanner for resource type %s: %w", resourceType, err)
	}

	regions := determineRegionToScan(resourceType, config)

	// Create a scanner based on the resource type
	switch resourceType {
	case constants.ResourceTypeS3:
		return NewS3Scanner(regions)
	case constants.ResourceTypeEC2:
		return NewEC2Scanner(regions)
	// Add more resource type scanners as needed
	default:
		return nil, fmt.Errorf("scanner not implemented for resource type: %s", resourceType)
	}
}

// determineRegionToScan intelligently selects AWS regions to scan based on the provided configuration.
//
// The function implements a hierarchical region selection strategy:
//  1. First, it checks for resource-type specific regions
//  2. If no resource-specific regions are defined, it checks the global scanning mode
//  3. If mode is "all", it returns all valid AWS regions
//  4. Otherwise, it uses explicitly configured regions or defaults to the default AWS region
//
// Parameters:
//   - resourceType: The type of AWS resource being scanned (e.g., "s3", "ec2")
//   - config: The comprehensive scan configuration containing region selection rules
//
// Returns:
//
//	A slice of AWS region strings to be scanned
func determineRegionToScan(resourceType string, config configuration.TaggyScanConfig) []string {
	var regions []string

	// Check for resource-specific region configuration
	resourceConfig, exists := config.Resources[resourceType]
	if exists && len(resourceConfig.Regions) > 0 {
		// Use regions explicitly defined for this resource type
		regions = resourceConfig.Regions
	} else if config.AWS.Regions.Mode == "all" {
		// Scan all available AWS regions when global mode is set to "all"
		regions = configuration.ValidAWSRegions()
	} else {
		// Fallback to configured or default regions
		if len(config.AWS.Regions.List) > 0 {
			// Use explicitly configured region list
			regions = config.AWS.Regions.List
		} else {
			// Default to the standard AWS default region if no regions are specified
			regions = []string{constants.DefaultAWSRegion}
		}
	}

	return regions
}

// ScannerManager manages the scanning process for multiple AWS resource types.
// It provides thread-safe mechanisms for initializing, tracking, and executing resource scans.
type ScannerManager struct {
	// scanners is a thread-safe map of resource type to its corresponding Scanner implementation.
	// Each scanner is responsible for scanning a specific type of AWS resource.
	scanners map[string]Scanner

	// config holds the comprehensive configuration for the scanning process,
	// including resource-specific settings, region selection, and other scanning parameters.
	config configuration.TaggyScanConfig

	// mu is a read-write mutex that protects concurrent access to the scanners and results,
	// ensuring thread-safe operations during parallel scanning.
	mu sync.RWMutex

	// results stores the scan results for each resource type, keyed by the resource type.
	// This allows aggregation and retrieval of scan results after the scanning process.
	results map[string]*ScanResult

	// errors captures any errors encountered during the scanning process.
	// This provides a comprehensive error tracking mechanism for post-scan analysis.
	errors []error
}

// NewScannerManager creates a new ScannerManager
// NewScannerManager creates a new ScannerManager instance for managing AWS resource scanning.
//
// This function initializes a ScannerManager with the following key behaviors:
// - Creates an empty map of scanners for tracking resource-specific scanning implementations
// - Configures the manager with the provided scanning configuration
// - Dynamically initializes scanners for each enabled resource type
//
// Parameters:
//   - config: A comprehensive configuration defining scanning parameters,
//     resource types, and scanning behavior
//
// Returns:
//   - A fully initialized *ScannerManager ready for scanning
//   - An error if any scanner initialization fails, with detailed error context
//
// Best practices:
// - Validates and initializes scanners before returning the manager
// - Provides granular error handling for individual scanner creation
// - Supports dynamic scanner configuration based on input
func NewScannerManager(config configuration.TaggyScanConfig) (*ScannerManager, error) {
	// Initialize the ScannerManager with empty collections and provided configuration
	manager := &ScannerManager{
		scanners: make(map[string]Scanner),
		config:   config,
		results:  make(map[string]*ScanResult),
	}

	// Iterate through configured resources and initialize enabled scanners
	for resourceType, resourceConfig := range config.Resources {
		if resourceConfig.Enabled {
			// Attempt to create a scanner for each enabled resource type
			scanner, err := NewScanner(resourceType, config)
			if err != nil {
				// Wrap and return any scanner initialization errors
				return nil, fmt.Errorf("failed to create scanner for %s: %w", resourceType, err)
			}
			// Store the successfully created scanner
			manager.scanners[resourceType] = scanner
		}
	}

	return manager, nil
}

// Scan performs concurrent scanning of all initialized scanners
// Scan performs a comprehensive, concurrent scanning of all initialized resource scanners.
//
// This method orchestrates a parallel scanning process across multiple resource types,
// providing robust error handling and performance tracking. Key features include:
//
// - Concurrent scanning of multiple resource types using goroutines
// - Thread-safe result and error collection
// - Detailed error reporting
// - Total scan duration tracking
//
// Best Practices:
// - Uses sync.WaitGroup for coordinating concurrent goroutines
// - Implements buffered channels to prevent goroutine leaks
// - Provides granular error tracking per resource type
// - Ensures thread-safe access to shared resources
//
// Parameters:
//   - ctx: A context.Context for controlling scan lifecycle, cancellation, and timeouts
//
// Returns:
//   - An error if any resource scanning fails, aggregating all encountered errors
//   - nil if all scans complete successfully
//
// Usage Example:
//
//	err := scannerManager.Scan(context.Background())
//	if err != nil {
//	    // Handle scanning errors
//	}
func (sm *ScannerManager) Scan(ctx context.Context) error {
	startTime := time.Now()

	// Create a wait group for concurrent scanning
	var wg sync.WaitGroup
	resultChan := make(chan *ScanResult, len(sm.scanners))
	errorChan := make(chan error, len(sm.scanners))

	// Scan each resource type concurrently
	for resourceType, scanner := range sm.scanners {
		wg.Add(1)
		go func(rt string, s Scanner) {
			defer wg.Done()

			// Create a resource instance
			resource := NewResourceType(rt)

			// Perform the scan
			result, err := s.Scan(ctx, resource, sm.config)
			if err != nil {
				// Log the specific error for each resource type
				errorChan <- fmt.Errorf("scan failed for %s: %w", rt, err)
				return
			}

			// Send result to channel
			resultChan <- result
		}(resourceType, scanner)
	}

	// Close channels when all goroutines are done
	go func() {
		wg.Wait()
		close(resultChan)
		close(errorChan)
	}()

	// Collect results and errors
	for result := range resultChan {
		sm.mu.Lock()
		sm.results[result.Region] = result
		sm.mu.Unlock()
	}

	// Collect any errors
	var scanErrors []error
	for err := range errorChan {
		scanErrors = append(scanErrors, err)
		sm.errors = append(sm.errors, err)
	}

	// Add total scan duration to the results
	totalDuration := time.Since(startTime)
	sm.mu.Lock()
	for _, result := range sm.results {
		result.TotalScanDuration = totalDuration
	}
	sm.mu.Unlock()

	// Return detailed error information
	if len(scanErrors) > 0 {
		var errorMessages []string
		for _, err := range scanErrors {
			errorMessages = append(errorMessages, err.Error())
		}
		return fmt.Errorf("scanning failed with %d errors:\n%s",
			len(scanErrors),
			strings.Join(errorMessages, "\n"))
	}

	return nil
}

// GetResults returns the scan results
func (sm *ScannerManager) GetResults() map[string]*ScanResult {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.results
}

// GetErrors returns any errors encountered during scanning
func (sm *ScannerManager) GetErrors() []error {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.errors
}
