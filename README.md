# AWS Taggy: Cloud Resource Tag Compliance Automation

## 🌟 Project Overview

AWS Taggy is an advanced CLI tool designed to automate and enforce tag compliance across cloud resources, with a primary focus on AWS infrastructure. The tool provides a comprehensive solution for managing, validating, and ensuring consistent tagging standards.

### Key Features

- 🏷️ Comprehensive tag validation through a flexible configuration file, for simple and more complex compliance rules (suitable for all kind of companies).
- 🔍 Discover/Inspect resources in your AWS account without a configuration, checking which ones are tagged, which aren't, or querying attributes of resources.
- 🌎 Multi-resource type support (RDS, S3, SNS, CloudWatch Logs, EC2, etc). More resources will be added in the future.
- 📊 Detailed compliance reporting (table, JSON, YAML, or directly in your `clipboard`)

## 🎯 Use Case

In modern cloud environments, maintaining consistent and meaningful resource tagging is crucial for:

- Cost allocation, and FinOps.
- Resource management. Just ensuring governance, specially when dealing with complex IaaC setups.
- Security compliance
- Operational efficiency

AWS Taggy solves these challenges by:

- Enforcing predefined tagging standards through a [configuration file](./docs/tag-compliance.yaml)
- Detecting and reporting non-compliant resources

## 🚀 Quick Start

### Installation

Using [Homebrew](https://brew.sh/):

```bash
brew tap excoriate/tap
brew install aws-taggy
```

---

## 📄 License

[MIT License](./LICENSE)

## 🔮 Roadmap

- [ ] Multi-cloud support
- [ ] Add support for AWS resources: SQS, Redshift, SES, SSM, EKS, ECS.
