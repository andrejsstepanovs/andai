
# README.md

# Workflow Summary

This workflow uses three main issue types: Story, Task, and Grooming, moving through states: Init, Backlog, Think, Work, Test, Fixme, QA, Deploy, and Done.

Story Initiation & Planning:

A Story starts in Init, moves to Backlog.
In Think, the AI analyzes the Story, comments, and context (summarize-task) to understand requirements and check if rework is needed from a previous failed Test state.
In Work, the AI uses the analysis from Think (via comments) to create specific Task issues (create-issues).
Task Execution & Testing:

Newly created Tasks start in Init, move through Backlog to Think.
In Think, the AI analyzes the Task, its parent Story, comments, and relevant files (context-files, summarize-task) to plan the implementation.
In Work, the AI uses aider (architect-code) to implement the code for the Task, reformats (project-cmd reformat), and commits the changes.
In Test, the AI runs linting and tests (project-cmd lint, project-cmd test). It then uses a general ai command to analyze the results. Based on this analysis, it creates Grooming issues for each detected linter/test error (create-issues action: Grooming). Finally, it uses evaluate (likely based on whether Grooming issues were created) to determine the next state:
Success (No errors): Moves to QA (Human QA).
Failure (Errors found): Moves to Fixme.
Grooming (Error Fixing):

Grooming issues are created by Tasks in the Test state when errors occur. They start in Init.
They move through Backlog to Think, where the AI analyzes the specific error and plans the fix (summarize-task).
In Work, the AI uses aider (code) to implement the fix, reformats, and commits.
Grooming issues skip the automated Test phase (using command: next) and move directly to QA. Correction: Based on the transitions, Grooming issues actually move from Work -> Test -> QA (if success) or Work -> Test -> Fixme (if fail). The Test step for Grooming currently just contains command: next, implying it doesn't re-run tests/linting within the Grooming task's Test state itself, but relies on the transition paths. This might be an area for refinement in the config.
Fixme State (Manual Intervention Loop):

If a Task fails the Test state (meaning Grooming issues were created), it moves to Fixme. This state seems designed for human intervention or review before potentially looping back.
The Fixme state applies only to Story according to workflow.states.Fixme.ai. This seems inconsistent with the transition (Test -> Fixme) which originates from Task issues. Assuming the transition takes precedence, a Task could enter Fixme.
From Fixme, a success transition leads to QA, while a failure transition loops back to Think (presumably for the Task to be re-analyzed).
Quality Assurance (QA) & Deployment:

Tasks and Grooming issues that pass Test (or Fixme) move to QA (Human review).
Stories move to QA automatically via a trigger when all their child Tasks and subsequent Grooming issues reach Done.
If QA passes (human approval assumed via manual state change), the issue moves to Deploy. Failure loops back to Think.
In Deploy, the issue's branch is merged into its parent (merge-into-parent). Tasks/Grooming merge into the Story branch; the Story merges into the project's final_branch.
Finally, the issue moves to Done.
Priorities & Triggers:

The priorities ensure that Grooming tasks are handled first, followed by Tasks, then Stories, generally processing items closer to Deploy before those in earlier states.
Triggers automatically move the parent (Story or Task) to the QA state once all its direct children (Task or Grooming) are Done.


---



# README.md

# Examples

This directory contains examples of how to configure `andai`.


---



# PROJECTS.md

# Projects configuration

Available arguments:
- `identifier` - Project identifier. Keep it unique, short and with no spaces.
- `name` - Human readable project name.
- `description` - Project description.
- `git_path` - Project path in Redmine to the project git repository. Depends on how project is mounted in redmine docker-compose volumes.
- `git_local_dir` - Local path to the project git repository. Best to have it full path to repository. If running AndAI from within docker, then adjust it accordingly.
- `final_branch` - Branch where all code should be merged. If not available, will be created.
- `commands` - Custom project commands. Used via `project-cmd` command in `workflow.issue_types[].jobs[].steps.command`.

Example:
```yaml
projects:
  - identifier: "test-project"
    name: "Test Project"
    description: "Python project doing nothing"
    git_path: "/test-project/.git"
    git_local_dir: "/home/user/www/test-project/.git"
    final_branch: "ANDAI-MAIN"
    wiki: |
      # Test Project
      Doing something important.
      # Important
      - Never change main.py file.
      - You are not allowed to introduce new dependencies.
```

