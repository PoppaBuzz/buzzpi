# BuzzPi Agent Build System

GO ?= go
GOOS ?= linux
GOARCH ?= arm64
OUTPUT_DIR ?= build
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
VERSION ?= v0.1.0-dev

LDFLAGS := -ldflags "\
	-X github.com/buzzpi/agent/internal/version.Version=$(VERSION) \
	-X github.com/buzzpi/agent/internal/version.Commit=$(COMMIT) \
	-X github.com/buzzpi/agent/internal/version.Date=$(DATE)"

.PHONY: all build runtime cli clean test lint fmt vet coverage

all: fmt vet lint test build

## Build: produces binaries for the current platform
build: runtime cli

## Build the Runtime daemon
runtime:
	@mkdir -p $(OUTPUT_DIR)
	$(GO) build $(LDFLAGS) -o $(OUTPUT_DIR)/buzzpi-runtime ./cmd/runtime/

## Build the CLI client
cli:
	@mkdir -p $(OUTPUT_DIR)
	$(GO) build $(LDFLAGS) -o $(OUTPUT_DIR)/buzzpi ./cmd/buzzpi/

## Cross-compile for all target platforms
cross: cross-runtime-arm64 cross-runtime-arm cross-runtime-amd64 cross-cli-all

cross-runtime-arm64:
	GOOS=linux GOARCH=arm64 $(GO) build $(LDFLAGS) -o $(OUTPUT_DIR)/buzzpi-runtime-linux-arm64 ./cmd/runtime/

cross-runtime-arm:
	GOOS=linux GOARCH=arm GOARM=7 $(GO) build $(LDFLAGS) -o $(OUTPUT_DIR)/buzzpi-runtime-linux-arm ./cmd/runtime/

cross-runtime-amd64:
	GOOS=linux GOARCH=amd64 $(GO) build $(LDFLAGS) -o $(OUTPUT_DIR)/buzzpi-runtime-linux-amd64 ./cmd/runtime/

cross-cli-all:
	GOOS=linux GOARCH=arm64 $(GO) build $(LDFLAGS) -o $(OUTPUT_DIR)/buzzpi-linux-arm64 ./cmd/buzzpi/
	GOOS=linux GOARCH=amd64 $(GO) build $(LDFLAGS) -o $(OUTPUT_DIR)/buzzpi-linux-amd64 ./cmd/buzzpi/
	GOOS=darwin GOARCH=arm64 $(GO) build $(LDFLAGS) -o $(OUTPUT_DIR)/buzzpi-darwin-arm64 ./cmd/buzzpi/

## Clean build artifacts
clean:
	rm -rf $(OUTPUT_DIR)/
	rm -f coverage.out

## Format: run gofmt on all source files
fmt:
	$(GO) fmt ./...

## Vet: run go vet on all packages
vet:
	$(GO) vet ./...

## Lint: run golangci-lint (if available)
lint:
	@which golangci-lint > /dev/null && golangci-lint run ./... || echo "golangci-lint not installed, skipping"

## Test: run all unit tests
test:
	$(GO) test -short -count=1 ./...

## Test with race detection
test-race:
	$(GO) test -race -short -count=1 ./...

## Test all (including integration)
test-all:
	$(GO) test -count=1 ./...

## Coverage: generate and open coverage report
coverage:
	$(GO) test -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

## Run the Runtime locally
run:
	$(GO) run $(LDFLAGS) ./cmd/runtime/

## Download dependencies
deps:
	$(GO) mod download
	$(GO) mod tidy

## Check for dependency updates
deps-update:
	$(GO) get -u ./...
	$(GO) mod tidy

## Help: list all targets
help:
	@echo "BuzzPi Agent Build System"
	@echo ""
	@echo "Targets:"
	@echo "  all          Default: fmt vet lint test build"
	@echo "  build        Build Runtime + CLI for current platform"
	@echo "  runtime      Build only the Runtime daemon"
	@echo "  cli          Build only the CLI client"
	@echo "  cross        Cross-compile for all target platforms"
	@echo "  clean        Remove build artifacts"
	@echo "  fmt          Format all Go source files"
	@echo "  vet          Run go vet"
	@echo "  lint         Run golangci-lint"
	@echo "  test         Run unit tests"
	@echo "  test-race    Run unit tests with race detector"
	@echo "  test-all     Run all tests (including integration)"
	@echo "  coverage     Generate code coverage report"
	@echo "  run          Run Runtime locally"
	@echo "  deps         Download and tidy dependencies"
