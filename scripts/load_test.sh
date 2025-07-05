#!/bin/bash

# Load Test Script for Bookstore API
# This script provides easy-to-use commands for load testing

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
API_URL="http://localhost:8080"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to print colored output
print_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to check if API is running
check_api() {
    print_info "Checking if API is running..."
    
    if curl -f -s "$API_URL/health" > /dev/null 2>&1; then
        print_info "✓ API is running"
        return 0
    else
        print_error "✗ API is not running"
        print_error "Please start the API with: docker-compose up -d"
        return 1
    fi
}

# Function to show help
show_help() {
    cat << EOF
Bookstore API Load Test Script

Usage: $0 [COMMAND] [OPTIONS]

Commands:
    light       Run light load test (100 requests, 5 workers, 30 seconds)
    medium      Run medium load test (500 requests, 10 workers, 60 seconds)
    heavy       Run heavy load test (1000 requests, 20 workers, 120 seconds)
    custom      Run custom load test (specify parameters)
    help        Show this help message

Custom Options:
    ./load_test.sh custom <requests> <concurrency> <duration_seconds>
    
    Example:
        ./load_test.sh custom 200 8 45
        
Environment Check:
    check       Check if API and monitoring services are running

Examples:
    $0 light          # Quick test
    $0 medium         # Standard test
    $0 heavy          # Stress test
    $0 custom 300 15 90   # Custom: 300 requests, 15 workers, 90 seconds
    $0 check          # Check services

The script will:
- Generate realistic CRUD operations (Create, Read, Update, Delete)
- Use weighted distribution: 40% Create, 30% Read, 20% Update, 10% Delete
- Create random delays between requests (10-500ms)
- Show real-time progress and final statistics
- Test with realistic book data

Monitoring:
- Metrics: http://localhost:8080/metrics
- Prometheus: http://localhost:9090
- Grafana: http://localhost:3000 (admin/admin)

EOF
}

# Function to run Go load test
run_load_test() {
    local requests=$1
    local concurrency=$2
    local duration=$3
    
    print_info "Starting load test..."
    print_info "Requests: $requests | Concurrency: $concurrency | Duration: ${duration}s"
    
    cd "$SCRIPT_DIR"
    
    # Check if Go is installed
    if ! command -v go &> /dev/null; then
        print_error "Go is not installed. Please install Go 1.23+ to run the load test."
        return 1
    fi
    
    # Run the load test
    go run loadtest.go "$requests" "$concurrency" "$duration"
}

# Function to check all services
check_services() {
    print_info "Checking all services..."
    
    # Check API
    if curl -f -s "$API_URL/health" > /dev/null 2>&1; then
        print_info "✓ API Server (port 8080)"
    else
        print_error "✗ API Server (port 8080)"
    fi
    
    # Check Prometheus
    if curl -f -s "http://localhost:9090/-/healthy" > /dev/null 2>&1; then
        print_info "✓ Prometheus (port 9090)"
    else
        print_error "✗ Prometheus (port 9090)"
    fi
    
    # Check Grafana
    if curl -f -s "http://localhost:3000/api/health" > /dev/null 2>&1; then
        print_info "✓ Grafana (port 3000)"
    else
        print_error "✗ Grafana (port 3000)"
    fi
    
    # Check PostgreSQL
    if docker-compose ps | grep -q "bookstore-postgres.*Up"; then
        print_info "✓ PostgreSQL (port 5432)"
    else
        print_error "✗ PostgreSQL (port 5432)"
    fi
    
    echo ""
    print_info "If any service is down, run: docker-compose up -d"
}

# Main script logic
case "${1:-help}" in
    light)
        check_api || exit 1
        run_load_test 100 5 30
        ;;
    medium)
        check_api || exit 1
        run_load_test 500 10 60
        ;;
    heavy)
        check_api || exit 1
        run_load_test 1000 20 120
        ;;
    custom)
        if [ $# -ne 4 ]; then
            print_error "Custom mode requires 3 parameters: requests, concurrency, duration"
            echo "Usage: $0 custom <requests> <concurrency> <duration_seconds>"
            exit 1
        fi
        check_api || exit 1
        run_load_test "$2" "$3" "$4"
        ;;
    check)
        check_services
        ;;
    help|--help|-h)
        show_help
        ;;
    *)
        print_error "Unknown command: $1"
        show_help
        exit 1
        ;;
esac