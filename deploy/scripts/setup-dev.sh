#!/bin/bash
# setup-dev.sh - Development environment setup script
set -e

echo "🚀 Setting up ISP Visual Monitor development environment..."

# Check prerequisites
check_prerequisites() {
    echo "📋 Checking prerequisites..."
    
    if ! command -v docker &> /dev/null; then
        echo "❌ Docker is not installed. Please install Docker first."
        exit 1
    fi
    
    if ! command -v docker-compose &> /dev/null && ! docker compose version &> /dev/null; then
        echo "❌ Docker Compose is not installed. Please install Docker Compose first."
        exit 1
    fi
    
    if ! command -v go &> /dev/null; then
        echo "⚠️  Go is not installed. You won't be able to run locally without Docker."
    else
        GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
        echo "✓ Go version: $GO_VERSION"
    fi
    
    echo "✓ Docker is installed"
    echo "✓ Docker Compose is installed"
}

# Create environment file
setup_env() {
    echo "📝 Setting up environment file..."
    
    if [ -f .env ]; then
        echo "⚠️  .env file already exists. Backing up to .env.backup"
        cp .env .env.backup
    fi
    
    if [ -f configs/config.yaml.example ]; then
        cp configs/config.yaml.example .env
        echo "✓ Created .env from example"
    else
        cat > .env << EOF
DEPLOYMENT_MODE=development
JWT_SECRET=dev-secret-change-in-production-$(openssl rand -hex 16)
DB_HOST=localhost
DB_PORT=5432
DB_USER=ispmonitor
DB_PASSWORD=ispmonitor
DB_NAME=ispmonitor
DB_SSLMODE=disable
REDIS_URL=redis://localhost:6379
API_PORT=8080
LOG_LEVEL=debug
ENABLE_METRICS=true
EOF
        echo "✓ Created default .env file"
    fi
}

# Start services
start_services() {
    echo "🐳 Starting Docker services..."
    
    # Use docker-compose or docker compose depending on what's available
    if command -v docker-compose &> /dev/null; then
        COMPOSE_CMD="docker-compose"
    else
        COMPOSE_CMD="docker compose"
    fi
    
    $COMPOSE_CMD up -d postgres redis
    
    echo "⏳ Waiting for services to be ready..."
    sleep 5
    
    # Wait for PostgreSQL
    echo "   Waiting for PostgreSQL..."
    for i in {1..30}; do
        if $COMPOSE_CMD exec -T postgres pg_isready -U ispmonitor &> /dev/null; then
            echo "   ✓ PostgreSQL is ready"
            break
        fi
        sleep 1
    done
    
    # Wait for Redis
    echo "   Waiting for Redis..."
    for i in {1..30}; do
        if $COMPOSE_CMD exec -T redis redis-cli ping &> /dev/null; then
            echo "   ✓ Redis is ready"
            break
        fi
        sleep 1
    done
}

# Install Go dependencies
install_deps() {
    if command -v go &> /dev/null; then
        echo "📦 Installing Go dependencies..."
        go mod download
        echo "✓ Dependencies installed"
    fi
}

# Run database migrations
run_migrations() {
    echo "🗄️  Running database migrations..."
    
    if command -v docker-compose &> /dev/null; then
        COMPOSE_CMD="docker-compose"
    else
        COMPOSE_CMD="docker compose"
    fi
    
    # Wait for migrations directory
    if [ -d "db/migrations" ]; then
        for migration in db/migrations/*.sql; do
            if [ -f "$migration" ]; then
                echo "   Running: $(basename $migration)"
                $COMPOSE_CMD exec -T postgres psql -U ispmonitor -d ispmonitor -f "/docker-entrypoint-initdb.d/$(basename $migration)" 2>/dev/null || true
            fi
        done
        echo "✓ Migrations complete"
    else
        echo "⚠️  No migrations directory found"
    fi
}

# Print success message
print_success() {
    echo ""
    echo "✅ Development environment is ready!"
    echo ""
    echo "📚 Quick start commands:"
    echo "   make run          - Run the application locally"
    echo "   make test         - Run tests"
    echo "   make docker-up    - Start all services with Docker"
    echo "   make docker-down  - Stop all services"
    echo ""
    echo "🌐 Services:"
    echo "   API:        http://localhost:8080"
    echo "   Prometheus: http://localhost:9090"
    echo "   Grafana:    http://localhost:3000 (admin/admin)"
    echo ""
    echo "📖 Documentation:"
    echo "   docs/DEPLOYMENT.md    - Deployment guide"
    echo "   docs/API.md           - API documentation"
    echo ""
}

# Main
main() {
    check_prerequisites
    setup_env
    start_services
    install_deps
    run_migrations
    print_success
}

main "$@"
