package setup

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/povarna/generative-ai-agents/eval-agent/internal/aggregator"
	"github.com/povarna/generative-ai-agents/eval-agent/internal/config"
	"github.com/povarna/generative-ai-agents/eval-agent/internal/executor"
	"github.com/povarna/generative-ai-agents/eval-agent/internal/judge"
	"github.com/povarna/generative-ai-agents/eval-agent/internal/llm"
	"github.com/povarna/generative-ai-agents/eval-agent/internal/llm/bedrock"
	"github.com/povarna/generative-ai-agents/eval-agent/internal/llm/gpt"
	"github.com/povarna/generative-ai-agents/eval-agent/internal/prechecks"
	"github.com/rs/zerolog"
)

type Config struct {
	AWSRegion          string
	ClaudeModelID      string
	OpenAIKey          string
	OpenAIModelID      string
	DefaultProvider    string
	PrecheckWeight     float64
	LLMJudgeWeight     float64
	EarlyExitThreshold float64
}

type Dependencies struct {
	Executor      *executor.Executor
	JudgeExecutor *executor.JudgeExecutor
	Logger        *zerolog.Logger
}

func LoadConfig() *Config {
	return &Config{
		AWSRegion:          getEnv("AWS_REGION", "us-east-1"),
		ClaudeModelID:      getEnv("CLAUDE_MODEL_ID", ""),
		OpenAIKey:          getEnv("OPEN_AI_KEY", ""),
		OpenAIModelID:      getEnv("OPEN_AI_MODEL_ID", ""),
		DefaultProvider:    getEnv("DEFAULT_LLM_PROVIDER", "bedrock"),
		PrecheckWeight:     getEnvFloat("PRECHECK_WEIGHT", 0.3),
		LLMJudgeWeight:     getEnvFloat("LLM_JUDGE_WEIGHT", 0.7),
		EarlyExitThreshold: getEnvFloat("EARLY_EXIT_THRESHOLD", 0.2),
	}
}

func Wire(ctx context.Context, cfg *Config, logger *zerolog.Logger) (*Dependencies, error) {
	llmClient, err := createLLMClient(ctx, cfg.DefaultProvider, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create Bedrock client: %w", err)
	}

	// PreChecks
	stageRunner := prechecks.NewStageRunner([]prechecks.Checker{
		&prechecks.LengthChecker{},
		&prechecks.OverlapChecker{MinOverlapThreshold: 0.3},
		&prechecks.FormatChecker{},
	})

	// Load judges configuration from YAML
	judgesConfig, err := config.LoadJudgesConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load judges config: %w", err)
	}

	// Create judge pool and build judges from config
	judgePool := judge.NewJudgePool(llmClient, logger)
	judges, err := judgePool.BuildFromConfig(judgesConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to build judges from config: %w", err)
	}

	// Create judge runner with config-driven judges
	judgeRunner := judge.NewJudgeRunner(judges, logger)

	// Judge factory for single judge execution (used by JudgeExecutor)
	judgeFactory := judge.NewJudgeFactory(llmClient, logger)

	// Aggregator
	agg := aggregator.NewAggregator(aggregator.Weights{
		PreChecks: cfg.PrecheckWeight,
		LLMJudge:  cfg.LLMJudgeWeight,
	}, logger)

	// Executors
	exec := executor.NewExecutor(stageRunner, judgeRunner, agg, cfg.EarlyExitThreshold, logger)
	judgeExec := executor.NewJudgeExecutor(judgeFactory, logger)

	return &Dependencies{
		Executor:      exec,
		JudgeExecutor: judgeExec,
		Logger:        logger,
	}, nil

}

func getEnv(key string, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		value = defaultValue
	}

	return value
}

func getEnvFloat(key string, defaultValue float64) float64 {
	valueStr := os.Getenv(key)
	value, err := strconv.ParseFloat(valueStr, 64)
	if err != nil {
		value = defaultValue
	}

	return value
}

func createLLMClient(ctx context.Context, provider string, cfg *Config) (llm.LLMClient, error) {
	switch provider {
	case "bedrock":
		return bedrock.NewClient(ctx, cfg.AWSRegion, cfg.ClaudeModelID)
	case "openai":
		return gpt.NewClient(cfg.OpenAIKey, cfg.OpenAIModelID)
	default:
		return bedrock.NewClient(ctx, cfg.AWSRegion, cfg.ClaudeModelID)
	}
}
