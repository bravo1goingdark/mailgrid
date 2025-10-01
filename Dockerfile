# Multi-stage build for minimal Docker image size
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy go mod files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build optimized binary with size reduction flags
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w -X main.version=docker -X main.buildTime=$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
    -trimpath \
    -o mailgrid \
    ./cmd/mailgrid

# Strip binary for additional size reduction
RUN strip mailgrid || true

# Final minimal image - use distroless for security
FROM gcr.io/distroless/static-debian12:nonroot

# Copy CA certificates and timezone data
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# Copy the binary
COPY --from=builder /app/mailgrid /usr/local/bin/mailgrid

# Create directories for data
USER 65532:65532

# Expose metrics port
EXPOSE 8090

# Set environment variables
ENV TZ=UTC
ENV MAILGRID_CONFIG=/app/config.json
ENV MAILGRID_DB=/app/data/mailgrid.db

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD ["/usr/local/bin/mailgrid", "--help"] || exit 1

# Default command runs the scheduler
CMD ["/usr/local/bin/mailgrid", "--scheduler-run", "--env", "/app/config.json"]

# Labels for better maintainability
LABEL org.opencontainers.image.title="MailGrid"
LABEL org.opencontainers.image.description="High-performance email orchestrator with advanced scheduling"
LABEL org.opencontainers.image.vendor="MailGrid Team"
LABEL org.opencontainers.image.licenses="BSD-3-Clause"
LABEL org.opencontainers.image.source="https://github.com/bravo1goingdark/mailgrid"
LABEL org.opencontainers.image.documentation="https://github.com/bravo1goingdark/mailgrid/blob/main/README.md"