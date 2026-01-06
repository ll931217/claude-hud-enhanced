.PHONY: build clean test run

BINARY_NAME=claude-hud
BUILD_DIR=bin
GO=go

build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GO) build -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/claude-hud

clean:
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)
	$(GO) clean

test:
	@echo "Running tests..."
	$(GO) test -v -race -cover ./...

run:
	@echo "Running..."
	$(GO) run ./cmd/claude-hud

fmt:
	@echo "Formatting..."
	$(GO) fmt ./...

lint:
	@echo "Linting..."
	golangci-lint run ./...
