#!/bin/bash

# Stress Testing Script for Dynapins Server
# Tests server limits and behavior under extreme load

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

SERVER_URL="${SERVER_URL:-http://localhost:8080}"
DOMAIN="${TEST_DOMAIN:-google.com}"
RESULTS_DIR="./performance/results"
mkdir -p "$RESULTS_DIR"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)

print_header() {
    echo -e "${BLUE}========================================${NC}"
    echo -e "${BLUE}$1${NC}"
    echo -e "${BLUE}========================================${NC}"
}

print_success() {
    echo -e "${GREEN}✓ $1${NC}"
}

print_error() {
    echo -e "${RED}✗ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠ $1${NC}"
}

print_info() {
    echo -e "${YELLOW}ℹ $1${NC}"
}

check_server() {
    if curl -sf "$SERVER_URL/health" > /dev/null; then
        return 0
    else
        return 1
    fi
}

# Progressive load test - find breaking point
run_progressive_load() {
    print_header "Progressive Load Test - Finding Breaking Point"
    
    local output_file="$RESULTS_DIR/stress_progressive_${TIMESTAMP}.txt"
    
    echo "Progressive Load Stress Test" > "$output_file"
    echo "============================" >> "$output_file"
    echo "Server: $SERVER_URL" >> "$output_file"
    echo "Domain: $DOMAIN" >> "$output_file"
    echo "" >> "$output_file"
    
    local connections=(10 25 50 100 200 500 1000 2000 5000)
    local duration=15
    
    for conn in "${connections[@]}"; do
        print_info "Testing with $conn concurrent connections..."
        
        echo "=== Connections: $conn ===" >> "$output_file"
        
        if ! check_server; then
            print_error "Server became unresponsive at $conn connections!"
            echo "ERROR: Server unresponsive at $conn connections" >> "$output_file"
            break
        fi
        
        if command -v hey &> /dev/null; then
            hey -z "${duration}s" -c "$conn" "$SERVER_URL/v1/pins?domain=$DOMAIN" >> "$output_file" 2>&1
            
            # Check error rate
            local error_count=$(grep -o "errors" "$output_file" | tail -1 | wc -l || echo 0)
            if [ "$error_count" -gt 0 ]; then
                print_warning "Errors detected at $conn connections"
            fi
        else
            print_error "hey not found, skipping test"
            return 1
        fi
        
        echo "" >> "$output_file"
        
        # Cool-down period
        print_info "Cool-down period (5s)..."
        sleep 5
    done
    
    print_success "Progressive load test completed: $output_file"
}

# Spike test - sudden load increase
run_spike_test() {
    print_header "Spike Test - Sudden Load Increase"
    
    local output_file="$RESULTS_DIR/stress_spike_${TIMESTAMP}.txt"
    
    echo "Spike Stress Test" > "$output_file"
    echo "=================" >> "$output_file"
    echo "" >> "$output_file"
    
    print_info "Phase 1: Baseline load (10 connections, 10s)"
    echo "=== Baseline Load ===" >> "$output_file"
    hey -z 10s -c 10 "$SERVER_URL/v1/pins?domain=$DOMAIN" >> "$output_file" 2>&1
    
    print_info "Phase 2: SPIKE! (1000 connections, 10s)"
    echo "" >> "$output_file"
    echo "=== SPIKE ===" >> "$output_file"
    hey -z 10s -c 1000 "$SERVER_URL/v1/pins?domain=$DOMAIN" >> "$output_file" 2>&1
    
    print_info "Phase 3: Return to baseline (10 connections, 10s)"
    echo "" >> "$output_file"
    echo "=== Return to Baseline ===" >> "$output_file"
    hey -z 10s -c 10 "$SERVER_URL/v1/pins?domain=$DOMAIN" >> "$output_file" 2>&1
    
    if check_server; then
        print_success "Server recovered from spike successfully"
    else
        print_error "Server did not recover from spike"
    fi
    
    print_success "Spike test completed: $output_file"
}

