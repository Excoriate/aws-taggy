package inspector

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Excoriate/aws-taggy/pkg/configuration"
	"github.com/Excoriate/aws-taggy/pkg/o11y"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/aws/aws-sdk-go-v2/service/rds/types"
)

// RDSClientCreator implements AWSClient for RDS
type RDSClientCreator struct{}

func (c *RDSClientCreator) CreateFromConfig(cfg *aws.Config) interface{} {
	return rds.NewFromConfig(*cfg)
}

// GetRDSClient retrieves an RDS client for the specified AWS region.
//
// This method creates or retrieves an existing RDS client configuration for the given region.
// It uses the AWSClientManager's internal client management to ensure efficient client reuse.
//
// Parameters:
//   - region: The AWS region for which to create or retrieve the RDS client
//
// Returns:
//   - *rds.Client: A configured AWS RDS client
//   - error: An error if client creation fails
func (m *AWSClientManager) GetRDSClient(region string) (*rds.Client, error) {
	client, err := m.GetClient(region, &RDSClientCreator{})
	if err != nil {
		return nil, err
	}
	return client.(*rds.Client), nil
}

// RDSInspector implements the Inspector interface for AWS RDS resources
type RDSInspector struct {
	Regions       []string
	ClientManager *AWSClientManager
	Logger        *o11y.Logger
}

// NewRDSInspector creates a new inspector with AWS client management
func NewRDSInspector(regions []string) (*RDSInspector, error) {
	// Create AWS client manager for the specified regions
	clientManager, err := NewAWSRegionalClientManager(regions)
	if err != nil {
		return nil, fmt.Errorf("failed to create AWS client manager: %w", err)
	}

	// Create a default logger
	logger := o11y.DefaultLogger()

	return &RDSInspector{
		Regions:       regions,
		ClientManager: clientManager,
		Logger:        logger,
	}, nil
}

// Inspect discovers RDS database instances and their metadata across specified regions
func (r *RDSInspector) Inspect(ctx context.Context, config configuration.TaggyScanConfig) (*InspectResult, error) {
	r.Logger.Info("Starting RDS resource scanning",
		"regions", r.Regions)

	result := &InspectResult{
		StartTime: time.Now(),
		Region:    r.Regions[0],
	}

	// Create async scanner with default config
	scanner := NewAsyncResourceInspector(DefaultInspectorConfig())

	// Define the resource discoverer function
	discoverer := func(ctx context.Context, region string) ([]interface{}, error) {
		// Get RDS client for this region
		rdsClient, err := r.ClientManager.GetRDSClient(region)
		if err != nil {
			return nil, fmt.Errorf("failed to get RDS client: %w", err)
		}

		// List database instances
		instances, err := r.listDatabaseInstances(ctx, rdsClient)
		if err != nil {
			return nil, fmt.Errorf("failed to list database instances: %w", err)
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
		instance := resource.(types.DBInstance)

		// Get RDS client for initial region
		rdsClient, err := r.ClientManager.GetRDSClient(r.Regions[0])
		if err != nil {
			return ResourceMetadata{}, fmt.Errorf("failed to get RDS client: %w", err)
		}

		// Fetch database instance tags
		tags, err := r.getDatabaseInstanceTags(ctx, rdsClient, *instance.DBInstanceArn)
		if err != nil {
			r.Logger.Warn("Failed to get database instance tags",
				"instance_arn", *instance.DBInstanceArn,
				"error", err)
			tags = make(map[string]string)
		}

		// Create resource metadata
		metadata := ResourceMetadata{
			ID:           *instance.DBInstanceArn,
			Type:         "rds",
			Provider:     "aws",
			Region:       r.Regions[0], // RDS is regional
			DiscoveredAt: time.Now(),
			Tags:         tags,
			RawResponse:  instance,
		}

		// Populate extended details
		metadata.Details.ARN = *instance.DBInstanceArn
		metadata.Details.Name = r.getDatabaseInstanceName(instance)
		metadata.Details.Status = aws.ToString(instance.DBInstanceStatus)
		metadata.Details.Properties = map[string]interface{}{
			"instance_class":    instance.DBInstanceClass,
			"engine":            instance.Engine,
			"engine_version":    instance.EngineVersion,
			"availability_zone": instance.AvailabilityZone,
		}

		return metadata, nil
	}

	// Perform the async scan
	resources, err := scanner.InspectResourcesAsync(ctx, r.Regions, discoverer, processor)
	if err != nil {
		return nil, fmt.Errorf("failed to scan RDS resources: %w", err)
	}

	// Update result with scanned resources
	result.Resources = resources
	result.TotalResources = len(resources)
	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)

	r.Logger.Info("RDS scanning completed",
		"total_resources", result.TotalResources,
		"duration", result.Duration)

	return result, nil
}

