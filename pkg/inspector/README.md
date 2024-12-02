# Inspector Package Documentation

## Overview

The `inspector` package provides a flexible and extensible framework for asynchronous resource discovery and inspection across multiple AWS services and regions. It is designed to efficiently scan, process, and analyze cloud resources with configurable concurrency and robust error handling.

## Key Components

### Core Interfaces and Types

1. **`Inspector` Interface**

   - Defines the core contract for resource inspection
   - Two primary methods:
     - `Inspect(ctx context.Context, config TaggyScanConfig) (*InspectResult, error)`
     - `Fetch(ctx context.Context, arn string, config TaggyScanConfig) (*ResourceMetadata, error)`

2. **`ResourceMetadata` Struct**
   - Represents detailed information about a discovered resource
   - Contains:
     - `ID`: Unique resource identifier
     - `Type`: Resource type (e.g., "s3", "vpc", "ec2")
     - `Provider`: Cloud provider (default: "aws")
     - `Region`: AWS region
     - `Tags`: Resource tags
     - `DiscoveredAt`: Timestamp of resource discovery
     - `RawResponse`: Original cloud provider response
     - `Details`: Extended resource information

### Async Resource Inspection

3. **`AsyncResourceInspector`**
   - Manages parallel resource discovery and processing
   - Key features:
     - Configurable worker count
     - Batch processing
     - Concurrent region scanning
     - Error aggregation

## Supported Resource Inspectors

### Current Implementations

1. **S3 Inspector**

   - Scans S3 buckets across specified regions
   - Retrieves bucket metadata and tags

2. **VPC Inspector**

   - Discovers VPC resources
   - Extracts VPC details, CIDR blocks, and tags

3. **EC2 Inspector**
   - Scans EC2 instances
   - Collects instance metadata and tags

## Usage Examples

### Creating an Inspector

```go
// Create an S3 scanner for specific regions
s3Scanner, err := NewS3Scanner([]string{"us-west-2", "us-east-1"})

// Create a VPC scanner
vpcScanner, err := NewVPCScanner([]string{"us-west-2"})
```

### Performing Resource Inspection

```go
// Configure scan parameters
config := configuration.TaggyScanConfig{
    ComplianceLevel: "high",
    // Other configuration options
}

// Inspect resources
result, err := s3Scanner.Inspect(context.Background(), config)
if err != nil {
    // Handle error
}

// Process discovered resources
for _, resource := range result.Resources {
    fmt.Printf("Resource ID: %s, Region: %s\n",
        resource.ID, resource.Region)
}
```

### Fetching Specific Resource Details

```go
// Fetch details for a specific resource by ARN
resourceDetails, err := s3Scanner.Fetch(
    context.Background(),
    "arn:aws:s3:::my-bucket",
    config
)
```

## Configuration Options

- **Regions**: Specify which AWS regions to scan
- **Compliance Levels**: Define resource tag compliance requirements
- **Batch Processing**: Configure concurrent worker count and batch size

## Error Handling

- Detailed error messages for resource discovery and processing
- Aggregates and reports multiple errors during scanning
- Provides context about failed resource inspections

## Performance Considerations

- Uses goroutines for parallel processing
- Configurable concurrency settings
- Efficient channel-based communication
- Minimal overhead with client caching

## Best Practices

1. Always specify regions explicitly
2. Handle potential errors during inspection
3. Use appropriate compliance levels
4. Monitor resource discovery logs
5. Adjust worker count based on your infrastructure

## Extensibility

- Easy to add new resource type inspectors
- Implement the `Inspector` interface for custom resource scanning
- Leverage the `AsyncResourceInspector` for parallel processing

## Dependencies

- AWS SDK for Go v2
- Zap logging
- Custom configuration management

## Logging

Uses the `o11y` package for structured logging with different log levels:

- Info: Resource discovery and processing
- Error: Failed resource inspections
- Debug: Detailed scanning information

## Limitations

- Currently supports AWS resources
- Requires AWS credentials and permissions
- Performance may vary based on resource count and network conditions

## Contributing

Refer to the project's `CONTRIBUTING.md` for guidelines on adding new inspectors or improving existing implementations.

## Advanced Configuration and Usage

### Configuration Strategies

The inspector package supports multiple configuration strategies:

1. **Dynamic Configuration**

   ```yaml
   aws:
     regions:
       mode: all # Options: all, specific
       list: # Optional list of regions when mode is 'specific'
         - us-west-2
         - us-east-1

   resources:
     s3:
       enabled: true
       regions: # Optional region override
         - us-west-2
     ec2:
       enabled: false
   ```

2. **Resource-Specific Filtering**
   - Filter resources by type, region, or specific attributes
   - Supports granular control over resource discovery

### CLI Integration Examples

#### Resource Discovery

```bash
# Discover S3 resources in a specific region
taggy discover s3 --region us-west-2

# Discover only untagged EC2 instances
taggy discover ec2 --untagged

# Discover resources with ARN details
taggy discover s3 --with-arn
```

