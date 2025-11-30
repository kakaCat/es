#!/usr/bin/env bash
set -euo pipefail

# Deploy script for ES Serverless system

ACTION=${1:-help}
NAMESPACE=${NAMESPACE:-es-serverless}

show_help() {
    echo "Usage: scripts/deploy.sh [ACTION]"
    echo ""
    echo "Actions:"
    echo "  help        - Show this help message"
    echo "  install     - Install the entire system"
    echo "  uninstall   - Uninstall the entire system"
    echo "  status      - Show system status"
    echo ""
    echo "Environment variables:"
    echo "  NAMESPACE   - Kubernetes namespace (default: es-serverless)"
}

install_system() {
    echo "Installing ES Serverless system in namespace: ${NAMESPACE}"
    
    # Apply Kubernetes manifests
    kubectl apply -k k8s/overlays/dev --validate=false
    
    # Wait for Elasticsearch to be ready
    echo "Waiting for Elasticsearch to be ready..."
    kubectl -n "$NAMESPACE" rollout status sts/elasticsearch --timeout=300s
    
    # Wait for Kibana to be ready
    echo "Waiting for Kibana to be ready..."
    kubectl -n "$NAMESPACE" rollout status deploy/kibana --timeout=300s
    
    # Wait for Manager to be ready
    echo "Waiting for Manager to be ready..."
    kubectl -n "$NAMESPACE" rollout status deploy/es-serverless-manager --timeout=300s
    
    # Wait for Shard Controller to be ready
    echo "Waiting for Shard Controller to be ready..."
    kubectl -n "$NAMESPACE" rollout status deploy/shard-controller --timeout=300s
    
    # Wait for Reporting Service to be ready
    echo "Waiting for Reporting Service to be ready..."
    kubectl -n "$NAMESPACE" rollout status deploy/reporting-service --timeout=300s
    
    echo "ES Serverless system installed successfully!"
    echo "Access the services:"
    echo "  Elasticsearch: kubectl -n $NAMESPACE port-forward svc/elasticsearch 9200:9200"
    echo "  Kibana: kubectl -n $NAMESPACE port-forward svc/kibana 5601:5601"
    echo "  Manager API: kubectl -n $NAMESPACE port-forward svc/es-serverless-manager 8080:8080"
    echo "  Reporting Service: kubectl -n $NAMESPACE port-forward svc/reporting-service 8081:8080"
}

uninstall_system() {
    echo "Uninstalling ES Serverless system from namespace: ${NAMESPACE}"
    kubectl delete ns "$NAMESPACE" --ignore-not-found=true
    echo "ES Serverless system uninstalled successfully!"
}

show_status() {
    echo "ES Serverless system status in namespace: ${NAMESPACE}"
    echo ""
    echo "Namespaces:"
    kubectl get ns | grep "$NAMESPACE" || echo "No namespaces found"
    echo ""
    echo "StatefulSets:"
    kubectl -n "$NAMESPACE" get sts 2>/dev/null || echo "No statefulsets found"
    echo ""
    echo "Deployments:"
    kubectl -n "$NAMESPACE" get deploy 2>/dev/null || echo "No deployments found"
    echo ""
    echo "Services:"
    kubectl -n "$NAMESPACE" get svc 2>/dev/null || echo "No services found"
    echo ""
    echo "Pods:"
    kubectl -n "$NAMESPACE" get pods 2>/dev/null || echo "No pods found"
    echo ""
    echo "Reporting Service Logs:"
    kubectl -n "$NAMESPACE" logs -l app=reporting-service --tail=20 2>/dev/null || echo "No reporting service logs found"
}

case "$ACTION" in
    help)
        show_help
        ;;
    install)
        install_system
        ;;
    uninstall)
        uninstall_system
        ;;
    status)
        show_status
        ;;
    *)
        echo "Unknown action: $ACTION"
        show_help
        exit 1
        ;;
esac