# 项目重构方案

## 📊 当前问题分析

### 1. 根目录文档混乱
- ❌ 25+ 个 Markdown 文件散落在根目录
- ❌ 多个重复的实现总结文档
- ❌ 临时文档没有归档

### 2. 部署配置重复
- ❌ `terraform/`, `helm/`, `scripts/` 和 `deployment-*/` 目录并存
- ❌ 软链接增加复杂度
- ❌ 不清楚应该使用哪个目录

### 3. 目录职责不清
- ❌ `demo/` 和 `examples/` 功能重叠
- ❌ `docker/` 和 `k8s/` 分散
- ❌ 缺少明确的开发文档位置

## 🎯 新的标准结构

```
es-paas/es/
│
├── README.md                       # 项目总览
├── CONTRIBUTING.md                 # 贡献指南（新增）
├── CHANGELOG.md                    # 变更日志（新增）
├── .gitignore
│
├── docs/                           # 📚 所有文档集中管理
│   ├── README.md                   # 文档索引
│   │
│   ├── architecture/               # 架构设计
│   │   ├── system-overview.md
│   │   ├── multi-tenancy.md
│   │   ├── gpu-acceleration.md
│   │   └── load-balancing.md
│   │
│   ├── deployment/                 # 部署文档
│   │   ├── README.md               # 部署总览
│   │   ├── terraform-helm.md       # Terraform+Helm方式
│   │   ├── shell-scripts.md        # Shell脚本方式
│   │   └── comparison.md           # 方式对比
│   │
│   ├── development/                # 开发文档
│   │   ├── setup.md                # 环境搭建
│   │   ├── es-plugin.md            # ES插件开发
│   │   ├── control-plane.md        # 控制平面开发
│   │   └── testing.md              # 测试指南
│   │
│   ├── operations/                 # 运维文档
│   │   ├── monitoring.md
│   │   ├── backup-restore.md
│   │   ├── disaster-recovery.md
│   │   └── troubleshooting.md
│   │
│   ├── api/                        # API文档
│   │   ├── rest-api.md
│   │   └── vector-search-api.md
│   │
│   └── archive/                    # 归档文档
│       ├── implementation-summary/ # 实现总结归档
│       └── requirements/           # 需求文档归档
│
├── src/                            # 💻 源代码（新目录名）
│   ├── control-plane/              # 控制平面（原server/）
│   │   ├── cmd/                    # 命令行入口
│   │   ├── pkg/                    # 核心包
│   │   ├── api/                    # API定义
│   │   ├── Dockerfile
│   │   └── README.md
│   │
│   ├── es-plugin/                  # ES插件（保持原位置）
│   │   ├── src/
│   │   ├── build.gradle
│   │   └── README.md
│   │
│   └── frontend/                   # 前端（移入src）
│       ├── index.html
│       ├── js/
│       └── README.md
│
├── deployments/                    # 🚀 部署配置（统一目录）
│   ├── README.md                   # 部署总览
│   │
│   ├── terraform/                  # Terraform + Helm方式
│   │   ├── main.tf
│   │   ├── variables.tf
│   │   ├── modules/
│   │   └── README.md
│   │
│   ├── helm/                       # Helm Charts
│   │   ├── elasticsearch/
│   │   ├── control-plane/
│   │   ├── monitoring/
│   │   └── README.md
│   │
│   ├── kubernetes/                 # 原生K8s YAML
│   │   ├── base/
│   │   ├── overlays/
│   │   └── README.md
│   │
│   └── docker/                     # Docker配置
│       ├── docker-compose.yml
│       └── README.md
│
├── scripts/                        # 🛠️ 工具脚本
│   ├── deploy/                     # 部署脚本
│   │   ├── deploy.sh
│   │   ├── cluster.sh
│   │   └── create-tenant.sh
│   │
│   ├── build/                      # 构建脚本
│   │   ├── build-plugin.sh
│   │   └── build-all.sh
│   │
│   ├── ops/                        # 运维脚本
│   │   ├── backup.sh
│   │   ├── monitor.sh
│   │   └── shard-management.sh
│   │
│   └── dev/                        # 开发辅助脚本
│       └── setup-dev.sh
│
├── tests/                          # 🧪 测试（新增）
│   ├── unit/                       # 单元测试
│   ├── integration/                # 集成测试
│   ├── e2e/                        # 端到端测试
│   └── README.md
│
├── examples/                       # 📖 示例代码（合并demo）
│   ├── basic-usage/
│   ├── multi-tenant/
│   ├── gpu-acceleration/
│   └── README.md
│
├── configs/                        # ⚙️ 配置文件模板
│   ├── elasticsearch.yml
│   ├── kind-config.yaml
│   └── README.md
│
└── tools/                          # 🔧 开发工具（新增）
    ├── linters/
    ├── formatters/
    └── README.md
```

## 🔄 迁移映射

### 目录迁移

