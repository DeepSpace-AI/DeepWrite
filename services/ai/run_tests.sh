#!/bin/bash
set -e

echo "============================================"
echo "AI Service Test Runner"
echo "============================================"
echo

# Check if config.yaml exists
if [ ! -f config.yaml ]; then
    echo "ERROR: config.yaml not found!"
    echo "Please copy config.yaml.example to config.yaml and update settings."
    exit 1
fi

# Create test data
echo "[Step 1] Creating test data..."
uv run python setup_test_data.py
echo

# Start server in background
echo "[Step 2] Starting AI service..."
uv run uvicorn main:app --host 0.0.0.0 --port 8010 &
SERVER_PID=$!
sleep 3
echo "Server started on http://localhost:8010 (PID: $SERVER_PID)"
echo

# Run tests
echo "[Step 3] Running tests..."
uv run python test_service.py
echo

# Cleanup
echo "[Step 4] Stopping server..."
kill $SERVER_PID 2>/dev/null || true

echo "============================================"
echo "Test complete!"
echo "============================================"