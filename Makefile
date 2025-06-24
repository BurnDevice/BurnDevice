.PHONY: help build run test clean proto generate nix-shell docker ci

# Variables with fallback for non-git environments
BINARY_NAME := burndevice
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE := $(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS := -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)

# Build optimization flags
OPTIMIZE_FLAGS := -ldflags="-s -w $(LDFLAGS)"
DEBUG_FLAGS := -ldflags="$(LDFLAGS)" -race

# Default target
help: ## Show this help message
	@echo "🔥 BurnDevice - 设备破坏性测试工具"
	@echo ""
	@echo "⚠️  警告：此工具仅用于授权的测试环境！"
	@echo ""
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

# Development environment
nix-shell: ## Enter Nix development shell
	nix develop

# Protocol Buffers
proto: ## Generate Protocol Buffer code
	buf generate

proto-lint: ## Lint Protocol Buffer files
	buf lint

proto-breaking: ## Check for breaking changes in proto files (requires git)
	@if git rev-parse --git-dir > /dev/null 2>&1; then \
		buf breaking --against '.git#branch=main'; \
	else \
		echo "⚠️  Git not initialized, skipping breaking change check"; \
	fi

# Build targets
build: proto ## Build the binary
	go build -ldflags="$(LDFLAGS)" -o bin/$(BINARY_NAME) ./cmd/burndevice

build-optimized: proto ## Build optimized binary (smaller size)
	go build $(OPTIMIZE_FLAGS) -o bin/$(BINARY_NAME)-optimized ./cmd/burndevice

build-debug: proto ## Build with debug info and race detection
	go build $(DEBUG_FLAGS) -o bin/$(BINARY_NAME)-debug ./cmd/burndevice

build-linux: proto ## Build for Linux
	GOOS=linux GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o bin/$(BINARY_NAME)-linux-amd64 ./cmd/burndevice

build-linux-arm64: proto ## Build for Linux ARM64
	GOOS=linux GOARCH=arm64 go build -ldflags="$(LDFLAGS)" -o bin/$(BINARY_NAME)-linux-arm64 ./cmd/burndevice

build-darwin: proto ## Build for macOS
	GOOS=darwin GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o bin/$(BINARY_NAME)-darwin-amd64 ./cmd/burndevice

build-darwin-arm64: proto ## Build for macOS ARM64 (Apple Silicon)
	GOOS=darwin GOARCH=arm64 go build -ldflags="$(LDFLAGS)" -o bin/$(BINARY_NAME)-darwin-arm64 ./cmd/burndevice

build-windows: proto ## Build for Windows
	GOOS=windows GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o bin/$(BINARY_NAME)-windows-amd64.exe ./cmd/burndevice

build-all: build-linux build-linux-arm64 build-darwin build-darwin-arm64 build-windows ## Build for all platforms

# Run targets
run: build ## Run the server
	./bin/$(BINARY_NAME) server

run-client: build ## Run the client
	./bin/$(BINARY_NAME) client

run-dev: ## Run in development mode with hot reload
	go run -ldflags="$(LDFLAGS)" ./cmd/burndevice server --config config.example.yaml

# Testing
test: ## Run all tests
	go test -v ./...

test-short: ## Run tests with short flag
	go test -short -v ./...

test-race: ## Run tests with race detection
	go test -race -v ./...

test-coverage: ## Run tests with coverage (excluding auto-generated protobuf code)
	go test -coverprofile=coverage.out ./cmd/... ./internal/...
	go tool cover -html=coverage.out -o coverage.html
	@echo "📊 Coverage report generated: coverage.html (excluding auto-generated protobuf code)"

test-coverage-func: ## Show test coverage by function (excluding auto-generated protobuf code)
	go test -coverprofile=coverage.out ./cmd/... ./internal/...
	go tool cover -func=coverage.out

test-coverage-summary: ## Show detailed test coverage summary by module
	@echo "🧪 Generating test coverage summary (excluding auto-generated protobuf code)..."
	@go test -coverprofile=coverage.out ./cmd/... ./internal/... 2>/dev/null
	@echo ""
	@echo "📊 Overall Coverage:"
	@echo "==================="
	@go tool cover -func=coverage.out | tail -1
	@echo ""
	@echo "📋 Module Coverage Details:"
	@echo "=========================="
	@go tool cover -func=coverage.out | grep -E "(cmd/|internal/)" | \
		awk '{print $$1, $$NF}' | sort -k2 -nr | \
		awk 'BEGIN{print "Module                          Coverage"} \
		     BEGIN{print "======                          ========"} \
		     {printf "%-30s %s\n", $$1, $$2}'
	@echo ""

benchmark: ## Run benchmarks
	go test -bench=. -benchmem ./...

# CI/CD targets
ci: deps proto-lint fmt vet test-race test-coverage security-check ## Run full CI pipeline

ci-quick: deps proto fmt vet test-short ## Run quick CI checks

quality-check: fmt vet lint ## Run all code quality checks

pre-commit: quality-check test-short ## Run pre-commit checks

