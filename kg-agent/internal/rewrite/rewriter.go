package rewrite

import (
	"context"
	"fmt"
	"strings"

	"github.com/povarna/generative-ai-with-go/kg-agent/internal/bedrock"
	"github.com/rs/zerolog/log"
)

type Rewriter struct {
	claudeClient *bedrock.Client
}

func NewRewriter(client *bedrock.Client) *Rewriter {
	return &Rewriter{
		claudeClient: client,
	}
}

func (r *Rewriter) RewriteQuery(ctx context.Context, originalQuery string) (string, error) {
	// Building a rewrite prompt

	prompt := fmt.Sprintf(`You are a query optimization assistant for a product documentation system.

Original query: "%s"

Rewrite this query to be:
1. More specific and clear
2. Better for semantic search
3. Free of typos and grammatical errors
4. Focused on technical documentation

Return ONLY the rewritten query, nothing else.`, originalQuery)

	response, err := r.claudeClient.InvokeModel(ctx, bedrock.ClaudeRequest{
		Prompt:      prompt,
		MaxTokens:   200,
		Temperature: 0.2, // Low temperature for consistent rewrite
	})

	if err != nil {
		log.Error().Err(err).Msg("Failed to rewrite query")
		// Fallback to original query
		return originalQuery, nil
	}

	rewriteQuery := strings.TrimSpace(response.Content)

	log.Info().
		Str("original", originalQuery).
		Str("rewritten", rewriteQuery).
		Msg("Query rewrite")

	return rewriteQuery, nil
}

func (r *Rewriter) ExpandQuery(query string, domain string) string {
	return fmt.Sprintf("%s in the context of %s", query, domain)
}

func (r *Rewriter) SimplifyQuery(ctx context.Context, complexQuery string) ([]string, error) {
	prompt := fmt.Sprintf(`Break down this complex question into 2-3 simpler sub-questions:

"%s"

Return only the sub-question, one per line. `, complexQuery)

	response, err := r.claudeClient.InvokeModel(ctx, bedrock.ClaudeRequest{
		Prompt:      prompt,
		MaxTokens:   300,
		Temperature: 0.3,
	})

	if err != nil {
		// Fallback in case we get an model invoke error
		return []string{complexQuery}, nil
	}

	// Split response into individual questions
	subQueries := strings.Split(strings.TrimSpace(response.Content), "\n")
	return subQueries, nil

}
