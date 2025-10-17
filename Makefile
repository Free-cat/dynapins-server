.PHONY: help build test test-coverage run clean fmt vet lint docker-build docker-run

# Configuration
IMAGE_NAME = dynapins-server
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")

# Default target
help:
	@echo "Available targets:"
	@echo "  make build          - Build the server binary"
	@echo "  make test           - Run all tests"
	@echo "  make test-coverage  - Run tests with coverage report"
	@echo "  make run            - Run the server locally"
	@echo "  make clean          - Remove build artifacts"
	@echo "  make fmt            - Format code with go fmt"
	@echo "  make vet            - Run go vet"
	@echo "  make lint           - Run all code quality checks"
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
