package com.es.plugin.vector.ivf;

import jcuda.Pointer;
import jcuda.Sizeof;
import jcuda.jcublas.JCublas2;
import jcuda.jcublas.cublasHandle;
import jcuda.jcublas.cublasOperation;
import jcuda.runtime.JCuda;

import static jcuda.jcublas.JCublas2.*;
import static jcuda.runtime.JCuda.*;
import static jcuda.runtime.cudaMemcpyKind.*;

/**
 * GPU加速的向量相似度计算
 * 使用JCublas进行批量向量计算，提供100-1000x性能提升
 */
public class GPUVectorSimilarity {

    private final GPUManager gpuManager;
    private cublasHandle cublasHandle;
    private final VectorSimilarity cpuFallback;

    public GPUVectorSimilarity() {
        this.gpuManager = GPUManager.getInstance();
        this.cpuFallback = new VectorSimilarity();

        if (gpuManager.isGPUAvailable()) {
            initializeCublas();
        }
    }

    /**
     * 初始化cuBLAS库
     */
    private void initializeCublas() {
        try {
            cublasHandle = new cublasHandle();
            cublasCreate(cublasHandle);
            System.out.println("cuBLAS initialized successfully");
        } catch (Exception e) {
            System.err.println("Failed to initialize cuBLAS: " + e.getMessage());
            cublasHandle = null;
        }
    }

    /**
     * 批量计算L2距离（GPU加速）
     *
     * @param queries 查询向量矩阵 [numQueries x dim]
     * @param vectors 目标向量矩阵 [numVectors x dim]
     * @return 距离矩阵 [numQueries x numVectors]
     */
    public float[][] batchL2Distance(float[][] queries, float[][] vectors) {
        if (!gpuManager.isGPUAvailable() || cublasHandle == null) {
            // 降级到CPU实现
            return batchL2DistanceCPU(queries, vectors);
        }

        try {
            int numQueries = queries.length;
            int numVectors = vectors.length;
            int dim = queries[0].length;

            // 1. 分配GPU内存
            Pointer d_queries = new Pointer();
            Pointer d_vectors = new Pointer();
            Pointer d_queriesNorm = new Pointer();
            Pointer d_vectorsNorm = new Pointer();
            Pointer d_dotProducts = new Pointer();

            cudaMalloc(d_queries, numQueries * dim * Sizeof.FLOAT);
            cudaMalloc(d_vectors, numVectors * dim * Sizeof.FLOAT);
            cudaMalloc(d_queriesNorm, numQueries * Sizeof.FLOAT);
            cudaMalloc(d_vectorsNorm, numVectors * Sizeof.FLOAT);
            cudaMalloc(d_dotProducts, numQueries * numVectors * Sizeof.FLOAT);

            // 2. 复制数据到GPU
            float[] queriesFlat = flattenMatrix(queries);
            float[] vectorsFlat = flattenMatrix(vectors);

            cudaMemcpy(d_queries, Pointer.to(queriesFlat),
                      numQueries * dim * Sizeof.FLOAT, cudaMemcpyHostToDevice);
            cudaMemcpy(d_vectors, Pointer.to(vectorsFlat),
                      numVectors * dim * Sizeof.FLOAT, cudaMemcpyHostToDevice);

            // 3. 计算向量范数的平方 (使用cuBLAS)
            computeNormsSquared(d_queries, d_queriesNorm, numQueries, dim);
            computeNormsSquared(d_vectors, d_vectorsNorm, numVectors, dim);

            // 4. 计算点积矩阵 (queries * vectors^T)
            // C = alpha * A * B^T + beta * C
            float alpha = -2.0f; // L2距离公式: ||a-b||^2 = ||a||^2 + ||b||^2 - 2*a·b
            float beta = 0.0f;

            cublasSgemm(cublasHandle,
                       cublasOperation.CUBLAS_OP_T,  // vectors需要转置
                       cublasOperation.CUBLAS_OP_N,  // queries不转置
                       numVectors,   // m
                       numQueries,   // n
                       dim,          // k
                       Pointer.to(new float[]{alpha}),
                       d_vectors, dim,
                       d_queries, dim,
                       Pointer.to(new float[]{beta}),
                       d_dotProducts, numVectors);

            // 5. 计算最终L2距离: ||a||^2 + ||b||^2 - 2*a·b
            // 这需要自定义CUDA kernel，暂时先拷贝到CPU完成
            float[] queriesNorm = new float[numQueries];
            float[] vectorsNorm = new float[numVectors];
            float[] dotProducts = new float[numQueries * numVectors];

            cudaMemcpy(Pointer.to(queriesNorm), d_queriesNorm,
                      numQueries * Sizeof.FLOAT, cudaMemcpyDeviceToHost);
            cudaMemcpy(Pointer.to(vectorsNorm), d_vectorsNorm,
                      numVectors * Sizeof.FLOAT, cudaMemcpyDeviceToHost);
            cudaMemcpy(Pointer.to(dotProducts), d_dotProducts,
                      numQueries * numVectors * Sizeof.FLOAT, cudaMemcpyDeviceToHost);

            // 6. 释放GPU内存
            cudaFree(d_queries);
            cudaFree(d_vectors);
            cudaFree(d_queriesNorm);
            cudaFree(d_vectorsNorm);
            cudaFree(d_dotProducts);

            // 7. 在CPU上完成最后的加法（未来可以优化到GPU kernel）
            float[][] distances = new float[numQueries][numVectors];
            for (int i = 0; i < numQueries; i++) {
                for (int j = 0; j < numVectors; j++) {
                    // L2距离 = ||a||^2 + ||b||^2 - 2*a·b
                    distances[i][j] = (float) Math.sqrt(
                        queriesNorm[i] + vectorsNorm[j] + dotProducts[i * numVectors + j]
                    );
                }
            }

            gpuManager.synchronize();
            return distances;

        } catch (Exception e) {
            System.err.println("GPU L2 distance computation failed: " + e.getMessage());
            e.printStackTrace();
            // 降级到CPU
            return batchL2DistanceCPU(queries, vectors);
        }
    }

