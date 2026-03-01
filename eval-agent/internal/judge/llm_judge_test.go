package judge

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/povarna/generative-ai-agents/eval-agent/internal/bedrock"
	"github.com/povarna/generative-ai-agents/eval-agent/internal/config"
	"github.com/povarna/generative-ai-agents/eval-agent/internal/models"
	"github.com/rs/zerolog"
)

func TestNewLLMJudge_Success(t *testing.T) {
	logger := zerolog.Nop()

	cfg := config.JudgeConfiguration{
		Name:            "test-judge",
		Enabled:         true,
		Description:     "Test judge",
		RequiresContext: false,
		Prompt:          "Score: {{.Answer}}",
		Model: &config.ModelConfig{
			MaxTokens:   256,
			Temperature: 0.0,
			Retry:       false,
		},
	}

	judge, err := NewLLMJudge(cfg, &MockLLMClient{}, &logger)
	if err != nil {
		t.Fatalf("NewLLMJudge failed: %v", err)
	}

	if judge.name != "test-judge" {
		t.Errorf("Expected name 'test-judge', got '%s'", judge.name)
	}
	if judge.requiresContext {
		t.Error("Expected requiresContext=false")
	}
	if judge.modelConfig.MaxTokens != 256 {
		t.Errorf("Expected MaxTokens=256, got %d", judge.modelConfig.MaxTokens)
	}
}

func TestNewLLMJudge_InvalidTemplate(t *testing.T) {
	logger := zerolog.Nop()

	cfg := config.JudgeConfiguration{
		Name:   "test-judge",
		Prompt: "{{.Invalid", // Invalid template syntax
		Model: &config.ModelConfig{
			MaxTokens: 256,
		},
	}

	_, err := NewLLMJudge(cfg, &MockLLMClient{}, &logger)
	if err == nil {
		t.Error("Expected error for invalid template")
	}
}

func TestNewLLMJudge_NilModelConfig(t *testing.T) {
	logger := zerolog.Nop()

	cfg := config.JudgeConfiguration{
		Name:   "test-judge",
		Prompt: "test",
		Model:  nil, // Should not happen after config loading
	}

	_, err := NewLLMJudge(cfg, &MockLLMClient{}, &logger)
	if err == nil {
		t.Error("Expected error for nil model config")
	}
}

func TestLLMJudge_Evaluate_Success(t *testing.T) {
	logger := zerolog.Nop()

	cfg := config.JudgeConfiguration{
		Name:            "relevance",
		Prompt:          "Query: {{.Query}}\nAnswer: {{.Answer}}",
		RequiresContext: false,
		Model: &config.ModelConfig{
			MaxTokens:   256,
			Temperature: 0.0,
			Retry:       false,
		},
	}

	mockClient := &MockLLMClient{
		ResponseToReturn: &bedrock.ClaudeResponse{
			Content: `{"score": 0.85, "reason": "Good match"}`,
		},
	}

	judge, err := NewLLMJudge(cfg, mockClient, &logger)
	if err != nil {
		t.Fatalf("NewLLMJudge failed: %v", err)
	}

	evalCtx := models.EvaluationContext{
		Query:  "What is AI?",
		Answer: "AI is artificial intelligence",
	}

	result := judge.Evaluate(context.Background(), evalCtx)

	if result.Score != 0.85 {
		t.Errorf("Expected score=0.85, got %f", result.Score)
	}
	if result.Reason != "Good match" {
		t.Errorf("Expected reason='Good match', got '%s'", result.Reason)
	}
	if result.Name != "relevance-judge" {
		t.Errorf("Expected name='relevance-judge', got '%s'", result.Name)
	}
}

func TestLLMJudge_Evaluate_MissingContext(t *testing.T) {
	logger := zerolog.Nop()

	cfg := config.JudgeConfiguration{
		Name:            "faithfulness",
		Prompt:          "Context: {{.Context}}\nAnswer: {{.Answer}}",
		RequiresContext: true, // Requires context
		Model: &config.ModelConfig{
			MaxTokens: 256,
		},
	}

	judge, _ := NewLLMJudge(cfg, &MockLLMClient{}, &logger)

	evalCtx := models.EvaluationContext{
		Query:   "test",
		Answer:  "test",
		Context: "", // Missing context
	}

	result := judge.Evaluate(context.Background(), evalCtx)

	if result.Score != 0.0 {
		t.Errorf("Expected score=0.0 for missing context, got %f", result.Score)
	}
	if result.Reason != "Context required but not provided" {
		t.Errorf("Expected context error, got '%s'", result.Reason)
	}
}

func TestLLMJudge_Evaluate_TemplateExecutionFails(t *testing.T) {
	logger := zerolog.Nop()

	cfg := config.JudgeConfiguration{
		Name:   "test",
		Prompt: "{{.NonExistentField}}", // Field doesn't exist in EvaluationContext
		Model: &config.ModelConfig{
			MaxTokens: 256,
		},
	}

	judge, _ := NewLLMJudge(cfg, &MockLLMClient{}, &logger)

	evalCtx := models.EvaluationContext{
		Query:  "test",
		Answer: "test",
	}

	result := judge.Evaluate(context.Background(), evalCtx)

	if result.Score != 0.0 {
		t.Errorf("Expected score=0.0 for template error, got %f", result.Score)
	}
	if result.Reason == "" || len(result.Reason) < 5 {
		t.Errorf("Expected error reason, got '%s'", result.Reason)
	}
}