# Dependencies
deps: ## Download dependencies
	go mod download
	go mod tidy

deps-update: ## Update dependencies
	go get -u ./...
	go mod tidy

deps-verify: ## Verify dependencies
	go mod verify

# Code quality
lint: ## Run linter
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "⚠️  golangci-lint not installed, skipping lint check"; \
	fi

fmt: ## Format code
	go fmt ./...

fmt-check: ## Check if code is formatted
	@if [ -n "$$(gofmt -s -l .)" ]; then \
		echo "❌ Code is not formatted. Run 'make fmt' to fix."; \
		gofmt -s -l .; \
		exit 1; \
	else \
		echo "✅ Code is properly formatted"; \
	fi

vet: ## Run go vet
	go vet ./...

# Security and vulnerability checks
security-check: ## Run security checks
	@echo "🔍 Running security checks..."
	@if command -v gosec >/dev/null 2>&1; then \
		gosec -exclude-dir=burndevice ./...; \
	else \
		echo "⚠️  gosec not installed, skipping security scan"; \
	fi
	@if command -v govulncheck >/dev/null 2>&1; then \
		govulncheck ./cmd/... ./internal/...; \
	else \
		echo "⚠️  govulncheck not installed, skipping vulnerability check"; \
	fi

# Docker
docker-build: ## Build Docker image
	docker build -t burndevice:$(VERSION) .

docker-build-multi: ## Build multi-platform Docker image
	docker buildx build --platform linux/amd64,linux/arm64 -t burndevice:$(VERSION) .

docker-run: ## Run Docker container
	docker run -p 8080:8080 burndevice:$(VERSION)

docker-compose-up: ## Start services with docker-compose
	docker-compose up -d

docker-compose-down: ## Stop services with docker-compose
	docker-compose down

# Cleanup
clean: ## Clean build artifacts
	rm -rf bin/
	rm -rf release/
	rm -f coverage.out coverage.html
	rm -f *.prof

clean-all: clean ## Clean everything including dependencies
	go clean -modcache

