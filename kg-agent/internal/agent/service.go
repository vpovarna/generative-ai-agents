package agent

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/povarna/generative-ai-agents/kg-agent/internal/bedrock"
	"github.com/povarna/generative-ai-agents/kg-agent/internal/cache"
	"github.com/povarna/generative-ai-agents/kg-agent/internal/conversation"
	"github.com/povarna/generative-ai-agents/kg-agent/internal/rewrite"
	"github.com/povarna/generative-ai-agents/kg-agent/internal/strategy"
	"github.com/rs/zerolog/log"
)

type Service struct {
	bedrockClient     *bedrock.Client
	miniClient        *bedrock.Client
	rewriter          *rewrite.Rewriter
	modelID           string
	searchClient      *SearchClient
	conversationStore conversation.ConversationStore
	retrievalStrategy *strategy.RetrievalStrategy
	searchCache       cache.SearchCache
}

func NewService(
	bedrockClient *bedrock.Client,
	miniClient *bedrock.Client,
	modelID string,
	rewriter *rewrite.Rewriter,
	searchClient *SearchClient,
	conversationStore conversation.ConversationStore,
	retrievalStrategy *strategy.RetrievalStrategy,
	searchCache cache.SearchCache) *Service {
	return &Service{
		bedrockClient:     bedrockClient,
		miniClient:        miniClient,
		rewriter:          rewriter,
		modelID:           modelID,
		searchClient:      searchClient,
		conversationStore: conversationStore,
		retrievalStrategy: retrievalStrategy,
		searchCache:       searchCache,
	}
}

func (s *Service) Query(ctx context.Context, queryRequest QueryRequest) (QueryResponse, error) {
	// Get or create session
	sessionID, conversationHistory := s.getOrCreateSession(ctx, queryRequest)

	// Decide if we need to search external documentation
	decision := s.retrievalStrategy.Decide(ctx, queryRequest.Prompt, conversationHistory)

	log.Info().Interface("decision", decision).Msg("Decision")

	// Rewrite query
	rewrittenQuery := s.rewriteQuery(ctx, queryRequest)
	var searchResults []SearchResult

	if decision.ShouldSearch {
		searchResults = s.search(ctx, queryRequest.Prompt, rewrittenQuery)
	} else {
		searchResults = nil
	}
	// Format context and build enhanced prompt
	enhancedPrompt := s.buildPromptWithContext(rewrittenQuery, searchResults, conversationHistory)

	selectedClient := s.selectModelForAnswer(decision, len(searchResults) > 0)

	// 4. Call Claude with context
	response, err := selectedClient.InvokeModel(ctx, bedrock.ClaudeRequest{
		Prompt:      enhancedPrompt,
		MaxTokens:   queryRequest.MaxToken,
		Temperature: queryRequest.Temperature,
	})
	if err != nil {
		log.Error().Err(err).Msg("Failed to invoke Claude")
		return QueryResponse{}, err
	}

	// Save user message and assistant response to conversation history
	s.saveConversationMessages(ctx, sessionID, queryRequest, response)

	queryResponse := QueryResponse{
		SessionID:  sessionID,
		Content:    response.Content,
		StopReason: response.StopReason,
		Model:      s.modelID,
	}

	return queryResponse, nil
}

