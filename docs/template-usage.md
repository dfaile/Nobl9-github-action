# Nobl9 Backstage Template Usage Guide

This guide explains how to use the Nobl9 Backstage template to create and manage Nobl9 projects through a self-service interface.

## Overview

The Nobl9 Backstage template (`nobl9-project-template`) enables teams to create Nobl9 projects with proper user role assignments through a simple form interface. The template integrates with GitHub Actions to automatically deploy configurations to Nobl9 using GitOps principles.

## Template Features

- **Self-service project creation** - Teams can create projects without manual intervention
- **Automated role management** - User roles are automatically assigned and managed
- **Email-to-UserID resolution** - Email addresses are automatically resolved to Okta User IDs
- **GitOps workflow** - All changes are version controlled and auditable
- **Validation and error handling** - Comprehensive validation ensures data quality

## Prerequisites

Before using the template, ensure you have:

1. **Backstage access** - Access to the Backstage instance where the template is installed
2. **Nobl9 account** - A valid Nobl9 account with Okta integration enabled
3. **GitHub access** - Access to the target GitHub repository for storing configurations
4. **User permissions** - Users must exist in Nobl9 with Okta integration

## Form Fields and Validation

### Project Information Section

#### Project Name
- **Type:** Text field (required)
- **Description:** Unique name for the Nobl9 project
- **Validation Rules:**
  - Must follow DNS RFC1123 standards
  - Only lowercase letters, numbers, and hyphens allowed
  - Cannot start or end with a hyphen
  - Length: 1-63 characters
  - Pattern: `^[a-z0-9]([a-z0-9-]*[a-z0-9])?$`
- **Examples:**
  - ✅ `my-project`
  - ✅ `project-123`
  - ✅ `monitoring-prod`
  - ❌ `My-Project` (uppercase not allowed)
  - ❌ `-project-` (cannot start/end with hyphen)
  - ❌ `project_123` (underscore not allowed)

#### Display Name
- **Type:** Text field (required)
- **Description:** Human-readable name shown in Nobl9 dashboard
- **Validation Rules:**
  - Length: 1-63 characters
  - Can contain spaces and special characters
- **Examples:**
  - ✅ `My Project Name`
  - ✅ `Production Monitoring`
  - ✅ `Team Alpha - SLOs`

#### Description
- **Type:** Multi-line text field (optional)
- **Description:** Optional description of the project's purpose
- **Validation Rules:**
  - No length restrictions
  - Supports markdown formatting
- **Examples:**
  - `Monitoring and SLO management for production services`
  - `Team Alpha's service level objectives and monitoring`

### User Management Section

#### Project Users
- **Type:** Dynamic array of user objects (required)
- **Description:** Add users and assign roles to this project
- **Validation Rules:**
  - At least one user required
  - At least one project owner required
  - No duplicate email addresses allowed
  - All email addresses must be valid Nobl9 users

#### User Email Address
- **Type:** Text field (required per user)
- **Description:** User's email address (must be a valid Nobl9 user with Okta integration)
- **Validation Rules:**
  - Must be a valid email format
  - Pattern: `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
  - User must exist in Nobl9 with Okta integration
- **Examples:**
  - ✅ `user@company.com`
  - ✅ `john.doe@example.org`
  - ❌ `invalid-email`
  - ❌ `user@` (incomplete email)

#### User Role
- **Type:** Dropdown selection (required per user)
- **Description:** User's role in the project
- **Options:**
  - **Project Owner (Full Access):** `project-owner`
    - Full administrative access to the project
    - Can manage all aspects including users, SLOs, services
    - Can delete the project
    - Required: At least one project owner per project
  - **Project Editor (Manage SLOs & Services):** `project-editor`
    - Can create and manage SLOs and services
    - Can view project metrics and dashboards
    - Cannot manage users or project settings
    - Cannot delete the project

## Step-by-Step Usage

### 1. Access the Template

1. Navigate to your Backstage instance
2. Go to the **Create** page
3. Find the **"Create Nobl9 Project"** template
4. Click **Choose** to start

### 2. Fill Out the Form

#### Project Information
1. **Project Name:** Enter a unique, DNS-compliant name
   - Use lowercase letters, numbers, and hyphens only
   - Keep it descriptive but concise
   - Example: `team-alpha-monitoring`

2. **Display Name:** Enter a human-readable name
   - This will appear in the Nobl9 dashboard
   - Can include spaces and special characters
   - Example: `Team Alpha - Production Monitoring`

3. **Description:** (Optional) Describe the project's purpose
   - Helps with project identification
   - Supports markdown formatting
   - Example: `Monitoring and SLO management for Team Alpha's production services`

#### User Management
1. **Add Users:** Click **"+ Add User"** for each user
2. **Email Address:** Enter the user's email address
   - Must be a valid Nobl9 user with Okta integration
   - The system will validate the email format
3. **Role Assignment:** Select the appropriate role
   - Ensure at least one user has **Project Owner** role
   - Assign **Project Editor** role to team members who need SLO management access

### 3. Validation and Review

The template performs several validation checks:

