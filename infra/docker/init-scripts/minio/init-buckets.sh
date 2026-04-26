#!/bin/bash
# DeepWrite Minio Bucket Initialization Script
# Creates required buckets on startup

set -e

echo "Initializing Minio buckets..."

# Wait for Minio to be ready
until mc alias set local http://localhost:9000 "${MINIO_ROOT_USER}" "${MINIO_ROOT_PASSWORD}" 2>/dev/null; do
    echo "Waiting for Minio to start..."
    sleep 2
done

# Create buckets
echo "Creating buckets..."
mc mb -p local/deepwrite-images || true
mc mb -p local/deepwrite-documents || true
mc mb -p local/deepwrite-code || true
mc mb -p local/deepwrite-exports || true

# Set bucket policies (public read for images, private for others)
mc anonymous set download local/deepwrite-images || true
mc anonymous set private local/deepwrite-documents || true
mc anonymous set private local/deepwrite-code || true
mc anonymous set private local/deepwrite-exports || true

echo "Minio buckets initialized successfully!"