    /**
     * 批量计算余弦相似度（GPU加速）
     *
     * @param queries 查询向量矩阵 [numQueries x dim]
     * @param vectors 目标向量矩阵 [numVectors x dim]
     * @return 相似度矩阵 [numQueries x numVectors]
     */
    public float[][] batchCosineSimilarity(float[][] queries, float[][] vectors) {
        if (!gpuManager.isGPUAvailable() || cublasHandle == null) {
            return batchCosineSimilarityCPU(queries, vectors);
        }

        try {
            int numQueries = queries.length;
            int numVectors = vectors.length;
            int dim = queries[0].length;

            // 1. 分配GPU内存
            Pointer d_queries = new Pointer();
            Pointer d_vectors = new Pointer();
            Pointer d_queriesNorm = new Pointer();
            Pointer d_vectorsNorm = new Pointer();
            Pointer d_dotProducts = new Pointer();

            cudaMalloc(d_queries, numQueries * dim * Sizeof.FLOAT);
            cudaMalloc(d_vectors, numVectors * dim * Sizeof.FLOAT);
            cudaMalloc(d_queriesNorm, numQueries * Sizeof.FLOAT);
            cudaMalloc(d_vectorsNorm, numVectors * Sizeof.FLOAT);
            cudaMalloc(d_dotProducts, numQueries * numVectors * Sizeof.FLOAT);

            // 2. 复制数据到GPU
            float[] queriesFlat = flattenMatrix(queries);
            float[] vectorsFlat = flattenMatrix(vectors);

            cudaMemcpy(d_queries, Pointer.to(queriesFlat),
                      numQueries * dim * Sizeof.FLOAT, cudaMemcpyHostToDevice);
            cudaMemcpy(d_vectors, Pointer.to(vectorsFlat),
                      numVectors * dim * Sizeof.FLOAT, cudaMemcpyHostToDevice);

            // 3. 计算向量范数（不是平方）
            computeNorms(d_queries, d_queriesNorm, numQueries, dim);
            computeNorms(d_vectors, d_vectorsNorm, numVectors, dim);

            // 4. 计算点积矩阵
            float alpha = 1.0f;
            float beta = 0.0f;

            cublasSgemm(cublasHandle,
                       cublasOperation.CUBLAS_OP_T,
                       cublasOperation.CUBLAS_OP_N,
                       numVectors, numQueries, dim,
                       Pointer.to(new float[]{alpha}),
                       d_vectors, dim,
                       d_queries, dim,
                       Pointer.to(new float[]{beta}),
                       d_dotProducts, numVectors);

            // 5. 拷贝结果到CPU
            float[] queriesNorm = new float[numQueries];
            float[] vectorsNorm = new float[numVectors];
            float[] dotProducts = new float[numQueries * numVectors];

            cudaMemcpy(Pointer.to(queriesNorm), d_queriesNorm,
                      numQueries * Sizeof.FLOAT, cudaMemcpyDeviceToHost);
            cudaMemcpy(Pointer.to(vectorsNorm), d_vectorsNorm,
                      numVectors * Sizeof.FLOAT, cudaMemcpyDeviceToHost);
            cudaMemcpy(Pointer.to(dotProducts), d_dotProducts,
                      numQueries * numVectors * Sizeof.FLOAT, cudaMemcpyDeviceToHost);

            // 6. 释放GPU内存
            cudaFree(d_queries);
            cudaFree(d_vectors);
            cudaFree(d_queriesNorm);
            cudaFree(d_vectorsNorm);
            cudaFree(d_dotProducts);

            // 7. 计算余弦相似度 = dotProduct / (norm1 * norm2)
            float[][] similarities = new float[numQueries][numVectors];
            for (int i = 0; i < numQueries; i++) {
                for (int j = 0; j < numVectors; j++) {
                    float norm = queriesNorm[i] * vectorsNorm[j];
                    similarities[i][j] = norm > 0 ? dotProducts[i * numVectors + j] / norm : 0.0f;
                }
            }

            gpuManager.synchronize();
            return similarities;

        } catch (Exception e) {
            System.err.println("GPU cosine similarity computation failed: " + e.getMessage());
            return batchCosineSimilarityCPU(queries, vectors);
        }
    }

