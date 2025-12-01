#!/bin/bash

# ES Serverless - 部署上报查询演示脚本
# 本脚本演示如何查询和分析创建容器过程中的上报记录

set -e

BASE_URL=${BASE_URL:-http://localhost:8080}
TEST_USER=${TEST_USER:-testuser}
TEST_SERVICE=${TEST_SERVICE:-test-service}
TEST_NAMESPACE=${TEST_NAMESPACE:-es-test}

echo "======================================"
echo "ES Serverless 部署上报查询演示"
echo "======================================"
echo ""

# 1. 创建一个测试集群
echo "📋 步骤 1: 创建测试集群..."
echo ""

curl -s -X POST $BASE_URL/clusters \
  -H "Content-Type: application/json" \
  -d '{
    "user": "'$TEST_USER'",
    "service_name": "'$TEST_SERVICE'",
    "namespace": "'$TEST_NAMESPACE'",
    "replicas": 1,
    "cpu_request": "500m",
    "cpu_limit": "2",
    "mem_request": "1Gi",
    "mem_limit": "2Gi",
    "disk_size": "10Gi",
    "gpu_count": 0,
    "dimension": 128,
    "vector_count": 10000,
    "index_limit": 100
  }' | jq '.'

echo ""
echo "✅ 创建请求已发送"
echo ""

# 等待一段时间让部署完成
echo "⏳ 等待10秒让部署过程进行..."
sleep 10
echo ""

# 2. 查询部署报告
echo "======================================"
echo "📊 步骤 2: 查询部署报告"
echo "======================================"
echo ""

echo "2.1 通过API查询最新的部署报告："
echo ""
curl -s -X GET "$BASE_URL/deployment/reports/$TEST_USER/$TEST_SERVICE" | jq '.'
echo ""

# 3. 查询部署状态
echo "======================================"
echo "📈 步骤 3: 查询部署状态"
echo "======================================"
echo ""

echo "3.1 查询特定部署的状态："
echo ""
curl -s -X GET "$BASE_URL/deployments?user=$TEST_USER&service_name=$TEST_SERVICE" | jq '.'
echo ""

# 4. 查询租户容器信息
echo "======================================"
echo "🏢 步骤 4: 查询租户容器信息"
echo "======================================"
echo ""

echo "4.1 查询特定租户容器："
echo ""
curl -s -X GET "$BASE_URL/tenant/containers/$TEST_USER/$TEST_SERVICE" | jq '.'
echo ""

# 5. 查看文件系统中的上报记录
echo "======================================"
echo "📁 步骤 5: 查看文件系统中的上报记录"
echo "======================================"
echo ""

echo "5.1 列出该服务的所有部署报告文件："
echo ""
if [ -d "server/deployment_reports" ]; then
    ls -lt server/deployment_reports/${TEST_USER}_${TEST_SERVICE}_*.json 2>/dev/null || echo "暂无报告文件"
else
    echo "报告目录不存在"
fi
echo ""

echo "5.2 查看最新的部署报告文件内容："
echo ""
if [ -d "server/deployment_reports" ]; then
    latest_report=$(ls -t server/deployment_reports/${TEST_USER}_${TEST_SERVICE}_*.json 2>/dev/null | head -1)
    if [ -n "$latest_report" ]; then
        cat "$latest_report" | jq '.'
    else
        echo "暂无报告文件"
    fi
else
    echo "报告目录不存在"
fi
echo ""

echo "5.3 查看部署日志（最后20行）："
echo ""
if [ -f "/tmp/deployment.log" ]; then
    tail -20 /tmp/deployment.log | grep "$TEST_USER.*$TEST_SERVICE" || echo "暂无相关日志"
else
    echo "日志文件不存在"
fi
echo ""

# 6. 分析上报步骤
echo "======================================"
echo "🔍 步骤 6: 分析上报步骤"
echo "======================================"
echo ""

