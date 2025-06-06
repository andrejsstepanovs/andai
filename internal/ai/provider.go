package ai

import (
	"encoding/json"
	"fmt"

	"github.com/teilomillet/gollm/config"
	"github.com/teilomillet/gollm/providers"
	"github.com/teilomillet/gollm/types"
	"github.com/teilomillet/gollm/utils"
)

type CustomOpenAIProvider struct {
	endpoint     string
	name         string
	apiKey       string
	model        string
	temperature  *float64
	extraHeaders map[string]string
	options      map[string]interface{}
	logger       utils.Logger
}

func NewCustomOpenAIProvider(name, endpoint, apiKey, model string, temperature *float64, extraHeaders map[string]string) providers.Provider {
	if extraHeaders == nil {
		extraHeaders = make(map[string]string)
	}
	return &CustomOpenAIProvider{
		name:         name,
		endpoint:     endpoint,
		apiKey:       apiKey,
		model:        model,
		extraHeaders: extraHeaders,
		options:      make(map[string]interface{}),
		logger:       utils.NewLogger(utils.LogLevelInfo),
		temperature:  temperature,
	}
}

func (p *CustomOpenAIProvider) SetLogger(logger utils.Logger) {
	p.logger = logger
}

func (p *CustomOpenAIProvider) SetOption(key string, value interface{}) {
	p.options[key] = value
	p.logger.Debug("Option set", "key", key, "value", value)
}

func (p *CustomOpenAIProvider) SetDefaultOptions(config *config.Config) {
	if p.temperature != nil {
		p.SetOption("temperature", p.temperature)
	}
	p.SetOption("max_tokens", config.MaxTokens)
	if config.Seed != nil {
		p.SetOption("seed", *config.Seed)
	}
	p.logger.Debug("Default options set", "temperature", config.Temperature, "max_tokens", config.MaxTokens, "seed", config.Seed)
}

func (p *CustomOpenAIProvider) Name() string {
	return p.name
}

func (p *CustomOpenAIProvider) Endpoint() string {
	return p.endpoint
}

func (p *CustomOpenAIProvider) SupportsJSONSchema() bool {
	return true
}

func (p *CustomOpenAIProvider) Headers() map[string]string {
	headers := map[string]string{
		"Content-Type":  "application/json",
		"Authorization": "Bearer " + p.apiKey,
	}

	for key, value := range p.extraHeaders {
		headers[key] = value
	}

	p.logger.Debug("Headers prepared", "headers", headers)
	return headers
}

func (p *CustomOpenAIProvider) PrepareRequest(prompt string, options map[string]interface{}) ([]byte, error) {
	request := map[string]interface{}{
		"model": p.model,
		"messages": []map[string]interface{}{
			{
				"role":    "user",
				"content": prompt,
			},
		},
	}

	// Handle tool_choice
	if toolChoice, ok := options["tool_choice"].(string); ok {
		request["tool_choice"] = toolChoice
	}

	// Handle tools
	if tools, ok := options["tools"].([]utils.Tool); ok && len(tools) > 0 {
		openAITools := make([]map[string]interface{}, len(tools))
		for i, tool := range tools {
			openAITools[i] = map[string]interface{}{
				"type": "function",
				"function": map[string]interface{}{
					"name":        tool.Function.Name,
					"description": tool.Function.Description,
					"parameters":  tool.Function.Parameters,
				},
				"strict": true, // Add this if you want strict mode
			}
		}
		request["tools"] = openAITools
	}

	// Add other options
	for k, v := range p.options {
		if k != "tools" && k != "tool_choice" {
			request[k] = v
		}
	}
	for k, v := range options {
		if k != "tools" && k != "tool_choice" {
			request[k] = v
		}
	}

	return json.Marshal(request)
}

