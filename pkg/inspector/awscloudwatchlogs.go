package inspector

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Excoriate/aws-taggy/pkg/configuration"
	"github.com/Excoriate/aws-taggy/pkg/o11y"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"
)

// CloudWatchLogsClientCreator implements AWSClient for CloudWatch Logs
type CloudWatchLogsClientCreator struct{}

// CreateFromConfig creates a new CloudWatch Logs client from the provided AWS configuration.
//
// This method implements the AWSClient interface for CloudWatch Logs client creation. It takes an AWS configuration
// pointer and returns a new CloudWatch Logs client instance that can be used to interact with AWS CloudWatch Logs services.
//
// The method performs the following key operations:
//  1. Dereferences the provided AWS configuration pointer
//  2. Creates a new CloudWatch Logs client using the cloudwatchlogs.NewFromConfig function
//  3. Returns the created CloudWatch Logs client as an interface{} to maintain flexibility
//
// Parameters:
//   - cfg: A pointer to an aws.Config configuration object containing AWS credentials, region, and other settings
//
// Returns:
//   - interface{}: A new CloudWatch Logs client instance that can be type-asserted to *cloudwatchlogs.Client if needed
//
// Example:
//
//	clientCreator := &CloudWatchLogsClientCreator{}
//	awsConfig := // load AWS configuration
//	cwLogsClient := clientCreator.CreateFromConfig(&awsConfig)
func (c *CloudWatchLogsClientCreator) CreateFromConfig(cfg *aws.Config) interface{} {
	return cloudwatchlogs.NewFromConfig(*cfg)
}

// GetCloudWatchLogsClient retrieves a CloudWatch Logs client for the specified AWS region.
//
// This method creates or retrieves an existing CloudWatch Logs client configuration for the given region.
// It uses the AWSClientManager's internal client management to ensure efficient client reuse.
//
// Parameters:
//   - region: The AWS region for which to create or retrieve the CloudWatch Logs client (e.g., "us-west-2", "eu-central-1")
//
// Returns:
//   - *cloudwatchlogs.Client: A configured AWS CloudWatch Logs client for the specified region
//   - error: An error if the client creation fails, otherwise nil
//
// The method is safe for concurrent use due to the underlying mutex-protected client management.
func (m *AWSClientManager) GetCloudWatchLogsClient(region string) (*cloudwatchlogs.Client, error) {
	client, err := m.GetClient(region, &CloudWatchLogsClientCreator{})
	if err != nil {
		return nil, err
	}
	return client.(*cloudwatchlogs.Client), nil
}

// CloudWatchLogsInspector implements the Scanner interface for AWS CloudWatch Logs resources.
// It provides functionality to discover and inspect CloudWatch Log Groups across multiple AWS regions.
type CloudWatchLogsInspector struct {
	Regions       []string
	ClientManager *AWSClientManager
	Logger        *o11y.Logger
}

// NewCloudWatchLogsInspector creates a new CloudWatchLogsScanner with AWS client management.
//
// This function initializes a new scanner with the specified regions and sets up the necessary
// AWS client manager and logger for CloudWatch Logs operations.
//
// Parameters:
//   - regions: A slice of AWS region identifiers where the scanner will operate
//
// Returns:
//   - *CloudWatchLogsScanner: A new scanner instance
//   - error: An error if initialization fails
func NewCloudWatchLogsInspector(regions []string) (*CloudWatchLogsInspector, error) {
	// Create AWS client manager for the specified regions
	clientManager, err := NewAWSRegionalClientManager(regions)
	if err != nil {
		return nil, fmt.Errorf("failed to create AWS client manager: %w", err)
	}

	// Create a default logger
	logger := o11y.DefaultLogger()

	return &CloudWatchLogsInspector{
		Regions:       regions,
		ClientManager: clientManager,
		Logger:        logger,
	}, nil
}

