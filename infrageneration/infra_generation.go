package infrageneration

type GenerationStage string

const (
	StageInit                   GenerationStage = "init"
	StageCollectAWSRequirements GenerationStage = "collect_aws_requirements"
	StageCollectAWSCredentials  GenerationStage = "collect_aws_credentials"
	StageGenerateAWSTerraform   GenerationStage = "generate_aws_terraform_iteration"
	StageValidateAWSTerraform   GenerationStage = "validate_aws_terraform_iteration"
	StageDeployAWSTerraform     GenerationStage = "deploy_aws_terraform_iteration"
)

var (
	validresourcetypes = []string{"ec2", "s3", "rds", "lambda", "vpc", "iam"}
)

type GeneratedResource struct {
	Type       string               `json:"type"`
	Identifier string               `json:"identifier"`
	Properties *[]GeneratedProperty `json:"properties,omitempty"`
}

type GeneratedProperty struct {
	Key   string      `json:"key"`
	Value interface{} `json:"value"`
	Code  string      `json:"code,omitempty"`
}

type GenerationResult struct {
	Stage     GenerationStage     `json:"stage"`
	Resources []GeneratedResource `json:"resources"`
	Messages  []string            `json:"messages"`
}
