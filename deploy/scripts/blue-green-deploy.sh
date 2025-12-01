#!/bin/bash

# ============================================
# Blue-Green Deployment Script
# Zero Downtime Deployment for GoShort
# ============================================

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
DEPLOY_PATH="${DEPLOY_PATH:-/opt/goshort}"
REGISTRY="${CI_REGISTRY}"
IMAGE_BACKEND="${CI_REGISTRY_IMAGE}/backend:${TAG:-latest}"
IMAGE_FRONTEND="${CI_REGISTRY_IMAGE}/frontend:${TAG:-latest}"
HEALTH_CHECK_TIMEOUT=60
HEALTH_CHECK_INTERVAL=2

# Functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to check health
check_health() {
    local container_name=$1
    local max_attempts=$((HEALTH_CHECK_TIMEOUT / HEALTH_CHECK_INTERVAL))
    
    log_info "Checking health of $container_name..."
    
    for i in $(seq 1 $max_attempts); do
        if docker exec "$container_name" curl -f http://localhost:8080/api/v1/health >/dev/null 2>&1; then
            log_success "$container_name is healthy!"
            return 0
        fi
        
        log_info "Attempt $i/$max_attempts: waiting for $container_name..."
        sleep $HEALTH_CHECK_INTERVAL
    done
    
    log_error "$container_name failed health check after $HEALTH_CHECK_TIMEOUT seconds"
    return 1
}

# Function to get active color
get_active_color() {
    if docker ps --format '{{.Names}}' | grep -q "goshort-backend-blue"; then
        if [ "$(docker inspect -f '{{.State.Running}}' goshort-backend-blue 2>/dev/null)" = "true" ]; then
            echo "blue"
            return
        fi
    fi
    echo "green"
}

# Function to switch nginx upstream
switch_nginx() {
    local target_color=$1
    
    log_info "Switching nginx to $target_color environment..."
    
    # Update upstream config
    echo "server backend-${target_color}:8080 max_fails=3 fail_timeout=30s;" | \
        sudo tee "${DEPLOY_PATH}/nginx/upstreams/backend_active.conf" > /dev/null
    
    # Test nginx config
    if ! docker exec goshort-nginx nginx -t 2>&1 | grep -q "successful"; then
        log_error "Nginx configuration test failed"
        return 1
    fi
    
    # Reload nginx
    docker exec goshort-nginx nginx -s reload
    
    log_success "Nginx switched to $target_color"
    return 0
}

# Function to rollback
rollback() {
    local previous_color=$1
    
    log_warning "Initiating rollback to $previous_color..."
    
    # Start previous environment
    docker-compose -f "${DEPLOY_PATH}/docker-compose.prod.yml" up -d "backend-${previous_color}"
    
    # Check health
    if ! check_health "goshort-backend-${previous_color}"; then
        log_error "Rollback failed - previous environment is unhealthy"
        return 1
    fi
    
    # Switch nginx back
    switch_nginx "$previous_color"
    
    log_success "Rollback completed successfully"
    return 0
}

# Function to send notification
send_notification() {
    local message=$1
    local status=$2
    
    if [ -n "$SLACK_WEBHOOK_URL" ]; then
        curl -X POST "$SLACK_WEBHOOK_URL" \
            -H 'Content-Type: application/json' \
            -d "{\"text\":\"🚀 GoShort Deployment: $message\",\"username\":\"DeployBot\",\"icon_emoji\":\":rocket:\"}" \
            >/dev/null 2>&1 || true
    fi
    
    if [ -n "$TELEGRAM_BOT_TOKEN" ] && [ -n "$TELEGRAM_CHAT_ID" ]; then
        local emoji="✅"
        [ "$status" = "error" ] && emoji="❌"
        
        curl -s -X POST "https://api.telegram.org/bot${TELEGRAM_BOT_TOKEN}/sendMessage" \
            -d "chat_id=${TELEGRAM_CHAT_ID}" \
            -d "text=${emoji} GoShort Deployment: ${message}" \
            >/dev/null 2>&1 || true
    fi
}

