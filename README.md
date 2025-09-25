# Nobl9 Backstage Template & GitHub Action

An automated workflow solution for Nobl9 project and user role management through a Backstage template and GitHub Action. This project enables self-service creation of Nobl9 projects and automated role assignments for any valid Nobl9 user.

## Overview

This solution consists of two main components:

1. **Backstage Template** - A user-friendly form interface that allows users to create Nobl9 projects and assign roles to themselves and others
2. **GitHub Action** - An automated processor that reads Nobl9 YAML configurations, resolves email addresses to Okta User IDs, and deploys projects and role bindings to Nobl9

## Problem Statement

Manual creation of Nobl9 projects and role assignments is time-consuming and error-prone. Users need a self-service way to create projects and manage access without requiring platform team intervention.

## Solution

This project automates the entire lifecycle of Nobl9 project creation and user role management through:

- **Self-Service Project Creation** - Any valid Nobl9 user can create projects through a simple Backstage interface
- **Automated Role Management** - Automatically assign project owners and editors based on template input
- **Email-to-UserID Resolution** - Convert email addresses to Okta User IDs for role binding using the Nobl9 API
- **GitOps Workflow** - Use GitHub repository as the source of truth for Nobl9 configurations
- **Multi-User Support** - Allow assignment of multiple users with different roles in a single template submission

## Key Features

### Self-Service Project Creation
- **User-Friendly Interface** - Simple Backstage form that any valid Nobl9 user can access
- **No Platform Team Dependency** - Users can create projects without waiting for approval or manual intervention
- **Standardized Process** - Consistent project creation following Nobl9 best practices and naming conventions
- **Real-Time Validation** - Immediate feedback on project names, email addresses, and form requirements

### Automated Role Management
- **Dynamic User Assignment** - Add multiple users with different roles in a single template submission
- **Role-Based Access Control** - Support for `project-owner` and `project-editor` roles
- **Self-Assignment Capability** - Users can assign roles to themselves and others
- **Bulk User Management** - Efficiently manage team access without individual requests

### Email-to-UserID Resolution
- **Automatic Conversion** - Seamlessly convert email addresses to Okta User IDs using Nobl9 API
- **Error Handling** - Graceful handling of invalid or non-existent email addresses
- **Comprehensive Logging** - Detailed logs of all resolution attempts and results
- **User Validation** - Verify user existence before creating role bindings

