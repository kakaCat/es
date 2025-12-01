package com.es.plugin.vector.ivf;

import java.util.ArrayList;
import java.util.List;
import java.util.Random;

/**
 * Simplified KMeans clustering for IVF algorithm.
 * Uses random initialization and fixed iteration count for simplicity.
 */
public class SimpleKMeansTrainer {

    private final int nlist;  // Number of clusters
    private final int maxIterations;  // Maximum iterations
    private final float convergenceThreshold;  // Early stopping threshold
    private final Random random;

    /**
     * Create a KMeans trainer.
     *
     * @param nlist Number of clusters to create
     * @param maxIterations Maximum number of iterations (default: 100)
     */
    public SimpleKMeansTrainer(int nlist, int maxIterations) {
        this.nlist = nlist;
        this.maxIterations = maxIterations;
        this.convergenceThreshold = 0.001f;  // Stop if centroids change less than 0.1%
        this.random = new Random(42);  // Fixed seed for reproducibility
    }

    /**
     * Create a KMeans trainer with default iterations.
     *
     * @param nlist Number of clusters to create
     */
    public SimpleKMeansTrainer(int nlist) {
        this(nlist, 100);
    }

    /**
     * Train KMeans clustering on the given vectors.
     *
     * @param vectors Training vectors
     * @return Array of cluster centroids
     */
    public float[][] train(float[][] vectors) {
        if (vectors == null || vectors.length == 0) {
            throw new IllegalArgumentException("Training vectors cannot be empty");
        }
        if (vectors.length < nlist) {
            throw new IllegalArgumentException("Number of vectors must be >= nlist");
        }

        int dimension = vectors[0].length;

        // Step 1: Random initialization
        float[][] centroids = randomInitialization(vectors, dimension);

        // Step 2: Iterative refinement
        int[] assignments = new int[vectors.length];

        for (int iter = 0; iter < maxIterations; iter++) {
            // Assign each vector to nearest cluster
            boolean changed = assignVectorsToClusters(vectors, centroids, assignments);

            // Update centroids
            float[][] newCentroids = updateCentroids(vectors, assignments, dimension);

            // Check for convergence
            if (!changed || hasConverged(centroids, newCentroids)) {
                System.out.println("KMeans converged at iteration " + (iter + 1));
                return newCentroids;
            }

            centroids = newCentroids;
        }

        System.out.println("KMeans finished after max iterations: " + maxIterations);
        return centroids;
    }

    /**
     * Initialize centroids by randomly selecting vectors.
     */
    private float[][] randomInitialization(float[][] vectors, int dimension) {
        float[][] centroids = new float[nlist][dimension];

        // Randomly select nlist vectors as initial centroids
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
     * Assign each vector to the nearest cluster.
     *
     * @return true if any assignment changed
     */
    private boolean assignVectorsToClusters(float[][] vectors, float[][] centroids, int[] assignments) {
        boolean changed = false;

        for (int i = 0; i < vectors.length; i++) {
            int nearestCluster = findNearestCluster(vectors[i], centroids);
            if (assignments[i] != nearestCluster) {
                assignments[i] = nearestCluster;
                changed = true;
            }
        }

        return changed;
    }

    /**
     * Find the nearest cluster for a given vector.
     */
    private int findNearestCluster(float[] vector, float[][] centroids) {
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
     * Update centroids based on current assignments.
     */
    private float[][] updateCentroids(float[][] vectors, int[] assignments, int dimension) {
        float[][] newCentroids = new float[nlist][dimension];
        int[] clusterSizes = new int[nlist];

        // Sum all vectors in each cluster
        for (int i = 0; i < vectors.length; i++) {
            int cluster = assignments[i];
            clusterSizes[cluster]++;
            for (int d = 0; d < dimension; d++) {
                newCentroids[cluster][d] += vectors[i][d];
            }
        }

        // Calculate average (centroid)
        for (int i = 0; i < nlist; i++) {
            if (clusterSizes[i] > 0) {
                for (int d = 0; d < dimension; d++) {
                    newCentroids[i][d] /= clusterSizes[i];
                }
            } else {
                // Handle empty cluster: reinitialize with random vector
                System.out.println("Warning: Cluster " + i + " is empty, reinitializing");
                int randomIdx = random.nextInt(vectors.length);
                System.arraycopy(vectors[randomIdx], 0, newCentroids[i], 0, dimension);
            }
        }

        return newCentroids;
    }

    /**
     * Check if centroids have converged.
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
     * Get cluster assignment for vectors.
     *
     * @param vectors Vectors to assign
     * @param centroids Cluster centroids
     * @return Array of cluster IDs for each vector
     */
    public static int[] assignClusters(float[][] vectors, float[][] centroids) {
        int[] assignments = new int[vectors.length];

        for (int i = 0; i < vectors.length; i++) {
            int nearestCluster = 0;
            float minDistance = Float.MAX_VALUE;

            for (int j = 0; j < centroids.length; j++) {
                float distance = VectorSimilarity.l2Distance(vectors[i], centroids[j]);
                if (distance < minDistance) {
                    minDistance = distance;
                    nearestCluster = j;
                }
            }

            assignments[i] = nearestCluster;
        }

        return assignments;
    }
}