# Release
release: clean build-all ## Create release artifacts
	mkdir -p release
	cp bin/* release/
	tar -czf release/burndevice-$(VERSION)-linux-amd64.tar.gz -C bin $(BINARY_NAME)-linux-amd64
	tar -czf release/burndevice-$(VERSION)-linux-arm64.tar.gz -C bin $(BINARY_NAME)-linux-arm64
	tar -czf release/burndevice-$(VERSION)-darwin-amd64.tar.gz -C bin $(BINARY_NAME)-darwin-amd64
	tar -czf release/burndevice-$(VERSION)-darwin-arm64.tar.gz -C bin $(BINARY_NAME)-darwin-arm64
	zip -j release/burndevice-$(VERSION)-windows-amd64.zip bin/$(BINARY_NAME)-windows-amd64.exe
	@echo "📦 Release artifacts created in release/ directory"

# Release management
.PHONY: release-check release-tag release-local version-patch version-minor version-major

release-check: ## Pre-release checks
	@echo "🔍 Running pre-release checks..."
	@echo "1. Checking git status..."
	@git diff --quiet || (echo "❌ Working directory is dirty" && exit 1)
	@git diff --cached --quiet || (echo "❌ Staged changes found" && exit 1)
	@echo "2. Checking current branch..."
	@[ "$$(git rev-parse --abbrev-ref HEAD)" = "main" ] || (echo "❌ Not on main branch, current: $$(git rev-parse --abbrev-ref HEAD)" && exit 1)
	@echo "3. Checking for unpushed commits..."
	@git fetch origin main
	@[ "$$(git rev-list HEAD...origin/main --count)" = "0" ] || (echo "❌ Local commits not pushed to origin/main" && exit 1)
	@echo "4. Running full CI pipeline..."
	@make ci
	@echo "✅ All pre-release checks passed"

release-tag: release-check ## Create and push release tag
	@if [ -z "$(VERSION)" ]; then \
		echo "❌ Usage: make release-tag VERSION=v1.0.0"; \
		echo ""; \
		echo "📋 Version format examples:"; \
		echo "  - Major release: v1.0.0"; \
		echo "  - Minor release: v1.1.0"; \
		echo "  - Patch release: v1.0.1"; \
		echo "  - Pre-release: v1.0.0-alpha.1, v1.0.0-beta.1, v1.0.0-rc.1"; \
		exit 1; \
	fi
	@echo "🏷️ Creating release tag $(VERSION)..."
	@git tag -a "$(VERSION)" -m "🔥 Release $(VERSION)"
	@git push origin "$(VERSION)"
	@echo ""
	@echo "✅ Release $(VERSION) tagged and pushed!"
	@echo "📦 GitHub Actions: https://github.com/BurnDevice/BurnDevice/actions"
	@echo "📋 Release page: https://github.com/BurnDevice/BurnDevice/releases"
	@echo ""
	@echo "⏰ Expected completion: 5-10 minutes"
	@echo "🎯 Release artifacts will include:"
	@echo "   - Multi-platform binaries (Linux, macOS, Windows)"
	@echo "   - Docker images (ghcr.io/burndevice/burndevice:$(VERSION))"
	@echo "   - Source code archives"

release-local: clean build-all ## Create local release artifacts for testing
	@echo "📦 Creating local release artifacts..."
	@mkdir -p release
	@cp bin/* release/ 2>/dev/null || true
	@cd release && \
		for file in burndevice-*; do \
			if [[ "$$file" == *".exe" ]]; then \
				zip "$${file%.*}.zip" "$$file" && rm "$$file"; \
			else \
				tar -czf "$${file}.tar.gz" "$$file" && rm "$$file"; \
			fi; \
		done
	@echo "✅ Local release artifacts created in release/"
	@ls -la release/

# Version helpers
version-current: ## Show current version
	@echo "Current version: $$(git describe --tags --abbrev=0 2>/dev/null || echo 'No tags found')"

version-patch: ## Suggest next patch version
	@echo "Current: $$(git describe --tags --abbrev=0 2>/dev/null || echo 'v0.0.0')"
	@echo "Next patch: $$(git describe --tags --abbrev=0 2>/dev/null | sed 's/v//' | awk -F. '{print "v" $$1 "." $$2 "." $$3+1}' || echo 'v0.0.1')"

version-minor: ## Suggest next minor version  
	@echo "Current: $$(git describe --tags --abbrev=0 2>/dev/null || echo 'v0.0.0')"
	@echo "Next minor: $$(git describe --tags --abbrev=0 2>/dev/null | sed 's/v//' | awk -F. '{print "v" $$1 "." $$2+1 ".0"}' || echo 'v0.1.0')"

version-major: ## Suggest next major version
	@echo "Current: $$(git describe --tags --abbrev=0 2>/dev/null || echo 'v0.0.0')"
	@echo "Next major: $$(git describe --tags --abbrev=0 2>/dev/null | sed 's/v//' | awk -F. '{print "v" $$1+1 ".0.0"}' || echo 'v1.0.0')"

# GoReleaser support (optional, for future use)
goreleaser-check: ## Check GoReleaser configuration
	@if command -v goreleaser >/dev/null 2>&1; then \
		goreleaser check; \
	else \
		echo "⚠️  GoReleaser not installed, skipping check"; \
	fi

goreleaser-snapshot: ## Build snapshot release with GoReleaser
	@if command -v goreleaser >/dev/null 2>&1; then \
		goreleaser release --snapshot --clean; \
	else \
		echo "⚠️  GoReleaser not installed, use 'make release-local' instead"; \
	fi

# Generate example scenarios
generate-example: build ## Generate example attack scenarios
	mkdir -p examples
	./bin/$(BINARY_NAME) generate --target "Linux test server" --severity MEDIUM --output examples/

# Validation
validate-config: build ## Validate configuration file
	./bin/$(BINARY_NAME) validate config config.example.yaml

validate-all: validate-config proto-lint ## Run all validation checks

# Installation
install: build ## Install the binary
	sudo cp bin/$(BINARY_NAME) /usr/local/bin/

uninstall: ## Uninstall the binary
	sudo rm -f /usr/local/bin/$(BINARY_NAME)

# Development helpers
dev-setup: deps proto ## Setup development environment
	@echo "✅ Development environment ready"
	@echo "Run 'make run-dev' to start the server"

# Watch mode (requires entr)
watch: ## Watch files and rebuild on changes
	@if command -v entr >/dev/null 2>&1; then \
		find . -name "*.go" -o -name "*.proto" | entr -r make run-dev; \
	else \
		echo "⚠️  entr not installed. Install with: brew install entr (macOS) or apt-get install entr (Ubuntu)"; \
	fi

# Performance profiling
profile-cpu: build ## Run CPU profiling
	go test -cpuprofile=cpu.prof -bench=. ./...

profile-mem: build ## Run memory profiling
	go test -memprofile=mem.prof -bench=. ./...

# Documentation
docs: ## Generate documentation
	@echo "📚 Generating documentation..."
	go doc -all ./... > docs/api.txt
	@echo "✅ Documentation generated in docs/api.txt"

# Version info
version: ## Show version information
	@echo "Binary: $(BINARY_NAME)"
	@echo "Version: $(VERSION)"
	@echo "Commit: $(COMMIT)"
	@echo "Date: $(DATE)"

# Health check
health-check: build ## Run comprehensive health check
	@echo "🏥 Running health check..."
	@echo "1. Testing build..."
	@make build >/dev/null 2>&1 && echo "✅ Build: OK" || echo "❌ Build: FAILED"
	@echo "2. Testing format..."
	@make fmt-check >/dev/null 2>&1 && echo "✅ Format: OK" || echo "❌ Format: FAILED"
	@echo "3. Testing vet..."
	@make vet >/dev/null 2>&1 && echo "✅ Vet: OK" || echo "❌ Vet: FAILED"
	@echo "4. Testing short tests..."
	@make test-short >/dev/null 2>&1 && echo "✅ Tests: OK" || echo "❌ Tests: FAILED"
	@echo "🏥 Health check complete"