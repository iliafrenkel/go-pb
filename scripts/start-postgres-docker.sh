#!/bin/bash
# Starts Postgres database with Adminer using Docker compose

set -e

if ! command -v docker &> /dev/null; then
    echo "Error: Docker could not be found."
    exit 1
fi

echo "Starting Postgres database..."
docker compose -f docker-compose-dev.yaml up -d

echo "Development environment started."
echo "DB Server: localhost:5432"
echo "Adminer:   http://localhost:8888"
echo ""
echo "To stop the environment, run:"
echo "docker compose -f docker-compose-dev.yaml down"
