package servergeneration

type ToolDefinition struct {
	Name        string
	Description string
	InputSchema map[string]interface{}
}

// SharedServerGenerationToolDefinitions centralizes server-generation tool schemas
// so providers can render them without duplicating large inline blocks.
func SharedServerGenerationToolDefinitions() []ToolDefinition {
	return []ToolDefinition{
		{
			Name:        ToolGenerateServerCode,
			Description: "Generate and validate Go server code for a custom tool server. Auto-includes: net/http, encoding/json, log. Returns a process_id for subsequent deployment steps.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"server_id":          map[string]interface{}{"type": "string", "description": "Unique server identifier (snake_case)"},
					"server_description": map[string]interface{}{"type": "string", "description": "Description of the server's purpose and capabilities"},
					"tools": map[string]interface{}{
						"type":        "array",
						"description": "Array of tool objects to implement. Each must have: tool_id, description, input_schema (JSON schema), and handler_code (Go function body)",
						"items": map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"tool_id":      map[string]interface{}{"type": "string", "description": "Unique tool identifier (snake_case). Will be POST /execute/{tool_id}"},
								"description":  map[string]interface{}{"type": "string", "description": "Clear description of what this tool does"},
								"input_schema": map[string]interface{}{"type": "object", "description": "JSON Schema with 'properties' and 'required' arrays. Input will be decoded from JSON request body."},
								"handler_code": map[string]interface{}{"type": "string", "description": "Go handler code (function body only). Access request body via 'var params map[string]interface{}' with json.Unmarshal. Write JSON response with w.Write."},
							},
							"required": []string{"tool_id", "description", "input_schema", "handler_code"},
						},
					},
					"imports": map[string]interface{}{
						"type":        "array",
						"description": "Optional additional Go imports needed for handler code (e.g. 'crypto/md5', 'strconv', 'strings'). Do not include: net/http, encoding/json, log (auto-included).",
						"items":       map[string]interface{}{"type": "string"},
					},
				},
				"required": []string{"server_id", "server_description", "tools"},
			},
		},
		{
			Name:        ToolDeployAndTestTools,
			Description: "Compile, start the server binary, and verify all tools work with test_params. Must have a test_params object for every tool defined in generate_server_code. On test failure, artifacts are automatically cleaned up. On success, proceed to deploy_and_register_server.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"process_id": map[string]interface{}{"type": "string", "description": "Process ID returned from generate_server_code"},
					"tool_tests": map[string]interface{}{
						"type":        "array",
						"description": "Test cases for each tool. Must match the tools defined in generate_server_code. Each test will POST to /execute/{tool_id} with test_params as JSON body.",
						"items": map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"tool_id":     map[string]interface{}{"type": "string", "description": "Tool ID to test"},
								"test_params": map[string]interface{}{"type": "object", "description": "Test params for the tool"},
							},
							"required": []string{"tool_id", "test_params"},
						},
					},
				},
				"required": []string{"process_id", "tool_tests"},
			},
		},
		{
			Name:        ToolDeployAndRegister,
			Description: "Register the tested server into the agent's registry and start the final server process. The server will remain running and available for tool calls. Clean up source files after this step with cleanup_server_generation.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"process_id": map[string]interface{}{"type": "string", "description": "Process ID from deploy_and_test_tools (after successful tests)"},
				},
				"required": []string{"process_id"},
			},
		},
		{
			Name:        ToolCleanupServerGeneration,
			Description: "Clean up temporary files after server deployment. Removes the Go source file (binary is kept and running). Call this after deploy_and_register_server completes successfully.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"process_id": map[string]interface{}{"type": "string", "description": "Process ID to clean up (returned from generate_server_code)"},
				},
				"required": []string{"process_id"},
			},
		},
		{
			Name:        ToolDeleteServer,
			Description: "Permanently delete a previously generated and deployed server. Stops the running process, removes files, and unregisters from the agent's registry.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"server_id": map[string]interface{}{
						"type":        "string",
						"description": "The unique identifier of the server to delete (e.g., 'csv_processor')",
					},
				},
				"required": []string{"server_id"},
			},
		},
	}
}

func IsServerGenerationTool(name string) bool {
	for _, toolDef := range SharedServerGenerationToolDefinitions() {
		if toolDef.Name == name {
			return true
		}
	}
	return false
}

func OpenAIToolSpecs() []map[string]interface{} {
	defs := SharedServerGenerationToolDefinitions()
	tools := make([]map[string]interface{}, 0, len(defs))
	for _, def := range defs {
		tools = append(tools, map[string]interface{}{
			"type": "function",
			"function": map[string]interface{}{
				"name":        def.Name,
				"description": def.Description,
				"parameters":  def.InputSchema,
			},
		})
	}
	return tools
}

func AnthropicToolSpecs() []map[string]interface{} {
	defs := SharedServerGenerationToolDefinitions()
	tools := make([]map[string]interface{}, 0, len(defs))
	for _, def := range defs {
		tools = append(tools, map[string]interface{}{
			"name":         def.Name,
			"description":  def.Description,
			"input_schema": def.InputSchema,
		})
	}
	return tools
}