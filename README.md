# AndAI

A local tool for organizing AI-assisted coding tasks with ticketing system and git integration.

Is using:
- golang
- [Aider](https://aider.chat/) (possible within docker)
- [Redmine](https://www.redmine.org/) (with mariadb in docker) 

kudos to [sameersbn/docker-redmine](https://github.com/sameersbn/docker-redmine) for redmine docker setup and [gollm](https://gollm.co/) team for cool llm lib

## What is AndAI?

**A no-nonsense tool for organizing AI-assisted coding tasks for real-world developers.**

AndAI is a disposable, configurable tool that combines local ticketing with git branching to manage AI-assisted development tasks.

## How Does AndAI Work?

You define ticket types, statuses, and workflows in a YAML config file and bind them with multiple git projects you are working with.

After the setup, you create a ticket in the UI and let AndAI work on it. When the job is done, you review the changes and approve or reject them (or auto merge everything ;]).

This results in a workflow where you keep creating tickets and reviewing results all day long. 
Jump in and commit your own changes on top of AI code if you want to, 
or just reject with a comment about what is wrong; it's all up to you.

## Key Benefits

### ✅ Engineering Accountability
Maintain full responsibility for your codebase while leveraging AI as a powerful development tool. Your name will be under final code, so you will need to QA and groom it. AI just helps you with the heavy lifting.

### ✅ Ticketing system
- **Local**: Benefit from a ticketing system intertwined with git branching without leaking commits to the cloud.
- **Convenience**: Define your AI tasks in browser and let AndAI do the rest. Review and approve/reject in browser. Open IDE when needed to refine if necessary.

### ✅ Structured Git Integration
- **Branch-per-task workflow**: Every ticket and subtask gets its own isolated git branch
- **Controlled merging**: Changes are only merged after explicit approval
- **Clean history**: Squash your and AI colab code into a clean commits before shipping

### ✅ Complete Workflow Control
- **Review flexibility**: Accept, reject, or edit AI-generated changes at any stage (or auto merge everything)
- **Customizable process**: Define your own ticket types, statuses, and workflows (find and benefit from your perfect workflow)
- **Iterative feedback**: Provide guidance to improve AI output via comments or new tasks, subtasks

### ✅ Developer-Friendly Design
- **State persistence**: Ticket state, comments, branches, etc. are preserved, allowing you to stop/restart anytime
- **Local-first**: Everything runs on your system - no one needs to know about your git branch mess or how you manage your tickets
- **Offline capable**: Works with local or external LLMs
- **Disposable**: Set it up in seconds, tear it down just as quickly

## Documentation

### Getting Started
- [Quick Start Guide](docs/setup/QUICKSTART.md)
- [Installation Instructions](docs/setup/INSTALLATION.md)

### Core Concepts
- [Workflow Overview](docs/setup/workflow/README.md)
- [Context Management](docs/setup/workflow/CONTEXT.md)
- [Available Commands](docs/setup/workflow/COMMANDS.md)

### Configuration
- [Project Configuration](docs/setup/PROJECTS.md)
- [AIDER Integration](docs/setup/AIDER.md)
- [LLM Configuration](docs/setup/LLM_MODELS.md)
- [Redmine Setup](docs/setup/REDMINE.md)

### Customization
- [Issue Types](docs/setup/workflow/ISSUE_TYPES.md)
- [Workflow States](docs/setup/workflow/STATES.md)

# License
TODO

# Contributing
Please reach out if you want to contribute to this project.

# Important
By using this project you will be programmatically calling LLMs resulting in costs that you will be responsible for.
Author of this project is not responsible for any costs that you may incur by using this software.
