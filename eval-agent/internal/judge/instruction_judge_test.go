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

func TestInstructionJudge_Evaluate_HappyPath(t *testing.T) {
	logger := zerolog.New(os.Stdout).Level(zerolog.Disabled)

	mockClient := &MockLLMClient{
		ResponseToReturn: &bedrock.ClaudeResponse{
			Content:    `{"score": 1.0, "reason": "All instructions followed: response is in bullet points as requested"}`,
			StopReason: "end_turn",
		},
		ErrorToReturn: nil,
	}

	judge := NewInstructionJudge(mockClient, &logger)

	evalContext := models.EvaluationContext{
		Query:  "List the benefits of encryption in bullet points",
		Answer: "• Protects data confidentiality\n• Ensures data integrity\n• Provides authentication",
	}

	result := judge.Evaluate(context.Background(), evalContext)

	if !mockClient.WasCalled {
		t.Error("Expected the mock LLM client to be called, but it wasn't")
	}

	if result.Name != "instruction-judge" {
		t.Errorf("Expected name='instruction-judge', got '%s'", result.Name)
	}

	if result.Score != 1.0 {
		t.Errorf("Expected score=1.0, got %f", result.Score)
	}

	if result.Reason != "All instructions followed: response is in bullet points as requested" {
		t.Errorf("Expected specific reason, got '%s'", result.Reason)
	}

	if result.Duration == 0 {
		t.Error("Expected duration to be measured")
	}

	// Verify request parameters specific to InstructionJudge
	if mockClient.LastRequest.MaxTokens != 300 {
		t.Errorf("Expected MaxTokens=300, got %d", mockClient.LastRequest.MaxTokens)
	}

	if mockClient.LastRequest.Temperature != 0.0 {
		t.Errorf("Expected Temperature=0.0, got %f", mockClient.LastRequest.Temperature)
	}
}

func TestInstructionJudge_Evaluate_LowScore(t *testing.T) {
	logger := zerolog.New(os.Stdout).Level(zerolog.Disabled)

	mockClient := &MockLLMClient{
		ResponseToReturn: &bedrock.ClaudeResponse{
			Content:    `{"score": 0.3, "reason": "Instructions largely ignored: requested 3 examples but provided none"}`,
			StopReason: "end_turn",
		},
		ErrorToReturn: nil,
	}

	judge := NewInstructionJudge(mockClient, &logger)

	evalContext := models.EvaluationContext{
		Query:  "Give me 3 examples of encryption algorithms",
		Answer: "Encryption is important for security.",
	}

	result := judge.Evaluate(context.Background(), evalContext)

	if result.Score != 0.3 {
		t.Errorf("Expected score=0.3, got %f", result.Score)
	}

	if result.Reason != "Instructions largely ignored: requested 3 examples but provided none" {
		t.Errorf("Expected specific reason, got '%s'", result.Reason)
	}
}

func TestInstructionJudge_Evaluate_LlmApiError(t *testing.T) {
	logger := zerolog.New(os.Stdout).Level(zerolog.Disabled)

	mockClient := &MockLLMClient{
		ResponseToReturn: nil,
		ErrorToReturn:    errors.New("bedrock service unavailable"),
	}

	judge := NewInstructionJudge(mockClient, &logger)

	evalContext := models.EvaluationContext{
		Query:  "Explain briefly",
		Answer: "Some answer",
	}

	result := judge.Evaluate(context.Background(), evalContext)

	if !mockClient.WasCalled {
		t.Error("Expected the mock LLM client to be called, but it wasn't")
	}

	if result.Name != "instruction-judge" {
		t.Errorf("Expected name='instruction-judge', got '%s'", result.Name)
	}

	if result.Score != 0.0 {
		t.Errorf("Expected score=0.0 on error, got %f", result.Score)
	}

	if result.Reason != "Failed to call LLM" {
		t.Errorf("Expected error reason, got '%s'", result.Reason)
	}
}

