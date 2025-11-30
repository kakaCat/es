package com.es.plugin.vector.ivf;

import org.apache.lucene.search.Query;
import org.apache.lucene.search.MatchAllDocsQuery;
import org.apache.lucene.search.BooleanQuery;
import org.apache.lucene.search.BooleanClause;
import org.apache.lucene.search.TermQuery;
import org.apache.lucene.index.Term;
import org.elasticsearch.common.io.stream.StreamInput;
import org.elasticsearch.common.io.stream.StreamOutput;
import org.elasticsearch.xcontent.XContentBuilder;
import org.elasticsearch.xcontent.XContentParser;
import org.elasticsearch.index.query.AbstractQueryBuilder;
import org.elasticsearch.index.query.SearchExecutionContext;

import java.io.IOException;
import java.util.Objects;
import java.util.List;
import java.util.Map;
import java.util.concurrent.ConcurrentHashMap;

public class IVFQueryBuilder extends AbstractQueryBuilder<IVFQueryBuilder> {
    public static final String NAME = "ann";

    // Cache for IVF indexes (index_name -> IVF index)
    private static final Map<String, InvertedFileIndex> indexCache = new ConcurrentHashMap<>();

    private String field;
    private float[] vector;
    private String algorithm = "ivf";
    private int nprobe = 10;
    private int k = 10;  // Number of results to return

    public IVFQueryBuilder() {}

    public IVFQueryBuilder(StreamInput in) throws IOException {
        super(in);
        field = in.readString();
        vector = in.readFloatArray();
        algorithm = in.readString();
        nprobe = in.readInt();
        k = in.readInt();
    }

    @Override
    protected void doWriteTo(StreamOutput out) throws IOException {
        out.writeString(field);
        out.writeFloatArray(vector);
        out.writeString(algorithm);
        out.writeInt(nprobe);
        out.writeInt(k);
    }

    @Override
    protected void doXContent(XContentBuilder builder, Params params) throws IOException {
        builder.startObject(NAME);
        builder.field("field", field);
        builder.field("vector", vector);
        builder.field("algorithm", algorithm);
        builder.field("nprobe", nprobe);
        builder.field("k", k);
        printBoostAndQueryName(builder);
        builder.endObject();
    }

    public static IVFQueryBuilder fromXContent(XContentParser parser) throws IOException {
        String field = null;
        float[] vector = null;
        String algorithm = "ivf";
        int nprobe = 10;
        int k = 10;

        XContentParser.Token token;
        String currentName = null;
        while ((token = parser.nextToken()) != XContentParser.Token.END_OBJECT) {
            if (token == XContentParser.Token.FIELD_NAME) {
                currentName = parser.currentName();
            } else if (token == XContentParser.Token.VALUE_STRING) {
                if ("field".equals(currentName)) {
                    field = parser.text();
                } else if ("algorithm".equals(currentName)) {
                    algorithm = parser.text();
                }
            } else if (token == XContentParser.Token.VALUE_NUMBER) {
                if ("nprobe".equals(currentName)) {
                    nprobe = parser.intValue();
                } else if ("k".equals(currentName)) {
                    k = parser.intValue();
                }
            } else if (token == XContentParser.Token.START_ARRAY) {
                if ("vector".equals(currentName)) {
                    vector = parser.floatArray();
                }
            }
        }

        IVFQueryBuilder builder = new IVFQueryBuilder();
        builder.field(field);
        builder.vector(vector);
        builder.algorithm(algorithm);
        builder.nprobe(nprobe);
        builder.k(k);
        return builder;
    }

    @Override
    protected Query doToQuery(SearchExecutionContext context) throws IOException {
        // Get or create IVF index for this field
        String indexName = context.index().getName() + "_" + field;
        InvertedFileIndex ivfIndex = getOrCreateIndex(indexName, context);

        // Perform IVF search
        List<InvertedFileIndex.SearchResult> results = ivfIndex.search(vector, k, nprobe);

        if (results.isEmpty()) {
            // Return empty query if no results found
            return new MatchAllDocsQuery();
        }

        // Create a boolean query with all matching doc IDs
        BooleanQuery.Builder booleanBuilder = new BooleanQuery.Builder();

        for (InvertedFileIndex.SearchResult result : results) {
            // Add each matching document as a term query
            TermQuery termQuery = new TermQuery(new Term("_id", result.docId));
            booleanBuilder.add(termQuery, BooleanClause.Occur.SHOULD);
        }

        return booleanBuilder.build();
    }

