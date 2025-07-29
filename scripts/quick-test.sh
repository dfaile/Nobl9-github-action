#!/bin/bash

# Quick Test Script for Nobl9 GitHub Action
# This script provides a quick way to test the action with your credentials

set -e

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}Nobl9 GitHub Action - Quick Test${NC}"
echo "======================================"
echo

# Check if we're in the right directory
if [ ! -f "action/go.mod" ]; then
    echo -e "${YELLOW}Warning: This script should be run from the project root directory${NC}"
    echo "Current directory: $(pwd)"
    echo "Expected to find: action/go.mod"
    echo
fi

# Check if credentials are provided
if [ -z "$NOBL9_CLIENT_ID" ] || [ -z "$NOBL9_CLIENT_SECRET" ]; then
    echo -e "${YELLOW}Nobl9 credentials not found in environment variables${NC}"
    echo "Please set your credentials:"
    echo "export NOBL9_CLIENT_ID=\"your-client-id\""
    echo "export NOBL9_CLIENT_SECRET=\"your-client-secret\""
    echo
    echo "Or run the setup script first:"
    echo "./scripts/setup-local.sh"
    exit 1
fi

echo -e "${GREEN}✓ Nobl9 credentials found${NC}"
echo

# Navigate to action directory
cd action

# Check if binary exists
if [ ! -f "nobl9-action" ]; then
    echo -e "${YELLOW}Binary not found. Building...${NC}"
    go build -o nobl9-action cmd/main.go
fi

echo -e "${GREEN}✓ Binary ready${NC}"
echo

# Create a simple test file if it doesn't exist
if [ ! -f "test-quick.yaml" ]; then
    echo "Creating test file..."
    cat > test-quick.yaml << 'EOF'
apiVersion: n9/v1alpha
kind: Project
metadata:
  name: quick-test-project
  displayName: Quick Test Project
  description: A quick test project
spec:
  description: Quick test project for local testing
EOF
    echo -e "${GREEN}✓ Test file created${NC}"
fi

echo "Running quick test..."
echo "Command: ./nobl9-action --dry-run --files test-quick.yaml --verbose"
echo

# Run the test
./nobl9-action --dry-run --files test-quick.yaml --verbose

echo
echo -e "${GREEN}✓ Quick test completed!${NC}"
echo
echo "Next steps:"
echo "1. Check the output above for any errors"
echo "2. Try with your own YAML files:"
echo "   ./nobl9-action --dry-run --files '*.yaml'"
echo "3. Run the full test suite:"
echo "   go test ./... -v"
echo "4. Read the full setup guide:"
echo "   docs/local-development-setup.md" 