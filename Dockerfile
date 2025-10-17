# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum* ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o server ./cmd/server

# Final stage
FROM alpine:3.20

# Install CA certificates and curl for healthcheck
RUN apk --no-cache add ca-certificates curl && \
    adduser -D -u 65532 -g appuser appuser

# Set default PORT (can be overridden)
ENV PORT=8080

WORKDIR /app

# Copy the binary from builder
COPY --from=builder /app/server .

# Change ownership to non-root user
RUN chown -R appuser:appuser /app

# Switch to non-root user
USER 65532:65532

# Expose port
EXPOSE 8080

# Health check with PORT env variable support
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD curl -f http://localhost:${PORT}/health || exit 1

# Run the application
CMD ["./server"]