// Inspect discovers CloudWatch Log Groups and their metadata across specified regions
func (s *CloudWatchLogsInspector) Inspect(ctx context.Context, config configuration.TaggyScanConfig) (*InspectResult, error) {
	s.Logger.Info("Starting CloudWatch Logs resource scanning",
		"regions", s.Regions)

	result := &InspectResult{
		StartTime: time.Now(),
		Region:    s.Regions[0],
	}

	// Create async scanner with default config
	scanner := NewAsyncResourceInspector(DefaultInspectorConfig())

	// Define the resource discoverer function
	discoverer := func(ctx context.Context, region string) ([]interface{}, error) {
		// Get CloudWatch Logs client for this region
		cwLogsClient, err := s.ClientManager.GetCloudWatchLogsClient(region)
		if err != nil {
			return nil, fmt.Errorf("failed to get CloudWatch Logs client: %w", err)
		}

		// List log groups
		logGroups, err := s.listLogGroups(ctx, cwLogsClient)
		if err != nil {
			return nil, fmt.Errorf("failed to list log groups: %w", err)
		}

		// Convert to interface slice
		resources := make([]interface{}, len(logGroups))
		for i, logGroup := range logGroups {
			resources[i] = logGroup
		}

		return resources, nil
	}

	// Define the resource processor function
	processor := func(ctx context.Context, resource interface{}) (ResourceMetadata, error) {
		logGroup, ok := resource.(types.LogGroup)
		if !ok {
			return ResourceMetadata{}, fmt.Errorf("invalid resource type: expected LogGroup")
		}

		// Get CloudWatch Logs client for the region
		cwLogsClient, err := s.ClientManager.GetCloudWatchLogsClient(s.Regions[0])
		if err != nil {
			return ResourceMetadata{}, fmt.Errorf("failed to get CloudWatch Logs client: %w", err)
		}

		// Get log group tags
		tags, err := s.getLogGroupTags(ctx, cwLogsClient, aws.ToString(logGroup.LogGroupName))
		if err != nil {
			s.Logger.Warn("Failed to get log group tags",
				"log_group", aws.ToString(logGroup.LogGroupName),
				"error", err)
			tags = make(map[string]string)
		}

		// Create resource metadata
		metadata := ResourceMetadata{
			ID:           aws.ToString(logGroup.LogGroupName),
			Type:         "cloudwatch_logs",
			Provider:     "aws",
			Region:       s.Regions[0],
			DiscoveredAt: time.Now(),
			Tags:         tags,
			RawResponse:  logGroup,
		}

		// Populate extended details
		metadata.Details.ARN = fmt.Sprintf("arn:aws:logs:%s:%s:log-group:%s:*",
			s.Regions[0], "unknown_account", aws.ToString(logGroup.LogGroupName))
		metadata.Details.Name = aws.ToString(logGroup.LogGroupName)
		metadata.Details.Properties = map[string]interface{}{
			"creation_time":     logGroup.CreationTime,
			"retention_in_days": logGroup.RetentionInDays,
			"stored_bytes":      logGroup.StoredBytes,
			"kms_key_id":        aws.ToString(logGroup.KmsKeyId),
		}

		return metadata, nil
	}

	// Perform the async scan
	resources, err := scanner.InspectResourcesAsync(ctx, s.Regions, discoverer, processor)
	if err != nil {
		return nil, fmt.Errorf("failed to scan CloudWatch Logs resources: %w", err)
	}

	// Update result with scanned resources
	result.Resources = resources
	result.TotalResources = len(resources)
	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)

	s.Logger.Info("CloudWatch Logs scanning completed",
		"total_resources", result.TotalResources,
		"duration", result.Duration)

	return result, nil
}

// listLogGroups retrieves all CloudWatch Log Groups in a region.
//
// This method uses pagination to retrieve all log groups in the specified region.
// It handles the AWS API pagination automatically and aggregates the results.
//
// Parameters:
//   - ctx: Context for the API calls
//   - client: The CloudWatch Logs client to use
//
// Returns:
//   - []types.LogGroup: A slice of discovered log groups
//   - error: An error if the operation fails
func (s *CloudWatchLogsInspector) listLogGroups(ctx context.Context, client *cloudwatchlogs.Client) ([]types.LogGroup, error) {
	var logGroups []types.LogGroup
	input := &cloudwatchlogs.DescribeLogGroupsInput{}
	paginator := cloudwatchlogs.NewDescribeLogGroupsPaginator(client, input)

	for paginator.HasMorePages() {
		output, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list log groups: %w", err)
		}
		if output != nil && len(output.LogGroups) > 0 {
			logGroups = append(logGroups, output.LogGroups...)
		}
	}

	return logGroups, nil
}

