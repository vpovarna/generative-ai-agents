package judge

import (
	"fmt"

	"github.com/povarna/generative-ai-with-go/eval-agent/internal/config"
	"github.com/rs/zerolog"
)

// JudgeFactory creates and manages judges by name for single-judge execution.
// It loads judges from YAML configuration.
type JudgeFactory struct {
	judges map[string]Judge
}

// NewJudgeFactory creates a factory with judges loaded from configuration.
func NewJudgeFactory(llmClient LLMClient, logger *zerolog.Logger) *JudgeFactory {
	// Load judges config
	judgesConfig, err := config.LoadJudgesConfig()
	if err != nil {
		logger.Error().Err(err).Msg("Failed to load judges config, factory will be empty")
		return &JudgeFactory{
			judges: make(map[string]Judge),
		}
	}

	// Build judges from config
	judgePool := NewJudgePool(llmClient, logger)
	judgesList, err := judgePool.BuildFromConfig(judgesConfig)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to build judges from config, factory will be empty")
		return &JudgeFactory{
			judges: make(map[string]Judge),
		}
	}

	// Create map by judge name for quick lookup
	judgesMap := make(map[string]Judge)
	for _, j := range judgesList {
		judgesMap[j.Name()] = j
	}

	logger.Info().Int("judge_count", len(judgesMap)).Msg("Judge factory initialized from config")

	return &JudgeFactory{
		judges: judgesMap,
	}
}

func (f *JudgeFactory) Get(judgeName string) (Judge, error) {
	judge, exist := f.judges[judgeName]
	if !exist {
		return nil, fmt.Errorf("judge not found")
	}

	return judge, nil
}
