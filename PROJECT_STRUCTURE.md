# 📁 项目结构说明

> ES Serverless 平台的标准化项目结构

**版本**: v2.0
**更新日期**: 2025-12-01

---

## 📊 目录树

```
es-paas/es/
│
├── README.md                       # 项目总览
├── CONTRIBUTING.md                 # 贡献指南
├── PROJECT_STRUCTURE.md            # 本文档
├── CLAUDE.md                       # AI 助手配置
│
├── src/                            # 💻 源代码
│   ├── control-plane/              # Go 控制平面服务
│   │   ├── *.go                    # 核心Go代码
│   │   ├── Dockerfile              # 容器化配置
│   │   └── README.md               # 控制平面文档
│   │
│   ├── es-plugin/                  # ES IVF 向量搜索插件
│   │   ├── src/main/java/          # Java源代码
│   │   ├── build.gradle            # Gradle构建配置
│   │   └── README.md               # 插件文档
│   │
│   └── frontend/                   # Web管理界面
│       ├── index.html              # 前端入口
│       ├── js/                     # JavaScript代码
│       └── README.md               # 前端文档
│
├── deployments/                    # 🚀 部署配置
│   ├── README.md                   # 部署总览
│   │
│   ├── terraform/                  # Terraform IaC配置
│   │   ├── main.tf                 # 主配置
│   │   ├── variables.tf            # 变量定义
│   │   ├── modules/                # Terraform模块
│   │   │   ├── tenant/             # 租户模块
│   │   │   ├── control-plane/      # 控制平面模块
│   │   │   ├── monitoring/         # 监控模块
│   │   │   ├── logging/            # 日志模块
│   │   │   └── gpu/                # GPU调度模块
│   │   └── README.md               # Terraform文档
│   │
│   ├── helm/                       # Helm Charts
│   │   ├── elasticsearch/          # ES集群Chart
│   │   ├── control-plane/          # 控制平面Chart
│   │   ├── monitoring/             # 监控Chart
│   │   └── README.md               # Helm文档
│   │
│   ├── kubernetes/                 # 原生K8s YAML
│   │   ├── base/                   # 基础配置
│   │   ├── overlays/               # 环境覆盖
│   │   │   ├── dev/                # 开发环境
│   │   │   └── prod/               # 生产环境
│   │   └── README.md               # Kubernetes文档
│   │
│   └── docker/                     # Docker配置
│       ├── docker-compose.yml      # Docker Compose
│       └── README.md               # Docker文档
│
├── scripts/                        # 🛠️ 工具脚本
│   ├── deploy/                     # 部署脚本
│   │   ├── deploy.sh               # 主部署脚本
│   │   ├── cluster.sh              # 集群管理
│   │   ├── create-tenant.sh        # 租户创建
│   │   └── deploy-terraform.sh     # Terraform部署
│   │
│   ├── build/                      # 构建脚本
│   │   ├── build-plugin.sh         # 构建ES插件
│   │   ├── build-reporting.sh      # 构建报告服务
│   │   └── build.sh                # 构建所有组件
│   │
│   ├── ops/                        # 运维脚本
│   │   ├── backup-es-snapshot.sh   # ES快照备份
│   │   ├── backup-metadata.sh      # 元数据备份
│   │   ├── restore-from-snapshot.sh# 快照恢复
│   │   ├── monitor.sh              # 监控脚本
│   │   └── shard-management.sh     # 分片管理
│   │
│   └── dev/                        # 开发辅助
│       ├── setup-dev.sh            # 开发环境搭建
│       └── test-ivf.sh             # IVF功能测试
│
├── docs/                           # 📚 文档中心
│   ├── README.md                   # 文档索引
│   │
│   ├── architecture/               # 架构设计
│   │   ├── multi-tenancy.md        # 多租户架构
│   │   ├── shard-replication.md    # 分片复制
│   │   ├── auto-scaling.md         # 自动扩展
│   │   └── gpu-acceleration.md     # GPU加速
│   │
│   ├── deployment/                 # 部署文档
│   │   ├── README.md               # 部署总览
│   │   ├── terraform-helm.md       # Terraform+Helm
│   │   ├── shell-scripts.md        # Shell脚本
│   │   └── gpu-setup.md            # GPU配置
│   │
│   ├── development/                # 开发文档
│   │   ├── setup.md                # 环境搭建
│   │   ├── es-plugin.md            # 插件开发
│   │   ├── control-plane.md        # 控制平面开发
│   │   └── testing.md              # 测试指南
│   │
│   ├── operations/                 # 运维手册
│   │   ├── monitoring.md           # 监控告警
│   │   ├── disaster-recovery.md    # 灾难恢复
│   │   ├── deployment-reporting.md # 部署上报
│   │   └── troubleshooting.md      # 故障排查
│   │
│   ├── api/                        # API文档
│   │   ├── rest-api.md             # REST API
│   │   └── vector-search.md        # 向量搜索API
│   │
│   └── archive/                    # 归档文档
│       ├── implementation-summary/ # 实现总结
│       └── requirements/           # 需求文档
│
├── tests/                          # 🧪 测试
│   ├── unit/                       # 单元测试
│   ├── integration/                # 集成测试
│   ├── e2e/                        # 端到端测试
│   └── README.md                   # 测试文档
│
├── examples/                       # 📖 示例代码
│   ├── basic-usage/                # 基础使用
│   ├── multi-tenant/               # 多租户示例
│   ├── gpu-acceleration/           # GPU加速示例
│   └── README.md                   # 示例说明
│
├── configs/                        # ⚙️ 配置文件
│   ├── elasticsearch.yml           # ES配置模板
│   ├── kind-config.yaml            # Kind集群配置
│   └── README.md                   # 配置说明
│
└── tools/                          # 🔧 开发工具
    └── README.md                   # 工具说明
```

