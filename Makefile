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
TEST_QUICK_TIMEOUT=15s

.PHONY: all build clean test coverage deps vendor help test-quick

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

# Run quick tests (very fast, under 15 seconds)
test-quick:
	@echo "Running quick tests (under 15 seconds)..."
	$(GOTEST) -timeout $(TEST_QUICK_TIMEOUT) -short ./...
	@echo "Running placeholder verification script..."
	chmod +x ./pkg/placeholder/example/verify_placeholder.sh
	./pkg/placeholder/example/verify_placeholder.sh
	@echo "Running C placeholder verification script..."
	chmod +x ./pkg/placeholder/example/verify_placeholder_c.sh
	./pkg/placeholder/example/verify_placeholder_c.sh

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
	@echo "  test-quick   - Run tests in short mode without race detector (under 15 seconds)"
	@echo "  coverage     - Run tests with coverage"
	@echo "  clean        - Clean build artifacts"
	@echo "  deps         - Ensure dependencies are up to date"
	@echo "  vendor       - Vendor dependencies"
	@echo "  all          - Run clean, deps, test, and build"
	@echo "  help         - Show this help message" 