# workflow.issue_types

Define issue types. Think about it as Jira issue types. Things like Epic, Story, Task, Sub-Task, etc.

Each issue type have a name (key value).

- `description` - Description of the issue type. Will be given to LLM as a context. See `workflow.issue_types[].jobs[].steps.context[] = "issue_types"`.
- `jobs` - List of jobs that AI should do (synchronously in order) when working on this issue type.

Keep in mind that each issue_type will be sharing same states. i.e. Each issue type will follow the same workflow of states (column to column (Backlog -> In Progress, etc.)).

This results in something like this:
```yaml
workflow:
  issue_types:
      Story:
        description: Single user story.
        jobs:
          - In Progress:
              steps:
                - command: create-issues
                  action: Task
                  context: ["ticket", "comments"]
                  prompt: Split Story into smaller workable coding Tasks.
          - Deployment:
              steps:
                - command: merge-into-parent
      Task:
        description: Single small coding task.
        jobs:
          - Backlog:
              steps:
                - command: next
          - In Progress:
              steps:
                - command: aider
                  action: architect-code
                  context: ["ticket", "comments"]
                  prompt: Implement (code) given Task issue based on available information.
          - Deployment:
              steps:
                - command: merge-into-parent
```

# workflow.issue_types[].jobs[].steps

Each `workflow.issue_types[].jobs` key is a state name in which job should be done. Should match `workflow.states` key value.

At this point we know what issue_type we are dealing with and in which state we are. Here we define what exactly needs to be done, how and in what order.

- `command` - Command that should be executed. See list of available commands in commands chapter.
- `action` - Command action to execute.
- `context` - Optional. But most commands use this. See list of available contexts in context chapter.
- `remember` - Optional Default false. If true, Stdout and Stderr of the command/action output will be stored and available in next step within current job.
- `comment` - Optional. Default false. If true, will comment command/action output to the ticketing system issue. One comment from Stdout and one from Stderr (if any).
- `prompt` - Optional. String. Depends on command/action.
- `summarize` - Optional. Default false. Depends on command/action.
- `comment-summary` - Optional. Default false. Depends on command/action.

# workflow.issue_types[].jobs[].steps.context

See [CONTEXT.md](CONTEXT.md) docs.

# workflow.issue_types[].jobs[].steps.command

There are multiple built in commands that you can use.

You can extend the command list for each `projects`. Useful when dealing with building, testing, linting etc.

## Available commands

See [COMMANDS.md](COMMANDS.md) docs.
