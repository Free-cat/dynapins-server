.PHONY: help build test test-coverage run clean fmt vet lint docker-build docker-run
.PHONY: bench bench-crypto bench-server bench-all bench-compare
.PHONY: load-test stress-test perf-test perf-clean

# Configuration
IMAGE_NAME = dynapins-server
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
PERF_RESULTS_DIR = ./performance/results

# Default target
help:
	@echo "Available targets:"
	@echo ""
	@echo "Build & Run:"
	@echo "  make build          - Build the server binary"
	@echo "  make run            - Run the server locally"
	@echo "  make clean          - Remove build artifacts"
	@echo ""
	@echo "Testing:"
	@echo "  make test           - Run all tests"
	@echo "  make test-coverage  - Run tests with coverage report"
	@echo ""
	@echo "Performance Testing:"
	@echo "  make bench          - Run all benchmarks"
	@echo "  make bench-crypto   - Run crypto benchmarks only"
	@echo "  make bench-server   - Run server benchmarks only"
	@echo "  make bench-compare  - Compare benchmark results"
	@echo "  make load-test      - Run HTTP load tests"
	@echo "  make stress-test    - Run stress tests"
	@echo "  make perf-test      - Run complete performance test suite"
	@echo "  make perf-clean     - Clean performance test results"
	@echo ""
	@echo "Code Quality:"
	@echo "  make fmt            - Format code with go fmt"
	@echo "  make vet            - Run go vet"
	@echo "  make lint           - Run all code quality checks"
	@echo ""
	@echo "Docker:"
	@echo "  make docker-build   - Build Docker image"
	@echo "  make docker-run     - Run Docker container"

# Build the server
build:
	@echo "Building server..."
	@go build -o bin/server ./cmd/server
	@echo "✓ Build complete: bin/server"

# Run tests
test:
	@echo "Running tests..."
	@go test ./... -v

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	@go test ./... -cover -coverprofile=coverage.out
	@go tool cover -html=coverage.out -o coverage.html
	@echo "✓ Coverage report: coverage.html"

# Run the server locally
run:
	@echo "Starting server..."
	@go run ./cmd/server

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf bin/
	@rm -f coverage.out coverage.html
	@rm -f server
	@echo "✓ Clean complete"

# Format code
fmt:
	@echo "Formatting code..."
	@go fmt ./...
	@echo "✓ Format complete"

# Run go vet
vet:
	@echo "Running go vet..."
	@go vet ./...
	@echo "✓ Vet complete"

# Run all linters
lint: fmt vet
	@echo "✓ All linters passed"

# Build Docker image
docker-build:
	@echo "Building Docker image..."
	@docker build -t $(IMAGE_NAME):$(VERSION) .
	@docker tag $(IMAGE_NAME):$(VERSION) $(IMAGE_NAME):latest
	@echo "✓ Built: $(IMAGE_NAME):$(VERSION)"

# Run Docker container
docker-run:
	@echo "Running Docker container..."
	@docker run -p 8080:8080 \
		-e ALLOWED_DOMAINS="example.com,*.example.com" \
		-e PRIVATE_KEY_PEM="$${PRIVATE_KEY_PEM}" \
		$(IMAGE_NAME):latest

# ========================================
# Performance Testing Targets
# ========================================

# Run all benchmarks
bench:
	@echo "Running all benchmarks..."
	@go test -bench=. -benchmem -run=^$$ ./...
	@echo "✓ Benchmarks complete"

# Run crypto benchmarks only
bench-crypto:
	@echo "Running crypto benchmarks..."
	@go test -bench=. -benchmem -run=^$$ ./internal/crypto/
	@echo "✓ Crypto benchmarks complete"

# Run server benchmarks only
bench-server:
	@echo "Running server benchmarks..."
	@go test -bench=. -benchmem -run=^$$ ./internal/server/
	@echo "✓ Server benchmarks complete"

# Run domain validator benchmarks
bench-domain:
	@echo "Running domain validator benchmarks..."
	@go test -bench=. -benchmem -run=^$$ ./internal/domain/
	@echo "✓ Domain benchmarks complete"

# Run certificate retrieval benchmarks
bench-cert:
	@echo "Running certificate retrieval benchmarks..."
	@go test -bench=. -benchmem -run=^$$ ./internal/cert/
	@echo "✓ Certificate benchmarks complete"

