package strategy

import (
	"context"
	"fmt"
	"strings"

	"github.com/povarna/generative-ai-agents/kg-agent/internal/bedrock"
	"github.com/povarna/generative-ai-agents/kg-agent/internal/conversation"
	"github.com/rs/zerolog/log"
)

type RetrievalStrategy struct {
	Client *bedrock.Client
}

func NewRetrievalStrategy(bedrockClient *bedrock.Client) *RetrievalStrategy {
	return &RetrievalStrategy{
		Client: bedrockClient,
	}
}

type Decision struct {
	ShouldSearch bool
	Reason       string
	Confidence   float64 //0.0 to 1.0
}

func (r *RetrievalStrategy) Decide(ctx context.Context, query string, history *conversation.Conversation) Decision {

	heuristicDecision := r.heuristicDecide(query, history)

	if heuristicDecision.Confidence > 0.85 {
		log.Info().Str("method", "heuristic").Msg("Using heuristic decision")
		return heuristicDecision
	}

	decision, err := r.llmDecide(ctx, query, history)
	if err != nil {
		log.Error().Err(err).Msg("Unable to call LLM classifier, using heuristic fallback")
		return heuristicDecision
	}

	return decision
}

// Heuristic Implementation Decision
func (r *RetrievalStrategy) heuristicDecide(query string, history *conversation.Conversation) Decision {
	query = strings.ToLower(strings.TrimSpace(query))

	// Rule 1: Greeting
	if r.isSimpleGreeting(query) {
		return Decision{
			ShouldSearch: false,
			Reason:       "Simple Greetings",
			Confidence:   0.95,
		}
	}

	// Rule 2: Short word
	if len(query) < 5 {
		return Decision{
			ShouldSearch: false,
			Reason:       "Too short to search",
			Confidence:   0.95,
		}
	}

	// Rule 3: Follow up question + history
	if r.isClearFollowUp(query) && r.hasRecentHistory(history) {
		return Decision{
			ShouldSearch: false,
			Reason:       "Obvious follow-up questions",
			Confidence:   0.85,
		}
	}

	// Rule 4: Simple pronoun + recent history
	if r.hasSimplePronoun(query) && r.hasRecentHistory(history) {
		return Decision{false, "References recent message", 0.85}
	}

	return Decision{true, "Default: search for quality", 0.70}
}

func (r *RetrievalStrategy) isSimpleGreeting(query string) bool {
	simpleGreetings := []string{
		"hello", "hi", "hey", "thanks", "thank you", "bye", "goodbye", "ok", "okay", "yes", "no",
	}

	for _, word := range simpleGreetings {
		if word == query {
			return true
		}
	}

	return false
}

func (r *RetrievalStrategy) isClearFollowUp(query string) bool {
	clearPhrases := []string{
		"tell me more",
		"explain that",
		"what does it mean",
		"can you elaborate",
	}

	for _, phrase := range clearPhrases {
		if strings.Contains(query, phrase) {
			return true
		}
	}

	return false
}

func (r *RetrievalStrategy) hasRecentHistory(history *conversation.Conversation) bool {
	if history == nil {
		return false
	}
	return len(history.Messages) >= 2
}

func (r *RetrievalStrategy) hasSimplePronoun(query string) bool {
	// Starts with simple pronouns (high confidence it's a reference)
	startsWithPronouns := []string{"it ", "that ", "this ", "these ", "those "}
	for _, pronoun := range startsWithPronouns {
		if strings.HasPrefix(query, pronoun) {
			return true
		}
	}
	return false
}

func (r *RetrievalStrategy) llmDecide(ctx context.Context, query string, history *conversation.Conversation) (Decision, error) {
	prompt := r.buildClassificationPrompt(query, history)

	response, err := r.Client.InvokeModel(ctx, bedrock.ClaudeRequest{
		Prompt:      prompt,
		MaxTokens:   100,
		Temperature: 0.0,
	})

	if err != nil {
		log.Warn().Err(err).Msg("LLM classification failed, defaulting to search")
		return Decision{
			ShouldSearch: true,
			Reason:       "LLM failed, safe default",
			Confidence:   0.60,
		}, nil
	}

	// Parse the response - check for NO_SEARCH
	content := strings.ToUpper(response.Content)
	shouldSearch := !strings.Contains(content, "NO_SEARCH")

	return Decision{
		ShouldSearch: shouldSearch,
		Reason:       "LLM classification",
		Confidence:   0.90,
	}, nil

}

func (r *RetrievalStrategy) buildClassificationPrompt(query string, history *conversation.Conversation) string {
	historyText := "None"
	if history != nil && len(history.Messages) > 0 {
		historyMessages := history.Messages
		if len(history.Messages) > 4 {
			historyMessages = history.Messages[len(history.Messages)-4:]
		}
		var sb strings.Builder
		for _, msg := range historyMessages {
			sb.WriteString(fmt.Sprintf("%s: %s\n", msg.Role, msg.Content))
		}
		historyText = sb.String()
	}

	return fmt.Sprintf(`You are a search decision classifier for a documentation assistant.

Recent Conversation History:
%s

Current User Query: "%s"

Task: Decide if we need to search external documentation.

Answer NO_SEARCH if:
- This is a greeting, pleasantry, or acknowledgment
- This is a follow-up clearly referencing previous conversation
- You can answer from general knowledge
- The conversation history contains sufficient context

Answer SEARCH if:
- This is a new technical question
- This requires specific documentation
- This is asking for details not in history

Respond EXACTLY in this format:
DECISION: [SEARCH or NO_SEARCH]
REASON: [brief explanation]`, historyText, query)
}
