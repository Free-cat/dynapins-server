#!/bin/bash

# Vegeta-specific load testing script
# Provides advanced load testing scenarios with Vegeta

set -e

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

RESULTS_DIR="./performance/results"
mkdir -p "$RESULTS_DIR"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)

print_header() {
    echo -e "${BLUE}========================================${NC}"
    echo -e "${BLUE}$1${NC}"
    echo -e "${BLUE}========================================${NC}"
}

print_info() {
    echo -e "${YELLOW}ℹ $1${NC}"
}

print_success() {
    echo -e "${GREEN}✓ $1${NC}"
}

# Check if vegeta is installed
if ! command -v vegeta &> /dev/null; then
    echo "Error: vegeta is not installed"
    echo "Install with: go install github.com/tsenart/vegeta@latest"
    exit 1
fi

# Constant rate test
run_constant_rate() {
    local rate=$1
    local duration=$2
    
    print_header "Constant Rate Test: ${rate} req/s for ${duration}"
    
    local output_bin="$RESULTS_DIR/vegeta_constant_${rate}rps_${TIMESTAMP}.bin"
    local output_txt="$RESULTS_DIR/vegeta_constant_${rate}rps_${TIMESTAMP}.txt"
    
    echo "GET http://localhost:8080/v1/pins?domain=google.com" | \
        vegeta attack -duration="${duration}" -rate="${rate}" -output="$output_bin"
    
    vegeta report "$output_bin" | tee "$output_txt"
    
    print_success "Results saved to: $output_txt"
}

# Ramping rate test
run_ramping_test() {
    print_header "Ramping Load Test"
    
    print_info "Phase 1: Warm-up (10 req/s for 10s)"
    echo "GET http://localhost:8080/v1/pins?domain=google.com" | \
        vegeta attack -duration=10s -rate=10 -output="$RESULTS_DIR/ramp_phase1_${TIMESTAMP}.bin"
    
    print_info "Phase 2: Normal load (50 req/s for 20s)"
    echo "GET http://localhost:8080/v1/pins?domain=google.com" | \
        vegeta attack -duration=20s -rate=50 -output="$RESULTS_DIR/ramp_phase2_${TIMESTAMP}.bin"
    
    print_info "Phase 3: High load (100 req/s for 20s)"
    echo "GET http://localhost:8080/v1/pins?domain=google.com" | \
        vegeta attack -duration=20s -rate=100 -output="$RESULTS_DIR/ramp_phase3_${TIMESTAMP}.bin"
    
    print_info "Phase 4: Peak load (200 req/s for 20s)"
    echo "GET http://localhost:8080/v1/pins?domain=google.com" | \
        vegeta attack -duration=20s -rate=200 -output="$RESULTS_DIR/ramp_phase4_${TIMESTAMP}.bin"
    
    print_info "Phase 5: Cool-down (25 req/s for 10s)"
    echo "GET http://localhost:8080/v1/pins?domain=google.com" | \
        vegeta attack -duration=10s -rate=25 -output="$RESULTS_DIR/ramp_phase5_${TIMESTAMP}.bin"
    
    # Generate reports
    local output_txt="$RESULTS_DIR/vegeta_ramping_${TIMESTAMP}.txt"
    echo "Ramping Load Test Results" > "$output_txt"
    echo "=========================" >> "$output_txt"
    
    for phase in 1 2 3 4 5; do
        echo "" >> "$output_txt"
        echo "Phase $phase:" >> "$output_txt"
        vegeta report "$RESULTS_DIR/ramp_phase${phase}_${TIMESTAMP}.bin" >> "$output_txt"
    done
    
    print_success "Ramping test results saved to: $output_txt"
}

# Multi-endpoint test
run_multi_endpoint() {
    print_header "Multi-Endpoint Test"
    
    local output_bin="$RESULTS_DIR/vegeta_multi_${TIMESTAMP}.bin"
    local output_txt="$RESULTS_DIR/vegeta_multi_${TIMESTAMP}.txt"
    
    print_info "Testing multiple endpoints simultaneously..."
    
    vegeta attack -duration=30s -rate=100 -targets=./performance/vegeta-targets.txt -output="$output_bin"
    
    vegeta report "$output_bin" | tee "$output_txt"
    
    print_success "Multi-endpoint results saved to: $output_txt"
}

