package tool

import (
	"fmt"
	"strings"
)

type Tool struct {
	ToolID      string
	Description string 
	InputSchema JSONSchema
	Handler     string
}

type JSONSchema struct {
	Properties map[string]PropertySchema `json:"properties"`
	Required   []string					 `json:"required"`
}

type PropertySchema struct {
    Type        string                      `json:"type"`
    Description string                      `json:"description,omitempty"`
    Items       *PropertySchema             `json:"items,omitempty"`       // For arrays
    Properties  map[string]PropertySchema   `json:"properties,omitempty"`  // For nested objects
    Required    []string                    `json:"required,omitempty"`    // For nested objects
}
// ValidateToolConfig validates tool configuration and returns sanitized values
func ValidateToolConfig(toolID, description, handler string, inputSchema JSONSchema) (string, string, string, JSONSchema, error) {
	// Strip whitespace
	toolID = strings.TrimSpace(toolID)
	description = strings.TrimSpace(description)
	handler = strings.TrimSpace(handler)

	// Validation
	if handler == "" {
		return "", "", "", JSONSchema{}, fmt.Errorf("handler cannot be empty")
	}
	if toolID == "" {
		return "", "", "", JSONSchema{}, fmt.Errorf("toolID cannot be empty")
	}
	if description == "" {
		return "", "", "", JSONSchema{}, fmt.Errorf("description cannot be empty")
	}

	// Validate required fields exist in properties
	for _, req := range inputSchema.Required {
		req = strings.TrimSpace(req)
		prop, exists := inputSchema.Properties[req]
		if !exists {
			return "", "", "", JSONSchema{}, fmt.Errorf("required field '%s' not found in properties", req)
		}
		// Validate property has a type
		if strings.TrimSpace(prop.Type) == "" {
			return "", "", "", JSONSchema{}, fmt.Errorf("property '%s' must have a type", req)
		}
	}

	// Validate property types
	validTypes := map[string]bool{
		"string":  true,
		"number":  true,
		"boolean": true,
		"array":   true,
		"object":  true,
		
	}

	for name, prop := range inputSchema.Properties {
		propType := strings.TrimSpace(prop.Type)
		if !validTypes[propType] {
			return "", "", "", JSONSchema{}, fmt.Errorf("property '%s': invalid type '%s'", name, propType)
		}
	}

	// Trim all property names and values
	trimmedProperties := make(map[string]PropertySchema)
	for key, prop := range inputSchema.Properties {
		trimmedKey := strings.TrimSpace(key)
		trimmedProperties[trimmedKey] = PropertySchema{
			Type:        strings.TrimSpace(prop.Type),
			Description: strings.TrimSpace(prop.Description),
			Items:       prop.Items,
			Properties:  prop.Properties,
			Required:    prop.Required,
		}
	}

	// Trim required fields
	trimmedRequired := make([]string, len(inputSchema.Required))
	for i, req := range inputSchema.Required {
		trimmedRequired[i] = strings.TrimSpace(req)
	}

	sanitizedSchema := JSONSchema{
		Properties: trimmedProperties,
		Required:   trimmedRequired,
	}

	return toolID, description, handler, sanitizedSchema, nil
}

func NewTool(
	toolID string,
	description string,
	inputSchema JSONSchema,
	handler string,
) *Tool {
	// Use the validation function
	cleanToolID, cleanDesc, cleanHandler, cleanSchema, err := ValidateToolConfig(toolID, description, handler, inputSchema)
	if err != nil {
		panic(err)
	}

	return &Tool{
		ToolID:      cleanToolID,
		Description: cleanDesc,
		InputSchema: cleanSchema,
		Handler:     cleanHandler,
	}
}
