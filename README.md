# 🔥 BurnDevice - 设备破坏性测试工具

[![Go Version](https://img.shields.io/badge/Go-1.24+-blue.svg)](https://golang.org)
[![Test Coverage](https://img.shields.io/badge/Coverage-53%25-yellow.svg)](coverage.out)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Security](https://img.shields.io/badge/Security-Research%20Only-red.svg)](docs/SECURITY.md)

BurnDevice 是一个专为**授权测试环境**设计的破坏性测试工具，用于评估系统的健壮性和恢复能力。

## ⚠️ 重要安全警告

**此工具具有极强的破坏性，仅限用于以下场景：**

- ✅ 授权的测试环境
- ✅ 个人拥有的测试设备
- ✅ 安全研究和教育目的
- ✅ 系统韧性测试

**严禁在以下场景使用：**

- ❌ 生产环境
- ❌ 他人设备
- ❌ 未经授权的系统
- ❌ 恶意攻击

**使用本工具即表示您同意承担所有风险和责任**

## 🚀 特性

- **多种破坏模式**: 文件删除、服务中断、内存耗尽、磁盘填满等
- **AI 驱动**: 集成 DeepSeek AI 生成智能攻击场景
- **安全控制**: 多层安全验证和目标限制机制
- **实时监控**: gRPC 流式接口实时监控破坏进度
- **可恢复性**: 支持不同严重级别的可恢复操作
- **审计日志**: 完整的操作记录和安全审计

## 🛠️ 技术栈

- **语言**: Go 1.24+
- **框架**: gRPC + Protocol Buffers
- **环境**: Nix Flakes
- **AI**: DeepSeek API 集成
- **配置**: Viper + YAML
- **日志**: Logrus 结构化日志

## 📦 安装和使用

### 1. 使用 Nix Flakes (推荐)

```bash
# 进入开发环境
nix develop

# 生成 Protocol Buffers 代码
buf generate

# 构建项目
make build

# 查看帮助
make help
```

### 2. 传统 Go 环境

```bash
# 确保安装了 Go 1.24+ 和 protoc
go version
protoc --version

# 克隆项目
git clone https://github.com/BurnDevice/BurnDevice.git
cd BurnDevice

# 安装依赖
go mod download

# 生成代码和构建
make dev-setup
make build
```

### 3. 配置

```bash
# 复制示例配置
cp config.example.yaml config.yaml

# 编辑配置文件
nvim config.yaml

# 设置环境变量
export BURNDEVICE_AI_API_KEY="your-deepseek-api-key"
```

### 4. 运行

```bash
# 启动服务器
./bin/burndevice server --config config.yaml

# 在另一个终端中使用客户端
./bin/burndevice client --help
```

## 🎯 使用示例

### 生成 AI 攻击场景

```bash
./bin/burndevice client generate-scenario \
  --target "Linux测试服务器 - Ubuntu 22.04, 4GB RAM, 100GB磁盘" \
  --max-severity MEDIUM \
  --server localhost:8080
```

### 执行文件删除测试

```bash
# 安全删除（可恢复）
./bin/burndevice client execute \
  --type FILE_DELETION \
  --targets "/tmp/test_file.txt" \
  --severity LOW \
  --confirm

# 查看系统信息
./bin/burndevice client system-info
```

### 内存耗尽测试

```bash
# 低强度内存压测
./bin/burndevice client execute \
  --type MEMORY_EXHAUSTION \
  --severity LOW \
  --confirm
```

## 🔧 配置选项

### 安全配置

```yaml
security:
  require_confirmation: true      # 需要明确确认
  max_severity: "MEDIUM"         # 最大严重级别
  enable_safe_mode: true         # 启用安全模式
  audit_log: true               # 启用审计日志
  
  # 白名单：允许的目标路径
  allowed_targets:
    - "/tmp/burndevice_test"
    - "/home/user/test"
  
  # 黑名单：禁止的目标路径
  blocked_targets:
    - "/"
    - "/bin"
    - "/usr"
    - "/etc"
```

### AI 配置

```yaml
ai:
  provider: "deepseek"
  api_key: "${BURNDEVICE_AI_API_KEY}"
  base_url: "https://api.deepseek.com"
  model: "deepseek-chat"
  max_tokens: 4096
  temperature: 0.7
```

## 🛡️ 安全机制

1. **多重确认**: 要求明确的破坏确认
2. **路径限制**: 白名单/黑名单机制
3. **严重级别**: 限制最大破坏级别
4. **安全模式**: 仿真而非真实执行
5. **审计日志**: 记录所有操作
6. **权限检查**: 验证操作权限

## 📋 破坏类型

| 类型 | 描述 | 严重级别 | 可恢复性 |
|------|------|----------|----------|
| FILE_DELETION | 文件删除攻击 | LOW-CRITICAL | 视级别而定 |
| SERVICE_TERMINATION | 服务终止攻击 | LOW-HIGH | 高 |
| MEMORY_EXHAUSTION | 内存耗尽攻击 | LOW-HIGH | 高 |
| DISK_FILL | 磁盘填满攻击 | LOW-HIGH | 中 |
| NETWORK_DISRUPTION | 网络中断攻击 | MEDIUM-HIGH | 高 |
| BOOT_CORRUPTION | 引导损坏攻击 | HIGH-CRITICAL | 低 |
| KERNEL_PANIC | 内核崩溃攻击 | CRITICAL | 低 |

## 🧪 开发和测试

```bash
# 运行测试
make test

# 运行竞态检测
make test-race

# 生成覆盖率报告
make test-coverage

# 代码质量检查
make lint
make vet

# 安全检查
make security-check
```

## 🐳 Docker 支持

```bash
# 构建 Docker 镜像
make docker-build

# 运行容器
make docker-run
```

## 📚 API 文档

### gRPC 服务

```protobuf
service BurnDeviceService {
  rpc ExecuteDestruction(DestructionRequest) returns (DestructionResponse);
  rpc GetSystemInfo(SystemInfoRequest) returns (SystemInfoResponse);
  rpc GenerateAttackScenario(AttackScenarioRequest) returns (AttackScenarioResponse);
  rpc StreamDestruction(DestructionRequest) returns (stream DestructionEvent);
}
```

详细的 API 文档请参考 [burndevice/v1/service.proto](burndevice/v1/service.proto)

## 🤝 贡献

1. Fork 项目
2. 创建功能分支 (`git checkout -b feature/amazing-feature`)
3. 提交更改 (`git commit -m 'Add amazing feature'`)
4. 推送到分支 (`git push origin feature/amazing-feature`)
5. 开启 Pull Request

**注意**: 所有贡献必须通过安全审查，不得包含真正的恶意代码。

## 📄 许可证

本项目采用 MIT 许可证 - 详见 [LICENSE](LICENSE) 文件

## ⚖️ 法律声明

本工具仅用于合法的安全研究和测试目的。使用者有责任确保其使用符合当地法律法规。作者不对任何误用或损害承担责任。

## 🆘 支持

- 📖 [文档](docs/)
- 🐛 [问题报告](https://github.com/BurnDevice/BurnDevice/issues)
- 💬 [讨论](https://github.com/BurnDevice/BurnDevice/discussions)

---

**再次提醒：此工具具有破坏性，请谨慎使用！** 🔥 