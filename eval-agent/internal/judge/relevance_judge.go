package judge

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/povarna/generative-ai-with-go/eval-agent/internal/bedrock"
	"github.com/povarna/generative-ai-with-go/eval-agent/internal/models"
)

// This is an LLM judge who validates if the answer address the query
type RelevanceJudge struct {
	llmClient *bedrock.Client
}

func NewRelevanceJudge(client *bedrock.Client) *RelevanceJudge {
	return &RelevanceJudge{
		llmClient: client,
	}
}


func (j *RelevanceJudge) Evaluate(ctx context.Context, evaluationContext models.EvaluationContext) models.StageResult {
	now := time.Now()
	result := models.StageResult{
		Name:  "relevance-judge",
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

func (j *RelevanceJudge) buildPrompt(evaluationContext models.EvaluationContext) string {
	return fmt.Sprintf(`You are an evaluation judge. 
Score how relevant the answer is to the query on a scale from 0.1 to 1.0

Query: %s
Answer: %s

Respond ONLY in JSON: {"score": <float>, "reason": "<string>"}`, evaluationContext.Query, evaluationContext.Answer)
}
