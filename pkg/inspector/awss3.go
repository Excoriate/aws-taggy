package inspector

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Excoriate/aws-taggy/pkg/configuration"
	"github.com/Excoriate/aws-taggy/pkg/o11y"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

// S3ClientCreator implements AWSClient for S3
type S3ClientCreator struct{}

func (c *S3ClientCreator) CreateFromConfig(cfg *aws.Config) interface{} {
	return s3.NewFromConfig(*cfg)
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

// S3Inspector implements the Scanner interface for AWS S3 resources
type S3Inspector struct {
	Regions       []string
	ClientManager *AWSClientManager
	Logger        *o11y.Logger
}

// NewS3Inspector creates a new S3Inspector with AWS client management
func NewS3Inspector(regions []string) (*S3Inspector, error) {
	// Create AWS client manager for the specified regions
	clientManager, err := NewAWSRegionalClientManager(regions)
	if err != nil {
		return nil, fmt.Errorf("failed to create AWS client manager: %w", err)
	}

	// Create a default logger
	logger := o11y.DefaultLogger()

	return &S3Inspector{
		Regions:       regions,
		ClientManager: clientManager,
		Logger:        logger,
	}, nil
}

// Inspect discovers S3 buckets and their metadata across specified regions
func (s *S3Inspector) Inspect(ctx context.Context, config configuration.TaggyScanConfig) (*InspectResult, error) {
	s.Logger.Info("Starting S3 resource scanning",
		"regions", s.Regions)

	result := &InspectResult{
		StartTime: time.Now(),
		Region:    s.Regions[0],
	}

	// Create async scanner with default config
	scanner := NewAsyncResourceInspector(DefaultInspectorConfig())

	// Define the resource discoverer function
	discoverer := func(ctx context.Context, region string) ([]interface{}, error) {
		// Get S3 client for this region
		s3Client, err := s.ClientManager.GetS3Client(region)
		if err != nil {
			return nil, fmt.Errorf("failed to get S3 client: %w", err)
		}

		// List buckets
		buckets, err := s.listBuckets(ctx, s3Client)
		if err != nil {
			return nil, fmt.Errorf("failed to list buckets: %w", err)
		}

		// Convert to interface slice
		resources := make([]interface{}, len(buckets))
		for i, bucket := range buckets {
			resources[i] = bucket
		}

		return resources, nil
	}

	// Define the resource processor function
	processor := func(ctx context.Context, resource interface{}) (ResourceMetadata, error) {
		bucket := resource.(types.Bucket)

		// Get S3 client for initial region
		s3Client, err := s.ClientManager.GetS3Client(s.Regions[0])
		if err != nil {
			return ResourceMetadata{}, fmt.Errorf("failed to get S3 client: %w", err)
		}

		// Get bucket location
		locationOutput, err := s3Client.GetBucketLocation(ctx, &s3.GetBucketLocationInput{
			Bucket: bucket.Name,
		})
		if err != nil {
			return ResourceMetadata{}, fmt.Errorf("failed to get bucket location: %w", err)
		}

		// Determine bucket region
		bucketRegion := string(locationOutput.LocationConstraint)
		if bucketRegion == "" {
			bucketRegion = "us-east-1"
		}

		// Get client for correct region if different
		if bucketRegion != s.Regions[0] {
			s3Client, err = s.ClientManager.GetS3Client(bucketRegion)
			if err != nil {
				return ResourceMetadata{}, fmt.Errorf("failed to get region-specific S3 client: %w", err)
			}
		}

		// Fetch bucket tags
		tags, err := s.getBucketTags(ctx, s3Client, *bucket.Name)
		if err != nil {
			s.Logger.Warn("Failed to get bucket tags",
				"bucket", *bucket.Name,
				"error", err)
			tags = make(map[string]string)
		}

		// Create resource metadata
		metadata := ResourceMetadata{
			ID:           *bucket.Name,
			Type:         "s3",
			Provider:     "aws",
			Region:       bucketRegion,
			DiscoveredAt: time.Now(),
			Tags:         tags,
			RawResponse:  bucket,
		}

		// Populate extended details
		metadata.Details.ARN = fmt.Sprintf("arn:aws:s3:::%s", *bucket.Name)
		metadata.Details.Name = *bucket.Name
		metadata.Details.Properties = map[string]interface{}{
			"creation_date": bucket.CreationDate,
			"region":        bucketRegion,
		}

		return metadata, nil
	}

	// Perform the async scan
	resources, err := scanner.InspectResourcesAsync(ctx, s.Regions, discoverer, processor)
	if err != nil {
		return nil, fmt.Errorf("failed to scan S3 resources: %w", err)
	}

	// Update result with scanned resources
	result.Resources = resources
	result.TotalResources = len(resources)
	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)

	s.Logger.Info("S3 scanning completed",
		"total_resources", result.TotalResources,
		"duration", result.Duration)

	return result, nil
}

