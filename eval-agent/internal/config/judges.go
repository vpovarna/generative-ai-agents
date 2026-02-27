package config

import (
	"os"

	"go.yaml.in/yaml/v3"
)

func LoadJudgesConfig() (*JudgeConfig, error) {

	path := os.Getenv("JUDGES_CONFIG_PATH")
	if path == "" {
		path = "configs/judges.yaml"
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg JudgeConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	applyDefaults(&cfg)

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func applyDefaults(cfg *JudgeConfig) {
	if cfg.ModelParams.MaxTokens == 0 {
		cfg.ModelParams.MaxTokens = 256
	}
}

func (j *JudgeConfig) Validate() error {
	return nil
}
