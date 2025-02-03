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

// VPCInspector implements the Inspector interface for AWS VPC resources
type VPCInspector struct {
	Regions       []string
	ClientManager *AWSClientManager
	Logger        *o11y.Logger
}

// NewVPCInspector creates a new VPCInspector with AWS client management
func NewVPCInspector(regions []string) (*VPCInspector, error) {
	clientManager, err := NewAWSRegionalClientManager(regions)
	if err != nil {
		return nil, fmt.Errorf("failed to create AWS client manager: %w", err)
	}

	// Create a default logger
	logger := o11y.DefaultLogger()

	return &VPCInspector{
		Regions:       regions,
		ClientManager: clientManager,
		Logger:        logger,
	}, nil
}

// Inspect discovers VPCs and their metadata across specified regions
func (s *VPCInspector) Inspect(ctx context.Context, config configuration.TaggyScanConfig) (*InspectResult, error) {
	s.Logger.Info("Starting VPC resource scanning",
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

		// List VPCs
		vpcs, err := s.listVPCs(ctx, ec2Client)
		if err != nil {
			return nil, fmt.Errorf("failed to list VPCs: %w", err)
		}

		// Convert to interface slice
		resources := make([]interface{}, len(vpcs))
		for i, vpc := range vpcs {
			resources[i] = vpc
		}

		return resources, nil
	}

	// Define the resource processor function
	processor := func(ctx context.Context, resource interface{}) (ResourceMetadata, error) {
		vpc := resource.(types.Vpc)

		// Get VPC tags
		tags := make(map[string]string)
		for _, tag := range vpc.Tags {
			tags[aws.ToString(tag.Key)] = aws.ToString(tag.Value)
		}

		// Create resource metadata
		metadata := ResourceMetadata{
			ID:           aws.ToString(vpc.VpcId),
			Type:         "vpc",
			Provider:     "aws",
			Region:       s.Regions[0], // VPCs are region-specific
			DiscoveredAt: time.Now(),
			Tags:         tags,
			RawResponse:  vpc,
		}

		// Populate extended details
		metadata.Details.ARN = fmt.Sprintf("arn:aws:ec2:%s:%s:vpc/%s",
			s.Regions[0], "unknown", aws.ToString(vpc.VpcId)) // VPCs don't have direct OwnerId
		metadata.Details.Name = s.getVPCName(vpc)
		metadata.Details.Status = s.getVPCStatus(vpc)
		metadata.Details.Properties = map[string]interface{}{
			"cidr_block": aws.ToString(vpc.CidrBlock),
			"is_default": vpc.IsDefault,
			"state":      vpc.State,
		}

		return metadata, nil
	}

	// Perform the async scan
	resources, err := scanner.InspectResourcesAsync(ctx, s.Regions, discoverer, processor)
	if err != nil {
		return nil, fmt.Errorf("failed to scan VPC resources: %w", err)
	}

	// Update result with scanned resources
	result.Resources = resources
	result.TotalResources = len(resources)
	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)

	s.Logger.Info("VPC scanning completed",
		"total_resources", result.TotalResources,
		"duration", result.Duration)

	return result, nil
}

// listVPCs retrieves all VPCs in a region
func (s *VPCInspector) listVPCs(ctx context.Context, client *ec2.Client) ([]types.Vpc, error) {
	input := &ec2.DescribeVpcsInput{}
	output, err := client.DescribeVpcs(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to list VPCs: %w", err)
	}

	return output.Vpcs, nil
}

// getVPCName extracts the Name tag or returns a default name
func (s *VPCInspector) getVPCName(vpc types.Vpc) string {
	for _, tag := range vpc.Tags {
		if aws.ToString(tag.Key) == "Name" {
			return aws.ToString(tag.Value)
		}
	}
	if vpc.VpcId != nil {
		return aws.ToString(vpc.VpcId)
	}
	return "Unnamed VPC"
}

// getVPCStatus determines the VPC status
func (s *VPCInspector) getVPCStatus(vpc types.Vpc) string {
	return string(vpc.State)
}

// Fetch implements the Inspector interface for retrieving specific VPC details
func (s *VPCInspector) Fetch(ctx context.Context, arn string, config configuration.TaggyScanConfig) (*ResourceMetadata, error) {
	// Parse VPC ID and region from ARN
	vpcID, region, err := ParseVPCARN(arn)
	if err != nil {
		return nil, fmt.Errorf("failed to parse VPC ARN: %w", err)
	}

	// Get EC2 client for the VPC's region
	ec2Client, err := s.ClientManager.GetEC2Client(region)
	if err != nil {
		return nil, fmt.Errorf("failed to create EC2 client: %w", err)
	}

	// Describe the specific VPC
	input := &ec2.DescribeVpcsInput{
		VpcIds: []string{vpcID},
	}
	output, err := ec2Client.DescribeVpcs(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch VPC: %w", err)
	}

	// Ensure we have a VPC
	if len(output.Vpcs) == 0 {
		return nil, fmt.Errorf("no VPC found with ID %s", vpcID)
	}

	vpc := output.Vpcs[0]

	// Get VPC tags
	tags := make(map[string]string)
	for _, tag := range vpc.Tags {
		tags[aws.ToString(tag.Key)] = aws.ToString(tag.Value)
	}

	// Create resource metadata
	resourceMeta := &ResourceMetadata{
		ID:   vpcID,
		Type: "vpc",

		Provider:     "aws",
		Region:       region,
		Tags:         tags,
		DiscoveredAt: time.Now(),
	}

	// Populate extended details
	resourceMeta.Details.ARN = arn
	resourceMeta.Details.Name = s.getVPCName(vpc)
	resourceMeta.Details.Status = s.getVPCStatus(vpc)
	resourceMeta.Details.Properties = map[string]interface{}{
		"cidr_block": aws.ToString(vpc.CidrBlock),
		"is_default": vpc.IsDefault,
		"state":      vpc.State,
	}

	return resourceMeta, nil
}

// ParseVPCARN extracts VPC ID and region from VPC ARN
func ParseVPCARN(arn string) (string, string, error) {
	// ARN format: arn:aws:ec2:region:account-id:vpc/vpc-id
	parts := strings.Split(arn, ":")
	if len(parts) != 6 {
		return "", "", fmt.Errorf("invalid VPC ARN format: %s", arn)
	}
	region := parts[3]
	vpcParts := strings.Split(parts[5], "/")
	if len(vpcParts) != 2 {
		return "", "", fmt.Errorf("invalid VPC ID format in ARN: %s", arn)
	}
	return vpcParts[1], region, nil
}
