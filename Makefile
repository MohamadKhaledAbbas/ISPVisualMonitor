.PHONY: help build build-all run test clean docker-build docker-up docker-down migrate lint

# Version info
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS := -ldflags "-w -s -X github.com/MohamadKhaledAbbas/ISPVisualMonitor/pkg/version.Version=$(VERSION) -X github.com/MohamadKhaledAbbas/ISPVisualMonitor/pkg/version.GitCommit=$(COMMIT) -X github.com/MohamadKhaledAbbas/ISPVisualMonitor/pkg/version.BuildDate=$(BUILD_DATE)"

# Default target
help:
	@echo "ISP Visual Monitor - Makefile Commands"
	@echo ""
	@echo "  make build          - Build the application binary"
	@echo "  make build-all      - Build for all platforms"
	@echo "  make run            - Run the application locally"
	@echo "  make test           - Run tests"
	@echo "  make test-coverage  - Run tests with coverage report"
	@echo "  make clean          - Clean build artifacts"
	@echo "  make docker-build   - Build Docker image"
	@echo "  make docker-up      - Start all services with Docker Compose"
	@echo "  make docker-down    - Stop all Docker services"
	@echo "  make docker-prod    - Start production services"
	@echo "  make migrate        - Run database migrations"
	@echo "  make lint           - Run linters (requires golangci-lint)"
	@echo "  make setup-dev      - Setup development environment"
	@echo "  make backup         - Create database backup"
	@echo ""

# Build the application
build:
	@echo "Building ISP Monitor ($(VERSION))..."
	@mkdir -p bin
	CGO_ENABLED=0 go build $(LDFLAGS) -o bin/ispmonitor ./cmd/ispmonitor
	@echo "Build complete: bin/ispmonitor"

# Build for all platforms
build-all:
	@echo "Building ISP Monitor for all platforms..."
	@mkdir -p bin
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o bin/ispmonitor-linux-amd64 ./cmd/ispmonitor
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o bin/ispmonitor-linux-arm64 ./cmd/ispmonitor
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o bin/ispmonitor-darwin-amd64 ./cmd/ispmonitor
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o bin/ispmonitor-darwin-arm64 ./cmd/ispmonitor
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o bin/ispmonitor-windows-amd64.exe ./cmd/ispmonitor
	@echo "Build complete for all platforms"

# Run the application locally
run:
	@echo "Running ISP Monitor..."
	go run ./cmd/ispmonitor

# Run tests
test:
	@echo "Running tests..."
	go test -v -race ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test -v -race -coverprofile=coverage.out -covermode=atomic ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

# Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -rf bin/
	rm -f coverage.out coverage.html
	@echo "Clean complete"

# Build Docker image
docker-build:
	@echo "Building Docker image..."
	docker build --build-arg VERSION=$(VERSION) -t ispmonitor:latest -t ispmonitor:$(VERSION) .
	@echo "Docker image built: ispmonitor:$(VERSION)"

# Start Docker Compose services (development)
docker-up:
	@echo "Starting Docker services..."
	docker-compose up -d
	@echo "Services started. Check status with: docker-compose ps"

# Stop Docker Compose services
docker-down:
	@echo "Stopping Docker services..."
	docker-compose down
	@echo "Services stopped"

# Start production services
docker-prod:
	@echo "Starting production services..."
	docker-compose -f docker-compose.prod.yml up -d
	@echo "Production services started"

# Run database migrations
migrate:
	@echo "Running database migrations..."
	docker-compose exec -T postgres psql -U ispmonitor -d ispmonitor < db/migrations/001_initial_schema.sql
	@echo "Migrations complete"

# Run linters (requires golangci-lint)
lint:
	@echo "Running linters..."
	golangci-lint run ./...

# Security scan
security:
	@echo "Running security scan..."
	gosec ./...

# Install development dependencies
dev-deps:
	@echo "Installing development dependencies..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/securego/gosec/v2/cmd/gosec@latest
	@echo "Dependencies installed"

# Setup development environment
setup-dev:
	@echo "Setting up development environment..."
	./deploy/scripts/setup-dev.sh

# Setup production environment
setup-prod:
	@echo "Setting up production environment..."
	./deploy/scripts/setup-prod.sh

# Create database backup
backup:
	@echo "Creating database backup..."
	./deploy/scripts/backup.sh backup

# Restore database from backup
restore:
	@echo "Restoring database..."
	./deploy/scripts/backup.sh restore $(BACKUP_FILE)

# Run go mod tidy
tidy:
	@echo "Tidying Go modules..."
	go mod tidy
	@echo "Go modules tidied"

# Format code
fmt:
	@echo "Formatting code..."
	go fmt ./...
	@echo "Code formatted"

# View logs from Docker services
logs:
	docker-compose logs -f

# View logs from specific service
logs-api:
	docker-compose logs -f api

logs-postgres:
	docker-compose logs -f postgres

logs-redis:
	docker-compose logs -f redis

# Open a shell in the app container
shell:
	docker-compose exec api sh

# Database shell
db-shell:
	docker-compose exec postgres psql -U ispmonitor -d ispmonitor

# Generate release (requires goreleaser)
release:
	@echo "Creating release..."
	goreleaser release --clean

# Dry run release
release-dry:
	@echo "Dry run release..."
	goreleaser release --snapshot --clean
