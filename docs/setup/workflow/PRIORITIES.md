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