1. **Project Name Validation:**
   - Checks DNS RFC1123 compliance
   - Ensures no invalid characters
   - Validates length requirements

2. **User Validation:**
   - Verifies at least one project owner is specified
   - Checks for duplicate email addresses
   - Validates email format for all users

3. **Form Validation:**
   - Ensures all required fields are completed
   - Validates field formats and patterns
   - Provides real-time feedback

### 4. Template Execution

When you click **Create**, the template executes the following steps:

1. **Validate User Input** - Performs comprehensive validation
2. **Validate Project Owner Requirement** - Ensures at least one owner exists
3. **Validate Email Duplicates** - Checks for duplicate email addresses
4. **Generate Nobl9 YAML** - Creates the configuration files
5. **Create Project Directory** - Sets up the repository structure
6. **Publish to GitHub** - Commits files to the repository
7. **Register in Catalog** - Adds the project to Backstage catalog

### 5. Generated Files

The template creates the following files in the repository:

#### `nobl9-project.yaml`
- Main Nobl9 configuration file
- Contains Project and RoleBinding definitions
- Includes metadata and labels for tracking

#### `catalog-info.yaml`
- Backstage catalog information
- Provides integration with Backstage features
- Includes links to related resources

#### `README.md`
- Project documentation
- Lists users and their roles
- Provides management information

## Generated Configuration Structure

### Project Definition
```yaml
apiVersion: n9/v1alpha
kind: Project
metadata:
  name: your-project-name
  displayName: Your Project Display Name
  description: Project description
  labels:
    source: backstage-template
    generated-by: nobl9-backstage-action
    created-date: 2024-01-01
    template-version: "1.0.0"
    project-type: "monitoring"
spec:
  # Project specification
```

### Role Binding Definition
```yaml
apiVersion: n9/v1alpha
kind: RoleBinding
metadata:
  name: your-project-name-role-bindings
  project: your-project-name
  displayName: "Role Bindings for Your Project Display Name"
spec:
  users:
    - email: user1@company.com
      roles:
        - project-owner
    - email: user2@company.com
      roles:
        - project-editor
```

## GitHub Action Integration

After the template creates the files, the GitHub Action automatically:

1. **Detects Changes** - Monitors the repository for new YAML files
2. **Validates Configuration** - Checks YAML structure and content
3. **Resolves Users** - Converts email addresses to Okta User IDs
4. **Creates Project** - Creates the project in Nobl9
5. **Applies Role Bindings** - Assigns users to their specified roles
6. **Provides Logs** - Generates detailed deployment logs

## Monitoring and Troubleshooting

### Check Deployment Status
1. Go to the GitHub repository
2. Navigate to the **Actions** tab
3. Look for the latest workflow run
4. Review the logs for any errors or warnings

### Common Issues

#### Project Name Validation Errors
- **Issue:** Project name contains invalid characters
- **Solution:** Use only lowercase letters, numbers, and hyphens
- **Example:** Change `My_Project` to `my-project`

#### User Validation Errors
- **Issue:** User email not found in Nobl9
- **Solution:** Ensure the user exists in Nobl9 with Okta integration
- **Action:** Contact your Nobl9 administrator to add the user

#### Duplicate Email Errors
- **Issue:** Same email address used multiple times
- **Solution:** Remove duplicate entries or use different email addresses

#### Missing Project Owner
- **Issue:** No project owner specified
- **Solution:** Ensure at least one user has the `project-owner` role

## Best Practices

### Project Naming
- Use descriptive but concise names
- Follow team naming conventions
- Include environment indicators when appropriate
- Examples: `team-alpha-prod`, `payment-service-monitoring`

### User Management
- Assign project owner role to team leads or administrators
- Use project editor role for team members who need SLO management
- Regularly review and update user assignments
- Document role assignments for audit purposes

### Description Writing
- Be specific about the project's purpose
- Include relevant context and scope
- Mention key services or systems being monitored
- Update descriptions when project scope changes

## Security Considerations

### Access Control
- Only assign necessary roles to users
- Regularly review user access and permissions
- Remove access for users who no longer need it
- Use the principle of least privilege

### Audit Trail
- All changes are version controlled in Git
- GitHub Actions provide detailed deployment logs
- Backstage catalog maintains project history
- Monitor access and changes regularly

## Support and Resources

### Documentation
- [Nobl9 Documentation](https://docs.nobl9.com)
- [Backstage Documentation](https://backstage.io/docs)
- [GitHub Actions Documentation](https://docs.github.com/en/actions)

### Getting Help
- Check GitHub Actions logs for deployment issues
- Contact your platform team for template-related questions
- Reach out to Nobl9 support for platform-specific issues
- Review the troubleshooting guide for common problems

### Template Updates
- Template versions are tracked in generated files
- Updates are deployed through the platform team
- Check release notes for new features and improvements
- Provide feedback to improve the template

## Template Version History

### Version 1.0.0
- Initial release
- Basic project and role binding creation
- Email-to-UserID resolution
- GitHub Action integration
- Backstage catalog integration

---

*This documentation is maintained by the platform team. For questions or suggestions, please contact the platform team or create an issue in the template repository.* 