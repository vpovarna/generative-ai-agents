package agent

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/povarna/generative-ai-with-go/kg-agent/internal/bedrock"
	"github.com/povarna/generative-ai-with-go/kg-agent/internal/rewrite"
	"github.com/rs/zerolog/log"
)

type Service struct {
	bedrockClient *bedrock.Client
	rewriter      *rewrite.Rewriter
	modelID       string
	searchClient  *SearchClient
}

func NewService(bedrockClient *bedrock.Client, modelID string, rewriter *rewrite.Rewriter, searchClient *SearchClient) *Service {
	return &Service{
		bedrockClient: bedrockClient,
		rewriter:      rewriter,
		modelID:       modelID,
		searchClient:  searchClient,
	}
}

func (s *Service) Query(ctx context.Context, queryRequest QueryRequest) (QueryResponse, error) {
	// Query rewrite
	rewrittenQuery, err := s.rewriter.RewriteQuery(ctx, queryRequest.Prompt)
	if err != nil {
		log.Error().Err(err).Msg("Query rewrite failed")
		// Continue with original query
		rewrittenQuery = queryRequest.Prompt
	}

	// Search for relevant context
	searchResults, err := s.searchClient.HybridSearch(ctx, rewrittenQuery, 5)
	if err != nil {
		log.Warn().Err(err).Msg("Search failed, continuing without context")
		searchResults = nil // Continue without context
	}

	// Format context and build enhanced prompt
	enhancedPrompt := s.buildPromptWithContext(rewrittenQuery, searchResults)

	// 4. Call Claude with context
	response, err := s.bedrockClient.InvokeModel(ctx, bedrock.ClaudeRequest{
		Prompt:      enhancedPrompt,
		MaxTokens:   queryRequest.MaxToken,
		Temperature: queryRequest.Temperature,
	})
	if err != nil {
		log.Error().Err(err).Msg("Failed to invoke Claude")
		return QueryResponse{}, err
	}

	queryResponse := QueryResponse{
		Content:    response.Content,
		StopReason: response.StopReason,
		Model:      s.modelID,
	}

	return queryResponse, nil
}

func (s *Service) QueryStream(ctx context.Context, queryRequest QueryRequest, flusher http.Flusher, writer io.Writer) error {
	// Query rewrite
	rewrittenQuery, err := s.rewriter.RewriteQuery(ctx, queryRequest.Prompt)
	if err != nil {
		log.Error().Err(err).Msg("Query rewrite failed")
		// Continue with original query
		rewrittenQuery = queryRequest.Prompt
	}

	// Search for relevant context
	searchResults, err := s.searchClient.HybridSearch(ctx, rewrittenQuery, 5)
	if err != nil {
		log.Warn().Err(err).Msg("Search failed, continuing without context")
		searchResults = nil // Continue without context
	}

	// Format context and build enhanced prompt
	enhancedPrompt := s.buildPromptWithContext(rewrittenQuery, searchResults)

	// Send starting event
	startEvent := SSEEvent{
		Event: "start",
		Data: StreamStartEvent{
			Model: s.modelID,
		},
	}

	if formatEvent, err := startEvent.Format(); err == nil {
		fmt.Fprint(writer, formatEvent)
		flusher.Flush()
	}

	// Call Claude with context (streaming)
	response, err := s.bedrockClient.InvokeModelStream(ctx, bedrock.ClaudeRequest{
		Prompt:      enhancedPrompt, // ← Changed: use enhanced prompt with context
		MaxTokens:   queryRequest.MaxToken,
		Temperature: queryRequest.Temperature,
	}, func(chunk string) error {
		// Send chunk event
		chunkEvent := SSEEvent{
			Event: "chunk",
			Data: StreamChunkEvent{
				Text: chunk,
			},
		}

		if formatEvent, ok := chunkEvent.Format(); ok == nil {
			fmt.Fprint(writer, formatEvent)
			flusher.Flush()
		}

		return nil
	})

	if err != nil {
		// Send error event
		errorEvent := SSEEvent{
			Event: "error",
			Data: StreamErrorEvent{
				Error: err.Error(),
			},
		}

		if formatEvent, ok := errorEvent.Format(); ok == nil {
			fmt.Fprint(writer, formatEvent)
			flusher.Flush()
		}
		return err
	}

	// Send end event
	doneEvent := SSEEvent{
		Event: "done",
		Data: StreamDoneEvent{
			StopReason: response.StopReason,
		},
	}
	if formatEvent, ok := doneEvent.Format(); ok == nil {
		fmt.Fprint(writer, formatEvent)
		flusher.Flush()
	}

	return nil
}

func (s *Service) buildPromptWithContext(userQuery string, searchResult []SearchResult) string {
	if len(searchResult) == 0 {
		// if no content retrieved from search API, return original query
		return userQuery
	}

	var builder strings.Builder

	for i, result := range searchResult {
		builder.WriteString(fmt.Sprintf("[%d] (relevance: %.2f)\n", i+1, result.Score))
		builder.WriteString(result.Content)
		builder.WriteString("\n\n")
	}

	context := builder.String()

	return fmt.Sprintf(`You are a helpful documentation assistant.
Use the following documentation excerpts to answer the user's question:
	
<context>
%s
</context>
	
User question: %s
	
Provide a clear, accurate answer based on the documentation provided. If the documentation doesn't contain the answer, say so.`, context, userQuery)
}
