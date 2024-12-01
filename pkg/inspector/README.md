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
