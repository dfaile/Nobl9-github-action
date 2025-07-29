#!/bin/bash
# Test the action using Docker

set -e

# Load environment variables
if [ -f ".env" ]; then
    export $(cat .env | grep -v '^#' | xargs)
fi

# Test with Docker
echo "Testing with Docker..."
docker run --rm \
  -e NOBL9_CLIENT_ID="$NOBL9_CLIENT_ID" \
  -e NOBL9_CLIENT_SECRET="$NOBL9_CLIENT_SECRET" \
  -v "$(pwd):/workspace" \
  -w /workspace \
  nobl9-action:local \
  process --dry-run --file-pattern "test-*.yaml" --client-id "$NOBL9_CLIENT_ID" --client-secret "$NOBL9_CLIENT_SECRET"

echo "Docker test completed!"
