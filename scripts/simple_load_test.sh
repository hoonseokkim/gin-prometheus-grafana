#!/bin/bash

# Simple Load Test Script using curl
# This script generates realistic load for monitoring demo

set -e

API_URL="http://localhost:8080/api/v1/books"
HEALTH_URL="http://localhost:8080/health"

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

print_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Sample book data
declare -a BOOKS=(
    '{"title":"The Go Programming Language","author":"Alan Donovan","isbn":"9780134190440","price":49.99,"published_at":"2015-11-16T00:00:00Z"}'
    '{"title":"Clean Code","author":"Robert C. Martin","isbn":"9780132350884","price":39.99,"published_at":"2008-08-11T00:00:00Z"}'
    '{"title":"Design Patterns","author":"Gang of Four","isbn":"9780201633610","price":54.99,"published_at":"1994-10-21T00:00:00Z"}'
    '{"title":"Refactoring","author":"Martin Fowler","isbn":"9780201485677","price":47.99,"published_at":"1999-07-08T00:00:00Z"}'
    '{"title":"Clean Architecture","author":"Robert C. Martin","isbn":"9780134494166","price":42.99,"published_at":"2017-09-20T00:00:00Z"}'
    '{"title":"Concurrency in Go","author":"Katherine Cox-Buday","isbn":"9781491941195","price":39.99,"published_at":"2017-07-19T00:00:00Z"}'
    '{"title":"Go in Action","author":"William Kennedy","isbn":"9781617291784","price":44.99,"published_at":"2015-11-04T00:00:00Z"}'
    '{"title":"Learning Go","author":"Jon Bodner","isbn":"9781492077213","price":49.99,"published_at":"2021-03-02T00:00:00Z"}'
    '{"title":"Building Microservices","author":"Sam Newman","isbn":"9781491950357","price":54.99,"published_at":"2015-02-20T00:00:00Z"}'
    '{"title":"The Pragmatic Programmer","author":"David Thomas","isbn":"9780201616224","price":49.99,"published_at":"1999-10-30T00:00:00Z"}'
)

# Function to check API health
check_api() {
    if ! curl -f -s "$HEALTH_URL" > /dev/null 2>&1; then
        print_error "API is not running on localhost:8080"
        print_error "Please start with: docker-compose up -d"
        exit 1
    fi
}

