#!/bin/bash
# backup.sh - Database backup script
set -e

# Configuration
BACKUP_DIR="${BACKUP_DIR:-/backups/ispmonitor}"
RETENTION_DAYS="${RETENTION_DAYS:-30}"
DATE=$(date +%Y%m%d_%H%M%S)

# Color codes
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

# Usage
usage() {
    echo "Usage: $0 [backup|restore|list] [options]"
    echo ""
    echo "Commands:"
    echo "  backup              Create a new backup"
    echo "  restore <file>      Restore from a backup file"
    echo "  list                List available backups"
    echo ""
    echo "Options:"
    echo "  --dir <path>        Backup directory (default: /backups/ispmonitor)"
    echo "  --retention <days>  Retention period in days (default: 30)"
    echo "  --s3 <bucket>       Upload to S3 bucket"
    echo ""
    echo "Environment variables:"
    echo "  DB_HOST             Database host"
    echo "  DB_USER             Database user"
    echo "  DB_NAME             Database name"
    echo "  PGPASSWORD          Database password"
}

# Parse arguments
parse_args() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            --dir)
                BACKUP_DIR="$2"
                shift 2
                ;;
            --retention)
                RETENTION_DAYS="$2"
                shift 2
                ;;
            --s3)
                S3_BUCKET="$2"
                shift 2
                ;;
            backup|restore|list)
                COMMAND="$1"
                shift
                ;;
            *)
                RESTORE_FILE="$1"
                shift
                ;;
        esac
    done
}

# Create backup
do_backup() {
    echo "📦 Creating backup..."
    
    # Create backup directory
    mkdir -p "$BACKUP_DIR"
    
    BACKUP_FILE="${BACKUP_DIR}/backup_${DATE}.sql.gz"
    
    # Get database connection info
    DB_HOST="${DB_HOST:-localhost}"
    DB_USER="${DB_USER:-ispmonitor}"
    DB_NAME="${DB_NAME:-ispmonitor}"
    
    # Check if running in Docker
    if [ -f /.dockerenv ]; then
        # Running inside container
        pg_dump -h "$DB_HOST" -U "$DB_USER" "$DB_NAME" | gzip > "$BACKUP_FILE"
    else
        # Running on host, use docker-compose
        if command -v docker-compose &> /dev/null; then
            COMPOSE_CMD="docker-compose"
        else
            COMPOSE_CMD="docker compose"
        fi
        
        $COMPOSE_CMD exec -T postgres pg_dump -U "$DB_USER" "$DB_NAME" | gzip > "$BACKUP_FILE"
    fi
    
    BACKUP_SIZE=$(du -h "$BACKUP_FILE" | cut -f1)
    echo -e "${GREEN}✓ Backup created: $BACKUP_FILE ($BACKUP_SIZE)${NC}"
    
    # Upload to S3 if configured
    if [ -n "$S3_BUCKET" ]; then
        echo "☁️  Uploading to S3..."
        aws s3 cp "$BACKUP_FILE" "s3://${S3_BUCKET}/backups/$(basename $BACKUP_FILE)"
        echo -e "${GREEN}✓ Uploaded to S3${NC}"
    fi
    
    # Cleanup old backups
    echo "🧹 Cleaning up old backups (older than $RETENTION_DAYS days)..."
    find "$BACKUP_DIR" -name "backup_*.sql.gz" -mtime +$RETENTION_DAYS -delete
    
    BACKUP_COUNT=$(ls -1 "$BACKUP_DIR"/backup_*.sql.gz 2>/dev/null | wc -l)
    echo -e "${GREEN}✓ Cleanup complete. $BACKUP_COUNT backups remaining.${NC}"
}

# Restore backup
do_restore() {
    if [ -z "$RESTORE_FILE" ]; then
        echo -e "${RED}❌ Please specify a backup file to restore${NC}"
        usage
        exit 1
    fi
    
    if [ ! -f "$RESTORE_FILE" ]; then
        echo -e "${RED}❌ Backup file not found: $RESTORE_FILE${NC}"
        exit 1
    fi
    
    echo -e "${YELLOW}⚠️  WARNING: This will overwrite the current database!${NC}"
    read -p "Are you sure you want to continue? (y/N) " -n 1 -r
    echo
    
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo "Restore cancelled."
        exit 0
    fi
    
    echo "📥 Restoring from $RESTORE_FILE..."
    
    DB_HOST="${DB_HOST:-localhost}"
    DB_USER="${DB_USER:-ispmonitor}"
    DB_NAME="${DB_NAME:-ispmonitor}"
    
    # Check if running in Docker
    if [ -f /.dockerenv ]; then
        gunzip -c "$RESTORE_FILE" | psql -h "$DB_HOST" -U "$DB_USER" "$DB_NAME"
    else
        if command -v docker-compose &> /dev/null; then
            COMPOSE_CMD="docker-compose"
        else
            COMPOSE_CMD="docker compose"
        fi
        
        gunzip -c "$RESTORE_FILE" | $COMPOSE_CMD exec -T postgres psql -U "$DB_USER" "$DB_NAME"
    fi
    
    echo -e "${GREEN}✓ Restore complete${NC}"
}

# List backups
do_list() {
    echo "📋 Available backups in $BACKUP_DIR:"
    echo ""
    
    if [ ! -d "$BACKUP_DIR" ]; then
        echo "   No backup directory found"
        exit 0
    fi
    
    ls -lh "$BACKUP_DIR"/backup_*.sql.gz 2>/dev/null | awk '{print "   " $9 " (" $5 ")"}'
    
    BACKUP_COUNT=$(ls -1 "$BACKUP_DIR"/backup_*.sql.gz 2>/dev/null | wc -l)
    echo ""
    echo "Total: $BACKUP_COUNT backups"
}

# Main
main() {
    parse_args "$@"
    
    case $COMMAND in
        backup)
            do_backup
            ;;
        restore)
            do_restore
            ;;
        list)
            do_list
            ;;
        *)
            usage
            exit 1
            ;;
    esac
}

main "$@"
