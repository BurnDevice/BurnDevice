# 🤝 贡献指南

感谢您对 BurnDevice 项目的关注！我们欢迎所有形式的贡献。

## ⚠️ 重要提醒

**BurnDevice 是一个破坏性测试工具，仅限用于授权的测试环境。所有贡献必须符合安全和道德标准。**

## 🚀 快速开始

### 环境准备

1. **Go 环境**：需要 Go 1.24+
2. **Nix 环境**（推荐）：
   ```bash
   nix develop
   ```
3. **传统环境**：
   ```bash
   # 安装依赖
   go mod download
   
   # 安装 buf
   curl -sSL "https://github.com/bufbuild/buf/releases/latest/download/buf-$(uname -s)-$(uname -m)" -o /usr/local/bin/buf
   chmod +x /usr/local/bin/buf
   ```

### 开发流程

1. **Fork 项目**
2. **创建分支**：
   ```bash
   git checkout -b feature/your-feature-name
   ```
3. **开发和测试**：
   ```bash
   # 生成代码
   buf generate
   
   # 运行测试
   make test
   
   # 代码检查
   make lint
   ```
4. **提交更改**：
   ```bash
   git commit -m "feat: add your feature description"
   ```
5. **推送分支**：
   ```bash
   git push origin feature/your-feature-name
   ```
6. **创建 Pull Request**

## 📋 贡献类型

### 🐛 Bug 修复
- 在 Issues 中报告 bug
- 提供复现步骤
- 包含系统信息
- 提交修复的 PR

### ✨ 新功能
- 先在 Issues 中讨论
- 确保功能符合项目目标
- 添加相应的测试
- 更新文档

### 📚 文档改进
- 修正错误
- 添加示例
- 改进可读性
- 翻译支持

### 🧪 测试改进
- 增加测试覆盖率
- 添加边界测试
- 性能测试
- 安全测试

## 🔧 开发标准

### 代码风格
- 遵循 Go 官方风格指南
- 使用 `gofmt` 格式化代码
- 通过 `golangci-lint` 检查
- 保持代码简洁和可读

### 提交规范
使用 [Conventional Commits](https://www.conventionalcommits.org/) 格式：

```
<type>[optional scope]: <description>

[optional body]

[optional footer(s)]
```

**类型**：
- `feat`: 新功能
- `fix`: Bug 修复
- `docs`: 文档更新
- `style`: 代码格式（不影响功能）
- `refactor`: 重构
- `test`: 测试相关
- `chore`: 构建过程或辅助工具的变动
- `perf`: 性能优化
- `ci`: CI/CD 相关
- `security`: 安全相关

**示例**：
```
feat(engine): add memory exhaustion attack type

Add new attack type for testing system memory limits.
Includes safety checks and configurable severity levels.

Closes #123
```

### 测试要求
- 新功能必须包含单元测试
- 测试覆盖率不能下降
- 包含集成测试（如适用）
- 测试应该快速和可靠

### 文档要求
- 公共 API 必须有文档注释
- 复杂逻辑需要内联注释
- 更新 README（如需要）
- 添加使用示例

## 🛡️ 安全指南

### 安全审查
所有贡献都将经过安全审查：
- 不得包含真正的恶意代码
- 必须有适当的安全控制
- 遵循最小权限原则
- 包含安全测试

### 禁止内容
- 无限制的破坏性代码
- 绕过安全机制的代码
- 针对生产环境的攻击
- 恶意软件或后门

### 安全报告
如发现安全漏洞，请通过以下方式报告：
- 发送邮件到：foxxcn.web@gmail.com
- 不要在公开 Issue 中讨论安全问题
- 我们将在 48 小时内响应

## 🧪 测试指南

### 运行测试
```bash
# 所有测试
make test

# 竞态检测
make test-race

# 覆盖率报告
make test-coverage

# 基准测试
make benchmark
```

### 测试分类
- **单元测试**：测试单个函数/方法
- **集成测试**：测试组件间交互
- **端到端测试**：测试完整工作流
- **性能测试**：测试性能指标

### 测试最佳实践
- 使用表驱动测试
- 测试边界条件
- 包含错误情况
- 使用有意义的测试名称
- 避免测试间依赖

## 📦 发布流程

### 版本控制
使用 [Semantic Versioning](https://semver.org/)：
- `MAJOR.MINOR.PATCH`
- `MAJOR`：不兼容的 API 变更
- `MINOR`：新功能（向后兼容）
- `PATCH`：Bug 修复（向后兼容）

### 发布步骤
1. 更新版本号
2. 更新 CHANGELOG
3. 创建 Git 标签
4. GitHub Actions 自动构建和发布

## 🤝 社区准则

### 行为准则
- 尊重所有参与者
- 使用包容性语言
- 专注于建设性反馈
- 帮助新贡献者

### 沟通渠道
- **Issues**：Bug 报告和功能请求
- **Discussions**：一般讨论和问答
- **Pull Requests**：代码审查和讨论

## 📞 获取帮助

如果您需要帮助：
1. 查看现有的 Issues 和 Discussions
2. 阅读项目文档
3. 创建新的 Issue 或 Discussion
4. 联系维护者

## 🙏 致谢

感谢所有贡献者的努力！您的贡献使 BurnDevice 变得更好。

---

**记住：我们共同致力于创建一个安全、有用、专业的测试工具。** 🔥 