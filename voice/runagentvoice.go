package voicechat

import (
	"log"

	agent "github.com/AnthonyL103/GOMCP/Agent"
	"github.com/AnthonyL103/GOMCP/chat"
	"github.com/AnthonyL103/GOMCP/transport"
)

type VoiceChatParser struct {
	chatSession *chat.Chat
	agent       *agent.Agent
	provider    transport.Provider

	voiceState   *VoiceSessionState
	runtimeState *RuntimeExecutionState
	policy       InterruptPolicy
}

func NewVoiceChatParser(chatSession *chat.Chat, ag *agent.Agent, provider transport.Provider) *VoiceChatParser {
	policy := DefaultInterruptPolicy()
	return &VoiceChatParser{
		chatSession:  chatSession,
		agent:        ag,
		provider:     provider,
		voiceState:   NewVoiceSessionState(policy),
		runtimeState: NewRuntimeExecutionState(),
		policy:       policy,
	}
}

func GetLastMessage(c *chat.Chat) *chat.Message {
	messages := c.GetMessages()
	if len(messages) == 0 {
		return nil
	}
	return &messages[len(messages)-1]
}

func (p *VoiceChatParser) Start() {
	// This is where you'd integrate with the actual voice recognition system

	for {

		lastMsg := GetLastMessage(p.chatSession)
		if lastMsg != nil && lastMsg.Role == "user" {
			// Simulate receiving a voice input as text for now
		}

		if lastMsg.Content == "exit" || lastMsg.Content == "quit" {
			log.Println("Shutting down...")
			break
		}
	}
}
