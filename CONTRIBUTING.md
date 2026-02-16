# Contributing to Mailgrid
 
Thank you for your interest in contributing to Mailgrid! This guide will help you understand the project structure, architecture, and development workflow.
 
## Table of Contents
 
- [Project Overview](#project-overview)
- [Architecture](#architecture)
- [Package Structure](#package-structure)
- [Data Flow](#data-flow)
- [Development Setup](#development-setup)
- [Contributing Guidelines](#contributing-guidelines)
- [Testing](#testing)
- [Code Standards](#code-standards)
 
## Project Overview
 
Mailgrid is a high-performance CLI tool for bulk email sending with advanced features like scheduling, monitoring, and resumable delivery. Built in Go for speed and reliability.
 
### Key Features
- Bulk email sending from CSV/Google Sheets
- Advanced scheduling with cron support
- Real-time monitoring dashboard
- Resumable delivery with offset tracking
- Template rendering with dynamic content
- SMTP connection pooling and batching
 
## Architecture
 
Mailgrid follows a modular architecture with clear separation of concerns:
 
```
CLI Layer       Application     Data Layer
(cmd, cli)      Logic Layer      (database)
                 (email, sched)
                 
Parsing &        Email Engine    Monitoring
Validation       & Scheduler    & Metrics
(parser)         (email, sched)   (monitor)
```
 
## Package Structure
 
### Core Packages
 
#### `cmd/mailgrid/`
**Purpose**: Application entry point
- `main.go`: CLI initialization and command routing
 
#### `cli/`
**Purpose**: Command-line interface and orchestration
- `cliargs.go`: Flag definitions and parsing using `spf13/pflag`
- `runner.go`: Main application flow orchestration
- `tasks.go`: Email task preparation and validation
 
**Key Functions**:
- `ParseFlags()`: Parses CLI arguments into `CLIArgs` struct
- `Run()`: Main orchestration function handling full lifecycle
- `PrepareEmailTasks()`: Converts recipients into email tasks
- `SendSingleEmail()`: Handles single email sending logic
 
#### `config/`
**Purpose**: Configuration management
- `config.go`: SMTP configuration loading and validation
 
**Key Types**:
- `SMTPConfig`: SMTP server settings
- `Config`: Root configuration structure
 
#### `parser/`
**Purpose**: Data parsing and filtering
- `csv.go`: CSV file parsing
- `sheet.go`: Google Sheets integration
- `types.go`: Data structures for recipients
- `filter.go`: Logical filtering of recipients
- `expr.go`: Expression parsing for filters
- `sheet_utils.go`: Google Sheets URL parsing
 
**Key Types**:
- `Recipient`: Represents an email recipient with template data
- `Expression`: Logical expression for filtering
 
#### `email/`
**Purpose**: Email sending engine
- `dispatcher.go`: Multi-threaded email distribution
- `worker.go`: Individual worker threads for sending
- `sender.go`: Core email sending logic
- `smtp.go`: SMTP connection management
- `template.go`: Template caching and rendering
- `batch.go`: Email batching optimization
- `pool.go`: Connection pooling
- `resilience.go`: Retry logic and circuit breaking
- `pipeline.go`: Email processing pipeline
 
**Key Components**:
- `Task`: Represents a single email to be sent
- `worker`: Goroutine that processes email tasks
- `StartDispatcher()`: Initiates multi-threaded email sending
 
#### `scheduler/`
**Purpose**: Job scheduling and management
- `scheduler.go`: Basic job scheduler
- `email_scheduler.go`: Email-specific scheduling logic
- `manager.go`: Scheduler lifecycle management
- `optimized_scheduler.go`: High-performance scheduler with auto-start
 
**Key Features**:
- Persistent job storage with BoltDB
- Auto-start/auto-shutdown capabilities
- Cron expression support
- Distributed locking for multi-instance safety
 
#### `monitor/`
**Purpose**: Real-time monitoring and metrics
- `server.go`: SSE-based monitoring server
- `dashboard.go`: HTML dashboard generation
- `interface.go`: Monitoring interface definitions
 
**Key Capabilities**:
- Real-time recipient status tracking
- Campaign performance metrics
- Server-Sent Events updates for live monitoring
 
#### `offset/`
**Purpose**: Resumable delivery tracking
- `tracker.go`: Offset tracking for interrupted campaigns
 
**Key Features**:
- Atomic offset updates
- Campaign resume functionality
- File-based persistence
 
### Supporting Packages
 
#### `database/`
**Purpose**: Persistent storage
- `boltdb.go`: BoltDB wrapper for job persistence
 
#### `webhook/`
**Purpose**: External notifications
- `webhook.go`: HTTP webhook client for campaign notifications
 
#### `utils/`
**Purpose**: Utility functions
- `strings.go`: String manipulation helpers
- `input.go`: Input validation and parsing
- `address.go`: Address parsing (CC/BCC)
- `validation.go`: Validation utilities
- `preview.go`: Email preview server
- `template.go`: Template rendering utilities
 
#### `internal/types/`
**Purpose**: Shared type definitions
- `types.go`: Common data structures used across packages
 
#### `logger/` & `metrics/`
**Purpose**: Observability
- Structured logging and metrics collection
 
## Data Flow
 
### 1. Initialization Phase
```
CLI Args → Config Loading → Recipient Parsing → Task Preparation
```
 
### 2. Execution Phase
```
Tasks → Dispatcher → Workers → SMTP → Monitoring → Completion
```
 
### 3. Detailed Flow
 
**1. Input Processing**:
   - CLI flags parsed into `CLIArgs` struct
   - SMTP config loaded from JSON file
   - CSV/Google Sheets parsed into `Recipient` structs
 
**2. Task Preparation**:
   - Recipients filtered using logical expressions
   - Templates rendered with recipient data
   - Email tasks created with attachments, CC/BCC
 
**3. Execution**:
   - Dispatcher creates worker goroutines
   - Workers process tasks in batches
   - SMTP connections managed via connection pooling
   - Progress tracked in real-time via monitoring
 
**4. Monitoring & Completion**:
   - Real-time status updates via Server-Sent Events
   - Offset tracking for resumable delivery
   - Webhook notifications on completion
   - Metrics collection and reporting
 
## Development Setup
 
### Prerequisites
- Go 1.18 or later
- Git
 
### Setup Steps
 
**1. Clone the repository**:
   ```bash
   git clone https://github.com/bravo1goingdark/mailgrid.git
   cd mailgrid
   ```
 
**2. Install dependencies**:
   ```bash
   go mod download
   ```
 
**3. Build the project**:
   ```bash
   go build -o mailgrid cmd/mailgrid/main.go
   ```
 
**4. Run tests**:
   ```bash
   go test ./...
   ```
 
**5. Run with test config**:
   ```bash
   ./mailgrid --env config.json --to test@example.com --subject "Test" --text "Hello" --dry-run
   ```
 
## Contributing Guidelines
 
### Before You Start
1. Check existing issues and PRs to avoid duplicates
2. For major changes, open an issue first to discuss the approach
3. Fork the repository and create a feature branch
 
### Making Changes
 
**1. Create a feature branch**:
   ```bash
   git checkout -b feature/your-feature-name
   ```
 
**2. Make your changes**:
   - Follow existing code patterns
   - Add tests for new functionality
   - Update documentation if needed
 
**3. Test your changes**:
   ```bash
   go test ./...
   go vet ./...
   gofmt -s -w .
   ```
 
**4. Commit with clear messages**:
   ```bash
   git commit -m "feat: add new monitoring feature"
   ```
 
### Pull Request Process
 
1. Update documentation for any new features
2. Add tests that cover your changes
3. Ensure all tests pass and code is formatted
4. Submit a PR with a clear description of changes
5. Respond to review feedback promptly
 
### Commit Message Format
```
type(scope): description
 
[optional body]
 
[optional footer]
```
 
Types: `feat`, `fix`, `docs`, `style`, `refactor`, `test`, `chore`
 
## Testing
 
### Test Structure
```
test/
├── cli/           # CLI integration tests
├── config/        # Configuration tests
├── parser/        # Data parsing and validation tests
├── preview/       # Preview server test helpers
└── scheduler/     # Scheduler tests
```
 
### Running Tests
 
```bash
# Run all tests
go test ./...
 
# Run tests with coverage
go test -cover ./...
 
# Run specific package tests
go test ./email/
 
# Run integration tests
go test ./test/cli/
```
 
### Test Categories
 
1. **Unit Tests**: Test individual functions and methods
2. **Integration Tests**: Test component interactions
3. **CLI Tests**: Test complete command-line workflows
 
## Code Standards
 
### Go Best Practices
- Follow effective Go guidelines
- Use `gofmt` and `goimports` for formatting
- Run `go vet` to catch common issues
- Handle errors explicitly
- Use meaningful variable and function names
 
### Project-Specific Guidelines
 
**1. Error Handling**:
   ```go
   if err != nil {
       return fmt.Errorf("operation failed: %w", err)
   }
   ```
 
**2. Logging**:
   ```go
   log.Printf("Sent to %s (attempt %d)", recipient.Email, attempt)
   ```
 
**3. Configuration**:
   - Use struct tags for JSON parsing
   - Validate configuration on load
   - Provide sensible defaults
 
**4. Concurrency**:
   - Use channels for communication
   - Properly handle goroutine lifecycle
   - Avoid shared mutable state
 
**5. Documentation**:
   - Add godoc comments for exported functions
   - Include usage examples in comments
   - Keep README.md updated
 
### Code Review Checklist
 
- [ ] Code follows Go conventions
- [ ] All tests pass
- [ ] Error handling is comprehensive
- [ ] Performance implications considered
- [ ] Documentation updated
- [ ] Backward compatibility maintained
- [ ] Security considerations addressed
 
## Reporting Issues
 
When reporting bugs, please include:
 
1. **Environment details**: OS, Go version, Mailgrid version
2. **Steps to reproduce**: Clear, minimal reproduction case
3. **Expected vs actual behavior**
4. **Relevant logs or error messages**
5. **Configuration details** (sanitized)
 
## Feature Requests
 
For new features:
 
1. **Check existing issues** to avoid duplicates
2. **Describe the use case** and motivation
3. **Propose implementation approach** if you have ideas
4. **Consider backward compatibility**
 
## Resources
 
- [Go Documentation](https://golang.org/doc/)
- [Effective Go](https://golang.org/doc/effective_go)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- [Project Issues](https://github.com/bravo1goingdark/mailgrid/issues)
 
## Getting Help
 
- **GitHub Issues**: For bugs and feature requests
- **Discussions**: For questions and general discussion
- **Documentation**: [Full docs](./docs/docs.md)
 
---
 
Thank you for contributing to Mailgrid! Your efforts help make bulk email automation better for everyone.
