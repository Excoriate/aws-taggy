package inspector

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Excoriate/aws-taggy/pkg/configuration"
	"github.com/Excoriate/aws-taggy/pkg/o11y"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/route53"
	"github.com/aws/aws-sdk-go-v2/service/route53/types"
)

// Route53ClientCreator implements AWSClient for Route 53
type Route53ClientCreator struct{}

func (c *Route53ClientCreator) CreateFromConfig(cfg *aws.Config) interface{} {
	return route53.NewFromConfig(*cfg)
}

// GetRoute53Client retrieves a Route 53 client for the specified AWS region.
//
// This method creates or retrieves an existing Route 53 client configuration for the given region.
// It uses the AWSClientManager's internal client management to ensure efficient client reuse.
//
// Parameters:
//   - region: The AWS region for which to create or retrieve the Route 53 client
//
// Returns:
//   - *route53.Client: A configured AWS Route 53 client
//   - error: An error if client creation fails
func (m *AWSClientManager) GetRoute53Client(region string) (*route53.Client, error) {
	client, err := m.GetClient(region, &Route53ClientCreator{})
	if err != nil {
		return nil, err
	}
	return client.(*route53.Client), nil
}

// Route53Inspector implements the Inspector interface for AWS Route 53 resources
type Route53Inspector struct {
	Regions       []string
	ClientManager *AWSClientManager
	Logger        *o11y.Logger
}

// NewRoute53Inspector creates a new inspector with AWS client management
func NewRoute53Inspector(regions []string) (*Route53Inspector, error) {
	// Create AWS client manager for the specified regions
	clientManager, err := NewAWSRegionalClientManager(regions)
	if err != nil {
		return nil, fmt.Errorf("failed to create AWS client manager: %w", err)
	}

	// Create a default logger
	logger := o11y.DefaultLogger()

	return &Route53Inspector{
		Regions:       regions,
		ClientManager: clientManager,
		Logger:        logger,
	}, nil
}

// Inspect discovers Route 53 hosted zones and their metadata across specified regions
func (r *Route53Inspector) Inspect(ctx context.Context, config configuration.TaggyScanConfig) (*InspectResult, error) {
	r.Logger.Info("Starting Route 53 resource scanning",
		"regions", r.Regions)

	result := &InspectResult{
		StartTime: time.Now(),
		Region:    r.Regions[0],
	}

	// Create async scanner with default config
	scanner := NewAsyncResourceInspector(DefaultInspectorConfig())

	// Define the resource discoverer function
	discoverer := func(ctx context.Context, region string) ([]interface{}, error) {
		// Get Route 53 client for this region
		route53Client, err := r.ClientManager.GetRoute53Client(region)
		if err != nil {
			return nil, fmt.Errorf("failed to get Route 53 client: %w", err)
		}

		// List hosted zones
		hostedZones, err := r.listHostedZones(ctx, route53Client)
		if err != nil {
			return nil, fmt.Errorf("failed to list hosted zones: %w", err)
		}

		// Convert to interface slice
		resources := make([]interface{}, len(hostedZones))
		for i, zone := range hostedZones {
			resources[i] = zone
		}

		return resources, nil
	}

	// Define the resource processor function
	processor := func(ctx context.Context, resource interface{}) (ResourceMetadata, error) {
		hostedZone := resource.(types.HostedZone)

		// Get Route 53 client for initial region
		route53Client, err := r.ClientManager.GetRoute53Client(r.Regions[0])
		if err != nil {
			return ResourceMetadata{}, fmt.Errorf("failed to get Route 53 client: %w", err)
		}

		// Fetch hosted zone tags
		tags, err := r.getHostedZoneTags(ctx, route53Client, *hostedZone.Id)
		if err != nil {
			r.Logger.Warn("Failed to get hosted zone tags",
				"zone_id", *hostedZone.Id,
				"error", err)
			tags = make(map[string]string)
		}

		// Create resource metadata
		metadata := ResourceMetadata{
			ID:           *hostedZone.Id,
			Type:         "route53_hosted_zone",
			Provider:     "aws",
			Region:       r.Regions[0], // Route 53 is a global service
			DiscoveredAt: time.Now(),
			Tags:         tags,
			RawResponse:  hostedZone,
		}

		// Populate extended details
		metadata.Details.ARN = fmt.Sprintf("arn:aws:route53:::hostedzone/%s", *hostedZone.Id)
		metadata.Details.Name = *hostedZone.Name
		metadata.Details.Properties = map[string]interface{}{
			"caller_reference": hostedZone.CallerReference,
			"config":           hostedZone.Config,
		}

		return metadata, nil
	}

	// Perform the async scan
	resources, err := scanner.InspectResourcesAsync(ctx, r.Regions, discoverer, processor)
	if err != nil {
		return nil, fmt.Errorf("failed to scan Route 53 resources: %w", err)
	}

	// Update result with scanned resources
	result.Resources = resources
	result.TotalResources = len(resources)
	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)

	r.Logger.Info("Route 53 scanning completed",
		"total_resources", result.TotalResources,
		"duration", result.Duration)

	return result, nil
}