| 原路径 | 新路径 | 操作 |
|--------|--------|------|
| `server/` | `src/control-plane/` | 重命名+重构 |
| `es-plugin/` | `src/es-plugin/` | 移动 |
| `frontend/` | `src/frontend/` | 移动 |
| `terraform/` | `deployments/terraform/` | 移动 |
| `helm/` | `deployments/helm/` | 移动 |
| `k8s/` | `deployments/kubernetes/` | 移动 |
| `docker/` | `deployments/docker/` | 移动 |
| `scripts/*.sh` | `scripts/deploy/` | 分类整理 |
| `demo/` + `examples/` | `examples/` | 合并 |

### 文档迁移

| 原文件 | 新路径 |
|--------|--------|
| `DEPLOYMENT.md` | `docs/deployment/README.md` |
| `具体要求.md` | `docs/archive/requirements/` |
| `IVF实现*.md` | `docs/archive/implementation-summary/` |
| `实现情况清单.md` | `docs/archive/` |
| `docs/多租户架构说明.md` | `docs/architecture/multi-tenancy.md` |
| `docs/灾难恢复手册.md` | `docs/operations/disaster-recovery.md` |

### 删除的文件

以下临时文档可以删除：
- `DEPLOYMENT_STATUS.md` (状态文档)
- `DEPLOYMENT_SUCCESS.md` (成功记录)
- `FILE_STRUCTURE.md` (旧结构说明)
- `IMPLEMENTATION_SUMMARY.md` (实现总结)
- `下一步操作指南.md` (临时指南)

## ✅ 迁移步骤

### 阶段1：准备工作
```bash
# 1. 创建新目录结构
mkdir -p src/{control-plane,es-plugin,frontend}
mkdir -p deployments/{terraform,helm,kubernetes,docker}
mkdir -p scripts/{deploy,build,ops,dev}
mkdir -p docs/{architecture,deployment,development,operations,api,archive}
mkdir -p tests/{unit,integration,e2e}
mkdir -p examples configs tools
```

### 阶段2：迁移源代码
```bash
# 2. 迁移server到control-plane
mv server src/control-plane

# 3. 迁移es-plugin
mv es-plugin src/

# 4. 迁移frontend
mv frontend src/
```

### 阶段3：整合部署配置
```bash
# 5. 迁移部署配置
mv terraform deployments/
mv helm deployments/
mv k8s deployments/kubernetes
mv docker deployments/

# 6. 删除deployment-*目录（软链接）
rm -rf deployment-terraform deployment-scripts
```

### 阶段4：整理脚本
```bash
# 7. 分类整理脚本
mv scripts/deploy.sh scripts/deploy/
mv scripts/cluster.sh scripts/deploy/
mv scripts/build-*.sh scripts/build/
mv scripts/backup*.sh scripts/ops/
mv scripts/monitor.sh scripts/ops/
```

### 阶段5：整理文档
```bash
# 8. 迁移文档
mv DEPLOYMENT.md docs/deployment/README.md
mv 具体要求.md docs/archive/requirements/
mv IVF*.md docs/archive/implementation-summary/

# 9. 合并examples
mv demo/* examples/
rmdir demo
```

### 阶段6：清理
```bash
# 10. 删除临时文档
rm DEPLOYMENT_STATUS.md DEPLOYMENT_SUCCESS.md FILE_STRUCTURE.md
rm IMPLEMENTATION_SUMMARY.md 下一步操作指南.md
```

## 📝 后续工作

### 1. 更新所有文档链接
- [ ] 更新README.md中的路径引用
- [ ] 更新docs/中的交叉引用
- [ ] 更新scripts/中的路径

### 2. 更新CI/CD配置
- [ ] 更新GitHub Actions路径
- [ ] 更新Dockerfile路径
- [ ] 更新测试脚本路径

### 3. 创建新文档
- [ ] CONTRIBUTING.md - 贡献指南
- [ ] CHANGELOG.md - 变更日志
- [ ] docs/development/setup.md - 开发环境搭建

## 🎯 预期收益

### 清晰度提升
- ✅ 代码、部署、文档、脚本完全分离
- ✅ 每个目录职责单一明确
- ✅ 新人可快速理解项目结构

### 可维护性提升
- ✅ 文档集中管理，易于查找
- ✅ 脚本分类清晰，易于维护
- ✅ 部署配置统一管理

### 专业性提升
- ✅ 符合开源项目标准结构
- ✅ 清晰的目录命名（src, deployments, docs）
- ✅ 完善的文档体系

## ⚠️ 风险评估

### 低风险
- ✅ 不影响代码逻辑
- ✅ Git历史完整保留
- ✅ 可逐步迁移

### 需注意
- ⚠️ 脚本中的路径引用需更新
- ⚠️ CI/CD配置需同步修改
- ⚠️ 文档链接需全面更新

## 📅 建议执行时间

**渐进式迁移**：
- Week 1: 创建新结构 + 迁移文档
- Week 2: 迁移源代码 + 更新引用
- Week 3: 迁移部署配置 + 测试
- Week 4: 清理临时文件 + 完善文档

---

**是否开始执行此重构方案？**
