# Multi-platform builder stage
FROM --platform=$BUILDPLATFORM golang:1.23-alpine AS builder

# Set build arguments for multi-arch support
ARG TARGETPLATFORM
ARG BUILDPLATFORM
ARG VERSION=devel

WORKDIR /app

# Install git for version tagging
RUN apk add --no-cache git

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the entire project
COPY . .

# Determine GOOS and GOARCH based on TARGETPLATFORM
RUN if [ -z "$GOOS" ] || [ -z "$GOARCH" ]; then \
    echo "GOOS or GOARCH is not set. Attempting to determine from TARGETPLATFORM: $TARGETPLATFORM"; \
    case "$TARGETPLATFORM" in \
        "linux/amd64") export GOOS=linux GOARCH=amd64 ;; \
        "linux/arm64") export GOOS=linux GOARCH=arm64 ;; \
        *) echo "Unsupported platform: $TARGETPLATFORM" && exit 1 ;; \
    esac; \
fi

RUN go mod tidy && go mod verify

RUN go build -ldflags="-X 'main.version=$VERSION'" -o /app/aws-taggy ./cli/main.go

# Debug: list files and show go version
RUN go version && ls -la /app

# Final lightweight image
FROM --platform=$TARGETPLATFORM alpine:3.18

WORKDIR /app

# Install AWS CLI and required tools
RUN apk add --no-cache aws-cli

# Copy the compiled binary
COPY --from=builder /app/aws-taggy /usr/local/bin/aws-taggy

# Set entrypoint
ENTRYPOINT ["/usr/local/bin/aws-taggy"]

# Default command (can be overridden)
CMD ["--help"]
