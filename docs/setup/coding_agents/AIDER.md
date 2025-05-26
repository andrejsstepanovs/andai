# aider

AndAI configuration aider setup.

Arguments:
- timeout - how long to wait for aider to respond
- config - path to `.andai.aider.yaml` file. See [Aider configuration documentation](https://aider.chat/docs/config/aider_conf.html)
- config_fallback - same as `config`, but used when `aider` hits token-limit error. Setup bigger context model here.
- model_metadata_file - Optional. Path to `.andai.aider.model.json` file. See [Aider model metadata documentation](https://aider.chat/docs/config/adv-model-settings.html)
- map_tokens - Optional. How many project map tokens aider will use. Default is 1024. It is good value.
- task_summary_prompt - Optional override prompt for `summarize-task` and aider `summarize` option. Example: `andai` command `command: aider` `summarize: True` configured in `workflow`.

```yaml
coding_agents:
  aider:
    timeout: "180m"
    config: "/full/path/to/here/.andai.aider.yaml"
    config_fallback: "/full/path/to/here/.andai.aider.fallback.yaml"
    model_metadata_file: "/full/path/to/here/.andai.aider.model.json"
    map_tokens:  1024

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

### LiteLLM
```yaml
model: openai/gpt4.1
weak-model: openai/gemini-2.0-flash
openai-api-key: "sk-1234"
openai-api-base: http://localhost:4000/v1/
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

