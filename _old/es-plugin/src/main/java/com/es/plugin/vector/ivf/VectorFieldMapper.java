package com.es.plugin.vector.ivf;

import org.apache.lucene.document.FieldType;
import org.apache.lucene.index.IndexOptions;
import org.elasticsearch.common.xcontent.XContentParser;
import org.elasticsearch.index.mapper.FieldMapper;
import org.elasticsearch.index.mapper.Mapper;
import org.elasticsearch.index.mapper.MapperParsingException;
import org.elasticsearch.index.mapper.ParametrizedFieldMapper;
import org.elasticsearch.index.mapper.ParseContext;
import org.elasticsearch.index.mapper.TextSearchInfo;
import org.elasticsearch.index.mapper.TypeParser;
import org.elasticsearch.index.mapper.ValueFetcher;

import java.io.IOException;
import java.util.ArrayList;
import java.util.Arrays;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.function.Supplier;

public class VectorFieldMapper extends ParametrizedFieldMapper {

    public static final String CONTENT_TYPE = "vector";
    
    private final int dimension;
    private final String metric;
    private final int nlist;
    private final int nprobe;

    protected VectorFieldMapper(String simpleName, FieldType fieldType, MappedFieldType mappedFieldType,
                               int dimension, String metric, int nlist, int nprobe,
                               MultiFields multiFields, CopyTo copyTo) {
        super(simpleName, fieldType, mappedFieldType, multiFields, copyTo);
        this.dimension = dimension;
        this.metric = metric;
        this.nlist = nlist;
        this.nprobe = nprobe;
    }

    public static class TypeParser implements Mapper.TypeParser {
        @Override
        public Mapper.Builder<?> parse(String name, Map<String, Object> node, ParserContext parserContext)
                throws MapperParsingException {
            
            int dimension = -1;
            String metric = "l2";
            int nlist = 100;
            int nprobe = 10;
            
            for (Map.Entry<String, Object> entry : node.entrySet()) {
                String propName = entry.getKey();
                Object propNode = entry.getValue();
                if ("type".equals(propName)) {
                    // Skip type property
                } else if ("dimension".equals(propName)) {
                    dimension = Integer.parseInt(propNode.toString());
                } else if ("metric".equals(propName)) {
                    metric = propNode.toString();
                } else if ("nlist".equals(propName)) {
                    nlist = Integer.parseInt(propNode.toString());
                } else if ("nprobe".equals(propName)) {
                    nprobe = Integer.parseInt(propNode.toString());
                }
            }
            
            if (dimension <= 0) {
                throw new MapperParsingException("Field [" + name + "] requires a dimension to be set");
            }
            
            return new Builder(name, dimension, metric, nlist, nprobe);
        }
    }

    public static class Builder extends ParametrizedFieldMapper.Builder {
        private final int dimension;
        private final String metric;
        private final int nlist;
        private final int nprobe;
        
        private final Parameter<Map<String, String>> meta = Parameter.metaParam();

        public Builder(String name, int dimension, String metric, int nlist, int nprobe) {
            super(name);
            this.dimension = dimension;
            this.metric = metric;
            this.nlist = nlist;
            this.nprobe = nprobe;
        }

        @Override
        protected List<Parameter<?>> getParameters() {
            return Arrays.asList(meta);
        }

        @Override
        public VectorFieldMapper build(Mapper.BuilderContext context) {
            FieldType fieldType = new FieldType();
            fieldType.setTokenized(false);
            fieldType.setIndexOptions(IndexOptions.NONE);
            fieldType.freeze();
            
            return new VectorFieldMapper(
                name(), fieldType, new VectorFieldType(buildFullName(context), dimension, metric, nlist, nprobe),
                dimension, metric, nlist, nprobe,
                multiFieldsBuilder.build(this, context),
                copyTo.build()
            );
        }
    }

    public static class VectorFieldType extends ParametrizedFieldMapper.MappedFieldType {
        private final int dimension;
        private final String metric;
        private final int nlist;
        private final int nprobe;
        
        public VectorFieldType(String name, int dimension, String metric, int nlist, int nprobe) {
            super(name);
            this.dimension = dimension;
            this.metric = metric;
            this.nlist = nlist;
            this.nprobe = nprobe;
        }
        
        public int dimension() {
            return dimension;
        }
        
        public String metric() {
            return metric;
        }
        
        public int nlist() {
            return nlist;
        }
        
        public int nprobe() {
            return nprobe;
        }

        @Override
        public String typeName() {
            return CONTENT_TYPE;
        }

        @Override
        public ValueFetcher valueFetcher(QueryShardContext context, SearchLookup searchLookup, String format) {
            return SourceValueFetcher.identity(name(), context);
        }
    }

    @Override
    protected void parseCreateField(ParseContext context) throws IOException {
        // Parse vector data from the document
        context.path().add(simpleName());

        XContentParser parser = context.parser();
        XContentParser.Token token = parser.currentToken();

        if (token == XContentParser.Token.START_ARRAY) {
            List<Float> vectorList = new ArrayList<>();
            while ((token = parser.nextToken()) != XContentParser.Token.END_ARRAY) {
                vectorList.add(parser.floatValue());
            }

            // Validate dimension
            if (vectorList.size() != dimension) {
                throw new IllegalArgumentException(
                    "Vector dimension mismatch: expected " + dimension + ", got " + vectorList.size()
                );
            }

            // Convert to float array
            float[] vector = new float[vectorList.size()];
            for (int i = 0; i < vectorList.size(); i++) {
                vector[i] = vectorList.get(i);
            }

            // Store the vector in Lucene document
            context.doc().add(new VectorField(fieldType().name(), vector));

            // Add vector to IVF index
            try {
                String indexName = context.index().getName() + "_" + fieldType().name();
                String docId = context.id();

                // Extract metadata from source
                Map<String, Object> metadata = new HashMap<>();
                if (context.sourceToParse().source() != null) {
                    // Add any metadata you want to store with the vector
                    metadata.put("_index", context.index().getName());
                    metadata.put("_id", docId);
                }

                // Add vector to IVF index (will be used during search)
                IVFQueryBuilder.addVectorToIndex(indexName, docId, vector, metadata);

            } catch (Exception e) {
                // Log error but don't fail indexing
                System.err.println("Failed to add vector to IVF index: " + e.getMessage());
            }

        } else {
            throw new IllegalArgumentException("Vector field must be an array of numbers");
        }

        context.path().remove();
    }

    @Override
    protected void mergeOptions(FieldMapper other, List<String> conflicts) {
        // Handle merging of field options
    }

    @Override
    public VectorFieldType fieldType() {
        return (VectorFieldType) super.fieldType();
    }
}