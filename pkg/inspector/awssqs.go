package inspector

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Excoriate/aws-taggy/pkg/configuration"
	"github.com/Excoriate/aws-taggy/pkg/o11y"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
)

// SQSClientCreator implements AWSClient for SQS
type SQSClientCreator struct{}

func (c *SQSClientCreator) CreateFromConfig(cfg *aws.Config) interface{} {
	return sqs.NewFromConfig(*cfg)
}

// GetSQSClient retrieves an SQS client for the specified AWS region.
//
// This method creates or retrieves an existing SQS client configuration for the given region.
// It uses the AWSClientManager's internal client management to ensure efficient client reuse.
//
// Parameters:
//   - region: The AWS region for which to create or retrieve the SQS client
//
// Returns:
//   - *sqs.Client: A configured AWS SQS client
//   - error: An error if client creation fails
func (m *AWSClientManager) GetSQSClient(region string) (*sqs.Client, error) {
	client, err := m.GetClient(region, &SQSClientCreator{})
	if err != nil {
		return nil, err
	}
	return client.(*sqs.Client), nil
}

// SQSInspector implements the Inspector interface for AWS SQS resources
type SQSInspector struct {
	Regions       []string
	ClientManager *AWSClientManager
	Logger        *o11y.Logger
}

// NewSQSInspector creates a new inspector with AWS client management
func NewSQSInspector(regions []string) (*SQSInspector, error) {
	// Create AWS client manager for the specified regions
	clientManager, err := NewAWSRegionalClientManager(regions)
	if err != nil {
		return nil, fmt.Errorf("failed to create AWS client manager: %w", err)
	}

	// Create a default logger
	logger := o11y.DefaultLogger()

	return &SQSInspector{
		Regions:       regions,
		ClientManager: clientManager,
		Logger:        logger,
	}, nil
}

// Inspect discovers SQS queues and their metadata across specified regions
func (s *SQSInspector) Inspect(ctx context.Context, config configuration.TaggyScanConfig) (*InspectResult, error) {
	s.Logger.Info("Starting SQS resource scanning",
		"regions", s.Regions)

	result := &InspectResult{
		StartTime: time.Now(),
		Region:    s.Regions[0],
	}

	// Create async scanner with default config
	scanner := NewAsyncResourceInspector(DefaultInspectorConfig())

	// Define the resource discoverer function
	discoverer := func(ctx context.Context, region string) ([]interface{}, error) {
		// Get SQS client for this region
		sqsClient, err := s.ClientManager.GetSQSClient(region)
		if err != nil {
			return nil, fmt.Errorf("failed to get SQS client: %w", err)
		}

		// List queues
		queues, err := s.listQueues(ctx, sqsClient)
		if err != nil {
			return nil, fmt.Errorf("failed to list queues: %w", err)
		}

		// Convert to interface slice
		resources := make([]interface{}, len(queues))
		for i, queueURL := range queues {
			resources[i] = queueURL
		}

		return resources, nil
	}

	// Define the resource processor function
	processor := func(ctx context.Context, resource interface{}) (ResourceMetadata, error) {
		queueURL := resource.(string)

		// Get SQS client for initial region
		sqsClient, err := s.ClientManager.GetSQSClient(s.Regions[0])
		if err != nil {
			return ResourceMetadata{}, fmt.Errorf("failed to get SQS client: %w", err)
		}

		// Get queue attributes to fetch ARN and other details
		attributes, err := s.getQueueAttributes(ctx, sqsClient, queueURL)
		if err != nil {
			return ResourceMetadata{}, fmt.Errorf("failed to get queue attributes: %w", err)
		}

		// Get queue tags
		tags, err := s.getQueueTags(ctx, sqsClient, queueURL)
		if err != nil {
			s.Logger.Warn("Failed to get queue tags",
				"queue_url", queueURL,
				"error", err)
			tags = make(map[string]string)
		}

		queueARN := attributes["QueueArn"]

		// Create resource metadata
		metadata := ResourceMetadata{
			ID:           queueARN,
			Type:         "sqs",
			Provider:     "aws",
			Region:       s.Regions[0], // SQS is regional
			DiscoveredAt: time.Now(),
			Tags:         tags,
			RawResponse:  attributes,
		}

		// Populate extended details
		metadata.Details.ARN = queueARN
		metadata.Details.Name = s.getQueueName(queueURL)
		metadata.Details.Properties = map[string]interface{}{
			"queue_url":          queueURL,
			"queue_arn":          queueARN,
			"visibility_timeout": attributes["VisibilityTimeout"],
			"delay_seconds":      attributes["DelaySeconds"],
			"queue_type":         attributes["FifoQueue"],
		}

		return metadata, nil
	}

	// Perform the async scan
	resources, err := scanner.InspectResourcesAsync(ctx, s.Regions, discoverer, processor)
	if err != nil {
		return nil, fmt.Errorf("failed to scan SQS resources: %w", err)
	}

	// Update result with scanned resources
	result.Resources = resources
	result.TotalResources = len(resources)
	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)

	s.Logger.Info("SQS scanning completed",
		"total_resources", result.TotalResources,
		"duration", result.Duration)

	return result, nil
}

