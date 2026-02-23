package judge

import (
	"context"
	"sync"

	"github.com/povarna/generative-ai-with-go/eval-agent/internal/models"
	"github.com/rs/zerolog"
)

type JudgeRunner struct {
	Judges []Judge
	logger *zerolog.Logger
}

func NewJudgeRunner(judges []Judge, logger *zerolog.Logger) *JudgeRunner {
	return &JudgeRunner{
		Judges: judges,
		logger: logger,
	}
}

func (c *JudgeRunner) Run(ctx context.Context, evaluationContext models.EvaluationContext) []models.StageResult {
	results := make(chan models.StageResult, len(c.Judges))
	var wg sync.WaitGroup

	for _, judge := range c.Judges {
		wg.Add(1)
		go func(j Judge) {
			defer wg.Done()
			results <- j.Evaluate(ctx, evaluationContext)
		}(judge)
	}

	wg.Wait()
	close(results)

	var stageResults []models.StageResult
	for result := range results {
		stageResults = append(stageResults, result)
	}
	c.logger.Debug().Int("judgeCount", len(stageResults)).Msg("all judges completed")

	return stageResults
}
