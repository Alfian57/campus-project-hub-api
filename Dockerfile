# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install dependencies
RUN apk add --no-cache git

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the server application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags="-w -s" -o server ./cmd/server

# Build the migrate application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags="-w -s" -o migrate ./cmd/migrate

# Build the seeder application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags="-w -s" -o seeder ./cmd/seeder

# Production stage
FROM alpine:3.19

WORKDIR /app

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates tzdata

# Create non-root user
RUN adduser -D -g '' appuser

# Copy binaries from builder
COPY --from=builder /app/server .
COPY --from=builder /app/migrate .
COPY --from=builder /app/seeder .
COPY --from=builder /app/configs ./configs

# Copy migrations for migrate tool
COPY --from=builder /app/migrations ./migrations

# Copy seeder assets
COPY --from=builder /app/cmd/seeder/assets ./cmd/seeder/assets

# Create uploads directory
RUN mkdir -p uploads && chown -R appuser:appuser /app

# Copy entrypoint script
COPY --from=builder --chown=appuser:appuser /app/entrypoint.sh ./entrypoint.sh
RUN chmod +x ./entrypoint.sh

USER appuser

EXPOSE 8000

# Health check
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8000/health || exit 1

# Run with entrypoint
ENTRYPOINT ["./entrypoint.sh"]
CMD ["./server"]
