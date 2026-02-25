package judge

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/povarna/generative-ai-with-go/eval-agent/internal/bedrock"
	"github.com/povarna/generative-ai-with-go/eval-agent/internal/models"
	"github.com/rs/zerolog"
)

// CompletenessJudge evaluates whether the model answer fully addresses all distinct
// questions or sub-requests in the user query. It scores 1.0 if every part is covered,
// 0.5 if some are missing or incomplete, and 0.0 if major parts are ignored.
type CompletenessJudge struct {
	llmClient LLMClient
	logger    *zerolog.Logger
}

func NewCompletenessJudge(client LLMClient, logger *zerolog.Logger) *CompletenessJudge {
	return &CompletenessJudge{
		llmClient: client,
		logger:    logger,
	}
}

func (j *CompletenessJudge) Evaluate(ctx context.Context, evaluationContext models.EvaluationContext) models.StageResult {
	now := time.Now()

	result := models.StageResult{
		Name:  "completeness-judge",
		Score: 0.0,
	}

	prompt := j.buildPrompt(evaluationContext)
	resp, err := j.llmClient.InvokeModel(ctx, bedrock.ClaudeRequest{
		Prompt:      prompt,
		MaxTokens:   256,
		Temperature: 0.0, // deterministic
	})

	if err != nil {
		j.logger.Error().Err(err).Str("judge", "completeness-judge").Msg("LLM call failed")

		result.Reason = "Failed to call LLM"
		result.Duration = time.Since(now)
		return result
	}

	var llmResponse judgeResponse
	if err := json.Unmarshal([]byte(resp.Content), &llmResponse); err != nil {
		result.Reason = "Failed to deserialize LLM response"
		result.Duration = time.Since(now)
		return result
	}

	if llmResponse.Score == 0.0 && llmResponse.Reason == "" {
		j.logger.Error().Msg("LLM returned empty score and reason")
		result.Reason = "Invalid LLM response: missing score and reason"
		result.Duration = time.Since(now)
		return result
	}

	result.Score = llmResponse.Score
	result.Reason = llmResponse.Reason
	result.Duration = time.Since(now)
	j.logger.Debug().Str("judge", "completeness-judge").Float64("score", result.Score).Msg("judge completed")

	return result
}

func (j *CompletenessJudge) buildPrompt(evaluationContext models.EvaluationContext) string {
	return fmt.Sprintf(`You are a completeness judge.
You are evaluating answer completeness.

Query: %s
Answer: %s

Task: Identify all distinct questions/requests in the query.
Does the answer address EACH one?
Score:
  - 1.0: All parts fully addressed
  - 0.5: Some parts missing or incomplete
  - 0.0: Major parts ignored

Respond ONLY in raw JSON with no markdown, no code blocks, no explanation:
{"score": <float>, "reason": "<which parts were addressed>"}`, evaluationContext.Query, evaluationContext.Answer)
}