    /**
     * Get or create IVF index for the field.
     * This is a simplified implementation - in production, you'd want to:
     * 1. Load index from persistent storage
     * 2. Build index incrementally as documents are added
     * 3. Handle index updates and deletions
     */
    private InvertedFileIndex getOrCreateIndex(String indexName, SearchExecutionContext context) throws IOException {
        return indexCache.computeIfAbsent(indexName, key -> {
            try {
                // Try to load existing index from file
                String indexPath = getIndexPath(indexName);
                java.io.File indexFile = new java.io.File(indexPath);

                if (indexFile.exists()) {
                    System.out.println("Loading existing IVF index from: " + indexPath);
                    return InvertedFileIndex.load(indexPath);
                } else {
                    System.out.println("Creating new IVF index: " + indexName);
                    // Create new index with default parameters
                    // In production, these should come from index settings
                    InvertedFileIndex newIndex = new InvertedFileIndex(100, vector.length, "l2");

                    // Index needs to be trained before use
                    // This is a placeholder - in production, training should happen
                    // when enough documents are indexed
                    System.out.println("Warning: IVF index not trained yet. Returning empty results.");

                    return newIndex;
                }
            } catch (Exception e) {
                throw new RuntimeException("Failed to load/create IVF index", e);
            }
        });
    }

    /**
     * Get file path for storing IVF index.
     */
    private String getIndexPath(String indexName) {
        // In production, this should use Elasticsearch's data directory
        String dataDir = System.getProperty("es.path.data", "/tmp/es-ivf-indexes");
        return dataDir + "/" + indexName + ".ivf";
    }

    /**
     * Static method to manually add vectors to an index.
     * This should be called during document indexing.
     */
    public static void addVectorToIndex(String indexName, String docId, float[] vector, Map<String, Object> metadata) {
        try {
            InvertedFileIndex index = indexCache.get(indexName);
            if (index != null && index.isTrained()) {
                index.addVector(docId, vector, metadata);

                // Periodically save index to disk
                if (index.size() % 1000 == 0) {
                    String indexPath = "/tmp/es-ivf-indexes/" + indexName + ".ivf";
                    index.save(indexPath);
                }
            }
        } catch (Exception e) {
            System.err.println("Failed to add vector to IVF index: " + e.getMessage());
        }
    }

    /**
     * Static method to train an index with vectors.
     * Should be called when index is created or when sufficient vectors are available.
     */
    public static void trainIndex(String indexName, float[][] trainingVectors, int dimension, String metricType) {
        try {
            InvertedFileIndex index = new InvertedFileIndex(100, dimension, metricType);
            index.train(trainingVectors);
            indexCache.put(indexName, index);

            // Save trained index
            String indexPath = "/tmp/es-ivf-indexes/" + indexName + ".ivf";
            new java.io.File("/tmp/es-ivf-indexes").mkdirs();
            index.save(indexPath);

            System.out.println("IVF index trained and saved: " + indexName);
        } catch (Exception e) {
            System.err.println("Failed to train IVF index: " + e.getMessage());
        }
    }

    @Override
    protected boolean doEquals(IVFQueryBuilder other) {
        return Objects.equals(field, other.field) &&
               java.util.Arrays.equals(vector, other.vector) &&
               Objects.equals(algorithm, other.algorithm) &&
               nprobe == other.nprobe &&
               k == other.k;
    }

    @Override
    protected int doHashCode() {
        return Objects.hash(field, java.util.Arrays.hashCode(vector), algorithm, nprobe, k);
    }

    @Override
    public String getWriteableName() {
        return NAME;
    }

    // Getters and setters
    public IVFQueryBuilder field(String field) {
        this.field = field;
        return this;
    }

    public String field() {
        return field;
    }

    public IVFQueryBuilder vector(float[] vector) {
        this.vector = vector;
        return this;
    }

    public float[] vector() {
        return vector;
    }

    public IVFQueryBuilder algorithm(String algorithm) {
        this.algorithm = algorithm;
        return this;
    }

    public String algorithm() {
        return algorithm;
    }

    public IVFQueryBuilder nprobe(int nprobe) {
        this.nprobe = nprobe;
        return this;
    }

    public int nprobe() {
        return nprobe;
    }

    public IVFQueryBuilder k(int k) {
        this.k = k;
        return this;
    }

    public int k() {
        return k;
    }
}