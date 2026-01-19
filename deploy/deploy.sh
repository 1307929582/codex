#!/bin/bash

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Functions
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

check_requirements() {
    log_info "Checking requirements..."

    if ! command -v docker &> /dev/null; then
        log_error "Docker is not installed. Please install Docker first."
        exit 1
    fi

    if ! command -v docker-compose &> /dev/null; then
        log_error "Docker Compose is not installed. Please install Docker Compose first."
        exit 1
    fi

    log_info "All requirements met."
}

setup_env() {
    log_info "Setting up environment variables..."

    if [ ! -f .env ]; then
        if [ -f .env.production.example ]; then
            cp .env.production.example .env
            log_warn ".env file created from template. Please edit it with your actual values."
            log_warn "Required variables: DB_PASSWORD, OPENAI_API_KEY, JWT_SECRET"
            read -p "Press Enter after editing .env file..."
        else
            log_error ".env.production.example not found!"
            exit 1
        fi
    else
        log_info ".env file already exists."
    fi

    # Validate required variables
    source .env
    if [ -z "$DB_PASSWORD" ] || [ -z "$OPENAI_API_KEY" ] || [ -z "$JWT_SECRET" ]; then
        log_error "Missing required environment variables in .env file!"
        log_error "Please set: DB_PASSWORD, OPENAI_API_KEY, JWT_SECRET"
        exit 1
    fi

    log_info "Environment variables configured."
}

build_images() {
    log_info "Building Docker images..."
    docker-compose build
    log_info "Docker images built successfully."
}

start_services() {
    log_info "Starting services..."
    docker-compose up -d
    log_info "Services started successfully."
}

stop_services() {
    log_info "Stopping services..."
    docker-compose down
    log_info "Services stopped."
}

restart_services() {
    log_info "Restarting services..."
    docker-compose restart
    log_info "Services restarted."
}

show_logs() {
    docker-compose logs -f
}

show_status() {
    log_info "Service status:"
    docker-compose ps
}

init_database() {
    log_info "Waiting for database to be ready..."
    sleep 5
    log_info "Database should be initialized automatically by the backend."
}

# Main menu
show_menu() {
    echo ""
    echo "================================"
    echo "  Codex Gateway Deployment CLI"
    echo "================================"
    echo "1. Deploy (First time setup)"
    echo "2. Start services"
    echo "3. Stop services"
    echo "4. Restart services"
    echo "5. View logs"
    echo "6. Show status"
    echo "7. Rebuild images"
    echo "8. Exit"
    echo "================================"
}

# Main script
main() {
    clear
    echo "Codex Gateway Deployment CLI"
    echo "=============================="
    echo ""

    check_requirements

    while true; do
        show_menu
        read -p "Select an option [1-8]: " choice

        case $choice in
            1)
                log_info "Starting first-time deployment..."
                setup_env
                build_images
                start_services
                init_database
                echo ""
                log_info "Deployment complete!"
                log_info "Backend: http://localhost:12322"
                log_info "Frontend: http://localhost:12321"
                log_info "API Docs: http://localhost:12322/health"
                ;;
            2)
                start_services
                ;;
            3)
                stop_services
                ;;
            4)
                restart_services
                ;;
            5)
                show_logs
                ;;
            6)
                show_status
                ;;
            7)
                build_images
                ;;
            8)
                log_info "Exiting..."
                exit 0
                ;;
            *)
                log_error "Invalid option. Please select 1-8."
                ;;
        esac

        echo ""
        read -p "Press Enter to continue..."
    done
}

# Run main
main
