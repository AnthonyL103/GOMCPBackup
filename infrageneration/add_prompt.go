package infrageneration

type ToolDefinition struct {
	Name        string
	Description string
	InputSchema map[string]interface{}
}

const awsInfraSystemPrompt = `You are an AWS infrastructure planning assistant for a local Terraform builder.

Your job is to help the user turn a vague idea into a safe, explicit, reviewable AWS Terraform plan.

Rules:
- Stay AWS-only for now.
- Use separate tools for separate phases.
- Start with requirements discovery and do not guess missing critical details.
- Treat the frontend project form as already completed and assume the backend received the full brief: project name, description, project type, data storage, background tasks, audience, expected usage, reliability needs, sensitive data, user location, and optional domain.
- Do not re-ask any of those frontend fields unless one is actually missing from the payload.
- Ask follow-up questions only for AWS-specific gaps that the form did not cover.
- If anything AWS-specific is ambiguous, ask again rather than assuming.
- Prefer plain language first, then show technical details.
- Always include a short summary of what you understood before moving to the next phase.
- Treat credentials as placeholders if the user says dummy values for now, but still require the credential shape.
- Generate an initial Terraform draft, validate it, explain it in beautiful markdown, then loop on revisions until the user confirms.
- Generate an initial Terraform draft, validate it, and then return a beautiful markdown explanation of the validated draft before looping on revisions.
- Do not deploy until the user explicitly approves the final draft.
- When deploying, mirror the step-by-step progress clearly and stop immediately if any step fails.

Discovery checklist:
- Which AWS region should be used?
- Is this dev, staging, or prod?
- What deployment shape fits the already-approved app brief?
- What resources are required: networking, compute, database, storage, DNS, secrets, monitoring, backup?
- What security, compliance, and budget constraints apply?
- What naming, tagging, and ownership rules should be used?
- What existing AWS assets already exist?
- What should never be created or opened publicly?

Output style:
- Be concise but thorough.
- Summarize every answer back to the user.
- Make the markdown explanation easy for a non-engineer to understand.
- Keep the plan reversible and safe by default.
`

const (
	ToolCollectAWSRequirements = "collect_aws_requirements"
	ToolCollectAWSCredentials  = "collect_aws_credentials"
	ToolGenerateAWSTerraform   = "generate_aws_terraform_iteration"
	ToolValidateAWSTerraform   = "validate_aws_terraform_iteration"
	ToolDeployAWSTerraform     = "deploy_aws_terraform_iteration"
)

// GetAWSInfraSystemPrompt returns the AWS-only system prompt used for the infra flow.
func GetAWSInfraSystemPrompt() string {
	return awsInfraSystemPrompt
}

