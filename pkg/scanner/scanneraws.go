package scanner

import (
	"context"
	"fmt"

	"github.com/Excoriate/aws-taggy/pkg/cloud"
	"github.com/Excoriate/aws-taggy/pkg/taggy"
	"github.com/aws/aws-sdk-go-v2/aws"
)

type ScannerAWS struct {
	TaggyClient *taggy.TaggyClient
	AWSRegion   string
	awsClient   *aws.Config
}

// NewScannerAWS creates a new AWS scanner instance, which is responsible for scanning and managing
// AWS resources with tagging capabilities. This constructor initializes the scanner with the necessary
// AWS configuration, client, and tagging client.
//
// Parameters:
//   - ctx: A context.Context for managing request cancellation, timeouts, and passing request-scoped values.
//   - tg: A pointer to the TaggyClient, which provides tagging-related functionality.
//   - awsRegion: The AWS region where the scanner will operate.
//
// Returns:
//   - *ScannerAWS: A configured AWS scanner instance.
//   - error: An error if AWS configuration or client initialization fails.
//
// The function performs the following steps:
//   1. Loads AWS client configuration from environment variables
//   2. Creates an AWS SDK configuration
//   3. Initializes a ScannerAWS struct with the provided parameters
//
// Example usage:
//   ctx := context.Background()
//   taggyClient := taggy.NewTaggyClient()
//   scanner, err := NewScannerAWS(ctx, taggyClient, "us-west-2")
func NewScannerAWS(ctx context.Context, tg *taggy.TaggyClient, awsRegion string) (*ScannerAWS, error) {
	aCfg := cloud.NewAWSClientConfigFromEnv(awsRegion)
	sdkCfg, err := cloud.NewAWSClient(ctx, aCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS configuration: %w", err)
	}

	sc := &ScannerAWS{
		TaggyClient: tg,
		AWSRegion:   awsRegion,
		awsClient:   sdkCfg,
	}

	return sc, nil
}
