# Rogue Planet Makefile

VERSION := 0.3.0
BINARY_NAME := rp
BIN_DIR := bin
COVERAGE_DIR := coverage

# Go parameters
GOCMD := go
GOBUILD := $(GOCMD) build
GOTEST := $(GOCMD) test
GOCLEAN := $(GOCMD) clean
GOMOD := $(GOCMD) mod
GOFMT := $(GOCMD) fmt
GOINSTALL := $(GOCMD) install

# Build flags
LDFLAGS := -ldflags="-s -w -X main.version=$(VERSION)"

.PHONY: all build test clean install fmt vet coverage help run examples

# Default target
all: clean fmt vet test build

## help: Show this help message
help:
	@echo "Rogue Planet - Makefile Commands"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Development:"
	@sed -n 's/^## dev://p' $(MAKEFILE_LIST) | column -t -s ':' | sed -e 's/^/  /'
	@echo ""
	@echo "Building:"
	@sed -n 's/^## build://p' $(MAKEFILE_LIST) | column -t -s ':' | sed -e 's/^/  /'
	@echo ""
	@echo "Testing:"
	@sed -n 's/^## test://p' $(MAKEFILE_LIST) | column -t -s ':' | sed -e 's/^/  /'
	@echo ""
	@echo "Other:"
	@sed -n 's/^## other://p' $(MAKEFILE_LIST) | column -t -s ':' | sed -e 's/^/  /'

## build: build: Build binary for current platform
build:
	@echo "Building $(BINARY_NAME) v$(VERSION)..."
	@mkdir -p $(BIN_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BIN_DIR)/$(BINARY_NAME) ./cmd/rp
	@echo "✓ Built $(BIN_DIR)/$(BINARY_NAME)"

## build: install: Install binary to GOPATH/bin
install:
	@INSTALL_PATH=$${GOBIN:-$${GOPATH:+$$GOPATH/bin}}; \
	INSTALL_PATH=$${INSTALL_PATH:-$$HOME/go/bin}; \
	echo "Installing $(BINARY_NAME) to $$INSTALL_PATH..."; \
	$(GOINSTALL) $(LDFLAGS) ./cmd/rp; \
	echo "✓ Installed to $$INSTALL_PATH/$(BINARY_NAME)"; \
	echo ""; \
	if echo "$$PATH" | grep -q "$$INSTALL_PATH"; then \
		echo "Run with: rp version"; \
	else \
		echo "⚠  $$INSTALL_PATH is not in your PATH"; \
		echo "Add it with: export PATH=\"$$INSTALL_PATH:\$$PATH\""; \
	fi

## test: test: Run all tests
test:
	@echo "Running tests..."
	@$(GOTEST) -v ./...
	@echo "✓ All tests passed"

## test: test-short: Run tests without verbose output
test-short:
	@$(GOTEST) ./...

## test: test-integration: Run integration tests
test-integration:
	@echo "Running integration tests..."
	@$(GOTEST) -v -tags=integration ./cmd/rp/... ./pkg/generator/...
	@echo "✓ Integration tests passed"

## test: coverage: Generate test coverage report
coverage:
	@echo "Running tests with coverage..."
	@mkdir -p $(COVERAGE_DIR)
	@$(GOTEST) -coverprofile=$(COVERAGE_DIR)/coverage.out -covermode=atomic ./...
	@$(GOCMD) tool cover -html=$(COVERAGE_DIR)/coverage.out -o $(COVERAGE_DIR)/coverage.html
	@$(GOCMD) tool cover -func=$(COVERAGE_DIR)/coverage.out | tail -1
	@echo "✓ Coverage report: $(COVERAGE_DIR)/coverage.html"

## test: test-race: Run tests with race detector
test-race:
	@echo "Running tests with race detector..."
	@$(GOTEST) -race ./...
	@echo "✓ No race conditions detected"

## test: bench: Run benchmarks
bench:
	@echo "Running benchmarks..."
	@$(GOTEST) -bench=. -benchmem ./...

