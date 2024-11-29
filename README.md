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

AWS Taggy includes interactive examples to demonstrate tag compliance scenarios.

#### Prerequisites

- Terraform
- AWS CLI
- Go 1.23+
- AWS Credentials

#### S3 Tag Compliance Example

##### Available Modes

```bash
# Create Terraform resources
just run-example 1-s3-specific-tags create

# Plan Terraform changes
just run-example 1-s3-specific-tags plan

# Run full scenario (create + compliance check)
just run-example 1-s3-specific-tags run

# Run compliance check on existing resources
just run-example 1-s3-specific-tags run-cli

# Destroy resources
just run-example 1-s3-specific-tags destroy
```

##### Direct CLI Execution

```bash
# Run compliance check from source code
go run cli/main.go compliance check \
  --config tests/examples/1-s3-specific-tags/tag-compliance.yaml \
  --resource aws-taggy \
  --output=table \
  --detailed
```

##### Example Configuration

- Location: `tests/examples/1-s3-specific-tags/`
- Terraform Config: `main.tf`
- Compliance Rules: `tag-compliance.yaml`
- Resource Name: `aws-taggy`

### AWS Credentials Setup

```bash
# Option 1: AWS CLI Configuration
aws configure

# Option 2: Environment Variables
export AWS_ACCESS_KEY_ID='your_access_key'
export AWS_SECRET_ACCESS_KEY='your_secret_key'
export AWS_REGION='us-east-1'
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
