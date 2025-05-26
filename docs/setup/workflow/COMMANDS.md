# Comment and Remember

All commands can take advantage of `comment` and `remember` flags.

## Comment
If `comment: true` then Stdout and Stderr of the command output will be commented to the ticketing system issue. One comment from Stdout and one from Stderr (if any).

## Remember
If `remember: true` then Stdout and Stderr of the command output will be stored and available in next step within current job.

# Available commands

Here is the list of available commands that you can use in your workflow job steps.

# next

Will do nothing and move the issue to next state. If multiple states are available will use the want that is marked with `success: true` in `workflow.transitions`.

```yaml
workflow:
  issue_types:
    My Issue Type Name:
      jobs:
        My Step Name:
          steps:
            - command: next
```

# merge-into-parent

Will merge the current issue branch into parent issue branch. If there is no parent available then will merge into `project.final_branch`.
If `project.final_branch` is not available then will create it.

```yaml
workflow:
  issue_types:
    My Issue Type Name:
      jobs:
        Deployment:
          steps:
            - command: merge-into-parent
```

# project-cmd

Executes command pre-defined in `projects[].commands`.

See how `project.commands` is defined in [../PROJECTS.md](../PROJECTS.md) docs.

Works nicely with `remember: true` and `comment: true` to store and comment command output.

```yaml
workflow:
  issue_types:
    My Issue Type Name:
      jobs:
        QA:
          steps:
            - command: project-cmd
              action: lint
            - command: project-cmd
              action: test
```

# create-issues

Will create new children issues based on context.

- `action` - Issue type to create. Should match `workflow.issue_types` key.
- `prompt` - Prompt for LLM.
- `context` - List of context sources. See [CONTEXT.md](CONTEXT.md) for more info.

```yaml
workflow:
  issue_types:
    Story:
      description: |
        A specific functionality or component. Scope: Feature components.
      jobs:
        In Progress:
          steps:
            - command: create-issues
              action: Task
              prompt: |
                Based on last comments and context that you are presented with,
                split current Story issue into Task issues using pre-planned idea mentioned in comments. 
                Write a detailed description of what needs to be done. 
                Comply to clarification comments if exists.
                Make the new issues detailed, clear and concise. Dont forget to mention scope of the files that need to be worked on.
                Keep in mind that Task issue are small enough to be implemented in a single code file or even better - one part of a single code file.
              context: ["issue_types", "ticket", "comments"]
```

It is useful to chain this command with `summarize-task` and `context-files`, `context-commits` commands.

# summarize-task

Based on context will summarize the task using LLM. This is useful when complexity rises and task is being processed multiple times.
Then just the issue description becomes outdated and more important things are located in issue comments.

```yaml
workflow:
  issue_types:
    Story:
      description: |
        A specific functionality or component. Scope: Feature components.
      jobs:
        Analysis:
          steps:
            - command: summarize-task
              comment: True
              context: ["issue_types", "project", "wiki", "ticket", "comments", "children"]
              prompt: |
                Break down the current Story issue into Task issues based on pre-planned ideas from comments, 
                ensuring detailed descriptions for each task. Review existing comments to clarify requirements, 
                check if the task was previously worked on, and determine if adjustments are needed. 
                If adjustments are needed then it means that the task failed QA, 
                create new issues to address QA findings. If there’s no evidence of prior work, 
                create new tasks to implement the Story. 
                Use comment timestamps and context to guide decisions, 
                ensuring alignment with QA requirements.
                Now using what you know, analyze current_issue and comments carefully.
                Suggest how to split current Story issue into smaller scope Task issues (make sure to stick with Task issue scope and create as many Task issues as necessary).
                Dont forget to mention scope of the Task issues you create.
                Make sure each Task issue is small enough to be implemented in a single code file or even better - one part of a single code file.
```

This command can be chained with other commands quite nicely. Either via `remember: true` and follow up with the implementation 
or via `comment: true` to store the output in the ticketing system.

# context-files

This command will traverse all given context text and match mentioned files that exist in repository. 
The file name and full content of the files will be included in next step prompt.

This command require mandatory `remember: true` flag to be set. So you want to do this before other command that will put the file content to good use.

```yaml
workflow:
  issue_types:
    Story:
      description: Story issue type that defines specific feature requirements that need to be implemented.
      jobs:
        Work:
          steps:
            - command: context-files
              context: ["wiki", "ticket", "parents", "comments"]
              remember: True
            - command: summarize-task
              comment: True
              context: ["issue_types", "project", "wiki", "ticket", "comments", "children"]
              prompt: Make a detailed analysis of the Story issue and comments.
            - command: create-issues
              action: Task
              prompt: |
                Create small workable Task issues that will result successful implementation of given Story issue.
```

# context-commits

This command will traverse all given context text and match existing git sha (long).
If it exists in git, will include the commit path message to context.

```yaml
workflow:
  issue_types:
    Bug:
      description: Bug issue type that defines specific bug that needs to be fixed.
      jobs:
        Work:
          steps:
            - command: context-commits
              context: ["parents", "comments", "parent-comments"]
              remember: True
            - command: code
              comment: True
              context: ["project", "wiki", "ticket", "comments"]
              prompt: Fix linter issues introduced in previous commits.
```

# commit

Will commit uncommited files. Using `git status` to find uncommited files and will commit them with `prompt` comment.

- `prompt` - Comment for the commit.

```yaml
workflow:
  issue_types:
    Sub-Task:
      description: Coding task
      jobs:
        In Progress:
          steps:
            - command: project-cmd
              action: reformat
            - command: commit
              prompt: linter changes
```


