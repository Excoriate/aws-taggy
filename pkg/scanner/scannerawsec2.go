package scanner

import (
	"context"
	"fmt"
	"time"

	"github.com/Excoriate/aws-taggy/pkg/configuration"
)

// EC2Scanner implements the Scanner interface for AWS EC2 resources
type EC2Scanner struct {
	Regions       []string
	ClientManager *AWSClientManager
}

// NewEC2Scanner creates a new EC2Scanner with AWS client management
func NewEC2Scanner(regions []string) (*EC2Scanner, error) {
	// Create AWS client manager for the specified regions
	clientManager, err := NewAWSClientManager(regions)
	if err != nil {
		return nil, fmt.Errorf("failed to create AWS client manager: %w", err)
	}

	return &EC2Scanner{
		Regions:       regions,
		ClientManager: clientManager,
	}, nil
}

// Scan discovers EC2 instances and their metadata across specified regions
func (s *EC2Scanner) Scan(ctx context.Context, resource Resource, config configuration.TaggyScanConfig) (*ScanResult, error) {
	result := &ScanResult{
		StartTime: time.Now(),
		Region:    resource.GetRegion(),
	}

	// Placeholder for actual EC2 scanning logic
	// This is where you'd implement actual EC2 resource discovery
	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)
	result.TotalResources = 0 // No resources discovered yet

	return result, nil
}
