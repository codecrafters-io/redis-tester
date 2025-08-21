#!/bin/bash

set -e

cd "$(dirname "$0")/.."

# Build image
docker_build() {
    docker build -t redis-tester -f docker_ci/ubuntu.Dockerfile .
}

case "$1" in
  test)
    echo "🧪 Running tests using Docker container"
    docker_build
    docker run --rm -it -v "$(pwd)":/app -w /app redis-tester make test
    ;;
  record_fixtures)
    echo "📝 Recording fixtures using Docker container"
    docker_build
    docker run --rm -it -v "$(pwd)":/app -w /app redis-tester make record_fixtures
    ;;
  *)
    echo "Usage:"
    echo "$0 [test|record_fixtures]"
    exit 1
    ;;
esac
