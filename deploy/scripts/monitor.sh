#!/bin/bash

# ============================================
# Monitoring Script
# ============================================

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

show_status() {
    echo -e "${BLUE}=========================================="
    echo "GoShort Service Status"
    echo -e "==========================================${NC}"
    echo ""
    
    # Container status
    echo -e "${YELLOW}Container Status:${NC}"
    docker ps --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}" | grep goshort || echo "No containers running"
    echo ""
    
    # Active environment
    echo -e "${YELLOW}Active Environment:${NC}"
    if [ -f /etc/nginx/upstreams/backend_active.conf ]; then
        grep "server" /etc/nginx/upstreams/backend_active.conf
    else
        echo "Configuration not found"
    fi
    echo ""
    
    # Resource usage
    echo -e "${YELLOW}Resource Usage:${NC}"
    docker stats --no-stream --format "table {{.Name}}\t{{.CPUPerc}}\t{{.MemUsage}}" | grep goshort
    echo ""
    
    # Recent logs
    echo -e "${YELLOW}Recent Logs (last 10 lines):${NC}"
    docker logs --tail 10 goshort-backend-blue 2>/dev/null || docker logs --tail 10 goshort-backend-green 2>/dev/null
    echo ""
}

check_metrics() {
    echo -e "${BLUE}=========================================="
    echo "Health Metrics"
    echo -e "==========================================${NC}"
    echo ""
    
    # Backend health
    echo -n "Backend Health: "
    if curl -f -s http://localhost:8080/api/v1/health >/dev/null 2>&1; then
        echo -e "${GREEN}Healthy${NC}"
    else
        echo -e "${RED}Unhealthy${NC}"
    fi
    
    # Frontend health
    echo -n "Frontend Health: "
    if curl -f -s http://localhost/ >/dev/null 2>&1; then
        echo -e "${GREEN}Healthy${NC}"
    else
        echo -e "${RED}Unhealthy${NC}"
    fi
    
    # Database connection
    echo -n "Database: "
    if docker exec goshort-postgres pg_isready -U postgres >/dev/null 2>&1; then
        echo -e "${GREEN}Connected${NC}"
    else
        echo -e "${RED}Disconnected${NC}"
    fi
    
    # Redis connection
    echo -n "Redis: "
    if docker exec goshort-redis redis-cli ping >/dev/null 2>&1; then
        echo -e "${GREEN}Connected${NC}"
    else
        echo -e "${RED}Disconnected${NC}"
    fi
    
    echo ""
}

show_logs() {
    local service=$1
    local lines=${2:-50}
    
    case $service in
        backend-blue)
            docker logs --tail "$lines" -f goshort-backend-blue
            ;;
        backend-green)
            docker logs --tail "$lines" -f goshort-backend-green
            ;;
        frontend)
            docker logs --tail "$lines" -f goshort-frontend
            ;;
        nginx)
            docker logs --tail "$lines" -f goshort-nginx
            ;;
        postgres)
            docker logs --tail "$lines" -f goshort-postgres
            ;;
        redis)
            docker logs --tail "$lines" -f goshort-redis
            ;;
        *)
            echo "Unknown service: $service"
            echo "Available services: backend-blue, backend-green, frontend, nginx, postgres, redis"
            exit 1
            ;;
    esac
}

case "$1" in
    status)
        show_status
        ;;
    health)
        check_metrics
        ;;
    logs)
        show_logs "$2" "$3"
        ;;
    *)
        echo "Usage: $0 {status|health|logs <service> [lines]}"
        echo ""
        echo "Commands:"
        echo "  status  - Show container status and active environment"
        echo "  health  - Check health of all services"
        echo "  logs    - Show logs for a specific service"
        echo ""
        echo "Examples:"
        echo "  $0 status"
        echo "  $0 health"
        echo "  $0 logs backend-blue 100"
        exit 1
        ;;
esac

