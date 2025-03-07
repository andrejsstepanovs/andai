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
