# AWS Taggy: Cloud Resource Tag Compliance Automation

## 🌟 Project Overview

AWS Taggy is an advanced CLI tool designed to automate and enforce tag compliance across cloud resources, with a primary focus on AWS infrastructure. The tool provides a comprehensive solution for managing, validating, and ensuring consistent tagging standards.

### Key Features

- 🏷️ Comprehensive tag validation through a flexible configuration file, for simple and more complex compliance rules (suitable for all kind of companies).
- 🔍 Discover/Inspect resources in your AWS account without a configuration, checking which ones are tagged, which aren't, or querying attributes of resources.
- 🌎 Multi-resource type support (RDS, S3, SNS, CloudWatch Logs, EC2, etc). More resources will be added in the future.
- 📊 Detailed compliance reporting (table, JSON, YAML, or directly in your `clipboard`)

### 🎯 Use Case

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
brew tap Excoriate/aws-taggy
# or also
brew tap Excoriate/homebrew-tap https://github.com/Excoriate/homebrew-tap.git
# And then install the cli
brew install aws-taggy
```

### Developer Experience 🌿

#### Prerequisites

- [Nix](https://nixos.org/download.html)
- [direnv](https://direnv.net/) (optional but recommended)
- [Just](https://github.com/casey/just)

#### Getting Started

1. **Automatic Environment Setup (Recommended)**:

   ```bash
   # If using direnv
   direnv allow
   ```

2. **Manual Nix Shell**:
   ```bash
   # Start the development shell
   just nix-shell
   ```

#### Available Commands

- `just nix-shell`: Start the Nix development shell
- `just ci`: Run the CI pipeline entirely, locally through Nix.

---

## 📚 Documentation

| Directory                  | Description                                             | Contents                                                                                                                     |
| -------------------------- | ------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------- |
| `docs/how-it-works/`       | Technical deep-dive into AWS Taggy's internal mechanics | - Compliance check flow documentation                                                                                        |
| `docs/user-guide/`         | Step-by-step guides for using AWS Taggy                 | - How to configure tag compliance<br>- How to query resources<br>- How to discover resources<br>- Tag compliance usage guide |
| `docs/examples/`           | Real-world configuration and usage examples             | - Sample configuration files<br>- S3 tag scanning scenarios                                                                  |
| `docs/tag-compliance.yaml` | Comprehensive tag compliance configuration template     | Detailed example of a full tag compliance configuration                                                                      |

For more details, explore the documentation in each directory.

## 📦 Quick Guide

### Resource Discovery

`aws-taggy` allows you (depending on your credentials) to discover resources in your AWS account.

```bash
aws-taggy discover <options>
# discover all the S3 buckets across your account.
aws-taggy discover --service s3
# discover all the S3 buckets, in a given region, and copy the result as a valid YAML in your clipboard.
aws-taggy discover --service s3 --region us-east-1 --clipboard
```

## 📄 License

[MIT License](./LICENSE)

## 🔮 Roadmap

- [ ] Multi-cloud support
- [ ] Add support for AWS resources: SQS, Redshift, SES, SSM, EKS, ECS.
