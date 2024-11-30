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

func (c *EC2ClientCreator) CreateFromConfig(cfg *aws.Config) interface{} {
	return ec2.NewFromConfig(*cfg)
}

// AWSClientManager manages AWS clients for different regions
type AWSClientManager struct {
	mu      sync.RWMutex
	clients map[string]*aws.Config
}

// NewAWSRegionalClientManager creates a new AWSClientManager synchronously
func NewAWSRegionalClientManager(regions []string) (*AWSClientManager, error) {
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

// GetClient retrieves an AWS client for a specific region
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
func (m *AWSClientManager) GetS3Client(region string) (*s3.Client, error) {
	client, err := m.GetClient(region, &S3ClientCreator{})
	if err != nil {
		return nil, err
	}
	return client.(*s3.Client), nil
}

// GetEC2Client retrieves an EC2 client for a specific region
func (m *AWSClientManager) GetEC2Client(region string) (*ec2.Client, error) {
	client, err := m.GetClient(region, &EC2ClientCreator{})
	if err != nil {
		return nil, err
	}
	return client.(*ec2.Client), nil
}
