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
- `siblings_comments` - Includes all siblings comments.
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
