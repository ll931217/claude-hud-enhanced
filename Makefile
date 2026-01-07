.PHONY: build clean test run release release-all benchmark help

BINARY_NAME=claude-hud
BUILD_DIR=bin
RELEASE_DIR=release
GO=go
GOLINT=golangci-lint

# Version information (can be overridden by make arguments)
VERSION?=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
GIT_COMMIT?=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE?=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
GO_VERSION?=$(shell go version | cut -d' ' -f3)

# Build flags for injecting version information
LDFLAGS=-ldflags "-X github.com/ll931217/claude-hud-enhanced/internal/version.Version=$(VERSION) \
                   -X github.com/ll931217/claude-hud-enhanced/internal/version.GitCommit=$(GIT_COMMIT) \
                   -X github.com/ll931217/claude-hud-enhanced/internal/version.BuildDate=$(BUILD_DATE) \
                   -X github.com/ll931217/claude-hud-enhanced/internal/version.GoVersion=$(GO_VERSION)"

# Platform-specific build targets
PLATFORMS=linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64

help:
	@echo "Claude HUD Enhanced - Build System"
	@echo "=================================="
	@echo ""
	@echo "Available targets:"
	@echo "  make build        - Build for current platform"
	@echo "  make release      - Build release for current platform"
	@echo "  make release-all  - Build releases for all platforms"
	@echo "  make clean        - Clean build artifacts"
	@echo "  make test         - Run tests"
	@echo "  make benchmark    - Run benchmarks"
	@echo "  make run          - Run the application"
	@echo "  make fmt          - Format code"
	@echo "  make lint         - Run linter"
	@echo ""
	@echo "Build variables:"
	@echo "  VERSION          - Version tag (default: git describe or 'dev')"
	@echo "  GIT_COMMIT       - Git commit hash (default: git rev-parse)"
	@echo "  BUILD_DATE       - Build timestamp (default: current time)"
	@echo "  GO_VERSION       - Go version (default: go version output)"

build:
	@echo "Building $(BINARY_NAME) v$(VERSION)..."
	@mkdir -p $(BUILD_DIR)
	$(GO) build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/claude-hud
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

release:
	@echo "Building release $(BINARY_NAME) v$(VERSION) for $(GOOS)/$(GOARCH)..."
	@mkdir -p $(RELEASE_DIR)
	$(GO) build $(LDFLAGS) -trimpath -o $(RELEASE_DIR)/$(BINARY_NAME)$(BINARY_EXT) ./cmd/claude-hud
	@echo "Release build complete: $(RELEASE_DIR)/$(BINARY_NAME)$(BINARY_EXT)"

release-all: $(PLATFORMS)

$(PLATFORMS):
	@echo "Building for $@..."
	@$(MAKE) release GOOS=$(word 1,$(subst /, ,$@)) GOARCH=$(word 2,$(subst /, ,$@))

# Platform-specific binary extensions
BINARY_EXT=
ifeq ($(GOOS),windows)
BINARY_EXT=.exe
endif

# Archive creation for releases
.PHONY: archives
archives: release-all
	@echo "Creating release archives..."
	@for platform in $(PLATFORMS); do \
		os=$$(echo $$platform | cut -d'/' -f1); \
		arch=$$(echo $$platform | cut -d'/' -f2); \
		binary=$(RELEASE_DIR)/$(BINARY_NAME); \
		if [ "$$os" = "windows" ]; then \
			binary=$$binary.exe; \
		fi; \
		if [ -f "$$binary" ]; then \
			archive=$(RELEASE_DIR)/$(BINARY_NAME)-$(VERSION)-$$os-$$arch; \
			if [ "$$os" = "windows" ]; then \
				cd $(RELEASE_DIR) && zip -q $$archive.zip $$(basename $$binary); \
				echo "Created: $$archive.zip"; \
			else \
				tar -C $(RELEASE_DIR) -czf $$archive.tar.gz $$(basename $$binary); \
				echo "Created: $$archive.tar.gz"; \
			fi; \
		fi; \
	done

clean:
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR) $(RELEASE_DIR)
	$(GO) clean

test:
	@echo "Running tests..."
	$(GO) test -v -race -cover ./...

benchmark:
	@echo "Running benchmarks..."
	$(GO) test -bench=. -benchmem ./...

run:
	@echo "Running $(BINARY_NAME) v$(VERSION)..."
	$(GO) run $(LDFLAGS) ./cmd/claude-hud

fmt:
	@echo "Formatting..."
	$(GO) fmt ./...

lint:
	@echo "Linting..."
	$(GOLINT) run ./...

# Development targets
.PHONY: dev dev-quick
dev: fmt lint test

dev-quick:
	@$(MAKE) build && $(BUILD_DIR)/$(BINARY_NAME)
