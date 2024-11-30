package inspector

import (
	"context"
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

// InspectResourcesAsync performs asynchronous resource scanning using the provided discoverer and processor functions
// InspectResourcesAsync performs an asynchronous, parallel resource inspection across multiple regions.
//
// This method orchestrates a concurrent resource discovery and processing workflow. It takes a context,
// a list of regions, a resource discoverer function, and a resource processor function as inputs.
// The method discovers resources across specified regions in parallel and processes them concurrently
// using a configurable number of worker goroutines.
//
// The method handles the entire lifecycle of resource inspection, including:
//   - Parallel resource discovery across multiple regions
//   - Concurrent resource processing
//   - Error handling and logging
//   - Aggregation of processed resource metadata
//
// Parameters:
//   - ctx: A context.Context for managing cancellation, timeouts, and request-scoped values
//   - regions: A slice of region identifiers to scan for resources
//   - discoverer: A ResourceDiscoverer function that finds resources in a specific region
//   - processor: A ResourceProcessor function that processes individual resources
//
// Returns:
//   - A slice of ResourceMetadata containing processed resource information
//   - An error if any critical failures occur during discovery or processing
//
// The method uses channels and wait groups to manage concurrent operations, ensuring
// efficient and controlled parallel processing of resources.
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
	for _, region := range regions {
		discoveryWg.Add(1)
		go func(r string) {
			defer discoveryWg.Done()

			// Discover resources in this region
			resources, err := discoverer(ctx, r)
			if err != nil {
				errorChan <- fmt.Errorf("failed to discover resources in region %s: %w", r, err)
				return
			}

			s.config.Logger.Info(fmt.Sprintf("Discovered resources in region %s", r),
				"region", r,
				"count", len(resources))

			// Send resources to processing channel
			for _, resource := range resources {
				resourceChan <- resource
				processingWg.Add(1)
			}
		}(region)
	}

	// Start resource processing workers
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

					resultChan <- metadata
				}()
			}
		}(i)
	}

	// Start a goroutine to close channels when all processing is done
	go func() {
		discoveryWg.Wait()  // Wait for all discovery goroutines
		close(resourceChan) // Close resource channel when discovery is done
		processingWg.Wait() // Wait for all processing to complete
		close(resultChan)   // Close result channel
		close(errorChan)    // Close error channel
	}()

	// Collect results and errors
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

	// Check for any errors
	if len(scanErrors) > 0 {
		return results, fmt.Errorf("scanning encountered %d errors", len(scanErrors))
	}

	return results, nil
}
