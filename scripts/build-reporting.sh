#!/usr/bin/env bash
set -euo pipefail

# Build script for Reporting Service

ACTION=${1:-help}
IMAGE_NAME=${IMAGE_NAME:-es-serverless/reporting-service}
IMAGE_TAG=${IMAGE_TAG:-latest}
DOCKERFILE=${DOCKERFILE:-Dockerfile.reporting}

show_help() {
    echo "Usage: scripts/build-reporting.sh [ACTION]"
    echo ""
    echo "Actions:"
    echo "  help        - Show this help message"
    echo "  build       - Build the reporting service image"
    echo "  push        - Push the reporting service image to registry"
    echo "  build-push  - Build and push the reporting service image"
    echo ""
    echo "Environment variables:"
    echo "  IMAGE_NAME  - Docker image name (default: es-serverless/reporting-service)"
    echo "  IMAGE_TAG   - Docker image tag (default: latest)"
    echo "  DOCKERFILE  - Dockerfile to use (default: Dockerfile.reporting)"
}

build_image() {
    echo "Building reporting service image: ${IMAGE_NAME}:${IMAGE_TAG}"
    
    # Create a temporary directory for the build
    BUILD_DIR=$(mktemp -d)
    trap "rm -rf $BUILD_DIR" EXIT
    
    # Copy necessary files to the build directory
    cp server/reporting.go "$BUILD_DIR/"
    cp server/es_client.go "$BUILD_DIR/" 2>/dev/null || echo "Warning: es_client.go not found"
    cp "$DOCKERFILE" "$BUILD_DIR/Dockerfile" 2>/dev/null || create_default_dockerfile "$BUILD_DIR/Dockerfile"
    
    # Build the image
    docker build -t "${IMAGE_NAME}:${IMAGE_TAG}" "$BUILD_DIR"
    
    echo "Reporting service image built successfully!"
}

create_default_dockerfile() {
    local dockerfile_path=$1
    cat > "$dockerfile_path" << 'EOF'
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY reporting.go ./
COPY es_client.go ./ 2>/dev/null || echo "es_client.go not found"

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o reporting-service .

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the binary
COPY --from=builder /app/reporting-service .

# Copy config files if they exist
COPY config.yaml ./config/ 2>/dev/null || echo "No config file found"

EXPOSE 8080

CMD ["./reporting-service"]
EOF
}

push_image() {
    echo "Pushing reporting service image: ${IMAGE_NAME}:${IMAGE_TAG}"
    docker push "${IMAGE_NAME}:${IMAGE_TAG}"
    echo "Reporting service image pushed successfully!"
}

build_and_push() {
    build_image
    push_image
}

case "$ACTION" in
    help)
        show_help
        ;;
    build)
        build_image
        ;;
    push)
        push_image
        ;;
    build-push)
        build_and_push
        ;;
    *)
        echo "Unknown action: $ACTION"
        show_help
        exit 1
        ;;
esac