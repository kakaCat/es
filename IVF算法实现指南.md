# IVF 算法实现指南

**目标:** 用最简单的方式实现可工作的 IVF 向量检索功能

---

## 第1步: 向量相似度计算 (2-3小时)

### 创建 VectorSimilarity.java

```java
// es-plugin/src/main/java/com/es/plugin/vector/ivf/VectorSimilarity.java

package com.es.plugin.vector.ivf;

public class VectorSimilarity {

    /**
     * L2 欧几里得距离
     * distance = sqrt(sum((v1[i] - v2[i])^2))
     */
    public static float l2Distance(float[] v1, float[] v2) {
        if (v1.length != v2.length) {
            throw new IllegalArgumentException("Vector dimensions must match");
        }

        float sum = 0.0f;
        for (int i = 0; i < v1.length; i++) {
            float diff = v1[i] - v2[i];
            sum += diff * diff;
        }
        return (float) Math.sqrt(sum);
    }

    /**
     * Cosine 相似度
     * similarity = dot(v1, v2) / (norm(v1) * norm(v2))
     * 返回值范围: [-1, 1], 1表示完全相同
     */
    public static float cosineSimilarity(float[] v1, float[] v2) {
        if (v1.length != v2.length) {
            throw new IllegalArgumentException("Vector dimensions must match");
        }

        float dotProduct = 0.0f;
        float norm1 = 0.0f;
        float norm2 = 0.0f;

        for (int i = 0; i < v1.length; i++) {
            dotProduct += v1[i] * v2[i];
            norm1 += v1[i] * v1[i];
            norm2 += v2[i] * v2[i];
        }

        if (norm1 == 0.0f || norm2 == 0.0f) {
            return 0.0f;
        }

        return dotProduct / (float)(Math.sqrt(norm1) * Math.sqrt(norm2));
    }

    /**
     * 点积 (内积)
     * dot = sum(v1[i] * v2[i])
     */
    public static float dotProduct(float[] v1, float[] v2) {
        if (v1.length != v2.length) {
            throw new IllegalArgumentException("Vector dimensions must match");
        }

        float sum = 0.0f;
        for (int i = 0; i < v1.length; i++) {
            sum += v1[i] * v2[i];
        }
        return sum;
    }

    /**
     * 批量计算距离 (优化性能)
     */
    public static float[] batchL2Distance(float[] query, float[][] vectors) {
        float[] distances = new float[vectors.length];
        for (int i = 0; i < vectors.length; i++) {
            distances[i] = l2Distance(query, vectors[i]);
        }
        return distances;
    }
}
```

**测试代码:**

```java
// 简单测试
float[] v1 = {1.0f, 2.0f, 3.0f};
float[] v2 = {4.0f, 5.0f, 6.0f};

float l2 = VectorSimilarity.l2Distance(v1, v2);
float cosine = VectorSimilarity.cosineSimilarity(v1, v2);
float dot = VectorSimilarity.dotProduct(v1, v2);

System.out.println("L2: " + l2);        // 5.196
System.out.println("Cosine: " + cosine); // 0.974
System.out.println("Dot: " + dot);       // 32.0
```

---

## 第2步: 简化版 KMeans 训练 (1天)

### 创建 SimpleKMeansTrainer.java

