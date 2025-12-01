package com.es.plugin.vector.ivf;

/**
 * Vector similarity calculation utilities.
 * Supports L2 distance, Cosine similarity, and Dot product.
 */
public class VectorSimilarity {

    /**
     * Calculate L2 (Euclidean) distance between two vectors.
     * Lower values indicate higher similarity.
     *
     * @param v1 First vector
     * @param v2 Second vector
     * @return L2 distance
     */
    public static float l2Distance(float[] v1, float[] v2) {
        if (v1.length != v2.length) {
            throw new IllegalArgumentException("Vectors must have the same dimension");
        }

        float sum = 0.0f;
        for (int i = 0; i < v1.length; i++) {
            float diff = v1[i] - v2[i];
            sum += diff * diff;
        }
        return (float) Math.sqrt(sum);
    }

    /**
     * Calculate Cosine similarity between two vectors.
     * Returns value in range [-1, 1], where 1 indicates identical direction.
     *
     * @param v1 First vector
     * @param v2 Second vector
     * @return Cosine similarity
     */
    public static float cosineSimilarity(float[] v1, float[] v2) {
        if (v1.length != v2.length) {
            throw new IllegalArgumentException("Vectors must have the same dimension");
        }

        float dotProduct = 0.0f;
        float normV1 = 0.0f;
        float normV2 = 0.0f;

        for (int i = 0; i < v1.length; i++) {
            dotProduct += v1[i] * v2[i];
            normV1 += v1[i] * v1[i];
            normV2 += v2[i] * v2[i];
        }

        if (normV1 == 0.0f || normV2 == 0.0f) {
            return 0.0f;
        }

        return dotProduct / ((float) Math.sqrt(normV1) * (float) Math.sqrt(normV2));
    }

    /**
     * Calculate Dot product between two vectors.
     * Higher values indicate higher similarity.
     *
     * @param v1 First vector
     * @param v2 Second vector
     * @return Dot product
     */
    public static float dotProduct(float[] v1, float[] v2) {
        if (v1.length != v2.length) {
            throw new IllegalArgumentException("Vectors must have the same dimension");
        }

        float sum = 0.0f;
        for (int i = 0; i < v1.length; i++) {
            sum += v1[i] * v2[i];
        }
        return sum;
    }

    /**
     * Batch calculate L2 distances from a query vector to multiple vectors.
     *
     * @param query Query vector
     * @param vectors Array of vectors to compare against
     * @return Array of L2 distances
     */
    public static float[] batchL2Distance(float[] query, float[][] vectors) {
        float[] distances = new float[vectors.length];
        for (int i = 0; i < vectors.length; i++) {
            distances[i] = l2Distance(query, vectors[i]);
        }
        return distances;
    }

    /**
     * Batch calculate Cosine similarities from a query vector to multiple vectors.
     *
     * @param query Query vector
     * @param vectors Array of vectors to compare against
     * @return Array of cosine similarities
     */
    public static float[] batchCosineSimilarity(float[] query, float[][] vectors) {
        float[] similarities = new float[vectors.length];
        for (int i = 0; i < vectors.length; i++) {
            similarities[i] = cosineSimilarity(query, vectors[i]);
        }
        return similarities;
    }

    /**
     * Batch calculate Dot products from a query vector to multiple vectors.
     *
     * @param query Query vector
     * @param vectors Array of vectors to compare against
     * @return Array of dot products
     */
    public static float[] batchDotProduct(float[] query, float[][] vectors) {
        float[] products = new float[vectors.length];
        for (int i = 0; i < vectors.length; i++) {
            products[i] = dotProduct(query, vectors[i]);
        }
        return products;
    }

    /**
     * Find indices of k nearest neighbors based on L2 distance.
     *
     * @param query Query vector
     * @param vectors Array of vectors to search
     * @param k Number of neighbors to find
     * @return Indices of k nearest neighbors (sorted by distance, ascending)
     */
    public static int[] findKNearestL2(float[] query, float[][] vectors, int k) {
        float[] distances = batchL2Distance(query, vectors);
        return findKSmallestIndices(distances, k);
    }

    /**
     * Find indices of k vectors with highest cosine similarity.
     *
     * @param query Query vector
     * @param vectors Array of vectors to search
     * @param k Number of neighbors to find
     * @return Indices of k most similar vectors (sorted by similarity, descending)
     */
    public static int[] findKNearestCosine(float[] query, float[][] vectors, int k) {
        float[] similarities = batchCosineSimilarity(query, vectors);
        return findKLargestIndices(similarities, k);
    }

    /**
     * Find indices of k smallest values in an array.
     */
    private static int[] findKSmallestIndices(float[] values, int k) {
        int n = values.length;
        k = Math.min(k, n);

        Integer[] indices = new Integer[n];
        for (int i = 0; i < n; i++) {
            indices[i] = i;
        }

        // Sort indices by values (ascending)
        java.util.Arrays.sort(indices, (a, b) -> Float.compare(values[a], values[b]));

        int[] result = new int[k];
        for (int i = 0; i < k; i++) {
            result[i] = indices[i];
        }
        return result;
    }

    /**
     * Find indices of k largest values in an array.
     */
    private static int[] findKLargestIndices(float[] values, int k) {
        int n = values.length;
        k = Math.min(k, n);

        Integer[] indices = new Integer[n];
        for (int i = 0; i < n; i++) {
            indices[i] = i;
        }

        // Sort indices by values (descending)
        java.util.Arrays.sort(indices, (a, b) -> Float.compare(values[b], values[a]));

        int[] result = new int[k];
        for (int i = 0; i < k; i++) {
            result[i] = indices[i];
        }
        return result;
    }
}
