# Workflow configuration

Don't be afraid with making mistakes in AndAI configuration. 
We have really strict validation process `andai validate config` that will help you to find any issues in your configuration files.

# AndAI workflow - Branches

Each `workflow.issue_types[].jobs[]` is started with preparing the environment. This includes:
- Changing active directory to the project directory.
- Open git repository.
- Switch to parent branch or (if no parent available) to `project.final_branch`.
- Creating (or check out if exists) to current issue branch (because parent branch was checked out before, this will be a branch derived from parent branch).
- Fetching all context information that is available in redmine. Current issue, parent, siblings, children, etc.
- Proceeding with job commands.

Every `AndAI` issue is backed up with git branch prefixed with `AI-`. If issue ID is 102 then branch will be `AI-102`.


# workflow.states

Define workflow states and for each issue_types that AndAI is allowed to work on.

See [STATES.md](STATES.md) docs.

## workflow.transitions

Defines available workflow state transitions. Think about it as Jira transitions from column to column.

See [TRANSITIONS.md](TRANSITIONS.md) docs.

# workflow.priorities

Defines in what order AI should pick up new tasks.

See [PRIORITIES.md](PRIORITIES.md) docs.

# workflow.triggers

Defines automatic state changes based on children issues state change. Useful when all children are finished and you need to proceed the work with a parent.

See [TRIGGERS.md](TRIGGERS.md) docs.

# workflow.issue_types

Define issue types. Think about it as Jira issue types. Things like Epic, Story, Task, Sub-Task, etc.

and how each issue type should be handled by AI in each step.

See [ISSUE_TYPES.md](ISSUE_TYPES.md) docs.

# workflow.issue_types[].jobs[].steps.context

See [CONTEXT.md](CONTEXT.md) docs.

# workflow.issue_types[].jobs[].steps.command

See [COMMANDS.md](COMMANDS.md) docs.
