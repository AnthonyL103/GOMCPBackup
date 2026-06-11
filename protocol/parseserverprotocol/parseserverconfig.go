// protocol/parseserverprotocol/parseserverprotocol.go
package parseserverprotocol

import (
	"fmt"
	"os"
	"gopkg.in/yaml.v3"

	"github.com/AnthonyL103/GOMCP/server"
	"github.com/AnthonyL103/GOMCP/tool"
)

// ServerConfig represents the YAML server configuration
type ServerConfig struct {
	ServerID    string       `yaml:"server_id"`
	Description string       `yaml:"description"`
	Tools       []ToolConfig `yaml:"tools"`
	Runtime     RuntimeConfigYAML `yaml:"runtime"` // YAML version
}

// RuntimeConfigYAML for deserializing from YAML
type RuntimeConfigYAML struct {
	Type    string   `yaml:"type"`
	Command string   `yaml:"command"`
	Args    []string `yaml:"args"`
	Port    int      `yaml:"port"`
}

// ToolConfig represents a tool in the YAML configuration
type ToolConfig struct {
	ToolID      string            `yaml:"tool_id"`
	Description string            `yaml:"description"`
	Handler     string            `yaml:"handler"`
	InputSchema InputSchemaConfig `yaml:"input_schema"`
}

// InputSchemaConfig represents the input schema in YAML
type InputSchemaConfig struct {
	Properties map[string]PropertyConfig `yaml:"properties"`
	Required   []string                  `yaml:"required"`
}

// PropertyConfig represents a property schema in YAML
type PropertyConfig struct {
	Type        string                     `yaml:"type"`
	Description string                     `yaml:"description"`
	Items       *PropertyConfig            `yaml:"items,omitempty"`
	Properties  map[string]PropertyConfig  `yaml:"properties,omitempty"`
	Required    []string                   `yaml:"required,omitempty"`
}

// convertPropertyConfig recursively converts PropertyConfig to PropertySchema
func convertPropertyConfig(prop PropertyConfig) tool.PropertySchema {
	schema := tool.PropertySchema{
		Type:        prop.Type,
		Description: prop.Description,
		Required:    prop.Required,
	}
	
	// Recursively convert Items if present (for arrays)
	if prop.Items != nil {
		converted := convertPropertyConfig(*prop.Items)
		schema.Items = &converted
	}
	
	// Recursively convert nested Properties if present (for objects)
	if len(prop.Properties) > 0 {
		schema.Properties = make(map[string]tool.PropertySchema)
		for name, nestedProp := range prop.Properties {
			schema.Properties[name] = convertPropertyConfig(nestedProp)
		}
	}
	
	return schema
}

// validateServerConfig validates the server configuration
func validateServerConfig(config *ServerConfig) error {
	if config.ServerID == "" {
		return fmt.Errorf("server_id cannot be empty")
	}
	if config.Description == "" {
		return fmt.Errorf("description cannot be empty")
	}
	if len(config.Tools) == 0 {
		return fmt.Errorf("at least one tool must be defined")
	}
	// Validate runtime
	if config.Runtime.Type == "" {
		return fmt.Errorf("runtime.type cannot be empty")
	}
	if config.Runtime.Command == "" {
		return fmt.Errorf("runtime.command cannot be empty")
	}
	return nil
}

func ParseServerConfig(filePath string) (*server.MCPServer, *server.RuntimeConfig, error) {
	// Read and unmarshal YAML
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read server config at %s: %w", filePath, err)
	}

	var config ServerConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, nil, fmt.Errorf("failed to parse YAML at %s: %w", filePath, err)
	}

	// Validate server-level config
	if err := validateServerConfig(&config); err != nil {
		return nil, nil, fmt.Errorf("invalid server config at %s: %w", filePath, err)
	}

	handlerSet := make(map[string]string)

	// Create tools
	tools := make([]*tool.Tool, 0, len(config.Tools))
	for _, tc := range config.Tools {
		if _, exists := handlerSet[tc.Handler]; exists {
			return nil, nil, fmt.Errorf("duplicate handler '%s' found in tools at %s", tc.Handler, filePath)
		}
		handlerSet[tc.Handler] = tc.ToolID

		props := make(map[string]tool.PropertySchema)
		for name, prop := range tc.InputSchema.Properties {
			props[name] = convertPropertyConfig(prop)
		}

		schema := tool.JSONSchema{
			Properties: props,
			Required:   tc.InputSchema.Required,
		}

		cleanToolID, cleanDesc, cleanHandler, cleanSchema, err := tool.ValidateToolConfig(
			tc.ToolID,
			tc.Description,
			tc.Handler,
			schema,
		)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid tool '%s': %w", tc.ToolID, err)
		}

		t := &tool.Tool{
			ToolID:      cleanToolID,
			Description: cleanDesc,
			InputSchema: cleanSchema,
			Handler:     cleanHandler,
		}

		tools = append(tools, t)
	}

	// Build RuntimeConfig from parsed YAML
	runtimeConfig := &server.RuntimeConfig{
		Type:    config.Runtime.Type,
		Command: config.Runtime.Command,
		Args:    config.Runtime.Args,
		Port:    config.Runtime.Port,
	}

	// Create server with runtime config
	mcpServer := server.NewMCPServer(
		config.ServerID,
		config.Description,
		tools,
		runtimeConfig,
	)

	return mcpServer, runtimeConfig, nil
}