# AWS Taggy: Cloud Resource Tag Compliance Automation

## 🌟 Project Overview

AWS Taggy is an advanced CLI tool designed to automate and enforce tag compliance across cloud resources, with a primary focus on AWS infrastructure. The tool provides a comprehensive solution for managing, validating, and ensuring consistent tagging standards.

### Key Features

- 🏷️ Comprehensive tag validation
- 🔍 Multi-resource type support
- 📊 Detailed compliance reporting
- 🚀 Easy integration with existing infrastructure
- 🛡️ Customizable compliance rules

## 🎯 Use Case

In modern cloud environments, maintaining consistent and meaningful resource tagging is crucial for:

- Cost allocation
- Resource management
- Security compliance
- Operational efficiency

AWS Taggy solves these challenges by:

- Enforcing predefined tagging standards
- Detecting and reporting non-compliant resources
- Providing actionable insights for tag improvements

## 🛠️ Project Structure

```
aws-taggy/
├── cli/           # CLI application source
├── pkg/           # Core package implementations
│   ├── cloud/     # Cloud provider interactions
│   ├── compliance/# Tag validation logic
│   └── ...
├── scripts/       # Utility scripts
│   ├── run_me.sh          # Generic compliance check script
│   └── terraform_manage.sh# Terraform management script
└── tests/
    └── examples/  # Example scenarios and test cases
        └── example-s3-specific-tags/
            ├── run.sh     # Example-specific workflow script
            └── ...
```

## 📦 Prerequisites

- Go 1.23+
- Terraform
- AWS CLI
- AWS Account

## 🚀 Quick Start

### Installation

```bash
# Clone the repository
git clone https://github.com/Excoriate/aws-taggy.git
cd aws-taggy

# Build the CLI
go build -o aws-taggy cli/main.go
```

### Running Examples

Each example in `tests/examples/` demonstrates a specific tagging compliance scenario.

#### S3 Bucket Tag Compliance Example

```bash
# Navigate to the S3 example
cd tests/examples/example-s3-specific-tags

# Run full example workflow (apply, check compliance, destroy)
./run.sh

# Run specific modes
./run.sh terraform     # Only apply Terraform
./run.sh compliance    # Only run compliance check
./run.sh destroy       # Destroy resources
```

#### Example Script Modes

The `run.sh` script supports multiple modes:

- `all` (default): Full workflow

  - Apply Terraform
  - Run compliance check
  - Destroy resources

- `terraform`: Only apply Terraform resources
- `compliance`: Only run tag compliance check
- `destroy`: Remove all created resources

### Compliance Check Modes

The underlying `run_me.sh` script supports multiple compliance check modes:

- `check`: Default compliance validation
- `validate`: Validate configuration
- `report`: Generate detailed report

### Example Compliance Check

```bash
# Run compliance check with different outputs
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

## 📝 Configuration

Tag compliance is defined in `tag-compliance.yaml`:

- Specify required tags
- Define validation rules
- Set compliance levels
- Configure notification channels

### Sample Configuration

```yaml
version: "1.0"
global:
  default_compliance_level: standard

resources:
  s3:
    enabled: true
    tag_criteria:
      minimum_required_tags: 5
      required_tags:
        - Environment
        - Owner
        - Project
```

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Commit your changes
4. Push to the branch
5. Create a Pull Request

## 📄 License

MIT License

## 🛡️ Best Practices

1. Use consistent, meaningful tag values
2. Follow naming conventions
3. Automate tag enforcement
4. Regularly audit resource tags

## 🔮 Roadmap

- [ ] Multi-cloud support
- [ ] Enhanced reporting capabilities
- [ ] More resource type integrations
- [ ] Custom compliance rule engine

## 💬 Support

Open an issue in the GitHub repository for any questions or problems.
