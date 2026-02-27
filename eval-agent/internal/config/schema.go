package config

// Config represents the complete evaluation configuration
type Config struct {
	LLMJudge       LLMJudgeConfig                `yaml:"llm_judge"`
	EvaluationType string                        `yaml:"evaluation_type"`
	Judges         map[string]JudgeConfiguration `yaml:"judges"`
	Aggregation    AggregationConfig             `yaml:"aggregation"`
}

// LLMJudgeConfig contains the global evaluation prompt and annotation labels
type LLMJudgeConfig struct {
	Prompt           string            `yaml:"prompt"`
	AnnotationLabels []AnnotationLabel `yaml:"annotation_labels"`
}

// AnnotationLabel defines a score label for human annotation (used in correlation analysis)
type AnnotationLabel struct {
	Score       int    `yaml:"score"`
	Label       string `yaml:"label"`
	Description string `yaml:"description"`
}

// JudgeConfig contains per-judge configuration
type JudgeConfig struct {
	Enabled        bool              `yaml:"enabled"`
	Weight         float64           `yaml:"weight"`
	PromptOverride string            `yaml:"prompt_override"`
	RequireContext bool              `yaml:"require_context"`
	ModelParams    ModelParamsConfig `yaml:"model_params"`
}

// ModelParamsConfig contains Bedrock model parameters for judge calls
type ModelParamsConfig struct {
	MaxTokens   int     `yaml:"max_tokens"`
	Temperature float64 `yaml:"temperature"`
	UseRetry    bool    `yaml:"use_retry"`
}

// AggregationConfig contains weights for aggregating precheck and judge scores
type AggregationConfig struct {
	PrecheckWeight float64 `yaml:"precheck_weight"`
	JudgeWeight    float64 `yaml:"judge_weight"`
}
