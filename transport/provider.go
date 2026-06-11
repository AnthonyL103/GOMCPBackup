package transport

import (
	agent "github.com/AnthonyL103/GOMCP/Agent"
	"github.com/AnthonyL103/GOMCP/chat"
)

// Provider handles LLM API communication
type Provider interface {
	// SendRequest sends a message and returns the assistant's response
	// It handles the full cycle: LLM call → tool execution → final response
	SendRequest(chat *chat.Chat, agent *agent.Agent, userMessage string) error

	// GetProviderName returns the provider name (e.g., "openai", "anthropic")
	GetProviderName() string
}