// listDatabaseInstances retrieves all RDS database instances
func (r *RDSInspector) listDatabaseInstances(ctx context.Context, client *rds.Client) ([]types.DBInstance, error) {
	var instances []types.DBInstance
	paginator := rds.NewDescribeDBInstancesPaginator(client, &rds.DescribeDBInstancesInput{})

	for paginator.HasMorePages() {
		output, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list database instances: %w", err)
		}
		instances = append(instances, output.DBInstances...)
	}

	return instances, nil
}

// getDatabaseInstanceTags retrieves tags for a specific RDS database instance
func (r *RDSInspector) getDatabaseInstanceTags(ctx context.Context, client *rds.Client, instanceARN string) (map[string]string, error) {
	// List tags for the database instance
	tagsOutput, err := client.ListTagsForResource(ctx, &rds.ListTagsForResourceInput{
		ResourceName: aws.String(instanceARN),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance tags: %w", err)
	}

	tags := make(map[string]string)
	for _, tag := range tagsOutput.TagList {
		tags[*tag.Key] = *tag.Value
	}

	return tags, nil
}

// getDatabaseInstanceName extracts the database instance name or returns a default name
func (r *RDSInspector) getDatabaseInstanceName(instance types.DBInstance) string {
	if instance.DBInstanceIdentifier != nil {
		return *instance.DBInstanceIdentifier
	}
	return "Unnamed RDS Instance"
}

// Fetch implements the Scanner interface for retrieving specific RDS database instance details
func (r *RDSInspector) Fetch(ctx context.Context, arn string, config configuration.TaggyScanConfig) (*ResourceMetadata, error) {
	// Parse database instance ARN
	instanceARN, region, err := ParseRDSARN(arn)
	if err != nil {
		return nil, fmt.Errorf("failed to parse RDS ARN: %w", err)
	}

	// Get RDS client for the database instance's region
	rdsClient, err := r.ClientManager.GetRDSClient(region)
	if err != nil {
		return nil, fmt.Errorf("failed to create RDS client: %w", err)
	}

	// Describe the specific database instance
	input := &rds.DescribeDBInstancesInput{
		Filters: []types.Filter{
			{
				Name:   aws.String("db-instance-id"),
				Values: []string{instanceARN},
			},
		},
	}
	output, err := rdsClient.DescribeDBInstances(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch RDS database instance: %w", err)
	}

	// Ensure we have a database instance
	if len(output.DBInstances) == 0 {
		return nil, fmt.Errorf("no database instance found with ARN %s", instanceARN)
	}

	instance := output.DBInstances[0]

	// Get database instance tags
	tags, err := r.getDatabaseInstanceTags(ctx, rdsClient, *instance.DBInstanceArn)
	if err != nil {
		r.Logger.Warn("Failed to get database instance tags", "instance_arn", instanceARN, "error", err)
		tags = make(map[string]string)
	}

	// Create resource metadata
	resourceMeta := &ResourceMetadata{
		ID:           instanceARN,
		Type:         "rds",
		Provider:     "aws",
		Region:       region,
		Tags:         tags,
		DiscoveredAt: time.Now(),
	}

	// Populate extended details
	resourceMeta.Details.ARN = arn
	resourceMeta.Details.Name = r.getDatabaseInstanceName(instance)
	resourceMeta.Details.Status = aws.ToString(instance.DBInstanceStatus)
	resourceMeta.Details.Properties = map[string]interface{}{
		"instance_class":    instance.DBInstanceClass,
		"engine":            instance.Engine,
		"engine_version":    instance.EngineVersion,
		"availability_zone": instance.AvailabilityZone,
	}

	return resourceMeta, nil
}

// ParseRDSARN extracts database instance ARN and region from RDS ARN
func ParseRDSARN(arn string) (string, string, error) {
	// ARN format: arn:aws:rds:region:account-id:db:db-instance-name
	parts := strings.Split(arn, ":")
	if len(parts) != 7 {
		return "", "", fmt.Errorf("invalid RDS ARN format: %s", arn)
	}
	region := parts[3]
	instanceName := parts[6]
	return instanceName, region, nil
}
