# Understanding Tag Compliance Configuration in aws-taggy

## Overview

The tag compliance configuration defines rules and standards for tagging AWS resources, ensuring consistent and meaningful resource metadata across your infrastructure.

## Configuration Structure

### Version

- Tracks schema version for future compatibility
- Current version: `1.0`

### AWS Configuration

- **Regions**:
  - Scan mode: `all` or specific regions
  - Batch processing size configurable

### Global Tagging Rules

- Enable/disable tag compliance
- Set minimum and maximum tag requirements
- Define required and forbidden tags

### Resource-Specific Configurations

- Customize tagging rules for different AWS services
- Currently supports:
  - S3 Buckets
  - EC2 Instances

## Key Configuration Elements

### Required Tags

Mandatory tags across resources:

- `Environment`: Deployment environment
- `Owner`: Responsible team/individual
- `Project`: Associated project

### Compliance Levels

1. **High Compliance**

   - Strictest tagging requirements
   - Additional security and monitoring tags
   - Explicit security validations

2. **Standard Compliance**
   - Moderate tagging requirements
   - Basic monitoring and ownership tags

## Tag Validation Rules

### Tag Key Restrictions

- Must start with lowercase letter
- Contain only letters, numbers, underscores, hyphens
- Maximum 128 characters

### Allowed Tag Values

Predefined allowed values for specific tags:

- `Environment`: production, staging, development
- `DataClassification`: public, private, confidential
- `SecurityLevel`: high, medium, low

### Case Sensitivity

- Some tags require specific case (lowercase, uppercase)
- Enforces consistent tagging format

## Notification Configuration

- Slack and email alerts for compliance issues
- Configurable reporting frequency
- Channels for different priority levels

## Example Configuration Snippet

```yaml
global:
  minimum_required_tags: 3
  required_tags:
    - Environment
    - Owner
    - Project

resources:
  s3:
    minimum_required_tags: 4
    required_tags:
      - DataClassification
      - BackupPolicy
```

## Best Practices

- Start with generated template
- Customize to match organizational standards
- Regularly review and update tagging rules
- Use consistent naming conventions

## Troubleshooting

- Validate configuration before deployment
- Use `aws-taggy config validate` to check syntax
- Enable debug mode for detailed information

## Related Commands

- `aws-taggy config validate`: Check configuration
- `aws-taggy config generate`: Create sample config
- `aws-taggy discover`: Find resources
- `aws-taggy compliance check`: Validate resource tags
