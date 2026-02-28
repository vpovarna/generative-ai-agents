package guardrails

import (
	"context"

	"github.com/povarna/generative-ai-agents/kg-agent/internal/bedrock"
	"github.com/rs/zerolog/log"
)

type Guardrails struct {
	staticValidator *StaticValidator
	claudeValidator *ClaudeValidator
	enableClaude    bool
}

func NewGuardrails(
	claudeClient *bedrock.Client,
) *Guardrails {
	return &Guardrails{
		staticValidator: NewStaticValidator(DefaultBanWords),
		claudeValidator: NewClaudeValidator(claudeClient),
		enableClaude:    true,
	}
}

func (g *Guardrails) ValidateInput(ctx context.Context, input string) ValidationResult {
	// Run static rules first (fast, free)
	result := g.staticValidator.Validate(input)
	if !result.IsValid {
		log.Info().Str("method", "static").Str("reason", result.Reason).Msg("Input blocked by static rules")
		return result
	}

	// If static passes and Claude is enabled, run Claude validator
	if g.enableClaude {
		result = g.claudeValidator.Validate(ctx, input)
		if !result.IsValid {
			log.Warn().
				Str("method", "claude").
				Str("category", result.Category).
				Str("reason", result.Reason).
				Msg("Input blocked by Claude validator")
		}
		return result
	}

	// All checks passed
	return ValidationResult{IsValid: true, Reason: "Input validated", Method: "static"}
}
