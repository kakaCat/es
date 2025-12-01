#!/usr/bin/env bash
set -euo pipefail

# Monitor script for ES Serverless system

ACTION=${1:-help}
NAMESPACE=${NAMESPACE:-es-serverless}

show_help() {
    echo "Usage: scripts/monitor.sh [ACTION]"
    echo ""
    echo "Actions:"
    echo "  help        - Show this help message"
    echo "  port-forward-prometheus - Port forward Prometheus service"
    echo "  port-forward-grafana   - Port forward Grafana service"
    echo "  port-forward-kibana    - Port forward Kibana service"
    echo "  logs           - Show logs for Elasticsearch pods"
    echo "  metrics        - Show current metrics for Elasticsearch"
    echo ""
    echo "Environment variables:"
    echo "  NAMESPACE   - Kubernetes namespace (default: es-serverless)"
}

port_forward_prometheus() {
    echo "Port forwarding Prometheus service..."
    echo "Access Prometheus at http://localhost:9090"
    kubectl -n "$NAMESPACE" port-forward svc/prometheus-service 9090:9090
}

port_forward_grafana() {
    echo "Port forwarding Grafana service..."
    echo "Access Grafana at http://localhost:3000"
    echo "Default credentials: admin/admin"
    kubectl -n "$NAMESPACE" port-forward svc/grafana-service 3000:3000
}

port_forward_kibana() {
    echo "Port forwarding Kibana service..."
    echo "Access Kibana at http://localhost:5601"
    kubectl -n "$NAMESPACE" port-forward svc/kibana 5601:5601
}

show_logs() {
    echo "Showing logs for Elasticsearch pods..."
    kubectl -n "$NAMESPACE" logs -f sts/elasticsearch
}

show_metrics() {
    echo "Showing current metrics for Elasticsearch..."
    kubectl -n "$NAMESPACE" top pods
}

case "$ACTION" in
    help)
        show_help
        ;;
    port-forward-prometheus)
        port_forward_prometheus
        ;;
    port-forward-grafana)
        port_forward_grafana
        ;;
    port-forward-kibana)
        port_forward_kibana
        ;;
    logs)
        show_logs
        ;;
    metrics)
        show_metrics
        ;;
    *)
        echo "Unknown action: $ACTION"
        show_help
        exit 1
        ;;
esac