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

*AWS Taggy* allows you (depending on your credentials) to discover resources in your AWS account.

```bash
aws-taggy discover <options>
# discover all the S3 buckets across your account.
aws-taggy discover --service s3
# discover all the S3 buckets, in a given region, and copy the result as a valid YAML in your clipboard.
aws-taggy discover --service s3 --region us-east-1 --clipboard
```

> NOTE: If you need to output a file in `json`, `yaml` or directly into your `clipboard`, you can use the `--output` flag.

```bash
aws-taggy discover --service s3 --region us-east-1 --output yaml --clipboard
```

### Query Tags on existing resources

*AWS Taggy* allows you to query tags on existing resources. You can use a combination of the `discover` commands, to get the resource's ARN, and then use the `query` command to get the tags.

```bash
aws-taggy query tags --service=s3 --arn arn:aws:s3:::contactservice-microserv-serverlessdeploymentbuck-1bhyuu --clipboard
```

### Create a new tag compliance configuration file

*AWS Taggy* allows you to create a new tag compliance configuration file, that you can customize to your needs. See this [link](./docs/tag-compliance.yaml) for more details, and this [guide](./docs/user-guide/how-to-configure-tag-compliance.md) to learn how to configure, and this [guide](./docs/how-it-works/compliance-check-flow.md) to learn how the compliance check works.

```bash
# Create a file in the current directory.
aws-taggy config generate --output .aws-taggy-tag-compliance.yaml
```

The file, when created, can easily be customized to your needs. If so, you can also use aws-taggy to validate if it's a valid configuration file, and if it's not, it will return a detailed error message, with the exact line and column where the error is.

```bash
# A configuration file is expected to be provided.
aws-taggy config validate --config .aws-taggy-tag-compliance.yaml
```

### Run the compliance check

The most relevant part of *AWS Taggy* is the compliance check. This is where the magic happens. You can run the compliance check for a given configuration file, and it will return a detailed report of the compliance of your resources.

```bash
aws-taggy compliance check --config .aws-taggy-tag-compliance.yaml
```

In the [examples](./docs/examples/) directory, you can find a sample configuration file, and a sample output of the compliance check, the terraform files to generate the resources used in those examples, and a `README.md` file that explain the scenario expressed in the example.







## 📄 License

[MIT License](./LICENSE)

## 🔮 Roadmap

- [ ] Multi-cloud support
- [ ] Add support for AWS resources: SQS, Redshift, SES, SSM, EKS, ECS.
