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

