#!/bin/bash

# Convenience script to test the Nobl9 action from the root directory

set -e

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}Nobl9 GitHub Action - Quick Test${NC}"
echo "======================================"
echo

# Check if we're in the right directory
if [ ! -f "action/go.mod" ]; then
    echo "Error: This script should be run from the project root directory"
    echo "Current directory: $(pwd)"
    echo "Expected to find: action/go.mod"
    exit 1
fi

# Navigate to action directory
cd action

# Load environment variables if .env exists in parent directory (root)
if [ -f "../.env" ]; then
    echo "Loading environment variables from ../.env"
    export $(cat ../.env | grep -v '^#' | xargs)
fi

# Check if credentials are available
if [ -z "$NOBL9_CLIENT_ID" ] || [ -z "$NOBL9_CLIENT_SECRET" ]; then
    echo "Warning: Nobl9 credentials not found in environment variables"
    echo "Please set your credentials:"
    echo "export NOBL9_CLIENT_ID=\"your-client-id\""
    echo "export NOBL9_CLIENT_SECRET=\"your-client-secret\""
    echo
    echo "Or edit the .env file in the root directory"
    exit 1
fi

echo -e "${GREEN}✓ Credentials found${NC}"
echo

# Check if binary exists
if [ ! -f "nobl9-action" ]; then
    echo "Building application..."
    go build -o nobl9-action cmd/main.go
fi

echo -e "${GREEN}✓ Binary ready${NC}"
echo

# Run the test
echo "Running Nobl9 action test..."
echo "Command: ./nobl9-action process --dry-run --file-pattern \"test-*.yaml\" --client-id \"$NOBL9_CLIENT_ID\" --client-secret \"***\""
echo

./nobl9-action process --dry-run --file-pattern "test-*.yaml" --client-id "$NOBL9_CLIENT_ID" --client-secret "$NOBL9_CLIENT_SECRET"

echo
echo -e "${GREEN}✓ Test completed successfully!${NC}"
echo
echo "Next steps:"
echo "1. Try with your own YAML files"
echo "2. Test validation: ./nobl9-action validate --file-pattern \"*.yaml\""
echo "3. Run Go tests: go test ./... -v"
echo "4. Read the full guide: docs/local-development-setup.md" 