// listQueues retrieves all SQS queues
func (s *SQSInspector) listQueues(ctx context.Context, client *sqs.Client) ([]string, error) {
	var queueURLs []string
	paginator := sqs.NewListQueuesPaginator(client, &sqs.ListQueuesInput{})

	for paginator.HasMorePages() {
		output, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list queues: %w", err)
		}
		queueURLs = append(queueURLs, output.QueueUrls...)
	}

	return queueURLs, nil
}

// getQueueAttributes retrieves the attributes for a specific SQS queue
func (s *SQSInspector) getQueueAttributes(ctx context.Context, client *sqs.Client, queueURL string) (map[string]string, error) {
	// Define the attributes we want to retrieve
	attributeNames := []types.QueueAttributeName{
		types.QueueAttributeNameVisibilityTimeout,
		types.QueueAttributeNameDelaySeconds,
		types.QueueAttributeNameFifoQueue,
		types.QueueAttributeNameQueueArn,
	}

	// Get queue attributes
	result, err := client.GetQueueAttributes(ctx, &sqs.GetQueueAttributesInput{
		QueueUrl:       aws.String(queueURL),
		AttributeNames: attributeNames,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve queue attributes: %w", err)
	}

	// Convert attributes to map[string]string for consistency
	attributes := make(map[string]string)
	for key, value := range result.Attributes {
		attributes[string(key)] = value
	}

	return attributes, nil
}

// getQueueTags retrieves the tags for a specific SQS queue
func (s *SQSInspector) getQueueTags(ctx context.Context, client *sqs.Client, queueURL string) (map[string]string, error) {
	// List tags for the queue
	tagsResult, err := client.ListQueueTags(ctx, &sqs.ListQueueTagsInput{
		QueueUrl: aws.String(queueURL),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list queue tags: %w", err)
	}

	return tagsResult.Tags, nil
}

// getQueueName extracts the queue name from the queue URL
func (s *SQSInspector) getQueueName(queueURL string) string {
	// Queue URL format: https://sqs.region.amazonaws.com/account-id/queue-name
	parts := strings.Split(queueURL, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return ""
}

// Fetch implements the Scanner interface for retrieving specific SQS queue details
func (s *SQSInspector) Fetch(ctx context.Context, arn string, config configuration.TaggyScanConfig) (*ResourceMetadata, error) {
	// Parse queue ARN
	queueName, region, err := ParseSQSARN(arn)
	if err != nil {
		return nil, fmt.Errorf("failed to parse SQS ARN: %w", err)
	}

	// Get SQS client for the queue's region
	sqsClient, err := s.ClientManager.GetSQSClient(region)
	if err != nil {
		return nil, fmt.Errorf("failed to create SQS client: %w", err)
	}

	// Get queue URL from ARN
	queueURL, err := s.getQueueURLFromARN(ctx, sqsClient, queueName)
	if err != nil {
		return nil, fmt.Errorf("failed to get queue URL: %w", err)
	}

	// Get queue attributes
	attributes, err := s.getQueueAttributes(ctx, sqsClient, queueURL)
	if err != nil {
		return nil, fmt.Errorf("failed to get queue attributes: %w", err)
	}

	// Get queue tags
	tags, err := s.getQueueTags(ctx, sqsClient, queueURL)
	if err != nil {
		s.Logger.Warn("Failed to get queue tags",
			"queue_url", queueURL,
			"error", err)
		tags = make(map[string]string)
	}

	// Create resource metadata
	resourceMeta := &ResourceMetadata{
		ID:           arn,
		Type:         "sqs",
		Provider:     "aws",
		Region:       region,
		Tags:         tags,
		DiscoveredAt: time.Now(),
		RawResponse:  attributes,
	}

	// Populate extended details
	resourceMeta.Details.ARN = arn
	resourceMeta.Details.Name = s.getQueueName(queueURL)
	resourceMeta.Details.Properties = map[string]interface{}{
		"queue_url":          queueURL,
		"queue_arn":          attributes["QueueArn"],
		"visibility_timeout": attributes["VisibilityTimeout"],
		"delay_seconds":      attributes["DelaySeconds"],
		"queue_type":         attributes["FifoQueue"],
	}

	return resourceMeta, nil
}

// ParseSQSARN extracts queue ARN and region from SQS ARN
func ParseSQSARN(arn string) (string, string, error) {
	// ARN format: arn:aws:sqs:region:account-id:queue-name
	parts := strings.Split(arn, ":")
	if len(parts) != 6 {
		return "", "", fmt.Errorf("invalid SQS ARN format: %s", arn)
	}
	region := parts[3]
	queueName := parts[5]
	return queueName, region, nil
}

// getQueueURLFromARN retrieves the queue URL using the ARN
func (s *SQSInspector) getQueueURLFromARN(ctx context.Context, client *sqs.Client, queueName string) (string, error) {
	// Get queue URL
	result, err := client.GetQueueUrl(ctx, &sqs.GetQueueUrlInput{
		QueueName: aws.String(queueName),
	})
	if err != nil {
		return "", fmt.Errorf("failed to get queue URL for queue %s: %w", queueName, err)
	}

	return *result.QueueUrl, nil
}
