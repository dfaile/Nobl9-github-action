# Nobl9 GitHub Action Setup Guide

This guide explains how to set up and configure the Nobl9 GitHub Action for automated project deployment and management.

## Overview

The Nobl9 GitHub Action automates the deployment of Nobl9 project configurations created through the Backstage template. It processes YAML files, resolves user email addresses to Okta User IDs, and creates/updates projects and role bindings in Nobl9.

## Prerequisites

Before setting up the action, ensure you have:

1. **Nobl9 Account** - Active Nobl9 account with API access
2. **Okta Integration** - Nobl9 configured with Okta for user management
3. **GitHub Repository** - Repository to store Nobl9 configurations
4. **GitHub Actions Access** - Ability to create workflows and secrets
5. **API Credentials** - Nobl9 client ID and client secret with appropriate permissions

## Required Secrets

The action requires the following GitHub secrets to authenticate with Nobl9 and Docker Hub:

### Core Secrets

#### `NOBL9_CLIENT_ID`
- **Description:** Your Nobl9 API client ID
- **Type:** String
- **Required:** Yes
- **Example:** `nobl9-client-12345`
- **How to obtain:**
  1. Go to Nobl9 Console > Settings > Access Keys
  2. Create a new API key or use existing one
  3. Copy the Client ID

#### `NOBL9_CLIENT_SECRET`
- **Description:** Your Nobl9 API client secret
- **Type:** String
- **Required:** Yes
- **Example:** `nobl9-secret-abcdef123456`
- **How to obtain:**
  1. Go to Nobl9 Console > Settings > Access Keys
  2. Create a new Access key or use existing one
  3. Copy the Client Secret
  4. **Important:** Store securely and never commit to repository

### Optional Secrets

#### `NOBL9_ORGANIZATION`
- **Description:** Nobl9 organization identifier (if using multi-org setup)
- **Type:** String
- **Required:** No (defaults to primary organization)
- **Example:** `my-company`
- **How to obtain:**
  1. Check Nobl9 Console URL: `https://app.nobl9.com/org/my-company`
  2. Use the organization identifier from the URL

#### `NOBL9_ENVIRONMENT`
- **Description:** Nobl9 environment (production/staging)
- **Type:** String
- **Required:** No (defaults to production)
- **Values:** `production`, `staging`
- **Example:** `production`

## Setting Up GitHub Secrets

### Method 1: Repository Secrets (Recommended)

1. **Navigate to Repository Settings**
   ```
   Repository > Settings > Secrets and variables > Actions
   ```

2. **Add Required Secrets**
   - Click **"New repository secret"**
   - Add each secret with exact names:
     - `NOBL9_CLIENT_ID`
     - `NOBL9_CLIENT_SECRET`
     - `NOBL9_ORGANIZATION` (if needed)
     - `NOBL9_ENVIRONMENT` (if needed)

3. **Verify Secrets**
   - Secrets are encrypted and cannot be viewed
   - Ensure names match exactly (case-sensitive)
   - Test with a simple workflow

### Method 2: Organization Secrets

For multiple repositories using the same Nobl9 credentials:

1. **Navigate to Organization Settings**
   ```
   Organization > Settings > Secrets and variables > Actions
   ```

2. **Add Organization Secrets**
   - Same process as repository secrets
   - Available to all repositories in the organization
   - Override with repository secrets if needed

### Method 3: Environment Secrets

For different environments (staging/production):

1. **Create Environments**
   ```
   Repository > Settings > Environments
   ```

2. **Add Environment-Specific Secrets**
   - Create `staging` and `production` environments
   - Add environment-specific secrets
   - Use environment protection rules if needed

### Docker Hub Secrets (for Releases)

