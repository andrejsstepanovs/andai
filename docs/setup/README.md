# AndAI setup and Configuration

App have a lot of configuration options at yor disposal.

Idea is simple but really powerful. Project yaml file is your window to the world of AI-assisted coding.

You can define projects, ticket types, statuses, workflows, AI models, state steps, etc. all via these yaml files.

## Installation
See [INSTALLATION.md](INSTALLATION.md)

## Quick Start
After installation follow [Quick Start](QUICKSTART.md) guide to get you up and running.

# andai configuration file

You must create `.andai.project.yaml` file that will contain all the necessary information for `andai` binary. 
`andai` will use this config file info to connect to redmine, configure it and gain info about projects and how to work with redmine tickets. 

All these root elements should exist within single yaml file.

## workflow
See [workflow/README.md](workflow/README.md) for more information.

## redmine
See [REDMINE.md](REDMINE.md) for more information.

## projects
See [PROJECTS.md](PROJECTS.md) for more information.

## aider
See [AIDER.md](AIDER.md) for more information.

## llm_models
See [LLM_MODELS.md](LLM_MODELS.md) for more information.
