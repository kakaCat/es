package com.es.plugin.vector.ivf;

import org.elasticsearch.action.ActionRequest;
import org.elasticsearch.action.ActionRequestValidationException;
import org.elasticsearch.action.ActionResponse;
import org.elasticsearch.action.ActionType;
import org.elasticsearch.common.io.stream.StreamInput;
import org.elasticsearch.common.io.stream.StreamOutput;
import org.elasticsearch.xcontent.ToXContentObject;
import org.elasticsearch.xcontent.XContentBuilder;

import java.io.IOException;
import java.util.Objects;

/**
 * Action for training IVF index.
 */
public class TrainIVFIndexAction extends ActionType<TrainIVFIndexAction.Response> {

    public static final TrainIVFIndexAction INSTANCE = new TrainIVFIndexAction();
    public static final String NAME = "indices:admin/ivf/train";

    private TrainIVFIndexAction() {
        super(NAME, Response::new);
    }

    /**
     * Request to train an IVF index.
     */
    public static class Request extends ActionRequest {
        private String indexName;
        private String fieldName;
        private int nlist;
        private String metric;

        public Request() {}

        public Request(String indexName, String fieldName, int nlist, String metric) {
            this.indexName = indexName;
            this.fieldName = fieldName;
            this.nlist = nlist;
            this.metric = metric;
        }

        public Request(StreamInput in) throws IOException {
            super(in);
            this.indexName = in.readString();
            this.fieldName = in.readString();
            this.nlist = in.readInt();
            this.metric = in.readString();
        }

        @Override
        public void writeTo(StreamOutput out) throws IOException {
            super.writeTo(out);
            out.writeString(indexName);
            out.writeString(fieldName);
            out.writeInt(nlist);
            out.writeString(metric);
        }

        @Override
        public ActionRequestValidationException validate() {
            if (indexName == null || indexName.isEmpty()) {
                ActionRequestValidationException e = new ActionRequestValidationException();
                e.addValidationError("indexName is required");
                return e;
            }
            if (fieldName == null || fieldName.isEmpty()) {
                ActionRequestValidationException e = new ActionRequestValidationException();
                e.addValidationError("fieldName is required");
                return e;
            }
            return null;
        }

        public String getIndexName() {
            return indexName;
        }

        public String getFieldName() {
            return fieldName;
        }

        public int getNlist() {
            return nlist;
        }

        public String getMetric() {
            return metric;
        }
    }

    /**
     * Response from training an IVF index.
     */
    public static class Response extends ActionResponse implements ToXContentObject {
        private boolean success;
        private String message;
        private int vectorCount;
        private int nlist;

        public Response() {}

        public Response(boolean success, String message, int vectorCount, int nlist) {
            this.success = success;
            this.message = message;
            this.vectorCount = vectorCount;
            this.nlist = nlist;
        }

        public Response(StreamInput in) throws IOException {
            super(in);
            this.success = in.readBoolean();
            this.message = in.readString();
            this.vectorCount = in.readInt();
            this.nlist = in.readInt();
        }

        @Override
        public void writeTo(StreamOutput out) throws IOException {
            out.writeBoolean(success);
            out.writeString(message);
            out.writeInt(vectorCount);
            out.writeInt(nlist);
        }

        @Override
        public XContentBuilder toXContent(XContentBuilder builder, Params params) throws IOException {
            builder.startObject();
            builder.field("success", success);
            builder.field("message", message);
            builder.field("vector_count", vectorCount);
            builder.field("nlist", nlist);
            builder.endObject();
            return builder;
        }

        public boolean isSuccess() {
            return success;
        }

        public String getMessage() {
            return message;
        }
    }
}
