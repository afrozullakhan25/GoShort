#!/bin/bash

# ============================================
# Server Setup Script
# Prepare server for GoShort deployment
# ============================================

set -e

# Configuration
DEPLOY_USER="deployuser"
DEPLOY_PATH="/opt/goshort"
SSH_PORT="22"

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m'

log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

main() {
    log_info "Starting server setup..."
    
    # Update system
    log_info "Updating system packages..."
    sudo apt update
    sudo apt upgrade -y
    
    # Install Docker
    log_info "Installing Docker..."
    curl -fsSL https://get.docker.com -o get-docker.sh
    sudo sh get-docker.sh
    rm get-docker.sh
    
    # Install Docker Compose
    log_info "Installing Docker Compose..."
    sudo curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
    sudo chmod +x /usr/local/bin/docker-compose
    
    # Create deploy user
    log_info "Creating deploy user..."
    sudo useradd -m -s /bin/bash "$DEPLOY_USER" || true
    sudo usermod -aG docker "$DEPLOY_USER"
    
    # Create deployment directory
    log_info "Creating deployment directory..."
    sudo mkdir -p "$DEPLOY_PATH"
    sudo mkdir -p "$DEPLOY_PATH/nginx/upstreams"
    sudo chown -R "$DEPLOY_USER:$DEPLOY_USER" "$DEPLOY_PATH"
    
    # Install additional tools
    log_info "Installing additional tools..."
    sudo apt install -y curl wget git vim htop nginx
    
    # Configure firewall
    log_info "Configuring firewall..."
    sudo ufw allow "$SSH_PORT/tcp"
    sudo ufw allow 80/tcp
    sudo ufw allow 443/tcp
    sudo ufw --force enable
    
    # Setup log rotation
    log_info "Setting up log rotation..."
    sudo tee /etc/logrotate.d/goshort > /dev/null <<EOF
/var/log/goshort/*.log {
    daily
    rotate 7
    compress
    delaycompress
    missingok
    notifempty
    create 0640 $DEPLOY_USER $DEPLOY_USER
}
EOF
    
    # Create systemd service for monitoring
    log_info "Creating monitoring service..."
    sudo tee /etc/systemd/system/goshort-monitor.service > /dev/null <<EOF
[Unit]
Description=GoShort Monitoring Service
After=docker.service
Requires=docker.service

[Service]
Type=oneshot
ExecStart=$DEPLOY_PATH/deploy/scripts/monitor.sh health
User=$DEPLOY_USER
StandardOutput=journal

[Install]
WantedBy=multi-user.target
EOF
    
    # Create monitoring timer
    sudo tee /etc/systemd/system/goshort-monitor.timer > /dev/null <<EOF
[Unit]
Description=GoShort Monitoring Timer
Requires=goshort-monitor.service

[Timer]
OnBootSec=5min
OnUnitActiveSec=5min

[Install]
WantedBy=timers.target
EOF
    
    # Enable monitoring timer
    sudo systemctl daemon-reload
    sudo systemctl enable goshort-monitor.timer
    sudo systemctl start goshort-monitor.timer
    
    log_success "Server setup completed!"
    log_info "Next steps:"
    log_info "1. Add SSH key for $DEPLOY_USER"
    log_info "2. Configure GitLab CI/CD variables"
    log_info "3. Push code to trigger deployment"
}

main "$@"

