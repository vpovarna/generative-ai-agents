package judge

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/povarna/generative-ai-with-go/eval-agent/internal/bedrock"
	"github.com/povarna/generative-ai-with-go/eval-agent/internal/models"
	"github.com/rs/zerolog"
)

// MockLLMClient is a fake LLM client for testing
type MockLLMClient struct {
	// What the mock should return when InvokeModel is called
	ResponseToReturn *bedrock.ClaudeResponse
	ErrorToReturn    error

	// Track if the mock was called (useful for verification)
	WasCalled   bool
	LastRequest *bedrock.ClaudeRequest
}

// InvokeModel implements the LLMClient interface
func (m *MockLLMClient) InvokeModel(ctx context.Context, request bedrock.ClaudeRequest) (*bedrock.ClaudeResponse, error) {
	m.WasCalled = true
	m.LastRequest = &request
	return m.ResponseToReturn, m.ErrorToReturn
}

func TestRelevanceJudge_Evaluate_HappyPath(t *testing.T) {
	logger := zerolog.New(os.Stdout).Level(zerolog.Disabled)

	mockClient := &MockLLMClient{
		ResponseToReturn: &bedrock.ClaudeResponse{
			Content:    `{"score": 0.9, "reason": "The answer directly addresses the query about encryption"}`,
			StopReason: "end_turn",
		},
		ErrorToReturn: nil,
	}

	judge := NewRelevanceJudge(mockClient, &logger)

	evalContext := models.EvaluationContext{
		Query:  "What is encryption?",
		Answer: "Encryption is the process of encoding data to protect it from unauthorized access.",
	}

	result := judge.Evaluate(context.Background(), evalContext)

	if !mockClient.WasCalled {
		t.Error("Expected the mock LLM client to be called, but it wasn't")
	}

	if result.Name != "relevance-judge" {
		t.Errorf("Expected name='relevance-judge', got '%s'", result.Name)
	}

	if result.Score != 0.9 {
		t.Errorf("Expected score=0.9, got %f", result.Score)
	}

	if result.Reason != "The answer directly addresses the query about encryption" {
		t.Errorf("Expected specific reason, got '%s'", result.Reason)
	}

	if result.Duration == 0 {
		t.Error("Expected duration to be measured")
	}

	if mockClient.LastRequest.MaxTokens != 256 {
		t.Errorf("Expected MaxTokens=256, got %d", mockClient.LastRequest.MaxTokens)
	}

	if mockClient.LastRequest.Temperature != 0.0 {
		t.Errorf("Expected Temperature=0.0, got %f", mockClient.LastRequest.Temperature)
	}
}

func TestRelevanceJudge_Evaluate_LlmApiError(t *testing.T) {
	logger := zerolog.New(os.Stdout).Level(zerolog.Disabled)

	mockClient := &MockLLMClient{
		ResponseToReturn: nil,
		ErrorToReturn:    errors.New("API failed"),
	}

	judge := NewRelevanceJudge(mockClient, &logger)

	evalCtx := models.EvaluationContext{
		Query:  "Some query",
		Answer: "Some answer",
	}

	result := judge.Evaluate(context.Background(), evalCtx)

	if !mockClient.WasCalled {
		t.Error("Expected the mock LLM client to be called, but it wasn't")
	}

	if result.Name != "relevance-judge" {
		t.Errorf("Expected name='relevance-judge', got '%s'", result.Name)
	}

	if result.Reason != "Failed to call LLM" {
		t.Errorf("Invalid reason message, got :%s", result.Reason)
	}

}

func TestRelevanceJudge_Evaluate_InvalidJsonFormat(t *testing.T) {
	logger := zerolog.New(os.Stdout).Level(zerolog.Disabled)

	tests := []struct {
		name    string
		content string
	}{
		{
			name:    "not JSON at all",
			content: "This is plain text, not JSON",
		},
		{
			name:    "malformed JSON - missing closing brace",
			content: `{"score": 0.8, "reason": "test"`,
		},
		{
			name:    "malformed JSON - trailing comma",
			content: `{"score": 0.8, "reason": "test",}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &MockLLMClient{
				ResponseToReturn: &bedrock.ClaudeResponse{
					Content:    tt.content,
					StopReason: "end_turn",
				},
			}

			judge := NewRelevanceJudge(mockClient, &logger)
			evalContext := models.EvaluationContext{
				Query:  "What is encryption?",
				Answer: "Encryption is the process of encoding data to protect it from unauthorized access.",
			}

			result := judge.Evaluate(context.Background(), evalContext)

			if !mockClient.WasCalled {
				t.Error("Expected the mock LLM client to be called, but it wasn't")
			}

			if result.Name != "relevance-judge" {
				t.Errorf("Expected name='relevance-judge', got '%s'", result.Name)
			}

			if result.Reason != "Failed to deserialize LLM response" {
				t.Errorf("Invalid reason message, got :%s", result.Reason)
			}
		})

	}
}

func TestRelevanceJudge_Evaluate_MissingFields(t *testing.T) {
	logger := zerolog.New(os.Stdout).Level(zerolog.Disabled)

	mockClient := &MockLLMClient{
		ResponseToReturn: &bedrock.ClaudeResponse{
			Content:    `{"test": "irrelevant field"}`,
			StopReason: "end_turn",
		},
		ErrorToReturn: nil,
	}

	judge := NewRelevanceJudge(mockClient, &logger)
	evalContext := models.EvaluationContext{
		Query:  "What is encryption?",
		Answer: "Test Answer",
	}

	result := judge.Evaluate(context.Background(), evalContext)

	if !mockClient.WasCalled {
		t.Error("Expected the mock LLM client to be called, but it wasn't")
	}

	if result.Name != "relevance-judge" {
		t.Errorf("Expected name='relevance-judge', got '%s'", result.Name)
	}

	if result.Reason == "Failed to deserialize LLM response" {
		t.Error("Should not fail on valid JSON with missing fields")
	}

	if result.Reason != "Invalid LLM response: missing score and reason" {
		t.Errorf("Expected empty reason, got '%s'", result.Reason)
	}

}
