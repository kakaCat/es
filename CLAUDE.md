# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

ES Serverless Platform - A serverless ElasticSearch platform with vector search capabilities using IVF (Inverted File) algorithms. This is a multi-tenant platform that provides automatic provisioning, scaling, and management of Elasticsearch clusters with custom vector search plugins.

## Essential Commands

### Building and Running

```bash
# Build the control plane Go service
cd server
go build -o manager .
./manager

# Build the ES IVF plugin
./scripts/build-plugin.sh

# Build the reporting component
./scripts/build-reporting.sh
```

### Deployment

```bash
# Deploy the entire system (requires Docker Desktop with Kubernetes enabled)
./scripts/deploy.sh install

# Check system status
./scripts/deploy.sh status

# Uninstall the system
./scripts/deploy.sh uninstall
```

### Cluster Management

```bash
# Create a cluster
./scripts/cluster.sh create <namespace> <replicas>

# Delete a cluster
./scripts/cluster.sh delete <namespace>

# Get cluster status
./scripts/cluster.sh status <namespace>

# Scale a cluster
./scripts/cluster.sh scale <namespace> <replicas>
```

### Frontend Development

```bash
cd frontend
python -m http.server 8000
# Access at http://localhost:8000
```

## Architecture Overview

### Three-Layer Architecture

1. **Control Plane** ([server/](server/))
   - Written in Go
   - Manages cluster lifecycle, auto-scaling, monitoring
   - Exposes REST API on port 8080
   - Components: Manager, AutoScaler, ShardController, ReplicationMonitor, ConsistencyChecker, MetadataService

2. **Data Plane** (Kubernetes-based)
   - Elasticsearch clusters deployed as StatefulSets
   - Each cluster runs in isolated namespaces
   - Custom IVF plugin for vector search
   - Monitoring via Prometheus/Grafana

3. **Plugin Layer** ([es-plugin/](es-plugin/))
   - Gradle-based Elasticsearch plugin (Java)
   - Implements IVF algorithm for vector search
   - Custom field types and query DSL

### Key Go Components

The server is organized into focused modules:

- **[main.go](server/main.go)** - HTTP API handlers, cluster orchestration, main entry point
- **[metadata.go](server/metadata.go)** - Metadata storage abstraction (PostgreSQL or filesystem)
- **[autoscaler.go](server/autoscaler.go)** - Auto-scaling logic with cooldown periods and trend analysis
- **[shard_controller.go](server/shard_controller.go)** - Shard rebalancing and allocation
- **[replication_monitor.go](server/replication_monitor.go)** - Replica synchronization monitoring
- **[consistency_checker.go](server/consistency_checker.go)** - Data consistency verification
- **[auto_recovery.go](server/auto_recovery.go)** - Automatic failure recovery
- **[monitoring.go](server/monitoring.go)** - Metrics collection and aggregation
- **[reporting.go](server/reporting.go)** - Deployment status reporting
- **[es_client.go](server/es_client.go)** - Elasticsearch API client

## Multi-Tenancy Model

The platform uses a three-level tenant hierarchy:

1. **Tenant Organization ID** (`tenant_org_id`) - Top-level organization isolation
2. **User** - Individual users within an organization
3. **Service Name** - Services owned by users

**Namespace Generation:**
```
{tenant_org_id}-{user}-{service_name}
```

**Kubernetes Labels:**
```yaml
labels:
  es-cluster: "true"
  tenant-org-id: "org-001"
  user: "alice"
  service-name: "vector-search"
```

All API requests creating clusters MUST include `tenant_org_id`, `user`, and `service_name`.

## REST API Endpoints

The control plane ([server/main.go](server/main.go)) exposes:

**Cluster Management:**
- `POST /clusters` - Create cluster (requires `tenant_org_id`, `user`, `service_name`)
- `DELETE /clusters` - Delete cluster
- `GET /clusters` - List all clusters
- `GET /clusters/{namespace}` - Get cluster details
- `POST /clusters/scale` - Scale cluster

**Vector Indexes:**
- `POST /vector-indexes` - Create vector index
- `GET /vector-indexes` - List vector indexes
- `DELETE /vector-indexes` - Delete vector index

**Monitoring:**
- `GET /deployments` - Deployment status
- `GET /metrics` - Monitoring metrics
- `GET /qps/{namespace}` - QPS metrics for namespace

**Tenant Management:**
- `GET /tenant/containers` - All tenant containers
- `GET /tenant/containers/{user}/{service}` - Specific container
- `GET /tenant/containers/org/{tenant_org_id}` - All containers for organization

## Data Storage

### Metadata Storage Modes

The platform supports **dual storage modes** (configured at startup):

1. **File-based storage** (default): `server/data/` directory
2. **PostgreSQL storage**: Set `USE_POSTGRES=true`

Files in `server/data/`:
- `tenant_{user}_{service}.json` - Tenant container metadata
- `deploy_{namespace}.json` - Deployment status
- `index_{namespace}_{index_name}.json` - Index metadata
- `metrics_{namespace}.json` - Monitoring metrics

### TenantContainer Structure

All tenant containers include:
- `tenant_org_id` - Organization identifier
- `user` - User identifier
- `service_name` - Service identifier
- `namespace` - Kubernetes namespace
- Resource specs (CPU, memory, disk, GPU)
- Vector configuration (dimension, vector_count)
- Status and timestamps

## Auto-Scaling Logic

Auto-scaling ([autoscaler.go](server/autoscaler.go)) implements:

