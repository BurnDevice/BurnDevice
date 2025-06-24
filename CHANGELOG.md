# 更新日志

本文档记录了BurnDevice项目的所有重要变更。

格式基于 [Keep a Changelog](https://keepachangelog.com/zh-CN/1.0.0/)，
项目遵循 [语义化版本控制](https://semver.org/lang/zh-CN/)。

## [未发布]

### 新增
- 增强的发布管理系统，支持自动化版本控制
- GoReleaser集成支持，提供更专业的发布流程
- 完善的发布前检查机制

### 变更
- 改进Makefile，添加release-tag、version-*等发布相关目标
- 优化Nix开发环境，添加goreleaser工具支持

## [v0.1.0] - 待发布

### 新增
- 🔥 BurnDevice核心功能
  - 文件系统破坏测试 (FILE_DELETION)
  - 系统服务中断测试 (SERVICE_TERMINATION)
  - 内存和磁盘耗尽测试 (MEMORY_EXHAUSTION, DISK_FILL)
  - 网络中断测试 (NETWORK_DISRUPTION)
  - 引导损坏测试 (BOOT_CORRUPTION)
  - 内核崩溃测试 (KERNEL_PANIC)

- 🏗️ 完整的架构设计
  - gRPC服务架构，支持高性能通信
  - Protocol Buffers API定义
  - 模块化内部结构 (engine, server, cli, config)

- 🤖 AI驱动功能
  - DeepSeek AI集成，智能生成攻击场景
  - 基于目标环境的自适应测试策略
  - 可配置的AI模型和参数

- 🔒 安全控制机制
  - 目标路径白名单/黑名单过滤
  - 操作严重级别限制 (LOW/MEDIUM/HIGH/CRITICAL)
  - 强制确认机制，防止意外执行
  - 详细审计日志记录
  - 操作备份和恢复功能

- 🐳 容器化支持
  - 多阶段Docker构建，优化镜像大小
  - 多架构支持 (linux/amd64, linux/arm64)
  - 非root用户运行，增强安全性
  - 健康检查和优雅关闭

- 📦 发布和分发
  - 多平台二进制构建 (Linux, macOS, Windows)
  - GitHub Actions自动化CI/CD
  - GitHub Container Registry集成
  - Docker Hub支持 (可选)
  - Homebrew formula生成 (可选)

- 🧪 测试和质量保证
  - 全面的单元测试覆盖
  - 竞态条件检测
  - 代码覆盖率报告
  - 安全扫描 (gosec, govulncheck)
  - 代码质量检查 (golangci-lint)

### 技术特性
- **语言**: Go 1.24+ 支持
- **构建系统**: Make + Nix Flake
- **API**: gRPC + Protocol Buffers
- **配置**: YAML配置文件
- **日志**: 结构化JSON日志
- **监控**: 实时操作指标和事件流

### 安全功能
- 操作前强制确认机制
- 目标验证和路径过滤
- 严重级别控制和限制
- 完整的操作审计日志
- 自动备份重要文件
- 安全的错误处理和信息泄露防护

### 开发体验
- Nix Flake提供可重现的开发环境
- 完整的Makefile任务集合
- 热重载开发模式
- 性能分析和基准测试工具
- 自动化代码格式化和检查

---

## 📋 发布说明

### 版本命名规则
- `v1.0.0` - 稳定版本，向后兼容
- `v1.1.0` - 新功能版本，向后兼容  
- `v1.0.1` - 补丁版本，bug修复
- `v1.0.0-alpha.1` - 预发布版本

### 发布流程
```bash
# 检查发布准备
make release-check

# 查看版本建议
make version-patch

# 执行发布
make release-tag VERSION=v1.0.0
```

### 支持的平台
- **Linux**: x86_64, ARM64
- **macOS**: Intel, Apple Silicon
- **Windows**: x86_64
- **Docker**: linux/amd64, linux/arm64

---

⚠️ **重要提醒**: 本工具仅用于授权的测试环境中，绝不可在生产环境使用！使用前请确保遵循相关法律法规和安全政策。 