## dev: fmt: Format Go code
fmt:
	@echo "Formatting code..."
	@$(GOFMT) ./...
	@echo "✓ Code formatted"

## dev: vet: Run go vet
vet:
	@echo "Running go vet..."
	@$(GOCMD) vet ./...
	@echo "✓ No issues found"

## dev: lint: Run golangci-lint (if installed)
lint:
	@echo "Running linters..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run ./...; \
		echo "✓ Linting complete"; \
	else \
		echo "⚠  golangci-lint not installed. Install with:"; \
		echo "  go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

## dev: check: Run all quality checks (fmt, vet, test, race)
check: fmt vet test test-race
	@echo "✓ All checks passed"

## dev: quick: Quick build for development (fmt + test + build)
quick: fmt test build
	@echo "✓ Quick build complete: $(BIN_DIR)/$(BINARY_NAME)"

## other: clean: Remove build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(BIN_DIR) $(COVERAGE_DIR)
	@$(GOCLEAN)
	@echo "✓ Cleaned"

## other: deps: Download and tidy dependencies
deps:
	@echo "Downloading dependencies..."
	@$(GOMOD) download
	@$(GOMOD) tidy
	@echo "✓ Dependencies updated"

## other: verify: Verify dependencies
verify:
	@echo "Verifying dependencies..."
	@$(GOMOD) verify
	@echo "✓ Dependencies verified"

## other: run: Build and run with example config
run: build
	@echo "Running rp..."
	@$(BIN_DIR)/$(BINARY_NAME) $(ARGS)

## other: run-example: Create example planet in /tmp
run-example: build
	@echo "Creating example planet in /tmp/rogue-planet-example..."
	@rm -rf /tmp/rogue-planet-example
	@mkdir -p /tmp/rogue-planet-example
	@cd /tmp/rogue-planet-example && $(CURDIR)/$(BIN_DIR)/$(BINARY_NAME) init
	@cd /tmp/rogue-planet-example && $(CURDIR)/$(BIN_DIR)/$(BINARY_NAME) add-feed https://blog.golang.org/feed.atom
	@cd /tmp/rogue-planet-example && $(CURDIR)/$(BIN_DIR)/$(BINARY_NAME) add-feed https://github.blog/feed/
	@cd /tmp/rogue-planet-example && $(CURDIR)/$(BIN_DIR)/$(BINARY_NAME) update
	@echo "✓ Example planet created at /tmp/rogue-planet-example/"
	@echo "  View: open /tmp/rogue-planet-example/public/index.html"

## other: setup-example: Run the example setup with custom feeds
setup-example: build
	@echo "Running example planet setup..."
	@./setup-example-planet.sh

