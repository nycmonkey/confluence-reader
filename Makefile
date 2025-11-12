.PHONY: build test clean run install help

# Default target
all: build

# Build the application
build:
	@echo "Building confluence-reader..."
	@go build -o confluence-reader

# Run tests
test:
	@echo "Running tests..."
	@go test ./... -v

# Run tests with coverage
coverage:
	@echo "Running tests with coverage..."
	@go test ./... -coverprofile=coverage.out
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -f confluence-reader
	@rm -f coverage.out coverage.html
	@rm -rf confluence-data/

# Run the application
run: build
	@./confluence-reader

# Install dependencies
deps:
	@echo "Downloading dependencies..."
	@go mod download
	@go mod tidy

# Format code
fmt:
	@echo "Formatting code..."
	@go fmt ./...

# Run linter (requires golangci-lint)
lint:
	@echo "Running linter..."
	@golangci-lint run || echo "golangci-lint not installed. Install with: brew install golangci-lint"

# Show help
help:
	@echo "Available targets:"
	@echo "  build     - Build the application"
	@echo "  test      - Run tests"
	@echo "  coverage  - Run tests with coverage report"
	@echo "  clean     - Remove build artifacts"
	@echo "  run       - Build and run the application"
	@echo "  deps      - Download and tidy dependencies"
	@echo "  fmt       - Format code"
	@echo "  lint      - Run linter"
	@echo "  help      - Show this help message"
