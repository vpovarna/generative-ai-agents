package judge

import (
	"testing"

	"github.com/povarna/generative-ai-with-go/eval-agent/internal/config"
	"github.com/rs/zerolog"
)

func TestNewJudgePool(t *testing.T) {
	logger := zerolog.Nop()
	mockClient := &MockLLMClient{}

	pool := NewJudgePool(mockClient, &logger)

	if pool == nil {
		t.Fatal("Expected pool to be created")
	}
	if pool.llmClient == nil {
		t.Error("Expected llmClient to be set")
	}
}

func TestJudgePool_BuildFromConfig_Success(t *testing.T) {
	logger := zerolog.Nop()
	mockClient := &MockLLMClient{}

	pool := NewJudgePool(mockClient, &logger)

	cfg := &config.JudgesConfig{
		Judges: config.Judges{
			DefaultModel: config.ModelConfig{
				MaxTokens:   256,
				Temperature: 0.0,
				Retry:       true,
			},
			Evaluators: []config.JudgeConfiguration{
				{
					Name:    "judge1",
					Enabled: true,
					Prompt:  "Score: {{.Answer}}",
					Model: &config.ModelConfig{
						MaxTokens: 256,
					},
				},
				{
					Name:    "judge2",
					Enabled: true,
					Prompt:  "Score: {{.Query}}",
					Model: &config.ModelConfig{
						MaxTokens: 128,
					},
				},
			},
		},
	}

	judges, err := pool.BuildFromConfig(cfg)
	if err != nil {
		t.Fatalf("BuildFromConfig failed: %v", err)
	}

	if len(judges) != 2 {
		t.Errorf("Expected 2 judges, got %d", len(judges))
	}
}

func TestJudgePool_BuildFromConfig_SkipsDisabledJudges(t *testing.T) {
	logger := zerolog.Nop()
	mockClient := &MockLLMClient{}

	pool := NewJudgePool(mockClient, &logger)

	cfg := &config.JudgesConfig{
		Judges: config.Judges{
			DefaultModel: config.ModelConfig{
				MaxTokens: 256,
			},
			Evaluators: []config.JudgeConfiguration{
				{
					Name:    "judge1",
					Enabled: true,
					Prompt:  "Score: {{.Answer}}",
					Model: &config.ModelConfig{
						MaxTokens: 256,
					},
				},
				{
					Name:    "judge2",
					Enabled: false, // Disabled
					Prompt:  "Score: {{.Query}}",
					Model: &config.ModelConfig{
						MaxTokens: 128,
					},
				},
				{
					Name:    "judge3",
					Enabled: true,
					Prompt:  "Score: {{.Context}}",
					Model: &config.ModelConfig{
						MaxTokens: 256,
					},
				},
			},
		},
	}

	judges, err := pool.BuildFromConfig(cfg)
	if err != nil {
		t.Fatalf("BuildFromConfig failed: %v", err)
	}

	if len(judges) != 2 {
		t.Errorf("Expected 2 enabled judges, got %d", len(judges))
	}
}

func TestJudgePool_BuildFromConfig_NilConfig(t *testing.T) {
	logger := zerolog.Nop()
	mockClient := &MockLLMClient{}

	pool := NewJudgePool(mockClient, &logger)

	_, err := pool.BuildFromConfig(nil)
	if err == nil {
		t.Error("Expected error for nil config")
	}
	if err.Error() != "judges config is nil" {
		t.Errorf("Expected 'judges config is nil' error, got: %v", err)
	}
}

func TestJudgePool_BuildFromConfig_NoEnabledJudges(t *testing.T) {
	logger := zerolog.Nop()
	mockClient := &MockLLMClient{}

	pool := NewJudgePool(mockClient, &logger)

	cfg := &config.JudgesConfig{
		Judges: config.Judges{
			DefaultModel: config.ModelConfig{
				MaxTokens: 256,
			},
			Evaluators: []config.JudgeConfiguration{
				{
					Name:    "judge1",
					Enabled: false,
					Prompt:  "Score: {{.Answer}}",
					Model: &config.ModelConfig{
						MaxTokens: 256,
					},
				},
			},
		},
	}

	_, err := pool.BuildFromConfig(cfg)
	if err == nil {
		t.Error("Expected error for no enabled judges")
	}

	expectedMsg := "no enabled judges found in config"
	if err.Error() != expectedMsg {
		t.Errorf("Expected '%s' error, got: %v", expectedMsg, err)
	}
}

func TestJudgePool_BuildFromConfig_InvalidJudge(t *testing.T) {
	logger := zerolog.Nop()
	mockClient := &MockLLMClient{}

	pool := NewJudgePool(mockClient, &logger)

	cfg := &config.JudgesConfig{
		Judges: config.Judges{
			DefaultModel: config.ModelConfig{
				MaxTokens: 256,
			},
			Evaluators: []config.JudgeConfiguration{
				{
					Name:    "bad-judge",
					Enabled: true,
					Prompt:  "{{.Invalid", // Invalid template
					Model: &config.ModelConfig{
						MaxTokens: 256,
					},
				},
			},
		},
	}

	_, err := pool.BuildFromConfig(cfg)
	if err == nil {
		t.Error("Expected error for invalid judge")
	}

	// Should mention the judge name in the error
	if !contains(err.Error(), "bad-judge") {
		t.Errorf("Expected error to mention 'bad-judge', got: %v", err)
	}
}
