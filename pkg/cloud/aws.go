package cloud

import (
	"context"
	"fmt"

	"github.com/Excoriate/aws-taggy/pkg/util"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
)

// AWSClientConfig defines a comprehensive configuration interface for AWS client creation.
// It provides methods to retrieve and validate AWS credentials, region, and configuration.
// The interface is designed to be flexible, allowing for easy extension, dependency injection,
// and simplified mocking during testing scenarios.
//
// The interface supports multiple ways of retrieving AWS configuration:
// - Direct configuration through struct fields
// - Environment variable-based configuration
// - Validation of credentials
// - Loading AWS SDK configuration
//
// Implementations of this interface should handle various authentication scenarios,
// including static credentials, environment-based credentials, and AWS SDK default credential chains.
type AWSClientConfig interface {
	// GetRegion returns the explicitly configured AWS region for the client.
	// If no region is set, it may return an empty string.
	//
	// Returns:
	//   - A string representing the AWS region (e.g., "us-east-1")
	GetRegion() string

	// GetRegionFromEnv attempts to retrieve the AWS region from standard environment variables.
	// This method provides a fallback mechanism for region configuration.
	//
	// Returns:
	//   - The AWS region as a string if successfully retrieved
	//   - An error if the region cannot be found or read from the environment
	GetRegionFromEnv() (string, error)

	// GetAccessKeyID returns the explicitly configured AWS access key ID.
	// If no access key ID is set, it may return an empty string.
	//
	// Returns:
	//   - A string representing the AWS access key ID
	GetAccessKeyID() string

	// GetAccessKeyIDFromEnv attempts to retrieve the AWS access key ID from environment variables.
	// This method provides a fallback mechanism for access key ID configuration.
	//
	// Returns:
	//   - The AWS access key ID as a string if successfully retrieved
	//   - An error if the access key ID cannot be found or read from the environment
	GetAccessKeyIDFromEnv() (string, error)

	// GetSecretAccessKey returns the explicitly configured AWS secret access key.
	// If no secret access key is set, it may return an empty string.
	//
	// Returns:
	//   - A string representing the AWS secret access key
	GetSecretAccessKey() string

	// GetSecretAccessKeyFromEnv attempts to retrieve the AWS secret access key from environment variables.
	// This method provides a fallback mechanism for secret access key configuration.
	//
	// Returns:
	//   - The AWS secret access key as a string if successfully retrieved
	//   - An error if the secret access key cannot be found or read from the environment
	GetSecretAccessKeyFromEnv() (string, error)

	// GetSessionToken returns the explicitly configured AWS session token.
	// This is typically used for temporary or assumed role credentials.
	// If no session token is set, it may return an empty string.
	//
	// Returns:
	//   - A string representing the AWS session token
	GetSessionToken() string

	// Validate performs comprehensive validation of the AWS credentials and configuration.
	// It checks for the presence and format of required credentials and settings.
	//
	// Returns:
	//   - An error if validation fails, indicating specific configuration issues
	//   - nil if all configurations are valid and complete
	Validate() error

	// LoadConfig creates and returns an AWS SDK configuration using the current credentials and settings.
	// This method abstracts the complexity of AWS SDK configuration initialization.
	//
	// Parameters:
	//   - ctx: A context.Context for controlling the configuration loading process
	//
	// Returns:
	//   - A fully configured aws.Config instance
	//   - An error if configuration loading fails
	LoadConfig(ctx context.Context) (aws.Config, error)
}

// AWSClientConfigOptions is a concrete implementation of the AWSClientConfig interface.
// It provides a straightforward way to configure AWS client settings using struct fields.
//
// This struct allows for direct configuration of AWS credentials and region,
// supporting both static configuration and environment-variable-based approaches.
//
// Fields:
//   - Region: The AWS region for service interactions
//   - AccessKeyID: The AWS access key ID for authentication
//   - SecretAccessKey: The AWS secret access key for authentication
//   - SessionToken: Optional session token for temporary credentials
type AWSClientConfigOptions struct {
	Region          string
	AccessKeyID     string
	SecretAccessKey string
	SessionToken    string
}

// GetRegion returns the configured AWS region for the client.
// If no region is explicitly set, this method will return an empty string.
//
// Returns:
//   - The AWS region as a string, which can be empty if not configured.
func (c *AWSClientConfigOptions) GetRegion() string {
	return c.Region
}

