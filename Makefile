# Go Core Git - Makefile
.PHONY: build test fmt lint clean cross demo help

# Build configuration
BINARY_NAME=gitmgr
BUILD_DIR=bin
VERSION?=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS=-ldflags "-X main.version=$(VERSION)"

# Default target
all: build

# Build for current platform
build:
	@echo "Building $(BINARY_NAME) v$(VERSION)..."
	@mkdir -p $(BUILD_DIR)
	go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/gitmgr

# Run tests
test:
	go test -v ./...

# Format code
fmt:
	go fmt ./...

# Lint code
lint:
	go vet ./...

# Clean build artifacts
clean:
	rm -rf $(BUILD_DIR)

# Cross-compile for multiple platforms
cross:
	@echo "Cross-compiling..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 ./cmd/gitmgr
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 ./cmd/gitmgr
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 ./cmd/gitmgr
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe ./cmd/gitmgr

# Run demo script
demo:
	@./scripts/demo.sh

# Show help
help:
	@echo "Available targets:"
	@echo "  build    - Build binary for current platform"
	@echo "  test     - Run tests"
	@echo "  fmt      - Format code"
	@echo "  lint     - Lint code"
	@echo "  clean    - Clean build artifacts"
	@echo "  cross    - Cross-compile for multiple platforms"
	@echo "  demo     - Run demo script"
	@echo "  help     - Show this help"