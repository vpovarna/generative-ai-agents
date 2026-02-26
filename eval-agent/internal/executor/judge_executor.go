package executor

import (
	"context"
	"errors"

	"github.com/povarna/generative-ai-with-go/eval-agent/internal/judge"
	"github.com/povarna/generative-ai-with-go/eval-agent/internal/models"
	"github.com/rs/zerolog"
)

type JudgeFactory interface {
	Get(judgeName string) (judge.Judge, error)
}

type JudgeExecutor struct {
	judges JudgeFactory
	logger *zerolog.Logger
}

func NewJudgeExecutor(judges JudgeFactory, logger *zerolog.Logger) *JudgeExecutor {
	return &JudgeExecutor{
		judges: judges,
		logger: logger,
	}
}

var ErrJudgeNotFound = errors.New("judge not found")

func (e *JudgeExecutor) Execute(ctx context.Context, judgeName string, threshold float64, evalCtx models.EvaluationContext) (models.EvaluationResult, error) {
	id := evalCtx.RequestID
	e.logger.Info().Str("requestID", id).Msg("starting evaluation")

	result := models.EvaluationResult{
		ID:     id,
		Stages: []models.StageResult{},
	}

	judge, err := e.judges.Get(judgeName)
	if err != nil {
		e.logger.Error().Err(err).Str("judgeName", judgeName).Msg("Judge not found")
		return result, ErrJudgeNotFound
	}

	judgeResponse := judge.Evaluate(ctx, evalCtx)

	result.Stages = append(result.Stages, judgeResponse)
	if judgeResponse.Score > threshold {
		result.Verdict = models.VerdictPass
	} else {
		result.Verdict = models.VerdictFail
	}
	result.Confidence = judgeResponse.Score

	return result, nil
}
