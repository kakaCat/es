package com.es.plugin.vector.ivf;

import java.io.*;
import java.util.*;

/**
 * Inverted File Index for IVF algorithm.
 * Stores vectors organized by cluster for efficient ANN search.
 */
public class InvertedFileIndex implements Serializable {

    private static final long serialVersionUID = 1L;

    private final int nlist;  // Number of clusters
    private final int dimension;  // Vector dimension
    private final String metricType;  // "l2", "cosine", or "dot"

    private float[][] centroids;  // Cluster centroids
    private Map<Integer, List<VectorDoc>> invertedLists;  // clusterId -> list of vectors
    private boolean isTrained;

    /**
     * Document containing a vector and metadata.
     */
    public static class VectorDoc implements Serializable {
        private static final long serialVersionUID = 1L;

        public final String docId;
        public final float[] vector;
        public final Map<String, Object> metadata;

        public VectorDoc(String docId, float[] vector, Map<String, Object> metadata) {
            this.docId = docId;
            this.vector = vector;
            this.metadata = metadata != null ? metadata : new HashMap<>();
        }
    }

    /**
     * Search result containing document ID, distance/score, and metadata.
     */
    public static class SearchResult implements Serializable {
        private static final long serialVersionUID = 1L;

        public final String docId;
        public final float score;
        public final Map<String, Object> metadata;

        public SearchResult(String docId, float score, Map<String, Object> metadata) {
            this.docId = docId;
            this.score = score;
            this.metadata = metadata;
        }
    }

    /**
     * Create an IVF index.
     *
     * @param nlist Number of clusters
     * @param dimension Vector dimension
     * @param metricType Distance metric: "l2", "cosine", or "dot"
     */
    public InvertedFileIndex(int nlist, int dimension, String metricType) {
        this.nlist = nlist;
        this.dimension = dimension;
        this.metricType = metricType;
        this.invertedLists = new HashMap<>();
        this.isTrained = false;

        // Initialize empty inverted lists
        for (int i = 0; i < nlist; i++) {
            invertedLists.put(i, new ArrayList<>());
        }
    }

    /**
     * Train the index using KMeans clustering.
     *
     * @param trainingVectors Vectors to use for training
     */
    public void train(float[][] trainingVectors) {
        if (trainingVectors == null || trainingVectors.length == 0) {
            throw new IllegalArgumentException("Training vectors cannot be empty");
        }

        System.out.println("Training IVF index with " + trainingVectors.length + " vectors...");

        SimpleKMeansTrainer trainer = new SimpleKMeansTrainer(nlist);
        this.centroids = trainer.train(trainingVectors);
        this.isTrained = true;

        System.out.println("IVF index trained successfully with " + nlist + " clusters");
    }

    /**
     * Add a vector to the index.
     *
     * @param docId Document ID
     * @param vector Vector to add
     * @param metadata Optional metadata
     */
    public void addVector(String docId, float[] vector, Map<String, Object> metadata) {
        if (!isTrained) {
            throw new IllegalStateException("Index must be trained before adding vectors");
        }
        if (vector.length != dimension) {
            throw new IllegalArgumentException("Vector dimension mismatch: expected " + dimension + ", got " + vector.length);
        }

        // Find nearest cluster
        int clusterId = findNearestCluster(vector);

        // Add to inverted list
        VectorDoc doc = new VectorDoc(docId, vector, metadata);
        invertedLists.get(clusterId).add(doc);
    }

    /**
     * Batch add vectors to the index.
     *
     * @param vectors List of vectors with their IDs and metadata
     */
    public void addVectors(List<VectorDoc> vectors) {
        for (VectorDoc doc : vectors) {
            addVector(doc.docId, doc.vector, doc.metadata);
        }
    }

    /**
     * Search for k nearest neighbors.
     *
     * @param queryVector Query vector
     * @param k Number of results to return
     * @param nprobe Number of clusters to search
     * @return List of search results
     */
    public List<SearchResult> search(float[] queryVector, int k, int nprobe) {
        if (!isTrained) {
            throw new IllegalStateException("Index must be trained before searching");
        }
        if (queryVector.length != dimension) {
            throw new IllegalArgumentException("Query vector dimension mismatch");
        }

        // Step 1: Find nprobe nearest clusters
        int[] nearestClusters = findNearestClusters(queryVector, nprobe);

        // Step 2: Collect all candidates from these clusters
        List<VectorDoc> candidates = new ArrayList<>();
        for (int clusterId : nearestClusters) {
            List<VectorDoc> clusterDocs = invertedLists.get(clusterId);
            if (clusterDocs != null) {
                candidates.addAll(clusterDocs);
            }
        }

        if (candidates.isEmpty()) {
            return new ArrayList<>();
        }

        // Step 3: Calculate distances/scores for all candidates
        List<SearchResult> results = new ArrayList<>();
        for (VectorDoc doc : candidates) {
            float score = calculateScore(queryVector, doc.vector);
            results.add(new SearchResult(doc.docId, score, doc.metadata));
        }

        // Step 4: Sort by score and return top-k
        sortResults(results);

        int resultSize = Math.min(k, results.size());
        return results.subList(0, resultSize);
    }

