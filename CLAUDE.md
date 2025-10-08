# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Mailgrid is a high-performance, ultra-lightweight CLI tool written in Go for sending bulk emails via SMTP from CSV or Google Sheets. It features a production-ready scheduler with auto-start capabilities, connection pooling, and comprehensive email templating.

## Development Commands

### Building and Testing
- `make build` - Build binary for current platform
- `make test` - Run unit tests
- `make test-coverage` - Run tests with coverage report
- `make lint` - Lint code with golangci-lint (or go vet/fmt if unavailable)
- `make clean` - Clean build artifacts

### Release and Distribution
- `make build-all` - Build optimized binaries for all platforms
- `make release` - Create full release with packages and checksums
- `make docker-build` - Build Docker image
- `make install` - Install to GOPATH/bin

### Development Workflow
- `make dev` - Development server (runs go run ./cmd/mailgrid --help)
- `make run` - Build and run locally
- `make security` - Run security scan with gosec

## Architecture

### Core Structure
```
cmd/mailgrid/     - CLI entry point and main function
cli/              - CLI argument parsing and orchestration
config/           - SMTP and application configuration
email/            - Email composition, templating, and SMTP sending
parser/           - CSV/Google Sheets parsing and filtering
scheduler/        - BoltDB-based job scheduling system
database/         - BoltDB persistence layer
monitor/          - Real-time campaign monitoring and progress tracking
utils/            - Utilities including preview server and validation
internal/types/   - Core data structures and types
```

### Key Components

**CLI Architecture**:
- `cmd/mailgrid/main.go` - Entry point that calls `cli.ParseFlags()` and `cli.Run()`
- `cli/runner.go` - Main orchestration logic for the email workflow
- `cli/runner_scheduler.go` - Scheduler-specific runner logic
- `cli/cliargs.go` - CLI flag definitions and parsing

**Email System**:
- Supports both CSV files and public Google Sheets as data sources
- Go text/template engine for dynamic email content
- Connection pooling and batch processing for performance
- Retry logic with exponential backoff

**Scheduler System**:
- BoltDB-backed persistence with distributed locking
- Supports one-time scheduling (RFC3339) and recurring (cron/interval)
- Auto-start/shutdown with configurable idle periods
- Built-in metrics endpoint on port 8090

**Monitoring System**:
- Real-time campaign progress tracking and recipient status monitoring
- Interface-based design supporting multiple monitor implementations
- Campaign metrics including successful/failed deliveries and retry counts
- Integration with email dispatcher for live status updates
- Support for tracking delivery times and error messages

### Key Configuration
- SMTP configuration via JSON files (see example in README)
- Default scheduler database: `mailgrid.db` in current directory
- Preview server default port: 8080
- Metrics server port: 8090 (when scheduler active)

### Dependencies
- `github.com/robfig/cron/v3` - Cron scheduling
- `go.etcd.io/bbolt` - BoltDB for job persistence
- `github.com/sirupsen/logrus` - Logging
- `github.com/spf13/pflag` - CLI flag parsing
- `github.com/stretchr/testify` - Testing framework

### Testing and Validation
- Unit tests use testify framework
- Mock SMTP server via `github.com/mocktools/go-smtp-mock/v2`
- Coverage reports generated to `coverage.html`
- Security scanning with gosec (optional)

### Performance Features
- Concurrent email sending with configurable workers (`--concurrency`)
- SMTP connection pooling and reuse
- Template caching (1-hour default)
- Circuit breaking for SMTP failures
- Adaptive polling in scheduler