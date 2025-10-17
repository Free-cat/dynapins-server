# Dynapins Server Performance Testing

Comprehensive performance testing suite for the Dynapins Server, including benchmarks, load tests, and stress tests.

## 📋 Table of Contents

- [Overview](#overview)
- [Quick Start](#quick-start)
- [Testing Tools](#testing-tools)
- [Benchmark Tests](#benchmark-tests)
- [Load Testing](#load-testing)
- [Stress Testing](#stress-testing)
- [Docker Environment](#docker-environment)
- [Interpreting Results](#interpreting-results)
- [Performance Targets](#performance-targets)

## 🎯 Overview

This directory contains a complete performance testing suite:

- **Benchmark Tests**: Go benchmarks for individual components
- **Load Tests**: HTTP endpoint load testing with multiple tools
- **Stress Tests**: System limits and breaking point analysis
- **Docker Environment**: Isolated environment for consistent testing

## 🚀 Quick Start

### Prerequisites

Install at least one load testing tool:

```bash
# Option 1: hey (recommended)
go install github.com/rakyll/hey@latest

# Option 2: wrk (macOS)
brew install wrk

# Option 3: vegeta
go install github.com/tsenart/vegeta@latest

# Option 4: Apache Bench (usually pre-installed)
# ab is typically available on most systems
```

### Run Performance Tests

```bash
# Navigate to server directory
cd dynapins-server

# Run Go benchmarks
make bench

# Run load tests
make load-test

# Run stress tests
make stress-test

# Run all performance tests
make perf-test
```

## 🛠️ Testing Tools

### 1. Go Benchmarks

Built-in Go benchmark tests for component-level performance:

```bash
# Run all benchmarks
go test -bench=. ./...

# Run with memory profiling
go test -bench=. -benchmem ./...

# Run specific package
go test -bench=. ./internal/crypto/

# Save results for comparison
go test -bench=. ./... > bench-baseline.txt
```

### 2. hey

Fast and simple HTTP load generator:

```bash
# Basic load test
hey -z 30s -c 100 http://localhost:8080/v1/pins?domain=google.com

# With rate limiting
hey -z 30s -c 100 -q 10 http://localhost:8080/v1/pins?domain=google.com

# Save results
hey -z 30s -c 100 http://localhost:8080/v1/pins?domain=google.com > results.txt
```

### 3. wrk

Modern HTTP benchmarking tool:

```bash
# Basic test
wrk -t4 -c100 -d30s http://localhost:8080/v1/pins?domain=google.com

# With custom script
wrk -t4 -c100 -d30s -s script.lua http://localhost:8080/v1/pins
```

### 4. vegeta

Constant throughput load tester:

```bash
# Constant rate
echo "GET http://localhost:8080/v1/pins?domain=google.com" | vegeta attack -duration=30s -rate=100 | vegeta report

# With targets file
vegeta attack -duration=30s -rate=100 -targets=vegeta-targets.txt | vegeta report

# Generate plot
vegeta attack -duration=30s -rate=100 -targets=vegeta-targets.txt -output=results.bin
vegeta plot results.bin > results.html
```

## 📊 Benchmark Tests

### Component Benchmarks

#### Crypto Operations

```bash
# Run crypto benchmarks
go test -bench=. ./internal/crypto/

# Key benchmarks:
# - BenchmarkGenerateSPKIHashes: Certificate hash generation
# - BenchmarkCreateJWS: JWS token creation
# - BenchmarkVerifyJWS: JWS token verification
# - BenchmarkParallelJWSCreation: Parallel token creation
```

**Expected Performance:**
- SPKI hash generation: ~100-200 µs per certificate
- JWS creation: ~50-100 µs per token
- JWS verification: ~80-150 µs per token

#### HTTP Handlers

```bash
# Run handler benchmarks
go test -bench=. ./internal/server/

# Key benchmarks:
# - BenchmarkHandleGetPins: Full /v1/pins endpoint
# - BenchmarkHandleHealth: Health check endpoint
# - BenchmarkHandleReadiness: Readiness endpoint
# - BenchmarkHandleGetPinsParallel: Concurrent requests
```

**Expected Performance:**
- Health endpoint: <5 µs per request
- Readiness endpoint: <10 µs per request
- Pins endpoint: 50-200 ms per request (includes external cert fetch)

#### Domain Validation

```bash
# Run domain validation benchmarks
go test -bench=. ./internal/domain/

# Key benchmarks:
# - BenchmarkIsAllowed: Domain whitelist validation
# - BenchmarkIsAllowedParallel: Parallel validation
```

**Expected Performance:**
- Domain validation: <1 µs per check

#### Certificate Retrieval

```bash
# Run cert retrieval benchmarks
go test -bench=. ./internal/cert/

# Key benchmarks:
# - BenchmarkGetCertificates: TLS cert retrieval
# - BenchmarkGetCertificatesParallel: Parallel retrieval
```

**Expected Performance:**
- Certificate retrieval: 50-150 ms (network dependent)

### Comparing Benchmarks

```bash
# Save baseline
go test -bench=. ./... > bench-baseline.txt

# Make changes...

# Compare with new results
go test -bench=. ./... > bench-new.txt
benchstat bench-baseline.txt bench-new.txt
```

## 🔥 Load Testing

### Using the Load Test Script

```bash
./performance/load-test.sh [OPTIONS] [TEST_TYPE]
```

#### Test Types

**1. Baseline Test**
Single requests to measure base latency:

```bash
./performance/load-test.sh baseline
```

**2. Load Test**
Standard load testing with configurable parameters:

```bash
# Default: 100 connections, 30s
./performance/load-test.sh load

# Custom configuration
./performance/load-test.sh -c 200 -t 60s load

# Specific tool
./performance/load-test.sh -T hey load
```

**3. Latency Test**
Measures latency distribution:

```bash
./performance/load-test.sh latency
```

**4. All Tests**
Run complete test suite:

```bash
./performance/load-test.sh all
```

#### Configuration Options

```bash
-u, --url URL           Server URL (default: http://localhost:8080)
-d, --domain DOMAIN     Domain to test (default: google.com)
-t, --duration TIME     Test duration (default: 30s)
-c, --connections NUM   Concurrent connections (default: 100)
-r, --rate NUM          Request rate per second (default: 0 = unlimited)
-T, --tool TOOL         Tool: auto, hey, wrk, vegeta, ab (default: auto)
```

#### Examples

```bash
# High-load test
./performance/load-test.sh -c 500 -t 120s load

# Rate-limited test
./performance/load-test.sh -r 100 -t 60s load

# Test different domain
./performance/load-test.sh -d github.com load

# Full test suite with custom config
./performance/load-test.sh -c 200 -t 60s all
```

### Using Vegeta Tests

Advanced load testing scenarios:

```bash
./performance/vegeta-test.sh [TEST_TYPE]
```

#### Test Types

**1. Constant Rate**
```bash
# 100 req/s for 30 seconds
./performance/vegeta-test.sh constant 100 30s
```

**2. Ramping Load**
Progressive load increase:
```bash
./performance/vegeta-test.sh ramping
```

Phases:
- Warm-up: 10 req/s for 10s
- Normal: 50 req/s for 20s
- High: 100 req/s for 20s
- Peak: 200 req/s for 20s
- Cool-down: 25 req/s for 10s

**3. Multi-Endpoint**
Test multiple endpoints simultaneously:
```bash
./performance/vegeta-test.sh multi
```

**4. Burst Test**
Simulate traffic bursts:
```bash
./performance/vegeta-test.sh burst
```

**5. Sustained Load**
Long-running test:
```bash
# 50 req/s for 10 minutes
./performance/vegeta-test.sh sustained 50 600s
```

## 💪 Stress Testing

### Using the Stress Test Script

```bash
./performance/stress-test.sh [TEST_TYPE]
```

#### Test Types

**1. Progressive Load**
Find the breaking point:
```bash
./performance/stress-test.sh progressive
```

Tests: 10, 25, 50, 100, 200, 500, 1000, 2000, 5000 connections

**2. Spike Test**
Sudden traffic spike:
```bash
./performance/stress-test.sh spike
```

Phases:
- Baseline: 10 connections
- SPIKE: 1000 connections
- Recovery: 10 connections

**3. Soak Test**
Sustained load over time:
```bash
# 5 minutes, 100 connections (default)
./performance/stress-test.sh soak

# Custom duration and connections
./performance/stress-test.sh soak 600 200
```

**4. Concurrency Test**
Multiple domains simultaneously:
```bash
./performance/stress-test.sh concurrency
```

**5. Memory Leak Test**
Repeated requests to detect memory leaks:
```bash
./performance/stress-test.sh memory
```

**6. Timeout Test**
Connection timeout handling:
```bash
./performance/stress-test.sh timeout
```

**7. Resource Exhaustion**
Extreme load test:
```bash
./performance/stress-test.sh resource
```

**8. All Tests**
⚠️ **WARNING**: Takes 30+ minutes
```bash
./performance/stress-test.sh all
```

## 🐳 Docker Environment

### Using Docker Compose for Testing

Provides an isolated, consistent testing environment:

```bash
# Start server with resource limits
cd dynapins-server
export PRIVATE_KEY_PEM=$(cat path/to/private_key.pem)
docker-compose -f performance/docker-compose.perf.yml up -d

# Run tests from load-tester container
docker-compose -f performance/docker-compose.perf.yml exec load-tester sh

# Inside container:
apk add --no-cache curl
go install github.com/rakyll/hey@latest
hey -z 30s -c 100 http://dynapins-server:8080/v1/pins?domain=google.com

# Cleanup
docker-compose -f performance/docker-compose.perf.yml down
```

### Resource Limits

The Docker setup includes resource constraints:

```yaml
limits:
  cpus: '2'
  memory: 512M
reservations:
  cpus: '1'
  memory: 256M
```

This ensures consistent testing conditions.

## 📈 Interpreting Results

### Key Metrics

#### Response Time
- **Mean**: Average response time
- **Median (P50)**: 50% of requests faster than this
- **P95**: 95% of requests faster than this
- **P99**: 99% of requests faster than this

**Good**: P95 < 200ms, P99 < 500ms

#### Throughput
- **Requests/sec**: Number of completed requests per second

**Good**: >100 req/s with <1% errors

#### Error Rate
- **Success Rate**: Percentage of successful requests
- **Error Rate**: Percentage of failed requests

**Good**: Error rate <1%

#### Resource Usage
- **CPU**: Processor utilization
- **Memory**: RAM usage
- **Connections**: Active connections

**Good**: Stable resource usage over time

### Example Output Analysis

#### hey Output
```
Summary:
  Total:        30.0023 secs
  Slowest:      0.5234 secs
  Fastest:      0.0234 secs
  Average:      0.0856 secs
  Requests/sec: 1166.58
  
Status code distribution:
  [200] 35000 responses
```

**Analysis**:
- ✅ High throughput (1166 req/s)
- ✅ Low average latency (85ms)
- ✅ All requests successful

#### vegeta Output
```
Requests      [total, rate, throughput]  3000, 100.03, 100.01
Duration      [total, attack, wait]      29.991s, 29.990s, 1.234ms
Latencies     [mean, 50, 95, 99, max]    82.345ms, 78.234ms, 156.789ms, 234.567ms, 456.789ms
Bytes In      [total, mean]              1234567, 411.52
Bytes Out     [total, mean]              0, 0.00
Success       [ratio]                    100.00%
```

**Analysis**:
- ✅ Consistent throughput (~100 req/s)
- ✅ Good latency (P95: 156ms, P99: 234ms)
- ✅ 100% success rate

### Warning Signs

🚨 **Problems to watch for**:

1. **High P99 latency**: Some requests taking very long
2. **Increasing latency over time**: Possible memory leak or resource exhaustion
3. **High error rate**: Server overloaded or errors
4. **Dropping throughput**: Performance degradation
5. **Uneven distribution**: Some requests much slower than others

## 🎯 Performance Targets

### Expected Performance Metrics

#### Endpoints

| Endpoint | Mean | P95 | P99 | Throughput |
|----------|------|-----|-----|------------|
| `/health` | <5ms | <10ms | <20ms | >5000 req/s |
| `/readiness` | <10ms | <20ms | <50ms | >2000 req/s |
| `/v1/pins` | <100ms | <200ms | <500ms | >100 req/s |

#### Concurrent Connections

| Connections | Success Rate | Mean Latency | Throughput |
|-------------|-------------|--------------|------------|
| 10 | 100% | <50ms | >200 req/s |
| 100 | >99% | <100ms | >1000 req/s |
| 500 | >99% | <200ms | >2000 req/s |
| 1000 | >95% | <500ms | >1500 req/s |

#### Stress Limits

- **Breaking Point**: >1000 concurrent connections
- **Soak Test**: Stable for >1 hour at 100 req/s
- **Spike Recovery**: <10s recovery time
- **Memory Leak**: <5% memory increase over 100k requests

### System Requirements

For optimal performance:

- **CPU**: 2+ cores
- **Memory**: 512MB+ RAM
- **Network**: <50ms latency to target domains
- **OS**: Linux (best performance), macOS, Windows

## 🔧 Troubleshooting

### Common Issues

**1. "hey not found"**
```bash
go install github.com/rakyll/hey@latest
export PATH=$PATH:$(go env GOPATH)/bin
```

**2. "Connection refused"**
```bash
# Check if server is running
curl http://localhost:8080/health

# Start server
cd dynapins-server
export PRIVATE_KEY_PEM=$(cat private_key.pem)
export ALLOWED_DOMAINS="google.com"
go run ./cmd/server
```

**3. "Too many open files"**
```bash
# Increase file descriptor limit (macOS/Linux)
ulimit -n 10000
```

**4. High latency results**
- Network latency to target domains
- Server under load from other processes
- Insufficient system resources

**5. Test failures**
- Ensure domain is in ALLOWED_DOMAINS
- Check certificate retrieval is working
- Verify network connectivity

## 📝 Best Practices

### Before Testing

1. ✅ Stop unnecessary services
2. ✅ Set consistent environment variables
3. ✅ Warm up the server (few requests first)
4. ✅ Use isolated environment (Docker)
5. ✅ Document system configuration

### During Testing

1. ✅ Monitor system resources (CPU, memory, network)
2. ✅ Start with low load, increase gradually
3. ✅ Include cool-down periods
4. ✅ Run tests multiple times for consistency
5. ✅ Save results with timestamps

### After Testing

1. ✅ Analyze all metrics, not just averages
2. ✅ Compare with baseline results
3. ✅ Document findings and anomalies
4. ✅ Clean up test data and processes
5. ✅ Archive results for future comparison

## 📚 Additional Resources

- [Go Benchmarking Guide](https://pkg.go.dev/testing#hdr-Benchmarks)
- [hey Documentation](https://github.com/rakyll/hey)
- [wrk Documentation](https://github.com/wg/wrk)
- [vegeta Documentation](https://github.com/tsenart/vegeta)
- [Load Testing Best Practices](https://www.nginx.com/blog/load-testing-best-practices/)

## 🤝 Contributing

When adding new performance tests:

1. Add benchmark tests for new components
2. Update load testing scenarios if needed
3. Document expected performance metrics
4. Include test results in PR description
5. Update this README with new test types

---

**Last Updated**: 2025-10-17
**Performance Suite Version**: 1.0.0

