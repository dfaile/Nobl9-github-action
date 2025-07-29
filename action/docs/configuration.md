# Configuration Management

This document describes how to configure the Nobl9 GitHub Action for different environments and use cases.

## Overview

The action uses a comprehensive configuration management system that supports:
- **GitHub Actions inputs** - Primary configuration method for CI/CD
- **Environment variables** - Fallback for local development
- **Auto-detection** - Automatic environment detection from credentials
- **Validation** - Comprehensive input validation with clear error messages

## Configuration Sources

The action loads configuration in the following order:

1. **GitHub Actions Inputs** (primary)
   - `INPUT_CLIENT_ID`
   - `INPUT_CLIENT_SECRET`
   - `INPUT_REPO_PATH`
   - `INPUT_FILE_PATTERN`
   - `INPUT_DRY_RUN`
   - `INPUT_FORCE`
   - `INPUT_VALIDATE_ONLY`
   - `INPUT_LOG_LEVEL`
   - `INPUT_LOG_FORMAT`

2. **Direct Environment Variables** (fallback for local development)
   - `NOBL9_CLIENT_ID`
   - `NOBL9_CLIENT_SECRET`

3. **GitHub Actions Context** (automatic)
   - `GITHUB_WORKSPACE`
   - `GITHUB_EVENT_PATH`
   - `GITHUB_TOKEN`
   - `GITHUB_ACTIONS`

## Required Configuration

### Nobl9 API Credentials

**Required for all operations:**

```yaml
# GitHub Actions workflow
- name: Sync Nobl9 Projects
  uses: ./
  with:
    client-id: ${{ secrets.NOBL9_CLIENT_ID }}
    client-secret: ${{ secrets.NOBL9_CLIENT_SECRET }}
```

**Local Development:**
```bash
export NOBL9_CLIENT_ID="your-client-id"
export NOBL9_CLIENT_SECRET="your-client-secret"
```

### GitHub Secrets Setup

1. **Create Nobl9 API Credentials:**
   - Go to your Nobl9 organization settings
   - Navigate to API Keys section
   - Create a new API key with appropriate permissions
   - Copy the Client ID and Client Secret

2. **Add to GitHub Repository Secrets:**
   - Go to your GitHub repository
   - Navigate to Settings → Secrets and variables → Actions
   - Add the following secrets:
     - `NOBL9_CLIENT_ID`: Your Nobl9 Client ID
     - `NOBL9_CLIENT_SECRET`: Your Nobl9 Client Secret

## Optional Configuration

### Repository Configuration

```yaml
# Default values
repo-path: "."                    # Repository path to scan
file-pattern: "**/*.yaml"        # Glob pattern for YAML files
```

**Examples:**
```yaml
# Scan specific directory
repo-path: "./nobl9-configs"

# Scan multiple file patterns
file-pattern: "**/nobl9-*.yaml"

# Scan specific file types
file-pattern: "**/*.{yaml,yml}"
```

### Processing Options

```yaml
# Default values
dry-run: false                   # Perform dry run without changes
force: false                     # Force processing despite validation errors
validate-only: false             # Only validate, don't deploy
```

**Use Cases:**
```yaml
# Development/testing
dry-run: true

# Override validation errors
force: true

# CI/CD validation step
validate-only: true
```

### Logging Configuration

```yaml
# Default values
log-level: "info"                # Log level (debug, info, warn, error)
log-format: "json"               # Log format (json, text)
```

**Examples:**
```yaml
# Verbose debugging
log-level: "debug"
log-format: "text"

# Production logging
log-level: "info"
log-format: "json"
```

## Environment Detection

The action automatically detects the Nobl9 environment from your credentials:

| Client ID Pattern | Detected Environment |
|------------------|---------------------|
| Contains "dev" | `dev` |
| Contains "staging" | `staging` |
| Contains "prod" | `prod` |
| Other | `unknown` |

**Example:**
```bash
# Development environment
NOBL9_CLIENT_ID="dev-api-client-123"

# Production environment  
NOBL9_CLIENT_ID="prod-api-client-456"
```

## Configuration Validation

The action validates all configuration before processing:

