#!/bin/bash
# Script to generate protobuf code

set -e

echo "Generating protobuf code..."

# Ensure we're in the project root
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
cd "$PROJECT_ROOT"

# Add Go bin to PATH
export PATH="$PATH:$(go env GOPATH)/bin"

# Generate Go code
protoc --go_out=. --go_opt=paths=source_relative \
       --go-grpc_out=. --go-grpc_opt=paths=source_relative \
       api/proto/compute.proto

echo "✓ Protobuf generation complete"
echo "Generated files:"
echo "  - api/proto/compute.pb.go"
echo "  - api/proto/compute_grpc.pb.go"
