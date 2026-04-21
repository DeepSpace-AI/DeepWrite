#!/bin/bash

set -e

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$REPO_ROOT"

IMAGE_TAG="${IMAGE_TAG:-latest}"
GITHUB_REPOSITORY="${GITHUB_REPOSITORY:-deepwrite}"

export IMAGE_TAG
export GITHUB_REPOSITORY

echo "=========================================="
echo "DeepWrite Deployment Script"
echo "=========================================="
echo "Image Tag: $IMAGE_TAG"
echo "Repository: $GITHUB_REPOSITORY"
echo ""

check_requirements() {
    echo "[1/5] Checking requirements..."
    
    if ! command -v docker &> /dev/null; then
        echo "Error: Docker is not installed"
        exit 1
    fi
    
    if ! command -v docker compose &> /dev/null; then
        echo "Error: Docker Compose is not installed"
        exit 1
    fi
    
    if [ -z "$DB_PASSWORD" ]; then
        echo "Error: DB_PASSWORD environment variable is required"
        exit 1
    fi
    
    if [ -z "$JWT_SECRET" ]; then
        echo "Error: JWT_SECRET environment variable is required"
        exit 1
    fi
    
    echo "✓ All requirements met"
}

pull_images() {
    echo ""
    echo "[2/5] Pulling Docker images..."
    
    docker compose -f docker-compose.prod.yml pull
    
    echo "✓ Images pulled"
}

backup_database() {
    echo ""
    echo "[3/5] Backing up database..."
    
    BACKUP_FILE="backup_$(date +%Y%m%d_%H%M%S).sql"
    docker compose -f docker-compose.prod.yml exec -T postgres \
        pg_dump -U "${DB_USER:-deepwrite}" "${DB_NAME:-deepwrite}" > "$BACKUP_FILE" 2>/dev/null || true
    
    if [ -f "$BACKUP_FILE" ] && [ -s "$BACKUP_FILE" ]; then
        echo "✓ Database backup created: $BACKUP_FILE"
    else
        echo "⚠ No existing database to backup or backup failed"
    fi
}

deploy_services() {
    echo ""
    echo "[4/5] Deploying services..."
    
    docker compose -f docker-compose.prod.yml up -d --remove-orphans
    
    echo "✓ Services deployed"
}

cleanup() {
    echo ""
    echo "[5/5] Cleaning up..."
    
    docker image prune -f
    
    echo "✓ Cleanup complete"
}

print_status() {
    echo ""
    echo "=========================================="
    echo "Deployment Complete!"
    echo "=========================================="
    echo ""
    echo "Services:"
    docker compose -f docker-compose.prod.yml ps
    echo ""
    echo "Access your application at:"
    echo "  - Web:  http://localhost:5173"
    echo "  - Admin: http://localhost:5174"
    echo "  - API:  http://localhost:8080"
    echo "  - AI:   http://localhost:8002"
}

main() {
    check_requirements
    pull_images
    backup_database
    deploy_services
    cleanup
    print_status
}

main "$@"