# Run all benchmarks and save results
bench-all:
	@echo "Running comprehensive benchmarks..."
	@mkdir -p $(PERF_RESULTS_DIR)
	@go test -bench=. -benchmem -run=^$$ ./... | tee $(PERF_RESULTS_DIR)/bench_$$(date +%Y%m%d_%H%M%S).txt
	@echo "✓ Results saved to $(PERF_RESULTS_DIR)"

# Save baseline benchmark for comparison
bench-baseline:
	@echo "Saving baseline benchmarks..."
	@mkdir -p $(PERF_RESULTS_DIR)
	@go test -bench=. -benchmem -run=^$$ ./... > $(PERF_RESULTS_DIR)/bench_baseline.txt
	@echo "✓ Baseline saved to $(PERF_RESULTS_DIR)/bench_baseline.txt"

# Compare with baseline (requires benchstat: go install golang.org/x/perf/cmd/benchstat@latest)
bench-compare:
	@echo "Comparing with baseline..."
	@if [ ! -f $(PERF_RESULTS_DIR)/bench_baseline.txt ]; then \
		echo "Error: No baseline found. Run 'make bench-baseline' first"; \
		exit 1; \
	fi
	@go test -bench=. -benchmem -run=^$$ ./... > $(PERF_RESULTS_DIR)/bench_new.txt
	@if command -v benchstat >/dev/null 2>&1; then \
		benchstat $(PERF_RESULTS_DIR)/bench_baseline.txt $(PERF_RESULTS_DIR)/bench_new.txt; \
	else \
		echo "benchstat not installed. Install with: go install golang.org/x/perf/cmd/benchstat@latest"; \
		echo "Showing new results:"; \
		cat $(PERF_RESULTS_DIR)/bench_new.txt; \
	fi

# Run CPU profiling
bench-cpu:
	@echo "Running benchmarks with CPU profiling..."
	@mkdir -p $(PERF_RESULTS_DIR)
	@go test -bench=. -cpuprofile=$(PERF_RESULTS_DIR)/cpu.prof -run=^$$ ./internal/crypto/
	@echo "✓ CPU profile saved to $(PERF_RESULTS_DIR)/cpu.prof"
	@echo "View with: go tool pprof $(PERF_RESULTS_DIR)/cpu.prof"

# Run memory profiling
bench-mem:
	@echo "Running benchmarks with memory profiling..."
	@mkdir -p $(PERF_RESULTS_DIR)
	@go test -bench=. -memprofile=$(PERF_RESULTS_DIR)/mem.prof -run=^$$ ./internal/crypto/
	@echo "✓ Memory profile saved to $(PERF_RESULTS_DIR)/mem.prof"
	@echo "View with: go tool pprof $(PERF_RESULTS_DIR)/mem.prof"

# Load testing (requires hey, wrk, or vegeta)
load-test:
	@echo "Running load tests..."
	@if [ ! -x ./performance/load-test.sh ]; then \
		chmod +x ./performance/load-test.sh; \
	fi
	@./performance/load-test.sh baseline
	@echo "✓ Load test complete"

# Load test with custom parameters
load-test-custom:
	@echo "Running custom load test..."
	@if [ ! -x ./performance/load-test.sh ]; then \
		chmod +x ./performance/load-test.sh; \
	fi
	@./performance/load-test.sh -c 200 -t 60s load

# Stress testing
stress-test:
	@echo "Running stress tests..."
	@if [ ! -x ./performance/stress-test.sh ]; then \
		chmod +x ./performance/stress-test.sh; \
	fi
	@./performance/stress-test.sh progressive
	@echo "✓ Stress test complete"

# Stress test - all scenarios
stress-test-all:
	@echo "Running all stress test scenarios..."
	@echo "WARNING: This may take 30+ minutes"
	@if [ ! -x ./performance/stress-test.sh ]; then \
		chmod +x ./performance/stress-test.sh; \
	fi
	@./performance/stress-test.sh all

# Vegeta load tests
vegeta-test:
	@echo "Running Vegeta load tests..."
	@if ! command -v vegeta >/dev/null 2>&1; then \
		echo "Error: vegeta not installed"; \
		echo "Install with: go install github.com/tsenart/vegeta@latest"; \
		exit 1; \
	fi
	@if [ ! -x ./performance/vegeta-test.sh ]; then \
		chmod +x ./performance/vegeta-test.sh; \
	fi
	@./performance/vegeta-test.sh constant 100 30s

