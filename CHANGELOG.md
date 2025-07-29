# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Initial project setup with Nobl9 GitHub Action
- Backstage template for Nobl9 project creation
- Comprehensive Go packages for Nobl9 integration
- Automated testing and CI/CD workflows
- Security scanning and dependency management
- Documentation and examples

### Changed
- N/A

### Deprecated
- N/A

### Removed
- N/A

### Fixed
- N/A

### Security
- N/A

---

## [v1.0.0] - 2024-01-XX

### Added
- **Core Nobl9 Integration**
  - Nobl9 client package with authentication and API interaction
  - User resolution system for email to UserID mapping
  - YAML parser for Nobl9 project configuration files
  - File scanner for recursive YAML file discovery
  - Processor for orchestrating Nobl9 project creation
  - Validator for configuration validation and error checking
  - Retry mechanism with exponential backoff
  - Structured logging with configurable levels

- **Backstage Template**
  - Complete Backstage template for Nobl9 project creation
  - Form-based project configuration with validation
  - Automated YAML generation for Nobl9 projects
  - Backstage catalog integration
  - Customizable project metadata and annotations
  - Environment-specific configuration support

- **GitHub Action**
  - Multi-platform GitHub Action (Linux, macOS, Windows)
  - Docker containerization for consistent execution
  - Dry-run mode for safe testing
  - Comprehensive error handling and reporting
  - Configurable logging and verbosity levels
  - Support for multiple input formats and sources

- **Testing and Quality Assurance**
  - Comprehensive unit tests for all Go packages
  - Integration tests for end-to-end workflows
  - Backstage template testing in real environments
  - Performance benchmarking and monitoring
  - Security scanning with multiple tools
  - Code quality checks and linting

- **CI/CD and Automation**
  - Automated testing workflows for Go code and templates
  - Security scanning with CodeQL, Trivy, and secret detection
  - Automated versioning and changelog generation
  - Release management with GitHub Container Registry
  - Dependency management with Dependabot
  - Multi-platform build and release automation

- **Documentation**
  - Comprehensive template usage documentation
  - Troubleshooting guide for common issues
  - Action setup and configuration guide
  - Examples and best practices
  - API documentation and integration guides
  - Security policy and vulnerability reporting

### Changed
- N/A

### Deprecated
- N/A

### Removed
- N/A

### Fixed
- N/A

### Security
- **Security Features**
  - Input validation and sanitization
  - Secure credential handling
  - Error message sanitization
  - Dependency vulnerability scanning
  - Secret detection and prevention
  - Container security scanning
  - CodeQL static analysis integration

---

## Version History

### Semantic Versioning
This project follows [Semantic Versioning](https://semver.org/spec/v2.0.0.html):

- **MAJOR** version for incompatible API changes
- **MINOR** version for backwards-compatible functionality additions
- **PATCH** version for backwards-compatible bug fixes

### Conventional Commits
This project uses [Conventional Commits](https://www.conventionalcommits.org/) for commit messages:

- `feat:` - New features (minor version bump)
- `fix:` - Bug fixes (patch version bump)
- `docs:` - Documentation changes (patch version bump)
- `style:` - Code style changes (patch version bump)
- `refactor:` - Code refactoring (patch version bump)
- `perf:` - Performance improvements (patch version bump)
- `test:` - Test additions or changes (patch version bump)
- `chore:` - Maintenance tasks (patch version bump)
- `BREAKING CHANGE:` - Breaking changes (major version bump)

### Automated Versioning
Version bumps are automatically determined by analyzing commit messages:
- Commits with `BREAKING CHANGE` or `!:` trigger major version bumps
- Commits starting with `feat:` trigger minor version bumps
- Commits starting with `fix:`, `docs:`, `style:`, `refactor:`, `perf:`, `test:`, or `chore:` trigger patch version bumps

### Release Process
1. **Automated Analysis** - Commit messages are analyzed to determine version bump type
2. **Changelog Generation** - Conventional changelog is automatically generated
3. **Version Updates** - Version is updated in all relevant files
4. **Tag Creation** - Git tag is created for the new version
5. **Release Creation** - GitHub release is created with assets and notes
6. **Container Publishing** - Docker image is published to GitHub Container Registry

---

## Contributing

When contributing to this project, please follow the conventional commit format to ensure proper versioning and changelog generation.

### Commit Message Format
```
<type>[optional scope]: <description>

[optional body]

[optional footer(s)]
```

### Examples
```
feat: add new Nobl9 project validation

Add comprehensive validation for Nobl9 project configurations including
required fields, format validation, and dependency checks.

Closes #123
```

```
fix: resolve user resolution timeout issue

Increase timeout for user resolution API calls and add retry logic
for better reliability.

Fixes #456
```

```
BREAKING CHANGE: update Nobl9 API client interface

The Nobl9 client interface has been updated to use the latest API
version. This requires updating all client initialization code.

Migration guide: docs/migration-v2.md
```

---

## Links

- [Project Repository](https://github.com/nobl9/nobl9-github-action)
- [Nobl9 Documentation](https://docs.nobl9.com/)
- [Backstage Documentation](https://backstage.io/docs)
- [GitHub Actions Documentation](https://docs.github.com/en/actions)
- [Conventional Commits](https://www.conventionalcommits.org/)
- [Semantic Versioning](https://semver.org/)
- [Keep a Changelog](https://keepachangelog.com/) 