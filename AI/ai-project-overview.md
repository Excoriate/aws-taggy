# 🏷️ AWS Tag Inspector (Taggy)

## Configuration File Logic and Structure

The tag-compliance.yaml configuration file is a sophisticated governance mechanism for AWS resource tagging, providing a flexible and powerful framework for defining, validating, and enforcing tagging standards across different resource types.

### Configuration file

```
---
version: "1.0"
# Global settings applied to all resources unless overridden
global:
  enabled: true
  batch_size: 20
  tag_criteria:
    minimum_required_tags: 3
    required_tags:
      - Environment
      - Owner
      - Project
    forbidden_tags:
      - Temporary
      - Test
    specific_tags:
      ComplianceLevel: high
      ManagedBy: terraform
    compliance_level: high

# Resource-specific configurations
resources:
  s3:
    enabled: true
    batch_size: 10
    tag_criteria:
      minimum_required_tags: 4
      required_tags:
        - DataClassification
        - BackupPolicy
        - Environment
        - Owner
      forbidden_tags:
        - Temporary
        - Test
      specific_tags:
        EncryptionRequired: "true"
      compliance_level: high
    excluded_resources:
      - pattern: terraform-state-*
        reason: Terraform state buckets managed separately
      - pattern: log-archive-*
        reason: Logging buckets excluded from standard compliance

  ec2:
    enabled: true
    batch_size: 15
    tag_criteria:
      minimum_required_tags: 3
      required_tags:
        - Application
        - PatchGroup
        - Environment
      forbidden_tags:
        - Temporary
        - Test
      specific_tags:
        AutoStop: enabled
      compliance_level: standard
    excluded_resources:
      - pattern: bastion-*
        reason: Bastion hosts managed by security team

# Compliance levels and their requirements
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
  standard:
    required_tags:
      - Owner
      - Project
      - Environment
    specific_tags:
      MonitoringEnabled: "true"

# Tag validation rules
tag_validation:
  allowed_values:
    Environment:
      - production
      - staging
      - development
      - sandbox
    DataClassification:
      - public
      - private
      - confidential
      - restricted
    SecurityLevel:
      - high
      - medium
      - low
  pattern_rules:
    CostCenter: ^[A-Z]{2}-[0-9]{4}$
    ProjectCode: ^PRJ-[0-9]{5}$
    Owner: ^[a-z0-9._%+-]+@company\.com$

# Notification settings for non-compliant resources
notifications:
  slack:
    enabled: true
    channels:
      high_priority: "compliance-alerts"
      standard: "compliance-reports"
  email:
    enabled: true
    recipients:
      - cloud-team@company.com
      - security-team@company.com
    frequency: daily

```

### Configuration Hierarchy and Principles

The configuration follows a multi-layered approach with the following key principles:

- Hierarchical configuration inheritance
- Granular resource-type specific overrides
- Comprehensive validation mechanisms
- Flexible compliance level definitions

### Configuration Sections

#### 1. Version Control

- Purpose: Tracks configuration schema version
- Enables future compatibility and schema evolution
- Allows for potential backward/forward compatibility strategies

#### 2. Global Configuration

- Establishes default tagging rules applicable across all resources
- Provides baseline compliance and processing parameters
- Key Components:
  - Global enablement flag
  - Default batch processing size
  - Minimum tag requirements
  - Default required and forbidden tags
  - Global compliance level specification

#### 3. Resource-Specific Configurations

- Allows granular control over individual resource types
- Supports resource-type specific:
  - Enablement flags
  - Batch processing sizes
  - Unique tag criteria
  - Resource exclusion patterns

#### 4. Compliance Levels

- Defines hierarchical compliance standards
- Supports multiple compliance tiers (e.g., high, standard)
- Specifies tag requirements for each compliance level
- Enables nuanced governance across different resource sensitivity levels

#### 5. Tag Validation Rules

- Implements strict validation mechanisms for tag values
- Supports:
  - Allowed value restrictions
  - Regex pattern matching
  - Comprehensive tag format enforcement

#### 6. Notification Configuration

- Manages reporting and alerting mechanisms
- Supports multiple notification channels
  - Slack integration
  - Email notifications
- Configurable reporting frequencies and recipients

### Scanning and Validation Logic

1. Configuration Loading

   - Parse YAML configuration
   - Validate configuration structure
   - Compile regex patterns
   - Prepare tag validation rules

2. Resource Discovery

   - Identify enabled resource types
   - Apply global and resource-specific configurations
   - Determine scanning parameters

3. Tag Scanning Process

   - Retrieve resource tags
   - Apply validation rules
   - Check against compliance levels
   - Generate detailed compliance results

4. Reporting Mechanism
   - Aggregate scan results
   - Apply notification rules
   - Distribute compliance reports

### Advanced Features

- Dynamic resource type support
- Concurrent resource scanning
- Flexible exclusion mechanisms
- Comprehensive error handling
- Detailed compliance reporting

The configuration provides a robust, extensible framework for AWS resource tag governance, with comprehensive validation, exclusion, and notification mechanisms.

### Potential Future Enhancements

- Dynamic compliance level assignment
- More granular exclusion patterns
- Enhanced error reporting
- Expanded resource type support

## Example Use Cases

1. **Cloud Governance**

   - Enforce standardized tagging across multi-account environments
   - Ensure consistent resource metadata

2. **Compliance and Security**

   - Implement tag-based access controls
   - Validate resource metadata for security and cost allocation

3. **Cost Management**
   - Enforce tagging standards for accurate cost tracking
   - Identify and remediate untagged or improperly tagged resources

## Limitations and Considerations

- Requires AWS credentials with read-only access
- Performance may vary with large numbers of resources
- Limited to supported AWS services

## Contributing

Contributions are welcome! Please refer to the project's contribution guidelines.

## License

Apache License 2.0
