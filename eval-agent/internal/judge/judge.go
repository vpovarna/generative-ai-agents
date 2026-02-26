package judge

import (
	"context"

	"github.com/povarna/generative-ai-with-go/eval-agent/internal/bedrock"
	"github.com/povarna/generative-ai-with-go/eval-agent/internal/models"
)

type Judge interface {
	Evaluate(ctx context.Context, evaluationContext models.EvaluationContext) models.StageResult
}

// LLMClient is an interface for invoking LLM models
// This allows mocking in tests without making real API calls
type LLMClient interface {
	InvokeModel(ctx context.Context, request bedrock.ClaudeRequest) (*bedrock.ClaudeResponse, error)
	InvokeModelWithRetry(ctx context.Context, request bedrock.ClaudeRequest) (*bedrock.ClaudeResponse, error)
}
