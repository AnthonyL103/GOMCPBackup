package infrageneration

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	agent "github.com/AnthonyL103/GOMCP/Agent"
)

func CollectAWSRequirementsTool(ag *agent.Agent, params map[string]interface{}) (string, bool) {
	_ = ag
	return formatInfraStagePreview(
		ToolCollectAWSRequirements,
		"AWS requirements collected",
		params,
		"Next: call collect_aws_credentials once the requirements are final.",
	), false
}

func CollectAWSCredentialsTool(ag *agent.Agent, params map[string]interface{}) (string, bool) {
	_ = ag
	return formatInfraStagePreview(
		ToolCollectAWSCredentials,
		"AWS credential context collected",
		params,
		"Next: call generate_aws_terraform_iteration with the approved requirements and credential summary.",
	), false
}

func GenerateAWSTerraformTool(ag *agent.Agent, params map[string]interface{}) (string, bool) {
	_ = ag

	// This tool accepts user-supplied Terraform HCL under the "terraform_code" key.
	// If terraform_code is provided, return it back for validation; otherwise
	// signal an error instructing callers to supply HCL. The tool should not
	// fabricate HCL — generation is the responsibility of the model or another tool.
	if v, ok := params["terraform_code"].(string); ok && strings.TrimSpace(v) != "" {
		outParams := map[string]interface{}{
			"terraform_code": v,
		}
		return formatInfraStagePreview(
			ToolGenerateAWSTerraform,
			"Terraform draft received",
			outParams,
			"Next: call validate_aws_terraform_iteration with terraform_code to validate and plan.",
		), false
	}

	return formatInfraStagePreview(
		ToolGenerateAWSTerraform,
		"No terraform_code provided",
		map[string]interface{}{"error": "please provide terraform_code in params; this tool accepts user-supplied HCL only."},
		"",
	), true
}

func ValidateAWSTerraformTool(ag *agent.Agent, params map[string]interface{}) (string, bool) {
	_ = ag

	//terraform fmt -check — ensures idiomatic HCL.
	//terraform init -backend=false — installs providers but avoids remote backends.
	//terraform validate — static syntax + config validation (doesn't require AWS creds for most checks).
	var tf string
	if v, ok := params["terraform_code"].(string); ok && strings.TrimSpace(v) != "" {
		tf = v
	} else if v, ok := params["terraform_summary"].(string); ok && strings.TrimSpace(v) != "" {
		// look for triple-backtick fenced block with optional hcl/terraform tag
		re := regexp.MustCompile("(?s)```(?:hcl|terraform)?\\n(.*?)```")
		if m := re.FindStringSubmatch(v); len(m) > 1 {
			tf = m[1]
		} else {
			// fall back to using the whole summary as the TF content
			tf = v
		}
	} else {
		return formatInfraStagePreview(
			ToolValidateAWSTerraform,
			"Terraform draft validated",
			map[string]interface{}{"error": "no terraform content found in terraform_code or terraform_summary"},
			"Next: call deploy_aws_terraform_iteration if you want the preview deploy stub.",
		), true
	}

	// Create temp dir and write main.tf
	dir, err := os.MkdirTemp("", "tfvalidate-*")
	if err != nil {
		return formatInfraStagePreview(
			ToolValidateAWSTerraform,
			"Terraform validation failed: could not create temp dir",
			map[string]interface{}{"error": err.Error()},
			"",
		), true
	}
	defer os.RemoveAll(dir)

	mainPath := filepath.Join(dir, "main.tf")
	if err := os.WriteFile(mainPath, []byte(tf), 0644); err != nil {
		return formatInfraStagePreview(
			ToolValidateAWSTerraform,
			"Terraform validation failed: could not write main.tf",
			map[string]interface{}{"error": err.Error()},
			"",
		), true
	}

	// Run terraform fmt -check to validate formatting
	fmtCmd := exec.Command("terraform", "fmt", "-check")
	fmtCmd.Dir = dir
	fmtOut, fmtErr := fmtCmd.CombinedOutput()

	// Prepare combined result
	result := make(map[string]interface{})
	result["fmt_output"] = string(fmtOut)

	if fmtErr != nil {
		result["error"] = fmtErr.Error()
		return formatInfraStagePreview(
			ToolValidateAWSTerraform,
			"Terraform fmt check failed",
			result,
			"",
		), true
	}

	// Run terraform init -backend=false
	initCmd := exec.Command("terraform", "init", "-backend=false")
	initCmd.Dir = dir
	initOut, initErr := initCmd.CombinedOutput()
	result["init_output"] = string(initOut)

	// If init failed, return its output
	if initErr != nil {
		result["error"] = initErr.Error()
		return formatInfraStagePreview(
			ToolValidateAWSTerraform,
			"Terraform init failed",
			result,
			"",
		), true
	}

	// Run terraform validate
	valCmd := exec.Command("terraform", "validate")
	valCmd.Dir = dir
	valOut, valErr := valCmd.CombinedOutput()
	result["validate_output"] = string(valOut)

	if valErr != nil {
		result["error"] = valErr.Error()
		return formatInfraStagePreview(
			ToolValidateAWSTerraform,
			"Terraform validation failed",
			result,
			"",
		), true
	}

	// Success
	return formatInfraStagePreview(
		ToolValidateAWSTerraform,
		"Terraform draft validated (fmt, init, validate succeeded)",
		result,
		"Next: call deploy_aws_terraform_iteration if you want the preview deploy stub.",
	), false
}

func DeployAWSTerraformTool(ag *agent.Agent, params map[string]interface{}) (string, bool) {
	_ = ag
	_ = params
	return "true", false
}

func formatInfraStagePreview(stageName, summary string, params map[string]interface{}, nextStep string) string {
	paramBytes, err := json.MarshalIndent(params, "", "  ")
	if err != nil {
		paramBytes = []byte(fmt.Sprintf("{\"marshal_error\": %q}", err.Error()))
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("%s\n", summary))
	sb.WriteString(fmt.Sprintf("Stage: %s\n", stageName))
	sb.WriteString("Mode: preview only; no AWS changes were made.\n")
	sb.WriteString("Inputs:\n")
	sb.WriteString("```json\n")
	sb.Write(paramBytes)
	sb.WriteString("\n```\n")
	if nextStep != "" {
		sb.WriteString(nextStep)
		sb.WriteString("\n")
	}
	return sb.String()
}