// nolint: unused
func (p *CustomOpenAIProvider) createBaseRequest(prompt string) map[string]interface{} {
	var request map[string]interface{}
	if err := json.Unmarshal([]byte(prompt), &request); err != nil {
		p.logger.Debug("Prompt is not a valid JSON, creating standard request", "error", err)
		request = map[string]interface{}{
			"model": p.model,
			"messages": []interface{}{
				map[string]interface{}{
					"role":    "user",
					"content": prompt,
				},
			},
		}
	}
	return request
}

// nolint: unused
func (p *CustomOpenAIProvider) processMessages(request map[string]interface{}) {
	p.logger.Debug("Processing messages")
	if messages, ok := request["messages"]; ok {
		switch msg := messages.(type) {
		case []interface{}:
			for i, m := range msg {
				if msgMap, ok := m.(map[string]interface{}); ok {
					p.processFunctionMessage(msgMap)
					p.processToolCalls(msgMap)
					msg[i] = msgMap
				}
			}
		case []map[string]string:
			newMessages := make([]interface{}, len(msg))
			for i, m := range msg {
				msgMap := make(map[string]interface{})
				for k, v := range m {
					msgMap[k] = v
				}
				p.processFunctionMessage(msgMap)
				p.processToolCalls(msgMap)
				newMessages[i] = msgMap
			}
			request["messages"] = newMessages
		default:
			p.logger.Warn("Unexpected type for messages", "type", fmt.Sprintf("%T", messages))
		}
	}
	p.logger.Debug("Messages processed", "messageCount", len(request["messages"].([]interface{})))
}

// nolint: unused
func (p *CustomOpenAIProvider) processFunctionMessage(msgMap map[string]interface{}) {
	if msgMap["role"] == "function" && msgMap["name"] == nil {
		if content, ok := msgMap["content"].(string); ok {
			var contentMap map[string]interface{}
			if err := json.Unmarshal([]byte(content), &contentMap); err == nil {
				if name, ok := contentMap["name"].(string); ok {
					msgMap["name"] = name
					p.logger.Debug("Function name extracted from content", "name", name)
				}
			}
		}
	}
}

// nolint: unused
func (p *CustomOpenAIProvider) processToolCalls(msgMap map[string]interface{}) {
	if toolCalls, ok := msgMap["tool_calls"].([]interface{}); ok {
		for j, call := range toolCalls {
			if callMap, ok := call.(map[string]interface{}); ok {
				if function, ok := callMap["function"].(map[string]interface{}); ok {
					if args, ok := function["arguments"].(string); ok {
						var parsedArgs map[string]interface{}
						if err := json.Unmarshal([]byte(args), &parsedArgs); err == nil {
							function["arguments"] = parsedArgs
							callMap["function"] = function
							toolCalls[j] = callMap
							p.logger.Debug("Tool call arguments parsed", "functionName", function["name"], "arguments", parsedArgs)
						}
					}
				}
			}
		}
		msgMap["tool_calls"] = toolCalls
	}
}

// nolint: unused
func (p *CustomOpenAIProvider) addOptions(request map[string]interface{}, options map[string]interface{}) {
	for k, v := range p.options {
		request[k] = v
	}
	for k, v := range options {
		request[k] = v
	}
	p.logger.Debug("Options added to request", "options", options)
}

func (p *CustomOpenAIProvider) PrepareRequestWithSchema(prompt string, options map[string]interface{}, schema interface{}) ([]byte, error) {
	p.logger.Debug("Preparing request with schema", "prompt", prompt, "schema", schema)
	request := map[string]interface{}{
		"model": p.model,
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
		"response_format": map[string]interface{}{
			"type":   "json_schema",
			"schema": schema,
		},
	}

	for k, v := range options {
		request[k] = v
	}

	reqJSON, err := json.Marshal(request)
	if err != nil {
		p.logger.Error("Failed to marshal request with schema", "error", err)
		return nil, err
	}

	p.logger.Debug("Request with schema prepared", "request", string(reqJSON))
	return reqJSON, nil
}

