# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

ES Serverless Platform - A serverless ElasticSearch platform with vector search capabilities using IVF (Inverted File) algorithms. Multi-tenant platform providing automatic provisioning, scaling, and management of Elasticsearch clusters with custom vector search plugins.

**Project Version**: v2.0
**Structure**: Reorganized to standard open-source layout (see PROJECT_STRUCTURE.md)

## Essential Commands

### Building and Running

```bash
# Build the control plane Go service
cd src/control-plane
go build -o manager .
./manager

# Build the ES IVF plugin
cd src/es-plugin
./gradlew build
# Output: build/distributions/es-ivf-plugin-*.zip

# Start frontend (development)
cd src/frontend
python -m http.server 8000
# Access at http://localhost:8000
```

### Deployment

```bash
# Deploy using Shell scripts (development/testing)
./scripts/deploy/deploy.sh install
./scripts/deploy/deploy.sh status
./scripts/deploy/deploy.sh uninstall

# Deploy using Terraform + Helm (production)
cd deployments/terraform
terraform init
terraform plan
terraform apply
```

### Cluster Management

```bash
# Create a cluster
./scripts/deploy/cluster.sh create <namespace> <replicas>

# Delete a cluster
./scripts/deploy/cluster.sh delete <namespace>

# Get cluster status
./scripts/deploy/cluster.sh status <namespace>

# Scale a cluster
./scripts/deploy/cluster.sh scale <namespace> <replicas>
```

### Testing

```bash
# Run IVF plugin tests
cd src/es-plugin
./gradlew test

# Run control plane tests
cd src/control-plane
go test ./...

# Integration tests
cd examples
# Run test scripts
```

## Architecture Overview

### Three-Layer Architecture

1. **Control Plane** (`src/control-plane/`)
   - Written in Go
   - Manages cluster lifecycle, auto-scaling, monitoring
   - Exposes REST API on port 8080
   - Components: Manager, AutoScaler, ShardController, ReplicationMonitor, ConsistencyChecker

2. **Data Plane** (Kubernetes-based)
   - Elasticsearch clusters deployed as StatefulSets
   - Each cluster runs in isolated namespaces
   - Custom IVF plugin for vector search
   - Monitoring via Prometheus/Grafana

3. **Plugin Layer** (`src/es-plugin/`)
   - Gradle-based Elasticsearch plugin (Java)
   - Implements IVF algorithm for vector search
   - Custom field types and query DSL

### Key Go Components

Located in `src/control-plane/`:

- **main.go** - HTTP API handlers, cluster orchestration, main entry point
- **metadata.go** - Metadata storage abstraction (PostgreSQL or filesystem)
- **autoscaler.go** - Auto-scaling logic with cooldown periods and trend analysis
- **shard_controller.go** - Shard rebalancing and allocation
- **replication_monitor.go** - Replica synchronization monitoring
- **consistency_checker.go** - Data consistency verification
- **auto_recovery.go** - Automatic failure recovery
- **monitoring.go** - Metrics collection and aggregation
- **reporting.go** - Deployment status reporting
- **es_client.go** - Elasticsearch API client

## Multi-Tenancy Model

Three-level tenant hierarchy:

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

**CRITICAL**: All cluster creation API requests MUST include `tenant_org_id`, `user`, and `service_name`.

## REST API Endpoints

The control plane (`src/control-plane/main.go`) exposes:

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

Dual storage modes (configured at startup):

1. **File-based storage** (default): `src/control-plane/data/` directory
2. **PostgreSQL storage**: Set `USE_POSTGRES=true` environment variable

**File-based storage structure:**
- `tenant_{user}_{service}.json` - Tenant container metadata
- `deploy_{namespace}.json` - Deployment status
- `index_{namespace}_{index_name}.json` - Index metadata
- `metrics_{namespace}.json` - Monitoring metrics

**PostgreSQL schema** (see `src/control-plane/metadata.go`):
- `tenant_containers` - Tenant container records
- `deployment_status` - Deployment tracking
- `index_metadata` - Vector index metadata
- `tenant_quotas` - Resource quotas per tenant
- `monitoring_metrics` - Time-series metrics

All tables include `tenant_org_id` for multi-tenant isolation.

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

Auto-scaling (`src/control-plane/autoscaler.go`) implements:

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

- **ShardController** (`src/control-plane/shard_controller.go`) - Monitors every 30s, triggers rebalancing
- **ReplicationMonitor** (`src/control-plane/replication_monitor.go`) - Monitors replica sync every 10s
- **ConsistencyChecker** (`src/control-plane/consistency_checker.go`) - Verifies data consistency
- **AutoRecoveryManager** (`src/control-plane/auto_recovery.go`) - Automatic recovery every 30s

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

