.PHONY: help build run test clean docker-build docker-up docker-down migrate

# Default target
help:
	@echo "ISP Visual Monitor - Makefile Commands"
	@echo ""
	@echo "  make build          - Build the application binary"
	@echo "  make run            - Run the application locally"
	@echo "  make test           - Run tests"
	@echo "  make clean          - Clean build artifacts"
	@echo "  make docker-build   - Build Docker image"
	@echo "  make docker-up      - Start all services with Docker Compose"
	@echo "  make docker-down    - Stop all Docker services"
	@echo "  make migrate        - Run database migrations"
	@echo "  make lint           - Run linters (requires golangci-lint)"
	@echo ""

# Build the application
build:
	@echo "Building ISP Monitor..."
	@mkdir -p bin
	go build -o bin/ispmonitor cmd/ispmonitor/main.go
	@echo "Build complete: bin/ispmonitor"

# Run the application locally
run:
	@echo "Running ISP Monitor..."
	go run cmd/ispmonitor/main.go

# Run tests
test:
	@echo "Running tests..."
	go test -v ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test -cover -coverprofile=coverage.out ./...
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
	docker build -t ispmonitor:latest .
	@echo "Docker image built: ispmonitor:latest"

# Start Docker Compose services
docker-up:
	@echo "Starting Docker services..."
	docker-compose up -d
	@echo "Services started. Check status with: docker-compose ps"

# Stop Docker Compose services
docker-down:
	@echo "Stopping Docker services..."
	docker-compose down
	@echo "Services stopped"

# Run database migrations
migrate:
	@echo "Running database migrations..."
	docker-compose exec -T postgres psql -U ispmonitor -d ispmonitor < db/migrations/001_initial_schema.sql
	@echo "Migrations complete"

# Run linters (requires golangci-lint)
lint:
	@echo "Running linters..."
	golangci-lint run ./...

# Install development dependencies
dev-deps:
	@echo "Installing development dependencies..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "Dependencies installed"

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
logs-app:
	docker-compose logs -f app

logs-postgres:
	docker-compose logs -f postgres

logs-redis:
	docker-compose logs -f redis

# Open a shell in the app container
shell:
	docker-compose exec app sh

# Database shell
db-shell:
	docker-compose exec postgres psql -U ispmonitor -d ispmonitor
