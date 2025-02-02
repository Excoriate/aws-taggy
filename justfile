# aws-taggy Justfile 🏷️🚀
# Manages build, test, and development workflows for the aws-taggy CLI

# Default project name
projectname := "aws-taggy"

# Display help information 📖
default:
    @just --list

# List all available commands with descriptions 🔍
help:
    @just --list

# Build the Go binary 🛠️
build:
    @go build -ldflags "-X cmd.version=$(git describe --abbrev=0 --tags || echo devel)" -o {{projectname}}

# Install the Go binary globally 📦
install:
    @go install -ldflags "-X main.version=$(git describe --abbrev=0 --tags)"

# Run the application directly 🚀🔧 Support arguments.
run *args:
    @echo "🌟 Launching aws-taggy CLI in Developer Mode 🖥️"
    @echo "🔍 Running from local source code..."
    @go run cli/main.go {{args}}

# Bootstrap development environment 🔧
bootstrap:
    @go generate -tags tools tools/tools.go

# Run tests with coverage reporting 🧪
test: clean
    @go test --cover -parallel=1 -v -coverprofile=coverage.out ./...
    @go tool cover -func=coverage.out | sort -rnk3

# Clean up build artifacts and temporary files 🧹
clean:
    @rm -rf coverage.out dist/ {{projectname}}

# Generate detailed test coverage report 📊
cover:
    @go test -v -race $(go list ./... | grep -v /vendor/) -v -coverprofile=coverage.out
    @go tool cover -func=coverage.out

# Format Go source code 🖌️
fmt:
    @gofumpt -w .
    @gci write .

# Run linters to ensure code quality 🕵️
lint:
    @golangci-lint run -c .golang-ci.yml

# Run pre-commit hooks for code quality checks 🏁
run-hooks:
    @echo "Updating pre-commit hooks 🧼"
    @pre-commit autoupdate
    @pre-commit run --all-files

# Run an example in the tests/examples directory 📚🔍
run-example dir mode:
    @echo "🚀 Running example in: {{dir}} 🔍"
    @./tests/examples/{{dir}}/run.sh {{mode}}

# Docker-related commands 🐳
# Build Docker image for Apple Silicon (arm64)
docker-build-arm:
    @docker buildx build \
        --platform linux/arm64 \
        -t aws-taggy:arm64 \
        --build-arg VERSION=$(git describe --abbrev=0 --tags 2>/dev/null || echo devel) \
        --load \
        .

# Build Docker image for Linux (amd64)
docker-build-linux:
    @docker buildx build \
        --platform linux/amd64 \
        -t aws-taggy:amd64 \
        --build-arg VERSION=$(git describe --abbrev=0 --tags 2>/dev/null || echo devel) \
        --load \
        .

# Build multi-platform Docker image
docker-build-multi:
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

# GitHub Actions-like Lint Workflow 🕵️
ci-lint:
    @echo "🔍 Running Golangci-Lint (GitHub Actions Style)"
    @golangci-lint run \
        --config .golangci.yml \
        --timeout=5m \
        --verbose

# GitHub Actions-like Test Workflow 🧪
ci-test:
    @echo "🚀 Running Tests with Coverage (GitHub Actions Style)"
    @go test \
        -race \
        -coverprofile="coverage.out" \
        -covermode=atomic \
        -parallel=1 \
        -v \
        ./...

# Comprehensive CI Check (Lint + Test) 🏁
ci-check: ci-lint ci-test
    @echo "✅ All CI checks passed successfully!"

# Nix Development Shell 🌿
# Commands for managing Nix development environment

# Start Nix development shell 🚀
nix-shell:
    @echo "🌿 Starting Nix Development Shell for AWS Taggy 🏷️"
    @nix develop . --extra-experimental-features nix-command --extra-experimental-features flakes

# Update Nix flake dependencies 🔄
nix-update:
    @echo "🔄 Updating Nix Flake Dependencies"
    @nix flake update

# Build project using Nix 🛠️
nix-build:
    @echo "🛠️ Building AWS Taggy with Nix"
    @nix build

# Clean Nix build artifacts 🧹
nix-clean:
    @echo "🧹 Cleaning Nix Build Artifacts"
    @nix clean
    @rm -rf result
