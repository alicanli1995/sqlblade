.PHONY: lint lint-fix test build clean install-tools fmt check quick-check build-examples

# Install golangci-lint
install-tools:
	@echo "📦 Installing golangci-lint..."
	@curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin latest
	@echo "✅ golangci-lint installed"

# Run linters
lint:
	@echo "🔍 Running golangci-lint..."
	@golangci-lint run --timeout=5m

# Run linters and auto-fix
lint-fix:
	@echo "🔧 Running golangci-lint with auto-fix..."
	@golangci-lint run --timeout=5m --fix

# Run tests
test:
	@echo "🧪 Running tests..."
	@go test -v -race -coverprofile=coverage.out -covermode=atomic ./...

# Build the project
build:
	@echo "🔨 Building..."
	@go build ./...

# Build examples
build-examples:
	@echo "🔨 Building examples..."
	@cd examples && go mod tidy && go build . || echo "Examples build skipped"
	@cd examples/mysql && go mod tidy && go build . || echo "MySQL examples build skipped"
	@cd examples/postgres && go mod tidy && go build . || echo "Postgres examples build skipped"

# Clean build artifacts
clean:
	@echo "🧹 Cleaning..."
	@go clean ./...
	@rm -f coverage.out

# Format code
fmt:
	@echo "✨ Formatting code..."
	@go fmt ./...

# Run all checks (lint + test)
check: lint test

# Quick check (lint only)
quick-check: lint