#### Compliance Checking

```bash
# Run compliance check with a specific configuration
taggy check /path/to/compliance-config.yaml

# Filter compliance check for a specific resource
taggy check /path/to/config.yaml --resource my-bucket-name

# Generate detailed compliance report
taggy check /path/to/config.yaml --detailed --output json
```

### Compliance Validation Rules

The inspector supports comprehensive tag compliance validation:

1. **Required Tags**

   - Ensure specific tags are present
   - Define mandatory tags in configuration

2. **Tag Value Validation**

   - Enforce tag value formats
   - Support regex patterns
   - Define allowed value sets

3. **Case Sensitivity**
   - Control tag key and value casing
   - Enforce naming conventions

### Output Formats

- `table`: Human-readable terminal output
- `json`: Machine-parseable format
- `yaml`: Alternative structured format
- Clipboard support for easy sharing

### Error Handling and Logging

- Detailed error messages
- Configurable logging levels
- Comprehensive error aggregation
- Support for partial scan success

## Performance Considerations

- Concurrent resource scanning
- Configurable worker pools
- Efficient AWS SDK v2 integration
- Minimal API call overhead

## Extensibility

### Adding New Resource Inspectors

1. Implement the `Inspector` interface
2. Create a new scanner struct
3. Define resource-specific discovery logic
4. Register with `InspectorManager`

### Custom Validation Rules

- Extend `TagValidator`
- Define complex validation logic
- Support domain-specific compliance requirements

## Security and Permissions

- Requires AWS IAM permissions for resource scanning
- Supports AWS credential resolution
- No sensitive data exposure in logs or output

## Limitations and Constraints

- AWS-specific implementation
- Region and service-level restrictions
- Performance dependent on AWS API limits

## Troubleshooting

- Use verbose logging for detailed diagnostics
- Check AWS credentials and permissions
- Validate configuration file syntax
- Monitor API rate limits

## Contributing

1. Follow Go best practices
2. Maintain high test coverage
3. Document new features
4. Adhere to existing code structure

## Adding a New AWS Resource Inspector

To add support for a new AWS resource type, follow these steps:

### 1. Add Resource Type Constant

In `pkg/constants/aws.go`, add the new resource type constant:

```go
const (
    // ... existing constants ...
    ResourceTypeNewService = "newservice"  // Add your new service constant
)
```

### 2. Register as Supported Resource

In `pkg/configuration/supported_resources.go`, add the resource to `SupportedAWSResources`:

```go
var SupportedAWSResources = map[string]bool{
    // ... existing resources ...
    constants.ResourceTypeNewService: true,  // Add your new service
}
```

Also, update the `NormalizeResourceType` function if your resource needs special name handling:

```go
func NormalizeResourceType(resource string) string {
    switch normalized {
    // ... existing cases ...
    case "new-service", "newservice":
        return constants.ResourceTypeNewService
    }
}
```

### 3. Create Resource Inspector

Create a new file `pkg/inspector/awsnewservice.go`:

```go
package inspector

import (
    "context"
    "fmt"
    "time"

    "github.com/Excoriate/aws-taggy/pkg/configuration"
    "github.com/Excoriate/aws-taggy/pkg/o11y"
    "github.com/aws/aws-sdk-go-v2/aws"
    "github.com/aws/aws-sdk-go-v2/service/newservice"
    "github.com/aws/aws-sdk-go-v2/service/newservice/types"
)

// NewServiceClientCreator implements AWSClient for NewService
type NewServiceClientCreator struct{}

func (c *NewServiceClientCreator) CreateFromConfig(cfg *aws.Config) interface{} {
    return newservice.NewFromConfig(*cfg)
}

// GetNewServiceClient retrieves a NewService client for the specified AWS region.
//
// This method creates or retrieves an existing NewService client configuration for the given region.
// It uses the AWSClientManager's internal client management to ensure efficient client reuse.
//
// Parameters:
//   - region: The AWS region for which to create or retrieve the NewService client
//
// Returns:
//   - *newservice.Client: A configured AWS NewService client
//   - error: An error if client creation fails
func (m *AWSClientManager) GetNewServiceClient(region string) (*newservice.Client, error) {
    client, err := m.GetClient(region, &NewServiceClientCreator{})
    if err != nil {
        return nil, err
    }
    return client.(*newservice.Client), nil
}

// NewServiceInspector implements the Inspector interface for AWS NewService resources
type NewServiceInspector struct {
    Regions       []string
    ClientManager *AWSClientManager
    Logger        *o11y.Logger
}

// NewNewServiceInspector creates a new inspector with AWS client management
func NewNewServiceInspector(regions []string) (*NewServiceInspector, error) {
    clientManager, err := NewAWSRegionalClientManager(regions)
    if err != nil {
        return nil, fmt.Errorf("failed to create AWS client manager: %w", err)
    }
    return &NewServiceInspector{
        Regions:       regions,
        ClientManager: clientManager,
        Logger:        o11y.DefaultLogger(),
    }, nil
}

// Implement required interface methods:
// - Inspect(ctx context.Context, config configuration.TaggyScanConfig) (*InspectResult, error)
// - Fetch(ctx context.Context, arn string, config configuration.TaggyScanConfig) (*ResourceMetadata, error)
```