#### `DOCKERHUB_USERNAME`
- **Description:** Your Docker Hub username
- **Type:** String
- **Required:** For releases only
- **Example:** `your-dockerhub-username`
- **How to obtain:**
  1. Create a Docker Hub account at [hub.docker.com](https://hub.docker.com/)
  2. Your username is displayed in your profile

#### `DOCKERHUB_TOKEN`
- **Description:** Docker Hub access token for authentication
- **Type:** String (sensitive)
- **Required:** For releases only
- **Example:** `dckr_pat_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx`
- **How to obtain:**
  1. Go to [Docker Hub](https://hub.docker.com/) → Account Settings → Security
  2. Click "New Access Token"
  3. Give it a name (e.g., "GitHub Actions")
  4. Select "Read, Write, Delete" permissions
  5. Copy the token and add it as a GitHub secret

## Workflow Configuration

### Basic Workflow Setup

Create `.github/workflows/nobl9-deploy.yml`:

```yaml
name: Nobl9 Project Deployment

on:
  push:
    branches: [main]
    paths:
      - 'projects/**/*.yaml'
      - 'projects/**/*.yml'
  pull_request:
    branches: [main]
    paths:
      - 'projects/**/*.yaml'
      - 'projects/**/*.yml'

jobs:
  deploy:
    runs-on: ubuntu-latest
    
    steps:
      - name: Checkout code
        uses: actions/checkout@v4
        
      - name: Deploy to Nobl9
        uses: docker://docker.io/your-dockerhub-username/nobl9-github-action:latest
        with:
          client-id: ${{ secrets.NOBL9_CLIENT_ID }}
          client-secret: ${{ secrets.NOBL9_CLIENT_SECRET }}
          organization: ${{ secrets.NOBL9_ORGANIZATION }}
          environment: ${{ secrets.NOBL9_ENVIRONMENT }}
```

### Advanced Workflow Configuration

```yaml
name: Nobl9 Project Deployment

on:
  push:
    branches: [main, develop]
    paths:
      - 'projects/**/*.yaml'
      - 'projects/**/*.yml'
  pull_request:
    branches: [main]
    paths:
      - 'projects/**/*.yaml'
      - 'projects/**/*.yml'
  workflow_dispatch:
    inputs:
      dry_run:
        description: 'Run in dry-run mode'
        required: false
        default: 'false'
        type: boolean

jobs:
  deploy:
    runs-on: ubuntu-latest
    timeout-minutes: 10
    
    strategy:
      matrix:
        environment: [production, staging]
      fail-fast: false
    
    steps:
      - name: Checkout code
        uses: actions/checkout@v4
        
      - name: Deploy to Nobl9
        uses: docker://docker.io/your-dockerhub-username/nobl9-github-action:latest
        with:
          client-id: ${{ secrets.NOBL9_CLIENT_ID }}
          client-secret: ${{ secrets.NOBL9_CLIENT_SECRET }}
          organization: ${{ secrets.NOBL9_ORGANIZATION }}
          environment: ${{ matrix.environment }}
          dry-run: ${{ github.event.inputs.dry_run || 'false' }}
          timeout: '60s'
          retry-attempts: '5'
        env:
          NOBL9_DEBUG: 'true'
          NOBL9_LOG_LEVEL: 'info'
```

## Action Inputs

### Required Inputs

#### `client-id`
- **Description:** Nobl9 API client ID
- **Type:** String
- **Required:** Yes
- **Example:** `${{ secrets.NOBL9_CLIENT_ID }}`

#### `client-secret`
- **Description:** Nobl9 API client secret
- **Type:** String
- **Required:** Yes
- **Example:** `${{ secrets.NOBL9_CLIENT_SECRET }}`

### Optional Inputs

#### `organization`
- **Description:** Nobl9 organization identifier
- **Type:** String
- **Required:** No
- **Default:** Primary organization
- **Example:** `${{ secrets.NOBL9_ORGANIZATION }}`

#### `environment`
- **Description:** Nobl9 environment
- **Type:** String
- **Required:** No
- **Default:** `production`
- **Values:** `production`, `staging`
- **Example:** `production`

#### `dry-run`
- **Description:** Run in dry-run mode (validate only)
- **Type:** Boolean
- **Required:** No
- **Default:** `false`
- **Example:** `true`

#### `timeout`
- **Description:** API request timeout
- **Type:** String
- **Required:** No
- **Default:** `30s`
- **Example:** `60s`

#### `retry-attempts`
- **Description:** Number of retry attempts for failed requests
- **Type:** String
- **Required:** No
- **Default:** `3`
- **Example:** `5`

#### `log-level`
- **Description:** Logging level
- **Type:** String
- **Required:** No
- **Default:** `info`
- **Values:** `debug`, `info`, `warn`, `error`
- **Example:** `debug`

## Environment Variables

### Debug and Logging

#### `NOBL9_DEBUG`
- **Description:** Enable debug mode
- **Type:** Boolean
- **Default:** `false`
- **Example:** `true`

#### `NOBL9_LOG_LEVEL`
- **Description:** Set logging level
- **Type:** String
- **Default:** `info`
- **Values:** `debug`, `info`, `warn`, `error`
- **Example:** `debug`

#### `NOBL9_LOG_FORMAT`
- **Description:** Log output format
- **Type:** String
- **Default:** `json`
- **Values:** `json`, `text`
- **Example:** `json`

### Performance and Reliability

#### `NOBL9_TIMEOUT`
- **Description:** API request timeout
- **Type:** String
- **Default:** `30s`
- **Example:** `60s`

#### `NOBL9_RETRY_ATTEMPTS`
- **Description:** Number of retry attempts
- **Type:** String
- **Default:** `3`
- **Example:** `5`

#### `NOBL9_RETRY_DELAY`
- **Description:** Initial retry delay
- **Type:** String
- **Default:** `1s`
- **Example:** `2s`

## Repository Structure

### Recommended Directory Structure

```
repository/
├── .github/
│   └── workflows/
│       └── nobl9-deploy.yml
├── projects/
│   ├── project-1/
│   │   ├── nobl9-project.yaml
│   │   ├── catalog-info.yaml
│   │   └── README.md
│   └── project-2/
│       ├── nobl9-project.yaml
│       ├── catalog-info.yaml
│       └── README.md
├── docs/
│   ├── template-usage.md
│   ├── action-setup.md
│   └── troubleshooting.md
└── README.md
```

### File Naming Conventions

- **YAML files:** `nobl9-project.yaml`, `*.yaml`, `*.yml`
- **Project directories:** Use DNS-compliant names (lowercase, hyphens)
- **Configuration files:** Follow Nobl9 naming conventions

## Security Considerations

### Secret Management

1. **Never commit secrets to repository**
   - Use GitHub secrets for all sensitive data
   - Rotate secrets regularly
   - Use different secrets for different environments

2. **Principle of least privilege**
   - Grant minimum required permissions to API clients
   - Use environment-specific credentials
   - Regularly review and audit access

3. **Secret rotation**
   - Rotate client secrets every 90 days
   - Update GitHub secrets immediately
   - Test new credentials before removing old ones

### Access Control

1. **Repository permissions**
   - Limit who can modify workflows
   - Use branch protection rules
   - Require pull request reviews

2. **Environment protection**
   - Use environment protection rules for production
   - Require manual approval for critical deployments
   - Implement deployment gates

## Testing and Validation

### Dry-Run Mode

Test configurations without making changes:

```yaml
- name: Validate Nobl9 Configuration
  uses: docker://docker.io/your-dockerhub-username/nobl9-github-action:latest
  with:
    client-id: ${{ secrets.NOBL9_CLIENT_ID }}
    client-secret: ${{ secrets.NOBL9_CLIENT_SECRET }}
    dry-run: 'true'
```

### Local Testing

Test configurations locally before pushing:

```bash
# Install Nobl9 CLI
npm install -g @nobl9/cli

# Validate YAML files
nobl9 validate projects/project-1/nobl9-project.yaml

# Test API connectivity
nobl9 projects list --client-id YOUR_CLIENT_ID --client-secret YOUR_CLIENT_SECRET
```

### Staging Environment

Test in staging before production:

```yaml
jobs:
  test-staging:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: docker://docker.io/your-dockerhub-username/nobl9-github-action:latest
        with:
          client-id: ${{ secrets.NOBL9_STAGING_CLIENT_ID }}
          client-secret: ${{ secrets.NOBL9_STAGING_CLIENT_SECRET }}
          environment: 'staging'
          dry-run: 'false'
```

## Monitoring and Observability

### GitHub Actions Monitoring

1. **Workflow status**
   - Monitor workflow runs in Actions tab
   - Set up notifications for failures
   - Track deployment metrics

2. **Log analysis**
   - Review logs for errors and warnings
   - Monitor performance metrics
   - Track user resolution success rates

### Nobl9 Monitoring

1. **Project status**
   - Monitor project creation success rates
   - Track role binding applications
   - Verify user assignments

2. **API usage**
   - Monitor API rate limits
   - Track authentication success rates
   - Monitor response times

## Troubleshooting

### Common Setup Issues

1. **Secret not found**
   - Verify secret names match exactly
   - Check repository vs. organization secrets
   - Ensure secrets are not empty

2. **Authentication failures**
   - Verify client credentials are correct
   - Check Nobl9 API key permissions
   - Ensure organization access

3. **Permission errors**
   - Verify client has required permissions
   - Check organization membership
   - Ensure project creation rights

### Debug Mode

Enable debug logging for troubleshooting:

```yaml
env:
  NOBL9_DEBUG: 'true'
  NOBL9_LOG_LEVEL: 'debug'
  NOBL9_LOG_FORMAT: 'text'
```

## Best Practices

### Workflow Design

1. **Trigger optimization**
   - Use path filters to trigger only on relevant changes
   - Avoid triggering on documentation changes
   - Use pull request workflows for validation

2. **Error handling**
   - Implement proper error handling and retries
   - Use conditional steps for different scenarios
   - Provide clear error messages

3. **Performance**
   - Use caching for dependencies
   - Optimize workflow execution time
   - Implement parallel processing where possible

### Security

1. **Secret management**
   - Rotate secrets regularly
   - Use environment-specific secrets
   - Implement secret scanning

2. **Access control**
   - Use least privilege principle
   - Implement proper approval workflows
   - Monitor access and changes

### Maintenance

1. **Regular updates**
   - Keep action version updated
   - Monitor for security updates
   - Test new versions in staging

2. **Documentation**
   - Keep setup documentation current
   - Document environment-specific configurations
   - Maintain troubleshooting guides

## Migration Guide

### From Manual Deployment

1. **Backup existing configurations**
   - Export current Nobl9 configurations
   - Document existing projects and users
   - Create backup of current state

2. **Set up GitHub repository**
   - Create repository for configurations
   - Set up directory structure
   - Configure secrets and workflows

3. **Migrate configurations**
   - Convert existing configurations to YAML
   - Test with dry-run mode
   - Deploy incrementally

### From Other CI/CD Systems

1. **Workflow conversion**
   - Convert existing deployment scripts
   - Adapt to GitHub Actions syntax
   - Test thoroughly before switching

2. **Secret migration**
   - Transfer secrets to GitHub
   - Update workflow references
   - Verify access and permissions

## Support and Resources

### Documentation
- [Nobl9 Documentation](https://docs.nobl9.com)
- [GitHub Actions Documentation](https://docs.github.com/en/actions)
- [Backstage Documentation](https://backstage.io/docs)

### Getting Help
- Check the [troubleshooting guide](troubleshooting.md)
- Review GitHub Actions logs for errors
- Contact the platform team for assistance
- Create issues in the action repository

### Community
- Nobl9 Community Forum
- GitHub Discussions
- Backstage Community

---

*This setup guide is maintained by the platform team. For questions or suggestions, please contact the platform team or create an issue in the repository.* 