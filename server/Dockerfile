FROM golang:1.21-alpine AS builder
RUN apk add --no-cache git bash curl ca-certificates gcc musl-dev
WORKDIR /app
COPY server/go.mod server/go.sum* ./
RUN go mod download || true
COPY server/*.go ./
COPY server/k8s ./k8s
COPY server/terraform ./terraform
COPY server/helm ./helm
RUN CGO_ENABLED=0 GOOS=linux go build -o manager .

FROM alpine:3.20
RUN apk add --no-cache bash curl ca-certificates postgresql-client
ARG KUBECTL_VERSION=v1.30.3
RUN curl -fsSL -o /usr/local/bin/kubectl https://storage.googleapis.com/kubernetes-release/release/${KUBECTL_VERSION}/bin/linux/amd64/kubectl && \
    chmod +x /usr/local/bin/kubectl
WORKDIR /app
COPY --from=builder /app/manager /app/manager
COPY --from=builder /app/k8s /app/k8s
COPY --from=builder /app/terraform /app/terraform
COPY --from=builder /app/helm /app/helm
# Set execute permissions
ENV PATH="/usr/local/bin:${PATH}"
EXPOSE 8080
CMD ["/app/manager"]