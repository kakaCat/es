// API基础URL
const API_BASE_URL = 'http://localhost:8080';

// DOM元素
const createClusterForm = document.getElementById('create-cluster-form');
const deleteClusterForm = document.getElementById('delete-cluster-form');
const listClustersForm = document.getElementById('list-clusters-form');
const clusterDetailsForm = document.getElementById('cluster-details-form');

// 结果显示区域
const createResult = document.getElementById('create-result');
const deleteResult = document.getElementById('delete-result');
const listResult = document.getElementById('list-result');
const detailsResult = document.getElementById('details-result');

// 创建集群表单提交事件
createClusterForm.addEventListener('submit', async (e) => {
    e.preventDefault();
    
    // 获取表单数据
    const formData = new FormData(createClusterForm);
    const clusterData = {
        tenant_org_id: formData.get('tenant_org_id'),
        user: formData.get('user'),
        service_name: formData.get('service_name'),
        namespace: formData.get('namespace'),
        replicas: parseInt(formData.get('replicas')),
        cpu_request: formData.get('cpu_request'),
        cpu_limit: formData.get('cpu_limit'),
        mem_request: formData.get('mem_request'),
        mem_limit: formData.get('mem_limit'),
        disk_size: formData.get('disk_size'),
        gpu_count: parseInt(formData.get('gpu_count')),
        dimension: parseInt(formData.get('dimension')),
        vector_count: parseInt(formData.get('vector_count')),
        index_limit: parseInt(formData.get('index_limit')),
        gitlab_url: formData.get('gitlab_url')
    };
    
    // 验证必填字段
    if (!clusterData.tenant_org_id) {
        showResult(createResult, '租户组织ID是必填项', 'error');
        return;
    }
    
    if (!clusterData.user) {
        showResult(createResult, '用户是必填项', 'error');
        return;
    }
    
    if (!clusterData.service_name) {
        showResult(createResult, '服务名称是必填项', 'error');
        return;
    }
    
    // 移除空值字段
    Object.keys(clusterData).forEach(key => {
        if (clusterData[key] === '' || clusterData[key] === null || clusterData[key] === undefined) {
            delete clusterData[key];
        }
    });
    
    try {
        showResult(createResult, '正在创建集群...', 'info');
        
        const response = await fetch(`${API_BASE_URL}/clusters`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(clusterData)
        });
        
        const result = await response.text();
        
        if (response.ok) {
            showResult(createResult, `集群创建成功!\n${result}`, 'success');
            createClusterForm.reset();
        } else {
            showResult(createResult, `创建失败: ${response.status} ${response.statusText}\n${result}`, 'error');
        }
    } catch (error) {
        showResult(createResult, `请求失败: ${error.message}`, 'error');
    }
});

// 删除集群表单提交事件
deleteClusterForm.addEventListener('submit', async (e) => {
    e.preventDefault();
    
    const namespace = document.getElementById('delete_namespace').value;
    
    if (!namespace) {
        showResult(deleteResult, '请提供命名空间', 'error');
        return;
    }
    
    try {
        showResult(deleteResult, '正在删除集群...', 'info');
        
        const response = await fetch(`${API_BASE_URL}/clusters`, {
            method: 'DELETE',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({ namespace: namespace })
        });
        
        const result = await response.text();
        
        if (response.ok) {
            showResult(deleteResult, `集群删除成功!\n${result}`, 'success');
            deleteClusterForm.reset();
        } else {
            showResult(deleteResult, `删除失败: ${response.status} ${response.statusText}\n${result}`, 'error');
        }
    } catch (error) {
        showResult(deleteResult, `请求失败: ${error.message}`, 'error');
    }
});

// 查询所有集群表单提交事件
listClustersForm.addEventListener('submit', async (e) => {
    e.preventDefault();
    
    try {
        showResult(listResult, '正在查询集群...', 'info');
        
        const response = await fetch(`${API_BASE_URL}/clusters`, {
            method: 'GET'
        });
        
        const clusters = await response.json();
        
        if (response.ok) {
            showResult(listResult, `查询成功! 共找到 ${clusters.length} 个集群:`, 'success');
            const pre = document.createElement('pre');
            pre.textContent = JSON.stringify(clusters, null, 2);
            listResult.appendChild(pre);
        } else {
            showResult(listResult, `查询失败: ${response.status} ${response.statusText}`, 'error');
        }
    } catch (error) {
        showResult(listResult, `请求失败: ${error.message}`, 'error');
    }
});

// 查询集群详情表单提交事件
clusterDetailsForm.addEventListener('submit', async (e) => {
    e.preventDefault();
    
    const namespace = document.getElementById('cluster_namespace').value;
    
    if (!namespace) {
        showResult(detailsResult, '请提供命名空间', 'error');
        return;
    }
    
    try {
        showResult(detailsResult, '正在查询集群详情...', 'info');
        
        const response = await fetch(`${API_BASE_URL}/clusters/${namespace}`, {
            method: 'GET'
        });
        
        const clusterDetails = await response.json();
        
        if (response.ok) {
            showResult(detailsResult, `查询成功!`, 'success');
            const pre = document.createElement('pre');
            pre.textContent = JSON.stringify(clusterDetails, null, 2);
            detailsResult.appendChild(pre);
        } else {
            showResult(detailsResult, `查询失败: ${response.status} ${response.statusText}`, 'error');
        }
    } catch (error) {
        showResult(detailsResult, `请求失败: ${error.message}`, 'error');
    }
});

// 显示结果函数
function showResult(element, message, type) {
    element.textContent = message;
    element.className = `result ${type}`;
    element.style.display = 'block';
    
    // 如果是成功或错误信息，3秒后自动隐藏
    if (type === 'success' || type === 'error') {
        setTimeout(() => {
            element.style.display = 'none';
        }, 5000);
    }
}

// 页面加载完成后检查API连接
document.addEventListener('DOMContentLoaded', async () => {
    try {
        const response = await fetch(`${API_BASE_URL}/health`);
        if (response.ok) {
            console.log('API服务连接正常');
        } else {
            console.warn('API服务可能不可用');
        }
    } catch (error) {
        console.warn('无法连接到API服务:', error.message);
    }
});