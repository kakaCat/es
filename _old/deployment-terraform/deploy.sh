#!/usr/bin/env bash
set -euo pipefail

# Terraform + Helm 部署快捷脚本

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TERRAFORM_DIR="$SCRIPT_DIR/terraform"

show_help() {
    cat <<EOF
Terraform + Helm 部署脚本

用法: $0 [命令]

命令:
  init        初始化 Terraform
  plan        预览基础设施变更
  apply       应用基础设施变更（部署）
  destroy     销毁所有基础设施
  output      显示 Terraform 输出
  status      显示部署状态
  help        显示此帮助信息

示例:
  # 首次部署
  $0 init
  $0 plan
  $0 apply

  # 查看状态
  $0 status
  $0 output

  # 清理
  $0 destroy

详细文档: $SCRIPT_DIR/README.md
EOF
}

init_terraform() {
    echo "🔧 初始化 Terraform..."
    cd "$TERRAFORM_DIR"
    terraform init
    echo "✅ Terraform 初始化完成"
}

plan_terraform() {
    echo "📋 预览基础设施变更..."
    cd "$TERRAFORM_DIR"
    terraform plan
}

apply_terraform() {
    echo "🚀 应用 Terraform 配置..."
    cd "$TERRAFORM_DIR"
    terraform apply
}

destroy_terraform() {
    echo "🗑️  销毁基础设施..."
    echo "⚠️  警告：这将删除所有资源！"
    read -p "确认继续? (yes/no): " confirm
    if [[ "$confirm" == "yes" ]]; then
        cd "$TERRAFORM_DIR"
        terraform destroy
    else
        echo "取消销毁操作"
    fi
}

show_output() {
    echo "📊 Terraform 输出..."
    cd "$TERRAFORM_DIR"
    terraform output
}

show_status() {
    echo "📊 部署状态..."
    echo ""
    echo "=== Kubernetes 集群信息 ==="
    kubectl cluster-info || echo "无法连接到集群"
    echo ""
    echo "=== 命名空间 ==="
    kubectl get ns | grep -E "NAME|es-|org-" || echo "未找到 ES 相关命名空间"
    echo ""
    echo "=== ES Serverless 资源 ==="
    kubectl get all -n es-serverless 2>/dev/null || echo "es-serverless 命名空间不存在"
}

main() {
    case "${1:-help}" in
        init)
            init_terraform
            ;;
        plan)
            plan_terraform
            ;;
        apply)
            apply_terraform
            ;;
        destroy)
            destroy_terraform
            ;;
        output)
            show_output
            ;;
        status)
            show_status
            ;;
        help|--help|-h)
            show_help
            ;;
        *)
            echo "未知命令: $1"
            echo ""
            show_help
            exit 1
            ;;
    esac
}

main "$@"
