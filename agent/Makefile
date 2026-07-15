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

.PHONY: all build runtime cli clean test lint fmt vet coverage run deps deps-update install uninstall verify deploy

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

## Package: build + create release archives
package: cross
	@mkdir -p $(OUTPUT_DIR)/release
	@cd $(OUTPUT_DIR) && tar czf release/buzzpi-runtime-linux-arm64.tar.gz buzzpi-runtime-linux-arm64
	@cd $(OUTPUT_DIR) && tar czf release/buzzpi-runtime-linux-arm.tar.gz buzzpi-runtime-linux-arm
	@cd $(OUTPUT_DIR) && tar czf release/buzzpi-runtime-linux-amd64.tar.gz buzzpi-runtime-linux-amd64
	@cp install.sh $(OUTPUT_DIR)/release/
	@cp verify.sh $(OUTPUT_DIR)/release/
	@cp setup.sh $(OUTPUT_DIR)/release/
	@echo "Release packages created in $(OUTPUT_DIR)/release/"

## Quick install: download + install on Pi via SSH (one-liner)
## Usage: make quick-install PI=192.168.1.50
PI ?= 
PI_USER ?= pi

quick-install:
ifndef PI
	$(error PI is not set. Usage: make quick-install PI=192.168.1.50)
endif
	@echo "Installing BuzzPi on $(PI_USER)@$(PI)..."
	ssh $(PI_USER)@$(PI) "curl -sSL https://jphat.net/buzzpi/setup.sh | sudo bash" || \
	ssh $(PI_USER)@$(PI) "wget -qO- https://jphat.net/buzzpi/setup.sh | sudo bash"
	@echo "Done! BuzzPi is running."

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

## Install on the local Pi (run on the Pi itself)
install:
	@sudo ./install.sh

## Uninstall from the local Pi (run on the Pi itself)
uninstall:
	@sudo ./install.sh --uninstall

## Upgrade binary on the local Pi (run on the Pi itself)
upgrade:
	@sudo ./install.sh --upgrade

## Verify installation (run on the Pi after install)
verify:
	@sudo ./verify.sh

## Deploy: cross-compile + copy to Pi over SSH
## Usage: make deploy PI=192.168.1.50 PI_USER=pi
PI ?= 
PI_USER ?= pi
PI_ARCH ?= arm64

deploy:
ifndef PI
	$(error PI is not set. Usage: make deploy PI=192.168.1.50)
endif
	@echo "Building for linux/$(PI_ARCH)..."
	@$(MAKE) cross-runtime-$(PI_ARCH)
	@echo "Copying to $(PI_USER)@$(PI)..."
	scp $(OUTPUT_DIR)/buzzpi-runtime-linux-$(PI_ARCH) $(PI_USER)@$(PI):~/
	@echo "Installing on Pi..."
	ssh $(PI_USER)@$(PI) "chmod +x ~/buzzpi-runtime-linux-$(PI_ARCH) && sudo ~/buzzpi-runtime-linux-$(PI_ARCH) --install"
	@echo "Verifying..."
	ssh $(PI_USER)@$(PI) "sudo ./install.sh" || true
	@echo "Deploy complete!"

## Deploy with install.sh (uses make deploy)
deploy-script:
ifndef PI
	$(error PI is not set. Usage: make deploy-script PI=192.168.1.50)
endif
	@echo "Building for linux/$(PI_ARCH)..."
	@$(MAKE) cross-runtime-$(PI_ARCH)
	@echo "Copying files to $(PI_USER)@$(PI)..."
	scp $(OUTPUT_DIR)/buzzpi-runtime-linux-$(PI_ARCH) $(PI_USER)@$(PI):~/buzzpi-runtime
	scp install.sh $(PI_USER)@$(PI):~/install.sh
	scp verify.sh $(PI_USER)@$(PI):~/verify.sh
	@echo "Installing on Pi..."
	ssh $(PI_USER)@$(PI) "chmod +x ~/buzzpi-runtime ~/install.sh ~/verify.sh && sudo ./install.sh"
	@echo "Deploy complete!"

## Help: list all targets
help:
	@echo "BuzzPi Agent Build System"
	@echo ""
	@echo "Build Targets:"
	@echo "  all          Default: fmt vet lint test build"
	@echo "  build        Build Runtime + CLI for current platform"
	@echo "  runtime      Build only the Runtime daemon"
	@echo "  cli          Build only the CLI client"
	@echo "  cross        Cross-compile for all target platforms"
	@echo "  clean        Remove build artifacts"
	@echo ""
	@echo "Code Quality:"
	@echo "  fmt          Format all Go source files"
	@echo "  vet          Run go vet"
	@echo "  lint         Run golangci-lint"
	@echo "  test         Run unit tests"
	@echo "  test-race    Run unit tests with race detector"
	@echo "  test-all     Run all tests (including integration)"
	@echo "  coverage     Generate code coverage report"
	@echo ""
	@echo "Run:"
	@echo "  run          Run Runtime locally"
	@echo "  deps         Download and tidy dependencies"
	@echo ""
	@echo "Deploy to Raspberry Pi:"
	@echo "  deploy       Cross-compile + SSH deploy to Pi"
	@echo "               Usage: make deploy PI=192.168.1.50 PI_USER=pi"
	@echo "  deploy-script  Same but uses install.sh"
	@echo ""
	@echo "Remote (run on Pi):"
	@echo "  install      Install BuzzPi on this machine"
	@echo "  uninstall    Remove BuzzPi from this machine"
	@echo "  upgrade      Upgrade binary and restart"
	@echo "  verify       Run post-install verification"