1. **Cooldown Periods:**
   - Scale-up: 5 minutes
   - Scale-down: 10 minutes

2. **Trend Analysis:**
   - Tracks last 5 metric snapshots
   - Adjusts thresholds based on trends
   - Prevents flapping

3. **Multi-dimensional Metrics:**
   - CPU usage
   - Memory usage
   - QPS (Queries Per Second)
   - Composite scoring

4. **Quota Enforcement:**
   - Checks tenant quotas before scaling
   - Prevents resource overcommitment

## Shard Management

### Components

- **ShardController** ([shard_controller.go](server/shard_controller.go)) - Monitors every 30s, triggers rebalancing
- **ReplicationMonitor** ([replication_monitor.go](server/replication_monitor.go)) - Monitors replica sync every 10s
- **ConsistencyChecker** ([consistency_checker.go](server/consistency_checker.go)) - Verifies data consistency
- **AutoRecoveryManager** ([auto_recovery.go](server/auto_recovery.go)) - Automatic recovery every 30s

### Key Operations

Shard rebalancing adjusts Elasticsearch cluster settings via API:
- `cluster.routing.rebalance.enable`
- `cluster.routing.allocation.node_concurrent_recoveries`
- `indices.recovery.max_bytes_per_sec`

Replica monitoring checks shard allocation and triggers recovery for:
- Unassigned shards
- Failed initializations
- Out-of-sync replicas

## IVF Plugin Architecture

Located in [es-plugin/](es-plugin/), the plugin implements:

**IVF Algorithm Components:**
- **nlist** - Number of inverted file clusters
- **nprobe** - Number of clusters to search
- **KMeans Training** - Cluster centroid calculation
- **Posting Lists** - Inverted index construction

**Supported Operations:**
- Index building and updates
- Batch vector queries
- Index metadata reporting
- QPS tracking

**Build Process:**
```bash
cd es-plugin
gradle build
# Output: build/distributions/es-ivf-plugin-*.zip
```

## Deployment Architecture

The system deploys on **Kubernetes** (Docker Desktop or Kind cluster):

1. **Prerequisites:**
   - Docker Desktop with Kubernetes enabled
   - kubectl CLI
   - Go 1.21+

2. **Kubernetes Resources per Cluster:**
   - Namespace (labeled with tenant info)
   - StatefulSet (Elasticsearch pods)
   - Service (cluster access)
   - PersistentVolumeClaims (data storage)

3. **Monitoring Stack:**
   - Prometheus (metrics collection)
   - Grafana (visualization)
   - Fluentd (log aggregation)

## Important Implementation Notes

### Metadata Management

When creating clusters:
1. **ALWAYS save metadata FIRST** (TenantContainer, DeploymentStatus)
2. Then create Kubernetes resources
3. Then update status to "created"

This order ensures metadata exists even if Kubernetes operations fail.

### Quota Checking

Before scaling or creating clusters:
1. Load tenant quota from metadata
2. Calculate current usage
3. Verify new operation won't exceed limits
4. Reject if quota exceeded

See [docs/自动扩展配额管理说明.md](docs/自动扩展配额管理说明.md)

### Logical Deletion

Tenant containers use **logical deletion**:
- Set `deleted: true` flag
- Preserve `deleted_at` timestamp
- Retain for audit/recovery
- Physical cleanup is separate process

See [docs/逻辑删除实现说明.md](docs/逻辑删除实现说明.md)

### Error Handling for Multi-Tenancy

Return HTTP 400 if missing required fields:
- `tenant_org_id is required for multi-tenancy`
- `user is required`
- `service_name is required`

## Key Documentation Files

- [README.md](README.md) - Project overview and quickstart
- [docs/architecture.md](docs/architecture.md) - Detailed architecture diagrams
- [docs/多租户架构说明.md](docs/多租户架构说明.md) - Multi-tenancy design
- [docs/分片数据同步实现方案.md](docs/分片数据同步实现方案.md) - Shard replication design
- [docs/自动扩展配额管理说明.md](docs/自动扩展配额管理说明.md) - Quota management
- [docs/部署上报机制说明.md](docs/部署上报机制说明.md) - Deployment reporting
- [具体要求.md](具体要求.md) - Original requirements (Chinese)
- [说明.md](说明.md) - Project goals and scope

## Testing

Test scripts in [demo/](demo/) directory cover all major functionality. Run tests after deployment to verify system health.

## Common Gotchas

1. **Namespace naming**: Must follow pattern `{tenant_org_id}-{user}-{service_name}` unless explicitly overridden
2. **Metadata files**: Check both database (if `USE_POSTGRES=true`) and filesystem for metadata
3. **Auto-scaling cooldown**: Don't expect immediate scaling; cooldown periods prevent thrashing
4. **Shard controller**: Only prints configs by default; actual API calls implemented in latest version
5. **ES plugin compatibility**: Built for Elasticsearch 8.x+
6. **Kubernetes context**: Ensure kubectl is configured for correct cluster before running scripts

## Database Schema (PostgreSQL Mode)

When `USE_POSTGRES=true`, the system uses PostgreSQL with schema in [server/metadata.go](server/metadata.go):

- `tenant_containers` - Tenant container records
- `deployment_status` - Deployment tracking
- `index_metadata` - Vector index metadata
- `tenant_quotas` - Resource quotas per tenant
- `monitoring_metrics` - Time-series metrics

All tables include `tenant_org_id` for multi-tenant isolation.
