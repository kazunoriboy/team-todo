#!/bin/sh
set -e

echo "=== Starting entrypoint ==="
echo "Current directory: $(pwd)"
echo "Go version: $(go version)"
echo ""

echo "=== Running go mod tidy ==="
go mod tidy -v
echo ""

echo "=== go.sum contents ==="
cat go.sum | head -20
echo ""

echo "=== Starting air ==="
exec air -c .air.toml

