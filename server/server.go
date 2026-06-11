package server

import (
	"fmt"
	"strings"

	"github.com/AnthonyL103/GOMCP/tool"
)

func isString(val interface{}) bool {
	_, ok := val.(string)
	return ok
}

type MCPServer struct {
	ServerID string
	Description string
	Tools map[string]*tool.Tool
	RuntimeConfig *RuntimeConfig
}

type RuntimeConfig struct {
	Type    string
	Command string
	Args    []string
	Port    int
}

func NewMCPServer(
	serverID string,
	description string,
	tools []*tool.Tool,
	runtimeconfig *RuntimeConfig,
) *MCPServer {
	serverID = strings.TrimSpace(serverID)
	description = strings.TrimSpace(description)
	//validate the inputs
	if serverID == "" || !isString(serverID) {
		panic("ServerID is required and must be a non-empty string")
	}

	if description == "" || !isString(description) {
		panic("Description is required and must be a non-empty string")
	}

	if len(tools) == 0 {
		panic("At least one tool must be provided")
	}

	
	//ensure that no tools are nil as a safeguard and create tool map
	toolMap := make(map[string]*tool.Tool)
	for _, t := range tools {
		if t == nil {
			panic("Tool cannot be nil")
		}
		toolMap[t.ToolID] = t
	}

	return &MCPServer{
		ServerID:    serverID,
		Description: description,
		Tools:       toolMap,
		RuntimeConfig: runtimeconfig,
	}
}

func (s *MCPServer) AddToolToServer(
	t *tool.Tool,
) *MCPServer{
	if s == nil {
		panic("Server cannot be nil")
	}
	if t == nil {
		panic("Tool cannot be nil")
	}
	if _, exists := s.Tools[t.ToolID]; exists {
		panic(fmt.Sprintf("Tool with id '%s' already exists in the server", t.ToolID))
	}

	s.Tools[t.ToolID] = t
	return s
}

func (s *MCPServer) RemoveToolFromServer(
	toolName string,
) *MCPServer {
	if s == nil {
		panic("Server cannot be nil")
	}
	toolName = strings.TrimSpace(toolName)
	if toolName == "" {
		panic("Tool name cannot be empty")
	}
	if _, exists := s.Tools[toolName]; !exists {
		panic(fmt.Sprintf("Tool with name '%s' does not exist in the server", toolName))
	}
	delete(s.Tools, toolName)
	return s
}

func (s *MCPServer) GetToolFromServer(
	toolName string,
) *tool.Tool {
	if s == nil {
		panic("Server cannot be nil")
	}
	toolName = strings.TrimSpace(toolName)
	if toolName == "" {
		panic("Tool name cannot be empty")
	}

	if t, exists := s.Tools[toolName]; exists {
		return t
	}
	panic(fmt.Sprintf("Tool with name '%s' does not exist in the server", toolName))
}