## projects[].commands

When defining custom project command you are forced to define it for all projects. i.e. all projects should have same commands available.

Tip: if there is no alternative, you can define command that dose nothing:
```yaml
projects:
    commands:
      - name: "test"
        command: ["echo", "OK"]
      - name: "lint"
        command: ["echo", "OK"]
```

## commands

- `name` - Command name. Will be used (matched) in `project-cmd` command.
- `command` - Command to execute. List of strings.
- `ignore_err` - Optional. Default false. If true, will ignore command exit code.
- `ignore_stdout_if_no_stderr` - Optional. Default false. If true, will ignore stdout if stderr is empty.
- `success_if_no_output` - Optional. Default false. If true, will consider command successful if there is no output. Useful if you want to comment it in redmine issue via `comment: True`.

Example:
```yaml
projects:
  - identifier: "my-project-001"
    name: "Project Name"
    description: "Description"
    git_path: "/redmine-path-to-repo/.git"
    git_local_dir: "/full/local/path/to/repo.git"
    final_branch: "main"
    wiki: |
      First line of wiki.
    commands:
      - name: "test"
        command: ["make", "test"]
        ignore_err: True
        ignore_stdout_if_no_stderr: True
      - name: "lint"
        command: ["make", "lint"]
        ignore_err: True
        success_if_no_output: True
      - name: "reformat"
        command: ["gofmt", "-s", "-w", "."]
        ignore_err: True
        success_if_no_output: True
```


---



# REDMINE.md

# redmine

Redmine is a flexible project management web application. Written using the Ruby on Rails framework, it is cross-platform and cross-database.

This part of confuguration is quite simple. Most of the values are hardcoded and mean nothing.

- db - Database connection. Because redmine is running in docker-compose, this value should be aligned with `.redmine.env` and `docker-compose.yaml` `database` setup.
- url - Redmine URL. Same story as with `db`.
- api_key - Redmine API key. Hardcode to anything you want really. We are sticking with single `admin` user. If you're running this locally, you can stick with this value.
- repositories - Path to repositories from where redmine is (container). Make sure that in `docker-compose.yaml` your project repositories are mounted to this path.

```yaml
redmine:
  db: redmine:redmine@tcp(localhost:3306)/redmine
  url: "http://localhost:10083"
  api_key: "2159cef2fb6c82c4f66981f199798781e161c694"
  repositories: "/var/repositories/"
```

`repositories` are not mandatory to be correct. But if correct you will have access to your project repositories from redmine web ui.


---



# INSTALLATION.md

# Installation

## Download

