package inspector

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/Excoriate/aws-taggy/pkg/configuration"
	"github.com/Excoriate/aws-taggy/pkg/o11y"
)

// InspectorManager manages scanning operations across multiple resource types
type InspectorManager struct {
	inspectors map[string]Inspector
	config     configuration.TaggyScanConfig
	results    map[string]*InspectResult
	logger     *o11y.Logger
	errors     []string
}

// NewInspectorManagerFromConfig creates a new inspector manager based on the configuration
func NewInspectorManagerFromConfig(config configuration.TaggyScanConfig) (*InspectorManager, error) {
	logger := o11y.DefaultLogger()
	inspectors := make(map[string]Inspector)
	results := make(map[string]*InspectResult)
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

		inspectors[resourceType] = scanner
	}

	return &InspectorManager{
		inspectors: inspectors,
		config:     config,
		results:    results,
		logger:     logger,
		errors:     errors,
	}, nil
}

// Inspect performs scanning for all configured resource types
func (sm *InspectorManager) Inspect(ctx context.Context) error {
	var wg sync.WaitGroup
	var mu sync.Mutex
	errChan := make(chan error, len(sm.inspectors))
	sm.errors = []string{} // Reset errors slice

	for resourceType, scanner := range sm.inspectors {
		wg.Add(1)
		go func(rt string, s Inspector) {
			defer wg.Done()

			sm.logger.Info(fmt.Sprintf("Scanning resource type: %s", rt))

			result, err := s.Inspect(ctx, sm.config)
			if err != nil {
				errorMsg := fmt.Sprintf("Scanning %s failed: %v", rt, err)
				sm.logger.Error(errorMsg)

				mu.Lock()
				sm.errors = append(sm.errors, errorMsg)
				errChan <- errors.New(errorMsg)
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
		return errors.Join(errs...)
	}

	return nil
}

// GetResults returns the scanning results for all resource types
func (sm *InspectorManager) GetResults() map[string]*InspectResult {
	return sm.results
}

// GetErrors returns the list of error messages encountered during scanning
func (sm *InspectorManager) GetErrors() []string {
	return sm.errors
}