# Function to create a random book
create_book() {
    local book_index=$((RANDOM % ${#BOOKS[@]}))
    local book_data="${BOOKS[$book_index]}"
    
    # Add random suffix to ISBN to make it unique
    local unique_isbn=$(echo "$book_data" | jq -r '.isbn')$((RANDOM % 10000))
    local modified_book=$(echo "$book_data" | jq --arg isbn "$unique_isbn" '.isbn = $isbn')
    
    curl -s -X POST "$API_URL" \
        -H "Content-Type: application/json" \
        -d "$modified_book" > /dev/null
}

# Function to read all books
read_all_books() {
    curl -s "$API_URL" > /dev/null
}

# Function to read a specific book by ID
read_book_by_id() {
    local book_id=$1
    curl -s "$API_URL/$book_id" > /dev/null 2>&1
}

# Function to update a book
update_book() {
    local book_id=$1
    local new_price=$((RANDOM % 50 + 20)).99
    local update_data="{\"price\": $new_price}"
    
    curl -s -X PUT "$API_URL/$book_id" \
        -H "Content-Type: application/json" \
        -d "$update_data" > /dev/null 2>&1
}

# Function to delete a book
delete_book() {
    local book_id=$1
    curl -s -X DELETE "$API_URL/$book_id" > /dev/null 2>&1
}

# Function to get a random book ID from the API
get_random_book_id() {
    local response=$(curl -s "$API_URL")
    
    # Check if response is valid JSON array
    if echo "$response" | jq -e '. | type == "array"' > /dev/null 2>&1; then
        local book_ids=$(echo "$response" | jq -r '.[].id' 2>/dev/null | head -10)
        if [ -n "$book_ids" ] && [ "$book_ids" != "" ]; then
            # Use awk instead of shuf for better compatibility
            echo "$book_ids" | awk 'BEGIN{srand()} {lines[NR]=$0} END{if(NR>0) print lines[int(rand()*NR)+1]}'
        else
            echo ""
        fi
    else
        echo ""
    fi
}

# Function to run load test
run_load_test() {
    local duration=$1
    local requests_per_second=$2
    local total_requests=$((duration * requests_per_second))
    
    print_info "Starting load test for ${duration}s at ${requests_per_second} req/s"
    print_info "Total target requests: $total_requests"
    
    # Initialize with some books
    print_info "Creating initial books..."
    for i in {1..10}; do
        create_book
        sleep 0.1
    done
    
    local request_count=0
    local start_time=$(date +%s)
    
    print_info "Starting main load test..."
    
    while [ $(($(date +%s) - start_time)) -lt $duration ]; do
        # Random operation selection (weighted)
        local rand=$((RANDOM % 100))
        
        if [ $rand -lt 35 ]; then
            # 35% - Create book
            create_book
            
        elif [ $rand -lt 65 ]; then
            # 30% - Read operations  
            if [ $((RANDOM % 2)) -eq 0 ]; then
                read_all_books
            else
                local book_id=$(get_random_book_id)
                if [ -n "$book_id" ]; then
                    read_book_by_id "$book_id"
                else
                    read_all_books
                fi
            fi
            
        elif [ $rand -lt 90 ]; then
            # 25% - Update book
            local book_id=$(get_random_book_id)
            if [ -n "$book_id" ]; then
                update_book "$book_id"
            fi
            
        else
            # 10% - Delete book
            local book_id=$(get_random_book_id)
            if [ -n "$book_id" ]; then
                delete_book "$book_id"
            fi
        fi
        
        ((request_count++))
        
        # Progress reporting
        if [ $((request_count % 50)) -eq 0 ]; then
            local elapsed=$(($(date +%s) - start_time))
            local current_rps=$((request_count / (elapsed + 1)))
            echo "Progress: $request_count requests in ${elapsed}s (~${current_rps} req/s)"
        fi
        
        # Sleep to maintain target RPS
        sleep $(echo "scale=3; 1.0 / $requests_per_second" | bc -l)
    done
    
    local final_elapsed=$(($(date +%s) - start_time))
    local final_rps=$((request_count / final_elapsed))
    
    print_info "Load test completed!"
    print_info "Total requests: $request_count in ${final_elapsed}s"
    print_info "Average RPS: $final_rps"
    
    # Show final book count
    local response=$(curl -s "$API_URL")
    local book_count=0
    if echo "$response" | jq -e '. | type == "array"' > /dev/null 2>&1; then
        book_count=$(echo "$response" | jq length 2>/dev/null || echo "0")
    fi
    print_info "Final book count: $book_count"
}

# Show help
show_help() {
    cat << EOF
Simple Load Test Script for Bookstore API

Usage: $0 [COMMAND]

Commands:
    light       30 seconds at 10 req/s (300 requests)
    medium      60 seconds at 15 req/s (900 requests)  
    heavy       120 seconds at 20 req/s (2400 requests)
    demo        300 seconds at 5 req/s (1500 requests) - For dashboard demo
    check       Check if services are running
    help        Show this help

The script performs weighted operations:
- 35% Create new books
- 30% Read operations (all books or by ID)
- 25% Update existing books  
- 10% Delete books

Monitoring URLs:
- API Metrics: http://localhost:8080/metrics
- Prometheus: http://localhost:9090
- Grafana: http://localhost:3000 (admin/admin)

EOF
}

# Check if bc is available for precise timing
if ! command -v bc &> /dev/null; then
    print_warning "bc calculator not found. Install with: brew install bc (macOS) or apt-get install bc (Linux)"
    print_warning "Using basic timing instead (less precise)"
    
    # Simple sleep function without bc
    precise_sleep() {
        sleep 0.1
    }
else
    precise_sleep() {
        local delay=$(echo "scale=3; 1.0 / $1" | bc -l)
        sleep "$delay"
    }
fi

# Main script logic
case "${1:-help}" in
    light)
        check_api
        run_load_test 30 10
        ;;
    medium)
        check_api
        run_load_test 60 15
        ;;
    heavy)
        check_api
        run_load_test 120 20
        ;;
    demo)
        check_api
        print_info "Starting 5-minute demo load test (perfect for dashboard watching)"
        run_load_test 300 5
        ;;
    check)
        print_info "Checking services..."
        if curl -f -s "$HEALTH_URL" > /dev/null 2>&1; then
            print_info "✓ API Server (port 8080)"
        else
            print_error "✗ API Server (port 8080)"
        fi
        
        if curl -f -s "http://localhost:9090/-/healthy" > /dev/null 2>&1; then
            print_info "✓ Prometheus (port 9090)"
        else
            print_error "✗ Prometheus (port 9090)"
        fi
        
        if curl -f -s "http://localhost:3000/api/health" > /dev/null 2>&1; then
            print_info "✓ Grafana (port 3000)"
        else
            print_error "✗ Grafana (port 3000)"
        fi
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