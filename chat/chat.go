package chat

import (
    "time"
)

// Message represents a single turn in the conversation
type Message struct {
    Role       string    `json:"role"`      // "user", "assistant"
    Content    string    `json:"content"`   // Final response text
    Timestamp  time.Time `json:"timestamp"`
    
    // Agent's complete turn (optional - only if tool was used)
    ToolCall   *ToolCall   `json:"tool_call,omitempty"`   // What tool the agent called
    ToolResult *ToolResult `json:"tool_result,omitempty"` // What the tool returned
}

//no specific object declaration for params as these are stored in chat and past context must be sent 
//and stored in a specific format for different providers.

type ToolCall struct {
    ServerID   string                 `json:"server_id"`
    ToolID     string                 `json:"tool_id"`     // Matches your tool registry
    Handler    string                 `json:"handler"`
    Parameters map[string]interface{} `json:"parameters"`
    Reasoning  string                 `json:"reasoning"`   // Agent's explanation
    ToolUseID  string                 `json:"tool_use_id"`
}

type ToolResult struct {
    ServerID string `json:"server_id"`
    ToolID  string `json:"tool_id"`  // Which tool was executed
    Content string `json:"content"`   // Tool output
    IsError bool   `json:"is_error"`  // Did execution fail?
    ToolUseID  string `json:"tool_use_id"`
}

// Chat stores the conversation history
type Chat struct {
    ChatID      string    `json:"chat_id"`
    Messages    []Message `json:"messages"`
    MaxMessages int       `json:"max_messages"` // Sliding window (0 = unlimited)
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
}

func NewChat(chatID string, maxMessages int) *Chat {
    if maxMessages < 0 {
        maxMessages = 0 // 0 = no limit
    }
    
    return &Chat{
        ChatID:      chatID,
        Messages:    []Message{},
        MaxMessages: maxMessages,
        CreatedAt:   time.Now(),
        UpdatedAt:   time.Now(),
    }
}

func (c *Chat) AddUserMessage(content string) {
    c.Messages = append(c.Messages, Message{
        Role:      "user",
        Content:   content,
        Timestamp: time.Now(),
    })
    c.UpdatedAt = time.Now()
    c.trimIfNeeded()
}

// AddAssistantMessage adds a complete agent turn
// If the agent used a tool, include toolCall and toolResult
func (c *Chat) AddAssistantMessage(content string, toolCall *ToolCall, toolResult *ToolResult) {
    c.Messages = append(c.Messages, Message{
        Role:       "assistant",
        Content:    content,
        ToolCall:   toolCall,
        ToolResult: toolResult,
        Timestamp:  time.Now(),
    })
    c.UpdatedAt = time.Now()
    c.trimIfNeeded()
}

func (c *Chat) GetMessages() []Message {
    return c.Messages
}

func (c *Chat) Clear() {
    c.Messages = []Message{}
    c.UpdatedAt = time.Now()
}

func (c *Chat) MessageCount() int {
    return len(c.Messages)
}

func (c *Chat) trimIfNeeded() {
    if c.MaxMessages > 0 && len(c.Messages) > c.MaxMessages {
        c.Messages = c.Messages[len(c.Messages)-c.MaxMessages:]
    }
}

func (c *Chat) GetRecentMessages(n int) []Message {
    if n <= 0 || n >= len(c.Messages) {
        return c.Messages
    }
    return c.Messages[len(c.Messages)-n:]
}