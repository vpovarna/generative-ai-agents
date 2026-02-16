package agent

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/povarna/generative-ai-with-go/kg-agent/internal/bedrock"
	"github.com/povarna/generative-ai-with-go/kg-agent/internal/rewrite"
	"github.com/rs/zerolog/log"
)

type Service struct {
	bedrockClient *bedrock.Client
	rewriter      *rewrite.Rewriter
	modelID       string
}

func NewService(bedrockClient *bedrock.Client, rewriter *rewrite.Rewriter, modelID string) *Service {
	return &Service{
		bedrockClient: bedrockClient,
		rewriter:      rewriter,
		modelID:       modelID,
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

	response, err := s.bedrockClient.InvokeModel(ctx, bedrock.ClaudeRequest{
		Prompt:      rewrittenQuery,
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

	response, err := s.bedrockClient.InvokeModelStream(ctx, bedrock.ClaudeRequest{
		Prompt:      rewrittenQuery,
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
