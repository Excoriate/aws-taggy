package inspector

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Excoriate/aws-taggy/pkg/configuration"
	"github.com/Excoriate/aws-taggy/pkg/constants"
	"github.com/Excoriate/aws-taggy/pkg/o11y"
)

// ScanResult represents the outcome of a resource scanning operation
type ScanResult struct {
	Resources      []ResourceMetadata `json:"resources"`
	StartTime      time.Time          `json:"start_time"`
	EndTime        time.Time          `json:"end_time"`
	Duration       time.Duration      `json:"duration"`
	Region         string             `json:"region"`
	TotalResources int                `json:"total_resources"`
	Errors         []string           `json:"errors,omitempty"`
}

// Inspector defines the interface for cloud resource inspection operations
type Inspector interface {
	// Scan performs a discovery operation for resources of a specific type
	Scan(ctx context.Context, config configuration.TaggyScanConfig) (*ScanResult, error)

	// Fetch retrieves detailed information about a specific resource
	Fetch(ctx context.Context, arn string, config configuration.TaggyScanConfig) (*ResourceMetadata, error)
}

// New creates a new inspector for a specific resource type
func New(resourceType string, cfg configuration.TaggyScanConfig) (Inspector, error) {
	// Validate regions
	regions := cfg.AWS.Regions.List
	if len(regions) == 0 {
		return nil, fmt.Errorf("no regions specified for resource type %s", resourceType)
	}

	// Normalize resource type
	resourceType = configuration.NormalizeResourceType(resourceType)

	// Create inspector based on resource type
	switch resourceType {
	case constants.ResourceTypeS3:
		return NewS3Scanner(regions)
	case constants.ResourceTypeEC2:
		return NewEC2Scanner(regions)
	case constants.ResourceTypeVPC:
		return NewVPCScanner(regions)
	default:
		return nil, fmt.Errorf("unsupported resource type: %s", resourceType)
	}
}

// ScannerManager manages scanning operations across multiple resource types
type ScannerManager struct {
	scanners map[string]Inspector
	config   configuration.TaggyScanConfig
	results  map[string]*ScanResult
	logger   *o11y.Logger
	errors   []string
}

// NewScannerManager creates a new scanner manager based on the configuration
func NewScannerManager(config configuration.TaggyScanConfig) (*ScannerManager, error) {
	logger := o11y.DefaultLogger()
	scanners := make(map[string]Inspector)
	results := make(map[string]*ScanResult)
	errors := []string{}

	// Iterate through configured resources and create scanners
	for resourceType, resourceConfig := range config.Resources {
		// Skip disabled resources
		if !resourceConfig.Enabled {
			logger.Info(fmt.Sprintf("Resource type %s is disabled, skipping", resourceType))
			continue
		}

		// Validate resource type
		if err := configuration.IsSupportedAWSResource(resourceType); err != nil {
			errorMsg := fmt.Sprintf("Resource type %s validation failed: %v", resourceType, err)
			logger.Error(errorMsg)
			errors = append(errors, errorMsg)
			continue
		}

		// Create scanner
		scanner, err := New(resourceType, config)
		if err != nil {
			errorMsg := fmt.Sprintf("Failed to create scanner for %s: %v", resourceType, err)
			logger.Error(errorMsg)
			errors = append(errors, errorMsg)
			continue
		}

		scanners[resourceType] = scanner
	}

	return &ScannerManager{
		scanners: scanners,
		config:   config,
		results:  results,
		logger:   logger,
		errors:   errors,
	}, nil
}

// Scan performs scanning for all configured resource types
func (sm *ScannerManager) Scan(ctx context.Context) error {
	var wg sync.WaitGroup
	var mu sync.Mutex
	errChan := make(chan error, len(sm.scanners))
	sm.errors = []string{} // Reset errors slice

	for resourceType, scanner := range sm.scanners {
		wg.Add(1)
		go func(rt string, s Inspector) {
			defer wg.Done()

			sm.logger.Info(fmt.Sprintf("Scanning resource type: %s", rt))

			result, err := s.Scan(ctx, sm.config)
			if err != nil {
				errorMsg := fmt.Sprintf("Scanning %s failed: %v", rt, err)
				sm.logger.Error(errorMsg)

				mu.Lock()
				sm.errors = append(sm.errors, errorMsg)
				errChan <- fmt.Errorf(errorMsg)
				mu.Unlock()
				return
			}

			// Store results by region for consistent access
			mu.Lock()
			sm.results[result.Region] = result
			mu.Unlock()
		}(resourceType, scanner)
	}

	wg.Wait()
	close(errChan)

	// Collect and return any errors
	var errs []error
	for err := range errChan {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		return fmt.Errorf("scanning encountered %d errors: %v", len(errs), errs)
	}

	return nil
}

// GetResults returns the scanning results for all resource types
func (sm *ScannerManager) GetResults() map[string]*ScanResult {
	return sm.results
}

// GetErrors returns the list of error messages encountered during scanning
func (sm *ScannerManager) GetErrors() []string {
	return sm.errors
}