# Complete performance test suite
perf-test:
	@echo "=========================================="
	@echo "Running Complete Performance Test Suite"
	@echo "=========================================="
	@echo ""
	@echo "[1/4] Running benchmarks..."
	@$(MAKE) bench-all
	@echo ""
	@echo "[2/4] Running baseline load test..."
	@$(MAKE) load-test
	@echo ""
	@echo "[3/4] Running stress tests..."
	@$(MAKE) stress-test
	@echo ""
	@echo "[4/4] Generating summary..."
	@echo ""
	@echo "=========================================="
	@echo "Performance Test Suite Complete"
	@echo "=========================================="
	@echo "Results saved to: $(PERF_RESULTS_DIR)"
	@echo ""
	@echo "View results:"
	@echo "  Benchmarks:  $(PERF_RESULTS_DIR)/bench_*.txt"
	@echo "  Load Tests:  $(PERF_RESULTS_DIR)/*_load_*.txt"
	@echo "  Stress Tests: $(PERF_RESULTS_DIR)/stress_*.txt"

# Clean performance test results
perf-clean:
	@echo "Cleaning performance test results..."
	@rm -rf $(PERF_RESULTS_DIR)
	@echo "✓ Performance results cleaned"

# Setup performance testing tools (macOS)
perf-setup-mac:
	@echo "Installing performance testing tools..."
	@if ! command -v hey >/dev/null 2>&1; then \
		echo "Installing hey..."; \
		go install github.com/rakyll/hey@latest; \
	fi
	@if ! command -v vegeta >/dev/null 2>&1; then \
		echo "Installing vegeta..."; \
		go install github.com/tsenart/vegeta@latest; \
	fi
	@if ! command -v benchstat >/dev/null 2>&1; then \
		echo "Installing benchstat..."; \
		go install golang.org/x/perf/cmd/benchstat@latest; \
	fi
	@if ! command -v wrk >/dev/null 2>&1; then \
		echo "Installing wrk..."; \
		brew install wrk 2>/dev/null || echo "brew not available, skipping wrk"; \
	fi
	@echo "✓ Performance tools installed"
	@echo ""
	@echo "Available tools:"
	@command -v hey >/dev/null 2>&1 && echo "  ✓ hey" || echo "  ✗ hey"
	@command -v vegeta >/dev/null 2>&1 && echo "  ✓ vegeta" || echo "  ✗ vegeta"
	@command -v wrk >/dev/null 2>&1 && echo "  ✓ wrk" || echo "  ✗ wrk"
	@command -v benchstat >/dev/null 2>&1 && echo "  ✓ benchstat" || echo "  ✗ benchstat"

# Setup performance testing tools (Linux)
perf-setup-linux:
	@echo "Installing performance testing tools..."
	@if ! command -v hey >/dev/null 2>&1; then \
		echo "Installing hey..."; \
		go install github.com/rakyll/hey@latest; \
	fi
	@if ! command -v vegeta >/dev/null 2>&1; then \
		echo "Installing vegeta..."; \
		go install github.com/tsenart/vegeta@latest; \
	fi
	@if ! command -v benchstat >/dev/null 2>&1; then \
		echo "Installing benchstat..."; \
		go install golang.org/x/perf/cmd/benchstat@latest; \
	fi
	@echo "✓ Performance tools installed"

# Docker-based performance testing
perf-docker:
	@echo "Running performance tests in Docker..."
	@if [ -z "$$PRIVATE_KEY_PEM" ]; then \
		echo "Error: PRIVATE_KEY_PEM not set"; \
		exit 1; \
	fi
	@docker-compose -f performance/docker-compose.perf.yml up -d
	@echo "Waiting for server to be ready..."
	@sleep 5
	@docker-compose -f performance/docker-compose.perf.yml exec load-tester sh -c "apk add --no-cache curl && curl http://dynapins-server:8080/health"
	@echo "Server ready. Run tests with:"
	@echo "  docker-compose -f performance/docker-compose.perf.yml exec load-tester sh"
	@echo ""
	@echo "Cleanup with:"
	@echo "  docker-compose -f performance/docker-compose.perf.yml down"
