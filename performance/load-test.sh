#!/bin/bash

# Load Testing Script for Dynapins Server
# This script provides comprehensive load testing using multiple tools

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Default configuration (environment overrides respected)
SERVER_URL="${SERVER_URL:-http://localhost:8080}"
DOMAIN="${TEST_DOMAIN:-example.com}"
DURATION="${DURATION:-30s}"
CONNECTIONS="${CONNECTIONS:-100}"
RATE="${RATE:-0}"   # 0 means unlimited (vegeta)
TOOL="${TOOL:-auto}" # auto (default), hey, wrk, vegeta, ab

# Results directory
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

print_info() {
    echo -e "${YELLOW}ℹ $1${NC}"
}

check_server() {
print_info "Checking server availability at $SERVER_URL..."
if curl -sf "$SERVER_URL/health" > /dev/null; then
        print_success "Server is running"
        return 0
    else
        print_error "Server is not responding"
        return 1
    fi
}

check_tool() {
    if command -v "$1" &> /dev/null; then
        return 0
    else
        return 1
    fi
}

detect_tool() {
    case "$TOOL" in
        hey|wrk|vegeta|ab)
            if ! check_tool "$TOOL"; then
                print_error "Tool '$TOOL' not found in PATH"
                exit 1
            fi
            ;;
        auto)
            for candidate in hey wrk vegeta ab; do
                if check_tool "$candidate"; then
                    TOOL="$candidate"
                    break
                fi
            done
            if [ "$TOOL" = "auto" ]; then
                print_error "No supported load testing tool found. Install hey (recommended)"
                exit 1
            fi
            ;;
        *)
            print_error "Unsupported tool: $TOOL"
            exit 1
            ;;
    esac

    print_info "Using load testing tool: $TOOL"
}

summarize_results() {
    local tool="$1"
    local file="$2"
    local summary_file="$3"

    echo "Summary:" > "$summary_file"
    case "$tool" in
        hey)
            local rps latency p95 p99 success non2xx
            rps=$(awk '/Requests\/sec/ {print $2}' "$file")
            latency=$(awk '/Latency:/ {print $2" "$3; exit}' "$file")
            p95=$(awk '/95%/ {print $2" "$3}' "$file")
            p99=$(awk '/99%/ {print $2" "$3}' "$file")
            non2xx=$(awk -F': ' '/Non-2xx or 3xx responses/ {print $2}' "$file")
            success=$(awk '/\[200\]/ {print $2" responses"}' "$file")

            echo "  Requests/sec: ${rps:-N/A}" >> "$summary_file"
            echo "  Avg latency: ${latency:-N/A}" >> "$summary_file"
            echo "  p95 latency: ${p95:-N/A}" >> "$summary_file"
            echo "  p99 latency: ${p99:-N/A}" >> "$summary_file"
            echo "  Success: ${success:-N/A}" >> "$summary_file"
            if [ -n "$non2xx" ] && [ "$non2xx" != "0" ]; then
                echo "  Non-2xx responses: $non2xx" >> "$summary_file"
            fi
            ;;
        wrk)
            local rps latency stdev p95 transfer
            rps=$(awk '/Requests\/sec/ {print $2}' "$file")
            latency=$(awk '/Latency/ {print $2" "$3; exit}' "$file")
            stdev=$(awk '/Latency/ {print $4" "$5; exit}' "$file")
            p95=$(awk '/95%/ {print $2" "$3}' "$file")
            transfer=$(awk '/Transfer\/sec/ {print $2" "$3}' "$file")

            echo "  Requests/sec: ${rps:-N/A}" >> "$summary_file"
            echo "  Avg latency: ${latency:-N/A}" >> "$summary_file"
            echo "  Stdev latency: ${stdev:-N/A}" >> "$summary_file"
            echo "  p95 latency: ${p95:-N/A}" >> "$summary_file"
            echo "  Transfer/sec: ${transfer:-N/A}" >> "$summary_file"
            ;;
        vegeta)
            local throughput avg p95
            throughput=$(awk '/throughput/ {print $3}' "$file")
            avg=$(awk '/latencies/ {getline; print $2" "$3}' "$file")
            p95=$(awk '/95th/ {print $2" "$3}' "$file")
            echo "  Throughput: ${throughput:-N/A}" >> "$summary_file"
            echo "  Avg latency: ${avg:-N/A}" >> "$summary_file"
            echo "  p95 latency: ${p95:-N/A}" >> "$summary_file"
            ;;
        ab)
            local rps time_per_request failed
            rps=$(awk '/Requests per second/ {print $4" "$5}' "$file")
            time_per_request=$(awk '/Time per request:/ {print $4" "$5; exit}' "$file")
            failed=$(awk '/Failed requests/ {print $3}' "$file")
            echo "  Requests/sec: ${rps:-N/A}" >> "$summary_file"
            echo "  Time per request: ${time_per_request:-N/A}" >> "$summary_file"
            echo "  Failed requests: ${failed:-N/A}" >> "$summary_file"
            ;;
    esac

    echo "  Raw output: $file" >> "$summary_file"
}

