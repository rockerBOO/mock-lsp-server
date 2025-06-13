# Mock LSP Server Makefile

# Variables
BINARY_NAME=mock-lsp-server
VERSION=1.0.0
BUILD_DIR=build
DIST_DIR=dist
GO_VERSION=1.24

# Default target
.PHONY: all
all: clean test lint build

# Build the application
.PHONY: build
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_NAME) -ldflags "-X main.version=$(VERSION)" .

# Run tests
.PHONY: test
test:
	@echo "Running tests..."
	go test -v ./...

# Run tests with coverage
.PHONY: test-coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run tests with race detection
.PHONY: test-race
test-race:
	@echo "Running tests with race detection..."
	go test -v -race ./...

# Lint the code
.PHONY: lint
lint:
	@echo "Linting code..."
	@which golangci-lint > /dev/null || (echo "golangci-lint not found, installing..." && go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
	golangci-lint run

# Format the code
.PHONY: fmt
fmt:
	@echo "Formatting code..."
	go fmt ./...

# Vet the code
.PHONY: vet
vet:
	@echo "Vetting code..."
	go vet ./...

# Build for multiple platforms
.PHONY: build-all
build-all: clean
	@echo "Building for multiple platforms..."
	@mkdir -p $(DIST_DIR)
	
	# Linux amd64
	GOOS=linux GOARCH=amd64 go build -o $(DIST_DIR)/$(BINARY_NAME)-linux-amd64 -ldflags "-X main.version=$(VERSION)" .
	
	# Linux arm64
	GOOS=linux GOARCH=arm64 go build -o $(DIST_DIR)/$(BINARY_NAME)-linux-arm64 -ldflags "-X main.version=$(VERSION)" .
	
	# macOS amd64
	GOOS=darwin GOARCH=amd64 go build -o $(DIST_DIR)/$(BINARY_NAME)-darwin-amd64 -ldflags "-X main.version=$(VERSION)" .
	
	# macOS arm64
	GOOS=darwin GOARCH=arm64 go build -o $(DIST_DIR)/$(BINARY_NAME)-darwin-arm64 -ldflags "-X main.version=$(VERSION)" .
	
	# Windows amd64
	GOOS=windows GOARCH=amd64 go build -o $(DIST_DIR)/$(BINARY_NAME)-windows-amd64.exe -ldflags "-X main.version=$(VERSION)" .
	
	# Windows arm64
	GOOS=windows GOARCH=arm64 go build -o $(DIST_DIR)/$(BINARY_NAME)-windows-arm64.exe -ldflags "-X main.version=$(VERSION)" .

# Create distribution packages
.PHONY: dist
dist: build-all
	@echo "Creating distribution packages..."
	@mkdir -p $(DIST_DIR)/packages
	
	# Create tar.gz for Linux and macOS
	for binary in $(DIST_DIR)/$(BINARY_NAME)-linux-* $(DIST_DIR)/$(BINARY_NAME)-darwin-*; do \
		if [ -f "$$binary" ]; then \
			basename=$$(basename $$binary); \
			tar -czf $(DIST_DIR)/packages/$$basename.tar.gz -C $(DIST_DIR) $$basename; \
		fi \
	done
	
	# Create zip for Windows
	for binary in $(DIST_DIR)/$(BINARY_NAME)-windows-*.exe; do \
		if [ -f "$$binary" ]; then \
			basename=$$(basename $$binary .exe); \
			cd $(DIST_DIR) && zip packages/$$basename.zip $$(basename $$binary); cd ..; \
		fi \
	done

# Install the binary to GOPATH/bin
.PHONY: install
install:
	@echo "Installing $(BINARY_NAME)..."
	go install -ldflags "-X main.version=$(VERSION)" .

# Run the application
.PHONY: run
run: build
	@echo "Running $(BINARY_NAME)..."
	./$(BUILD_DIR)/$(BINARY_NAME)

# Clean build artifacts
.PHONY: clean
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)
	@rm -rf $(DIST_DIR)
	@rm -f coverage.out coverage.html

# Download dependencies
.PHONY: deps
deps:
	@echo "Downloading dependencies..."
	go mod download
	go mod tidy

# Update dependencies
.PHONY: deps-update
deps-update:
	@echo "Updating dependencies..."
	go get -u ./...
	go mod tidy

# Check for security vulnerabilities
.PHONY: security
security:
	@echo "Checking for security vulnerabilities..."
	@which govulncheck > /dev/null || (echo "govulncheck not found, installing..." && go install golang.org/x/vuln/cmd/govulncheck@latest)
	govulncheck ./...

# Development setup
.PHONY: dev-setup
dev-setup:
	@echo "Setting up development environment..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install golang.org/x/vuln/cmd/govulncheck@latest

# Docker build
.PHONY: docker-build
docker-build:
	@echo "Building Docker image..."
	docker build -t $(BINARY_NAME):$(VERSION) .

# Show help
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  all          - Run clean, test, lint, and build"
	@echo "  build        - Build the application"
	@echo "  test         - Run tests"
	@echo "  test-coverage- Run tests with coverage report"
	@echo "  test-race    - Run tests with race detection"
	@echo "  lint         - Lint the code"
	@echo "  fmt          - Format the code"
	@echo "  vet          - Vet the code"
	@echo "  build-all    - Build for multiple platforms"
	@echo "  dist         - Create distribution packages"
	@echo "  install      - Install binary to GOPATH/bin"
	@echo "  run          - Build and run the application"
	@echo "  clean        - Clean build artifacts"
	@echo "  deps         - Download dependencies"
	@echo "  deps-update  - Update dependencies"
	@echo "  security     - Check for security vulnerabilities"
	@echo "  dev-setup    - Setup development environment"
	@echo "  docker-build - Build Docker image"
	@echo "  help         - Show this help message"