func (s *Service) QueryStream(ctx context.Context, queryRequest QueryRequest, flusher http.Flusher, writer io.Writer) error {
	// Get or create session
	sessionID, conversationHistory := s.getOrCreateSession(ctx, queryRequest)

	// Decide if we need to search external documentation
	decision := s.retrievalStrategy.Decide(ctx, queryRequest.Prompt, conversationHistory)

	// Query rewrite
	rewrittenQuery := s.rewriteQuery(ctx, queryRequest)

	// Conditionally search
	var searchResults []SearchResult
	if decision.ShouldSearch {
		searchResults = s.search(ctx, queryRequest.Prompt, rewrittenQuery)
	} else {
		searchResults = nil
	}

	// Format context and build enhanced prompt
	enhancedPrompt := s.buildPromptWithContext(rewrittenQuery, searchResults, conversationHistory)

	// Select model based on query complexity
	selectedClient := s.selectModelForAnswer(decision, len(searchResults) > 0)

	// Send starting event
	startEvent := SSEEvent{
		Event: "start",
		Data: StreamStartEvent{
			SessionID: sessionID,
			Model:     s.modelID,
		},
	}

	if formatEvent, err := startEvent.Format(); err == nil {
		fmt.Fprint(writer, formatEvent)
		flusher.Flush()
	}

	// Call Claude with context (streaming)
	response, err := selectedClient.InvokeModelStream(ctx, bedrock.ClaudeRequest{
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

	// Save user message and assistant response to conversation history
	s.saveConversationMessages(ctx, sessionID, queryRequest, response)

	return nil
}

func (s *Service) rewriteQuery(ctx context.Context, queryRequest QueryRequest) string {
	rewrittenQuery, err := s.rewriter.RewriteQuery(ctx, queryRequest.Prompt)
	if err != nil {
		log.Error().Err(err).Msg("Query rewrite failed")
		// Continue with original query
		rewrittenQuery = queryRequest.Prompt
	}
	return rewrittenQuery
}

func (s *Service) search(ctx context.Context, originalQuery string, rewrittenQuery string) []SearchResult {
	// Use ORIGINAL query for cache key (rewrite is non-deterministic)
	cacheKey := s.generateCacheKey(originalQuery, "hybrid", 5)
	value, err := s.searchCache.Get(ctx, cacheKey)

	if err != nil {
		log.Info().Msg("Cache miss!. Calling search api... ")
		searchResults, err := s.searchClient.HybridSearch(ctx, rewrittenQuery, 5)
		if err != nil {
			log.Warn().Err(err).Msg("Search failed, continuing without context")
			searchResults = nil // Continue without context
		}

		// Update cache
		data, err := json.Marshal(searchResults)
		if err == nil {
			if err := s.searchCache.Set(ctx, cacheKey, data, 30*time.Minute); err != nil {
				log.Warn().Err(err).Msg("Unable to cache query search result")
			}
		}

		return searchResults
	}

	// CACHE HIT
	log.Info().Str("original_query", originalQuery).Msg("Cache hit!")
	var searchResult []SearchResult
	if err := json.Unmarshal(value, &searchResult); err != nil {
		log.Error().Err(err).Msg("Unable to deserialize response")
		return nil
	}

	return searchResult

}

func (s *Service) saveConversationMessages(ctx context.Context, sessionID string, queryRequest QueryRequest, response *bedrock.ClaudeResponse) {
	if sessionID == "" {
		return // No session to save to
	}

	userMsg := conversation.Message{
		Role:      "user",
		Content:   queryRequest.Prompt,
		Timestamp: time.Now(),
	}
	if err := s.conversationStore.AddMessage(ctx, sessionID, userMsg); err != nil {
		log.Warn().Err(err).Msg("Failed to save user message")
	}

	assistantMsg := conversation.Message{
		Role:      "assistant",
		Content:   response.Content,
		Timestamp: time.Now(),
	}
	if err := s.conversationStore.AddMessage(ctx, sessionID, assistantMsg); err != nil {
		log.Warn().Err(err).Msg("Failed to save assistant message")
	}

}

func (s *Service) getOrCreateSession(ctx context.Context, queryRequest QueryRequest) (string, *conversation.Conversation) {
	var sessionID string
	var conversationHistory *conversation.Conversation
	var err error

	if queryRequest.SessionID == "" {
		// Create new Session
		session, err := s.conversationStore.CreateSession(ctx)
		if err != nil {
			log.Error().Err(err).Msg("failed to create session")
		} else {
			sessionID = session.ID
		}
	} else {
		// Retrieve conversation
		sessionID = queryRequest.SessionID
		conversationHistory, err = s.conversationStore.GetConversation(ctx, sessionID)
		if err != nil {
			log.Warn().Err(err).Str("sessionID", sessionID).Msg("failed to retrieve conversation, continuing without history")
			conversationHistory = nil
		}
	}
	return sessionID, conversationHistory
}

func (s *Service) buildPromptWithContext(userQuery string, searchResult []SearchResult, conversationHistory *conversation.Conversation) string {
	historySection := ""

	if conversationHistory != nil && len(conversationHistory.Messages) > 0 {
		var hb strings.Builder
		maxMessages := 10

		hb.WriteString("Conversation history:\n")
		messages := conversationHistory.Messages
		if len(messages) > maxMessages {
			messages = messages[len(messages)-maxMessages:]
		}

		for _, msg := range messages {
			hb.WriteString(fmt.Sprintf("%s: %s\n", msg.Role, msg.Content))
		}
		historySection = hb.String() + "\n"
	}

	docsSection := ""
	if len(searchResult) > 0 {
		var db strings.Builder
		db.WriteString("Relevant documentation:\n<context>\n")
		for i, r := range searchResult {
			db.WriteString(fmt.Sprintf("[%d] (relevance: %.2f)\n%s\n\n", i+1, r.Score, r.Content))
		}
		db.WriteString("</context>\n")
		docsSection = db.String() + "\n"
	}

	return fmt.Sprintf(`You are a helpful documentation assistant.
	
	%s%sCurrent question: %s
	
	Provide a clear, accurate answer based on the information provided.`,
		historySection, docsSection, userQuery)
}

func (s *Service) selectModelForAnswer(decision strategy.Decision, hasSearchResults bool) *bedrock.Client {
	// Simple queries without search → Use Haiku
	if !decision.ShouldSearch && decision.Confidence > 0.90 && !hasSearchResults {
		log.Info().Msg("Using Haiku for simple query")
		return s.miniClient
	}
	// Complex queries or with search results → Use Sonnet
	log.Info().Msg("Using Sonnet for complex query")
	return s.bedrockClient
}

func (s *Service) generateCacheKey(query string, searchType string, limit int) string {
	input := fmt.Sprintf("%s:%s:%d", query, searchType, limit)
	hash := sha256.Sum256([]byte(input))
	return fmt.Sprintf("%x", hash)
}

// ClearSearchCache clears all cached search results
func (s *Service) ClearSearchCache(ctx context.Context) error {
	return s.searchCache.Clear(ctx)
}
