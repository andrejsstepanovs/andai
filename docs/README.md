# AndAI

A flexible local tool for setting up AI-assisted workflows with customizable configurations.

## Quick Start

```bash
andai lets go
```
This single command runs the complete setup and starts the work loop.

## Setup

For detailed installation instructions, see [SETUP.md](setup/README.md).

## Overview

AndAI provides a flexible framework to configure and automate AI-assisted workflows. It integrates with Redmine for project management and supports various LLM configurations to enhance your development process.

The tool is designed to be adaptable to different project requirements, allowing you to experiment with different workflow configurations until you find the optimal setup for your needs.

## Environment Configuration

AndAI works with the `PROJECT` environment variable that defaults to `project`. This determines which configuration file is used:
- Default: `.andai.project.yaml`
- Custom: `.andai.{PROJECT_NAME}.yaml` (e.g., `.andai.project1.yaml` when `PROJECT=project1`)

## CLI Commands

### Help
```bash
andai help
```
Display help information for available commands.

### Setup Commands
```bash
andai setup all             # Configure everything at once
andai setup admin           # Configure admin access without password changes
andai setup auto-increments # Optimize issue, project, and user ID numbering
andai setup projects        # Update project configurations
andai setup settings        # Enable Redmine REST API
andai setup token           # Configure (or retrieve) Redmine admin token
andai setup workflow        # Configure Redmine workflow settings
```

### Validation
```bash
andai validate config       # Validate the AndAI configuration file
```

### Connectivity Tests
```bash
andai ping all              # Test all connections
andai ping aider            # Test aider connection
andai ping api              # Test Redmine API connection
andai ping db               # Test database connection
andai ping git              # Test Git repository access
andai ping llm              # Test LLM connection
```

### Workflow Execution
```bash
andai work next             # Run a single work cycle
andai work loop             # Run continuous work cycles
```

### Issue Management
```bash
andai issue create <type> <subject> <description>   # Create a new issue
andai issue move <subject> <success|fail>           # Move an issue to next step
andai issue move-children <subject> <success|fail>  # Move all child issues to next step
```

### Utility Commands
```bash
andai nothing               # Sleep indefinitely (useful for Docker containers)
andai lets go               # Comprehensive setup and continuous execution
```

## Development Philosophy

AndAI is designed for experimentation and flexibility. The local nature of the tool makes it easy to:
- Try different workflow configurations
- Start over with a clean slate when needed
- Maintain multiple workflow setups for different projects
- Find the optimal command sequence for your specific AI-assisted development needs

## Contributing

Feel free to experiment with different workflows and configurations. Share your setups to help the community.
