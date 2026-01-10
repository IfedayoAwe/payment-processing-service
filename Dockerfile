# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install git for potential private dependencies
RUN apk add --no-cache git

# Copy go mod files first for better layer caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
# CGO_ENABLED=0 creates a static binary that works in any Linux container
RUN CGO_ENABLED=0 GOOS=linux go build -o payment-service .

# Runtime stage - using alpine for smaller image
FROM alpine:latest

WORKDIR /app

# Install CA certificates for HTTPS/TLS connections
# tzdata for timezone support (useful for logging and date operations)
RUN apk add --no-cache ca-certificates tzdata && \
    update-ca-certificates

# Create non-root user for security
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

# Copy the binary from builder
COPY --from=builder /app/payment-service /usr/local/bin/payment-service

# Copy migrations directory
COPY --from=builder /app/migrations ./migrations

# Change ownership to non-root user
RUN chown -R appuser:appuser /app /usr/local/bin/payment-service

# Switch to non-root user
USER appuser

# Expose the API port
EXPOSE 8080

# Run the service
ENTRYPOINT ["/usr/local/bin/payment-service"]
CMD []