// SharedAWSToolDefinitions centralizes the infra phase tool contracts.
func SharedAWSToolDefinitions() []ToolDefinition {
	return []ToolDefinition{
		{
			Name:        ToolCollectAWSRequirements,
			Description: "Collect and refine AWS infrastructure requirements using the frontend form as pre-collected context for product basics. Only ask for AWS-specific gaps, constraints, and guardrails needed to draft Terraform safely.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"project_goal":        map[string]interface{}{"type": "string", "description": "Plain-language description of what the user wants to build, including the frontend form answers if available"},
					"environment":         map[string]interface{}{"type": "string", "description": "dev, staging, prod, or similar"},
					"region":              map[string]interface{}{"type": "string", "description": "Primary AWS region"},
					"workload_type":       map[string]interface{}{"type": "string", "description": "Web app, API, database, worker, static site, batch job, etc."},
					"networking":          map[string]interface{}{"type": "string", "description": "Public/private access, ingress, egress, VPC, subnets, DNS expectations"},
					"security":            map[string]interface{}{"type": "string", "description": "Encryption, IAM, secrets, exposure limits, compliance constraints"},
					"data_services":       map[string]interface{}{"type": "string", "description": "RDS, DynamoDB, S3, EFS, cache, or none"},
					"observability":       map[string]interface{}{"type": "string", "description": "Logs, metrics, alarms, tracing, retention needs"},
					"backup_and_recovery": map[string]interface{}{"type": "string", "description": "Backups, recovery target, retention, disaster recovery expectations"},
					"cost_guardrails":     map[string]interface{}{"type": "string", "description": "Budget or cost sensitivity constraints"},
					"naming_and_tags":     map[string]interface{}{"type": "string", "description": "Naming conventions, tags, owner metadata, project labels"},
					"existing_aws_assets": map[string]interface{}{"type": "string", "description": "Anything already in AWS that should be reused or avoided"},
					"open_questions":      map[string]interface{}{"type": "array", "items": map[string]interface{}{"type": "string"}, "description": "Remaining questions the assistant still needs answered"},
				},
				"required": []string{"project_goal"},
			},
		},
		{
			Name:        ToolCollectAWSCredentials,
			Description: "Capture AWS credential context separately from the frontend product form. Use placeholder values if needed, but still require the full shape needed for later validation and deploy steps.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"credential_mode":      map[string]interface{}{"type": "string", "description": "profile, access keys, assumed role, dummy, or sandbox"},
					"aws_profile":          map[string]interface{}{"type": "string", "description": "Local AWS profile name if used"},
					"account_id":           map[string]interface{}{"type": "string", "description": "AWS account identifier"},
					"region":               map[string]interface{}{"type": "string", "description": "AWS region to use"},
					"role_arn":             map[string]interface{}{"type": "string", "description": "Optional role ARN if assuming a role"},
					"dummy_credentials_ok": map[string]interface{}{"type": "boolean", "description": "Whether placeholder values are intentionally being used for now"},
					"notes":                map[string]interface{}{"type": "string", "description": "Any credential caveats or constraints"},
				},
				"required": []string{"credential_mode", "dummy_credentials_ok"},
			},
		},
		{
			Name:        ToolGenerateAWSTerraform,
			Description: "Generate the first Terraform iteration for AWS from the collected requirements and credentials context. Do not rediscover product basics that were already captured in the frontend form.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"requirements_summary": map[string]interface{}{"type": "string", "description": "Compact summary of the approved requirements"},
					"credentials_summary":  map[string]interface{}{"type": "string", "description": "Compact summary of the AWS credential context"},
					"generation_notes":     map[string]interface{}{"type": "string", "description": "Revision notes, assumptions, and guardrails for the first draft"},
					"preferred_template":   map[string]interface{}{"type": "string", "description": "Optional template choice such as static site, API, database-backed app, or worker"},
				},
				"required": []string{"requirements_summary"},
			},
		},
		{
			Name:        ToolValidateAWSTerraform,
			Description: "Run formatting, validation, and plan-style checks against the generated Terraform draft, then return the results plus a beautiful markdown explanation of every meaningful part of the validated draft. Focus on AWS implementation details rather than re-questioning the project brief.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"terraform_summary": map[string]interface{}{"type": "string", "description": "Summary of the current Terraform draft"},
					"validation_scope":  map[string]interface{}{"type": "string", "description": "What should be checked in this pass"},
					"known_risks":       map[string]interface{}{"type": "string", "description": "Known concerns that should be rechecked"},
				},
				"required": []string{"terraform_summary"},
			},
		},
		{
			Name:        ToolDeployAWSTerraform,
			Description: "Execute the approved AWS deployment flow and mirror the progress back to the user step by step. The deployment step should assume the requirements were already finalized upstream.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"approval_status":         map[string]interface{}{"type": "string", "description": "Explicit confirmation that the user approved deployment"},
					"deployment_target":       map[string]interface{}{"type": "string", "description": "Account, stack, or environment being deployed to"},
					"last_validation_summary": map[string]interface{}{"type": "string", "description": "Most recent successful validation summary"},
					"terminal_mirroring":      map[string]interface{}{"type": "boolean", "description": "Whether terminal progress should be echoed back to the frontend"},
				},
				"required": []string{"approval_status", "deployment_target"},
			},
		},
	}
}

// OpenAIToolSpecs converts the AWS infra tool definitions into OpenAI function specs.
func OpenAIToolSpecs() []map[string]interface{} {
	defs := SharedAWSToolDefinitions()
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

// AnthropicToolSpecs converts the AWS infra tool definitions into Anthropic tool specs.
func AnthropicToolSpecs() []map[string]interface{} {
	defs := SharedAWSToolDefinitions()
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

// IsInfraGenerationTool reports whether the name belongs to the AWS infra flow.
func IsInfraGenerationTool(name string) bool {
	for _, def := range SharedAWSToolDefinitions() {
		if def.Name == name {
			return true
		}
	}
	return false
}
