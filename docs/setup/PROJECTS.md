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
