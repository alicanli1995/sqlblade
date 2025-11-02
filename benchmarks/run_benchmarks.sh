#!/bin/bash

# Benchmark test setup script
# This script helps you set up and run benchmark tests

set -e

echo "🚀 SQLBlade Benchmark Test Setup"
echo "=================================="
echo ""

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    echo "❌ Docker is not running. Please start Docker and try again."
    exit 1
fi

# Check if docker-compose is available
if ! command -v docker-compose &> /dev/null; then
    echo "❌ docker-compose is not installed. Please install docker-compose."
    exit 1
fi

echo "📦 Starting PostgreSQL container..."
docker-compose up -d

echo "⏳ Waiting for PostgreSQL to be ready..."
sleep 5

# Wait for PostgreSQL to be ready
max_attempts=30
attempt=0
while [ $attempt -lt $max_attempts ]; do
    if docker-compose exec -T postgres pg_isready -U benchmark_user > /dev/null 2>&1; then
        echo "✅ PostgreSQL is ready!"
        break
    fi
    attempt=$((attempt + 1))
    sleep 1
done

if [ $attempt -eq $max_attempts ]; then
    echo "❌ PostgreSQL failed to start. Check logs with: docker-compose logs"
    exit 1
fi

echo ""
echo "🧪 Running benchmark tests..."
echo ""

# Set connection string environment variable
export DB_CONN="postgres://benchmark_user:benchmark_pass@localhost:5433/benchmark_db?sslmode=disable"

# Run benchmarks
go test -bench=. -benchmem -benchtime=3s .

echo ""
echo "✅ Benchmark tests completed!"
echo ""
echo "To stop PostgreSQL: docker-compose down"
echo "To view logs: docker-compose logs postgres"

