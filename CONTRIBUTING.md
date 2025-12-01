# 贡献指南

感谢您对 ES Serverless Platform 项目的关注！我们欢迎各种形式的贡献。

---

## 📖 目录

- [开发环境搭建](#开发环境搭建)
- [项目结构](#项目结构)
- [开发流程](#开发流程)
- [代码规范](#代码规范)
- [测试要求](#测试要求)
- [提交规范](#提交规范)
- [Pull Request 流程](#pull-request-流程)

---

## 🛠️ 开发环境搭建

### 必需工具

1. **Go 1.21+** - 控制平面开发
2. **Java 11+** 和 **Gradle** - ES 插件开发
3. **Docker Desktop** with Kubernetes - 本地测试
4. **kubectl** - Kubernetes 命令行工具
5. **Git** - 版本控制

### 克隆项目

```bash
git clone https://github.com/your-org/es-paas.git
cd es-paas/es
```

### 安装依赖

#### 控制平面

```bash
cd src/control-plane
go mod download
go build -o manager .
```

#### ES 插件

```bash
cd src/es-plugin
./gradlew build
```

### 启动开发环境

```bash
# 启动本地 Kubernetes 集群
# Docker Desktop 需启用 Kubernetes

# 部署系统
./scripts/deploy/deploy.sh install

# 启动控制平面（本地开发）
cd src/control-plane
./manager
```

详细说明请查看 [开发环境搭建文档](docs/development/setup.md)。

---

## 📂 项目结构

请先阅读 [PROJECT_STRUCTURE.md](PROJECT_STRUCTURE.md) 了解项目组织方式。

**关键目录**：
```
src/control-plane/    # Go 控制平面服务
src/es-plugin/        # Java ES IVF 插件
src/frontend/         # 前端管理界面
deployments/          # 部署配置（Terraform、Helm、K8s）
scripts/              # 工具脚本（deploy、build、ops、dev）
docs/                 # 文档中心（按类型分类）
tests/                # 测试代码
```

---

## 🔄 开发流程

### 1. 创建功能分支

```bash
git checkout -b feature/your-feature-name
# 或
git checkout -b fix/your-bug-fix
```

分支命名规范：
- `feature/` - 新功能
- `fix/` - Bug 修复
- `docs/` - 文档更新
- `refactor/` - 代码重构
- `test/` - 测试相关

### 2. 进行开发

根据您修改的模块，遵循相应的开发指南：

- **控制平面**: [控制平面开发](docs/development/control-plane.md)
- **ES 插件**: [ES 插件开发](docs/development/es-plugin.md)
- **前端**: [前端开发](docs/development/frontend.md)

### 3. 编写测试

所有新功能和 Bug 修复都应包含测试：

```bash
# 单元测试
tests/unit/

# 集成测试
tests/integration/

# 端到端测试
tests/e2e/
```

### 4. 运行测试

```bash
# Go 控制平面测试
cd src/control-plane
go test ./...

# ES 插件测试
cd src/es-plugin
./gradlew test

# 集成测试
cd tests
./run-integration-tests.sh
```

### 5. 提交代码

遵循 [提交规范](#提交规范)。

---

## 📝 代码规范

### Go 代码规范

- 遵循 [Effective Go](https://golang.org/doc/effective_go.html)
- 使用 `gofmt` 格式化代码
- 使用 `golint` 检查代码质量
- 错误处理：显式处理所有错误，不要忽略
- 注释：公开的函数、结构体必须有注释

**示例**：

```go
// GetClusterStatus retrieves the current status of a cluster
// Returns ClusterStatus and error if cluster not found
func GetClusterStatus(namespace string) (*ClusterStatus, error) {
    if namespace == "" {
        return nil, fmt.Errorf("namespace cannot be empty")
    }
    // ...
}
```

### Java 代码规范（ES 插件）

- 遵循 [Google Java Style Guide](https://google.github.io/styleguide/javaguide.html)
- 使用 4 空格缩进
- 类、方法必须有 Javadoc 注释
- 异常处理：捕获具体异常，避免空 catch

**示例**：

```java
/**
 * Inverted File Index for vector search
 *
 * @param dimension vector dimension
 * @param nlist number of clusters
 */
public class InvertedFileIndex {
    // ...
}
```

### 前端代码规范

- 使用 ES6+ 语法
- 函数名使用驼峰命名
- 常量使用大写字母
- 添加必要的注释

---

## 🧪 测试要求

### 单元测试

- 所有新功能必须有单元测试
- 代码覆盖率 > 80%
- 测试文件与源文件同目录，命名为 `*_test.go` 或 `*Test.java`

### 集成测试

对于控制平面和 ES 插件的集成功能：
- 测试完整的业务流程
- 包含成功和失败场景
- 测试多租户隔离

### 端到端测试

- 测试完整的用户工作流
- 包括 UI 操作和 API 调用
- 验证系统各组件的集成

---

## 📋 提交规范

使用 [Conventional Commits](https://www.conventionalcommits.org/) 规范：

```
<type>(<scope>): <subject>

<body>

<footer>
```

### Type

- `feat`: 新功能
- `fix`: Bug 修复
- `docs`: 文档更新
- `style`: 代码格式调整（不影响功能）
- `refactor`: 代码重构
- `test`: 测试相关
- `chore`: 构建或辅助工具变动

### Scope

- `control-plane`: 控制平面
- `es-plugin`: ES 插件
- `frontend`: 前端
- `deployment`: 部署配置
- `docs`: 文档

### 示例

```bash
git commit -m "feat(control-plane): add GPU node scheduling support"

git commit -m "fix(es-plugin): fix IVF index memory leak

The IVF index was not releasing memory properly after rebuild.
This commit adds proper cleanup in the destructor.

Closes #123"
```

---

## 🔀 Pull Request 流程

### 1. 确保代码质量

在提交 PR 前：

```bash
# 运行测试
cd src/control-plane && go test ./...
cd src/es-plugin && ./gradlew test

# 格式化代码
cd src/control-plane && gofmt -w .

# 检查 lint
cd src/control-plane && golint ./...
```

### 2. 推送分支

```bash
git push origin feature/your-feature-name
```

### 3. 创建 Pull Request

在 GitHub 上创建 PR，填写以下信息：

**标题**：简洁描述变更内容

**描述模板**：
```markdown
## 变更内容
<!-- 描述此 PR 的主要变更 -->

## 变更类型
- [ ] 新功能
- [ ] Bug 修复
- [ ] 文档更新
- [ ] 代码重构
- [ ] 性能优化

## 测试
- [ ] 单元测试通过
- [ ] 集成测试通过
- [ ] 手动测试通过

## 检查清单
- [ ] 代码遵循项目规范
- [ ] 添加了必要的测试
- [ ] 更新了相关文档
- [ ] 提交信息符合规范

## 相关 Issue
Closes #<issue-number>

## 截图（如果适用）
```

### 4. Code Review

- 至少需要 1 位维护者 Review
- 解决所有 Review 意见
- 确保 CI 检查通过

### 5. 合并

Review 通过后，维护者会合并您的 PR。

---

## 🐛 报告 Bug

### 提交 Issue

在 GitHub Issues 中提交，包含以下信息：

1. **Bug 描述**：清晰描述问题
2. **复现步骤**：详细的复现步骤
3. **期望行为**：期望的正确行为
4. **实际行为**：实际发生的行为
5. **环境信息**：
   - OS 版本
   - Go 版本
   - Kubernetes 版本
   - ES 版本
6. **日志和错误信息**：相关的日志输出
7. **截图**：如果适用

---

## 💡 提出新功能

### 提交 Feature Request

在 GitHub Issues 中提交，包含：

1. **功能描述**：详细描述建议的功能
2. **使用场景**：为什么需要这个功能
3. **预期效果**：这个功能应该如何工作
4. **可选方案**：是否有替代方案

---

## 📚 文档贡献

文档同样重要！您可以：

1. **修正文档错误**
2. **补充缺失文档**
3. **添加使用示例**
4. **翻译文档**

文档位于 `docs/` 目录，遵循 Markdown 格式。

文档结构说明：
- `docs/architecture/` - 架构设计文档
- `docs/deployment/` - 部署指南
- `docs/development/` - 开发文档
- `docs/operations/` - 运维手册
- `docs/api/` - API 文档

---

## ❓ 需要帮助？

- 查看 [文档中心](docs/README.md)
- 查看 [FAQ](docs/FAQ.md)
- 提交 [GitHub Issue](https://github.com/your-org/es-paas/issues)
- 加入讨论组（待添加）

---

## 📜 许可证

通过贡献代码，您同意您的贡献将在 MIT 许可证下发布。

---

**感谢您的贡献！** 🎉
