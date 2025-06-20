# 🔥 BurnDevice Dockerfile
# Multi-stage build for production-ready container

# Build stage
FROM golang:1.24-alpine AS builder

# Install build dependencies
RUN apk add --no-cache \
    git \
    make \
    curl \
    unzip

# Install buf
ARG BUF_VERSION=1.28.1
RUN curl -sSL "https://github.com/bufbuild/buf/releases/download/v${BUF_VERSION}/buf-Linux-x86_64" -o /usr/local/bin/buf && \
    chmod +x /usr/local/bin/buf

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Generate protobuf code
RUN buf generate

# Build arguments
ARG VERSION=dev
ARG COMMIT=unknown
ARG DATE=unknown

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-s -w -X main.version=${VERSION} -X main.commit=${COMMIT} -X main.date=${DATE}" \
    -o burndevice ./cmd/burndevice

# Runtime stage
FROM alpine:3.19

# Install runtime dependencies
RUN apk add --no-cache \
    ca-certificates \
    tzdata \
    && rm -rf /var/cache/apk/*

# Create non-root user
RUN addgroup -g 1000 burndevice && \
    adduser -D -s /bin/sh -u 1000 -G burndevice burndevice

# Create necessary directories
RUN mkdir -p /app/config /app/data /app/logs && \
    chown -R burndevice:burndevice /app

# Copy binary from builder stage
COPY --from=builder /app/burndevice /usr/local/bin/burndevice

# Copy default config
COPY --chown=burndevice:burndevice config.example.yaml /app/config/config.yaml

# Switch to non-root user
USER burndevice

# Set working directory
WORKDIR /app

# Expose gRPC port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD burndevice client system-info --server localhost:8080 || exit 1

# Default command
CMD ["burndevice", "server", "--config", "/app/config/config.yaml"]

# Labels for metadata
LABEL org.opencontainers.image.title="BurnDevice"
LABEL org.opencontainers.image.description="🔥 Device destructive testing tool for authorized test environments"
LABEL org.opencontainers.image.version="${VERSION}"
LABEL org.opencontainers.image.source="https://github.com/BurnDevice/BurnDevice"
LABEL org.opencontainers.image.licenses="MIT"
LABEL org.opencontainers.image.documentation="https://github.com/BurnDevice/BurnDevice/blob/main/README.md"