# Main deployment logic
main() {
    log_info "=========================================="
    log_info "Starting Blue-Green Deployment"
    log_info "=========================================="
    
    # Change to deployment directory
    cd "$DEPLOY_PATH" || {
        log_error "Failed to change to deployment directory: $DEPLOY_PATH"
        exit 1
    }
    
    # Login to registry
    log_info "Logging in to container registry..."
    echo "${CI_REGISTRY_PASSWORD}" | docker login -u "${CI_REGISTRY_USER}" --password-stdin "${CI_REGISTRY}" || {
        log_error "Failed to login to registry"
        exit 1
    }
    
    # Pull latest images
    log_info "Pulling latest images..."
    docker pull "$IMAGE_BACKEND" || {
        log_error "Failed to pull backend image"
        exit 1
    }
    docker pull "$IMAGE_FRONTEND" || {
        log_error "Failed to pull frontend image"
        exit 1
    }
    
    # Determine active and inactive colors
    ACTIVE_COLOR=$(get_active_color)
    
    if [ "$ACTIVE_COLOR" = "blue" ]; then
        INACTIVE_COLOR="green"
    else
        INACTIVE_COLOR="blue"
    fi
    
    log_info "Current active environment: $ACTIVE_COLOR"
    log_info "Deploying to: $INACTIVE_COLOR"
    
    # Update inactive environment
    log_info "Starting $INACTIVE_COLOR environment..."
    docker-compose -f docker-compose.prod.yml up -d "backend-${INACTIVE_COLOR}" || {
        log_error "Failed to start $INACTIVE_COLOR environment"
        exit 1
    }
    
    # Wait for health check
    if ! check_health "goshort-backend-${INACTIVE_COLOR}"; then
        log_error "Health check failed for $INACTIVE_COLOR"
        
        # Attempt rollback
        log_warning "Attempting to keep $ACTIVE_COLOR running..."
        docker-compose -f docker-compose.prod.yml stop "backend-${INACTIVE_COLOR}"
        
        send_notification "Deployment failed - health check timeout for $INACTIVE_COLOR environment" "error"
        exit 1
    fi
    
    # Perform smoke tests
    log_info "Running smoke tests on $INACTIVE_COLOR..."
    
    # Test health endpoint
    if ! docker exec "goshort-backend-${INACTIVE_COLOR}" curl -f http://localhost:8080/api/v1/health >/dev/null 2>&1; then
        log_error "Smoke test failed: health endpoint"
        rollback "$ACTIVE_COLOR"
        send_notification "Deployment failed - smoke tests failed" "error"
        exit 1
    fi
    
    # Test ready endpoint
    if ! docker exec "goshort-backend-${INACTIVE_COLOR}" curl -f http://localhost:8080/api/v1/ready >/dev/null 2>&1; then
        log_warning "Ready endpoint check failed, continuing..."
    fi
    
    log_success "Smoke tests passed"
    
    # Switch nginx upstream
    if ! switch_nginx "$INACTIVE_COLOR"; then
        log_error "Failed to switch nginx"
        rollback "$ACTIVE_COLOR"
        send_notification "Deployment failed - nginx switch failed" "error"
        exit 1
    fi
    
    # Wait for connections to drain
    log_info "Waiting for connections to drain from $ACTIVE_COLOR..."
    sleep 5
    
    # Stop old active environment
    log_info "Stopping $ACTIVE_COLOR environment..."
    docker-compose -f docker-compose.prod.yml stop "backend-${ACTIVE_COLOR}"
    
    # Update frontend
    log_info "Updating frontend..."
    docker-compose -f docker-compose.prod.yml up -d frontend || {
        log_warning "Failed to update frontend, but backend is running"
    }
    
    # Cleanup old images
    log_info "Cleaning up old images..."
    docker image prune -af --filter "until=72h" >/dev/null 2>&1 || true
    
    # Final verification
    log_info "Verifying deployment..."
    sleep 3
    
    if ! check_health "goshort-backend-${INACTIVE_COLOR}"; then
        log_error "Final verification failed"
        rollback "$ACTIVE_COLOR"
        send_notification "Deployment failed - final verification failed" "error"
        exit 1
    fi
    
    log_success "=========================================="
    log_success "Deployment completed successfully!"
    log_success "Active environment: $INACTIVE_COLOR"
    log_success "Previous environment: $ACTIVE_COLOR (stopped)"
    log_success "=========================================="
    
    # Send success notification
    send_notification "Deployment completed successfully! Active: $INACTIVE_COLOR" "success"
    
    # Display deployment info
    echo ""
    log_info "Deployment Summary:"
    log_info "  Backend Image: $IMAGE_BACKEND"
    log_info "  Frontend Image: $IMAGE_FRONTEND"
    log_info "  Active Environment: $INACTIVE_COLOR"
    log_info "  Commit: ${CI_COMMIT_SHORT_SHA:-unknown}"
    log_info "  Deployed by: ${GITLAB_USER_NAME:-unknown}"
    log_info "  Timestamp: $(date)"
    echo ""
}

# Trap errors
trap 'log_error "Deployment failed with error on line $LINENO"' ERR

# Run main function
main "$@"

