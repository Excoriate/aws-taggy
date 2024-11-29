# S3 Non-Compliant Tags Example

## Scenario Overview

This example demonstrates multiple tag compliance violations for an AWS S3 bucket, showcasing various ways resources can fail tag compliance checks.

## Compliance Violations

### 1. Tag Count Violation

- Total tags: 10 (exceeds maximum allowed of 8)

### 2. Prohibited Tags (3 violations)

- `temp:test`: Temporary tags are explicitly forbidden
- `test:example`: Test-related tags are not allowed
- `random:tag`: Random tags are not permitted

### 3. Invalid Tag Key Format (7 violations)

Tags must be lowercase and start with a letter. Found:

- `ENVIRONMENT`: Uppercase not permitted
- `Owner`: Incorrect capitalization
- `Sensitive`: Incorrect capitalization
- `RANDOM_TAG`: Uppercase not permitted
- And other non-compliant keys

### 4. Case Sensitivity Violations

Multiple tags violate case sensitivity rules:

- `ENVIRONMENT`: Should be lowercase
- `Owner`: Should be lowercase
- `Sensitive`: Should be lowercase

## Compliance Rules

The example enforces strict tagging rules:

- Maximum 8 tags per resource
- Lowercase tag keys only (`^[a-z][a-z0-9_-]*$`)
- No tags with prefixes: "temp:", "test:", "random:"
- Required tags:
  - project
  - environment
  - owner
  - data_class
  - cost_center

## Expected Compliance Check Results

The compliance check will report multiple violations:

```
📊 Compliance Summary:
Total Resources: 1
Compliant: 0
Non-Compliant: 1

Violation Types:
🚨 excess_tags: 1 occurrences
🚨 prohibited_tag: 3 occurrences
🚨 invalid_key_format: 7 occurrences
🚨 case_violation: 1 occurrences
```

## Learning Objectives

- Understand and identify common tag compliance violations
- Learn AWS resource tagging best practices
- Experience with real-world tag validation scenarios

## Usage

```bash
# Create infrastructure and run compliance check
./run.sh run

# Run compliance check on existing infrastructure
./run.sh run-cli

# Destroy infrastructure
./run.sh destroy
```

## Recommended Fixes

1. Reduce total number of tags to 8 or fewer
2. Remove prohibited tags:
   - Remove `temp:test`
   - Remove `test:example`
   - Remove `random:tag`
3. Fix tag key formatting:
   - Convert all uppercase tags to lowercase
   - Ensure all tag keys start with a letter
4. Follow case sensitivity rules consistently
