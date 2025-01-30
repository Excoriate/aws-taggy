# Querying AWS Resources with aws-taggy

## Overview

The `aws-taggy query` command provides powerful capabilities to retrieve detailed information about AWS resources. It offers two primary subcommands:

1. `query tags`: Retrieve tags for a specific AWS resource
2. `query info`: Get comprehensive details about a specific AWS resource

## Prerequisites

- AWS credentials configured (via AWS CLI, environment variables, or AWS config file)
- Proper IAM permissions to describe resources
- `aws-taggy` CLI installed

## Query Info Command

### Purpose

The `query info` command retrieves comprehensive details about a specific AWS resource, including:

- Resource identifier
- Resource type
- AWS region
- Provider
- Number of tags
- Full ARN
- Additional resource-specific properties

### Usage

```bash
aws-taggy query info --arn=RESOURCE_ARN --service=SERVICE_TYPE [options]
```

### Required Parameters

- `--arn`: The complete Amazon Resource Name (ARN) of the resource

  - **Must be the full, exact ARN**
  - Example: `arn:aws:s3:::my-bucket`

- `--service`: The AWS service type
  - Supports services like: `s3`, `ec2`, `rds`, `lambda`, etc.
  - **Must match the service of the resource**

### Optional Flags

- `--output`: Specify the output format

  - Supported formats:
    - `table` (default, human-readable)
    - `json` (machine-readable)
    - `yaml` (machine-readable)
  - Example: `--output=json`

- `--clipboard`: Copy the output directly to your system clipboard

  - Useful for quick sharing or further processing
  - Example: `--clipboard`

- `--debug`: Enable detailed debug information
  - Provides additional context about the resource query
  - Helpful for troubleshooting

### Examples

```bash
# Query information for a Serverless Deployment S3 Bucket
aws-taggy query info \
  --arn=arn:aws:s3:::contactservice-microserv-serverlessdeploymentbuck-1v5kalz3bhyuu \
  --service=s3

# Example Output:
# ID:        contactservice-microserv-serverlessdeploymentbuck-1v5kalz3bh
# Type:      s3
# Region:    us-east-1
# Provider:  aws
# Tag Count: 3
# ARN:       arn:aws:s3:::contactservice-microserv-serverlessdeploymentbuck-1v5kalz3bhyuu

# Query information for an S3 bucket with JSON output
aws-taggy query info \
  --arn=arn:aws:s3:::contactservice-microserv-serverlessdeploymentbuck-1v5kalz3bhyuu \
  --service=s3 \
  --output=json

# Query information and copy to clipboard
aws-taggy query info \
  --arn=arn:aws:s3:::contactservice-microserv-serverlessdeploymentbuck-1v5kalz3bhyuu \
  --service=s3 \
  --clipboard
```

## Query Tags Command

### Purpose

The `query tags` command retrieves all tags associated with a specific AWS resource.

### Usage

```bash
aws-taggy query tags --arn=RESOURCE_ARN --service=SERVICE_TYPE [options]
```

### Required Parameters

- `--arn`: The complete Amazon Resource Name (ARN) of the resource
- `--service`: The AWS service type

### Optional Flags

- `--output`: Specify the output format (table, json, yaml)
- `--clipboard`: Copy tags to clipboard

### Examples

```bash
# Query tags for a Serverless Deployment S3 Bucket
aws-taggy query tags \
  --arn=arn:aws:s3:::contactservice-microserv-serverlessdeploymentbuck-1v5kalz3bhyuu \
  --service=s3

# Example Output:
# Key           Value
# environment   production
# STAGE         prod

# Query tags with JSON output
aws-taggy query tags \
  --arn=arn:aws:s3:::contactservice-microserv-serverlessdeploymentbuck-1v5kalz3bhyuu \
  --service=s3 \
  --output=json
```

## Troubleshooting

### Common Issues

1. **Invalid ARN**: Ensure you're using the complete, exact ARN
2. **Service Mismatch**: Verify the service type matches the resource
3. **Permissions**: Confirm you have the necessary IAM permissions

### Debugging

- Use the `--debug` flag for additional information
- Verify AWS credentials and configuration

## Best Practices

1. Always use the full, exact ARN
2. Match the service type precisely
3. Use appropriate output format based on your use case
4. Leverage the clipboard feature for quick sharing

## Supported Services

Supported services include, but are not limited to:

- S3
- EC2
- RDS
- Lambda
- ECS
- EKS

_Note: The exact list of supported services may vary. Use `--help` for the most up-to-date information._

## Performance Considerations

- Querying resources requires AWS API calls
- Large numbers of queries may impact performance
- Consider using filters or specific regions to optimize query speed

## Security

- Ensure your AWS credentials have read-only permissions
- Avoid sharing ARNs or query results containing sensitive information