### Additional Features
- **Backstage Integration** - Seamless integration with Backstage's template system
- **Official Nobl9 SDK** - Built using the [Nobl9 Go SDK](https://github.com/nobl9/nobl9-go) for reliable API interactions
- **Comprehensive Logging** - Detailed logs and error handling for all operations
- **Environment Support** - Multi-environment support (dev, staging, prod) with automatic environment detection
- **Security** - Secure credential management using GitHub secrets
- **Validation** - Robust validation of YAML configurations and user inputs
- **GitOps Workflow** - Full audit trail with Git-based configuration management
- **Local Development** - Complete local development environment with testing tools

## Architecture

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   Backstage     │    │   GitHub Repo    │    │     Nobl9       │
│   Template      │───▶│   (YAML Files)   │───▶│     API         │
│   (User Input)  │    │                  │    │                 │
└─────────────────┘    └──────────────────┘    └─────────────────┘
                                │
                                ▼
                       ┌──────────────────┐
                       │   GitHub Action  │
                       │   (Go + SDK)     │
                       └──────────────────┘
```

## Installation and Setup

### Prerequisites

- **Backstage Instance** - A properly configured Backstage instance with GitHub integration
- **Nobl9 Account** - Access to a Nobl9 environment with API credentials
- **GitHub Repository** - A repository for storing Nobl9 YAML configurations
- **Okta Integration** - Nobl9 must be configured with Okta for user management

### Local Development Setup

For developers who want to test and contribute to the project:

1. **Clone the Repository**
   ```bash
   git clone https://github.com/dfaile/Nobl9-github-action.git
   cd Nobl9-github-action
   ```

2. **Configure Credentials**
   ```bash
   # Create a .env file with your Nobl9 credentials
   echo "NOBL9_CLIENT_ID=your-client-id" > .env
   echo "NOBL9_CLIENT_SECRET=your-client-secret" >> .env
   ```

3. **Build and Test**
   ```bash
   cd action
   go build -o nobl9-action ./cmd/main.go
   ./nobl9-action --help
   ```

### Backstage Template Setup

1. **Add Template to Backstage Catalog**
   ```yaml
   # Add to your Backstage catalog-info.yaml
   apiVersion: backstage.io/v1alpha1
   kind: Template
   metadata:
     name: nobl9-project-template
     title: Create Nobl9 Project
     description: Create a new Nobl9 project with user role assignments
   spec:
     type: service
     owner: platform-team
     parameters:
       # Template parameters will be defined here
   ```

2. **Configure Template Repository**
   - The template is pre-configured to commit to a specific GitHub repository
   - Update the repository URL in the template configuration if needed

### GitHub Action Setup

1. **Add Action to Your Repository**
   ```yaml
   # .github/workflows/nobl9-sync.yml
   name: Nobl9 Project Sync
   on:
     push:
       paths:
         - '**/*.yaml'
         - '**/*.yml'
   
   jobs:
     sync-nobl9:
       runs-on: ubuntu-latest
       steps:
         - uses: actions/checkout@v4
         - name: Sync Nobl9 Projects
           uses: docker://docker.io/dfaile/nobl9-github-action:latest
           with:
             client-id: ${{ secrets.NOBL9_CLIENT_ID }}
             client-secret: ${{ secrets.NOBL9_CLIENT_SECRET }}
             dry-run: false
             file-pattern: "**/*.yaml"
   ```

2. **Configure Required Secrets**
   - `NOBL9_CLIENT_ID` - Your Nobl9 API client ID
   - `NOBL9_CLIENT_SECRET` - Your Nobl9 API client secret
   - `DOCKERHUB_USERNAME` - Your Docker Hub username (for releases)
   - `DOCKERHUB_TOKEN` - Your Docker Hub access token (for releases)

3. **Environment Configuration**
   - The action automatically detects the Nobl9 environment from your credentials
   - Supports multiple environments (dev, staging, prod)

4. **Docker Hub Setup (for Releases)**
   - Create a Docker Hub account if you don't have one
   - Generate an access token at [Docker Hub](https://hub.docker.com/) → Account Settings → Security
   - Add the token as `DOCKERHUB_TOKEN` secret in your GitHub repository
   - The release workflow will automatically build and push Docker images to your Docker Hub account

## Usage

### Using the GitHub Action

#### Basic Usage
```yaml
# .github/workflows/nobl9-sync.yml
name: Nobl9 Project Sync
on:
  push:
    paths: ['**/*.yaml']

jobs:
  sync:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
         - uses: docker://docker.io/dfaile/nobl9-github-action:latest
        with:
          client-id: ${{ secrets.NOBL9_CLIENT_ID }}
          client-secret: ${{ secrets.NOBL9_CLIENT_SECRET }}
          dry-run: false
          file-pattern: "**/*.yaml"
```

#### Action Inputs

| Input | Description | Required | Default |
|-------|-------------|----------|---------|
| `client-id` | Nobl9 API client ID | Yes | - |
| `client-secret` | Nobl9 API client secret | Yes | - |
| `dry-run` | Validate files without making changes | No | `false` |
| `file-pattern` | File pattern to process | No | `**/*.yaml` |
| `repo-path` | Repository path to scan | No | `.` |
| `force` | Force processing even if validation fails | No | `false` |
| `validate-only` | Only validate files, don't process | No | `false` |
| `log-level` | Log level (debug, info, warn, error) | No | `info` |
| `log-format` | Log format (json, text) | No | `json` |

#### Action Outputs

| Output | Description |
|--------|-------------|
| `processed-files` | Number of files processed |
| `projects-created` | Number of projects created |
| `projects-updated` | Number of projects updated |
| `role-bindings-created` | Number of role bindings created |
| `role-bindings-updated` | Number of role bindings updated |
| `users-resolved` | Number of email addresses resolved to User IDs |
| `users-unresolved` | Number of email addresses that couldn't be resolved |

### Using the Backstage Template

1. **Navigate to Backstage**
   - Go to your Backstage instance
   - Click "Create" in the sidebar

2. **Select Template**
   - Choose "Create Nobl9 Project" template

3. **Fill in Details**
   - **Project Name**: Unique project identifier (e.g., `my-team-project`)
   - **Display Name**: Human-readable project name (e.g., `My Team Project`)
   - **Description**: Project description
   - **Owner**: Your email address
   - **Additional Users**: Add team members with roles

4. **Submit**
   - Click "Create" to generate the project
   - The template will create YAML files in your repository
   - GitHub Actions will automatically process and deploy to Nobl9

### Nobl9 YAML Configuration

The action processes Nobl9 YAML files with the following structure:

#### Project Configuration
```yaml
apiVersion: n9/v1alpha
kind: Project
metadata:
  name: my-project
  displayName: My Project
  description: A sample Nobl9 project
spec:
  description: Project description
```

#### Role Binding Configuration
```yaml
apiVersion: n9/v1alpha
kind: RoleBinding
metadata:
  name: my-role-binding
  project: my-project
spec:
  role: project-owner
  users:
    - user@example.com
    - another.user@example.com
```

## GitHub Actions Workflows

This repository includes several GitHub Actions workflows for CI/CD and quality assurance:

### Available Workflows

1. **CI Workflow** (`.github/workflows/ci.yml`)
   - Runs on push and pull requests
   - Tests Go code compilation and unit tests
   - Runs linting with golangci-lint
   - Builds the action and tests Docker build

2. **Security Workflow** (`.github/workflows/security.yml`)
   - Runs on push and pull requests
   - Vulnerability scanning with Trivy
   - Go vulnerability checking with govulncheck
   - Secret scanning with TruffleHog and Gitleaks
   - ⚠️ **Note**: Automatic weekly scans are disabled to save GitHub Actions minutes

3. **Template Validation** (`.github/workflows/template-validation.yml`)
   - Runs on changes to template files
   - Validates YAML syntax with yamllint
   - Checks template structure and required fields

4. **Release Workflow** (`.github/workflows/release-simple.yml`)
   - Runs on tag pushes and manual dispatch
   - Builds multi-platform binaries
   - Builds and pushes Docker images to Docker Hub
   - Creates GitHub releases with changelog

### Workflow Features

- ✅ **Docker Hub Integration**: All workflows use Docker Hub for container images
- ✅ **Error Resilient**: Proper error handling and fallbacks
- ✅ **Resource Efficient**: Optimized to minimize GitHub Actions minutes usage
- ✅ **Multi-Platform**: Supports Linux and macOS builds
- ✅ **Security Focused**: Comprehensive security scanning and validation

## Development

### Local Development

1. **Setup Environment**
   ```bash
   cd action
   go mod download
   ```

2. **Test Locally**
   ```bash
   go build -o nobl9-action ./cmd/main.go
   ./nobl9-action --help
   ```

3. **Run Tests**
   ```bash
   cd action
   go test ./... -v
   ```

4. **Build and Test**
   ```bash
   cd action
   go build -o nobl9-action cmd/main.go
   ./nobl9-action process --dry-run --file-pattern "test-*.yaml" \
     --client-id "$NOBL9_CLIENT_ID" --client-secret "$NOBL9_CLIENT_SECRET"
   ```

### Project Structure

```
nobl9-github-action/
├── action/                    # GitHub Action source code
│   ├── cmd/                   # Main application entry point
│   ├── pkg/                   # Go packages
│   │   ├── config/           # Configuration management
│   │   ├── errors/           # Error handling
│   │   ├── logger/           # Logging utilities
│   │   ├── nobl9/            # Nobl9 API client
│   │   ├── parser/           # YAML parsing
│   │   ├── processor/        # File processing
│   │   ├── resolver/         # Email-to-UserID resolution
│   │   ├── retry/            # Retry logic
│   │   ├── scanner/          # File scanning
│   │   └── validator/        # Validation logic
│   ├── action.yml            # GitHub Action definition
│   └── Dockerfile            # Container definition
├── template/                  # Backstage template
│   ├── template.yaml         # Template definition
│   └── template/             # Template files
├── docs/                      # Documentation
└── .github/                   # GitHub workflows
```

## Troubleshooting

### Common Issues

1. **Authentication Errors**
   - Verify your Nobl9 credentials are correct
   - Check that your API key has the required permissions
   - Ensure Nobl9 is configured with Okta for user management

2. **User Resolution Failures**
   - Verify email addresses exist in your Nobl9 organization
   - Check that users are active in Okta
   - Review logs for specific error messages

3. **YAML Validation Errors**
   - Check YAML syntax using `yamllint`
   - Verify required fields are present
   - Ensure proper Nobl9 API version and kind

4. **GitHub Action Failures**
   - Check GitHub secrets are properly configured
   - Review action logs for detailed error messages
   - Verify file patterns match your YAML files

5. **Docker Hub Authentication Issues**
   - Verify `DOCKERHUB_USERNAME` and `DOCKERHUB_TOKEN` secrets are set
   - Check that the Docker Hub token has the correct permissions
   - Ensure the token is not expired

6. **Workflow Execution Issues**
   - Check that workflows are not disabled in repository settings
   - Verify branch protection rules allow workflow execution
   - Review workflow permissions and ensure they're properly configured

### Getting Help

- **Documentation**: Check the `docs/` directory for detailed guides
- **Issues**: Report bugs and feature requests on GitHub
- **Local Testing**: Use the local development setup for debugging

## Contributing

1. **Fork the Repository**
2. **Create a Feature Branch**
3. **Make Your Changes**
4. **Test Locally**
   ```bash
   cd action
   go build -o nobl9-action ./cmd/main.go
   ./nobl9-action --help
   ```
5. **Submit a Pull Request**

## License

This project is licensed under the Mozilla Public License 2.0 - see the [LICENSE](LICENSE) file for details. 