run_hey() {
    print_header "Load Testing with 'hey'"
    
    local output_file="$RESULTS_DIR/hey_${TIMESTAMP}.txt"
    local endpoint="$SERVER_URL/v1/pins?domain=$DOMAIN"
    
    print_info "Running hey test..."
    print_info "  URL: $endpoint"
    print_info "  Duration: $DURATION"
    print_info "  Connections: $CONNECTIONS"
    
    hey -z "$DURATION" -c "$CONNECTIONS" "$endpoint" | tee "$output_file"

    local summary_file="$RESULTS_DIR/hey_summary_${TIMESTAMP}.txt"
    summarize_results "hey" "$output_file" "$summary_file"
    print_success "Summary saved to: $summary_file"
    print_success "Raw results saved to: $output_file"
    print_info "--- Summary ---"
    cat "$summary_file"
}

run_wrk() {
    print_header "Load Testing with 'wrk'"
    
    local output_file="$RESULTS_DIR/wrk_${TIMESTAMP}.txt"
    local endpoint="$SERVER_URL/v1/pins?domain=$DOMAIN"
    
    print_info "Running wrk test..."
    print_info "  URL: $endpoint"
    print_info "  Duration: $DURATION"
    print_info "  Connections: $CONNECTIONS"
    print_info "  Threads: 4"
    
    wrk -t4 -c "$CONNECTIONS" -d "$DURATION" "$endpoint" | tee "$output_file"

    local summary_file="$RESULTS_DIR/wrk_summary_${TIMESTAMP}.txt"
    summarize_results "wrk" "$output_file" "$summary_file"
    print_success "Summary saved to: $summary_file"
    print_success "Raw results saved to: $output_file"
    print_info "--- Summary ---"
    cat "$summary_file"
}

run_vegeta() {
    print_header "Load Testing with 'vegeta'"
    
    local output_file="$RESULTS_DIR/vegeta_${TIMESTAMP}.txt"
    local bin_file="$RESULTS_DIR/vegeta_${TIMESTAMP}.bin"
    local endpoint="$SERVER_URL/v1/pins?domain=$DOMAIN"
    
    print_info "Running vegeta test..."
    print_info "  URL: $endpoint"
    print_info "  Duration: $DURATION"
    print_info "  Rate: ${RATE}/s (0 = unlimited)"
    
    # Create target file
    echo "GET $endpoint" | vegeta attack \
        -duration="$DURATION" \
        -rate="$RATE" \
        -output="$bin_file"
    
    # Generate report
    vegeta report "$bin_file" | tee "$output_file"
    
    # Generate plot (optional)
    if check_tool "gnuplot"; then
        vegeta plot "$bin_file" > "$RESULTS_DIR/vegeta_${TIMESTAMP}.html"
        print_success "HTML plot saved to: $RESULTS_DIR/vegeta_${TIMESTAMP}.html"
    fi
    
    local summary_file="$RESULTS_DIR/vegeta_summary_${TIMESTAMP}.txt"
    summarize_results "vegeta" "$output_file" "$summary_file"
    print_success "Summary saved to: $summary_file"
    print_success "Raw results saved to: $output_file"
    print_info "--- Summary ---"
    cat "$summary_file"
}

