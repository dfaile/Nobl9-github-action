# Nobl9 Project Skeleton

This directory contains example files for Nobl9 projects that can be used as templates.

## Files

- `nobl9-project.yaml` - Example Nobl9 project configuration with Project and RoleBinding definitions
- `README.md` - This file with project information

## How to Use

This skeleton is provided as a reference for the structure of Nobl9 project files. The actual project creation is handled through the Backstage template at `../template.yaml`.

### For Backstage Template Users

1. Use the Backstage template "Create Nobl9 Project" 
2. Fill in the form with your project details
3. The template will automatically trigger the GitHub Action to create your project

### For Manual Creation

1. Copy `nobl9-project.yaml` to your repository
2. Replace the example values with your actual project details
3. Commit the file to your repository
4. The GitHub Action will automatically process it and create the project in Nobl9

## Example

See `../../example-nobl9-project.yaml` for a complete example with real values.

## Management

This project is managed through:
1. **Git Repository** - For version control and audit trail
2. **GitHub Action** - For automatically deploying changes to Nobl9

## Next Steps

1. Access the [Nobl9 Dashboard](https://app.nobl9.com) to view your project
2. Create SLOs and services within this project
3. Monitor project metrics and performance

## Support

For issues or questions about this project:
- Check the GitHub Actions logs for deployment status
- Contact the platform team for assistance
- Review the [Nobl9 Documentation](https://docs.nobl9.com) 