# Tag Compliance Validation Package

## Overview

The `compliance` package provides a robust and flexible tag validation system for AWS resources. It implements comprehensive tag compliance checks to ensure resources meet organizational tagging standards, governance requirements, and best practices.

## Validation Implementation Details

### Package Structure

```
pkg/compliance/
├── validator.go        # Core validation logic
├── result.go           # Compliance result handling
├── rules.go            # Violation type and rule definitions
└── README.md           # Package documentation
```

### Key Validation Functions

| Validation Type   | Function                    | Location       | Description                                               |
| ----------------- | --------------------------- | -------------- | --------------------------------------------------------- |
| Prohibited Tags   | `checkProhibitedTags()`     | `validator.go` | Prevents tags with specific substrings or prefixes        |
| Key Format        | `validateKeyFormat()`       | `validator.go` | Validates tag key formatting using regex patterns         |
| Key Prefix/Suffix | `validateKeyPrefixSuffix()` | `validator.go` | Checks allowed prefixes, suffixes, and key length         |
| Value Characters  | `validateValueCharacters()` | `validator.go` | Validates allowed characters and blocks disallowed values |
| Value Length      | `validateValueLength()`     | `validator.go` | Enforces per-tag length constraints                       |
| Allowed Values    | `validateAllowedValues()`   | `validator.go` | Restricts tag values to predefined lists                  |
| Case Sensitivity  | `validateCaseRules()`       | `validator.go` | Handles case transformation and validation                |
| Pattern Matching  | `validatePatternRules()`    | `validator.go` | Applies regex-based validation for specific tags          |

### Main Validation Entry Point

The primary method for tag validation is `ValidateTags()` in `validator.go`:

```go
func (v *TagValidator) ValidateTags(tags map[string]string) *ComplianceResult {
    result := &ComplianceResult{
        IsCompliant:  true,
        Violations:   []Violation{},
        ResourceTags: tags,
    }

    // Validation method calls
    v.checkTagCount(tags, result)
    v.checkRequiredTags(tags, result)
    v.checkProhibitedTags(tags, result)
    v.validateKeyPrefixSuffix(tags, result)
    v.validateValueCharacters(tags, result)
    v.validateCaseSensitivity(tags, result)
    v.validateKeyFormat(tags, result)
    v.validateValueLength(tags, result)
    v.validateCaseRules(tags, result)
    v.validateAllowedValues(tags, result)
    v.validatePatternRules(tags, result)

    return result
}
```

### Violation Handling

Key classes for violation management:

- `Violation` struct (in `result.go`): Represents individual tag violations
- `ComplianceResult` struct (in `result.go`): Aggregates validation results
- `Summary` struct (in `result.go`): Provides comprehensive compliance summary

### Configuration Integration

The validator uses configuration from `pkg/configuration/config.go`:

- Loads validation rules from YAML configuration
- Supports dynamic rule updates
- Provides flexible, configuration-driven validation

## Validation Coverage

The package supports full coverage of tag validation rules as specified in the configuration schema:

### 1. Prohibited Tags Validation

- Prevents tags containing specific substrings or prefixes
- Blocks tags like "aws:", "internal:", "temp:", "test:"
- Helps maintain clean and standardized tag namespaces

### 2. Key Format Validation

- Enforces tag key formatting rules using regex patterns
- Supports multiple validation rules per configuration
- Validates:
  - Lowercase letter start
  - Allowed characters (letters, numbers, underscores, hyphens)
  - Maximum key length (up to 128 characters)
- Provides custom error messages for each rule

### 3. Key Validation

- Checks tag key prefixes and suffixes
- Supports predefined allowed prefixes (e.g., "project-", "env-")
- Supports predefined allowed suffixes (e.g., "-prod", "-dev")
- Enforces maximum key length
- Prevents non-compliant tag key structures

### 4. Value Validation

- Validates tag value character sets
- Restricts allowed characters (e.g., alphanumeric, dots, underscores)
- Blocks disallowed values like "undefined", "null", "none"
- Ensures meaningful and specific tag content

### 5. Length Constraints

- Implements per-tag length validation
- Supports minimum and maximum length rules
- Provides tag-specific length constraints
- Examples:
  - Environment tag: 2-15 characters
  - Owner tag: 3-50 characters
  - Project tag: 4-30 characters

### 6. Allowed Values

- Restricts tag values to predefined lists
- Supports tag-specific allowed value sets
- Examples:
  - Environment: production, staging, development
  - DataClassification: public, private, confidential
  - SecurityLevel: high, medium, low

### 7. Case Sensitivity

- Multiple case validation modes:
  - Strict: Exact case matching
  - Relaxed: Case-insensitive matching
- Case transformation rules:
  - Lowercase enforcement
  - Uppercase enforcement
  - Mixed case with optional pattern validation
- Supports custom case patterns

### 8. Pattern Matching

- Advanced regex-based validation
- Supports complex pattern rules for specific tags
- Examples:
  - CostCenter: Specific format like AA-1234
  - ProjectCode: PRJ-12345 format
  - Owner: Company email validation

## Validation Mechanisms

- Global rules applied across all resources
- Resource-type specific overrides
- Compliance level-based validation
- Detailed violation tracking and reporting

## Violation Handling

Supports comprehensive violation types:

1. Missing tags
2. Case violations
3. Invalid values
4. Pattern mismatches
5. Prohibited tags
6. Length constraint violations
7. Invalid key formats

## Performance and Scalability

- In-memory validation
- Concurrent processing support
- Minimal computational overhead
- Easily extensible validation framework

## Security Considerations

- Prevents sensitive tag configurations
- Enforces organizational tagging standards
- Provides governance and compliance tracking

## Future Roadmap

- Enhanced machine learning tag recommendations
- More granular compliance reporting
- Advanced pattern matching capabilities
- Cloud cost management integrations

## Best Practices

1. Define clear, comprehensive tagging standards
2. Regularly update validation rules
3. Use compliance levels strategically
4. Implement resource-type specific rules
5. Leverage detailed violation reporting

## Advanced Usage Example

```go
// Create a tag validator with configuration
config := loadTagComplianceConfig()
validator := compliance.NewTagValidator(config)

// Validate resource tags
resourceTags := map[string]string{
    "Environment": "production",
    "Owner": "team@company.com",
}

// Perform comprehensive validation
result := validator.ValidateTags(resourceTags)

// Handle violations
if !result.IsCompliant {
    for _, violation := range result.Violations {
        fmt.Printf("Violation: %s - %s\n", violation.Type, violation.Message)
    }
}

// Generate compliance summary
summary := compliance.GenerateSummary([]*ComplianceResult{result})
fmt.Printf("Compliance Rate: %d%%\n", summary.ComplianceRate)
```

## Debugging and Logging

- Use verbose logging in `pkg/o11y` for detailed validation insights
- Enable debug mode in configuration for comprehensive validation traces

## Extension Points

- Implement custom `Validator` interface for specialized validations
- Add new violation types in `rules.go`
- Extend `ComplianceResult` for additional metadata

## Performance Optimization

- Validations are executed concurrently
- Minimal memory allocation
- Efficient regex compilation and caching
