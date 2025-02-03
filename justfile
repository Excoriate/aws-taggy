# Default project name
projectname := "aws-taggy"

# Enable dotenv loading
set dotenv-load

# Display help information 📖
default:
    @just --list

# List all available commands with descriptions 🔍
help:
    @just --list

# Build the Go binary 🛠️
build: clean-build
    @echo "🚀 Building AWS Taggy CLI..."
    @go mod tidy
    @cd cli && go build -ldflags "-X cmd.version=$(git describe --abbrev=0 --tags 2>/dev/null || echo devel)" -o ../{{projectname}}
    @echo "🚀 AWS Taggy CLI built successfully!"

# Clean the build directory 🧹
clean-build:
    @echo "🧹 Cleaning AWS Taggy CLI build directory..."
    @if [ -f "{{projectname}}" ]; then rm "{{projectname}}"; fi
    @echo "🧹 AWS Taggy CLI compiled binary removed successfully!"

# Run the application directly 🚀🔧 Support arguments.
run *args:
    @echo "🌟 Launching aws-taggy CLI in Developer Mode 🖥️"
    @echo "🔍 Running from local source code..."
    @go run cli/main.go {{args}}

# Run the application directly 🚀🔧 Support arguments.
runbin *args: build
    @./{{projectname}} {{args}}

# Bootstrap development environment 🔧
bootstrap:
    @go generate -tags tools tools/tools.go

# Run tests with coverage reporting 🧪
test: clean
    @go test --cover -parallel=1 -v -coverprofile=coverage.out ./...
    @go tool cover -func=coverage.out | sort -rnk3

# Clean up build artifacts and temporary files 🧹
clean: clean-build
    @echo "🧹 Cleaning coverage.out, dist/ and compiled binary..."
    @rm -rf coverage.out dist/ {{projectname}}

# Generate detailed test coverage report 📊
cover:
    @go test -v -race $(go list ./... | grep -v /vendor/) -v -coverprofile=coverage.out
    @go tool cover -func=coverage.out

# Format Go source code 🖌️
fmt:
    @echo "📜 Formatting Go source code..."
    @echo "✅ Formatting complete. Check formatted_files.log for details."
    @gofumpt -w .
    @go fmt ./...

# Run linters to ensure code quality 🕵️
lint:
    @golangci-lint run --config=./.golangci.yml --timeout=5m --verbose

# Run pre-commit hooks for code quality checks 🏁
run-hooks:
    @echo "Updating pre-commit hooks 🧼"
    @pre-commit autoupdate
    @pre-commit run --all-files

# Docker-related commands 🐳
# Build Docker image for Apple Silicon (arm64)
build-docker-arm:
    @docker buildx build \
        --platform linux/arm64 \
        -t aws-taggy:arm64 \
        --build-arg VERSION=$(git describe --abbrev=0 --tags 2>/dev/null || echo devel) \
        --load \
        .

# Build Docker image for Linux (amd64)
build-docker-linux:
    @docker buildx build \
        --platform linux/amd64 \
        -t aws-taggy:amd64 \
        --build-arg VERSION=$(git describe --abbrev=0 --tags 2>/dev/null || echo devel) \
        --load \
        .

# Build multi-platform Docker image
build-docker-multi:
    @docker buildx build \
        --platform linux/amd64,linux/arm64 \
        -t aws-taggy:latest \
        --build-arg VERSION=$(git describe --abbrev=0 --tags 2>/dev/null || echo devel) \
        --push \
        .

# Run Docker container for Apple Silicon
rundocker-arm *args:
    @just docker-build-arm
    @docker run --rm \
        -v "$(HOME)/.aws:/root/.aws" \
        -e AWS_PROFILE \
        -e AWS_DEFAULT_REGION \
        -e AWS_ACCESS_KEY_ID \
        -e AWS_SECRET_ACCESS_KEY \
        -e AWS_SESSION_TOKEN \
        aws-taggy:arm64 {{args}}

# Run Docker container for Linux
rundocker-linux *args:
    @just docker-build-linux
    @docker run --rm \
        -v "$(HOME)/.aws:/root/.aws" \
        -e AWS_PROFILE \
        -e AWS_DEFAULT_REGION \
        -e AWS_ACCESS_KEY_ID \
        -e AWS_SECRET_ACCESS_KEY \
        -e AWS_SESSION_TOKEN \
        aws-taggy:amd64 {{args}}

# Clean up Docker resources
docker-clean:
    @docker rmi aws-taggy:arm64 2>/dev/null || true
    @docker rmi aws-taggy:amd64 2>/dev/null || true
    @docker rmi aws-taggy:latest 2>/dev/null || true
    @docker system prune -f

# GitHub Actions-like Go Workflow 🔍
ci-go: fmt lint build
    @echo "🔍 Running Go CI (fmt, lint, build)"

# Comprehensive CI Check (Lint + Test) 🏁
ci: ci-go test run-hooks
    @echo "✅ All CI checks passed successfully!"

# Nix Development Shell 🌿
# Commands for managing Nix development environment

# Start Nix development shell 🚀
nix-shell:
    @echo "🌿 Starting Nix Development Shell for AWS Taggy 🏷️"
    @nix develop . --extra-experimental-features nix-command --extra-experimental-features flakes

# Run Goreleaser to build the release artifacts
run-goreleaser:
    @goreleaser release --snapshot --clean

# Install Go development utilities
install-dev-tools:
    #!/usr/bin/env bash
    set -euo pipefail

    # Function to check and install a Go tool
    install_go_tool() {
        local tool_name="$1"
        local package="$2"

        if ! command -v "${tool_name}" &> /dev/null; then
            echo "🚀 Installing ${tool_name}..."
            go install "${package}"
        else
            echo "✅ ${tool_name} is already installed."
        fi
    }

    # Install gofumpt
    install_go_tool gofumpt mvdan.cc/gofumpt@latest

    # Install goimports
    install_go_tool goimports golang.org/x/tools/cmd/goimports@latest

    echo "🎉 Go development tools are ready!"
