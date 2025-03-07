# aider

AndAI configuration aider setup.

Arguments:
- timeout - how long to wait for aider to respond
- config - path to `.andai.aider.yaml` file. See [Aider configuration documentation](https://aider.chat/docs/config/aider_conf.html)
- model_metadata_file - Optional. Path to `.andai.aider.model.json` file. See [Aider model metadata documentation](https://aider.chat/docs/config/adv-model-settings.html)
- map_tokens - Optional. How many project map tokens aider will use. Default is 1024. It is good value.
- task_summary_prompt - Optional. Used with `andai` command `command: aider` `summarize: True` configured in `workflow`.

```yaml
aider:
  timeout: "180m"
  config: "/full/path/to/here/.andai.aider.yaml"
  model_metadata_file: "/full/path/to/here/.andai.aider.model.json"
  map_tokens:  1024
  task_summary_prompt: |
    I need you to REFORMAT the technical information above into a structured developer task.
    DO NOT implement any technical solution - your role is ONLY to organize and present the information.
    
    ### Your Analysis Process:
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
    
    REMEMBER: Your task is ONLY to format and clarify the existing information, not to solve the technical problem or create new solutions.
```

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
