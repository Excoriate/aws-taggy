package inspector

import (
	"context"
	"errors"
	"fmt"
	"sync"
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
		select {
		case <-ctx.Done():
			s.config.Logger.Error("Context cancelled during resource discovery",
				"error", ctx.Err())
			return
		default:
			discoveryWg.Add(1)
			go func(r string) {
				defer discoveryWg.Done()

				resources, err := discoverer(ctx, r)
				if err != nil {
					s.config.Logger.Error("Failed to discover resources",
						"region", r,
						"error", err)
					select {
					case errorChan <- fmt.Errorf("failed to discover resources in region %s: %w", r, err):
					case <-ctx.Done():
						s.config.Logger.Error("Context cancelled while sending discovery error",
							"region", r,
							"error", err)
					}
					return
				}

				s.config.Logger.Info(fmt.Sprintf("Discovered resources in region %s", r),
					"region", r,
					"count", len(resources))

				processingWg.Add(len(resources))

				for _, resource := range resources {
					select {
					case resourceChan <- resource:
					case <-ctx.Done():
						s.config.Logger.Error("Context cancelled while sending resource",
							"region", r)
						processingWg.Add(-1)
						return
					}
				}
			}(region)
		}
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
	workerWg := &sync.WaitGroup{}

	for i := 0; i < s.config.NumWorkers; i++ {
		workerWg.Add(1)
		go func(workerID int) {
			defer workerWg.Done()
			for {
				select {
				case resource, ok := <-resourceChan:
					if !ok {
						return
					}
					func() {
						defer processingWg.Done()
						metadata, err := processor(ctx, resource)
						if err != nil {
							s.config.Logger.Error("Failed to process resource",
								"worker", workerID,
								"error", err)
							return
						}

						s.config.Logger.Info("Processed resource",
							"worker", workerID,
							"type", metadata.Type,
							"id", metadata.ID,
							"region", metadata.Region,
							"has_tags", len(metadata.Tags) > 0,
							"tag_count", len(metadata.Tags))

						select {
						case resultChan <- metadata:
						case <-ctx.Done():
							s.config.Logger.Error("Context cancelled while sending result",
								"worker", workerID,
								"resource_id", metadata.ID)
						}
					}()
				case <-ctx.Done():
					s.config.Logger.Error("Context cancelled for worker",
						"worker", workerID,
						"error", ctx.Err())
					return
				}
			}
		}(i)
	}

	// Wait for all workers to finish before closing result channel
	go func() {
		workerWg.Wait()
		close(resultChan)
	}()
}

// manageChannelLifecycle handles closing of channels and waiting for goroutines
func (s *AsyncResourceInspector) manageChannelLifecycle(
	ctx context.Context,
	resourceChan chan interface{},
	resultChan chan ResourceMetadata,
	errorChan chan error,
	discoveryWg *sync.WaitGroup,
	processingWg *sync.WaitGroup,
) {
	// Create a channel to coordinate closing
	done := make(chan struct{})
	closeOnce := sync.Once{}

	// Close channels in the correct order
	go func() {
		defer close(done)

		// First wait for discovery to complete
		discoveryWg.Wait()
		s.config.Logger.Info("Resource discovery completed")

		// Then close the resource channel
		closeOnce.Do(func() {
			close(resourceChan)
			s.config.Logger.Info("Resource channel closed")
		})

		// Wait for processing to complete
		processingWg.Wait()
		s.config.Logger.Info("Resource processing completed")

		// Finally close the error channel
		close(errorChan)
		s.config.Logger.Info("Error channel closed")
	}()

	// Wait for completion or context cancellation
	select {
	case <-done:
		s.config.Logger.Info("Channel lifecycle management completed normally")
		return
	case <-ctx.Done():
		s.config.Logger.Error("Context cancelled during channel lifecycle management",
			"error", ctx.Err())
		// Ensure channels are closed even on cancellation
		closeOnce.Do(func() {
			close(resourceChan)
			s.config.Logger.Info("Resource channel closed on cancellation")
		})
		return
	}
}

// collectScanResults aggregates processed resources and errors
func (s *AsyncResourceInspector) collectScanResults(
	ctx context.Context,
	resultChan chan ResourceMetadata,
	errorChan chan error,
) ([]ResourceMetadata, []error) {
	var results []ResourceMetadata
	var scanErrors []error
	var mu sync.Mutex
	var wg sync.WaitGroup

	wg.Add(2)

	// Collect results
	go func() {
		defer wg.Done()
		for {
			select {
			case result, ok := <-resultChan:
				if !ok {
					return
				}
				mu.Lock()
				results = append(results, result)
				mu.Unlock()
			case <-ctx.Done():
				return
			}
		}
	}()

	// Collect errors with improved error handling
	go func() {
		defer wg.Done()
		for {
			select {
			case err, ok := <-errorChan:
				if !ok {
					return
				}
				mu.Lock()
				if err != nil {
					s.config.Logger.Error("Error during resource scanning",
						"error", err)
					scanErrors = append(scanErrors, err)
				}
				mu.Unlock()
			case <-ctx.Done():
				if ctx.Err() != nil {
					mu.Lock()
					scanErrors = append(scanErrors, ctx.Err())
					mu.Unlock()
				}
				return
			}
		}
	}()

	wg.Wait()
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
	// Create buffered channels with larger capacity
	resourceChan := make(chan interface{}, s.config.BatchSize*len(regions))
	resultChan := make(chan ResourceMetadata, s.config.BatchSize*len(regions))
	errorChan := make(chan error, len(regions)*2)

	var discoveryWg, processingWg sync.WaitGroup

	// Start resource discovery
	s.startResourceDiscovery(ctx, regions, discoverer, resourceChan, errorChan, &discoveryWg, &processingWg)

	// Start resource processing
	s.startResourceProcessing(ctx, resourceChan, resultChan, processor, &processingWg)

	// Manage channel lifecycle
	s.manageChannelLifecycle(ctx, resourceChan, resultChan, errorChan, &discoveryWg, &processingWg)

	// Collect results and errors
	results, scanErrors := s.collectScanResults(ctx, resultChan, errorChan)

	if len(scanErrors) > 0 {
		// Create a detailed error message
		errMsg := fmt.Sprintf("scanning encountered %d errors:\n", len(scanErrors))
		for i, err := range scanErrors {
			errMsg += fmt.Sprintf("  %d. %v\n", i+1, err)
		}

		return results, errors.New(errMsg)
	}

	return results, nil
}