---

## 📂 目录职责说明

### `src/` - 源代码

所有应用程序源代码的集中位置。

#### `src/control-plane/`
- **职责**: Go编写的控制平面服务
- **内容**:
  - REST API服务器
  - 自动扩展器（AutoScaler）
  - 分片控制器（ShardController）
  - 复制监控器（ReplicationMonitor）
  - 一致性检查器（ConsistencyChecker）
- **关键文件**:
  - `main.go` - 主入口
  - `autoscaler.go` - 自动扩展逻辑
  - `shard_controller.go` - 分片管理
  - `Dockerfile` - 容器化

#### `src/es-plugin/`
- **职责**: Elasticsearch IVF向量搜索插件
- **内容**:
  - IVF算法实现（Java）
  - GPU加速支持（JCuda）
  - 自定义字段类型
  - 向量搜索DSL
- **关键文件**:
  - `InvertedFileIndex.java` - IVF索引
  - `GPUVectorSimilarity.java` - GPU加速
  - `build.gradle` - 构建配置

#### `src/frontend/`
- **职责**: Web管理界面
- **内容**:
  - 集群管理UI
  - 多租户配置界面
  - 监控仪表板
- **技术栈**: HTML + JavaScript

---

### `deployments/` - 部署配置

所有部署相关的配置文件。

#### `deployments/terraform/`
- **职责**: 基础设施即代码（IaC）
- **适用**: 生产环境、多租户、大规模部署
- **特点**:
  - 模块化设计
  - 状态管理
  - 多云支持

#### `deployments/helm/`
- **职责**: Kubernetes应用包管理
- **内容**:
  - `elasticsearch/` - ES集群Chart
  - `control-plane/` - 控制平面Chart
  - `monitoring/` - Prometheus+Grafana
- **优势**: 版本管理、回滚能力