# Soak test - sustained load over time
run_soak_test() {
    print_header "Soak Test - Sustained Load"
    
    local duration=${1:-300} # Default 5 minutes
    local connections=${2:-100}
    
    local output_file="$RESULTS_DIR/stress_soak_${TIMESTAMP}.txt"
    
    print_info "Running sustained load test..."
    print_info "Duration: ${duration}s ($(($duration / 60)) minutes)"
    print_info "Connections: $connections"
    
    echo "Soak Test" > "$output_file"
    echo "=========" >> "$output_file"
    echo "Duration: ${duration}s" >> "$output_file"
    echo "Connections: $connections" >> "$output_file"
    echo "" >> "$output_file"
    
    if command -v hey &> /dev/null; then
        hey -z "${duration}s" -c "$connections" "$SERVER_URL/v1/pins?domain=$DOMAIN" >> "$output_file" 2>&1
    else
        print_error "hey not found"
        return 1
    fi
    
    if check_server; then
        print_success "Server stable after soak test"
    else
        print_error "Server became unstable during soak test"
    fi
    
    print_success "Soak test completed: $output_file"
}

# Concurrency test - multiple domains simultaneously
run_concurrency_test() {
    print_header "Concurrency Test - Multiple Domains"
    
    local output_file="$RESULTS_DIR/stress_concurrency_${TIMESTAMP}.txt"
    
    echo "Concurrency Stress Test" > "$output_file"
    echo "=======================" >> "$output_file"
    echo "" >> "$output_file"
    
    local domains=("google.com" "github.com" "cloudflare.com")
    local pids=()
    
    print_info "Launching concurrent requests to multiple domains..."
    
    for domain in "${domains[@]}"; do
        print_info "Starting load test for $domain"
        (
            hey -z 30s -c 100 "$SERVER_URL/v1/pins?domain=$domain" > "${output_file}.${domain}.tmp" 2>&1
        ) &
        pids+=($!)
    done
    
    # Wait for all background jobs
    print_info "Waiting for all tests to complete..."
    for pid in "${pids[@]}"; do
        wait "$pid"
    done
    
    # Combine results
    for domain in "${domains[@]}"; do
        echo "=== Domain: $domain ===" >> "$output_file"
        cat "${output_file}.${domain}.tmp" >> "$output_file"
        echo "" >> "$output_file"
        rm "${output_file}.${domain}.tmp"
    done
    
    print_success "Concurrency test completed: $output_file"
}

# Memory leak test - repeated requests to detect leaks
run_memory_leak_test() {
    print_header "Memory Leak Test - Repeated Requests"
    
    local output_file="$RESULTS_DIR/stress_memory_${TIMESTAMP}.txt"
    
    echo "Memory Leak Test" > "$output_file"
    echo "================" >> "$output_file"
    echo "" >> "$output_file"
    
    local iterations=10
    local requests_per_iteration=10000
    
    print_info "Running $iterations iterations of $requests_per_iteration requests each"
    print_warning "This may take several minutes..."
    
    for i in $(seq 1 $iterations); do
        print_info "Iteration $i/$iterations"
        
        echo "=== Iteration $i ===" >> "$output_file"
        
        if command -v hey &> /dev/null; then
            hey -n "$requests_per_iteration" -c 100 "$SERVER_URL/v1/pins?domain=$DOMAIN" >> "$output_file" 2>&1
        else
            print_error "hey not found"
            return 1
        fi
        
        # Check if server is still responsive
        if ! check_server; then
            print_error "Server became unresponsive at iteration $i"
            echo "ERROR: Server unresponsive at iteration $i" >> "$output_file"
            break
        fi
        
        echo "" >> "$output_file"
        sleep 2
    done
    
    print_success "Memory leak test completed: $output_file"
}

