# Go Calculator Makefile
.PHONY: all build build-all build-linux build-windows build-darwin test clean fmt vet lint deps tidy tag-stable tag-alpha help

# Variables
APP_NAME := calc
MODULE := github.com/jondkelley/cicd_golang_calculator
CMD_DIR := ./cmd/calculator
VERSION := $(shell git describe --tags --always --dirty)
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
	go build $(LDFLAGS) -o $(APP_NAME) $(CMD_DIR)
	@echo "Built: ./$(APP_NAME)"

# Build for all platforms
build-all: deps fmt $(BINARIES)

# Build for Linux
build-linux: deps fmt
	@echo "Building for Linux..."
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(APP_NAME)-linux $(CMD_DIR)

# Build for Windows
build-windows: deps fmt
	@echo "Building for Windows..."
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(APP_NAME)-windows.exe $(CMD_DIR)

# Build for macOS Intel
build-darwin: deps fmt
	@echo "Building for macOS (Intel)..."
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(APP_NAME)-macos $(CMD_DIR)

# Build for macOS Apple Silicon
build-darwin-arm64: deps fmt
	@echo "Building for macOS (Apple Silicon)..."
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(APP_NAME)-macos-arm64 $(CMD_DIR)

# Generic build rule for all platforms
$(APP_NAME)-%: deps fmt
	$(eval GOOS := $(word 1,$(subst -, ,$*)))
	$(eval GOARCH := $(word 2,$(subst -, ,$*)))
	$(eval EXT := $(if $(filter windows,$(GOOS)),.exe,))
	@echo "Building for $(GOOS)/$(GOARCH)..."
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
	@echo "Building release version..."
	go build $(LDFLAGS) -a -installsuffix cgo -o $(APP_NAME) $(CMD_DIR)

# Create and push stable tag
tag-stable:
	@if [ -z "$(TAG)" ]; then \
		echo "Usage: make tag-stable TAG=v1.0.1"; \
		exit 1; \
	fi
	@echo "Creating stable tag: $(TAG)"
	git tag $(TAG)
	git push origin $(TAG)

# Create and push alpha tag
tag-alpha:
	@if [ -z "$(TAG)" ]; then \
		echo "Usage: make tag-alpha TAG=v1.0.1-alpha"; \
		exit 1; \
	fi
	@echo "Creating alpha tag: $(TAG)"
	git tag $(TAG)
	git push origin $(TAG)

# Create and push tag (auto-detect stable vs alpha)
tag:
	@if [ -z "$(TAG)" ]; then \
		echo "Usage: make tag TAG=v1.0.1 or make tag TAG=v1.0.1-alpha"; \
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

# Show help
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
	@echo "  tag-stable    - Create stable tag: make tag-stable TAG=v1.0.1"
	@echo "  tag-alpha     - Create alpha tag: make tag-alpha TAG=v1.0.1-alpha"
	@echo "  tag           - Create any tag: make tag TAG=v1.0.1"
	@echo "  list-tags     - List all git tags"
	@echo "  delete-tag    - Delete tag: make delete-tag TAG=v1.0.1"
	@echo "  version       - Show current version info"
	@echo "  dev           - Full development cycle (clean, build, run)"
	@echo "  ci            - Simulate CI pipeline"
	@echo "  install       - Install binary to GOPATH/bin"
	@echo "  help          - Show this help message"