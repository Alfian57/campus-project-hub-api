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
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o server ./cmd/server

# Build the seeder application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o seeder ./cmd/seeder

# Production stage
FROM alpine:3.19

WORKDIR /app

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates tzdata

# Create non-root user
RUN adduser -D -g '' appuser

# Copy binaries from builder
COPY --from=builder /app/server .
COPY --from=builder /app/seeder .
COPY --from=builder /app/configs ./configs

# Copy seeder assets
COPY --from=builder /app/cmd/seeder/assets ./cmd/seeder/assets

# Create uploads directory
RUN mkdir -p uploads && chown -R appuser:appuser /app

USER appuser

EXPOSE 8000

CMD ["./server"]
