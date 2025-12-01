# 如何拉取和使用代码

本文档说明如何从 Git 仓库拉取代码并开始使用。

## 场景 1: 第一次使用 (Clone 仓库)

### 前置要求

```bash
# 安装 Git
brew install git

# 配置 Git (首次使用)
git config --global user.name "Your Name"
git config --global user.email "your.email@example.com"
```

### 克隆仓库

```bash
# 方法 1: HTTPS (推荐)
git clone https://github.com/your-org/es-serverless.git
cd es-serverless

# 方法 2: SSH (需要先配置 SSH key)
git clone git@github.com:your-org/es-serverless.git
cd es-serverless

# 查看分支
git branch -a

# 切换到开发分支 (如果有)
git checkout develop
```

### 验证代码

```bash
# 查看文件
ls -la

# 应该看到:
# - terraform/
# - helm/
# - scripts/
# - docs/
# - Makefile
# - README.md

# 查看提交历史
git log --oneline -10

# 查看当前状态
git status
```

## 场景 2: 已有本地代码,初始化 Git

### 当前状态

代码已在本地目录: `/Users/yunpeng/Documents/es项目`

### 初始化 Git 仓库

```bash
# 进入项目目录
cd /Users/yunpeng/Documents/es项目

# 运行初始化脚本
./scripts/init-git.sh

# 或手动初始化
git init
git add .
git commit -m "feat: 初始提交"
```

### 添加远程仓库

```bash
# 方法 1: GitHub
git remote add origin https://github.com/your-org/es-serverless.git

# 方法 2: GitLab
git remote add origin https://gitlab.com/your-org/es-serverless.git

# 方法 3: 自建 Git 服务器
git remote add origin git@your-server.com:your-org/es-serverless.git

# 验证远程仓库
git remote -v
```

### 推送到远程

```bash
# 推送到 main 分支
git push -u origin main

# 或推送到 master 分支
git push -u origin master

# 推送所有分支和标签
git push --all origin
git push --tags origin
```

## 场景 3: 更新已有代码

### 拉取最新代码

```bash
# 进入项目目录
cd /path/to/es-serverless

# 查看当前状态
git status

# 拉取最新代码
git pull

# 或分两步
git fetch origin
git merge origin/main
```

### 处理冲突

如果有冲突:

```bash
# 查看冲突文件
git status

# 手动编辑冲突文件
vim <conflicted-file>

# 标记为已解决
git add <conflicted-file>

# 完成合并
git commit
```

### 更新依赖

```bash
# 更新 Terraform
cd terraform
terraform init -upgrade

# 更新 Helm
helm repo update

# 更新 Go 依赖 (如果有)
cd server
go mod download
```

## 场景 4: 团队协作

### 创建功能分支

```bash
# 从 main 创建新分支
git checkout -b feature/new-tenant-api

# 开发功能...
# 编辑文件

# 查看变更
git status
git diff

# 提交变更
git add .
git commit -m "feat: 添加新的租户 API"

# 推送到远程
git push -u origin feature/new-tenant-api
```

### 提交 Pull Request

```bash
# 在 GitHub/GitLab 网页上创建 PR
# 或使用 CLI (GitHub)
gh pr create --title "添加新的租户 API" --body "详细说明..."

# 或使用 CLI (GitLab)
glab mr create --title "添加新的租户 API"
```

### 代码审查后合并

```bash
# 拉取最新的 main
git checkout main
git pull

# 删除本地功能分支
git branch -d feature/new-tenant-api

# 删除远程功能分支
git push origin --delete feature/new-tenant-api
```

## 常用 Git 命令

### 查看状态和历史

```bash
# 查看当前状态
git status

# 查看变更
git diff

# 查看提交历史
git log
git log --oneline
git log --graph --oneline --all

# 查看某个文件的历史
git log -- path/to/file
```

### 分支操作

```bash
# 列出分支
git branch
git branch -a  # 包括远程分支

# 创建分支
git branch feature/new-feature

# 切换分支
git checkout feature/new-feature

# 创建并切换 (推荐)
git checkout -b feature/new-feature

# 删除分支
git branch -d feature/new-feature

# 重命名分支
git branch -m old-name new-name
```

### 提交操作

```bash
# 添加文件
git add file.txt
git add .  # 添加所有

# 提交
git commit -m "message"

# 修改上次提交
git commit --amend

# 撤销提交 (保留变更)
git reset --soft HEAD~1

# 撤销提交 (丢弃变更)
git reset --hard HEAD~1
```

### 远程操作

```bash
# 查看远程仓库
git remote -v

# 添加远程仓库
git remote add origin <url>

# 拉取
git fetch origin
git pull origin main

# 推送
git push origin main
git push -u origin main  # 首次推送

# 删除远程分支
git push origin --delete branch-name
```

### 标签操作

```bash
# 创建标签
git tag v1.0.0
git tag -a v1.0.0 -m "Version 1.0.0"

# 查看标签
git tag
git show v1.0.0

# 推送标签
git push origin v1.0.0
git push --tags  # 推送所有标签

# 删除标签
git tag -d v1.0.0
git push origin --delete v1.0.0
```

## 项目特定的 Git 工作流

### 开发新功能

