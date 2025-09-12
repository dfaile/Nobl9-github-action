# Multi-stage build for Nobl9 GitHub Action
# Stage 1: Build the Go application
FROM golang:1.24.4-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application with optimizations
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s -extldflags '-static'" \
    -o nobl9-action \
    ./cmd/main.go

# Stage 2: Create minimal runtime image
FROM alpine:3.19

# Install runtime dependencies
RUN apk add --no-cache \
    ca-certificates \
    tzdata \
    && rm -rf /var/cache/apk/*

# Create non-root user for security
RUN addgroup -g 1001 -S nobl9 && \
    adduser -u 1001 -S nobl9 -G nobl9

# Set working directory
WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/nobl9-action /app/nobl9-action

# Copy entrypoint script and any additional runtime files
COPY --from=builder /app/entrypoint.sh /app/entrypoint.sh
COPY --from=builder /app/action.yml /app/action.yml

# Make entrypoint executable and change ownership to non-root user
RUN chmod +x /app/entrypoint.sh && chown -R nobl9:nobl9 /app

# Switch to non-root user
USER nobl9

# Set entrypoint
ENTRYPOINT ["/app/entrypoint.sh"]

# Default command (will be overridden by action.yml args)
CMD ["--help"] 