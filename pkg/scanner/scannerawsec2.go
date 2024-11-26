package scanner

import (
	"context"
	"fmt"
	"strings"
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

// Fetch implements the Scanner interface for retrieving specific EC2 instance details
func (s *EC2Scanner) Fetch(ctx context.Context, arn string, config configuration.TaggyScanConfig) (*ResourceMetadata, error) {
	// Parse instance ID from ARN
	instanceID, region, err := ParseEC2ARN(arn)
	if err != nil {
		return nil, fmt.Errorf("failed to parse EC2 ARN: %w", err)
	}

	// Get EC2 client for the specific region
	// Note: Client creation is prepared for future implementation
	if _, err := s.ClientManager.GetEC2Client(region); err != nil {
		return nil, fmt.Errorf("failed to create EC2 client: %w", err)
	}

	// Implementation for EC2 instance fetching will go here
	// For now, return a placeholder since EC2 scanning isn't fully implemented
	return &ResourceMetadata{
		ID:       instanceID,
		Type:     "ec2",
		Provider: "aws",
		Region:   region,
		Details: struct {
			ARN        string                 `json:"arn,omitempty"`
			Name       string                 `json:"name,omitempty"`
			Status     string                 `json:"status,omitempty"`
			Properties map[string]interface{} `json:"properties,omitempty"`
			Compliance struct {
				IsCompliant bool      `json:"is_compliant"`
				Violations  []string  `json:"violations,omitempty"`
				LastCheck   time.Time `json:"last_check"`
			} `json:"compliance"`
		}{
			ARN: arn,
		},
	}, nil
}

// ParseEC2ARN extracts instance ID and region from EC2 ARN
func ParseEC2ARN(arn string) (string, string, error) {
	// ARN format: arn:aws:ec2:region:account-id:instance/instance-id
	parts := strings.Split(arn, ":")
	if len(parts) != 6 {
		return "", "", fmt.Errorf("invalid EC2 ARN format: %s", arn)
	}
	region := parts[3]
	instanceParts := strings.Split(parts[5], "/")
	if len(instanceParts) != 2 {
		return "", "", fmt.Errorf("invalid EC2 instance ID format in ARN: %s", arn)
	}
	return instanceParts[1], region, nil
}
