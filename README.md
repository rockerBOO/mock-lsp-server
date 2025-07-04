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

# Run with custom log directory
./mock-lsp-server -log_dir /path/to/logs

# Log logging configuration
./mock-lsp-server -info
```

### Logging Configuration

The server supports flexible logging configuration:

- `-log_dir`: Specify a custom log directory
- `-config`: Use a custom configuration file
- `-info`: Log logging configuration details

Create a `config.json` for advanced logging setup:

```json
{
  "log_dir": "/path/to/custom/logs",
  "log_level": "info",
  "log_file": "custom_logfile.log"
}
```

Logs are created in the following priority:

1. CLI-specified directory
2. Configuration file directory
3. User-specific default directory

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

   Copyright 2025 Dave lage

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
