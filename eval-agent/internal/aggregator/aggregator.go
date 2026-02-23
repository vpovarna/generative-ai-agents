package aggregator

import (
	"github.com/povarna/generative-ai-with-go/eval-agent/internal/models"
	"github.com/rs/zerolog"
)

type Weights struct {
	PreChecks float64
	LLMJudge  float64
}

type Aggregator struct {
	Weights Weights
	logger  *zerolog.Logger
}

func NewAggregator(weights Weights, logger *zerolog.Logger) *Aggregator {
	return &Aggregator{
		Weights: weights,
		logger:  logger,
	}
}

func (a *Aggregator) Aggregate(id string, stage1 []models.StageResult, stage2 []models.StageResult) models.EvaluationResult {
	result := models.EvaluationResult{
		ID:     id,
		Stages: append(stage1, stage2...),
	}

	stage1Score, stage2Score := 0.0, 0.0

	for _, stage := range stage1 {
		stage1Score += stage.Score
	}

	for _, stage := range stage2 {
		stage2Score += stage.Score
	}

	if len(stage1) == 0 || len(stage2) == 0 {
		result.Verdict = models.VerdictFail
		return result
	}

	stage1Avg := stage1Score / float64(len(stage1))
	stage2Avg := stage2Score / float64(len(stage2))

	confidence := (stage1Avg * a.Weights.PreChecks) + (stage2Avg * a.Weights.LLMJudge)

	result.Confidence = confidence
	result.Verdict = a.calculateVerdict(confidence)

	a.logger.
		Info().
		Float64("confidence", confidence).
		Str("verdict", string(result.Verdict)).
		Msg("aggregation complete")
	return result
}

func (a *Aggregator) calculateVerdict(confidence float64) models.Verdict {
	if confidence > 0.8 {
		return models.VerdictPass
	}
	if confidence > 0.5 {
		return models.VerdictReview
	}
	return models.VerdictFail
}