echo "6.1 统计各状态的上报次数："
echo ""
if [ -d "server/deployment_reports" ]; then
    echo "状态                      上报次数"
    echo "----------------------------------------"
    for status in starting namespace_created gitlab_pulled k8s_applied resources_configured disk_configured gpu_configured rollout_completed tenant_synced completed; do
        count=$(grep -l "\"status\": \"$status\"" server/deployment_reports/${TEST_USER}_${TEST_SERVICE}_*.json 2>/dev/null | wc -l)
        printf "%-25s %d\n" "$status" "$count"
    done
else
    echo "报告目录不存在"
fi
echo ""

# 7. 查看时序图
echo "======================================"
echo "📊 步骤 7: 查看创建流程时序图"
echo "======================================"
echo ""

echo "7.1 时序图位置："
echo "    /docs/时序图集合.md - 第1节：创建容器组"
echo ""

echo "7.2 详细上报机制说明："
echo "    /docs/部署上报机制说明.md"
echo ""

# 8. 实时监控上报过程
echo "======================================"
echo "🔴 步骤 8: 实时监控上报过程"
echo "======================================"
echo ""

echo "如果需要实时监控部署上报过程，可以运行："
echo ""
echo "  tail -f /tmp/deployment.log | grep '$TEST_USER.*$TEST_SERVICE'"
echo ""
echo "或者监控报告文件变化："
echo ""
echo "  watch -n 1 'ls -lt server/deployment_reports/${TEST_USER}_${TEST_SERVICE}_*.json | head -5'"
echo ""

# 9. 查询完整部署历史
echo "======================================"
echo "📜 步骤 9: 查询完整部署历史"
echo "======================================"
echo ""

echo "9.1 该服务的所有部署报告（按时间倒序）："
echo ""
if [ -d "server/deployment_reports" ]; then
    echo "时间戳          状态                      消息"
    echo "--------------------------------------------------------------------------------"
    for report in $(ls -t server/deployment_reports/${TEST_USER}_${TEST_SERVICE}_*.json 2>/dev/null); do
        timestamp=$(basename "$report" | sed 's/.*_\([0-9]*\).json/\1/')
        status=$(jq -r '.status' "$report" 2>/dev/null || echo "N/A")
        message=$(jq -r '.message' "$report" 2>/dev/null || echo "N/A")
        date_str=$(date -r "$timestamp" "+%Y-%m-%d %H:%M:%S" 2>/dev/null || echo "$timestamp")
        printf "%-15s %-25s %s\n" "$date_str" "$status" "$message"
    done
else
    echo "报告目录不存在"
fi
echo ""

# 10. 清理测试资源（可选）
echo "======================================"
echo "🗑️  步骤 10: 清理测试资源"
echo "======================================"
echo ""

read -p "是否删除测试集群？(y/N) " -n 1 -r
echo ""
if [[ $REPLY =~ ^[Yy]$ ]]; then
    echo "正在删除测试集群..."
    curl -s -X DELETE "$BASE_URL/clusters/$TEST_NAMESPACE" | jq '.'
    echo ""
    echo "✅ 测试集群已删除"
else
    echo "保留测试集群"
fi
echo ""

# 总结
echo "======================================"
echo "📋 上报机制总结"
echo "======================================"
echo ""
echo "✅ 上报步骤：10次（从starting到completed）"
echo "✅ 存储位置："
echo "   - 部署报告：server/deployment_reports/"
echo "   - 部署日志：/tmp/deployment.log"
echo "   - 元数据服务：server/deployments.json"
echo "   - 租户数据：server/tenant_data/"
echo ""
echo "✅ 查询方式："
echo "   - API查询：GET $BASE_URL/deployment/reports/{user}/{service}"
echo "   - 文件查询：cat server/deployment_reports/{user}_{service}_{timestamp}.json"
echo "   - 日志查询：tail -f /tmp/deployment.log"
echo ""
echo "📖 详细文档："
echo "   - /docs/部署上报机制说明.md"
echo "   - /docs/时序图集合.md"
echo ""
echo "======================================"
echo "演示完成！"
echo "======================================"
