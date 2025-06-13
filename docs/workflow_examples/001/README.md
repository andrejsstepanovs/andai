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
