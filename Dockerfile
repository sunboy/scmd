# Dockerfile for scmd
# Multi-stage build for minimal production image

# Stage 1: Build stage
FROM golang:1.24-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git make ca-certificates tzdata

WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o scmd ./cmd/scmd

# Stage 2: Runtime stage
FROM alpine:latest

# Install runtime dependencies
RUN apk add --no-cache ca-certificates bash curl wget

# Create non-root user
RUN addgroup -S scmd && adduser -S scmd -G scmd

# Create necessary directories
RUN mkdir -p /home/scmd/.scmd/models \
    && chown -R scmd:scmd /home/scmd

WORKDIR /home/scmd

# Copy binary from builder
COPY --from=builder /build/scmd /usr/local/bin/scmd
RUN chmod +x /usr/local/bin/scmd

# Switch to non-root user
USER scmd

# Set environment variables
ENV SCMD_DATA_DIR=/home/scmd/.scmd
ENV PATH="/usr/local/bin:${PATH}"

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD scmd --version || exit 1

# Default command
ENTRYPOINT ["scmd"]
CMD ["--help"]
