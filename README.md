# Goldmane Prometheus Exporter

A Prometheus exporter for Calico Goldmane flow logs that connects to the Goldmane gRPC API and exposes flow metrics in Prometheus format.

## Overview

This application runs as a daemon in Kubernetes and:
- Connects to the Calico Goldmane API via gRPC
- Polls for flow data at configurable intervals (default: 15 seconds)
- Exposes Prometheus metrics at `/metrics` endpoint

## Metrics

The exporter provides two counter metrics:

### `calico_flow_allow`
Number of allowed network flows in Calico

### `calico_flow_deny`
Number of denied network flows in Calico

### Labels

Both metrics include the following dimensions:
- `reporter` - Source (src) or destination (dst)
- `protocol` - Network protocol (TCP, UDP, etc.)
- `src_namespace` - Source namespace
- `src_pod` - Source pod name
- `src_port` - Source port (currently fixed to "0" as not directly available in FlowKey)
- `dst_namespace` - Destination namespace
- `dst_object` - Destination object name
- `dst_port` - Destination port

## Configuration

Configuration is done via environment variables:

| Variable | Description | Default |
|----------|-------------|---------|
| `GOLDMANE_ADDR` | Address of the Goldmane gRPC API | `localhost:9094` |
| `METRICS_ADDR` | Address where metrics endpoint will be exposed | `:9090` |
| `POLL_INTERVAL` | Polling interval in seconds | `15` |
| `TLS_ENABLED` | Enable TLS for Goldmane connection | `false` |
| `TLS_CERT_PATH` | Path to TLS certificate file | `""` |
| `TLS_KEY_PATH` | Path to TLS key file | `""` |
| `TLS_CA_PATH` | Path to CA certificate file | `""` |

## Building

### Prerequisites

- Go 1.24 or later
- `protoc` compiler (for generating gRPC code)
- Docker (for containerized builds)

### Generate Proto Files

Before building, you need to generate Go code from the proto files:

```bash
# Make the script executable
chmod +x generate.sh

# Generate proto files (requires protoc to be installed)
./generate.sh

# Or use the Makefile
make proto
```

### Local Build

```bash
# Build the binary
make build

# Or manually
go build -o bin/goldmane-prometheus ./cmd/goldmane-prometheus
```

### Docker Build

```bash
# Build Docker image
make docker-build

# Or manually
docker build -t goldmane-prometheus:latest .
```

## Running

### Local

```bash
# Set environment variables
export GOLDMANE_ADDR=goldmane-api.calico-system.svc:9094
export METRICS_ADDR=:9090
export POLL_INTERVAL=15

# Run the application
./bin/goldmane-prometheus

# Or use go run
go run ./cmd/goldmane-prometheus
```

### Docker

```bash
docker run -d \
  -e GOLDMANE_ADDR=goldmane-api.calico-system.svc:9094 \
  -e METRICS_ADDR=:9090 \
  -e POLL_INTERVAL=15 \
  -p 9090:9090 \
  goldmane-prometheus:latest
```

### Kubernetes

Example deployment:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: goldmane-prometheus
  namespace: calico-system
spec:
  replicas: 1
  selector:
    matchLabels:
      app: goldmane-prometheus
  template:
    metadata:
      labels:
        app: goldmane-prometheus
    spec:
      containers:
      - name: goldmane-prometheus
        image: goldmane-prometheus:latest
        env:
        - name: GOLDMANE_ADDR
          value: "goldmane-api.calico-system.svc:9094"
        - name: METRICS_ADDR
          value: ":9090"
        - name: POLL_INTERVAL
          value: "15"
        ports:
        - containerPort: 9090
          name: metrics
          protocol: TCP
        livenessProbe:
          httpGet:
            path: /health
            port: metrics
          initialDelaySeconds: 10
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /ready
            port: metrics
          initialDelaySeconds: 5
          periodSeconds: 5
---
apiVersion: v1
kind: Service
metadata:
  name: goldmane-prometheus
  namespace: calico-system
  labels:
    app: goldmane-prometheus
spec:
  ports:
  - port: 9090
    targetPort: metrics
    name: metrics
  selector:
    app: goldmane-prometheus
```

## Endpoints

- `/metrics` - Prometheus metrics endpoint
- `/health` - Health check endpoint
- `/ready` - Readiness check endpoint

## Development

### Project Structure

```
.
├── cmd/
│   └── goldmane-prometheus/    # Main application entry point
│       └── main.go
├── internal/
│   ├── collector/               # Flow collection and metrics logic
│   │   ├── collector.go
│   │   └── metrics.go
│   └── config/                  # Configuration management
│       └── config.go
├── proto/                       # Protocol buffer definitions
│   └── api.proto
├── Dockerfile                   # Container image definition
├── Makefile                    # Build automation
├── generate.sh                 # Proto generation script
└── README.md
```

## License

This project is provided as-is for use with Calico Goldmane flow logs.
