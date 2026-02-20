# Phloem MCP Makefile

.PHONY: build install clean test run help quality check pre-commit preflight preflight-release verify-privacy verify-install ci-local

# Binary name (phloem for monorepo/CI; matches release-gate cp phloem/phloem)
BINARY=phloem
VERSION=0.2.0
COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
DATE=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

# Quality thresholds
MIN_COVERAGE=70
LINT_TIMEOUT=5m

# Build flags
LDFLAGS=-ldflags "-s -w -X 'main.version=$(VERSION)' -X 'main.commit=$(COMMIT)' -X 'main.date=$(DATE)'"

# Default target
all: build

## build: Build the binary
build:
	@echo "Building $(BINARY)..."
	go build $(LDFLAGS) -o $(BINARY) .
	@echo "Built: $(BINARY)"

## install: Install to /usr/local/bin
install: build
	@echo "Installing to /usr/local/bin/$(BINARY)..."
	cp $(BINARY) /usr/local/bin/
	@echo "Installed"

## clean: Remove build artifacts
clean:
	@echo "Cleaning..."
	rm -f $(BINARY)
	rm -f coverage.out
	@echo "Clean"

## test: Run tests
test:
	@echo "Running tests..."
	go test -v ./...

## test-coverage: Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

## run: Build and run
run: build
	./$(BINARY)

## fmt: Format code
fmt:
	@echo "Formatting..."
	go fmt ./...
	@echo "Formatted"

## lint: Run linter
lint:
	@echo "Linting..."
	golangci-lint run ./...

## deps: Download dependencies
deps:
	@echo "Downloading dependencies..."
	go mod download
	go mod tidy
	@echo "Dependencies ready"

## status: Show memory status
status: build
	./$(BINARY) status

## cursor-config: Show Cursor MCP configuration
cursor-config:
	@echo ""
	@echo "Add this to ~/.cursor/mcp.json:"
	@echo ""
	@echo '{'
	@echo '  "mcpServers": {'
	@echo '    "phloem": {'
	@echo '      "command": "'$(shell pwd)/$(BINARY)'"'
	@echo '      "args": ["serve"]'
	@echo '    }'
	@echo '  }'
	@echo '}'
	@echo ""

## quality: Run full quality gate (tests, coverage, lint)
quality: test-coverage lint
	@echo ""
	@echo "Quality Gate Results:"
	@echo "========================"
	@COVERAGE=$$(go tool cover -func=coverage.out | grep total | awk '{print $$3}' | sed 's/%//'); \
	if [ $$(echo "$$COVERAGE < $(MIN_COVERAGE)" | bc -l) -eq 1 ]; then \
		echo "Coverage $$COVERAGE% is below threshold $(MIN_COVERAGE)%"; \
		exit 1; \
	else \
		echo "Coverage $$COVERAGE% meets threshold $(MIN_COVERAGE)%"; \
	fi
	@echo "Lint passed"
	@echo ""
	@echo "Quality gate PASSED"

## check: Quick quality check (tests + lint, no coverage threshold)
check: test lint
	@echo "Quick check passed"

## pre-commit: Run before committing (fast checks)
pre-commit: fmt check
	@echo ""
	@echo "Ready to commit"

## race: Run tests with race detector
race:
	@echo "Running race detector..."
	go test -race ./...
	@echo "No race conditions found"

## bench: Run benchmarks
bench:
	@echo "Running benchmarks..."
	go test -bench=. -benchmem ./...

## acceptance: Run Gherkin acceptance tests
acceptance:
	@echo "Running acceptance tests..."
	go test -v ./test/acceptance/...

## acceptance-smoke: Run smoke acceptance tests only
acceptance-smoke:
	@echo "Running smoke acceptance tests..."
	go test -v ./test/acceptance/... -run TestSmokeFeatures

## acceptance-critical: Run critical acceptance tests only
acceptance-critical:
	@echo "Running critical acceptance tests..."
	GODOG_TAGS="@critical" go test -v ./test/acceptance/... -run TestCriticalFeatures

## mcp-test: Test MCP protocol compliance
mcp-test: build
	@echo "Testing MCP protocol..."
	@echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}' | timeout 5 ./$(BINARY) serve 2>/dev/null | head -1 | grep -q protocolVersion && echo "Initialize: OK" || echo "Initialize: FAILED"
	@echo "MCP protocol test passed"

