package com.es.plugin.vector.ivf;

import java.util.ArrayList;
import java.util.List;
import java.util.Random;

/**
 * GPU加速的KMeans聚类训练器
 * 使用GPU进行批量距离计算，大幅提升训练速度
 */
public class GPUKMeansTrainer {

    private final int nlist;
    private final int maxIterations;
    private final float convergenceThreshold;
    private final Random random;
    private final GPUVectorSimilarity gpuSimilarity;
    private final boolean useGPU;

    public GPUKMeansTrainer(int nlist, int maxIterations) {
        this.nlist = nlist;
        this.maxIterations = maxIterations;
        this.convergenceThreshold = 0.001f;
        this.random = new Random(42);

        // 初始化GPU
        this.gpuSimilarity = new GPUVectorSimilarity();
        this.useGPU = gpuSimilarity.isGPUAvailable();

        if (useGPU) {
            System.out.println("GPU KMeans trainer initialized - GPU acceleration enabled");
        } else {
            System.out.println("GPU KMeans trainer initialized - using CPU fallback");
        }
    }

    public GPUKMeansTrainer(int nlist) {
        this(nlist, 100);
    }

    /**
     * 训练KMeans聚类（GPU加速）
     */
    public float[][] train(float[][] vectors) {
        if (vectors == null || vectors.length == 0) {
            throw new IllegalArgumentException("Training vectors cannot be empty");
        }
        if (vectors.length < nlist) {
            throw new IllegalArgumentException("Number of vectors must be >= nlist");
        }

        int dimension = vectors[0].length;
        int numVectors = vectors.length;

        System.out.println("Starting GPU KMeans training:");
        System.out.println("  Vectors: " + numVectors);
        System.out.println("  Dimension: " + dimension);
        System.out.println("  Clusters: " + nlist);
        System.out.println("  GPU enabled: " + useGPU);

        long startTime = System.currentTimeMillis();

        // Step 1: 随机初始化
        float[][] centroids = randomInitialization(vectors, dimension);

        // Step 2: 迭代优化
        int[] assignments = new int[vectors.length];

        for (int iter = 0; iter < maxIterations; iter++) {
            long iterStart = System.currentTimeMillis();

            // GPU批量分配向量到聚类
            boolean changed = assignVectorsToClustersGPU(vectors, centroids, assignments);

            // 更新质心
            float[][] newCentroids = updateCentroids(vectors, assignments, dimension);

            long iterTime = System.currentTimeMillis() - iterStart;

            // 检查收敛
            if (!changed || hasConverged(centroids, newCentroids)) {
                long totalTime = System.currentTimeMillis() - startTime;
                System.out.println("KMeans converged at iteration " + (iter + 1));
                System.out.println("Total training time: " + totalTime + " ms");
                System.out.println("Average iteration time: " + (totalTime / (iter + 1)) + " ms");
                return newCentroids;
            }

            centroids = newCentroids;

            if ((iter + 1) % 10 == 0) {
                System.out.println("Iteration " + (iter + 1) + " completed in " + iterTime + " ms");
            }
        }

        long totalTime = System.currentTimeMillis() - startTime;
        System.out.println("KMeans finished after max iterations: " + maxIterations);
        System.out.println("Total training time: " + totalTime + " ms");

        return centroids;
    }

    /**
     * 随机初始化质心
     */
    private float[][] randomInitialization(float[][] vectors, int dimension) {
        float[][] centroids = new float[nlist][dimension];

        List<Integer> selectedIndices = new ArrayList<>();
        while (selectedIndices.size() < nlist) {
            int idx = random.nextInt(vectors.length);
            if (!selectedIndices.contains(idx)) {
                selectedIndices.add(idx);
                System.arraycopy(vectors[idx], 0, centroids[selectedIndices.size() - 1], 0, dimension);
            }
        }

        return centroids;
    }

