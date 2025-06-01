# Go Calculator Makefile Documentation

This Makefile generates build, test, and deployment targets

## Quick Start

```bash
# Build everything (clean, deps, format, vet, test, build)
make all

# Just build for current platform
make build

# Run tests
make test
```

## Build Targets

### Basic Building
- `make build` - Build for your current platform
- `make build-all` - Build for all supported platforms (Linux, Windows, macOS Intel, macOS ARM64)
- `make release` - Build optimized release version with full checks

### Platform-Specific Builds
- `make build-linux` - Build for Linux (amd64)
- `make build-windows` - Build for Windows (amd64)
- `make build-darwin` - Build for macOS Intel (amd64)
- `make build-darwin-arm64` - Build for macOS Apple Silicon (arm64)

### Output Files
When you run `make build-all`, you'll get these binaries:
- `calc-linux` - Linux executable
- `calc-windows.exe` - Windows executable
- `calc-macos` - macOS Intel executable
- `calc-macos-arm64` - macOS Apple Silicon executable

## Testing Targets

- `make test` - Run all tests with verbose output
- `make test-coverage` - Run tests and generate HTML coverage report
- `make bench` - Run benchmark tests

## Code Quality Targets

- `make fmt` - Format all Go code
- `make vet` - Run `go vet` to catch potential issues
- `make lint` - Run golangci-lint (if installed)
- `make ci` - Full CI pipeline (clean, deps, fmt, vet, lint, test, build-all)

## Development Targets

- `make run` - Build and run the calculator
- `make run-args ARGS="1 + 2"` - Run with specific arguments
- `make dev` - Full development cycle (clean, build, run)

## Dependency Management

- `make deps` - Initialize Go module and download dependencies
- `make tidy` - Clean up go.mod and go.sum
- `make clean` - Remove all build artifacts and clear Go cache

## Tagging and Releases

### Creating Tags
```bash
# Create a stable release tag
make tag-stable TAG=v1.0.1

# Create an alpha release tag
make tag-alpha TAG=v1.0.2-alpha

# Create any tag (auto-detects stable vs alpha)
make tag TAG=v1.0.1
```

### Tag Management
```bash
# List all existing tags
make list-tags

# Delete a tag (both local and remote)
make delete-tag TAG=v1.0.1

# Show current version info
make version
```

## Installation and Docker

- `make install` - Install the binary to `$GOPATH/bin`
- `make docker` - Build Docker image (requires Dockerfile)

## Examples

### Basic Development Workflow
```bash
# Start fresh
make clean

# Build and test everything
make all

# Run the calculator
make run
```

### Multi-Platform Release
```bash
# Build for all platforms
make build-all

# Check what was built
ls -la calc-*

# Test the Linux version (if on Linux)
./calc-linux
```

### Release Process
```bash
# Run full CI checks
make ci

# Create and push a release tag
make tag-stable TAG=v1.2.0

# GitHub Actions will automatically:
# - Build multi-platform binaries
# - Create GitHub release
# - Update version.json
```

### Testing Workflow
```bash
# Run basic tests
make test

# Get coverage report
make test-coverage
# Opens coverage.html in browser

# Run benchmarks
make bench
```

### Code Quality Checks
```bash
# Format code
make fmt

# Check for issues
make vet

# Run linter (if golangci-lint installed)
make lint

# Or run all quality checks
make ci
```

## Environment Variables

The Makefile uses these variables that you can override:

```bash
# Change the app name
make build APP_NAME=mycalc

# Change module path
make deps MODULE=github.com/myuser/mycalc

# Pass arguments to run target
make run-args ARGS="5 * 10"
```

## Help

Run `make help` to see all available targets with descriptions.

## Integration with GitHub Actions

This Makefile is designed to work with your existing GitHub Actions workflow. When you push a tag:

1. GitHub Actions triggers on tag push
2. Runs tests and builds multi-platform binaries
3. Creates GitHub release with binaries
4. Updates version.json with release info

The Makefile's tagging commands make this process easy:

```bash
# For stable releases
make tag-stable TAG=v1.0.1

# For alpha releases
make tag-alpha TAG=v1.0.2-alpha
```

## Tips

- Always run `make ci` before pushing to ensure your code passes all checks
- Use `make dev` for quick development iterations
- The build includes version info from git tags automatically
- All binaries include build time and version information