func TestInstructionJudge_Evaluate_InvalidJsonFormat(t *testing.T) {
	logger := zerolog.New(os.Stdout).Level(zerolog.Disabled)

	tests := []struct {
		name    string
		content string
	}{
		{
			name:    "plain text response",
			content: "The answer follows instructions well",
		},
		{
			name:    "malformed JSON - unclosed object",
			content: `{"score": 0.9, "reason": "good"`,
		},
		{
			name:    "malformed JSON - invalid syntax",
			content: `{score: 0.8, reason: "test"}`,
		},
		{
			name:    "JSON wrapped in markdown",
			content: "```json\n{\"score\": 0.7, \"reason\": \"test\"}\n```",
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

			judge := NewInstructionJudge(mockClient, &logger)

			result := judge.Evaluate(context.Background(), models.EvaluationContext{
				Query:  "Write concisely",
				Answer: "Done",
			})

			if !mockClient.WasCalled {
				t.Error("Expected the mock LLM client to be called, but it wasn't")
			}

			if result.Name != "instruction-judge" {
				t.Errorf("Expected name='instruction-judge', got '%s'", result.Name)
			}

			if result.Reason != "Failed to deserialize LLM response" {
				t.Errorf("Expected deserialization error, got '%s'", result.Reason)
			}

			if result.Score != 0.0 {
				t.Errorf("Expected score=0.0 on JSON error, got %f", result.Score)
			}
		})
	}
}

func TestInstructionJudge_Evaluate_MissingFields(t *testing.T) {
	logger := zerolog.New(os.Stdout).Level(zerolog.Disabled)

	mockClient := &MockLLMClient{
		ResponseToReturn: &bedrock.ClaudeResponse{
			Content:    `{"other_field": "value"}`,
			StopReason: "end_turn",
		},
		ErrorToReturn: nil,
	}

	judge := NewInstructionJudge(mockClient, &logger)

	evalContext := models.EvaluationContext{
		Query:  "Answer in one sentence",
		Answer: "This is a sentence.",
	}

	result := judge.Evaluate(context.Background(), evalContext)

	if !mockClient.WasCalled {
		t.Error("Expected the mock LLM client to be called, but it wasn't")
	}

	if result.Name != "instruction-judge" {
		t.Errorf("Expected name='instruction-judge', got '%s'", result.Name)
	}

	if result.Reason == "Failed to deserialize LLM response" {
		t.Error("Should not fail on valid JSON with missing fields")
	}

	if result.Reason != "Invalid LLM response: missing score and reason" {
		t.Errorf("Expected validation error, got '%s'", result.Reason)
	}

	if result.Score != 0.0 {
		t.Errorf("Expected score=0.0, got %f", result.Score)
	}
}

func TestInstructionJudge_Evaluate_PartiallyMissingFields(t *testing.T) {
	logger := zerolog.New(os.Stdout).Level(zerolog.Disabled)

	tests := []struct {
		name           string
		content        string
		expectedScore  float64
		expectedReason string
		shouldPass     bool
	}{
		{
			name:           "missing reason field",
			content:        `{"score": 0.8}`,
			expectedScore:  0.8,
			expectedReason: "",
			shouldPass:     true, // Has score, even if reason is empty
		},
		{
			name:           "missing score field",
			content:        `{"reason": "Instructions followed"}`,
			expectedScore:  0.0,
			expectedReason: "Invalid LLM response: missing score and reason", // Validation catches this
			shouldPass:     false,
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

			judge := NewInstructionJudge(mockClient, &logger)

			result := judge.Evaluate(context.Background(), models.EvaluationContext{
				Query:  "Test query",
				Answer: "Test answer",
			})

			if result.Score != tt.expectedScore {
				t.Errorf("Expected score=%f, got %f", tt.expectedScore, result.Score)
			}

			if tt.shouldPass && result.Reason != tt.expectedReason {
				t.Errorf("Expected reason='%s', got '%s'", tt.expectedReason, result.Reason)
			}
		})
	}
}

func TestInstructionJudge_Evaluate_NoExplicitInstructions(t *testing.T) {
	logger := zerolog.New(os.Stdout).Level(zerolog.Disabled)

	mockClient := &MockLLMClient{
		ResponseToReturn: &bedrock.ClaudeResponse{
			Content:    `{"score": 1.0, "reason": "No explicit instructions in query"}`,
			StopReason: "end_turn",
		},
	}

	judge := NewInstructionJudge(mockClient, &logger)

	evalContext := models.EvaluationContext{
		Query:  "What is encryption?",
		Answer: "Encryption is the process of encoding data to protect it from unauthorized access using cryptographic algorithms.",
	}

	result := judge.Evaluate(context.Background(), evalContext)

	if result.Score != 1.0 {
		t.Errorf("Expected score=1.0 when no instructions present, got %f", result.Score)
	}

	if result.Reason != "No explicit instructions in query" {
		t.Errorf("Expected specific reason, got '%s'", result.Reason)
	}
}