// GetRegionFromEnv attempts to retrieve the AWS region from environment variables.
// It uses the utility function to extract the region from standard AWS environment variables.
//
// Returns:
//   - The AWS region as a string if successfully retrieved from environment.
//   - An error if the region cannot be found or read from the environment.
func (c *AWSClientConfigOptions) GetRegionFromEnv() (string, error) {
	return util.GetAWSRegionEnvVar()
}

// GetAccessKeyID returns the configured AWS access key ID for the client.
// If no access key ID is explicitly set, this method will return an empty string.
//
// Returns:
//   - The AWS access key ID as a string, which can be empty if not configured.
func (c *AWSClientConfigOptions) GetAccessKeyID() string {
	return c.AccessKeyID
}

// GetSecretAccessKey returns the configured AWS secret access key for the client.
// If no secret access key is explicitly set, this method will return an empty string.
//
// Returns:
//   - The AWS secret access key as a string, which can be empty if not configured.
func (c *AWSClientConfigOptions) GetSecretAccessKey() string {
	return c.SecretAccessKey
}

// GetSessionToken returns the configured AWS session token for the client.
// If no session token is explicitly set, this method will return an empty string.
//
// Returns:
//   - The AWS session token as a string, which can be empty if not configured.
func (c *AWSClientConfigOptions) GetSessionToken() string {
	return c.SessionToken
}

// GetAccessKeyIDFromEnv attempts to retrieve the AWS access key ID from environment variables.
// It uses the utility function to extract the access key ID from standard AWS environment variables.
//
// Returns:
//   - The AWS access key ID as a string if successfully retrieved from environment.
//   - An error if the access key ID cannot be found or read from the environment.
func (c *AWSClientConfigOptions) GetAccessKeyIDFromEnv() (string, error) {
	return util.GetAWSAccessKeyIDEnvVar()
}

// GetSecretAccessKeyFromEnv attempts to retrieve the AWS secret access key from environment variables.
// It uses the utility function to extract the secret access key from standard AWS environment variables.
//
// Returns:
//   - The AWS secret access key as a string if successfully retrieved from environment.
//   - An error if the secret access key cannot be found or read from the environment.
func (c *AWSClientConfigOptions) GetSecretAccessKeyFromEnv() (string, error) {
	return util.GetAWSSecretAccessKeyEnvVar()
}

// Validate checks if the required AWS credentials are present
// Validate checks the AWS client configuration for completeness and attempts to populate
// missing credentials from environment variables. This method ensures that all required
// AWS credentials (region, access key ID, and secret access key) are present.
//
// The validation process follows these steps:
//   1. If Region is not set, attempt to retrieve it from the AWS_REGION or AWS_DEFAULT_REGION environment variable
//   2. If AccessKeyID is not set, attempt to retrieve it from the AWS_ACCESS_KEY_ID environment variable
//   3. If SecretAccessKey is not set, attempt to retrieve it from the AWS_SECRET_ACCESS_KEY environment variable
//
// Returns an error if any of the required credentials cannot be found or retrieved.
// The method modifies the AWSClientConfigOptions struct by populating missing fields
// with values from environment variables when possible.
//
// Possible error scenarios include:
//   - Unable to retrieve region from environment
//   - Unable to retrieve access key ID from environment
//   - Unable to retrieve secret access key from environment
//
// Example:
//   config := &AWSClientConfigOptions{}
//   if err := config.Validate(); err != nil {
//       // Handle configuration validation error
//   }
func (c *AWSClientConfigOptions) Validate() error {
	if c.Region == "" {
		// Try to get region from environment if not set
		region, err := c.GetRegionFromEnv()
		if err != nil {
			return fmt.Errorf("region is required: %w", err)
		}
		c.Region = region
	}

	if c.AccessKeyID == "" {
		// Try to get access key ID from environment if not set
		accessKeyID, err := c.GetAccessKeyIDFromEnv()
		if err != nil {
			return fmt.Errorf("access key ID is required: %w", err)
		}
		c.AccessKeyID = accessKeyID
	}

	if c.SecretAccessKey == "" {
		// Try to get secret access key from environment if not set
		secretAccessKey, err := c.GetSecretAccessKeyFromEnv()
		if err != nil {
			return fmt.Errorf("secret access key is required: %w", err)
		}
		c.SecretAccessKey = secretAccessKey
	}

	return nil
}

