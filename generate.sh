#!/bin/bash
# Generate Go code from proto files
# This script requires protoc to be installed

set -e

echo "Generating Go code from proto files..."

# Ensure output directory exists
mkdir -p proto

# Generate Go code
protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    proto/api.proto

echo "Proto files generated successfully"
