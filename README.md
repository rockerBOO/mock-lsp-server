# Mock LSP Server

## Overview

A mock Language Server Protocol (LSP) server implemented in Go. This server provides a simulated LSP implementation for testing and development purposes.

## Features

- Implements key LSP methods:
  - Initialize
  - Completion
  - Hover
  - Definition
  - References
  - Document Symbols
- Supports basic document lifecycle events:
  - Open
  - Change
  - Save
  - Close
- Generates mock diagnostics
- Runs via stdio

## Requirements

- Go 1.24+
- Dependencies:
  - github.com/myleshyson/lsprotocol-go/protocol
  - github.com/sourcegraph/jsonrpc2

## Quick Start

### Using Make (Recommended)

```bash
# Download dependencies
make deps

# Run tests
make test

# Build the application
make build

# Run the server
make run
```

### Manual Installation

```bash
# Download dependencies
go mod download

# Run tests
go test -v ./...

# Build manually
go build -o mock-lsp-server .

# Run the server
./mock-lsp-server
```

## Development

### Available Make Targets

#### Core Development

- `make all` - Run clean, test, lint, and build
- `make build` - Build the application
- `make test` - Run all tests
- `make test-coverage` - Run tests with HTML coverage report
- `make test-race` - Run tests with race condition detection
- `make lint` - Lint the code (auto-installs golangci-lint)
- `make fmt` - Format the code
- `make vet` - Vet the code

#### Distribution

- `make build-all` - Build for multiple platforms (Linux, macOS, Windows)
- `make dist` - Create distribution packages (.tar.gz and .zip)
- `make install` - Install binary to GOPATH/bin

#### Utilities

- `make clean` - Clean build artifacts
- `make deps` - Download and tidy dependencies
- `make deps-update` - Update all dependencies
- `make run` - Build and run the application
- `make security` - Check for security vulnerabilities
- `make dev-setup` - Setup development environment
- `make help` - Show all available targets

### Development Setup

```bash
# Setup development environment (installs linting tools)
make dev-setup

# Install dependencies
make deps

# Run tests with coverage
make test-coverage
```

### Cross-Platform Builds

Build for all supported platforms:

```bash
make build-all
```

This creates binaries for:

- Linux (amd64, arm64)
- macOS (amd64, arm64)
- Windows (amd64, arm64)

### Creating Distribution Packages

```bash
make dist
```

Creates compressed packages in `dist/packages/`:

- `.tar.gz` files for Linux and macOS
- `.zip` files for Windows

## Testing

Run the comprehensive test suite:

```bash
# Basic tests
make test

# With coverage report
make test-coverage

# With race detection
make test-race
```

## Code Quality

```bash
# Format code
make fmt

# Lint code
make lint

# Vet code
make vet

# Check for security vulnerabilities
make security
```

## Docker

Build Docker image:

```bash
make docker-build
```

## Purpose

This mock LSP server is designed for testing language server integrations, providing simulated responses for various LSP requests. It's useful for:

- Testing LSP client implementations
- Development and debugging of language server features
- Integration testing of editor/IDE LSP support
- Learning LSP protocol implementation

## License

[Insert license information]
