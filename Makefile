.PHONY: help init plan apply destroy status clean tenant-create tenant-list

# 默认目标
.DEFAULT_GOAL := help

# 颜色定义
GREEN  := $(shell tput -Txterm setaf 2)
YELLOW := $(shell tput -Txterm setaf 3)
BLUE   := $(shell tput -Txterm setaf 6)
RESET  := $(shell tput -Txterm sgr0)

# 配置
TERRAFORM_DIR := terraform
NAMESPACE     := es-serverless

##@ 帮助

help: ## 显示帮助信息
	@echo ''
	@echo '$(BLUE)ES Serverless Platform - Terraform/Helm 管理工具$(RESET)'
	@echo ''
	@echo '使用方式:'
	@echo '  make $(GREEN)<target>$(RESET)'
	@echo ''
	@awk 'BEGIN {FS = ":.*##"; printf "可用命令:\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  $(GREEN)%-20s$(RESET) %s\n", $$1, $$2 } /^##@/ { printf "\n$(BLUE)%s$(RESET)\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ 平台管理

init: ## 初始化 Terraform
	@echo "$(BLUE)初始化 Terraform...$(RESET)"
	cd $(TERRAFORM_DIR) && terraform init
	@echo "$(GREEN)✓ 初始化完成$(RESET)"

plan: ## 查看部署计划
	@echo "$(BLUE)生成执行计划...$(RESET)"
	cd $(TERRAFORM_DIR) && terraform plan
	@echo "$(GREEN)✓ 计划生成完成$(RESET)"

apply: ## 部署平台
	@echo "$(BLUE)部署 ES Serverless 平台...$(RESET)"
	cd $(TERRAFORM_DIR) && terraform apply
	@echo "$(GREEN)✓ 部署完成$(RESET)"
	@make show-urls

apply-auto: ## 自动部署 (不需要确认)
	@echo "$(BLUE)自动部署 ES Serverless 平台...$(RESET)"
	cd $(TERRAFORM_DIR) && terraform apply -auto-approve
	@echo "$(GREEN)✓ 部署完成$(RESET)"
	@make show-urls

destroy: ## 销毁所有资源
	@echo "$(YELLOW)警告: 这将销毁所有 Terraform 管理的资源!$(RESET)"
	@read -p "确认继续? (yes/no): " confirm; \
	if [ "$$confirm" = "yes" ]; then \
		cd $(TERRAFORM_DIR) && terraform destroy; \
		echo "$(GREEN)✓ 销毁完成$(RESET)"; \
	else \
		echo "$(YELLOW)已取消$(RESET)"; \
	fi

status: ## 查看系统状态
	@echo "$(BLUE)ES Serverless 平台状态:$(RESET)"
	@echo ""
	@echo "Terraform 状态:"
	@cd $(TERRAFORM_DIR) && terraform show -json | jq -r '.values.outputs | to_entries[] | "  \(.key): \(.value.value)"' || echo "  未部署"
	@echo ""
	@echo "Kubernetes 资源:"
	@echo "  Namespaces:"
	@kubectl get ns -l app.kubernetes.io/name=es-serverless 2>/dev/null || echo "    无"
	@echo "  Helm Releases:"
	@helm list -n $(NAMESPACE) 2>/dev/null || echo "    无"
	@echo "  Pods:"
	@kubectl get pods -n $(NAMESPACE) 2>/dev/null || echo "    无"

show-urls: ## 显示服务访问地址
	@echo ""
	@echo "$(BLUE)服务访问地址:$(RESET)"
	@cd $(TERRAFORM_DIR) && terraform output -json 2>/dev/null | jq -r '
		"  Elasticsearch:  kubectl -n " + .namespace.value + " port-forward svc/elasticsearch 9200:9200",
		"  Manager API:    kubectl -n " + .namespace.value + " port-forward svc/es-control-plane-manager 8080:8080",
		"  Grafana:        kubectl -n " + .namespace.value + " port-forward svc/monitoring-grafana 3000:3000",
		"  Prometheus:     kubectl -n " + .namespace.value + " port-forward svc/monitoring-prometheus 9090:9090"
	' || echo "  平台未部署"

##@ 租户管理

tenant-create: ## 创建租户 (需要参数: ORG, USER, SERVICE)
ifndef ORG
	@echo "$(YELLOW)错误: 缺少 ORG 参数$(RESET)"
	@echo "使用方式: make tenant-create ORG=org-001 USER=alice SERVICE=vector-search"
	@exit 1
endif
ifndef USER
	@echo "$(YELLOW)错误: 缺少 USER 参数$(RESET)"
	@echo "使用方式: make tenant-create ORG=org-001 USER=alice SERVICE=vector-search"
	@exit 1
endif
ifndef SERVICE
	@echo "$(YELLOW)错误: 缺少 SERVICE 参数$(RESET)"
	@echo "使用方式: make tenant-create ORG=org-001 USER=alice SERVICE=vector-search"
	@exit 1