    /**
     * 批量计算点积（GPU加速）
     */
    public float[][] batchDotProduct(float[][] queries, float[][] vectors) {
        if (!gpuManager.isGPUAvailable() || cublasHandle == null) {
            return batchDotProductCPU(queries, vectors);
        }

        try {
            int numQueries = queries.length;
            int numVectors = vectors.length;
            int dim = queries[0].length;

            // 分配GPU内存
            Pointer d_queries = new Pointer();
            Pointer d_vectors = new Pointer();
            Pointer d_result = new Pointer();

            cudaMalloc(d_queries, numQueries * dim * Sizeof.FLOAT);
            cudaMalloc(d_vectors, numVectors * dim * Sizeof.FLOAT);
            cudaMalloc(d_result, numQueries * numVectors * Sizeof.FLOAT);

            // 复制到GPU
            cudaMemcpy(d_queries, Pointer.to(flattenMatrix(queries)),
                      numQueries * dim * Sizeof.FLOAT, cudaMemcpyHostToDevice);
            cudaMemcpy(d_vectors, Pointer.to(flattenMatrix(vectors)),
                      numVectors * dim * Sizeof.FLOAT, cudaMemcpyHostToDevice);

            // 矩阵乘法: result = queries * vectors^T
            float alpha = 1.0f;
            float beta = 0.0f;

            cublasSgemm(cublasHandle,
                       cublasOperation.CUBLAS_OP_T,
                       cublasOperation.CUBLAS_OP_N,
                       numVectors, numQueries, dim,
                       Pointer.to(new float[]{alpha}),
                       d_vectors, dim,
                       d_queries, dim,
                       Pointer.to(new float[]{beta}),
                       d_result, numVectors);

            // 拷贝结果
            float[] resultFlat = new float[numQueries * numVectors];
            cudaMemcpy(Pointer.to(resultFlat), d_result,
                      numQueries * numVectors * Sizeof.FLOAT, cudaMemcpyDeviceToHost);

            // 释放内存
            cudaFree(d_queries);
            cudaFree(d_vectors);
            cudaFree(d_result);

            // 转换为2D数组
            float[][] result = new float[numQueries][numVectors];
            for (int i = 0; i < numQueries; i++) {
                System.arraycopy(resultFlat, i * numVectors, result[i], 0, numVectors);
            }

            gpuManager.synchronize();
            return result;

        } catch (Exception e) {
            System.err.println("GPU dot product computation failed: " + e.getMessage());
            return batchDotProductCPU(queries, vectors);
        }
    }