# Timeout test - verify timeout handling
run_timeout_test() {
    print_header "Timeout Test - Connection Handling"
    
    local output_file="$RESULTS_DIR/stress_timeout_${TIMESTAMP}.txt"
    
    echo "Timeout Stress Test" > "$output_file"
    echo "===================" >> "$output_file"
    echo "" >> "$output_file"
    
    print_info "Testing connection timeouts with slow clients..."
    
    # Test with very low timeout
    for delay in 0.1 0.5 1 2 5; do
        print_info "Testing with ${delay}s delay between requests"
        
        echo "=== Delay: ${delay}s ===" >> "$output_file"
        
        for i in {1..10}; do
            curl -w "Time: %{time_total}s, Status: %{http_code}\n" \
                -o /dev/null -s \
                "$SERVER_URL/v1/pins?domain=$DOMAIN" >> "$output_file" 2>&1
            
            sleep "$delay"
        done
        
        echo "" >> "$output_file"
    done
    
    print_success "Timeout test completed: $output_file"
}

# Resource exhaustion test
run_resource_exhaustion_test() {
    print_header "Resource Exhaustion Test"
    
    local output_file="$RESULTS_DIR/stress_resource_${TIMESTAMP}.txt"
    
    echo "Resource Exhaustion Test" > "$output_file"
    echo "========================" >> "$output_file"
    echo "" >> "$output_file"
    
    print_warning "This test attempts to exhaust server resources"
    print_info "Testing with extremely high connection count..."
    
    local max_connections=10000
    local duration=30
    
    echo "=== Extreme Load: $max_connections connections ===" >> "$output_file"
    
    if command -v hey &> /dev/null; then
        hey -z "${duration}s" -c "$max_connections" "$SERVER_URL/v1/pins?domain=$DOMAIN" >> "$output_file" 2>&1
    else
        print_error "hey not found"
        return 1
    fi
    
    # Check server recovery
    print_info "Checking server recovery..."
    sleep 10
    
    if check_server; then
        print_success "Server recovered successfully"
        echo "Result: Server recovered" >> "$output_file"
    else
        print_error "Server failed to recover"
        echo "Result: Server failed to recover" >> "$output_file"
    fi
    
    print_success "Resource exhaustion test completed: $output_file"
}

show_usage() {
    cat << EOF
Usage: $0 [TEST_TYPE]

Stress Testing Script for Dynapins Server

TEST_TYPES:
    progressive     - Progressive load increase to find breaking point
    spike           - Sudden traffic spike test
    soak [duration] [connections] - Sustained load test (default: 300s, 100 conns)
    concurrency     - Multiple domains simultaneously
    memory          - Memory leak detection
    timeout         - Timeout handling test
    resource        - Resource exhaustion test
    all             - Run all stress tests (WARNING: May take 30+ minutes)

EXAMPLES:
    $0 progressive
    $0 soak 600 200
    $0 all

ENVIRONMENT VARIABLES:
    SERVER_URL      Server URL (default: http://localhost:8080)
    TEST_DOMAIN     Domain to test (default: google.com)

EOF
}

# Main execution
TEST_TYPE=${1:-progressive}

case $TEST_TYPE in
    progressive)
        run_progressive_load
        ;;
    spike)
        run_spike_test
        ;;
    soak)
        duration=${2:-300}
        connections=${3:-100}
        run_soak_test "$duration" "$connections"
        ;;
    concurrency)
        run_concurrency_test
        ;;
    memory)
        run_memory_leak_test
        ;;
    timeout)
        run_timeout_test
        ;;
    resource)
        run_resource_exhaustion_test
        ;;
    all)
        print_warning "Running all stress tests - this may take 30+ minutes"
        echo ""
        run_progressive_load
        echo ""
        run_spike_test
        echo ""
        run_concurrency_test
        echo ""
        run_timeout_test
        echo ""
        run_soak_test 180 100
        echo ""
        run_memory_leak_test
        echo ""
        run_resource_exhaustion_test
        ;;
    -h|--help)
        show_usage
        exit 0
        ;;
    *)
        print_error "Unknown test type: $TEST_TYPE"
        show_usage
        exit 1
        ;;
esac

print_success "Stress testing completed!"
print_info "Results saved to: $RESULTS_DIR"

