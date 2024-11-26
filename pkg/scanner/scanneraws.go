package scanner

import (
	"context"
	"fmt"
	"sync"

	"github.com/Excoriate/aws-taggy/pkg/cloud"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

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