endif
	@echo "$(BLUE)创建租户: $(ORG)-$(USER)-$(SERVICE)$(RESET)"
	./scripts/create-tenant.sh \
		--org $(ORG) \
		--user $(USER) \
		--service $(SERVICE) \
		$(if $(CPU),--cpu $(CPU),) \
		$(if $(MEMORY),--memory $(MEMORY),) \
		$(if $(DISK),--disk $(DISK),) \
		$(if $(GPU),--gpu $(GPU),) \
		$(if $(DIMENSION),--dimension $(DIMENSION),) \
		$(if $(VECTORS),--vectors $(VECTORS),) \
		$(if $(REPLICAS),--replicas $(REPLICAS),)
	@echo "$(GREEN)✓ 租户创建完成$(RESET)"

tenant-list: ## 列出所有租户
	@echo "$(BLUE)租户列表:$(RESET)"
	@kubectl get ns -l es-cluster=true -o custom-columns=\
NAME:.metadata.name,\
ORG:.metadata.labels.tenant-org-id,\
USER:.metadata.labels.user,\
SERVICE:.metadata.labels.service-name,\
CREATED:.metadata.creationTimestamp \
		2>/dev/null || echo "  无租户"

tenant-status: ## 查看租户状态 (需要参数: TENANT)
ifndef TENANT
	@echo "$(YELLOW)错误: 缺少 TENANT 参数$(RESET)"
	@echo "使用方式: make tenant-status TENANT=org-001-alice-vector-search"
	@exit 1
endif
	@echo "$(BLUE)租户状态: $(TENANT)$(RESET)"
	@echo ""
	@echo "Pods:"
	@kubectl get pods -n $(TENANT) 2>/dev/null || echo "  无"
	@echo ""
	@echo "Services:"
	@kubectl get svc -n $(TENANT) 2>/dev/null || echo "  无"
	@echo ""
	@echo "PVCs:"
	@kubectl get pvc -n $(TENANT) 2>/dev/null || echo "  无"
	@echo ""
	@echo "元数据:"
	@kubectl get configmap tenant-metadata -n $(TENANT) -o yaml 2>/dev/null | grep -A 20 "^data:" || echo "  无"

tenant-delete: ## 删除租户 (需要参数: TENANT)
ifndef TENANT
	@echo "$(YELLOW)错误: 缺少 TENANT 参数$(RESET)"
	@echo "使用方式: make tenant-delete TENANT=org-001-alice-vector-search"
	@exit 1
endif
	@echo "$(YELLOW)警告: 这将删除租户 $(TENANT) 的所有资源!$(RESET)"
	@read -p "确认继续? (yes/no): " confirm; \
	if [ "$$confirm" = "yes" ]; then \
		cd $(TERRAFORM_DIR)/tenants/$(TENANT) && terraform destroy -auto-approve; \
		echo "$(GREEN)✓ 租户删除完成$(RESET)"; \
	else \
		echo "$(YELLOW)已取消$(RESET)"; \
	fi

##@ 监控和日志

logs-manager: ## 查看 Manager 日志
	@echo "$(BLUE)Manager 日志 (Ctrl+C 退出):$(RESET)"
	kubectl -n $(NAMESPACE) logs -l app=es-control-plane-manager -f --tail=50

logs-elasticsearch: ## 查看 Elasticsearch 日志
	@echo "$(BLUE)Elasticsearch 日志 (Ctrl+C 退出):$(RESET)"
	kubectl -n $(NAMESPACE) logs elasticsearch-0 -f --tail=50

logs-shard: ## 查看 Shard Controller 日志
	@echo "$(BLUE)Shard Controller 日志 (Ctrl+C 退出):$(RESET)"
	kubectl -n $(NAMESPACE) logs -l component=shard-controller -f --tail=50

logs-tenant: ## 查看租户日志 (需要参数: TENANT)
ifndef TENANT
	@echo "$(YELLOW)错误: 缺少 TENANT 参数$(RESET)"
	@echo "使用方式: make logs-tenant TENANT=org-001-alice-vector-search"
	@exit 1
endif
	@echo "$(BLUE)租户 $(TENANT) 日志 (Ctrl+C 退出):$(RESET)"
	kubectl -n $(TENANT) logs -l app=elasticsearch -f --tail=50

metrics: ## 查看资源使用情况
	@echo "$(BLUE)节点资源使用:$(RESET)"
	@kubectl top nodes
	@echo ""
	@echo "$(BLUE)平台 Pod 资源使用:$(RESET)"
	@kubectl top pods -n $(NAMESPACE) 2>/dev/null || echo "  Metrics server 未安装"

##@ 访问服务

port-forward-manager: ## 端口转发 Manager API (8080)
	@echo "$(BLUE)转发 Manager API 到 localhost:8080$(RESET)"
	@echo "访问: http://localhost:8080"
	kubectl -n $(NAMESPACE) port-forward svc/es-control-plane-manager 8080:8080

