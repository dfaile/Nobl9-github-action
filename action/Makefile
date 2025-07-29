# Makefile for Nobl9 GitHub Action

# Variables
IMAGE_NAME = nobl9-action
IMAGE_TAG = latest
REGISTRY = ghcr.io/your-org

# Go variables
GOOS = linux
GOARCH = amd64
CGO_ENABLED = 0

# Build flags
LDFLAGS = -w -s -extldflags '-static'

.PHONY: help build test clean docker-build docker-push docker-run lint fmt

# Default target
help:
	@echo "Available targets:"
	@echo "  build        - Build the Go binary"
	@echo "  test         - Run tests"
	@echo "  clean        - Clean build artifacts"
	@echo "  docker-build - Build Docker image"
	@echo "  docker-push  - Push Docker image to registry"
	@echo "  docker-run   - Run Docker container locally"
	@echo "  lint         - Run linter"
	@echo "  fmt          - Format code"

# Build the Go binary
build:
	@echo "Building nobl9-action binary..."
	GOOS=$(GOOS) GOARCH=$(GOARCH) CGO_ENABLED=$(CGO_ENABLED) go build \
		-ldflags="$(LDFLAGS)" \
		-o nobl9-action \
		./cmd/main.go
	@echo "Binary built successfully: nobl9-action"

# Run tests
test:
	@echo "Running tests..."
	go test -v ./...
	@echo "Tests completed"

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -f nobl9-action
	rm -rf dist/
	@echo "Clean completed"

# Build Docker image
docker-build:
	@echo "Building Docker image..."
	docker build -t $(IMAGE_NAME):$(IMAGE_TAG) .
	@echo "Docker image built: $(IMAGE_NAME):$(IMAGE_TAG)"

# Build Docker image for specific registry
docker-build-registry:
	@echo "Building Docker image for registry..."
	docker build -t $(REGISTRY)/$(IMAGE_NAME):$(IMAGE_TAG) .
	@echo "Docker image built: $(REGISTRY)/$(IMAGE_NAME):$(IMAGE_TAG)"

# Push Docker image to registry
docker-push:
	@echo "Pushing Docker image to registry..."
	docker push $(REGISTRY)/$(IMAGE_NAME):$(IMAGE_TAG)
	@echo "Docker image pushed: $(REGISTRY)/$(IMAGE_NAME):$(IMAGE_TAG)"

# Run Docker container locally
docker-run:
	@echo "Running Docker container..."
	docker run --rm -it $(IMAGE_NAME):$(IMAGE_TAG) --help

# Run linter
lint:
	@echo "Running linter..."
	golangci-lint run
	@echo "Linting completed"

# Format code
fmt:
	@echo "Formatting code..."
	go fmt ./...
	@echo "Code formatting completed"

# Install dependencies
deps:
	@echo "Installing dependencies..."
	go mod download
	go mod tidy
	@echo "Dependencies installed"

# Verify dependencies
deps-check:
	@echo "Checking dependencies..."
	go mod verify
	@echo "Dependencies verified"

# Development setup
dev-setup: deps fmt lint test
	@echo "Development setup completed"

# Release build
release: clean build test docker-build
	@echo "Release build completed" 