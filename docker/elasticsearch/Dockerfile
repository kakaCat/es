FROM gradle:8-jdk17 AS builder
WORKDIR /app
COPY es-plugin/ .
RUN gradle assemble --no-daemon

FROM docker.elastic.co/elasticsearch/elasticsearch:8.15.3

# Create plugin directory
RUN mkdir -p /usr/share/elasticsearch/plugins/es-vector-ivf

# Copy built plugin jar and descriptor
COPY --from=builder /app/build/libs/*.jar /usr/share/elasticsearch/plugins/es-vector-ivf/
COPY --from=builder /app/src/main/resources/plugin-descriptor.properties /usr/share/elasticsearch/plugins/es-vector-ivf/

# Fix permissions
USER root
RUN chown -R elasticsearch:elasticsearch /usr/share/elasticsearch/plugins/es-vector-ivf
USER elasticsearch
