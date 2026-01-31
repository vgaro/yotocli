# Build Stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git make

# Copy module files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN go build -o yoto main.go

# Runtime Stage
FROM alpine:latest

WORKDIR /app

# Install runtime dependencies
# ffmpeg: For audio normalization
# python3 & py3-pip: For yt-dlp
# curl: For healthcheck
RUN apk add --no-cache \
    ffmpeg \
    python3 \
    py3-pip \
    curl \
    ca-certificates

# Install yt-dlp (managed by system package if possible, or pip with break-system-packages on newer alpine)
# Alpine's python environment is managed. 'pipx' is often preferred, or --break-system-packages.
# Let's try pip directly first, usually allowed in containers.
RUN pip3 install --no-cache-dir yt-dlp --break-system-packages

# Copy binary from builder
COPY --from=builder /app/yoto /usr/local/bin/yoto

# Create config directory
RUN mkdir -p /root/.config/yotocli

# Expose SSE port
EXPOSE 8080

# Healthcheck
HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \
  CMD curl -f http://localhost:8080/health || exit 1

# Default command: Run MCP in SSE mode
CMD ["yoto", "mcp", "--transport", "sse", "--port", "8080", "--addr", "0.0.0.0"]
