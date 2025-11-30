#!/usr/bin/env bash
set -euo pipefail

# Build script for ES Serverless Manager

VERSION=${VERSION:-latest}
IMAGE_NAME=${IMAGE_NAME:-es-serverless-manager}
REGISTRY=${REGISTRY:-""}  # Set to your registry, e.g., "docker.io/yourusername/"

# Full image name
if [ -n "$REGISTRY" ]; then
  FULL_IMAGE_NAME="${REGISTRY}${IMAGE_NAME}:${VERSION}"
else
  FULL_IMAGE_NAME="${IMAGE_NAME}:${VERSION}"
fi

echo "Building ${FULL_IMAGE_NAME}"

# Build the Docker image
docker build -t "${FULL_IMAGE_NAME}" -f server/Dockerfile .

echo "Build completed successfully!"

# Optionally push to registry
if [ "${PUSH:-false}" = "true" ]; then
  echo "Pushing ${FULL_IMAGE_NAME} to registry..."
  docker push "${FULL_IMAGE_NAME}"
  echo "Push completed!"
fi