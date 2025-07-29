# Nobl9 Project: ${{ values.displayName }}

This Nobl9 project was created using the Backstage template and is managed through GitOps.

## Project Information

- **Project Name:** `${{ values.projectName }}`
- **Display Name:** ${{ values.displayName }}
- **Description:** ${{ values.description or "Nobl9 project for monitoring and SLO management" }}
- **Created:** ${{ now | date("2006-01-02 15:04:05") }}

## Users and Roles

| Email | Role |
|-------|------|
{% for user in values.users %}| ${{ user.email }} | ${{ user.role }} |
{% endfor %}

## Configuration Files

- `nobl9-project.yaml` - Main Nobl9 configuration with Project and RoleBinding definitions
- `catalog-info.yaml` - Backstage catalog information

## Management

This project is managed through:
1. **Backstage Template** - For creating and updating project configurations
2. **GitHub Action** - For automatically deploying changes to Nobl9
3. **Git Repository** - For version control and audit trail

## Next Steps

1. Access the [Nobl9 Dashboard](https://app.nobl9.com) to view your project
2. Create SLOs and services within this project
3. Monitor project metrics and performance

## Support

For issues or questions about this project:
- Check the GitHub Actions logs for deployment status
- Contact the platform team for assistance
- Review the [Nobl9 Documentation](https://docs.nobl9.com) 