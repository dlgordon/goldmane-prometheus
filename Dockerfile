# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy go mod files and download dependencies first (cache layer)
COPY go.mod go.sum ./
RUN go mod download

# Copy source code (proto .pb.go files must be pre-generated)
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o goldmane-prometheus ./cmd/goldmane-prometheus

# Runtime stage
FROM alpine:3.21

RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=builder /app/goldmane-prometheus .

EXPOSE 9090

ENTRYPOINT ["./goldmane-prometheus"]
