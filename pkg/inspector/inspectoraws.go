package inspector

import (
	"context"
	"fmt"
	"sync"

	"github.com/Excoriate/aws-taggy/pkg/cloud"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// ResourceProcessor is a function type that processes a single resource and returns its metadata
type ResourceProcessor func(ctx context.Context, resource interface{}) (ResourceMetadata, error)

// ResourceDiscoverer is a function type that discovers resources and sends them to a channel
type ResourceDiscoverer func(ctx context.Context, region string) ([]interface{}, error)

// AWSClient is an interface for AWS service clients
type AWSClient interface {
	CreateFromConfig(cfg *aws.Config) interface{}
}

// S3ClientCreator implements AWSClient for S3
type S3ClientCreator struct{}

func (c *S3ClientCreator) CreateFromConfig(cfg *aws.Config) interface{} {
	return s3.NewFromConfig(*cfg)
}

// EC2ClientCreator implements AWSClient for EC2
type EC2ClientCreator struct{}

// CreateFromConfig creates a new EC2 client from the provided AWS configuration.
//
// This method implements the AWSClient interface for EC2 client creation. It takes an AWS configuration
// pointer and returns a new EC2 client instance that can be used to interact with AWS EC2 services.
//
// The method performs the following key operations:
//  1. Dereferences the provided AWS configuration pointer
//  2. Creates a new EC2 client using the ec2.NewFromConfig function
//  3. Returns the created EC2 client as an interface{} to maintain flexibility
//
// Parameters:
//   - cfg: A pointer to an aws.Config configuration object containing AWS credentials, region, and other settings
//
// Returns:
//   - interface{}: A new EC2 client instance that can be type-asserted to *ec2.Client if needed
//
// Example:
//
//	clientCreator := &EC2ClientCreator{}
//	awsConfig := // load AWS configuration
//	ec2Client := clientCreator.CreateFromConfig(&awsConfig)
func (c *EC2ClientCreator) CreateFromConfig(cfg *aws.Config) interface{} {
	return ec2.NewFromConfig(*cfg)
}

// AWSClientManager manages AWS clients for different regions
// AWSClientManager is a thread-safe manager for AWS client configurations across multiple regions.
//
// This struct provides a concurrent-safe mechanism to store and retrieve AWS client configurations
// for different AWS regions. It uses a read-write mutex to ensure safe concurrent access to the
// client configurations.
//
// Fields:
//   - mu: A read-write mutex (sync.RWMutex) to provide thread-safe access to the clients map
//   - clients: A map storing AWS client configurations, keyed by region name
//
// The AWSClientManager is designed to support multi-region AWS operations by maintaining
// a collection of pre-configured AWS client configurations that can be easily retrieved
// and used across different parts of the application.
type AWSClientManager struct {
	// mu provides concurrent access control for the clients map
	mu sync.RWMutex

	// clients stores AWS configurations indexed by region
	clients map[string]*aws.Config
}

// NewAWSRegionalClientManager creates a new AWSClientManager with AWS client configurations for specified regions.
//
// This function initializes an AWSClientManager by creating AWS client configurations
// for each provided region. It performs the following key operations:
//  1. Creates a new AWSClientManager with an empty clients map
//  2. Iterates through the provided regions
//  3. For each region, creates an AWS client configuration using cloud.NewAWSClientConfig
//  4. Loads the AWS configuration for the region using LoadConfig
//  5. Stores the configuration in the manager's clients map
//
// The method creates client configurations synchronously, which means it will attempt
// to create configurations for all specified regions before returning.
//
// Parameters:
//   - regions: A slice of AWS region strings (e.g., ["us-west-2", "us-east-1"])
//
// Returns:
//   - *AWSClientManager: A fully initialized AWSClientManager with region configurations
//   - error: An error if any region's client configuration fails to load, otherwise nil
//
// Example:
//
//	manager, err := NewAWSRegionalClientManager([]string{"us-west-2", "us-east-1"})
//	if err != nil {
//	    // Handle error
//	}
func NewAWSRegionalClientManager(regions []string) (*AWSClientManager, error) {
	manager := &AWSClientManager{
		clients: make(map[string]*aws.Config),
	}

	// Synchronous client creation for each specified region
	for _, region := range regions {
		// Create AWS client configuration for the current region
		awsClientConfig := cloud.NewAWSClientConfig(region)
		cfg, err := awsClientConfig.LoadConfig(context.Background())
		if err != nil {
			return nil, fmt.Errorf("failed to create AWS client for region %s: %w", region, err)
		}

		// Store the region-specific AWS configuration
		manager.clients[region] = cfg
	}

	return manager, nil
}

// GetClient retrieves an AWS client for a specific region
// GetClient retrieves or creates an AWS service client for a specific region.
//
// This method is responsible for managing AWS service clients across different regions
// in a thread-safe manner. It follows a lazy initialization pattern, creating clients
// on-demand and caching them for future use.
//
// The method performs the following key operations:
//  1. Checks if a client configuration for the specified region already exists
//  2. If not, creates a new AWS client configuration for the region
//  3. Stores the new configuration in a thread-safe manner
//  4. Uses the provided creator to instantiate the specific AWS service client
//
// Parameters:
//   - region: The AWS region for which to retrieve or create a client (e.g., "us-west-2")
//   - creator: An AWSClient implementation that knows how to create a specific AWS service client
//
// Returns:
//   - interface{}: A configured AWS service client for the specified region
//   - error: An error if client configuration or creation fails, otherwise nil
//
// Thread-safety: The method uses read-write mutex to ensure safe concurrent access
// and modification of the client configurations.
func (m *AWSClientManager) GetClient(region string, creator AWSClient) (interface{}, error) {
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

	return creator.CreateFromConfig(cfg), nil
}

// GetS3Client retrieves an S3 client for a specific region
// GetS3Client retrieves an Amazon S3 (Simple Storage Service) client for the specified AWS region.
//
// This method creates or retrieves an existing S3 client configuration for the given region.
// It uses the AWSClientManager's internal client management to ensure efficient client reuse.
//
// Parameters:
//   - region: The AWS region for which to create or retrieve the S3 client (e.g., "us-west-2", "eu-central-1")
//
// Returns:
//   - *s3.Client: A configured AWS S3 client for the specified region
//   - error: An error if the client creation fails, otherwise nil
//
// The method is safe for concurrent use due to the underlying mutex-protected client management.
func (m *AWSClientManager) GetS3Client(region string) (*s3.Client, error) {
	client, err := m.GetClient(region, &S3ClientCreator{})
	if err != nil {
		return nil, err
	}
	return client.(*s3.Client), nil
}

// GetEC2Client retrieves an Amazon EC2 (Elastic Compute Cloud) client for the specified AWS region.
//
// This method creates or retrieves an existing EC2 client configuration for the given region.
// It uses the AWSClientManager's internal client management to ensure efficient client reuse.
//
// Parameters:
//   - region: The AWS region for which to create or retrieve the EC2 client (e.g., "us-west-2", "eu-central-1")
//
// Returns:
//   - *ec2.Client: A configured AWS EC2 client for the specified region
//   - error: An error if the client creation fails, otherwise nil
//
// The method is safe for concurrent use due to the underlying mutex-protected client management.
func (m *AWSClientManager) GetEC2Client(region string) (*ec2.Client, error) {
	client, err := m.GetClient(region, &EC2ClientCreator{})
	if err != nil {
		return nil, err
	}
	return client.(*ec2.Client), nil
}
