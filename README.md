# 🏷️ AWS Taggy

## Overview

AWS Taggy is a powerful CLI tool for comprehensive AWS resource tagging compliance and governance. It helps organizations maintain consistent, secure, and well-managed cloud infrastructure by enforcing tagging standards across AWS resources.

## ✨ Features

- 🕵️ Comprehensive AWS resource tag scanning
- 🚨 Flexible compliance rule configuration
- 📊 Detailed reporting of untagged or non-compliant resources
- 🛡️ Customizable tagging policies
- 📝 Supports multiple output formats (JSON, CSV, CLI)

## 🚀 Installation

### Homebrew

```bash
brew tap Excoriate/aws-taggy
brew install aws-taggy
```

### Go Install

```bash
go install github.com/Excoriate/aws-taggy@latest
```

## 🛠️ Quick Start

### Scan Current AWS Account

```bash
# Basic scan with default rules
aws-taggy scan

# Scan with custom configuration
aws-taggy scan --config ./taggy-rules.yaml
```

### Generate Compliance Report

```bash
# Generate JSON report
aws-taggy scan --output json > compliance-report.json

# Generate CSV report
aws-taggy scan --output csv > compliance-report.csv
```

## 📝 Configuration

Create a `taggy-rules.yaml` to define custom tagging policies:

```yaml
# Basic configuration example
version: "1.0"

# Global tagging rules applied across all resources
global:
  enabled: true
  tag_criteria:
    # Minimum number of tags required for compliance
    minimum_required_tags: 3
    max_tags: 50

    # Core required tags for all resources
    required_tags:
      - Environment # Identifies deployment environment
      - Owner # Indicates responsible team/individual
      - Project # Associates resource with a specific project

    # Example scenario: Preventing unmanaged resources
    forbidden_tags:
      - Temporary # Blocks resources marked as temporary
      - Test # Prevents test resources from being considered compliant

    # Enforcing governance standards
    specific_tags:
      ComplianceLevel: high
      ManagedBy: terraform

# Resource-specific tagging rules
resources:
  s3:
    enabled: true
    tag_criteria:
      # Example scenario: Stricter tagging for sensitive S3 buckets
      # Imagine a financial services company with strict data governance requirements
      minimum_required_tags: 4
      required_tags:
        - DataClassification # Specifies data sensitivity
        - BackupPolicy # Defines backup strategy
        - Environment # Deployment environment
        - Owner # Resource ownership

      specific_tags:
        EncryptionRequired: "true" # Mandatory encryption for data protection

  ec2:
    enabled: true
    tag_criteria:
      # Example scenario: Managing cloud infrastructure costs and maintenance
      minimum_required_tags: 3
      required_tags:
        - Application # Identifies application running on instance
        - PatchGroup # Indicates patch management group
        - Environment # Deployment environment

      specific_tags:
        AutoStop: enabled # Enables automatic instance stopping to manage costs

# Compliance levels for different governance requirements
compliance_levels:
  high:
    required_tags:
      - SecurityLevel # Security classification
      - DataClassification # Data sensitivity
      - Backup # Backup strategy
      - Owner # Resource ownership
      - CostCenter # Precise cost allocation

    specific_tags:
      SecurityApproved: "true"
      MonitoringEnabled: "true"
```

## 🔍 Compliance Check Command

AWS Taggy provides a powerful compliance check command to validate your AWS resource tagging:

```bash
# Basic compliance scan
aws-taggy compliance scan

# Scan with custom configuration file
aws-taggy compliance scan --config ./taggy-rules.yaml

# Generate detailed compliance report
aws-taggy compliance scan --output json > compliance-report.json

# Scan specific resource types
aws-taggy compliance scan --resources s3,ec2

# Scan with verbose output for detailed insights
aws-taggy compliance scan --verbose
```

### Compliance Check Scenarios

1. **Cost Allocation and Governance**

   - **Scenario**: A multinational corporation needs precise cost tracking across multiple departments and projects.
   - **Example**: Ensure all resources have `CostCenter`, `Project`, and `Owner` tags to enable accurate financial reporting and chargeback.
   - **Benefits**:
     - Prevents untagged or improperly tagged resources from being deployed
     - Enables granular cost allocation and tracking
     - Supports financial governance and budget management

2. **Security and Compliance**

   - **Scenario**: A healthcare organization must maintain strict data protection standards.
   - **Example**: Validate security-related tags like `DataClassification`, `SecurityLevel`, and `Backup` to ensure data protection compliance.
   - **Benefits**:
     - Enforces encryption and monitoring requirements
     - Blocks resources that don't meet security tagging standards
     - Supports regulatory compliance (e.g., HIPAA, GDPR)

3. **Environment Management**
   - **Scenario**: A software development company with multiple environments and complex infrastructure.
   - **Example**: Track resources across production, staging, development, and sandbox environments.
   - **Benefits**:
     - Prevents mixing of resources between environments
     - Enables precise environment-based filtering and management
     - Supports infrastructure isolation and security

### Compliance Validation Rules

AWS Taggy implements comprehensive tag validation:

- **Tag Key Validation**

  - Must start with a lowercase letter
  - Contains only letters, numbers, underscores, and hyphens
  - Maximum length of 128 characters

- **Tag Value Validation**

  - Allows alphanumeric characters, dots, underscores, and hyphens
  - Prevents generic values like "undefined", "null", or "n/a"

- **Length Constraints**

  - Environment tag: 2-15 characters
  - Owner tag: 3-50 characters
  - Project tag: 4-30 characters

- **Case Sensitivity**
  - Strict mode for sensitive tags
  - Lowercase enforcement for environment and security tags
  - Uppercase for cost center tags

## 🚨 Compliance Levels

- **High**: Strictest tagging requirements

  - Mandatory security and data classification tags
  - Explicit security and monitoring approvals
  - Example: Financial services, healthcare, government sectors

- **Standard**: Moderate tagging requirements
  - Basic ownership and environment tracking
  - Ensures fundamental resource management
  - Example: Small to medium businesses, less regulated industries

## 🛡️ Exclusion Mechanisms

- Exclude specific resources from compliance checks
- Define exclusion patterns (e.g., Terraform state buckets)
- Provide reasons for exclusions
- Example: Excluding bastion hosts or logging infrastructure from standard compliance rules

## 🔍 Supported AWS Resources

- EC2 Instances
- RDS Databases
- S3 Buckets
- EBS Volumes
- ELB/ALB
- Lambda Functions
- And more...

## 🤝 Contributing

Contributions are welcome! Please see [CONTRIBUTING.md](CONTRIBUTING.md)

## 🛡️ Security

- Follows AWS best practices
- Supports IAM roles and temporary credentials
- Minimal AWS permissions required

## 📄 License

[MIT](LICENSE)

## 🙌 Acknowledgements

Crafted with ❤️ by Alex T. to make FinOps, and Security teams life easier.