### Required Fields
- ✅ Nobl9 Client ID
- ✅ Nobl9 Client Secret
- ✅ GitHub Workspace (automatically set in GitHub Actions)

### Optional Field Validation
- ✅ Log level must be: `debug`, `info`, `warn`, `error`
- ✅ Log format must be: `json`, `text`
- ✅ Boolean values must be: `true`, `false`, `1`, `0`, `yes`, `no`, `on`, `off`

### Error Messages

The action provides clear error messages for configuration issues:

```
❌ Configuration validation failed: Nobl9 client ID is required
❌ Configuration validation failed: invalid log level: invalid (valid: [debug info warn error])
❌ Configuration validation failed: invalid boolean value: maybe
```

## Security Best Practices

### Credential Management
- ✅ **Never commit credentials** to version control
- ✅ **Use GitHub Secrets** for sensitive data
- ✅ **Rotate credentials** regularly
- ✅ **Use least privilege** API keys

### Environment Separation
- ✅ **Separate credentials** for different environments
- ✅ **Use environment-specific** client IDs
- ✅ **Validate environment** detection
- ✅ **Test in staging** before production

### Access Control
- ✅ **Limit repository access** to authorized users
- ✅ **Use branch protection** rules
- ✅ **Require PR reviews** for configuration changes
- ✅ **Monitor action usage** and logs

## Troubleshooting

### Common Issues

**1. Missing Credentials**
```
Error: Nobl9 client ID is required
```
**Solution:** Ensure `NOBL9_CLIENT_ID` and `NOBL9_CLIENT_SECRET` are set in GitHub Secrets.

**2. Invalid Log Level**
```
Error: invalid log level: debug (valid: [debug info warn error])
```
**Solution:** Use one of the supported log levels: `debug`, `info`, `warn`, `error`.

**3. Missing GitHub Workspace**
```
Error: GitHub workspace is required
```
**Solution:** This is automatically set in GitHub Actions. Check if running in the correct context.

**4. Environment Detection Issues**
```
Warning: Unable to detect Nobl9 environment
```
**Solution:** Ensure your client ID contains environment indicators (dev, staging, prod).

### Debug Mode

Enable debug logging for troubleshooting:

```yaml
- name: Debug Nobl9 Action
  uses: ./
  with:
    client-id: ${{ secrets.NOBL9_CLIENT_ID }}
    client-secret: ${{ secrets.NOBL9_CLIENT_SECRET }}
    log-level: "debug"
    log-format: "text"
    dry-run: true
```

### Local Testing

Test configuration locally:

```bash
# Set environment variables
export NOBL9_CLIENT_ID="your-client-id"
export NOBL9_CLIENT_SECRET="your-client-secret"
export GITHUB_WORKSPACE="/path/to/repo"

# Run with debug logging
./nobl9-action process --log-level debug --dry-run
```

## Examples

### Basic Configuration
```yaml
- name: Sync Nobl9 Projects
  uses: ./
  with:
    client-id: ${{ secrets.NOBL9_CLIENT_ID }}
    client-secret: ${{ secrets.NOBL9_CLIENT_SECRET }}
```

### Advanced Configuration
```yaml
- name: Advanced Nobl9 Sync
  uses: ./
  with:
    client-id: ${{ secrets.NOBL9_CLIENT_ID }}
    client-secret: ${{ secrets.NOBL9_CLIENT_SECRET }}
    repo-path: "./nobl9-configs"
    file-pattern: "**/nobl9-*.yaml"
    dry-run: false
    force: false
    validate-only: false
    log-level: "info"
    log-format: "json"
```

### Validation Workflow
```yaml
- name: Validate Nobl9 Configs
  uses: ./
  with:
    client-id: ${{ secrets.NOBL9_CLIENT_ID }}
    client-secret: ${{ secrets.NOBL9_CLIENT_SECRET }}
    validate-only: true
    log-level: "info"
```

### Development Workflow
```yaml
- name: Test Nobl9 Changes
  uses: ./
  with:
    client-id: ${{ secrets.NOBL9_DEV_CLIENT_ID }}
    client-secret: ${{ secrets.NOBL9_DEV_CLIENT_SECRET }}
    dry-run: true
    log-level: "debug"
    log-format: "text"
``` 