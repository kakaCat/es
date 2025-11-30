package com.es.plugin.vector.ivf;

import org.apache.lucene.document.Field;
import org.apache.lucene.document.FieldType;
import org.apache.lucene.index.IndexOptions;
import org.apache.lucene.util.BytesRef;

import java.nio.ByteBuffer;
import java.util.List;

/**
 * Lucene field for storing vector data.
 */
public class VectorField extends Field {

    private static final FieldType VECTOR_FIELD_TYPE = createFieldType();

    private static FieldType createFieldType() {
        FieldType ft = new FieldType();
        ft.setTokenized(false);
        ft.setIndexOptions(IndexOptions.NONE);
        ft.setStored(true);
        ft.setOmitNorms(true);
        ft.freeze();
        return ft;
    }

    /**
     * Create a vector field from a list of floats.
     *
     * @param name Field name
     * @param vector Vector values
     */
    public VectorField(String name, List<Float> vector) {
        super(name, encodeVector(vector), VECTOR_FIELD_TYPE);
    }

    /**
     * Create a vector field from a float array.
     *
     * @param name Field name
     * @param vector Vector values
     */
    public VectorField(String name, float[] vector) {
        super(name, encodeVector(vector), VECTOR_FIELD_TYPE);
    }

    /**
     * Encode a list of floats to BytesRef for storage.
     */
    private static BytesRef encodeVector(List<Float> vector) {
        float[] array = new float[vector.size()];
        for (int i = 0; i < vector.size(); i++) {
            array[i] = vector.get(i);
        }
        return encodeVector(array);
    }

    /**
     * Encode a float array to BytesRef for storage.
     */
    private static BytesRef encodeVector(float[] vector) {
        ByteBuffer buffer = ByteBuffer.allocate(vector.length * Float.BYTES);
        for (float value : vector) {
            buffer.putFloat(value);
        }
        return new BytesRef(buffer.array());
    }

    /**
     * Decode BytesRef back to float array.
     */
    public static float[] decodeVector(BytesRef bytes) {
        ByteBuffer buffer = ByteBuffer.wrap(bytes.bytes, bytes.offset, bytes.length);
        float[] vector = new float[bytes.length / Float.BYTES];
        for (int i = 0; i < vector.length; i++) {
            vector[i] = buffer.getFloat();
        }
        return vector;
    }

    /**
     * Get the vector as a float array.
     */
    public float[] getVector() {
        return decodeVector((BytesRef) fieldsData);
    }
}
