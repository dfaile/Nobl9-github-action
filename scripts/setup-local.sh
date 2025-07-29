#!/bin/bash

# Nobl9 GitHub Action - Local Development Setup Script
# This script automates the setup process for local development

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Function to check Go version
check_go_version() {
    if ! command_exists go; then
        print_error "Go is not installed. Please install Go 1.21+ from https://golang.org/dl/"
        exit 1
    fi
    
    GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
    REQUIRED_VERSION="1.21"
    
    if [ "$(printf '%s\n' "$REQUIRED_VERSION" "$GO_VERSION" | sort -V | head -n1)" != "$REQUIRED_VERSION" ]; then
        print_error "Go version $GO_VERSION is too old. Please install Go 1.21+"
        exit 1
    fi
    
    print_success "Go version $GO_VERSION is compatible"
}

# Function to check Docker
check_docker() {
    if ! command_exists docker; then
        print_warning "Docker is not installed. Some features may not work."
        print_warning "Install Docker from https://www.docker.com/products/docker-desktop/"
    else
        print_success "Docker is available"
    fi
}

# Function to check Node.js
check_nodejs() {
    if ! command_exists node; then
        print_warning "Node.js is not installed. Backstage template testing will not be available."
        print_warning "Install Node.js 18+ from https://nodejs.org/"
    else
        NODE_VERSION=$(node --version | sed 's/v//')
        REQUIRED_VERSION="18.0.0"
        
        if [ "$(printf '%s\n' "$REQUIRED_VERSION" "$NODE_VERSION" | sort -V | head -n1)" != "$REQUIRED_VERSION" ]; then
            print_warning "Node.js version $NODE_VERSION may be too old. Consider upgrading to 18+"
        else
            print_success "Node.js version $NODE_VERSION is compatible"
        fi
    fi
}

# Function to setup environment variables
setup_environment() {
    print_status "Setting up environment variables..."
    
    # Check if .env file exists
    if [ -f ".env" ]; then
        print_warning ".env file already exists. Backing up to .env.backup"
        cp .env .env.backup
    fi
    
    # Create .env file
    cat > .env << 'EOF'
# Nobl9 GitHub Action - Local Development Environment
# Replace these values with your actual Nobl9 credentials

# Nobl9 API Credentials
NOBL9_CLIENT_ID="your-client-id-here"
NOBL9_CLIENT_SECRET="your-client-secret-here"

# GitHub Actions Environment (for testing)
GITHUB_WORKSPACE="$(pwd)"
GITHUB_ACTIONS="true"

# Logging Configuration
LOG_LEVEL="info"
LOG_FORMAT="json"

# Development Settings
DRY_RUN="true"
VALIDATE_ONLY="false"
EOF

    print_success "Created .env file"
    print_warning "Please edit .env file and add your actual Nobl9 credentials"
}

# Function to install Go dependencies
install_dependencies() {
    print_status "Installing Go dependencies..."
    
    # Check if we're in the action directory or need to navigate there
    if [ -f "action/go.mod" ]; then
        cd action
        print_status "Navigated to action directory"
    elif [ ! -f "go.mod" ]; then
        print_error "go.mod file not found. Are you in the correct directory?"
        print_error "Expected to find: action/go.mod or go.mod"
        exit 1
    fi
    
    go mod download
    go mod tidy
    
    print_success "Go dependencies installed"
}

# Function to build the application
build_application() {
    print_status "Building the application..."
    
    # Ensure we're in the action directory
    if [ ! -f "go.mod" ] && [ -f "action/go.mod" ]; then
        cd action
    fi
    
    # Build Go binary
    go build -o nobl9-action cmd/main.go
    
    if [ -f "nobl9-action" ]; then
        print_success "Application built successfully"
    else
        print_error "Failed to build application"
        exit 1
    fi
    
    # Build Docker image if Docker is available
    if command_exists docker; then
        print_status "Building Docker image..."
        docker build -t nobl9-action:local .
        print_success "Docker image built successfully"
    fi
}

