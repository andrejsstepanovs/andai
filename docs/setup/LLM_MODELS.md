# llm_models 

`andai` have `ai` command that will use LLM directly and also other functionality of `andai` workflows is using LLM for some tasks.
You are required to configure this part and set up proper api-key.

- name - Hardcoded "normal" value.

- model - Model name that will be used
- temperature - Temperature for the model
- provider - LLM inference provider 
- base_url - Base URL for the model
- api_key - API key for the model
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
    api_key: "****"
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
