package judge

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/povarna/generative-ai-with-go/eval-agent/internal/bedrock"
	"github.com/povarna/generative-ai-with-go/eval-agent/internal/models"
)

// This is a LLM judge which checks if the answer is internally logically consistent — independent of the query or context.
type CoherenceJudge struct {
	llmClient *bedrock.Client
}

func NewCoherenceJudge(client *bedrock.Client) *CoherenceJudge {
	return &CoherenceJudge{
		llmClient: client,
	}
}

func (j *CoherenceJudge) Evaluate(ctx context.Context, evaluationContext models.EvaluationContext) models.StageResult {
	now := time.Now()

	result := models.StageResult{
		Name:  "coherence-judge",
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

func (j *CoherenceJudge) buildPrompt(evaluationContext models.EvaluationContext) string {
	return fmt.Sprintf(`You are an evaluation judge.
Score how logically coherent and internally consistent the answer is, on a scale from 0.0 to 1.0.
Do NOT consider whether the answer is correct or relevant — only evaluate its internal logic.

Answer: %s

Respond ONLY in JSON: {"score": <float>, "reason": "<string>"}`, evaluationContext.Answer)
}
