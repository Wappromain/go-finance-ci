#!/usr/bin/env bash
# Builds the API image and runs it until you press Ctrl+C.
# Usage: ./scripts/run_api.sh            (serves on port 8080)
#        PORT=9000 ./scripts/run_api.sh  (serves on port 9000)
set -euo pipefail

cd "$(dirname "$0")/.."   # run from the repository root

IMAGE="loan-api"
PORT="${PORT:-8080}"

echo "==> Building $IMAGE"
docker build -f Dockerfile.api -t "$IMAGE" .

echo "==> Starting API on http://localhost:$PORT (Ctrl+C to stop)"
# --rm deletes the container when it stops; -p maps your port to the container's 8080.
docker run --rm --name loan-api -p "$PORT:8080" "$IMAGE"
