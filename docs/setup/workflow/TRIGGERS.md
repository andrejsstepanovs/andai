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
