# Campus Project Hub API - Makefile

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GORUN=$(GOCMD) run
GOTEST=$(GOCMD) test
GOMOD=$(GOCMD) mod
BINARY_NAME=server

# Database
DB_HOST ?= localhost
DB_PORT ?= 5432
DB_USER ?= postgres
DB_PASSWORD ?= postgres
DB_NAME ?= campus_project_hub
DB_SSLMODE ?= disable
DATABASE_URL ?= postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSLMODE)

# Migration
MIGRATE=migrate
MIGRATIONS_PATH=./migrations

.PHONY: all build run test clean deps migrate-up migrate-down migrate-create migrate-force migrate-version help

# Default target
all: build

# Build the application
build:
	@echo "Building..."
	$(GOBUILD) -o $(BINARY_NAME) ./cmd/server

# Run the application
run:
	@echo "Running..."
	$(GORUN) ./cmd/server/main.go

# Run database seeder
seed:
	@echo "Running seeder..."
	$(GORUN) ./cmd/seeder/main.go

# Run tests
test:
	@echo "Testing..."
	$(GOTEST) -v ./...

# Clean build files
clean:
	@echo "Cleaning..."
	rm -f $(BINARY_NAME)
	rm -rf ./tmp

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy

# Install migrate tool
install-migrate:
	@echo "Installing golang-migrate..."
	go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Run all migrations up
migrate-up:
	@echo "Running migrations..."
	$(MIGRATE) -path $(MIGRATIONS_PATH) -database "$(DATABASE_URL)" up

# Rollback last migration
migrate-down:
	@echo "Rolling back last migration..."
	$(MIGRATE) -path $(MIGRATIONS_PATH) -database "$(DATABASE_URL)" down 1

# Rollback all migrations
migrate-down-all:
	@echo "Rolling back all migrations..."
	$(MIGRATE) -path $(MIGRATIONS_PATH) -database "$(DATABASE_URL)" down -all

# Create new migration
# Usage: make migrate-create name=create_users_table
migrate-create:
	@echo "Creating migration: $(name)"
	$(MIGRATE) create -ext sql -dir $(MIGRATIONS_PATH) -seq $(name)

# Force set migration version
# Usage: make migrate-force version=1
migrate-force:
	@echo "Forcing migration version: $(version)"
	$(MIGRATE) -path $(MIGRATIONS_PATH) -database "$(DATABASE_URL)" force $(version)

# Show current migration version
migrate-version:
	@echo "Current migration version:"
	$(MIGRATE) -path $(MIGRATIONS_PATH) -database "$(DATABASE_URL)" version

# Show migration status
migrate-status:
	@echo "Migration status:"
	$(MIGRATE) -path $(MIGRATIONS_PATH) -database "$(DATABASE_URL)" version

# Reset database (drop all and migrate up)
db-reset: migrate-down-all migrate-up
	@echo "Database reset complete"

# Docker commands
docker-build:
	@echo "Building Docker image..."
	docker build -t campus-hub-api .

docker-run:
	@echo "Running Docker container..."
	docker run -p 8000:8000 --env-file .env campus-hub-api

# Development with hot reload (requires air)
dev:
	@echo "Running with hot reload..."
	air

# Install air for hot reload
install-air:
	@echo "Installing air..."
	go install github.com/air-verse/air@latest

# Lint code
lint:
	@echo "Linting..."
	golangci-lint run

# Format code
fmt:
	@echo "Formatting..."
	$(GOCMD) fmt ./...

# Install swag CLI
install-swag:
	@echo "Installing swag..."
	go install github.com/swaggo/swag/cmd/swag@latest

# Generate swagger docs
swagger:
	@echo "Generating swagger documentation..."
	$(shell go env GOPATH)/bin/swag init -g cmd/server/main.go -o docs

# Format swagger comments
swagger-fmt:
	@echo "Formatting swagger comments..."
	$(shell go env GOPATH)/bin/swag fmt

# Help
help:
	@echo "Available commands:"
	@echo "  make build           - Build the application"
	@echo "  make run             - Run the application"
	@echo "  make seed            - Run database seeder"
	@echo "  make test            - Run tests"
	@echo "  make clean           - Clean build files"
	@echo "  make deps            - Download dependencies"
	@echo ""
	@echo "Migration commands:"
	@echo "  make install-migrate - Install golang-migrate tool"
	@echo "  make migrate-up      - Run all migrations"
	@echo "  make migrate-down    - Rollback last migration"
	@echo "  make migrate-down-all- Rollback all migrations"
	@echo "  make migrate-create name=xxx - Create new migration"
	@echo "  make migrate-force version=x - Force set migration version"
	@echo "  make migrate-version - Show current version"
	@echo "  make db-reset        - Reset database"
	@echo ""
	@echo "Swagger commands:"
	@echo "  make install-swag    - Install swag CLI"
	@echo "  make swagger         - Generate swagger docs"
	@echo "  make swagger-fmt     - Format swagger comments"
	@echo ""
	@echo "Docker commands:"
	@echo "  make docker-build    - Build Docker image"
	@echo "  make docker-run      - Run Docker container"
	@echo ""
	@echo "Development commands:"
	@echo "  make dev             - Run with hot reload (requires air)"
	@echo "  make install-air     - Install air for hot reload"
	@echo "  make lint            - Lint code"
	@echo "  make fmt             - Format code"