Located in `src/es-plugin/`, the plugin implements:

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
cd src/es-plugin
./gradlew build
# Output: build/distributions/es-ivf-plugin-*.zip
```

## Deployment Architecture

The system deploys on **Kubernetes** (Docker Desktop or production cluster):

1. **Prerequisites:**
   - Docker Desktop with Kubernetes enabled (local)
   - kubectl CLI
   - Go 1.21+
   - Java 11+ and Gradle (for plugin development)

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

See [docs/architecture/auto-scaling.md](docs/architecture/auto-scaling.md)

### Logical Deletion

Tenant containers use **logical deletion**:
- Set `deleted: true` flag
- Preserve `deleted_at` timestamp
- Retain for audit/recovery
- Physical cleanup is separate process

See [docs/operations/logical-deletion.md](docs/operations/logical-deletion.md)

### Error Handling for Multi-Tenancy

Return HTTP 400 if missing required fields:
- `tenant_org_id is required for multi-tenancy`
- `user is required`
- `service_name is required`

## Load Balancing

4-layer load balancing architecture:

1. **L4 Network Layer**: Kubernetes Service with Round-Robin (ClusterIP)
2. **L7 Application Layer**: Elasticsearch coordinating nodes with intelligent routing
3. **Data Layer**: ShardController hot shard migration (every 30s check)
4. **Auto-scaling Layer**: HPA + ShardController dynamic node scaling

## Key Documentation Files

### Core Documentation
- [README.md](README.md) - Project overview and quickstart
- [PROJECT_STRUCTURE.md](PROJECT_STRUCTURE.md) - Directory structure explanation
- [CONTRIBUTING.md](CONTRIBUTING.md) - Development guidelines
- [MIGRATION_GUIDE.md](MIGRATION_GUIDE.md) - v1.0 to v2.0 migration guide

### Architecture Documentation
- [docs/architecture.md](docs/architecture.md) - Detailed architecture diagrams
- [docs/architecture/multi-tenancy.md](docs/architecture/multi-tenancy.md) - Multi-tenancy design
- [docs/architecture/shard-replication.md](docs/architecture/shard-replication.md) - Shard replication design
- [docs/architecture/auto-scaling.md](docs/architecture/auto-scaling.md) - Quota management and auto-scaling

### Deployment Documentation
- [docs/deployment/README.md](docs/deployment/README.md) - Deployment overview
- [docs/deployment/terraform-helm.md](docs/deployment/terraform-helm.md) - Terraform + Helm deployment
- [docs/deployment/shell-scripts.md](docs/deployment/shell-scripts.md) - Shell script deployment

### Operations Documentation
- [docs/operations/monitoring.md](docs/operations/monitoring.md) - Monitoring and alerting
- [docs/operations/disaster-recovery.md](docs/operations/disaster-recovery.md) - Backup and recovery
- [docs/operations/deployment-reporting.md](docs/operations/deployment-reporting.md) - Deployment reporting
- [docs/operations/logical-deletion.md](docs/operations/logical-deletion.md) - Logical deletion mechanism

## Common Gotchas

1. **Path References**: Project was restructured in v2.0. Use new paths:
   - `server/` → `src/control-plane/`
   - `es-plugin/` → `src/es-plugin/`
   - `frontend/` → `src/frontend/`
   - `scripts/*.sh` → `scripts/{deploy,build,ops,dev}/`

2. **Namespace Naming**: Must follow pattern `{tenant_org_id}-{user}-{service_name}` unless explicitly overridden

3. **Metadata Files**: Check both database (if `USE_POSTGRES=true`) and filesystem for metadata

4. **Auto-scaling Cooldown**: Don't expect immediate scaling; cooldown periods prevent thrashing (scale-up: 5min, scale-down: 10min)

5. **Shard Controller**: Only prints configs by default; actual API calls implemented in latest version

6. **ES Plugin Compatibility**: Built for Elasticsearch 8.x+

7. **Kubernetes Context**: Ensure kubectl is configured for correct cluster before running scripts

8. **Old Files**: Files in `_old/` directory are deprecated backups from v1.0, scheduled for deletion in v3.0

## Project Structure Notes

The project follows a standard open-source layout (v2.0):
- `src/` - All source code
- `deployments/` - All deployment configurations
- `scripts/` - Tool scripts organized by function
- `docs/` - Documentation center organized by type
- `tests/` - Test code
- `examples/` - Usage examples
- `_old/` - Deprecated files from v1.0 (will be removed)

For detailed structure explanation, see [PROJECT_STRUCTURE.md](PROJECT_STRUCTURE.md).
