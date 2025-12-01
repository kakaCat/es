package com.es.plugin.vector.ivf;

import org.junit.Before;
import org.junit.Test;
import static org.junit.Assert.*;

import java.util.*;

/**
 * Unit tests for IVF vector search components.
 */
public class IVFIndexTest {

    private static final int DIMENSION = 128;
    private static final int NUM_VECTORS = 1000;
    private static final int NLIST = 10;

    @Test
    public void testVectorSimilarityL2() {
        float[] v1 = {1.0f, 2.0f, 3.0f};
        float[] v2 = {1.0f, 2.0f, 3.0f};
        float[] v3 = {4.0f, 5.0f, 6.0f};

        // Same vector should have distance 0
        assertEquals(0.0f, VectorSimilarity.l2Distance(v1, v2), 0.0001f);

        // Different vectors should have distance > 0
        assertTrue(VectorSimilarity.l2Distance(v1, v3) > 0);

        // Distance should be symmetric
        assertEquals(
            VectorSimilarity.l2Distance(v1, v3),
            VectorSimilarity.l2Distance(v3, v1),
            0.0001f
        );
    }

    @Test
    public void testVectorSimilarityCosine() {
        float[] v1 = {1.0f, 0.0f, 0.0f};
        float[] v2 = {1.0f, 0.0f, 0.0f};
        float[] v3 = {0.0f, 1.0f, 0.0f};

        // Same direction should have similarity 1
        assertEquals(1.0f, VectorSimilarity.cosineSimilarity(v1, v2), 0.0001f);

        // Orthogonal vectors should have similarity 0
        assertEquals(0.0f, VectorSimilarity.cosineSimilarity(v1, v3), 0.0001f);

        // Cosine similarity is symmetric
        assertEquals(
            VectorSimilarity.cosineSimilarity(v1, v3),
            VectorSimilarity.cosineSimilarity(v3, v1),
            0.0001f
        );
    }

    @Test
    public void testVectorSimilarityDotProduct() {
        float[] v1 = {1.0f, 2.0f, 3.0f};
        float[] v2 = {4.0f, 5.0f, 6.0f};

        // Dot product: 1*4 + 2*5 + 3*6 = 32
        assertEquals(32.0f, VectorSimilarity.dotProduct(v1, v2), 0.0001f);

        // Dot product is symmetric
        assertEquals(
            VectorSimilarity.dotProduct(v1, v2),
            VectorSimilarity.dotProduct(v2, v1),
            0.0001f
        );
    }

    @Test
    public void testKMeansTraining() {
        // Generate random training data
        float[][] trainingVectors = generateRandomVectors(100, 10);

        SimpleKMeansTrainer trainer = new SimpleKMeansTrainer(5, 50);
        float[][] centroids = trainer.train(trainingVectors);

        // Should return correct number of centroids
        assertEquals(5, centroids.length);

        // Each centroid should have correct dimension
        for (float[] centroid : centroids) {
            assertEquals(10, centroid.length);
        }

        // Centroids should be distinct (not all zeros)
        for (float[] centroid : centroids) {
            float norm = 0;
            for (float val : centroid) {
                norm += val * val;
            }
            assertTrue("Centroid should not be all zeros", norm > 0.001f);
        }
    }

    @Test
    public void testIVFIndexTrainAndAdd() {
        // Create index
        InvertedFileIndex index = new InvertedFileIndex(10, 128, "l2");

        // Train with random vectors
        float[][] trainingVectors = generateRandomVectors(500, 128);
        index.train(trainingVectors);

        assertTrue("Index should be trained", index.isTrained());

        // Add vectors
        for (int i = 0; i < 100; i++) {
            float[] vector = generateRandomVector(128);
            Map<String, Object> metadata = new HashMap<>();
            metadata.put("id", i);

            index.addVector("doc_" + i, vector, metadata);
        }

        assertEquals("Index should contain 100 vectors", 100, index.size());
    }