func TestLLMJudge_Evaluate_LLMCallFails(t *testing.T) {
	logger := zerolog.Nop()

	cfg := config.JudgeConfiguration{
		Name:   "test",
		Prompt: "Score: {{.Answer}}",
		Model: &config.ModelConfig{
			MaxTokens: 256,
			Retry:     false,
		},
	}

	mockClient := &MockLLMClient{
		ErrorToReturn: errors.New("API error"),
	}

	judge, _ := NewLLMJudge(cfg, mockClient, &logger)

	evalCtx := models.EvaluationContext{
		Answer: "test",
	}

	result := judge.Evaluate(context.Background(), evalCtx)

	if result.Score != 0.0 {
		t.Errorf("Expected score=0.0 for LLM error, got %f", result.Score)
	}
	if result.Reason != "Failed to call LLM" {
		t.Errorf("Expected LLM error reason, got '%s'", result.Reason)
	}
}

func TestLLMJudge_Evaluate_WithRetry(t *testing.T) {
	logger := zerolog.Nop()

	cfg := config.JudgeConfiguration{
		Name:   "test",
		Prompt: "Score: {{.Answer}}",
		Model: &config.ModelConfig{
			MaxTokens: 256,
			Retry:     true, // Should use InvokeModelWithRetry
		},
	}

	mockClient := &MockLLMClient{
		ResponseToReturn: &bedrock.ClaudeResponse{
			Content: `{"score": 0.9, "reason": "test"}`,
		},
	}

	judge, _ := NewLLMJudge(cfg, mockClient, &logger)

	evalCtx := models.EvaluationContext{
		Answer: "test",
	}

	result := judge.Evaluate(context.Background(), evalCtx)

	if result.Score != 0.9 {
		t.Errorf("Expected score=0.9, got %f", result.Score)
	}
	// Note: Cannot verify InvokeModelWithRetry was called vs InvokeModel
	// without modifying the existing MockLLMClient
}

func TestLLMJudge_Evaluate_InvalidJSON(t *testing.T) {
	logger := zerolog.Nop()

	cfg := config.JudgeConfiguration{
		Name:   "test",
		Prompt: "Score: {{.Answer}}",
		Model: &config.ModelConfig{
			MaxTokens: 256,
		},
	}

	mockClient := &MockLLMClient{
		ResponseToReturn: &bedrock.ClaudeResponse{
			Content: `not valid json`,
		},
	}

	judge, _ := NewLLMJudge(cfg, mockClient, &logger)

	evalCtx := models.EvaluationContext{
		Answer: "test",
	}

	result := judge.Evaluate(context.Background(), evalCtx)

	if result.Score != 0.0 {
		t.Errorf("Expected score=0.0 for invalid JSON, got %f", result.Score)
	}
	if result.Reason != "Failed to deserialize LLM response" {
		t.Errorf("Expected deserialization error, got '%s'", result.Reason)
	}
}

func TestLLMJudge_Evaluate_EmptyScoreAndReason(t *testing.T) {
	logger := zerolog.Nop()

	cfg := config.JudgeConfiguration{
		Name:   "test",
		Prompt: "Score: {{.Answer}}",
		Model: &config.ModelConfig{
			MaxTokens: 256,
		},
	}

	mockClient := &MockLLMClient{
		ResponseToReturn: &bedrock.ClaudeResponse{
			Content: `{"score": 0.0, "reason": ""}`,
		},
	}

	judge, _ := NewLLMJudge(cfg, mockClient, &logger)

	evalCtx := models.EvaluationContext{
		Answer: "test",
	}

	result := judge.Evaluate(context.Background(), evalCtx)

	if result.Score != 0.0 {
		t.Errorf("Expected score=0.0, got %f", result.Score)
	}
	if result.Reason != "Invalid LLM response: missing score and reason" {
		t.Errorf("Expected invalid response error, got '%s'", result.Reason)
	}
}

func TestLLMJudge_Evaluate_ScoreOutOfRange(t *testing.T) {
	tests := []struct {
		name  string
		score float64
	}{
		{"negative", -0.5},
		{"too high", 1.5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := zerolog.Nop()

			cfg := config.JudgeConfiguration{
				Name:   "test",
				Prompt: "Score: {{.Answer}}",
				Model: &config.ModelConfig{
					MaxTokens: 256,
				},
			}

			mockClient := &MockLLMClient{
				ResponseToReturn: &bedrock.ClaudeResponse{
					Content: fmt.Sprintf(`{"score": %f, "reason": "test"}`, tt.score),
				},
			}

			judge, _ := NewLLMJudge(cfg, mockClient, &logger)

			evalCtx := models.EvaluationContext{
				Answer: "test",
			}

			result := judge.Evaluate(context.Background(), evalCtx)

			if result.Score != 0.0 {
				t.Errorf("Expected score=0.0 for out of range score, got %f", result.Score)
			}
			if !contains(result.Reason, "out of range") {
				t.Errorf("Expected out of range error, got '%s'", result.Reason)
			}
		})
	}
}

// MockLLMClient for testing
type MockLLMClient struct {
	ResponseToReturn *bedrock.ClaudeResponse
	ErrorToReturn    error
	WasCalled        bool
	LastRequest      *bedrock.ClaudeRequest
}

func (m *MockLLMClient) InvokeModel(ctx context.Context, request bedrock.ClaudeRequest) (*bedrock.ClaudeResponse, error) {
	m.WasCalled = true
	m.LastRequest = &request
	if m.ErrorToReturn != nil {
		return nil, m.ErrorToReturn
	}
	return m.ResponseToReturn, nil
}

func (m *MockLLMClient) InvokeModelWithRetry(ctx context.Context, request bedrock.ClaudeRequest) (*bedrock.ClaudeResponse, error) {
	m.WasCalled = true
	m.LastRequest = &request
	if m.ErrorToReturn != nil {
		return nil, m.ErrorToReturn
	}
	return m.ResponseToReturn, nil
}

// Helper
func contains(s, substr string) bool {
	return len(s) >= len(substr) && containsHelper(s, substr)
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
