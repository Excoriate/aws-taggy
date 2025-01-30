# Configuring Tag Compliance with aws-taggy

## Overview

The `config` command provides powerful configuration management tools for the AWS Taggy CLI, helping you validate and generate tag compliance configuration files. These configuration files define your organization's tagging standards, compliance rules, and resource management policies.

## Command Syntax

```bash
aws-taggy config <command> [flags]
```

## Available Commands

### 1. `config validate`

Validate an existing tag compliance configuration file to ensure it meets the required standards.

#### Usage

```bash
aws-taggy config validate --config=path/to/config.yaml [flags]
```

#### Flags

- `--config`: Path to the tag compliance configuration file (required)
- `--output`: Output format for validation results
  - Options: `table` (default), `json`, `yaml`
- `--debug`: Enable detailed debug information

#### Example

```bash
# Validate a configuration file
aws-taggy config validate --config=tag-compliance.yaml

# Validate with JSON output
aws-taggy config validate --config=tag-compliance.yaml --output=json
```

#### What Validation Checks

- Configuration file syntax
- Required configuration sections
- Tag validation rules
- Compliance level definitions
- Resource-specific configurations

#### Validation Output

- Confirmation of file validity
- Summary of global and resource-specific settings
- Potential warnings or errors in configuration

### 2. `config generate`

Generate a sample tag compliance configuration file to help you get started with defining your tagging standards.

#### Usage

```bash
aws-taggy config generate [flags]
```

#### Flags

- `--directory`: Specify the output directory for the configuration file
- `--filename`: Custom filename for the generated configuration
- `--overwrite`: Overwrite existing configuration file if present

#### Example

```bash
# Generate a sample configuration in the current directory
aws-taggy config generate

# Generate a configuration in a specific directory with a custom filename
aws-taggy config generate --directory=./config --filename=my-tag-compliance.yaml
```

#### Generated Configuration Contents

The generated file includes:

- Version information
- Global tagging rules
- Resource-specific configurations
- Compliance level definitions
- Tag validation rules
- Notification settings

## Configuration File Structure

A typical `tag-compliance.yaml` file includes:

1. **Version**: Schema version for compatibility
2. **AWS Configuration**:
   - Region scanning mode
   - Batch processing settings
3. **Global Settings**:
   - Minimum required tags
   - Required and forbidden tags
4. **Resource-Specific Rules**:
   - Service-level tagging requirements
   - Exclusion patterns
5. **Compliance Levels**:
   - High and standard compliance definitions
6. **Tag Validation**:
   - Key and value constraints
   - Allowed values and patterns
7. **Notifications**:
   - Slack and email alert configurations

## Best Practices

1. Start with the generated configuration template
2. Customize the template to match your organization's standards
3. Validate the configuration before deployment
4. Regularly review and update your tagging rules
5. Use consistent naming conventions for tags
6. Define clear compliance levels

## Troubleshooting

- Ensure the configuration file is a valid YAML format
- Check for syntax errors in tag definitions
- Verify that required tags are correctly specified
- Use the `--debug` flag for detailed error information

## Example Configuration Scenarios

### Scenario 1: Basic Tagging Requirements

```yaml
global:
  minimum_required_tags: 3
  required_tags:
    - Environment
    - Owner
    - Project
```

### Scenario 2: S3 Bucket Specific Rules

```yaml
resources:
  s3:
    minimum_required_tags: 4
    required_tags:
      - DataClassification
      - BackupPolicy
      - Environment
      - Owner
```

## Related Commands

- `aws-taggy discover`: Find resources and their current tagging status
- `aws-taggy compliance check`: Validate resources against configuration

## Performance Considerations

- Large configuration files may increase validation time
- Complex tag validation rules can impact resource scanning performance

## Security

- Never commit configuration files with sensitive information to version control
- Use environment variables or secure secret management for sensitive data

## Limitations

- Configuration validation is syntax-based
- Actual tag compliance requires runtime resource scanning

## Version Compatibility

Ensure you're using the latest version of aws-taggy for the most up-to-date configuration features.

## Getting Help

- Use `aws-taggy config --help` for command-line help
- Consult documentation for detailed configuration guidelines