## other: examples: Generate example outputs in tmp/ folder
examples: build
	@echo "Generating examples in tmp/ folder..."
	@rm -rf tmp
	@mkdir -p tmp/default tmp/classic tmp/elegant tmp/dark
	@echo ""
	@echo "==> Generating default theme example..."
	@cd tmp/default && $(CURDIR)/$(BIN_DIR)/$(BINARY_NAME) init
	@cp examples/config.ini tmp/default/config.ini
	@cd tmp/default && $(CURDIR)/$(BIN_DIR)/$(BINARY_NAME) add-feed https://blog.golang.org/feed.atom
	@cd tmp/default && $(CURDIR)/$(BIN_DIR)/$(BINARY_NAME) add-feed https://github.blog/feed/
	@cd tmp/default && $(CURDIR)/$(BIN_DIR)/$(BINARY_NAME) update
	@cd tmp/default && $(CURDIR)/$(BIN_DIR)/$(BINARY_NAME) generate
	@echo "✓ Default theme output: tmp/default/public/index.html"
	@echo ""
	@echo "==> Generating classic theme example..."
	@cd tmp/classic && $(CURDIR)/$(BIN_DIR)/$(BINARY_NAME) init
	@cp examples/config.ini tmp/classic/config.ini
	@mkdir -p tmp/classic/theme
	@cp -r examples/themes/classic/* tmp/classic/theme/
	@sed -i '' 's|#   template = ./themes/classic/template.html|template = $(CURDIR)/tmp/classic/theme/template.html|' tmp/classic/config.ini
	@cd tmp/classic && $(CURDIR)/$(BIN_DIR)/$(BINARY_NAME) add-feed https://blog.golang.org/feed.atom
	@cd tmp/classic && $(CURDIR)/$(BIN_DIR)/$(BINARY_NAME) add-feed https://github.blog/feed/
	@cd tmp/classic && $(CURDIR)/$(BIN_DIR)/$(BINARY_NAME) update
	@cd tmp/classic && $(CURDIR)/$(BIN_DIR)/$(BINARY_NAME) generate
	@echo "✓ Classic theme output: tmp/classic/public/index.html"
	@echo "✓ Classic theme files: tmp/classic/theme/"
	@echo ""
	@echo "==> Generating elegant theme example..."
	@cd tmp/elegant && $(CURDIR)/$(BIN_DIR)/$(BINARY_NAME) init
	@cp examples/config.ini tmp/elegant/config.ini
	@mkdir -p tmp/elegant/theme
	@cp -r examples/themes/elegant/* tmp/elegant/theme/
	@sed -i '' 's|#   template = ./themes/elegant/template.html|template = $(CURDIR)/tmp/elegant/theme/template.html|' tmp/elegant/config.ini
	@cd tmp/elegant && $(CURDIR)/$(BIN_DIR)/$(BINARY_NAME) add-feed https://blog.golang.org/feed.atom
	@cd tmp/elegant && $(CURDIR)/$(BIN_DIR)/$(BINARY_NAME) add-feed https://github.blog/feed/
	@cd tmp/elegant && $(CURDIR)/$(BIN_DIR)/$(BINARY_NAME) update
	@cd tmp/elegant && $(CURDIR)/$(BIN_DIR)/$(BINARY_NAME) generate
	@echo "✓ Elegant theme output: tmp/elegant/public/index.html"
	@echo "✓ Elegant theme files: tmp/elegant/theme/"
	@echo ""
	@echo "==> Generating dark theme example..."
	@cd tmp/dark && $(CURDIR)/$(BIN_DIR)/$(BINARY_NAME) init
	@cp examples/config.ini tmp/dark/config.ini
	@mkdir -p tmp/dark/theme
	@cp -r examples/themes/dark/* tmp/dark/theme/
	@sed -i '' 's|#   template = ./themes/elegant/template.html|template = $(CURDIR)/tmp/dark/theme/template.html|' tmp/dark/config.ini
	@cd tmp/dark && $(CURDIR)/$(BIN_DIR)/$(BINARY_NAME) add-feed https://blog.golang.org/feed.atom
	@cd tmp/dark && $(CURDIR)/$(BIN_DIR)/$(BINARY_NAME) add-feed https://github.blog/feed/
	@cd tmp/dark && $(CURDIR)/$(BIN_DIR)/$(BINARY_NAME) update
	@cd tmp/dark && $(CURDIR)/$(BIN_DIR)/$(BINARY_NAME) generate
	@echo "✓ Dark theme output: tmp/dark/public/index.html"
	@echo "✓ Dark theme files: tmp/dark/theme/"
	@echo ""
	@echo "✓ All examples generated in tmp/ folder"
	@echo ""
	@echo "  View default: open tmp/default/public/index.html"
	@echo "  View classic: open tmp/classic/public/index.html"
	@echo "  View elegant: open tmp/elegant/public/index.html"
	@echo "  View dark:    open tmp/dark/public/index.html"

# Development workflow targets
.PHONY: dev watch
dev: quick

# Watch for changes (requires entr: brew install entr)
watch:
	@if command -v entr >/dev/null 2>&1; then \
		find . -name "*.go" | entr -c make quick; \
	else \
		echo "⚠  entr not installed. Install with:"; \
		echo "  macOS: brew install entr"; \
		echo "  Linux: apt-get install entr or yum install entr"; \
	fi
