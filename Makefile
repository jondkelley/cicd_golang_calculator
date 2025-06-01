# Go Calculator Makefile
.PHONY: all build build-all build-linux build-windows build-darwin test clean fmt vet lint deps tidy help

# Variables
APP_NAME := calc
MODULE := github.com/jondkelley/cicd_golang_calculator
CMD_DIR := ./cmd/calculator

# Version handling - prefer environment variable, fallback to git describe
ifdef VERSION
	VERSION := $(VERSION)
else
	VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev-unknown")
endif

BUILD_TIME := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS := -ldflags "-X main.version=$(VERSION) -X main.buildTime=$(BUILD_TIME)"

# Architecture targets
PLATFORMS := linux/amd64 windows/amd64 darwin/amd64 darwin/arm64
BINARIES := $(foreach platform,$(PLATFORMS),$(APP_NAME)-$(subst /,-,$(platform)))

# Default target
all: clean deps fmt vet test build

# Clean and setup
clean:
	@echo "Cleaning up..."
	go clean -cache
	rm -f $(APP_NAME) $(APP_NAME)-*
	rm -f go.mod go.sum

# Initialize go module
init:
	@echo "Initializing Go module..."
	@echo "module $(MODULE)" > go.mod
	@echo "go 1.21" >> go.mod
	go mod edit -replace $(MODULE)=./

# Install dependencies
deps: init
	@echo "Getting dependencies..."
	go mod tidy

# Format code
fmt:
	@echo "Formatting code..."
	go fmt ./...

# Vet code
vet:
	@echo "Vetting code..."
	go vet ./...

# Lint code (requires golangci-lint)
lint:
	@echo "Linting code..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed, skipping..."; \
	fi

# Tidy dependencies
tidy:
	@echo "Tidying dependencies..."
	go mod tidy

# Run tests
test:
	@echo "Running tests..."
	go test ./... -v

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test ./... -v -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html

# Run benchmarks
bench:
	@echo "Running benchmarks..."
	go test ./... -bench=.

# Build for current platform
build: deps fmt
	@echo "Building calculator for current platform..."
	@echo "Using version: $(VERSION)"
	go build $(LDFLAGS) -o $(APP_NAME) $(CMD_DIR)
	@echo "Built: ./$(APP_NAME)"

# Build for all platforms
build-all: deps fmt $(BINARIES)
	@echo "All binaries built with version: $(VERSION)"

# Build for Linux
build-linux: deps fmt
	@echo "Building for Linux with version: $(VERSION)"
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(APP_NAME)-linux-amd64 $(CMD_DIR)

# Build for Windows  
build-windows: deps fmt
	@echo "Building for Windows with version: $(VERSION)"
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(APP_NAME)-windows-amd64.exe $(CMD_DIR)

# Build for macOS Intel
build-darwin: deps fmt
	@echo "Building for macOS (Intel) with version: $(VERSION)"
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(APP_NAME)-darwin-amd64 $(CMD_DIR)

# Build for macOS Apple Silicon
build-darwin-arm64: deps fmt
	@echo "Building for macOS (Apple Silicon) with version: $(VERSION)"
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(APP_NAME)-darwin-arm64 $(CMD_DIR)

# Generic build rule for all platforms
$(APP_NAME)-%: deps fmt
	$(eval GOOS := $(word 1,$(subst -, ,$*)))
	$(eval GOARCH := $(word 2,$(subst -, ,$*)))
	$(eval EXT := $(if $(filter windows,$(GOOS)),.exe,))
	@echo "Building for $(GOOS)/$(GOARCH) with version: $(VERSION)"
	GOOS=$(GOOS) GOARCH=$(GOARCH) go build $(LDFLAGS) -o $@$(EXT) $(CMD_DIR)

# Development run
run: build
	@echo "Running calculator..."
	./$(APP_NAME)

# Development run with arguments
run-args: build
	@echo "Running calculator with arguments: $(ARGS)"
	./$(APP_NAME) $(ARGS)

# Release build (optimized)
release: deps fmt vet test
	@echo "Building release version with version: $(VERSION)"
	go build $(LDFLAGS) -a -installsuffix cgo -o $(APP_NAME) $(CMD_DIR)

tag:
	@if [ -z "$(TAG)" ]; then \
		echo "Usage: make tag TAG=v1.0.1 or make tag TAG=v1.0.1-alpha or make tag TAG=v1.0.1-beta"; \
		exit 1; \
	fi
	@echo "Creating tag: $(TAG)"
	git tag $(TAG)
	git push origin $(TAG)

# List all tags
list-tags:
	@echo "All tags:"
	git tag -l

# Delete a tag (local and remote)
delete-tag:
	@if [ -z "$(TAG)" ]; then \
		echo "Usage: make delete-tag TAG=v1.0.1"; \
		exit 1; \
	fi
	@echo "Deleting tag: $(TAG)"
	git tag -d $(TAG) || true
	git push origin --delete $(TAG) || true

# Show current version
version:
	@echo "Current version: $(VERSION)"
	@echo "Build time: $(BUILD_TIME)"

# Development cycle
dev: clean all run

# CI/CD simulation
ci: clean deps fmt vet lint test build-all

# Docker build (if you have a Dockerfile)
docker:
	@echo "Building Docker image..."
	docker build -t $(APP_NAME):$(VERSION) .

# Install binary to GOPATH/bin
install: build
	@echo "Installing $(APP_NAME) to $(GOPATH)/bin"
	cp $(APP_NAME) $(GOPATH)/bin/

help:
	@echo "Available targets:"
	@echo "  all           - Clean, deps, format, vet, test, and build"
	@echo "  build         - Build for current platform"
	@echo "  build-all     - Build for all platforms (Linux, Windows, macOS)"
	@echo "  build-linux   - Build for Linux"
	@echo "  build-windows - Build for Windows"
	@echo "  build-darwin  - Build for macOS (Intel)"
	@echo "  build-darwin-arm64 - Build for macOS (Apple Silicon)"
	@echo "  test          - Run tests"
	@echo "  test-coverage - Run tests with coverage report"
	@echo "  bench         - Run benchmarks"
	@echo "  clean         - Clean build artifacts and cache"
	@echo "  deps          - Initialize module and get dependencies"
	@echo "  fmt           - Format code"
	@echo "  vet           - Vet code"
	@echo "  lint          - Lint code (requires golangci-lint)"
	@echo "  run           - Build and run the application"
	@echo "  run-args      - Run with arguments: make run-args ARGS='1 + 2'"
	@echo "  release       - Build optimized release version"
	@echo "  tag           - Create any tag: make tag TAG=v1.0.1 or TAG=v1.0.1-alpha"
	@echo "  list-tags     - List all git tags"
	@echo "  delete-tag    - Delete tag: make delete-tag TAG=v1.0.1"
	@echo "  version       - Show current version info"
	@echo "  dev           - Full development cycle (clean, build, run)"
	@echo "  ci            - Simulate CI pipeline"
	@echo "  install       - Install binary to GOPATH/bin"
	@echo "  help          - Show this help message"
	@echo ""
	@echo "Environment variables:"
	@echo "  VERSION       - Override version (useful in CI): make build VERSION=v1.2.3"