    @Test
    public void testIVFIndexSearch() {
        // Create and train index
        InvertedFileIndex index = new InvertedFileIndex(10, 128, "l2");
        float[][] trainingVectors = generateRandomVectors(500, 128);
        index.train(trainingVectors);

        // Add test vectors
        List<float[]> testVectors = new ArrayList<>();
        for (int i = 0; i < 100; i++) {
            float[] vector = generateRandomVector(128);
            testVectors.add(vector);

            Map<String, Object> metadata = new HashMap<>();
            metadata.put("id", i);
            index.addVector("doc_" + i, vector, metadata);
        }

        // Search for first test vector
        float[] queryVector = testVectors.get(0);
        List<InvertedFileIndex.SearchResult> results = index.search(queryVector, 5, 3);

        assertNotNull("Search results should not be null", results);
        assertTrue("Should find at least one result", results.size() > 0);
        assertTrue("Should find at most 5 results", results.size() <= 5);

        // First result should be the query vector itself (distance ≈ 0)
        InvertedFileIndex.SearchResult topResult = results.get(0);
        assertEquals("Top result should be the query document", "doc_0", topResult.docId);
        assertTrue("Distance to itself should be very small", topResult.score < 0.001f);
    }

    @Test
    public void testIVFIndexSearchWithDifferentMetrics() {
        // Test with L2
        testSearchWithMetric("l2");

        // Test with Cosine
        testSearchWithMetric("cosine");

        // Test with Dot product
        testSearchWithMetric("dot");
    }

    private void testSearchWithMetric(String metric) {
        InvertedFileIndex index = new InvertedFileIndex(10, 32, metric);
        float[][] trainingVectors = generateRandomVectors(200, 32);
        index.train(trainingVectors);

        // Add vectors
        for (int i = 0; i < 50; i++) {
            float[] vector = generateRandomVector(32);
            index.addVector("doc_" + i, vector, null);
        }

        // Search
        float[] queryVector = generateRandomVector(32);
        List<InvertedFileIndex.SearchResult> results = index.search(queryVector, 10, 3);

        assertNotNull("Results should not be null for metric: " + metric, results);
        assertTrue("Should find results for metric: " + metric, results.size() > 0);
    }

    @Test
    public void testIVFIndexPersistence() throws Exception {
        // Create and train index
        InvertedFileIndex index = new InvertedFileIndex(10, 64, "l2");
        float[][] trainingVectors = generateRandomVectors(300, 64);
        index.train(trainingVectors);

        // Add vectors
        for (int i = 0; i < 50; i++) {
            float[] vector = generateRandomVector(64);
            index.addVector("doc_" + i, vector, null);
        }

        int originalSize = index.size();

        // Save to file
        String tempFile = "/tmp/test_ivf_index.ivf";
        index.save(tempFile);

        // Load from file
        InvertedFileIndex loadedIndex = InvertedFileIndex.load(tempFile);

        assertEquals("Loaded index should have same size", originalSize, loadedIndex.size());
        assertTrue("Loaded index should be trained", loadedIndex.isTrained());

        // Search should work on loaded index
        float[] queryVector = generateRandomVector(64);
        List<InvertedFileIndex.SearchResult> results = loadedIndex.search(queryVector, 5, 3);

        assertNotNull("Search should work on loaded index", results);
        assertTrue("Should find results", results.size() > 0);

        // Clean up
        new java.io.File(tempFile).delete();
    }

    @Test
    public void testIVFIndexStats() {
        InvertedFileIndex index = new InvertedFileIndex(10, 128, "l2");
        float[][] trainingVectors = generateRandomVectors(500, 128);
        index.train(trainingVectors);

        // Add vectors
        for (int i = 0; i < 100; i++) {
            float[] vector = generateRandomVector(128);
            index.addVector("doc_" + i, vector, null);
        }

        Map<String, Object> stats = index.getStats();

        assertEquals(10, stats.get("nlist"));
        assertEquals(128, stats.get("dimension"));
        assertEquals("l2", stats.get("metricType"));
        assertEquals(true, stats.get("isTrained"));
        assertEquals(100, stats.get("totalVectors"));
    }

    // Helper methods

    private float[] generateRandomVector(int dimension) {
        Random random = new Random();
        float[] vector = new float[dimension];
        for (int i = 0; i < dimension; i++) {
            vector[i] = random.nextFloat() * 2.0f - 1.0f;  // Range: [-1, 1]
        }
        return vector;
    }

    private float[][] generateRandomVectors(int count, int dimension) {
        float[][] vectors = new float[count][dimension];
        for (int i = 0; i < count; i++) {
            vectors[i] = generateRandomVector(dimension);
        }
        return vectors;
    }
}
