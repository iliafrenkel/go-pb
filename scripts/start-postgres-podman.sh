#!/bin/bash
# Starts Postgres database with Adminer using Podman native Kubernetes support.

set -e

if ! command -v podman &> /dev/null; then
    echo "Error: podman could not be found."
    exit 1
fi

echo "Starting Postgres database..."
podman play kube podman-kube-dev.yaml

echo "Development environment started."
echo "DB Server: localhost:5432"
echo "Adminer:   http://localhost:8888"
echo ""
echo "To stop the environment, run:"
echo "podman play kube --down podman-kube-dev.yaml"