### 4. Register in Inspector Factory

In `pkg/inspector/inspector.go`, add your resource to the `New` function:

```go
func New(resourceType string, cfg configuration.TaggyScanConfig) (Inspector, error) {
    switch resourceType {
    // ... existing cases ...
    case constants.ResourceTypeNewService:
        return NewNewServiceInspector(regions)
    }
}
```

### Implementation Guidelines

1. **Resource Inspector Structure**

   - Place all service-specific code in the resource's file (e.g., `awsnewservice.go`)
   - Include the client creator and getter in the same file
   - Follow the `ResourceInspector` naming convention
   - Implement both `Inspect` and `Fetch` methods

2. **AWS Client Management**

   - Define service-specific `ClientCreator` struct
   - Implement `CreateFromConfig` method
   - Add type-safe client getter method
   - Use the generic `GetClient` from `AWSClientManager`

3. **Resource Discovery**

   - Use AWS SDK v2 pagination where available
   - Handle region-specific resources appropriately
   - Implement proper error handling and logging

4. **Tag Management**

   - Use the appropriate AWS SDK v2 methods for tag retrieval
   - Handle missing or empty tags gracefully
   - Follow AWS tagging best practices

5. **Error Handling**
   - Use proper error wrapping with `fmt.Errorf("message: %w", err)`
   - Handle AWS-specific errors using `errors.As`
   - Provide meaningful error messages

### Example Usage

```go
// Initialize configuration
cfg := configuration.TaggyScanConfig{
    AWS: configuration.AWSConfig{
        Regions: configuration.RegionsConfig{
            Mode: "specific",
            List: []string{"us-west-2"},
        },
    },
}

// Create inspector
inspector, err := inspector.New(constants.ResourceTypeNewService, cfg)
if err != nil {
    log.Fatal(err)
}

// Perform inspection
result, err := inspector.Inspect(context.Background(), cfg)
if err != nil {
    log.Fatal(err)
}
```

## Best Practices

1. **File Organization**

   - Keep all service-specific code in one file
   - Follow consistent naming patterns
   - Include comprehensive documentation

2. **Client Management**

   - Implement proper client caching
   - Handle client errors appropriately
   - Use type-safe client getters

3. **Resource Scanning**

   - Use asynchronous scanning for better performance
   - Implement proper pagination
   - Handle resource-specific limitations

4. **Documentation**
   - Add comprehensive GoDoc comments
   - Document any region-specific behavior
   - Include examples in documentation

## Common Pitfalls

1. **Region Handling**

   - Some services are global (like S3)
   - Others are strictly regional
   - Handle region-specific errors

2. **API Limitations**

   - Be aware of AWS API rate limits
   - Implement proper backoff strategies
   - Handle service quotas appropriately

3. **Resource Cleanup**
   - Close clients properly
   - Clean up any temporary resources
   - Handle context cancellation

## Adding a New AWS Resource Inspector

### Prerequisites

Before creating a new inspector, ensure you have the required AWS SDK v2 dependencies:

```bash
# Add the AWS SDK v2 base packages
go get github.com/aws/aws-sdk-go-v2
go get github.com/aws/aws-sdk-go-v2/config

# Add the specific service package (example for CloudWatch Logs)
go get github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs

# Add AWS common types if needed
go get github.com/aws/aws-sdk-go-v2/aws
```

These dependencies should be added to your `go.mod` file before implementing the inspector.

## Dependency Management

When adding a new AWS service inspector, follow these dependency management best practices:

1. **Version Compatibility**

   - Use compatible versions of AWS SDK v2 packages
   - Add dependencies explicitly to `go.mod`
   - Avoid mixing SDK versions

2. **Required Packages**

   ```go
   import (
       "github.com/aws/aws-sdk-go-v2/aws"
       "github.com/aws/aws-sdk-go-v2/service/<service-name>"
       "github.com/aws/aws-sdk-go-v2/service/<service-name>/types"
   )
   ```

3. **Dependency Updates**

   - Run `go mod tidy` after adding new dependencies
   - Verify dependency tree with `go mod graph`
   - Test with updated dependencies

4. **Common Issues**

   - Missing service package in `go.mod`
   - Incompatible SDK versions
   - Indirect dependencies conflicts

5. **Resolution Steps**

   ```bash
   # Update dependencies
   go get -u github.com/aws/aws-sdk-go-v2/...

   # Clean up module
   go mod tidy

   # Verify build
   go build ./...
   ```
