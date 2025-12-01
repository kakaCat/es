#!/usr/bin/env bash
set -euo pipefail

# Git 仓库初始化脚本

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

echo "==================================="
echo "ES Serverless - Git 初始化"
echo "==================================="
echo ""

cd "$PROJECT_ROOT"

# 检查是否已经是 Git 仓库
if [ -d ".git" ]; then
    echo "✓ Git 仓库已存在"
    echo ""
    echo "当前状态:"
    git status --short
    exit 0
fi

echo "初始化 Git 仓库..."
git init

# 创建 .gitignore
echo "创建 .gitignore..."
cat > .gitignore <<'EOF'
# ============================================
# Terraform
# ============================================
**/.terraform/*
*.tfstate
*.tfstate.*
crash.log
crash.*.log
*.tfvars
*.tfvars.json
!terraform.tfvars.example
override.tf
override.tf.json
*_override.tf
*_override.tf.json
.terraform.lock.hcl
.terraformrc
terraform.rc

# 租户配置 (自动生成)
terraform/tenants/*
!terraform/tenants/.gitkeep

# ============================================
# Helm
# ============================================
helm/*/charts/
helm/*/Chart.lock
*.tgz

# ============================================
# 备份和临时文件
# ============================================
backups/
*.backup
*.bak
*.tmp
*.temp

# ============================================
# IDE 和编辑器
# ============================================
.vscode/
*.code-workspace
.idea/
*.iml
*.iws
*.swp
*.swo
*~
\#*\#
.\#*
*.sublime-project
*.sublime-workspace

# ============================================
# 日志
# ============================================
*.log

# ============================================
# Kubernetes
# ============================================
kubeconfig
*.kubeconfig
secrets/
*.secret.yaml

# ============================================
# Go
# ============================================
server/manager
server/*.exe
server/*.dll
server/*.so
server/*.dylib
__pycache__/
*.py[cod]
*$py.class

# Go modules
go.sum
vendor/

# ============================================
# 构建产物
# ============================================
es-plugin/build/
es-plugin/.gradle/
node_modules/
npm-debug.log
yarn-error.log

# ============================================
# 数据目录
# ============================================
server/data/*
!server/data/.gitkeep

# ============================================
# 环境变量
# ============================================
.env
.env.local
.env.*.local

# ============================================
# 压缩文件
# ============================================
*.zip
*.tar.gz
*.tar
*.rar

# ============================================
# 操作系统
# ============================================
.DS_Store
.AppleDouble
.LSOverride
Icon
._*
.DocumentRevisions-V100
.fseventsd
.Spotlight-V100
.TemporaryItems
.Trashes
.VolumeIcon.icns
.com.apple.timemachine.donotpresent
Thumbs.db
ehthumbs.db
Desktop.ini
$RECYCLE.BIN/
*.lnk
*~
.fuse_hidden*
.directory
.Trash-*
.nfs*
EOF

# 创建 .gitkeep 文件
echo "创建 .gitkeep 文件..."
mkdir -p terraform/tenants
touch terraform/tenants/.gitkeep
mkdir -p server/data
touch server/data/.gitkeep

# 添加所有文件
echo "添加文件到 Git..."
git add .

# 创建初始提交
echo "创建初始提交..."
git commit -m "feat: 初始提交 - ES Serverless Terraform/Helm 平台

包含:
- Terraform 配置 (5 个模块)
- Helm Charts (3 个 Charts)
- 部署脚本
- Makefile (30+ 命令)
- 完整文档 (25,000+ 字)
- Go + Helm SDK 集成示例

功能:
✅ 平台一键部署
✅ 租户快速创建
✅ 多租户隔离
✅ 完整监控 (Prometheus + Grafana)
✅ 自动化运维工具
"

echo ""
echo "==================================="
echo "✓ Git 仓库初始化完成!"
echo "==================================="
echo ""
echo "提交信息:"
git log --oneline -1
echo ""
echo "文件统计:"
git ls-files | wc -l | xargs echo "  提交文件数:"
echo ""
echo "下一步:"
echo "  1. 添加远程仓库:"
echo "     git remote add origin <repository-url>"
echo ""
echo "  2. 推送到远程:"
echo "     git push -u origin main"
echo ""
echo "  或者继续本地开发:"
echo "     git status"
echo "     git log"
echo ""