// listBuckets retrieves all S3 buckets
func (s *S3Inspector) listBuckets(ctx context.Context, client *s3.Client) ([]types.Bucket, error) {
	output, err := client.ListBuckets(ctx, &s3.ListBucketsInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to list buckets: %w", err)
	}
	return output.Buckets, nil
}

// getBucketTags retrieves tags for a specific bucket
func (s *S3Inspector) getBucketTags(ctx context.Context, client *s3.Client, bucketName string) (map[string]string, error) {
	// First, try to get the bucket location
	locationOutput, err := client.GetBucketLocation(ctx, &s3.GetBucketLocationInput{
		Bucket: aws.String(bucketName),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get bucket location: %w", err)
	}

	// If the bucket is in a different region, create a new client for that region
	bucketRegion := string(locationOutput.LocationConstraint)
	if bucketRegion == "" {
		bucketRegion = "us-east-1" // Default region for buckets without explicit location
	}

	// If the bucket is in a different region, create a new client
	if bucketRegion != s.Regions[0] {
		s.Logger.Debug("Bucket in different region",
			"bucket", bucketName,
			"detected_region", bucketRegion)

		// Create a new S3 client for the specific region
		regionClient, err := s.ClientManager.GetS3Client(bucketRegion)
		if err != nil {
			return nil, fmt.Errorf("failed to create client for bucket region %s: %w", bucketRegion, err)
		}
		client = regionClient
	}

	// Attempt to get bucket tags
	tagsOutput, err := client.GetBucketTagging(ctx, &s3.GetBucketTaggingInput{
		Bucket: aws.String(bucketName),
	})
	if err != nil {
		// Check if the error indicates a permanent redirect
		if strings.Contains(err.Error(), "PermanentRedirect") {
			s.Logger.Warn("Permanent redirect error",
				"bucket", bucketName,
				"error", err)
			return nil, fmt.Errorf("bucket requires specific endpoint: %w", err)
		}
		// If NoSuchTagSet, return empty tags map (bucket exists but has no tags)
		if strings.Contains(err.Error(), "NoSuchTagSet") {
			s.Logger.Debug("No tags found for bucket",
				"bucket", bucketName)
			return make(map[string]string), nil
		}
		return nil, fmt.Errorf("failed to get bucket tags: %w", err)
	}

	tags := make(map[string]string)
	for _, tag := range tagsOutput.TagSet {
		tags[*tag.Key] = *tag.Value
	}

	return tags, nil
}

// Fetch implements the Scanner interface for retrieving specific S3 bucket details
func (s *S3Inspector) Fetch(ctx context.Context, arn string, config configuration.TaggyScanConfig) (*ResourceMetadata, error) {
	// Parse bucket name from ARN
	bucketName, err := ParseS3ARN(arn)
	if err != nil {
		return nil, fmt.Errorf("failed to parse S3 ARN: %w", err)
	}

	// Get the bucket's region first
	s3Client, err := s.ClientManager.GetS3Client("us-east-1") // Start with default region
	if err != nil {
		return nil, fmt.Errorf("failed to create S3 client: %w", err)
	}

	// Get bucket location
	locationOutput, err := s3Client.GetBucketLocation(ctx, &s3.GetBucketLocationInput{
		Bucket: aws.String(bucketName),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get bucket location: %w", err)
	}

	// Get the correct region
	bucketRegion := string(locationOutput.LocationConstraint)
	if bucketRegion == "" {
		bucketRegion = "us-east-1"
	}

	// Get client for the correct region if different
	if bucketRegion != "us-east-1" {
		s3Client, err = s.ClientManager.GetS3Client(bucketRegion)
		if err != nil {
			return nil, fmt.Errorf("failed to create S3 client for region %s: %w", bucketRegion, err)
		}
	}

	// Get bucket tags
	tags, err := s.getBucketTags(ctx, s3Client, bucketName)
	if err != nil {
		s.Logger.Warn("Failed to get bucket tags", "bucket", bucketName, "error", err)
		tags = make(map[string]string)
	}

	// Create resource metadata
	resourceMeta := &ResourceMetadata{
		ID:           bucketName,
		Type:         "s3",
		Provider:     "aws",
		Region:       bucketRegion,
		Tags:         tags,
		DiscoveredAt: time.Now(),
	}

	// Populate extended details
	resourceMeta.Details.ARN = arn
	resourceMeta.Details.Name = bucketName
	resourceMeta.Details.Properties = map[string]interface{}{
		"region": bucketRegion,
	}

	return resourceMeta, nil
}

// ParseS3ARN extracts bucket name from S3 ARN
func ParseS3ARN(arn string) (string, error) {
	// ARN format: arn:aws:s3:::bucket-name
	parts := strings.Split(arn, ":")
	if len(parts) != 6 {
		return "", fmt.Errorf("invalid S3 ARN format: %s", arn)
	}
	return parts[5], nil
}
