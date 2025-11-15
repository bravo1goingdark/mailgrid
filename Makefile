# MailGrid Build and Release Makefile
# Optimized for minimal binary sizes and cross-platform compatibility

# Build variables
BINARY_NAME=mailgrid
VERSION ?= $(shell git describe --tags --exact-match 2>/dev/null || git rev-parse --short HEAD)
BUILD_TIME=$(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS=-s -w -X main.version=$(VERSION) -X main.buildTime=$(BUILD_TIME)

# Optimization flags for minimal binary size
BUILD_FLAGS=-ldflags="$(LDFLAGS)" -trimpath
GO_ENV=CGO_ENABLED=0

# Output directories
DIST_DIR=dist
RELEASES_DIR=releases

# Platform targets
PLATFORMS=\
	windows/amd64 \
	windows/arm64 \
	linux/amd64 \
	linux/arm64 \
	linux/386 \
	darwin/amd64 \
	darwin/arm64 \
	freebsd/amd64

# Default target
.PHONY: all
all: clean test build

# Clean build artifacts
.PHONY: clean
clean:
	@echo "🧹 Cleaning build artifacts..."
	@rm -rf $(DIST_DIR) $(RELEASES_DIR) $(BINARY_NAME) $(BINARY_NAME).exe
	@go clean -cache
	@echo "✅ Cleaned!"

# Run tests
.PHONY: test
test:
	@echo "🧪 Running tests..."
	@go test -v ./...
	@echo "✅ Tests passed!"

# Run tests with coverage
.PHONY: test-coverage
test-coverage:
	@echo "📊 Running tests with coverage..."
	@go test -v -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "✅ Coverage report generated: coverage.html"

# Build for current platform
.PHONY: build
build:
	@echo "🔨 Building $(BINARY_NAME) for current platform..."
	@$(GO_ENV) go build $(BUILD_FLAGS) -o $(BINARY_NAME) ./cmd/mailgrid
	@echo "✅ Built $(BINARY_NAME)"
	@ls -lah $(BINARY_NAME) 2>/dev/null || dir $(BINARY_NAME)* 2>nul || true

# Build optimized release binaries for all platforms
.PHONY: build-all
build-all: clean
	@echo "🏗️  Building optimized binaries for all platforms..."
	@mkdir -p $(RELEASES_DIR)
	@$(foreach platform,$(PLATFORMS), \
		echo "Building for $(platform)..."; \
		GOOS=$(word 1,$(subst /, ,$(platform))) \
		GOARCH=$(word 2,$(subst /, ,$(platform))) \
		$(GO_ENV) go build $(BUILD_FLAGS) \
			-o $(RELEASES_DIR)/$(BINARY_NAME)-$(subst /,-,$(platform))$(if $(findstring windows,$(platform)),.exe,) \
			./cmd/mailgrid; \
	)
	@echo "✅ All binaries built!"
	@ls -lah $(RELEASES_DIR)/ 2>/dev/null || dir $(RELEASES_DIR)\* 2>nul || true

# Create compressed release packages
.PHONY: release-packages
release-packages: build-all
	@echo "📦 Creating release packages..."
	@cd $(RELEASES_DIR) && \
	for file in mailgrid-windows-*.exe; do \
		if [ -f "$$file" ]; then \
			echo "Creating $$file.zip..."; \
			zip -9 "$$file.zip" "$$file" ../README.md ../LICENSE ../RELEASE_NOTES_v1.0.0.md; \
			rm "$$file"; \
		fi \
	done
	@cd $(RELEASES_DIR) && \
	for file in mailgrid-*; do \
		if [ -f "$$file" ] && [[ $$file != *.zip ]]; then \
			echo "Creating $$file.tar.gz..."; \
			tar -czf "$$file.tar.gz" "$$file" ../README.md ../LICENSE ../RELEASE_NOTES_v1.0.0.md; \
			rm "$$file"; \
		fi \
	done
	@echo "✅ Release packages created!"

# Generate checksums
.PHONY: checksums
checksums: release-packages
	@echo "🔐 Generating checksums..."
	@cd $(RELEASES_DIR) && \
	if command -v sha256sum >/dev/null 2>&1; then \
		sha256sum * > checksums.txt; \
	else \
		shasum -a 256 * > checksums.txt; \
	fi
	@echo "✅ Checksums generated!"

# Full release build
.PHONY: release
release: test build-all release-packages checksums
	@echo "🎉 Release build complete!"
	@echo "📋 Release assets:"
	@ls -lah $(RELEASES_DIR)/ 2>/dev/null || dir $(RELEASES_DIR)\* 2>nul || true

# Run locally
.PHONY: run
run: build
	@echo "🚀 Running MailGrid..."
	@./$(BINARY_NAME) --help

# Install to GOPATH/bin
.PHONY: install
install:
	@echo "📥 Installing MailGrid..."
	@$(GO_ENV) go install $(BUILD_FLAGS) ./cmd/mailgrid
	@echo "✅ MailGrid installed to GOPATH/bin"

# Development server with live reload
.PHONY: dev
dev:
	@echo "🔄 Starting development server..."
	@go run ./cmd/mailgrid --help

# Lint code
.PHONY: lint
lint:
	@echo "🔍 Linting code..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "⚠️  golangci-lint not found, skipping..."; \
		go vet ./...; \
		go fmt ./...; \
	fi
	@echo "✅ Code linted!"

# Security scan
.PHONY: security
security:
	@echo "🔒 Running security scan..."
	@if command -v gosec >/dev/null 2>&1; then \
		gosec ./...; \
	else \
		echo "⚠️  gosec not found, running basic vulnerability check..."; \
		go list -json -m all | nancy sleuth; \
	fi
	@echo "✅ Security scan complete!"

# Create GitHub release (requires gh CLI)
.PHONY: github-release
github-release: release
	@echo "🎯 Creating GitHub release..."
	@if command -v gh >/dev/null 2>&1; then \
		gh release create $(VERSION) \
			$(RELEASES_DIR)/* \
			--title "MailGrid $(VERSION) - Production-Ready Email Orchestrator" \
			--notes-file RELEASE_NOTES_v1.0.0.md; \
	else \
		echo "❌ GitHub CLI (gh) not found. Please install it or create release manually."; \
		echo "📁 Release assets are ready in $(RELEASES_DIR)/"; \
	fi

# Show build info
.PHONY: info
info:
	@echo "📋 Build Information"
	@echo "===================="
	@echo "Binary Name: $(BINARY_NAME)"
	@echo "Version: $(VERSION)"
	@echo "Build Time: $(BUILD_TIME)"
	@echo "Go Version: $(shell go version)"
	@echo "Platforms: $(PLATFORMS)"
	@echo "LDFLAGS: $(LDFLAGS)"

# Help target
.PHONY: help
help:
	@echo "MailGrid Build and Release Makefile"
	@echo "==================================="
	@echo
	@echo "Available targets:"
	@echo "  all              - Clean, test, and build for current platform"
	@echo "  build            - Build binary for current platform"
	@echo "  build-all        - Build optimized binaries for all platforms"
	@echo "  release          - Create full release with packages and checksums"
	@echo "  release-packages - Create compressed release packages"
	@echo "  checksums        - Generate SHA256 checksums"
	@echo "  test             - Run unit tests"
	@echo "  test-coverage    - Run tests with coverage report"
	@echo "  lint             - Lint code with golangci-lint"
	@echo "  security         - Run security scan"
	@echo "  clean            - Clean build artifacts"
	@echo "  install          - Install to GOPATH/bin"
	@echo "  run              - Build and run locally"
	@echo "  dev              - Development server"
	@echo "  github-release   - Create GitHub release (requires gh CLI)"
	@echo "  info             - Show build information"
	@echo "  help             - Show this help message"
	@echo
	@echo "Environment variables:"
	@echo "  VERSION          - Override version (default: git tag or commit hash)"
	@echo
	@echo "Examples:"
	@echo "  make build                    # Build for current platform"
	@echo "  make release VERSION=v1.0.1  # Create v1.0.1 release"
