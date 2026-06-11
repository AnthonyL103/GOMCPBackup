// transport/anthropic.go
package transport

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	agent "github.com/AnthonyL103/GOMCP/Agent"
	"github.com/AnthonyL103/GOMCP/chat"
	"github.com/AnthonyL103/GOMCP/infrageneration"
	"github.com/AnthonyL103/GOMCP/protocol/llmprotocol"
	"github.com/AnthonyL103/GOMCP/servergeneration"
)

type AnthropicProvider struct {
	APIKey      string
	Model       string
	Temperature float32
	MaxTokens   int
	// Optional callback fired immediately after each tool cycle completes,
	// before the next LLM call. Used by the HTTP server to broadcast over WS.
	OnToolCall func(msg chat.Message)
}

func NewAnthropicProvider(config *agent.LLMConfig) *AnthropicProvider {
	return &AnthropicProvider{
		APIKey:      config.APIKey,
		Model:       config.Model,
		Temperature: config.Temperature,
		MaxTokens:   config.MaxTokens,
	}
}

func (p *AnthropicProvider) GetProviderName() string {
	return "anthropic"
}

func (p *AnthropicProvider) SendRequest(c *chat.Chat, ag *agent.Agent, userMessage string) error {
	c.AddUserMessage(userMessage)

	agentInstructions := llmprotocol.GetAgentInstructions(ag)
	availableTools := llmprotocol.ExtractTools(ag)
	formattedTools := p.buildTools(availableTools, ag)
	messages := p.buildMessages(c)

	requestBody := map[string]interface{}{
		"model":       p.Model,
		"max_tokens":  p.MaxTokens,
		"temperature": p.Temperature,
		"system":      agentInstructions,
		"messages":    messages,
	}

	if len(formattedTools) > 0 {
		requestBody["tools"] = formattedTools
	}

	response, err := p.sendHTTPRequest(requestBody)
	if err != nil {
		return err
	}

	responseText, toolCallID, toolName, toolParams, stopReason, err := p.parseResponse(response)
	if err != nil {
		return err
	}

	toolsProcessed := false
	for stopReason == "tool_use" && toolName != "" {
		toolsProcessed = true

		currentToolCallID := toolCallID
		currentToolName := toolName
		currentToolParams := toolParams

		toolInfo, exists := availableTools[currentToolName]
		if !exists {
			if !isServerGenerationToolAnthropic(currentToolName) && !isInfraGenerationToolAnthropic(currentToolName) {
				return fmt.Errorf("tool %s not found", currentToolName)
			}
			toolInfo = llmprotocol.ToolInfo{ServerID: "infrastructure_generation", Handler: currentToolName}
		}

		if isServerGenerationToolAnthropic(currentToolName) && !ag.ServerGeneration {
			return fmt.Errorf("tool %s not available; enable server generation in config", currentToolName)
		}

		if isInfraGenerationToolAnthropic(currentToolName) && !ag.InfraGeneration {
			return fmt.Errorf("tool %s not available; enable infra generation in config", currentToolName)
		}

		toolResult, isError := llmprotocol.ExecuteTool(ag, &chat.ToolCall{
			ServerID:   toolInfo.ServerID,
			ToolID:     currentToolName,
			Handler:    toolInfo.Handler,
			Parameters: currentToolParams,
		})

		truncatetoolparams := make(map[string]interface{})
		for k, v := range currentToolParams {
			if strVal, ok := v.(string); ok && len(strVal) > 100 {
				truncatetoolparams[k] = strVal[:100] + "..."
			} else {
				truncatetoolparams[k] = v
			}
		}
		fmt.Println("Calling tool:", currentToolName, "with params:", truncatetoolparams)

		// Build the completed tool cycle message
		toolMsg := chat.Message{
			Role: "assistant",
			ToolCall: &chat.ToolCall{
				ServerID:   toolInfo.ServerID,
				ToolID:     currentToolName,
				Handler:    toolInfo.Handler,
				Parameters: currentToolParams,
				Reasoning:  "",
				ToolUseID:  currentToolCallID,
			},
			ToolResult: &chat.ToolResult{
				ServerID:  toolInfo.ServerID,
				ToolID:    currentToolName,
				Content:   toolResult,
				IsError:   isError,
				ToolUseID: currentToolCallID,
			},
		}

		if p.OnToolCall != nil {
			p.OnToolCall(toolMsg)
		}

		messages = p.buildMessages(c)

		messages = append(messages, map[string]interface{}{
			"role": "assistant",
			"content": []map[string]interface{}{
				{
					"type":  "tool_use",
					"id":    currentToolCallID,
					"name":  currentToolName,
					"input": currentToolParams,
				},
			},
		})

		messages = append(messages, map[string]interface{}{
			"role": "user",
			"content": []map[string]interface{}{
				{
					"type":        "tool_result",
					"tool_use_id": currentToolCallID,
					"content":     toolResult,
					"is_error":    isError,
				},
			},
		})

		requestBody["messages"] = messages
		response, err = p.sendHTTPRequest(requestBody)
		if err != nil {
			return err
		}

		responseText, toolCallID, toolName, toolParams, stopReason, err = p.parseResponse(response)
		if err != nil {
			return err
		}

		// Save to chat history after callback so history is consistent
		c.AddAssistantMessage(
			responseText,
			toolMsg.ToolCall,
			toolMsg.ToolResult,
		)
	}

	if !toolsProcessed {
		c.AddAssistantMessage(responseText, nil, nil)
	}

	return nil
}