```java
// es-plugin/src/main/java/com/es/plugin/vector/ivf/SimpleKMeansTrainer.java

package com.es.plugin.vector.ivf;

import java.util.Random;
import java.util.Arrays;

public class SimpleKMeansTrainer {

    private final int nClusters;      // nlist: 聚类数量
    private final int maxIterations;   // 最大迭代次数
    private float[][] centroids;       // 聚类中心

    public SimpleKMeansTrainer(int nClusters, int maxIterations) {
        this.nClusters = nClusters;
        this.maxIterations = maxIterations;
    }

    /**
     * 训练 KMeans 模型
     * @param vectors 训练向量
     * @return 聚类中心
     */
    public float[][] train(float[][] vectors) {
        int dimension = vectors[0].length;

        // 1. 随机初始化聚类中心
        centroids = initializeCentroids(vectors, nClusters, dimension);

        // 2. 迭代优化
        int[] assignments = new int[vectors.length];

        for (int iter = 0; iter < maxIterations; iter++) {
            // 2.1 分配每个向量到最近的聚类
            boolean changed = false;
            for (int i = 0; i < vectors.length; i++) {
                int newCluster = findNearestCentroid(vectors[i], centroids);
                if (assignments[i] != newCluster) {
                    changed = true;
                    assignments[i] = newCluster;
                }
            }

            // 如果没有变化,提前结束
            if (!changed) {
                System.out.println("KMeans converged at iteration " + iter);
                break;
            }

            // 2.2 重新计算聚类中心
            updateCentroids(vectors, assignments, centroids);
        }

        return centroids;
    }

    /**
     * 随机初始化聚类中心 (从数据中随机选择)
     */
    private float[][] initializeCentroids(float[][] vectors, int k, int dim) {
        Random rand = new Random(42); // 固定种子便于调试
        float[][] centers = new float[k][dim];

        // 随机选择 k 个向量作为初始中心
        boolean[] selected = new boolean[vectors.length];
        for (int i = 0; i < k; i++) {
            int idx;
            do {
                idx = rand.nextInt(vectors.length);
            } while (selected[idx]);

            selected[idx] = true;
            System.arraycopy(vectors[idx], 0, centers[i], 0, dim);
        }

        return centers;
    }

    /**
     * 找到最近的聚类中心
     */
    private int findNearestCentroid(float[] vector, float[][] centroids) {
        int nearest = 0;
        float minDist = Float.MAX_VALUE;

        for (int i = 0; i < centroids.length; i++) {
            float dist = VectorSimilarity.l2Distance(vector, centroids[i]);
            if (dist < minDist) {
                minDist = dist;
                nearest = i;
            }
        }

        return nearest;
    }

    /**
     * 更新聚类中心 (计算每个簇的平均值)
     */
    private void updateCentroids(float[][] vectors, int[] assignments,
                                  float[][] centroids) {
        int dimension = vectors[0].length;
        int[] counts = new int[centroids.length];

        // 重置聚类中心
        for (int i = 0; i < centroids.length; i++) {
            Arrays.fill(centroids[i], 0.0f);
        }

        // 累加每个簇的向量
        for (int i = 0; i < vectors.length; i++) {
            int cluster = assignments[i];
            counts[cluster]++;

            for (int d = 0; d < dimension; d++) {
                centroids[cluster][d] += vectors[i][d];
            }
        }

        // 计算平均值
        for (int i = 0; i < centroids.length; i++) {
            if (counts[i] > 0) {
                for (int d = 0; d < dimension; d++) {
                    centroids[i][d] /= counts[i];
                }
            }
        }
    }

    public float[][] getCentroids() {
        return centroids;
    }
}
```

**使用示例:**

```java
// 准备训练数据
float[][] trainingVectors = loadVectorsFromIndex(); // 从索引加载向量

// 训练 KMeans (nlist=100, 最大迭代100次)
SimpleKMeansTrainer trainer = new SimpleKMeansTrainer(100, 100);
float[][] centroids = trainer.train(trainingVectors);

// 保存聚类中心到索引元数据
saveCentroidsToMetadata(centroids);
```

---

## 第3步: 倒排索引构建 (1天)

### 创建 InvertedFileIndex.java

