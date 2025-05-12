package ai

// ForceJSON is a constant that contains a message to force JSON responses.
const ForceJSON = "No yapping. " +
	"Answer only with JSON content. " +
	"Don't explain your choice (no explanation). " +
	"No other explanations or unrelated text is necessary. " +
	"Be careful generating JSON, it needs to be valid."

// TaskSummaryPrompt is a constant that contains a prompt for generating a task summary.
const TaskSummaryPrompt = `I need you to REFORMAT the technical information above into a structured developer task.
DO NOT implement any technical solution - your role is ONLY to organize and present the information.

### Your Think Process:
1. PRIORITY INFORMATION SOURCES (analyze in this order):
	- Current issue descriptions and requirements
	- Latest comments and discussions on the issue
	- Project wiki and documentation
	- Parent issues and dependencies

2. CONTEXT TO INCORPORATE:
	- Project documentation and technical constraints
	- System architecture and integration points
	- Related tickets and dependencies
	- Previous implementation patterns and solutions
	
	### Task Content Requirements:
	- Clearly identify the specific problem/feature to implement
	- Extract all technical requirements and acceptance criteria
	- Highlight potential obstacles, edge cases, and dependencies
	- Include relevant code references, API endpoints, data structures
	- Specify exact files to be modified
	- Identify specific methods to be changed (if known)
	- Reference similar implementations to follow existing patterns

### DELIVERABLE: A FORMATTED TASK WITH THESE SECTIONS:
	
	1. **Summary** (1-2 sentences describing the core task)
	2. **Background** (Essential context for understanding why this work matters)
	3. **Requirements** (Specific, measurable criteria for success)
	4. **Implementation Guide**:
	- Recommended approach
	- Specific steps with technical details
	- Code areas to modify
	- Potential challenges and considerations
	5. **Resources** (Code files, references that developer should work with)
	6. **Constraints** (Limitations or restrictions that may impact development)

### Output Style Requirements:
- Format as an official assignment/directive to a developer
- Use precise technical language appropriate for the development environment
- Prioritize clarity and actionability over comprehensiveness
- Include code snippets or pseudocode where helpful
- Provide context and high-level understanding (marked as contextual information)
- Highlight any areas of uncertainty requiring clarification
- Use clear headings, bullet points, and code blocks for readability

REMEMBER: Your task is ONLY to format and clarify the existing information, not to solve the technical problem or create new solutions.`
