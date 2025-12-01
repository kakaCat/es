-- Create tables for ES Serverless metadata storage

-- Tenant Containers table
CREATE TABLE IF NOT EXISTS tenant_containers (
    id VARCHAR(255) PRIMARY KEY,
    tenant_org_id VARCHAR(255) NOT NULL,
    user_id VARCHAR(255) NOT NULL,
    service_name VARCHAR(255) NOT NULL,
    namespace VARCHAR(255) NOT NULL,
    replicas INTEGER NOT NULL,
    cpu VARCHAR(50) NOT NULL,
    memory VARCHAR(50) NOT NULL,
    disk VARCHAR(50) NOT NULL,
    gpu_count INTEGER NOT NULL,
    dimension INTEGER NOT NULL,
    vector_count INTEGER NOT NULL,
    status VARCHAR(50) NOT NULL,
    created_at TIMESTAMP NOT NULL,
    sync_time TIMESTAMP NOT NULL,
    deleted BOOLEAN NOT NULL DEFAULT FALSE
);

-- Indexes for tenant_containers
CREATE INDEX IF NOT EXISTS idx_tenant_containers_user_service ON tenant_containers(user_id, service_name);
CREATE INDEX IF NOT EXISTS idx_tenant_containers_org ON tenant_containers(tenant_org_id);
CREATE INDEX IF NOT EXISTS idx_tenant_containers_namespace ON tenant_containers(namespace);
CREATE INDEX IF NOT EXISTS idx_tenant_containers_deleted ON tenant_containers(deleted);

-- Index Metadata table
CREATE TABLE IF NOT EXISTS index_metadata (
    id VARCHAR(255) PRIMARY KEY,
    index_name VARCHAR(255) NOT NULL,
    namespace VARCHAR(255) NOT NULL,
    dimension INTEGER NOT NULL,
    metric VARCHAR(50) NOT NULL,
    ivf_nlist INTEGER NOT NULL,
    ivf_nprobe INTEGER NOT NULL,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    created_by VARCHAR(255) NOT NULL,
    status VARCHAR(50) NOT NULL,
    document_count INTEGER NOT NULL,
    storage_size VARCHAR(50) NOT NULL
);

-- Indexes for index_metadata
CREATE INDEX IF NOT EXISTS idx_index_metadata_name ON index_metadata(index_name);
CREATE INDEX IF NOT EXISTS idx_index_metadata_namespace ON index_metadata(namespace);

-- Tenant Quota table
CREATE TABLE IF NOT EXISTS tenant_quota (
    id VARCHAR(255) PRIMARY KEY,
    tenant_id VARCHAR(255) NOT NULL UNIQUE,
    max_indices INTEGER NOT NULL,
    max_storage VARCHAR(50) NOT NULL,
    current_indices INTEGER NOT NULL,
    current_storage VARCHAR(50) NOT NULL,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
);

-- Deployment Status table
CREATE TABLE IF NOT EXISTS deployment_status (
    id VARCHAR(255) PRIMARY KEY,
    tenant_org_id VARCHAR(255) NOT NULL,
    namespace VARCHAR(255) NOT NULL UNIQUE,
    user_id VARCHAR(255) NOT NULL,
    service_name VARCHAR(255) NOT NULL,
    status VARCHAR(50) NOT NULL,
    cpu_usage DOUBLE PRECISION NOT NULL,
    memory_usage DOUBLE PRECISION NOT NULL,
    disk_usage DOUBLE PRECISION NOT NULL,
    qps DOUBLE PRECISION NOT NULL,
    gpu_count INTEGER NOT NULL,
    dimension INTEGER NOT NULL,
    vector_count INTEGER NOT NULL,
    replicas INTEGER NOT NULL,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
);

-- Indexes for deployment_status
CREATE INDEX IF NOT EXISTS idx_deployment_status_namespace ON deployment_status(namespace);
CREATE INDEX IF NOT EXISTS idx_deployment_status_user ON deployment_status(user_id);

-- Monitoring Metrics table
CREATE TABLE IF NOT EXISTS monitoring_metrics (
    id VARCHAR(255) PRIMARY KEY,
    namespace VARCHAR(255) NOT NULL,
    cpu_usage DOUBLE PRECISION NOT NULL,
    memory_usage DOUBLE PRECISION NOT NULL,
    disk_usage DOUBLE PRECISION NOT NULL,
    qps DOUBLE PRECISION NOT NULL,
    timestamp TIMESTAMP NOT NULL
);

-- Indexes for monitoring_metrics
CREATE INDEX IF NOT EXISTS idx_monitoring_metrics_namespace ON monitoring_metrics(namespace);
CREATE INDEX IF NOT EXISTS idx_monitoring_metrics_timestamp ON monitoring_metrics(timestamp);

-- Container Metrics table
CREATE TABLE IF NOT EXISTS container_metrics (
    id VARCHAR(255) PRIMARY KEY,
    namespace VARCHAR(255) NOT NULL,
    cpu_usage DOUBLE PRECISION NOT NULL,
    memory_usage DOUBLE PRECISION NOT NULL,
    disk_usage DOUBLE PRECISION NOT NULL,
    plugin_qps DOUBLE PRECISION NOT NULL,
    timestamp TIMESTAMP NOT NULL
);

-- Indexes for container_metrics
CREATE INDEX IF NOT EXISTS idx_container_metrics_namespace ON container_metrics(namespace);
CREATE INDEX IF NOT EXISTS idx_container_metrics_timestamp ON container_metrics(timestamp);