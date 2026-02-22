package judge

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/povarna/generative-ai-with-go/eval-agent/internal/bedrock"
	"github.com/povarna/generative-ai-with-go/eval-agent/internal/models"
)

// FaithfulnessJudge evaluates whether the answer is grounded in the provided context.
// It penalizes answers that introduce facts or claims not supported by the context (hallucinations).
type FaithfulnessJudge struct {
	llmClient *bedrock.Client
}

func NewFaithfulnessJudge(client *bedrock.Client) *FaithfulnessJudge {
	return &FaithfulnessJudge{
		llmClient: client,
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
		Pormpt:      prompt,
		MaxTokens:   256,
		Temperature: 0.0, // determinist
	})
	if err != nil {
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

	result.Score = llmResponse.Score
	result.Reason = llmResponse.Reason
	result.Duration = time.Since(now)

	return result

}

func (j *FaithfulnessJudge) buildPrompt(evaluationContext models.EvaluationContext) string {
	return fmt.Sprintf(`You are an evaluation judge.
Score how faithful the answer is to the provided context, on a scale from 0.0 to 1.0.
Penalize if the answer introduces facts not present in the context.

Context: %s
Answer: %s

Does the answer stay faithful to the context? Score 0.0-1.0.
Penalize if the answer introduces facts not present in the context.

Respond ONLY in JSON: {"score": <float>, "reason": "<string>"}`, evaluationContext.Context, evaluationContext.Answer)
}