#### `deployments/kubernetes/`
- **职责**: 原生Kubernetes YAML配置
- **适用**: 开发测试、快速验证
- **结构**:
  - `base/` - 基础配置
  - `overlays/dev/` - 开发环境
  - `overlays/prod/` - 生产环境

#### `deployments/docker/`
- **职责**: Docker Compose本地部署
- **适用**: 本地开发、单机部署

---

### `scripts/` - 工具脚本

自动化脚本，按功能分类。

#### `scripts/deploy/`
- 集群创建和管理
- 租户配置
- 一键部署

#### `scripts/build/`
- 编译ES插件
- 构建容器镜像
- CI/CD集成

#### `scripts/ops/`
- 备份恢复
- 监控告警
- 分片管理

#### `scripts/dev/`
- 开发环境搭建
- 测试工具
- 调试脚本

---

### `docs/` - 文档中心

所有项目文档的集中管理。

#### 文档分类
- **architecture/** - 系统设计和架构决策
- **deployment/** - 部署指南和配置说明
- **development/** - 开发者指南
- **operations/** - 运维手册
- **api/** - API参考文档
- **archive/** - 历史文档归档

#### 文档规范
- 使用Markdown格式
- 包含代码示例
- 保持更新日期
- 交叉链接清晰

---

### `tests/` - 测试

测试代码分类管理。

- **unit/** - 单元测试
- **integration/** - 集成测试
- **e2e/** - 端到端测试

---

### `examples/` - 示例

使用示例和最佳实践。

- 基础使用教程
- 多租户配置示例
- GPU加速示例
- 性能优化案例

---

### `configs/` - 配置

配置文件模板和示例。

- Elasticsearch配置
- Kubernetes集群配置
- 监控配置

---

## 🔄 与旧结构的映射

| 旧路径 | 新路径 | 说明 |
|--------|--------|------|
| `server/` | `src/control-plane/` | 控制平面代码 |
| `es-plugin/` | `src/es-plugin/` | ES插件代码 |
| `frontend/` | `src/frontend/` | 前端代码 |
| `terraform/` | `deployments/terraform/` | Terraform配置 |
| `helm/` | `deployments/helm/` | Helm Charts |
| `k8s/` | `deployments/kubernetes/` | K8s配置 |
| `docker/` | `deployments/docker/` | Docker配置 |
| `scripts/*.sh` | `scripts/{deploy,build,ops,dev}/` | 脚本分类 |
| `demo/` | `examples/` | 示例代码 |
| `docs/*.md` | `docs/{architecture,deployment,operations}/` | 文档分类 |

---

## 🎯 目录使用指南

### 新开发者
1. 阅读 `README.md`
2. 查看 `docs/development/setup.md`
3. 运行 `scripts/dev/setup-dev.sh`
4. 查看 `examples/` 中的示例

### 部署运维
1. 阅读 `docs/deployment/README.md`
2. 选择部署方式（Terraform 或 Shell脚本）
3. 使用 `scripts/deploy/` 中的脚本
4. 参考 `docs/operations/` 运维文档

### 贡献代码
1. 阅读 `CONTRIBUTING.md`
2. 在 `src/` 中修改代码
3. 在 `tests/` 中添加测试
4. 更新 `docs/` 中的相关文档

---

## 📝 维护规范

### 目录创建
- 新建目录必须包含 `README.md`
- 说明目录用途和内容
- 列出重要文件索引

### 文件命名
- 使用小写字母和短横线
- 避免空格和特殊字符
- 文件名要有描述性

### 文档更新
- 代码变更必须同步更新文档
- 保持文档版本号和日期
- 定期清理过时文档到archive/

---

## 🔗 相关链接

- [项目README](README.md)
- [文档中心](docs/README.md)
- [部署指南](docs/deployment/README.md)
- [贡献指南](CONTRIBUTING.md)

---

**版本历史**:
- v2.0 (2025-12-01) - 标准化重构
- v1.0 (2025-11-01) - 初始结构