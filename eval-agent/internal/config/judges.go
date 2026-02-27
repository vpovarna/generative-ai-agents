package config

import (
	"fmt"
	"os"
	"text/template"

	"gopkg.in/yaml.v3"
)

// JudgesConfig is the root configuration structure
type JudgesConfig struct {
	Judges Judges `yaml:"judges"`
}

// Judges contains default model config and list of evaluators
type Judges struct {
	DefaultModel ModelConfig          `yaml:"default_model"`
	Evaluators   []JudgeConfiguration `yaml:"evaluators"`
}

// JudgeConfiguration defines a single judge configuration
type JudgeConfiguration struct {
	Name            string       `yaml:"name"`
	Enabled         bool         `yaml:"enabled"`
	Description     string       `yaml:"description"`
	RequiresContext bool         `yaml:"requires_context"`
	Prompt          string       `yaml:"prompt"`
	Model           *ModelConfig `yaml:"model,omitempty"` // Optional override
}

// ModelConfig defines LLM model parameters
type ModelConfig struct {
	MaxTokens   int     `yaml:"max_tokens,omitempty"`
	Temperature float64 `yaml:"temperature,omitempty"`
	Retry       bool    `yaml:"retry,omitempty"`
}

// LoadJudgesConfig loads and validates the judges configuration from YAML
func LoadJudgesConfig() (*JudgesConfig, error) {
	path := os.Getenv("JUDGES_CONFIG_PATH")
	if path == "" {
		path = "configs/judges.yaml"
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", path, err)
	}

	var cfg JudgesConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	applyDefaults(&cfg)

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &cfg, nil
}

func applyDefaults(cfg *JudgesConfig) {
	if cfg.Judges.DefaultModel.MaxTokens == 0 {
		cfg.Judges.DefaultModel.MaxTokens = 256
	}
	if cfg.Judges.DefaultModel.Temperature == 0.0 {
		cfg.Judges.DefaultModel.Temperature = 0.0
	}

	// For each judge, apply defaults
	for i := range cfg.Judges.Evaluators {
		judge := &cfg.Judges.Evaluators[i]

		if judge.Model == nil {
			judge.Model = &ModelConfig{
				MaxTokens:   cfg.Judges.DefaultModel.MaxTokens,
				Temperature: cfg.Judges.DefaultModel.Temperature,
				Retry:       cfg.Judges.DefaultModel.Retry,
			}
		} else {
			if judge.Model.MaxTokens == 0 {
				judge.Model.MaxTokens = cfg.Judges.DefaultModel.MaxTokens
			}
			if judge.Model.Temperature == 0.0 {
				judge.Model.Temperature = cfg.Judges.DefaultModel.Temperature
			}
		}
	}
}

func (cfg *JudgesConfig) Validate() error {
	if len(cfg.Judges.Evaluators) == 0 {
		return fmt.Errorf("no judges configured in evaluators list")
	}

	seen := make(map[string]bool)

	for i, judge := range cfg.Judges.Evaluators {
		if judge.Name == "" {
			return fmt.Errorf("judge at index %d is missing name", i)
		}

		if seen[judge.Name] {
			return fmt.Errorf("duplicate judge name: %s", judge.Name)
		}
		seen[judge.Name] = true

		if judge.Prompt == "" {
			return fmt.Errorf("judge %s is missing prompt", judge.Name)
		}

		if _, err := template.New(judge.Name).Parse(judge.Prompt); err != nil {
			return fmt.Errorf("judge %s has invalid prompt template: %w", judge.Name, err)
		}

		if judge.Model != nil {
			if judge.Model.MaxTokens < 0 {
				return fmt.Errorf("judge %s has negative max_tokens: %d", judge.Name, judge.Model.MaxTokens)
			}
			if judge.Model.Temperature < 0.0 || judge.Model.Temperature > 1.0 {
				return fmt.Errorf("judge %s has invalid temperature: %f (must be 0.0-1.0)", judge.Name, judge.Model.Temperature)
			}
		}
	}

	if cfg.Judges.DefaultModel.MaxTokens < 0 {
		return fmt.Errorf("default model has negative max_tokens: %d", cfg.Judges.DefaultModel.MaxTokens)
	}
	if cfg.Judges.DefaultModel.Temperature < 0.0 || cfg.Judges.DefaultModel.Temperature > 1.0 {
		return fmt.Errorf("default model has invalid temperature: %f (must be 0.0-1.0)", cfg.Judges.DefaultModel.Temperature)
	}

	return nil
}