```java
// es-plugin/src/main/java/com/es/plugin/vector/ivf/InvertedFileIndex.java

package com.es.plugin.vector.ivf;

import java.util.*;
import java.io.*;

/**
 * 向量文档
 */
class VectorDoc implements Serializable {
    private static final long serialVersionUID = 1L;

    public int docId;          // Elasticsearch 文档 ID
    public float[] vector;     // 向量数据
    public Map<String, Object> metadata; // 元数据

    public VectorDoc(int docId, float[] vector, Map<String, Object> metadata) {
        this.docId = docId;
        this.vector = vector;
        this.metadata = metadata;
    }
}

/**
 * 倒排索引 (IVF)
 */
public class InvertedFileIndex {

    private float[][] centroids;  // 聚类中心
    private Map<Integer, List<VectorDoc>> invertedLists; // 倒排表
    private String metric;         // 距离度量: l2, cosine, dot

    public InvertedFileIndex(float[][] centroids, String metric) {
        this.centroids = centroids;
        this.metric = metric;
        this.invertedLists = new HashMap<>();

        // 初始化倒排表
        for (int i = 0; i < centroids.length; i++) {
            invertedLists.put(i, new ArrayList<>());
        }
    }

    /**
     * 添加向量到索引
     */
    public void addVector(int docId, float[] vector, Map<String, Object> metadata) {
        // 1. 找到最近的聚类簇
        int clusterId = findNearestCluster(vector);

        // 2. 添加到对应的倒排列表
        VectorDoc doc = new VectorDoc(docId, vector, metadata);
        invertedLists.get(clusterId).add(doc);
    }

    /**
     * 批量添加向量
     */
    public void addVectors(List<VectorDoc> docs) {
        for (VectorDoc doc : docs) {
            addVector(doc.docId, doc.vector, doc.metadata);
        }
    }

    /**
     * ANN 搜索
     * @param queryVector 查询向量
     * @param k 返回Top-K结果
     * @param nprobe 搜索的聚类簇数量
     * @return 排序后的搜索结果
     */
    public List<SearchResult> search(float[] queryVector, int k, int nprobe) {
        // 1. 找到最近的 nprobe 个聚类簇
        int[] nearestClusters = findNearestClusters(queryVector, nprobe);

        // 2. 从这些簇中收集候选向量
        List<VectorDoc> candidates = new ArrayList<>();
        for (int clusterId : nearestClusters) {
            candidates.addAll(invertedLists.get(clusterId));
        }

        // 3. 计算距离并排序
        List<SearchResult> results = new ArrayList<>();
        for (VectorDoc doc : candidates) {
            float distance = calculateDistance(queryVector, doc.vector);
            results.add(new SearchResult(doc.docId, distance, doc.metadata));
        }

        // 4. 按距离排序
        results.sort(Comparator.comparingDouble(r -> r.distance));

        // 5. 返回 Top-K
        return results.subList(0, Math.min(k, results.size()));
    }

    /**
     * 找到最近的聚类簇
     */
    private int findNearestCluster(float[] vector) {
        int nearest = 0;
        float minDist = Float.MAX_VALUE;

        for (int i = 0; i < centroids.length; i++) {
            float dist = VectorSimilarity.l2Distance(vector, centroids[i]);
            if (dist < minDist) {
                minDist = dist;
                nearest = i;
            }
        }

        return nearest;
    }

    /**
     * 找到最近的 nprobe 个聚类簇
     */
    private int[] findNearestClusters(float[] vector, int nprobe) {
        // 计算到所有聚类中心的距离
        float[] distances = new float[centroids.length];
        for (int i = 0; i < centroids.length; i++) {
            distances[i] = VectorSimilarity.l2Distance(vector, centroids[i]);
        }

        // 找到最小的 nprobe 个
        int[] indices = new int[centroids.length];
        for (int i = 0; i < centroids.length; i++) {
            indices[i] = i;
        }

        // 简单排序 (可优化为部分排序)
        for (int i = 0; i < Math.min(nprobe, indices.length); i++) {
            for (int j = i + 1; j < indices.length; j++) {
                if (distances[indices[j]] < distances[indices[i]]) {
                    int temp = indices[i];
                    indices[i] = indices[j];
                    indices[j] = temp;
                }
            }
        }

        return Arrays.copyOf(indices, Math.min(nprobe, indices.length));
    }

    /**
     * 根据配置的 metric 计算距离
     */
    private float calculateDistance(float[] v1, float[] v2) {
        switch (metric.toLowerCase()) {
            case "cosine":
                return 1.0f - VectorSimilarity.cosineSimilarity(v1, v2);
            case "dot":
                return -VectorSimilarity.dotProduct(v1, v2); // 负值使得越大越相似
            case "l2":
            default:
                return VectorSimilarity.l2Distance(v1, v2);
        }
    }

    /**
     * 持久化索引到文件 (简单的序列化)
     */
    public void saveToFile(String filepath) throws IOException {
        try (ObjectOutputStream oos = new ObjectOutputStream(
                new FileOutputStream(filepath))) {
            oos.writeObject(centroids);
            oos.writeObject(invertedLists);
            oos.writeObject(metric);
        }
    }

    /**
     * 从文件加载索引
     */
    @SuppressWarnings("unchecked")
    public static InvertedFileIndex loadFromFile(String filepath)
            throws IOException, ClassNotFoundException {
        try (ObjectInputStream ois = new ObjectInputStream(
                new FileInputStream(filepath))) {
            float[][] centroids = (float[][]) ois.readObject();
            Map<Integer, List<VectorDoc>> invertedLists =
                (Map<Integer, List<VectorDoc>>) ois.readObject();
            String metric = (String) ois.readObject();

            InvertedFileIndex index = new InvertedFileIndex(centroids, metric);
            index.invertedLists = invertedLists;
            return index;
        }
    }

    /**
     * 获取统计信息
     */
    public IndexStats getStats() {
        int totalVectors = 0;
        for (List<VectorDoc> list : invertedLists.values()) {
            totalVectors += list.size();
        }
        return new IndexStats(centroids.length, totalVectors);
    }
}

/**
 * 搜索结果
 */
class SearchResult {
    public int docId;
    public float distance;
    public Map<String, Object> metadata;

    public SearchResult(int docId, float distance, Map<String, Object> metadata) {
        this.docId = docId;
        this.distance = distance;
        this.metadata = metadata;
    }
}

/**
 * 索引统计信息
 */
class IndexStats {
    public int nClusters;
    public int totalVectors;

    public IndexStats(int nClusters, int totalVectors) {
        this.nClusters = nClusters;
        this.totalVectors = totalVectors;
    }
}
```

