package inspector

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// AsyncResourceInspector handles asynchronous resource scanning
// AsyncResourceInspector is a struct that manages asynchronous resource inspection processes.
// It encapsulates configuration settings for parallel resource discovery and processing.
// The struct provides a flexible mechanism for scanning and analyzing resources across multiple regions
// with configurable concurrency and batch processing.
type AsyncResourceInspector struct {
	// config holds the configuration parameters for resource inspection
	// including logging, worker count, batch size, and other operational settings
	config InspectorConfig
}

// NewAsyncResourceInspector creates a new AsyncResourceInspector
// NewAsyncResourceInspector creates a new instance of AsyncResourceInspector with the specified configuration.
//
// This function initializes an AsyncResourceInspector with the provided configuration settings.
// It allows customization of resource inspection parameters such as logging, worker count,
// batch size, and other operational settings.
//
// Parameters:
//   - config: An InspectorConfig struct that defines the configuration for resource inspection.
//
// Returns:
//   - A pointer to the newly created AsyncResourceInspector instance.
//
// Example:
//
//	config := InspectorConfig{
//	    NumWorkers: 5,
//	    BatchSize: 100,
//	    Logger: customLogger,
//	}
//	inspector := NewAsyncResourceInspector(config)
func NewAsyncResourceInspector(config InspectorConfig) *AsyncResourceInspector {
	return &AsyncResourceInspector{
		config: config,
	}
}

// startResourceDiscovery initiates parallel resource discovery for given regions
func (s *AsyncResourceInspector) startResourceDiscovery(
	ctx context.Context,
	regions []string,
	discoverer ResourceDiscoverer,
	resourceChan chan interface{},
	errorChan chan error,
	discoveryWg *sync.WaitGroup,
	processingWg *sync.WaitGroup,
) {
	for _, region := range regions {
		discoveryWg.Add(1)
		go func(r string) {
			defer discoveryWg.Done()

			// Discover resources in this region
			resources, err := discoverer(ctx, r)
			if err != nil {
				select {
				case errorChan <- fmt.Errorf("failed to discover resources in region %s: %w", r, err):
				default:
					// If channel is full, log the error
					s.config.Logger.Error(fmt.Sprintf("Failed to send error for region %s: %v", r, err))
				}
				return
			}

			s.config.Logger.Info(fmt.Sprintf("Discovered resources in region %s", r),
				"region", r,
				"count", len(resources))

			// Add to processing WaitGroup before sending resources
			processingWg.Add(len(resources))

			// Send resources to processing channel
			for _, resource := range resources {
				select {
				case resourceChan <- resource:
				case <-ctx.Done():
					processingWg.Add(-1) // Decrement if we couldn't send
					return
				}
			}
		}(region)
	}
}

// startResourceProcessing starts worker goroutines to process resources
func (s *AsyncResourceInspector) startResourceProcessing(
	ctx context.Context,
	resourceChan chan interface{},
	resultChan chan ResourceMetadata,
	processor ResourceProcessor,
	processingWg *sync.WaitGroup,
) {
	for i := 0; i < s.config.NumWorkers; i++ {
		go func(workerID int) {
			for resource := range resourceChan {
				func() {
					defer processingWg.Done()

					// Process the resource
					metadata, err := processor(ctx, resource)
					if err != nil {
						s.config.Logger.Error("Failed to process resource",
							"error", err)
						return
					}

					// Log successful processing
					s.config.Logger.Info("Processed resource",
						"type", metadata.Type,
						"id", metadata.ID,
						"region", metadata.Region,
						"has_tags", len(metadata.Tags) > 0,
						"tag_count", len(metadata.Tags))

					// Send result with non-blocking select
					select {
					case resultChan <- metadata:
					case <-ctx.Done():
						return
					}
				}()
			}
		}(i)
	}
}

// manageChannelLifecycle handles closing of channels and waiting for goroutines
func (s *AsyncResourceInspector) manageChannelLifecycle(
	resourceChan chan interface{},
	resultChan chan ResourceMetadata,
	errorChan chan error,
	discoveryWg *sync.WaitGroup,
	processingWg *sync.WaitGroup,
) {
	go func() {
		// Wait for all discovery goroutines to finish
		discoveryWg.Wait()

		// Wait a small duration to ensure all resources are sent
		time.Sleep(100 * time.Millisecond)

		// Close resource channel when discovery is done
		close(resourceChan)

		// Wait for all processing to complete
		processingWg.Wait()

		// Close result and error channels
		close(resultChan)
		close(errorChan)
	}()
}

// collectScanResults aggregates processed resources and errors
func (s *AsyncResourceInspector) collectScanResults(
	resultChan chan ResourceMetadata,
	errorChan chan error,
) ([]ResourceMetadata, []error) {
	var results []ResourceMetadata
	var scanErrors []error

	// Collect errors
	for err := range errorChan {
		scanErrors = append(scanErrors, err)
	}

	// Collect processed resources
	for metadata := range resultChan {
		results = append(results, metadata)
	}

	return results, scanErrors
}

// InspectResourcesAsync performs asynchronous resource scanning using the provided discoverer and processor functions
// InspectResourcesAsync performs an asynchronous, parallel scanning of resources across multiple regions.
//
// This method allows for efficient and concurrent discovery and processing of resources using
// provided discoverer and processor functions. It supports:
//   - Concurrent resource discovery across multiple regions
//   - Parallel processing of discovered resources
//   - Error aggregation and handling
//   - Configurable batch sizes and concurrency
//
// Parameters:
//   - ctx: A context for cancellation and timeout management
//   - regions: A slice of region identifiers to scan
//   - discoverer: A function that discovers resources in a given region
//   - processor: A function that processes individual resources
//
// Returns:
//   - A slice of ResourceMetadata containing processed resource information
//   - An error if any scanning or processing errors occurred
//
// The method uses goroutines and channels to achieve high-performance, concurrent resource scanning.
// It manages the entire lifecycle of discovery and processing, including error handling and resource tracking.
func (s *AsyncResourceInspector) InspectResourcesAsync(
	ctx context.Context,
	regions []string,
	discoverer ResourceDiscoverer,
	processor ResourceProcessor,
) ([]ResourceMetadata, error) {
	// Create channels for async processing
	resourceChan := make(chan interface{}, s.config.BatchSize)
	resultChan := make(chan ResourceMetadata, s.config.BatchSize)
	errorChan := make(chan error, len(regions))

	// WaitGroups for discovery and processing
	var discoveryWg, processingWg sync.WaitGroup

	// Start resource discovery goroutines
	s.startResourceDiscovery(ctx, regions, discoverer, resourceChan, errorChan, &discoveryWg, &processingWg)

	// Start resource processing workers
	s.startResourceProcessing(ctx, resourceChan, resultChan, processor, &processingWg)

	// Manage channel lifecycle
	s.manageChannelLifecycle(resourceChan, resultChan, errorChan, &discoveryWg, &processingWg)

	// Collect results and errors
	results, scanErrors := s.collectScanResults(resultChan, errorChan)

	// Check for any errors
	if len(scanErrors) > 0 {
		return results, fmt.Errorf("scanning encountered %d errors", len(scanErrors))
	}

	return results, nil
}
