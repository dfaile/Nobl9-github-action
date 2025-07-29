# Local Development Setup Guide

This guide will help you set up and test the Nobl9 GitHub Action on your local machine.

## Prerequisites

### Required Software
- **Go 1.21+** - [Download](https://golang.org/dl/)
- **Docker** - [Download](https://www.docker.com/products/docker-desktop/)
- **Git** - [Download](https://git-scm.com/downloads)
- **Node.js 18+** (for Backstage template testing) - [Download](https://nodejs.org/)

### Required Accounts
- **Nobl9 Account** - [Sign up](https://www.nobl9.com/)
- **GitHub Account** - [Sign up](https://github.com/)

## Step 1: Clone and Setup Repository

```bash
# Clone the repository
git clone https://github.com/your-org/nobl9-github-action.git
cd nobl9-github-action

# Navigate to the action directory
cd action

# Install Go dependencies
go mod download
go mod tidy
```

## Step 2: Configure Nobl9 Credentials

### Get Your Nobl9 Credentials

1. **Log into your Nobl9 account**
2. **Navigate to Settings → API Keys**
3. **Create a new API key** with the following permissions:
   - `projects:read`
   - `projects:write`
   - `rolebindings:read`
   - `rolebindings:write`
   - `users:read`

4. **Note down your credentials:**
   - Client ID (e.g., `0oa2s2ag3kxbcvyIa417`)
   - Client Secret (e.g., `IuMmTiqEvU7XWI1jLtwrrcBU8Tri2YmIfHdCI4Iz`)

### Set Environment Variables

```bash
# Set your Nobl9 credentials
export NOBL9_CLIENT_ID="your-client-id"
export NOBL9_CLIENT_SECRET="your-client-secret"

# Set GitHub Actions environment (optional)
export GITHUB_WORKSPACE="$(pwd)"
export GITHUB_ACTIONS="true"
```

## Step 3: Build the Action

```bash
# Build the Go binary
go build -o nobl9-action cmd/main.go

# Build Docker image
docker build -t nobl9-action:local .
```

## Step 4: Test the Action Components

### Test Go Packages

```bash
# Run all tests
go test ./... -v

# Run specific package tests
go test ./pkg/nobl9/... -v
go test ./pkg/parser/... -v
go test ./pkg/resolver/... -v
go test ./pkg/processor/... -v
go test ./pkg/scanner/... -v
go test ./pkg/validator/... -v
go test ./pkg/retry/... -v
go test ./pkg/errors/... -v
go test ./pkg/logger/... -v
go test ./pkg/config/... -v
```

### Test the Main Application

```bash
# Test with help
./nobl9-action --help

# Test with dry-run
./nobl9-action --dry-run --repo-path .

# Test with specific files
./nobl9-action --dry-run --files "*.yaml"

# Test with validation only
./nobl9-action --validate-only --repo-path .
```

## Step 5: Create Test Data

### Create Sample Nobl9 YAML Files

Create test files in your repository:

```bash
# Create a test project file
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

# Create a test role binding file
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
```

### Test File Processing

```bash
# Test processing your sample files
./nobl9-action --dry-run --files "test-*.yaml"

# Test with specific project
./nobl9-action --dry-run --files "test-project.yaml"
```

## Step 6: Test Backstage Template

### Setup Backstage Environment

```bash
# Navigate to template directory
cd ../template

# Install Node.js dependencies
npm install

# Create a test Backstage app
npx @backstage/create-app@latest test-backstage-app
cd test-backstage-app

# Install scaffolder plugin
npm install @backstage/plugin-scaffolder-backend

# Copy the template
cp -r ../../template/template ./packages/backend/src/plugins/scaffolder/templates/

# Start Backstage
npm start
```

### Test Template Generation

1. **Open Backstage** at `http://localhost:3000`
2. **Navigate to Create** in the sidebar
3. **Select "Nobl9 Project"** template
4. **Fill in the form** with test data:
   - Project Name: `test-project`
   - Display Name: `Test Project`
   - Description: `Test project for local development`
   - Owner: `your-email@example.com`
5. **Click "Create"** to generate the project

## Step 7: Test GitHub Action Integration

### Create Test Workflow

Create a test workflow file:

```bash
# Create test workflow
cat > .github/workflows/test-local.yml << 'EOF'
name: Test Local Action

on:
  workflow_dispatch:

jobs:
  test-action:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout
        uses: actions/checkout@v4

      - name: Test Nobl9 Action
        uses: ./
        with:
          client-id: ${{ secrets.NOBL9_CLIENT_ID }}
          client-secret: ${{ secrets.NOBL9_CLIENT_SECRET }}
          dry-run: true
          files: "test-*.yaml"
        env:
          NOBL9_CLIENT_ID: ${{ secrets.NOBL9_CLIENT_ID }}
          NOBL9_CLIENT_SECRET: ${{ secrets.NOBL9_CLIENT_SECRET }}
EOF
```

### Test with Docker

```bash
# Test the action using Docker
docker run --rm \
  -e NOBL9_CLIENT_ID="$NOBL9_CLIENT_ID" \
  -e NOBL9_CLIENT_SECRET="$NOBL9_CLIENT_SECRET" \
  -v "$(pwd):/workspace" \
  -w /workspace \
  nobl9-action:local \
  --dry-run --files "test-*.yaml"
```

## Step 8: Debugging and Troubleshooting

### Enable Debug Logging

```bash
# Set debug level logging
export LOG_LEVEL="debug"

# Run with verbose output
./nobl9-action --dry-run --files "test-*.yaml" --verbose
```

### Check Configuration

```bash
# Validate configuration
./nobl9-action --validate-only --repo-path .

# Test configuration loading
go run cmd/main.go --config-test
```

### Common Issues and Solutions

#### Issue: "Failed to connect to Nobl9"
**Solution:**
- Verify your credentials are correct
- Check network connectivity
- Ensure your Nobl9 account has the required permissions

#### Issue: "No YAML files found"
**Solution:**
- Check file patterns: `--files "*.yaml"`
- Verify files are in the correct directory
- Check file permissions

#### Issue: "Invalid YAML format"
**Solution:**
- Validate YAML syntax: `yamllint test-*.yaml`
- Check for proper Nobl9 API version and kind
- Ensure required fields are present

#### Issue: "User resolution failed"
**Solution:**
- Verify email addresses are valid
- Check if users exist in your Nobl9 organization
- Ensure your API key has user read permissions

## Step 9: Performance Testing

### Load Testing

```bash
# Create multiple test files
for i in {1..10}; do
  cat > "test-project-$i.yaml" << EOF
apiVersion: n9/v1alpha
kind: Project
metadata:
  name: test-project-$i
  displayName: Test Project $i
spec:
  description: Test project $i
EOF
done

# Test processing multiple files
./nobl9-action --dry-run --files "test-project-*.yaml"
```

### Benchmark Tests

```bash
# Run Go benchmarks
go test -bench=. ./pkg/...

# Run specific benchmarks
go test -bench=BenchmarkParseYAML ./pkg/parser/...
go test -bench=BenchmarkResolveEmails ./pkg/resolver/...
```

## Step 10: Integration Testing

### Test End-to-End Workflow

1. **Create a test repository** with Nobl9 YAML files
2. **Set up GitHub secrets** with your Nobl9 credentials
3. **Create a workflow** that uses the action
4. **Push changes** and monitor the workflow execution
5. **Verify results** in your Nobl9 dashboard

### Test Error Scenarios

```bash
# Test with invalid credentials
export NOBL9_CLIENT_ID="invalid"
./nobl9-action --dry-run --files "test-*.yaml"

# Test with non-existent files
./nobl9-action --dry-run --files "nonexistent-*.yaml"

# Test with invalid YAML
echo "invalid: yaml: content" > invalid.yaml
./nobl9-action --dry-run --files "invalid.yaml"
```

## Step 11: Development Workflow

### Code Changes

```bash
# Make changes to the code
# Run tests to ensure nothing is broken
go test ./... -v

# Build and test
go build -o nobl9-action cmd/main.go
./nobl9-action --dry-run --files "test-*.yaml"
```

### Docker Development

```bash
# Rebuild Docker image after changes
docker build -t nobl9-action:local .

# Test with Docker
docker run --rm \
  -e NOBL9_CLIENT_ID="$NOBL9_CLIENT_ID" \
  -e NOBL9_CLIENT_SECRET="$NOBL9_CLIENT_SECRET" \
  -v "$(pwd):/workspace" \
  -w /workspace \
  nobl9-action:local \
  --dry-run --files "test-*.yaml"
```

## Environment Variables Reference

| Variable | Description | Required | Default |
|----------|-------------|----------|---------|
| `NOBL9_CLIENT_ID` | Nobl9 API Client ID | Yes | - |
| `NOBL9_CLIENT_SECRET` | Nobl9 API Client Secret | Yes | - |
| `GITHUB_WORKSPACE` | GitHub Actions workspace path | No | Current directory |
| `GITHUB_ACTIONS` | Whether running in GitHub Actions | No | `false` |
| `LOG_LEVEL` | Logging level (debug, info, warn, error) | No | `info` |
| `LOG_FORMAT` | Log format (json, text) | No | `json` |

## Command Line Options

| Option | Description | Default |
|--------|-------------|---------|
| `--dry-run` | Validate files without making changes | `false` |
| `--validate-only` | Only validate files, don't process | `false` |
| `--files` | File pattern to process | `*.yaml` |
| `--repo-path` | Repository path to scan | Current directory |
| `--verbose` | Enable verbose logging | `false` |
| `--help` | Show help information | - |

## Next Steps

1. **Set up CI/CD** using the provided GitHub Actions workflows
2. **Configure monitoring** and alerting for the action
3. **Document your specific use cases** and configurations
4. **Contribute improvements** back to the project

## Support

- **Documentation**: Check the `docs/` directory for detailed guides
- **Issues**: Report bugs and feature requests on GitHub
- **Discussions**: Join community discussions for help and ideas

---

**Note**: Keep your Nobl9 credentials secure and never commit them to version control. Use environment variables or GitHub secrets for production deployments. 