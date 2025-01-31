package inspector

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Excoriate/aws-taggy/pkg/configuration"
	"github.com/Excoriate/aws-taggy/pkg/o11y"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/aws/aws-sdk-go-v2/service/sns/types"
)

// SNSClientCreator implements AWSClient for SNS
type SNSClientCreator struct{}

func (c *SNSClientCreator) CreateFromConfig(cfg *aws.Config) interface{} {
	return sns.NewFromConfig(*cfg)
}

// GetSNSClient retrieves an SNS client for the specified AWS region.
//
// This method creates or retrieves an existing SNS client configuration for the given region.
// It uses the AWSClientManager's internal client management to ensure efficient client reuse.
//
// Parameters:
//   - region: The AWS region for which to create or retrieve the SNS client
//
// Returns:
//   - *sns.Client: A configured AWS SNS client
//   - error: An error if client creation fails
func (m *AWSClientManager) GetSNSClient(region string) (*sns.Client, error) {
	client, err := m.GetClient(region, &SNSClientCreator{})
	if err != nil {
		return nil, err
	}
	return client.(*sns.Client), nil
}

// SNSInspector implements the Inspector interface for AWS SNS resources
type SNSInspector struct {
	Regions       []string
	ClientManager *AWSClientManager
	Logger        *o11y.Logger
}

// NewSNSInspector creates a new inspector with AWS client management
func NewSNSInspector(regions []string) (*SNSInspector, error) {
	// Create AWS client manager for the specified regions
	clientManager, err := NewAWSRegionalClientManager(regions)
	if err != nil {
		return nil, fmt.Errorf("failed to create AWS client manager: %w", err)
	}

	// Create a default logger
	logger := o11y.DefaultLogger()

	return &SNSInspector{
		Regions:       regions,
		ClientManager: clientManager,
		Logger:        logger,
	}, nil
}

// Inspect discovers SNS topics and their metadata across specified regions
func (s *SNSInspector) Inspect(ctx context.Context, config configuration.TaggyScanConfig) (*InspectResult, error) {
	s.Logger.Info("Starting SNS resource scanning",
		"regions", s.Regions)

	result := &InspectResult{
		StartTime: time.Now(),
		Region:    s.Regions[0],
	}

	// Create async scanner with default config
	scanner := NewAsyncResourceInspector(DefaultInspectorConfig())

	// Define the resource discoverer function
	discoverer := func(ctx context.Context, region string) ([]interface{}, error) {
		// Get SNS client for this region
		snsClient, err := s.ClientManager.GetSNSClient(region)
		if err != nil {
			return nil, fmt.Errorf("failed to get SNS client: %w", err)
		}

		// List topics
		topics, err := s.listTopics(ctx, snsClient)
		if err != nil {
			return nil, fmt.Errorf("failed to list topics: %w", err)
		}

		// Convert to interface slice
		resources := make([]interface{}, len(topics))
		for i, topic := range topics {
			resources[i] = topic
		}

		return resources, nil
	}

	// Define the resource processor function
	processor := func(ctx context.Context, resource interface{}) (ResourceMetadata, error) {
		topic := resource.(types.Topic)

		// Get SNS client for initial region
		snsClient, err := s.ClientManager.GetSNSClient(s.Regions[0])
		if err != nil {
			return ResourceMetadata{}, fmt.Errorf("failed to get SNS client: %w", err)
		}

		// Fetch topic tags
		tags, err := s.getTopicTags(ctx, snsClient, *topic.TopicArn)
		if err != nil {
			s.Logger.Warn("Failed to get topic tags",
				"topic_arn", *topic.TopicArn,
				"error", err)
			tags = make(map[string]string)
		}

		// Create resource metadata
		metadata := ResourceMetadata{
			ID:           *topic.TopicArn,
			Type:         "sns",
			Provider:     "aws",
			Region:       s.Regions[0], // SNS is regional
			DiscoveredAt: time.Now(),
			Tags:         tags,
			RawResponse:  topic,
		}

		// Populate extended details
		metadata.Details.ARN = *topic.TopicArn
		metadata.Details.Name = s.getTopicName(*topic.TopicArn)
		metadata.Details.Properties = map[string]interface{}{
			"topic_arn": *topic.TopicArn,
		}

		return metadata, nil
	}

	// Perform the async scan
	resources, err := scanner.InspectResourcesAsync(ctx, s.Regions, discoverer, processor)
	if err != nil {
		return nil, fmt.Errorf("failed to scan SNS resources: %w", err)
	}

	// Update result with scanned resources
	result.Resources = resources
	result.TotalResources = len(resources)
	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)

	s.Logger.Info("SNS scanning completed",
		"total_resources", result.TotalResources,
		"duration", result.Duration)

	return result, nil
}

