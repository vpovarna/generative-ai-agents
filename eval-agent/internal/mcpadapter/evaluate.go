package mcpadapter

import (
	"context"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/povarna/generative-ai-with-go/eval-agent/internal/executor"
	"github.com/povarna/generative-ai-with-go/eval-agent/internal/models"
)

// EvaluateInput is the MCP tool input schema (matches HTTP API field names).
type EvaluateInput struct {
	EventID string `json:"event_id" jsonschema:"unique event identifier"`
	Query   string `json:"user_query" jsonschema:"user's original query"`
	Answer  string `json:"answer" jsonschema:"agent response to evaluate"`
	Context string `json:"context,omitempty" jsonschema:"optional context or retrieved documents"`
}

// NewEvaluateHandler returns a tool handler that uses the given executor.
// Pass the returned function to mcp.AddTool.
func NewEvaluateHandler(exec *executor.Executor) func(context.Context, *mcp.CallToolRequest, EvaluateInput) (*mcp.CallToolResult, models.EvaluationResult, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input EvaluateInput) (*mcp.CallToolResult, models.EvaluationResult, error) {
		return EvaluateResponse(ctx, exec, req, input)
	}
}

// EvaluateResponse runs the evaluation pipeline and returns the result.
func EvaluateResponse(
	ctx context.Context,
	exec *executor.Executor,
	req *mcp.CallToolRequest,
	input EvaluateInput,
) (*mcp.CallToolResult, models.EvaluationResult, error) {
	evalCtx := models.EvaluationContext{
		RequestID: input.EventID,
		Query:     input.Query,
		Context:   input.Context,
		Answer:    input.Answer,
		CreatedAt: time.Now(),
	}

	result := exec.Execute(ctx, evalCtx)
	return nil, result, nil
}