port-forward-grafana: ## 端口转发 Grafana (3000)
	@echo "$(BLUE)转发 Grafana 到 localhost:3000$(RESET)"
	@echo "访问: http://localhost:3000 (admin/admin)"
	kubectl -n $(NAMESPACE) port-forward svc/monitoring-grafana 3000:3000

port-forward-prometheus: ## 端口转发 Prometheus (9090)
	@echo "$(BLUE)转发 Prometheus 到 localhost:9090$(RESET)"
	@echo "访问: http://localhost:9090"
	kubectl -n $(NAMESPACE) port-forward svc/monitoring-prometheus 9090:9090

port-forward-es: ## 端口转发 Elasticsearch (9200)
	@echo "$(BLUE)转发 Elasticsearch 到 localhost:9200$(RESET)"
	@echo "访问: http://localhost:9200"
	kubectl -n $(NAMESPACE) port-forward svc/elasticsearch 9200:9200

##@ 开发和测试

validate: ## 验证 Terraform 配置
	@echo "$(BLUE)验证 Terraform 配置...$(RESET)"
	cd $(TERRAFORM_DIR) && terraform validate
	@echo "$(GREEN)✓ 配置有效$(RESET)"

format: ## 格式化 Terraform 代码
	@echo "$(BLUE)格式化 Terraform 代码...$(RESET)"
	cd $(TERRAFORM_DIR) && terraform fmt -recursive
	@echo "$(GREEN)✓ 格式化完成$(RESET)"

lint-helm: ## 检查 Helm Charts
	@echo "$(BLUE)检查 Helm Charts...$(RESET)"
	helm lint helm/elasticsearch
	helm lint helm/control-plane
	helm lint helm/monitoring
	@echo "$(GREEN)✓ 检查完成$(RESET)"

test-cluster: ## 测试集群连接
	@echo "$(BLUE)测试 Kubernetes 连接...$(RESET)"
	@kubectl cluster-info
	@echo ""
	@echo "$(BLUE)测试 Helm...$(RESET)"
	@helm version
	@echo ""
	@echo "$(GREEN)✓ 环境正常$(RESET)"

##@ 清理和维护

clean-state: ## 清理 Terraform 状态文件 (谨慎!)
	@echo "$(YELLOW)警告: 这将删除 Terraform 状态文件!$(RESET)"
	@read -p "确认继续? (yes/no): " confirm; \
	if [ "$$confirm" = "yes" ]; then \
		rm -f $(TERRAFORM_DIR)/terraform.tfstate*; \
		rm -rf $(TERRAFORM_DIR)/.terraform; \
		echo "$(GREEN)✓ 状态文件已清理$(RESET)"; \
	else \
		echo "$(YELLOW)已取消$(RESET)"; \
	fi

clean-tenants: ## 清理所有租户配置文件
	@echo "$(YELLOW)警告: 这将删除所有租户配置文件!$(RESET)"
	@read -p "确认继续? (yes/no): " confirm; \
	if [ "$$confirm" = "yes" ]; then \
		rm -rf $(TERRAFORM_DIR)/tenants/*; \
		echo "$(GREEN)✓ 租户配置文件已清理$(RESET)"; \
	else \
		echo "$(YELLOW)已取消$(RESET)"; \
	fi

backup-state: ## 备份 Terraform 状态
	@echo "$(BLUE)备份 Terraform 状态...$(RESET)"
	@mkdir -p backups
	@cp $(TERRAFORM_DIR)/terraform.tfstate backups/terraform.tfstate.$$(date +%Y%m%d_%H%M%S)
	@echo "$(GREEN)✓ 备份完成: backups/terraform.tfstate.$$(date +%Y%m%d_%H%M%S)$(RESET)"

##@ 快速开始

quick-start: ## 快速开始 (初始化 + 部署)
	@echo "$(BLUE)快速开始 ES Serverless 平台...$(RESET)"
	@make init
	@make apply-auto
	@echo ""
	@echo "$(GREEN)✓ 平台部署完成!$(RESET)"
	@make show-urls

quick-demo: ## 快速演示 (部署平台 + 创建示例租户)
	@echo "$(BLUE)运行快速演示...$(RESET)"
	@make quick-start
	@echo ""
	@echo "$(BLUE)创建示例租户...$(RESET)"
	@make tenant-create ORG=demo USER=alice SERVICE=vector-search CPU=2000m MEMORY=4Gi
	@echo ""
	@echo "$(GREEN)✓ 演示完成!$(RESET)"
	@make tenant-list

##@ 文档

docs: ## 打开文档
	@echo "$(BLUE)可用文档:$(RESET)"
	@echo "  - TERRAFORM_HELM_README.md: 主文档"
	@echo "  - docs/terraform-helm-guide.md: 完整使用指南"
	@echo "  - docs/helm-charts-reference.md: Helm Charts 参考"
	@echo "  - docs/terraform-architecture-diagram.md: 架构图"
