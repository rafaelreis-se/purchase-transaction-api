# Build stage
FROM golang:1.25.1 AS builder

# Install git and build dependencies
RUN apt-get update && apt-get install -y git ca-certificates tzdata && rm -rf /var/lib/apt/lists/*

# Create appuser for security
RUN useradd -r -s /bin/false appuser

# Set working directory
WORKDIR /app

# Copy go mod files first for better caching
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -o server cmd/server/main.go

# Final stage - minimal image
FROM alpine:latest

# Install ca-certificates for HTTPS calls
RUN apk --no-cache add ca-certificates tzdata

# Create appuser for security (Alpine syntax)
RUN adduser -D -g '' appuser

# Copy timezone data
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# Create directory for app
WORKDIR /app

# Copy the binary from builder stage
COPY --from=builder /app/server .

# Create directory for SQLite database and set ownership
RUN mkdir -p /app/data && chown -R appuser:appuser /app/data && chmod -R 755 /app/data

# Use non-root user
USER appuser

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Run the application
CMD ["./server"]