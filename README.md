# 🏷️ AWS Taggy

## Overview

AWS Taggy is a powerful CLI tool for comprehensive AWS resource tagging compliance and governance.

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
```

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