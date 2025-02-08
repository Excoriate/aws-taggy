# How to Customize Tag Compliance Check

## Overview

This guide provides detailed instructions on how to customize the tag compliance check using the `tag-compliance.yaml` configuration file. AWS Taggy allows you to enforce tagging standards across AWS resources, ensuring compliance with organizational policies.

---

### Configuration Sections

#### 1. **Version**

- **Purpose**: Tracks the schema version of the tag compliance configuration.
- **Example**:
  ```yaml
  version: "1.0"
  ```
  *No specific Terraform tagging example needed for this section.*

#### 2. **AWS Configuration**

- **Regions**: Specifies which AWS regions to scan. Options are `all` or a list of specific regions.
  - **Example**:
    ```yaml
    aws:
      regions:
        mode: all
    ```
  - **Terraform Example**:
    ```hcl
    resource "aws_s3_bucket" "example" {
      bucket = "my-bucket-in-us-east-1"

      tags = {
        Region = "us-east-1"
        # Ensures the tag matches the region configuration
      }
    }
    ```

- **Batch Size**: Controls the number of resources processed in a single batch.
  - **Example**:
    ```yaml
    batch_size: 20
    ```
  *No specific Terraform tagging example needed for this section.*

#### 3. **Global Settings**

- **Enabled**: Activates the tag compliance process for all resources.
  - **Example**:
    ```yaml
    global:
      enabled: true
    ```

- **Tag Criteria**: Defines global tagging rules.
  - **Terraform Example**:
    ```hcl
    resource "aws_instance" "example" {
      ami           = "ami-0c55b159cbfafe1f0"
      instance_type = "t2.micro"

      tags = {
        # Meets minimum required tags
        Environment = "production"
        Owner       = "cloud-team@company.com"
        Project     = "infrastructure-core"

        # Avoids forbidden tags
        # No "Temporary" or "Test" tags allowed

        # Matches specific tag requirements
        ComplianceLevel = "high"
      }
    }
    ```

#### 4. **Resource-Specific Configurations**

- **S3 Specific Configuration**:
  - **Terraform Example**:
    ```hcl
    resource "aws_s3_bucket" "compliance_bucket" {
      bucket = "my-compliant-bucket"

      tags = {
        # Meets S3-specific tag requirements
        DataClassification = "confidential"
        BackupPolicy       = "daily-backup"
        Environment        = "production"
        Owner              = "data-team@company.com"

        # Ensures encryption is required
        EncryptionRequired = "true"
      }
    }
    ```

- **EC2 Specific Configuration**:
  - **Terraform Example**:
    ```hcl
    resource "aws_instance" "application_server" {
      ami           = "ami-0c55b159cbfafe1f0"
      instance_type = "t2.medium"

      tags = {
        Application  = "web-backend"
        PatchGroup   = "monthly-patch"
        Environment  = "staging"

        # Enables auto-stop for cost management
        AutoStop     = "enabled"
      }
    }
    ```

#### 5. **Compliance Levels**

- **High Compliance Level**:
  - **Terraform Example**:
    ```hcl
    resource "aws_eks_cluster" "high_security_cluster" {
      name     = "production-cluster"
      role_arn = aws_iam_role.eks_cluster.arn

      tags = {
        # High compliance level tags
        SecurityLevel        = "high"
        DataClassification   = "restricted"
        Backup               = "weekly"
        Owner                = "security-team@company.com"
        CostCenter           = "IT-0123"

        # Additional security validations
        SecurityApproved     = "true"
        MonitoringEnabled    = "true"
        ComplianceLevel      = "high"
      }
    }
    ```

- **Standard Compliance Level**:
  - **Terraform Example**:
    ```hcl
    resource "aws_rds_cluster" "standard_database" {
      cluster_identifier = "standard-db-cluster"
      engine             = "aurora-postgresql"

      tags = {
        Owner       = "database-team@company.com"
        Project     = "customer-portal"
        Environment = "development"

        # Standard monitoring requirement
        MonitoringEnabled = "true"
        ComplianceLevel   = "standard"
      }
    }
    ```

#### 6. **Tag Validation Rules**

- **Terraform Examples Demonstrating Validation**:
  ```hcl
  resource "aws_vpc" "compliant_network" {
    cidr_block = "10.0.0.0/16"

    tags = {
      # Follows key format rules (lowercase, alphanumeric)
      network_tier = "private"

      # Matches pattern rules
      CostCenter  = "IT-0123"  # Matches ^[A-Z]{2}-[0-9]{4}$ pattern

      # Follows case sensitivity rules
      environment = "production"  # Lowercase

      # Avoids prohibited prefixes
      project_name = "core-infra"  # Not using "aws:" or "internal:" prefixes
    }
  }
  ```

#### 7. **Notification Configuration**

- **Terraform Example with Notification Tags**:
  ```hcl
  resource "aws_cloudwatch_log_group" "compliance_logs" {
    name = "/aws/taggy/compliance-logs"

    tags = {
      NotificationGroup = "compliance-alerts"
      AlertRecipient    = "cloud-team@company.com"
      ReportFrequency   = "daily"
    }
  }
  ```

---

### Best Practices for Terraform Tagging

1. Use consistent tag naming conventions
2. Leverage Terraform variables for tag values
3. Consider using a `locals` block for shared tags
4. Implement tag validation in your Terraform code

**Example of Best Practices**:
```hcl
locals {
  common_tags = {
    Environment     = var.environment
    Owner           = var.team_email
    Project         = var.project_name
    ComplianceLevel = "high"
  }
}

resource "aws_s3_bucket" "example" {
  bucket = "my-compliant-bucket"

  tags = merge(
    local.common_tags,
    {
      DataClassification = "confidential"
    }
  )
}
```

This guide provides comprehensive Terraform tag examples that align with the AWS Taggy tag compliance configuration. By following these examples, you can ensure your infrastructure-as-code meets your organization's tagging standards.

If you have any specific questions or need further examples, feel free to ask!