// LoadConfig creates an AWS SDK configuration using static credentials or environment variables
// LoadConfig creates an AWS SDK configuration using the configured credentials.
//
// This method validates the AWS client configuration, ensures a region is set,
// and creates an AWS SDK configuration with static credentials. If no region
// is specified, it defaults to "us-east-1".
//
// The method performs the following steps:
//   1. Validate the AWS credentials using the Validate() method
//   2. Set a default region if not already specified
//   3. Create a static credentials provider using the configured access key,
//      secret access key, and optional session token
//   4. Load the AWS SDK configuration with the specified region and credentials
//
// Parameters:
//   - ctx: A context.Context for controlling the AWS configuration loading process.
//     This allows for potential timeouts or cancellation of the configuration process.
//
// Returns:
//   - aws.Config: A fully configured AWS SDK configuration ready for use with AWS service clients
//   - error: An error if credential validation fails or configuration cannot be loaded
func (c *AWSClientConfigOptions) LoadConfig(ctx context.Context) (aws.Config, error) {
	// Validate credentials
	if err := c.Validate(); err != nil {
		return aws.Config{}, fmt.Errorf("invalid AWS configuration: %w", err)
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
	return config.LoadDefaultConfig(ctx,
		config.WithRegion(c.Region),
		config.WithCredentialsProvider(credProvider),
	)
}

// NewAWSClientConfig creates a new AWS client configuration
// NewAWSClientConfig creates a new AWS client configuration with the specified credentials and region.
//
// This function initializes an AWSClientConfig with the provided AWS credentials and region.
// If no region is specified, it defaults to "us-east-1".
//
// Parameters:
//   - region: The AWS region to use for the client configuration. If empty, defaults to "us-east-1".
//   - accessKeyID: The AWS access key ID for authentication.
//   - secretAccessKey: The AWS secret access key for authentication.
//   - sessionToken: Optional session token for temporary credentials.
//
// Returns:
//   - AWSClientConfig: A configured AWS client configuration ready for use.
//
// Example usage:
//   awsConfig := NewAWSClientConfig("us-west-2", "AKIAIOSFODNN7EXAMPLE", "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY", "")
func NewAWSClientConfig(
	region string,
	accessKeyID string,
	secretAccessKey string,
	sessionToken string,
) AWSClientConfig {
	// Default to us-east-1 if no region provided
	if region == "" {
		region = "us-east-1"
	}

	return &AWSClientConfigOptions{
		Region:          region,
		AccessKeyID:     accessKeyID,
		SecretAccessKey: secretAccessKey,
		SessionToken:    sessionToken,
	}
}

// NewAWSClientConfigFromEnv creates an AWS client configuration using environment variables.
//
// This function attempts to create an AWS client configuration by reading credentials
// from environment variables. If no environment variables are set or they are empty,
// it returns a default configuration with empty credentials.
//
// The function uses the following standard AWS environment variables:
//   - AWS_REGION or AWS_DEFAULT_REGION: Sets the AWS region
//   - AWS_ACCESS_KEY_ID: Sets the AWS access key ID
//   - AWS_SECRET_ACCESS_KEY: Sets the AWS secret access key
//   - AWS_SESSION_TOKEN: Sets the optional session token for temporary credentials
//
// Returns:
//   - AWSClientConfig: A configured AWS client configuration, potentially using environment variables.
//
// Example usage:
//   awsConfig := NewAWSClientConfigFromEnv()
//   // This will use credentials from environment variables if available
func NewAWSClientConfigFromEnv() AWSClientConfig {
	return NewAWSClientConfig("", "", "", "")
}

// NewAWSClient is a convenience function to create an AWS configuration
// NewAWSClient creates a new AWS SDK configuration using the provided AWS client configuration.
//
// This function takes a context and an AWSClientConfig, and returns an AWS SDK configuration.
// It is a convenience wrapper around the LoadConfig method, which validates credentials,
// sets a default region if not specified, and creates a configuration with static credentials.
//
// Parameters:
//   - ctx: A context.Context for controlling the AWS configuration loading process.
//   - cfg: An AWSClientConfig containing AWS credentials and optional region information.
//
// Returns:
//   - aws.Config: A configured AWS SDK configuration ready for use with AWS service clients.
//   - error: An error if the configuration loading fails, such as invalid credentials.
//
// Example usage:
//   ctx := context.Background()
//   awsConfig := NewAWSClientConfig("us-west-2", "accessKey", "secretKey", "")
//   sdkConfig, err := NewAWSClient(ctx, awsConfig)
//   if err != nil {
//       // Handle configuration error
//   }
func NewAWSClient(
	ctx context.Context,
	cfg AWSClientConfig,
) (aws.Config, error) {
	// Load and return AWS configuration
	return cfg.LoadConfig(ctx)
} 