package prechecks

import (
	"github.com/povarna/generative-ai-agents/eval-agent/internal/models"
)

type Checker interface {
	Check(evaluationContext models.EvaluationContext) models.StageResult
}