```bash
# 1. 更新主分支
git checkout main
git pull

# 2. 创建功能分支
git checkout -b feature/multi-gpu-support

# 3. 开发功能
# 编辑 terraform/modules/tenant/variables.tf
# 添加 GPU 配置

# 4. 测试
make validate
make plan

# 5. 提交
git add terraform/modules/tenant/
git commit -m "feat: 添加多 GPU 支持

- 添加 gpu_type 变量
- 支持指定 GPU 型号
- 更新文档
"

# 6. 推送
git push -u origin feature/multi-gpu-support

# 7. 创建 PR
gh pr create
```

### 修复 Bug

```bash
# 1. 创建修复分支
git checkout -b fix/terraform-lock-issue

# 2. 修复问题
# 编辑文件...

# 3. 提交
git add .
git commit -m "fix: 修复 Terraform state 锁定问题

问题: Terraform 并发执行导致 state 锁定
解决: 添加重试逻辑和超时配置

Closes #123
"

# 4. 推送并创建 PR
git push -u origin fix/terraform-lock-issue
gh pr create
```

### 更新文档

```bash
# 1. 创建文档分支
git checkout -b docs/update-helm-guide

# 2. 更新文档
vim docs/terraform-helm-guide.md

# 3. 提交
git add docs/
git commit -m "docs: 更新 Helm 使用指南

- 添加故障排查章节
- 更新配置示例
- 修正错别字
"

# 4. 推送
git push -u origin docs/update-helm-guide
```

## Git 提交规范

### 提交消息格式

```
<type>(<scope>): <subject>

<body>

<footer>
```

### Type 类型

- `feat`: 新功能
- `fix`: Bug 修复
- `docs`: 文档更新
- `style`: 代码格式 (不影响功能)
- `refactor`: 重构
- `perf`: 性能优化
- `test`: 测试
- `chore`: 构建/工具链

### 示例

```bash
# 新功能
git commit -m "feat(tenant): 添加租户资源配额管理"

# Bug 修复
git commit -m "fix(helm): 修复 values 传递错误"

# 文档
git commit -m "docs: 更新快速开始指南"

# 重构
git commit -m "refactor(manager): 重构 Helm 集成模块"

# 详细提交
git commit -m "feat(monitoring): 添加自定义 Grafana Dashboard

添加以下 Dashboard:
- Elasticsearch 集群健康
- 租户资源使用
- QPS 监控

相关 Issue: #45
"
```

## 常见问题

### Q1: 如何撤销最后一次提交?

```bash
# 撤销但保留变更
git reset --soft HEAD~1

# 撤销并丢弃变更
git reset --hard HEAD~1
```

### Q2: 如何临时保存当前工作?

```bash
# 保存工作区
git stash

# 查看 stash
git stash list

# 恢复
git stash pop

# 或指定
git stash apply stash@{0}
```

### Q3: 如何查看某次提交的变更?

```bash
# 查看提交详情
git show <commit-hash>

# 查看文件变更
git diff <commit-hash>~1 <commit-hash>
```

### Q4: 如何回退到某个版本?

```bash
# 创建新提交来回退
git revert <commit-hash>

# 或强制回退 (危险!)
git reset --hard <commit-hash>
git push --force
```

### Q5: 如何合并多个提交?

```bash
# 交互式 rebase
git rebase -i HEAD~3

# 在编辑器中选择 squash
# 保存退出
```

## .gitignore 说明

项目的 `.gitignore` 已配置忽略:

- ✅ Terraform 状态文件 (`*.tfstate`)
- ✅ Terraform 变量文件 (`*.tfvars`,保留 `.example`)
- ✅ 租户配置目录 (`terraform/tenants/*`)
- ✅ 备份文件 (`backups/`, `*.backup`)
- ✅ IDE 配置 (`.vscode/`, `.idea/`)
- ✅ 日志文件 (`*.log`)
- ✅ 临时文件 (`*.tmp`, `.DS_Store`)

## 文件结构预览

```bash
# 查看 Git 管理的文件
git ls-files

# 查看忽略的文件
git status --ignored

# 查看文件统计
git ls-files | wc -l
```

## 最佳实践

### 1. 频繁提交

```bash
# 小步提交,易于回滚
git commit -m "feat: 添加 GPU 变量定义"
git commit -m "feat: 实现 GPU 分配逻辑"
git commit -m "test: 添加 GPU 配置测试"
```

### 2. 清晰的提交消息

```bash
# 好 ✅
git commit -m "feat(tenant): 添加多 GPU 支持,支持指定 GPU 型号"

# 不好 ❌
git commit -m "更新代码"
```

### 3. 定期推送

```bash
# 每天结束前推送
git push origin feature/your-branch
```

### 4. 保持分支更新

```bash
# 定期从 main 合并
git checkout feature/your-branch
git merge main
```

### 5. 使用标签标记版本

```bash
# 发布版本时
git tag -a v1.0.0 -m "Release 1.0.0"
git push --tags
```

## 获取帮助

```bash
# Git 帮助
git help
git help <command>

# 查看配置
git config --list

# 项目文档
cat README.md
cat GETTING_STARTED.md
```

---

**快速开始**: 从 [GETTING_STARTED.md](GETTING_STARTED.md) 开始!
