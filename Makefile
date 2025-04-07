# Makefile for unisign

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# Binary name
BINARY_NAME=unisign
CMD_DIR=./cmd/unisign

# Build settings
BUILD_DIR=./bin
BUILD_FLAGS=-ldflags="-s -w"
LDFLAGS=-ldflags="-X main.Version=$$(git describe --tags --always --dirty 2>/dev/null || echo 'dev')"

# Testing settings
TEST_FLAGS=-race
TEST_TIMEOUT=600s

.PHONY: all build clean test coverage deps vendor help

all: clean deps test build

# Build the binary
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(BUILD_FLAGS) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(CMD_DIR)
	@echo "Build complete. Binary at $(BUILD_DIR)/$(BINARY_NAME)"

# Run tests
test:
	@echo "Running tests..."
	$(GOTEST) $(TEST_FLAGS) -timeout $(TEST_TIMEOUT) ./...

# Run tests with coverage
coverage:
	@echo "Running tests with coverage..."
	$(GOTEST) $(TEST_FLAGS) -timeout $(TEST_TIMEOUT) -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html

# Run short tests (faster)
test-short:
	@echo "Running short tests..."
	$(GOTEST) $(TEST_FLAGS) -timeout $(TEST_TIMEOUT) -short ./...

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	$(GOCLEAN)
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html

# Install dependencies
deps:
	@echo "Verifying dependencies..."
	$(GOMOD) tidy
	$(GOMOD) verify

# Vendor dependencies
vendor:
	@echo "Vendoring dependencies..."
	$(GOMOD) vendor

# Help information
help:
	@echo "Available targets:"
	@echo "  build        - Build the binary"
	@echo "  test         - Run tests"
	@echo "  test-short   - Run tests with -short flag"
	@echo "  coverage     - Run tests with coverage"
	@echo "  clean        - Clean build artifacts"
	@echo "  deps         - Ensure dependencies are up to date"
	@echo "  vendor       - Vendor dependencies"
	@echo "  all          - Run clean, deps, test, and build"
	@echo "  help         - Show this help message" 