package inspector

import (
	"context"
	"fmt"
	"sync"

	"github.com/Excoriate/aws-taggy/pkg/cloud"
	"github.com/Excoriate/aws-taggy/pkg/o11y"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// ResourceProcessor is a function type that processes a single resource and returns its metadata
type ResourceProcessor func(ctx context.Context, resource interface{}) (ResourceMetadata, error)

// ResourceDiscoverer is a function type that discovers resources and sends them to a channel
type ResourceDiscoverer func(ctx context.Context, region string) ([]interface{}, error)

// ScanConfig holds configuration for the scanning process
type ScanConfig struct {
	Logger     *o11y.Logger
	NumWorkers int
	BatchSize  int
}

// DefaultScanConfig returns a default scan configuration
func DefaultScanConfig() ScanConfig {
	return ScanConfig{
		Logger:     o11y.DefaultLogger(),
		NumWorkers: 10,
		BatchSize:  100,
	}
}

// AsyncResourceScanner handles asynchronous resource scanning
type AsyncResourceScanner struct {
	config ScanConfig
}

// NewAsyncResourceScanner creates a new AsyncResourceScanner
func NewAsyncResourceScanner(config ScanConfig) *AsyncResourceScanner {
	return &AsyncResourceScanner{
		config: config,
	}
}

// ScanResources performs asynchronous resource scanning using the provided discoverer and processor functions
func (s *AsyncResourceScanner) ScanResources(
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

// AWSClientManager manages AWS clients for different regions
type AWSClientManager struct {
	mu      sync.RWMutex
	clients map[string]*aws.Config
}

// NewAWSClientManager creates a new AWSClientManager synchronously
func NewAWSClientManager(regions []string) (*AWSClientManager, error) {
	manager := &AWSClientManager{
		clients: make(map[string]*aws.Config),
	}

	// Synchronous client creation
	for _, region := range regions {
		// Create AWS client configuration for the region
		awsClientConfig := cloud.NewAWSClientConfig(region)
		cfg, err := awsClientConfig.LoadConfig(context.Background())
		if err != nil {
			return nil, fmt.Errorf("failed to create AWS client for region %s: %w", region, err)
		}

		manager.clients[region] = cfg
	}

	return manager, nil
}

// GetS3Client retrieves an S3 client for a specific region
func (m *AWSClientManager) GetS3Client(region string) (*s3.Client, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	cfg, exists := m.clients[region]
	if !exists {
		// If the specific region client doesn't exist, create it
		awsClientConfig := cloud.NewAWSClientConfig(region)
		newCfg, err := awsClientConfig.LoadConfig(context.Background())
		if err != nil {
			return nil, fmt.Errorf("failed to create AWS client for region %s: %w", region, err)
		}

		// Store the new client configuration
		m.mu.RUnlock()
		m.mu.Lock()
		m.clients[region] = newCfg
		m.mu.Unlock()
		m.mu.RLock()

		cfg = newCfg
	}

	return s3.NewFromConfig(*cfg), nil
}

// GetEC2Client retrieves an EC2 client for a specific region
func (m *AWSClientManager) GetEC2Client(region string) (*aws.Config, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	cfg, exists := m.clients[region]
	if !exists {
		return nil, fmt.Errorf("no AWS client found for region %s", region)
	}

	return cfg, nil
}
