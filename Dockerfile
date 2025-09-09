# Multi-stage Dockerfile for mailgrid
# Stage 1: Build stage
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata make

# Set working directory
WORKDIR /build

# Copy go mod files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build information
ARG VERSION=dev
ARG COMMIT_HASH=unknown
ARG BUILD_DATE

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags "-w -s -X main.version=${VERSION} -X main.commit=${COMMIT_HASH} -X main.date=${BUILD_DATE}" \
    -o mailgrid cmd/mailgrid/main.go

# Stage 2: Runtime stage
FROM alpine:3.18

# Install runtime dependencies
RUN apk --no-cache add \
    ca-certificates \
    tzdata \
    && addgroup -g 1001 mailgrid \
    && adduser -D -s /bin/sh -u 1001 -G mailgrid mailgrid

# Copy binary from build stage
COPY --from=builder /build/mailgrid /usr/local/bin/mailgrid

# Copy example files (optional)
COPY --from=builder /build/example /app/example

# Create necessary directories
RUN mkdir -p /app/logs /app/data /app/attachments \
    && chown -R mailgrid:mailgrid /app

# Set working directory
WORKDIR /app

# Switch to non-root user
USER mailgrid

# Expose metrics port
EXPOSE 8090

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8090/health || exit 1

# Set default command
ENTRYPOINT ["/usr/local/bin/mailgrid"]
CMD ["--help"]

# Metadata
LABEL maintainer="Mailgrid Team"
LABEL description="High-performance bulk email sender"
LABEL version="${VERSION}"
