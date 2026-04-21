#!/bin/bash

set -e

echo "=========================================="
echo "DeepWrite Server Setup Script"
echo "=========================================="
echo ""

install_docker() {
    echo "[1/6] Installing Docker..."
    
    if command -v docker &> /dev/null; then
        echo "✓ Docker already installed"
        return
    fi
    
    curl -fsSL https://get.docker.com -o get-docker.sh
    sh get-docker.sh
    rm get-docker.sh
    
    usermod -aG docker "${USER:-root}"
    
    echo "✓ Docker installed"
}

install_dependencies() {
    echo ""
    echo "[2/6] Installing additional dependencies..."
    
    apt-get update
    apt-get install -y \
        curl \
        git \
        nginx \
        certbot \
        python3-certbot-nginx \
        ufw
    
    echo "✓ Dependencies installed"
}

setup_firewall() {
    echo ""
    echo "[3/6] Configuring firewall..."
    
    ufw --force reset
    ufw default deny incoming
    ufw default allow outgoing
    ufw allow ssh
    ufw allow 'Nginx Full'
    ufw --force enable
    
    echo "✓ Firewall configured"
}

setup_ssl() {
    echo ""
    echo "[4/6] Setting up SSL certificates..."
    
    read -p "Enter your domain name: " DOMAIN
    
    if [ -n "$DOMAIN" ]; then
        certbot --nginx -d "$DOMAIN" --non-interactive --agree-tos --email "admin@${DOMAIN}" || true
        echo "✓ SSL configured for $DOMAIN"
    else
        echo "⚠ Skipping SSL setup (no domain provided)"
    fi
}

clone_repository() {
    echo ""
    echo "[5/6] Cloning repository..."
    
    DEPLOY_PATH="${DEPLOY_PATH:-/opt/deepwrite}"
    
    if [ -d "$DEPLOY_PATH" ]; then
        echo "Directory exists, pulling latest changes..."
        cd "$DEPLOY_PATH"
        git pull || true
    else
        read -p "Enter Git repository URL: " REPO_URL
        git clone "$REPO_URL" "$DEPLOY_PATH"
        cd "$DEPLOY_PATH"
    fi
    
    echo "✓ Repository ready at $DEPLOY_PATH"
}

create_env_file() {
    echo ""
    echo "[6/6] Creating environment file..."
    
    ENV_FILE=".env.production"
    
    if [ -f "$ENV_FILE" ]; then
        echo "⚠ Environment file already exists"
        return
    fi
    
    read -p "Enter database password: " DB_PASS
    read -p "Enter JWT secret (or press enter to generate): " JWT_KEY
    
    JWT_SECRET="${JWT_KEY:-$(openssl rand -hex 32)}"
    
    cat > "$ENV_FILE" << EOF
# DeepWrite Production Environment
# Generated on $(date)

# Database
DB_USER=deepwrite
DB_PASSWORD=${DB_PASS}
DB_NAME=deepwrite

# Security
JWT_SECRET=${JWT_SECRET}

# Optional: External ports
# DB_EXTERNAL_PORT=5432
# REDIS_EXTERNAL_PORT=6379
EOF
    
    chmod 600 "$ENV_FILE"
    
    echo "✓ Environment file created: $ENV_FILE"
    echo ""
    echo "⚠ Please edit $ENV_FILE to add any additional configuration"
}

print_next_steps() {
    echo ""
    echo "=========================================="
    echo "Server Setup Complete!"
    echo "=========================================="
    echo ""
    echo "Next steps:"
    echo "  1. Review and edit .env.production"
    echo "  2. Configure nginx (see nginx/nginx.conf.example)"
    echo "  3. Run: ./scripts/deploy.sh"
    echo ""
    echo "For GitHub Actions deployment, add these secrets:"
    echo "  - DEPLOY_HOST: $(curl -s ifconfig.me || echo 'your-server-ip')"
    echo "  - DEPLOY_USER: ${USER:-root}"
    echo "  - DEPLOY_KEY: (SSH private key)"
    echo "  - DEPLOY_PATH: /opt/deepwrite"
    echo "  - DB_PASSWORD: (from .env.production)"
    echo "  - JWT_SECRET: (from .env.production)"
}

main() {
    install_docker
    install_dependencies
    setup_firewall
    setup_ssl
    clone_repository
    create_env_file
    print_next_steps
}

main "$@"