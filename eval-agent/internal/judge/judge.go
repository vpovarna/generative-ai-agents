package judge

import (
	"context"

	"github.com/povarna/generative-ai-with-go/eval-agent/internal/models"
)

type Judge interface {
	Evaluate(ctx context.Context, evaluationContext models.EvaluationContext) models.StageResult
}