run_ab() {
    print_header "Load Testing with 'ab' (Apache Bench)"
    
    local output_file="$RESULTS_DIR/ab_${TIMESTAMP}.txt"
    local endpoint="$SERVER_URL/v1/pins?domain=$DOMAIN"
    
    # Convert duration to number of requests (approximation)
    local duration_seconds=$(echo "$DURATION" | sed 's/s$//')
    local total_requests=$((duration_seconds * 100))
    
    print_info "Running ab test..."
    print_info "  URL: $endpoint"
    print_info "  Requests: $total_requests"
    print_info "  Concurrency: $CONNECTIONS"
    
    ab -n "$total_requests" -c "$CONNECTIONS" "$endpoint" | tee "$output_file"

    local summary_file="$RESULTS_DIR/ab_summary_${TIMESTAMP}.txt"
    summarize_results "ab" "$output_file" "$summary_file"
    print_success "Summary saved to: $summary_file"
    print_success "Raw results saved to: $output_file"
    print_info "--- Summary ---"
    cat "$summary_file"
}

run_baseline_test() {
    print_header "Baseline Performance Test"

    local output_file="$RESULTS_DIR/baseline_${TIMESTAMP}.txt"

    cat <<EOF | tee "$output_file"
Endpoint Performance Summary
============================
Server: $SERVER_URL
Domain: $DOMAIN

EOF

    run_baseline_endpoint() {
        local name="$1"
        local url="$2"
        local result
        result=$(curl -w "dns=%{time_namelookup}s connect=%{time_connect}s tls=%{time_appconnect}s ttfb=%{time_starttransfer}s total=%{time_total}s" -o /dev/null -s "$url")
        echo "$name: $url" | tee -a "$output_file"
        echo "  $result" | tee -a "$output_file"
        echo "" | tee -a "$output_file"
    }

    run_baseline_endpoint "Health" "$SERVER_URL/health"
    run_baseline_endpoint "Readiness" "$SERVER_URL/readiness"
    run_baseline_endpoint "Pins" "$SERVER_URL/v1/pins?domain=$DOMAIN"

    print_success "Baseline summary saved to: $output_file"
}

run_stress_test() {
    print_header "Stress Test - Progressive Load"
    
    local output_file="$RESULTS_DIR/stress_${TIMESTAMP}.txt"
    
    echo "Progressive Load Test Results" > "$output_file"
    echo "=============================" >> "$output_file"
    echo "" >> "$output_file"
    
    for connections in 10 50 100 250 500 1000; do
        print_info "Testing with $connections concurrent connections..."
        echo "Connections: $connections" >> "$output_file"
        
        if check_tool "hey"; then
            hey -z 10s -c "$connections" "$SERVER_URL/v1/pins?domain=$DOMAIN" >> "$output_file" 2>&1
        else
            print_error "hey not found, skipping stress test"
            return 1
        fi
        
        echo "" >> "$output_file"
        echo "---" >> "$output_file"
        echo "" >> "$output_file"
        
        sleep 2
    done
    
    print_success "Stress test results saved to: $output_file"
}

run_latency_test() {
    print_header "Latency Distribution Test"
    
    local output_file="$RESULTS_DIR/latency_${TIMESTAMP}.txt"
    
    print_info "Running 1000 sequential requests to measure latency distribution..."
    
    echo "Latency Distribution Test" > "$output_file"
    echo "=========================" >> "$output_file"
    echo "" >> "$output_file"
    
    for i in {1..1000}; do
        curl -w "%{time_total}\n" -o /dev/null -s "$SERVER_URL/v1/pins?domain=$DOMAIN" >> "$output_file"
        
        if [ $((i % 100)) -eq 0 ]; then
            print_info "Completed $i/1000 requests"
        fi
    done
    
    # Calculate statistics if bc is available
    if check_tool "bc"; then
        local avg=$(awk '{ total += $1; count++ } END { print total/count }' "$output_file")
        local min=$(sort -n "$output_file" | head -1)
        local max=$(sort -n "$output_file" | tail -1)
        
        echo "" >> "$output_file"
        echo "Statistics:" >> "$output_file"
        echo "  Min: ${min}s" >> "$output_file"
        echo "  Max: ${max}s" >> "$output_file"
        echo "  Avg: ${avg}s" >> "$output_file"
        
        print_info "Latency Stats: Min=${min}s, Max=${max}s, Avg=${avg}s"
    fi
    
    print_success "Latency test results saved to: $output_file"
}

