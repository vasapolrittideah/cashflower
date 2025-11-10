#!/bin/bash
set -e

SERVICE="$1"

if [ -z "$SERVICE" ]; then
  echo "Usage: compile_service.sh <service-name>"
  exit 1
fi

BINARY_NAME="$SERVICE"

echo "Building service: $SERVICE"
echo "Output binary: $BINARY_NAME"

mkdir -p ./build
rm -f "build/$BINARY_NAME"

CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o "build/$BINARY_NAME" "./services/$SERVICE/cmd/main.go"

echo "Build completed: $BINARY_NAME"
