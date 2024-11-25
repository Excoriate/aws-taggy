package cloud

import (
	"context"
	"fmt"

	"github.com/Excoriate/aws-taggy/pkg/util"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
)

// AWSClientConfig defines the configuration interface for AWS client creation
type AWSClientConfig interface {
	GetRegion() string
	GetAccessKeyID() string
	GetSecretAccessKey() string
	GetSessionToken() string
	Validate() error
	LoadConfig(ctx context.Context) (*aws.Config, error)
}

// AWSClientConfigOptions implements AWSClientConfig
type AWSClientConfigOptions struct {
	Region          string
	AccessKeyID     string
	SecretAccessKey string
	SessionToken    string
}

// Implement interface methods (existing methods remain the same)
func (c *AWSClientConfigOptions) GetRegion() string {
	return c.Region
}

func (c *AWSClientConfigOptions) GetAccessKeyID() string {
	return c.AccessKeyID
}

func (c *AWSClientConfigOptions) GetSecretAccessKey() string {
	return c.SecretAccessKey
}

func (c *AWSClientConfigOptions) GetSessionToken() string {
	return c.SessionToken
}

// Validate checks if the required AWS credentials are present
func (c *AWSClientConfigOptions) Validate() error {
	if c.Region == "" {
		region, err := util.GetAWSRegionEnvVar()
		if err != nil {
			return fmt.Errorf("region is required: %w", err)
		}
		c.Region = region
	}

	if c.AccessKeyID == "" {
		accessKeyID, err := util.GetAWSAccessKeyIDEnvVar()
		if err != nil {
			return fmt.Errorf("access key ID is required: %w", err)
		}
		c.AccessKeyID = accessKeyID
	}

	if c.SecretAccessKey == "" {
		secretAccessKey, err := util.GetAWSSecretAccessKeyEnvVar()
		if err != nil {
			return fmt.Errorf("secret access key is required: %w", err)
		}
		c.SecretAccessKey = secretAccessKey
	}

	return nil
}

// LoadConfig creates an AWS SDK configuration using static credentials or environment variables
func (c *AWSClientConfigOptions) LoadConfig(ctx context.Context) (*aws.Config, error) {
	// Validate credentials
	if err := c.Validate(); err != nil {
		return nil, fmt.Errorf("invalid AWS configuration: %w", err)
	}

	// Default to us-east-1 if no region specified
	if c.Region == "" {
		c.Region = "us-east-1"
	}

	// Create static credentials provider
	credProvider := credentials.NewStaticCredentialsProvider(
		c.AccessKeyID,
		c.SecretAccessKey,
		c.SessionToken,
	)

	// Load configuration with static credentials and specified region
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(c.Region),
		config.WithCredentialsProvider(credProvider),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS configuration: %w", err)
	}

	return &cfg, nil
}

// NewAWSClientConfig creates a new AWS client configuration
func NewAWSClientConfig(
	awsRegion string,
	accessKeyID string,
	secretAccessKey string,
	sessionToken string,
) AWSClientConfig {
	// Default to us-east-1 if no region provided
	if awsRegion == "" {
		awsRegion = "us-east-1"
	}

	return &AWSClientConfigOptions{
		Region:          awsRegion,
		AccessKeyID:     accessKeyID,
		SecretAccessKey: secretAccessKey,
		SessionToken:    sessionToken,
	}
}

// NewAWSClientConfigFromEnv creates an AWS client configuration using environment variables
func NewAWSClientConfigFromEnv(awsRegion string) AWSClientConfig {
	return NewAWSClientConfig(awsRegion, "", "", "")
}

// NewAWSClient is a convenience function to create an AWS configuration
func NewAWSClient(
	ctx context.Context,
	cfg AWSClientConfig,
) (*aws.Config, error) {
	// Load and return AWS configuration
	return cfg.LoadConfig(ctx)
} 