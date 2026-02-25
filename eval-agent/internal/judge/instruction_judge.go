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

// InstructionJudge evaluates whether the answer follows explicit instructions in the query.
// It checks for format requirements, count specifications, style constraints, and other directives.
type InstructionJudge struct {
	llmClient LLMClient
	logger    *zerolog.Logger
}

func NewInstructionJudge(client LLMClient, logger *zerolog.Logger) *InstructionJudge {
	return &InstructionJudge{
		llmClient: client,
		logger:    logger,
	}
}

func (j *InstructionJudge) Evaluate(ctx context.Context, evaluationContext models.EvaluationContext) models.StageResult {
	now := time.Now()

	result := models.StageResult{
		Name:  "instruction-judge",
		Score: 0.0,
	}

	prompt := j.buildPrompt(evaluationContext)

	resp, err := j.llmClient.InvokeModel(ctx, bedrock.ClaudeRequest{
		Prompt:      prompt,
		MaxTokens:   300,
		Temperature: 0.0,
	})
	if err != nil {
		j.logger.Error().Err(err).Str("judge", "instruction-judge").Msg("LLM call failed")

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

	result.Score = llmResponse.Score
	result.Reason = llmResponse.Reason
	result.Duration = time.Since(now)

	j.logger.Debug().Str("judge", "instruction-judge").Float64("score", result.Score).Msg("judge completed")
	return result
}

func (j *InstructionJudge) buildPrompt(evaluationContext models.EvaluationContext) string {
	return fmt.Sprintf(`You are an evaluation judge for instruction-following.

Your task:
1. Carefully analyze the query for any EXPLICIT instructions or requirements
2. Check if the answer follows each instruction
3. Score based on compliance

Query: %s
Answer: %s

Types of instructions to look for:
- Format requirements: "as JSON", "in bullet points", "as a list", "in code format", "as a table", "step by step"
- Count specifications: "3 examples", "list 5 items", "top 10", "at least 2"
- Style directives: "be concise", "in detail", "briefly", "explain simply", "comprehensively"
- Length constraints: "in one sentence", "in 50 words or less", "in a paragraph"
- Content constraints: "without technical jargon", "for beginners", "with examples", "include code"

Scoring guidelines:
- 1.0: All instructions followed perfectly, OR no explicit instructions in query
- 0.7-0.9: Most instructions followed, minor deviations
- 0.4-0.6: Some instructions followed, some ignored
- 0.0-0.3: Instructions largely ignored

IMPORTANT: Only evaluate EXPLICIT instructions. Do not penalize for general quality issues.

Respond ONLY in raw JSON with no markdown, no code blocks, no explanation:
{"score": <float>, "reason": "<which parts were addressed>"}`, evaluationContext.Query, evaluationContext.Answer)
}