show_usage() {
    cat << EOF
Usage: $0 [OPTIONS] [TEST_TYPE]

Load Testing Script for Dynapins Server

TEST_TYPE:
    baseline    - Run baseline performance test (single requests)
    load        - Run load test (default)
    stress      - Run progressive stress test
    latency     - Run latency distribution test
    all         - Run all test types

OPTIONS:
    -u, --url URL           Server URL (default: http://localhost:8080)
    -d, --domain DOMAIN     Domain to test (default: google.com)
    -t, --duration TIME     Test duration (default: 30s)
    -c, --connections NUM   Number of concurrent connections (default: 100)
    -r, --rate NUM          Request rate per second (default: 0 = unlimited)
    -T, --tool TOOL         Load testing tool: auto, hey, wrk, vegeta, ab (default: auto)
    -h, --help              Show this help message

EXAMPLES:
    # Run default load test
    $0

    # Run load test with custom configuration
    $0 -u http://localhost:8080 -d google.com -t 60s -c 200

    # Run all tests
    $0 all

    # Run stress test
    $0 stress

    # Use specific tool
    $0 -T hey load

ENVIRONMENT VARIABLES:
    SERVER_URL      Server URL
    TEST_DOMAIN     Domain to test
    DURATION        Test duration
    CONNECTIONS     Number of concurrent connections
    RATE            Request rate per second
    TOOL            Load testing tool to use

EOF
}

# Parse command line arguments
TEST_TYPE="load"

while [[ $# -gt 0 ]]; do
    case $1 in
        -u|--url)
            SERVER_URL="$2"
            shift 2
            ;;
        -d|--domain)
            DOMAIN="$2"
            shift 2
            ;;
        -t|--duration)
            DURATION="$2"
            shift 2
            ;;
        -c|--connections)
            CONNECTIONS="$2"
            shift 2
            ;;
        -r|--rate)
            RATE="$2"
            shift 2
            ;;
        -T|--tool)
            TOOL="$2"
            shift 2
            ;;
        -h|--help)
            show_usage
            exit 0
            ;;
        baseline|load|stress|latency|all)
            TEST_TYPE="$1"
            shift
            ;;
        *)
            print_error "Unknown option: $1"
            show_usage
            exit 1
            ;;
    esac
done

# Main execution
print_header "Dynapins Server Load Testing"
echo "Configuration:"
echo "  Server: $SERVER_URL"
echo "  Domain: $DOMAIN"
echo "  Duration: $DURATION"
echo "  Connections: $CONNECTIONS"
echo "  Rate: $RATE/s"
echo "  Test Type: $TEST_TYPE"
echo ""

# Check server availability
if ! check_server; then
    print_error "Server is not available. Please start the server first."
    exit 1
fi

# Run tests based on type
case $TEST_TYPE in
    baseline)
        run_baseline_test
        ;;
    load)
        detect_tool
        case $TOOL in
            hey)
                run_hey
                ;;
            wrk)
                run_wrk
                ;;
            vegeta)
                run_vegeta
                ;;
            ab)
                run_ab
                ;;
        esac
        ;;
    stress)
        run_stress_test
        ;;
    latency)
        run_latency_test
        ;;
    all)
        run_baseline_test
        echo ""
        run_latency_test
        echo ""
        run_stress_test
        echo ""
        detect_tool
        case $TOOL in
            hey)
                run_hey
                ;;
            wrk)
                run_wrk
                ;;
            vegeta)
                run_vegeta
                ;;
            ab)
                run_ab
                ;;
        esac
        ;;
esac

print_success "All tests completed successfully!"
print_info "Results directory: $RESULTS_DIR"

