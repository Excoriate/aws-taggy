## Project Structure

```
.
├── .cursorrules
├── .direnv
│ └── flake-profile-a5d5b61aa8a61b7d9d765e1daf971a9a578f1cfa.rc
├── .editorconfig
├── .gitattributes
├── .github
│ ├── CODEOWNERS
│ ├── ISSUE_TEMPLATE
│ │ ├── bug_report.md
│ │ └── feature_request.md
│ ├── PULL_REQUEST_TEMPLATE.md
│ ├── auto-comment.yml
│ ├── config.yml
│ ├── dependabot.yml
│ ├── labeler.yml
│ ├── no-response.yml
│ ├── pr-labeler.yml
│ ├── settings.yml
│ ├── stale.yml
│ └── workflows
│ ├── go-ci-lint.yaml
│ ├── go-ci-test.yaml
│ ├── issue-comment-created.yml
│ ├── labels-assigner.yml
│ ├── lock-threads.yml
│ ├── release.yml
│ ├── semantic-pr.yml
│ └── test.yml
├── .gitignore
├── .golang-ci.yml
├── .golangci.yml
├── .goreleaser.yml
├── .pre-commit-config.yaml
├── .release-please-manifest.json
├── AI
│ ├── ai-conventions.md
│ ├── ai-project-overview.md
│ └── ai-project-structure.md
├── CONTRIBUTING.md
├── LICENSE
├── README.md
├── cli
│ ├── cmd
│ │ └── root.go
│ ├── go.mod
│ ├── internal
│ │ ├── configuration
│ │ │ └── app.go
│ │ └── tui
│ │ └── banner.go
│ └── main.go
├── codecov.yml
├── docs
│ └── tag-compliance.yaml
├── go.mod
├── go.work
├── go.work.sum
├── justfile
├── pkg
│ └── o11y
│ ├── logger.go
│ └── logger_test.go
├── release-please-config.json
├── scripts
│ ├── completions.sh
│ └── manpages.sh
└── tools
└── tools.go
```

## Technology Stack and Tools

### Primary Technologies

- **Language**: Go (Golang)
- **Cloud Provider**: AWS
- **Configuration**: YAML-based configuration management

### Development Tools

- **Package Management**: Go Modules
- **Linting**: golangci-lint
- **Testing**: Go's built-in testing framework
- **CI/CD**: GitHub Actions
- **Release Management**: GoReleaser
- **Binary Distribution**: Homebrew through GoReleaser.

### Key Dependencies

- **AWS SDK**: aws-sdk-go-v2
- **Configuration Parsing**: gopkg.in/yaml.v3
- **Logging**: Zap or standard Go log
- **CLI Framework**: Kong https://pkg.go.dev/github.com/alecthomas/kong@v1.4.0#section-readme

### Development Workflow

It's a mono repo, with a go.work with two modules:

1. cli/ -> https://github.com/Excoriate/aws-taggy/cli
2. (root) -> https://github.com/Excoriate/aws-taggy

- **Version Control**: Git
- **Dependency Management**: Go Modules
- **Pre-commit Hooks**: Configured for code quality
- **Semantic Versioning**: Implemented via release-please
- **Nix**: direnv, and flakes.nix.
- **Task Management**: Justfile https://just.systems/

### Deployment and Distribution

- **Containerization**: Docker

### Observability

- **Logging**: Custom logging package in `pkg/o11y`

### Key Project Characteristics

- Modular architecture
- Extensible resource scanning
- Configuration-driven design
- Cloud-native approach to resource governance