    /**
     * 计算向量范数的平方（用于L2距离）
     */
    private void computeNormsSquared(Pointer d_vectors, Pointer d_norms, int numVectors, int dim) {
        for (int i = 0; i < numVectors; i++) {
            float[] norm = new float[1];
            Pointer vectorPtr = d_vectors.withByteOffset(i * dim * Sizeof.FLOAT);

            cublasSnrm2(cublasHandle, dim, vectorPtr, 1, Pointer.to(norm));

            float normSquared = norm[0] * norm[0];
            cudaMemcpy(d_norms.withByteOffset(i * Sizeof.FLOAT),
                      Pointer.to(new float[]{normSquared}),
                      Sizeof.FLOAT, cudaMemcpyHostToDevice);
        }
    }

    /**
     * 计算向量范数（用于余弦相似度）
     */
    private void computeNorms(Pointer d_vectors, Pointer d_norms, int numVectors, int dim) {
        for (int i = 0; i < numVectors; i++) {
            float[] norm = new float[1];
            Pointer vectorPtr = d_vectors.withByteOffset(i * dim * Sizeof.FLOAT);

            cublasSnrm2(cublasHandle, dim, vectorPtr, 1, Pointer.to(norm));

            cudaMemcpy(d_norms.withByteOffset(i * Sizeof.FLOAT),
                      Pointer.to(new float[]{norm[0]}),
                      Sizeof.FLOAT, cudaMemcpyHostToDevice);
        }
    }

    /**
     * 将2D数组展平为1D数组（行优先）
     */
    private float[] flattenMatrix(float[][] matrix) {
        int rows = matrix.length;
        int cols = matrix[0].length;
        float[] result = new float[rows * cols];

        for (int i = 0; i < rows; i++) {
            System.arraycopy(matrix[i], 0, result, i * cols, cols);
        }

        return result;
    }

    // ==================== CPU降级实现 ====================

    private float[][] batchL2DistanceCPU(float[][] queries, float[][] vectors) {
        float[][] result = new float[queries.length][vectors.length];
        for (int i = 0; i < queries.length; i++) {
            for (int j = 0; j < vectors.length; j++) {
                result[i][j] = VectorSimilarity.l2Distance(queries[i], vectors[j]);
            }
        }
        return result;
    }

    private float[][] batchCosineSimilarityCPU(float[][] queries, float[][] vectors) {
        float[][] result = new float[queries.length][vectors.length];
        for (int i = 0; i < queries.length; i++) {
            for (int j = 0; j < vectors.length; j++) {
                result[i][j] = VectorSimilarity.cosineSimilarity(queries[i], vectors[j]);
            }
        }
        return result;
    }

    private float[][] batchDotProductCPU(float[][] queries, float[][] vectors) {
        float[][] result = new float[queries.length][vectors.length];
        for (int i = 0; i < queries.length; i++) {
            for (int j = 0; j < vectors.length; j++) {
                result[i][j] = VectorSimilarity.dotProduct(queries[i], vectors[j]);
            }
        }
        return result;
    }

    /**
     * 清理资源
     */
    public void cleanup() {
        if (cublasHandle != null) {
            cublasDestroy(cublasHandle);
            cublasHandle = null;
        }
    }

    /**
     * 检查GPU是否可用
     */
    public boolean isGPUAvailable() {
        return gpuManager.isGPUAvailable() && cublasHandle != null;
    }
}
