package inspector

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Excoriate/aws-taggy/pkg/configuration"
	"github.com/Excoriate/aws-taggy/pkg/o11y"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

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

// EC2Inspector implements the Inspector interface for AWS EC2 resources
type EC2Inspector struct {
	Regions       []string
	ClientManager *AWSClientManager
	Logger        *o11y.Logger
}

// NewEC2Scanner creates a new EC2Scanner with AWS client management
func NewEC2Scanner(regions []string) (*EC2Inspector, error) {
	// Create AWS client manager for the specified regions
	clientManager, err := NewAWSRegionalClientManager(regions)
	if err != nil {
		return nil, fmt.Errorf("failed to create AWS client manager: %w", err)
	}

	// Create a default logger
	logger := o11y.DefaultLogger()

	return &EC2Inspector{
		Regions:       regions,
		ClientManager: clientManager,
		Logger:        logger,
	}, nil
}

// getRegionFromAZ extracts the region from an availability zone
func (s *EC2Inspector) getRegionFromAZ(az string) string {
	// AZ format is like "us-east-1a", so remove the last character to get the region
	if len(az) > 0 {
		return az[:len(az)-1]
	}
	return s.Regions[0] // fallback to first configured region
}

// Inspect discovers EC2 instances and their metadata across specified regions
func (s *EC2Inspector) Inspect(ctx context.Context, config configuration.TaggyScanConfig) (*InspectResult, error) {
	s.Logger.Info("Starting EC2 resource scanning",
		"regions", s.Regions)

	result := &InspectResult{
		StartTime: time.Now(),
		Region:    s.Regions[0],
	}

	// Create async scanner with default config
	scanner := NewAsyncResourceInspector(DefaultInspectorConfig())

	// Define the resource discoverer function
	discoverer := func(ctx context.Context, region string) ([]interface{}, error) {
		// Get EC2 client for this region
		ec2Client, err := s.ClientManager.GetEC2Client(region)
		if err != nil {
			return nil, fmt.Errorf("failed to get EC2 client: %w", err)
		}

		// List instances
		instances, err := s.listInstances(ctx, ec2Client)
		if err != nil {
			return nil, fmt.Errorf("failed to list instances: %w", err)
		}

		// Convert to interface slice
		resources := make([]interface{}, len(instances))
		for i, instance := range instances {
			resources[i] = instance
		}

		return resources, nil
	}

	// Define the resource processor function
	processor := func(ctx context.Context, resource interface{}) (ResourceMetadata, error) {
		instance := resource.(types.Instance)

		// Get EC2 client for the instance's region
		region := s.getRegionFromAZ(aws.ToString(instance.Placement.AvailabilityZone))

		// Get instance tags
		tags := make(map[string]string)
		for _, tag := range instance.Tags {
			tags[aws.ToString(tag.Key)] = aws.ToString(tag.Value)
		}

		// Create resource metadata
		metadata := ResourceMetadata{
			ID:           aws.ToString(instance.InstanceId),
			Type:         "ec2",
			Provider:     "aws",
			Region:       region,
			DiscoveredAt: time.Now(),
			Tags:         tags,
			RawResponse:  instance,
		}

		// Populate extended details
		accountID := "unknown" // EC2 instances don't have direct OwnerId field
		metadata.Details.ARN = fmt.Sprintf("arn:aws:ec2:%s:%s:instance/%s",
			region, accountID, aws.ToString(instance.InstanceId))
		metadata.Details.Name = s.getInstanceName(instance)
		metadata.Details.Status = string(instance.State.Name)
		metadata.Details.Properties = map[string]interface{}{
			"instance_type":     instance.InstanceType,
			"availability_zone": instance.Placement.AvailabilityZone,
			"launch_time":       instance.LaunchTime,
		}

		return metadata, nil
	}

	// Perform the async scan
	resources, err := scanner.InspectResourcesAsync(ctx, s.Regions, discoverer, processor)
	if err != nil {
		return nil, fmt.Errorf("failed to scan EC2 resources: %w", err)
	}

	// Update result with scanned resources
	result.Resources = resources
	result.TotalResources = len(resources)
	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)

	s.Logger.Info("EC2 scanning completed",
		"total_resources", result.TotalResources,
		"duration", result.Duration)

	return result, nil
}

// listInstances retrieves all EC2 instances in a region
func (s *EC2Inspector) listInstances(ctx context.Context, client *ec2.Client) ([]types.Instance, error) {
	input := &ec2.DescribeInstancesInput{}
	output, err := client.DescribeInstances(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to list instances: %w", err)
	}

	var instances []types.Instance
	for _, reservation := range output.Reservations {
		instances = append(instances, reservation.Instances...)
	}
	return instances, nil
}

// getInstanceName extracts the Name tag or returns a default name
func (s *EC2Inspector) getInstanceName(instance types.Instance) string {
	for _, tag := range instance.Tags {
		if aws.ToString(tag.Key) == "Name" {
			return aws.ToString(tag.Value)
		}
	}
	if instance.InstanceId != nil {
		return aws.ToString(instance.InstanceId)
	}
	return "Unnamed Instance"
}

// Fetch implements the Inspector interface for retrieving specific EC2 instance details
func (s *EC2Inspector) Fetch(ctx context.Context, arn string, config configuration.TaggyScanConfig) (*ResourceMetadata, error) {
	// Parse instance ID and region from ARN
	instanceID, region, err := ParseEC2ARN(arn)
	if err != nil {
		return nil, fmt.Errorf("failed to parse EC2 ARN: %w", err)
	}

	// Get EC2 client for the instance's region
	ec2Client, err := s.ClientManager.GetEC2Client(region)
	if err != nil {
		return nil, fmt.Errorf("failed to create EC2 client: %w", err)
	}

	// Describe the specific instance
	input := &ec2.DescribeInstancesInput{
		InstanceIds: []string{instanceID},
	}
	output, err := ec2Client.DescribeInstances(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch EC2 instance: %w", err)
	}

	// Ensure we have an instance
	if len(output.Reservations) == 0 || len(output.Reservations[0].Instances) == 0 {
		return nil, fmt.Errorf("no instance found with ID %s", instanceID)
	}

	instance := output.Reservations[0].Instances[0]

	// Get instance tags
	tags := make(map[string]string)
	for _, tag := range instance.Tags {
		tags[aws.ToString(tag.Key)] = aws.ToString(tag.Value)
	}

	// Create resource metadata
	resourceMeta := &ResourceMetadata{
		ID:           instanceID,
		Type:         "ec2",
		Provider:     "aws",
		Region:       region,
		Tags:         tags,
		DiscoveredAt: time.Now(),
	}

	// Populate extended details
	resourceMeta.Details.ARN = arn
	resourceMeta.Details.Name = s.getInstanceName(instance)
	resourceMeta.Details.Status = string(instance.State.Name)
	resourceMeta.Details.Properties = map[string]interface{}{
		"instance_type":     instance.InstanceType,
		"availability_zone": instance.Placement.AvailabilityZone,
		"launch_time":       instance.LaunchTime,
	}

	return resourceMeta, nil
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
