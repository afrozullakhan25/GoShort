#!/bin/bash

# ============================================
# Rollback Script
# ============================================

set -e

DEPLOY_PATH="${DEPLOY_PATH:-/opt/goshort}"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

main() {
    cd "$DEPLOY_PATH"
    
    log_info "Initiating rollback..."
    
    # Determine current active
    if docker ps --format '{{.Names}}' | grep -q "goshort-backend-blue"; then
        CURRENT="blue"
        PREVIOUS="green"
    else
        CURRENT="green"
        PREVIOUS="blue"
    fi
    
    log_info "Current active: $CURRENT"
    log_info "Rolling back to: $PREVIOUS"
    
    # Start previous environment
    log_info "Starting $PREVIOUS environment..."
    docker-compose -f docker-compose.prod.yml up -d "backend-${PREVIOUS}"
    
    # Wait for health
    log_info "Waiting for health check..."
    sleep 10
    
    if ! docker exec "goshort-backend-${PREVIOUS}" curl -f http://localhost:8080/api/v1/health >/dev/null 2>&1; then
        log_error "Health check failed for $PREVIOUS"
        exit 1
    fi
    
    # Switch nginx
    log_info "Switching nginx to $PREVIOUS..."
    echo "server backend-${PREVIOUS}:8080 max_fails=3 fail_timeout=30s;" | \
        sudo tee "${DEPLOY_PATH}/nginx/upstreams/backend_active.conf" > /dev/null
    
    docker exec goshort-nginx nginx -t
    docker exec goshort-nginx nginx -s reload
    
    # Stop current
    log_info "Stopping $CURRENT environment..."
    docker-compose -f docker-compose.prod.yml stop "backend-${CURRENT}"
    
    log_info "Rollback completed successfully!"
    log_info "Active environment is now: $PREVIOUS"
}

main "$@"