func (p *CustomOpenAIProvider) ParseResponse(body []byte) (string, error) {
	var response struct {
		Choices []struct {
			Message struct {
				Content   string `json:"content"`
				ToolCalls []struct {
					ID       string `json:"id"`
					Type     string `json:"type"`
					Function struct {
						Name      string `json:"name"`
						Arguments string `json:"arguments"`
					} `json:"function"`
				} `json:"tool_calls"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return "", err
	}

	if len(response.Choices) == 0 {
		return "", fmt.Errorf("empty response from API")
	}

	message := response.Choices[0].Message
	if message.Content != "" {
		return message.Content, nil
	}

	if len(message.ToolCalls) > 0 {
		toolCallJSON, err := json.Marshal(message.ToolCalls)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("<function_call>%s</function_call>", toolCallJSON), nil
	}

	return "", fmt.Errorf("no content or tool calls in response")
}

func (p *CustomOpenAIProvider) HandleFunctionCalls(body []byte) ([]byte, error) {
	var response struct {
		Choices []struct {
			Message struct {
				ToolCalls []struct {
					Function struct {
						Name      string `json:"name"`
						Arguments string `json:"arguments"`
					} `json:"function"`
				} `json:"tool_calls"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("error parsing response: %w", err)
	}

	if len(response.Choices) == 0 || len(response.Choices[0].Message.ToolCalls) == 0 {
		return nil, fmt.Errorf("no tool calls found in response")
	}

	toolCalls := response.Choices[0].Message.ToolCalls
	result := make([]map[string]interface{}, len(toolCalls))
	for i, call := range toolCalls {
		var args map[string]interface{}
		if err := json.Unmarshal([]byte(call.Function.Arguments), &args); err != nil {
			return nil, fmt.Errorf("error parsing arguments: %w", err)
		}
		result[i] = map[string]interface{}{
			"name":      call.Function.Name,
			"arguments": args,
		}
	}

	return json.Marshal(result)
}

func (p *CustomOpenAIProvider) SetExtraHeaders(extraHeaders map[string]string) {
	p.extraHeaders = extraHeaders
	p.logger.Debug("Extra headers set", "headers", extraHeaders)
}

// PrepareRequestWithMessages creates a request body using structured message objects
// rather than a flattened prompt string.
func (p *CustomOpenAIProvider) PrepareRequestWithMessages(messages []types.MemoryMessage, options map[string]interface{}) ([]byte, error) {
	request := map[string]interface{}{
		"model":    p.model,
		"messages": []map[string]interface{}{},
	}

	// Add system prompt if present
	if systemPrompt, ok := options["system_prompt"].(string); ok && systemPrompt != "" {
		request["messages"] = append(request["messages"].([]map[string]interface{}), map[string]interface{}{
			"role":    "system",
			"content": systemPrompt,
		})
	}

	// Convert structured messages to Groq format (OpenAI compatible)
	for _, msg := range messages {
		request["messages"] = append(request["messages"].([]map[string]interface{}), map[string]interface{}{
			"role":    msg.Role,
			"content": msg.Content,
		})
	}

	// Add other options from provider and request
	for k, v := range p.options {
		if k != "messages" {
			request[k] = v
		}
	}
	for k, v := range options {
		if k != "messages" && k != "system_prompt" && k != "structured_messages" {
			request[k] = v
		}
	}

	return json.Marshal(request)
}

func (p *CustomOpenAIProvider) SupportsStreaming() bool {
	return false
}

// PrepareStreamRequest prepares a request body for streaming
func (p *CustomOpenAIProvider) PrepareStreamRequest(prompt string, options map[string]interface{}) ([]byte, error) {
	options["stream"] = true
	return p.PrepareRequest(prompt, options)
}

func (p *CustomOpenAIProvider) ParseStreamResponse(chunk []byte) (string, error) {
	var response struct {
		Choices []struct {
			Delta struct {
				Content string `json:"content"`
			} `json:"delta"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(chunk, &response); err != nil {
		return "", err
	}
	if len(response.Choices) == 0 {
		return "", nil
	}
	return response.Choices[0].Delta.Content, nil
}