# Burst test
run_burst_test() {
    print_header "Burst Test"
    
    print_info "Simulating traffic bursts..."
    
    local output_txt="$RESULTS_DIR/vegeta_burst_${TIMESTAMP}.txt"
    echo "Burst Test Results" > "$output_txt"
    echo "==================" >> "$output_txt"
    
    for burst in 1 2 3 4 5; do
        print_info "Burst $burst: 500 req/s for 5s"
        
        local output_bin="$RESULTS_DIR/burst_${burst}_${TIMESTAMP}.bin"
        echo "GET http://localhost:8080/v1/pins?domain=google.com" | \
            vegeta attack -duration=5s -rate=500 -output="$output_bin"
        
        echo "" >> "$output_txt"
        echo "Burst $burst:" >> "$output_txt"
        vegeta report "$output_bin" >> "$output_txt"
        
        # Cool-down period
        print_info "Cool-down period (5s)"
        sleep 5
    done
    
    print_success "Burst test results saved to: $output_txt"
}

# Sustained load test
run_sustained_test() {
    local rate=$1
    local duration=${2:-300s}
    
    print_header "Sustained Load Test: ${rate} req/s for ${duration}"
    
    local output_bin="$RESULTS_DIR/vegeta_sustained_${rate}rps_${TIMESTAMP}.bin"
    local output_txt="$RESULTS_DIR/vegeta_sustained_${rate}rps_${TIMESTAMP}.txt"
    
    echo "GET http://localhost:8080/v1/pins?domain=google.com" | \
        vegeta attack -duration="${duration}" -rate="${rate}" -output="$output_bin"
    
    vegeta report "$output_bin" | tee "$output_txt"
    
    # Generate histogram if gnuplot is available
    if command -v gnuplot &> /dev/null; then
        vegeta plot "$output_bin" > "$RESULTS_DIR/vegeta_sustained_${rate}rps_${TIMESTAMP}.html"
        print_success "HTML plot saved to: $RESULTS_DIR/vegeta_sustained_${rate}rps_${TIMESTAMP}.html"
    fi
    
    print_success "Sustained test results saved to: $output_txt"
}

# Show usage
show_usage() {
    cat << EOF
Usage: $0 [TEST_TYPE]

Vegeta-specific load testing for Dynapins Server

TEST_TYPES:
    constant [rate] [duration]  - Constant rate test (default: 100 req/s, 30s)
    ramping                     - Progressive load ramp-up and cool-down
    multi                       - Test multiple endpoints
    burst                       - Burst traffic simulation
    sustained [rate] [duration] - Long-running sustained load (default: 50 req/s, 300s)
    all                         - Run all test scenarios

EXAMPLES:
    $0 constant 100 60s      # 100 req/s for 60 seconds
    $0 ramping               # Progressive load test
    $0 sustained 50 600s     # 50 req/s for 10 minutes
    $0 all                   # Run all scenarios

EOF
}

# Main execution
case ${1:-constant} in
    constant)
        rate=${2:-100}
        duration=${3:-30s}
        run_constant_rate "$rate" "$duration"
        ;;
    ramping)
        run_ramping_test
        ;;
    multi)
        run_multi_endpoint
        ;;
    burst)
        run_burst_test
        ;;
    sustained)
        rate=${2:-50}
        duration=${3:-300s}
        run_sustained_test "$rate" "$duration"
        ;;
    all)
        run_constant_rate 100 30s
        echo ""
        run_multi_endpoint
        echo ""
        run_burst_test
        echo ""
        run_ramping_test
        ;;
    -h|--help)
        show_usage
        exit 0
        ;;
    *)
        echo "Unknown test type: $1"
        show_usage
        exit 1
        ;;
esac

print_success "Test completed!"