func (p *AnthropicProvider) buildMessages(c *chat.Chat) []map[string]interface{} {
	messages := []map[string]interface{}{}

	for _, msg := range c.GetMessages() {
		switch msg.Role {
		case "user":
			messages = append(messages, map[string]interface{}{
				"role":    "user",
				"content": msg.Content,
			})

		case "assistant":
			if msg.ToolCall != nil && msg.ToolResult != nil {
				messages = append(messages, map[string]interface{}{
					"role": "assistant",
					"content": []map[string]interface{}{
						{
							"type":  "tool_use",
							"id":    msg.ToolCall.ToolUseID,
							"name":  msg.ToolCall.ToolID,
							"input": msg.ToolCall.Parameters,
						},
					},
				})

				messages = append(messages, map[string]interface{}{
					"role": "user",
					"content": []map[string]interface{}{
						{
							"type":        "tool_result",
							"tool_use_id": msg.ToolResult.ToolUseID,
							"content":     msg.ToolResult.Content,
							"is_error":    msg.ToolResult.IsError,
						},
					},
				})

				if msg.Content != "" {
					messages = append(messages, map[string]interface{}{
						"role": "assistant",
						"content": []map[string]interface{}{
							{
								"type": "text",
								"text": msg.Content,
							},
						},
					})
				}
			} else {
				content := []map[string]interface{}{}
				if msg.Content != "" {
					content = append(content, map[string]interface{}{
						"type": "text",
						"text": msg.Content,
					})
				}
				if len(content) > 0 {
					messages = append(messages, map[string]interface{}{
						"role":    "assistant",
						"content": content,
					})
				}
			}
		}
	}

	return messages
}

func (p *AnthropicProvider) buildTools(availableTools map[string]llmprotocol.ToolInfo, ag *agent.Agent) []map[string]interface{} {
	tools := []map[string]interface{}{}

	for toolID, toolInfo := range availableTools {
		formatschema := toolInfo.Schema
		formatschema["type"] = "object"
		tools = append(tools, map[string]interface{}{
			"name":         toolID,
			"description":  toolInfo.Description,
			"input_schema": formatschema,
		})
	}

	if ag.ServerGeneration {
		tools = append(tools, servergeneration.AnthropicToolSpecs()...)
	}

	if ag.InfraGeneration {
		tools = append(tools, infrageneration.AnthropicToolSpecs()...)
	}

	return tools
}

func (p *AnthropicProvider) sendHTTPRequest(requestBody map[string]interface{}) (map[string]interface{}, error) {
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", "https://api.anthropic.com/v1/messages", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", p.APIKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	var response map[string]interface{}
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return response, nil
}

func (p *AnthropicProvider) parseResponse(response map[string]interface{}) (string, string, string, map[string]interface{}, string, error) {
	content, ok := response["content"].([]interface{})
	if !ok || len(content) == 0 {
		return "", "", "", nil, "", fmt.Errorf("no content in response")
	}

	stopReason := response["stop_reason"].(string)
	responseText := ""
	toolCallID := ""
	toolName := ""
	var toolParams map[string]interface{}

	for _, block := range content {
		blockMap := block.(map[string]interface{})
		blockType := blockMap["type"].(string)

		switch blockType {
		case "text":
			responseText = blockMap["text"].(string)
		case "tool_use":
			toolCallID = blockMap["id"].(string)
			toolName = blockMap["name"].(string)
			toolParams = blockMap["input"].(map[string]interface{})
		}
	}

	return responseText, toolCallID, toolName, toolParams, stopReason, nil
}

func isServerGenerationToolAnthropic(name string) bool {
	return servergeneration.IsServerGenerationTool(name)
}

func isInfraGenerationToolAnthropic(name string) bool {
	return infrageneration.IsInfraGenerationTool(name)
}
