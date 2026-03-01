package judge

import (
	"context"

	"github.com/povarna/generative-ai-agents/eval-agent/internal/models"
)

type Judge interface {
	Name() string
	Evaluate(ctx context.Context, evaluationContext models.EvaluationContext) models.StageResult
}