// getLogGroupTags retrieves tags for a specific log group.
//
// This method constructs the ARN for the log group and uses the ListTagsForResource API
// to retrieve its tags. It handles cases where the log group might not have any tags.
//
// Parameters:
//   - ctx: Context for the API calls
//   - client: The CloudWatch Logs client to use
//   - logGroupName: The name of the log group
//
// Returns:
//   - map[string]string: A map of tag key-value pairs
//   - error: An error if the operation fails
func (s *CloudWatchLogsInspector) getLogGroupTags(ctx context.Context, client *cloudwatchlogs.Client, logGroupName string) (map[string]string, error) {
	// Construct the ARN for the log group
	logGroupARN := fmt.Sprintf("arn:aws:logs:%s:%s:log-group:%s:*",
		s.Regions[0], "unknown_account", logGroupName)

	input := &cloudwatchlogs.ListTagsForResourceInput{
		ResourceArn: aws.String(logGroupARN),
	}

	// Retrieve log group tags
	tagsOutput, err := client.ListTagsForResource(ctx, input)
	if err != nil {
		var awsErr *types.ResourceNotFoundException
		if errors.As(err, &awsErr) {
			return make(map[string]string), nil
		}
		return nil, fmt.Errorf("failed to get log group tags: %w", err)
	}

	// Return tags directly as they are already in map[string]string format
	if tagsOutput != nil && tagsOutput.Tags != nil {
		return tagsOutput.Tags, nil
	}

	return make(map[string]string), nil
}

// Fetch implements the Scanner interface for retrieving specific CloudWatch Log Group details
func (s *CloudWatchLogsInspector) Fetch(ctx context.Context, arn string, config configuration.TaggyScanConfig) (*ResourceMetadata, error) {
	// Parse log group name from ARN
	logGroupName, region, err := ParseCloudWatchLogsARN(arn)
	if err != nil {
		return nil, fmt.Errorf("failed to parse CloudWatch Logs ARN: %w", err)
	}

	// Get CloudWatch Logs client for the region
	cwLogsClient, err := s.ClientManager.GetCloudWatchLogsClient(region)
	if err != nil {
		return nil, fmt.Errorf("failed to create CloudWatch Logs client: %w", err)
	}

	// Get log group details
	input := &cloudwatchlogs.DescribeLogGroupsInput{
		LogGroupNamePrefix: aws.String(logGroupName),
	}
	output, err := cwLogsClient.DescribeLogGroups(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to get log group details: %w", err)
	}

	// Find the exact log group
	var logGroup *types.LogGroup
	if output != nil && len(output.LogGroups) > 0 {
		for _, lg := range output.LogGroups {
			if aws.ToString(lg.LogGroupName) == logGroupName {
				logGroup = &lg
				break
			}
		}
	}
	if logGroup == nil {
		return nil, fmt.Errorf("log group not found: %s", logGroupName)
	}

	// Get log group tags
	tags, err := s.getLogGroupTags(ctx, cwLogsClient, logGroupName)
	if err != nil {
		s.Logger.Warn("Failed to get log group tags", "log_group", logGroupName, "error", err)
		tags = make(map[string]string)
	}

	// Create resource metadata
	resourceMeta := &ResourceMetadata{
		ID:           logGroupName,
		Type:         "cloudwatch_logs",
		Provider:     "aws",
		Region:       region,
		Tags:         tags,
		DiscoveredAt: time.Now(),
		RawResponse:  logGroup,
	}

	// Populate extended details
	resourceMeta.Details.ARN = arn
	resourceMeta.Details.Name = logGroupName
	resourceMeta.Details.Properties = map[string]interface{}{
		"creation_time":     logGroup.CreationTime,
		"retention_in_days": logGroup.RetentionInDays,
		"stored_bytes":      logGroup.StoredBytes,
		"kms_key_id":        aws.ToString(logGroup.KmsKeyId),
	}

	return resourceMeta, nil
}

// ParseCloudWatchLogsARN extracts log group name and region from CloudWatch Logs ARN
//
// This function parses an ARN for a CloudWatch Logs log group and extracts the log group name
// and region. It handles ARNs in the format:
// arn:aws:logs:region:account-id:log-group:log-group-name:*
//
// The function also handles log group names that may contain ':' characters by joining
// all remaining parts after the log-group prefix.
//
// Parameters:
//   - arn: The ARN string to parse
//
// Returns:
//   - string: The log group name
//   - string: The AWS region
//   - error: An error if the ARN format is invalid
func ParseCloudWatchLogsARN(arn string) (string, string, error) {
	// ARN format: arn:aws:logs:region:account-id:log-group:log-group-name:*
	parts := strings.Split(arn, ":")
	if len(parts) < 7 {
		return "", "", fmt.Errorf("invalid CloudWatch Logs ARN format: %s", arn)
	}

	// Extract region
	region := parts[3]

	// Handle log group names that may contain ':' characters
	// Join all parts after "log-group" until the last part (which should be "*")
	logGroupParts := parts[6 : len(parts)-1]
	logGroupName := strings.Join(logGroupParts, ":")

	return logGroupName, region, nil
}
