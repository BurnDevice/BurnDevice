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

# Release management - 最佳实践版本
.PHONY: release-check release-build release-tag release-upload release-test

# 发布前检查
release-check: ## 发布前检查
	@echo "🔍 发布前检查..."
	@git diff --quiet || (echo "❌ 工作目录不干净" && exit 1)
	@[ "$$(git rev-parse --abbrev-ref HEAD)" = "main" ] || (echo "❌ 请在main分支发布" && exit 1)
	@make test-short
	@echo "✅ 检查通过"

# 构建发布包
release-build: clean ## 构建发布包
	@if [ -z "$(VERSION)" ]; then \
		echo "❌ 请指定版本: make release-build VERSION=v1.0.0"; \
		exit 1; \
	fi
	@echo "📦 构建 $(VERSION) 发布包..."
	@make build-all
	@mkdir -p release
	@echo "🗜️ 创建压缩包..."
	@cd bin && \
		tar -czf ../release/burndevice-$(VERSION)-linux-amd64.tar.gz burndevice-linux-amd64 && \
		tar -czf ../release/burndevice-$(VERSION)-linux-arm64.tar.gz burndevice-linux-arm64 && \
		tar -czf ../release/burndevice-$(VERSION)-darwin-amd64.tar.gz burndevice-darwin-amd64 && \
		tar -czf ../release/burndevice-$(VERSION)-darwin-arm64.tar.gz burndevice-darwin-arm64 && \
		tar -czf ../release/burndevice-$(VERSION)-windows-amd64.tar.gz burndevice-windows-amd64.exe
	@echo "✅ 发布包构建完成:"
	@ls -la release/

# 创建并推送标签（触发GitHub Actions）
release-tag: release-check ## 创建并推送发布标签
	@if [ -z "$(VERSION)" ]; then \
		echo "❌ 请指定版本: make release-tag VERSION=v1.0.0"; \
		exit 1; \
	fi
	@echo "🏷️ 创建发布标签 $(VERSION)..."
	@if git tag -l | grep -q "^$(VERSION)$$"; then \
		echo "❌ 标签 $(VERSION) 已存在"; \
		exit 1; \
	fi
	@git tag -a $(VERSION) -m "🔥 Release $(VERSION)"
	@echo "📤 推送标签到远程仓库..."
	@git push origin $(VERSION)
	@echo "✅ 标签 $(VERSION) 已创建并推送！"
	@echo "⏰ GitHub Actions 正在创建 Release 页面..."

