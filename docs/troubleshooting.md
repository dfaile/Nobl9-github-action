# Nobl9 GitHub Action Troubleshooting Guide

This guide helps you diagnose and resolve common issues with the Nobl9 GitHub Action and Backstage template integration.

## Quick Diagnosis

### Check GitHub Actions Status
1. Go to your repository's **Actions** tab
2. Look for the latest workflow run
3. Check the job status and logs
4. Look for error messages in the logs

### Common Error Patterns
- **Authentication failures** - Invalid credentials or expired tokens
- **API timeouts** - Network connectivity or Nobl9 service issues
- **User resolution errors** - Email addresses not found in Nobl9
- **YAML validation errors** - Invalid configuration format
- **Permission errors** - Insufficient access to Nobl9 resources

## Authentication Issues

### Error: "Invalid credentials" or "Authentication failed"

#### Symptoms
```
Error: failed to authenticate with Nobl9
Error: invalid client credentials
Error: authentication token expired
```

#### Causes
1. **Invalid Client ID or Secret**
   - Incorrect credentials in GitHub secrets
   - Credentials copied with extra spaces or characters
   - Wrong environment credentials used

2. **Expired Credentials**
   - Client secret has expired
   - Token refresh failed
   - Credentials rotated but not updated

3. **Wrong Environment**
   - Using production credentials for staging environment
   - Environment mismatch between action and Nobl9

#### Solutions

**1. Verify GitHub Secrets**
```bash
# Check if secrets are properly set
# Go to: Settings > Secrets and variables > Actions
# Verify these secrets exist:
# - NOBL9_CLIENT_ID
# - NOBL9_CLIENT_SECRET
# - NOBL9_ORGANIZATION (if required)
```

**2. Test Credentials Manually**
```bash
# Test with curl (replace with your values)
curl -X POST https://app.nobl9.com/oauth/token \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "grant_type=client_credentials" \
  -d "client_id=YOUR_CLIENT_ID" \
  -d "client_secret=YOUR_CLIENT_SECRET"
```

**3. Regenerate Credentials**
1. Go to Nobl9 Console > Settings > API Keys
2. Create new client credentials
3. Update GitHub secrets with new values
4. Test the action again

**4. Check Environment Configuration**
```yaml
# In your workflow file, ensure environment is correct
env:
  NOBL9_ENVIRONMENT: "production"  # or "staging"
```

### Error: "Insufficient permissions"

#### Symptoms
```
Error: user does not have permission to create projects
Error: insufficient privileges for role binding
Error: access denied to Nobl9 organization
```

#### Solutions

**1. Check Client Permissions**
- Verify the client has organization-level access
- Ensure client can create projects and manage users
- Check if client is restricted to specific projects

**2. Verify Organization Access**
```bash
# Check organization access
# The client should have access to the target organization
```

**3. Update Client Permissions**
1. Go to Nobl9 Console > Settings > API Keys
2. Edit the client permissions
3. Grant necessary permissions:
   - Project creation
   - User management
   - Role binding management

## API Timeout Issues

### Error: "Request timeout" or "Connection timeout"

#### Symptoms
```
Error: context deadline exceeded
Error: request timeout after 30s
Error: connection to Nobl9 API failed
```

#### Causes
1. **Network Connectivity Issues**
   - Firewall blocking outbound connections
   - DNS resolution problems
   - Network latency

2. **Nobl9 Service Issues**
   - Nobl9 API experiencing high load
   - Service maintenance or outages
   - Rate limiting

3. **Configuration Issues**
   - Timeout values too low
   - Retry configuration inadequate

#### Solutions

**1. Check Network Connectivity**
```bash
# Test connectivity to Nobl9
curl -I https://app.nobl9.com
curl -I https://api.nobl9.com

# Check DNS resolution
nslookup app.nobl9.com
nslookup api.nobl9.com
```

**2. Increase Timeout Values**
```yaml
# In your workflow, increase timeout
env:
  NOBL9_TIMEOUT: "60s"  # Increase from default 30s
  NOBL9_RETRY_ATTEMPTS: "5"  # Increase retry attempts
```

