#!/bin/bash
# Lint script for SQLBlade

set -e

echo "🔍 Running golangci-lint..."

# Check if golangci-lint is installed
if ! command -v golangci-lint &> /dev/null; then
    echo "❌ golangci-lint is not installed"
    echo "📦 Installing golangci-lint..."
    curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin latest
    export PATH=$PATH:$(go env GOPATH)/bin
fi

# Run lint
golangci-lint run --timeout=5m

echo "✅ Lint completed successfully!"

