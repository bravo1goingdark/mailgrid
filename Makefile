# Mailgrid Makefile
.PHONY: build clean test lint format security check install deps help

# Variables
GO_VERSION := 1.21
BINARY_NAME := mailgrid
PACKAGE := github.com/bravo1goingdark/mailgrid
BUILD_DIR := build
DIST_DIR := dist

# Build information
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT_HASH := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS := -ldflags "-X main.version=$(VERSION) -X main.commit=$(COMMIT_HASH) -X main.date=$(BUILD_DATE)"

# Default target
all: clean deps format lint test security build

# Help target
help: ## Show this help message
	@echo "Mailgrid Build System"
	@echo "Usage: make <target>"
	@echo ""
	@echo "Targets:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# Install dependencies
deps: ## Install project dependencies
	@echo "Installing dependencies..."
	go mod download
	go mod tidy
	go install golang.org/x/vuln/cmd/govulncheck@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Format code
format: ## Format Go code
	@echo "Formatting code..."
	go fmt ./...
	gofmt -s -w .

# Lint code
lint: ## Run linters
	@echo "Running linters..."
	go vet ./...
	golangci-lint run ./... || true

# Run tests
test: ## Run tests
	@echo "Running tests..."
	go test -v -race -cover ./...

# Security check
security: ## Run security checks
	@echo "Running security checks..."
	govulncheck ./...

# Type checking
check: ## Run all checks (format, lint, test, security)
	@echo "Running all checks..."
	$(MAKE) format
	$(MAKE) lint
	$(MAKE) test
	$(MAKE) security

# Build for current platform
build: ## Build binary for current platform
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) cmd/$(BINARY_NAME)/main.go

# Build for all platforms
build-all: ## Build binaries for all platforms
	@echo "Building for all platforms..."
	@mkdir -p $(DIST_DIR)
	
	# Linux AMD64
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build $(LDFLAGS) \
		-o $(DIST_DIR)/$(BINARY_NAME)-linux-amd64 cmd/$(BINARY_NAME)/main.go
	
	# Linux ARM64
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build $(LDFLAGS) \
		-o $(DIST_DIR)/$(BINARY_NAME)-linux-arm64 cmd/$(BINARY_NAME)/main.go
	
	# macOS AMD64
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build $(LDFLAGS) \
		-o $(DIST_DIR)/$(BINARY_NAME)-darwin-amd64 cmd/$(BINARY_NAME)/main.go
	
	# macOS ARM64
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build $(LDFLAGS) \
		-o $(DIST_DIR)/$(BINARY_NAME)-darwin-arm64 cmd/$(BINARY_NAME)/main.go
	
	# Windows AMD64
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build $(LDFLAGS) \
		-o $(DIST_DIR)/$(BINARY_NAME)-windows-amd64.exe cmd/$(BINARY_NAME)/main.go
	
	# Windows ARM64
	GOOS=windows GOARCH=arm64 CGO_ENABLED=0 go build $(LDFLAGS) \
		-o $(DIST_DIR)/$(BINARY_NAME)-windows-arm64.exe cmd/$(BINARY_NAME)/main.go
	
	@echo "Cross-compilation complete. Binaries in $(DIST_DIR)/"

# Install binary
install: build ## Install binary to GOPATH/bin
	@echo "Installing $(BINARY_NAME)..."
	go install $(LDFLAGS) cmd/$(BINARY_NAME)/main.go

# Create release archives
release: build-all ## Create release archives
	@echo "Creating release archives..."
	@mkdir -p $(DIST_DIR)/archives
	
	# Create archives for each platform
	cd $(DIST_DIR) && tar -czf archives/$(BINARY_NAME)-$(VERSION)-linux-amd64.tar.gz $(BINARY_NAME)-linux-amd64
	cd $(DIST_DIR) && tar -czf archives/$(BINARY_NAME)-$(VERSION)-linux-arm64.tar.gz $(BINARY_NAME)-linux-arm64
	cd $(DIST_DIR) && tar -czf archives/$(BINARY_NAME)-$(VERSION)-darwin-amd64.tar.gz $(BINARY_NAME)-darwin-amd64
	cd $(DIST_DIR) && tar -czf archives/$(BINARY_NAME)-$(VERSION)-darwin-arm64.tar.gz $(BINARY_NAME)-darwin-arm64
	cd $(DIST_DIR) && zip archives/$(BINARY_NAME)-$(VERSION)-windows-amd64.zip $(BINARY_NAME)-windows-amd64.exe
	cd $(DIST_DIR) && zip archives/$(BINARY_NAME)-$(VERSION)-windows-arm64.zip $(BINARY_NAME)-windows-arm64.exe
	
	@echo "Release archives created in $(DIST_DIR)/archives/"

# Clean build artifacts
clean: ## Clean build artifacts
	@echo "Cleaning..."
	rm -rf $(BUILD_DIR) $(DIST_DIR)
	go clean

# Run development server with file watching (requires air)
dev: ## Run development server with hot reload
	@echo "Starting development server..."
	air -c .air.toml || go run cmd/$(BINARY_NAME)/main.go --help

# Generate example config
example-config: ## Generate example configuration file
	@echo "Generating example config..."
	@cat > example/config-production.json << 'EOF'
{
  "smtp": {
    "host": "smtp.gmail.com",
    "port": 587,
    "username": "your-email@gmail.com",
    "password": "your-app-password",
    "from": "your-email@gmail.com",
    "use_tls": true,
    "insecure_skip_verify": false,
    "connection_timeout": "10s",
    "read_timeout": "30s",
    "write_timeout": "30s"
  },
  "rate_limit": 10,
  "burst_limit": 20,
  "log": {
    "level": "info",
    "format": "json",
    "file": "logs/mailgrid.log",
    "max_size": 100,
    "max_backups": 3,
    "max_age": 28
  },
  "metrics": {
    "enabled": true,
    "port": 8090
  },
  "max_attachment_size": 10485760,
  "max_concurrency": 10,
  "max_batch_size": 50,
  "max_retries": 3
}
EOF
	@echo "Example production config created: example/config-production.json"

# Docker targets
docker-build: ## Build Docker image
	@echo "Building Docker image..."
	docker build -t $(BINARY_NAME):$(VERSION) .

docker-run: ## Run Docker container
	@echo "Running Docker container..."
	docker run --rm -it $(BINARY_NAME):$(VERSION)

# Performance benchmark
benchmark: ## Run performance benchmarks
	@echo "Running benchmarks..."
	go test -bench=. -benchmem ./...

# Code coverage
coverage: ## Generate code coverage report
	@echo "Generating coverage report..."
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Check for outdated dependencies
deps-check: ## Check for outdated dependencies
	@echo "Checking for outdated dependencies..."
	go list -u -m all

# Update dependencies
deps-update: ## Update all dependencies
	@echo "Updating dependencies..."
	go get -u ./...
	go mod tidy