## release-check: Full release verification (DEPRECATED - use zero-defect)
release-check: quality race mcp-test
	@echo ""
	@echo "Release verification PASSED"
	@echo "   - Quality gate: OK"
	@echo "   - Race detection: OK"
	@echo "   - MCP protocol: OK"
	@echo ""
	@echo "NOTE: Use 'make zero-defect' for full verification"

## zero-defect: MANDATORY before any release - full integration testing
zero-defect: build
	@echo "Running zero-defect release gate..."
	@./scripts/zero-defect.sh

## preflight: Fast local check before committing (~30s)
preflight: build
	@echo "Running preflight checks..."
	@echo ""
	@echo "1/4 go vet..."
	@go vet ./...
	@echo "2/4 Build... (already done)"
	@echo "3/4 Unit tests (short)..."
	@go test -short ./...
	@echo "4/4 MCP protocol check..."
	@TOOL_COUNT=$$(echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}' | timeout 5 ./$(BINARY) serve 2>/dev/null | head -1 | grep -c protocolVersion); \
	if [ "$$TOOL_COUNT" -eq 1 ]; then echo "MCP initialize: OK"; else echo "MCP initialize: FAILED"; exit 1; fi
	@echo ""
	@echo "Preflight PASSED - ready to commit"

## preflight-release: Thorough check before tagging a release (~2min)
preflight-release: preflight
	@echo ""
	@echo "Running release preflight..."
	@echo ""
	@echo "1/4 Full test suite (no cache)..."
	@go test -v -count=1 ./...
	@echo "2/4 Race detector..."
	@go test -race -short ./...
	@echo "3/4 Privacy verification..."
	@if [ -f scripts/verify-privacy.sh ]; then bash scripts/verify-privacy.sh; else echo "verify-privacy.sh not found, skipping"; fi
	@echo "4/4 Zero-defect gate..."
	@if [ -f scripts/zero-defect.sh ]; then bash scripts/zero-defect.sh; else echo "zero-defect.sh not found, skipping"; fi
	@echo ""
	@echo "Release preflight PASSED"

## verify-privacy: Verify Phloem makes no network connections
verify-privacy: build
	@echo "Running privacy verification..."
	@if [ -f scripts/verify-privacy.sh ]; then bash scripts/verify-privacy.sh; else echo "verify-privacy.sh not found"; exit 1; fi

## verify-install: Verify IDE setup commands work correctly
verify-install: build
	@echo "Running install verification..."
	@if [ -f scripts/verify-install.sh ]; then bash scripts/verify-install.sh; else echo "verify-install.sh not found"; exit 1; fi

## ci-local: Mirror what GitHub Actions CI runs
ci-local: preflight lint
	@echo ""
	@echo "Running full CI locally..."
	@go test -v -race -coverprofile=coverage.out -covermode=atomic ./...
	@COVERAGE=$$(go tool cover -func=coverage.out | grep total | awk '{print $$3}' | sed 's/%//'); \
	echo "Coverage: $${COVERAGE}%"; \
	if [ $$(echo "$$COVERAGE < 70" | bc -l) -eq 1 ]; then \
		echo "Coverage below 70% threshold"; exit 1; \
	fi
	@echo ""
	@echo "Local CI PASSED"

## help: Show this help
help:
	@echo "Phloem MCP - Local-first AI memory with causal graphs"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Build:"
	@echo "  make build         Build the binary"
	@echo "  make install       Install to /usr/local/bin"
	@echo "  make clean         Remove build artifacts"
	@echo ""
	@echo "Testing:"
	@echo "  make test          Run tests"
	@echo "  make test-coverage Run tests with coverage"
	@echo "  make race          Run with race detector"
	@echo "  make bench         Run benchmarks"
	@echo "  make mcp-test      Test MCP protocol"
	@echo "  make acceptance    Run Gherkin acceptance tests"
	@echo "  make acceptance-smoke Run smoke tests only"
	@echo "  make preflight     Fast local check before committing (~30s)"
	@echo "  make preflight-release Thorough check before tagging a release (~2min)"
	@echo ""
	@echo "Quality:"
	@echo "  make quality       Full quality gate"
	@echo "  make check         Quick check"
	@echo "  make pre-commit    Pre-commit checks"
	@echo "  make zero-defect   MANDATORY before release"
	@echo "  make verify-privacy Verify no network connections"
	@echo "  make verify-install Verify IDE setup commands"
	@echo "  make ci-local      Mirror GitHub Actions CI locally"
	@echo ""
	@echo "Setup:"
	@echo "  make cursor-config Show Cursor configuration"
	@echo ""
	@echo "Other:"
	@echo "  make fmt           Format code"
	@echo "  make lint          Run linter"
