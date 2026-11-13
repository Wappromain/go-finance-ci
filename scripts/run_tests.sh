#!/usr/bin/env bash
# Builds the test image, runs every test inside it, and exits with the
# tests' status: 0 if they all pass, non zero if any fail.
# Usage: ./scripts/run_tests.sh
set -euo pipefail

cd "$(dirname "$0")/.."   # run from the repository root

IMAGE="loan-api-test"

echo "==> Building $IMAGE"
docker build -f Dockerfile.test -t "$IMAGE" .

echo "==> Running tests"
# set -e is on, so if the tests fail this script stops here and
# returns the same non zero status that docker run returned.
docker run --rm "$IMAGE"

echo "==> All tests passed"
