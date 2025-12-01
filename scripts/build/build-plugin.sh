#!/usr/bin/env bash
set -euo pipefail

# Build script for ES IVF Plugin

PLUGIN_DIR="es-plugin"
BUILD_DIR="${PLUGIN_DIR}/build"

echo "Building Elasticsearch IVF Plugin..."

# Check if Gradle is installed
if ! command -v gradle &> /dev/null; then
    echo "Gradle is not installed. Please install Gradle to build the plugin."
    exit 1
fi

# Navigate to plugin directory
cd "${PLUGIN_DIR}"

# Clean previous builds
./gradlew clean

# Build the plugin
./gradlew assemble

# Check if build was successful
if [ -f "${BUILD_DIR}/libs/es-vector-ivf-plugin-1.0.0.jar" ]; then
    echo "Plugin built successfully!"
    echo "Plugin location: ${BUILD_DIR}/libs/es-vector-ivf-plugin-1.0.0.jar"
else
    echo "Plugin build failed!"
    exit 1
fi

# Return to original directory
cd ..