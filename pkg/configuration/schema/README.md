# AWS Taggy Tag Compliance Schema

## Overview

This document provides a comprehensive guide to the AWS Taggy tag compliance configuration schema. The schema defines how AWS resources are validated for tagging compliance, enabling organizations to enforce consistent tagging standards across their cloud infrastructure.

## Schema Version

The schema uses semantic versioning to track configuration compatibility. The current version is `1.0`.

```yaml
version: "1.0"
```

## Global Configuration

The global configuration defines baseline tagging rules applied across all AWS resources.

### Tag Criteria

```yaml
global:
  enabled: true
  tag_criteria:
    minimum_required_tags: 3
    max_tags: 50
    required_tags:
      - Environment
      - Owner
      - Project
```

#### Example Scenario: Enterprise Governance

Imagine a large enterprise with multiple departments and projects. The global configuration ensures:

- Every resource has at least 3 tags
- No resource exceeds 50 tags
- Core metadata tags like Environment, Owner, and Project are mandatory

### Forbidden and Specific Tags

```yaml
global:
  tag_criteria:
    forbidden_tags:
      - Temporary
      - Test
    specific_tags:
      ComplianceLevel: high
      ManagedBy: terraform
```

#### Example Scenario: Preventing Shadow IT

- Blocks resources tagged as "Temporary" or "Test" from compliance
- Enforces infrastructure management through Terraform
- Prevents ad-hoc resource creation outside approved processes

## Resource-Specific Configurations

Different AWS resource types can have unique tagging requirements.

### S3 Bucket Example

```yaml
resources:
  s3:
    enabled: true
    tag_criteria:
      minimum_required_tags: 4
      required_tags:
        - DataClassification
        - BackupPolicy
        - Environment
        - Owner
      specific_tags:
        EncryptionRequired: "true"
```

#### Scenario: Financial Data Storage

For S3 buckets storing sensitive financial data:

- Requires 4 tags for comprehensive metadata
- Mandates data classification and backup policy tags
- Ensures encryption is explicitly enabled

### EC2 Instance Example

```yaml
resources:
  ec2:
    enabled: true
    tag_criteria:
      minimum_required_tags: 3
      required_tags:
        - Application
        - PatchGroup
        - Environment
      specific_tags:
        AutoStop: enabled
```

#### Scenario: Cost-Efficient Infrastructure

For EC2 instances:

- Tracks application, patch management, and environment
- Enables automatic instance stopping to manage costs

## Compliance Levels

Define different compliance standards for varying governance requirements.

```yaml
compliance_levels:
  high:
    required_tags:
      - SecurityLevel
      - DataClassification
      - Backup
      - Owner
      - CostCenter
    specific_tags:
      SecurityApproved: "true"
      MonitoringEnabled: "true"
```

### Compliance Level Scenarios

- **High Level**: Strict requirements for regulated industries

  - Comprehensive security and data management tags
  - Explicit security approvals
  - Suitable for healthcare, finance, government sectors

- **Standard Level**: Basic governance for less regulated environments
  - Core ownership and environment tracking
  - Suitable for small to medium businesses

## Tag Validation Rules

Comprehensive validation mechanisms to ensure tag quality and consistency.

### Key Validation Rules

```yaml
tag_validation:
  key_format_rules:
    - pattern: "^[a-z][a-z0-9_-]*$"
    - pattern: "^.{1,128}$"
```

- Must start with lowercase letter
- Contains only letters, numbers, underscores, hyphens
- Maximum length of 128 characters

### Value Validation

```yaml
tag_validation:
  value_validation:
    allowed_characters: "a-zA-Z0-9._-"
    disallowed_values:
      - "undefined"
      - "null"
      - "none"
      - "n/a"
```

- Prevents generic or meaningless tag values
- Ensures meaningful and specific tag content

### Length Constraints

```yaml
tag_validation:
  length_rules:
    environment:
      min_length: 2
      max_length: 15
    owner:
      min_length: 3
      max_length: 50
```

- Enforces meaningful tag lengths
- Prevents overly short or excessively long tags

### Case Sensitivity

```yaml
tag_validation:
  case_sensitivity:
    Environment:
      mode: strict
    DataClassification:
      mode: relaxed
```

- Strict mode for sensitive tags
- Flexible case rules for different tag types

## Exclusion Mechanisms

```yaml
resources:
  s3:
    excluded_resources:
      - pattern: terraform-state-*
        reason: Terraform state buckets managed separately
```

- Define patterns to exclude specific resources from compliance checks
- Provide reasons for exclusions
- Useful for infrastructure-specific resources like Terraform state buckets

## Best Practices

1. Start with a standard compliance level
2. Gradually increase complexity and strictness
3. Regularly review and update tagging policies
4. Use automation to enforce compliance
5. Educate teams about tagging importance

## Contributing

Contributions to improve the tag compliance schema are welcome! Please follow the project's contribution guidelines.

## License

[MIT License](LICENSE)
