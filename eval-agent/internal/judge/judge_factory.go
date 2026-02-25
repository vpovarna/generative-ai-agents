package judge

import (
	"fmt"

	"github.com/rs/zerolog"
)

type JudgeFactory struct {
	judges map[string]Judge
}

func NewJudgeFactory(llmClient LLMClient, logger *zerolog.Logger) *JudgeFactory {
	return &JudgeFactory{
		judges: map[string]Judge{
			"relevance":    NewRelevanceJudge(llmClient, logger),
			"faithfulness": NewFaithfulnessJudge(llmClient, logger),
			"coherence":    NewCoherenceJudge(llmClient, logger),
			"completeness": NewCompletenessJudge(llmClient, logger),
			"instruction":  NewInstructionJudge(llmClient, logger),
		},
	}
}

func (f *JudgeFactory) Get(judgeName string) (Judge, error) {
	judge, exist := f.judges[judgeName]
	if !exist {
		return nil, fmt.Errorf("judge not found")
	}

	return judge, nil
}
