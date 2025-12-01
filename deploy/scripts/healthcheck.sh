#!/bin/bash

# ============================================
# Health Check Script
# ============================================

set -e

# Configuration
MAX_RETRIES=5
RETRY_INTERVAL=2
TIMEOUT=5

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

check_endpoint() {
    local url=$1
    local name=$2
    
    echo -n "Checking $name... "
    
    for i in $(seq 1 $MAX_RETRIES); do
        if curl -f -s -m $TIMEOUT "$url" >/dev/null 2>&1; then
            echo -e "${GREEN}OK${NC}"
            return 0
        fi
        
        if [ $i -lt $MAX_RETRIES ]; then
            sleep $RETRY_INTERVAL
        fi
    done
    
    echo -e "${RED}FAILED${NC}"
    return 1
}

main() {
    echo "=========================================="
    echo "Health Check Report"
    echo "=========================================="
    echo ""
    
    local failed=0
    
    # Check backend health
    if ! check_endpoint "http://localhost:8080/api/v1/health" "Backend Health"; then
        ((failed++))
    fi
    
    # Check backend ready
    if ! check_endpoint "http://localhost:8080/api/v1/ready" "Backend Ready"; then
        ((failed++))
    fi
    
    # Check frontend
    if ! check_endpoint "http://localhost/" "Frontend"; then
        ((failed++))
    fi
    
    # Check nginx
    if ! check_endpoint "http://localhost/nginx-health" "Nginx"; then
        ((failed++))
    fi
    
    echo ""
    echo "=========================================="
    
    if [ $failed -eq 0 ]; then
        echo -e "${GREEN}All checks passed!${NC}"
        echo "=========================================="
        return 0
    else
        echo -e "${RED}$failed check(s) failed!${NC}"
        echo "=========================================="
        return 1
    fi
}

main "$@"

