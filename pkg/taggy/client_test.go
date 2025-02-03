package taggy

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Excoriate/aws-taggy/pkg/configuration"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createTempConfigFile(t *testing.T, content string) string {
	t.Helper()

	// Create a temporary directory
	tempDir := t.TempDir()

	// Create a temporary config file
	configPath := filepath.Join(tempDir, "test_config.yaml")
	err := os.WriteFile(configPath, []byte(content), 0o644)
	require.NoError(t, err)

	return configPath
}

func TestNewTaggyClient(t *testing.T) {
	t.Parallel()

	// Test case 1: Valid configuration file
	validConfigContent := `
version: "1.0"
aws:
  regions:
    mode: specific
    list:
      - us-west-2
global:
  enabled: true
  tag_criteria:
    required_tags:
      - environment
      - project
resources:
  ec2:
    enabled: true
`
	validConfigPath := createTempConfigFile(t, validConfigContent)

	client, err := New(validConfigPath)
	require.NoError(t, err)
	require.NotNil(t, client)

	assert.NotNil(t, client.Config())
	assert.Equal(t, "1.0", client.Config().Version)
	assert.Equal(t, "specific", client.Config().AWS.Regions.Mode)
	assert.Equal(t, 1, len(client.Config().AWS.Regions.List))
	assert.Equal(t, "us-west-2", client.Config().AWS.Regions.List[0])
}

func TestNewTaggyClientWithInvalidConfigFile(t *testing.T) {
	t.Parallel()

	// Test case 2: Invalid configuration file path
	client, err := New("/path/to/nonexistent/config.yaml")
	assert.Error(t, err)
	assert.Nil(t, client)
}

func TestNewWithConfig(t *testing.T) {
	t.Parallel()

	// Test case 1: Valid configuration
	validConfig := &configuration.TaggyScanConfig{
		Version: "1.0",
		AWS: configuration.AWSConfig{
			Regions: configuration.RegionsConfig{
				Mode: "specific",
				List: []string{"us-east-1"},
			},
		},
		Global: configuration.GlobalConfig{
			Enabled: true,
			TagCriteria: configuration.TagCriteria{
				RequiredTags: []string{"owner"},
			},
		},
		Resources: map[string]configuration.ResourceConfig{
			"s3": {
				Enabled: true,
			},
		},
	}

	client, err := NewWithConfig(validConfig)
	require.NoError(t, err)
	require.NotNil(t, client)

	assert.Equal(t, validConfig, client.Config())
}

func TestNewWithConfigNil(t *testing.T) {
	t.Parallel()

	// Test case 2: Nil configuration
	client, err := NewWithConfig(nil)
	assert.Error(t, err)
	assert.Nil(t, client)
}

func TestConfigMethod(t *testing.T) {
	t.Parallel()

	// Prepare a configuration
	testConfig := &configuration.TaggyScanConfig{
		Version: "1.0",
		AWS: configuration.AWSConfig{
			Regions: configuration.RegionsConfig{
				Mode: "specific",
				List: []string{"eu-west-1"},
			},
		},
		Resources: map[string]configuration.ResourceConfig{
			"ec2": {
				Enabled: true,
				Regions: []string{"eu-west-1"},
			},
			"rds": {
				Enabled: true,
			},
		},
	}

	client, err := NewWithConfig(testConfig)
	require.NoError(t, err)

	// Test the Config method
	retrievedConfig := client.Config()
	assert.Equal(t, testConfig, retrievedConfig)
	assert.Equal(t, testConfig.AWS.Regions.Mode, retrievedConfig.AWS.Regions.Mode)
	assert.Equal(t, testConfig.AWS.Regions.List, retrievedConfig.AWS.Regions.List)
	assert.Equal(t, testConfig.Resources, retrievedConfig.Resources)
}