---

## 第4步: 集成到 IVFQueryBuilder (1.5天)

### 修改 IVFQueryBuilder.java

```java
// es-plugin/src/main/java/com/es/plugin/vector/ivf/IVFQueryBuilder.java

@Override
protected Query doToQuery(QueryShardContext context) throws IOException {
    // 1. 获取或构建 IVF 索引
    InvertedFileIndex ivfIndex = getOrBuildIndex(context, field);

    // 2. 执行 ANN 搜索
    List<SearchResult> results = ivfIndex.search(vector, k, nprobe);

    // 3. 转换为 Lucene Query
    if (results.isEmpty()) {
        return new MatchNoDocsQuery();
    }

    // 3.1 创建 DocID 查询
    BooleanQuery.Builder builder = new BooleanQuery.Builder();
    for (SearchResult result : results) {
        // 将 docId 转换为 Lucene Term Query
        builder.add(new TermQuery(new Term("_id", String.valueOf(result.docId))),
                    BooleanClause.Occur.SHOULD);
    }

    return builder.build();
}

/**
 * 获取或构建 IVF 索引
 */
private InvertedFileIndex getOrBuildIndex(QueryShardContext context, String fieldName)
        throws IOException {
    // 1. 尝试从缓存加载
    String indexCacheKey = context.index().getName() + "_" + fieldName;
    InvertedFileIndex cachedIndex = indexCache.get(indexCacheKey);

    if (cachedIndex != null) {
        return cachedIndex;
    }

    // 2. 从索引构建
    // 2.1 读取所有向量
    List<VectorDoc> allVectors = readAllVectorsFromIndex(context, fieldName);

    // 2.2 训练 KMeans
    float[][] vectors = allVectors.stream()
        .map(doc -> doc.vector)
        .toArray(float[][]::new);

    SimpleKMeansTrainer trainer = new SimpleKMeansTrainer(nlist, 100);
    float[][] centroids = trainer.train(vectors);

    // 2.3 构建倒排索引
    InvertedFileIndex ivfIndex = new InvertedFileIndex(centroids, metric);
    ivfIndex.addVectors(allVectors);

    // 2.4 缓存索引
    indexCache.put(indexCacheKey, ivfIndex);

    return ivfIndex;
}

/**
 * 从 Elasticsearch 索引读取所有向量
 */
private List<VectorDoc> readAllVectorsFromIndex(QueryShardContext context,
                                                  String fieldName)
        throws IOException {
    List<VectorDoc> vectors = new ArrayList<>();

    // 使用 Lucene IndexReader 遍历所有文档
    IndexReader reader = context.getIndexReader();
    for (LeafReaderContext leaf : reader.leaves()) {
        LeafReader leafReader = leaf.reader();

        for (int docId = 0; docId < leafReader.maxDoc(); docId++) {
            Document doc = leafReader.document(docId);

            // 读取向量字段
            BytesRef vectorBytes = doc.getBinaryValue(fieldName);
            if (vectorBytes != null) {
                float[] vector = deserializeVector(vectorBytes);
                vectors.add(new VectorDoc(docId, vector, null));
            }
        }
    }

    return vectors;
}

/**
 * 反序列化向量 (假设存储为 float[] 的字节数组)
 */
private float[] deserializeVector(BytesRef bytes) {
    ByteBuffer buffer = ByteBuffer.wrap(bytes.bytes, bytes.offset, bytes.length);
    float[] vector = new float[bytes.length / 4]; // 4 bytes per float

    for (int i = 0; i < vector.length; i++) {
        vector[i] = buffer.getFloat();
    }

    return vector;
}
```