**3. Check Nobl9 Status**
- Visit [Nobl9 Status Page](https://status.nobl9.com)
- Check for ongoing maintenance or outages
- Monitor Nobl9 community forums for issues

**4. Implement Exponential Backoff**
```yaml
# The action already implements retry logic
# Check logs for retry attempts and backoff delays
```

## User Resolution Issues

### Error: "User not found" or "Email resolution failed"

#### Symptoms
```
Error: user user@company.com not found in Nobl9
Error: failed to resolve email to User ID
Error: email address not associated with Okta account
```

#### Causes
1. **User Not in Nobl9**
   - User doesn't exist in Nobl9
   - User not provisioned through Okta
   - User account disabled or deleted

2. **Okta Integration Issues**
   - User not synced from Okta
   - Okta integration misconfigured
   - User email mismatch between systems

3. **Email Format Issues**
   - Email address typo
   - Different email format than expected
   - Case sensitivity issues

#### Solutions

**1. Verify User Exists in Nobl9**
1. Go to Nobl9 Console > Users
2. Search for the user by email address
3. Verify user status is "Active"
4. Check user's Okta integration status

**2. Check Okta Integration**
1. Verify Okta integration is enabled
2. Check user sync status in Okta
3. Ensure user is in the correct Okta groups
4. Verify email address matches between systems

**3. Add User to Nobl9**
```bash
# If user doesn't exist, add them through Nobl9 Console
# Or use Nobl9 API to create user
curl -X POST https://app.nobl9.com/api/v1/users \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@company.com",
    "firstName": "John",
    "lastName": "Doe"
  }'
```

**4. Check Email Format**
- Verify email address spelling
- Check for extra spaces or characters
- Ensure consistent email format across systems

### Error: "Duplicate user assignment"

#### Symptoms
```
Error: user already assigned to project
Error: duplicate role binding found
Error: user already has role in project
```

#### Solutions

**1. Check Existing Role Bindings**
1. Go to Nobl9 Console > Projects > [Project Name] > Users
2. Check if user already has a role
3. Remove existing role binding if needed

**2. Update Instead of Create**
- The action should handle updates automatically
- Check if the action is configured for update mode
- Verify YAML configuration is correct

## YAML Validation Issues

### Error: "Invalid YAML format" or "Schema validation failed"

#### Symptoms
```
Error: invalid YAML structure
Error: missing required field 'metadata'
Error: invalid project name format
```

#### Causes
1. **YAML Syntax Errors**
   - Missing quotes around strings
   - Incorrect indentation
   - Invalid characters in values

2. **Schema Validation Errors**
   - Missing required fields
   - Invalid field values
   - Wrong data types

3. **Template Generation Issues**
   - Backstage template errors
   - Variable substitution problems
   - Invalid template syntax

#### Solutions

**1. Validate YAML Syntax**
```bash
# Use online YAML validator
# Or use command line tools
python -c "import yaml; yaml.safe_load(open('nobl9-project.yaml'))"
```

**2. Check Required Fields**
```yaml
# Ensure all required fields are present
apiVersion: n9/v1alpha
kind: Project
metadata:
  name: project-name  # Required
  displayName: "Display Name"  # Required
spec:
  # Project specification
```

**3. Fix Common YAML Issues**
```yaml
# Correct YAML format
apiVersion: n9/v1alpha
kind: Project
metadata:
  name: my-project  # Use hyphens, not underscores
  displayName: "My Project"  # Quotes for spaces
  description: "Project description"  # Optional
spec: {}  # Empty spec is valid
```

**4. Validate with Nobl9 Schema**
```bash
# Use Nobl9 CLI to validate
nobl9 validate nobl9-project.yaml
```

## Deployment Failures

### Error: "Project creation failed" or "Role binding failed"

#### Symptoms
```
Error: failed to create project
Error: project already exists
Error: role binding creation failed
Error: insufficient permissions for role assignment
```

#### Solutions

**1. Check Project Existence**
1. Go to Nobl9 Console > Projects
2. Check if project already exists
3. Use different project name or update existing project

**2. Verify Project Name Format**
```yaml
# Project name must follow DNS RFC1123
metadata:
  name: my-project  # ✅ Valid
  name: my_project  # ❌ Invalid (underscore)
  name: My-Project  # ❌ Invalid (uppercase)
```

**3. Check Role Binding Permissions**
- Ensure client has permission to create role bindings
- Verify user roles are valid for the project
- Check if role binding name conflicts

**4. Handle Existing Resources**
```yaml
# The action should handle updates automatically
# If not, manually update existing resources
# Or delete and recreate if needed
```

## GitHub Actions Issues

### Error: "Workflow failed" or "Job cancelled"

#### Symptoms
```
Error: workflow run failed
Error: job timeout exceeded
Error: insufficient resources
```

#### Solutions

**1. Check Workflow Configuration**
```yaml
# Ensure proper workflow setup
name: Nobl9 Project Deployment
on:
  push:
    paths:
      - 'projects/**/*.yaml'  # Trigger on YAML changes
jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: your-org/nobl9-action@v1
        with:
          client-id: ${{ secrets.NOBL9_CLIENT_ID }}
          client-secret: ${{ secrets.NOBL9_CLIENT_SECRET }}
```

**2. Increase Job Timeout**
```yaml
# Add timeout to job
jobs:
  deploy:
    timeout-minutes: 10  # Increase timeout
    runs-on: ubuntu-latest
```

**3. Check Resource Limits**
- GitHub Actions has resource limits
- Large repositories may timeout
- Consider using self-hosted runners for large deployments

### Error: "Secret not found" or "Invalid secret"

#### Solutions

**1. Verify Secrets Configuration**
1. Go to repository Settings > Secrets and variables > Actions
2. Check if all required secrets are set:
   - `NOBL9_CLIENT_ID`
   - `NOBL9_CLIENT_SECRET`
   - `NOBL9_ORGANIZATION` (if required)

**2. Check Secret Names**
```yaml
# Ensure secret names match exactly
with:
  client-id: ${{ secrets.NOBL9_CLIENT_ID }}  # Exact match
  client-secret: ${{ secrets.NOBL9_CLIENT_SECRET }}  # Exact match
```

**3. Update Secrets**
- Secrets are encrypted and cannot be viewed
- If unsure, delete and recreate secrets
- Use repository-level secrets for sensitive data

## Performance Issues

### Slow Deployment or Timeouts

#### Symptoms
- Deployments taking longer than expected
- Frequent timeout errors
- High resource usage

#### Solutions

**1. Optimize YAML Files**
- Reduce file size by removing unnecessary comments
- Use efficient YAML structure
- Consider splitting large configurations

**2. Implement Caching**
```yaml
# Add caching to workflow
- uses: actions/cache@v3
  with:
    path: ~/.cache/nobl9
    key: ${{ runner.os }}-nobl9-${{ hashFiles('**/nobl9-project.yaml') }}
```

**3. Use Parallel Processing**
```yaml
# Process multiple files in parallel
jobs:
  deploy:
    strategy:
      matrix:
        file: [project1.yaml, project2.yaml, project3.yaml]
```

## Monitoring and Debugging

### Enable Debug Logging

**1. Set Debug Environment Variable**
```yaml
env:
  NOBL9_DEBUG: "true"
  NOBL9_LOG_LEVEL: "debug"
```

**2. Check Detailed Logs**
- Look for detailed error messages
- Check API request/response logs
- Monitor retry attempts and backoff delays

### Common Log Patterns

**Authentication Success**
```
INFO: Nobl9 client created successfully
INFO: Authentication successful
```

**User Resolution Success**
```
INFO: Resolved user user@company.com to UserID: user-123
INFO: User validation successful
```

**Deployment Success**
```
INFO: Project created successfully
INFO: Role binding applied successfully
INFO: Deployment completed
```

**Error Patterns**
```
ERROR: Authentication failed: invalid credentials
ERROR: User not found: user@company.com
ERROR: API timeout: request exceeded 30s
ERROR: YAML validation failed: invalid format
```

## Getting Help

### 1. Check Documentation
- [Nobl9 Documentation](https://docs.nobl9.com)
- [GitHub Actions Documentation](https://docs.github.com/en/actions)
- [Backstage Documentation](https://backstage.io/docs)

### 2. Review Logs
- GitHub Actions logs provide detailed error information
- Check both workflow and step-level logs
- Look for specific error codes and messages

### 3. Contact Support
- **Platform Team:** For template and action issues
- **Nobl9 Support:** For Nobl9 platform issues
- **GitHub Support:** For GitHub Actions issues

### 4. Create Issues
- Use the repository's issue tracker
- Include relevant logs and error messages
- Provide steps to reproduce the issue

## Prevention Best Practices

### 1. Regular Maintenance
- Rotate credentials regularly
- Monitor Nobl9 API usage
- Update action versions periodically

### 2. Testing
- Test changes in staging environment first
- Use dry-run mode for validation
- Implement automated testing

### 3. Monitoring
- Set up alerts for deployment failures
- Monitor Nobl9 API response times
- Track user resolution success rates

### 4. Documentation
- Keep runbooks updated
- Document common issues and solutions
- Maintain troubleshooting procedures

---

*This troubleshooting guide is maintained by the platform team. For additional help, contact the platform team or create an issue in the repository.* 