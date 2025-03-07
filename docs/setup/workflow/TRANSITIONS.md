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
