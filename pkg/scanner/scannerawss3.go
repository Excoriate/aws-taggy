package scanner

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/Excoriate/aws-taggy/pkg/configuration"
	"github.com/Excoriate/aws-taggy/pkg/o11y"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

// S3Scanner implements the Scanner interface for AWS S3 resources
type S3Scanner struct {
	Regions       []string
	ClientManager *AWSClientManager
	Logger        *o11y.Logger
}

// NewS3Scanner creates a new S3Scanner with AWS client management
func NewS3Scanner(regions []string) (*S3Scanner, error) {
	// Create AWS client manager for the specified regions
	clientManager, err := NewAWSClientManager(regions)
	if err != nil {
		return nil, fmt.Errorf("failed to create AWS client manager: %w", err)
	}

	// Create a default logger
	logger := o11y.DefaultLogger()

	return &S3Scanner{
		Regions:       regions,
		ClientManager: clientManager,
		Logger:        logger,
	}, nil
}

// Scan discovers S3 buckets and their metadata across specified regions
func (s *S3Scanner) Scan(ctx context.Context, resource Resource, config configuration.TaggyScanConfig) (*ScanResult, error) {
	s.Logger.Info("Starting S3 resource scanning",
		"resource_type", resource.GetType(),
		"regions", s.Regions)

	result := &ScanResult{
		StartTime: time.Now(),
		Region:    resource.GetRegion(),
	}

	// Scan buckets across all specified regions concurrently
	var scanWg sync.WaitGroup
	resultChan := make(chan []ResourceMetadata, len(s.Regions))
	errorChan := make(chan error, len(s.Regions))

	for _, region := range s.Regions {
		scanWg.Add(1)
		go func(r string) {
			defer scanWg.Done()

			s.Logger.Debug("Scanning S3 buckets in region", "region", r)

			// Get S3 client for this region
			s3Client, err := s.ClientManager.GetS3Client(r)
			if err != nil {
				errorMsg := fmt.Sprintf("failed to get S3 client for region %s", r)
				s.Logger.Error(errorMsg, "error", err)
				errorChan <- fmt.Errorf(errorMsg+": %w", err)
				return
			}

			// List buckets in this region
			buckets, err := s.listBuckets(ctx, s3Client)
			if err != nil {
				errorMsg := fmt.Sprintf("failed to list buckets in region %s", r)
				s.Logger.Error(errorMsg, "error", err)
				errorChan <- fmt.Errorf(errorMsg+": %w", err)
				return
			}

			s.Logger.Info("Discovered S3 buckets",
				"region", r,
				"bucket_count", len(buckets))

			// Convert buckets to ResourceMetadata
			regionResources := make([]ResourceMetadata, 0, len(buckets))
			for _, bucket := range buckets {
				// Fetch bucket tags first
				tags, err := s.getBucketTags(ctx, s3Client, *bucket.Name)
				if err != nil {
					errorMsg := fmt.Sprintf("failed to get tags for bucket %s", *bucket.Name)
					s.Logger.Warn(errorMsg, "error", err)
					result.Errors = append(result.Errors,
						fmt.Sprintf("%s in region %s: %v", errorMsg, r, err))
					tags = make(map[string]string) // Initialize empty tags map on error
				} else {
					s.Logger.Debug("Bucket tags retrieved",
						"bucket", *bucket.Name,
						"tag_count", len(tags))
				}

				resourceMeta := ResourceMetadata{
					ID:           *bucket.Name,
					Type:         "s3",
					Provider:     "aws",
					Region:       r,
					DiscoveredAt: time.Now(),
					Tags:         tags,
					RawResponse:  bucket,
				}

				// Populate extended details
				resourceMeta.Details.ARN = fmt.Sprintf("arn:aws:s3:::%s", *bucket.Name)
				resourceMeta.Details.Name = *bucket.Name
				resourceMeta.Details.Properties = map[string]interface{}{
					"creation_date": bucket.CreationDate,
				}

				regionResources = append(regionResources, resourceMeta)
			}

			resultChan <- regionResources
		}(region)
	}

	// Wait for all region scans to complete
	scanWg.Wait()
	close(resultChan)
	close(errorChan)

	// Collect results and errors
	var scanErrors []error
	for err := range errorChan {
		scanErrors = append(scanErrors, err)
	}

	for regionResources := range resultChan {
		result.Resources = append(result.Resources, regionResources...)
	}

	// Check for any errors
	if len(scanErrors) > 0 {
		var errorMessages []string
		for _, err := range scanErrors {
			errorMessages = append(errorMessages, err.Error())
		}

		s.Logger.Error("S3 scanning encountered errors",
			"error_count", len(scanErrors))

		return nil, fmt.Errorf("S3 scanning failed with %d errors:\n%s",
			len(scanErrors),
			strings.Join(errorMessages, "\n"))
	}

	result.TotalResources = len(result.Resources)
	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)

	s.Logger.Info("S3 scanning completed",
		"total_resources", result.TotalResources,
		"duration", result.Duration)

	return result, nil
}

// listBuckets retrieves all S3 buckets
func (s *S3Scanner) listBuckets(ctx context.Context, client *s3.Client) ([]types.Bucket, error) {
	output, err := client.ListBuckets(ctx, &s3.ListBucketsInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to list buckets: %w", err)
	}
	return output.Buckets, nil
}

// getBucketTags retrieves tags for a specific bucket
func (s *S3Scanner) getBucketTags(ctx context.Context, client *s3.Client, bucketName string) (map[string]string, error) {
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
		return nil, fmt.Errorf("failed to get bucket tags: %w", err)
	}

	tags := make(map[string]string)
	for _, tag := range tagsOutput.TagSet {
		tags[*tag.Key] = *tag.Value
	}

	return tags, nil
}