    /**
     * Find the nearest cluster for a vector.
     */
    private int findNearestCluster(float[] vector) {
        int nearestCluster = 0;
        float bestScore = calculateScore(vector, centroids[0]);

        for (int i = 1; i < nlist; i++) {
            float score = calculateScore(vector, centroids[i]);
            if (isBetterScore(score, bestScore)) {
                bestScore = score;
                nearestCluster = i;
            }
        }

        return nearestCluster;
    }

    /**
     * Find nprobe nearest clusters.
     */
    private int[] findNearestClusters(float[] vector, int nprobe) {
        nprobe = Math.min(nprobe, nlist);

        // Calculate scores for all clusters
        float[] scores = new float[nlist];
        for (int i = 0; i < nlist; i++) {
            scores[i] = calculateScore(vector, centroids[i]);
        }

        // Find indices of best nprobe clusters
        Integer[] indices = new Integer[nlist];
        for (int i = 0; i < nlist; i++) {
            indices[i] = i;
        }

        // Sort by score
        Arrays.sort(indices, (a, b) -> {
            if (metricType.equals("l2")) {
                return Float.compare(scores[a], scores[b]);  // Lower is better for L2
            } else {
                return Float.compare(scores[b], scores[a]);  // Higher is better for cosine/dot
            }
        });

        int[] result = new int[nprobe];
        for (int i = 0; i < nprobe; i++) {
            result[i] = indices[i];
        }

        return result;
    }

    /**
     * Calculate score between two vectors based on metric type.
     */
    private float calculateScore(float[] v1, float[] v2) {
        switch (metricType) {
            case "l2":
                return VectorSimilarity.l2Distance(v1, v2);
            case "cosine":
                return VectorSimilarity.cosineSimilarity(v1, v2);
            case "dot":
                return VectorSimilarity.dotProduct(v1, v2);
            default:
                throw new IllegalArgumentException("Unknown metric type: " + metricType);
        }
    }

    /**
     * Check if score1 is better than score2.
     */
    private boolean isBetterScore(float score1, float score2) {
        if (metricType.equals("l2")) {
            return score1 < score2;  // Lower is better for L2
        } else {
            return score1 > score2;  // Higher is better for cosine/dot
        }
    }

    /**
     * Sort results based on metric type.
     */
    private void sortResults(List<SearchResult> results) {
        results.sort((a, b) -> {
            if (metricType.equals("l2")) {
                return Float.compare(a.score, b.score);  // Ascending for L2
            } else {
                return Float.compare(b.score, a.score);  // Descending for cosine/dot
            }
        });
    }

    /**
     * Get index statistics.
     */
    public Map<String, Object> getStats() {
        Map<String, Object> stats = new HashMap<>();
        stats.put("nlist", nlist);
        stats.put("dimension", dimension);
        stats.put("metricType", metricType);
        stats.put("isTrained", isTrained);

        if (isTrained) {
            int totalVectors = 0;
            int minClusterSize = Integer.MAX_VALUE;
            int maxClusterSize = 0;

            for (List<VectorDoc> cluster : invertedLists.values()) {
                int size = cluster.size();
                totalVectors += size;
                minClusterSize = Math.min(minClusterSize, size);
                maxClusterSize = Math.max(maxClusterSize, size);
            }

            stats.put("totalVectors", totalVectors);
            stats.put("minClusterSize", minClusterSize);
            stats.put("maxClusterSize", maxClusterSize);
            stats.put("avgClusterSize", totalVectors / (float) nlist);
        }

        return stats;
    }

    /**
     * Save index to file.
     */
    public void save(String filepath) throws IOException {
        try (ObjectOutputStream oos = new ObjectOutputStream(new FileOutputStream(filepath))) {
            oos.writeObject(this);
        }
        System.out.println("IVF index saved to " + filepath);
    }

    /**
     * Load index from file.
     */
    public static InvertedFileIndex load(String filepath) throws IOException, ClassNotFoundException {
        try (ObjectInputStream ois = new ObjectInputStream(new FileInputStream(filepath))) {
            InvertedFileIndex index = (InvertedFileIndex) ois.readObject();
            System.out.println("IVF index loaded from " + filepath);
            return index;
        }
    }

    /**
     * Check if index is trained.
     */
    public boolean isTrained() {
        return isTrained;
    }

    /**
     * Get number of vectors in the index.
     */
    public int size() {
        int total = 0;
        for (List<VectorDoc> cluster : invertedLists.values()) {
            total += cluster.size();
        }
        return total;
    }

    /**
     * Clear all vectors from the index (keep training).
     */
    public void clear() {
        for (int i = 0; i < nlist; i++) {
            invertedLists.get(i).clear();
        }
    }
}
