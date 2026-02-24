package mcpadapter

import (
	"context"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/povarna/generative-ai-with-go/eval-agent/internal/executor"
	"github.com/povarna/generative-ai-with-go/eval-agent/internal/models"
)

// NewEvaluateHandler returns a tool handler that uses the given executor.
// Pass the returned function to mcp.AddTool.
func NewEvaluateHandler(exec *executor.Executor) func(context.Context, *mcp.CallToolRequest, models.EvaluationContext) (*mcp.CallToolResult, models.EvaluationResult, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input models.EvaluationContext) (*mcp.CallToolResult, models.EvaluationResult, error) {
		return EvaluateResponse(ctx, exec, req, input)
	}
}

// Tool handler
func EvaluateResponse(
	ctx context.Context,
	exec *executor.Executor,
	req *mcp.CallToolRequest,
	input models.EvaluationContext,
) (*mcp.CallToolResult, models.EvaluationResult, error) {

	// Normalize to internal model
	evalCtx := models.EvaluationContext{
		RequestID: input.RequestID,
		Query:     input.Query,
		Context:   input.Context,
		Answer:    input.Answer,
		CreatedAt: time.Now(),
	}

	// Call your EXISTING executor
	result := exec.Execute(ctx, evalCtx)

	// Convert to output format
	stages := make([]models.StageResult, len(result.Stages))
	for i, s := range result.Stages {
		stages[i] = models.StageResult{
			Name:   s.Name,
			Score:  s.Score,
			Reason: s.Reason,
		}
	}

	output := models.EvaluationResult{
		ID:         result.ID,
		Confidence: result.Confidence,
		Verdict:    result.Verdict,
		Stages:     stages,
	}

	return nil, output, nil
}