# 等待GitHub Actions创建release并上传本地构建的包
release-upload: release-build release-tag ## 上传发布包到GitHub Release
	@if [ -z "$(VERSION)" ]; then \
		echo "❌ 请指定版本: make release-upload VERSION=v1.0.0"; \
		exit 1; \
	fi
	@echo "⏳ 等待 GitHub Actions 创建 Release..."
	@echo "💡 提示：如果等待过程中断，可以手动运行："
	@echo "   gh release upload $(VERSION) release/*.tar.gz --clobber"
	@echo ""
	@WAIT_COUNT=0; \
	MAX_WAIT=60; \
	RETRY_INTERVAL=5; \
	SUCCESS=false; \
	trap 'echo ""; echo "⚠️  等待被中断，但发布包已构建完成"; echo "💡 手动上传命令："; echo "   gh release upload $(VERSION) release/*.tar.gz --clobber"; exit 130' INT TERM; \
	while [ $$WAIT_COUNT -lt $$MAX_WAIT ]; do \
		if gh release view $(VERSION) --json state --jq '.state' 2>/dev/null | grep -q "published"; then \
			echo "✅ Release $(VERSION) 已创建并发布！"; \
			SUCCESS=true; \
			break; \
		elif gh release view $(VERSION) >/dev/null 2>&1; then \
			echo "✅ Release $(VERSION) 已创建（可能仍在处理中）"; \
			SUCCESS=true; \
			break; \
		fi; \
		WAIT_COUNT=$$((WAIT_COUNT + 1)); \
		REMAINING=$$((MAX_WAIT - WAIT_COUNT)); \
		printf "等待中... [%d/%d] (剩余 %d 秒)\r" $$WAIT_COUNT $$MAX_WAIT $$((REMAINING * RETRY_INTERVAL)); \
		if [ $$((WAIT_COUNT % 12)) -eq 0 ] && [ $$WAIT_COUNT -gt 0 ]; then \
			echo ""; \
			echo "📋 检查 GitHub Actions 状态: https://github.com/BurnDevice/BurnDevice/actions"; \
		fi; \
		sleep $$RETRY_INTERVAL; \
	done; \
	echo ""; \
	if [ "$$SUCCESS" = "false" ]; then \
		echo "⏰ GitHub Actions 创建 Release 超时（等待了 $$((MAX_WAIT * RETRY_INTERVAL)) 秒）"; \
		echo "📋 请检查 Actions 状态: https://github.com/BurnDevice/BurnDevice/actions"; \
		echo ""; \
		echo "💡 如果 Release 已创建，可以手动上传发布包："; \
		echo "   gh release upload $(VERSION) release/*.tar.gz --clobber"; \
		echo ""; \
		read -p "是否继续尝试上传发布包？(y/N): " confirm; \
		if [ "$$confirm" != "y" ] && [ "$$confirm" != "Y" ]; then \
			echo "❌ 发布流程中止"; \
			exit 1; \
		fi; \
	fi
	@echo "📦 上传本地构建的发布包..."
	@if gh release upload $(VERSION) release/*.tar.gz --clobber; then \
		echo ""; \
		echo "🎉 发布完成!"; \
		echo "📋 Release页面: https://github.com/BurnDevice/BurnDevice/releases/tag/$(VERSION)"; \
		echo "🔍 验证安装: curl -fsSL https://github.com/BurnDevice/BurnDevice/releases/download/$(VERSION)/burndevice-$(VERSION)-linux-amd64.tar.gz | tar -xz"; \
	else \
		echo ""; \
		echo "❌ 上传发布包失败"; \
		echo "💡 请手动上传："; \
		echo "   gh release upload $(VERSION) release/*.tar.gz --clobber"; \
		exit 1; \
	fi

# 快速发布（不等待GitHub Actions，直接尝试上传）
release-quick: release-build release-tag ## 快速发布（跳过等待，直接尝试上传）
	@if [ -z "$(VERSION)" ]; then \
		echo "❌ 请指定版本: make release-quick VERSION=v1.0.0"; \
		exit 1; \
	fi
	@echo "🚀 快速发布模式：跳过等待，直接尝试上传..."
	@sleep 10  # 给GitHub Actions一点时间
	@echo "📦 尝试上传发布包..."
	@RETRY_COUNT=0; \
	MAX_RETRY=6; \
	while [ $$RETRY_COUNT -lt $$MAX_RETRY ]; do \
		if gh release upload $(VERSION) release/*.tar.gz --clobber 2>/dev/null; then \
			echo "✅ 发布包上传成功！"; \
			echo "🎉 发布完成!"; \
			echo "📋 Release页面: https://github.com/BurnDevice/BurnDevice/releases/tag/$(VERSION)"; \
			exit 0; \
		fi; \
		RETRY_COUNT=$$((RETRY_COUNT + 1)); \
		if [ $$RETRY_COUNT -lt $$MAX_RETRY ]; then \
			echo "⏳ Release尚未创建，等待 10 秒后重试... ($$RETRY_COUNT/$$MAX_RETRY)"; \
			sleep 10; \
		fi; \
	done; \
	echo "❌ 快速上传失败，Release可能尚未创建"; \
	echo "📋 请检查 GitHub Actions: https://github.com/BurnDevice/BurnDevice/actions"; \
	echo "💡 稍后可手动上传："; \
	echo "   gh release upload $(VERSION) release/*.tar.gz --clobber"

# 仅上传发布包（假设Release已存在）
release-assets-only: ## 仅上传发布包到已存在的Release
	@if [ -z "$(VERSION)" ]; then \
		echo "❌ 请指定版本: make release-assets-only VERSION=v1.0.0"; \
		exit 1; \
	fi
	@if [ ! -d "release" ]; then \
		echo "❌ 发布包目录不存在，请先运行: make release-build VERSION=$(VERSION)"; \
		exit 1; \
	fi
	@echo "📦 上传发布包到 Release $(VERSION)..."
	@if gh release upload $(VERSION) release/*.tar.gz --clobber; then \
		echo "✅ 发布包上传成功！"; \
		echo "📋 Release页面: https://github.com/BurnDevice/BurnDevice/releases/tag/$(VERSION)"; \
	else \
		echo "❌ 上传失败，请检查 Release 是否存在"; \
		echo "📋 Release列表: gh release list"; \
		exit 1; \
	fi

# 一键发布 (推荐使用)
release: ## 一键发布 (使用方法: make release VERSION=v1.0.0)
	@if [ -z "$(VERSION)" ]; then \
		echo ""; \
		echo "🚀 BurnDevice 发布工具"; \
		echo ""; \
		echo "使用方法:"; \
		echo "  make release VERSION=v1.0.0        # 标准发布（等待GitHub Actions）"; \
		echo "  make release-quick VERSION=v1.0.0  # 快速发布（跳过等待）"; \
		echo "  make release-assets-only VERSION=v1.0.0  # 仅上传发布包"; \
		echo ""; \
		echo "版本格式:"; \
		echo "  主版本: v1.0.0"; \
		echo "  次版本: v1.1.0"; \
		echo "  补丁版本: v1.0.1"; \
		echo "  预发布: v1.0.0-beta.1"; \
		echo ""; \
		echo "当前版本: $$(git describe --tags --abbrev=0 2>/dev/null || echo '未找到标签')"; \
		echo ""; \
		echo "发布选项说明:"; \
		echo "  release       - 完整流程，等待GitHub Actions创建Release"; \
		echo "  release-quick - 快速模式，跳过等待直接尝试上传"; \
		echo "  release-assets-only - 仅上传发布包到已存在的Release"; \
		echo ""; \
		echo "标准发布流程:"; \
		echo "  1. 发布前检查（代码格式、测试等）"; \
		echo "  2. 构建多平台二进制文件"; \
		echo "  3. 创建并推送 Git 标签"; \
		echo "  4. 等待 GitHub Actions 创建 Release"; \
		echo "  5. 上传本地构建的发布包"; \
		echo ""; \
		echo "💡 如果标准流程中断，可以使用:"; \
		echo "   make release-assets-only VERSION=v1.0.0"; \
		echo ""; \
		exit 1; \
	fi
	@make release-upload VERSION=$(VERSION)

# 版本信息
version-current: ## 显示当前版本
	@echo "当前版本: $$(git describe --tags --abbrev=0 2>/dev/null || echo '未找到标签')"

# 本地测试发布包
release-test: release-build ## 本地测试发布包
	@echo "🧪 测试发布包..."
	@cd /tmp && \
		tar -xzf $(PWD)/release/burndevice-$(VERSION)-linux-amd64.tar.gz && \
		./burndevice-linux-amd64 --version && \
		rm burndevice-linux-amd64
	@echo "✅ 发布包测试通过"

# 删除标签（用于重新发布）
release-delete-tag: ## 删除指定标签 (使用方法: make release-delete-tag VERSION=v1.0.0)
	@if [ -z "$(VERSION)" ]; then \
		echo "❌ 请指定版本: make release-delete-tag VERSION=v1.0.0"; \
		exit 1; \
	fi
	@echo "🗑️ 删除标签 $(VERSION)..."
	@git tag -d $(VERSION) || true
	@git push origin :refs/tags/$(VERSION) || true
	@echo "✅ 标签 $(VERSION) 已删除"

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