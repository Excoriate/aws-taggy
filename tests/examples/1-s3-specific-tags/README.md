# AWS Taggy S3 Tag Compliance Example

## Scenario Overview

This example demonstrates a specific tag compliance validation scenario for an S3 bucket, showcasing how aws-taggy enforces tagging standards.

### Test Scenario

**Goal**: Validate that an S3 bucket meets predefined tagging compliance rules.

#### Terraform Resource Tags

```hcl
tags = {
  Name                = "aws-taggy-test-bucket"
  Environment         = "development"
  Owner               = "data-engineering-team@company.com"
  Project             = "aws-taggy-demo"
  DataClassification  = "internal"
  CostCenter          = "DE-1234"
}
```

#### Compliance Configuration Validation Rules

- **Minimum Required Tags**: 5 tags
- **Required Tags**:
  - `Environment`: Must be `development`, `production`, `staging`, or `sandbox`
  - `Owner`: Must be an email from `@company.com`
  - `Project`: Must be present
  - `DataClassification`: Must be `public`, `internal`, `confidential`, or `restricted`
  - `CostCenter`: Must match pattern `XX-NNNN`

### Expected Compliance Result

**Scenario**: The provided S3 bucket tags should pass all compliance checks

- ✅ 5 required tags present
- ✅ `Environment` is a valid value (`development`)
- ✅ `Owner` matches email pattern
- ✅ `CostCenter` matches `XX-NNNN` pattern
- ✅ `DataClassification` is a valid value (`internal`)

## Running the Example

### Prerequisites

- Terraform
- AWS CLI
- aws-taggy CLI
- AWS Credentials

### AWS Credentials Setup

Before running the example, you must set up AWS credentials:

```bash
# Option 1: AWS CLI Configuration
aws configure

# Option 2: Environment Variables
export AWS_ACCESS_KEY_ID='your_access_key'
export AWS_SECRET_ACCESS_KEY='your_secret_key'
export AWS_REGION='us-east-1'  # Or your preferred region
```

### Workflow

1. **Terraform Operations**

```bash
# Plan Terraform changes
./run.sh terraform

# Apply Terraform changes
./run.sh apply
```

2. **Compliance Check**

```bash
# Run compliance check
./run.sh compliance

# Alternative direct command
aws-taggy compliance check \
  --config ./tag-compliance.yaml \
  --output=table \
  --detailed
```

### Run Modes

The `run.sh` script supports multiple modes:

- `all` (default): Full workflow (apply, check, destroy)
- `terraform`: Only Terraform operations
- `compliance`: Only run compliance check
- `destroy`: Remove created resources

### Compliance Check Flags

```bash
# Show detailed table output
aws-taggy compliance check \
  --config ./tag-compliance.yaml \
  --output=table \
  --detailed

# Generate JSON output
aws-taggy compliance check \
  --config ./tag-compliance.yaml \
  --output=json \
  --output-file=compliance_results.json
```

## Compliance Check with Specific Resource

### Resource-Specific Compliance Validation

You can run a compliance check for the S3 bucket using two methods:

#### Method 1: Using Justfile (Recommended)

```bash
# Run full example workflow (create resources and check compliance)
just run-example 1-s3-specific-tags run

# Run compliance check assuming infrastructure is already created
just run-example 1-s3-specific-tags run-cli
```

#### Method 2: Direct CLI Execution

```bash
# Run compliance check from source code
go run cli/main.go compliance check \
  --config tests/examples/1-s3-specific-tags/tag-compliance.yaml \
  --resource aws-taggy \
  --output=table \
  --detailed
```

### Compliance Check Options

- `just run-example 1-s3-specific-tags run`: Creates resources and runs compliance check
- `just run-example 1-s3-specific-tags run-cli`: Runs compliance check on existing resources
- Direct CLI command supports various output formats and detailed reporting

### Benefits of Resource-Specific Checks

- Quickly validate a single resource's tag compliance
- Useful for spot-checking specific resources
- Supports both resource names and full ARNs
- Provides detailed insights into tag compliance for the specified resource

### Troubleshooting

- Ensure the resource name exactly matches the existing resource
- Verify AWS credentials are correctly configured
- Check that the resource exists in the specified AWS region
- Use `--detailed` flag for comprehensive compliance information

## Troubleshooting

- Ensure AWS credentials are correctly configured
- Verify AWS CLI is installed and working
- Check that you have necessary permissions for S3 bucket creation
- Review AWS region settings
