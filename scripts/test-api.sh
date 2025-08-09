#!/bin/bash
set -e

echo "=== gitmgr HTTP API Test ==="
echo

# Start server in background
echo "Starting API server..."
./bin/gitmgr-server -addr=127.0.0.1:8081 &
SERVER_PID=$!

# Wait for server to start
sleep 2

# Function to cleanup
cleanup() {
    echo "Stopping server..."
    kill $SERVER_PID 2>/dev/null || true
    wait $SERVER_PID 2>/dev/null || true
}
trap cleanup EXIT

BASE_URL="http://127.0.0.1:8081"

echo "Testing health endpoint..."
curl -s "$BASE_URL/health" | jq .
echo

echo "Testing repository info..."
curl -s "$BASE_URL/v1/repo?path=." | jq .
echo

echo "Testing status..."
curl -s "$BASE_URL/v1/status?path=." | jq .
echo

echo "Testing log..."
curl -s "$BASE_URL/v1/log?path=.&max=3" | jq .
echo

echo "Testing diff..."
curl -s "$BASE_URL/v1/diff?path=.&stat=true" | jq .
echo

echo "Testing raw command..."
curl -s -X POST "$BASE_URL/v1/raw" \
  -H "Content-Type: application/json" \
  -d '{"path":".","args":["branch","--show-current"]}' | jq .
echo

echo "API tests completed successfully!"