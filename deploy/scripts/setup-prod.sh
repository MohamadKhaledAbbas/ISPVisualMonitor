#!/bin/bash
# setup-prod.sh - Production environment setup script
set -e

echo "🚀 Setting up ISP Visual Monitor production environment..."

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
REQUIRED_VARS=(
    "JWT_SECRET"
    "DATABASE_URL"
    "REDIS_URL"
)

# Check prerequisites
check_prerequisites() {
    echo "📋 Checking prerequisites..."
    
    if ! command -v docker &> /dev/null; then
        echo -e "${RED}❌ Docker is not installed.${NC}"
        exit 1
    fi
    
    if ! command -v docker-compose &> /dev/null && ! docker compose version &> /dev/null; then
        echo -e "${RED}❌ Docker Compose is not installed.${NC}"
        exit 1
    fi
    
    echo -e "${GREEN}✓ Prerequisites met${NC}"
}

# Check required environment variables
check_env() {
    echo "📝 Checking environment configuration..."
    
    if [ ! -f .env ]; then
        echo -e "${RED}❌ .env file not found${NC}"
        echo "Please create a .env file with production settings"
        echo "See configs/config.yaml.example for reference"
        exit 1
    fi
    
    source .env
    
    missing=()
    for var in "${REQUIRED_VARS[@]}"; do
        if [ -z "${!var}" ]; then
            missing+=("$var")
        fi
    done
    
    if [ ${#missing[@]} -ne 0 ]; then
        echo -e "${RED}❌ Missing required environment variables:${NC}"
        for var in "${missing[@]}"; do
            echo "   - $var"
        done
        exit 1
    fi
    
    # Check JWT_SECRET strength
    if [ ${#JWT_SECRET} -lt 32 ]; then
        echo -e "${YELLOW}⚠️  JWT_SECRET should be at least 32 characters${NC}"
    fi
    
    echo -e "${GREEN}✓ Environment configuration valid${NC}"
}

# Pull latest images
pull_images() {
    echo "📥 Pulling latest images..."
    
    if command -v docker-compose &> /dev/null; then
        COMPOSE_CMD="docker-compose"
    else
        COMPOSE_CMD="docker compose"
    fi
    
    $COMPOSE_CMD -f docker-compose.prod.yml pull
    
    echo -e "${GREEN}✓ Images pulled${NC}"
}

# Deploy services
deploy() {
    echo "🚀 Deploying services..."
    
    if command -v docker-compose &> /dev/null; then
        COMPOSE_CMD="docker-compose"
    else
        COMPOSE_CMD="docker compose"
    fi
    
    # Deploy with production compose file
    $COMPOSE_CMD -f docker-compose.prod.yml up -d
    
    echo "⏳ Waiting for services to start..."
    sleep 10
    
    # Health check
    echo "🏥 Running health check..."
    for i in {1..30}; do
        if curl -sf http://localhost:8080/health > /dev/null 2>&1; then
            echo -e "${GREEN}✓ Application is healthy${NC}"
            break
        fi
        if [ $i -eq 30 ]; then
            echo -e "${RED}❌ Health check failed${NC}"
            echo "Check logs with: docker-compose -f docker-compose.prod.yml logs"
            exit 1
        fi
        sleep 2
    done
}

# Setup SSL (optional)
setup_ssl() {
    if [ "$SETUP_SSL" == "true" ]; then
        echo "🔒 Setting up SSL..."
        
        # Create SSL directory
        mkdir -p docker/ssl
        
        # Generate self-signed certificate for testing
        # In production, use Let's Encrypt or your CA
        if [ ! -f docker/ssl/server.crt ]; then
            openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
                -keyout docker/ssl/server.key \
                -out docker/ssl/server.crt \
                -subj "/CN=localhost"
            
            echo -e "${YELLOW}⚠️  Self-signed certificate generated${NC}"
            echo "   For production, replace with certificates from your CA"
        fi
    fi
}

# Print status
print_status() {
    echo ""
    echo "═══════════════════════════════════════════════════════"
    echo -e "${GREEN}✅ Production deployment complete!${NC}"
    echo "═══════════════════════════════════════════════════════"
    echo ""
    echo "🌐 Services:"
    echo "   API:     http://localhost:8080"
    echo "   Metrics: http://localhost:9090/metrics"
    echo ""
    echo "📊 Health check:"
    echo "   curl http://localhost:8080/health"
    echo ""
    echo "📋 View logs:"
    echo "   docker-compose -f docker-compose.prod.yml logs -f"
    echo ""
    echo "🛑 Stop services:"
    echo "   docker-compose -f docker-compose.prod.yml down"
    echo ""
}

# Main
main() {
    check_prerequisites
    check_env
    setup_ssl
    pull_images
    deploy
    print_status
}

main "$@"