---

## 第5步: 测试验证

### 5.1 创建测试脚本

```bash
# demo/test-ivf-vector-search.sh

#!/bin/bash

ES_URL="http://localhost:9200"

echo "=== Step 1: 创建向量索引 ==="
curl -X PUT "$ES_URL/test-vectors" -H 'Content-Type: application/json' -d'{
  "mappings": {
    "properties": {
      "embedding": {
        "type": "vector",
        "dimension": 128,
        "metric": "l2",
        "nlist": 10,
        "nprobe": 3
      },
      "title": {
        "type": "text"
      }
    }
  }
}'

echo -e "\n\n=== Step 2: 插入测试向量 ==="
for i in {1..100}; do
  # 生成随机向量
  vector=$(python3 -c "import random; print([random.random() for _ in range(128)])")

  curl -X POST "$ES_URL/test-vectors/_doc/$i" -H 'Content-Type: application/json' -d"{
    \"embedding\": $vector,
    \"title\": \"Document $i\"
  }"
done

echo -e "\n\n=== Step 3: 等待索引刷新 ==="
sleep 2

echo -e "\n\n=== Step 4: 执行向量搜索 ==="
query_vector=$(python3 -c "import random; print([random.random() for _ in range(128)])")

curl -X POST "$ES_URL/test-vectors/_search" -H 'Content-Type: application/json' -d"{
  \"size\": 10,
  \"query\": {
    \"ann\": {
      \"field\": \"embedding\",
      \"vector\": $query_vector,
      \"algorithm\": \"ivf\",
      \"nprobe\": 3
    }
  }
}"
```

### 5.2 验证结果

运行测试:
```bash
chmod +x demo/test-ivf-vector-search.sh
./demo/test-ivf-vector-search.sh
```

期望输出:
```json
{
  "took": 150,
  "hits": {
    "total": { "value": 10 },
    "hits": [
      {
        "_id": "42",
        "_score": 0.85,
        "_source": {
          "title": "Document 42"
        }
      },
      ...
    ]
  }
}
```

---

## 优化建议

### 性能优化 (后期可做)

1. **索引持久化**
   - 当前每次查询都重建索引
   - 优化: 构建一次,持久化到磁盘
   - 在索引更新时重新训练

2. **并行计算**
   - KMeans 训练并行化
   - 距离计算使用多线程
   - SIMD 向量化操作

3. **内存优化**
   - 使用 off-heap 内存存储向量
   - 压缩向量 (Product Quantization)
   - LRU 缓存热点向量

### 功能增强 (后期可做)

1. **增量更新**
   - 新增向量时不重新训练
   - 设置重训练阈值 (如新增10%数据)

2. **查询优化**
   - Early termination (找到足够多结果即停止)
   - 自适应 nprobe (根据召回率动态调整)

---

## 时间估算

| 任务 | 预估时间 | 实际建议 |
|------|---------|---------|
| VectorSimilarity | 2小时 | 先写单元测试 |
| SimpleKMeansTrainer | 6小时 | 用小数据集验证 |
| InvertedFileIndex | 8小时 | 分步实现和测试 |
| IVFQueryBuilder集成 | 10小时 | 最复杂,多测试 |
| 测试和调试 | 8小时 | 充分测试 |

**总计: 约 34小时 (4-5个工作日)**

---

## 常见问题

### Q1: 如何调试?
A: 在每个方法中添加日志:
```java
System.out.println("KMeans iteration " + iter + ", changed: " + changed);
```

### Q2: 性能不达标怎么办?
A: 先不要优化,确保功能可用。后续可以:
- 减少 nlist (降低训练时间)
- 增加 nprobe (提高召回率)
- 使用缓存避免重复计算

### Q3: 内存溢出怎么办?
A: 分批处理数据:
```java
// 不要一次加载所有向量
// 改为分批训练
for (int batch = 0; batch < totalBatches; batch++) {
    float[][] batchVectors = loadBatch(batch);
    processVectors(batchVectors);
}
```

---

## 下一步

完成 IVF 算法后:
1. ✅ 基本功能测试 (100个向量)
2. ✅ 中等规模测试 (10,000个向量)
3. ✅ 修复 Bug
4. ✅ 启用 Elasticsearch Security
5. ✅ 创建简单 UI

**目标: 2周内完成可演示的版本!**
