#!/bin/bash
# Test the action locally

set -e

# Load environment variables
if [ -f ".env" ]; then
    export $(cat .env | grep -v '^#' | xargs)
fi

# Run tests
echo "Running local tests..."
./nobl9-action process --dry-run --file-pattern "test-*.yaml" --client-id "$NOBL9_CLIENT_ID" --client-secret "$NOBL9_CLIENT_SECRET"

echo "Tests completed!"
