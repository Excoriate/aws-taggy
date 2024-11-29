# S3 Non-Compliant Tags Example

## Scenario Overview

This example demonstrates multiple tag compliance violations for an AWS S3 bucket, showcasing various ways resources can fail tag compliance checks.

## Compliance Violations

### 1. Prohibited Tags

- `temp:test`: Temporary tags are explicitly forbidden
- `test:example`: Test-related tags are not allowed

### 2. Invalid Tag Keys

- `ENVIRONMENT`: Uppercase tag keys are not permitted
- `Owner`: Incorrect capitalization
- `Sensitive`: Unexpected capitalized tag

### 3. Incorrect Tag Values

- `environment`: "development" is not an allowed value
- `data_class`: "unknown" is not a valid classification
- `Owner`: Email does not match required domain
- `project`: Contains uppercase letters
- `cost_center`: Incorrect format

### 4. Additional Issues

- Unexpected tags like `random:tag`
- Inconsistent tag naming conventions

## Compliance Rules

The example enforces strict tagging rules:

- Lowercase tag keys only
- Specific allowed values for environment and data classification
- Company email domain requirement
- Specific project name and cost center formats
- No temporary or test-related tags

## Expected Compliance Check Results

The compliance check will report multiple violations:

- Prohibited tag detection
- Uppercase tag key warnings
- Invalid email format errors
- Incorrect tag value formats
- Missing or improperly formatted required tags

## Learning Objectives

- Understand complex tag compliance rules
- Recognize common tagging mistakes
- Learn best practices for resource tagging

## Recommended Fixes

1. Remove temporary and test tags
2. Use lowercase tag keys
3. Provide a valid company email
4. Use allowed tag values
5. Follow specified tag formatting rules

## Usage

```bash
# Run Terraform plan and compliance check
./run.sh plan

# Apply infrastructure with compliance validation
./run.sh apply

# Destroy infrastructure
./run.sh destroy
```
