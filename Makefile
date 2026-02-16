.PHONY: proto build run clean docker-build docker-run docker-stop

# Application name
APP_NAME := goldmane-prometheus
IMAGE_NAME := $(APP_NAME)
IMAGE_TAG := latest

# Runtime configuration (override with env vars or make args)
GOLDMANE_ADDR ?= goldmane-api.calico-system.svc:9094
METRICS_ADDR ?= :9090
METRICS_PORT ?= 9090
POLL_INTERVAL ?= 15
TLS_ENABLED ?= false
TLS_CERT_PATH ?=
TLS_KEY_PATH ?=
TLS_CA_PATH ?=

# Generate Go code from proto files
proto:
	@echo "Generating Go code from proto files..."
	@mkdir -p proto
	@sh generate.sh
	@echo "Proto files generated successfully"

# Build the application
build:
	@echo "Building $(APP_NAME)..."
	@go build -o bin/$(APP_NAME) ./cmd/$(APP_NAME)
	@echo "Build complete: bin/$(APP_NAME)"

# Run the application locally
run:
	GOLDMANE_ADDR=$(GOLDMANE_ADDR) \
	METRICS_ADDR=$(METRICS_ADDR) \
	POLL_INTERVAL=$(POLL_INTERVAL) \
	TLS_ENABLED=$(TLS_ENABLED) \
	TLS_CERT_PATH=$(TLS_CERT_PATH) \
	TLS_KEY_PATH=$(TLS_KEY_PATH) \
	TLS_CA_PATH=$(TLS_CA_PATH) \
	go run ./cmd/$(APP_NAME)

# Clean build artifacts
clean:
	@rm -rf bin/
	@echo "Cleaned build artifacts"

# Build Docker image
docker-build:
	docker build -t $(IMAGE_NAME):$(IMAGE_TAG) .

# Run Docker container with all config exposed
docker-run: docker-stop
	docker run -d \
		--name $(APP_NAME) \
		-p $(METRICS_PORT):9090 \
		-e GOLDMANE_ADDR=$(GOLDMANE_ADDR) \
		-e METRICS_ADDR=$(METRICS_ADDR) \
		-e POLL_INTERVAL=$(POLL_INTERVAL) \
		-e TLS_ENABLED=$(TLS_ENABLED) \
		$(if $(TLS_CERT_PATH),-e TLS_CERT_PATH=/certs/tls.crt -v $(TLS_CERT_PATH):/certs/tls.crt:ro) \
		$(if $(TLS_KEY_PATH),-e TLS_KEY_PATH=/certs/tls.key -v $(TLS_KEY_PATH):/certs/tls.key:ro) \
		$(if $(TLS_CA_PATH),-e TLS_CA_PATH=/certs/ca.crt -v $(TLS_CA_PATH):/certs/ca.crt:ro) \
		$(IMAGE_NAME):$(IMAGE_TAG)
	@echo "Container $(APP_NAME) started on port $(METRICS_PORT)"
	@echo "  Metrics: http://localhost:$(METRICS_PORT)/metrics"
	@echo "  Health:  http://localhost:$(METRICS_PORT)/health"

# Stop and remove the Docker container
docker-stop:
	@docker rm -f $(APP_NAME) 2>/dev/null || true

# Tail container logs
docker-logs:
	docker logs -f $(APP_NAME)
