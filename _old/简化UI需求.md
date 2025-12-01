# 简化版 UI 功能需求

基于《说明.md》的要求,但采用简单实用的实现方式。

---

## UI-2: 索引管理页面 (简化版)

**目标:** 提供基本的索引创建和查看功能

### 必需功能
```html
<!-- frontend/index-management.html -->

<h2>创建向量索引</h2>
<form>
  <input name="index_name" placeholder="索引名称" required>
  <input name="dimension" type="number" placeholder="向量维度 (如128)" required>
  <select name="metric">
    <option value="l2">L2 距离</option>
    <option value="cosine">Cosine 相似度</option>
    <option value="dot">点积</option>
  </select>
  <input name="nlist" type="number" placeholder="nlist (默认100)" value="100">
  <input name="nprobe" type="number" placeholder="nprobe (默认10)" value="10">
  <button type="submit">创建索引</button>
</form>

<h2>索引列表</h2>
<button onclick="loadIndexes()">刷新列表</button>
<table id="indexes-table">
  <thead>
    <tr>
      <th>索引名</th>
      <th>维度</th>
      <th>距离度量</th>
      <th>nlist/nprobe</th>
      <th>状态</th>
      <th>操作</th>
    </tr>
  </thead>
  <tbody></tbody>
</table>
```

**实现工作量:** 1-2小时

---

## UI-3: 监控页面 (简化版)

**目标:** 展示基本的实时指标和历史趋势

### 必需功能

#### 3.1 简单的实时指标展示
```html
<!-- frontend/monitoring.html -->

<div class="metrics-grid">
  <!-- 当前QPS -->
  <div class="metric-card">
    <h3>当前 QPS</h3>
    <div id="qps-value" class="metric-value">0</div>
  </div>

  <!-- 平均延迟 -->
  <div class="metric-card">
    <h3>平均延迟 (ms)</h3>
    <div id="latency-value" class="metric-value">0</div>
  </div>

  <!-- 节点数量 -->
  <div class="metric-card">
    <h3>在线节点</h3>
    <div id="nodes-value" class="metric-value">0</div>
  </div>

  <!-- CPU使用率 -->
  <div class="metric-card">
    <h3>CPU 使用率</h3>
    <div id="cpu-value" class="metric-value">0%</div>
  </div>
</div>

<!-- 使用简单的文本表格显示节点状态 -->
<h3>节点状态</h3>
<table id="nodes-table">
  <thead>
    <tr>
      <th>节点名</th>
      <th>状态</th>
      <th>CPU</th>
      <th>内存</th>
    </tr>
  </thead>
  <tbody></tbody>
</table>

<!-- 使用文本列表显示历史记录 -->
<h3>QPS 历史 (最近10条)</h3>
<ul id="qps-history"></ul>
```

**JavaScript 实现:**
```javascript
// 每10秒刷新一次数据
setInterval(async () => {
  const response = await fetch('/metrics');
  const data = await response.json();

  document.getElementById('qps-value').textContent = data.qps;
  document.getElementById('latency-value').textContent = data.latency;
  document.getElementById('nodes-value').textContent = data.nodes;
  document.getElementById('cpu-value').textContent = data.cpu + '%';

  updateNodesTable(data.nodesList);
  updateQPSHistory(data.qpsHistory);
}, 10000);
```

**实现工作量:** 2-3小时

#### 3.2 向量分布可视化 (超简化版)

**《说明.md》要求 TSNE 降维,但可以先用简单方式:**

```html
<!-- 选项1: 纯文本统计 (最简单) -->
<h3>向量统计</h3>
<div>
  <p>总向量数: <span id="total-vectors">0</span></p>
  <p>平均向量长度: <span id="avg-norm">0</span></p>
  <p>最大距离: <span id="max-distance">0</span></p>
</div>

<!-- 选项2: 简单的HTML Canvas散点图 (中等) -->
<h3>向量分布 (随机采样1000个)</h3>
<canvas id="vector-canvas" width="600" height="400"></canvas>
<script>
// 使用前两个维度简单绘制散点图
function drawSimpleScatter(vectors) {
  const canvas = document.getElementById('vector-canvas');
  const ctx = canvas.getContext('2d');

  vectors.forEach(v => {
    const x = (v[0] + 1) * 300;  // 归一化到canvas坐标
    const y = (v[1] + 1) * 200;
    ctx.fillRect(x, y, 2, 2);
  });
}
</script>
```

**实现工作量:**
- 选项1 (纯文本): 30分钟
- 选项2 (简单散点图): 2小时

**如果真的需要 TSNE (后期可选):**
```python
# 创建一个超简单的Python服务
# services/vector-viz/app.py (仅50行代码)

from flask import Flask, request, jsonify
from sklearn.manifold import TSNE
import numpy as np

app = Flask(__name__)

@app.route('/tsne', methods=['POST'])
def tsne_endpoint():
    vectors = request.json['vectors']
    coords = TSNE(n_components=2).fit_transform(vectors)
    return jsonify(coords.tolist())

if __name__ == '__main__':
    app.run(port=5000)
```