// listHostedZones retrieves all Route 53 hosted zones
func (r *Route53Inspector) listHostedZones(ctx context.Context, client *route53.Client) ([]types.HostedZone, error) {
	var hostedZones []types.HostedZone
	paginator := route53.NewListHostedZonesPaginator(client, &route53.ListHostedZonesInput{})

	for paginator.HasMorePages() {
		output, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list hosted zones: %w", err)
		}
		hostedZones = append(hostedZones, output.HostedZones...)
	}

	return hostedZones, nil
}

// getHostedZoneTags retrieves tags for a specific hosted zone
func (r *Route53Inspector) getHostedZoneTags(ctx context.Context, client *route53.Client, hostedZoneID string) (map[string]string, error) {
	// List tags for the hosted zone
	tagsOutput, err := client.ListTagsForResource(ctx, &route53.ListTagsForResourceInput{
		ResourceId:   aws.String(hostedZoneID),
		ResourceType: types.TagResourceTypeHostedzone,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get hosted zone tags: %w", err)
	}

	tags := make(map[string]string)
	for _, tag := range tagsOutput.ResourceTagSet.Tags {
		tags[*tag.Key] = *tag.Value
	}

	return tags, nil
}

// Fetch implements the Scanner interface for retrieving specific Route 53 hosted zone details
func (r *Route53Inspector) Fetch(ctx context.Context, arn string, config configuration.TaggyScanConfig) (*ResourceMetadata, error) {
	// Parse hosted zone ID from ARN
	hostedZoneID, err := ParseRoute53ARN(arn)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Route 53 ARN: %w", err)
	}

	// Get Route 53 client (global service)
	route53Client, err := r.ClientManager.GetRoute53Client(r.Regions[0])
	if err != nil {
		return nil, fmt.Errorf("failed to create Route 53 client: %w", err)
	}

	// Get hosted zone details
	zoneOutput, err := route53Client.GetHostedZone(ctx, &route53.GetHostedZoneInput{
		Id: aws.String(hostedZoneID),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get hosted zone details: %w", err)
	}

	// Get hosted zone tags
	tags, err := r.getHostedZoneTags(ctx, route53Client, hostedZoneID)
	if err != nil {
		r.Logger.Warn("Failed to get hosted zone tags", "zone_id", hostedZoneID, "error", err)
		tags = make(map[string]string)
	}

	// Create resource metadata
	resourceMeta := &ResourceMetadata{
		ID:           hostedZoneID,
		Type:         "route53_hosted_zone",
		Provider:     "aws",
		Region:       r.Regions[0], // Route 53 is a global service
		Tags:         tags,
		DiscoveredAt: time.Now(),
	}

	// Populate extended details
	resourceMeta.Details.ARN = arn
	resourceMeta.Details.Name = *zoneOutput.HostedZone.Name
	resourceMeta.Details.Properties = map[string]interface{}{
		"caller_reference": zoneOutput.HostedZone.CallerReference,
		"config":           zoneOutput.HostedZone.Config,
	}

	return resourceMeta, nil
}

// ParseRoute53ARN extracts hosted zone ID from Route 53 ARN
func ParseRoute53ARN(arn string) (string, error) {
	// ARN format: arn:aws:route53:::hostedzone/ZONEID
	parts := strings.Split(arn, "/")
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid Route 53 ARN format: %s", arn)
	}
	return parts[1], nil
}