    /**
     * GPU加速的向量分配到聚类
     * 使用批量GPU计算所有向量到所有质心的距离
     */
    private boolean assignVectorsToClustersGPU(float[][] vectors, float[][] centroids, int[] assignments) {
        boolean changed = false;

        // 如果GPU可用且向量数量足够大，使用GPU批量计算
        if (useGPU && vectors.length > 100) {
            try {
                // GPU批量计算所有距离: [numVectors x nlist]
                float[][] distanceMatrix = gpuSimilarity.batchL2Distance(vectors, centroids);

                // 找到每个向量的最近质心
                for (int i = 0; i < vectors.length; i++) {
                    int nearestCluster = 0;
                    float minDistance = distanceMatrix[i][0];

                    for (int j = 1; j < nlist; j++) {
                        if (distanceMatrix[i][j] < minDistance) {
                            minDistance = distanceMatrix[i][j];
                            nearestCluster = j;
                        }
                    }

                    if (assignments[i] != nearestCluster) {
                        assignments[i] = nearestCluster;
                        changed = true;
                    }
                }
            } catch (Exception e) {
                System.err.println("GPU assignment failed, falling back to CPU: " + e.getMessage());
                // 降级到CPU
                changed = assignVectorsToCentersCPU(vectors, centroids, assignments);
            }
        } else {
            // CPU计算
            changed = assignVectorsToCentersCPU(vectors, centroids, assignments);
        }

        return changed;
    }

    /**
     * CPU版本的向量分配（降级使用）
     */
    private boolean assignVectorsToCentersCPU(float[][] vectors, float[][] centroids, int[] assignments) {
        boolean changed = false;

        for (int i = 0; i < vectors.length; i++) {
            int nearestCluster = findNearestClusterCPU(vectors[i], centroids);
            if (assignments[i] != nearestCluster) {
                assignments[i] = nearestCluster;
                changed = true;
            }
        }

        return changed;
    }

    /**
     * CPU版本的最近质心查找
     */
    private int findNearestClusterCPU(float[] vector, float[][] centroids) {
        int nearestCluster = 0;
        float minDistance = Float.MAX_VALUE;

        for (int i = 0; i < centroids.length; i++) {
            float distance = VectorSimilarity.l2Distance(vector, centroids[i]);
            if (distance < minDistance) {
                minDistance = distance;
                nearestCluster = i;
            }
        }

        return nearestCluster;
    }

    /**
     * 根据当前分配更新质心
     */
    private float[][] updateCentroids(float[][] vectors, int[] assignments, int dimension) {
        float[][] newCentroids = new float[nlist][dimension];
        int[] clusterSizes = new int[nlist];

        // 累加每个聚类中的向量
        for (int i = 0; i < vectors.length; i++) {
            int cluster = assignments[i];
            clusterSizes[cluster]++;
            for (int d = 0; d < dimension; d++) {
                newCentroids[cluster][d] += vectors[i][d];
            }
        }

        // 计算平均值（质心）
        for (int i = 0; i < nlist; i++) {
            if (clusterSizes[i] > 0) {
                for (int d = 0; d < dimension; d++) {
                    newCentroids[i][d] /= clusterSizes[i];
                }
            } else {
                // 处理空聚类：用随机向量重新初始化
                System.out.println("Warning: Cluster " + i + " is empty, reinitializing");
                int randomIdx = random.nextInt(vectors.length);
                System.arraycopy(vectors[randomIdx], 0, newCentroids[i], 0, dimension);
            }
        }

        return newCentroids;
    }

    /**
     * 检查质心是否收敛
     */
    private boolean hasConverged(float[][] oldCentroids, float[][] newCentroids) {
        float totalChange = 0.0f;
        float totalNorm = 0.0f;

        for (int i = 0; i < nlist; i++) {
            float distance = VectorSimilarity.l2Distance(oldCentroids[i], newCentroids[i]);
            totalChange += distance;

            float norm = 0.0f;
            for (float val : newCentroids[i]) {
                norm += val * val;
            }
            totalNorm += Math.sqrt(norm);
        }

        float relativeChange = totalChange / (totalNorm + 1e-8f);
        return relativeChange < convergenceThreshold;
    }

    /**
     * 清理GPU资源
     */
    public void cleanup() {
        if (gpuSimilarity != null) {
            gpuSimilarity.cleanup();
        }
    }

    /**
     * 检查是否使用GPU
     */
    public boolean isUsingGPU() {
        return useGPU;
    }
}
