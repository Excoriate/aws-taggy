# Compliance Package

## Overview

The `compliance` package provides a flexible and extensible tag validation system for AWS resources. It supports complex tag validation rules, including:

- Required tags
- Case sensitivity rules
- Allowed value validation
- Pattern matching

## Key Components

### Validator

The `Validator` interface defines the core method for tag validation:

```go
type Validator interface {
    ValidateTags(tags map[string]string) *ComplianceResult
}
```

### Compliance Levels

Three compliance levels are supported:

- `high`: Strictest validation
- `standard`: Moderate validation
- `low`: Relaxed validation

### Violation Types

- `missing_tags`: Required tags are missing
- `case_violation`: Tag case doesn't match rules
- `invalid_value`: Tag value not in allowed list
- `pattern_violation`: Tag doesn't match regex pattern

## Usage Example

```go
config := loadConfigFromFile()
validator := compliance.NewTagValidator(config)

resourceTags := map[string]string{
    "Environment": "production",
    "Owner": "team@company.com",
}

result := validator.ValidateTags(resourceTags)
if !result.IsCompliant {
    for _, violation := range result.Violations {
        fmt.Println(violation.Message)
    }
}
```

## Extensibility

The package is designed to be easily extended:

- Add new validation rules
- Implement custom validators
- Extend violation types

## Best Practices

1. Use configuration files to define validation rules
2. Implement comprehensive test coverage
3. Handle violations gracefully