# Function to create test files
create_test_files() {
    print_status "Creating test files..."
    
    # Ensure we're in the action directory for test files
    if [ ! -f "go.mod" ] && [ -f "action/go.mod" ]; then
        cd action
    fi
    
    # Create test project file
    cat > test-project.yaml << 'EOF'
apiVersion: n9/v1alpha
kind: Project
metadata:
  name: test-project
  displayName: Test Project
  description: A test project for local development
spec:
  description: Test project created during local development
EOF

    # Create test role binding file
    cat > test-rolebinding.yaml << 'EOF'
apiVersion: n9/v1alpha
kind: RoleBinding
metadata:
  name: test-rolebinding
  project: test-project
spec:
  role: admin
  users:
    - user@example.com
EOF

    # Create test SLO file
    cat > test-slo.yaml << 'EOF'
apiVersion: n9/v1alpha
kind: SLO
metadata:
  name: test-slo
  project: test-project
  displayName: Test SLO
spec:
  description: Test SLO for local development
  indicator:
    metricSource:
      name: test-metric-source
  objectives:
    - displayName: Test Objective
      target: 0.99
      value: 0.99
      op: lte
EOF

    print_success "Test files created:"
    echo "  - test-project.yaml"
    echo "  - test-rolebinding.yaml"
    echo "  - test-slo.yaml"
}

# Function to run initial tests
run_tests() {
    print_status "Running initial tests..."
    
    # Ensure we're in the action directory
    if [ ! -f "go.mod" ] && [ -f "action/go.mod" ]; then
        cd action
    fi
    
    # Test the application
    if [ -f "nobl9-action" ]; then
        print_status "Testing application help..."
        ./nobl9-action --help > /dev/null 2>&1 && print_success "Application help test passed" || print_error "Application help test failed"
        
        print_status "Testing dry-run mode..."
        ./nobl9-action --dry-run --files "test-*.yaml" > /dev/null 2>&1 && print_success "Dry-run test passed" || print_warning "Dry-run test failed (may need credentials)"
    fi
    
    # Run Go tests
    print_status "Running Go tests..."
    go test ./pkg/... -v > test-results.log 2>&1 && print_success "Go tests completed" || print_warning "Some Go tests failed (check test-results.log)"
}

# Function to create development scripts
create_dev_scripts() {
    print_status "Creating development scripts..."
    
    # Ensure we're in the action directory
    if [ ! -f "go.mod" ] && [ -f "action/go.mod" ]; then
        cd action
    fi
    
    # Create scripts directory if it doesn't exist
    mkdir -p scripts
    
    # Create test script
    cat > scripts/test-local.sh << 'EOF'
#!/bin/bash
# Test the action locally

set -e

# Load environment variables
if [ -f ".env" ]; then
    export $(cat .env | grep -v '^#' | xargs)
fi

# Run tests
echo "Running local tests..."
./nobl9-action --dry-run --files "test-*.yaml" --verbose

echo "Tests completed!"
EOF

    # Create docker test script
    cat > scripts/test-docker.sh << 'EOF'
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
  --dry-run --files "test-*.yaml" --verbose

echo "Docker test completed!"
EOF

    # Make scripts executable
    chmod +x scripts/test-local.sh
    chmod +x scripts/test-docker.sh
    
    print_success "Development scripts created:"
    echo "  - scripts/test-local.sh"
    echo "  - scripts/test-docker.sh"
}

# Function to display next steps
display_next_steps() {
    echo
    print_success "Local development setup completed!"
    echo
    echo "Next steps:"
    echo "1. Edit .env file and add your Nobl9 credentials:"
    echo "   nano .env"
    echo
    echo "2. Test the application:"
    echo "   ./scripts/test-local.sh"
    echo
    echo "3. Test with Docker (if available):"
    echo "   ./scripts/test-docker.sh"
    echo
    echo "4. Run Go tests:"
    echo "   go test ./... -v"
    echo
    echo "5. Read the full setup guide:"
    echo "   docs/local-development-setup.md"
    echo
    echo "6. Start developing!"
    echo
}

# Main setup function
main() {
    echo "=========================================="
    echo "Nobl9 GitHub Action - Local Setup Script"
    echo "=========================================="
    echo
    
    # Check prerequisites
    print_status "Checking prerequisites..."
    check_go_version
    check_docker
    check_nodejs
    echo
    
    # Setup environment
    setup_environment
    echo
    
    # Install dependencies
    install_dependencies
    echo
    
    # Build application
    build_application
    echo
    
    # Create test files
    create_test_files
    echo
    
    # Run initial tests
    run_tests
    echo
    
    # Create development scripts
    create_dev_scripts
    echo
    
    # Display next steps
    display_next_steps
}

# Run main function
main "$@" 