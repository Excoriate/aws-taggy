package cloud

import (
	"context"
	"fmt"

	"github.com/Excoriate/aws-taggy/pkg/constants"
	"github.com/Excoriate/aws-taggy/pkg/util"
	"github.com/aws/aws-sdk-go-v2/aws"
	awscfg "github.com/aws/aws-sdk-go-v2/config"
)

// AWSClientConfig defines the configuration interface for AWS client creation
type AWSClientConfig interface {
	GetRegion() string
	Validate() error
	LoadConfig(ctx context.Context) (*aws.Config, error)
}

// AWSClientConfigOptions implements AWSClientConfig
type AWSClientConfigOptions struct {
	Region string
}

func (c *AWSClientConfigOptions) GetRegion() string {
	return c.Region
}

func (c *AWSClientConfigOptions) Validate() error {
	if c.Region == "" {
		_, err := util.GetAWSRegionEnvVar()
		if err != nil {
			c.Region = constants.DefaultAWSRegion
		}
	}
	return nil
}

func (c *AWSClientConfigOptions) LoadConfig(ctx context.Context) (*aws.Config, error) {
	if err := c.Validate(); err != nil {
		return nil, fmt.Errorf("invalid AWS configuration: %w", err)
	}

	cfg, err := awscfg.LoadDefaultConfig(ctx,
		awscfg.WithRegion(c.Region),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS configuration: %w", err)
	}

	return &cfg, nil
}

// NewAWSClientConfig creates a new AWS client configuration
func NewAWSClientConfig(region string) AWSClientConfig {
	if region == "" {
		region = constants.DefaultAWSRegion
	}

	return &AWSClientConfigOptions{
		Region: region,
	}
}
