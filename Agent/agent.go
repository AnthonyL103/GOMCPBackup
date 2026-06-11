package agent

import (
	"fmt"
	"strings"

	"github.com/AnthonyL103/GOMCP/registry"
	"github.com/AnthonyL103/GOMCP/server"
	"github.com/AnthonyL103/GOMCP/tool"
)

func isString(val interface{}) bool {
	_, ok := val.(string)
	return ok
}

func isInt(val interface{}) bool {
	_, ok := val.(int)
	return ok
}

func isFloat(val interface{}) bool {
	_, ok := val.(float32)
	return ok
}

type AgentDetails struct {
	AgentID     string
	Description string
	ServerCount int
	ToolCount   int
	// ServerTools maps ServerID -> map of ToolName -> Tool
	ServerTools      map[string]map[string]*tool.Tool
	ServerGeneration bool
	VoiceChat        bool
	InfraGeneration  bool
}

type LLMConfig struct {
	APIKey      string
	Model       string
	Temperature float32
	MaxTokens   int
}

type Agent struct {
	AgentID          string
	Description      string
	Registry         *registry.Registry
	LLMConfig        *LLMConfig
	ServerGeneration bool
	VoiceChat        bool
	InfraGeneration  bool
	// Remote response / tooling settings
	EnableRemoteResponses bool
	ToolingOutput         bool
	RemoteRoute           string
}

// return list of valid models for the user
func getModelList() []string {
	return []string{
		"gpt-4o", "gpt-4o-mini", "gpt-4-turbo", "o1-preview", "o1-mini",
		"claude-opus-4-5-20251101", "claude-sonnet-4-5-20250929", "claude-haiku-4-5-20251001",
	}
}

func validateLLMConfig(LLMConfig *LLMConfig) {
	if LLMConfig == nil {
		panic("LLMConfig cannot be nil")
	}

	if LLMConfig.APIKey == "" || !isString(LLMConfig.APIKey) {
		panic("LLMConfig.APIKey is required and must be a non-empty string")
	}
	if LLMConfig.Model == "" || !isString(LLMConfig.Model) {
		panic("LLMConfig.Model is required and must be a non-empty string")
	}

	if LLMConfig.Temperature == 0 || !isFloat(LLMConfig.Temperature) {
		panic("LLMConfig.Temperature is required and must be an float")
	}

	if LLMConfig.MaxTokens == 0 || !isInt(LLMConfig.MaxTokens) {
		panic("LLMConfig.MaxTokens is required and must be an int")
	}

	validModels := map[string]bool{
		// OpenAI GPT models (current)
		"gpt-4o":              true,
		"gpt-4o-mini":         true,
		"gpt-4-turbo":         true,
		"gpt-4-turbo-preview": true,
		"o1-preview":          true,
		"o1-mini":             true,
		"gpt-3.5-turbo":       true, // legacy but still supported

		// Claude 4.5 models (latest)
		"claude-opus-4-5-20251101":   true,
		"claude-sonnet-4-5-20250929": true,
		"claude-haiku-4-5-20251001":  true,

		// Claude 3.5 models (legacy)
		"claude-3-5-sonnet-20241022": true,
		"claude-3-5-haiku-20241022":  true,
	}

	if !validModels[LLMConfig.Model] {
		panic(fmt.Sprintf("Invalid model '%s'. Supported models: %v", LLMConfig.Model, getModelList()))
	}
}

func NewAgent(
	agentID string,
	description string,
	registry *registry.Registry,
	LLMConfig *LLMConfig,
	serverGeneration bool,
	voiceChat bool,
	infraGeneration bool,
	enableRemoteResponses bool,
	toolingOutput bool,
	remoteRoute string,

) *Agent {
	agentID = strings.TrimSpace(agentID)
	description = strings.TrimSpace(description)

	if agentID == "" || !isString(agentID) {
		panic("AgentID is required and must be a non-empty string")
	}
	if description == "" || !isString(description) {
		panic("Description is required and must be a non-empty string")
	}

	if registry == nil {
		panic("Registry cannot be nil")
	}

	validateLLMConfig(LLMConfig)

	return &Agent{
		AgentID:               agentID,
		Description:           description,
		Registry:              registry,
		LLMConfig:             LLMConfig,
		ServerGeneration:      serverGeneration,
		VoiceChat:             voiceChat,
		InfraGeneration:       infraGeneration, // default to false, can be set via config
		EnableRemoteResponses: enableRemoteResponses,
		ToolingOutput:         toolingOutput,
		RemoteRoute:           remoteRoute,
	}
}

func (a *Agent) GetAgentDetails(agent *Agent) *AgentDetails {
	if a == nil {
		panic("Agent cannot be nil")
	}

	servers := make(map[string]*server.MCPServer)
	serverTools := make(map[string]map[string]*tool.Tool)
	serverGeneration := a.ServerGeneration
	toolCount := 0

	for _, server := range a.Registry.Servers {
		servers[server.ServerID] = server
		serverTools[server.ServerID] = make(map[string]*tool.Tool)

		for toolName, tool := range server.Tools {
			serverTools[server.ServerID][toolName] = tool
			toolCount++
		}
	}

	return &AgentDetails{
		AgentID:          a.AgentID,
		Description:      a.Description,
		ServerCount:      len(servers),
		ToolCount:        toolCount,
		ServerTools:      serverTools,
		ServerGeneration: serverGeneration,
		VoiceChat:        a.VoiceChat,
		InfraGeneration:  a.InfraGeneration,
	}
}