Download `andai` executable from [github releases](https://github.com/andrejsstepanovs/andai/tags)

## Build

Checkout the repo and build it from source.

```bash
# building from source
git clone git@github.com:andrejsstepanovs/andai.git
cd andai
make build
ls -l ./build/andai
# add it to PATH or create alias to it
```

## Aider

If you run `andai` in docker, this step is not necessary.

If you plan to run `andai` locally, you will need to install [aider](https://aider.chat/).

Install it and make sure it is available in your PATH. No other configuration is necessary, 
as most of the `aider` configuration will be done via command line arguments.

```
uv tool install aider
uv tool upgrade aider-chat
```

## Quick Start

After this follow [Quick Start](QUICKSTART.md) guide to get you up and running.


---



# CODING_AGENTS.md

# AndAI Coding Agents configuration

Here you should define what coding agent are available.

## Supported
- aider

```yaml
coding_agents:
    aider:
        ....
```

## aider
See [AIDER.md](AIDER.md) for more information.


---



# AIDER.md

# aider

AndAI configuration aider setup.

Arguments:
- timeout - how long to wait for aider to respond
- config - path to `.andai.aider.yaml` file. See [Aider configuration documentation](https://aider.chat/docs/config/aider_conf.html)
- config_fallback - same as `config`, but used when `aider` hits token-limit error. Setup bigger context model here.
- model_metadata_file - Optional. Path to `.andai.aider.model.json` file. See [Aider model metadata documentation](https://aider.chat/docs/config/adv-model-settings.html) and [default values](https://github.com/BerriAI/litellm/blob/main/model_prices_and_context_window.json)
- map_tokens - Optional. How many project map tokens aider will use. Default is 1024. It is good value.
- task_summary_prompt - Optional override prompt for `summarize-task` and aider `summarize` option. Example: `andai` command `command: aider` `summarize: True` configured in `workflow`.

```yaml
coding_agents:
  aider:
    timeout: "180m"
    config: "/full/path/to/here/.andai.aider.yaml"
    config_fallback: "/full/path/to/here/.andai.aider.fallback.yaml"
    model_metadata_file: "/full/path/to/here/.andai.aider.model.json"
    map_tokens:  1024

(!) You are required to configure api key and url in aider config file.

## Aider configuration file

Store your Aider configuration in `.andai.aider.yaml` file. (location defined in `aider.config`)

### Anthropic
```yaml
sonnet: true
model: claude-3-5-sonnet-latest
weak-model: claude-3-5-sonnet-latest
anthropic-api-key: "sk-****"
subtree-only: true
```

### LiteLLM
```yaml
model: openai/gpt4.1
weak-model: openai/gemini-2.0-flash
openai-api-key: "sk-1234"
openai-api-base: http://localhost:4000/v1/
```

### LM Studio
```yaml
model: "openai/qwen2.5-coder-32b-instruct@q4_k_m"
openai-api-key: "lmstudio"
openai-api-base: "http://localhost:1234/v1"
subtree-only: true
```

### Ollama
```yaml
model: ollama_chat/deepseek-r1:32b-qwen-distill-q4_K_M
openai-api-key: "ollama"
set-env:
  - OLLAMA_API_BASE=http://localhost:11434
subtree-only: true
```

## Aider custom model setting
If you configured `model_metadata_file` this is how it could look like (`andai.aider.model.json`):

```json
{
  "openai/qwen2.5-coder-32b-instruct@q4_k_m": {
    "use_temperature": 0.2
  }
}
```



---



# LLM_MODELS.md

# llm_models 

`andai` have `ai` command that will use LLM directly and also other functionality of `andai` workflows is using LLM for some tasks.
You are required to configure this part and set up proper api-key.

- name - Hardcoded "normal" value.

- model - Model name that will be used
- temperature - Temperature for the model
- provider - LLM inference provider 
- base_url - Base URL for the model
- api_key - API key for the model. Can be env variable or hardcoded value. For env variable prefix with `os.environ/YOUR_ENV_VAR_API_KEY`.
- commands - Optional (evaluate, summarize-task, create-issues, ai). List of commands model must be used for. If not set, all commands will use mandatory "normal" model.


```yaml
llm_models:
  - name: "normal"
    temperature: 0.2
    model: "claude-3-7-sonnet-latest"
    provider: "anthropic"
    api_key: "****"
```

```yaml
llm_models:
  - name: "normal"
    temperature: 0.2
    provider: "custom"
    base_url: "https://llm.provider.url.com/v1/chat/completions"
    api_key: os.environ/YOUR_ENV_VAR_API_KEY
```

## External Providers
- `anthropic`
- `cohere`
- `groq`
- `mistral`
- `openai`
- `google`
- `groq`
- `openrouter`
- `deepseek`

## Local Providers
- `provider: custom`
- `base_url` - URL to the model
- `api_key` - API key for the model

### Ollama
```yaml
llm_models:
  - name: "normal"
    temperature: 0.2
    provider: "ollama"
    model: "deepseek-r1:14b-qwen-distill-q4_K_M"
    base_url: "http://localhost:11434"
    api_key: "ollama"
```

### LM Studio
```yaml
llm_models:
  - name: "normal"
    temperature: 0.2
    provider: "custom"
    model: "phi-4"
    base_url: "http://localhost:1234/v1/chat/completions"
    api_key: "lmstudio"
```

# Specific command model example

First command model will be used.

```yaml
llm_models:
  - name: "normal"
    temperature: 0.2
    model: "claude-3-7-sonnet-latest"
    provider: "anthropic"
    api_key: "****"

  - name: "simple"
    temperature: 0.2
    provider: "custom"
    model: "phi-4"
    base_url: "http://localhost:1234/v1/chat/completions"
    api_key: "lmstudio"
    commands:
      - evaluate

  - name: "long"
    temperature: 0.2
    model: "gemini-2.0-flash"
    provider: google
    max_tokens: 2000000
    api_key: "****"
    commands:
      - summarize-task

```


---



# STATES.md

# workflow.states

Define ticketing system (Redmine) states. Think about it as Jira columns in Kanban/Sprint board.
We need to set default state and closed state.

Each state have a name (key value). Name can be anything. Call it a Rhino if you want.

## States arguments:
- `is_default` - When creating new ticket, this state will be set as default.
- `is_first` - Tells Redmine that this is first state in workflow.
- `is_closed` - Tells Redmine that this is closed state.
- `ai` - List of issue Types (defined in `workflow.issue_types`) that AI is allowed to work on in this state. If empty, AI will not work on any of this issue type in this state.
- `description` - Description of the state.

Example:
```yaml
workflow:
  states:
    Init:
      ai: ["Task", "Grooming"]
      is_default: True
      is_first: True
    Backlog:
      ai: ["Story", "Task", "Grooming"]
      description: Issue is ready to be worked on.
    Analysis:
      ai: ["Story", "Task", "Grooming"]
      description: Analyze Issue and plan how to work on it.
    In Progress:
      ai: ["Story", "Task", "Grooming"]
      description: Analyze Issue and plan how to work on it.
    Deployment:
      ai: ["Story", "Task", "Grooming"]
      description: Merge code into to parent Issue branch.
    QA:
      ai: ["Story", "Task", "Grooming"]
      description: |
        Check current state of the story code.
        If everything is fine, move to Done. 
        If not, move to In Progress where new Tasks will try to fix discovered issues.
    Review:
      # Notice how Story is not part of it - user will review Story code. If he/she will move to Deployment, AI will merge it, if he/she will move it to Analysis AndAI will pick up again.
      ai: ["Task", "Grooming"] 
      description: Human QA to check Story code.
    Done:
      description: Issue is completed.
      ai: []
      is_closed: True
```


---



# ISSUE_TYPES.md

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


---



# CONTEXT.md

# workflow.issue_types[].jobs[].steps.context

Context is building knowledge prompt that is given to LLM. Available context values are:
- `ticket` - Includes redmine issue subject, id, type name and description.
- `comments` - Includes all comments from redmine issue.
- `last-comment` - Includes last comment from redmine issue.
- `last-2-comment` - Includes last two comments from redmine issue.
- `last-3-comment` - Includes last three comments from redmine issue.
- `last-4-comment` - Includes last four comments from redmine issue.
- `last-5-comment` - Includes last five comments from redmine issue.
- `project` - Includes project name, identifier and description.
- `wiki` - Includes project wiki. Defined in `projects[]` ends up in redmine project wiki. We pick data from there.
- `children` - Includes all children issues with same info as `ticket`. Is not including Closed issues.
- `siblings` - Includes all siblings issues with same info as `ticket`. Is not including Closed issues.
- `siblings-comments` - Includes all siblings comments.
- `parent` - Includes first single parent issue with same info as `ticket`.
- `parents` - Includes parent and parent parents issues with same info as `ticket`.
- `parent-comments` - Includes parent comment messages.
- `issue_types` - Includes all issue type names with descriptions.
- `affected-files` - Includes all files that were touched by closed children git commits.

Other knowledge info you should know about:
- If previous step had `remember: true` set, then that will be automatically included in knowledge context.
  As example this will include case like this: `context-files` command was executed (before) with `Remember: true`,
  so in this step `context-files` will be available in context.

## Under the hood
AndAI will gather all information necessary and combine it all into single prompt file in temp directory.
This file will be given to LLM for processing. After LLM is done, prompt file will be deleted.


---



# PRIORITIES.md

# workflow.priorities

This is ordered list of tasks that AI should consider first when picking up new task to work on.
It is good idea to organize this in reverse order. For example, first work on Sub-Task that is in Deployment state, then on Sub-Task that is in QA state, etc.
and lastly work on Story or Epic that is available in Deployment and then QA and then In Progress, etc.

This is important because `AndAI` is preserving state in the ticketing system. We can stop it any time and continue work later.
We should be able to pick up where we left off, i.e. most urgent (and smallest) task first.

- `type` - Issue type. Should match `workflow.issue_types` key.
- `state` - State in which task is. Should match `workflow.states` key.

Example:
```yaml
workflow:
  priorities:
    - type: Grooming
      state: Deployment
    - type: Grooming
      state: Review
    - type: Grooming
      state: QA
    - type: Grooming
      state: In Progress
    - type: Grooming
      state: Analysis
    - type: Grooming
      state: Backlog
    - type: Grooming
      state: Init

    - type: Task
      state: Deployment
    - type: Task
      state: Review
    - type: Task
      state: QA
    - type: Task
      state: In Progress
    - type: Task
      state: Analysis
    - type: Task
      state: Backlog
    - type: Task
      state: Init

    - type: Story
      state: Deployment
    - type: Story
      state: QA
    - type: Story
      state: In Progress
    - type: Story
      state: Analysis
    - type: Story
      state: Backlog
```

Note: it is probably possible to deduct this list from other workflow information that is available programmatically.
I will try to do this someday, then `workflow.priorities` would be optional.


---



# TRANSITIONS.md

# workflow.transitions

Here we define how states are connected. This is a simple list of transitions from one state to another.

- `source` - State from which transition. Should match `workflow.states` key.
- `target` - State to which transition. Should match `workflow.states` key.
- `fail` - Optional (true|false). Failure path. See `evaluate` command.
- `success` - Optional (true|false). Success path. See `evaluate` command.

It is OK to create multiple connections between states. It helps when working with the system via browser and often also is necessary to build more complex workflows.

Example:
```yaml
workflow:
  transitions:
    - source: Init
      target: Backlog

    - source: Backlog
      target: Analysis
    - source: Analysis
      target: In Progress

    - source: In Progress
      target: QA

    - source: QA
      success: true
      target: Review
    - source: QA
      fail: true
      target: Analysis

    - source: Review
      success: true
      target: Deployment

    - source: Review
      fail: true
      target: Analysis

    - source: Deployment
      success: true
      target: Done

    - source: Done
      target: QA
```


---



# TRIGGERS.md

# workflow.triggers

Triggers allow to move parent issue to next state when all children issues are in specific state.
For example. After the very last Sub-Task is moved to Done state, then move parent Story task from QA to Review or Deployment.

- `issue_type` - Issue type. Should match `workflow.issue_types` key.
- `if` - List of conditions that need to be met to trigger the action.
    - `moved_to` - Issue (`triggers[].issue_type`) was moved to this state. Must be `workflow.states` value.
    - `all_siblings_status` - All siblings issues are in this state. Must be `workflow.states` value.
    - `transition` - Action to take.
        - `who` - Who should be moved. Can be `parent` or `children`.
        - `to` - State to move to. Must be `workflow.states` value.

Example that will move `Task` issue parent `Story` issue to `Review` state when all `Task` issues are in `Done` state.
Also this will do the same with `Grooming` and `Task` issues. After all `Task` issues are in `Done` state,
AndAI will automatically move parent `Grooming` issue to `Review` state.

```yaml
workflow:
  triggers:
    - issue_type: Task
      if:
        - moved_to: Done
          all_siblings_status: Done
          transition:
            who: parent
            to: Review
    - issue_type: Grooming
      if:
        - moved_to: Done
          all_siblings_status: Done
          transition:
            who: parent
            to: Review
```


---



# COMMANDS.md

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


---



# README.md

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


---



# QUICKSTART.md

# Quick Start

We will create new `andai` project config folder in `/tmp/test` (use any folder you like),
copy configuration files from [config_template/](../config_template/) folder and edit them.

```bash
mkdir /tmp/test ; cd /tmp/test # or any other folder you like

# Download configuration files
wget https://raw.githubusercontent.com/andrejsstepanovs/andai/refs/heads/main/docs/config_template/.andai.aider.yaml
wget https://raw.githubusercontent.com/andrejsstepanovs/andai/refs/heads/main/docs/config_template/.andai.project.yaml
wget https://raw.githubusercontent.com/andrejsstepanovs/andai/refs/heads/main/docs/config_template/docker-compose.yaml
wget https://raw.githubusercontent.com/andrejsstepanovs/andai/refs/heads/main/docs/config_template/.redmine.env

# Check that files are there
➜  test tree -a /tmp/test
/tmp/test
├── .andai.aider.yaml
├── .andai.project.yaml
├── docker-compose.yaml
└── .redmine.env

0 directories, 4 files
```

Then (!) Edit these 3 files:
- `.andai.aider.yaml`
- `.andai.project.yaml`
- `docker-compose.yaml`

## Start Ticketing system

Now that you have configuration files in place (and adjusted with your project and llm config), you can start ticketing system.

```bash
cd /tmp/test
docker-compose up -d
```

This will create new redmine (ticketing system) instance with database.

*(!) Do not configure it. `AndAI` will handle it in next step.*

It will take few seconds until redmine is up and running.

## andai binary
Copy `andai` binary there as well or add it to PATH so it's available from everywhere.

## Configure ticketing system
Now it's time to start using `AndAI` binary.
There are multiple commands that are focusing on setup tasks and ping commands to make sure that all is configured correctly and ready for work.

```bash
cd /tmp/test/

# check that .andai.project.yaml is valid
andai validate config

# this command will only set things up
andai setup all

# this command will ping services. useful to check if everything is in order.
andai ping all

# run single work loop
andai work next

# never ending work loop
andai work loop
```

Alternative (all in one):
```bash
# from same folder where you have `.andai.project.yaml`
andai go
```

## Create Ticket

Open Redmine in browser. If you used provided configuration files, it should be available at `http://localhost:10083`.

With username: `admin`, Password: `admin`.

We do not care about ACL and other security issues, because this is local setup that you in full control.

Let's create simple ticket like:
```
Improve README.md documentation
```

After that observe terminal command that is running `andai`. It should pick up this ticket and start working on it.

## Stop and cleanup
```bash
docker-compose down
```
Be aware that `docker-compose down` will destroy images. If you want to stop and continue then use `docker-compose stop`.

## Follow up

Now that you created simple setup, and it is working as expected, 
you probably want to implement more complex workflows and add more real projects into the mix.

See other examples in [/docs/workflow_examples](../workflow_examples/) folder and [workflow/README.md](workflow/README.md) documentation to see what is available.


---



# README.md

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

Go to [config_template/](../config_template/) and copy all files into new directory.

## workflow
See [workflow/README.md](workflow/README.md) for more information.

## redmine
See [REDMINE.md](REDMINE.md) for more information.

## projects
See [PROJECTS.md](PROJECTS.md) for more information.

## coding_agents
See [coding_agents/CODING_AGENTS.md](coding_agents/CODING_AGENTS.md) for more information.

## llm_models
See [LLM_MODELS.md](LLM_MODELS.md) for more information.


---



# README.md

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
andai ping git              # Test Git repository access and existence of git command line tool
andai ping llm              # Test LLM connection for all configured models
```

### Workflow Execution
```bash
andai work next             # Run a single work cycle.                                                              Optional parameter --project <identifier>
andai work loop             # Run continuous work cycles.                                                           Optional parameter --project <identifier>
andai go                    # Setups everything (setup all) and continuous execution (work loop) in single command. Optional parameter --project <identifier>
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
```

## Development Philosophy

AndAI is designed for experimentation and flexibility. The local nature of the tool makes it easy to:
- Try different workflow configurations
- Start over with a clean slate when needed
- Maintain multiple workflow setups for different projects
- Find the optimal command sequence for your specific AI-assisted development needs

## Contributing

Feel free to experiment with different workflows and configurations. Share your setups to help the community.


---



# README.md

# [AndAI](https://www.andai.pro)

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
Apache License 2.0

# Contributing
Please reach out if you want to contribute to this project.

# Important
By using this project you will be programmatically calling LLMs resulting in costs that you will be responsible for.
Author of this project is not responsible for any costs that you may incur by using this software.


---


