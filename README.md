# ES Serverless Platform

A Serverless ElasticSearch platform with vector search capabilities based on IVF algorithms.

## Prerequisites

Before deploying the system, ensure you have the following prerequisites installed:

- Docker Desktop with Kubernetes enabled
- kubectl CLI
- Go 1.21+ (for local development)
- Bash shell

### Kubernetes Environment Setup

1. **Enable Kubernetes in Docker Desktop**:
   - Open Docker Desktop
   - Go to Settings > Kubernetes
   - Check "Enable Kubernetes"
   - Click "Apply & Restart"

2. **Verify Kubernetes is running**:
   ```bash
   kubectl cluster-info
   kubectl get nodes
   ```

If you encounter any Kubernetes setup issues, please refer to [KUBERNETES_SETUP_ISSUES.md](KUBERNETES_SETUP_ISSUES.md) for detailed troubleshooting steps.

## Features

- 🏢 **多租户架构**：通过租户组织ID实现组织级别的资源隔离（详见 [/docs/多租户架构说明.md](/docs/多租户架构说明.md)）
- **Serverless Elasticsearch Clusters**: Automatic provisioning and scaling of Elasticsearch clusters
- **Vector Search**: High-dimensional vector search using IVF algorithms (similar to FAISS)
- **Multi-tenancy**: Isolated environments for different users/organizations
- **Auto-scaling**: Dynamic scaling based on CPU, memory, and QPS metrics
- 🔒 **配额管理**：自动扩展时检查租户配额，防止资源超限（详见 [/docs/自动扩展配额管理说明.md](/docs/自动扩展配额管理说明.md)）
- 🗑️ **逻辑删除**：租户容器记录采用逻辑删除机制，确保数据安全性和可恢复性（详见 [/docs/逻辑删除实现说明.md](/docs/逻辑删除实现说明.md)）
- 🗄️ **数据库存储**：元数据存储支持PostgreSQL数据库和文件系统双模式（详见 [/docs/数据库存储实现说明.md](/docs/数据库存储实现说明.md)）
- 🔄 **副本数据同步**：实时监控分片副本同步状态，自动故障恢复（详见 [/docs/分片数据同步实现进度.md](/docs/分片数据同步实现进度.md)）
- 📊 **监控系统**：集成Prometheus和Grafana的监控系统（详见 [/docs/监控系统架构说明.md](/docs/监控系统架构说明.md)）
- 💾 **数据备份与恢复**：自动备份Elasticsearch快照和元数据到MinIO(S3兼容)，支持按租户隔离（详见 [/docs/灾难恢复手册.md](/docs/灾难恢复手册.md)）

## Deployment

To deploy the entire system:

```bash
./scripts/deploy.sh install
```

To check the system status:

```bash
./scripts/deploy.sh status
```

To uninstall the system:

```bash
./scripts/deploy.sh uninstall
```

## Development

To build and run the control plane locally:

```bash
cd server
go build -o manager .
./manager
```

### 前端管理界面

项目包含一个简单的前端管理界面，可以通过Web浏览器访问和管理ES Serverless集群。

1. 进入前端目录：
   ```bash
   cd frontend
   ```

2. 启动一个简单的HTTP服务器（例如使用Python）：
   ```bash
   # Python 3
   python -m http.server 8000
   
   # 或者 Python 2
   python -m SimpleHTTPServer 8000
   ```

3. 在浏览器中访问：
   ```
   http://localhost:8000
   ```

4. 前端界面功能包括：
   - 创建集群（支持多租户组织ID）
   - 删除集群
   - 查询所有集群
   - 查询集群详情

📖 详细交互流程请查看 [/docs/前端与后端交互时序图.md](/docs/前端与后端交互时序图.md)

## Architecture

The system consists of the following components:

1. **Control Plane**: Manages cluster lifecycle, auto-scaling, and monitoring
2. **Data Plane**: Elasticsearch clusters deployed on Kubernetes StatefulSets
3. **Monitoring**: Prometheus and Grafana for metrics collection and visualization
4. **Logging**: Fluentd for log collection and aggregation
5. **Reporting**: Service for collecting and reporting usage statistics
6. **Frontend**: Web-based management interface
7. **Backup & Recovery**: Automated backup and recovery system using MinIO (S3 compatible)

## API Endpoints

The control plane exposes the following REST API endpoints:

- `POST /clusters` - Create a new Elasticsearch cluster
- `DELETE /clusters` - Delete an existing Elasticsearch cluster
- `GET /clusters` - List all clusters
- `POST /vector-indexes` - Create a new vector index
- `GET /vector-indexes` - List all vector indexes
- `DELETE /vector-indexes` - Delete a vector index
- `POST /clusters/scale` - Scale a cluster
- `GET /deployments` - List deployment status
- `GET /metrics` - List monitoring metrics

For detailed API documentation, see [/docs/API接口文档.md](/docs/API接口文档.md)