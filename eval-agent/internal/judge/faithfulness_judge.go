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

// FaithfulnessJudge evaluates whether the answer is grounded in the provided context.
// It penalizes answers that introduce facts or claims not supported by the context (hallucinations).
type FaithfulnessJudge struct {
	llmClient LLMClient
	logger    *zerolog.Logger
}

func NewFaithfulnessJudge(client LLMClient, logger *zerolog.Logger) *FaithfulnessJudge {
	return &FaithfulnessJudge{
		llmClient: client,
		logger:    logger,
	}
}

func (j *FaithfulnessJudge) Evaluate(ctx context.Context, evaluationContext models.EvaluationContext) models.StageResult {
	now := time.Now()

	result := models.StageResult{
		Name:  "faithfulness-judge",
		Score: 0.0,
	}

	prompt := j.buildPrompt(evaluationContext)

	resp, err := j.llmClient.InvokeModel(ctx, bedrock.ClaudeRequest{
		Prompt:      prompt,
		MaxTokens:   256,
		Temperature: 0.0, // determinist
	})
	if err != nil {
		j.logger.Error().Err(err).Str("judge", "faithfulness-judge").Msg("LLM call failed")

		result.Reason = "Failed to call LLM"
		result.Duration = time.Since(now)
		return result
	}

	var llmResponse judgeResponse
	if err := json.Unmarshal([]byte(resp.Content), &llmResponse); err != nil {
		j.logger.Error().Err(err).Msg("Failed to deserialize")
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

	if llmResponse.Score < 0.0 || llmResponse.Score > 1.0 {
		j.logger.Error().Msg("LLM returned invalid score")
		result.Reason = "Invalid LLM response: missing score and reason"
		result.Duration = time.Since(now)
		return result
	}

	result.Score = llmResponse.Score
	result.Reason = llmResponse.Reason
	result.Duration = time.Since(now)

	j.logger.Debug().Str("judge", "faithfulness-judge").Float64("score", result.Score).Msg("judge completed")
	return result

}

func (j *FaithfulnessJudge) buildPrompt(evaluationContext models.EvaluationContext) string {
	return fmt.Sprintf(`You are an evaluation judge.
Score how faithful the answer is to the provided context, on a scale from 0.0 to 1.0.
Penalize if the answer introduces facts not present in the context.

Context: %s
Answer: %s

Respond ONLY in raw JSON with no markdown, no code blocks, no explanation:
{"score": <float>, "reason": "<string>"}`, evaluationContext.Context, evaluationContext.Answer)
}
