# Discovering AWS Resources with aws-taggy

## Overview

The `discover` command is a powerful tool in the `aws-taggy` CLI that allows you to scan and identify AWS resources across different services and regions, providing detailed insights into their tagging status.

## Command Syntax

```bash
aws-taggy discover [options]
```

## Available Options

### Service Selection

- `--service=SERVICE`: Specify the AWS service to discover
  - Supported services: `s3` (tested), likely to expand to other services in future versions
  - Example: `aws-taggy discover --service=s3`

### Region Filtering

- `--region=REGION`: Limit discovery to a specific AWS region
  - Supports standard AWS region formats (e.g., `us-east-1`, `eu-central-1`)
  - Example: `aws-taggy discover --service=s3 --region=eu-central-1`

### Tagging Filters

- `--untagged`: Display only resources without tags
  - Useful for identifying resources that need tagging
  - Example: `aws-taggy discover --service=s3 --untagged`

### Output Options

- `--with-arn`: Include Amazon Resource Names (ARNs) in the output
  - Provides full resource identification
  - Example: `aws-taggy discover --service=s3 --with-arn`

### Display Customization

- `--output=[table|json|yaml]`: Specify the output format
  - Default is likely `table`
  - Allows integration with other tools and scripts

## Example Scenarios

### 1. Discover All S3 Buckets

```bash
aws-taggy discover --service=s3
```

- Scans all S3 buckets across default region
- Shows total resources, tagged and untagged counts

### 2. Find Untagged S3 Buckets

```bash
aws-taggy discover --service=s3 --untagged
```

- Highlights resources without any tags
- Helps in identifying resources needing tag management

### 3. Discover S3 Buckets in a Specific Region

```bash
aws-taggy discover --service=s3 --region=eu-central-1
```

- Focuses on resources in the specified region
- Useful for region-specific audits

### 4. Detailed Resource Discovery with ARNs

```bash
aws-taggy discover --service=s3 --with-arn
```

- Provides comprehensive resource information
- Includes full Amazon Resource Names

## Output Explanation

The command generates a table with the following columns:

- **Resource**: Bucket/resource name
- **Region**: AWS region of the resource
- **Has Tags**: Boolean indicating tag presence
- **Tag Count**: Number of tags associated with the resource
- **ARN** (optional): Full Amazon Resource Name when `--with-arn` is used

## Best Practices

1. Regularly run discovery to maintain an updated inventory
2. Use `--untagged` to identify resources needing tag standardization
3. Combine with `tag` command for efficient resource management

## Troubleshooting

- Ensure AWS credentials are correctly configured
- Check network connectivity to AWS services
- Verify IAM permissions for resource discovery

## Limitations

- Currently supports S3 service
- Scanning large numbers of resources may take time
- Dependent on AWS API rate limits

## Related Commands

- `aws-taggy tag`: Apply tags to discovered resources
- `aws-taggy report`: Generate comprehensive tagging reports

## Performance Notes

- Scanning duration depends on the number of resources
- Typical S3 scan takes 7-12 seconds for ~88 resources
- Performance may vary based on AWS environment complexity
