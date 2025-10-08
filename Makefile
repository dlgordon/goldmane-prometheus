.PHONY: proto build run clean docker-build

# Generate Go code from proto files
proto:
	@echo "Generating Go code from proto files..."
	@mkdir -p proto
	@docker run --rm -v $(PWD):/workspace -w /workspace \
		namely/protoc-all:1.51_1 \
		-f proto/api.proto \
		-l go \
		-o .
	@echo "Proto files generated successfully"

# Build the application
build:
	@echo "Building goldmane-prometheus..."
	@go build -o bin/goldmane-prometheus ./cmd/goldmane-prometheus
	@echo "Build complete: bin/goldmane-prometheus"

# Run the application
run:
	@go run ./cmd/goldmane-prometheus

# Clean build artifacts
clean:
	@rm -rf bin/
	@rm -rf proto/*.pb.go
	@echo "Cleaned build artifacts"

# Build Docker image
docker-build:
	@docker build -t goldmane-prometheus:latest .
	@echo "Docker image built: goldmane-prometheus:latest"
