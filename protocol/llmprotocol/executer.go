package llmprotocol

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	agent "github.com/AnthonyL103/GOMCP/Agent"
	"github.com/AnthonyL103/GOMCP/chat"
	"github.com/AnthonyL103/GOMCP/infrageneration"
	"github.com/AnthonyL103/GOMCP/server"
	"github.com/AnthonyL103/GOMCP/servergeneration"
)

func ExecuteTool(ag *agent.Agent, tc *chat.ToolCall) (string, bool) {
	if ag == nil {
		panic("Agent does not exist")
	}
	if tc == nil {
		panic("Tool Call does not exist")
	}

	if tc.ToolID == servergeneration.ToolGenerateServerCode {
		return servergeneration.GenerateServerCodeTool(ag, tc.Parameters)
	}

	if tc.ToolID == servergeneration.ToolDeployAndTestTools {
		return servergeneration.DeployAndTestToolsTool(ag, tc.Parameters)
	}

	if tc.ToolID == servergeneration.ToolDeployAndRegister {
		return servergeneration.DeployAndRegisterServerTool(ag, tc.Parameters)
	}

	if tc.ToolID == servergeneration.ToolCleanupServerGeneration {
		return servergeneration.CleanupServerGenerationTool(ag, tc.Parameters)
	}

	if tc.ToolID == servergeneration.ToolDeleteServer {
		return servergeneration.DeleteServerTool(ag, tc.Parameters)
	}

	if tc.ToolID == infrageneration.ToolCollectAWSRequirements {
		return infrageneration.CollectAWSRequirementsTool(ag, tc.Parameters)
	}

	if tc.ToolID == infrageneration.ToolCollectAWSCredentials {
		return infrageneration.CollectAWSCredentialsTool(ag, tc.Parameters)
	}

	if tc.ToolID == infrageneration.ToolGenerateAWSTerraform {
		return infrageneration.GenerateAWSTerraformTool(ag, tc.Parameters)
	}

	if tc.ToolID == infrageneration.ToolValidateAWSTerraform {
		return infrageneration.ValidateAWSTerraformTool(ag, tc.Parameters)
	}

	if tc.ToolID == infrageneration.ToolDeployAWSTerraform {
		return infrageneration.DeployAWSTerraformTool(ag, tc.Parameters)
	}

	// Get the server
	srv, exists := ag.Registry.Servers[tc.ServerID]
	if !exists {
		return fmt.Sprintf("Server '%s' not found", tc.ServerID), true
	}

	// Get the tool
	_, exists = srv.Tools[tc.ToolID]
	if !exists {
		return fmt.Sprintf("Tool '%s' not found in server '%s'", tc.ToolID, tc.ServerID), true
	}

	runtimeConfig := srv.RuntimeConfig

	// Execute external tool
	return executeExternalTool(tc, runtimeConfig)
}

// executeExternalTool makes HTTP request to external server, completely language agnostic
func executeExternalTool(tc *chat.ToolCall, config *server.RuntimeConfig) (string, bool) {

	// Marshal parameters
	jsonData, err := json.Marshal(tc.Parameters) // Fixed: tc.Parameters not tc.params
	if err != nil {
		return fmt.Sprintf("Failed to marshal request: %v", err), true
	}

	// Make HTTP request to handler route
	url := fmt.Sprintf("http://localhost:%d/execute/%s", config.Port, tc.Handler)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Sprintf("Failed to create request: %v", err), true
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Sprintf("Failed to send request to %s: %v", url, err), true
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Sprintf("Failed to read response: %v", err), true
	}

	if resp.StatusCode != 200 {
		return fmt.Sprintf("Server error (status %d): %s", resp.StatusCode, string(body)), true
	}

	return string(body), false
}
