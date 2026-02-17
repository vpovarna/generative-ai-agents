package agent

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/povarna/generative-ai-with-go/kg-agent/internal/bedrock"
	"github.com/povarna/generative-ai-with-go/kg-agent/internal/conversation"
	"github.com/povarna/generative-ai-with-go/kg-agent/internal/rewrite"
	"github.com/povarna/generative-ai-with-go/kg-agent/internal/strategy"
	"github.com/rs/zerolog/log"
)

type Service struct {
	bedrockClient     *bedrock.Client
	rewriter          *rewrite.Rewriter
	modelID           string
	searchClient      *SearchClient
	conversationStore conversation.ConversationStore
	retrievalStrategy *strategy.RetrievalStrategy
}

func NewService(
	bedrockClient *bedrock.Client,
	modelID string,
	rewriter *rewrite.Rewriter,
	searchClient *SearchClient,
	conversationStore conversation.ConversationStore,
	retrievalStrategy *strategy.RetrievalStrategy) *Service {
	return &Service{
		bedrockClient:     bedrockClient,
		rewriter:          rewriter,
		modelID:           modelID,
		searchClient:      searchClient,
		conversationStore: conversationStore,
		retrievalStrategy: retrievalStrategy,
	}
}

func (s *Service) Query(ctx context.Context, queryRequest QueryRequest) (QueryResponse, error) {
	// Get or create session
	sessionID, conversationHistory := s.getOrCreateSession(ctx, queryRequest)

	// Decide if we need to search external documentation
	decision := s.retrievalStrategy.Decide(ctx, queryRequest.Prompt, conversationHistory)

	// Rewrite query
	rewrittenQuery := s.rewriteQuery(ctx, queryRequest)
	var searchResults []SearchResult

	if decision.ShouldSearch {
		searchResults = s.search(ctx, rewrittenQuery)
	} else {
		searchResults = nil
	}
	// Format context and build enhanced prompt
	enhancedPrompt := s.buildPromptWithContext(rewrittenQuery, searchResults, conversationHistory)

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

func (s *Service) search(ctx context.Context, rewrittenQuery string) []SearchResult {
	searchResults, err := s.searchClient.HybridSearch(ctx, rewrittenQuery, 5)
	if err != nil {
		log.Warn().Err(err).Msg("Search failed, continuing without context")
		searchResults = nil // Continue without context
	}
	return searchResults
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
		searchResults = s.search(ctx, rewrittenQuery)
	} else {
		searchResults = nil
	}

	// Format context and build enhanced prompt
	enhancedPrompt := s.buildPromptWithContext(rewrittenQuery, searchResults, conversationHistory)

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