**实现工作量:** 半天 (包括Docker部署)

---

## 简化方案总结

### 必须实现 (满足基本需求)
- [x] 索引管理页面: 创建表单 + 列表展示 - **2小时**
- [x] 监控页面: 实时指标卡片 + 节点表格 - **3小时**
- [ ] 向量统计: 纯文本统计信息 - **30分钟**

**总工作量: 5.5小时 (1个工作日内完成)**

### 可选增强 (时间充裕时)
- [ ] 简单散点图 (不用TSNE,只用前2维) - **+2小时**
- [ ] TSNE可视化 (Python服务) - **+半天**
- [ ] Chart.js 图表 (QPS曲线图) - **+2小时**

---

## 实现优先级

### Phase 1: 最小可用版本 (1天)
```bash
1. 索引管理页面 - 基本表单和列表
2. 监控页面 - 实时数字指标
3. 向量统计 - 纯文本展示
```

### Phase 2: 增强版 (可选,+2天)
```bash
4. 添加简单的Canvas散点图
5. 添加Chart.js折线图 (QPS趋势)
6. 美化CSS样式
```

### Phase 3: 完整版 (可选,+1周)
```bash
7. TSNE可视化服务
8. 交互式图表 (D3.js)
9. 实时更新 (WebSocket)
```

---

## 前端文件结构

```
frontend/
├── index.html              # 集群管理 (已有)
├── index-management.html   # 索引管理 (新建 - 简单)
├── monitoring.html         # 监控页面 (新建 - 简单)
├── styles.css             # 统一样式 (已有)
├── script.js              # 集群管理脚本 (已有)
├── index-management.js    # 索引管理脚本 (新建 - ~100行)
└── monitoring.js          # 监控脚本 (新建 - ~150行)
```

**总代码量:** ~250行 JavaScript + ~200行 HTML = **450行代码**

---

## 代码示例: 极简监控页面

```html
<!DOCTYPE html>
<html>
<head>
    <title>ES Serverless 监控</title>
    <style>
        .metrics-grid { display: grid; grid-template-columns: repeat(4, 1fr); gap: 20px; }
        .metric-card { border: 1px solid #ddd; padding: 20px; text-align: center; }
        .metric-value { font-size: 36px; font-weight: bold; color: #007bff; }
        table { width: 100%; border-collapse: collapse; margin-top: 20px; }
        th, td { border: 1px solid #ddd; padding: 8px; text-align: left; }
    </style>
</head>
<body>
    <h1>ES Serverless 监控</h1>

    <div class="metrics-grid">
        <div class="metric-card">
            <h3>当前 QPS</h3>
            <div id="qps" class="metric-value">-</div>
        </div>
        <div class="metric-card">
            <h3>延迟 P95 (ms)</h3>
            <div id="latency" class="metric-value">-</div>
        </div>
        <div class="metric-card">
            <h3>在线节点</h3>
            <div id="nodes" class="metric-value">-</div>
        </div>
        <div class="metric-card">
            <h3>CPU 使用率</h3>
            <div id="cpu" class="metric-value">-</div>
        </div>
    </div>

    <h2>节点状态</h2>
    <table id="nodes-table">
        <thead>
            <tr><th>节点</th><th>状态</th><th>CPU</th><th>内存</th></tr>
        </thead>
        <tbody></tbody>
    </table>

    <script>
        async function loadMetrics() {
            const res = await fetch('http://localhost:8080/metrics');
            const data = await res.json();

            document.getElementById('qps').textContent = data.qps || 0;
            document.getElementById('latency').textContent = data.latency || 0;
            document.getElementById('nodes').textContent = data.nodes_count || 0;
            document.getElementById('cpu').textContent = (data.cpu_usage || 0) + '%';

            const tbody = document.querySelector('#nodes-table tbody');
            tbody.innerHTML = data.nodes_list?.map(node => `
                <tr>
                    <td>${node.name}</td>
                    <td>${node.status}</td>
                    <td>${node.cpu}%</td>
                    <td>${node.memory}%</td>
                </tr>
            `).join('') || '<tr><td colspan="4">暂无数据</td></tr>';
        }

        // 每10秒刷新
        setInterval(loadMetrics, 10000);
        loadMetrics();
    </script>
</body>
</html>
```

**这个文件只有70行,但满足基本监控需求!**

---

## 建议

1. **先做最简单的版本** - 纯HTML表单+表格,无需复杂图表
2. **数据优先,展示其次** - 确保API返回正确数据,UI只是简单展示
3. **TSNE可以后补** - 先用文本统计,后期有时间再加可视化
4. **避免过度工程** - 不用React/Vue,纯HTML+Vanilla JS即可

**核心原则: 功能完整 > 界面华丽**