# git

Executes your custom `git` command where `action` contains follow up command.

```yaml
workflow:
  issue_types:
    Sub-Task:
      description: Coding task
      jobs:
        In Progress:
          steps:
            - command: git
              action: status
              comment: true
```

Not sure where this can be used. Explore other commands before falling back to this one.

# bash

Executes your custom `bash` command where `action` contains command itself (space separated).

```yaml
workflow:
  issue_types:
    Sub-Task:
      description: Coding task
      jobs:
        In Progress:
          steps:
            - command: bash
              action: "echo hello"
              comment: true
```

# evaluate

This is really useful command. It will evaluate success or failure of the given context and move the issue to next state based on that.

- The state must have multiple transitions
- There must be 1 `success: true` and one `fail: true` transition. See (TRANSITIONS.md)[TRANSITIONS.md] docs.
- `context` - mandatory
- `prompot` - Optional

```yaml
workflow:
  issue_types:
    Sub-Task:
      description: Coding task
      jobs:
        QA:
          steps:
            - command: evaluate
              context: ["comments"]
```

# ai

Custom LLM command. Will ask LLM to generate a response based on the given context and prompt.

- `prompt` - Prompt for LLM.
- `context` - List of context sources. See [CONTEXT.md](CONTEXT.md) for more info.

```yaml
workflow:
  issue_types:
    Sub-Task:
      description: Coding task
      jobs:
        QA:
          steps:
            - command: ai
              prompt: |
                If there are any linter errors then pinpoint exact files and place in file (using linter output) that is at fault.
                Answer with clear and concise explanation of the sources for these linter errors.
                If there are any testing errors then pinpoint exact files and test cases that are faulty (using tests output).
                Keep the answer straight to the point of what was test cases are broken and in what file and where.
                If there are no linter and testing errors then answer with "Linter and tests are OK".
              comment: True
              remember: True
```

# aider

Aider is bread and butter of the `AndAI`. It is external tool [aider](https://aider.chat/) that is used for coding and code analysis tasks.

- `action` - Available values are `commit`, `architect`, `code`, `architect-code`
- `prompt` - Prompt for Aider.
- `summarize` - Optional. Default false. If true, will summarize the context and give it to aider as a task to work on. You can override default summarization prompt via `aider.task_summary_prompt`.
- `comment-summary` - Optional. Default false. Works together with `summarize` (`summarize` must be `true`). If true, will comment the summary as issue comment.
- `context` - List of context sources. See [CONTEXT.md](CONTEXT.md) for more info.

## actions

Use full power of `aider` tool with these actions.

(!) Aider commands will result in real code changes and lost uncommited changes in your repository. So, please, commit or stash your uncommited changes before using `aider` commands.

### commit

This command will use aider to commit uncommited code. Works same as `commit` command but via aider i.e. LLM.

### architect

This command will use aider to `chat mode`. No code changes will be made. It is used for analyzing issue together with code and plan the next steps.
It can be pre-processing step before issue splitting into children tasks (`create-issues` command) or pre-analyze before coding steps.

### code

This will use aider `code` mode. Aider will make code changes in current working git branch and will commit it. All uncommited changes will be lost.

### architect-code

This will use aider `architect` mode. This will do more thinking when making code changes resulting in better results but more LLM tokens will be used.

Examples:
```yaml
workflow:
  issue_types:
    Grooming:
      description: |
        A unit of work focused on adjusting parent task code and fixing linter and tests code.
      jobs:
        In Progress:
          steps:
            - command: aider
              action: code
              summarize: True
              comment-summary: False
              comment: False
              context: ["issue_types", "project", "wiki", "parents", "ticket", "last-comment"]
              prompt: |
                Implement (code) given Task issue based on ticket description and last comments.
                It is not allowed to create new files! 
                Evaluate other project files to learn how to do problematic parts 
                and make necessary changes using that knowledge.
            - command: project-cmd
              action: reformat
            - command: commit
              prompt: linter changes
```

Run custom project commands to:
- reformat code
- commit reformatting changes
- run linter and remember
- run tests and remember
- run aider architect (chat mode) via summary of the context
- this summary will be commented to the ticketing system because `comment-summary: True`
- evaluate the issue based on aider output that was commented via `comment: True`
- issue will be moved to success path if all is OK or to reject path if there are issues to be resolved.
- idea is that `andai` will pick this Story task again and will create new sub-tasks based on comments that we just saved here.

```yaml
workflow:
  issue_types:
    Story:
      description: "A specific functionality or component. Scope: Feature components."
      jobs:
        QA:
        steps:
          - command: project-cmd
            action: reformat
          - command: commit
            prompt: linter changes
          - command: project-cmd
            action: lint
            remember: true
          - command: project-cmd
            action: test
            remember: true
          - command: context-files
            context: [ "wiki", "ticket", "comments", "children" ]
            remember: True
          - command: aider
            action: architect
            prompt: |
              Check recently implemented code necessary to complete Story issue.
              Check in recent comments test and linter results and code logical errors.
              Evaluate issues you found and summarize in positive or negative feedback.
              Unit tests and linter results (see previous comments) are mandatory for positive feedback.
              If feedback is mostly positive and code will work as expected it should be positive.
              If code have obvious syntactic or logical errors it should be negative.
              All other context information available to you is meant only to help you figure out what code you should evaluate.
              Answer strictly based on code that is present.
            context: ["project", "wiki", "ticket", "children", "affected-files", "comments"] # TODO add children-closed
            comment: True
            summarize: True
            comment-summary: True
          - command: evaluate
            context: ["comments"]
```