// listTopics retrieves all SNS topics
func (s *SNSInspector) listTopics(ctx context.Context, client *sns.Client) ([]types.Topic, error) {
	var topics []types.Topic
	paginator := sns.NewListTopicsPaginator(client, &sns.ListTopicsInput{})

	for paginator.HasMorePages() {
		output, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list topics: %w", err)
		}
		topics = append(topics, output.Topics...)
	}

	return topics, nil
}

// getTopicTags retrieves tags for a specific SNS topic
func (s *SNSInspector) getTopicTags(ctx context.Context, client *sns.Client, topicARN string) (map[string]string, error) {
	// List tags for the topic
	tagsOutput, err := client.ListTagsForResource(ctx, &sns.ListTagsForResourceInput{
		ResourceArn: aws.String(topicARN),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get topic tags: %w", err)
	}

	tags := make(map[string]string)
	for _, tag := range tagsOutput.Tags {
		tags[*tag.Key] = *tag.Value
	}

	return tags, nil
}

// getTopicName extracts the topic name from its ARN
func (s *SNSInspector) getTopicName(topicARN string) string {
	parts := strings.Split(topicARN, ":")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return "Unnamed Topic"
}

// Fetch implements the Scanner interface for retrieving specific SNS topic details
func (s *SNSInspector) Fetch(ctx context.Context, arn string, config configuration.TaggyScanConfig) (*ResourceMetadata, error) {
	// Parse topic ARN
	topicARN, region, err := ParseSNSARN(arn)
	if err != nil {
		return nil, fmt.Errorf("failed to parse SNS ARN: %w", err)
	}

	// Get SNS client for the topic's region
	snsClient, err := s.ClientManager.GetSNSClient(region)
	if err != nil {
		return nil, fmt.Errorf("failed to create SNS client: %w", err)
	}

	// Get topic tags
	tags, err := s.getTopicTags(ctx, snsClient, topicARN)
	if err != nil {
		s.Logger.Warn("Failed to get topic tags", "topic_arn", topicARN, "error", err)
		tags = make(map[string]string)
	}

	// Create resource metadata
	resourceMeta := &ResourceMetadata{
		ID:           topicARN,
		Type:         "sns",
		Provider:     "aws",
		Region:       region,
		Tags:         tags,
		DiscoveredAt: time.Now(),
	}

	// Populate extended details
	resourceMeta.Details.ARN = arn
	resourceMeta.Details.Name = s.getTopicName(topicARN)
	resourceMeta.Details.Properties = map[string]interface{}{
		"topic_arn": topicARN,
	}

	return resourceMeta, nil
}

// ParseSNSARN extracts topic ARN and region from SNS ARN
func ParseSNSARN(arn string) (string, string, error) {
	// ARN format: arn:aws:sns:region:account-id:topic-name
	parts := strings.Split(arn, ":")
	if len(parts) != 6 {
		return "", "", fmt.Errorf("invalid SNS ARN format: %s", arn)
	}
	region := parts[3]
	return arn, region, nil
}
